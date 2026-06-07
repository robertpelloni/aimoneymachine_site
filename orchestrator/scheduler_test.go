package orchestrator

import (
	"os"
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

	err := scheduler.Tasks[0].Execute(orch)
	if err != nil {
		t.Errorf("Task execution failed: %v", err)
	}
	if !taskExecuted {
		t.Error("Task function was not called")
	}
}

func TestSchedulerPersistence(t *testing.T) {
	tmpFile := "test_tasks.json"
	defer os.Remove(tmpFile)

	orch := NewOrchestrator()
	s1 := NewScheduler(orch)

	lastRun := time.Now().Add(-10 * time.Minute)
	s1.Tasks = append(s1.Tasks, &Task{Name: "PTask", LastRun: lastRun})

	err := s1.SaveState(tmpFile)
	if err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	s2 := NewScheduler(orch)
	s2.Register("PTask", 5*time.Minute, func(o *Orchestrator) error { return nil })

	err = s2.LoadState(tmpFile, NewHustleProtocol())
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if !s2.Tasks[0].LastRun.Equal(lastRun) {
		t.Errorf("Persistence failed. Expected %v, got %v", lastRun, s2.Tasks[0].LastRun)
	}
}
