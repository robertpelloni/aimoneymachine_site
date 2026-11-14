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
- [x] **Binance price source** (v1.7.0: BinancePriceFetcher over the existing exchange/binance provider, TRADING_PRICE_SOURCE=binance; public tickers, no keys needed, verified reachable from prod via api.binance.us)
- [x] **Kraken price source** (v1.7.1) - same PriceFetcher adapter pattern, public Ticker endpoint — Binance, Kraken plugins.
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
- [~] **Lead Generation + Outreach** - web discovery + honest SMTP done (v1.8.0); Google Maps/Places discovery + email enrichment remain (need API key) — Scrape Google Maps, LLM personalize pitch, SMTP rate-limited delivery.
- [x] **Digital Product Factory** (v1.9.0; Gumroad listing needs GUMROAD_ACCESS_TOKEN) — Generate Notion templates, presets, guides, auto-list on Gumroad via API.
- [~] **YouTube Shorts Factory** - script->scenes->FFmpeg MP4 done (v1.4.x, verified on prod); TTS + upload remain, credential-blocked below — Script -> Image generation -> TTS -> Auto Upload.

## v1.2.0 Dashboard & Prod
- [x] Consolidated one-page dashboard with dispatchAction() actually implemented (was undefined  buttons were dead)
- [x] Removed fake hardcoded dashboard data ( balance, always-green status dots)  now live /api/status
- [x] Wire hustle://calendar + hustle://publisher handlers (v1.2.1)
- [ ] Fix TestSyncProtocolIntegration on Windows (execs bash sync.sh  skip on GOOS=windows)
- [ ] Regenerate X/Twitter API keys (401 in prod, docs/ai/maintenance-2026-08-01.md)
- [ ] Move FreeLLM keys out of the systemd unit into an EnvironmentFile with 600 perms, then rotate
- [x] Re-enable + start all aimm-* services on Hetzner (v1.2.0; all active, smoke-verified)
- [x] Wire hustle://calendar + hustle://publisher handlers (v1.2.1)
- [x] Fix TestRollback destroying uncommitted work: it ran git checkout/clean in the real repo; now uses a temp repo (v1.2.1)
- [x] Fix TestSyncProtocolIntegration on Windows  skips without bash (v1.2.1)
- [x] CoinGecko 429s on prod - resolved v1.3.1 (batched fetch + stale fallback; a key was in fact present in prod .env). Binance is now an alternative source (v1.7.0, TRADING_PRICE_SOURCE=binance).
- [x] Embed dashboard in Go binary, retire Python aimm-dashboard service (v1.3.0)
- [x] Per-symbol trading history (v1.3.3; the unchecked duplicate of this item above was stale).
- [x] CoinGecko batch + stale fallback (v1.3.1)
- [x] Per-symbol trading history (v1.3.3)
- [x] Shorts FFmpeg render pipeline (v1.4.0/v1.4.1) - MP4 verified on prod: 1080x1920 h264+aac, 38s, drawtext cards
- [ ] Shorts TTS: set ELEVENLABS_API_KEY in prod orchestrator .env to enable voice (currently silent cards, manifest says so)
- [ ] Shorts YouTube upload: needs YOUTUBE_CLIENT_ID/YOUTUBE_REFRESH_TOKEN OAuth from user
- [x] System tray controller (v1.5.0, tray/)
- [ ] Apply sanitize-style output cleanup to social.GenerateContent (affiliate done v1.9.9; raw LLM scaffolding can leak there too)
