package publisher

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// WordPressAuthMethod indicates which auth strategy to use
type WordPressAuthMethod int

const (
	AuthJWT          WordPressAuthMethod = iota // JWT plugin (jwt-auth/v1/token)
	AuthApplicationPassword                      // Application Passwords (Basic Auth)
)

// WordPressPublisher publishes content to WordPress via REST API
type WordPressPublisher struct {
	BaseURL       string // e.g., https://yoursite.com
	Username      string
	Password      string // WordPress login password or Application Password
	AuthMethod    WordPressAuthMethod
	Token         string // JWT token (cached)
	HTTPClient    *http.Client
	DryRun        bool
}

type wpPostRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Status     string   `json:"status"` // publish, draft, pending
	Categories []int    `json:"categories,omitempty"`
	Tags       []int    `json:"tags,omitempty"`
	Slug       string   `json:"slug,omitempty"`
	Excerpt    string   `json:"excerpt,omitempty"`
}

type wpPostResponse struct {
	ID      int    `json:"id"`
	Title   struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Link    string `json:"link"`
	Status  string `json:"status"`
}

type wpJWTResponse struct {
	Token string `json:"token"`
}

// NewWordPressPublisher creates a new WordPress publisher from env vars
func NewWordPressPublisher() *WordPressPublisher {
	authMethod := AuthJWT
	if os.Getenv("WORDPRESS_AUTH_METHOD") == "application-password" {
		authMethod = AuthApplicationPassword
	}

	return &WordPressPublisher{
		BaseURL:    os.Getenv("WORDPRESS_URL"),
		Username:   os.Getenv("WORDPRESS_USERNAME"),
		Password:   os.Getenv("WORDPRESS_PASSWORD"),
		AuthMethod: authMethod,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// IsConfigured returns true if WordPress credentials are set
func (w *WordPressPublisher) IsConfigured() bool {
	return w.BaseURL != "" && w.Username != "" && w.Password != ""
}

// getJWTToken authenticates and gets a JWT token
func (w *WordPressPublisher) getJWTToken() (string, error) {
	if w.Token != "" {
		return w.Token, nil
	}
	
	reqBody := map[string]string{
		"username": w.Username,
		"password": w.Password,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	
	req, _ := http.NewRequest("POST", w.BaseURL+"/wp-json/jwt-auth/v1/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := w.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("JWT auth failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("JWT auth error %d: %s", resp.StatusCode, string(body))
	}
	
	var jwtResp wpJWTResponse
	if err := json.Unmarshal(body, &jwtResp); err != nil {
		return "", fmt.Errorf("failed to parse JWT response: %v", err)
	}
	
	w.Token = jwtResp.Token
	return jwtResp.Token, nil
}

// PublishPost publishes a post to WordPress
func (w *WordPressPublisher) PublishPost(title, content, status string, categories []string, tags []string) (*wpPostResponse, error) {
	if !w.IsConfigured() {
		return nil, fmt.Errorf("WordPress not configured (set WORDPRESS_URL, WORDPRESS_USERNAME, WORDPRESS_PASSWORD)")
	}

	if w.DryRun {
		fmt.Printf("[WordPress] 🧪 DRY RUN: Would publish '%s' to %s\n", title, w.BaseURL)
		return &wpPostResponse{
			Title: struct {
				Rendered string `json:"rendered"`
			}{Rendered: title},
			Link:   "#dry-run",
			Status: "draft",
		}, nil
	}

	// Clean markdown content for WordPress (convert to basic HTML)
	htmlContent := markdownToHTML(content)

	reqBody := wpPostRequest{
		Title:   title,
		Content: htmlContent,
		Status:  status,
		Excerpt: extractExcerpt(content),
	}

	bodyBytes, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/wp-json/wp/v2/posts", w.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set authentication based on method
	switch w.AuthMethod {
	case AuthApplicationPassword:
		// Application Passwords use Basic Auth: base64(username:application_password)
		auth := base64.StdEncoding.EncodeToString([]byte(w.Username + ":" + w.Password))
		req.Header.Set("Authorization", "Basic "+auth)
		fmt.Printf("[WordPress] Using Application Password auth\n")
	case AuthJWT:
		// JWT uses Bearer token
		token, err := w.getJWTToken()
		if err != nil {
			return nil, fmt.Errorf("JWT authentication failed: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		fmt.Printf("[WordPress] Using JWT auth\n")
	}

	resp, err := w.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WordPress API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("WordPress API error %d: %s", resp.StatusCode, string(body))
	}

	var postResp wpPostResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		return nil, fmt.Errorf("failed to parse WordPress response: %v", err)
	}

	fmt.Printf("[WordPress] ✅ Published: %s\n", postResp.Link)
	return &postResp, nil
}

// PublishMarkdown publishes a markdown file's content to WordPress
func (w *WordPressPublisher) PublishMarkdown(filePath string) (*wpPostResponse, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	content := string(data)
	
	// Extract title from markdown (first H1)
	title := extractTitle(content)
	if title == "" {
		title = "Untitled Post"
	}

	return w.PublishPost(title, content, "publish", nil, nil)
}

// markdownToHTML converts basic markdown to HTML
func markdownToHTML(md string) string {
	lines := strings.Split(md, "\n")
	var result strings.Builder
	inCodeBlock := false

	for _, line := range lines {
		// Skip YAML frontmatter
		if strings.HasPrefix(line, "---") {
			continue
		}

		// Code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				result.WriteString("</code></pre>\n")
			} else {
				result.WriteString("<pre><code>")
			}
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			result.WriteString(line + "\n")
			continue
		}

		// Headers
		if strings.HasPrefix(line, "# ") {
			result.WriteString("<h1>" + strings.TrimPrefix(line, "# ") + "</h1>\n")
		} else if strings.HasPrefix(line, "## ") {
			result.WriteString("<h2>" + strings.TrimPrefix(line, "## ") + "</h2>\n")
		} else if strings.HasPrefix(line, "### ") {
			result.WriteString("<h3>" + strings.TrimPrefix(line, "### ") + "</h3>\n")
		} else if strings.HasPrefix(line, "#### ") {
			result.WriteString("<h4>" + strings.TrimPrefix(line, "#### ") + "</h4>\n")
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			result.WriteString("<li>" + strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ") + "</li>\n")
		} else if strings.HasPrefix(line, "> ") {
			result.WriteString("<blockquote>" + strings.TrimPrefix(line, "> ") + "</blockquote>\n")
		} else if strings.TrimSpace(line) == "" {
			result.WriteString("<br>\n")
		} else {
			// Bold
			line = strings.ReplaceAll(line, "**", "<b>")
			// Italic
			line = strings.ReplaceAll(line, "*", "<i>")
			// Inline code
			line = strings.ReplaceAll(line, "`", "<code>")
			result.WriteString("<p>" + line + "</p>\n")
		}
	}

	return result.String()
}

// extractTitle extracts the H1 title from markdown
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// extractExcerpt extracts the first paragraph as an excerpt
func extractExcerpt(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "*") {
			// Clean markdown formatting
			line = strings.ReplaceAll(line, "**", "")
			line = strings.ReplaceAll(line, "*", "")
			line = strings.ReplaceAll(line, "`", "")
			if len(line) > 200 {
				line = line[:200] + "..."
			}
			return line
		}
	}
	return ""
}

