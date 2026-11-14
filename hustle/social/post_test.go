package social

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
