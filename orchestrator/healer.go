package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

type Healer struct {
	Orchestrator *Orchestrator
	RetryLimit   int
	RetryCount   int
}

func NewHealer(o *Orchestrator) *Healer {
	return &Healer{
		Orchestrator: o,
		RetryLimit:   3,
		RetryCount:   0,
	}
}

func (h *Healer) Diagnose(issue string) string {
	fmt.Printf("Diagnosing issue: %s\n", issue)

	prompt := fmt.Sprintf("Act as a system architect. Diagnose the following system issue: %s. Provide a concise explanation of the root cause.", issue)
	diagnosis, err := h.Orchestrator.LLM.Generate(prompt)
	if err != nil {
		diagnosis = "Unable to generate AI diagnosis; defaulting to 'Incomplete system state'."
	}

	// Log diagnosis to memory
	h.Orchestrator.L1.Add(MemoryEntry{
		ID:        fmt.Sprintf("diag-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Issue: %s, Diagnosis: %s", issue, diagnosis),
		Timestamp: time.Now(),
		Tags:      []string{"healer", "diagnosis"},
	})
	return diagnosis
}

func (h *Healer) Fix(diagnosis string) bool {
	h.RetryCount++
	fmt.Printf("Applying fix attempt %d/%d based on diagnosis.\n", h.RetryCount, h.RetryLimit)

	prompt := fmt.Sprintf("Based on this diagnosis: %s, generate a step-by-step fix strategy for the system.", diagnosis)
	fixStrategy, err := h.Orchestrator.LLM.Generate(prompt)
	if err != nil {
		fixStrategy = "Apply generic system reset."
	}

	fmt.Printf("[Healer] Strategy: %s\n", fixStrategy)

	// Log fix attempt to memory
	h.Orchestrator.L1.Add(MemoryEntry{
		ID:        fmt.Sprintf("fix-att-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Attempt %d: %s", h.RetryCount, fixStrategy),
		Timestamp: time.Now(),
		Tags:      []string{"healer", "fix_attempt"},
	})

	return true
}

func (h *Healer) Verify(fix string) bool {
	fmt.Printf("Verifying fix: %s\n", fix)
	// Mock verification failure logic
	return h.RetryCount > 1
}

// WealthPreservation analyzes active hustles and terminates underperforming ones
func (h *Healer) WealthPreservation() {
	fmt.Println("[Healer] Running Wealth Preservation ROI Audit...")

	analysis := h.Orchestrator.Ledger.AnalyzeProfitability()
	if strings.Contains(strings.ToLower(analysis), "terminate") || strings.Contains(strings.ToLower(analysis), "underperforming") {
		fmt.Printf("[Healer] ROI Warning Detected: %s\n", analysis)

		// In a real system, we'd find the task in the scheduler and unschedule it.
		// For now, we log the corrective action to memory for the scheduler to see.
		h.Orchestrator.L1.Add(MemoryEntry{
			ID:        fmt.Sprintf("roi-correction-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Wealth Preservation Action: Requesting termination of underperforming hustles. Reason: %s", analysis),
			Timestamp: time.Now(),
			Tags:      []string{"healer", "wealth_preservation", "roi_correction"},
		})
	} else {
		fmt.Println("[Healer] All active hustles meet ROI thresholds.")
	}
}

func (h *Healer) Loop(issue string) {
	// Periodic Wealth Audit
	h.WealthPreservation()

	diagnosis := h.Diagnose(issue)

	for h.RetryCount < h.RetryLimit {
		if h.Fix(diagnosis) {
			if h.Verify("Attempted Fix") {
				fmt.Println("System healed successfully.")
				h.Orchestrator.L2.Add(MemoryEntry{
					ID:        fmt.Sprintf("fix-success-%d", time.Now().Unix()),
					Content:   fmt.Sprintf("Resolved issue: %s", issue),
					Timestamp: time.Now(),
					Tags:      []string{"healer", "resolution"},
				})
				return
			}

			// Consult Council if initial fixes fail
			if h.RetryCount == h.RetryLimit-1 {
				fmt.Println("Fixes failing. Consulting Multi-Agent Council for strategy...")
				council := NewCouncil(h.Orchestrator)
				debate, _ := council.Debate("How to fix persistent issue: " + issue)
				fmt.Printf("Council consensus: %s\n", debate.Consensus)
			}

			fmt.Println("Verification failed, retrying...")
		}
	}

	fmt.Println("Retry limit reached, triggering mandatory rollback.")
	rollback := NewRollbackHandler(h.Orchestrator)
	rollback.Execute()
}
