package research

import (
	"testing"
)

func TestResearchPipelineE2E(t *testing.T) {
	// 1. Mock Search
	searcher := &MockSearch{}
	results, err := searcher.Query("How to build a money machine")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Expected at least one search result")
	}

	// 2. Synthesize Report
	report := &Report{Title: "E2E Test Report"}
	report.Synthesize(results)
	if report.Content == "" {
		t.Fatal("Report synthesis produced empty content")
	}

	// 3. Export PDF
	err = report.ExportPDF("test_report.pdf")
	if err != nil {
		t.Fatalf("PDF export failed: %v", err)
	}
}
