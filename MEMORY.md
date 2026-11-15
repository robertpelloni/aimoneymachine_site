[PROJECT_MEMORY]

# AI Hustle Machine — Project Memory & Architecture Summary (`v2.0.0`)

## 1. Core Vision & Philosophy
The **AI Hustle Machine** (`aimoneymachine_site`) is a self-orchestrating, LLM-driven autonomous agent monorepo designed to execute revenue-generating workflows (`Observe -> Think -> Act -> Learn -> Evaluate`) using zero-cost/low-cost LLM routers (`FreeLLM`, `LiteLLM`, `LM Studio`, `Ollama`).

## 2. Monorepo Architecture (Go 1.25 `go.work`, 33 Modules)
- **`orchestrator/`**: Core engine housing `AgentLoop` & `MultiAgentOrchestrator`, 4-persona `Council`, `L1/L2/L3` memory + pure-Go `modernc.org/sqlite` vector store (`hustle.db`), `CachingLLM` (SHA-256 prompt deduplication), `PromptOptimizer` (epsilon-greedy multi-armed bandit), `Identity` (`did:hustle` Ed25519 signing), `RingLogBuffer` (stdout tee for `/api/log`), `HustleProtocol` (`hustle://` router), `A2ABroker` + `MemorySwarm`, `Healer`, and the embedded Consolidated Dashboard (`orchestrator/dashboard.html`).
- **`hustle/*` (31 Domain Modules -> 36 Protocol Handlers)**:
  - **Tier 1 (Direct Markets & Arbitrage):** `trading` (CoinGecko/Binance/Kraken/Coinbase + RSI/SMA/MACD/Bollinger), `stocks` (Alpaca), `domains`, `retail` (FBA), `realestate` (STR).
  - **Tier 2 (Commerce & Products):** `affiliate`, `products` (Digital Product Factory), `ecommerce`, `saas` (ideate + single-file HTML/JS MVP), `publishing` (`kdp`), `pod`.
  - **Tier 3 (Content & Publishing):** `content` (blog/seo/newsletter/thread + 100K expander + static HTML/RSS CMS), `publisher` (WordPress/Newsletter/Shopify + `calendar`).
  - **Tier 4 (Growth, Media & B2B Funnel):** `research`, `curation` (+ `rss`), `social` (Bluesky/Twitter/LinkedIn), `media` (FFmpeg vertical Shorts + `/api/hymns`), `outreach`, `crm` (+ `leadgen` & `synergy_leadgen`).
  - **Tier 5 (Confluence & 11-Module AI Agency):** `confluence`, `automation`, `bi`, `finance`, `legal`, `design`, `devagency`, `marketing`, `localization`, `support`, `careers`, `qa`.
- **`tray/` (`bin/aimm-tray.exe`)**: Windows `-H=windowsgui` system tray controller with single-instance mutex, auto-managed hidden SSH port-forwarding (`localhost:8082 -> hetzner:8082`), `BelowNormal` local server toggle, and 1-click quick dispatches.

## 3. Critical Operational Invariants Learned
1. **Cloudflare Bot Fight / WAF vs Python `urllib` & Go `net/http`:** `aimoneymachine.site` sits behind Cloudflare, which returns `403 Forbidden` on default `Python-urllib/*` user agents. Every WordPress REST client (`fixes/expand.py`, `fixes/batch_publisher.py`, `content_audit.py`, `hustle/publisher/wordpress.go`) MUST send a browser-compatible `User-Agent` header and load credentials from `.env` rather than source literals.
2. **Dashboard & Smoke Contracts:** `orchestrator/api_dashboard_test.go` requires `"Consolidated Dashboard"` in `/`, and `scripts/smoke_prod.py` requires `"diskPct"`, `"dispatchAction"`, and `"dispatched successfully"` on `POST /api/dispatch`.
3. **Nondestructive & Low-Priority Discipline:** Heavy builds and local daemons run at `BelowNormal` (`nice -n 10` on Linux), superseded files are archived in `archive/` and `fixes/archive/` (never deleted), compiled binaries are preserved, and `.memory/`, `.pi/`, and `hustle.db` are tracked in git while secrets (`.env`, `node.key`, `token_*.json`) remain strictly gitignored.
