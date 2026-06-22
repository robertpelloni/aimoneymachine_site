package orchestrator

import (
	"fmt"
	"os/exec"
	"time"
)

type RollbackHandler struct {
	Orchestrator *Orchestrator
}

func NewRollbackHandler(o *Orchestrator) *RollbackHandler {
	return &RollbackHandler{Orchestrator: o}
}

func (r *RollbackHandler) Execute() error {
	fmt.Println("[Rollback] Executing automated rollback to previous stable state...")

	// 1. Revert tracked files
	fmt.Println("[Rollback] Reverting modified files via git checkout...")
	cmdCheckout := exec.Command("git", "checkout", ".")
	if err := cmdCheckout.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %v", err)
	}

	// 2. Remove untracked files/directories
	fmt.Println("[Rollback] Cleaning untracked files via git clean...")
	cmdClean := exec.Command("git", "clean", "-fd")
	if err := cmdClean.Run(); err != nil {
		return fmt.Errorf("git clean failed: %v", err)
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
