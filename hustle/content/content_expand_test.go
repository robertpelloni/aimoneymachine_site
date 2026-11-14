package content

import (
	"strings"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

// A terse first draft (the 98-word prod blog) must trigger exactly one
// expansion pass and keep the longer result; threads never expand.
func TestGenerateExpandsShortDraftOnce(t *testing.T) {
	calls := 0
	orch := orchestrator.NewOrchestrator()
	orch.LLM = &orchestrator.MockLLM{GenerateFunc: func(prompt string) (string, error) {
		calls++
		if strings.Contains(prompt, "Expand the following") {
			return "# Expanded Title\n" + strings.Repeat("deep useful content ", 40), nil
		}
		return "# Draft Title\nshort draft body", nil
	}}
	m := NewContentModule(orch, t.TempDir())
	res, err := m.Generate(ContentRequest{Topic: "AI", Type: BlogPost, TargetWords: 100})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if calls != 2 {
		t.Errorf("LLM calls = %d, want 2 (draft + one expansion)", calls)
	}
	if !strings.Contains(res.Body, "Expanded Title") {
		t.Errorf("expanded body not used: %.60q", res.Body)
	}
	if res.Title != "Expanded Title" {
		t.Errorf("title = %q, want from expanded body", res.Title)
	}

	// Threads are supposed to be short: no expansion call.
	calls = 0
	res2, err := m.Generate(ContentRequest{Topic: "AI", Type: SocialThread, TargetWords: 100})
	if err != nil {
		t.Fatalf("thread Generate: %v", err)
	}
	if calls != 1 {
		t.Errorf("thread LLM calls = %d, want 1 (no expansion)", calls)
	}
	if !strings.Contains(res2.Body, "Draft Title") {
		t.Errorf("thread body changed: %.60q", res2.Body)
	}
}