// PublishHTMLPost publishes pre‑rendered HTML content to WordPress without markdown conversion.
func (w *WordPressPublisher) PublishHTMLPost(title, htmlContent, status string, categories []string, tags []string) (*wpPostResponse, error) {
	if !w.IsConfigured() {
		return nil, fmt.Errorf("WordPress not configured (set WORDPRESS_URL, WORDPRESS_USERNAME, WORDPRESS_PASSWORD)")
	}

	if w.DryRun {
		fmt.Printf("[WordPress] 🧪 DRY RUN: Would publish HTML post '%s' to %s\n", title, w.BaseURL)
		return &wpPostResponse{
			Title: struct {
				Rendered string `json:"rendered"`
			}{Rendered: title},
			Link:   "#dry-run",
			Status: "draft",
		}, nil
	}

	reqBody := wpPostRequest{
		Title:   title,
		Content: htmlContent,
		Status:  status,
		Excerpt: extractExcerpt(htmlContent),
	}

	bodyBytes, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/wp-json/wp/v2/posts", w.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set authentication based on method
	switch w.AuthMethod {
	case AuthApplicationPassword:
		auth := base64.StdEncoding.EncodeToString([]byte(w.Username + ":" + w.Password))
		req.Header.Set("Authorization", "Basic "+auth)
		fmt.Printf("[WordPress] Using Application Password auth\n")
	case AuthJWT:
		token, err := w.getJWTToken()
		if err != nil {
			return nil, fmt.Errorf("JWT authentication failed: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		fmt.Printf("[WordPress] Using JWT auth\n")
	}

	resp, err := w.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WordPress API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("WordPress API error %d: %s", resp.StatusCode, string(body))
	}

	var postResp wpPostResponse
	if err := json.Unmarshal(body, &postResp); err != nil {
		return nil, fmt.Errorf("failed to parse WordPress response: %v", err)
	}

	fmt.Printf("[WordPress] ✅ Published HTML post: %s\n", postResp.Link)
	return &postResp, nil
}