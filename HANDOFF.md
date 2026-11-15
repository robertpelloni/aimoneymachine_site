# SESSION HANDOFF — 2026-11-13 (Muse Code / muse-wonder-1.0 / aimoneymachine_site)

## What was broken (verified, not guessed)
- Repo did not build: unresolved merge-conflict markers were committed in go.work, 7 go.mod files, orchestrator+social go.sum, orchestrator/cmd/orchestrator/main.go, orchestrator/agent_loop.go, hustle/social/post.go, VERSION.md, STATUS.json, ROADMAP/TODO/HANDOFF. go build failed with 'errors parsing go.work'.
- dashboard.html buttons called dispatchAction() which was never defined — every action button was dead. Account/System cards were hardcoded fakes ( balance, always-green dots). /bot was a separate subpage on prod.
- hustle/social/post.go (newer side) used io without importing it; hustle/media + hustle/affiliate called LLM.GenerateJSON which only existed on OpenAICompatProvider, not the LLMProvider interface.
- Hetzner prod: ALL aimm-* services inactive + disabled (only nginx/MySQL/Redis/Ollama up). WordPress site itself is up. aimm-healthcheck + certbot failed. Prod VERSION 1.0.0-alpha.89, orchestrator binary from Jun 24.
- SECURITY: aimm-freellm.service contains ~50 plaintext API keys in the unit file. Rotate them and move to a 600-perm EnvironmentFile. (Keys are NOT reproduced here.)

