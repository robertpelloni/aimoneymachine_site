package outreach

import (
	"fmt"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type Module struct {
	Orch *orchestrator.Orchestrator
}

func NewModule(orch *orchestrator.Orchestrator) *Module {
	return &Module{
		Orch: orch,
	}
}

func (m *Module) GenerateTemplate(platform, niche, topic string) (string, error) {
	prompt := fmt.Sprintf(`Generate a highly converting, personalized cold outreach template for the platform: %s.
Target Niche: %s
Value Proposition / Topic: %s

Requirements:
- Keep it under 4 sentences.
- Be professional but highly persuasive.
- Don't be overly salesy.
- Include placeholders like {{Name}} and {{Company}}.`, platform, niche, topic)

	body, err := m.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate %s outreach template: %w", platform, err)
	}

	m.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("outreach-template-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Generated %s template for %s: %s", platform, niche, topic),
		Timestamp: time.Now(),
		Tags:      []string{"outreach", "template", platform},
	})

	return body, nil
}

func (m *Module) ScheduleCadence(targetEmail, platform, template string) error {
	fmt.Printf("[Outreach Module] Scheduling %s cadence for %s\n", platform, targetEmail)
	fmt.Printf("[Outreach Module] Template: \n%s\n", template)

	m.Orch.Ledger.Add(orchestrator.Transaction{
		Amount: 0.05,
		Type:   orchestrator.Expense,
		Hustle: "OutreachCadence",
		Note:   fmt.Sprintf("Scheduled %s cadence for %s", platform, targetEmail),
	})

	return nil
}
