# Todo List (`v2.0.0`)

## Completed in `v2.0.0` (2026-11-14)
- [x] Restore all 20+ hollow `hustle/*` Go modules (`stocks`, `domains`, `retail`, `realestate`, `ecommerce`, `saas`, `kdp`, `pod`, `confluence`, `outreach`, `crm`, `automation`, `bi`, `finance`, `legal`, `design`, `devagency`, `marketing`, `localization`, `support`, `careers`, `qa`) and wire all **36 `hustle://` protocol handlers** in `orchestrator/cmd/orchestrator/main.go`.
- [x] Restore and wire `orchestrator/llm_cache.go` (`CachingLLM`), `orchestrator/identity.go` (`did:hustle` Ed25519), `orchestrator/log_buffer.go` (`RingLogBuffer`), and `/api/hymns`.
- [x] Redesign `orchestrator/dashboard.html` into one Consolidated Single-Page Command Center ordered by highest ROI, folding `/bot` into an inline SVG Bot Inspector, adding `ⓘ` tooltips across every control, and surfacing all 36 modules + 10 companion service logs.
- [x] Upgrade `tray/` (`bin\aimm-tray.exe`) with `-H=windowsgui`, hidden child windows, automatic SSH port-forwarding (`localhost:8082 -> hetzner:8082`), BelowNormal local server start/stop, and 1-click quick dispatches.
- [x] Fix Cloudflare `403 Forbidden` on Hetzner WordPress REST scripts (`expand.py`, `batch_publisher.py`, `content_audit.py`, `wordpress.go`) and remove hardcoded WordPress/Twitter credentials from repo scripts.
- [x] Fix `twitter_poster.py` `<think>` reasoning leaks + `aimm-twitter.timer` (`OnUnitInactiveSec=6h`, `Persistent=true`) + `watchdog.sh` unit list & PM2 false alarms.
- [x] Nondestructively organize 32 superseded experimental files/patches into `archive/` and `fixes/archive/`, and enforce `BelowNormal` priority in `scripts/build-all.ps1`, `scripts/build-tray.ps1`, `scripts/manage.ps1`, and `build.sh`.

## Remaining Operator / Credential-Gated Tasks (`v2.1+`)
- [ ] **X / Twitter Credits (`402 Payment Required`):** Top up X Developer Portal pay-per-use credits (or switch default social distribution solely to Bluesky/LinkedIn, where Bluesky is already posting live).
- [ ] **Binance.US API Key / IP Whitelist (`-2015` / HTTP `401`):** Whitelist Hetzner IP `5.161.250.43` and enable Spot Read/Trade on `BINANCE_API_KEY` in `/opt/aimoneymachine/.env` (or switch Python trader to paper/Coinbase).
- [ ] **Gumroad / Shopify Storefront Keys:** Set `GUMROAD_ACCESS_TOKEN` (and optional `SHOPIFY_ACCESS_TOKEN` + `SHOPIFY_SHOP_NAME`) in `/opt/aimoneymachine/.env` so generated digital products auto-publish live checkout links.
- [ ] **YouTube & Voiceover Keys:** Set `ELEVENLABS_API_KEY` + `YOUTUBE_CLIENT_ID` / `YOUTUBE_CLIENT_SECRET` / `YOUTUBE_REFRESH_TOKEN` to auto-upload rendered Shorts and `/opt/hymnmania/youtube_ready/*.mp4` videos.
- [ ] **FreeLLM Unit Hardening:** Move inline API keys from `/etc/systemd/system/aimm-freellm.service` into `/opt/aimoneymachine/.freellm.env` (`chmod 600`) and rotate old free-tier keys.
