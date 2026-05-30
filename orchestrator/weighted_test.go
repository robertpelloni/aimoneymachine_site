package orchestrator

import (
	"testing"
	"time"
)

func TestWeightedCouncil(t *testing.T) {
	orch := NewOrchestrator()
	council := NewCouncil(orch)

	result, _ := council.Debate("Test Weighted Voting")

	if result.WeightedScore <= 0 {
		t.Errorf("Expected positive weighted score, got %v", result.WeightedScore)
	}
}

func TestRankedSearch(t *testing.T) {
	orch := NewOrchestrator()

	// Add fresh entry
	fresh := MemoryEntry{ID: "fresh", Content: "Keyword match", BaseScore: 100, Timestamp: time.Now()}
	orch.L1.Add(fresh)

	// Add old entry
	old := MemoryEntry{ID: "old", Content: "Keyword match", BaseScore: 100, Timestamp: time.Now().Add(-100 * time.Hour)}
	orch.L1.Add(old)

	results := orch.L1.RankedSearch("Keyword")

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if results[0].ID != "fresh" {
		t.Errorf("RankedSearch failed. Expected 'fresh' to be first, got %s", results[0].ID)
	}
}
