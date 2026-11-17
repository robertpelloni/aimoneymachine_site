# AI Hustle Machine — Operator & User Manual (`v2.0.1`)

## 1. Quick Orientation

The **AI Hustle Machine** (`aimoneymachine_site`) runs **exclusively live on Hetzner** (`5.161.250.43`, `/opt/aimoneymachine`) — **zero local servers or local port-forwarding tunnels** are needed on your workstation.

- **Public WordPress Site:** https://aimoneymachine.site
- **Live Consolidated Dashboard (Single Page):** https://aimoneymachine.site/dashboard
- **Live Telemetry Endpoint:** https://aimoneymachine.site/api/status
- **Production Server (Hetzner):** `ssh hetzner` (`5.161.250.43`, `/opt/aimoneymachine`)
- **API Token (`AIMM_API_TOKEN`):** Stored in `/opt/aimoneymachine/.env`. Click **🔐 API Token** in the top-right bar of `https://aimoneymachine.site/dashboard` and paste `AIMM_API_TOKEN` once (saved in browser `localStorage`) to unlock `POST /api/dispatch` actions and live multi-service daemon logs.

---

## 2. Optional Remote System Tray (`bin\aimm-tray.exe`)

`aimm-tray.exe` is a lightweight **remote-only** status indicator (never starts any local server or local port tunnel):

```powershell
# One-shot CLI check against live https://aimoneymachine.site/api/status:
.\bin\aimm-tray.exe --status
.\scripts\manage.ps1 status
```

### Menu Controls (When Launched)
- **Status Indicator:** Green (`Active`), Amber (`Disk >= 85%` or connecting), Red (`Offline`), polling `https://aimoneymachine.site/api/status` every 15s.
- **Open Live Hetzner Dashboard:** Opens `https://aimoneymachine.site/dashboard` in your default browser.
- **Quick Run Dispatches:** 1-click triggers for BTC Trading Cycle, Affiliate Pipeline, SEO Blog Post, and Self-Healer Audit on Hetzner (uses `AIMM_API_TOKEN` env var).
- **Hetzner Prod Controls:** Start, Restart, or Stop all `aimm-*` systemd units remotely via `ssh hetzner`.

---

## 3. Consolidated Single-Page Dashboard (`https://aimoneymachine.site/dashboard`)

All **36 `hustle://` backend protocol modules** are wired into one single page (zero subpages; historical `/bot?sym=BTC` URLs redirect to `/dashboard?bot=BTC#trading` and open the inline SVG Bot Inspector):

1. **Tier 1 — Quant Trading & Arbitrage (`#trading`):**
   - `hustle://trading?symbol=BTC|ETH|SOL|DOGE|XRP|ADA|LINK` (`action=watchlist` to add custom tickers)
   - `hustle://stocks?action=analyze|order&symbol=NVDA` (Alpaca paper/live execution)
   - `hustle://domains?domain=...` & `action=bid`
   - `hustle://retail?category=...&buy=...&sell=...` (Amazon FBA spread)
   - `hustle://realestate?location=...&rent=...&rate=...` & `action=manage` (Airbnb STR cashflow & AI host replies)
2. **Tier 2 — Affiliate, Products, Ecommerce & SaaS (`#commerce`):**
   - `hustle://affiliate?action=run|discover|batch_insert`
   - `hustle://products?topic=...&type=template|guide|preset`
   - `hustle://ecommerce?action=discover|listing|ads|fulfill`
   - `hustle://saas?action=ideate|mvp`
   - `hustle://kdp` (`action=keywords`) & `hustle://pod` (`action=listing`)
3. **Tier 3 — Content, 100K Pillar Expander, Static CMS & Calendar (`#content`):**
   - `hustle://content?topic=...&type=blog|seo|newsletter|thread` (`action=brainstorm|expand|cms`)
   - `hustle://publisher?action=status|wordpress|wordpress_markdown|newsletter|shopify`
   - `hustle://calendar?action=status|add|schedule_week|process|fill` (`fill` generates content for due topic-only shells, capped via `limit`, duplicates marked superseded; the 90-minute CalendarProcess job runs fill then process automatically)
4. **Tier 4 — Audience, Video & B2B LeadGen/CRM (`#growth`):**
   - `hustle://research?query=...`, `hustle://curation?topic=...`, `hustle://rss?action=add|remove|list`
   - `hustle://social?platform=Bluesky|Twitter|LinkedIn&topic=...`
   - `hustle://media?action=shorts|plan&topic=...` + `/api/hymns` (`HymnMania` queue)
   - `hustle://leadgen`, `hustle://outreach`, `hustle://synergy_leadgen`, `hustle://crm`
5. **Tier 5 — Autonomous Agents, Council, Chains & 11-Module Agency Workbench (`#automation`):**
   - `hustle://agent?type=general&iterations=5`, `hustle://autoplan`, `hustle://council?topic=...`
   - `hustle://chain?name=...` (`action=discover|create`), `hustle://confluence?recipe=saas_growth|dropship_loop|trading_alpha`
   - Tabbed Agency Suite: `automation`, `bi`, `finance`, `legal`, `design`, `devagency`, `marketing`, `localization`, `support`, `careers`, `qa`
6. **Tier 6 — Ledger, SQLite Task History, Vector Memory, Mesh & Multi-Service Logs (`#system`):**
   - `hustle://swarm?action=sync|aggregate|set_goal|broadcast_alpha`, `hustle://healer`, `hustle://sync`
   - Live Multi-Service Log Console (`/api/service-logs?service=orchestrator|publisher|content|audit|trader|hustle-gen|freellm|bluesky|twitter|watchdog`)

---

## 4. Building & Deploying to Hetzner (`BelowNormal` Priority)

```powershell
# Cross-compile binaries locally (never starts any local server):
.\scripts\build-all.ps1

# Run full smoke check against live HTTPS Hetzner deployment:
python .\scripts\smoke_prod.py https://aimoneymachine.site
```

On Linux / Git Bash:
```bash
./scripts/deploy-hetzner.sh
```
