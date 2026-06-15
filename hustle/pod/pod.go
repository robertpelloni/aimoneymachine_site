package pod

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type PODModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewPODModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *PODModule {
	return &PODModule{
		Orch:   orch,
		Broker: broker,
	}
}

// PlanDesigns generates a set of design concepts for a niche
func (p *PODModule) PlanDesigns(niche string) (string, error) {
	fmt.Printf("[POD] Planning designs for niche: %s\n", niche)

	prompt := fmt.Sprintf(`Act as a successful Print-on-Demand (POD) designer.
Generate 5 unique, trending design concepts (quotes, slogans, or visuals) for the niche: "%s".

Requirements:
- Target audience specific
- High-converting slogans/visual descriptions
- SEO-friendly titles for Etsy/Redbubble
- List of suggested keywords for each

Format: Professional Markdown list.`, niche)

	plan, err := p.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	p.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("pod-plan-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("POD Design Plan for %s", niche),
		Timestamp: time.Now(),
		Tags:      []string{"pod", "design", niche},
	})

	return plan, nil
}

// OptimizeListing generates an optimized product listing for a POD platform
func (p *PODModule) OptimizeListing(designName, niche string) (string, error) {
	prompt := fmt.Sprintf(`Generate a high-converting POD product listing for the design: "%s" in the niche: "%s".

Requirements:
- Click-worthy Title
- Engaging Description (Benefits + Features)
- 15-20 relevant Tags (comma-separated)
- Optimized for Redbubble/Etsy search algorithms

Format: Professional Markdown.`, designName, niche)

	return p.Orch.LLM.Generate(prompt)
}
