package orchestrator

import (
	"strings"
	"testing"
)

func TestHealerLoop(t *testing.T) {
	orch := NewOrchestrator()
	mockLLM := &MockLLM{Response: "FAILURE"}
	orch.LLM = mockLLM
	healer := NewHealer(orch)

	issue := "Test failure simulation"

	// Test that it retries when LLM says FAILURE
	// We want it to succeed on the second attempt
	// Note: We need a way to change the response BETWEEN Fix and Verify or between iterations.
	// Since Loop calls Fix then Verify in a tight loop, we use a more clever mock.

	count := 0
	orch.LLM = &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			if strings.Contains(prompt, "Analyze if this fix is likely to have resolved the underlying issue") {
				count++
				if count > 1 {
					return "SUCCESS", nil
				}
				return "FAILURE", nil
			}
			return "MOCK_DIAGNOSIS_OR_FIX", nil
		},
	}

	healer.Loop(issue)

	if healer.RetryCount < 2 {
		t.Errorf("Expected at least 2 retries to succeed, got %d", healer.RetryCount)
	}

	if len(orch.L2.Entries) != 1 {
		t.Errorf("Expected resolution to be logged in L2 memory, got %d entries", len(orch.L2.Entries))
	}
}
