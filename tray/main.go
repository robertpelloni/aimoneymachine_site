// aimm-tray v2.0.0: system-tray / menubar controller for the AI Hustle Machine.
//
// Features:
//   - Live green/amber/red status icon + tooltip + menu telemetry (polled every 15s)
//   - Auto-managed background SSH tunnel (`localhost:8082 -> hetzner:8082`) when `:8082` is closed,
//     cleanly tracking ONLY its own spawned child PID (never killing unrelated ssh/go processes)
//   - Start / Stop a local BelowNormal-priority `orchestrator.exe` daemon for offline/local testing
//   - 1-click Quick Hustle Dispatches (BTC Trading Cycle, Affiliate Pipeline, SEO Blog, Self-Healer)
//   - Start / Restart / Stop Hetzner production systemd services via `ssh hetzner`
//   - Hidden child console windows on Windows (`SysProcAttr{HideWindow: true}`) + `tray.log` telemetry
//
// Headless check: `aimm-tray --status` prints one status line and exits.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"fyne.io/systray"
)

type childManager struct {
	mu        sync.Mutex
	tunnelCmd *exec.Cmd
	localCmd  *exec.Cmd
}

var children = &childManager{}

func trayLog(format string, args ...interface{}) {
	line := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(trayLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		_, _ = f.WriteString(line)
		_ = f.Close()
	}
}

func trayLogPath() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if filepath.Base(dir) == "bin" || filepath.Base(dir) == "tray" {
			return filepath.Join(filepath.Dir(dir), "tray.log")
		}
		return filepath.Join(dir, "tray.log")
	}
	return "tray.log"
}

func apiBase() string {
	if v := os.Getenv("AIMM_API_URL"); v != "" {
		return v
	}
	return "http://localhost:8082"
}

func sshHost() string {
	if v := os.Getenv("AIMM_SSH_HOST"); v != "" {
		return v
	}
	return "hetzner"
}

func localPortFromAPI() string {
	u, err := url.Parse(apiBase())
	if err != nil {
		return "8082"
	}
	if p := u.Port(); p != "" {
		return p
	}
	return "8082"
}

func portOpen(port string) bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 900*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (m *childManager) startTunnel(host, port string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tunnelCmd != nil && m.tunnelCmd.Process != nil {
		return nil
	}
	if portOpen(port) {
		return nil
	}
	bin, args := TunnelCommand(host, port)
	cmd := exec.Command(bin, args...)
	configureHiddenProcess(cmd, true)
	if err := cmd.Start(); err != nil {
		trayLog("SSH tunnel start failed (%s:%s): %v", host, port, err)
		return err
	}
	pid := cmd.Process.Pid
	m.tunnelCmd = cmd
	trayLog("Started background SSH tunnel (PID %d): localhost:%s -> %s:8082", pid, port, host)
	go func(c *exec.Cmd, p int) {
		err := c.Wait()
		m.mu.Lock()
		if m.tunnelCmd == c {
			m.tunnelCmd = nil
		}
		m.mu.Unlock()
		trayLog("SSH tunnel PID %d exited (%v)", p, err)
	}(cmd, pid)
	return nil
}

func (m *childManager) stopTunnel() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.tunnelCmd != nil && m.tunnelCmd.Process != nil {
		pid := m.tunnelCmd.Process.Pid
		_ = m.tunnelCmd.Process.Kill()
		m.tunnelCmd = nil
		trayLog("Stopped tray-managed SSH tunnel PID %d", pid)
	}
}

func (m *childManager) tunnelActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tunnelCmd != nil && m.tunnelCmd.Process != nil
}

func findLocalOrchestratorBinary() (string, string) {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(dir, "orchestrator.exe"),
			filepath.Join(dir, "..", "bin", "orchestrator.exe"),
			filepath.Join(dir, "..", "orchestrator_windows.exe"),
		}
		for _, c := range candidates {
			if abs, aerr := filepath.Abs(c); aerr == nil {
				if _, serr := os.Stat(abs); serr == nil {
					workDir := filepath.Dir(abs)
					if filepath.Base(workDir) == "bin" {
						workDir = filepath.Dir(workDir)
					}
					return abs, workDir
				}
			}
		}
	}
	return filepath.Join("bin", "orchestrator.exe"), "."
}

