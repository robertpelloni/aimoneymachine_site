package publisher

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Buttondown's POST /v1/emails rejects unknown fields with 422
// extra_forbidden (observed on prod: loc body.payload.tags). The create
// payload must carry only subject/body — tags are not accepted here.
func TestButtondownPublishOmitsTags(t *testing.T) {
	var got map[string]interface{}
	bp := &ButtondownPublisher{
		APIKey: "test-key",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			data, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("request body is not JSON: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(`{"id":"e1","subject":"S","status":"draft"}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}
	if err := bp.Publish("S", "body text", []string{"ai", "automation"}); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if _, ok := got["tags"]; ok {
		t.Errorf("payload contains forbidden tags field: %v", got)
	}
	if got["subject"] != "S" || got["body"] != "body text" {
		t.Errorf("payload = %v, want subject+body preserved", got)
	}
}
