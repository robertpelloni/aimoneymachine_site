package affiliate

import (
	"fmt"
	"os"
	"strings"

	"github.com/robertpelloni/hustle/orchestrator"
	"github.com/robertpelloni/hustle/hustle/social"
	"github.com/robertpelloni/hustle/hustle/publisher"
)

// AffiliateModule manages affiliate marketing pipelines
type AffiliateModule struct {
	Orch *orchestrator.Orchestrator
}

// NewAffiliateModule creates a new instance
func NewAffiliateModule(orch *orchestrator.Orchestrator) *AffiliateModule {
	return &AffiliateModule{
		Orch: orch,
	}
}

// Run executes a full affiliate marketing pipeline
func (m *AffiliateModule) Run(niche string) error {
	fmt.Printf("[Affiliate] Running full pipeline for niche: %s\n", niche)

	// Step 1: Discover products
	products, err := m.DiscoverProducts(niche)
	if err != nil {
		return err
	}
	if len(products) == 0 {
		return fmt.Errorf("no products found for niche: %s", niche)
	}

	// Step 2: Pick top product
	topProduct := products[0]
	fmt.Printf("[Affiliate] Selected top product: %s\n", topProduct.Name)

	// Step 3: Generate review
	review, err := m.GenerateReview(topProduct)
	if err != nil {
		return err
	}

	// Step 4: Add to dashboard / publish queue
	fmt.Printf("[Affiliate] Generated Review:\n%s\n", review)

	// Route to the social queue for auto-posting
	provider := social.NewTwitterProvider(
		os.Getenv("TWITTER_API_KEY"),
		os.Getenv("TWITTER_API_SECRET"),
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		os.Getenv("TWITTER_ACCESS_SECRET"),
	)

	// Enforce 280-character limit
	reviewRunes := []rune(review)
	socialReview := review
	if len(reviewRunes) > 280 {
		socialReview = string(reviewRunes[:277]) + "..."
	}

	// Call SchedulePost so the generated review hits Twitter. Note SchedulePost generates its own content, so we will use the provider directly.
	if err := provider.Post(m.Orch, "Twitter", socialReview); err != nil {
		fmt.Printf("[Affiliate] Failed to auto-post to Twitter: %v\n", err)
	} else {
		fmt.Println("[Affiliate] Successfully queued/posted affiliate review to Twitter")
	}

	// Embed existing affiliate links using the AffiliateInserter
	inserter := publisher.NewAffiliateInserter()
	review = inserter.ProcessContent(review)

	// Save to DB or file (simple logging for now)
	m.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("affiliate-review-%s", topProduct.ID),
		Content:   fmt.Sprintf("Generated affiliate review for %s", topProduct.Name),
		Tags:      []string{"affiliate", "review"},
	})

	return nil
}

// Product represents an affiliate product
type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price string `json:"price"`
	URL   string `json:"url"`
	Desc  string `json:"description"`
}

// DiscoverProducts finds hot products via LLM simulation (or future scraping)
func (m *AffiliateModule) DiscoverProducts(niche string) ([]Product, error) {
	prompt := fmt.Sprintf(`Simulate an Amazon affiliate product discovery for the niche: "%s".
Respond with a JSON array of 3 realistic, high-converting product objects:
[
  {
    "id": "B08N5WRWNW",
    "name": "Example Product",
    "price": "$99.99",
    "url": "https://amazon.com/dp/B08N5WRWNW?tag=your-tag-20",
    "description": "Short feature description"
  }
]
Respond ONLY with valid JSON.`, niche)

	var products []Product
	if err := m.Orch.LLM.GenerateJSON(prompt, &products); err != nil {
		// Mock fallback if JSON parsing from LLM fails
		products = append(products, Product{
			ID: "B0MOCK123",
			Name: "Mock Affiliate Product for " + niche,
			Price: "$149.99",
			URL: "https://amazon.com/dp/B0MOCK123?tag=hustlemachine-20",
			Desc: "A highly rated product automatically discovered.",
		})
	}

	// If the LLM returned an empty array but didn't error, add the mock
	if len(products) == 0 {
	    products = append(products, Product{
			ID: "B0MOCK123",
			Name: "Mock Affiliate Product for " + niche,
			Price: "$149.99",
			URL: "https://amazon.com/dp/B0MOCK123?tag=hustlemachine-20",
			Desc: "A highly rated product automatically discovered.",
		})
	}

	return products, nil
}

// GenerateReview uses the LLM to write a high-converting affiliate review
func (m *AffiliateModule) GenerateReview(product Product) (string, error) {
	prompt := fmt.Sprintf(`Write a high-converting, honest-sounding product review for the following item.
Product: %s
Price: %s
Features: %s

Include the affiliate link (%s) naturally in the text. Add a clear disclaimer that it is an affiliate link. Ensure the tone is helpful and not overly salesy. Format as markdown.`, product.Name, product.Price, product.Desc, product.URL)

	review, err := m.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate review: %w", err)
	}
	return review, nil
}

// InsertAffiliateLinksIntoContent reads a directory of markdown files and inserts affiliate links where appropriate
func (m *AffiliateModule) InsertAffiliateLinksIntoContent(dir string) error {
	fmt.Printf("[Affiliate] Starting batch affiliate link insertion in %s\n", dir)
	inserter := publisher.NewAffiliateInserter()

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	updatedCount := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			path := dir + "/" + file.Name()
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			original := string(data)
			processed := inserter.ProcessContent(original)

			if original != processed {
				os.WriteFile(path, []byte(processed), 0644)
				updatedCount++
			}
		}
	}

	fmt.Printf("[Affiliate] Completed batch insertion. Updated %d files.\n", updatedCount)
	return nil
}
