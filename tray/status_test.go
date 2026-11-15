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
		fmt.Fprint(w, `{"status":"Active","version":"2.0.0","profit":12.5,"memories":42,"peers":3,"disk_used_pct":91,"disk_free_gb":7.2}`)
	}))
	defer srv.Close()

	st, err := FetchStatus(srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("FetchStatus: %v", err)
	}
	if !st.Healthy() {
		t.Error("expected healthy")
	}
	if got := st.Tooltip(); !strings.Contains(got, "v2.0.0") || !strings.Contains(got, "$12.50") {
		t.Errorf("tooltip = %q", got)
	}
	if got := st.Tooltip(); !strings.Contains(got, "disk 91%") {
		t.Errorf("tooltip should warn at >=85%% disk: %q", got)
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

func TestControlAndTunnelCommands(t *testing.T) {
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
	if _, args, _ := ControlCommand("status", "myhost"); args[len(args)-2] != "myhost" {
		t.Errorf("host arg = %q, want myhost", args[len(args)-2])
	}
	tBin, tArgs := TunnelCommand("hetzner", "8082")
	if tBin != "ssh" || !strings.Contains(strings.Join(tArgs, " "), "8082:localhost:8082") {
		t.Errorf("unexpected tunnel command: %s %v", tBin, tArgs)
	}
}

func TestDispatchURI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dispatch" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, "URI hustle://trading?symbol=BTC dispatched successfully")
	}))
	defer srv.Close()

	out, err := DispatchURI(srv.Client(), srv.URL, "hustle://trading?symbol=BTC")
	if err != nil {
		t.Fatalf("DispatchURI: %v", err)
	}
	if !strings.Contains(out, "dispatched successfully") {
		t.Errorf("unexpected output: %q", out)
	}
}
