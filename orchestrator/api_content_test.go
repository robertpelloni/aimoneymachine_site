package orchestrator

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// Missing output/content is a normal empty-library state: the dashboard
// must get 200 + [], not a 500 it renders as a broken panel.
func TestHandleContentMissingDirReturnsEmptyList(t *testing.T) {
	t.Chdir(t.TempDir()) // no output/content here
	rec := httptest.NewRecorder()
	(&API{}).handleContent(rec, httptest.NewRequest("GET", "/content", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Fatalf("body = %q, want []", got)
	}
}
