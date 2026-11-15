# Deployment & Operations Guide (`v2.0.1`)

> **Operating Policy:** **Hetzner live deployment only (`https://aimoneymachine.site/dashboard`)** — no local server or local port-forwarding daemon runs on operator workstations.

## Prerequisites

- **Go 1.25+** (`go.work` workspace with 33 modules — pure-Go `modernc.org/sqlite`, **no CGO / gcc required** on Windows or Linux)
- **Python 3.10+** (stdlib-only smoke tests & companion WordPress/trading daemons)
- **Hetzner Production Host:** `ssh hetzner` (`root@5.161.250.43`, `/opt/aimoneymachine`)

---

## 1. Build Binaries (`BelowNormal` CPU Priority, No Local Server Started)

Never delete compiled binaries. Build cross-platform artifacts at `BelowNormal` priority:

```powershell
# Windows PowerShell — builds Linux orchestrator_linux_new + CLI/remote-tray binaries:
.\scripts\build-all.ps1
```

On Linux / macOS / Git Bash:
```bash
./build.sh linux    # cross-compile orchestrator_linux_new for Hetzner amd64
```

---

## 2. Live Hetzner Access & Optional Remote Tray

- **Live Consolidated Dashboard:** https://aimoneymachine.site/dashboard
- **Live Telemetry Endpoint:** https://aimoneymachine.site/api/status
- **API Token (`AIMM_API_TOKEN`):** Stored in `/opt/aimoneymachine/.env` on Hetzner (`ssh hetzner "grep ^AIMM_API_TOKEN /opt/aimoneymachine/.env"`). Paste into the **🔐 API Token** modal in the top-right bar of `https://aimoneymachine.site/dashboard` to unlock `POST /api/dispatch` actions and live multi-service daemon logs.

```powershell
# Query live Hetzner status or run external HTTPS smoke checks (zero local servers):
.\scripts\manage.ps1 status
.\scripts\manage.ps1 smoke
.\bin\aimm-tray.exe --status
```

---

## 3. Hetzner Production (`aimoneymachine.site` / `5.161.250.43`)

- **SSH Alias:** `ssh hetzner` (`root@5.161.250.43`, app root `/opt/aimoneymachine`, scripts `/opt/aimoneymachine/bin`, WordPress `/var/www/aimoneymachine`).
- **Nginx Reverse Proxy (`/etc/nginx/sites-available/aimoneymachine`):**
  - `/` and all WordPress posts/pages/REST API continue serving WordPress (`/var/www/aimoneymachine`).
  - `/dashboard`, `/dashboard/`, `/bot`, and `/api/` reverse-proxy directly to `http://127.0.0.1:8082` (`aimm-orchestrator`), guarded by `AIMM_API_TOKEN` for all state-changing dispatches and raw service logs.
- **Systemd Units:**
  - `aimm-orchestrator.service` — Go `v2.0.1` daemon (`-daemon -real-prices -api 8082`), serving the embedded Consolidated Dashboard (`/dashboard`) and `/api/*` on `127.0.0.1:8082`.
  - `aimm-freellm.service` (`:4000`) & `aimm-litellm.service` (`:4001`) — Free/fallback OpenAI-compatible LLM routers.
  - `aimm-content.service` (`expand.py` — 100K pillar expansion), `aimm-publisher.service` (`batch_publisher.py`), `aimm-audit.service` (`content_audit.py`), `aimm-hustle-gen.service` (`hustle_generator.py`), `aimm-python-trader.service` (`trader.py`), `aimm-watchdog.service` (`watchdog.sh`).
  - Timers: `aimm-bluesky.timer`, `aimm-twitter.timer` (`Persistent=true`, `OnUnitInactiveSec=6h`), `aimm-healthcheck.timer`.

### One-Command Production Deploy + Smoke Gate
```bash
./scripts/deploy-hetzner.sh
```
This builds `orchestrator_linux_new` at low CPU priority, creates timestamped `.bak-*` backups on Hetzner, uploads the binary + companion scripts, restarts services via `nice -n 10`, and runs `python3 /opt/aimoneymachine/smoke_prod.py`.

---

## 4. Key Environment Variables (`/opt/aimoneymachine/.env`)

| Variable | Purpose |
|---|---|
| `LLM_BASE_URL` / `LLM_FALLBACK_BASE_URL` | Primary & waterfall OpenAI-compatible LLM endpoints |
| `AIMM_API_TOKEN` | Bearer token guarding state-changing `POST` API endpoints & `/api/service-logs` |
| `TRADING_PRICE_SOURCE` | `coingecko` (default), `binance`, `kraken`, or `coinbase` |
| `WORDPRESS_URL`, `WORDPRESS_USER`, `WORDPRESS_PASSWORD` | WordPress REST API Application Password for `aimoneymachine.site` |
| `BLUESKY_HANDLE`, `BLUESKY_APP_PASSWORD` | Live Bluesky ATProto social publishing |
| `TWITTER_API_KEY`, `TWITTER_API_SECRET`, `TWITTER_ACCESS_TOKEN`, `TWITTER_ACCESS_SECRET` | X/Twitter OAuth1 & v2 publishing |
| `LINKEDIN_ACCESS_TOKEN`, `LINKEDIN_AUTHOR_URN` | LinkedIn UGC v2 publishing |
| `GUMROAD_ACCESS_TOKEN` / `SHOPIFY_ACCESS_TOKEN` | Storefront listing creation for digital & dropshipped products |
| `ALPACA_API_KEY`, `ALPACA_API_SECRET`, `BINANCE_API_KEY` | Equity & crypto exchange order execution |
| `ELEVENLABS_API_KEY`, `YOUTUBE_CLIENT_ID`, `YOUTUBE_REFRESH_TOKEN` | Shorts TTS voiceover & automated YouTube uploads |
