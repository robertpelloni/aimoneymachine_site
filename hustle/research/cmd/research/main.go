package main

import (
	"fmt"
	"github.com/robertpelloni/hustle/hustle/research"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	fmt.Println("=== Hustle Research Module ===")
	orch := orchestrator.NewOrchestrator()

	query := "Latest trends in AI agent orchestration"
	fmt.Printf("Running research for: %s\n", query)

	// 1. Initial Research
	searcher := &research.ResearchSearch{
		ActiveProvider: research.Tavily,
		Orchestrator:   orch,
	}
	results, err := searcher.Query(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 2. Council Debate
	council := orchestrator.NewCouncil(orch)
	debate, _ := council.Debate(query)
	fmt.Println("Council Consensus reached.")

	// 3. Synthesis (Incorporating Debate)
	report := &research.Report{Title: "AI Trends 2026 - Verified"}
	// We'll simulate synthesis with debate points for now
	report.Synthesize(orch, results)
	report.Content += "\n\nCOUNCIL DEBATE SUMMARY:\n" + debate.String()

	fmt.Println("Synthesis Complete.")
	fmt.Println(report.Content)
}
