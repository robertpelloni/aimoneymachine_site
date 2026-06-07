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

	t.Run("GenerateTopicIdeas", func(t *testing.T) {
		orch.LLM = &orchestrator.MockLLM{Response: `["Topic 1", "Topic 2"]`}
		topics, err := m.GenerateTopicIdeas("tech", 2)
		if err != nil {
			t.Errorf("GenerateTopicIdeas failed: %v", err)
		}
		if len(topics) != 2 {
			t.Errorf("Expected 2 topics, got %d", len(topics))
		}
	})
}
