package orchestrator

import (
	"fmt"
	"time"
)

func ShowDashboard(orch *Orchestrator) {
	fmt.Println("\033[H\033[2J") // Clear terminal
	fmt.Println("==================================================")
	fmt.Println("          AI HUSTLE MACHINE DASHBOARD             ")
	fmt.Println("==================================================")
	fmt.Printf(" Time:    %s\n", time.Now().Format("15:04:05"))
	fmt.Printf(" Memory:  L1:%d, L2:%d, L3:%d\n", len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries))
	fmt.Println("--------------------------------------------------")
	fmt.Printf(" Revenue: $%.2f\n", orch.Ledger.TotalRevenue())
	fmt.Printf(" Expense: $%.2f\n", orch.Ledger.TotalExpenses())
	fmt.Printf(" Profit:  $%.2f\n", orch.Ledger.Profit())
	fmt.Println("--------------------------------------------------")
	fmt.Println(" Recent Events:")

	// Show last 3 L1 entries as events
	count := 0
	for i := len(orch.L1.Entries) - 1; i >= 0 && count < 3; i-- {
		fmt.Printf(" [%s] %s\n", orch.L1.Entries[i].Timestamp.Format("15:04"), orch.L1.Entries[i].Content)
		count++
	}

	fmt.Println("==================================================")
}

func StartLiveDashboard(orch *Orchestrator) {
	for {
		ShowDashboard(orch)
		time.Sleep(2 * time.Second)
	}
}
