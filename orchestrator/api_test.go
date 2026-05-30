package orchestrator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicAPIClient(t *testing.T) {
	// Mock server to simulate Anthropic API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"content": [{"text": "Mocked Claude response"}]}`)
	}))
	defer server.Close()

	_ = &AnthropicProvider{APIKey: "test-key"}
	t.Logf("Mock server running at: %s", server.URL)
}

func TestTavilyAPIClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"results": [{"title": "Tavily Result", "url": "https://test.com", "content": "Tavily content"}]}`)
	}))
	defer server.Close()
	t.Logf("Mock Tavily server running at: %s", server.URL)
}
