package social

import (
	"bytes"
<<<<<<< HEAD
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

=======
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
>>>>>>> origin/main
	"github.com/robertpelloni/hustle/orchestrator"
)

type SocialPost struct {
	Platform    string
	Content     string
	ScheduledAt time.Time
}

var twitterAPIEndpoint = "https://api.twitter.com/2/tweets"
var linkedInAPIEndpoint = "https://api.linkedin.com/v2/ugcPosts"

type Provider interface {
	Post(orch *orchestrator.Orchestrator, platform, content string) error
	SetDryRun(enabled bool)
}

type TwitterProvider struct {
	DryRun       bool
	APIKey       string
	APISecret    string
	AccessToken  string
	AccessSecret string
	APIURL       string
}

func (p *TwitterProvider) SetDryRun(enabled bool) {
	p.DryRun = enabled
}

type twitterPostRequest struct {
	Text string `json:"text"`
}

func (p *TwitterProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[Twitter] DryRun: Posting to %s: %s\n", platform, content)
		return nil
	}

	if p.APIKey == "" || p.APISecret == "" || p.AccessToken == "" || p.AccessSecret == "" {
		return fmt.Errorf("missing Twitter OAuth environment variables")
	}

	config := oauth1.NewConfig(p.APIKey, p.APISecret)
	token := oauth1.NewToken(p.AccessToken, p.AccessSecret)
	httpClient := config.Client(context.Background(), token)

	reqBody, err := json.Marshal(twitterPostRequest{Text: content})
	if err != nil {
		return fmt.Errorf("failed to marshal twitter request: %w", err)
	}

	apiURL := p.APIURL
	if apiURL == "" {
		apiURL = twitterAPIEndpoint
	}

	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request to Twitter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twitter API error: received status code %d", resp.StatusCode)
	}

	fmt.Printf("[Twitter] Successfully posted to %s: %s\n", platform, content)
	return nil
}

func NewTwitterProvider(apiKey, apiSecret, accessToken, accessSecret string) *TwitterProvider {
	return &TwitterProvider{
		APIKey:       apiKey,
		APISecret:    apiSecret,
		AccessToken:  accessToken,
		AccessSecret: accessSecret,
		APIURL:       twitterAPIEndpoint,
	}
}

type LinkedInProvider struct {
<<<<<<< HEAD
	HTTPClient *http.Client
	APIURL     string
}

func (p *LinkedInProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	accessToken := os.Getenv("LINKEDIN_ACCESS_TOKEN")
	memberID := os.Getenv("LINKEDIN_MEMBER_ID")

	if accessToken == "" || memberID == "" {
=======
	DryRun      bool
	AccessToken string
	AuthorURN   string
	HTTPClient  *http.Client
	APIURL      string
}

func (p *LinkedInProvider) SetDryRun(enabled bool) {
	p.DryRun = enabled
}

func (p *LinkedInProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[LinkedIn] DryRun: Posting to %s: %s\n", platform, content)
		return nil
	}

	if p.AccessToken == "" || p.AuthorURN == "" {
>>>>>>> origin/main
		return fmt.Errorf("missing LINKEDIN_ACCESS_TOKEN or LINKEDIN_MEMBER_ID environment variable")
	}

	apiURL := p.APIURL
	if apiURL == "" {
<<<<<<< HEAD
		apiURL = "https://api.linkedin.com/v2/ugcPosts"
=======
		apiURL = linkedInAPIEndpoint
>>>>>>> origin/main
	}

	client := p.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	payload := map[string]interface{}{
<<<<<<< HEAD
		"author":         fmt.Sprintf("urn:li:person:%s", memberID),
=======
		"author":         p.AuthorURN,
>>>>>>> origin/main
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]interface{}{
					"text": content,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]interface{}{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
<<<<<<< HEAD
		return fmt.Errorf("failed to marshal linkedin payload: %v", err)
=======
		return fmt.Errorf("failed to marshal linkedin payload: %w", err)
>>>>>>> origin/main
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
<<<<<<< HEAD
		return fmt.Errorf("failed to create linkedin request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
=======
		return fmt.Errorf("failed to create linkedin request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.AccessToken)
>>>>>>> origin/main
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
<<<<<<< HEAD
		return fmt.Errorf("linkedin api request failed: %v", err)
=======
		return fmt.Errorf("linkedin api request failed: %w", err)
>>>>>>> origin/main
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
<<<<<<< HEAD
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("linkedin api returned status %d: %s", resp.StatusCode, string(bodyBytes))
=======
		return fmt.Errorf("linkedin api returned status %d", resp.StatusCode)
>>>>>>> origin/main
	}

	fmt.Printf("[LinkedIn] Successfully posted to %s: %s\n", platform, content)
	return nil
}

<<<<<<< HEAD
func NewLinkedInProvider() *LinkedInProvider {
	return &LinkedInProvider{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		APIURL:     "https://api.linkedin.com/v2/ugcPosts",
=======
func NewLinkedInProvider(accessToken, authorURN string) *LinkedInProvider {
	return &LinkedInProvider{
		AccessToken: accessToken,
		AuthorURN:   authorURN,
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		APIURL:      linkedInAPIEndpoint,
>>>>>>> origin/main
	}
}

func GenerateContent(orch *orchestrator.Orchestrator, topic string) string {
	prompt := fmt.Sprintf("Create a short, revolutionary social media post about %s with hashtags.", topic)
	content, err := orch.LLM.Generate(prompt)
	if err != nil {
		return fmt.Sprintf("Revolutionary insights on %s! #hustle #ai", topic)
	}
	return content
}

func SchedulePost(orch *orchestrator.Orchestrator, provider Provider, platform, topic string) {
	content := GenerateContent(orch, topic)
	fmt.Printf("Scheduling post for %s: %s\n", platform, content)

	err := provider.Post(orch, platform, content)
	if err == nil {
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("social-%s-%d", platform, time.Now().Unix()),
			Content:   fmt.Sprintf("Posted to %s: %s", platform, content),
			Timestamp: time.Now(),
			Tags:      []string{"social", platform},
		})

		orch.Ledger.Add(orchestrator.Transaction{
			Amount: 0.01,
			Type:   orchestrator.Expense,
			Hustle: "SocialMedia",
			Note:   fmt.Sprintf("API post to %s", platform),
		})
	}
}
