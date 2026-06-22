package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// NewsletterPublisher publishes newsletters to various platforms
type NewsletterPublisher struct {
	Platform    string // buttondown, substack, convertkit
	Buttondown  *ButtondownPublisher
	Substack    *SubstackPublisher
	HTTPClient  *http.Client
}

// ButtondownPublisher publishes to Buttondown.email
type ButtondownPublisher struct {
	APIKey     string
	HTTPClient *http.Client
}

type buttondownEmail struct {
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	Tags       []string `json:"tags,omitempty"`
	PublishDate string  `json:"publish_date,omitempty"`
}

type buttondownResponse struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Status  string `json:"status"`
}

// SubstackPublisher publishes to Substack
type SubstackPublisher struct {
	APIKey     string
	PublicationURL string
	HTTPClient *http.Client
}

// NewNewsletterPublisher creates a newsletter publisher from env vars
func NewNewsletterPublisher() *NewsletterPublisher {
	p := &NewsletterPublisher{
		Platform:   os.Getenv("NEWSLETTER_PLATFORM"),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}

	if p.Platform == "" {
		p.Platform = "buttondown" // default
	}

	// Buttondown
	p.Buttondown = &ButtondownPublisher{
		APIKey:     os.Getenv("BUTTONDOWN_API_KEY"),
		HTTPClient: p.HTTPClient,
	}

	// Substack
	p.Substack = &SubstackPublisher{
		APIKey:         os.Getenv("SUBSTACK_API_KEY"),
		PublicationURL: os.Getenv("SUBSTACK_PUBLICATION_URL"),
		HTTPClient:     p.HTTPClient,
	}

	return p
}

// IsConfigured returns true if newsletter credentials are set
func (n *NewsletterPublisher) IsConfigured() bool {
	switch n.Platform {
	case "buttondown":
		return n.Buttondown.APIKey != ""
	case "substack":
		return n.Substack.APIKey != ""
	default:
		return false
	}
}

// PublishNewsletter publishes a newsletter
func (n *NewsletterPublisher) PublishNewsletter(subject, body string, tags []string) error {
	if !n.IsConfigured() {
		return fmt.Errorf("newsletter not configured (set NEWSLETTER_PLATFORM and API key)")
	}

	switch n.Platform {
	case "buttondown":
		return n.Buttondown.Publish(subject, body, tags)
	case "substack":
		return n.Substack.Publish(subject, body, tags)
	default:
		return fmt.Errorf("unsupported newsletter platform: %s", n.Platform)
	}
}

// Publish publishes to Buttondown
func (b *ButtondownPublisher) Publish(subject, body string, tags []string) error {
	if b.APIKey == "" {
		return fmt.Errorf("BUTTONDOWN_API_KEY not set")
	}

	reqBody := buttondownEmail{
		Subject: subject,
		Body:    body,
		Tags:    tags,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.buttondown.email/v1/emails", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+b.APIKey)

	resp, err := b.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Buttondown API request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Buttondown API error %d: %s", resp.StatusCode, string(respBody))
	}

	var emailResp buttondownResponse
	if err := json.Unmarshal(respBody, &emailResp); err != nil {
		return fmt.Errorf("failed to parse Buttondown response: %v", err)
	}

	fmt.Printf("[Buttondown] ✅ Newsletter queued: %s (ID: %s)\n", emailResp.Subject, emailResp.ID)
	return nil
}

// Publish publishes to Substack
func (s *SubstackPublisher) Publish(subject, body string, tags []string) error {
	if s.APIKey == "" {
		return fmt.Errorf("SUBSTACK_API_KEY not set")
	}

	// Substack uses a different API format
	reqBody := map[string]interface{}{
		"title":          subject,
		"body_html":      markdownToHTML(body),
		"post_date":      time.Now().Format(time.RFC3339),
		"tags":           tags,
		"type":           "newsletter",
		"audience":       "everyone",
		"share_draft":    false,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/api/v1/posts", s.PublicationURL)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Substack API request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Substack API error %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("[Substack] ✅ Newsletter published: %s\n", subject)
	return nil
}

// GenerateNewsletter generates a newsletter from content
func GenerateNewsletter(topic, content string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 🚀 Weekly AI Hustle Digest: %s\n\n", topic))
	sb.WriteString("Hey there,\n\n")
	sb.WriteString("Here's your weekly dose of AI-powered money-making insights:\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString(content)
	sb.WriteString("\n\n---\n\n")
	sb.WriteString("## 🎯 This Week's Action Items\n\n")
	sb.WriteString("1. **Try it yourself**: Set up a local LLM and start automating\n")
	sb.WriteString("2. **Share the love**: Forward this to a friend who'd benefit\n")
	sb.WriteString("3. **Take action**: Pick one hustle and start today\n\n")
	sb.WriteString("---\n\n")
	sb.WriteString("*Sent from the AI Hustle Machine — zero-cost, autonomous income generation.*\n")

	return sb.String()
}