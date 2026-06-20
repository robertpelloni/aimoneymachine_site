[PROJECT_MEMORY]

# AI Hustle Machine — Project Memory & Architecture Summary

## 1. Core Vision & Philosophy
The **AI Hustle Machine** (also referred to as the "Fully Automated Luxury Protocol") is a self-orchestrating, LLM-driven autonomous agent system designed to run revenue-generating "hustles" using local or free LLM models. The system operates on continuous autonomous execution: Observe → Think → Act → Learn → Evaluate. It is designed to require zero human intervention once deployed, aggressively targeting high-ROI, low-maintenance workflows.

## 2. Monorepo Architecture
The project is built in **Go 1.25.0** using a `go.work` monorepo structure. It moved away from submodules to a unified repository to prevent sync issues.
*   **Orchestrator (`/orchestrator`)**: The core engine. It manages the Agent Loop, multi-agent Council, Memory tiers, Ledger, Scheduler, and Healer.
*   **Hustle Modules (`/hustle/*`)**: Specialized domain execution environments:
    *   `content`: Markdown CMS, static site generation, blogs, newsletters.
    *   `curation`: RSS aggregation and summarizing.
    *   `research`: Web search (Tavily/Brave) and alpha discovery.
    *   `social`: Twitter/LinkedIn automated posting and engagement.
    *   `trading`: Crypto trading and TA using CoinGecko.

## 3. Key Design Patterns & Protocols
*   **The `hustle://` Protocol**: Internal routing mechanism. The LLM decides actions by emitting URIs like `hustle://content?topic=AI&type=blog` or `hustle://trading?symbol=BTC`. The orchestrator parses this and dispatches it to the relevant module.
*   **Agent Loop**: The main autonomous driver. It observes context from memory, uses the LLM to *think* (produce a `hustle://` URI), *acts* on it, and evaluates the financial/system result.
*   **Tiered Memory**:
    *   `L1`: Immediate short-term context.
    *   `L2`: Episodic memory.
    *   `L3`: Long-term vector-embedded memory (using pure-Go cosine similarity fallback or `sqlite-vec`).
*   **Waterfall LLM & OpenAI Compatibility**: Built to interact with local servers (LM Studio, Ollama) via standard OpenAI schemas. Includes fallback wrappers (`WaterfallLLM`) for resilience.
*   **Caching & Optimization (Phase 5 additions)**:
    *   `CachingLLM`: SHA-256 hashed prompt caching with TTL to save tokens and time on duplicate requests.
    *   `PromptOptimizer`: Uses an epsilon-greedy multi-arm bandit algorithm to randomly A/B test prompt variations and organically adopt the highest-performing prompts over time.

## 4. Self-Healing & Wealth Preservation
*   **The Healer**: A self-reflection module. If a task fails or an error is thrown, the Healer queries the LLM with the error logs, generates a diagnostic strategy, applies it, and audits if the fix was successful.
*   **ROI Auditing**: The `Ledger` tracks financial performance for every registered task. The Scheduler continuously audits this ledger. If a hustle runs a deficit past a certain threshold, the system autonomously terminates and unregisters it ("Wealth Preservation").

## 5. Mesh & Swarm Intelligence
*   **A2A (Agent-to-Agent)**: The system supports decentralized peering. Nodes discover each other and broadcast their status, sharing insights ("alpha") and aggregating total mesh-wide profit ("Luxury Space Communism" collective wealth tracking).

## 6. Engineering Decisions & Constraints
*   **Pure Go SQLite**: Shifted from `go-sqlite3` to `modernc.org/sqlite` to permanently fix Windows CGO compilation blockers.
*   **The Executive Protocol (`sync.sh`)**: A rigid bash script used to maintain repository hygiene. It handles upstream tracking, intelligent dual-direction merging between feature branches and `main`, submodule updates, and version synchronization before pushing.
*   **Documentation Governance**: The project relies heavily on strict markdown file maintenance. `TODO.md`, `ROADMAP.md`, `CHANGELOG.md`, `VISION.md`, `HANDOFF.md`, and `VERSION.md` are the ground truth for system state and cross-agent communication. Any version bumps require atomic commits documenting the exact version string.
