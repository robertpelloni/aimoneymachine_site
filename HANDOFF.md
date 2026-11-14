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
