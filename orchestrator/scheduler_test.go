package orchestrator

import (
	"testing"
	"time"
)

func TestSchedulerTask(t *testing.T) {
	orch := NewOrchestrator()
	scheduler := NewScheduler(orch)

	taskExecuted := false
	scheduler.Register("TestTask", 100*time.Millisecond, func(o *Orchestrator) error {
		taskExecuted = true
		return nil
	})

	if len(scheduler.Tasks) != 1 {
		t.Errorf("Expected 1 registered task, got %d", len(scheduler.Tasks))
	}

	// Manual execution check
	err := scheduler.Tasks[0].Execute(orch)
	if err != nil {
		t.Errorf("Task execution failed: %v", err)
	}
	if !taskExecuted {
		t.Error("Task function was not called")
	}
}
