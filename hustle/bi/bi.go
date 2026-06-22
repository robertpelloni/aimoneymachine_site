package bi

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type BIModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewBIModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *BIModule {
	return &BIModule{
		Orch:   orch,
		Broker: broker,
	}
}

// GenerateInsights analyzes memory patterns to identify hidden business opportunities
func (b *BIModule) GenerateInsights() (string, error) {
	fmt.Printf("[BI] Analyzing memory patterns for strategic insights\n")

	// Collect high-value L2 entries
	discoveries := b.Orch.L2.Entries
	var history string
	for i, d := range discoveries {
		history += fmt.Sprintf("- %s (%v)\n", d.Content, d.Tags)
		if i > 20 { break }
	}

	prompt := fmt.Sprintf(`Act as a business intelligence analyst. Analyze the following discovery history from an autonomous agent swarm and identify 3 "hidden gems" or cross-domain opportunities.

HISTORY:
%s

Requirements:
- Identify correlation between niches
- Propose a "High-Confluence" business strategy combining 2+ modules
- Estimate potential impact on net profit

Format: Professional Markdown Report.`, history)

	insights, err := b.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	b.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("bi-insight-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("BI Strategic Insight: %s", insights[:100]+"..."),
		Timestamp: time.Now(),
		Tags:      []string{"bi", "strategy", "intelligence"},
	})

	return insights, nil
}
