package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/hustle/content"
	"github.com/robertpelloni/hustle/hustle/curation"
	"github.com/robertpelloni/hustle/hustle/research"
	"github.com/robertpelloni/hustle/hustle/social"
	"github.com/robertpelloni/hustle/hustle/trading"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	hustleTask := flag.String("hustle", "", "Run a specific hustle module (research, social, curation, content, chain, trading, swarm)")
	hustleURI := flag.String("uri", "", "Run a hustle module via hustle:// URI")
	syncMode := flag.Bool("sync", false, "Run repository synchronization protocol")
	params := flag.String("params", "{}", "JSON parameters for the hustle module")
	dashboard := flag.Bool("dashboard", false, "Start the live terminal dashboard")
	daemon := flag.Bool("daemon", false, "Run the Orchestrator as a background task scheduler")
	interactive := flag.Bool("interactive", false, "Launch the interactive command menu")
	apiPort := flag.String("api", "", "Start the HTTP API on specified port (e.g. 8080)")
	realPrices := flag.Bool("real-prices", false, "Use real-world market prices via CoinGecko")
	seedURL := flag.String("seed", "", "URL of a seed node for mesh federation")
	agentMode := flag.Bool("agent", false, "Run as autonomous agent loop (LLM-driven decision making)")
	agentType := flag.String("agent-type", "general", "Agent hustle focus: general, research, content, trading, social")
	agentIter := flag.Int("agent-iterations", 20, "Max iterations for agent loop")
	autoPlan := flag.Bool("autoplan", false, "LLM generates and executes a strategic hustle plan")
	dryRun := flag.Bool("dry-run", false, "Execute in dry-run mode (no external mutations)")
	refreshRate := flag.Int("refresh", 1000, "Dashboard refresh rate in milliseconds")
	flag.Parse()

	// Source version from VERSION.md
	versionData, err := os.ReadFile("VERSION.md")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionData))
	}

	fmt.Printf("=== AI Hustle Machine Orchestrator v%s ===\n", version)

	// ── Signal Handling ──
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// ── Initialize Orchestrator with REAL LLM ──
	orch := orchestrator.NewOrchestrator()
	orch.Load("memory.json")
	orch.DryRun = *dryRun

	// Initialize SQLite Store
	db, err := orchestrator.NewSQLiteStore("hustle.db")
	if err != nil {
		fmt.Printf("[SQLite] ⚠️ Failed to initialize database: %v\n", err)
	} else {
		orch.DB = db
	}

	// Wire up real LLM via LM Studio / OpenAI-compatible provider
	llmProvider := orchestrator.NewOpenAICompatProvider()
	// Auto-detect and test the LLM connection
	if model, err := llmProvider.DetectModel(); err == nil {
		fmt.Printf("[LLM] ✅ Connected to local LLM: %s\n", model)
		orch.LLM = llmProvider
		// Also set up real embeddings
		orch.Embedder = orchestrator.NewOpenAICompatEmbedder()
	} else {
		fmt.Printf("[LLM] ⚠️  Local LLM not available (%v). Falling back to MockLLM.\n", err)
		fmt.Printf("[LLM]    Start LM Studio or set LLM_BASE_URL to enable real AI.\n")
	}

	protocol := orchestrator.NewHustleProtocol()
	chainManager := orchestrator.NewChainManager(orch, protocol)
	chainManager.LoadState("chains.json")
	discoverer := orchestrator.NewChainDiscoverer(orch, chainManager)
	broker := orchestrator.NewA2ABroker(orch)
	swarm := orchestrator.NewMemorySwarm(orch, broker)
	multiAgent := orchestrator.NewMultiAgentOrchestrator(orch, protocol, broker)

	// ── Initialize Trading Module ──
	var fetcher trading.PriceFetcher = &trading.MockPriceFetcher{}
	if *realPrices {
		fmt.Println("[Trading] Real-world price fetching enabled via CoinGecko.")
		cgKey := os.Getenv("COINGECKO_API_KEY")
		if cgKey != "" {
			fmt.Println("[Trading] CoinGecko Pro API Key detected.")
		}
		fetcher = &trading.CoinGeckoFetcher{APIKey: cgKey}
	}
	traderModule := &trading.TradingModule{
		Orchestrator: orch,
		Broker:       broker,
		Fetcher:      fetcher,
	}

	// ── Initialize Content Module ──
	contentModule := content.NewContentModule(orch, "output/content")

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
				protocol.HandleURI(msg.Payload)
			}
		}
	}()

	// ── Register Protocol Handlers ──
	protocol.Register("research", func(p url.Values) error {
		query := p.Get("query")
		if query == "" {
			query = "AI Trends"
		}
		searcher := research.NewResearchSearch(research.Tavily, orch, broker)
		_, err := searcher.Query(query)
		return err
	})

	protocol.Register("curation", func(p url.Values) error {
		topic := p.Get("topic")
		if topic == "" {
			topic = "AI"
		}
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

	protocol.Register("social", func(p url.Values) error {
		platform := p.Get("platform")
		if platform == "" {
			platform = "Twitter"
		}
		topic := p.Get("topic")
		if topic == "" {
			topic = "AI"
		}
		contentStr := p.Get("content")

		var provider social.Provider
		if strings.ToLower(platform) == "linkedin" {
			provider = social.NewLinkedInProvider(
				os.Getenv("LINKEDIN_ACCESS_TOKEN"),
				os.Getenv("LINKEDIN_AUTHOR_URN"),
			)
		} else {
			provider = social.NewTwitterProvider(
				os.Getenv("TWITTER_API_KEY"),
				os.Getenv("TWITTER_API_SECRET"),
				os.Getenv("TWITTER_ACCESS_TOKEN"),
				os.Getenv("TWITTER_ACCESS_SECRET"),
			)
		}

		if *dryRun {
			provider.SetDryRun(true)
		}

		if contentStr != "" {
			fmt.Printf("[Social] Posting explicit content to %s: %s\n", platform, contentStr)
			return provider.Post(orch, platform, contentStr)
		}

		social.SchedulePost(orch, provider, platform, topic)
		return nil
	})

	protocol.Register("trading", func(p url.Values) error {
		symbol := p.Get("symbol")
		if symbol == "" {
			symbol = "BTC"
		}
		traderModule.Symbol = symbol
		return traderModule.ExecuteStrategy()
	})

	protocol.Register("content", func(p url.Values) error {
		topic := p.Get("topic")
		if topic == "" {
			topic = "AI automation trends"
		}
		contentType := p.Get("type")
		if contentType == "" {
			contentType = "blog"
		}
		niche := p.Get("niche")
		if niche == "" {
			niche = "AI & automation"
		}
		keywords := p.Get("keywords")
		kwList := []string{"AI", "automation"}
		if keywords != "" {
			kwList = strings.Split(keywords, ",")
		}

		req := content.ContentRequest{
			Topic:       topic,
			Type:        content.ContentType(contentType),
			Keywords:    kwList,
			TargetWords: 800,
			Niche:       niche,
		}
		_, err := contentModule.Generate(req)
		return err
	})

	protocol.Register("swarm", func(p url.Values) error {
		action := p.Get("action")
		peerID := p.Get("peer_id")
		if peerID == "" {
			peerID = "unknown-peer"
		}
		switch action {
		case "sync":
			swarm.Sync()
		case "sync_request":
			swarm.HandleSyncRequest(peerID)
		case "provide_index":
			data := p.Get("data")
			swarm.ReconcileIndex(peerID, data)
		case "request_entry":
			id := p.Get("id")
			swarm.ProvideEntry(peerID, id)
		case "provide_entry":
			id := p.Get("id")
			contentStr := p.Get("content")
			orch.L2.Add(orchestrator.MemoryEntry{
				ID:        id,
				Content:   contentStr,
				Timestamp: time.Now(),
				Tags:      []string{"swarm", "received", "from:" + peerID},
			})
			fmt.Printf("[Swarm] Successfully ingested entry %s from %s\n", id, peerID)
		case "aggregate":
			swarm.AggregateStatus()
		case "status_request":
			swarm.ProvideStatus(peerID)
		case "provide_status":
			data := p.Get("data")
			swarm.HandleStatusResponse(peerID, data)
		case "set_goal":
			amountStr := p.Get("amount")
			var amount float64
			if _, err := fmt.Sscanf(amountStr, "%f", &amount); err == nil && amount > 0 {
				orch.WealthGoal = amount
				fmt.Printf("[Swarm] Unified Collective Wealth Goal set to $%.2f\n", amount)
			}
		}
		return nil
	})

	protocol.Register("chain", func(p url.Values) error {
		action := p.Get("action")
		if action == "discover" {
			_, err := discoverer.Discover()
			return err
		}
		name := p.Get("name")
		if name == "" {
			name = "curation"
		}
		return chainManager.Execute(name)
	})

	protocol.Register("healer", func(p url.Values) error {
		issue := p.Get("issue")
		if issue == "" {
			issue = "Unknown system instability"
		}
		h := orchestrator.NewHealer(orch)
		h.Loop(issue)
		return nil
	})

	protocol.Register("sync", func(p url.Values) error {
		return runSyncProtocol()
	})

	// ── Mesh Seed ──
	if *seedURL != "" {
		fmt.Printf("[Swarm] Joining mesh via seed: %s\n", *seedURL)
		broker.RegisterPeer("seed-node", *seedURL)
		protocol.HandleURI("hustle://swarm?action=sync")
	}

	// ── API Server ──
	if *apiPort != "" {
		api := orchestrator.NewAPI(orch, protocol, broker, chainManager, discoverer)
		api.MultiAgent = multiAgent
		go api.Start(*apiPort)
	}

	// ── Agent Mode (NEW): LLM-driven autonomous loop ──
	if *agentMode {
		agent := orchestrator.NewAgentLoop(orch, protocol, broker, *agentType)
		agent.State.MaxIter = *agentIter
		fmt.Printf("[Agent] 🤖 Launching autonomous agent: %s (%d iterations)\n", *agentType, *agentIter)
		if err := agent.Run(); err != nil {
			fmt.Printf("[Agent] ❌ Agent failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ── Auto-Plan Mode (NEW): LLM generates strategic plan and executes it ──
	if *autoPlan {
		fmt.Println("[AutoPlan] 🧠 Asking LLM to generate strategic hustle plan...")
		plans, err := orchestrator.PlanHustles(orch)
		if err != nil {
			fmt.Printf("[AutoPlan] ❌ Planning failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("[AutoPlan] 📋 Generated %d hustle plans:\n", len(plans))
		for i, plan := range plans {
			fmt.Printf("  %d. [%s] %s — %s (every %d min, priority: %s)\n",
				i+1, plan.Category, plan.Name, plan.Description, plan.IntervalMin, plan.Priority)
		}

		// Execute the plans as agents
		for _, plan := range plans {
			if plan.Priority == "high" || plan.Priority == "medium" {
				// Register chain from plan steps
				chain := &orchestrator.Chain{
					Name:        plan.Name,
					Description: plan.Description,
					Steps:       plan.Steps,
				}
				chainManager.Register(chain)
				multiAgent.AddAgent(plan.Category, 10)
			}
		}

		multiAgent.RunAll()
		return
	}

	// ── Interactive Mode ──
	if *interactive {
		runInteractiveMenu(orch, protocol, broker, traderModule, contentModule, version)
		return
	}

	// ── Dashboard Mode ──
	if *dashboard {
		orchestrator.StartLiveDashboard(orch, time.Duration(*refreshRate)*time.Millisecond)
		return
	}

	// ── Daemon Mode ──
	if *daemon {
		scheduler := orchestrator.NewScheduler(orch)
		scheduler.LoadState("tasks.json", protocol)

		scheduler.Register("Research", 1*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://research?query=AI+Agent+Trends")
		})
		scheduler.Register("CurationChain", 2*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://chain")
		})
		scheduler.Register("Trading", 30*time.Minute, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://trading?symbol=BTC")
		})
		scheduler.Register("ContentGeneration", 3*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://content?topic=AI+automation+trends&type=blog")
		})
		scheduler.Register("SwarmSync", 4*time.Hour, func(o *orchestrator.Orchestrator) error {
			protocol.HandleURI("hustle://swarm?action=aggregate")
			return protocol.HandleURI("hustle://swarm?action=sync")
		})
		scheduler.Register("Sync", 6*time.Hour, func(o *orchestrator.Orchestrator) error {
			return runSyncProtocol()
		})
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
			return orchestrator.WriteStatusReport(version, "Active", "Scheduler Heartbeat", o.Ledger)
		})

		fmt.Println("Orchestrator running in Daemon mode.")
		scheduler.Start()
		return
	}

	// ── Sync Mode ──
	if *syncMode {
		fmt.Println("Triggering repository sync protocol...")
		if err := runSyncProtocol(); err != nil {
			fmt.Printf("Sync error: %v\n", err)
		}
	}

	// ── Single Task / URI Mode ──
	if *hustleURI != "" {
		fmt.Printf("Executing hustle URI: %s\n", *hustleURI)
		if err := protocol.HandleURI(*hustleURI); err != nil {
			fmt.Printf("Protocol error: %v\n", err)
		}
	} else if *hustleTask != "" {
		fmt.Printf("Launching hustle module: %s with params: %s\n", *hustleTask, *params)
		uri := fmt.Sprintf("hustle://%s", *hustleTask)
		if err := protocol.HandleURI(uri); err != nil {
			fmt.Printf("Protocol error: %v\n", err)
		}
	}

	// Wait for shutdown signal or exit if not in long-running mode
	if *apiPort != "" || *daemon || *dashboard || *agentMode {
		fmt.Println("Orchestrator running. Press Ctrl+C to terminate.")
		select {
		case sig := <-sigChan:
			fmt.Printf("\n[Main] Received signal: %v\n", sig)
			orch.Shutdown()
			os.Exit(0)
		}
	}

	// Real-time status reporting with financial metrics
	orchestrator.WriteStatusReport(version, "Running", "Orchestrator command processed", orch.Ledger)
	fmt.Printf("Orchestrator initialized with L1: %d, L2: %d, L3: %d entries. Profit: $%.2f\n",
		len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries), orch.Ledger.Profit())
}

