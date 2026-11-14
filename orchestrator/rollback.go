package orchestrator

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RollbackHandler struct {
	Orchestrator *Orchestrator
	// Dir is the explicit git repository root to revert. It is required:
	// a zero-value handler refuses to run, so no caller can accidentally
	// `git checkout .` + `git clean -fd` whatever directory the process
	// happens to sit in (this destroyed uncommitted work twice via tests).
	Dir string
}

func NewRollbackHandler(o *Orchestrator) *RollbackHandler {
	return &RollbackHandler{Orchestrator: o}
}

// NewRollbackHandlerInDir targets an explicit repository root.
func NewRollbackHandlerInDir(o *Orchestrator, dir string) *RollbackHandler {
	return &RollbackHandler{Orchestrator: o, Dir: dir}
}

func (r *RollbackHandler) Execute() error {
	if r.Dir == "" {
		return fmt.Errorf("rollback refused: no explicit Dir set (use NewRollbackHandlerInDir)")
	}
	abs, err := filepath.Abs(r.Dir)
	if err != nil {
		return fmt.Errorf("rollback refused: bad Dir %q: %w", r.Dir, err)
	}
	// Only ever operate on a git top-level that is exactly Dir.
	top, err := exec.Command("git", "-C", abs, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("rollback refused: %s is not a git repository: %w", abs, err)
	}
	if filepath.Clean(strings.TrimSpace(string(top))) != filepath.Clean(abs) {
		return fmt.Errorf("rollback refused: %s is not the top level of its git repository", abs)
	}

	fmt.Printf("[Rollback] Executing automated rollback in %s...\n", abs)

	run := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = abs
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git %s failed: %w (%s)", strings.Join(args, " "), err, out)
		}
		return nil
	}
	// 1. Revert tracked files. 2. Remove untracked files/directories.
	fmt.Println("[Rollback] Reverting modified files via git checkout...")
	if err := run("checkout", "--", "."); err != nil {
		return err
	}
	fmt.Println("[Rollback] Cleaning untracked files via git clean...")
	if err := run("clean", "-fd"); err != nil {
		return err
	}
	fmt.Println("[Rollback] System successfully reverted to last committed state.")

	// Log event to memory
	r.Orchestrator.L2.Add(MemoryEntry{
		ID:        fmt.Sprintf("rollback-%d", time.Now().Unix()),
		Content:   "Automated Rollback Triggered: System reverted to last stable Git state.",
		Timestamp: time.Now(),
		Tags:      []string{"rollback", "recovery", "git"},
	})

	return nil
}
