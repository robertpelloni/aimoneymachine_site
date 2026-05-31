package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/robertpelloni/hustle/hustle/curation"
	"github.com/robertpelloni/hustle/hustle/research"
	"github.com/robertpelloni/hustle/hustle/social"
	"github.com/robertpelloni/hustle/orchestrator"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	hustleTask := flag.String("hustle", "", "Run a specific hustle module (research, social, curation, chain)")
	hustleURI := flag.String("uri", "", "Run a hustle module via hustle:// URI")
	syncMode := flag.Bool("sync", false, "Run repository synchronization protocol")
	params := flag.String("params", "{}", "JSON parameters for the hustle module")
	dashboard := flag.Bool("dashboard", false, "Start the live terminal dashboard")
	daemon := flag.Bool("daemon", false, "Run the Orchestrator as a background task scheduler")
	interactive := flag.Bool("interactive", false, "Launch the interactive command menu")
	flag.Parse()

	// Source version from VERSION.md
	versionData, err := os.ReadFile("VERSION.md")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionData))
	}

	fmt.Printf("=== AI Hustle Machine Orchestrator v%s ===\n", version)

	orch := orchestrator.NewOrchestrator()
	protocol := orchestrator.NewHustleProtocol()

	// Register Handlers
	protocol.Register("research", func(p url.Values) error {
		query := p.Get("query")
		if query == "" { query = "AI Trends" }
		searcher := &research.ResearchSearch{Orchestrator: orch}
		_, err := searcher.Query(query)
		return err
	})
	protocol.Register("curation", func(p url.Values) error {
		topic := p.Get("topic")
		if topic == "" { topic = "AI" }
		c := &curation.CurationModule{
			Orchestrator: orch,
			Fetcher:      curation.NewRSSFetcher(),
			Feeds:        []string{"https://news.ycombinator.com/rss"},
		}
		return c.Curate(topic)
	})
	protocol.Register("social", func(p url.Values) error {
		platform := p.Get("platform")
		if platform == "" { platform = "Twitter" }
		topic := p.Get("topic")
		if topic == "" { topic = "AI" }

		var provider social.Provider = social.NewTwitterProvider()
		if strings.ToLower(platform) == "linkedin" {
			provider = social.NewLinkedInProvider()
		}
		social.SchedulePost(orch, provider, platform, topic)
		return nil
	})
	protocol.Register("chain", func(p url.Values) error {
		return runCurationChain(orch)
	})

	if *interactive {
		runInteractiveMenu(orch, protocol, version)
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

		scheduler.Register("Heartbeat", 5*time.Minute, func(o *orchestrator.Orchestrator) error {
			return orchestrator.WriteStatusReport(version, "Active", "Scheduler Heartbeat", o.Ledger)
		})

		fmt.Println("Orchestrator running in Daemon mode.")
		scheduler.Start()
		return
	}

	if *syncMode {
		fmt.Println("Triggering repository sync protocol...")
		// Internal call to sync logic would go here
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

	// Real-time status reporting with financial metrics
	orchestrator.WriteStatusReport(version, "Running", "Orchestrator command processed", orch.Ledger)

	fmt.Printf("Orchestrator initialized with L1: %d, L2: %d, L3: %d entries. Profit: $%.2f\n",
		len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries), orch.Ledger.Profit())
}

func runCurationChain(orch *orchestrator.Orchestrator) error {
	fmt.Println("--- STARTING CURATION CHAIN ---")

	// 1. Curate
	c := &curation.CurationModule{
		Orchestrator: orch,
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
	provider := social.NewTwitterProvider()
	social.SchedulePost(orch, provider, "Twitter", "the following curated insight: "+lastCurated)

	fmt.Println("--- CURATION CHAIN COMPLETE ---")
	return nil
}

func runInteractiveMenu(orch *orchestrator.Orchestrator, protocol *orchestrator.HustleProtocol, version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- INTERACTIVE COMMAND MENU ---")
		fmt.Println("1. Launch Research Hustle")
		fmt.Println("2. Launch Social Hustle")
		fmt.Println("3. Launch Curation Hustle")
		fmt.Println("4. Launch FULL CHAIN (Curate -> Post)")
		fmt.Println("5. View Dashboard")
		fmt.Println("6. Run Repository Sync")
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
			protocol.HandleURI("hustle://chain")
		case "5":
			orchestrator.ShowDashboard(orch)
			fmt.Println("\nPress Enter to return to menu...")
			reader.ReadString('\n')
		case "6":
			fmt.Println("Running Sync Protocol...")
		case "q":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
