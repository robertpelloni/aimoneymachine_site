// Package products is the Digital Product Factory: it generates a
// sellable digital product (Notion-style template, guide, or preset pack)
// as files on disk plus a storefront listing, and can publish the listing
// to Gumroad when GUMROAD_ACCESS_TOKEN is configured.
//
// Honesty contract (same as the Shorts pipeline): generation always
// works locally; the listing step records "not_configured" in the
// manifest instead of pretending a product is live in a store.
package products

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

// ProductType selects the kind of product to generate.
type ProductType string

const (
	Template ProductType = "template" // Notion-style template / checklist system
	Guide    ProductType = "guide"    // Written guide / playbook
	Preset   ProductType = "preset"   // Preset / config / prompt pack
)

// Listing is the storefront metadata for a generated product.
type Listing struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int    `json:"price_cents"`
	URL         string `json:"url,omitempty"`
	StoreID     string `json:"store_product_id,omitempty"`
	Status      string `json:"status"` // not_configured | listed | failed
	Note        string `json:"note,omitempty"`
}

// Manifest records what a generation run produced.
type Manifest struct {
	Topic     string      `json:"topic"`
	Type      ProductType `json:"type"`
	Slug      string      `json:"slug"`
	Dir       string      `json:"dir"`
	Files     []string    `json:"files"`
	Listing   Listing     `json:"listing"`
	CreatedAt time.Time   `json:"created_at"`
}

// Factory generates digital products under output/products.
type Factory struct {
	Orch     *orchestrator.Orchestrator
	OutDir   string
	Gumroad  *GumroadClient
}

// NewFactory builds a factory; the Gumroad client is configured only
// when GUMROAD_ACCESS_TOKEN is set.
func NewFactory(orch *orchestrator.Orchestrator) *Factory {
	return &Factory{
		Orch:    orch,
		OutDir:  "output/products",
		Gumroad: NewGumroadClientFromEnv(),
	}
}

// Generate creates the product content, listing copy, and manifest.
// It returns the manifest directory.
func (f *Factory) Generate(topic string, ptype ProductType) (*Manifest, error) {
	if topic == "" {
		return nil, fmt.Errorf("topic is required")
	}
	if ptype != Template && ptype != Guide && ptype != Preset {
		return nil, fmt.Errorf("unknown product type %q (want template|guide|preset)", ptype)
	}

	content, err := f.Orch.LLM.Generate(productPrompt(topic, ptype))
	if err != nil {
		return nil, fmt.Errorf("generating product content: %w", err)
	}
	listingCopy, err := f.Orch.LLM.Generate(listingPrompt(topic, ptype))
	if err != nil {
		return nil, fmt.Errorf("generating listing copy: %w", err)
	}

	slug := slugify(fmt.Sprintf("%s-%s", topic, ptype))
	dir := filepath.Join(f.OutDir, fmt.Sprintf("%s-%d", slug, time.Now().Unix()))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	productPath := filepath.Join(dir, "product.md")
	if err := os.WriteFile(productPath, []byte(content), 0644); err != nil {
		return nil, err
	}
	listingPath := filepath.Join(dir, "listing.md")
	if err := os.WriteFile(listingPath, []byte(listingCopy), 0644); err != nil {
		return nil, err
	}

	m := &Manifest{
		Topic:  topic,
		Type:   ptype,
		Slug:   slug,
		Dir:    dir,
		Files:  []string{productPath, listingPath},
		Listing: Listing{
			Name:        titleFor(topic, ptype),
			Description: firstLine(listingCopy),
			PriceCents:  suggestedPriceCents(ptype),
			Status:      "not_configured",
			Note:        "set GUMROAD_ACCESS_TOKEN to list on Gumroad",
		},
		CreatedAt: time.Now(),
	}

	// Optional real listing.
	if f.Gumroad != nil {
		id, link, err := f.Gumroad.CreateProduct(m.Listing.Name, m.Listing.Description, m.Listing.PriceCents)
		if err != nil {
			m.Listing.Status = "failed"
			m.Listing.Note = err.Error()
			fmt.Printf("[Products] Gumroad listing failed: %v\n", err)
		} else {
			m.Listing.Status = "listed"
			m.Listing.StoreID = id
			m.Listing.URL = link
			m.Listing.Note = ""
			fmt.Printf("[Products] Listed on Gumroad: %s\n", link)
		}
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if err := orchestrator.WriteFileAtomic(manifestPath, data, 0644); err != nil {
		return nil, err
	}
	m.Files = append(m.Files, manifestPath)

	f.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("product-%s-%d", slug, time.Now().Unix()),
		Content:   fmt.Sprintf("Generated %s product %q in %s (listing: %s)", ptype, topic, dir, m.Listing.Status),
		Timestamp: time.Now(),
		Tags:      []string{"products", string(ptype), "digital-product"},
	})
	fmt.Printf("[Products] Generated %s in %s (listing: %s)\n", slug, dir, m.Listing.Status)
	return m, nil
}

