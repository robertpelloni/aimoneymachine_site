package automation

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type AutomationModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewAutomationModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *AutomationModule {
	return &AutomationModule{
		Orch:   orch,
		Broker: broker,
	}
}

// BuildWorkflow generates an n8n or Zapier workflow configuration
func (a *AutomationModule) BuildWorkflow(businessProblem string) (string, error) {
	fmt.Printf("[Automation] Building workflow for: %s\n", businessProblem)

	prompt := fmt.Sprintf(`Act as an automation architect. Build a step-by-step n8n or Zapier workflow JSON schema for the following problem: "%s".

Requirements:
- Identify trigger (e.g. Webhook, Schedule, Email)
- List all intermediate action nodes (e.g. HTTP Request, Google Sheets, LLM)
- Define final output
- Provide the configuration in structured JSON format

Respond with ONLY the JSON object.`, businessProblem)

	workflow, err := a.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	a.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("auto-flow-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Automation Workflow Built for: %s", businessProblem),
		Timestamp: time.Now(),
		Tags:      []string{"automation", "workflow", "agency"},
	})

	return workflow, nil
}
