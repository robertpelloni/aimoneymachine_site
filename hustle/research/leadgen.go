package research

import (
	"fmt"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type Lead struct {
	Name        string
	Email       string
	Website     string
	Description string
	Niche       string
	// Source is where the lead came from: "demo-seed" for the built-in
	// examples, "web-search" for real discovery. Demo leads use .test
	// addresses and must never be mailed outside dry-run.
	Source string
}

type LeadGenerator struct {
	Orch *orchestrator.Orchestrator
}

func NewLeadGenerator(orch *orchestrator.Orchestrator) *LeadGenerator {
	return &LeadGenerator{Orch: orch}
}

// FindLeads returns the built-in DEMO seed leads. Despite the historical
// name this discovers nothing: the two entries are examples with .test
// addresses, labeled Source "demo-seed" so no code path mistakes them for
// real prospects. Real discovery is FindLeadsFromSearch.
func (lg *LeadGenerator) FindLeads(niche string) ([]Lead, error) {
	fmt.Printf("[LeadGen] DEMO seed leads for niche %q (not real discovery — set a search source for real leads)\n", niche)

	leads := []Lead{
		{
			Name:        "Acme Plumbing Services",
			Email:       "contact@acmeplumbing.test",
			Website:     "https://acmeplumbing.test",
			Description: "A regional plumbing service that lacks an automated appointment booking system.",
			Niche:       niche,
			Source:      "demo-seed",
		},
		{
			Name:        "Downtown Legal Partners",
			Email:       "info@downtownlegal.test",
			Website:     "https://downtownlegal.test",
			Description: "A boutique law firm that manually triages all their intake emails.",
			Niche:       niche,
			Source:      "demo-seed",
		},
	}

	lg.storeLeads(leads, niche)
	return leads, nil
}

// FindLeadsFromSearch discovers real businesses from web search results.
// It deliberately does NOT invent email addresses: Email stays empty, so
// SendEmail will generate-and-log pitches but refuse delivery until a
// real address is attached by an enrichment step.
func (lg *LeadGenerator) FindLeadsFromSearch(searcher SearchInterface, niche string) ([]Lead, error) {
	if searcher == nil {
		return nil, fmt.Errorf("no searcher configured")
	}
	results, err := searcher.Query(niche + " businesses contact")
	if err != nil {
		return nil, err
	}
	var leads []Lead
	for _, r := range results {
		if r.Title == "" || r.URL == "" {
			continue
		}
		leads = append(leads, Lead{
			Name:        r.Title,
			Website:     r.URL,
			Description: r.Snippet,
			Niche:       niche,
			Source:      "web-search",
		})
		if len(leads) >= 10 {
			break
		}
	}
	fmt.Printf("[LeadGen] Found %d web leads for niche %q\n", len(leads), niche)
	lg.storeLeads(leads, niche)
	return leads, nil
}

func (lg *LeadGenerator) storeLeads(leads []Lead, niche string) {
	for _, l := range leads {
		lg.Orch.L2.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("lead-%s-%d", l.Name, time.Now().UnixNano()),
			Content:   fmt.Sprintf("Found lead %s (%s) [%s]. Needs: %s", l.Name, l.Website, l.Source, l.Description),
			Timestamp: time.Now(),
			Tags:      []string{"lead", "prospect", niche},
		})
	}
}
