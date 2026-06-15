package publishing

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type KDPModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewKDPModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *KDPModule {
	return &KDPModule{
		Orch:   orch,
		Broker: broker,
	}
}

// PlanInterior generates a design plan for a low-content book interior
func (k *KDPModule) PlanInterior(niche string) (string, error) {
	fmt.Printf("[KDP] Planning interior for niche: %s\n", niche)

	prompt := fmt.Sprintf(`Act as a successful KDP publisher.
Generate a detailed interior layout plan for a low-content book in the niche: "%s".

Requirements:
- Page types (e.g. prompt pages, checklist, habit tracker)
- Layout descriptions
- Font and style recommendations
- Unique selling point for this interior

Format: Professional Markdown layout.`, niche)

	plan, err := k.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	k.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("kdp-interior-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("KDP Interior Plan for %s", niche),
		Timestamp: time.Now(),
		Tags:      []string{"kdp", "publishing", niche},
	})

	return plan, nil
}

// GenerateKeywords provides high-converting Amazon keywords for a book
func (k *KDPModule) GenerateKeywords(title, niche string) (string, error) {
	prompt := fmt.Sprintf(`Generate 7 high-converting backend keyword phrases for an Amazon KDP book titled: "%s" in the niche: "%s".

Format: Comma-separated phrases.`, title, niche)

	return k.Orch.LLM.Generate(prompt)
}
