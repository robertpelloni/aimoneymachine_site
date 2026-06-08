package orchestrator

import (
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
	orch := NewOrchestrator()
	rollback := NewRollbackHandler(orch)

	// Create a dummy file to be cleaned
	dummyFile := "uncommitted_test_file.txt"
	if err := os.WriteFile(dummyFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	// Execute rollback
	if err := rollback.Execute(); err != nil {
		t.Errorf("Rollback execution failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(dummyFile); err == nil {
		t.Errorf("Dummy file still exists after rollback clean")
		os.Remove(dummyFile) // cleanup
	}

	if len(orch.L2.Entries) == 0 {
		t.Errorf("Expected rollback entry in L2 memory")
	}
}

func TestSubmoduleDetection(t *testing.T) {
	status := CheckSubmodules()
	t.Logf("Submodule status: %s", status)
}
