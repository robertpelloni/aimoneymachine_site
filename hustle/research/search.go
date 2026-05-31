package research

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"os"
	"strings"
	"time"
)

type SearchResult struct {
	URL      string
	Snippet  string
	Title    string
	Provider string
}

type SearchInterface interface {
	Query(q string) ([]SearchResult, error)
}

type Provider string

const (
	Tavily Provider = "Tavily"
	Brave  Provider = "Brave"
	Google Provider = "Google"
)

type ResearchSearch struct {
	ActiveProvider Provider
	Orchestrator   *orchestrator.Orchestrator
	Broker         *orchestrator.A2ABroker
	APIKey         string
}

func NewResearchSearch(p Provider, orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *ResearchSearch {
	key := ""
	if p == Tavily {
		key = os.Getenv("TAVILY_API_KEY")
	}
	return &ResearchSearch{
		ActiveProvider: p,
		Orchestrator:   orch,
		Broker:         broker,
		APIKey:         key,
	}
}

func (s *ResearchSearch) Query(q string) ([]SearchResult, error) {
	fmt.Printf("Searching via %s for: %s\n", s.ActiveProvider, q)

	if s.ActiveProvider == Tavily && s.APIKey == "" {
		fmt.Println("Warning: TAVILY_API_KEY not set, using mock data.")
	}

	results := []SearchResult{
		{URL: "https://hustle.com/info", Title: "Hustle Strategy", Snippet: "Key insights for automated revenue with $BTC trends.", Provider: string(s.ActiveProvider)},
	}

	if s.Orchestrator != nil {
		for _, res := range results {
			entry := orchestrator.MemoryEntry{
				ID:        res.URL,
				Content:   fmt.Sprintf("%s: %s", res.Title, res.Snippet),
				BaseScore: 50.0,
				Timestamp: time.Now(),
				Tags:      []string{"research", string(s.ActiveProvider)},
			}
			s.Orchestrator.L1.Add(entry)

			// Alpha Discovery Handoff
			if strings.Contains(strings.ToUpper(res.Snippet), "$") {
				// Extract symbol (naive for alpha)
				symbol := "BTC" // Simulated extraction
				fmt.Printf("[Research] Potential Alpha discovered: %s\n", symbol)

				if s.Broker != nil {
					msg := orchestrator.Message{
						ID:        fmt.Sprintf("alpha-%d", time.Now().Unix()),
						Source:    "research-module",
						Type:      orchestrator.Event,
						Topic:     "alpha_discovery",
						Payload:   symbol,
						Timestamp: time.Now(),
					}
					s.Broker.Publish(msg)
				}
			}
		}
	}

	return results, nil
}
