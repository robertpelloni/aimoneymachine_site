# AI Hustle Machine — Fully Automated Luxury Protocol

A self-orchestrating, LLM-driven autonomous agent system that runs revenue-generating "hustles" using free local AI models. The machine thinks, acts, learns, and evolves — without human intervention.

**Current Version:** v1.0.0-alpha.63 · **Status:** Active Development · **Language:** Go 1.24.3

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ORCHESTRATOR (Core)                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ hustle:// │  │  Agent   │  │  Memory  │  │  Ledger  │   │
│  │ Protocol  │  │   Loop   │  │ L1/L2/L3 │  │  (ROI)   │   │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘   │
│        │             │             │             │          │
│  ┌─────┴────┐  ┌─────┴────┐  ┌─────┴────┐  ┌─────┴────┐   │
│  │ Scheduler │  │  Healer  │  │  Council │  │  A2A     │   │
│  │ (Daemon)  │  │ (Fix)    │  │ (Debate) │  │  Mesh    │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              LLM Waterfall (Free → Paid)             │  │
│  │  LM Studio → Ollama → OpenRouter Free → Paid APIs   │  │
│  └──────────────────────────────────────────────────────┘  │
└──────────────────────────┬──────────────────────────────────┘
                           │ hustle:// URIs
        ┌──────────┬───────┼───────┬──────────┐
        ▼          ▼       ▼       ▼          ▼
   ┌────────┐ ┌────────┐ ┌──────┐ ┌──────┐ ┌────────┐
   │Research│ │Curation│ │Social│ │Trading│ │Content │
   │(Search)│ │(RSS)   │ │(Post)│ │(Crypto)│ │(Blog)  │
   └────────┘ └────────┘ └──────┘ └──────┘ └────────┘
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.24.3** (required toolchain)
- **LM Studio** (recommended) or **Ollama** for free local LLM
- A loaded model (e.g., `gemma-4-26b`, `qwen3-27b`)

### 1. Start Your Local LLM
```bash
# Option A: LM Studio — load a model and start the server (default port 1234)
# Option B: Ollama
ollama serve
ollama pull gemma3:27b
```

### 2. Build & Run
```bash
git clone --recursive https://github.com/robertpelloni/fully_automated_gay_luxury_space_communism
cd fully_automated_gay_luxury_space_communism
go work sync
./build.sh
```

### 3. Run Modes

```bash
# 🤖 AUTONOMOUS AGENT — LLM drives all decisions (the main event)
./bin/orchestrator -agent -agent-type general -agent-iterations 50

# 🧠 AUTO-PLAN — LLM generates strategy, then executes it
./bin/orchestrator -autoplan

# 🔁 DAEMON — Scheduled tasks run on timers (no LLM decision-making)
./bin/orchestrator -daemon -real-prices

# 🖥️ INTERACTIVE — Manual control menu
./bin/orchestrator -interactive

# 📊 DASHBOARD — Live terminal UI
./bin/orchestrator -dashboard

# 🌐 API SERVER — HTTP endpoints for remote control
./bin/orchestrator -api 8080

# 🧪 SINGLE TASK — Run one hustle directly
./bin/orchestrator -hustle trading -real-prices
./bin/orchestrator -uri "hustle://content?topic=AI+agents&type=blog"
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_BASE_URL` | `http://localhost:1234/v1` | OpenAI-compatible LLM server URL |
| `LLM_MODEL` | *(auto-detect)* | Specific model name (auto-detected if empty) |
| `LLM_API_KEY` | *(from `LM_STUDIO_KEY`)* | API key (not needed for local servers) |
| `EMBED_BASE_URL` | `http://localhost:1234/v1` | Embedding server URL |
| `EMBED_MODEL` | *(auto-detect)* | Embedding model (auto-detects `nomic`/`embed`) |
| `TAVILY_API_KEY` | *(none)* | For real web search (falls back to mock) |

## 🧩 Hustle Modules

| Module | Protocol URI | What It Does | Revenue Model |
|--------|-------------|-------------|---------------|
| **Research** | `hustle://research?query=TOPIC` | Web search, sentiment extraction, alpha discovery | Feeds Trading + Content |
| **Curation** | `hustle://curation?topic=TOPIC` | RSS aggregation, newsletter blurb generation | Ad revenue, subscribers |
| **Social** | `hustle://social?platform=Twitter&topic=TOPIC` | LLM-generated posts with hashtags | Audience growth → monetization |
| **Trading** | `hustle://trading?symbol=BTC` | Technical analysis (SMA, RSI, Divergence), strategy execution | Trade profits |
| **Content** | `hustle://content?topic=X&type=blog` | Blog/newsletter/SEO article/social thread generation | Ads, affiliates, SEO traffic |

## 🧠 Core Systems

### hustle:// Protocol
All modules communicate via custom URIs (`hustle://module?params`). This decouples modules and allows both manual and LLM-driven orchestration through the same interface.

### Agent Loop (Observe → Think → Act → Learn → Evaluate)
The autonomous agent loop uses the LLM as the "brain":
1. **Observe**: Gathers recent memory, financial state, past successes
2. **Think**: LLM decides the next `hustle://` URI to execute
3. **Act**: Executes via the protocol
4. **Learn**: Stores outcome in memory, promotes milestones to L2
5. **Evaluate**: Stops if too many errors or wealth preservation triggers

### LLM Waterfall (Zero-Cost AI)
Cascades through providers on failure:
```
LM Studio (local, free) → Ollama (local, free) → OpenRouter (free tier) → Paid APIs
```

### Tiered Memory
| Tier | Name | Purpose | Persistence |
|------|------|---------|-------------|
| L1 | Scratchpad | Volatile events, recent activity | In-memory + SQLite |
| L2 | Vault | Successes, discoveries, chain evolution | In-memory + SQLite |
| L3 | Archive | Long-term system knowledge | In-memory + SQLite |

### Wealth Preservation
- Ledger tracks Revenue/Expense per hustle
- Healer runs ROI audits — flags underperforming hustles
- Scheduler auto-unregisters tasks hemorrhaging capital
- Agent loop stops if profit drops below -$1000

### Multi-Agent Council
Bull (argue FOR) · Bear (argue AGAINST) · Critic (synthesize) — weighted voting for strategic decisions.

### A2A Mesh
Federated peer-to-peer networking with:
- Direct message routing
- Topic pub/sub (`alpha_discovery`, `trade_execution`, `swarm_sync`)
- Delta-sync memory reconciliation (checksum-based)
- Mesh-wide status aggregation

## 🛠️ Development

### Build
```bash
./build.sh
```

### Test
```bash
cd orchestrator && go test ./... -v
cd ../hustle/trading && go test ./... -v
cd ../hustle/curation && go test ./... -v
```

### Project Structure
```
orchestrator/          # Core: protocol, memory, scheduler, healer, LLM, mesh
hustle/
  ├── research/        # Web search + alpha discovery
  ├── curation/        # RSS feed curation
  ├── social/          # Social media posting
  ├── trading/         # Crypto trading + technical analysis
  └── content/         # Blog/newsletter/SEO content generation
.agent/workflows/      # Engineering workflows (office-hours, ship, review)
docs/designs/          # Design documents
```

## 📋 Status

See [ROADMAP.md](ROADMAP.md) for the phased plan, [TODO.md](TODO.md) for task tracking, and [VISION.md](VISION.md) for the long-term vision.
