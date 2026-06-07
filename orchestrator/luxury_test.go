package orchestrator

import (
	"testing"
	"strings"
)

func TestWealthPreservation(t *testing.T) {
	orch := NewOrchestrator()
	healer := NewHealer(orch)

	// Case 1: Healthy ROI
	orch.Ledger.Add(Transaction{Amount: 100, Type: Revenue, Hustle: "Trading"})
	healer.WealthPreservation()
	if len(orch.L1.Entries) != 0 {
		t.Errorf("Expected 0 corrections for healthy ROI, got %d", len(orch.L1.Entries))
	}

	// Case 2: Underperforming ROI
	orch.Ledger.Add(Transaction{Amount: 500, Type: Expense, Hustle: "BadBot"})

	healer.WealthPreservation()
	if len(orch.L1.Entries) != 1 {
		t.Errorf("Expected 1 correction for underperforming ROI, got %d", len(orch.L1.Entries))
	}

	entry := orch.L1.Entries[0]
	if !strings.Contains(entry.Content, "Terminate immediately") {
		t.Errorf("Expected termination message, got: %s", entry.Content)
	}

	found := false
	for _, tag := range entry.Tags {
		if tag == "wealth_preservation" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected tag 'wealth_preservation', got %v", entry.Tags)
	}
}
