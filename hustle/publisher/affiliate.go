package publisher

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// AffiliateLink represents an affiliate link
type AffiliateLink struct {
	Keyword    string `json:"keyword"`
	URL        string `json:"url"`
	Program    string `json:"program"` // amazon, shareasale, etc.
	Commission string `json:"commission"`
}

// AffiliateInserter manages affiliate link insertion
type AffiliateInserter struct {
	Links []AffiliateLink
}

// NewAffiliateInserter creates a new affiliate link inserter
func NewAffiliateInserter() *AffiliateInserter {
	a := &AffiliateInserter{}
	a.LoadLinks()
	return a
}

// LoadLinks loads affiliate links from config file
func (a *AffiliateInserter) LoadLinks() error {
	data, err := os.ReadFile("affiliate_links.json")
	if err != nil {
		// Create default config if not exists
		a.Links = getDefaultAffiliateLinks()
		a.SaveLinks()
		return nil
	}

	return json.Unmarshal(data, &a.Links)
}

// SaveLinks saves affiliate links to config file
func (a *AffiliateInserter) SaveLinks() error {
	data, err := json.MarshalIndent(a.Links, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("affiliate_links.json", data, 0644)
}

// InsertAffiliateLinks inserts affiliate links into content
func (a *AffiliateInserter) InsertAffiliateLinks(content string) string {
	result := content

	for _, link := range a.Links {
		// Only insert if keyword appears and not already linked
		if strings.Contains(strings.ToLower(result), strings.ToLower(link.Keyword)) {
			// Check if not already linked
			if !strings.Contains(result, link.URL) {
				// Replace first occurrence with linked version
				keyword := link.Keyword
				linked := fmt.Sprintf("[%s](%s)", keyword, link.URL)
				result = strings.Replace(result, keyword, linked, 1)
			}
		}
	}

	return result
}

// InsertDisclosure adds an affiliate disclosure to content
func (a *AffiliateInserter) InsertDisclosure(content string) string {
	disclosure := "\n\n---\n*Disclosure: This post contains affiliate links. We may earn a commission if you make a purchase through these links at no extra cost to you.*\n"
	return content + disclosure
}

// ProcessContent processes content with affiliate links and disclosure
func (a *AffiliateInserter) ProcessContent(content string) string {
	result := a.InsertAffiliateLinks(content)
	result = a.InsertDisclosure(result)
	return result
}

// AddLink adds a new affiliate link
func (a *AffiliateInserter) AddLink(keyword, url, program, commission string) {
	a.Links = append(a.Links, AffiliateLink{
		Keyword:    keyword,
		URL:        url,
		Program:    program,
		Commission: commission,
	})
	a.SaveLinks()
}

// GetLinksForContent returns affiliate links relevant to the content
func (a *AffiliateInserter) GetLinksForContent(content string) []AffiliateLink {
	var relevant []AffiliateLink
	contentLower := strings.ToLower(content)

	for _, link := range a.Links {
		if strings.Contains(contentLower, strings.ToLower(link.Keyword)) {
			relevant = append(relevant, link)
		}
	}

	return relevant
}

// getDefaultAffiliateLinks returns a starter set of affiliate links
func getDefaultAffiliateLinks() []AffiliateLink {
	return []AffiliateLink{
		{
			Keyword:    "LM Studio",
			URL:        "https://lmstudio.ai/?ref=hustle-machine",
			Program:    "direct",
			Commission: "N/A",
		},
		{
			Keyword:    "Ollama",
			URL:        "https://ollama.ai/?ref=hustle-machine",
			Program:    "direct",
			Commission: "N/A",
		},
		{
			Keyword:    "GPU",
			URL:        "https://www.amazon.com/s?k=gpu&tag=hustlemachine-20",
			Program:    "amazon",
			Commission: "1-4%",
		},
		{
			Keyword:    "mechanical keyboard",
			URL:        "https://www.amazon.com/s?k=mechanical+keyboard&tag=hustlemachine-20",
			Program:    "amazon",
			Commission: "1-4%",
		},
		{
			Keyword:    "web hosting",
			URL:        "https://www.bluehost.com/track/hustlemachine/",
			Program:    "bluehost",
			Commission: "$65+",
		},
		{
			Keyword:    "VPN",
			URL:        "https://www.expressvpn.com/go/hustlemachine",
			Program:    "expressvpn",
			Commission: "$13-36",
		},
	}
}