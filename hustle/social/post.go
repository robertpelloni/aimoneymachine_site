package social

import (
	"bytes"
<<<<<<< HEAD
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
=======
	"context"
	"encoding/json"
	"fmt"
<<<<<<< HEAD
	"github.com/dghubble/oauth1"
	"github.com/robertpelloni/hustle/orchestrator"
=======
>>>>>>> origin/main
	"net/http"
	"net/url"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/robertpelloni/hustle/orchestrator"
)

type SocialPost struct {
	ID          string
	Platform    string
	Content     string
	Author      string
	ScheduledAt time.Time
}

var twitterAPIEndpoint = "https://api.twitter.com/2/tweets"
var twitterSearchEndpoint = "https://api.twitter.com/2/tweets/search/recent"
var linkedInAPIEndpoint = "https://api.linkedin.com/v2/ugcPosts"

type Provider interface {
	Post(orch *orchestrator.Orchestrator, platform, content string) error
	Search(query string) ([]SocialPost, error)
	Reply(postID, content string) error
	SetDryRun(enabled bool)
}

type TwitterProvider struct {
<<<<<<< HEAD
=======
	DryRun       bool
>>>>>>> origin/main
	APIKey       string
	APISecret    string
	AccessToken  string
	AccessSecret string
<<<<<<< HEAD
}

func (p *TwitterProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.APIKey == "" || p.APISecret == "" || p.AccessToken == "" || p.AccessSecret == "" {
		fmt.Printf("[Twitter] Missing API keys. Mock posting to %s: %s\n", platform, content)
		return nil
	}

=======
	APIURL       string
}

func (p *TwitterProvider) SetDryRun(enabled bool) {
	p.DryRun = enabled
>>>>>>> origin/main
}

type twitterPostRequest struct {
	Text string `json:"text"`
}

func (p *TwitterProvider) Search(query string) ([]SocialPost, error) {
	if p.DryRun {
		return []SocialPost{{ID: "mock-id", Content: "Mock tweet about " + query}}, nil
	}

	config := oauth1.NewConfig(p.APIKey, p.APISecret)
	token := oauth1.NewToken(p.AccessToken, p.AccessSecret)
	httpClient := config.Client(context.Background(), token)

	url := fmt.Sprintf("%s?query=%s&max_results=10", twitterSearchEndpoint, url.QueryEscape(query))
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Data []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var posts []SocialPost
	for _, d := range data.Data {
		posts = append(posts, SocialPost{ID: d.ID, Content: d.Text, Platform: "Twitter"})
	}
	return posts, nil
}

func (p *TwitterProvider) Reply(postID, content string) error {
	if p.DryRun {
		fmt.Printf("[Twitter] DryRun: Replying to %s: %s\n", postID, content)
		return nil
	}

	config := oauth1.NewConfig(p.APIKey, p.APISecret)
	token := oauth1.NewToken(p.AccessToken, p.AccessSecret)
	httpClient := config.Client(context.Background(), token)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"text": content,
		"reply": map[string]string{
			"in_reply_to_tweet_id": postID,
		},
	})

	resp, err := httpClient.Post(twitterAPIEndpoint, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twitter reply error: %d", resp.StatusCode)
	}

	return nil
}

func (p *TwitterProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[Twitter] DryRun: Posting to %s: %s\n", platform, content)
		return nil
	}

<<<<<<< HEAD
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
		return errors.New("missing Twitter OAuth environment variables")
	}

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
=======
	if p.APIKey == "" || p.APISecret == "" || p.AccessToken == "" || p.AccessSecret == "" {
		return fmt.Errorf("missing Twitter OAuth environment variables")
	}

>>>>>>> origin/main
	config := oauth1.NewConfig(p.APIKey, p.APISecret)
	token := oauth1.NewToken(p.AccessToken, p.AccessSecret)
	httpClient := config.Client(context.Background(), token)
>>>>>>> origin/main