func (m *childManager) startLocalServer(port string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.localCmd != nil && m.localCmd.Process != nil {
		return nil
	}
	bin, workDir := findLocalOrchestratorBinary()
	cmd := exec.Command(bin, "-api", port, "-daemon", "-dry-run")
	cmd.Dir = workDir
	configureHiddenProcess(cmd, true) // BelowNormal + HideWindow
	if err := cmd.Start(); err != nil {
		trayLog("Local orchestrator start failed: %v", err)
		return err
	}
	pid := cmd.Process.Pid
	m.localCmd = cmd
	trayLog("Started local BelowNormal orchestrator (PID %d) on :%s", pid, port)
	go func(c *exec.Cmd, p int) {
		err := c.Wait()
		m.mu.Lock()
		if m.localCmd == c {
			m.localCmd = nil
		}
		m.mu.Unlock()
		trayLog("Local orchestrator PID %d exited (%v)", p, err)
	}(cmd, pid)
	return nil
}

func (m *childManager) stopLocalServer() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.localCmd != nil && m.localCmd.Process != nil {
		pid := m.localCmd.Process.Pid
		_ = m.localCmd.Process.Kill()
		m.localCmd = nil
		trayLog("Stopped tray-managed local orchestrator PID %d", pid)
	}
}

func (m *childManager) localActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.localCmd != nil && m.localCmd.Process != nil
}

func main() {
	once := flag.Bool("status", false, "print orchestrator status once and exit (no tray)")
	noAutoTunnel := flag.Bool("no-tunnel", false, "disable automatic SSH tunnel to Hetzner :8082")
	flag.Parse()

	if *once {
		st, err := FetchStatus(nil, apiBase())
		if err != nil {
			fmt.Printf("offline (%s): %v\n", apiBase(), err)
			os.Exit(1)
		}
		fmt.Println(st.Tooltip())
		return
	}

	release, err := acquireInstance()
	if err != nil {
		trayLog("Second instance exited cleanly: %v", err)
		fmt.Println("aimm-tray:", err)
		return
	}
	defer release()

	autoTunnelDefault := !*noAutoTunnel && os.Getenv("AIMM_AUTO_TUNNEL") != "0"
	trayLog("aimm-tray v2.0.0 starting (api=%s, ssh=%s, autoTunnel=%v)", apiBase(), sshHost(), autoTunnelDefault)

	// Proactively establish the SSH tunnel before systray.Run so localhost:8082 is
	// reachable immediately even if Windows Explorer defers Shell_NotifyIcon / onReady.
	if autoTunnelDefault {
		_ = children.startTunnel(sshHost(), localPortFromAPI())
	}

	systray.Run(func() { onReady(autoTunnelDefault) }, func() {
		children.stopTunnel()
		children.stopLocalServer()
		trayLog("aimm-tray exited cleanly")
	})
}

