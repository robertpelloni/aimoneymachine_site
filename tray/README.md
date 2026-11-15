# AIMM Remote System Tray Controller (`bin/aimm-tray.exe`)

Remote-only desktop menubar / Windows notification-area controller for the **AI Hustle Machine** (`v2.0.1`).

> **Zero Local Servers:** `aimm-tray` never spawns a local `orchestrator` server or local SSH port-forwarding tunnel. It communicates directly with the live Hetzner deployment over HTTPS (`https://aimoneymachine.site`).

## Capabilities
- **Live Hetzner Status Badge:** Green (`Active`), Amber (`Disk >= 85%` or connecting), Red (`Offline`), polling `https://aimoneymachine.site/api/status` every 15 seconds.
- **Open Live Dashboard:** Opens `https://aimoneymachine.site/dashboard` in your default browser.
- **1-Click Quick Dispatches:** Run BTC trading cycle, Affiliate pipeline, SEO blog generation, or Self-Healer audit on Hetzner (authenticated via `AIMM_API_TOKEN`).
- **Hetzner Prod Lifecycle:** Start, Restart, or Stop all `aimm-*` systemd units remotely on Hetzner (`5.161.250.43`) via `ssh hetzner`.
- **Start at Login:** Optional HKCU Windows Registry Run toggle.

## Build & Check
```powershell
.\scripts\build-tray.ps1
.\bin\aimm-tray.exe --status
```

Logs are written to `tray.log` in the repository root.
