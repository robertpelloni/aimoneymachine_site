package social

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

func TestLinkedInProvider_Post(t *testing.T) {
	mockOrch := &orchestrator.Orchestrator{}

	t.Run("Missing Credentials", func(t *testing.T) {
		p := NewLinkedInProvider("", "")
		err := p.Post(mockOrch, "LinkedIn", "Test content")
		if err == nil {
			t.Fatalf("expected error due to missing credentials, got nil")
		}
	})

	t.Run("Successful Post", func(t *testing.T) {
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
