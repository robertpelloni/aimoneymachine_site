package affiliate

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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

	// Route to the configured social platform for auto-posting
	// (SOCIAL_DEFAULT_PLATFORM; Bluesky on prod while X returns 402).
	platform := social.DefaultPlatform()
	provider := social.NewProvider(platform)
	socialReview := social.FormatForPlatform(sanitizeSocialReview(review, topProduct.URL), platform)

	if social.ContainsPlaceholder(socialReview) {
		// Never auto-post fill-in template text publicly (a LinkedIn
		// post shipped "[Insert your bio here]" via this class of
		// output). The review is still saved below; only the post is
		// skipped, and the skip is recorded like other post failures.
		fmt.Printf("[Affiliate] Review contains placeholder text; skipping auto-post to %s\n", platform)
		m.Orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("affiliate-post-skipped-%s", topProduct.ID),
			Content:   fmt.Sprintf("Affiliate post to %s skipped: placeholder text in generated review", platform),
			Timestamp: time.Now(),
			Tags:      []string{"affiliate", "social", "post-failed", platform},
		})
	} else if err := provider.Post(m.Orch, platform, socialReview); errors.Is(err, social.ErrDuplicatePost) {
		fmt.Printf("[Affiliate] Review already posted to %s (duplicate suppressed)\n", platform)
	} else if err != nil {
		// Visible, not fatal: the review itself is the product; record
		// the distribution failure so dashboards/memory show it.
		fmt.Printf("[Affiliate] Failed to auto-post to %s: %v\n", platform, err)
		m.Orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("affiliate-post-failed-%s", topProduct.ID),
			Content:   fmt.Sprintf("Affiliate post to %s failed: %v", platform, err),
			Timestamp: time.Now(),
			Tags:      []string{"affiliate", "social", "post-failed", platform},
		})
	} else {
		fmt.Printf("[Affiliate] Successfully posted affiliate review to %s\n", platform)
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
	// The review's destination is a single social post (280-300 chars),
	// so ask for exactly that: an unbounded markdown review took >480s
	// to generate through the FreeLLM proxy and timed out (2026-11-14),
	// only to be truncated to 300 runes anyway. Long-form reviews are
	// the content module's job.
	prompt := fmt.Sprintf(`Write a punchy, honest one-post product recommendation (MAX 240 characters total, plain text, no markdown).
Product: %s
Price: %s
Features: %s
End with this affiliate link: %s
Mention briefly that it is an affiliate link. Helpful tone, not salesy.`, product.Name, product.Price, product.Desc, product.URL)

	review, err := m.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate review: %w", err)
	}
	return review, nil
}

// sanitizeSocialReview makes raw LLM review output safe for a public
// one-post feed. Observed on prod (2026-11-14): the model appended its
// own reasoning ("But does it mention it's an affiliate link? Yes...")
// and echoed the prompt ("The user wants a punchy...") around the copy.
// Rules: strip code fences; when the product URL is present, the post
// ends at the URL (anything after it is commentary); collapse to one
// line. Without a URL it only normalizes whitespace — it never invents
// or rewrites copy.
func sanitizeSocialReview(review, productURL string) string {
	s := strings.TrimSpace(review)
	s = strings.TrimPrefix(s, "```markdown")
	s = strings.TrimPrefix(s, "```md")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	s = strings.TrimSpace(s)
	// Strip a leading instruction-echo sentence ("The user wants a
	// punchy... <real copy>") when a link later proves real copy follows.
	lower := strings.ToLower(s)
	for _, opener := range []string{"the user wants", "the user asked", "here is a", "here's a", "here is the", "here's the", "sure! here", "certainly! here", "write a "} {
		if strings.HasPrefix(lower, opener) {
			if idx := strings.Index(s, ". "); idx >= 0 && strings.Contains(s[idx:], "http") {
				s = strings.TrimSpace(s[idx+2:])
			}
			break
		}
	}
	if productURL != "" {
		if idx := strings.Index(s, productURL); idx >= 0 {
			s = strings.TrimSpace(s[:idx+len(productURL)])
			s = strings.Trim(s, `"'`)
		}
	}
	// One line for a social post.
	s = strings.Join(strings.Fields(s), " ")
	return s
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
