package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Persisted DiscoveredChain tasks whose chain no longer exists must be
// dropped at load, not re-registered to fail on every cycle. Known chains
// (and the no-hook legacy path) still load.
func TestLoadStateDropsOrphanedDiscoveredChains(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	state := `[
	  {"name": "DiscoveredChain:gone", "interval": 60000000000, "last_run": "2026-01-01T00:00:00Z"},
	  {"name": "DiscoveredChain:kept", "interval": 60000000000, "last_run": "2026-01-01T00:00:00Z"}
	]`
	path := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(path, []byte(state), 0644); err != nil {
		t.Fatal(err)
	}

	s := NewScheduler(NewOrchestrator())
	s.KnownChain = func(name string) bool { return name == "kept" }
	if err := s.LoadState(path, nil); err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(s.Tasks) != 1 || s.Tasks[0].Name != "DiscoveredChain:kept" {
		names := []string{}
		for _, task := range s.Tasks {
			names = append(names, task.Name)
		}
		t.Fatalf("tasks = %v, want only DiscoveredChain:kept", names)
	}
	if s.Tasks[0].Interval != time.Minute {
		t.Fatalf("interval = %v, want 1m", s.Tasks[0].Interval)
	}
}
