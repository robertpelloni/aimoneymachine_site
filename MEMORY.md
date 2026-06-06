# Architectural Observations & Preferences

## Core Functionality Summary

### 1. Orchestrator (`orchestrator/`)
- **Protocol Handler:** Implements `hustle://` URI routing for decoupled module execution.
- **A2A Broker:** Distributed message broker for agent-to-agent communication and mesh scaling.
- **Tiered Memory:** L1 (Volatile), L2 (Vault/Successes), L3 (Archive) with semantic and vector search (SQLite-vec).
- **Self-Evolving Scheduler:** Manages persistent tasks with autonomous interval adjustment based on ROI.
- **Healer:** Context-aware self-healing with "Wealth Preservation" ROI audits and automated task termination.
- **Multi-Agent Council:** Weighted voting for high-stakes strategy verification.
- **Service Layer:** HTTP API for remote orchestration, status aggregation, and mesh management.

### 2. Research Hustle (`hustle/research/`)
- **LLM-Driven Discovery:** Utilizes Tavily and LLMs to extract market sentiment and ticker symbols.
- **Reporting:** Generates automated PDF reports for discovered alpha.

### 3. Curation Hustle (`hustle/curation/`)
- **RSS Integration:** Fetches and processes real feeds (e.g., Hacker News) for automated content generation.
- **Content Pipeline:** Summarizes and blurbs content for downstream social posting.

### 4. Social Hustle (`hustle/social/`)
- **Multi-Platform:** Authenticated posting logic for Twitter and LinkedIn.
- **OAuth2 Flow:** State persistence for production-grade API interaction.

### 5. Trading Hustle (`hustle/trading/`)
- **Market Confluence:** Filters technical signals (RSI/SMA) through Research Sentiment from memory.
- **Real-world Pricing:** Integrated CoinGecko fetcher with HTTP timeouts for production stability.

### 6. Mesh Swarm & Luxury Logic
- **Federated Intelligence:** Distributed memory and status aggregation across the mesh.
- **Luxury Discovery:** Focus on high-ROI, low-maintenance autonomous workflow generation.

## Design Preferences
- **Minimal Dependencies:** Prefer standard library or high-signal tools.
- **Fail-Fast:** Robust stashing and rollback logic in `sync.sh`.
- **Autonomous Evolution:** System independently discovers and optimizes its own workflows.

## Go Version
- Strictly pinned to **Go 1.24.3** across all modules and workspace.
