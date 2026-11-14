package orchestrator

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }

func TestRequireToken(t *testing.T) {
	// No token configured: open (historical localhost behavior).
	open := &API{}
	rec := httptest.NewRecorder()
	open.requireToken(okHandler)(rec, httptest.NewRequest(http.MethodPost, "/dispatch", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("open API: status = %d, want 200", rec.Code)
	}

	guarded := &API{Token: "s3cret"}
	for name, mutate := range map[string]func(*http.Request){
		"missing":    func(r *http.Request) {},
		"wrong":      func(r *http.Request) { r.Header.Set("X-API-Key", "nope") },
		"x-api-key":  func(r *http.Request) { r.Header.Set("X-API-Key", "s3cret") },
		"bearer":     func(r *http.Request) { r.Header.Set("Authorization", "Bearer s3cret") },
		"wrong bear": func(r *http.Request) { r.Header.Set("Authorization", "Bearer nope") },
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/dispatch", nil)
		mutate(req)
		guarded.requireToken(okHandler)(rec, req)
		want := http.StatusUnauthorized
		if name == "x-api-key" || name == "bearer" {
			want = http.StatusOK
		}
		if rec.Code != want {
			t.Errorf("%s: status = %d, want %d", name, rec.Code, want)
		}
	}
}
