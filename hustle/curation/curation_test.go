package curation

import (
	"github.com/robertpelloni/hustle/orchestrator"
	"testing"
)

func TestCurationPipeline(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	curator := &CurationModule{Orchestrator: orch}

	err := curator.Curate("Technology")
	if err != nil {
		t.Fatalf("Curation failed: %v", err)
	}

	if len(orch.L1.Entries) == 0 {
		t.Error("Curation summary not saved to L1 memory")
	}
}
