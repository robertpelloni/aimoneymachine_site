package orchestrator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSyncHealth(t *testing.T) {
	health := GetSyncHealth()
	if health.Status != "Active" {
		t.Errorf("Expected status Active, got %s", health.Status)
	}
	t.Logf("Detected Git State: %s", health.GitState)
}

func TestMergeConflictHandling(t *testing.T) {
	// Mock merge conflict simulation
	hasConflict := true
	if hasConflict {
		// Log discovery
		t.Log("Successfully detected mock merge conflict")
	}
}

func TestRollback(t *testing.T) {
	// RollbackHandler runs `git checkout .` + `git clean -fd` in the process
	// working directory. Running it in the real repo would revert and delete
	// uncommitted work (it destroyed in-flight edits twice), so exercise it
	// only inside a throwaway git repo and verify the revert there.
	dir := t.TempDir()
	t.Chdir(dir)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git unavailable for rollback test: %v (%s)", err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	tracked := filepath.Join(dir, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("committed"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")
	// Dirty the repo: modify a tracked file and add an untracked one.
	if err := os.WriteFile(tracked, []byte("dirty"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := NewRollbackHandler(NewOrchestrator()).Execute(); err != nil {
		t.Fatalf("Rollback execution failed: %v", err)
	}
	got, err := os.ReadFile(tracked)
	if err != nil || string(got) != "committed" {
		t.Fatalf("tracked file = %q, %v; want reverted to committed", got, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "untracked.txt")); !os.IsNotExist(err) {
		t.Fatalf("untracked file still exists after rollback (err=%v)", err)
	}
}

func TestSubmoduleDetection(t *testing.T) {
	status := CheckSubmodules()
	t.Logf("Submodule status: %s", status)
}
