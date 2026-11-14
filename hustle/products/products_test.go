package products

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

func TestGenerateWritesProductAndHonestListing(t *testing.T) {
	out := t.TempDir()
	f := NewFactory(orchestrator.NewOrchestrator())
	f.OutDir = out
	f.Gumroad = nil // no token in tests

	m, err := f.Generate("Notion CRM for freelancers", Template)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if m.Listing.Status != "not_configured" {
		t.Errorf("listing status = %q, want not_configured without token", m.Listing.Status)
	}
	if m.Listing.PriceCents != 2400 {
		t.Errorf("price = %d, want 2400", m.Listing.PriceCents)
	}
	for _, want := range []string{"product.md", "listing.md", "manifest.json"} {
		found := false
		for _, p := range m.Files {
			if filepath.Base(p) == want {
				found = true
				if info, err := os.Stat(p); err != nil || info.Size() == 0 {
					t.Errorf("file %s missing/empty: %v", p, err)
				}
			}
		}
		if !found {
			t.Errorf("manifest files missing %s: %v", want, m.Files)
		}
	}
	if _, err := f.Generate("x", ProductType("bogus")); err == nil {
		t.Error("unknown type must be rejected")
	}
}

func TestGumroadClientCreateProduct(t *testing.T) {
	var gotForm url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/products" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = r.ParseForm()
		gotForm = r.PostForm
		fmt.Fprint(w, `{"success":true,"product":{"id":"prod_123","short_url":"https://gumroad.com/l/abc"}}`)
	}))
	defer ts.Close()

	g := &GumroadClient{Token: "tok", BaseURL: ts.URL, Client: ts.Client()}
	id, link, err := g.CreateProduct("My Template", "desc", 2400)
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	if id != "prod_123" || link != "https://gumroad.com/l/abc" {
		t.Errorf("id/link = %q/%q", id, link)
	}
	if gotForm.Get("access_token") != "tok" || gotForm.Get("price_cents") != "2400" {
		t.Errorf("form wrong: %v", gotForm)
	}
}

func TestGumroadClientFailureSurfaces(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"success":false,"message":"Invalid access token"}`)
	}))
	defer ts.Close()

	g := &GumroadClient{Token: "bad", BaseURL: ts.URL, Client: ts.Client()}
	if _, _, err := g.CreateProduct("x", "y", 100); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("err = %v, want 401 surfaced", err)
	}
	if NewGumroadClientFromEnv() != nil && os.Getenv("GUMROAD_ACCESS_TOKEN") == "" {
		t.Error("client must be nil without token")
	}
}