func runSyncProtocol() error {
	cmd := exec.Command("./sync.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runRSSMenu(orch *orchestrator.Orchestrator, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- RSS FEED MANAGEMENT ---")
		fmt.Println(" 1. List Current Feeds")
		fmt.Println(" 2. Add New RSS Feed")
		fmt.Println(" 3. Remove RSS Feed")
		fmt.Println(" r. Return to Main Menu")
		fmt.Print("Select option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			fmt.Println("Active RSS Feeds:")
			if len(orch.RSSFeeds) == 0 {
				fmt.Println("  (Using default: https://news.ycombinator.com/rss)")
			} else {
				for i, f := range orch.RSSFeeds {
					fmt.Printf("  %d. %s\n", i+1, f)
				}
			}
		case "2":
			fmt.Print("Enter RSS Feed URL: ")
			urlStr, _ := reader.ReadString('\n')
			urlStr = strings.TrimSpace(urlStr)
			if urlStr != "" {
				orch.RSSFeeds = append(orch.RSSFeeds, urlStr)
				fmt.Println("Feed added.")
			}
		case "3":
			fmt.Print("Enter feed number to remove: ")
			idxStr, _ := reader.ReadString('\n')
			idxStr = strings.TrimSpace(idxStr)
			var idx int
			fmt.Sscanf(idxStr, "%d", &idx)
			if idx > 0 && idx <= len(orch.RSSFeeds) {
				orch.RSSFeeds = append(orch.RSSFeeds[:idx-1], orch.RSSFeeds[idx:]...)
				fmt.Println("Feed removed.")
			} else {
				fmt.Println("Invalid index.")
			}
		case "r":
			return
		}
	}
}

