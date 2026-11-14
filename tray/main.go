// aimm-tray: system-tray / menubar controller for the AI Hustle Machine.
//
// It answers "where is the tray icon?": there wasn't one — the orchestrator
// is a headless daemon. This app polls the orchestrator API and offers:
//
//   - live status (icon colour + tooltip + first menu item)
//   - Open Dashboard (the API base URL in a browser)
//   - Start / Stop / Restart the Hetzner prod services via the existing
//     `ssh hetzner` alias — Stop is the requested "quit the servers"
//   - Quit (the tray itself)
//
// Config (env): AIMM_API_URL (default http://localhost:8082 — prod via
// `ssh -L 8082:localhost:8082 hetzner`), AIMM_SSH_HOST (default hetzner).
//
// Headless check: `aimm-tray --status` prints one status line and exits.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/systray"
)

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

func main() {
	once := flag.Bool("status", false, "print orchestrator status once and exit (no tray)")
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
		// Second launch: the first instance owns the tray. Exit quietly
		// (message for console launches) instead of stacking icons.
		fmt.Println("aimm-tray:", err)
		return
	}
	defer release()
	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetTitle("AIMM")
	systray.SetTooltip("AI Hustle Machine: starting…")
	systray.SetIcon(iconFor(false))

	mStatus := systray.AddMenuItem("● Starting…", "Live orchestrator status (polled every 15s)")
	mStatus.Disable()
	mOpen := systray.AddMenuItem("Open Dashboard", "Open the consolidated dashboard in a browser")
	systray.AddSeparator()
	mRefresh := systray.AddMenuItem("Refresh Now", "Poll the orchestrator immediately")
	mAutostart := systray.AddMenuItemCheckbox("Start at Login", "Launch aimm-tray when you sign in (Windows)", autostartEnabled())
	if !autostartSupported() {
		mAutostart.Disable()
		mAutostart.SetTooltip("Start-at-login is only implemented on Windows")
	}
	systray.AddSeparator()
	mStart := systray.AddMenuItem("Start Prod Services", "ssh: start all aimm-* services on Hetzner")
	mRestart := systray.AddMenuItem("Restart Orchestrator", "ssh: systemctl restart aimm-orchestrator")
	mStop := systray.AddMenuItem("Quit Servers (Stop Prod)", "ssh: stop all aimm-* services on Hetzner")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Tray", "Exit this tray app (servers keep running)")

	client := &http.Client{Timeout: 5 * time.Second}
	refresh := func() {
		st, err := FetchStatus(client, apiBase())
		if err != nil {
			systray.SetIcon(iconFor(false))
			systray.SetTooltip((&Status{}).Tooltip())
			mStatus.SetTitle((&Status{}).MenuLabel())
			return
		}
		systray.SetIcon(iconFor(st.Healthy()))
		systray.SetTooltip(st.Tooltip())
		mStatus.SetTitle(st.MenuLabel())
	}
	go func() {
		refresh()
		for range time.Tick(15 * time.Second) {
			refresh()
		}
	}()

	control := func(action string) {
		name, args, err := ControlCommand(action, sshHost())
		if err != nil {
			return
		}
		out, err := exec.Command(name, args...).CombinedOutput()
		if err != nil {
			mStatus.SetTitle(fmt.Sprintf("⚠ %s failed: %v", action, err))
			return
		}
		_ = out
		refresh()
	}

	for {
		select {
		case <-mOpen.ClickedCh:
			openBrowser(apiBase() + "/")
		case <-mRefresh.ClickedCh:
			refresh()
		case <-mAutostart.ClickedCh:
			want := !mAutostart.Checked()
			if err := setAutostart(want); err != nil {
				mStatus.SetTitle(fmt.Sprintf("⚠ start-at-login failed: %v", err))
			} else if want {
				mAutostart.Check()
			} else {
				mAutostart.Uncheck()
			}
		case <-mStart.ClickedCh:
			mStatus.SetTitle("… starting prod services")
			go control("start")
		case <-mRestart.ClickedCh:
			mStatus.SetTitle("… restarting orchestrator")
			go control("restart-orchestrator")
		case <-mStop.ClickedCh:
			mStatus.SetTitle("… stopping prod services")
			go control("stop")
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		_ = exec.Command("open", url).Start()
	default:
		_ = exec.Command("xdg-open", url).Start()
	}
}

// iconFor renders a 32x32 status dot as a PNG (ICO-wrapped on Windows by
// systray when needed; PNG bytes are accepted by the library).
func iconFor(healthy bool) []byte {
	const n = 32
	img := image.NewRGBA(image.Rect(0, 0, n, n))
	c := color.RGBA{239, 68, 68, 255} // red: offline
	if healthy {
		c = color.RGBA{34, 197, 94, 255} // green: active
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			dx, dy := float64(x-n/2)+0.5, float64(y-n/2)+0.5
			if dx*dx+dy*dy <= 13*13 {
				img.Set(x, y, c)
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
