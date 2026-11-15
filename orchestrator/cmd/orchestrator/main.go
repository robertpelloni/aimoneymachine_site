package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/robertpelloni/hustle/hustle/affiliate"
	"github.com/robertpelloni/hustle/hustle/automation"
	"github.com/robertpelloni/hustle/hustle/bi"
	"github.com/robertpelloni/hustle/hustle/careers"
	"github.com/robertpelloni/hustle/hustle/confluence"
	"github.com/robertpelloni/hustle/hustle/content"
	"github.com/robertpelloni/hustle/hustle/crm"
	"github.com/robertpelloni/hustle/hustle/curation"
	"github.com/robertpelloni/hustle/hustle/design"
	"github.com/robertpelloni/hustle/hustle/devagency"
	"github.com/robertpelloni/hustle/hustle/domains"
	"github.com/robertpelloni/hustle/hustle/ecommerce"
	"github.com/robertpelloni/hustle/hustle/finance"
	"github.com/robertpelloni/hustle/hustle/legal"
	"github.com/robertpelloni/hustle/hustle/localization"
	"github.com/robertpelloni/hustle/hustle/marketing"
	"github.com/robertpelloni/hustle/hustle/media"
	"github.com/robertpelloni/hustle/hustle/outreach"
	"github.com/robertpelloni/hustle/hustle/pod"
	"github.com/robertpelloni/hustle/hustle/products"
	"github.com/robertpelloni/hustle/hustle/publisher"
	"github.com/robertpelloni/hustle/hustle/publishing"
	"github.com/robertpelloni/hustle/hustle/qa"
	"github.com/robertpelloni/hustle/hustle/realestate"
	"github.com/robertpelloni/hustle/hustle/research"
	"github.com/robertpelloni/hustle/hustle/retail"
	"github.com/robertpelloni/hustle/hustle/saas"
	"github.com/robertpelloni/hustle/hustle/social"
	"github.com/robertpelloni/hustle/hustle/stocks"
	"github.com/robertpelloni/hustle/hustle/support"
	"github.com/robertpelloni/hustle/hustle/trading"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	hustleTask := flag.String("hustle", "", "Run a specific hustle module")
	hustleURI := flag.String("uri", "", "Run a hustle module via hustle:// URI")
	syncMode := flag.Bool("sync", false, "Run repository synchronization protocol")
	params := flag.String("params", "{}", "JSON parameters for the hustle module")
	dashboard := flag.Bool("dashboard", false, "Start the live terminal dashboard")
	daemon := flag.Bool("daemon", false, "Run the Orchestrator as a background task scheduler")
	interactive := flag.Bool("interactive", false, "Launch the interactive command menu")
	apiPort := flag.String("api", "", "Start the HTTP API on specified port (e.g. 8082)")
	realPrices := flag.Bool("real-prices", false, "Use real-world market prices via CoinGecko/Binance/Kraken/Coinbase")
	seedURL := flag.String("seed", "", "URL of a seed node for mesh federation")
	agentMode := flag.Bool("agent", false, "Run as autonomous agent loop (LLM-driven decision making)")
	agentType := flag.String("agent-type", "general", "Agent hustle focus: general, research, content, trading, social")
	agentIter := flag.Int("agent-iterations", 20, "Max iterations for agent loop")
	autoPlan := flag.Bool("autoplan", false, "LLM generates and executes a strategic hustle plan")
	dryRun := flag.Bool("dry-run", false, "Execute in dry-run mode (no external mutations)")
	flag.Parse()

	// Tee stdout into GlobalLogBuffer when serving HTTP or running as a daemon so
	// GET /api/log always has live output even under systemd journald.
	if *apiPort != "" || *daemon {
		orchestrator.InstallStdoutTee()
	}

	versionData, err := os.ReadFile("VERSION.md")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionData))
	}

	fmt.Printf("=== AI Hustle Machine Orchestrator v%s ===\n", version)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	orch := orchestrator.NewOrchestrator()
	orch.Version = version
	_ = orch.Load("memory.json")
	orch.DryRun = *dryRun

	// Load or generate persistent Ed25519 node identity (node.key is gitignored).
	if id, ierr := orchestrator.NewIdentity(); ierr == nil {
		orch.Identity = id
		fmt.Printf("[Identity] Node DID: %s\n", id.GetDID()[:minInt(len(id.GetDID()), 36)]+"...")
	}

	// Seed default multi-armed bandit prompt variations for telemetry & prompt optimization.
	orch.Optimizer.AddVariation("social", "Write a punchy, high-engagement post with a contrarian hook about: %s")
	orch.Optimizer.AddVariation("social", "Write a concise data-backed insight post with a practical takeaway about: %s")
	orch.Optimizer.AddVariation("content", "Write an authoritative, step-by-step operator playbook on: %s")
	orch.Optimizer.AddVariation("affiliate", "Write an honest, conversion-focused mini-review highlighting ROI for: %s")

	db, err := orchestrator.NewSQLiteStore("hustle.db")
	if err != nil {
		fmt.Printf("[SQLite] Failed to initialize database: %v\n", err)
	} else {
		orch.DB = db
	}

	// Wire up real LLM via OpenAI-compatible proxy -> optional Waterfall fallback -> SHA-256 CachingLLM.
	llmProvider := orchestrator.NewOpenAICompatProvider()
	if model, err := llmProvider.DetectModelWithRetry(3, 3*time.Second); err == nil {
		fmt.Printf("[LLM] Connected to LLM: %s\n", model)
		var baseLLM orchestrator.LLMProvider = llmProvider
		if fbURL := os.Getenv("LLM_FALLBACK_BASE_URL"); fbURL != "" {
			fb := orchestrator.NewOpenAICompatProviderWithBaseURL(fbURL)
			if fbModel, ferr := fb.DetectModelWithRetry(2, 2*time.Second); ferr == nil {
				fmt.Printf("[LLM] Fallback provider ready: %s (%s)\n", fbURL, fbModel)
				baseLLM = &orchestrator.WaterfallLLM{Providers: []orchestrator.LLMProvider{llmProvider, fb}}
			} else {
				fmt.Printf("[LLM] Fallback provider %s unavailable (%v); primary only.\n", fbURL, ferr)
			}
		}
		cacheLLM := orchestrator.NewCachingLLM(baseLLM)
		if ttlStr := os.Getenv("LLM_CACHE_TTL_MIN"); ttlStr != "" {
			if mins, perr := strconv.Atoi(ttlStr); perr == nil && mins > 0 {
				cacheLLM.SetMaxAge(time.Duration(mins) * time.Minute)
			}
		}
		orch.Cache = cacheLLM
		orch.LLM = cacheLLM
		orch.Embedder = orchestrator.NewOpenAICompatEmbedder()
	} else {
		fmt.Printf("[LLM] Local LLM not available (%v). Falling back to MockLLM.\n", err)
		cacheLLM := orchestrator.NewCachingLLM(orch.LLM)
		orch.Cache = cacheLLM
	}

	protocol := orchestrator.NewHustleProtocol()
	chainManager := orchestrator.NewChainManager(orch, protocol)
	_ = chainManager.LoadState("chains.json")
	discoverer := orchestrator.NewChainDiscoverer(orch, chainManager)
	broker := orchestrator.NewA2ABroker(orch)
	swarm := orchestrator.NewMemorySwarm(orch, broker)
	multiAgent := orchestrator.NewMultiAgentOrchestrator(orch, protocol, broker)
	scheduler := orchestrator.NewScheduler(orch)

	// ── Initialize Trading Module ──
	var fetcher trading.PriceFetcher = &trading.MockPriceFetcher{}
	priceSource := "mock"
	if *realPrices {
		switch strings.ToLower(os.Getenv("TRADING_PRICE_SOURCE")) {
		case "binance":
			priceSource = "binance"
			fmt.Println("[Trading] Real-world price fetching enabled via Binance.")
			fetcher = trading.NewBinancePriceFetcher()
		case "kraken":
			priceSource = "kraken"
			fmt.Println("[Trading] Real-world price fetching enabled via Kraken.")
			fetcher = trading.NewKrakenPriceFetcher()
		case "coinbase":
			priceSource = "coinbase"
			fmt.Println("[Trading] Real-world price fetching enabled via Coinbase.")
			fetcher = trading.NewCoinbasePriceFetcher()
		default:
			priceSource = "coingecko"
			fmt.Println("[Trading] Real-world price fetching enabled via CoinGecko.")
			fetcher = trading.NewCoinGeckoFetcher()
		}
	}
	traderModule := &trading.TradingModule{
		Orchestrator: orch,
		Broker:       broker,
		Fetcher:      fetcher,
		HistoryFile:  "trading_history.json",
	}
	if err := traderModule.LoadHistory(); err != nil {
		fmt.Printf("[Trading] Could not restore history (starting fresh): %v\n", err)
	}

	contentModule := content.NewContentModule(orch, "output/content")
	contentCalendar := publisher.NewContentCalendar()
	_ = contentCalendar.LoadCalendar()

	// ── Mesh Event Listeners ──
	alphaCh := broker.SubscribeTopic("alpha_discovery")
	go func() {
		for msg := range alphaCh {
			fmt.Printf("[Mesh] Received Alpha Discovery: %s\n", msg.Payload)
			traderModule.AddToWatchlist(msg.Payload)
		}
	}()

	syncCh := broker.SubscribeTopic("swarm_sync")
	go func() {
		for msg := range syncCh {
			fmt.Printf("[Mesh] Swarm Sync Notification from %s\n", msg.Source)
			swarm.HandleSyncRequest(msg.Source)
		}
	}()

	agentCh := broker.Subscribe("local-node")
	go func() {
		for msg := range agentCh {
			if strings.HasPrefix(msg.Payload, "hustle://") {
				fmt.Printf("[A2A] Executing Mesh Protocol URI from %s: %s\n", msg.Source, msg.Payload)
				_ = protocol.HandleURI(msg.Payload)
			}
		}
	}()

	// ── Register ALL 36 Protocol Handlers ──
	protocol.Register("research", func(p url.Values) error {
		query := defVal(p.Get("query"), "AI Trends")
		searcher := research.NewResearchSearch(research.DuckDuckGo, orch, broker)
		_, err := searcher.Query(query)
		return err
	})

	protocol.Register("curation", func(p url.Values) error {
		topic := defVal(p.Get("topic"), "AI")
		feeds := orch.RSSFeeds
		if len(feeds) == 0 {
			feeds = []string{"https://news.ycombinator.com/rss"}
		}
		c := &curation.CurationModule{
			Orchestrator: orch,
			Fetcher:      curation.NewRSSFetcher(),
			Feeds:        feeds,
		}
		return c.Curate(topic)
	})

	protocol.Register("rss", func(p url.Values) error {
		action := defVal(p.Get("action"), "list")
		feedURL := strings.TrimSpace(p.Get("url"))
		switch action {
		case "add":
			if feedURL == "" {
				return fmt.Errorf("rss add requires url")
			}
			for _, existing := range orch.RSSFeeds {
				if existing == feedURL {
					return nil
				}
			}
			orch.RSSFeeds = append(orch.RSSFeeds, feedURL)
			_ = orch.Save("memory.json")
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("rss-add-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Added RSS feed: %s (total: %d)", feedURL, len(orch.RSSFeeds)),
				Timestamp: time.Now(),
				Tags:      []string{"rss", "curation"},
			})
			return nil
		case "remove":
			if feedURL == "" {
				return fmt.Errorf("rss remove requires url")
			}
			filtered := make([]string, 0, len(orch.RSSFeeds))
			for _, f := range orch.RSSFeeds {
				if f != feedURL {
					filtered = append(filtered, f)
				}
			}
			orch.RSSFeeds = filtered
			_ = orch.Save("memory.json")
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("rss-rm-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Removed RSS feed: %s", feedURL),
				Timestamp: time.Now(),
				Tags:      []string{"rss", "curation"},
			})
			return nil
		default:
			fmt.Printf("[RSS] Active feeds: %v\n", orch.RSSFeeds)
			return nil
		}
	})

	protocol.Register("social", func(p url.Values) error {
		platform := defVal(p.Get("platform"), social.DefaultPlatform())
		topic := defVal(p.Get("topic"), "AI")
		contentStr := p.Get("content")

		var provider social.Provider
		switch strings.ToLower(platform) {
		case "linkedin":
			provider = social.NewLinkedInProvider(os.Getenv("LINKEDIN_ACCESS_TOKEN"), os.Getenv("LINKEDIN_AUTHOR_URN"))
		case "bluesky":
			provider = social.NewBlueskyProvider(os.Getenv("BLUESKY_HANDLE"), os.Getenv("BLUESKY_APP_PASSWORD"))
		default:
			provider = social.NewTwitterProvider(
				os.Getenv("TWITTER_API_KEY"),
				os.Getenv("TWITTER_API_SECRET"),
				os.Getenv("TWITTER_ACCESS_TOKEN"),
				os.Getenv("TWITTER_ACCESS_SECRET"),
			)
		}
		if orch.DryRun {
			provider.SetDryRun(true)
		}
		if contentStr != "" {
			return provider.Post(orch, platform, contentStr)
		}

		_, varIdx, _ := orch.Optimizer.GetPrompt("social")
		generatedContent := social.GenerateContent(orch, topic)
		affModule := affiliate.NewModule(orch)
		generatedContent = affModule.InjectAffiliateLink(generatedContent, topic)
		generatedContent = social.FormatForPlatform(generatedContent, platform)

		err := provider.Post(orch, platform, generatedContent)
		if err == nil {
			orch.Optimizer.RecordWin("social", varIdx)
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("social-%s-%d", platform, time.Now().Unix()),
				Content:   fmt.Sprintf("Posted to %s: %s", platform, generatedContent),
				Timestamp: time.Now(),
				Tags:      []string{"social", platform},
			})
			orch.Ledger.Add(orchestrator.Transaction{
				Amount: 0.01,
				Type:   orchestrator.Expense,
				Hustle: "SocialMedia",
				Note:   fmt.Sprintf("API post to %s", platform),
			})
		} else {
			orch.Optimizer.RecordLoss("social", varIdx)
		}
		return err
	})

	protocol.Register("trading", func(p url.Values) error {
		action := p.Get("action")
		symbol := defVal(p.Get("symbol"), "BTC")
		if action == "watchlist" {
			traderModule.AddToWatchlist(symbol)
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("watchlist-%s-%d", symbol, time.Now().Unix()),
				Content:   fmt.Sprintf("Added %s to active trading watchlist", strings.ToUpper(symbol)),
				Timestamp: time.Now(),
				Tags:      []string{"trading", "watchlist", strings.ToUpper(symbol)},
			})
			return nil
		}
		return traderModule.ExecuteFor(symbol)
	})

	protocol.Register("stocks", func(p url.Values) error {
		symbol := defVal(p.Get("symbol"), "NVDA")
		action := defVal(p.Get("action"), "analyze")
		mod := stocks.NewStockModule(orch, broker)
		if action == "order" {
			if orch.DryRun {
				orch.L1.Add(orchestrator.MemoryEntry{
					ID:        fmt.Sprintf("stock-dry-%d", time.Now().Unix()),
					Content:   fmt.Sprintf("[DRY-RUN] Alpaca %s order for %s", defVal(p.Get("side"), "buy"), symbol),
					Timestamp: time.Now(),
					Tags:      []string{"stocks", "alpaca", symbol},
				})
				return nil
			}
			qty := parseFloat(p.Get("qty"), 1.0)
			return stocks.NewAlpacaExecutor().ExecuteOrder(symbol, defVal(p.Get("side"), "buy"), qty)
		}
		_, err := mod.AnalyzeStock(symbol)
		return err
	})

	protocol.Register("content", func(p url.Values) error {
		action := p.Get("action")
		topic := p.Get("topic")

		switch action {
		case "cms":
			cfg := content.DefaultSiteConfig()
			if theme := p.Get("theme"); theme != "" {
				cfg.Theme = theme
			}
			if title := p.Get("title"); title != "" {
				cfg.Title = title
			}
			if err := content.GenerateSite(contentModule.OutputDir, cfg); err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("cms-site-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Generated static pSEO HTML site + RSS feed in %s (theme: %s)", cfg.OutputDir, cfg.Theme),
				Timestamp: time.Now(),
				Tags:      []string{"content", "cms", "static-site"},
			})
			return nil
		case "brainstorm":
			niche := defVal(p.Get("niche"), defVal(topic, "AI & automation"))
			ideas, err := contentModule.GenerateTopicIdeas(niche, 8)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("brainstorm-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Topic Ideas for %s: %s", niche, strings.Join(ideas, " | ")),
				Timestamp: time.Now(),
				Tags:      []string{"content", "brainstorm", niche},
			})
			return nil
		case "expand":
			if topic == "" {
				topic = "AI automation trends"
			}
			targetChars := 100000
			initial := "Initial core article about " + topic + ". Expanding toward comprehensive deep-dive pillar guide."
			if files, err := os.ReadDir(contentModule.OutputDir); err == nil {
				for _, file := range files {
					if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") && strings.Contains(strings.ToLower(file.Name()), strings.ToLower(strings.ReplaceAll(topic, " ", "-"))) {
						path := filepath.Join(contentModule.OutputDir, file.Name())
						if data, readErr := os.ReadFile(path); readErr == nil {
							initial = string(data)
							break
						}
					}
				}
			}
			expanded, err := contentModule.ExpandContent(topic, initial, targetChars)
			if err != nil {
				return err
			}
			affModule := affiliate.NewAffiliateModule(orch)
			_, _ = affModule.DiscoverProducts(topic)
			inserter := publisher.NewAffiliateInserter()
			expandedWithLinks := inserter.ProcessContent(expanded)
			fmt.Printf("[Content] Expansion & affiliate link pass complete (%d bytes)\n", len(expandedWithLinks))
			return nil
		}

		if topic == "" {
			topic = "AI automation trends"
		}
		contentType := defVal(p.Get("type"), "blog")
		niche := defVal(p.Get("niche"), "AI & automation")
		kwList := []string{"AI", "automation"}
		if keywords := p.Get("keywords"); keywords != "" {
			kwList = strings.Split(keywords, ",")
		}
		_, varIdx, _ := orch.Optimizer.GetPrompt("content")
		_, err := contentModule.Generate(content.ContentRequest{
			Topic:       topic,
			Type:        content.ContentType(contentType),
			Keywords:    kwList,
			TargetWords: 800,
			Niche:       niche,
		})
		if err == nil {
			orch.Optimizer.RecordWin("content", varIdx)
		} else {
			orch.Optimizer.RecordLoss("content", varIdx)
		}
		return err
	})

	protocol.Register("swarm", func(p url.Values) error {
		action := p.Get("action")
		peerID := defVal(p.Get("peer_id"), "unknown-peer")
		switch action {
		case "sync":
			swarm.Sync()
		case "sync_request":
			swarm.HandleSyncRequest(peerID)
		case "provide_index":
			swarm.ReconcileIndex(peerID, p.Get("data"))
		case "request_entry":
			swarm.ProvideEntry(peerID, p.Get("id"))
		case "provide_entry":
			orch.L2.Add(orchestrator.MemoryEntry{
				ID:        p.Get("id"),
				Content:   p.Get("content"),
				Timestamp: time.Now(),
				Tags:      []string{"swarm", "received", "from:" + peerID},
			})
		case "aggregate":
			swarm.AggregateStatus()
		case "status_request":
			swarm.ProvideStatus(peerID)
		case "provide_status":
			swarm.HandleStatusResponse(peerID, p.Get("data"))
		case "set_goal":
			if amount := parseFloat(p.Get("amount"), 0); amount > 0 {
				orch.WealthGoal = amount
				_ = orch.Save("memory.json")
				orch.L1.Add(orchestrator.MemoryEntry{
					ID:        fmt.Sprintf("goal-%d", time.Now().Unix()),
					Content:   fmt.Sprintf("Collective Wealth Goal updated to $%.2f", amount),
					Timestamp: time.Now(),
					Tags:      []string{"swarm", "wealth_goal"},
				})
			}
		case "broadcast_alpha":
			analysis := orch.Ledger.AnalyzeProfitability()
			broker.Publish(orchestrator.Message{
				ID:        fmt.Sprintf("strategy-%d", time.Now().Unix()),
				Source:    "orchestrator-dashboard",
				Type:      orchestrator.Event,
				Topic:     "collective_strategy",
				Payload:   "COLLECTIVE_ALPHA: " + analysis,
				Timestamp: time.Now(),
			})
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("collective_strategy-%d", time.Now().Unix()),
				Content:   "COLLECTIVE_ALPHA: " + analysis,
				Timestamp: time.Now(),
				Tags:      []string{"collective_strategy", "swarm"},
			})
		}
		return nil
	})

	protocol.Register("chain", func(p url.Values) error {
		switch p.Get("action") {
		case "discover":
			_, err := discoverer.Discover()
			return err
		case "create":
			name := strings.TrimSpace(p.Get("name"))
			desc := defVal(p.Get("description"), "Custom workflow chain")
			stepsRaw := strings.TrimSpace(p.Get("steps"))
			if name == "" || stepsRaw == "" {
				return fmt.Errorf("chain create requires name and comma-separated steps")
			}
			steps := []string{}
			for _, s := range strings.Split(stepsRaw, ",") {
				if st := strings.TrimSpace(s); st != "" {
					steps = append(steps, st)
				}
			}
			if err := orchestrator.ValidateChainSteps(steps, protocol); err != nil {
				return err
			}
			chainManager.Register(&orchestrator.Chain{Name: name, Description: desc, Steps: steps})
			return nil
		default:
			return chainManager.Execute(defVal(p.Get("name"), "curation"))
		}
	})

	protocol.Register("confluence", func(p url.Values) error {
		recipe := defVal(p.Get("recipe"), "saas_growth")
		return confluence.NewConfluenceEngine(orch, protocol, broker).ExecuteRecipe(confluence.Recipe(recipe))
	})

	protocol.Register("council", func(p url.Values) error {
		topic := defVal(p.Get("topic"), "Highest-ROI autonomous AI revenue allocation for next 30 days")
		res, err := orchestrator.NewCouncil(orch).Debate(topic)
		if err != nil {
			return err
		}
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("council-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Council Verdict on %q (Score %.2f): %s", topic, res.WeightedScore, res.Consensus),
			Timestamp: time.Now(),
			Tags:      []string{"council", "debate", "strategy"},
		})
		return nil
	})

	protocol.Register("agent", func(p url.Values) error {
		focus := defVal(p.Get("type"), "general")
		iters, _ := strconv.Atoi(defVal(p.Get("iterations"), "3"))
		if iters <= 0 {
			iters = 3
		}
		ag := multiAgent.LaunchAgentAsync(focus, iters)
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("agent-launch-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Launched autonomous agent %s (focus: %s, max_iter: %d)", ag.State.AgentID, focus, iters),
			Timestamp: time.Now(),
			Tags:      []string{"agent_loop", focus},
		})
		return nil
	})

	protocol.Register("autoplan", func(p url.Values) error {
		plans, err := orchestrator.PlanHustles(orch)
		if err != nil {
			return err
		}
		registered := 0
		var names []string
		for _, plan := range plans {
			names = append(names, fmt.Sprintf("%s (%s)", plan.Name, plan.Priority))
			if err := orchestrator.ValidateChainSteps(plan.Steps, protocol); err == nil {
				chainManager.Register(&orchestrator.Chain{
					Name:        plan.Name,
					Description: plan.Description,
					Steps:       plan.Steps,
				})
				registered++
			}
		}
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("autoplan-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Auto-Plan generated %d strategic plans (%d registered as chains): %s", len(plans), registered, strings.Join(names, ", ")),
			Timestamp: time.Now(),
			Tags:      []string{"autoplan", "strategy"},
		})
		return nil
	})

	protocol.Register("synergy_leadgen", func(p url.Values) error {
		topic := defVal(p.Get("topic"), "AI automation services")
		niche := defVal(p.Get("niche"), "B2B SaaS")

		campaign := research.NewOutreachCampaign(orch)
		campaign.WebSearch = research.NewResearchSearch(research.DuckDuckGo, orch, broker)
		if err := campaign.Run(niche, topic); err != nil {
			return err
		}

		affModule := affiliate.NewModule(orch)
		socialContent := social.GenerateContent(orch, topic)
		socialContent = affModule.InjectAffiliateLink(socialContent, topic)
		platform := social.DefaultPlatform()
		socialContent = social.FormatForPlatform(socialContent, platform)
		provider := social.NewProvider(platform)
		if orch.DryRun {
			provider.SetDryRun(true)
		}
		return provider.Post(orch, platform, socialContent)
	})

	protocol.Register("leadgen", func(p url.Values) error {
		niche := defVal(p.Get("niche"), "AI automation")
		lg := research.NewLeadGenerator(orch)
		searcher := research.NewResearchSearch(research.DuckDuckGo, orch, broker)
		leads, err := lg.FindLeadsFromSearch(searcher, niche)
		if err != nil || len(leads) == 0 {
			leads, err = lg.FindLeads(niche)
		}
		if err != nil {
			return err
		}
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("leadgen-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Discovered %d leads for niche %q", len(leads), niche),
			Timestamp: time.Now(),
			Tags:      []string{"leadgen", niche},
		})
		return nil
	})

	protocol.Register("outreach", func(p url.Values) error {
		mod := outreach.NewModule(orch)
		platform := defVal(p.Get("platform"), "Email")
		niche := defVal(p.Get("niche"), "B2B SaaS")
		topic := defVal(p.Get("topic"), "AI workflow automation")
		tmpl, err := mod.GenerateTemplate(platform, niche, topic)
		if err != nil {
			return err
		}
		if email := strings.TrimSpace(p.Get("email")); email != "" {
			return mod.ScheduleCadence(email, platform, tmpl)
		}
		return nil
	})

	protocol.Register("crm", func(p url.Values) error {
		mod := crm.NewCRMModule(orch, broker)
		email := defVal(p.Get("email"), "prospect@example.com")
		status := defVal(p.Get("status"), string(crm.Contacted))
		if p.Get("action") == "followup" {
			reply, err := mod.GenerateFollowUp(crm.ClientLead{
				Name:   defVal(p.Get("name"), "Founder"),
				Email:  email,
				Status: crm.LeadStatus(status),
				Niche:  defVal(p.Get("niche"), "B2B AI"),
			})
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("crm-followup-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("CRM Follow-Up for %s: %s", email, reply),
				Timestamp: time.Now(),
				Tags:      []string{"crm", "followup", email},
			})
			return nil
		}
		return mod.UpdateLeadStatus(email, crm.LeadStatus(status))
	})

	protocol.Register("healer", func(p url.Values) error {
		issue := defVal(p.Get("issue"), "Routine wealth preservation and health check")
		h := orchestrator.NewHealer(orch)
		h.Loop(issue)
		return nil
	})

	protocol.Register("affiliate", func(p url.Values) error {
		action := p.Get("action")
		niche := defVal(p.Get("niche"), "AI productivity tools")
		mod := affiliate.NewAffiliateModule(orch)
		switch action {
		case "run", "":
			return mod.Run(niche)
		case "discover":
			prods, err := mod.DiscoverProducts(niche)
			if err != nil {
				return err
			}
			var names []string
			for _, pr := range prods {
				names = append(names, fmt.Sprintf("%s (%s)", pr.Name, pr.Price))
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("aff-disc-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Affiliate Discovery (%s): %s", niche, strings.Join(names, " | ")),
				Timestamp: time.Now(),
				Tags:      []string{"affiliate", "discovery", niche},
			})
			return nil
		case "batch_insert":
			return mod.InsertAffiliateLinksIntoContent(defVal(p.Get("dir"), "output/content"))
		default:
			return fmt.Errorf("unknown affiliate action: %s", action)
		}
	})

	protocol.Register("media", func(p url.Values) error {
		topic := defVal(p.Get("topic"), "AI Automation")
		if p.Get("action") == "plan" {
			_, err := media.NewMediaModule(orch, broker).PlanProduction(topic, defVal(p.Get("type"), "YouTube Short"))
			return err
		}
		return media.NewShortsFactory(orch).Run(topic)
	})

	protocol.Register("calendar", func(p url.Values) error {
		switch p.Get("action") {
		case "status", "":
			contentCalendar.PrintStatus()
			return nil
		case "process":
			n, err := contentCalendar.ProcessPending()
			fmt.Printf("[Calendar] Processed %d due entries\n", n)
			return err
		case "add":
			topic := p.Get("topic")
			if topic == "" {
				return fmt.Errorf("calendar add requires topic")
			}
			platform := defVal(p.Get("platform"), "wordpress")
			contentType := defVal(p.Get("type"), "blog")
			entry := contentCalendar.AddEntry(topic, contentType, platform, time.Now())
			if body := p.Get("content"); body != "" {
				for i := range contentCalendar.Entries {
					if contentCalendar.Entries[i].ID == entry.ID {
						contentCalendar.Entries[i].Content = body
						contentCalendar.Entries[i].Title = topic
						contentCalendar.Entries[i].Status = "ready"
					}
				}
				_ = contentCalendar.SaveCalendar()
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("cal-add-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Scheduled %s (%s) -> %s", topic, contentType, platform),
				Timestamp: time.Now(),
				Tags:      []string{"calendar", platform},
			})
			return nil
		case "schedule_week":
			weekly := publisher.GenerateWeeklySchedule()
			for _, item := range weekly {
				contentCalendar.AddEntry(item["topic"], item["type"], item["platform"], time.Now())
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("cal-week-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Populated %d weekly content calendar entries", len(weekly)),
				Timestamp: time.Now(),
				Tags:      []string{"calendar", "weekly"},
			})
			return nil
		default:
			return fmt.Errorf("unknown calendar action: %s", p.Get("action"))
		}
	})

	protocol.Register("publisher", func(p url.Values) error {
		switch p.Get("action") {
		case "status", "":
			wpOk := publisher.NewWordPressPublisher().IsConfigured()
			nlOk := publisher.NewNewsletterPublisher().IsConfigured()
			shOk := publisher.NewShopifyPublisher().IsConfigured()
			fmt.Printf("[Publisher] WordPress=%v Newsletter=%v Shopify=%v\n", wpOk, nlOk, shOk)
			return nil
		case "wordpress":
			title := p.Get("title")
			body := p.Get("content")
			if title == "" || body == "" {
				return fmt.Errorf("publisher wordpress requires title and content")
			}
			wp := publisher.NewWordPressPublisher()
			wp.DryRun = orch.DryRun
			resp, err := wp.PublishPost(title, body, defVal(p.Get("status"), "draft"), nil, nil)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("wp-pub-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Published to WordPress: %s (%s)", title, resp.Link),
				Timestamp: time.Now(),
				Tags:      []string{"publisher", "wordpress"},
			})
			return nil
		case "wordpress_markdown":
			path := p.Get("path")
			if path == "" {
				return fmt.Errorf("publisher wordpress_markdown requires path")
			}
			wp := publisher.NewWordPressPublisher()
			wp.DryRun = orch.DryRun
			resp, err := wp.PublishMarkdown(path)
			if err != nil {
				return err
			}
			fmt.Printf("[Publisher] WordPress markdown published: %s\n", resp.Link)
			return nil
		case "newsletter":
			subject := p.Get("subject")
			body := p.Get("content")
			if subject == "" || body == "" {
				return fmt.Errorf("publisher newsletter requires subject and content")
			}
			return publisher.NewNewsletterPublisher().PublishNewsletter(subject, body, []string{"ai", "automation"})
		case "shopify":
			title := defVal(p.Get("title"), "AI Workflow Bundle")
			body := defVal(p.Get("content"), "<p>Instant digital productivity bundle.</p>")
			sh := publisher.NewShopifyPublisher()
			sh.DryRun = orch.DryRun || !sh.IsConfigured()
			id, err := sh.CreateProduct(publisher.ShopifyProduct{
				Title:       title,
				BodyHTML:    body,
				Vendor:      "AI Money Machine",
				ProductType: defVal(p.Get("type"), "Digital"),
				Status:      defVal(p.Get("status"), "draft"),
			})
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("shopify-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Shopify listing prepared: %s (ID: %s)", title, id),
				Timestamp: time.Now(),
				Tags:      []string{"publisher", "shopify"},
			})
			return nil
		default:
			return fmt.Errorf("unknown publisher action: %s", p.Get("action"))
		}
	})

	protocol.Register("products", func(p url.Values) error {
		if action := p.Get("action"); action != "" && action != "generate" {
			return fmt.Errorf("unknown products action: %s", action)
		}
		topic := defVal(p.Get("topic"), "AI productivity system")
		ptype := products.ProductType(defVal(p.Get("type"), string(products.Template)))
		m, err := products.NewFactory(orch).Generate(topic, ptype)
		if err != nil {
			return err
		}
		fmt.Printf("[Products] %s listing=%s dir=%s\n", m.Slug, m.Listing.Status, m.Dir)
		return nil
	})

	protocol.Register("ecommerce", func(p url.Values) error {
		mod := ecommerce.NewEcommerceModule(orch, broker)
		action := defVal(p.Get("action"), "discover")
		niche := defVal(p.Get("niche"), "smart home")
		prodName := defVal(p.Get("product"), niche+" Pro Kit")
		sampleProd := ecommerce.Product{
			Name:        prodName,
			Description: defVal(p.Get("description"), "High-margin trending product in "+niche),
			Niche:       niche,
			Keywords:    []string{niche, "trending", "2026"},
			PriceRange:  defVal(p.Get("price"), "$39-$79"),
			DemandScore: 9.0,
		}
		switch action {
		case "discover":
			prods, err := mod.DiscoverProducts(niche)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("ecom-disc-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Discovered %d trending ecommerce products for %s", len(prods), niche),
				Timestamp: time.Now(),
				Tags:      []string{"ecommerce", niche},
			})
			return nil
		case "listing":
			listing, err := mod.GenerateListing(sampleProd)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("ecom-list-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Shopify Listing for %s: %s", prodName, listing),
				Timestamp: time.Now(),
				Tags:      []string{"ecommerce", "listing"},
			})
			return nil
		case "ads":
			platform := defVal(p.Get("platform"), "TikTok & Instagram")
			ad, err := mod.GenerateAds(sampleProd, platform)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("ecom-ad-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("%s Ad Copy for %s: %s", platform, prodName, ad),
				Timestamp: time.Now(),
				Tags:      []string{"ecommerce", "ads", platform},
			})
			return nil
		case "fulfill":
			plan, err := mod.FulfillOrder(defVal(p.Get("order_id"), "ORD-1001"), prodName)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("ecom-fulfill-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Fulfillment workflow for %s: %s", prodName, plan),
				Timestamp: time.Now(),
				Tags:      []string{"ecommerce", "fulfillment"},
			})
			return nil
		default:
			return fmt.Errorf("unknown ecommerce action: %s", action)
		}
	})

	protocol.Register("saas", func(p url.Values) error {
		mod := saas.NewSaaSModule(orch, broker)
		niche := defVal(p.Get("niche"), "AI workflow productivity")
		proj, err := mod.IdeateProduct(niche)
		if err != nil {
			return err
		}
		if p.Get("action") == "mvp" {
			code, merr := mod.GenerateMVP(*proj)
			if merr != nil {
				return merr
			}
			outDir := filepath.Join("output", "saas")
			_ = os.MkdirAll(outDir, 0755)
			outFile := filepath.Join(outDir, fmt.Sprintf("mvp-%d.html", time.Now().Unix()))
			_ = os.WriteFile(outFile, []byte(code), 0644)
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("saas-mvp-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Generated Micro-SaaS MVP %q -> %s", proj.Name, outFile),
				Timestamp: time.Now(),
				Tags:      []string{"saas", "mvp", niche},
			})
			return nil
		}
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("saas-idea-l1-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Micro-SaaS Concept: %s — %s (Stack: %s, Model: %s)", proj.Name, proj.Description, proj.TechStack, proj.Monetization),
			Timestamp: time.Now(),
			Tags:      []string{"saas", "ideation", niche},
		})
		return nil
	})

	protocol.Register("domains", func(p url.Values) error {
		mod := domains.NewDomainModule(orch, broker)
		dom := defVal(p.Get("domain"), "aiworkflowhub.com")
		if p.Get("action") == "bid" {
			strategy, err := mod.AuctionBid(dom, parseFloat(p.Get("bid"), 120.0))
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("domain-bid-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Auction Strategy for %s: %s", dom, strategy),
				Timestamp: time.Now(),
				Tags:      []string{"domains", "auction", dom},
			})
			return nil
		}
		_, _, err := mod.EvaluateDomain(dom)
		return err
	})

	protocol.Register("kdp", func(p url.Values) error {
		mod := publishing.NewKDPModule(orch, broker)
		niche := defVal(p.Get("niche"), "Daily AI Productivity Planner")
		if p.Get("action") == "keywords" {
			kw, err := mod.GenerateKeywords(defVal(p.Get("title"), niche), niche)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("kdp-kw-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("KDP Backend Keywords (%s): %s", niche, kw),
				Timestamp: time.Now(),
				Tags:      []string{"kdp", "keywords", niche},
			})
			return nil
		}
		_, err := mod.PlanInterior(niche)
		return err
	})

	protocol.Register("pod", func(p url.Values) error {
		mod := pod.NewPODModule(orch, broker)
		niche := defVal(p.Get("niche"), "Developer Humor & AI")
		if p.Get("action") == "listing" {
			listing, err := mod.OptimizeListing(defVal(p.Get("design"), "Ship Fast Prompt Faster"), niche)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("pod-list-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("POD Listing Optimized (%s): %s", niche, listing),
				Timestamp: time.Now(),
				Tags:      []string{"pod", "listing", niche},
			})
			return nil
		}
		_, err := mod.PlanDesigns(niche)
		return err
	})

	protocol.Register("retail", func(p url.Values) error {
		buy := parseFloat(p.Get("buy"), 18.0)
		sell := parseFloat(p.Get("sell"), 54.0)
		cat := defVal(p.Get("category"), "Smart Home Electronics")
		_, _, err := retail.NewRetailModule(orch, broker).CalculateArbitrageROI(buy, sell, cat)
		return err
	})

	protocol.Register("realestate", func(p url.Values) error {
		mod := realestate.NewRealEstateModule(orch, broker)
		loc := defVal(p.Get("location"), "Austin, TX")
		if p.Get("action") == "manage" {
			_, err := mod.ManageListing(loc, defVal(p.Get("message"), "Hi! Is early check-in at 1 PM available tomorrow?"))
			return err
		}
		rent := parseFloat(p.Get("rent"), 2200.0)
		rate := parseFloat(p.Get("rate"), 185.0)
		_, _, err := mod.CalculateSTRProfit(rent, rate, loc)
		return err
	})

	protocol.Register("automation", func(p url.Values) error {
		_, err := automation.NewAutomationModule(orch, broker).BuildWorkflow(
			defVal(p.Get("problem"), "Sync inbound Stripe leads to CRM, enrich company domain, and send personalized onboarding email"),
		)
		return err
	})

	protocol.Register("bi", func(p url.Values) error {
		_, err := bi.NewBIModule(orch, broker).GenerateInsights()
		return err
	})

	protocol.Register("finance", func(p url.Values) error {
		mod := finance.NewFinanceModule(orch, broker)
		csvData := defVal(p.Get("csv"), "2026-11-01,Hetzner Cloud VPS,24.50\n2026-11-05,Domain Renewal,12.99\n2026-11-10,OpenRouter API Credits,15.00")
		txs, err := mod.ClassifyTransactions(csvData)
		if err != nil {
			return err
		}
		if p.Get("action") == "tax" {
			summary, terr := mod.GenerateTaxSummary(txs)
			if terr != nil {
				return terr
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("finance-tax-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Tax Deduction Summary: %s", summary),
				Timestamp: time.Now(),
				Tags:      []string{"finance", "tax"},
			})
		}
		return nil
	})

	protocol.Register("legal", func(p url.Values) error {
		_, err := legal.NewLegalModule(orch, broker).GenerateDocument(
			defVal(p.Get("type"), "Terms of Service & Affiliate Disclosure"),
			defVal(p.Get("business"), "AI Money Machine"),
		)
		return err
	})

	protocol.Register("design", func(p url.Values) error {
		_, err := design.NewDesignModule(orch, broker).IdeateBrand(
			defVal(p.Get("name"), "Apex AI Studio"),
			defVal(p.Get("niche"), "Autonomous B2B Workflows"),
		)
		return err
	})

	protocol.Register("devagency", func(p url.Values) error {
		mod := devagency.NewDevAgencyModule(orch, broker)
		if p.Get("action") == "propose" {
			prop, err := mod.ProposeService(defVal(p.Get("project"), "Build an AI customer-intake agent integrated with HubSpot and Slack"))
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("dev-prop-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Freelance Dev Proposal: %s", prop),
				Timestamp: time.Now(),
				Tags:      []string{"devagency", "proposal"},
			})
			return nil
		}
		_, err := mod.AuditCode(defVal(p.Get("path"), "orchestrator"))
		return err
	})

	protocol.Register("marketing", func(p url.Values) error {
		_, err := marketing.NewMarketingModule(orch, broker).AnalyzeROI()
		return err
	})

	protocol.Register("localization", func(p url.Values) error {
		mod := localization.NewLocalizationModule(orch, broker)
		lang := defVal(p.Get("lang"), "Spanish")
		if p.Get("action") == "seo" {
			kws, err := mod.LocalizeSEO(defVal(p.Get("niche"), "AI productivity tools"), lang)
			if err != nil {
				return err
			}
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("loc-seo-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Localized SEO Keywords (%s): %s", lang, strings.Join(kws, ", ")),
				Timestamp: time.Now(),
				Tags:      []string{"localization", "seo", lang},
			})
			return nil
		}
		_, err := mod.TranslateContent(defVal(p.Get("text"), "Automate your business workflows and multiply output with AI agents."), lang)
		return err
	})

	protocol.Register("support", func(p url.Values) error {
		mod := support.NewSupportModule(orch, broker)
		if p.Get("action") == "faq" {
			return mod.GenerateFAQ()
		}
		_, err := mod.ResolveTicket(
			defVal(p.Get("customer"), "Alex"),
			defVal(p.Get("issue"), "How do I connect my WordPress application password for automated publishing?"),
		)
		return err
	})

	protocol.Register("careers", func(p url.Values) error {
		mod := careers.NewCareerModule(orch, broker)
		if p.Get("action") == "tailor" {
			_, err := mod.TailorResume(
				defVal(p.Get("resume"), "Senior Full-Stack & AI Engineer experienced in Go, Python, distributed systems, and LLM orchestration."),
				defVal(p.Get("job"), "Principal AI Systems Engineer — Autonomous Agent Infrastructure"),
			)
			return err
		}
		_, err := mod.SearchJobs(defVal(p.Get("title"), "AI Automation Engineer"), defVal(p.Get("location"), "Remote"))
		return err
	})

	protocol.Register("qa", func(p url.Values) error {
		mod := qa.NewQAModule(orch, broker)
		if p.Get("action") == "stability" {
			return mod.VerifyStability()
		}
		_, err := mod.RunTests(defVal(p.Get("module"), "./hustle/affiliate/..."))
		return err
	})

	protocol.Register("sync", func(p url.Values) error {
		return runSyncProtocol()
	})

	// ── Mesh Seed ──
	if *seedURL != "" {
		fmt.Printf("[Swarm] Joining mesh via seed: %s\n", *seedURL)
		broker.RegisterPeer("seed-node", *seedURL)
		_ = protocol.HandleURI("hustle://swarm?action=sync")
	}

	// ── API Server ──
	if *apiPort != "" {
		api := orchestrator.NewAPI(orch, protocol, broker, chainManager, discoverer)
		api.MultiAgent = multiAgent
		api.PriceSource = priceSource
		api.TradingSnapshot = func() interface{} {
			return traderModule.SnapshotAll()
		}
		api.CalendarSnapshot = func() interface{} {
			return contentCalendar.Entries
		}
		api.SchedulerSnapshot = func() interface{} {
			return scheduler.Snapshot()
		}
		go func() {
			if err := api.Start(*apiPort); err != nil {
				fmt.Printf("[API] Server exited: %v\n", err)
			}
		}()

		if *seedURL != "" {
			selfURL := os.Getenv("AIMM_PEER_URL")
			if selfURL == "" {
				selfURL = "http://localhost:" + *apiPort
			}
			go announceToSeed(*seedURL, selfURL, *apiPort)
		}
	}

	// ── Agent Mode ──
	if *agentMode {
		agent := orchestrator.NewAgentLoop(orch, protocol, broker, *agentType)
		agent.State.MaxIter = *agentIter
		fmt.Printf("[Agent] Launching autonomous agent: %s (%d iterations)\n", *agentType, *agentIter)
		if err := agent.Run(); err != nil {
			fmt.Printf("[Agent] Agent failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ── Auto-Plan Mode ──
	if *autoPlan {
		fmt.Println("[AutoPlan] Asking LLM to generate strategic hustle plan...")
		plans, err := orchestrator.PlanHustles(orch)
		if err != nil {
			fmt.Printf("[AutoPlan] Planning failed: %v\n", err)
			os.Exit(1)
		}
		for _, plan := range plans {
			if plan.Priority == "high" || plan.Priority == "medium" {
				if err := orchestrator.ValidateChainSteps(plan.Steps, protocol); err == nil {
					chainManager.Register(&orchestrator.Chain{
						Name:        plan.Name,
						Description: plan.Description,
						Steps:       plan.Steps,
					})
				}
				multiAgent.AddAgent(plan.Category, 10)
			}
		}
		multiAgent.RunAll()
		return
	}

	if *interactive {
		runInteractiveMenu(orch, protocol, broker, traderModule, contentModule, version)
		return
	}

	if *dashboard {
		orchestrator.StartLiveDashboard(orch)
		return
	}

	// ── Daemon Mode ──
	if *daemon {
		scheduler.KnownChain = func(name string) bool {
			return chainManager.Has(name)
		}
		_ = scheduler.LoadState("tasks.json", protocol)

		scheduler.Register("Research", 1*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://research?query=AI+Agent+Trends")
		})
		scheduler.Register("CurationChain", 2*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://chain")
		})
		symbols := []string{"BTC", "ETH", "SOL", "DOGE", "XRP", "ADA", "LINK"}
		for i, sym := range symbols {
			interval := time.Duration(i+1) * time.Minute
			s := sym
			scheduler.Register("Trading"+s, interval, func(o *orchestrator.Orchestrator) error {
				return protocol.HandleURI("hustle://trading?symbol=" + s)
			})
		}
		scheduler.Register("ContentGeneration", 3*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://content?topic=AI+automation+trends&type=blog")
		})
		scheduler.Register("CalendarProcess", 90*time.Minute, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://calendar?action=process")
		})
		scheduler.Register("SwarmSync", 4*time.Hour, func(o *orchestrator.Orchestrator) error {
			_ = protocol.HandleURI("hustle://swarm?action=aggregate")
			return protocol.HandleURI("hustle://swarm?action=sync")
		})
		if syncAvailable() {
			scheduler.Register("Sync", 6*time.Hour, func(o *orchestrator.Orchestrator) error {
				return runSyncProtocol()
			})
		} else {
			fmt.Printf("[Scheduler] Sync task skipped: %s not present on this host\n", syncScriptPath())
		}
		scheduler.Register("ProfitAnalysis", 12*time.Hour, func(o *orchestrator.Orchestrator) error {
			h := orchestrator.NewHealer(o)
			h.WealthPreservation()
			suggestion := o.Ledger.AnalyzeProfitability()
			fmt.Printf("[Scheduler] Financial Analysis: %s\n", suggestion)
			scheduler.ReevaluateStrategy(suggestion)
			o.L2.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("profit-analysis-%d", time.Now().Unix()),
				Content:   suggestion,
				Timestamp: time.Now(),
				Tags:      []string{"ledger", "analysis", "recommendation"},
			})
			return nil
		})
		scheduler.Register("WorkflowDiscovery", 24*time.Hour, func(o *orchestrator.Orchestrator) error {
			discovered, err := discoverer.Discover()
			if err != nil {
				return err
			}
			scheduler.Register("DiscoveredChain:"+discovered.Name, time.Duration(discovered.IntervalMinutes)*time.Minute, func(o *orchestrator.Orchestrator) error {
				return protocol.HandleURI(fmt.Sprintf("hustle://chain?name=%s", discovered.Name))
			})
			return nil
		})
		scheduler.Register("Heartbeat", 5*time.Minute, func(o *orchestrator.Orchestrator) error {
			if o.Cache != nil {
				o.Cache.ClearExpired()
			}
			return orchestrator.WriteStatusReport(version, "Active", "Scheduler Heartbeat", &o.Ledger)
		})

		// Run scheduler in background so signalChan handles clean SIGINT/SIGTERM state persistence!
		go scheduler.Start()
		fmt.Println("Orchestrator running in Daemon mode.")
		sig := <-sigChan
		fmt.Printf("\n[Main] Received signal: %v\n", sig)
		scheduler.Stop()
		orch.Shutdown()
		return
	}

	if *syncMode {
		if err := runSyncProtocol(); err != nil {
			fmt.Printf("Sync error: %v\n", err)
		}
	}

	if *hustleURI != "" {
		if err := protocol.HandleURI(*hustleURI); err != nil {
			fmt.Printf("Protocol error: %v\n", err)
		}
	} else if *hustleTask != "" {
		fmt.Printf("Launching hustle module: %s with params: %s\n", *hustleTask, *params)
		if err := protocol.HandleURI(fmt.Sprintf("hustle://%s", *hustleTask)); err != nil {
			fmt.Printf("Protocol error: %v\n", err)
		}
	}

	if *apiPort != "" {
		fmt.Println("Orchestrator API running. Press Ctrl+C to terminate.")
		sig := <-sigChan
		fmt.Printf("\n[Main] Received signal: %v\n", sig)
		orch.Shutdown()
		return
	}

	_ = orchestrator.WriteStatusReport(version, "Running", "Orchestrator command processed", &orch.Ledger)
	fmt.Printf("Orchestrator initialized with L1: %d, L2: %d, L3: %d entries. Profit: $%.2f\n",
		orch.L1.Len(), orch.L2.Len(), orch.L3.Len(), orch.Ledger.Profit())
}