func productPrompt(topic string, ptype ProductType) string {
	kind := map[ProductType]string{
		Template: "a ready-to-use template system (structure, sections, checklists, and fill-in fields)",
		Guide:    "a practical step-by-step guide (chapters, actionable steps, examples, and a quick-start checklist)",
		Preset:   "a preset pack (named presets with exact settings/values and when to use each)",
	}[ptype]
	return fmt.Sprintf("Create %s about %q that a buyer can use immediately. Write the full product content in markdown, not an outline.", kind, topic)
}

func listingPrompt(topic string, ptype ProductType) string {
	return fmt.Sprintf("Write storefront listing copy for a %s about %q: one benefit-led first line, then 3-5 bullet benefits, who it is for, and a call to action. Plain text.", ptype, topic)
}

func titleFor(topic string, ptype ProductType) string {
	suffix := map[ProductType]string{Template: "Template", Guide: "Guide", Preset: "Preset Pack"}[ptype]
	return fmt.Sprintf("%s %s", topic, suffix)
}

func suggestedPriceCents(ptype ProductType) int {
	switch ptype {
	case Guide:
		return 2900
	case Preset:
		return 1900
	default: // Template
		return 2400
	}
}

func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(strings.TrimPrefix(line, "# ")); t != "" {
			return t
		}
	}
	return ""
}

func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// GumroadClient creates products via the Gumroad API v2.
type GumroadClient struct {
	Token   string
	BaseURL string
	Client  *http.Client
}

// NewGumroadClientFromEnv returns nil when GUMROAD_ACCESS_TOKEN is unset.
func NewGumroadClientFromEnv() *GumroadClient {
	token := os.Getenv("GUMROAD_ACCESS_TOKEN")
	if token == "" {
		return nil
	}
	return &GumroadClient{
		Token:   token,
		BaseURL: getEnvDefault("GUMROAD_API_URL", "https://api.gumroad.com"),
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// CreateProduct lists a product and returns its store ID and URL.
func (g *GumroadClient) CreateProduct(name, description string, priceCents int) (string, string, error) {
	form := url.Values{
		"access_token": {g.Token},
		"name":         {name},
		"description":  {description},
		"price_cents":  {strconv.Itoa(priceCents)},
	}
	client := g.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.PostForm(strings.TrimRight(g.BaseURL, "/")+"/v2/products", form)
	if err != nil {
		return "", "", fmt.Errorf("gumroad request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("gumroad status %d: %s", resp.StatusCode, strings.TrimSpace(string(body))[:min(len(body), 200)])
	}
	var parsed struct {
		Success bool `json:"success"`
		Product struct {
			ID        string `json:"id"`
			ShortURL  string `json:"short_url"`
			ProductURL string `json:"url"`
		} `json:"product"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", fmt.Errorf("gumroad response parse: %w", err)
	}
	if !parsed.Success || parsed.Product.ID == "" {
		return "", "", fmt.Errorf("gumroad did not create a product: %s", strings.TrimSpace(string(body))[:min(len(body), 200)])
	}
	link := parsed.Product.ShortURL
	if link == "" {
		link = parsed.Product.ProductURL
	}
	return parsed.Product.ID, link, nil
}
