package orchestrator

import (
	"testing"
	"time"
)

// After a restart every persisted task is "due" at once. The stagger
// deadline must hold tasks back even when LastRun is ancient, and let
// them through once it passes.
func TestTaskDueBootStagger(t *testing.T) {
	now := time.Now()
	ancient := now.Add(-24 * time.Hour)

	staggered := &Task{Name: "late", Interval: time.Minute, LastRun: ancient, notBefore: now.Add(time.Hour)}
	if taskDue(staggered, now) {
		t.Error("task inside stagger window must not be due")
	}
	if !taskDue(staggered, now.Add(2*time.Hour)) {
		t.Error("task past stagger deadline with old LastRun must be due")
	}

	ready := &Task{Name: "first", Interval: time.Minute, LastRun: ancient, notBefore: now}
	if !taskDue(ready, now) {
		t.Error("first task (zero stagger offset) must be due immediately")
	}

	recent := &Task{Name: "recent", Interval: time.Hour, LastRun: now, notBefore: now.Add(-time.Hour)}
	if taskDue(recent, now.Add(time.Minute)) {
		t.Error("task run a minute ago with 1h interval must not be due")
	}
}
