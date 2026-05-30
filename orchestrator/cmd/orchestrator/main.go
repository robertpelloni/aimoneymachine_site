package main

import (
	"flag"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"os"
	"strings"
)

func main() {
	hustleTask := flag.String("hustle", "", "Run a specific hustle module (research, social)")
	syncMode := flag.Bool("sync", false, "Run repository synchronization protocol")
	params := flag.String("params", "{}", "JSON parameters for the hustle module")
	dashboard := flag.Bool("dashboard", false, "Start the live terminal dashboard")
	flag.Parse()

	fmt.Println("=== AI Hustle Machine Orchestrator ===")

	// Source version from VERSION.md
	versionData, err := os.ReadFile("VERSION.md")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionData))
	}

	orch := orchestrator.NewOrchestrator()

	if *dashboard {
		orchestrator.ShowDashboard(orch)
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

	fmt.Printf("Orchestrator (v%s) initialized with L1: %d, L2: %d, L3: %d entries. Profit: $%.2f\n",
		version, len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries), orch.Ledger.Profit())
}
