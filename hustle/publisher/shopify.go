package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ShopifyProduct struct {
	Title       string `json:"title"`
	BodyHTML    string `json:"body_html"`
	Vendor      string `json:"vendor"`
	ProductType string `json:"product_type"`
	Status      string `json:"status"` // active, draft
}

type ShopifyPublisher struct {
	AccessToken string
	ShopName    string // e.g. "my-store"
	DryRun      bool
	Client      *http.Client
}

func NewShopifyPublisher() *ShopifyPublisher {
	return &ShopifyPublisher{
		AccessToken: os.Getenv("SHOPIFY_ACCESS_TOKEN"),
		ShopName:    os.Getenv("SHOPIFY_SHOP_NAME"),
		Client:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *ShopifyPublisher) IsConfigured() bool {
	return s.AccessToken != "" && s.ShopName != ""
}

func (s *ShopifyPublisher) CreateProduct(p ShopifyProduct) (string, error) {
	if s.DryRun {
		fmt.Printf("[Shopify] DRY RUN: Would create product %s\n", p.Title)
		return "dry-run-id", nil
	}

	if !s.IsConfigured() {
		return "", fmt.Errorf("Shopify not configured")
	}

	url := fmt.Sprintf("https://%s.myshopify.com/admin/api/2024-01/products.json", s.ShopName)

	payload := map[string]interface{}{
		"product": p,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("X-Shopify-Access-Token", s.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Shopify error %d: %s", resp.StatusCode, string(data))
	}

	var result struct {
		Product struct {
			ID int64 `json:"id"`
		} `json:"product"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("[Shopify] ✅ Successfully created product: %s (ID: %d)\n", p.Title, result.Product.ID)
	return fmt.Sprintf("%d", result.Product.ID), nil
}
