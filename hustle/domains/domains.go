package domains

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type DomainModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewDomainModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *DomainModule {
	return &DomainModule{
		Orch:   orch,
		Broker: broker,
	}
}

// EvaluateDomain estimates the resale value of a domain name
func (d *DomainModule) EvaluateDomain(domain string) (float64, string, error) {
	fmt.Printf("[Domains] Evaluating domain: %s\n", domain)

	prompt := fmt.Sprintf(`Act as a domain valuation expert. Evaluate the potential resale value of the domain: "%s".

Requirements:
- Estimated value range (USD)
- Keyword demand analysis
- TLD strength (.com vs others)
- Potential buyer niches

Respond with a JSON object:
{
  "estimated_value": 500.0,
  "analysis": "Professional valuation summary"
}

Respond ONLY with valid JSON.`, domain)

	var result struct {
		Value    float64 `json:"estimated_value"`
		Analysis string  `json:"analysis"`
	}

	if err := d.Orch.LLM.GenerateJSON(prompt, &result); err != nil {
		return 0, "Error during valuation", err
	}

	// Store in memory
	d.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("domain-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Domain Valuation: %s ($%.2f) - %s", domain, result.Value, result.Analysis),
		Timestamp: time.Now(),
		Tags:      []string{"domains", "flipping", domain},
	})

	return result.Value, result.Analysis, nil
}

// AuctionBid generates a bidding strategy for a domain auction
func (d *DomainModule) AuctionBid(domain string, currentBid float64) (string, error) {
	fmt.Printf("[Domains] Analyzing auction for: %s\n", domain)

	prompt := fmt.Sprintf(`Act as a domain auction strategist. Analyze the auction for domain: "%s" with a current bid of $%.2f.

Requirements:
- Recommend next bid amount
- Set a maximum walk-away price
- Propose a psychological bidding strategy (e.g. sniping vs early dominance)

Format: Professional Markdown.`, domain, currentBid)

	return d.Orch.LLM.Generate(prompt)
}
