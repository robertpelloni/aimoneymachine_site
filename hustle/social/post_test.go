package social

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

func TestLinkedInProvider_Post(t *testing.T) {
	mockOrch := &orchestrator.Orchestrator{}

	t.Run("Missing Env Vars", func(t *testing.T) {
		p := NewLinkedInProvider("", "")
		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err == nil {
			t.Fatalf("expected error due to missing credentials, got nil")
		}
	})

	t.Run("Successful Post", func(t *testing.T) {
		// Create a mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer mock-token" {
				t.Errorf("Expected Bearer mock-token")
			}
			if r.Header.Get("X-Restli-Protocol-Version") != "2.0.0" {
				t.Errorf("Expected X-Restli-Protocol-Version 2.0.0")
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id": "urn:li:share:12345"}`)
		}))
		defer server.Close()

		p := &LinkedInProvider{
			AccessToken: "mock-token",
			AuthorURN:   "urn:li:person:mock-member",
			HTTPClient:  server.Client(),
			APIURL:      server.URL,
		}

		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("API Error Response", func(t *testing.T) {
		// Create a mock HTTP server returning 401 Unauthorized
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message": "Unauthorized"}`)
		}))
		defer server.Close()

		p := &LinkedInProvider{
			AccessToken: "mock-token",
			AuthorURN:   "urn:li:person:mock-member",
			HTTPClient:  server.Client(),
			APIURL:      server.URL,
		}

		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err == nil {
			t.Fatalf("expected error from API response, got nil")
		}
	})
}

func TestTwitterProvider_DryRun(t *testing.T) {
	provider := &TwitterProvider{
		DryRun: true,
	}

	orch := orchestrator.NewOrchestrator()
	err := provider.Post(orch, "Twitter", "Test Content")
	if err != nil {
		t.Fatalf("expected no error during dry run, got: %v", err)
	}
}

func TestTwitterProvider_MissingEnv(t *testing.T) {
	provider := NewTwitterProvider("", "", "", "")

	orch := orchestrator.NewOrchestrator()
	err := provider.Post(orch, "Twitter", "Test Content")
	if err == nil {
		t.Fatalf("expected error due to missing env variables, got nil")
	}
}

