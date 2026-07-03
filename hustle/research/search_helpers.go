package research

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *ResearchSearch) queryDuckDuckGo(q string) ([]SearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(q))
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("DuckDuckGo returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	doc.Find(".result").Each(func(i int, sel *goquery.Selection) {
		if len(results) >= 5 {
			return
		}
		title := strings.TrimSpace(sel.Find(".result__title").Text())
		snippet := strings.TrimSpace(sel.Find(".result__snippet").Text())
		link, _ := sel.Find(".result__url").Attr("href")
		if title != "" && snippet != "" {
			results = append(results, SearchResult{
				Title:   title,
				URL:     link,
				Snippet: snippet,
				Provider: "DuckDuckGo",
			})
		}
	})
	return results, nil
}

func (s *ResearchSearch) queryFearGreed(q string) []SearchResult {
	if !strings.Contains(strings.ToLower(q), "crypto") && !strings.Contains(strings.ToLower(q), "bitcoin") && !strings.Contains(strings.ToLower(q), "btc") {
		return nil
	}

	resp, err := http.Get("https://api.alternative.me/fng/")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			Value             string `json:"value"`
			ValueClassification string `json:"value_classification"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	if len(data.Data) > 0 {
		return []SearchResult{{
			Title:   "Crypto Fear & Greed Index",
			Snippet: fmt.Sprintf("Current market sentiment is %s with a score of %s.", data.Data[0].ValueClassification, data.Data[0].Value),
			Provider: "alternative.me",
		}}
	}
	return nil
}
