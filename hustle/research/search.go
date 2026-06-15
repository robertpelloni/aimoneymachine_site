package research

import (
	"encoding/json"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"net/http"
	"net/url"
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

			// Dynamic Niche Arbitrage: Sentiment Spike Detection
			s.detectSentimentSpike(res)
		}
	}

	return results, nil
}

func (s *ResearchSearch) detectSentimentSpike(res SearchResult) {
	if s.Orchestrator == nil || s.Orchestrator.LLM == nil {
		return
	}

	// Heuristic: look for words indicating high interest or "viral"
	viralWords := []string{"viral", "trending", "explosion", "massive interest", "huge gap", "nobody is talking about"}
	isPotentiallyViral := false
	for _, word := range viralWords {
		if strings.Contains(strings.ToLower(res.Snippet), word) {
			isPotentiallyViral = true
			break
		}
	}

	if isPotentiallyViral {
		fmt.Printf("[Research] Detected potential sentiment spike in: %s\n", res.Title)

		prompt := fmt.Sprintf("Analyze if this search result represents a 'niche arbitrage' opportunity (high demand, low content). Result: %s. Respond with ONLY 'SPIKE' or 'NORMAL'.", res.Snippet)
		resp, _ := s.Orchestrator.LLM.Generate(prompt)

		if strings.Contains(strings.ToUpper(resp), "SPIKE") {
			fmt.Printf("[Research] 🚀 SENTIMENT SPIKE CONFIRMED for: %s\n", res.Title)

			// Trigger autonomous content generation via A2A
			if s.Broker != nil {
				msg := orchestrator.Message{
					ID:        fmt.Sprintf("spike-%d", time.Now().Unix()),
					Source:    "research-module",
					Type:      orchestrator.Command,
					Topic:     "niche_arbitrage",
					Payload:   fmt.Sprintf("hustle://content?topic=%s&type=seo&publish=true", url.QueryEscape(res.Title)),
					Timestamp: time.Now(),
				}
				s.Broker.Publish(msg)
			}

			// Log to L2
			s.Orchestrator.L2.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("niche-arb-%d", time.Now().Unix()),
				Content:   fmt.Sprintf("Niche Arbitrage Spike: %s", res.Title),
				BaseScore: 95.0,
				Timestamp: time.Now(),
				Tags:      []string{"research", "arbitrage", "spike", res.Title},
			})
		}
	}
}

func (s *ResearchSearch) queryTavily(q string) ([]SearchResult, error) {
	url := "https://api.tavily.com/search"
	payload := map[string]interface{}{
		"api_key":      s.APIKey,
		"query":        q,
		"search_depth": "basic",
	}
	body, _ := json.Marshal(payload)

	var resp *http.Response
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		resp, err = http.Post(url, "application/json", strings.NewReader(string(body)))
		if err == nil {
			if resp.StatusCode == http.StatusOK {
				break
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			return nil, fmt.Errorf("tavily API returned status: %d", resp.StatusCode)
		}
	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("tavily API returned status: %d (retries exhausted)", resp.StatusCode)
	}
	defer resp.Body.Close()

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
