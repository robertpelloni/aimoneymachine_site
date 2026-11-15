# aimoneymachine_site — AI Hustle Machine (`v2.0.0`)

Autonomous, LLM-driven multi-module revenue engine (`orchestrator` + 31 `hustle/*` Go modules + Windows `aimm-tray` desktop controller) powering **https://aimoneymachine.site** and a single-page Consolidated Command Center.

## Architecture at a Glance

- **Core Engine (`orchestrator/` + `hustle/*`, Go 1.25 `go.work`):**
  - **36 `hustle://` Protocol Handlers** spanning crypto quant trading (`BTC`, `ETH`, `SOL`, `DOGE`, `XRP`, `ADA`, `LINK` via CoinGecko/Binance/Kraken/Coinbase), Alpaca equities, domain flipping, Amazon FBA & Airbnb STR arbitrage, affiliate pipelines, digital product & single-file Micro-SaaS MVP factories, WordPress/Newsletter/Shopify publishing, 100K-character SEO pillar expansion, YouTube Shorts + HymnMania queues, B2B LeadGen/Outreach/CRM, async multi-agent loops, 4-persona Council debates, and an 11-module AI Agency workbench.
  - **Single-Page Consolidated Dashboard (`orchestrator/dashboard.html`):** Embedded directly into the Go binary on `:8082` (`/` and `/dashboard`) — zero subpages, ROI-ordered sections, inline SVG Bot Inspector, instant Dispatch Output Drawer, `ⓘ` tooltips on every control, and a 10-service live log viewer.
- **Desktop System Tray Controller (`tray/` -> `bin\aimm-tray.exe`):**
  - Live green/amber/red status icon, automatic background SSH tunnel (`localhost:8082 -> hetzner:8082`), `BelowNormal`-priority local server start/stop, 1-click hustle dispatches, and remote Hetzner systemd controls.
- **Research Corpus & Memory (`.memory/`, `.pi/`, root `.md` library):**
  - Extensive architecture & agent-monetization reference library plus tracked Brain memory and SQLite state (`hustle.db`).

## Quick Start

```powershell
# Build all Windows + Linux binaries at BelowNormal CPU priority:
.\scripts\build-all.ps1

# Launch the Windows System Tray Icon (auto-connects SSH tunnel to Hetzner :8082):
.\scripts\manage.ps1 start-tray

# Verify health & run smoke suite:
.\bin\aimm-tray.exe --status
python .\scripts\smoke_prod.py http://localhost:8082
```

See [USER_MANUAL.md](USER_MANUAL.md) for full dashboard & protocol usage, [DEPLOY.md](DEPLOY.md) for Hetzner production operations, [HANDOFF.md](HANDOFF.md) for live telemetry & handoff notes, and [CHANGELOG.md](CHANGELOG.md) for version history.

## Honest Financial Status

 Recorded historical ledger baseline in `STATUS.json`: revenue \$171.60, infrastructure/setup expenses \$1,005.00, net **-\$833.40**. All 36 software pipelines, WordPress publishing/expansion daemons, Bluesky auto-posting, and quant/product/agency engines are active and verified; external marketplace keys (`GUMROAD_ACCESS_TOKEN`, Binance IP whitelist, X API credits, YouTube OAuth) are documented in [TODO.md](TODO.md) to unlock direct external settlement.
