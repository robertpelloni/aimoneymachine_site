package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomicRoundtripAndNoLeftovers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	if err := WriteFileAtomic(path, []byte(`{"v":1}`), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != `{"v":1}` {
		t.Fatalf("read = %q, %v", got, err)
	}
	// Overwrite works and leaves no temp debris.
	if err := WriteFileAtomic(path, []byte(`{"v":2}`), 0644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	got, _ = os.ReadFile(path)
	if string(got) != `{"v":2}` {
		t.Fatalf("after rewrite = %q", got)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".aimm-tmp-") {
			t.Fatalf("temp leftover: %s", e.Name())
		}
	}
	if len(entries) != 1 {
		t.Fatalf("dir entries = %d, want 1", len(entries))
	}
}

// A leftover temp from a crashed write must not affect reading the real
// file: the last completed rename is the state.
func TestWriteFileAtomicCrashLeftoverIgnored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := WriteFileAtomic(path, []byte("good"), 0644); err != nil {
		t.Fatal(err)
	}
	// Simulate a crash mid-write: partial temp file abandoned.
	if err := os.WriteFile(filepath.Join(dir, ".aimm-tmp-999"), []byte("part"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "good" {
		t.Fatalf("state = %q, %v; want good", got, err)
	}
}

// Scheduler + chain state survive a save/load roundtrip through the
// atomic path (this is the deploy-restart scenario).
func TestSchedulerChainAtomicRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	s := NewScheduler(NewOrchestrator())
	s.Register("Research", 1000000000, nil)
	if err := s.SaveState(filepath.Join(dir, "tasks.json")); err != nil {
		t.Fatalf("scheduler save: %v", err)
	}
	s2 := NewScheduler(NewOrchestrator())
	if err := s2.LoadState(filepath.Join(dir, "tasks.json"), nil); err != nil {
		t.Fatalf("scheduler load: %v", err)
	}

	m := NewChainManager(NewOrchestrator(), NewHustleProtocol())
	m.Chains["curation"] = &Chain{Name: "curation", Steps: []string{"hustle://curation"}}
	if err := m.SaveState(filepath.Join(dir, "chains.json")); err != nil {
		t.Fatalf("chain save: %v", err)
	}
	m2 := NewChainManager(NewOrchestrator(), NewHustleProtocol())
	if err := m2.LoadState(filepath.Join(dir, "chains.json")); err != nil {
		t.Fatalf("chain load: %v", err)
	}
	if _, ok := m2.Chains["curation"]; !ok {
		t.Fatal("chain lost in roundtrip")
	}
}
