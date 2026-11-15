package saas

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type SaaSProject struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TechStack   string `json:"tech_stack"`
	Monetization string `json:"monetization"`
}

type SaaSModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewSaaSModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *SaaSModule {
	return &SaaSModule{
		Orch:   orch,
		Broker: broker,
	}
}

// IdeateProduct finds a gap in the micro-SaaS market using research
func (s *SaaSModule) IdeateProduct(niche string) (*SaaSProject, error) {
	fmt.Printf("[SaaS] Ideating micro-product for niche: %s\n", niche)

	prompt := fmt.Sprintf(`Act as a successful SaaS founder. Analyze the niche: "%s" and identify a high-potential Micro-SaaS product gap.

Requirements:
- Solves a specific pain point
- Can be built as a single-page application or small tool
- Clear monetization strategy (Stripe/Affiliate/Freemium)

Respond with a JSON object:
{
  "name": "Product Name",
  "description": "What it does and who it's for",
  "tech_stack": "e.g. Next.js, Tailwind, Supabase",
  "monetization": "Subscription/One-time"
}

Respond ONLY with valid JSON.`, niche)

	var project SaaSProject
	if err := s.Orch.LLM.GenerateJSON(prompt, &project); err != nil {
		return nil, err
	}

	// Store in memory
	s.Orch.L2.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("saas-idea-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("SaaS Idea: %s (%s)", project.Name, project.Description),
		BaseScore: 85.0,
		Timestamp: time.Now(),
		Tags:      []string{"saas", "ideation", niche},
	})

	return &project, nil
}

// GenerateMVP creates a full-stack codebase for the product
func (s *SaaSModule) GenerateMVP(project SaaSProject) (string, error) {
	fmt.Printf("[SaaS] Generating MVP codebase for: %s\n", project.Name)

	prompt := fmt.Sprintf(`Generate a production-ready, single-file MVP codebase (HTML/Tailwind/JavaScript) for: "%s".

Description: %s
Tech Stack: %s

Requirements:
- Modern, professional UI
- Functional core logic (e.g. calculation, generation)
- Stripe checkout placeholder
- SEO-optimized landing page structure within the same file

Format: Complete HTML/JS code.`, project.Name, project.Description, project.TechStack)

	return s.Orch.LLM.Generate(prompt)
}
