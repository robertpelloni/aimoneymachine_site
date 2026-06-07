package orchestrator

import (
	"os"
	"testing"
)

func TestGStackNativeSkills(t *testing.T) {
	// The gstack submodule has been consolidated into native workflows.
	// We verify that the skills exist in the .agent/workflows directory.
	skills := []string{
		"office-hours.md",
		"plan-ceo-review.md",
		"plan-eng-review.md",
		"ship.md",
	}

	for _, skill := range skills {
		path := "../.agent/workflows/" + skill
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Native skill %s not found in .agent/workflows/", skill)
		}
	}
}