func onReady(autoTunnelInit bool) {
	systray.SetTitle("AIMM")
	systray.SetTooltip("AI Hustle Machine: connecting…")
	systray.SetIcon(iconFor("starting"))

	mStatus := systray.AddMenuItem("● Connecting…", "Live orchestrator telemetry (polled every 15s)")
	mStatus.Disable()
	mOpen := systray.AddMenuItem("🌐 Open Consolidated Dashboard", "Open http://localhost:8082/ in your default browser")
	mRefresh := systray.AddMenuItem("↻ Refresh Status Now", "Poll /api/status immediately")

	systray.AddSeparator()
	mTunnel := systray.AddMenuItemCheckbox("🔌 Auto SSH Tunnel to Hetzner (:8082)", "Maintain a background SSH port-forward localhost:8082 -> hetzner:8082 when port 8082 is closed", autoTunnelInit)
	mLocal := systray.AddMenuItem("▶ Start Local Server (BelowNormal)", "Launch bin\\orchestrator.exe locally at BelowNormal CPU priority (Dry-Run)")

	systray.AddSeparator()
	mQuickTrade := systray.AddMenuItem("⚡ Quick Run: BTC Trading Cycle", "Dispatch hustle://trading?symbol=BTC")
	mQuickAff := systray.AddMenuItem("💰 Quick Run: Affiliate Pipeline", "Dispatch hustle://affiliate?action=run&niche=AI+productivity+tools")
	mQuickBlog := systray.AddMenuItem("✍ Quick Run: SEO Blog Post", "Dispatch hustle://content?topic=Autonomous+AI+Workflows&type=seo")
	mQuickHeal := systray.AddMenuItem("🩺 Quick Run: Self-Healer Audit", "Dispatch hustle://healer?issue=Tray+health+audit")

	systray.AddSeparator()
	mStartProd := systray.AddMenuItem("☁ Start Hetzner Prod Services", "ssh hetzner: start all aimm-* systemd units")
	mRestartProd := systray.AddMenuItem("♻ Restart Hetzner Orchestrator", "ssh hetzner: systemctl restart aimm-orchestrator")
	mStopProd := systray.AddMenuItem("⏹ Quit Prod Servers (Stop Hetzner)", "ssh hetzner: stop all aimm-* systemd units")

	systray.AddSeparator()
	mAutostart := systray.AddMenuItemCheckbox("Start Tray at Windows Login", "Launch aimm-tray automatically when you sign in", autostartEnabled())
	if !autostartSupported() {
		mAutostart.Disable()
	}
	mQuit := systray.AddMenuItem("✕ Quit Tray", "Close this tray icon and any tunnel/local server it started")

	client := &http.Client{Timeout: 5 * time.Second}
	port := localPortFromAPI()

	refresh := func() {
		st, err := FetchStatus(client, apiBase())
		if err != nil && mTunnel.Checked() && !children.localActive() {
			_ = children.startTunnel(sshHost(), port)
			time.Sleep(1400 * time.Millisecond)
			st, err = FetchStatus(client, apiBase())
		}
		if children.localActive() {
			mLocal.SetTitle("⏹ Stop Local Server (Running)")
		} else {
			mLocal.SetTitle("▶ Start Local Server (BelowNormal)")
		}
		if err != nil {
			systray.SetIcon(iconFor("offline"))
			systray.SetTooltip((&Status{}).Tooltip())
			mStatus.SetTitle((&Status{}).MenuLabel())
			return
		}
		state := "healthy"
		if st.DiskUsedPct >= 85 {
			state = "warn"
		}
		systray.SetIcon(iconFor(state))
		systray.SetTooltip(st.Tooltip())
		mStatus.SetTitle(st.MenuLabel())
		trayLog("Status OK: v%s profit=$%.2f disk=%.1f%% tunnel=%v local=%v", st.Version, st.Profit, st.DiskUsedPct, children.tunnelActive(), children.localActive())
	}

	go func() {
		refresh()
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			refresh()
		}
	}()

	controlProd := func(action string) {
		name, args, err := ControlCommand(action, sshHost())
		if err != nil {
			return
		}
		cmd := exec.Command(name, args...)
		configureHiddenProcess(cmd, true)
		out, err := cmd.CombinedOutput()
		if err != nil {
			trayLog("Prod control %s failed: %v (%s)", action, err, string(out))
			mStatus.SetTitle(fmt.Sprintf("⚠ %s failed: %v", action, err))
			return
		}
		trayLog("Prod control %s succeeded", action)
		time.Sleep(1200 * time.Millisecond)
		refresh()
	}

	quickDispatch := func(label, uri string) {
		mStatus.SetTitle("⏳ Running " + label + "…")
		out, err := DispatchURI(nil, apiBase(), uri)
		if err != nil {
			trayLog("Dispatch %s failed: %v", uri, err)
			mStatus.SetTitle("⚠ Dispatch failed — check API / token")
			return
		}
		trayLog("Dispatch %s OK: %s", uri, firstLine(out))
		refresh()
	}

	for {
		select {
		case <-mOpen.ClickedCh:
			openBrowser(apiBase() + "/")
		case <-mRefresh.ClickedCh:
			go refresh()
		case <-mTunnel.ClickedCh:
			if mTunnel.Checked() {
				mTunnel.Uncheck()
				children.stopTunnel()
			} else {
				mTunnel.Check()
				_ = children.startTunnel(sshHost(), port)
			}
			go refresh()
		case <-mLocal.ClickedCh:
			if children.localActive() {
				children.stopLocalServer()
			} else {
				children.stopTunnel()
				mTunnel.Uncheck()
				if err := children.startLocalServer(port); err != nil {
					mStatus.SetTitle(fmt.Sprintf("⚠ local start failed: %v", err))
				}
				time.Sleep(1000 * time.Millisecond)
			}
			go refresh()
		case <-mQuickTrade.ClickedCh:
			go quickDispatch("BTC Cycle", "hustle://trading?symbol=BTC")
		case <-mQuickAff.ClickedCh:
			go quickDispatch("Affiliate", "hustle://affiliate?action=run&niche=AI+productivity+tools")
		case <-mQuickBlog.ClickedCh:
			go quickDispatch("SEO Blog", "hustle://content?topic=Autonomous+AI+Workflows&type=seo")
		case <-mQuickHeal.ClickedCh:
			go quickDispatch("Healer", "hustle://healer?issue=Tray+health+audit")
		case <-mStartProd.ClickedCh:
			mStatus.SetTitle("… starting Hetzner services")
			go controlProd("start")
		case <-mRestartProd.ClickedCh:
			mStatus.SetTitle("… restarting Hetzner orchestrator")
			go controlProd("restart-orchestrator")
		case <-mStopProd.ClickedCh:
			mStatus.SetTitle("… stopping Hetzner services")
			go controlProd("stop")
		case <-mAutostart.ClickedCh:
			want := !mAutostart.Checked()
			if err := setAutostart(want); err != nil {
				mStatus.SetTitle(fmt.Sprintf("⚠ start-at-login failed: %v", err))
			} else if want {
				mAutostart.Check()
			} else {
				mAutostart.Uncheck()
			}
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			return s[:i]
		}
	}
	return s
}