## What was done (v1.2.0)
- Resolved all conflicts (kept newer Go 1.25/sqlite HEAD side; kept Twitter retry+error-body side; main.go got clean affiliate+media handlers using real APIs; dropped handlers for nonexistent packages bi/careers/kdp/retail/domains/realestate/calendar whose imports/functions do not exist).
- Added GenerateJSON to LLMProvider + all impls (GenerateJSONViaGenerate helper).
- Rewrote dashboard.html as ONE consolidated page: Overview > Money (trading/affiliate/content/publisher) > Growth (research/curation/social/shorts) > Automation (chains/leadgen/agents/strategy) > System (mesh/healer/sync/quit) > Logs. Every card: label, description, ? tooltip, form + working dispatch. Bot detail merged inline; no subpages.
- New dashboard_server.py (serves /, proxies /api/* to :8082, POST /api/dispatch). scripts/deploy-hetzner.sh added. build.sh now builds affiliate+media and supports ./build.sh linux.
- Verified: go build ./... = 0, go vet = 0 (orchestrator module). go test: only pre-existing Windows failure TestSyncProtocolIntegration (execs bash sync.sh) — unrelated, tracked in TODO.

## Next (in priority order)
1. Start prod services in dependency order (freellm -> orchestrator -> dashboard -> others) and watch journalctl for 10+ min.
2. Regenerate X/Twitter keys (401). 3. Rotate FreeLLM unit keys. 4. Wire hustle://calendar/publisher. 5. Shorts TTS/FFmpeg/upload. 6. Tray app (IDEAS #6).

## Tray icon answer
There is none, anywhere in the repo. Prod control = systemctl stop/start/restart aimm-*; local control = Ctrl+C. Dashboard System section documents this.


## Update 2026-11-14 (v1.2.1, Muse Code / muse-wonder-1.0)
- Wired hustle://calendar (status/process/add/schedule_week) + hustle://publisher (status/wordpress/wordpress_markdown/newsletter) in main.go; dashboard Publisher card now has live forms. publisher.ProcessPending() publishes due entries by index (GetPendingEntries copies lost status).
- ROOT-CAUSE FIND: TestRollback (orchestrator/sync_test.go) executed RollbackHandler in the real working tree — git checkout . + git clean -fd in orchestrator/ — silently reverting uncommitted orchestrator/ edits and deleting new untracked files during full test runs (hit in both v1.2.0 and v1.2.1 sessions). Fixed: test now runs in a temp git repo and asserts the revert there. Do NOT reintroduce a bare Execute() in the repo root; consider hardening RollbackHandler.Execute to require an explicit directory.
- GET /content now returns [] not 500 for an empty library. Full orchestrator suite green on Windows.

## Update 2026-11-14 (v1.3.0)
- Dashboard embedded via go:embed in orchestrator/api.go, served at / + /dashboard on :8082 with /api/* aliases incl. new Go handleLog/handleProgress. dashboard.html moved to orchestrator/dashboard.html (git rename). dashboard_server.py is fallback only.
- RollbackHandler hardened: Execute() refuses without explicit Dir; Dir must be the git top level; commands scoped with cmd.Dir. Healer passes os.Getwd() explicitly. TestRollback covers refusal + revert.
- Prod: aimm-dashboard (Python :8083) RETIRED (disable --now). Dashboard is now http://localhost:8082/ via ssh tunnel. healthcheck.sh on prod updated to check :8082/ instead of the dashboard unit (backup healthcheck.sh.bak-20261114). All other aimm-* active; healthcheck HEALTHY; site 200.
- New prod noise: CoinGecko 429 rate limiting (no API key) failing some trading cycles — see TODO.

## Update 2026-11-14 (v1.3.3)
- Per-symbol trading history (see CHANGELOG). Prod proof: old process showed ADA $0.16 with SMA $27,338.90 (BTC contamination); new process shows ADA SMA $0.16 / BTC SMA $67,359, each its own series.
- Watchdog was resurrecting the retired Python dashboard: /opt/aimoneymachine/bin/watchdog.sh SERVICES included aimm-dashboard + a :8083 port check, restarting it every 60s. Fixed on prod (backup watchdog.sh.bak-20261114) and canonical copy committed at scripts/watchdog.sh. Verified dashboard stays down across a full watchdog cycle; watchdog reports All services OK.

## Update 2026-11-14 (v1.4.0/v1.4.1)
- ShortsFactory renders real MP4s: SplitScenes -> optional ElevenLabs TTS -> ffmpeg 1080x1920 drawtext cards -> output/shorts/<slug>-<ts>/{script.txt,short.mp4,manifest.json}. Verified on prod: h264+aac, 38s, 185KB, 5 scenes, frame extract 49K (text rendered). Manifest honestly records skipped TTS (no ELEVENLABS_API_KEY) and not_configured upload.
- Gotcha: ffmpeg filtergraph does NOT treat \\' as an escape inside drawtext text='...' - raw ' and \\' both fail rc=8 'Filter not found' on ffmpeg 6.1; Unicode right single quote works (probe-verified with exact argv, no shell). escapeDrawtext uses that. Local minimal ffmpeg lacks drawtext+libx264: Assemble falls back to blank cards / mpeg4 and notes it in the manifest.

## Incident 2026-11-14: Hetzner root disk 100% -> site 504
- /dev/sda1 hit 100% (0 avail): nginx logged 'No space left on device', WordPress timed out (mysqld/php-fpm unwritable). Freed my own deploy backups (kept newest 2); external action then freed ~16G (disk back to ~79%) and site recovered to 200. Remaining /tmp ercm+contentplanner backups (~2G) deliberately kept.
- Watchdog disk alert tightened: WARNING >85%, CRITICAL >=95% with top /tmp+/opt consumers named in the log line (prod + scripts/watchdog.sh, backup watchdog.sh.bak-20261114b). Watch for recurrence: biggest steady consumers are /var/www/fwber (6.6G), /opt/tormentnexus (6.2G), shared-models GGUF (2.9G).

## Update 2026-11-14 (v1.9.0/v1.9.1 + overload incident)
- Digital Product Factory verified on prod: dispatch generated output/products/notion-crm-for-freelancers-template-*/{product.md 6.3KB real LLM content, listing.md, manifest.json listing=not_configured $24}. First attempt appeared to hang: box was at load 61-182 with swap exhausted (fwber next-build + sales_bot chrome/postgres + Ollama judge), SSH/HTTPS timed out; it recovered when the fwber build finished (load 0.37). No other project's processes were touched.
- LLM calls had a hardcoded 600s timeout in 3 constructors; the timed-out dispatch left NO journal trace. v1.9.1: llmTimeout()/llmHTTPClient() honoring LLM_TIMEOUT_SEC (prod .env now 240), and API dispatch failures log '[API] Dispatch failed for URI ...'. FreeLLM log confirmed client context-cancel at ~600s during the incident.
- Also observed: scheduler Sync task fails every cycle on prod ('fork/exec ./sync.sh: no such file') - sync.sh was never deployed to /opt/aimoneymachine; harmless noise, candidate fix: skip when absent.

## Update 2026-11-14 (v1.9.12 + agent loop verified)
- Agent loop verified end-to-end on prod (bounded dry-run, 3 iterations, real FreeLLM): chose research -> content (real 484-word blog written to output/content) -> social on the same topic, exactly the intended strategy chain. Dry-run respected (no public post).
- CORRECTION to the v1.9.12 diagnosis: the MockLLM fallback in manual runs was NOT a transient proxy blip. Prod .env contains a 'source /opt/shared-models/llm-env.sh' line: bash-sourcing .env executes it and that file overrides LLM_BASE_URL to a bare Ollama URL (no /v1, 404s), while systemd EnvironmentFile ignores the line so the daemon is unaffected. Manual prod runs must set LLM_BASE_URL explicitly (see /tmp/run_agent2.sh pattern) - never source .env in bash. The DetectModel status-check + retry hardening stays (it made this diagnosable in one log line).
- Also in .env: GMAIL_APP_PASSWORD is unquoted with spaces (bash sourcing mangles it; systemd reads it literally). Rotate/requote when touching that file. Do not print .env values in logs or chat.


---

## Update 2026-11-14 (`v2.0.0`) — [Gemini in Antigravity in c:\Users\hyper\workspace\aimoneymachine_site on laptop]

### 1. Summary of Completed Work (`v2.0.0`)
- **Full 36-Module Go Backend Restoration & Test Suite (`100% PASS`)**:
  - Reconciled all dropped modules/features across historical branches (`e74c915`, `503a61e`, `a6319b0`, `0ca10b2`) and `stash@{0}` into `main`.
  - Restored and wired all 36 `hustle://` protocol handlers in `orchestrator/cmd/orchestrator/main.go` (`trading`, `stocks`, `affiliate`, `products`, `content`, `publisher`, `calendar`, `ecommerce`, `saas`, `domains`, `kdp`, `pod`, `retail`, `realestate`, `media`, `curation`, `social`, `research`, `leadgen`, `synergy_leadgen`, `outreach`, `crm`, `automation`, `bi`, `careers`, `confluence`, `design`, `devagency`, `legal`, `localization`, `marketing`, `support`, `education`, `hymns`, `agent`, `auto_plan`, `chain`, `swarm`, `healer`, `sync`) plus `CachingLLM` (`orchestrator/llm_cache.go`), Ed25519 DID identity (`orchestrator/identity.go`), `LogBuffer` (`orchestrator/log_buffer.go`), bidirectional `-seed` peer registration (`announceToSeed`), and `AIMM_API_TOKEN` auth.
  - Verified `go test` passes 100% across `orchestrator`, `tray`, `trading`, and all 31 `hustle/*` Go workspace modules.
- **Single-Page Consolidated Dashboard (`orchestrator/dashboard.html`)**:
  - Embedded via `go:embed` at `/` and `/dashboard` on `:8082` with zero separate subpages (`/bot` folded into an inline SVG Bot Inspector).
  - Ordered strictly by direct ROI/value across 6 tiers (including a condensed 11-tab AI Agency Workbench), with KPI header strip, Integration Readiness bar, Direct `hustle://` Omnibox, Inline Dispatch Output Drawer, SQLite Task History, L1/L2/L3 + Vector Memory Search, Prompt Bandit/Cache stats, HymnMania queue (`/api/hymns`), 10-service Live Log Console (`/api/service-logs`), and `ⓘ` tooltips on every metric, input, and button.
- **Windows System Tray Controller v2 (`tray/` -> `bin/aimm-tray.exe`)**:
  - Added valid 22-byte Windows `.ico` dynamic status badges (`green` / `amber` / `red`), proactive background SSH tunnel startup (`localhost:8082 -> hetzner:8082` before `systray.Run`), hidden child console windows (`SysProcAttr{HideWindow: true}`), `BelowNormal` local server toggle, 1-click quick dispatches, and clean Quit/Stop server controls.
  - Launched locally as a persistent background daemon; verified `python scripts\smoke_prod.py http://localhost:8082` passes `16/16 SMOKE OK` on Windows through the tray's auto-SSH tunnel.
- **Hetzner Production Deployment & Two-Layer WordPress REST Fix (`5.161.250.43`)**:
  - Deployed `v2.0.0` `orchestrator`, `dashboard.html`, `dashboard.py`, `expand.py`, `batch_publisher.py`, `content_audit.py`, `twitter_poster.py`, `watchdog.sh`, and `smoke_prod.py` (with `.bak-20261114-v2` backups).
  - Resolved **both** root causes blocking WordPress REST automation:
    1. Added browser-compatible `Mozilla/5.0` `User-Agent` headers across Go (`hustle/publisher/wordpress.go`) and Python (`fixes/expand.py`, `fixes/batch_publisher.py`, `fixes/content_audit.py`) to clear Cloudflare Bot Fight Mode `403 Forbidden`.
    2. Reactivated `jwt-authentication-for-wp-rest-api` (`1.5.0`) in `/var/www/aimoneymachine` (was inactive, causing `404 Not Found` on `/wp-json/jwt-auth/v1/token`) and provisioned a dedicated WordPress Application Password (`AIMM-Orchestrator-v2`) in `/opt/aimoneymachine/.env` (`chmod 600`, backed up to `.env.bak-20261114`). Verified live: `JWT_AUTH_VERIFIED: HTTP 200`, `aimm-content` fetched all 398 posts, `aimm-audit` fetched Page 1 (100 posts), and `aimm-publisher` checked and skipped duplicates cleanly.
  - Upgraded `fixes/twitter_poster.py` (strips `<think>...</think>` reasoning leaks, logs clean `[BLOCKED-BILLING]` on X `402 Payment Required` without crashing systemd) and armed `aimm-twitter.timer` (`OnCalendar=*-*-* 00,06,12,18:00:00`, `NEXT` verified scheduled).
  - Updated `/opt/aimoneymachine/bin/watchdog.sh` to cover all 9 active `aimm-*` systemd units without false PM2 offline spam (`All monitored services OK`).
- **Nondestructive Repo Hygiene**:
  - Moved 32 superseded root WordPress/CSS/patch scripts into `archive/wordpress_experiments/`, `archive/patches/`, and `fixes/archive/` (redacting legacy Twitter OAuth test keys to `getenv(...)`).
  - Tracked `.memory/`, `.pi/`, and `hustle.db` in git; preserved all compiled binaries in `bin/` and root.

### 2. Remaining External / Credential-Gated Items for Next Cycle
1. **Binance.US Live Order Execution (`aimm-python-trader`)**: Public price/RSI/SMA/MACD/BB signals work on both Go (`aimm-orchestrator`) and Python (`aimm-python-trader`), but `trader.log` reports `Binance API error 401: {"code":-2015,"msg":"Invalid API-key, IP, or permissions for action."}` on live balance/order calls. Regenerate `BINANCE_API_KEY` + `BINANCE_SECRET_KEY` on Binance.US with Spot Trading enabled and whitelist Hetzner IP `5.161.250.43`.
2. **X / Twitter Credits (`402 Payment Required`)**: `aimm-twitter` and Go `social` authenticate cleanly via OAuth 1.0a, but X v2 returns `402 Payment Required (credits depleted)`. Automated social distribution currently routes through **Bluesky** (`chains.json` + `SOCIAL_DEFAULT_PLATFORM=Bluesky`) until X Developer Portal credits are topped up.
3. **Inbound Mail (`admin@aimoneymachine.site`)**: Postfix (`master`) listens on `0.0.0.0:25` on Hetzner, so inbound SMTP timeouts from Gmail stem from external port-25 filtering / Cloudflare DNS proxy; enable **Cloudflare Email Routing** for `aimoneymachine.site` (free forwarding of `admin@aimoneymachine.site` without exposing port 25) or open inbound TCP `25` in Hetzner Cloud Firewall + UFW with an unproxied `mail.aimoneymachine.site` MX record.
4. **Optional Storefront / Media Keys**: Populate `GUMROAD_ACCESS_TOKEN`, `SHOPIFY_SHOP_DOMAIN` + `SHOPIFY_ACCESS_TOKEN`, `ELEVENLABS_API_KEY`, `YOUTUBE_*` OAuth, and `SMTP_HOST`/`SMTP_USER`/`SMTP_PASS` in `/opt/aimoneymachine/.env` to switch those modules from honest local/dry-run artifact generation to live external publishing.
5. **Security Cleanup**: Move plaintext provider keys out of `/etc/systemd/system/aimm-freellm.service` into a `0600` `EnvironmentFile` and rotate any stale keys when convenient.

---

## Update 2026-11-14 (`v2.0.1`) — [Gemini in Antigravity in c:\Users\hyper\workspace\aimoneymachine_site on laptop]

### 1. Hetzner-Live-Only Migration (`v2.0.1` — Zero Local Servers)
- **Eliminated All Local Servers & Local Tunnels**:
  - Stopped the local Windows tray/SSH-tunnel daemon (`task-1159`) and verified zero local `orchestrator.exe`, `aimm-tray.exe`, or `:8082` listeners remain on the Windows workstation.
  - Stripped `startLocalServer`, `stopLocalServer`, `findLocalOrchestratorBinary`, `startTunnel`, `stopTunnel`, `TunnelCommand`, and `start-local`/`stop-local` from `tray/main.go`, `tray/status.go`, and `scripts/manage.ps1`.
  - Pointed `tray/` and `scripts/manage.ps1` / `scripts/smoke_prod.py` directly at the live Hetzner HTTPS deployment (`https://aimoneymachine.site/dashboard` and `https://aimoneymachine.site/api/status`) with Cloudflare-compatible `Mozilla/5.0` `User-Agent` headers.
- **Live HTTPS Exposure on Hetzner Nginx + Bearer Hardening**:
  - Provisioned `AIMM_API_TOKEN` in `/opt/aimoneymachine/.env` (`0600`, backed up first).
  - Hardened `orchestrator/api.go` so `/api/log` and `/api/service-logs` also enforce `requireToken` when `AIMM_API_TOKEN` is set (preventing anonymous scraping of `journalctl`/`freellm.log`), added `/api/register`, `/api/message`, `/api/sync`, and `/api/healer` routes, added `X-Robots-Tag: noindex, nofollow` + `<meta name="robots" content="noindex, nofollow">` on `/dashboard`, and updated `/bot` redirects to `/dashboard?bot=<SYM>#trading`.
  - Configured `/etc/nginx/sites-available/aimoneymachine` on Hetzner (backed up first) to reverse-proxy `/dashboard`, `/dashboard/`, `/bot`, and `/api/` to `http://127.0.0.1:8082`, replacing the stale broken WordPress `/dashboard` placeholder page while leaving `/` (WordPress homepage) and all 398 blog posts untouched.
  - Verified both internal (`python3 /opt/aimoneymachine/smoke_prod.py` with `.env` Bearer token) and external (`python scripts\smoke_prod.py https://aimoneymachine.site`) smoke suites pass `100% SMOKE OK`!
