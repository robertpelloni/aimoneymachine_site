package orchestrator

import (
	"testing"
)

func TestCouncilDebate(t *testing.T) {
	orch := NewOrchestrator()
	council := NewCouncil(orch)

	topic := "Invest all capital into dogecoin"
	result, err := council.Debate(topic)
	if err != nil {
		t.Fatalf("Debate failed: %v", err)
	}

	if result.Consensus == "" {
		t.Error("Expected non-empty consensus")
	}

	if len(result.Points) != 2 {
		t.Errorf("Expected 2 points (Bull and Bear), got %d", len(result.Points))
	}
}
