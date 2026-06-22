# Quick Start Guide

## Prerequisites

1. **Go 1.24.3** - [Download](https://go.dev/dl/)
2. **LM Studio** (recommended) or **Ollama** for local LLM

## Option 1: LM Studio Setup (Recommended)

1. Download LM Studio from https://lmstudio.ai
2. Load a model (e.g., `gemma-4-26b-it` or `qwen3-27b`)
3. Start the server (default: `http://localhost:1234/v1`)
4. Optionally load an embedding model (e.g., `nomic-embed-text`)

## Option 2: Ollama Setup

```bash
ollama serve
ollama pull gemma3:27b
# Server runs at http://localhost:11434/v1
```

## Build

```bash
git clone https://github.com/robertpelloni/fully_automated_gay_luxury_space_communism
cd fully_automated_gay_luxury_space_communism
go work sync
./build.sh
```

## Test Run (No API Keys Required)

The system works in **stub/mock mode** without any API keys:

```bash
# Interactive menu
./bin/orchestrator -interactive

# Autonomous agent (50 iterations)
./bin/orchestrator -agent -agent-iterations 50

# Daemon mode with scheduled tasks
./bin/orchestrator -daemon
```

## Production Mode (Real APIs)

1. Copy and configure environment variables:
```bash
copy .env.example .env
```

2. Fill in your API keys:
- `TAVILY_API_KEY` - Get from https://tavily.com
- `TWITTER_BEARER_TOKEN` - Get from https://developer.twitter.com
- `LINKEDIN_ACCESS_TOKEN` - Get from https://www.linkedin.com/developers

3. Run with real APIs:
```bash
./bin/orchestrator -agent -agent-type content -agent-iterations 20
```

## Verify FreeLLM Proxy

If you want to use a FreeLLM proxy for all LLM connections:

```bash
set FREE_LLM_PROXY_URL=http://your-proxy:8000/v1
./bin/orchestrator -agent -agent-iterations 10
```

The system will automatically use the proxy instead of the local LLM server.

## Verify Research APIs

Test real web search:

```bash
set TAVILY_API_KEY=your_key_here
./bin/orchestrator -uri "hustle://research?query=AI agents 2026"
```

## Verify Social Posting

Test Twitter posting:

```bash
set TWITTER_BEARER_TOKEN=your_token_here
./bin/orchestrator -uri "hustle://social?platform=Twitter&topic=AI automation"
```

## Common Issues

### CGO Build Error on Windows
✅ **Fixed!** The project now uses `modernc.org/sqlite` (pure Go) instead of `github.com/mattn/go-sqlite3` (CGO).

### LLM Not Responding
- Make sure LM Studio server is running
- Check `LLM_BASE_URL` or `FREE_LLM_PROXY_URL`
- Try: `curl http://localhost:1234/v1/models`

### API Errors
- Check your API keys in `.env`
- Verify the API is accessible
- Some APIs have rate limits on free tiers

## Next Steps

1. Start with `-interactive` mode to explore
2. Try `-agent` mode for autonomous operation
3. Configure API keys for production use
4. Read [VISION.md](VISION.md) and [ROADMAP.md](ROADMAP.md)

---
**Current Version:** v1.0.0-alpha.66
**Status:** ✅ Build Fixed on Windows, ✅ FreeLLM Proxy Support, ✅ Real API Integration Ready