# Release Notes - v1.0.0-alpha.63

## "Real AI Integration" — Phase 3

This release bridges the gap between a mock-LLM prototype and a real autonomous agent system. The machine can now **think for free** using local LLMs.

### New Core Features

- **OpenAI-Compatible LLM Provider**: Connects to LM Studio, Ollama, vLLM, or any OpenAI-compatible server. Auto-detects available models. Graceful fallback to MockLLM if no server is running.
- **Real Embedding Provider**: Generates actual vector embeddings via local Nomic/embed models for semantic memory search.
- **Agent Loop (Observe → Think → Act → Learn → Evaluate)**: The LLM acts as the "brain" — it reads memory context, decides which `hustle://` URI to execute next, observes the result, and plans the next step. This is the core autonomous execution paradigm.
- **Multi-Agent Orchestrator**: Run multiple specialized agents (research, content, trading, social) concurrently with live status monitoring.
- **HustlePlan Strategic Planner**: Ask the LLM to analyze your system state and generate 5 prioritized hustle plans.
- **Content Hustle Module**: Generate blogs, newsletters, SEO articles, and social media threads. Saves to markdown with YAML frontmatter. Includes topic brainstorming.

### New CLI Modes

```bash
./bin/orchestrator -agent -agent-iterations 20    # LLM-driven autonomous loop
./bin/orchestrator -autoplan                       # LLM generates strategy, then executes
```

### Bug Fixes

- Fixed duplicate `trading.PriceFetcher` initialization in `main.go`
- Fixed missing `-seed` flag declaration
- Fixed `go.work` missing `./hustle/content` module

### Known Issues

- **🔴 Windows build failure**: `go-sqlite3` CGO incompatibility with gcc 15.2.0. Migration to `modernc.org/sqlite` (pure Go) is planned.
- **No tests yet** for agent_loop, openai_compat, or content modules.
- Social posting providers are stubs (no real API calls).
- Healer loop is not closed (diagnoses but cannot apply and verify fixes).

### Upgrade Notes

If you have LM Studio running with a model loaded, the orchestrator will automatically connect on startup. No configuration required. Set `LLM_BASE_URL` environment variable to override the default `http://localhost:1234/v1`.

---

*Previous: v1.0.0-alpha.62 — "Fully Automated Luxury Protocol" (stable monorepo)*
