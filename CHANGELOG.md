# Changelog

## [1.0.0-alpha.37] - 2026-06-02
### Added
- **Global Stash Logic**: Refactored `sync.sh` to use a robust global stash/pop strategy for the Executive Protocol.
- **Profit Analysis Engine**: Implemented autonomous financial evaluation that logs strategy optimization suggestions to L2 memory.
- **Multi-Symbol Trading**: Added `ExecuteAll` capability to the Trading module for automated watchlist scanning.
- **Repository Hygiene**: Standardized `.gitignore` and purged all compiled binaries from the source tree.

## [1.0.0-alpha.36] - 2026-06-02
### Added
- **AI-Driven Self-Healing (BETA)**: Enhanced the `Healer` module to utilize LLM for context-aware system diagnosis and fix generation.
- **Intelligent Merge Engine (STABLE)**: Finalized Step 2 of the Executive Protocol, supporting automated forward/reverse merges and permanent sync integration testing.
- **Mesh-wide Trading Coordination**: Integrated `TradingModule` with the A2A Broker for broadcasting trade decisions across the mesh.
- **Distributed Curation**: `CurationModule` now supports multi-feed management and automatically broadcasts content discoveries to the mesh.
- **Persistent Financial Ledger**: Implemented transaction persistence in `ledger.json` and a `hustle://ledger` protocol handler.
- **Dynamic Research Analysis**: Enhanced the Research module with AI-driven symbol extraction and a sentiment synthesis handoff for Trading confluence.

## [1.0.0-alpha.35] - 2026-06-02
### Added
- **Automated Delta-Sync (STABLE)**: Completed the federated memory reconciliation loop.
- **Protocol Handshaking**: Enhanced `hustle://swarm` with peer identification for bidirectional data exchange.
- **Remote URI Execution**: Added direct message listeners that execute incoming `hustle://` commands from the mesh.
- **Mesh-Aware API**: Updated Orchestrator API and protocol handlers to support routed sync responses.
- **OAuth2 Lifecycle**: Introduced `hustle/social/oauth.go` to handle token exchange and persistence for social platforms.
