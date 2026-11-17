package content

import (
	"github.com/robertpelloni/hustle/orchestrator"
	"testing"
)

func TestContentModule(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	orch.LLM = &orchestrator.MockLLM{Response: "# Mock Title\nMock Content"}

	m := NewContentModule(orch, "test_output")

	t.Run("Generate", func(t *testing.T) {
		req := ContentRequest{
			Topic: "AI Future",
			Type:  BlogPost,
		}
		result, err := m.Generate(req)
		if err != nil {
			t.Errorf("Generate failed: %v", err)
		}
		if result == nil {
			t.Fatal("Result is nil")
		}
		if result.Title != "Mock Title" {
			t.Errorf("Expected title 'Mock Title', got '%s'", result.Title)
		}
	})

	// Long-form output with bracketed fill-ins ("[Insert your ...]")
	// must be rejected, not saved/published: the social pipeline shipped
	// exactly this class of text to LinkedIn (2026-11-17). Callers
	// already treat a Generate error honestly (calendar shells stay
	// pending, chain steps fail visibly).
	t.Run("GenerateRejectsPlaceholders", func(t *testing.T) {
		orch.LLM = &orchestrator.MockLLM{Response: "# Stealth Wealth\n\nBuild quietly. [Insert your success story here] and [insert relevant statistic here] to inspire readers."}
		result, err := m.Generate(ContentRequest{Topic: "Stealth Wealth", Type: BlogPost})
		if err == nil {
			t.Fatalf("Generate with placeholders: err = nil, result = %+v; want rejection", result)
		}
		if result != nil {
			t.Errorf("result = %+v, want nil on placeholder rejection", result)
		}
	})

	t.Run("GenerateTopicIdeas", func(t *testing.T) {
		orch.LLM = &orchestrator.MockLLM{Response: `["Topic 1", "Topic 2"]`}
		topics, err := m.GenerateTopicIdeas("tech", 2)
		if err != nil {
			t.Errorf("GenerateTopicIdeas failed: %v", err)
		}
		if len(topics) < 2 {
			t.Errorf("Expected 2 topics, got %d", len(topics))
		}
	})
}