func TestTwitterProvider_PostSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var reqData twitterPostRequest
		json.Unmarshal(body, &reqData)
		if reqData.Text != "Test Content" {
			t.Errorf("expected text 'Test Content', got '%s'", reqData.Text)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	provider := &TwitterProvider{
		APIKey:       "test_key",
		APISecret:    "test_secret",
		AccessToken:  "test_token",
		AccessSecret: "test_token_secret",
		DryRun:       false,
		APIURL:       ts.URL,
	}

	orch := orchestrator.NewOrchestrator()
	err := provider.Post(orch, "Twitter", "Test Content")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

// Shapes observed leaking into real prod posts (affiliate, v1.9.9):
// prompt echo at the start, model reasoning appended at the end.
func TestSanitizeSocialContent(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{
			"leading echo stripped",
			"Here is a revolutionary post about AI tools. AI agents are eating SaaS alive. #AI #hustle",
			"AI agents are eating SaaS alive. #AI #hustle",
		},
		{
			"user-wants echo stripped",
			"The user wants a short post about crypto. Bitcoin is eating the banks. #crypto",
			"Bitcoin is eating the banks. #crypto",
		},
		{
			"trailing reasoning cut",
			"AI agents are eating SaaS! #AI #hustle. But does it mention it's revolutionary? The user said revolutionary, so maybe it should.",
			"AI agents are eating SaaS! #AI #hustle.",
		},
		{
			"fences stripped",
			"```\nAI tools are changing everything fast. #AI\n```",
			"AI tools are changing everything fast. #AI",
		},
		{
			"clean copy untouched",
			"AI agents are eating SaaS alive. #AI #hustle",
			"AI agents are eating SaaS alive. #AI #hustle",
		},
	}
	for _, c := range cases {
		if got := sanitizeSocialContent(c.in); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
	// A lone echo sentence has no copy to rescue: it must stay marked
	// as echo so GenerateContent's validation rejects it to fallback.
	if got := sanitizeSocialContent("Create a short, revolutionary social media post about AI with hashtags."); !hasAnyPrefix(strings.ToLower(got), echoOpeners) {
		t.Errorf("lone echo not detectable: %q", got)
	}
}

func TestGenerateContentSanitizesEchoAndReasoning(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	orch.LLM = &orchestrator.MockLLM{GenerateFunc: func(string) (string, error) {
		return "Here is a revolutionary post about AI automation. AI agents are replacing busywork everywhere. #AI #hustle. But does it mention it's revolutionary enough?", nil
	}}
	got := GenerateContent(orch, "AI automation")
	want := "AI agents are replacing busywork everywhere. #AI #hustle."
	if got != want {
		t.Errorf("GenerateContent = %q, want %q", got, want)
	}
	// Pure junk still falls back.
	orch.LLM = &orchestrator.MockLLM{GenerateFunc: func(string) (string, error) {
		return "Create a short, revolutionary social media post about AI automation with hashtags.", nil
	}}
	if got := GenerateContent(orch, "AI automation"); !strings.Contains(got, "Revolutionary insights on AI automation!") {
		t.Errorf("echo-only output should fall back, got %q", got)
	}
}

// LinkedIn tokens expire under a long-lived daemon; a 401 with refresh
// material present must refresh once and retry with the new token.
func TestLinkedInRefreshOn401ThenSucceeds(t *testing.T) {
	var apiCalls int
	var secondAuth string
	var tokenForm url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ugc":
			apiCalls++
			if apiCalls == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"message":"expired"}`)
				return
			}
			secondAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusCreated)
		case "/token":
			_ = r.ParseForm()
			tokenForm = r.PostForm
			fmt.Fprint(w, `{"access_token":"new-token"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	p := &LinkedInProvider{
		AccessToken: "old-token", AuthorURN: "urn:li:person:1",
		APIURL: ts.URL + "/ugc", TokenURL: ts.URL + "/token",
		RefreshToken: "rt-1", ClientID: "cid", ClientSecret: "sec",
		HTTPClient: ts.Client(),
	}
	if err := p.Post(orchestrator.NewOrchestrator(), "LinkedIn", "hello"); err != nil {
		t.Fatalf("Post after refresh: %v", err)
	}
	if apiCalls != 2 {
		t.Errorf("API calls = %d, want 2 (401 + retry)", apiCalls)
	}
	if secondAuth != "Bearer new-token" {
		t.Errorf("retry auth = %q, want Bearer new-token", secondAuth)
	}
	if tokenForm.Get("grant_type") != "refresh_token" || tokenForm.Get("refresh_token") != "rt-1" {
		t.Errorf("token form wrong: %v", tokenForm)
	}
}

// Regression: LinkedIn token responses exceed 500 bytes. The first
// prod refresh actually succeeded but the 500-byte read truncated the
// JSON and the parse rejected it. A >500-byte response must parse.
func TestLinkedInRefreshParsesLongTokenResponse(t *testing.T) {
	longToken := strings.Repeat("A", 400)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ugc":
			if r.Header.Get("Authorization") == "Bearer "+longToken {
				w.WriteHeader(http.StatusCreated)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		case "/token":
			fmt.Fprintf(w, `{"access_token":%q,"expires_in":5183999,"refresh_token":%q}`,
				longToken, strings.Repeat("R", 300))
		}
	}))
	defer ts.Close()
	p := &LinkedInProvider{
		AccessToken: "old", AuthorURN: "urn:li:person:1",
		APIURL: ts.URL + "/ugc", TokenURL: ts.URL + "/token",
		RefreshToken: "rt", ClientID: "cid", ClientSecret: "sec",
		HTTPClient: ts.Client(),
	}
	if err := p.Post(orchestrator.NewOrchestrator(), "LinkedIn", "hello"); err != nil {
		t.Fatalf("Post with long token response: %v", err)
	}
	if p.AccessToken != longToken {
		t.Errorf("adopted token length = %d, want %d", len(p.AccessToken), len(longToken))
	}
}

// Refreshed tokens (including a rotated refresh token) must persist
// so a daemon restart starts fresh instead of paying another 401.
func TestLinkedInTokenPersistenceAcrossRestart(t *testing.T) {
	tokenFile := filepath.Join(t.TempDir(), "li_tokens.json")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ugc":
			if r.Header.Get("Authorization") == "Bearer fresh-token" {
				w.WriteHeader(http.StatusCreated)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		case "/token":
			fmt.Fprint(w, `{"access_token":"fresh-token","refresh_token":"rotated-rt"}`)
		}
	}))
	defer ts.Close()

	p1 := &LinkedInProvider{
		AccessToken: "expired", AuthorURN: "urn:li:person:1",
		APIURL: ts.URL + "/ugc", TokenURL: ts.URL + "/token",
		RefreshToken: "old-rt", ClientID: "cid", ClientSecret: "sec",
		HTTPClient: ts.Client(), TokenFile: tokenFile,
	}
	if err := p1.Post(orchestrator.NewOrchestrator(), "LinkedIn", "hello"); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if _, err := os.Stat(tokenFile); err != nil {
		t.Fatalf("token file not persisted: %v", err)
	}

	// Simulated restart: env still holds the expired pair.
	p2 := &LinkedInProvider{
		AccessToken: "expired", RefreshToken: "old-rt", TokenFile: tokenFile,
	}
	if !p2.loadStoredTokens() {
		t.Fatal("stored tokens not loaded")
	}
	if p2.AccessToken != "fresh-token" {
		t.Errorf("access token = %q, want fresh-token", p2.AccessToken)
	}
	if p2.RefreshToken != "rotated-rt" {
		t.Errorf("refresh token = %q, want rotated-rt (rotation adopted)", p2.RefreshToken)
	}
}

