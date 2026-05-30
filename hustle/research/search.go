package research

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"os"
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
	APIKey         string
}

func NewResearchSearch(p Provider, orch *orchestrator.Orchestrator) *ResearchSearch {
	key := ""
	if p == Tavily {
		key = os.Getenv("TAVILY_API_KEY")
	}
	return &ResearchSearch{
		ActiveProvider: p,
		Orchestrator:   orch,
		APIKey:         key,
	}
}

func (s *ResearchSearch) Query(q string) ([]SearchResult, error) {
	fmt.Printf("Searching via %s for: %s\n", s.ActiveProvider, q)

	if s.ActiveProvider == Tavily && s.APIKey == "" {
		fmt.Println("Warning: TAVILY_API_KEY not set, using mock data.")
	}

	results := []SearchResult{
		{URL: "https://hustle.com/info", Title: "Hustle Strategy", Snippet: "Key insights for automated revenue.", Provider: string(s.ActiveProvider)},
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
		}
	}

	return results, nil
}

type MockSearch struct {
	Orchestrator *orchestrator.Orchestrator
}

func (s *MockSearch) Query(q string) ([]SearchResult, error) {
	results := []SearchResult{
		{URL: "https://example.com/mock", Title: "Mock Result", Snippet: "Mock data.", Provider: "Mock"},
	}

	if s.Orchestrator != nil {
		for _, res := range results {
			entry := orchestrator.MemoryEntry{
				ID:        res.URL,
				Content:   res.Snippet,
				Timestamp: time.Now(),
			}
			s.Orchestrator.L1.Add(entry)
		}
	}

	return results, nil
}
