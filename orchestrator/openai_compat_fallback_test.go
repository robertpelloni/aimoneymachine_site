package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The fallback provider must not inherit the primary's model name:
// Ollama rejected 'gpt-4o' (from LLM_MODEL) on the first prod waterfall
// run. WithBaseURL clears Model so Generate re-detects against the
// fallback's own /models (skipping embedding models).
func TestWithBaseURLClearsModelAndRedetects(t *testing.T) {
	t.Setenv("LLM_MODEL", "gpt-4o")
	var askedModel string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/models"):
			fmt.Fprint(w, `{"data":[{"id":"nomic-embed-text:latest"},{"id":"gemma-test:latest"}]}`)
		case strings.HasSuffix(r.URL.Path, "/chat/completions"):
			var body struct {
				Model string `json:"model"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			askedModel = body.Model
			fmt.Fprint(w, `{"choices":[{"message":{"role":"assistant","content":"ok from fallback"}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	p := NewOpenAICompatProviderWithBaseURL(ts.URL)
	if p.Model != "" {
		t.Fatalf("Model = %q, want cleared for re-detection", p.Model)
	}
	got, err := p.Generate("hello")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "ok from fallback" {
		t.Fatalf("Generate = %q", got)
	}
	if askedModel != "gemma-test:latest" {
		t.Errorf("asked model = %q, want gemma-test:latest (non-embedding)", askedModel)
	}
}

// Ollama binds the IPv4 loopback only; a "localhost" fallback URL can
// resolve to [::1] first and the refused dial surfaced as a fatal
// DetectModel failure on prod (CurationChain, 2026-11-17) while
// 127.0.0.1 answered fine. The fallback constructor normalizes the
// host; non-loopback URLs pass through untouched.
func TestWithBaseURLNormalizesLoopback(t *testing.T) {
	if got := NewOpenAICompatProviderWithBaseURL("http://localhost:11434/v1").BaseURL; got != "http://127.0.0.1:11434/v1" {
		t.Errorf("BaseURL = %q, want http://127.0.0.1:11434/v1", got)
	}
	if got := NewOpenAICompatProviderWithBaseURL("http://example.com:11434/v1").BaseURL; got != "http://example.com:11434/v1" {
		t.Errorf("BaseURL = %q, want non-loopback untouched", got)
	}
}
