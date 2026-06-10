package research

import (
	"encoding/json"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"net/http"
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

	var results []SearchResult
	var err error

	if s.ActiveProvider == Tavily && s.APIKey != "" {
		results, err = s.queryTavily(q)
		if err != nil {
			fmt.Printf("[Research] Tavily query failed: %v. Falling back to mock.\n", err)
		}
	}

	if len(results) == 0 {
		if s.ActiveProvider == Tavily && s.APIKey == "" {
			fmt.Println("Warning: TAVILY_API_KEY not set, using mock data.")
		}
		results = []SearchResult{
			{URL: "https://hustle.com/info", Title: "Hustle Strategy", Snippet: "Key insights for automated revenue with $BTC trends.", Provider: string(s.ActiveProvider)},
		}
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

func (s *ResearchSearch) queryTavily(q string) ([]SearchResult, error) {
	url := "https://api.tavily.com/search"
	payload := map[string]interface{}{
		"api_key":      s.APIKey,
		"query":        q,
		"search_depth": "basic",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tavily API returned status: %d", resp.StatusCode)
	}

	var data struct {
		Results []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Snippet string `json:"content"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, r := range data.Results {
		results = append(results, SearchResult{
			URL:      r.URL,
			Title:    r.Title,
			Snippet:  r.Snippet,
			Provider: "Tavily",
		})
	}
	return results, nil
}
