package orchestrator

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// The embedded dashboard must be served at / and /dashboard, unknown paths
// must 404 (not fall through to the page), and /api/progress must answer
// valid empty JSON when no progress file exists yet.
func TestHandleDashboardServesEmbeddedPage(t *testing.T) {
	api := &API{}
	for _, path := range []string{"/", "/dashboard", "/dashboard/"} {
		rec := httptest.NewRecorder()
		api.handleDashboard(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 200 {
			t.Fatalf("%s: status = %d, want 200", path, rec.Code)
		}
		if got := rec.Header().Get("X-Robots-Tag"); got != "noindex, nofollow" {
			t.Fatalf("%s: X-Robots-Tag = %q, want noindex, nofollow", path, got)
		}
		if !strings.Contains(rec.Body.String(), "Consolidated Dashboard") {
			t.Fatalf("%s: body is not the dashboard page", path)
		}
	}
	rec := httptest.NewRecorder()
	api.handleDashboard(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != 404 {
		t.Fatalf("/nope: status = %d, want 404", rec.Code)
	}

	botRec := httptest.NewRecorder()
	api.handleBotRedirect(botRec, httptest.NewRequest("GET", "/bot?sym=ETH", nil))
	if botRec.Code != 302 || botRec.Header().Get("Location") != "/dashboard?bot=ETH#trading" {
		t.Fatalf("/bot redirect = %d %q", botRec.Code, botRec.Header().Get("Location"))
	}
}

func TestHandleProgressDefaultsToEmpty(t *testing.T) {
	t.Chdir(t.TempDir())
	rec := httptest.NewRecorder()
	(&API{}).handleProgress(rec, httptest.NewRequest("GET", "/api/progress", nil))
	if got := strings.TrimSpace(rec.Body.String()); got != `{"done":[]}` {
		t.Fatalf("body = %q, want {\"done\":[]}", got)
	}
}
