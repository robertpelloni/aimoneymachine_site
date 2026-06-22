package qa

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"os/exec"
	"time"
)

type QAModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewQAModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *QAModule {
	return &QAModule{
		Orch:   orch,
		Broker: broker,
	}
}

// RunTests autonomously executes module tests and reports outcomes
func (q *QAModule) RunTests(modulePath string) (string, error) {
	fmt.Printf("[QA] Running tests for module: %s\n", modulePath)

	cmd := exec.Command("go", "test", modulePath)
	output, err := cmd.CombinedOutput()

	status := "success"
	if err != nil {
		status = "failure"
	}

	report := string(output)

	// Store in memory
	q.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("qa-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("QA Test Run for %s: %s", modulePath, status),
		Timestamp: time.Now(),
		Tags:      []string{"qa", "test", modulePath, status},
	})

	return report, nil
}

// VerifyStability performs a system-wide health check
func (q *QAModule) VerifyStability() error {
	modules := []string{"./orchestrator/...", "./hustle/content/...", "./hustle/trading/..."}

	for _, m := range modules {
		res, err := q.RunTests(m)
		if err != nil {
			fmt.Printf("[QA] ❌ Stability check failed for %s:\n%s\n", m, res)
			// Trigger healer via A2A
			if q.Broker != nil {
				q.Broker.Publish(orchestrator.Message{
					ID:        fmt.Sprintf("alert-%d", time.Now().Unix()),
					Source:    "qa-module",
					Type:      orchestrator.Command,
					Topic:     "swarm_fix",
					Payload:   fmt.Sprintf("Module %s failed stability check: %v", m, err),
					Timestamp: time.Now(),
				})
			}
		} else {
			fmt.Printf("[QA] ✅ Module %s is stable.\n", m)
		}
	}
	return nil
}