func defVal(val, fallback string) string {
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

func parseFloat(val string, fallback float64) float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err == nil {
		return v
	}
	return fallback
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func syncScriptPath() string {
	if v := os.Getenv("SYNC_SCRIPT"); v != "" {
		return v
	}
	return "./sync.sh"
}

func syncAvailable() bool {
	_, err := os.Stat(syncScriptPath())
	return err == nil
}

func announceToSeed(seedURL, selfURL, apiPort string) {
	body, _ := json.Marshal(map[string]string{
		"id":  "peer-" + apiPort,
		"url": selfURL,
	})
	client := &http.Client{Timeout: 10 * time.Second}
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err := client.Post(strings.TrimRight(seedURL, "/")+"/register", "application/json", bytes.NewReader(body))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Printf("[Swarm] Announced to seed %s as %s\n", seedURL, selfURL)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	fmt.Printf("[Swarm] Could not announce to seed %s\n", seedURL)
}

func runSyncProtocol() error {
	if !syncAvailable() {
		return fmt.Errorf("repository sync unavailable: %s not found on this host (prod updates via scripts/deploy-hetzner.sh; see DEPLOY.md)", syncScriptPath())
	}
	cmd := exec.Command(syncScriptPath())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runInteractiveMenu(orch *orchestrator.Orchestrator, protocol *orchestrator.HustleProtocol, broker *orchestrator.A2ABroker, traderModule *trading.TradingModule, contentModule *content.ContentModule, version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- AI HUSTLE MACHINE INTERACTIVE MENU ---")
		fmt.Println(" 1. Research | 2. Social | 3. Curation | 4. Trading | 5. Chain | 6. Swarm Sync")
		fmt.Println(" 7. Blog | 8. Affiliate | 9. Digital Product | 10. Council Debate | 11. Auto-Plan")
		fmt.Println(" 12. Terminal Dashboard | 13. Task History | uri. Custom hustle:// URI | q. Quit")
		fmt.Print("Select: ")
		in, _ := reader.ReadString('\n')
		in = strings.TrimSpace(in)
		switch in {
		case "1":
			_ = protocol.HandleURI("hustle://research")
		case "2":
			_ = protocol.HandleURI("hustle://social")
		case "3":
			_ = protocol.HandleURI("hustle://curation")
		case "4":
			_ = protocol.HandleURI("hustle://trading")
		case "5":
			_ = protocol.HandleURI("hustle://chain")
		case "6":
			_ = protocol.HandleURI("hustle://swarm?action=sync")
		case "7":
			_ = protocol.HandleURI("hustle://content?type=blog")
		case "8":
			_ = protocol.HandleURI("hustle://affiliate?action=run")
		case "9":
			_ = protocol.HandleURI("hustle://products?type=template")
		case "10":
			_ = protocol.HandleURI("hustle://council")
		case "11":
			_ = protocol.HandleURI("hustle://autoplan")
		case "12":
			orchestrator.ShowDashboard(orch)
		case "13":
			orchestrator.ShowTaskHistory(orch)
		case "uri":
			fmt.Print("Enter hustle:// URI: ")
			u, _ := reader.ReadString('\n')
			_ = protocol.HandleURI(strings.TrimSpace(u))
		case "q":
			return
		}
	}
}
