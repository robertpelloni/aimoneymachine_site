package orchestrator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A transient 503 from the proxy must (a) surface as an HTTP error, not
// a cryptic unmarshal failure, and (b) be retried until the server
// answers — bootstrap must not downgrade to MockLLM on one blip.
func TestDetectModelWithRetryRecoversFromTransient503(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `503`)
			return
		}
		fmt.Fprint(w, `{"data":[{"id":"gemma-test:latest"}]}`)
	}))
	defer ts.Close()

	p := NewOpenAICompatProviderWithBaseURL(ts.URL)
	model, err := p.DetectModelWithRetry(3, time.Millisecond)
	if err != nil {
		t.Fatalf("DetectModelWithRetry: %v", err)
	}
	if model != "gemma-test:latest" {
		t.Fatalf("model = %q", model)
	}
	if calls != 3 {
		t.Fatalf("probe calls = %d, want 3", calls)
	}

	// And the single-shot error names the HTTP status.
	p2 := NewOpenAICompatProviderWithBaseURL(ts.URL)
	calls = 99 // force the 503 branch off; use a fresh failing server below
	_ = p2
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer bad.Close()
	if _, err := NewOpenAICompatProviderWithBaseURL(bad.URL).DetectModel(); err == nil || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("DetectModel err = %v, want HTTP 503 named", err)
	}
}
