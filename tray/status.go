package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Status mirrors the orchestrator GET /api/status payload.
type Status struct {
	Status      string  `json:"status"`
	Version     string  `json:"version"`
	Profit      float64 `json:"profit"`
	Revenue     float64 `json:"revenue"`
	Expenses    float64 `json:"expenses"`
	Memories    int     `json:"memories"`
	Peers       int     `json:"peers"`
	PriceSource string  `json:"price_source"`
	DryRun      bool    `json:"dry_run"`
	DiskUsedPct float64 `json:"disk_used_pct"`
	DiskFreeGB  float64 `json:"disk_free_gb"`
}

// FetchStatus GETs <base>/api/status, falling back to <base>/status.
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

// DispatchURI sends a POST /api/dispatch request for a hustle:// URI.
func DispatchURI(client *http.Client, base, uri string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	base = strings.TrimRight(base, "/")
	payload, _ := json.Marshal(map[string]string{"uri": uri})
	req, err := http.NewRequest("POST", base+"/api/dispatch", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if tok := os.Getenv("AIMM_API_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return strings.TrimSpace(string(body)), nil
}

// Tooltip is the one-line tray hover text.
func (s *Status) Tooltip() string {
	if s == nil || s.Status == "offline" || s.Status == "" {
		return "AI Hustle Machine: orchestrator offline"
	}
	tip := fmt.Sprintf("AI Hustle Machine v%s: %s, profit $%.2f, %d memories, %d peers",
		s.Version, s.Status, s.Profit, s.Memories, s.Peers)
	if s.DiskUsedPct >= 85 {
		tip += fmt.Sprintf(" | ⚠ disk %.0f%% used (%.1f GB free)", s.DiskUsedPct, s.DiskFreeGB)
	}
	return tip
}

// MenuLabel is the (disabled) first menu item text.
func (s *Status) MenuLabel() string {
	if s == nil || s.Status == "offline" || s.Status == "" {
		return "● Offline — auto-tunnel / start local server"
	}
	mode := ""
	if s.DryRun {
		mode = " [DryRun]"
	}
	return fmt.Sprintf("● %s  v%s  $%.2f  (Disk %.0f%%)%s", s.Status, s.Version, s.Profit, s.DiskUsedPct, mode)
}

// Healthy drives the icon colour.
func (s *Status) Healthy() bool {
	return s != nil && s.Status != "" && s.Status != "offline"
}

// ControlCommand builds the ssh invocation for a prod service action.
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
	return "ssh", []string{"-o", "BatchMode=yes", "-o", "ConnectTimeout=10", host, remote}, nil
}

// TunnelCommand returns the background ssh port-forward invocation mapping localPort -> remote localhost:8082.
func TunnelCommand(host, localPort string) (string, []string) {
	if host == "" {
		host = "hetzner"
	}
	if localPort == "" {
		localPort = "8082"
	}
	return "ssh", []string{
		"-N",
		"-o", "BatchMode=yes",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"-o", "ConnectTimeout=10",
		"-L", localPort + ":localhost:8082",
		host,
	}
}
