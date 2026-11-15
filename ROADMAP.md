# Project Roadmap

## Phase 1: Federated Foundation (`v1.0.0-alpha.1` – `v1.0.0-alpha.30`) — COMPLETE
- [x] Tiered Memory (`L1` / `L2` / `L3`)
- [x] Multi-Agent Council of Supervisors
- [x] Agent-to-Agent (`A2A`) Broker
- [x] Initial Hustle Modules (`research`, `social`, `curation`)
- [x] `hustle://` Protocol Router

## Phase 2: Intelligence & Autonomy (`v1.0.0-alpha.31` – `v1.0.0-alpha.43`) — COMPLETE
- [x] Delta-Sync Memory Swarming
- [x] Chain Orchestration & Discovery
- [x] Mesh Aggregation & Collective Wealth Goal
- [x] Wealth Preservation (ROI auditing -> auto-termination)

## Phase 3: Real AI Integration (`v1.0.0-alpha.44` – `v1.0.0-alpha.66`) — COMPLETE
- [x] OpenAI-compatible LLM provider (`LM Studio`, `Ollama`, `FreeLLM`, `LiteLLM`)
- [x] Vector Embeddings (`modernc.org/sqlite` + cosine similarity search)
- [x] Autonomous Agent Loop (`Observe -> Think -> Act -> Learn -> Evaluate`)
- [x] Multi-Agent Concurrent Orchestrator & `PlanHustles` Strategic Planner
- [x] LLM-Verified Self-Healer

## Phase 4: Production Hustle Operations (`v1.0.0-alpha.67` – `v1.9.19`) — COMPLETE
- [x] Pure-Go SQLite (`modernc.org/sqlite`) — zero CGO requirement on Windows & Linux
- [x] Real Multi-Source Crypto Feeds (`CoinGecko`, `Binance`, `Kraken`, `Coinbase`) + RSI/SMA/MACD/Bollinger TA
- [x] Real Social Providers (`Bluesky` ATProto, `Twitter/X` OAuth1/v2, `LinkedIn` UGC v2) + Dry-Run mode
- [x] Digital Product Factory (`output/products/*/manifest.json`) & FFmpeg YouTube Shorts Factory (`output/shorts/`)
- [x] Bidirectional Mesh Peer Registration (`POST /register` + `-seed` callback) & Optional Bearer Auth (`AIMM_API_TOKEN`)

## Phase 5: Consolidated Command Center & Full 36-Module Suite (`v2.0.0`) — COMPLETE
- [x] **36/36 `hustle://` Modules Restored & Wired:** Full Go workspace integration across trading, equities (`Alpaca`), domain flipping, retail FBA, STR real estate, affiliate, digital products, dropshipping/Shopify, Micro-SaaS MVP generation, KDP, POD, content/100K/CMS, publisher/calendar, B2B leadgen/outreach/CRM, Confluence recipes, and 11 AI Agency modules.
- [x] **SHA-256 `CachingLLM` & Ed25519 Node Identity (`did:hustle`):** Active in runtime and surfaced on `/api/status`.
- [x] **Single-Page Consolidated Dashboard (`orchestrator/dashboard.html`):** ROI-prioritized layout, inline SVG Bot Inspector (replacing `/bot`), Inline Dispatch Output Drawer, `ⓘ` tooltips on every control, and 10-service live log console.
- [x] **System Tray Controller v2 (`bin/aimm-tray.exe`):** Auto-SSH tunnel to Hetzner `:8082`, `BelowNormal` local server toggle, hidden console windows, and 1-click quick dispatches.
- [x] **Cloudflare-Safe WordPress & Hetzner Pipeline:** Fixed `403 Forbidden` across `expand.py`, `batch_publisher.py`, `content_audit.py`, and `wordpress.go`.

## Phase 6: External Monetization Activation (`v2.1+`) — NEXT
- [ ] Configure `GUMROAD_ACCESS_TOKEN` or Stripe Payment Links to auto-publish `output/products/*` listings for direct checkout
- [ ] Whitelist Hetzner IP (`5.161.250.43`) or rotate `BINANCE_API_KEY` on Binance.US to transition `aimm-python-trader` from public-data/401 to live/authenticated execution
- [ ] Add `BUTTONDOWN_API_KEY` / `CONVERTKIT_API_SECRET` to activate automated email newsletter issues
- [ ] Add `YOUTUBE_CLIENT_ID` / `YOUTUBE_REFRESH_TOKEN` + `ELEVENLABS_API_KEY` for voiced automated YouTube Shorts & HymnMania uploads
- [ ] Move `aimm-freellm.service` inline provider keys into `/opt/aimoneymachine/.freellm.env` (`chmod 600`) and rotate exposed keys
