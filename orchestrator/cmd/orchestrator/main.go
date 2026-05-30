package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"os"
	"strings"
	"time"
)

func main() {
	hustleTask := flag.String("hustle", "", "Run a specific hustle module (research, social, curation)")
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

	if *interactive {
		runInteractiveMenu(orch, version)
		return
	}

	if *dashboard {
		orchestrator.StartLiveDashboard(orch)
		return
	}

	if *daemon {
		scheduler := orchestrator.NewScheduler(orch)
		// Register default heartbeat task
		scheduler.Register("Heartbeat", 5*time.Minute, func(o *orchestrator.Orchestrator) error {
			return orchestrator.WriteStatusReport(version, "Active", "Scheduler Heartbeat", o.Ledger)
		})

		fmt.Println("Orchestrator running in Daemon mode.")
		scheduler.Start()
		return
	}

	if *syncMode {
		fmt.Println("Triggering repository sync protocol...")
		// In a real app, this would execute sync.sh or internal git logic
	}

	if *hustleTask != "" {
		fmt.Printf("Launching hustle module: %s with params: %s\n", *hustleTask, *params)
	}

	// Real-time status reporting with financial metrics
	orchestrator.WriteStatusReport(version, "Running", "Orchestrator command processed", orch.Ledger)

	fmt.Printf("Orchestrator initialized with L1: %d, L2: %d, L3: %d entries. Profit: $%.2f\n",
		len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries), orch.Ledger.Profit())
}

func runInteractiveMenu(orch *orchestrator.Orchestrator, version string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- INTERACTIVE COMMAND MENU ---")
		fmt.Println("1. Launch Research Hustle")
		fmt.Println("2. Launch Social Hustle")
		fmt.Println("3. Launch Curation Hustle")
		fmt.Println("4. View Dashboard")
		fmt.Println("5. Run Repository Sync")
		fmt.Println("q. Quit")
		fmt.Print("Select an option: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			fmt.Println("Launching Research Hustle...")
			// Logic to invoke research module
		case "2":
			fmt.Println("Launching Social Hustle...")
		case "3":
			fmt.Println("Launching Curation Hustle...")
		case "4":
			orchestrator.ShowDashboard(orch)
			fmt.Println("\nPress Enter to return to menu...")
			reader.ReadString('\n')
		case "5":
			fmt.Println("Running Sync Protocol...")
		case "q":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option, please try again.")
		}
	}
}
