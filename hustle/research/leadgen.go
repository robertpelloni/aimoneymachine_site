package research

import (
	"fmt"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type Lead struct {
	Name        string
	Email       string
	Website     string
	Description string
	Niche       string
}

type LeadGenerator struct {
	Orch *orchestrator.Orchestrator
}

func NewLeadGenerator(orch *orchestrator.Orchestrator) *LeadGenerator {
	return &LeadGenerator{Orch: orch}
}

// FindLeads simulates discovering business leads for a specific niche.
func (lg *LeadGenerator) FindLeads(niche string) ([]Lead, error) {
	fmt.Printf("[LeadGen] Finding high-value leads in niche: %s\n", niche)

	// Use Tavily to source leads
	searcher := NewResearchSearch(Tavily, lg.Orch, nil)
	query := fmt.Sprintf("%s startups companies", niche)
	results, err := searcher.Query(query)
	if err != nil {
		fmt.Printf("[LeadGen] Tavily search failed, falling back to dummy leads: %v\n", err)
	}

	var leads []Lead
	if len(results) > 0 {
		for i, r := range results {
			if i >= 3 {
				break
			}
			// Extract email or create dummy based on domain
			email := fmt.Sprintf("contact@%s.test", strings.ReplaceAll(strings.ToLower(niche), " ", ""))

			leads = append(leads, Lead{
				Name:        r.Title,
				Email:       email,
				Website:     r.URL,
				Description: r.Snippet,
				Niche:       niche,
			})
		}
	} else {
		leads = []Lead{
			{
				Name:        "Acme Plumbing Services",
				Email:       "contact@acmeplumbing.test",
				Website:     "https://acmeplumbing.test",
				Description: "A regional plumbing service that lacks an automated appointment booking system.",
				Niche:       niche,
			},
			{
				Name:        "Downtown Legal Partners",
				Email:       "info@downtownlegal.test",
				Website:     "https://downtownlegal.test",
				Description: "A boutique law firm that manually triages all their intake emails.",
				Niche:       niche,
			},
		}
	}

	for _, l := range leads {
		lg.Orch.L2.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("lead-%s-%d", l.Email, time.Now().UnixNano()),
			Content:   fmt.Sprintf("Found lead %s (%s). Needs: %s", l.Name, l.Email, l.Description),
			Timestamp: time.Now(),
			Tags:      []string{"lead", "prospect", niche},
		})
	}

	return leads, nil
}
