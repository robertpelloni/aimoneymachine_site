package social

import (
<<<<<<< HEAD
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
=======
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
>>>>>>> origin/main
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

func TestLinkedInProvider_Post(t *testing.T) {
	// Helper function to mock orchestrator
	mockOrch := &orchestrator.Orchestrator{}

	t.Run("Missing Env Vars", func(t *testing.T) {
<<<<<<< HEAD
		os.Unsetenv("LINKEDIN_ACCESS_TOKEN")
		os.Unsetenv("LINKEDIN_MEMBER_ID")

		p := NewLinkedInProvider()
=======
		p := NewLinkedInProvider("", "")
>>>>>>> origin/main
		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err == nil {
			t.Fatalf("expected error due to missing env vars, got nil")
		}
	})

	t.Run("Successful Post", func(t *testing.T) {
<<<<<<< HEAD
		t.Setenv("LINKEDIN_ACCESS_TOKEN", "mock-token")
		t.Setenv("LINKEDIN_MEMBER_ID", "mock-member-id")

=======
>>>>>>> origin/main
		// Create a mock HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer mock-token" {
				t.Errorf("Expected Bearer mock-token, got %s", authHeader)
			}
			protocolVersion := r.Header.Get("X-Restli-Protocol-Version")
			if protocolVersion != "2.0.0" {
				t.Errorf("Expected X-Restli-Protocol-Version 2.0.0, got %s", protocolVersion)
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id": "urn:li:share:12345"}`)
		}))
		defer server.Close()

		p := &LinkedInProvider{
<<<<<<< HEAD
			HTTPClient: server.Client(),
			APIURL:     server.URL,
=======
			AccessToken: "mock-token",
			AuthorURN:   "mock-member-id",
			HTTPClient:  server.Client(),
			APIURL:      server.URL,
>>>>>>> origin/main
		}

		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("API Error Response", func(t *testing.T) {
<<<<<<< HEAD
		t.Setenv("LINKEDIN_ACCESS_TOKEN", "mock-token")
		t.Setenv("LINKEDIN_MEMBER_ID", "mock-member-id")

=======
>>>>>>> origin/main
		// Create a mock HTTP server returning 401 Unauthorized
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message": "Unauthorized"}`)
		}))
		defer server.Close()

		p := &LinkedInProvider{
<<<<<<< HEAD
			HTTPClient: server.Client(),
			APIURL:     server.URL,
=======
			AccessToken: "mock-token",
			AuthorURN:   "mock-member-id",
			HTTPClient:  server.Client(),
			APIURL:      server.URL,
>>>>>>> origin/main
		}

		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err == nil {
			t.Fatalf("expected error from API response, got nil")
		}
	})
}
<<<<<<< HEAD
=======

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
>>>>>>> origin/main
