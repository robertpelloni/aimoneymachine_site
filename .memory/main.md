# Brain Memory — AI Hustle Machine (`aimoneymachine_site`)

## Current Architecture (`v2.0.0`)
- **Single Source of Truth Version**: `VERSION.md` (`2.0.0`).
- **Core Go Daemon (`orchestrator/`)**:
  - Multi-tier memory (`L1` Scratchpad, `L2` Vault, `L3` Archive) persisted atomically to `memory.json` + pure-Go SQLite (`hustle.db` via `modernc.org/sqlite` with cosine-similarity vector search).
  - Real OpenAI-compatible LLM client (`openai_compat.go`) connected to `FreeLLM` (`:4000`, 473 models) with `WaterfallLLM` fallback (`LLM_FALLBACK_BASE_URL`), `CachingLLM` (`llm_cache.go`) prompt deduplication, and `PromptOptimizer` epsilon-greedy multi-armed bandit (`prompt_optimizer.go`).
  - Ed25519 decentralized node identity (`identity.go` -> `did:hustle:<pubkey>`), A2A mesh broker (`a2a.go`), bidirectional `-seed` peer registration (`announceToSeed`), and checksummed `MemorySwarm` (`memory_swarm.go`).
  - Concurrent bounded `Scheduler` (`MaxConcurrent=3`, `BootStagger=20s`) preventing LLM stampedes while keeping per-symbol crypto bots (`BTC`, `ETH`, `SOL`, `DOGE`, `XRP`, `ADA`, `LINK`) running every 1-7 minutes with persisted `trading_history.json`.
  - Self-contained single-page **Consolidated Dashboard** (`orchestrator/dashboard.html` via `go:embed` in `orchestrator/api.go`) exposing 100% of all 36 `hustle://` protocol modules, inline SVG Bot Inspector, 11-tab AI Agency Workbench, Multi-Service Log Viewer (`/api/service-logs`), Hymn Remixes queue (`/api/hymns`), SQLite Task History, Memory/Bandit/Cache inspector, and optional `AIMM_API_TOKEN` auth.
- **Desktop System Tray (`tray/` -> `bin/aimm-tray.exe`)**:
  - Windows GUI tray controller (`-H=windowsgui`) with valid 22-byte Windows `.ico` dynamic status badges, proactive background SSH tunnel (`localhost:8082 -> hetzner:8082`), local `BelowNormal` server start/stop, 1-click quick hustle dispatches, and clean Quit/Stop controls.
- **Hetzner Production (`5.161.250.43`, `/opt/aimoneymachine`)**:
  - Systemd services: `aimm-orchestrator`, `aimm-freellm`, `aimm-litellm`, `aimm-content`, `aimm-publisher`, `aimm-audit`, `aimm-hustle-gen`, `aimm-python-trader`, `aimm-watchdog`, plus `aimm-healthcheck.timer` & `aimm-twitter.timer`.

## Key Operational Decisions & Lessons
1. **Two-Layer WordPress REST Auth Requirement (Cloudflare + JWT App Password)**:
   - Layer 1: Cloudflare WAF blocks `Python-urllib/*` and default Go HTTP clients (`403 Forbidden`); all Go (`hustle/publisher/wordpress.go`) and Python (`fixes/expand.py`, `fixes/batch_publisher.py`, `fixes/content_audit.py`) clients must send an explicit `Mozilla/5.0` User-Agent.
   - Layer 2: `jwt-authentication-for-wp-rest-api` v1.5.0 must remain active in `/var/www/aimoneymachine` AND `WORDPRESS_PASSWORD` in `/opt/aimoneymachine/.env` must be a valid WordPress **Application Password** (`AIMM-Orchestrator-v2`), not the wp-admin login password.
2. **Consolidated Single-Page UI**: Every backend module must be represented on the single consolidated page (`orchestrator/dashboard.html`) ordered by ROI/value — zero separate subpages. Keep `Consolidated Dashboard`, `diskPct`, and `dispatchAction` markers intact for automated smoke & unit tests.
3. **Nondestructive Repo & Process Hygiene**: Archive superseded scripts into `archive/` instead of deleting; track `.memory/`, `.pi/`, and `*.db` in git; run heavyweight builds/daemons at `BelowNormal` priority (`-buildvcs=false`); never delete compiled binaries during final builds.