// A 422 DUPLICATE_POST means the content is already live: Post must
// return the ErrDuplicatePost sentinel (not a generic failure) so
// callers can treat distribution as achieved.
func TestLinkedInDuplicatePostSentinel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"message":"com.linkedin.content.common.exception.BadRequestResponseException: Content is a duplicate","errorDetails":{"code":"DUPLICATE_POST"},"status":422}`)
	}))
	defer ts.Close()
	p := &LinkedInProvider{
		AccessToken: "tok", AuthorURN: "urn:li:person:1",
		APIURL: ts.URL, HTTPClient: ts.Client(),
	}
	err := p.Post(orchestrator.NewOrchestrator(), "LinkedIn", "same text as before")
	if !errors.Is(err, ErrDuplicatePost) {
		t.Fatalf("err = %v, want ErrDuplicatePost", err)
	}
}

// Without refresh material a 401 must surface as an error naming the
// status (and body), not hang or silently pass.
func TestLinkedIn401WithoutRefreshFailsClearly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"bad token"}`)
	}))
	defer ts.Close()
	p := &LinkedInProvider{
		AccessToken: "old-token", AuthorURN: "urn:li:person:1",
		APIURL: ts.URL, HTTPClient: ts.Client(),
	}
	err := p.Post(orchestrator.NewOrchestrator(), "LinkedIn", "hello")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v, want 401 surfaced", err)
	}
}

// The constructor computed env fallbacks into locals but returned the
// raw (empty) params — any caller passing "" got a broken provider.
func TestNewLinkedInProviderEnvFallback(t *testing.T) {
	t.Setenv("LINKEDIN_ACCESS_TOKEN", "env-token")
	t.Setenv("LINKEDIN_AUTHOR_URN", "urn:li:person:env")
	t.Setenv("LINKEDIN_REFRESH_TOKEN", "env-rt")
	p := NewLinkedInProvider("", "")
	if p.AccessToken != "env-token" || p.AuthorURN != "urn:li:person:env" {
		t.Errorf("env fallback lost: token=%q urn=%q", p.AccessToken, p.AuthorURN)
	}
	if p.RefreshToken != "env-rt" {
		t.Errorf("refresh token = %q, want env-rt", p.RefreshToken)
	}
}

func TestDefaultPlatformAndProviderSelection(t *testing.T) {
	t.Setenv("SOCIAL_DEFAULT_PLATFORM", "")
	if got := DefaultPlatform(); got != "Twitter" {
		t.Errorf("default platform = %q, want Twitter", got)
	}
	t.Setenv("SOCIAL_DEFAULT_PLATFORM", "Bluesky")
	if got := DefaultPlatform(); got != "Bluesky" {
		t.Errorf("env platform = %q, want Bluesky", got)
	}
	if _, ok := NewProvider("Bluesky").(*BlueskyProvider); !ok {
		t.Error("NewProvider(Bluesky) did not return *BlueskyProvider")
	}
	if _, ok := NewProvider("Twitter").(*TwitterProvider); !ok {
		t.Error("NewProvider(Twitter) did not return *TwitterProvider")
	}
	if _, ok := NewProvider("LinkedIn").(*LinkedInProvider); !ok {
		t.Error("NewProvider(LinkedIn) did not return *LinkedInProvider")
	}
}

// Bluesky rejects posts over 300 graphemes (XRPC InvalidRequest); the
// formatter must cap it the way it already capped Twitter at 280.
func TestFormatForPlatformLimits(t *testing.T) {
	long := strings.Repeat("a", 400)
	if got := FormatForPlatform(long, "Bluesky"); len([]rune(got)) != 300 {
		t.Errorf("Bluesky len = %d runes, want 300", len([]rune(got)))
	}
	if got := FormatForPlatform(long, "Twitter"); len([]rune(got)) != 280 {
		t.Errorf("Twitter len = %d runes, want 280", len([]rune(got)))
	}
	short := "hello"
	if got := FormatForPlatform(short, "Bluesky"); got != short {
		t.Errorf("short Bluesky post modified: %q", got)
	}
	if got := FormatForPlatform(long, "LinkedIn"); got != long {
		t.Error("LinkedIn must not be truncated at these limits")
	}
}
