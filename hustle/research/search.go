package research

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type SearchResult struct {
	URL       string
	Snippet   string
	Title     string
	Provider  string
	Sentiment string // BULLISH, BEARISH, or NEUTRAL (set by sentiment sources)
}

type SearchInterface interface {
	Query(q string) ([]SearchResult, error)
}

type Provider string

const (
	DuckDuckGo Provider = "DuckDuckGo"
	Tavily     Provider = "Tavily"
	Brave      Provider = "Brave"
	Google     Provider = "Google"
)

type ResearchSearch struct {
	ActiveProvider Provider
	Orchestrator   *orchestrator.Orchestrator
	Broker         *orchestrator.A2ABroker
	APIKey         string
	lastCoinGecko  time.Time // rate limiter for CoinGecko API
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

	// 1. Try DuckDuckGo (free, no key needed)
	if s.ActiveProvider == DuckDuckGo {
		results, _ = s.queryDuckDuckGo(q)
	}

	// 2. If DuckDuckGo fails, try Fear & Greed Index (free, no key)
	//    This gives real crypto market sentiment
	if len(results) == 0 {
		fg := s.queryFearGreed(q)
		if len(fg) > 0 {
			results = fg
		}
	}

	// 3. Try CoinGecko sentiment (we have the API key)
	if len(results) == 0 {
		cg := s.queryCoinGecko(q)
		if len(cg) > 0 {
			results = cg
		}
	}

	// 4. Fallback to Tavily if available
	if len(results) == 0 && s.APIKey != "" {
		results, _ = s.queryTavily(q)
	}

	// 5. Absolute last resort: mock data
	if len(results) == 0 {
		fmt.Println("[Research] All sentiment sources failed, using fallback.")
		results = []SearchResult{{
			URL:      "https://hustle.com/fallback",
			Title:    "Market Overview",
			Snippet:  "Neutral market conditions for " + q + ". No strong directional signals detected.",
			Provider: "Fallback",
		}}
	}

	s.processResults(results)
	return results, nil
}

func (s *ResearchSearch) processResults(results []SearchResult) {
	if s.Orchestrator == nil {
		return
	}
	for _, res := range results {
		s.Orchestrator.L1.Add(orchestrator.MemoryEntry{
			ID:        res.URL,
			Content:   fmt.Sprintf("%s: %s", res.Title, res.Snippet),
			BaseScore: 50.0,
			Timestamp: time.Now(),
			Tags:      []string{"research", res.Provider},
		})
		if strings.Contains(strings.ToUpper(res.Snippet), "$") {
			fmt.Printf("[Research] Potential Alpha discovered: BTC\n")
			if s.Broker != nil {
				s.Broker.Publish(orchestrator.Message{
					ID:        fmt.Sprintf("alpha-%d", time.Now().Unix()),
					Source:    "research-module",
					Type:      orchestrator.Event,
					Topic:     "alpha_discovery",
					Payload:   "BTC",
					Timestamp: time.Now(),
				})
			}
		}
		s.detectSentimentSpike(res)
	}
}

func (s *ResearchSearch) detectSentimentSpike(res SearchResult) {
	if s.Orchestrator == nil || s.Orchestrator.LLM == nil {
		return
	}
	viralWords := []string{"viral", "trending", "explosion", "massive interest", "huge gap", "nobody is talking about"}
	for _, word := range viralWords {
		if strings.Contains(strings.ToLower(res.Snippet), word) {
			fmt.Printf("[Research] Detected potential sentiment spike in: %s\n", res.Title)
			resp, _ := s.Orchestrator.LLM.Generate(
				fmt.Sprintf("Analyze if this result is a niche arbitrage opportunity. Result: %s. Respond ONLY 'SPIKE' or 'NORMAL'.", res.Snippet))
			if strings.Contains(strings.ToUpper(resp), "SPIKE") {
				fmt.Printf("[Research] SENTIMENT SPIKE CONFIRMED for: %s\n", res.Title)
				if s.Broker != nil {
					s.Broker.Publish(orchestrator.Message{
						ID:        fmt.Sprintf("spike-%d", time.Now().Unix()),
						Source:    "research-module",
						Type:      orchestrator.Command,
						Topic:     "niche_arbitrage",
						Payload:   fmt.Sprintf("hustle://content?topic=%s&type=seo&publish=true", url.QueryEscape(res.Title)),
						Timestamp: time.Now(),
					})
				}
			}
			break
		}
	}
}

// ── Real Sentiment Sources ──

