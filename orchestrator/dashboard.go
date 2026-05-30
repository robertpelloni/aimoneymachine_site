package orchestrator

import (
	"fmt"
	"time"
)

// ShowDashboard renders a static view of the current machine state
func ShowDashboard(orch *Orchestrator) {
	fmt.Println("\033[H\033[2J") // Clear terminal
	fmt.Println("==================================================")
	fmt.Println("          AI HUSTLE MACHINE DASHBOARD             ")
	fmt.Println("   (Real-time visualization of machine health)    ")
	fmt.Println("==================================================")
	fmt.Printf(" [SYSTEM TIME]    %s\n", time.Now().Format("15:04:05"))
	fmt.Printf(" [MEMORY STATE]   L1:%d, L2:%d, L3:%d entries\n", len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries))
	fmt.Println("--------------------------------------------------")
	fmt.Println(" [FINANCIAL PERFORMANCE]")
	fmt.Printf("  Revenue:        $%.2f\n", orch.Ledger.TotalRevenue())
	fmt.Printf("  Expenses:       $%.2f\n", orch.Ledger.TotalExpenses())
	fmt.Printf("  NET PROFIT:     $%.2f\n", orch.Ledger.Profit())
	fmt.Println("--------------------------------------------------")
	fmt.Println(" [RECENT ACTIVITY LOG]")

	// Show last 3 L1 entries as events
	count := 0
	for i := len(orch.L1.Entries) - 1; i >= 0 && count < 3; i-- {
		fmt.Printf("  (%s) %s\n", orch.L1.Entries[i].Timestamp.Format("15:04"), orch.L1.Entries[i].Content)
		count++
	}
	if count == 0 {
		fmt.Println("  (No recent activity recorded)")
	}

	fmt.Println("==================================================")
	fmt.Println(" Tip: Use 'orchestrator -interactive' for controls.")
}

// StartLiveDashboard launches the dashboard in a refresh loop
func StartLiveDashboard(orch *Orchestrator) {
	for {
		ShowDashboard(orch)
		time.Sleep(2 * time.Second)
	}
}
