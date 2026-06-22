package orchestrator

import (
	"fmt"
	"os"
<<<<<<< HEAD
=======
	"sort"
	"strings"
>>>>>>> origin/main
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// ShowDashboard renders a static view of the current machine state
func ShowDashboard(orch *Orchestrator) {
	fmt.Println("\033[H\033[2J") // Clear terminal
	fmt.Printf("%s==================================================%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s          AI HUSTLE MACHINE DASHBOARD             %s\n", colorBold, colorYellow, colorReset)
	fmt.Println("   (Real-time visualization of machine health)    ")
	if orch.DryRun {
		fmt.Printf("      %s⚠️  [DRY-RUN MODE ACTIVE: NO POSTING]%s       \n", colorRed, colorReset)
	}
	fmt.Printf("%s==================================================%s\n", colorCyan, colorReset)
	fmt.Printf(" [SYSTEM TIME]    %s\n", time.Now().Format("15:04:05"))
	fmt.Printf(" [MEMORY STATE]   L1:%d, L2:%d, L3:%d entries\n", len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries))

	// Scheduler Status
	if len(orch.TaskQueue) > 0 {
		fmt.Printf(" [SCHEDULER]      Next: %s\n", strings.Join(orch.TaskQueue, " | "))
	}

	// Hyper-Diverse Hustle Metrics
	contentCount := len(orch.L1.Search("content"))
	productCount := len(orch.L1.Search("ecommerce"))
	auditCount := len(orch.L1.Search("devagency"))
	ticketCount := len(orch.L1.Search("support"))
	podCount := len(orch.L1.Search("pod"))
	mediaCount := len(orch.L1.Search("media"))
	careerCount := len(orch.L1.Search("careers"))
	financeCount := len(orch.L1.Search("finance"))
	biCount := len(orch.L1.Search("bi"))
	designCount := len(orch.L1.Search("design"))
	confCount := len(orch.L1.Search("confluence"))

	if contentCount > 0 || productCount > 0 || auditCount > 0 || ticketCount > 0 || podCount > 0 || mediaCount > 0 || careerCount > 0 || financeCount > 0 || biCount > 0 || designCount > 0 || confCount > 0 {
		fmt.Printf(" [HUSTLE METRICS] Content:%d | Product:%d | Agency:%d | Support:%d | POD:%d | Media:%d | Careers:%d | Fin:%d | BI:%d | Design:%d | Strategic:%d\n",
			contentCount, productCount, auditCount, ticketCount, podCount, mediaCount, careerCount, financeCount, biCount, designCount, confCount)
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

<<<<<<< HEAD
	fmt.Println("--------------------------------------------------")
	fmt.Println(" [SOCIAL PROVIDERS]")

	twitterStatus := "[✗ OFFLINE]"
	if os.Getenv("TWITTER_API_KEY") != "" && os.Getenv("TWITTER_ACCESS_TOKEN") != "" {
		twitterStatus = "[✓ ONLINE]"
	}
	fmt.Printf("  Twitter:        %s\n", twitterStatus)

	linkedInStatus := "[✗ OFFLINE]"
	if os.Getenv("LINKEDIN_ACCESS_TOKEN") != "" && os.Getenv("LINKEDIN_AUTHOR_URN") != "" {
		linkedInStatus = "[✓ ONLINE]"
	}
	fmt.Printf("  LinkedIn:       %s\n", linkedInStatus)

	fmt.Println("--------------------------------------------------")
	fmt.Println(" [FINANCIAL PERFORMANCE]")
=======
	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [SOCIAL PROVIDERS]\n")

	twitterStatus := fmt.Sprintf("%s[✗ OFFLINE]%s", colorRed, colorReset)
	if os.Getenv("TWITTER_API_KEY") != "" && os.Getenv("TWITTER_ACCESS_TOKEN") != "" {
		twitterStatus = fmt.Sprintf("%s[✓ ONLINE]%s", colorGreen, colorReset)
	}
	fmt.Printf("  Twitter:        %s\n", twitterStatus)

	linkedInStatus := fmt.Sprintf("%s[✗ OFFLINE]%s", colorRed, colorReset)
	if os.Getenv("LINKEDIN_ACCESS_TOKEN") != "" && os.Getenv("LINKEDIN_AUTHOR_URN") != "" {
		linkedInStatus = fmt.Sprintf("%s[✓ ONLINE]%s", colorGreen, colorReset)
	}
	fmt.Printf("  LinkedIn:       %s\n", linkedInStatus)

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [%sFINANCIAL PERFORMANCE%s]\n", colorBold, colorReset)
>>>>>>> origin/main
	fmt.Printf("  Revenue:        $%.2f\n", orch.Ledger.TotalRevenue())
	fmt.Printf("  Expenses:       $%.2f\n", orch.Ledger.TotalExpenses())

	profitColor := colorGreen
	if orch.Ledger.Profit() < 0 {
		profitColor = colorRed
	}
	fmt.Printf("  %sNET PROFIT:     $%.2f%s\n", profitColor, orch.Ledger.Profit(), colorReset)

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [%s%sLUXURY SPACE COMMUNISM (FEDERATED WEALTH)%s]\n", colorBold, colorYellow, colorReset)

	type peerProfit struct {
		id     string
		profit float64
	}
	var leaderboard []peerProfit

	collectiveProfit := orch.Ledger.Profit()
	leaderboard = append(leaderboard, peerProfit{"local-node", orch.Ledger.Profit()})

	meshEntries := orch.L1.Search("mesh_status")
	for _, e := range meshEntries {
		if idx := strings.Index(e.Content, "PROFIT: $"); idx != -1 {
			var p float64
			fmt.Sscanf(e.Content[idx:], "PROFIT: $%f", &p)
			collectiveProfit += p

			// Extract peer ID
			peerID := "unknown"
			fmt.Sscanf(e.Content, "Mesh Peer %s Status:", &peerID)
			leaderboard = append(leaderboard, peerProfit{peerID, p})
		}
	}
	collProfitColor := colorGreen
	if collectiveProfit < 0 {
		collProfitColor = colorRed
	}
	fmt.Printf("  COLLECTIVE MESH PROFIT: %s$%.2f%s\n", collProfitColor, collectiveProfit, colorReset)

	// Collective Goal Progress
	currentGoal := orch.WealthGoal
	if currentGoal <= 0 {
		currentGoal = 10000 // Default goal for display if not set
	}
	progress := (collectiveProfit / currentGoal) * 100
	if progress < 0 { progress = 0 }

	// ASCII Progress Bar
	barWidth := 20
	filled := int((progress / 100) * float64(barWidth))
	if filled < 0 { filled = 0 }
	if filled > barWidth { filled = barWidth }
	if filled < 0 { filled = 0 }
	empty := barWidth - filled
	if empty < 0 { empty = 0 }
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	fmt.Printf("  GOAL PROGRESS:  [%s] %s%.1f%%%s\n", bar, colorCyan, progress, colorReset)
	fmt.Printf("  TARGET WEALTH:  $%.2f\n", currentGoal)

	// Display leaderboard
	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].profit > leaderboard[j].profit
	})
	// Load Leaderboard Sync Entries from L1
	syncEntries := orch.L1.Search("leaderboard")
	for _, e := range syncEntries {
		if idx := strings.Index(e.Content, "Profit: $"); idx != -1 {
			var amount float64
			fmt.Sscanf(e.Content[idx:], "Profit: $%f", &amount)

			// Extract peer ID
			peerID := "unknown"
			fmt.Sscanf(e.Content, "Leaderboard Sync: Peer %s Profit:", &peerID)

			// Update or Add
			found := false
			for j, lp := range leaderboard {
				if lp.id == peerID {
					leaderboard[j].profit = amount
					found = true
					break
				}
			}
			if !found {
				leaderboard = append(leaderboard, peerProfit{peerID, amount})
			}
		}
	}

	fmt.Printf("\n  %sPEER LEADERBOARD:%s\n", colorBold, colorReset)
	fmt.Printf("  %-15s %-15s %-10s\n", "RANK", "PEER ID", "PROFIT")
	fmt.Println("  " + strings.Repeat("-", 42))
	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].profit > leaderboard[j].profit
	})
	for i := 0; i < len(leaderboard) && i < 10; i++ {
		rankColor := colorReset
		if i == 0 { rankColor = colorYellow }
		fmt.Printf("  %s#%-14d %-15s $%-10.2f%s\n", rankColor, i+1, leaderboard[i].id, leaderboard[i].profit, colorReset)
	}

	// Collective Alpha Insight
	strategies := orch.L1.Search("collective_strategy")
	if len(strategies) > 0 {
		latest := strategies[len(strategies)-1]
		fmt.Printf("  COLLECTIVE ALPHA: %s\n", latest.Content)
	}

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	// Active Swarm Fixes
	swarmFixes := orch.L1.Search("swarm_fix")
	if len(swarmFixes) > 0 {
		latest := swarmFixes[len(swarmFixes)-1]
		fmt.Printf(" [SWARM HEALING]  Active Fix: %s\n", latest.Content)
	}

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [MESH SWARM STATUS]\n")
	fmt.Printf("  %-20s %-15s %-10s\n", "PEER ID", "HEALTH", "VERSION")
	fmt.Println("  " + strings.Repeat("-", 47))

	// Find mesh status entries in L1
	if len(meshEntries) > 0 {
		for _, e := range meshEntries {
			// Extract data from "Mesh Peer test-peer Status: Active, PROFIT: $500.50"
			peerID := "unknown"
			status := "unknown"
			fmt.Sscanf(e.Content, "Mesh Peer %s Status: %s", &peerID, &status)
			status = strings.TrimSuffix(status, ",")

			healthColor := colorGreen
			if status != "Active" { healthColor = colorRed }

			fmt.Printf("  %-20s %s%-15s%s %-10s\n", peerID, healthColor, status, colorReset, "v1.0.0")
		}
	} else {
		fmt.Println("  (No remote mesh data aggregated yet)")
	}

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [SPACE COMMS (MESH TRAFFIC)]\n")

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

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [RECENT TASK LOG (SQL)]\n")
	if orch.DB != nil {
		history, _ := orch.DB.GetTaskHistory(3)
		if len(history) > 0 {
			for _, h := range history {
				duration := time.Duration(h.DurationMs) * time.Millisecond
				statusColor := colorGreen
				if h.Status != "success" {
					statusColor = colorRed
				}
				fmt.Printf("  [%s] %s (%v) - %s%s%s\n", h.Timestamp.Format("15:04"), h.TaskID, duration, statusColor, h.Status, colorReset)
			}
		} else {
			fmt.Println("  (No task history logged yet)")
		}
	} else {
		fmt.Printf("  %s(Database offline)%s\n", colorRed, colorReset)
	}

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [RECENT ACTIVITY LOG (L1)]\n")

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

// ShowTaskHistory displays the recent task execution logs
func ShowTaskHistory(orch *Orchestrator) {
	if orch.DB == nil {
		fmt.Println("Error: SQLite store not initialized.")
		return
	}

	history, err := orch.DB.GetTaskHistory(20)
	if err != nil {
		fmt.Printf("Error fetching task history: %v\n", err)
		return
	}

	fmt.Println("\n--- RECENT TASK EXECUTION HISTORY ---")
	fmt.Printf("%-20s %-10s %-10s %-20s\n", "TASK NAME", "DURATION", "STATUS", "MESSAGE")
	fmt.Println(strings.Repeat("-", 65))

	for _, h := range history {
		msg := h.Message
		if len(msg) > 20 {
			msg = msg[:17] + "..."
		}
		duration := time.Duration(h.DurationMs) * time.Millisecond
		fmt.Printf("%-20s %-10v %-10s %-20s\n", h.TaskID, duration, h.Status, msg)
	}
	fmt.Println("--------------------------------------------------")
}

// StartLiveDashboard launches the dashboard in a refresh loop
func StartLiveDashboard(orch *Orchestrator) {
	for {
		ShowDashboard(orch)
		time.Sleep(2 * time.Second)
	}
}
