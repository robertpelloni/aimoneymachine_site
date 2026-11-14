package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Status mirrors the orchestrator GET /api/status payload.
type Status struct {
	Status   string  `json:"status"`
	Version  string  `json:"version"`
	Profit   float64 `json:"profit"`
	Memories int     `json:"memories"`
	Peers    int     `json:"peers"`
}

// FetchStatus GETs <base>/api/status, falling back to <base>/status (the
// bare orchestrator route) so the tray works against either.
func FetchStatus(client *http.Client, base string) (*Status, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	base = strings.TrimRight(base, "/")
	var lastErr error
	for _, path := range []string{"/api/status", "/status"} {
		resp, err := client.Get(base + path)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("%s: HTTP %d", path, resp.StatusCode)
			continue
		}
		var st Status
		if err := json.Unmarshal(body, &st); err != nil {
			lastErr = err
			continue
		}
		if st.Status == "" {
			st.Status = "offline"
		}
		return &st, nil
	}
	return nil, lastErr
}

// Tooltip is the one-line tray hover text.
func (s *Status) Tooltip() string {
	if s == nil || s.Status == "offline" || s.Status == "" {
		return "AI Hustle Machine: orchestrator offline"
	}
	return fmt.Sprintf("AI Hustle Machine v%s: %s, profit $%.2f, %d memories, %d peers",
		s.Version, s.Status, s.Profit, s.Memories, s.Peers)
}

// MenuLabel is the (disabled) first menu item text.
func (s *Status) MenuLabel() string {
	if s == nil || s.Status == "offline" || s.Status == "" {
		return "● Offline — check orchestrator / tunnel"
	}
	return fmt.Sprintf("● %s  v%s  $%.2f", s.Status, s.Version, s.Profit)
}

// Healthy drives the icon colour.
func (s *Status) Healthy() bool {
	return s != nil && s.Status != "" && s.Status != "offline"
}

// ControlCommand builds the ssh invocation for a prod service action.
// host is an ssh alias (default "hetzner" from ~/.ssh/config). Quitting
// the servers = action "stop" (aimm-orchestrator + aimm-freellm).
func ControlCommand(action, host string) (name string, args []string, err error) {
	if host == "" {
		host = "hetzner"
	}
	var remote string
	switch action {
	case "start":
		remote = "systemctl start aimm-freellm aimm-orchestrator aimm-litellm aimm-content aimm-publisher aimm-audit aimm-hustle-gen aimm-python-trader aimm-watchdog"
	case "stop":
		remote = "systemctl stop aimm-python-trader aimm-hustle-gen aimm-audit aimm-publisher aimm-content aimm-litellm aimm-orchestrator aimm-freellm aimm-watchdog"
	case "restart-orchestrator":
		remote = "systemctl restart aimm-orchestrator"
	case "status":
		remote = "systemctl is-active aimm-freellm aimm-orchestrator aimm-litellm aimm-content aimm-publisher aimm-audit aimm-hustle-gen aimm-python-trader aimm-watchdog; curl -s http://localhost:8082/api/status"
	default:
		return "", nil, fmt.Errorf("unknown action %q", action)
	}
	return "ssh", []string{"-o", "ConnectTimeout=10", host, remote}, nil
}