<<<<<<< HEAD
	payload := map[string]string{
		"text": content,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal twitter payload: %w", err)
	}

	resp, err := httpClient.Post(twitterAPIEndpoint, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to post to twitter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return fmt.Errorf("twitter API returned status %d", resp.StatusCode)
=======
	reqBody, err := json.Marshal(twitterPostRequest{Text: content})
	if err != nil {
		return fmt.Errorf("failed to marshal twitter request: %w", err)
	}

	apiURL := p.APIURL
	if apiURL == "" {
<<<<<<< HEAD
		apiURL = "https://api.twitter.com/2/tweets"
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
=======
		apiURL = twitterAPIEndpoint
	}

	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		resp, err = httpClient.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				break
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			return fmt.Errorf("twitter API error: received status code %d", resp.StatusCode)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to send request to Twitter API: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return fmt.Errorf("twitter API error: received status code %d (retries exhausted)", resp.StatusCode)
	}
	defer resp.Body.Close()

	fmt.Printf("[Twitter] Successfully posted to %s: %s\n", platform, content)
	return nil
}

<<<<<<< HEAD
func NewTwitterProvider() *TwitterProvider {
	return &TwitterProvider{
		DryRun: os.Getenv("TWITTER_DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "true",
		APIURL: "https://api.twitter.com/2/tweets",
=======
func NewTwitterProvider(apiKey, apiSecret, accessToken, accessSecret string) *TwitterProvider {
	return &TwitterProvider{
		APIKey:       apiKey,
		APISecret:    apiSecret,
		AccessToken:  accessToken,
		AccessSecret: accessSecret,
<<<<<<< HEAD
=======
		APIURL:       twitterAPIEndpoint,
>>>>>>> origin/main
	}
}

type LinkedInProvider struct {
<<<<<<< HEAD
	AccessToken string
	AuthorURN   string
}

func (p *LinkedInProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.AccessToken == "" || p.AuthorURN == "" {
		fmt.Printf("[LinkedIn] Missing API keys. Mock posting to %s: %s\n", platform, content)
		return nil
	}

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

func (p *LinkedInProvider) Search(query string) ([]SocialPost, error) {
	// LinkedIn search is highly restricted; returning stub
	return nil, fmt.Errorf("LinkedIn search not implemented")
}

func (p *LinkedInProvider) Reply(postID, content string) error {
	// LinkedIn comments API is distinct; returning stub
	return fmt.Errorf("LinkedIn reply not implemented")
}

func (p *LinkedInProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[LinkedIn] DryRun: Posting to %s: %s\n", platform, content)
		return nil
	}

	if p.AccessToken == "" || p.AuthorURN == "" {
		return fmt.Errorf("missing LINKEDIN_ACCESS_TOKEN or LINKEDIN_MEMBER_ID environment variable")
	}

	apiURL := p.APIURL
	if apiURL == "" {
		apiURL = linkedInAPIEndpoint
	}

	client := p.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

>>>>>>> origin/main
	payload := map[string]interface{}{
		"author":         p.AuthorURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
<<<<<<< HEAD
				"shareCommentary": map[string]string{
=======
				"shareCommentary": map[string]interface{}{
>>>>>>> origin/main
					"text": content,
				},
				"shareMediaCategory": "NONE",
			},
		},
<<<<<<< HEAD
		"visibility": map[string]string{
=======
		"visibility": map[string]interface{}{
>>>>>>> origin/main
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal linkedin payload: %w", err)
	}

<<<<<<< HEAD
	req, err := http.NewRequest("POST", linkedInAPIEndpoint, bytes.NewBuffer(payloadBytes))
=======
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
>>>>>>> origin/main
	if err != nil {
		return fmt.Errorf("failed to create linkedin request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.AccessToken)
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		resp, err = client.Do(req)
		if err == nil {
			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
				break
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			return fmt.Errorf("linkedin api returned status %d", resp.StatusCode)
		}
	}

	if err != nil {
		return fmt.Errorf("linkedin api request failed: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("linkedin api returned status %d (retries exhausted)", resp.StatusCode)
	}
	defer resp.Body.Close()

	fmt.Printf("[LinkedIn] Successfully posted to %s: %s\n", platform, content)
	return nil
}

func NewLinkedInProvider(accessToken, authorURN string) *LinkedInProvider {
	return &LinkedInProvider{
		AccessToken: accessToken,
		AuthorURN:   authorURN,
<<<<<<< HEAD
=======
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
