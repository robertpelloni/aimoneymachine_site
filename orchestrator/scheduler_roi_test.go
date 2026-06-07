package orchestrator

import (
	"testing"
	"time"
)

func TestSchedulerWealthPreservationWiring(t *testing.T) {
	orch := NewOrchestrator()
	scheduler := NewScheduler(orch)

	scheduler.Register("BadHustle", 1*time.Hour, func(o *Orchestrator) error {
		return nil
	})

	if len(scheduler.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(scheduler.Tasks))
	}

	// Simulate Wealth Preservation Action in memory
	orch.L1.Add(MemoryEntry{
		ID: "roi-kill-1",
		Content: "Wealth Preservation Action: Requesting termination of underperforming hustles. Reason: Underperforming hustle detected: BadHustle",
		Timestamp: time.Now(),
		Tags: []string{"healer", "wealth_preservation"},
	})

	// Trigger the check
	scheduler.checkROICorrections()

	if len(scheduler.Tasks) != 0 {
		t.Errorf("Expected 0 tasks after Wealth Preservation, but %s still exists", scheduler.Tasks[0].Name)
	}
}
