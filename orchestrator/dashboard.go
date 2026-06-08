package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

// ShowDashboard renders a static view of the current machine state
func ShowDashboard(orch *Orchestrator) {
	fmt.Println("\033[H\033[2J") // Clear terminal
	fmt.Println("==================================================")
	fmt.Println("          AI HUSTLE MACHINE DASHBOARD             ")
	fmt.Println("   (Real-time visualization of machine health)    ")
	if orch.DryRun {
		fmt.Println("      ⚠️  [DRY-RUN MODE ACTIVE: NO POSTING]       ")
	}
	fmt.Println("==================================================")
	fmt.Printf(" [SYSTEM TIME]    %s\n", time.Now().Format("15:04:05"))
	fmt.Printf(" [MEMORY STATE]   L1:%d, L2:%d, L3:%d entries\n", len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries))

	// Scheduler Status
	if len(orch.TaskQueue) > 0 {
		fmt.Printf(" [SCHEDULER]      Next: %s\n", strings.Join(orch.TaskQueue, " | "))
	}

	// Content Metrics
	contentCount := 0
	contentEntries := orch.L1.Search("content")
	contentCount = len(contentEntries)
	if contentCount > 0 {
		fmt.Printf(" [CONTENT HUB]    Generated: %d pieces\n", contentCount)
	}

	// Agent Observability
	agentEntries := orch.L1.Search("agent_loop")
	if len(agentEntries) > 0 {
		successCount := 0
		failCount := 0
		activeAgents := make(map[string]bool)
		for _, e := range agentEntries {
			for _, t := range e.Tags {
				if t == "success" { successCount++ }
				if t == "failure" { failCount++ }
				if !strings.HasPrefix(t, "agent_loop") && t != "success" && t != "failure" {
					activeAgents[t] = true
				}
			}
		}

		var agentList []string
		for a := range activeAgents { agentList = append(agentList, a) }
		fmt.Printf(" [ACTIVE AGENTS]  %s\n", strings.Join(agentList, ", "))
		fmt.Printf(" [AGENT METRICS]  Success: %d | Errors: %d\n", successCount, failCount)

		// Type-specific Metrics
		for _, aType := range agentList {
			sType := 0
			eType := 0
			for _, e := range agentEntries {
				hasType := false
				for _, t := range e.Tags { if t == aType { hasType = true; break } }
				if !hasType { continue }
				for _, t := range e.Tags {
					if t == "success" { sType++ }
					if t == "failure" { eType++ }
				}
			}
			fmt.Printf("   - %s: %d OK, %d ERR\n", aType, sType, eType)
		}

		lastAgent := agentEntries[len(agentEntries)-1]
		fmt.Printf(" [LAST ACTION]    %s\n", lastAgent.Content)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println(" [FINANCIAL PERFORMANCE]")
	fmt.Printf("  Revenue:        $%.2f\n", orch.Ledger.TotalRevenue())
	fmt.Printf("  Expenses:       $%.2f\n", orch.Ledger.TotalExpenses())
	fmt.Printf("  NET PROFIT:     $%.2f\n", orch.Ledger.Profit())
	fmt.Println("--------------------------------------------------")
	fmt.Println(" [LUXURY SPACE COMMUNISM (FEDERATED WEALTH)]")

	collectiveProfit := orch.Ledger.Profit()
	meshEntries := orch.L1.Search("mesh_status")
	for _, e := range meshEntries {
		// Parse peer's profit from the status message "PROFIT: $..."
		if idx := strings.Index(e.Content, "PROFIT: $"); idx != -1 {
			var p float64
			fmt.Sscanf(e.Content[idx:], "PROFIT: $%f", &p)
			collectiveProfit += p
		}
	}
	fmt.Printf("  COLLECTIVE MESH PROFIT: $%.2f\n", collectiveProfit)

	fmt.Println("--------------------------------------------------")
	fmt.Println(" [MESH SWARM OVERVIEW]")

	// Find mesh status entries in L1
	if len(meshEntries) > 0 {
		for _, e := range meshEntries {
			fmt.Printf("  [PEER] %s\n", e.Content)
		}
	} else {
		fmt.Println("  (No remote mesh data aggregated yet)")
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println(" [SPACE COMMS (MESH TRAFFIC)]")

	// Filter for mesh-specific messages
	meshMsgs := orch.L1.Search("mesh")
	if len(meshMsgs) > 0 {
		if len(meshMsgs) > 3 {
			meshMsgs = meshMsgs[len(meshMsgs)-3:]
		}
		for _, m := range meshMsgs {
			fmt.Printf("  [COMMS] %s\n", m.Content)
		}
	} else {
		fmt.Println("  (No active space communication detected)")
	}

	fmt.Println("--------------------------------------------------")
	fmt.Println(" [RECENT ACTIVITY LOG]")

	// Show last 3 L1 entries as events (excluding mesh_status to avoid duplication)
	count := 0
	for i := len(orch.L1.Entries) - 1; i >= 0 && count < 3; i-- {
		isMesh := false
		for _, t := range orch.L1.Entries[i].Tags {
			if t == "mesh_status" {
				isMesh = true
				break
			}
		}
		if isMesh { continue }

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
