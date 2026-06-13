package orchestrator

import (
	"fmt"
	"os"
	"strings"
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

func ShowDashboard(orch *Orchestrator) {
	fmt.Println("\033[H\033[2J")
	fmt.Printf("%s==================================================%s\n", colorCyan, colorReset)
	fmt.Printf("%s%s          AI HUSTLE MACHINE DASHBOARD             %s\n", colorBold, colorYellow, colorReset)
	fmt.Println("   (Real-time visualization of machine health)    ")
	if orch.DryRun {
		fmt.Printf("      %s⚠️  [DRY-RUN MODE ACTIVE: NO POSTING]%s       \n", colorRed, colorReset)
	}
	fmt.Printf("%s==================================================%s\n", colorCyan, colorReset)
	fmt.Printf(" [SYSTEM TIME]    %s\n", time.Now().Format("15:04:05"))
	fmt.Printf(" [MEMORY STATE]   L1:%d, L2:%d, L3:%d entries\n", len(orch.L1.Entries), len(orch.L2.Entries), len(orch.L3.Entries))

	if orch.DB != nil {
		count, _ := orch.DB.GetCacheCount()
		cached, isCached := orch.LLM.(*CachedLLM)
		status := "DISABLED"
		statusColor := colorRed
		if isCached && cached.Enabled {
			status = "ENABLED"
			statusColor = colorGreen
		}
		fmt.Printf(" [LLM CACHE]      %d responses | Status: %s%s%s\n", count, statusColor, status, colorReset)
	}

	if len(orch.TaskQueue) > 0 {
		fmt.Printf(" [SCHEDULER]      Next: %s\n", strings.Join(orch.TaskQueue, " | "))
	}

	contentEntries := orch.L1.Search("content")
	if len(contentEntries) > 0 {
		fmt.Printf(" [CONTENT HUB]    Generated: %d pieces\n", len(contentEntries))
	}

	agentEntries := orch.L1.Search("agent_loop")
	if len(agentEntries) > 0 {
		successCount, failCount := 0, 0
		activeAgents := make(map[string]bool)
		for _, e := range agentEntries {
			for _, t := range e.Tags {
				if t == "success" { successCount++ }
				if t == "failure" { failCount++ }
				if !strings.HasPrefix(t, "agent_loop") && t != "success" && t != "failure" { activeAgents[t] = true }
			}
		}
		var agentList []string
		for a := range activeAgents { agentList = append(agentList, a) }
		fmt.Printf(" [ACTIVE AGENTS]  %s\n", strings.Join(agentList, ", "))
		fmt.Printf(" [AGENT METRICS]  Success: %d | Errors: %d\n", successCount, failCount)
		lastAgent := agentEntries[len(agentEntries)-1]
		fmt.Printf(" [LAST ACTION]    %s\n", lastAgent.Content)
	}

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
	fmt.Printf("  Revenue:        $%.2f\n", orch.Ledger.TotalRevenue())
	fmt.Printf("  Expenses:       $%.2f\n", orch.Ledger.TotalExpenses())
	profitColor := colorGreen
	if orch.Ledger.Profit() < 0 { profitColor = colorRed }
	fmt.Printf("  %sNET PROFIT:     $%.2f%s\n", profitColor, orch.Ledger.Profit(), colorReset)

	fmt.Printf("%s--------------------------------------------------%s\n", colorCyan, colorReset)
	fmt.Printf(" [%s%sLUXURY SPACE COMMUNISM (FEDERATED WEALTH)%s]\n", colorBold, colorYellow, colorReset)
	collectiveProfit := orch.Ledger.Profit()
	meshEntries := orch.L1.Search("mesh_status")
	for _, e := range meshEntries {
		if idx := strings.Index(e.Content, "PROFIT: $"); idx != -1 {
			var p float64
			fmt.Sscanf(e.Content[idx:], "PROFIT: $%f", &p)
			collectiveProfit += p
		}
	}
	collProfitColor := colorGreen
	if collectiveProfit < 0 { collProfitColor = colorRed }
	fmt.Printf("  COLLECTIVE MESH PROFIT: %s$%.2f%s\n", collProfitColor, collectiveProfit, colorReset)

	meshGoal := orch.WealthGoal
	if meshGoal <= 0 { meshGoal = 1 }
	progress := (collectiveProfit / meshGoal) * 100
	if progress < 0 { progress = 0 }
	barWidth := 20
	filled := int((progress / 100) * float64(barWidth))
	if filled > barWidth { filled = barWidth }
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("  GOAL PROGRESS:  [%s] %s%.1f%%%s\n", bar, colorCyan, progress, colorReset)
	fmt.Printf("  TARGET WEALTH:  $%.2f\n", meshGoal)

	fmt.Println("==================================================")
	fmt.Println(" Tip: Use 'orchestrator -interactive' for controls.")
}

func ShowTaskHistory(orch *Orchestrator) {
	if orch.DB == nil { return }
	history, _ := orch.DB.GetTaskHistory(20)
	fmt.Println("\n--- RECENT TASK EXECUTION HISTORY ---")
	for _, h := range history {
		duration := time.Duration(h.DurationMs) * time.Millisecond
		fmt.Printf("%-20s %-10v %-10s %-20s\n", h.TaskID, duration, h.Status, h.Message)
	}
}

func StartLiveDashboard(orch *Orchestrator, refreshRate time.Duration) {
	if refreshRate <= 0 {
		refreshRate = 1 * time.Second
	}
	for {
		ShowDashboard(orch)
		time.Sleep(refreshRate)
	}
}
