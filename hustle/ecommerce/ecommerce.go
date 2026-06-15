package ecommerce

import (
	"fmt"
	"github.com/robertpelloni/hustle/hustle/research"
	"github.com/robertpelloni/hustle/orchestrator"
	"strings"
	"time"
)

type Product struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Niche       string   `json:"niche"`
	Keywords    []string `json:"keywords"`
	PriceRange  string   `json:"price_range"`
	DemandScore float64  `json:"demand_score"`
}

type EcommerceModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewEcommerceModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *EcommerceModule {
	return &EcommerceModule{
		Orch:   orch,
		Broker: broker,
	}
}

// DiscoverProducts finds trending products for a niche using research results
func (e *EcommerceModule) DiscoverProducts(niche string) ([]Product, error) {
	fmt.Printf("[Ecommerce] Discovering products for niche: %s\n", niche)

	searcher := research.NewResearchSearch(research.Tavily, e.Orch, e.Broker)
	results, err := searcher.Query(fmt.Sprintf("trending products %s 2026", niche))
	if err != nil {
		return nil, err
	}

	var combined string
	for _, res := range results {
		combined += fmt.Sprintf("- %s: %s\n", res.Title, res.Snippet)
	}

	prompt := fmt.Sprintf(`Analyze these search results and identify 3 high-demand products in the "%s" niche.

SEARCH RESULTS:
%s

Respond with a JSON array of objects:
[
  {
    "name": "Product Name",
    "description": "Short description of why it is trending",
    "niche": "%s",
    "keywords": ["key1", "key2"],
    "price_range": "$20-$50",
    "demand_score": 9.5
  }
]

Respond ONLY with valid JSON.`, niche, combined, niche)

	var products []Product
	if err := e.Orch.LLM.GenerateJSON(prompt, &products); err != nil {
		return nil, err
	}

	// Store in memory
	for _, p := range products {
		e.Orch.L2.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("product-%s-%d", p.Name, time.Now().Unix()),
			Content:   fmt.Sprintf("Trending Product Discovery: %s (Niche: %s, Demand: %.1f)", p.Name, p.Niche, p.DemandScore),
			BaseScore: p.DemandScore * 10.0,
			Timestamp: time.Now(),
			Tags:      []string{"ecommerce", "product", p.Niche},
		})
	}

	return products, nil
}

// GenerateListing creates a Shopify-ready product listing
func (e *EcommerceModule) GenerateListing(p Product) (string, error) {
	prompt := fmt.Sprintf(`Generate a high-converting Shopify product listing for: "%s".

Product Details:
- Description: %s
- Niche: %s
- Target Price: %s
- Keywords: %s

Requirements:
- Catchy, SEO-optimized title
- Engaging product description with benefits
- Bulleted list of key features
- "Why Choose Us" section
- Shipping and Returns placeholder text
- SEO Meta title and description

Format: HTML-ready Markdown.`, p.Name, p.Description, p.Niche, p.PriceRange, strings.Join(p.Keywords, ", "))

	return e.Orch.LLM.Generate(prompt)
}
