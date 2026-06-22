# Deployment & Environment Setup

## Prerequisites

- **Go 1.24.0+** (Required toolchain)
- **LM Studio** (Recommended) or **Ollama** for free local LLM inference
- **A loaded model** in LM Studio (e.g., `gemma-4-26b-it`) or Ollama (e.g., `gemma3:27b`)

## Quick Start

### 1. Start Your Local LLM Server

**Option A: LM Studio** (recommended for GPU)
1. Open LM Studio
2. Load a model (e.g., `gemma-4-26b-it-qat-heretic`)
3. Start the server (default: `http://localhost:1234/v1`)
4. Also load an embedding model (e.g., `nomic-embed-text-v1.5`) for vector search

**Option B: Ollama**
```bash
ollama serve
ollama pull gemma3:27b
# Server runs at http://localhost:11434/v1
```

### 2. Clone & Build

```bash
git clone https://github.com/robertpelloni/fully_automated_gay_luxury_space_communism
cd fully_automated_gay_luxury_space_communism
go work sync
./build.sh
```

> **Note**: The project uses `modernc.org/sqlite`, a pure-Go SQLite driver. No CGO or `gcc` is required for compilation.

### 3. Verify LLM Connection

```bash
# Check LM Studio is serving models
curl http://localhost:1234/v1/models

# Test a chat completion
curl -s http://localhost:1234/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gemma-4-26b-it","messages":[{"role":"user","content":"Say hello"}],"max_tokens":10}'
```

### 4. Run

```bash
# Autonomous agent (the main event — LLM drives all decisions)
./bin/orchestrator -agent -agent-iterations 20

# LLM generates strategy, then executes it
./bin/orchestrator -autoplan

# Daemon mode with scheduled tasks + real crypto prices
./bin/orchestrator -daemon -real-prices

# Interactive menu
./bin/orchestrator -interactive

# API server for remote control
./bin/orchestrator -api 8080
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_BASE_URL` | `http://localhost:1234/v1` | OpenAI-compatible LLM server |
| `LLM_MODEL` | *(auto-detect)* | Model name (auto-detected from /models) |
| `LLM_API_KEY` | *(from `LM_STUDIO_KEY`)* | API key (not needed for local) |
| `LLM_CACHE_ENABLED` | `true` | Enable SQLite-based response caching |
| `EMBED_BASE_URL` | `http://localhost:1234/v1` | Embedding server URL |
| `EMBED_MODEL` | *(auto-detect)* | Embedding model name |
| `TAVILY_API_KEY` | *(none)* | Real web search (research module) |
| `ANTHROPIC_API_KEY` | *(none)* | Anthropic provider in LLM waterfall |

## Mesh Deployment (Cluster Setup)

To launch a federated luxury mesh:

1. **Seed Node**:
```bash
./bin/orchestrator -daemon -api 8080 -real-prices
```

2. **Peer Node**:
```bash
./bin/orchestrator -daemon -api 8081 -seed http://seed-ip:8080 -real-prices
```

## Performance Monitoring

- **Global Ledger**: Monitor `ledger.json` across all nodes.
- **Terminal Dashboard**: Run `./bin/orchestrator -dashboard` to see aggregated mesh profits.
- **Wealth Preservation**: The system automatically audits ROI and terminates underperforming hustles.
- **Agent Loop Status**: When running in `-agent` mode, real-time status is visible on the Dashboard (`-dashboard`).

## Automated Maintenance

The Orchestrator includes a self-healing `sync` protocol. To trigger manually:

```bash
./bin/orchestrator -uri "hustle://sync"
```
