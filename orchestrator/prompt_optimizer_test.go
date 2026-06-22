package orchestrator

import (
	"testing"
)

func TestPromptOptimizer(t *testing.T) {
	po := NewPromptOptimizer()

	po.AddVariation("test", "template 1")
	po.AddVariation("test", "template 2")

	// Get a prompt
	template, idx, err := po.GetPrompt("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if idx != 0 {
		t.Fatalf("expected index 0 for first unused prompt, got %v", idx)
	}

	if template != "template 1" {
		t.Fatalf("expected template 1, got %v", template)
	}

	po.RecordWin("test", idx)

	if po.Variations["test"][0].Wins != 1 {
		t.Fatalf("expected 1 win for template 1, got %v", po.Variations["test"][0].Wins)
	}
}
