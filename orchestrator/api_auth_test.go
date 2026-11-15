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

	// Verify /api/service-logs and /api/log are also protected when Token != "".
	mux := http.NewServeMux()
	guarded.RegisterRoutes(mux)
	for _, p := range []string{"/api/log", "/api/service-logs?service=orchestrator"} {
		unauth := httptest.NewRecorder()
		mux.ServeHTTP(unauth, httptest.NewRequest(http.MethodGet, p, nil))
		if unauth.Code != http.StatusUnauthorized {
			t.Errorf("%s unauth: status = %d, want 401", p, unauth.Code)
		}
		authReq := httptest.NewRequest(http.MethodGet, p, nil)
		authReq.Header.Set("Authorization", "Bearer s3cret")
		authRec := httptest.NewRecorder()
		mux.ServeHTTP(authRec, authReq)
		if authRec.Code != http.StatusOK {
			t.Errorf("%s auth: status = %d, want 200", p, authRec.Code)
		}
	}
}
