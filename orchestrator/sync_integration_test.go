package orchestrator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSyncProtocolIntegration simulates a multi-branch environment and verifies sync.sh logic.
func TestSyncProtocolIntegration(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	// Create a temporary directory for the test repository
	tempDir, err := os.MkdirTemp("", "sync-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	runCmd := func(dir string, name string, args ...string) (string, error) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	// 1. Initialize "remote" repository
	remoteDir := filepath.Join(tempDir, "remote")
	if err := os.Mkdir(remoteDir, 0755); err != nil {
		t.Fatalf("failed to create remote dir: %v", err)
	}
	runCmd(remoteDir, "git", "init", "--bare")

	// 2. Initialize "local" repository
	localDir := filepath.Join(tempDir, "local")
	runCmd(tempDir, "git", "clone", remoteDir, "local")

	// Setup user config for commits
	runCmd(localDir, "git", "config", "user.email", "test@example.com")
	runCmd(localDir, "git", "config", "user.name", "Test User")

	// 3. Create initial state on main
	os.WriteFile(filepath.Join(localDir, "VERSION.md"), []byte("1.0.0"), 0644)
	os.WriteFile(filepath.Join(localDir, "README.md"), []byte("Initial content"), 0644)
	// Copy sync.sh to localDir
	syncPath, _ := filepath.Abs("../sync.sh")
	syncContent, _ := os.ReadFile(syncPath)
	os.WriteFile(filepath.Join(localDir, "sync.sh"), syncContent, 0755)

	runCmd(localDir, "git", "add", ".")
	runCmd(localDir, "git", "commit", "-m", "Initial commit")
	runCmd(localDir, "git", "push", "origin", "main")

	// 4. Create a feature branch
	runCmd(localDir, "git", "checkout", "-b", "feat/test-branch")
	os.WriteFile(filepath.Join(localDir, "feature.txt"), []byte("Feature content"), 0644)
	runCmd(localDir, "git", "add", ".")
	runCmd(localDir, "git", "commit", "-m", "Add feature")

	// 5. Update main (simulate upstream change)
	runCmd(localDir, "git", "checkout", "main")
	os.WriteFile(filepath.Join(localDir, "README.md"), []byte("Updated content on main"), 0644)
	runCmd(localDir, "git", "add", ".")
	runCmd(localDir, "git", "commit", "-m", "Update main")
	runCmd(localDir, "git", "push", "origin", "main")

	// Ensure remote tracking branch is updated and LOCAL main exists
	runCmd(localDir, "git", "fetch", "origin", "main:main")

	// 6. Go back to feature branch and make it "ready" for forward merge
	runCmd(localDir, "git", "checkout", "feat/test-branch")
	runCmd(localDir, "git", "checkout", "-b", "feat/ready-test")

	// 7. Add uncommitted changes to test stashing
	os.WriteFile(filepath.Join(localDir, "uncommitted.txt"), []byte("Stash me"), 0644)

	// 8. Run sync.sh
	t.Log("Running sync.sh in test environment...")

	out, err := runCmd(localDir, "./sync.sh")
	if err != nil {
		t.Errorf("sync.sh failed: %v\nOutput: %s", err, out)
	}
	t.Logf("sync.sh output: %s", out)

	// 9. Verify Reverse Merge (feat/ready-test should have main's updates)
	runCmd(localDir, "git", "checkout", "feat/ready-test")
	readmeContent, _ := os.ReadFile(filepath.Join(localDir, "README.md"))
	if !strings.Contains(string(readmeContent), "Updated content on main") {
		t.Errorf("Reverse merge failed: feat/ready-test does not contain main's updates. Content: %s", string(readmeContent))
	}

	// 10. Verify Forward Merge (main should have feat/ready-test's content)
	runCmd(localDir, "git", "checkout", "main")
	if _, err := os.Stat(filepath.Join(localDir, "feature.txt")); os.IsNotExist(err) {
		t.Errorf("Forward merge failed: main does not contain feature.txt from feat/ready-test")
	}

	// 11. Verify Stash Restore (uncommitted.txt should still exist)
	runCmd(localDir, "git", "checkout", "feat/ready-test")
	if _, err := os.Stat(filepath.Join(localDir, "uncommitted.txt")); os.IsNotExist(err) {
		t.Errorf("Stash restore failed: uncommitted.txt is missing")
	}
}
