package research

import (
	"github.com/robertpelloni/hustle/orchestrator"
	"os"
	"testing"
)

func TestResearchPipelineE2E(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	orch.LLM = &orchestrator.MockLLM{Response: "The AI agent market is expanding rapidly."}

	// 1. Multi-Provider Search with Orchestrator Integration
	searcher := &MockSearch{
		Orchestrator:   orch,
	}
	results, err := searcher.Query("Advanced automated trading strategies")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// 2. Enhanced Report Synthesis
	report := &Report{Title: "Trading Alpha Report"}
	report.Synthesize(orch, results)

	// 3. Real File Export
	testFile := "test_trading_alpha.pdf"
	defer os.Remove(testFile)

	err = report.ExportPDF(testFile)
	if err != nil {
		t.Fatalf("PDF export failed: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Exported file does not exist")
	}
}
