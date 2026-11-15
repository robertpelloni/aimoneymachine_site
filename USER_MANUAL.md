# AI Hustle Machine — Operator & User Manual (`v2.0.0`)

## 1. Quick Orientation

The **AI Hustle Machine** (`aimoneymachine_site`) is a Go `1.25` monorepo that runs an autonomous revenue & content engine (`orchestrator`) alongside a Consolidated Single-Page Web Dashboard (`/` on `:8082`) and a Windows System Tray Controller (`bin\aimm-tray.exe`).

- **Public WordPress Site:** https://aimoneymachine.site
- **Production Server (Hetzner):** `ssh hetzner` (`5.161.250.43`, `/opt/aimoneymachine`)
- **Consolidated Dashboard URL:** `http://localhost:8082/` (served directly from the Go binary; automatically port-forwarded from Hetzner by `bin\aimm-tray.exe`)

---

## 2. System Tray Controller (`bin\aimm-tray.exe`)

Launch on Windows with zero console windows:

```powershell
.\scripts\build-tray.ps1
.\scripts\manage.ps1 start-tray
```

### Menu Controls
- **Status Indicator:** Green (`Active`), Amber (`Disk >= 85%` or connecting), Red (`Offline`), polling `/api/status` every 15s.
- **Open Consolidated Dashboard:** Opens `http://localhost:8082/` in your default browser.
- **Auto SSH Tunnel to Hetzner (`:8082`):** Enabled by default. Whenever `localhost:8082` is closed, `aimm-tray` spawns a hidden `BelowNormal` background `ssh -N -L 8082:localhost:8082 hetzner` tunnel and cleanly terminates **only** its own tunnel PID when unchecked or when quitting the tray.
- **Start / Stop Local Server (`BelowNormal`):** Launches `bin\orchestrator.exe -api 8082 -daemon -dry-run` locally at `BelowNormal` CPU priority.
- **Quick Run Dispatches:** 1-click triggers for BTC Trading Cycle, Affiliate Pipeline, SEO Blog Post, and Self-Healer Audit.
- **Hetzner Prod Controls:** Start, Restart, or Stop (`Quit Prod Servers`) all `aimm-*` systemd units remotely.

---

## 3. Consolidated Single-Page Dashboard (`http://localhost:8082/`)

All **36 `hustle://` backend protocol modules** are wired into one single page (zero subpages; historical `/bot?sym=BTC` URLs redirect to `/?bot=BTC#trading` and open the inline SVG Bot Inspector):

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
   - `hustle://calendar?action=status|add|schedule_week|process`
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

## 4. Building & Deploying (`BelowNormal` Priority)

```powershell
# Build all Windows + Linux binaries without deleting existing artifacts:
.\scripts\build-all.ps1

# Run full smoke check against localhost:8082 (local or SSH-tunneled prod):
python .\scripts\smoke_prod.py http://localhost:8082
```

On Linux / Git Bash:
```bash
./scripts/deploy-hetzner.sh
```
