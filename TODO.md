# Todo List

## 🔴 Blockers
- [x] **Fix Windows CGO build** (v1.0.0-alpha.78) — Migrate `go-sqlite3` to `modernc.org/sqlite` (pure Go).
- [x] **Real Social Posting** (v1.0.0-alpha.82) — Real Twitter/LinkedIn API integrations (OAuth1, Bearer token) with DryRun support, error handling, and production tests.

## 🟠 High Priority
- [x] **Real Trading Data** (v1.0.0-alpha.83) — CoinGecko fetcher production-hardened with caching, retry, rate-limit handling, and API key support.
- [x] **Markdown CMS** (v1.0.0-alpha.82) — Static site generator in `hustle/content/deploy.go` with RSS feed, responsive CSS, and local preview server.
- [x] **Content Deployment Pipeline** — Automated deploy to WordPress or GitHub Pages from the content output.

## 🟡 Medium Priority
- [x] **Dashboard Styling** (v1.0.0-alpha.78) — Add color codes to the terminal UI (green for profit, red for error).
- [x] **Task History** (v1.0.0-alpha.78) — Log task execution times and durations to SQLite.

## 🟢 Lower Priority
- [ ] **Multi-exchange support** — Binance, Kraken plugins.
- [x] **LLM response cache** — Content-addressable local cache.

## ✅ Completed
- [x] **Dynamic RSS Management** (v1.0.0-alpha.74)
- [x] **Scheduler Observability** (v1.0.0-alpha.74)
- [x] **Real Research API** (v1.0.0-alpha.68)
- [x] **Graceful Shutdown** (v1.0.0-alpha.67)
- [x] **Git Rollback** (v1.0.0-alpha.69)
- [x] **Mesh Wealth UI** (v1.0.0-alpha.70/75)
- [x] **REST API Expansion** (v1.0.0-alpha.72)
- [x] **Social Dry-Run & Real Posting** (v1.0.0-alpha.82) — Twitter OAuth1 + LinkedIn Bearer with tests
- [x] **Markdown CMS / Static Site Generator** (v1.0.0-alpha.82) — `deploy.go` with RSS, CSS, server
- [x] **CoinGecko Fetcher Hardening** (v1.0.0-alpha.83) — Caching, retry, rate-limit, API key, 14 tests

## 🟣 Fully Automated Real Estate Engine
- [x] **Affiliate Marketing Engine** — Scrape hot products, LLM review, auto-tweet, and blog.
- [ ] **Lead Generation + Outreach** — Scrape Google Maps, LLM personalize pitch, SMTP rate-limited delivery.
- [ ] **Digital Product Factory** — Generate Notion templates, presets, guides, auto-list on Gumroad via API.
- [ ] **YouTube Shorts Factory** — Script -> Image generation -> TTS -> Auto Upload.

## v1.2.0 Dashboard & Prod
- [x] Consolidated one-page dashboard with dispatchAction() actually implemented (was undefined � buttons were dead)
- [x] Removed fake hardcoded dashboard data ( balance, always-green status dots) � now live /api/status
- [ ] Wire hustle://calendar + hustle://publisher handlers (publisher.PublishEntry exists)
- [ ] Fix TestSyncProtocolIntegration on Windows (execs bash sync.sh � skip on GOOS=windows)
- [ ] Regenerate X/Twitter API keys (401 in prod, docs/ai/maintenance-2026-08-01.md)
- [ ] Move FreeLLM keys out of the systemd unit into an EnvironmentFile with 600 perms, then rotate
- [ ] Re-enable + start all aimm-* services on Hetzner (found all inactive/disabled, 2026-11-13)
- [x] Wire hustle://calendar + hustle://publisher handlers (v1.2.1)
- [x] Fix TestRollback destroying uncommitted work: it ran git checkout/clean in the real repo; now uses a temp repo (v1.2.1)
- [x] Fix TestSyncProtocolIntegration on Windows � skips without bash (v1.2.1)
- [ ] CoinGecko 429s on prod (no COINGECKO_API_KEY set): trading tasks for XRP/ADA failing with rate limited after 3 retries. User: set a free CoinGecko Demo key in prod .env, or raise per-symbol intervals in daemon scheduler.
- [x] Embed dashboard in Go binary, retire Python aimm-dashboard service (v1.3.0)
- [ ] TradingModule keeps ONE shared History/RSIHistory across all symbols (orchestrator reuses a single traderModule and mutates .Symbol): SMA/RSI mix BTC/XRP/ADA prices, so indicators are meaningless across symbols. Fix: per-symbol history map in TradingModule (v1.3.1 observation from prod logs: ADA SMA $0.58 vs price $0.16).
- [x] CoinGecko batch + stale fallback (v1.3.1)
- [x] Per-symbol trading history (v1.3.3)
- [x] Shorts FFmpeg render pipeline (v1.4.0/v1.4.1) - MP4 verified on prod: 1080x1920 h264+aac, 38s, drawtext cards
- [ ] Shorts TTS: set ELEVENLABS_API_KEY in prod orchestrator .env to enable voice (currently silent cards, manifest says so)
- [ ] Shorts YouTube upload: needs YOUTUBE_CLIENT_ID/YOUTUBE_REFRESH_TOKEN OAuth from user
- [x] System tray controller (v1.5.0, tray/)
