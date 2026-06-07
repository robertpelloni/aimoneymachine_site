# Architectural Observations & Preferences

## Core Traits
- **Protocol-Driven:** The system heavily relies on `hustle://` URIs for decoupling. This allows for easy mesh routing, manual testing, and LLM-driven agent orchestration through the same interface.
- **Tiered Memory:** L1 (Volatile/Events), L2 (Successes/Discoveries), L3 (System/Long-term). This separation prevents "context rot" in long-running sessions.
- **Go 1.24.3:** Strict adherence to the latest Go toolchain for performance and security.
- **A2A Mesh:** Distributed by design. No single point of failure; memory and status are replicated across peers.
- **Zero-Cost AI:** Free local LLMs (LM Studio, Ollama) make all decisions. No paid API calls required for the core agent loop.

## Design Preferences
- **Minimal Dependencies:** Prefer standard library or high-signal open source tools (like `sqlite-vec`).
- **Fail-Fast:** Rollback logic is first-class. If a sync or build fails, the system reverts immediately.
- **Autonomous Evolution:** The system should generate its own tasks and code improvements (ChainDiscoverer + Agent Loop).
- **Graceful Degradation:** If local LLM is unavailable, fall back to MockLLM. If sqlite-vec is missing, fall back to Go-level cosine similarity. If CGO fails, fall back to in-memory only.

## Discovered Optimizations
- **LLM Sentiment Confluence:** Technical signals are noisy; filtering them through LLM-extracted market sentiment significantly reduces false positives in Trading.
- **Mesh Aggregation:** Centralizing status in L1 memory allows the Dashboard to remain stateless and high-performance.
- **Agent Loop > Daemon Mode:** The LLM-driven agent loop (Observe→Think→Act→Learn) produces more varied and strategic behavior than fixed-interval daemon scheduling. The daemon is still useful for predictable baseline execution.
- **Content Generation is Highest-ROI Module:** With a local LLM, generating blog/newsletter/SEO content costs zero marginal compute and produces directly monetizable output. This is the "easiest money" hustle.

## Deployment Verification (v1.0.0-alpha.63)
- **Real LLM Connection:** LM Studio at localhost:1234 serving Gemma 27B + Nomic embeddings. Chat completions and embedding generation both verified working.
- **Agent Loop:** Think→Act cycle verified with real LLM generating `hustle://` URIs based on context.
- **Build:** Blocked on Windows by CGO/gcc incompatibility with `go-sqlite3`. Needs migration to `modernc.org/sqlite`.
- **API Performance:** Verified real-world price fetching with CoinGecko typically completes under 2s.
- **Mesh Resilience:** Mesh status aggregation is asynchronous and does not block the primary task loop unless many peers are unresponsive.
- **Wealth Preservation:** ROI audits successfully detect deficits from persistent storage (`ledger.json`) and propose corrective termination.

## Known Technical Debt
- `go-sqlite3` requires CGO — should migrate to `modernc.org/sqlite` (pure Go) for Windows compatibility
- Council debate scoring is mock heuristic, not LLM-driven analysis
- Healer `Verify()` is a mock that always returns true after 2nd attempt
- Rollback handler `Execute()` is a stub with no real git revert logic
- Anthropic provider `Generate()` returns a hardcoded string, not a real API call
- Social posting providers (Twitter, LinkedIn) are stubs that print but don't post
- No `.env` file support — all config via environment variables only
