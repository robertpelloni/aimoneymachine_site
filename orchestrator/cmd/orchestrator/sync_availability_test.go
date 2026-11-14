package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Prod (no sync.sh in the working directory) must get an honest
// "unavailable" from both the probe and runSyncProtocol, and the probe
// must turn positive when the script exists (SYNC_SCRIPT override).
func TestSyncAvailability(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("SYNC_SCRIPT", "")
	if syncAvailable() {
		t.Fatal("syncAvailable = true with no sync.sh present")
	}
	err := runSyncProtocol()
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("runSyncProtocol err = %v, want unavailable", err)
	}

	script := filepath.Join(t.TempDir(), "sync.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\ntrue\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SYNC_SCRIPT", script)
	if !syncAvailable() {
		t.Fatal("syncAvailable = false with SYNC_SCRIPT pointing at a real file")
	}
}
