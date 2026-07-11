# Todo List

<<<<<<< HEAD
## 🔴 Blockers
- [x] **Fix Windows CGO build** (v1.0.0-alpha.78) — Migrate `go-sqlite3` to `modernc.org/sqlite` (pure Go).
- [x] **Real Social Posting** (v1.0.0-alpha.82) — Real Twitter/LinkedIn API integrations (OAuth1, Bearer token) with DryRun support, error handling, and production tests.

## 🟠 High Priority
- [x] **Real Trading Data** (v1.0.0-alpha.83) — CoinGecko fetcher production-hardened with caching, retry, rate-limit handling, and API key support.
- [x] **Markdown CMS** (v1.0.0-alpha.82) — Static site generator in `hustle/content/deploy.go` with RSS feed, responsive CSS, and local preview server.
- [x] **Content Deployment Pipeline** — Automated deploy to WordPress or GitHub Pages from the content output.
=======
## High Priority (v1.1.0)
- [x] Fix Twitter/X auto-posting (OAuth1 debug — keys exist, posts fail silently)
- [x] Build affiliate marketing engine (Amazon scrape → LLM → auto-post)
- [ ] Enable CoinGecko API key for real trading prices

## Medium Priority
- [x] YouTube Shorts factory (AI script → image → TTS → upload)
- [ ] Multi-platform content repurposing pipeline
- [ ] Lead generation + email outreach automation
- [ ] Digital product factory (LLM-generated templates → Gumroad)
>>>>>>> e74c915 (I have validated the system state. Here is a summary of the steps I took:)

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