func runSpaceCommsSubmenu(orch *orchestrator.Orchestrator, protocol *orchestrator.HustleProtocol, broker *orchestrator.A2ABroker, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- SPACE COMMUNICATION CONTROL ---")
		fmt.Println(" 1. List Active Mesh Peers")
		fmt.Println(" 2. Broadcast Global Directive")
		fmt.Println(" 3. Trigger Mesh Synchronization")
		fmt.Println(" 4. Sync Collective Strategy")
		fmt.Println(" 5. View Collective Strategy Hub")
		fmt.Println(" r. Return to Main Menu")
		fmt.Print("Select option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			fmt.Println("Active Peers:")
			for id, url := range broker.Peers {
				fmt.Printf("  - %s: %s\n", id, url)
			}
		case "2":
			fmt.Print("Enter Global Directive: ")
			directive, _ := reader.ReadString('\n')
			directive = strings.TrimSpace(directive)
			broker.Publish(orchestrator.Message{
				ID:        fmt.Sprintf("directive-%d", time.Now().Unix()),
				Source:    "orchestrator-ui",
				Type:      orchestrator.Command,
				Topic:     "global_directives",
				Payload:   directive,
				Timestamp: time.Now(),
			})
			fmt.Println("Directive broadcasted to space.")
		case "3":
			protocol.HandleURI("hustle://swarm?action=sync")
		case "4":
			fmt.Println("Broadcasting local best hustle to mesh...")
			analysis := orch.Ledger.AnalyzeProfitability()
			broker.Publish(orchestrator.Message{
				ID:        fmt.Sprintf("strategy-%d", time.Now().Unix()),
				Source:    "orchestrator-ui",
				Type:      orchestrator.Event,
				Topic:     "collective_strategy",
				Payload:   "COLLECTIVE_ALPHA: " + analysis,
				Timestamp: time.Now(),
			})
			fmt.Println("Strategy synced.")
		case "5":
			fmt.Println("\n📊 COLLECTIVE STRATEGY HUB:")
			strategies := orch.L1.Search("collective_strategy")
			if len(strategies) == 0 {
				fmt.Println("  (No shared strategies discovered in mesh yet)")
			} else {
				for _, s := range strategies {
					fmt.Printf("  [%s] %s\n", s.Timestamp.Format("15:04"), s.Content)
				}
			}
		case "r":
			return
		}
	}
}

