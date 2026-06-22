# AI Hustle Machine — Project Roadmap

## Project Purpose

An autonomous AI-powered trading, content, and income automation system called "AI Hustle Machine." It runs trading bots on Binance.US, generates blog content via LLM, does web research, manages a WordPress site (aimoneymachine.site), and auto-posts to social media — all orchestrated through a central Go-based scheduler.

## Current State

### Running on Hetzner (24/7)
- **Orchestrator** (systemd) — 7 trading bots: BTC, ETH, SOL, DOGE, XRP, ADA, LINK
- **FreeLLM** (systemd) — LLM proxy with 15+ provider API keys, Tokdiet disabled, port 4000
- **Dashboard** (systemd) — Web UI at port 8083 with per-bot dashboards
- **Content expansion** (systemd) — Expanding 217 posts to 100K chars each, ~60 done
- **fwber backend** (PM2) — Node.js API on port 4002, nginx proxied

### Running Locally (previously, now stopped)
- Local orchestrator instances killed
- Local watchdog task removed
- Everything runs on Hetzner now

### WordPress Site
- aimoneymachine.site — Twenty Twenty-Five theme, dark redesign CSS injected
- 180+ blog posts on AI/automation topics
- LinkedIn auto-posting active (~2 posts so far)
- Twitter keys saved but not yet active

## Key Decisions Made

| Decision | Rationale |
|----------|-----------|
| **Hetzner deployment** | Local machine keeps dying (sleep, process crashes, watchdog fails). Linux systemd services are reliable. |
| **FreeLLM Linux fork** | FreeLLM is Windows-only (kernel32.dll, systray). Created `freellm-linux` branch removing GUI deps, adding UIServer, disabling Tokdiet. |
| **LLM → Ollama → LiteLLM → FreeLLM** | Tried multiple LLM proxies. FreeLLM has the most free provider integrations. Needed to strip GUI and disable Tokdiet for headless Linux. |
| **Binance.US instead of Binance.com** | User is in US. Binance.com blocks US IPs with 451 error. |
| **RSI-based strategy** | Simple, proven, works without LLM. BUY at RSI<30, SELL at RSI>70, with SMA(5) confirmation. |
| **Stop-loss 5%, Trailing stop 8%** | Hardcoded defaults to protect positions. |
| **Content expansion in chunks** | LLMs can't output 100K chars in one call. Each post built from ~10 chunks of 25K chars each. |
| **Per-symbol trading modules** | Each bot gets its own history/RSI so BTC doesn't pollute ETH's calculations. |
| **Port mapping** | 4000=FreeLLM, 4002=fwber API, 8082=orchestrator API, 8083=dashboard. Documented in /etc/services on both Linux and Windows. |

## Milestones

- [x] Initial prototype — basic RSI trading on Binance
- [x] WordPress site setup with 180+ auto-generated posts
- [x] Content expansion script — 60/217 posts at 100K chars
- [x] Web dashboards — main dashboard + per-bot pages
- [x] LinkedIn auto-posting via curation chain
- [x] Stop-loss + trailing stop implementation
- [x] FreeLLM Linux fork — headless build (freellm-linux branch)
- [x] Hetzner deploy — all services via systemd/PM2
- [x] FreeLLM API keys loaded (15+ providers)
- [x] Tokdiet bypass — direct provider routing
- [ ] All 217 posts expanded to 100K chars
- [ ] Twitter auto-posting (keys saved, need daemon restart with TWITTER_* vars)
- [ ] Newsletter subscriber growth
- [ ] Trading bots generating real P&L (need market volatility)

## Open Questions / Risks

- **FreeLLM stability on Linux** — the headless fork is new, may have edge cases
- **Binance.US balance** — $55 balance, trades will fire when RSI hits extremes
- **Content expansion speed** — ~60/217 done, ~25K chars per LLM call, ~10 calls per post
- **No real P&L yet** — strategy works but market has been flat, no trades executed
