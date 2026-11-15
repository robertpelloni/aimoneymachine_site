# Release Notes — v2.0.0 (2026-11-14)

## "Full-Spectrum Consolidated Command Center & 36-Module Restoration"

`v2.0.0` unifies every historical feature branch, restores all 20+ hollow Go modules in `hustle/*` and `orchestrator/`, redesigns the embedded dashboard into a single-page ROI-prioritized command center covering **100% of all 36 `hustle://` protocol handlers**, upgrades the Windows system tray controller with automatic SSH tunneling and `BelowNormal` process management, and resolves the Cloudflare `403 Forbidden` blockade on the Hetzner WordPress publishing/expansion/audit pipelines.

### Highlights

- **36/36 `hustle://` Modules Live & Interactive:** Restored and hardened `stocks` (Alpaca), `domains`, `retail` (FBA), `realestate` (STR), `ecommerce`, `saas`, `kdp`, `pod`, `confluence`, `outreach`, `crm`, `automation`, `bi`, `finance`, `legal`, `design`, `devagency`, `marketing`, `localization`, `support`, `careers`, and `qa`, plus `orchestrator/llm_cache.go` (SHA-256 prompt cache), `orchestrator/identity.go` (`did:hustle` Ed25519 identity), `orchestrator/log_buffer.go` (in-memory stdout tee for systemd), and `/api/hymns`.
- **Single-Page Consolidated Dashboard (`orchestrator/dashboard.html`):** Zero subpages (`/bot` folded into an inline SVG price/SMA/Bollinger Bot Inspector). Organized from highest direct ROI (`Trading & Arbitrage` -> `Affiliate, Products & SaaS` -> `Content, 100K & Static CMS` -> `Social, Video & B2B LeadGen` -> `Async Agents, Council, Chains & 11-Tab Agency Workbench` -> `Ledger, SQLite History, Vector Memory & 10-Service Live Logs`), with `ⓘ` hover tooltips on every card and an instant Inline Dispatch Output Drawer.
- **Upgraded System Tray (`bin\aimm-tray.exe`):** Compiled with `-H=windowsgui` and hidden child windows (`SysProcAttr{HideWindow: true}`), featuring automatic background SSH port-forwarding (`localhost:8082 -> hetzner:8082`), BelowNormal local server start/stop, 1-click quick dispatches, and remote Hetzner service controls.
- **Production Pipeline Fixes (Hetzner `5.161.250.43`):**
  - Added browser-grade `User-Agent` headers and `.env` credential loading to `expand.py`, `batch_publisher.py`, `content_audit.py`, and `hustle/publisher/wordpress.go`, eliminating Cloudflare `403 Forbidden` errors and removing hardcoded credentials.
  - Sanitized chain-of-thought leakage (`<think>` tags) and handled X API `402 Payment Required` cleanly in `twitter_poster.py` + added `OnUnitInactiveSec=6h` / `Persistent=true` to `aimm-twitter.timer`.
  - Updated `watchdog.sh` to cover all active `aimm-*` systemd units without false PM2 warnings.
- **Nondestructive Workspace Hygiene:** Archived 32 superseded experimental scripts/patches into `archive/` and `fixes/archive/` with sanitized `getenv(...)` placeholders and `BelowNormal` build scripts (`scripts/build-all.ps1`, `scripts/manage.ps1`, `build.sh`).