func runInteractiveMenu(orch *orchestrator.Orchestrator, protocol *orchestrator.HustleProtocol, broker *orchestrator.A2ABroker, traderModule *trading.TradingModule, contentModule *content.ContentModule, version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- INTERACTIVE COMMAND MENU ---")
		fmt.Println(" 1. Launch Research Hustle")
		fmt.Println(" 2. Launch Social Hustle")
		fmt.Println(" 3. Launch Curation Hustle")
		fmt.Println(" 4. Launch Trading Hustle")
		fmt.Println(" 5. Launch FULL CHAIN (Curate -> Post)")
		fmt.Println(" 6. Trigger Swarm Sync")
		fmt.Println(" 7. Broadcast Custom Mesh Event")
		fmt.Println(" 8. Trading: SELL ALL & Clear History")
		fmt.Println(" 9. View Dashboard")
		fmt.Println("10. Run Repository Sync")
		fmt.Println("11. 🆕 Generate Blog Content")
		fmt.Println("12. 🆕 Generate Newsletter")
		fmt.Println("13. 🆕 Generate SEO Article")
		fmt.Println("14. 🆕 Generate Social Thread")
		fmt.Println("15. 🆕 Brainstorm Content Topics")
		fmt.Println("16. 🆕 Start Agent Loop (10 iterations)")
		fmt.Println("17. 🆕 Auto-Plan Hustles (LLM strategy)")
		fmt.Println("18. 🆕 View Content Library")
		fmt.Println("19. 🆕 Space Communication (Mesh Control)")
		fmt.Println("20. 🆕 Manual System Diagnosis")
		fmt.Println("21. 🆕 RSS Feed Management")
		fmt.Println("22. 🆕 View Task History (SQLite)")
		fmt.Println("23. 🆕 Configure Collective Wealth Goal")
		fmt.Println(" q. Quit")
		fmt.Print("Select an option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			protocol.HandleURI("hustle://research")
		case "2":
			protocol.HandleURI("hustle://social")
		case "3":
			protocol.HandleURI("hustle://curation")
		case "4":
			protocol.HandleURI("hustle://trading")
		case "5":
			protocol.HandleURI("hustle://chain")
		case "6":
			protocol.HandleURI("hustle://swarm?action=sync")
		case "7":
			fmt.Print("Enter event payload: ")
			payload, _ := reader.ReadString('\n')
			payload = strings.TrimSpace(payload)
			broker.Publish(orchestrator.Message{
				ID:        fmt.Sprintf("manual-%d", time.Now().Unix()),
				Source:    "interactive-user",
				Type:      orchestrator.Event,
				Topic:     "manual_event",
				Payload:   payload,
				Timestamp: time.Now(),
			})
		case "8":
			fmt.Println("Executing SELL ALL and clearing technical history...")
			traderModule.ExecuteStrategy()
			broker.Publish(orchestrator.Message{
				ID:        fmt.Sprintf("sell-all-%d", time.Now().Unix()),
				Source:    "interactive-user",
				Type:      orchestrator.Command,
				Topic:     "trade_execution",
				Payload:   "ACTION: SELL_ALL, Reason: Manual intervention",
				Timestamp: time.Now(),
			})
			fmt.Println("History cleared. SELL_ALL broadcasted to mesh.")
		case "9":
			orchestrator.ShowDashboard(orch)
			fmt.Println("\nPress Enter to return to menu...")
			reader.ReadString('\n')
		case "10":
			if err := runSyncProtocol(); err != nil {
				fmt.Printf("Sync error: %v\n", err)
			}
		case "11":
			fmt.Print("Topic (default: AI automation trends): ")
			topic, _ := reader.ReadString('\n')
			topic = strings.TrimSpace(topic)
			if topic == "" {
				topic = "AI automation trends"
			}
			protocol.HandleURI(fmt.Sprintf("hustle://content?topic=%s&type=blog", url.QueryEscape(topic)))
		case "12":
			fmt.Print("Topic (default: This week in AI): ")
			topic, _ := reader.ReadString('\n')
			topic = strings.TrimSpace(topic)
			if topic == "" {
				topic = "This week in AI"
			}
			protocol.HandleURI(fmt.Sprintf("hustle://content?topic=%s&type=newsletter", url.QueryEscape(topic)))
		case "13":
			fmt.Print("Target keyword (default: best AI tools 2026): ")
			topic, _ := reader.ReadString('\n')
			topic = strings.TrimSpace(topic)
			if topic == "" {
				topic = "best AI tools 2026"
			}
			protocol.HandleURI(fmt.Sprintf("hustle://content?topic=%s&type=seo", url.QueryEscape(topic)))
		case "14":
			fmt.Print("Thread topic (default: How AI agents are replacing SaaS): ")
			topic, _ := reader.ReadString('\n')
			topic = strings.TrimSpace(topic)
			if topic == "" {
				topic = "How AI agents are replacing SaaS"
			}
			protocol.HandleURI(fmt.Sprintf("hustle://content?topic=%s&type=thread", url.QueryEscape(topic)))
		case "15":
			fmt.Print("Niche (default: AI & automation): ")
			niche, _ := reader.ReadString('\n')
			niche = strings.TrimSpace(niche)
			if niche == "" {
				niche = "AI & automation"
			}
			ideas, err := contentModule.GenerateTopicIdeas(niche, 10)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("\n📊 Content Topic Ideas:")
				for i, idea := range ideas {
					fmt.Printf("  %d. %s\n", i+1, idea)
				}
			}
		case "16":
			fmt.Println("🤖 Starting agent loop (10 iterations)...")
			agent := orchestrator.NewAgentLoop(orch, protocol, broker, "general")
			agent.State.MaxIter = 10
			agent.Run()
		case "17":
			fmt.Println("🧠 Asking LLM to plan hustles...")
			plans, err := orchestrator.PlanHustles(orch)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				for i, p := range plans {
					fmt.Printf("  %d. [%s/%s] %s — %s\n", i+1, p.Category, p.Priority, p.Name, p.Description)
				}
			}
		case "18":
			path, err := contentModule.GenerateGallery()
			if err != nil {
				fmt.Printf("Error generating gallery: %v\n", err)
			} else {
				fmt.Printf("✅ Content Gallery generated: %s\n", path)
				files, _ := os.ReadDir("output/content")
				fmt.Println("\n📂 CONTENT LIBRARY:")
				for _, f := range files {
					if !f.IsDir() {
						fmt.Printf("  - %s\n", f.Name())
					}
				}
			}
		case "19":
			runSpaceCommsSubmenu(orch, protocol, broker, reader)
		case "20":
			fmt.Print("Describe the system issue to diagnose: ")
			issue, _ := reader.ReadString('\n')
			issue = strings.TrimSpace(issue)
			if issue == "" {
				issue = "Manual override diagnostic"
			}
			h := orchestrator.NewHealer(orch)
			fmt.Println("Starting LLM-verified diagnostic loop...")
			h.Loop(issue)
		case "21":
			runRSSMenu(orch, reader)
		case "22":
			orchestrator.ShowTaskHistory(orch)
			fmt.Println("\nPress Enter to return to menu...")
			reader.ReadString('\n')
		case "23":
			fmt.Printf("Current Wealth Goal: $%.2f\n", orch.WealthGoal)
			fmt.Print("Enter new goal amount: ")
			goalStr, _ := reader.ReadString('\n')
			goalStr = strings.TrimSpace(goalStr)
			var newGoal float64
			if _, err := fmt.Sscanf(goalStr, "%f", &newGoal); err == nil && newGoal > 0 {
				orch.WealthGoal = newGoal
				fmt.Printf("Collective Wealth Goal updated to $%.2f. Broadcasting to mesh...\n", newGoal)
				protocol.HandleURI(fmt.Sprintf("hustle://swarm?action=set_goal&amount=%.2f", newGoal))
			} else {
				fmt.Println("Invalid amount. Goal unchanged.")
			}
		case "q":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
