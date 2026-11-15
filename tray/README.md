# AIMM System Tray Controller (`bin/aimm-tray.exe`)

Desktop menubar / Windows notification-area controller for the **AI Hustle Machine** (`v2.0.0`).

## Capabilities
- **Live Status Badge:** Green (`Active`), Amber (`Disk >= 85%` or connecting), Red (`Offline`), polling `/api/status` every 15 seconds.
- **Auto-Managed SSH Tunnel (`:8082`):** Automatically opens a hidden background `ssh -N -L 8082:localhost:8082 hetzner` port-forward when `localhost:8082` is closed, tracking strictly its own child PID and closing it cleanly on exit.
- **Local BelowNormal Server Mode:** Start/stop `bin\orchestrator.exe -api 8082 -daemon -dry-run` locally at `BelowNormal` CPU priority with hidden console windows.
- **1-Click Quick Dispatches:** Run BTC trading cycle, Affiliate pipeline, SEO blog generation, or Self-Healer audit right from the tray menu.
- **Hetzner Prod Lifecycle:** Start, Restart, or Stop (`Quit Prod Servers`) all `aimm-*` systemd units on Hetzner (`5.161.250.43`).
- **Start at Login:** Optional HKCU Windows Registry Run toggle.

## Build & Run
```powershell
.\scripts\build-tray.ps1
.\scripts\manage.ps1 start-tray
.\bin\aimm-tray.exe --status
```

Logs are written to `tray.log` in the repository root.
