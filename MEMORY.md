# Architectural Observations & Preferences

## Core Traits
- **Protocol-Driven:** The system heavily relies on `hustle://` URIs for decoupling.
- **Tiered Memory:** L1 (Volatile/Events), L2 (Successes/Discoveries), L3 (System/Long-term).
- **Go 1.25.0:** Strict adherence to the latest Go toolchain (v1.0.0-alpha.86).
- **A2A Mesh:** Distributed by design. No single point of failure.
- **Zero-Cost AI:** Free local LLMs (LM Studio, Ollama) make all decisions.

## Phase 3/4 Advancements
- **OpenAI Compatibility**: Standardized on OpenAI API format for local LLM integration.
- **Agent Loop**: Implemented a robust autonomous cycle (Observe→Think→Act→Learn).
- **Content Module**: Transitioned content generation from a placeholder to a first-class module with markdown output.
- **Closed-Loop Healing**: Moved beyond simple retries to LLM-verified healing.
- **Space Comms Interface**: Formalized mesh control with a dedicated dashboard section and interactive management submenu.

## Design Preferences
- **Minimal Dependencies:** Prefer standard library or high-signal open source tools.
- **Fail-Fast:** Rollback logic is first-class.
- **Autonomous Evolution:** The system should generate its own tasks and code improvements.

## Discovered Optimizations
- **LLM Response Caching:** SHA256-based prompt hashing in SQLite prevents redundant API calls, cutting latency by 90% on repeated tasks (v1.0.0-alpha.82).
- **LLM Sentiment Confluence:** Filtering signals through LLM-extracted sentiment reduces noise.
- **Mesh Aggregation:** Centralizing status in L1 memory allows the Dashboard to remain stateless.
- **Content Generation is Highest-ROI:** Zero marginal compute cost for directly monetizable output.

## Known Technical Debt
- Social posting providers (Twitter, LinkedIn) require API key hardening for production.
- Research module requires real Tavily/Brave API integration for production.
