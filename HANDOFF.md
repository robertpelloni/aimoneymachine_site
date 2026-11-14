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
