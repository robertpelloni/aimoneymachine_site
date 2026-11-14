package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchStatusParsesAndFormats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/status" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"status":"Active","version":"1.5.0","profit":12.5,"memories":42,"peers":3}`)
	}))
	defer srv.Close()

	st, err := FetchStatus(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("FetchStatus: %v", err)
	}
	if !st.Healthy() {
		t.Error("expected healthy")
	}
	if got := st.Tooltip(); !strings.Contains(got, "v1.5.0") || !strings.Contains(got, "$12.50") {
		t.Errorf("tooltip = %q", got)
	}
	if got := st.MenuLabel(); !strings.Contains(got, "Active") {
		t.Errorf("menu label = %q", got)
	}
}

func TestFetchStatusOfflineFormats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer srv.Close()
	if _, err := FetchStatus(srv.Client(), srv.URL); err == nil {
		t.Fatal("expected error for 502")
	}
	st := &Status{}
	if st.Healthy() {
		t.Error("empty status must not be healthy")
	}
	if got := st.Tooltip(); !strings.Contains(got, "offline") {
		t.Errorf("offline tooltip = %q", got)
	}
}

func TestControlCommand(t *testing.T) {
	name, args, err := ControlCommand("stop", "")
	if err != nil {
		t.Fatalf("stop: %v", err)
	}
	joined := name + " " + strings.Join(args, " ")
	for _, want := range []string{"ssh", "hetzner", "systemctl stop", "aimm-orchestrator"} {
		if !strings.Contains(joined, want) {
			t.Errorf("stop command %q missing %q", joined, want)
		}
	}
	if _, _, err := ControlCommand("bogus", "hetzner"); err == nil {
		t.Error("unknown action must error")
	}
	if _, args, _ := ControlCommand("status", "myhost"); args[2] != "myhost" {
		t.Errorf("host arg = %q, want myhost", args[2])
	}
}
