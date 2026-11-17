package social

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
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

// --- TwitterProvider ---
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

	var httpClient *http.Client

	if p.APIKey != "" && p.APISecret != "" && p.AccessToken != "" && p.AccessSecret != "" {
		config := oauth1.NewConfig(p.APIKey, p.APISecret)
		token := oauth1.NewToken(p.AccessToken, p.AccessSecret)
		httpClient = config.Client(context.Background(), token)
	} else {
		return fmt.Errorf("missing Twitter OAuth environment variables")
	}

	reqBody, err := json.Marshal(twitterPostRequest{Text: content})
	if err != nil {
		return fmt.Errorf("failed to marshal twitter request: %w", err)
	}

	apiURL := p.APIURL
	if apiURL == "" {
		apiURL = twitterAPIEndpoint
	}

	// Retry up to 3 times with backoff
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		req, reqErr := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
		if reqErr != nil {
		    return fmt.Errorf("failed to create request: %w", reqErr)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = httpClient.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				break
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				continue
			}
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("twitter API error: received status code %d. Full Response bytes: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	if err != nil {
		return fmt.Errorf("failed to send request to Twitter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twitter API error: received status code %d (retries exhausted). Response: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Printf("[Twitter] Successfully posted to %s: %s\n", platform, content)
	return nil
}

func NewTwitterProvider(apiKey, apiSecret, accessToken, accessSecret string) *TwitterProvider {
	return &TwitterProvider{
		DryRun:       os.Getenv("TWITTER_DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "true",
		APIKey:       apiKey,
		APISecret:    apiSecret,
		AccessToken:  accessToken,
		AccessSecret: accessSecret,
		APIURL:       twitterAPIEndpoint,
	}
}

// --- LinkedInProvider ---
type LinkedInProvider struct {
	DryRun      bool
	AccessToken string
	AuthorURN   string
	HTTPClient  *http.Client
	APIURL      string
	// Refresh material for the OAuth2 refresh flow. LinkedIn access
	// tokens expire (~60 days); the daemon outlives them, which showed
	// up as recurring 401s killing discovered chains at their last
	// step. On a 401, Post exchanges the refresh token once and retries.
	RefreshToken string
	ClientID     string
	ClientSecret string
	TokenURL     string
	// TokenFile, when set, persists refreshed tokens (0600, atomic) so
	// a daemon restart starts from the fresh token instead of paying
	// a 401 â†’ refresh cycle, and survives LinkedIn rotating the
	// refresh token itself. Env: LINKEDIN_TOKEN_FILE (default
	// .linkedin_token.json in the working directory).
	TokenFile string
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

	payload := map[string]interface{}{
		"author":         p.AuthorURN,
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
		return fmt.Errorf("failed to marshal linkedin payload: %w", err)
	}

	status, body, err := p.doPost(client, apiURL, payloadBytes)
	if err != nil {
		return err
	}

	// Expired access token: one refresh + one retry, in-process. The
	// daemon is long-lived and LinkedIn tokens live ~60 days, so
	// without this every chain ending in a LinkedIn step dies at 401
	// until a human rotates the token by hand.
	if status == http.StatusUnauthorized && p.RefreshToken != "" && p.ClientID != "" && p.ClientSecret != "" {
		fmt.Println("[LinkedIn] 401 received; refreshing access token")
		if rerr := p.refreshAccessToken(client); rerr != nil {
			return fmt.Errorf("linkedin token refresh failed: %w (original: status %d: %s)", rerr, status, body)
		}
		status, body, err = p.doPost(client, apiURL, payloadBytes)
		if err != nil {
			return err
		}
	}

	if status == http.StatusUnprocessableEntity && strings.Contains(body, "DUPLICATE_POST") {
		fmt.Printf("[LinkedIn] Duplicate of an already-published post; treating as success\n")
		return ErrDuplicatePost
	}
	if status != http.StatusCreated && status != http.StatusOK {
		return fmt.Errorf("linkedin api returned status %d: %s", status, body)
	}

	fmt.Printf("[LinkedIn] Successfully posted to %s: %s\n", platform, content)
	return nil
}

// ErrDuplicatePost means LinkedIn rejected the post as a duplicate of
// one already published (422 DUPLICATE_POST). The distribution goal is
// already met, so callers treat this as success, not failure â€”
// identical generated text for a repeated topic otherwise kills chains
// at their final step (observed live 2026-11-16).
var ErrDuplicatePost = errors.New("linkedin: duplicate post (already published)")

// doPost performs one UGC post attempt and returns (status, body, err).
func (p *LinkedInProvider) doPost(client *http.Client, apiURL string, payloadBytes []byte) (int, string, error) {
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return 0, "", fmt.Errorf("failed to create linkedin request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.AccessToken)
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("linkedin api request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
	return resp.StatusCode, strings.TrimSpace(string(respBody)), nil
}

// refreshAccessToken exchanges LINKEDIN_REFRESH_TOKEN for a new access
// token and adopts it for this process. (Persisting it back to the env
// file is deliberately not done from inside the daemon.)
func (p *LinkedInProvider) refreshAccessToken(client *http.Client) error {
	tokenURL := p.TokenURL
	if tokenURL == "" {
		tokenURL = "https://www.linkedin.com/oauth/v2/accessToken"
	}
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {p.RefreshToken},
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
	}
	resp, err := client.PostForm(tokenURL, form)
	if err != nil {
		return fmt.Errorf("token endpoint request: %w", err)
	}
	defer resp.Body.Close()
	// Token responses run ~700 bytes (two long tokens + expiry): the
	// original 500-byte cap truncated the JSON mid-token and the parse
	// then "failed" on a perfectly good refresh. Never echo this body
	// into errors/logs â€” it IS credential material.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token endpoint status %d", resp.StatusCode)
	}
	var parsed struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.AccessToken == "" {
		return fmt.Errorf("token endpoint returned %d without a usable access_token", resp.StatusCode)
	}
	p.AccessToken = parsed.AccessToken
	// LinkedIn may rotate the refresh token; adopt the new one or the
	// next refresh (after a restart) would use a dead token.
	if parsed.RefreshToken != "" {
		p.RefreshToken = parsed.RefreshToken
	}
	fmt.Println("[LinkedIn] Access token refreshed")
	p.persistTokens()
	return nil
}

// storedTokens is the on-disk shape of the persisted LinkedIn tokens.
type storedTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// persistTokens writes the current tokens to TokenFile (0600, atomic).
// Best-effort: a failed write is logged, never fatal to the post.
func (p *LinkedInProvider) persistTokens() {
	if p.TokenFile == "" {
		return
	}
	data, err := json.Marshal(storedTokens{AccessToken: p.AccessToken, RefreshToken: p.RefreshToken})
	if err != nil {
		return
	}
	if err := orchestrator.WriteFileAtomic(p.TokenFile, data, 0600); err != nil {
		fmt.Printf("[LinkedIn] Could not persist refreshed tokens: %v\n", err)
	}
}

// loadStoredTokens reads TokenFile into the provider when present.
// Returns true when a stored access token was adopted.
func (p *LinkedInProvider) loadStoredTokens() bool {
	if p.TokenFile == "" {
		return false
	}
	data, err := os.ReadFile(p.TokenFile)
	if err != nil {
		return false
	}
	var stored storedTokens
	if err := json.Unmarshal(data, &stored); err != nil || stored.AccessToken == "" {
		return false
	}
	p.AccessToken = stored.AccessToken
	if stored.RefreshToken != "" {
		p.RefreshToken = stored.RefreshToken
	}
	return true
}

func NewLinkedInProvider(accessToken, authorURN string) *LinkedInProvider {
	token := accessToken
	urn := authorURN
	if token == "" {
		token = os.Getenv("LINKEDIN_ACCESS_TOKEN")
	}
	if urn == "" {
		urn = os.Getenv("LINKEDIN_AUTHOR_URN")
		if urn == "" {
			urn = os.Getenv("LINKEDIN_MEMBER_ID")
		}
	}
	tokenFile := os.Getenv("LINKEDIN_TOKEN_FILE")
	if tokenFile == "" {
		tokenFile = ".linkedin_token.json"
	}
	p := &LinkedInProvider{
		AccessToken:  token,
		AuthorURN:    urn,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		APIURL:       linkedInAPIEndpoint,
		RefreshToken: os.Getenv("LINKEDIN_REFRESH_TOKEN"),
		ClientID:     os.Getenv("LINKEDIN_CLIENT_ID"),
		ClientSecret: os.Getenv("LINKEDIN_CLIENT_SECRET"),
		TokenFile:    tokenFile,
	}
	// A previously refreshed (and persisted) token beats the env one:
	// the env access token is expired by the time this matters.
	if p.loadStoredTokens() {
		fmt.Println("[LinkedIn] Loaded persisted refreshed token")
	}
	return p
}

// metaSentencePrefixes mark model scaffolding, not post copy: thinking
// about the prompt, checking its own work, echoing instructions. Any
// trailing run of such sentences is reasoning, never publishable text.
var metaSentencePrefixes = []string{
	"the user wants", "the user asked", "the user said", "the prompt ",
	"but does it", "does it mention", "is that clear", "should i ",
	"let me ", "note:", "i should ", "maybe i should", "this post should",
	"as an ai",
}

// echoOpeners are instruction-echo first sentences ("Here is a post
// about X. <actual post>") that precede the real copy.
var echoOpeners = []string{
	"the user wants", "the user asked", "here is a", "here's a",
	"here is the", "here's the", "here is your", "here's your",
	"sure! here", "certainly! here", "create a ",
}

func splitSentences(s string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range s {
		cur.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			if t := strings.TrimSpace(cur.String()); t != "" {
				out = append(out, t)
			}
			cur.Reset()
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" {
		out = append(out, t)
	}
	return out
}

func hasAnyPrefix(lower string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// sanitizeSocialContent makes raw LLM output safe for a public post,
// mirroring the affiliate sanitizer (v1.9.9): strip code fences and
// wrapping quotes, drop a leading instruction-echo sentence, and cut
// any trailing run of meta/reasoning sentences. It never rewrites the
// copy itself; anything unrecognizable passes through for the caller's
// junk validation to reject.
func sanitizeSocialContent(content string) string {
	s := strings.TrimSpace(content)
	s = strings.TrimPrefix(s, "```markdown")
	s = strings.TrimPrefix(s, "```md")
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}

	sentences := splitSentences(s)
	// Leading echo sentence, only when real copy follows it.
	if len(sentences) > 1 && hasAnyPrefix(strings.ToLower(sentences[0]), echoOpeners) {
		sentences = sentences[1:]
	}
	// Trailing meta/reasoning sentences, cut as a run from the end.
	for len(sentences) > 1 {
		last := strings.ToLower(sentences[len(sentences)-1])
		if !hasAnyPrefix(last, metaSentencePrefixes) {
			break
		}
		sentences = sentences[:len(sentences)-1]
	}
	return strings.Join(sentences, " ")
}

// placeholderMarkers match bracketed fill-in text ("[Insert your bio
// here]") that template-style LLM output ships verbatim. A post went to
// LinkedIn carrying three of these (2026-11-17); they are never valid
// finished copy.
var placeholderMarkers = []string{
	"[insert", "[your", "[name]", "[handle]", "[bio]", "[link]",
	"[placeholder", "[todo", "[company name", "[website",
}

// ContainsPlaceholder reports whether copy carries bracketed fill-in
// text. Exported for the other public-posting pipelines (affiliate)
// that do not pass through GenerateContent.
func ContainsPlaceholder(content string) bool {
	lower := strings.ToLower(content)
	for _, m := range placeholderMarkers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func GenerateContent(orch *orchestrator.Orchestrator, topic string) string {
	prompt := fmt.Sprintf("Create a short, revolutionary social media post about %s with hashtags. Write finished copy only â€” never include bracketed placeholders such as [insert ...] or [your ...].", topic)
	raw, err := orch.LLM.Generate(prompt)
	content := sanitizeSocialContent(raw)

	// Bracketed placeholders get ONE corrective regeneration; if the
	// retry still has them, the validity check below falls back to the
	// canned line instead of posting fill-in text publicly.
	if err == nil && ContainsPlaceholder(content) {
		fmt.Printf("[Social] Placeholder text in generated post; regenerating once\n")
		raw2, err2 := orch.LLM.Generate(prompt + " The previous attempt contained bracketed placeholder text; write the complete final post with real words everywhere a placeholder appeared.")
		if err2 == nil {
			if c2 := sanitizeSocialContent(raw2); !ContainsPlaceholder(c2) {
				content = c2
			}
		}
	}

	// Validate LLM output â€” reject junk: metadata, JSON, short garbage.
	// (Prompt echo and reasoning are already stripped by the sanitizer;
	// anything still matching these markers is unrecoverable.)
	if err != nil || content == "" || len(content) < 20 || ContainsPlaceholder(content) ||
		hasAnyPrefix(strings.ToLower(content), echoOpeners) ||
		strings.Contains(content, "[Model:") ||
		strings.Contains(content, "Generated content") ||
		strings.Contains(content, "```json") ||
		strings.Contains(content, "```") ||
		strings.Contains(content, "gt-oss") ||
		strings.Contains(content, "[assistant") ||
		strings.Contains(content, "<|assistant|") {
		fallback := fmt.Sprintf("Revolutionary insights on %s! #hustle #ai", topic)
		if err != nil {
			fmt.Printf("[Social] LLM error, using fallback: %v\n", err)
		} else {
			fmt.Printf("[Social] LLM returned invalid content, using fallback: %q\n", content)
		}
		return fallback
	}

	return content
}

// ValidatePlatform rejects platform names with no provider. Discovered
// chains invent platforms (observed live: platform=Pinterest), and the
// handler's switch silently routed every unknown name to Twitter â€”
// where the post died on X's 402 instead of failing where it belonged.
func ValidatePlatform(platform string) error {
	switch strings.ToLower(platform) {
	case "twitter", "linkedin", "bluesky":
		return nil
	default:
		return fmt.Errorf("unsupported social platform %q (supported: Twitter, LinkedIn, Bluesky)", platform)
	}
}

// DefaultPlatform is the social platform used by pipelines that post on
// the user's behalf (affiliate, synergy lead-gen). SOCIAL_DEFAULT_PLATFORM
// overrides it; default Twitter for historical compatibility. Prod sets
// Bluesky while X returns 402 (credits depleted).
func DefaultPlatform() string {
	if v := os.Getenv("SOCIAL_DEFAULT_PLATFORM"); v != "" {
		return v
	}
	return "Twitter"
}

// NewProvider builds the provider for a platform name, reading that
// platform's credentials from the environment.
func NewProvider(platform string) Provider {
	switch strings.ToLower(platform) {
	case "linkedin":
		return NewLinkedInProvider(os.Getenv("LINKEDIN_ACCESS_TOKEN"), os.Getenv("LINKEDIN_AUTHOR_URN"))
	case "bluesky":
		return NewBlueskyProvider(os.Getenv("BLUESKY_HANDLE"), os.Getenv("BLUESKY_APP_PASSWORD"))
	default:
		return NewTwitterProvider(
			os.Getenv("TWITTER_API_KEY"),
			os.Getenv("TWITTER_API_SECRET"),
			os.Getenv("TWITTER_ACCESS_TOKEN"),
			os.Getenv("TWITTER_ACCESS_SECRET"),
		)
	}
}

func FormatForPlatform(content, platform string) string {
	// Per-platform length caps. Bluesky rejects records over 300
	// graphemes with XRPC InvalidRequest â€” the curation chain's
	// affiliate-injected posts exceeded it because only Twitter was
	// truncated here.
	limit := 0
	switch strings.ToLower(platform) {
	case "twitter":
		limit = 280
	case "bluesky":
		limit = 300
	}
	if limit > 0 {
		runes := []rune(content)
		if len(runes) > limit {
			return string(runes[:limit-3]) + "..."
		}
	}
	return content
}


// recordIntegrationHealth feeds the /api/status integration dots:
// success marks the platform ok; a 401/402-class rejection marks it
// auth_failed so a configured-but-dead integration stops showing
// green (Twitter did, through weeks of X 402 responses).
func recordIntegrationHealth(orch *orchestrator.Orchestrator, platform string, err error) {
	if err == nil {
		orch.SetIntegrationHealth(platform, "ok")
		return
	}
	s := err.Error()
	if strings.Contains(s, "401") || strings.Contains(s, "402") {
		orch.SetIntegrationHealth(platform, "auth_failed")
	}
}

func SchedulePost(orch *orchestrator.Orchestrator, provider Provider, platform, topic string) {
	content := GenerateContent(orch, topic)
	content = FormatForPlatform(content, platform)
	fmt.Printf("Scheduling post for %s: %s\n", platform, content)

	err := provider.Post(orch, platform, content)
	recordIntegrationHealth(orch, platform, err)
	if errors.Is(err, ErrDuplicatePost) {
		// Already live from an earlier identical generation: goal met,
		// no new ledger charge (the original post was charged then).
		fmt.Printf("[Social] Already posted to %s (duplicate suppressed)\n", platform)
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("social-%s-dup-%d", platform, time.Now().Unix()),
			Content:   fmt.Sprintf("Duplicate post suppressed on %s: %s", platform, content),
			Timestamp: time.Now(),
			Tags:      []string{"social", platform, "duplicate"},
		})
	} else if err == nil {
		fmt.Printf("[Social] âœ… Successfully posted to %s\n", platform)
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
	} else {
		fmt.Printf("[Social] âŒ Failed to post to %s: %v\n", platform, err)
	}
}

