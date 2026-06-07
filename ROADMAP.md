# Project Roadmap

## Phase 1: Federated Foundation (v1.0.0-alpha.1 – v1.0.0-alpha.30) ✅ COMPLETE
- Tiered Memory (L1/L2/L3)
- Basic Multi-Agent Council
- Agent-to-Agent (A2A) Broker
- Initial Hustle Modules (Research, Social, Curation)
- hustle:// Protocol routing

## Phase 2: Intelligence & Autonomy (v1.0.0-alpha.31 – v1.0.0-alpha.43) ✅ COMPLETE
- Delta-Sync Memory Swarming
- Chain Orchestration & Discovery
- Autonomous Luxury Logic
- Mesh Aggregation & Status Scaling
- Wealth Preservation (ROI auditing → auto-termination)

## Phase 3: Real AI Integration (v1.0.0-alpha.63 – v1.0.0-alpha.75) 🔧 IN PROGRESS
> The critical transition from mock LLM to real local AI, enabling actual autonomous agent loops.

- [x] OpenAI-compatible LLM provider (LM Studio, Ollama, vLLM)
- [x] Real embedding provider (Nomic via LM Studio)
- [x] Agent Loop (Observe → Think → Act → Learn → Evaluate)
- [x] Multi-Agent Orchestrator (concurrent agent loops)
- [x] Content Hustle module (blog, newsletter, SEO, social thread)
- [x] HustlePlan strategic planner (LLM-generated plans)
- [x] Interactive menu with all 5 hustle modules
- [x] Agent mode CLI flag (`-agent`, `-autoplan`)
- [ ] **Build system working** — CGO/sqlite3 compilation on Windows (gcc version mismatch)
- [ ] **All tests passing** with new LLM provider (agent_loop, openai_compat tests)
- [ ] **Ollama fallback** tested and verified
- [ ] **End-to-end agent loop test** that runs 3 iterations with real LLM
- [ ] **Content module test** (content_test.go)
- [ ] **Agent loop integration** — agent uses content hustle in its action list
- [ ] **Environment config** (.env.example + startup validation)

## Phase 4: Production Hustle Operations (v1.0.0-alpha.76 – v1.0.0-beta.10) 📋 PLANNED
> Make the hustles actually produce output and track real value.

- [ ] **Research with real web search** — Tavily API integration tested end-to-end
- [ ] **Trading with real CoinGecko data** — verified in daemon mode with -real-prices
- [ ] **Content output pipeline** — generated markdown stored, indexed, searchable
- [ ] **Content topic queue** — LLM generates topic list, agent works through it
- [ ] **RSS feed expansion** — configurable feed list (not just HN)
- [ ] **Social posting safety** — dry-run mode, human approval gate before real posts
- [ ] **PDF report generation** — fix gofpdf dependency, test real PDF output
- [ ] **Revenue estimation model** — per-hustle projected vs actual tracking
- [ ] **Daemon mode with agent loop** — agent runs continuously inside scheduler
- [ ] **Graceful shutdown** — persist state on SIGINT, resume on restart

## Phase 5: Advanced Autonomy (v1.0.0-beta.11 – v1.0.0-rc.1) 📋 PLANNED
> Self-improving, self-healing, self-optimizing system.

- [ ] **Closed-loop healer** — healer applies fixes, verifies, retries (max 3)
- [ ] **Stop hooks** — intercept before session ends, verify promises fulfilled
- [ ] **Self-optimizing prompts** — A/B test prompt variations, track win rates
- [ ] **Advanced trading indicators** — MACD, Bollinger Bands, volume analysis
- [ ] **Multi-exchange trading** — Binance, Kraken exchange plugins
- [ ] **Cross-hustle feedback** — research discoveries feed content topics, content feeds social
- [ ] **Agent memory → tool selection** — remember which tools succeed in which contexts
- [ ] **Real social API integration** — OAuth2 flow for Twitter/X, LinkedIn
- [ ] **Browser automation hustle** — web scraping for lead generation, market research
- [ ] **LLM cache** — content-addressable cache for identical prompts

## Phase 6: Federation & Scale (v1.0.0-rc.2 – v1.0.0) 📋 PLANNED
> Multi-node mesh, global leaderboard, production hardening.

- [ ] **Multi-node cluster testing** — 10+ peers with real mesh sync
- [ ] **Global profit leaderboard** — mesh-wide ranking via A2A topic
- [ ] **Decentralized identity** — DID for peer verification
- [ ] **Adaptive sync intervals** — sync more during volatility, less during calm
- [ ] **Token budget manager** — allocate context budget per tool/agent
- [ ] **L1 memory TTL** — prevent unbounded growth in long-running nodes
- [ ] **WASM sandbox** — safe code execution for discovered chains
- [ ] **Browser extension** — remote hustle monitoring dashboard
- [ ] **Production CI/CD** — GitHub Actions with real build + test
