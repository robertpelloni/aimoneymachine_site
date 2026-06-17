package design

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type DesignModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewDesignModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *DesignModule {
	return &DesignModule{
		Orch:   orch,
		Broker: broker,
	}
}

// IdeateBrand generates visual concepts and style guides for a business
func (d *DesignModule) IdeateBrand(businessName, niche string) (string, error) {
	fmt.Printf("[Design] Ideating brand for: %s (%s)\n", businessName, niche)

	prompt := fmt.Sprintf(`Act as a professional brand designer. Generate a comprehensive visual brand strategy for the business: "%s" in the niche: "%s".

Requirements:
- Logo concepts (3 variants with visual descriptions)
- Primary and secondary color palette (Hex codes)
- Typography recommendations
- Brand "vibe" and personality traits
- Visual direction for social media and website

Format: Professional Markdown.`, businessName, niche)

	strategy, err := d.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	d.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("brand-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Brand Strategy Generated for: %s", businessName),
		Timestamp: time.Now(),
		Tags:      []string{"design", "branding", businessName},
	})

	return strategy, nil
}
