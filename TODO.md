# Todo List

## 🔴 Blockers
- [x] **Fix Windows CGO build** (v1.0.0-alpha.78) — Migrate `go-sqlite3` to `modernc.org/sqlite` (pure Go).
- [x] **Real Social Posting** (v1.0.0-alpha.82) — Real Twitter/LinkedIn API integrations (OAuth1, Bearer token) with DryRun support, error handling, and production tests.

## 🟠 High Priority
- [x] **Real Trading Data** (v1.0.0-alpha.83) — CoinGecko fetcher production-hardened with caching, retry, rate-limit handling, and API key support.
- [x] **Markdown CMS** (v1.0.0-alpha.82) — Static site generator in `hustle/content/deploy.go` with RSS feed, responsive CSS, and local preview server.
- [x] **Content Deployment Pipeline** (v1.0.0-alpha.84) — Automated deploy to WordPress or GitHub Pages from the content output.
- [x] **Affiliate Automation** (v1.0.0-alpha.84) — Automatic link insertion in all generated content.
- [x] **Lead Discovery** (v1.0.0-alpha.84) — LLM-driven identification of business leads from research.
- [x] **LLM response cache** (v1.0.0-alpha.86) — Content-addressable local SQLite cache.
- [x] **Bollinger Bands & MACD** (v1.0.0-alpha.87) — Advanced technical indicators for trading.
- [x] **Ghost Protocol Base** (v1.0.0-alpha.87) — Stealth mode task jitter.
- [x] **Binance Execution** (v1.0.0-alpha.88) — Real trade implementation.
- [x] **Outreach Delivery** (v1.0.0-alpha.88) — SMTP pitch distribution.

## 🟡 Medium Priority
- [x] **Dashboard Styling** (v1.0.0-alpha.78) — Add color codes to the terminal UI (green for profit, red for error).
- [x] **Task History** (v1.0.0-alpha.78) — Log task execution times and durations to SQLite.

## 🟢 Lower Priority
- [ ] **Multi-exchange support** — Binance, Kraken plugins.

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
