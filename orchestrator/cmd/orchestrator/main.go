package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/robertpelloni/hustle/hustle/curation"
	"github.com/robertpelloni/hustle/hustle/research"
	"github.com/robertpelloni/hustle/hustle/social"
	"github.com/robertpelloni/hustle/hustle/trading"
	"github.com/robertpelloni/hustle/orchestrator"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	hustleTask := flag.String("hustle", "", "Run a specific hustle module (research, social, curation, chain, trading, swarm)")
	hustleURI := flag.String("uri", "", "Run a hustle module via hustle:// URI")
	syncMode := flag.Bool("sync", false, "Run repository synchronization protocol")
	params := flag.String("params", "{}", "JSON parameters for the hustle module")
	dashboard := flag.Bool("dashboard", false, "Start the live terminal dashboard")
	daemon := flag.Bool("daemon", false, "Run the Orchestrator as a background task scheduler")
	interactive := flag.Bool("interactive", false, "Launch the interactive command menu")
	apiPort := flag.String("api", "", "Start the HTTP API on specified port (e.g. 8080)")
	flag.Parse()

	// Source version from VERSION.md
	versionData, err := os.ReadFile("VERSION.md")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionData))
	}

	fmt.Printf("=== AI Hustle Machine Orchestrator v%s ===\n", version)

	orch := orchestrator.NewOrchestrator()
	orch.Version = version
	protocol := orchestrator.NewHustleProtocol()
	broker := orchestrator.NewA2ABroker(orch)
	swarm := orchestrator.NewMemorySwarm(orch, broker)

	// Initialize Trading for event handling
	traderModule := &trading.TradingModule{
		Orchestrator: orch,
		Broker:       broker,
		Fetcher:      &trading.MockPriceFetcher{},
	}

	socialModule := social.NewSocialModule(orch)

	// Mesh Event Listener: Alpha Discovery
	alphaCh := broker.SubscribeTopic("alpha_discovery")
	go func() {
		for msg := range alphaCh {
			fmt.Printf("[Mesh] Received Alpha Discovery: %s\n", msg.Payload)
			traderModule.AddToWatchlist(msg.Payload)
		}
	}()

	// Mesh Event Listener: Swarm Sync
	syncCh := broker.SubscribeTopic("swarm_sync")
	go func() {
		for msg := range syncCh {
			fmt.Printf("[Mesh] Swarm Sync Notification from %s\n", msg.Source)
			swarm.HandleSyncRequest(msg.Source)
		}
	}()

	// Agent Direct Message Listener (Handles incoming hustle:// URIs from mesh)
	agentCh := broker.Subscribe("local-node")
	go func() {
		for msg := range agentCh {
			if strings.HasPrefix(msg.Payload, "hustle://") {
				fmt.Printf("[A2A] Executing Mesh Protocol URI from %s: %s\n", msg.Source, msg.Payload)
				protocol.HandleURI(msg.Payload)
			}
		}
	}()

	// Register Handlers
	protocol.Register("research", func(p url.Values) error {
		action := p.Get("action")
		query := p.Get("query")
		if query == "" { query = "AI Trends" }

		searcher := research.NewResearchSearch(research.Tavily, orch, broker)
		results, err := searcher.Query(query)
		if err != nil { return err }

		if action == "analyze" && len(results) > 0 {
			fmt.Println("[Research] Performing deep AI analysis of findings...")
			for _, res := range results {
				prompt := fmt.Sprintf("Summarize the revolutionary potential of this finding: %s", res.Snippet)
				summary, _ := orch.LLM.Generate(prompt)
				fmt.Printf("[Research] Deep Insight: %s\n", summary)
			}
		}

		return nil
	})
	protocol.Register("curation", func(p url.Values) error {
		topic := p.Get("topic")
		if topic == "" { topic = "AI" }

		c := &curation.CurationModule{
			Orchestrator: orch,
			Broker:       broker,
			Fetcher:      curation.NewRSSFetcher(),
			Feeds:        []string{"https://news.ycombinator.com/rss"},
		}

		// Support dynamic feed addition via URI
		extraFeed := p.Get("feed")
		if extraFeed != "" {
			c.AddFeed(extraFeed)
		}

		return c.Curate(topic)
	})
	protocol.Register("social", func(p url.Values) error {
		action := p.Get("action")
		platform := p.Get("platform")
		if platform == "" { platform = "Twitter" }

		// Standardize platform key
		platKey := "Twitter"
		if strings.ToLower(platform) == "linkedin" {
			platKey = "LinkedIn"
		}

		provider, ok := socialModule.Providers[platKey]
		if !ok {
			return fmt.Errorf("social provider %s not found", platform)
		}

		switch action {
		case "auth":
			code := p.Get("code")
			if code == "" {
				authURL := provider.Auth.Config.AuthCodeURL("state-token")
				fmt.Printf("[Social] Please authorize %s at: %s\n", platform, authURL)
				return nil
			}
			return provider.Auth.Exchange(context.Background(), code)
		default:
			topic := p.Get("topic")
			if topic == "" { topic = "AI" }
			social.SchedulePost(orch, provider, platform, topic)
		}
		return nil
	})
	protocol.Register("trading", func(p url.Values) error {
		action := p.Get("action")
		symbol := p.Get("symbol")
		if symbol == "" { symbol = "BTC" }
		traderModule.Symbol = symbol

		fetcherType := p.Get("fetcher")
		if fetcherType == "coingecko" {
			traderModule.Fetcher = &trading.CoinGeckoFetcher{}
		} else {
			traderModule.Fetcher = &trading.MockPriceFetcher{}
		}

		if action == "all" {
			return traderModule.ExecuteAll()
		}

		return traderModule.ExecuteStrategy()
	})
	protocol.Register("swarm", func(p url.Values) error {
		action := p.Get("action")
		peerID := p.Get("peer_id")
		if peerID == "" { peerID = "unknown-peer" }

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
			content := p.Get("content")
			orch.L2.Add(orchestrator.MemoryEntry{
				ID:        id,
				Content:   content,
				Timestamp: time.Now(),
				Tags:      []string{"swarm", "received", "from:" + peerID},
			})
			fmt.Printf("[Swarm] Successfully ingested entry %s from %s\n", id, peerID)
		case "status":
			fmt.Println("--- MESH SWARM STATUS ---")
			fmt.Printf("Connected Peers: %d\n", len(broker.Peers))
			for id, url := range broker.Peers {
				fmt.Printf("  - %s: %s\n", id, url)
			}
			fmt.Printf("Subscribed Topics: %d\n", len(broker.Topics))
			for topic := range broker.Topics {
				fmt.Printf("  - %s\n", topic)
			}
		}
		return nil
	})
	protocol.Register("chain", func(p url.Values) error {
		return runCurationChain(orch, broker)
	})
	protocol.Register("healer", func(p url.Values) error {
		issue := p.Get("issue")
		if issue == "" { issue = "Unknown system instability" }
		h := orchestrator.NewHealer(orch)
		h.Loop(issue)
		return nil
	})
	protocol.Register("ledger", func(p url.Values) error {
		fmt.Println("--- FINANCIAL LEDGER SUMMARY ---")
		fmt.Printf("Total Revenue:  $%.2f\n", orch.Ledger.TotalRevenue())
		fmt.Printf("Total Expenses: $%.2f\n", orch.Ledger.TotalExpenses())
		fmt.Printf("Total Profit:   $%.2f\n", orch.Ledger.Profit())
		fmt.Printf("Transaction Count: %d\n", len(orch.Ledger.Transactions))
		return nil
	})

	if *apiPort != "" {
		api := orchestrator.NewAPI(orch, protocol, broker)
		go api.Start(*apiPort)
	}

	if *interactive {
		runInteractiveMenu(orch, protocol, broker, traderModule, version)
		return
	}

	if *dashboard {
		orchestrator.StartLiveDashboard(orch)
		return
	}

	if *daemon {
		scheduler := orchestrator.NewScheduler(orch)

		// Register Hustle Tasks using protocol
		scheduler.Register("Research", 1*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://research?query=AI+Agent+Trends")
		})

		scheduler.Register("CurationChain", 2*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://chain")
		})

		scheduler.Register("Trading", 30*time.Minute, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://trading?symbol=BTC")
		})

		scheduler.Register("SwarmSync", 4*time.Hour, func(o *orchestrator.Orchestrator) error {
			return protocol.HandleURI("hustle://swarm?action=sync")
		})

		// Continuous Autonomous Execution: Sync Protocol
		scheduler.Register("Sync", 6*time.Hour, func(o *orchestrator.Orchestrator) error {
			return runSyncProtocol()
		})

		scheduler.Register("Heartbeat", 5*time.Minute, func(o *orchestrator.Orchestrator) error {
			return orchestrator.WriteStatusReport(version, "Active", "Scheduler Heartbeat", o.Ledger)
		})

		scheduler.Register("ProfitAnalysis", 12*time.Hour, func(o *orchestrator.Orchestrator) error {
			suggestion := o.Ledger.AnalyzeProfitability()
			fmt.Printf("[Scheduler] Financial Analysis: %s\n", suggestion)

			// Self-Evolving: Dynamically adjust scheduler based on profit
			scheduler.ReevaluateStrategy(suggestion)

			o.L2.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("profit-analysis-%d", time.Now().Unix()),
				Content:   suggestion,
				Timestamp: time.Now(),
				Tags:      []string{"ledger", "analysis", "recommendation"},
			})
			return nil
		})

		fmt.Println("Orchestrator running in Daemon mode.")
		scheduler.Start()
		return
	}

	if *syncMode {
		fmt.Println("Triggering repository sync protocol...")
		if err := runSyncProtocol(); err != nil {
			fmt.Printf("Sync error: %v\n", err)
		}
	}

	if *hustleURI != "" {
		fmt.Printf("Executing hustle URI: %s\n", *hustleURI)
		if err := protocol.HandleURI(*hustleURI); err != nil {
			fmt.Printf("Protocol error: %v\n", err)
		}
	} else if *hustleTask != "" {
		fmt.Printf("Launching hustle module: %s with params: %s\n", *hustleTask, *params)
		// Map old style task to URI for consistency
		uri := fmt.Sprintf("hustle://%s", *hustleTask)
		if err := protocol.HandleURI(uri); err != nil {
			fmt.Printf("Protocol error: %v\n", err)
		}
	}

	// If we started API but no other long running mode, we need to wait
	if *apiPort != "" {
		fmt.Println("API running. Press Ctrl+C to terminate.")
		select {}
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

func runCurationChain(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) error {
	fmt.Println("--- STARTING CURATION CHAIN ---")

	// 1. Curate
	c := &curation.CurationModule{
		Orchestrator: orch,
		Broker:       broker,
		Fetcher:      curation.NewRSSFetcher(),
		Feeds:        []string{"https://news.ycombinator.com/rss"},
	}
	err := c.Curate("AI")
	if err != nil {
		return fmt.Errorf("curation failed: %v", err)
	}

	// 2. Fetch last curated blurb from L1 memory
	memories := orch.L1.Search("curation")
	if len(memories) == 0 {
		return fmt.Errorf("no curated content found in memory")
	}
	lastCurated := memories[len(memories)-1].Content

	// 3. Post to Social
	fmt.Println("Forwarding curated blurb to Social module...")
	mod := social.NewSocialModule(orch)
	provider := mod.Providers["Twitter"]
	// We use the curated content as a basis for the social post
	social.SchedulePost(orch, provider, "Twitter", "the following curated insight: "+lastCurated)

	fmt.Println("--- CURATION CHAIN COMPLETE ---")
	return nil
}

func runInteractiveMenu(orch *orchestrator.Orchestrator, protocol *orchestrator.HustleProtocol, broker *orchestrator.A2ABroker, traderModule *trading.TradingModule, version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- INTERACTIVE COMMAND MENU ---")
		fmt.Println("1. Launch Research Hustle")
		fmt.Println("2. Launch Social Hustle")
		fmt.Println("3. Launch Curation Hustle")
		fmt.Println("4. Launch Trading Hustle")
		fmt.Println("5. Launch FULL CHAIN (Curate -> Post)")
		fmt.Println("6. Trigger Swarm Sync")
		fmt.Println("7. Broadcast Custom Mesh Event")
		fmt.Println("8. Trading: SELL ALL & Clear History")
		fmt.Println("9. Data: Reset L1 Memory")
		fmt.Println("10. Data: Backup Ledger")
		fmt.Println("11. View Dashboard")
		fmt.Println("12. Run Repository Sync")
		fmt.Println("q. Quit")
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
			traderModule.History = []float64{}
			traderModule.RSIHistory = []float64{}
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
			fmt.Println("Resetting L1 (Audit) memory...")
			orch.L1.Entries = []orchestrator.MemoryEntry{}
			fmt.Println("L1 memory cleared.")
		case "10":
			backupFile := fmt.Sprintf("ledger_backup_%d.json", time.Now().Unix())
			if err := orch.Ledger.Save(backupFile); err != nil {
				fmt.Printf("Backup failed: %v\n", err)
			} else {
				fmt.Printf("Ledger backed up to %s\n", backupFile)
			}
		case "11":
			orchestrator.ShowDashboard(orch)
			fmt.Println("\nPress Enter to return to menu...")
			reader.ReadString('\n')
		case "12":
			if err := runSyncProtocol(); err != nil {
				fmt.Printf("Sync error: %v\n", err)
			}
		case "q":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
