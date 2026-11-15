# Brain Memory — AI Hustle Machine (`aimoneymachine_site`)

## Current Architecture (`v2.0.1` — Hetzner Live Deployment Only)
- **Single Source of Truth Version**: `VERSION.md` (`2.0.1`).
- **Strict Operating Policy**: **NO local servers or local port-forwarding tunnels** on operator workstations. All runtime execution, dashboard serving, and API dispatches happen exclusively on live Hetzner production (`https://aimoneymachine.site/dashboard` + `https://aimoneymachine.site/api/*`).
- **Core Go Daemon (`orchestrator/`)**:
  - Multi-tier memory (`L1` Scratchpad, `L2` Vault, `L3` Archive) persisted atomically to `memory.json` + pure-Go SQLite (`hustle.db` via `modernc.org/sqlite` with cosine-similarity vector search).
  - Real OpenAI-compatible LLM client (`openai_compat.go`) connected to `FreeLLM` (`:4000`, 473 models) with `WaterfallLLM` fallback (`LLM_FALLBACK_BASE_URL`), `CachingLLM` (`llm_cache.go`) prompt deduplication, and `PromptOptimizer` epsilon-greedy multi-armed bandit (`prompt_optimizer.go`).
  - Ed25519 decentralized node identity (`identity.go` -> `did:hustle:<pubkey>`), A2A mesh broker (`a2a.go`), bidirectional `-seed` peer registration (`announceToSeed`), and checksummed `MemorySwarm` (`memory_swarm.go`).
  - Concurrent bounded `Scheduler` (`MaxConcurrent=3`, `BootStagger=20s`) preventing LLM stampedes while keeping per-symbol crypto bots (`BTC`, `ETH`, `SOL`, `DOGE`, `XRP`, `ADA`, `LINK`) running every 1-7 minutes with persisted `trading_history.json`.
  - Self-contained single-page **Consolidated Dashboard** (`orchestrator/dashboard.html` via `go:embed` in `orchestrator/api.go`) served live at `https://aimoneymachine.site/dashboard` (via Nginx reverse proxy to `127.0.0.1:8082`), exposing 100% of all 36 `hustle://` protocol modules, inline SVG Bot Inspector, 11-tab AI Agency Workbench, Bearer-protected Multi-Service Log Viewer (`/api/service-logs`), Hymn Remixes queue (`/api/hymns`), SQLite Task History, Memory/Bandit/Cache inspector, `X-Robots-Tag: noindex, nofollow`, and `AIMM_API_TOKEN` auth.
- **Remote-Only Tray Helper (`tray/` -> `bin/aimm-tray.exe`)**:
  - Optional remote status indicator polling `https://aimoneymachine.site/api/status` directly over HTTPS (zero local server, zero SSH tunnel).
- **Hetzner Production (`5.161.250.43`, `/opt/aimoneymachine`)**:
  - Nginx (`/etc/nginx/sites-available/aimoneymachine`): `/` serves WordPress; `/dashboard`, `/dashboard/`, `/bot`, and `/api/` proxy to `http://127.0.0.1:8082`.
  - Systemd services: `aimm-orchestrator`, `aimm-freellm`, `aimm-litellm`, `aimm-content`, `aimm-publisher`, `aimm-audit`, `aimm-hustle-gen`, `aimm-python-trader`, `aimm-watchdog`, plus `aimm-healthcheck.timer` & `aimm-twitter.timer`.

## Key Operational Decisions & Lessons
1. **Hetzner Live Only (Zero Local Server)**: Never launch `orchestrator.exe`, `dashboard_server.py`, or local `:8082` port-forwarding tunnels on the operator's Windows laptop; deploy and verify directly on Hetzner (`https://aimoneymachine.site/dashboard`).
2. **Two-Layer WordPress & Dashboard Cloudflare Requirement**:
   - Cloudflare WAF blocks `Python-urllib/*` and default Go HTTP clients (`403 Forbidden`); all Go (`hustle/publisher/wordpress.go`, `tray/status.go`) and Python (`fixes/*.py`, `scripts/smoke_prod.py`) clients must send an explicit `Mozilla/5.0` User-Agent.
   - `jwt-authentication-for-wp-rest-api` v1.5.0 must remain active in `/var/www/aimoneymachine` AND `WORDPRESS_PASSWORD` in `/opt/aimoneymachine/.env` must be a valid WordPress **Application Password** (`AIMM-Orchestrator-v2`).
3. **Public API Security (`AIMM_API_TOKEN`)**: Because `/api/` is proxied publicly on `https://aimoneymachine.site/api/`, `AIMM_API_TOKEN` MUST remain set in `/opt/aimoneymachine/.env` so all `POST` endpoints (`/api/dispatch`, `/api/register`, `/api/message`, `/api/sync`, `/api/healer`) and raw daemon log endpoints (`/api/log`, `/api/service-logs`) require Bearer authentication.
4. **Consolidated Single-Page UI & Repo Hygiene**: Every backend module stays on the single consolidated page (`orchestrator/dashboard.html`) ordered by ROI/value; never delete compiled binaries during builds; run builds at `BelowNormal` priority (`-buildvcs=false`).
