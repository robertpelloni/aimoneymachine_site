# 🚧 UNDER CONSTRUCTION — ALPHA STATE! 🚧

# AI Hustle Machine — Fully Automated Luxury Protocol

A self-orchestrating, LLM-driven autonomous agent system that runs revenue-generating "hustles" using free local AI models. The machine thinks, acts, learns, and evolves — without human intervention.

**Current Version:** 1.0.0-alpha.81 · **Status:** Active Development · **Language:** Go 1.25.0

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

## 🌐 API Reference

The Orchestrator provides a REST API for remote monitoring and control.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/status` | GET | Current system health, version, and profit |
| `/agents` | GET | Status of all active specialized agents |
| `/content` | GET | List of generated files in the content library |
| `/chains` | GET | All registered workflow chains |
| `/strategy`| GET | Shared alpha/strategies from the mesh |
| `/dispatch`| POST | Execute a `hustle://` URI (JSON: `{"uri": "..."}`) |
| `/healer` | POST | Trigger manual diagnosis (JSON: `{"issue": "..."}`) |
| `/sync` | POST | Trigger repository synchronization |
| `/register`| POST | Register a new mesh peer (JSON: `{"id": "...", "url": "..."}`) |

## 🚀 Quick Start

### Prerequisites
- **Go 1.25.0** (required toolchain)
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
git clone https://github.com/robertpelloni/fully_automated_gay_luxury_space_communism
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

## 🧩 Hustle Modules

| Module | Protocol URI | What It Does | Revenue Model |
|--------|-------------|-------------|---------------|
| **Research** | `hustle://research?query=TOPIC` | Web search, sentiment extraction, alpha discovery | Feeds Trading + Content |
| **Curation** | `hustle://curation?topic=TOPIC` | RSS aggregation, newsletter blurb generation | Ad revenue, subscribers |
| **Social** | `hustle://social?platform=Twitter&topic=TOPIC` | LLM-generated posts with hashtags | Audience growth → monetization |
| **Trading** | `hustle://trading?symbol=BTC` | Technical analysis (SMA, RSI, Divergence), strategy execution | Trade profits |
| **Content** | `hustle://content?topic=X&type=blog` | Blog/newsletter/SEO article/social thread generation | Ads, affiliates, SEO traffic |

## 📋 Status

See [ROADMAP.md](ROADMAP.md) for the phased plan, [TODO.md](TODO.md) for task tracking, and [VISION.md](VISION.md) for the long-term vision.
