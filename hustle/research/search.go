package research

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"io"
	"net/http"
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
	if s.ActiveProvider == Tavily {
		return s.queryTavily(q)
	}

	// Fallback/Mock for others
	results := []SearchResult{
		{URL: "https://hustle.com/info", Title: "Mock Strategy", Snippet: "Fallback mock data.", Provider: string(s.ActiveProvider)},
	}
	s.storeInMemory(results)
	return results, nil
}

func (s *ResearchSearch) queryTavily(q string) ([]SearchResult, error) {
	if s.APIKey == "" {
		return nil, fmt.Errorf("TAVILY_API_KEY not set")
	}

	url := "https://api.tavily.com/search"
	payload := struct {
		APIKey string `json:"api_key"`
		Query  string `json:"query"`
	}{
		APIKey: s.APIKey,
		Query:  q,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily error: %s", string(body))
	}

	var result struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var finalResults []SearchResult
	for _, r := range result.Results {
		finalResults = append(finalResults, SearchResult{
			URL:      r.URL,
			Title:    r.Title,
			Snippet:  r.Content,
			Provider: "Tavily",
		})
	}

	s.storeInMemory(finalResults)
	return finalResults, nil
}

func (s *ResearchSearch) storeInMemory(results []SearchResult) {
	if s.Orchestrator != nil {
		for _, res := range results {
			entry := orchestrator.MemoryEntry{
				ID:        res.URL,
				Content:   fmt.Sprintf("%s: %s", res.Title, res.Snippet),
				BaseScore: 50.0,
				Timestamp: time.Now(),
				Tags:      []string{"research", res.Provider},
			}
			s.Orchestrator.L1.Add(entry)
		}
	}
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