func openBrowser(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	case "darwin":
		cmd = exec.Command("open", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	configureHiddenProcess(cmd, false)
	_ = cmd.Start()
}

// iconFor renders a crisp 32x32 status badge with a subtle outer ring.
func iconFor(state string) []byte {
	const n = 32
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	fill := color.RGBA{239, 68, 68, 255} // red: offline
	ring := color.RGBA{127, 29, 29, 255}
	switch state {
	case "healthy":
		fill = color.RGBA{34, 197, 94, 255} // green: active
		ring = color.RGBA{20, 83, 45, 255}
	case "warn", "starting":
		fill = color.RGBA{245, 158, 11, 255} // amber: disk warning / connecting
		ring = color.RGBA{120, 53, 15, 255}
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			dx, dy := float64(x-n/2)+0.5, float64(y-n/2)+0.5
			r2 := dx*dx + dy*dy
			if r2 <= 11.5*11.5 {
				img.Set(x, y, fill)
			} else if r2 <= 14*14 {
				img.Set(x, y, ring)
			}
		}
	}
	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, img)
	pngBytes := pngBuf.Bytes()
	if runtime.GOOS != "windows" {
		return pngBytes
	}
	// Windows LoadImageW(IMAGE_ICON) requires a valid 22-byte ICO header wrapping the PNG payload.
	var ico bytes.Buffer
	_ = binary.Write(&ico, binary.LittleEndian, uint16(0))                // Reserved
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1))                // Type: 1 = ICON
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1))                // Image count = 1
	ico.WriteByte(byte(n))                                                // Width (32)
	ico.WriteByte(byte(n))                                                // Height (32)
	ico.WriteByte(0)                                                      // Palette colors
	ico.WriteByte(0)                                                      // Reserved
	_ = binary.Write(&ico, binary.LittleEndian, uint16(1))                // Color planes
	_ = binary.Write(&ico, binary.LittleEndian, uint16(32))               // Bits per pixel
	_ = binary.Write(&ico, binary.LittleEndian, uint32(len(pngBytes)))    // Size of PNG data
	_ = binary.Write(&ico, binary.LittleEndian, uint32(6+16))             // Offset to PNG data (22)
	ico.Write(pngBytes)
	return ico.Bytes()
}
