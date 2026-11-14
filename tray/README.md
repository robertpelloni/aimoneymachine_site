# aimm-tray — system tray controller

The tray icon that did not exist before v1.5.0. A small Go app
(`fyne.io/systray`) that polls the orchestrator API and controls the
Hetzner prod services over the existing `ssh hetzner` alias.

## Menu

- **Status line** — live `● Active  v<version>  $<profit>` (polled every 15s)
- **Open Dashboard** — opens `AIMM_API_URL` in the browser
- **Refresh Now** — poll immediately
- **Start Prod Services** — `ssh hetzner systemctl start aimm-*`
- **Restart Orchestrator** — `ssh hetzner systemctl restart aimm-orchestrator`
- **Quit Servers (Stop Prod)** — `ssh hetzner systemctl stop aimm-*`
- **Quit Tray** — exits the tray only; servers keep running

The icon is a green dot when the orchestrator answers, red when offline.

## Config (env)

| Var | Default | Meaning |
|---|---|---|
| `AIMM_API_URL` | `http://localhost:8082` | Orchestrator API base. For prod: `ssh -L 8082:localhost:8082 hetzner` then keep the default. |
| `AIMM_SSH_HOST` | `hetzner` | ssh alias used for service control. |

## Build & run

```powershell
./scripts/build-tray.ps1   # writes bin\aimm-tray.exe
.\bin\aimm-tray.exe        # tray icon appears
.\bin\aimm-tray.exe --status   # headless one-line status check
```

Single-instance: a second launch exits with 'aimm-tray is already running'. The Start at Login checkbox writes the HKCU Run key (Windows).
