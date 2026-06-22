package social

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTwitterProviderIntegration(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("Expected path '/', got '%s'", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode payload: %v", err)
		}

		if payload["text"] != "Revolutionary Test Content!" {
			t.Errorf("Expected text 'Revolutionary Test Content!', got '%s'", payload["text"])
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	// Override the endpoint
	oldEndpoint := twitterAPIEndpoint
	twitterAPIEndpoint = mockServer.URL
	defer func() { twitterAPIEndpoint = oldEndpoint }()

	provider := NewTwitterProvider("key", "secret", "token", "tsecret")
	err := provider.Post(nil, "Twitter", "Revolutionary Test Content!")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLinkedInProviderIntegration(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Errorf("Expected path '/', got '%s'", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-access-token" {
			t.Errorf("Expected auth header 'Bearer test-access-token', got '%s'", authHeader)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer mockServer.Close()

	oldEndpoint := linkedInAPIEndpoint
	linkedInAPIEndpoint = mockServer.URL
	defer func() { linkedInAPIEndpoint = oldEndpoint }()

	provider := NewLinkedInProvider("test-access-token", "urn:li:person:test")
	err := provider.Post(nil, "LinkedIn", "Revolutionary Test Content!")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