func (s *ResearchSearch) queryFearGreed(q string) []SearchResult {
	// Free Crypto Fear & Greed Index — no API key required
	type fngData struct {
		Data []struct {
			Value          string `json:"value"`
			Classification string `json:"value_classification"`
		} `json:"data"`
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.alternative.me/fng/?limit=1")
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	var data fngData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || len(data.Data) == 0 {
		return nil
	}

	val := data.Data[0].Value
	cls := data.Data[0].Classification
	// Determine sentiment directly from Fear & Greed value
	var sentiment string
	v := 0
	fmt.Sscanf(val, "%d", &v)
	switch {
	case v <= 25:
		sentiment = "BEARISH"
	case v <= 45:
		sentiment = "NEUTRAL"
	case v <= 55:
		sentiment = "NEUTRAL"
	case v <= 75:
		sentiment = "NEUTRAL"
	default:
		sentiment = "BULLISH"
	}

	fmt.Printf("[Research] Fear & Greed Index: %s/100 (%s) → %s\n", val, cls, sentiment)
	return []SearchResult{{
		URL:       "https://alternative.me/crypto/fear-and-greed-index/",
		Title:     fmt.Sprintf("Fear & Greed: %s (%s/100)", cls, val),
		Snippet:   fmt.Sprintf("Fear & Greed Index: %s/100. Market is in %s. Sentiment: %s.", val, cls, sentiment),
		Provider:  "FearGreedIndex",
		Sentiment: sentiment,
	}}
}

func (s *ResearchSearch) queryCoinGecko(q string) []SearchResult {
	apiKey := os.Getenv("COINGECKO_API_KEY")
	if apiKey == "" {
		return nil
	}
	// Only query CoinGecko for BTC to stay within free tier limits (max 1 call per 60s)
	if !strings.HasPrefix(strings.ToUpper(q), "BTC") {
		return nil
	}
	if time.Since(s.lastCoinGecko) < 60*time.Second {
		return nil
	}
	s.lastCoinGecko = time.Now()

	coin := coinName(strings.ToLower(parseSymbol(q)))
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s?localization=false&tickers=false&community_data=true&developer_data=false&sparkline=false", coin)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-cg-demo-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil
	}

	sup, _ := raw["sentiment_votes_up_percentage"].(float64)
	if sup == 0 {
		return nil
	}

	sentiment := "NEUTRAL"
	if sup >= 55 {
		sentiment = "BULLISH"
	} else if sup <= 45 {
		sentiment = "BEARISH"
	}

	fmt.Printf("[Research] CoinGecko: %.0f%% bullish on %s (%s)\n", sup, coin, sentiment)
	return []SearchResult{{
		URL:       fmt.Sprintf("https://www.coingecko.com/en/coins/%s", coin),
		Title:     fmt.Sprintf("CoinGecko: %.0f%% Bullish", sup),
		Snippet:   fmt.Sprintf("Community sentiment for %s: %.0f%% bullish. Overall: %s.", coin, sup, sentiment),
		Provider:  "CoinGecko",
		Sentiment: sentiment,
	}}
}

// ── Helpers ──

func parseSymbol(q string) string {
	parts := strings.Fields(q)
	if len(parts) > 0 {
		return strings.ToUpper(parts[0])
	}
	return "BTC"
}

func coinName(symbol string) string {
	m := map[string]string{
		"BTC": "bitcoin", "ETH": "ethereum", "SOL": "solana",
		"DOGE": "dogecoin", "XRP": "ripple", "ADA": "cardano",
		"LINK": "chainlink",
	}
	if v, ok := m[symbol]; ok {
		return v
	}
	return "bitcoin"
}

// ── Web Search Providers ──

func (s *ResearchSearch) queryDuckDuckGo(q string) ([]SearchResult, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(q))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo returned status: %d", resp.StatusCode)
	}
	var data struct {
		AbstractText string `json:"AbstractText"`
		AbstractURL  string `json:"AbstractURL"`
		Results      []struct {
			FirstURL string `json:"FirstURL"`
			Text     string `json:"Text"`
		} `json:"Results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	var results []SearchResult
	if data.AbstractText != "" {
		results = append(results, SearchResult{URL: data.AbstractURL, Title: "Summary", Snippet: data.AbstractText, Provider: "DuckDuckGo"})
	}
	for _, r := range data.Results {
		results = append(results, SearchResult{URL: r.FirstURL, Title: r.Text, Snippet: r.Text, Provider: "DuckDuckGo"})
	}
	return results, nil
}

func (s *ResearchSearch) queryTavily(q string) ([]SearchResult, error) {
	url := "https://api.tavily.com/search"
	payload := map[string]interface{}{"api_key": s.APIKey, "query": q, "search_depth": "basic"}
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
		results = append(results, SearchResult{URL: r.URL, Title: r.Title, Snippet: r.Snippet, Provider: "Tavily"})
	}
	return results, nil
}
