# Deployment & Operations Guide (`v2.0.0`)

## Prerequisites

- **Go 1.25+** (`go.work` workspace with 33 modules — pure-Go `modernc.org/sqlite`, **no CGO / gcc required** on Windows or Linux)
- **Python 3.10+** (stdlib-only smoke tests & companion WordPress/trading daemons)
- **OpenAI-Compatible LLM Endpoint:** Local `LM Studio` (`:1234`), `Ollama` (`:11434`), or Hetzner `FreeLLM` (`:4000`) / `LiteLLM` (`:4001`)

---

## 1. Local Build (`BelowNormal` CPU Priority)

Never delete compiled binaries. Build everything at `BelowNormal` priority so desktop responsiveness is unaffected:

```powershell
# Windows PowerShell — builds bin\orchestrator.exe, bin\aimm-tray.exe, 7 CLI tools, and Linux orchestrator_linux_new:
.\scripts\build-all.ps1

# Or build only the Windows system tray icon:
.\scripts\build-tray.ps1
```

On Linux / macOS / Git Bash:
```bash
./build.sh          # native binaries into bin/
./build.sh linux    # cross-compile orchestrator_linux_new for Hetzner amd64
```

---

## 2. Windows System Tray & Local Management

```powershell
# Launch the system tray icon (auto-opens hidden SSH tunnel localhost:8082 -> hetzner:8082):
.\scripts\manage.ps1 start-tray

# Check live status from CLI:
.\bin\aimm-tray.exe --status
.\scripts\manage.ps1 status

# Optional: run a local dry-run orchestrator at BelowNormal priority instead of SSH tunnel:
.\scripts\manage.ps1 start-local
.\scripts\manage.ps1 stop-local
```

---

## 3. Hetzner Production (`aimoneymachine.site` / `5.161.250.43`)

- **SSH Alias:** `ssh hetzner` (`root@5.161.250.43`, app root `/opt/aimoneymachine`, scripts `/opt/aimoneymachine/bin`, WordPress `/var/www/aimoneymachine`).
- **Systemd Units:**
  - `aimm-orchestrator.service` — Go v2.0.0 daemon (`-daemon -real-prices -api 8082`), serving both the embedded Consolidated Dashboard (`/`) and `/api/*` on `127.0.0.1:8082` (firewalled from public internet).
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
| `AIMM_API_TOKEN` | Optional Bearer token guarding state-changing `POST` API endpoints |
| `TRADING_PRICE_SOURCE` | `coingecko` (default), `binance`, `kraken`, or `coinbase` |
| `WORDPRESS_URL`, `WORDPRESS_USER`, `WORDPRESS_PASSWORD` | WordPress REST API Application Password for `aimoneymachine.site` |
| `BLUESKY_HANDLE`, `BLUESKY_APP_PASSWORD` | Live Bluesky ATProto social publishing |
| `TWITTER_API_KEY`, `TWITTER_API_SECRET`, `TWITTER_ACCESS_TOKEN`, `TWITTER_ACCESS_SECRET` | X/Twitter OAuth1 & v2 publishing |
| `LINKEDIN_ACCESS_TOKEN`, `LINKEDIN_AUTHOR_URN` | LinkedIn UGC v2 publishing |
| `GUMROAD_ACCESS_TOKEN` / `SHOPIFY_ACCESS_TOKEN` | Storefront listing creation for digital & dropshipped products |
| `ALPACA_API_KEY`, `ALPACA_API_SECRET`, `BINANCE_API_KEY` | Equity & crypto exchange order execution |
| `ELEVENLABS_API_KEY`, `YOUTUBE_CLIENT_ID`, `YOUTUBE_REFRESH_TOKEN` | Shorts TTS voiceover & automated YouTube uploads |
