# aimoneymachine_site — AI Hustle Machine (`v2.0.1`)

Autonomous, LLM-driven multi-module revenue engine (`orchestrator` + 31 `hustle/*` Go modules) running **live on Hetzner** and powering **https://aimoneymachine.site** and the single-page Consolidated Command Center at **https://aimoneymachine.site/dashboard** (zero local servers).

## Architecture at a Glance

- **Core Engine (`orchestrator/` + `hustle/*`, Go 1.25 `go.work`):**
  - **36 `hustle://` Protocol Handlers** spanning crypto quant trading (`BTC`, `ETH`, `SOL`, `DOGE`, `XRP`, `ADA`, `LINK` via CoinGecko/Binance/Kraken/Coinbase), Alpaca equities, domain flipping, Amazon FBA & Airbnb STR arbitrage, affiliate pipelines, digital product & single-file Micro-SaaS MVP factories, WordPress/Newsletter/Shopify publishing, 100K-character SEO pillar expansion, YouTube Shorts + HymnMania queues, B2B LeadGen/Outreach/CRM, async multi-agent loops, 4-persona Council debates, and an 11-module AI Agency workbench.
  - **Live Single-Page Consolidated Dashboard (`https://aimoneymachine.site/dashboard`):** Embedded directly into the Go binary (`orchestrator/dashboard.html`) and reverse-proxied over Cloudflare HTTPS via Nginx — zero subpages, ROI-ordered sections, inline SVG Bot Inspector, instant Dispatch Output Drawer, `ⓘ` tooltips on every control, and a Bearer-protected 10-service live log viewer.
- **Remote-Only System Tray Helper (`tray/` -> `bin\aimm-tray.exe`):**
  - Optional remote status indicator polling `https://aimoneymachine.site/api/status` directly over HTTPS (never starts any local server or local port-forwarding tunnel) with 1-click remote dispatches and `ssh hetzner` systemd controls.
- **Research Corpus & Memory (`.memory/`, `.pi/`, root `.md` library):**
  - Extensive architecture & agent-monetization reference library plus tracked Brain memory and SQLite state (`hustle.db`).

## Live Operations (Zero Local Server)

- **Public Blog & Pillar Library:** https://aimoneymachine.site
- **Live Consolidated Dashboard:** https://aimoneymachine.site/dashboard
- **Live Status JSON:** https://aimoneymachine.site/api/status

```powershell
# Check live Hetzner status & run external HTTPS smoke suite (no local server):
.\scripts\manage.ps1 status
python .\scripts\smoke_prod.py https://aimoneymachine.site
```

See [USER_MANUAL.md](USER_MANUAL.md) for full dashboard & protocol usage, [DEPLOY.md](DEPLOY.md) for Hetzner production operations, [HANDOFF.md](HANDOFF.md) for live telemetry & handoff notes, and [CHANGELOG.md](CHANGELOG.md) for version history.

## Honest Financial Status

Recorded historical ledger baseline in `STATUS.json`: revenue $171.60, infrastructure/setup expenses $1,005.00, net **-$833.40**. All 36 software pipelines, WordPress publishing/expansion daemons, Bluesky auto-posting, and quant/product/agency engines are active and verified live on Hetzner; external marketplace keys (`GUMROAD_ACCESS_TOKEN`, Binance IP whitelist, X API credits, YouTube OAuth) are documented in [TODO.md](TODO.md) to unlock direct external settlement.
