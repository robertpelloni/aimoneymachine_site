# Changelog

## [1.0.0-alpha.35] - 2026-05-30
### Added
- **Automated Delta-Sync (STABLE)**: Completed the federated memory reconciliation loop.
- **Protocol Handshaking**: Enhanced `hustle://swarm` with peer identification for bidirectional data exchange.
- **Remote URI Execution**: Added direct message listeners that execute incoming `hustle://` commands from the mesh.
- **Mesh-Aware API**: Updated Orchestrator API and protocol handlers to support routed sync responses.

## [1.0.0-alpha.34] - 2026-05-30
### Added
- **Alpha Handoff (BETA)**: Implemented automated Research-to-Trading data flow via A2A mesh events.
- **Mesh Event Broadcaster**: Added manual event broadcasting to the interactive CLI menu.
- **Watchlist Integration**: Trading module now supports dynamic watchlist updates from mesh discovery events.

## [1.0.0-alpha.33] - 2026-05-30
### Added
- **Delta-Sync Memory Swarming (BETA)**: Implemented context-aware entry requests in `MemorySwarm`.
- **Targeted Entry Retrieval**: Added `Get(id)` methods to all memory tiers (L1/L2/L3) for specific context sharing.
- **Mesh Data Transfer**: Enhanced `hustle://swarm` protocol with `request_entry` and `provide_entry` actions for peer-to-peer data ingestion.

## [1.0.0-alpha.32] - 2026-05-30
### Added
- **Merkle Tree Context Hashing (BETA)**: Implemented `Checksum()` methods for memory entries and tiers to enable efficient mesh state comparison.
- **RSI Divergence Detection**: Added advanced momentum-based signal detection to the Trading module.
- **Delta-Aware Swarming**: Enhanced `MemorySwarm` to exchange checksums before synchronization to minimize bandwidth.
- **High-Confidence Strategy**: Decisions now utilize confluence between RSI Divergence and SMA crossovers.

## [1.0.0-alpha.31] - 2026-05-30
### Added
- **Networked Pub/Sub (BETA)**: Extended `A2ABroker` to support topic-based message forwarding across remote peers.
- **RSI Indicator**: Implemented Relative Strength Index (RSI) calculation in the Trading module.
- **Sophisticated Trading Strategy**: Decision engine now uses RSI and SMA crossovers for BUY/SELL signals.
- **Mesh Convergence**: HTTP `/message` endpoint now supports incoming topic traffic for federated event handling.

## [1.0.0-alpha.30] - 2026-05-30
### Added
- **Asynchronous Mesh (BETA)**: Implemented NATS-style Topic Pub/Sub in `A2ABroker`.
- **Trading Indicators (EXPERIMENTAL)**: Added Simple Moving Average (SMA) calculation to the Trading module.
- **Price Sourcing**: Introduced `PriceFetcher` interface for modular price data delivery.
- **Event-Driven Pub/Sub**: Added `Publish` and `SubscribeTopic` methods for broadcast messaging.

## [1.0.0-alpha.29] - 2026-05-30
### Added
- **P2P Memory Swarm (EXPERIMENTAL)**: Implemented `MemorySwarm` for federated synchronization of L2/L3 memories across the mesh.
- **`hustle://swarm` Protocol**: Added protocol handler for triggering and responding to memory sync requests.
- **Scheduled Synchronization**: Integrated memory swarming into the Orchestrator's automated task scheduler.

## [1.0.0-alpha.28] - 2026-05-30
### Added
- **Remote A2A Forwarding (BETA)**: Implemented cross-host message forwarding in `A2ABroker`.
- **Peer Discovery API**: Added `/register` endpoint for orchestrators to handshake and form a mesh.
- **Trading Module (BETA)**: Scaffolded automated trading module with strategy execution and symbol tracking.
- **Networked Dispatch**: Hardened the HTTP `/message` endpoint for receiving remote A2A traffic.

## [1.0.0-alpha.27] - 2026-05-30
### Added
- **A2A Broker (EXPERIMENTAL)**: Implemented the Agent-to-Agent message broker for inter-agent communication.
- **Mesh Messaging**: Defined `Message` and `MessageType` (Query, Command, Response, Event) structures for federated orchestration.
- **External Dispatch API**: Exposed the `hustle://` protocol via an HTTP `/dispatch` endpoint for remote task triggering.
- **CLI API Support**: Added `-api` flag to Orchestrator CLI to launch the message listener.

## [1.0.0-alpha.26] - 2026-05-30
### Added
- **`hustle://` Protocol Handler (BETA)**: Introduced a URI-based routing system for dispatching hustle tasks.
- **Protocol-Driven CLI**: Added `-uri` flag to Orchestrator CLI for direct protocol execution.
- **Handler Registration**: Implemented a flexible registry for module handlers to prevent import cycles.
- **Integrated Routing**: Daemon and Interactive modes now utilize the protocol handler for consistent task execution.

## [1.0.0-alpha.25] - 2026-05-30
### Added
- **Curation Chain (BETA)**: Implemented automated pipeline connecting content curation to social media posting.
- **Functional Vector Search**: Integrated Go-level cosine similarity for semantic memory retrieval in `SQLiteStore`.
- **Chain Orchestration**: Added "chain" hustle type and interactive menu option for executing full multi-module workflows.
- **Scheduler Evolution**: Transitioned from individual hustle tasks to complex "CurationChain" scheduling.

## [1.0.0-alpha.24] - 2026-05-30
### Added
- **Social Media API Scaffolding (BETA)**: Implemented real-world API structures for Twitter and LinkedIn posting.
- **Executive Protocol: Forward Merge (BETA)**: Enhanced `sync.sh` with automated forward merging for `feat/ready-*` branches.
- **Interactive Module Launching (BETA)**: Wired Orchestrator CLI menu to launch specific hustle modules.
- **Autonomous Scheduling (BETA)**: Scheduler now automatically triggers Research and Curation tasks in daemon mode.
- **Hardened SQLite Persistence**: Added strict error handling for metadata and embedding serialization.

## [1.0.0-alpha.23] - 2026-05-30
### Added
- **Vector Memory (BETA)**: Implemented vector storage and retrieval scaffold in `SQLiteStore`.
- **Semantic Search (EXPERIMENTAL)**: Added `EmbeddingProvider` interface and updated `RankedSearch` to support future vector re-ranking.
- **Mock Embedder**: Added `MockEmbedder` for testing semantic search pipelines.

## [1.0.0-alpha.22] - 2026-05-30
### Added
- **Curation Module: Real Feeds (BETA)**: Integrated `gofeed` to fetch and process real RSS/Atom feeds for content curation.
- **Rate Limiting (STABLE)**: Implemented `RateLimiter` in `orchestrator` to manage API consumption, integrated into `WaterfallLLM`.
- **Executive Protocol (BETA)**: Enhanced `sync.sh` to reconcile all local feature branches with main automatically.
- **Interactive Menu (BETA)**: Added `-interactive` flag to Orchestrator for manual control and module selection.
- **UI Enhancements**: Added labels and descriptions to the terminal dashboard for improved observability.
- **Go Standardization**: Standardized all modules and CI to Go 1.24 for compatibility.

## [1.0.0-alpha.21] - 2026-05-30
### Added
- Task state persistence for Scheduler (Saves to `tasks.json`).
- Initial Content Curation hustle module.
- Auto-saving task history after execution.

## [1.0.0-alpha.20] - 2026-05-30
### Added
- Task Scheduler in `orchestrator/scheduler.go` for periodic execution.
- Daemon mode (`-daemon`) in Orchestrator CLI.
- Heartbeat task for automated status reporting.

## [1.0.0-alpha.19] - 2026-05-30
### Added
- Real PDF generation in Research module using `gofpdf`.
- Intelligent error classification and retry logic in `WaterfallLLM`.

## [1.0.0-alpha.18] - 2026-05-30
### Added
- Intelligent Merge Engine in `sync.sh` with multi-branch reconciliation.
- Hardened Git rollback implementation.

## [1.0.0-alpha.17] - 2026-05-30
### Added
- Real HTTP client for Anthropic messages API.
- Real HTTP client for Tavily search API.

## [1.0.0-alpha.16] - 2026-05-30
### Added
- LLM Waterfall implementation for provider failover.
- Enhanced AnthropicProvider scaffolding.

## [1.0.0-alpha.15] - 2026-05-30
### Added
- Weighted voting system for the Multi-Agent Council.
- Hybrid Memory Ranking sorting.

## [1.0.0-alpha.14] - 2026-05-30
### Added
- Terminal-optimized Live Dashboard.
- Council-integrated healing.

## [1.0.0-alpha.13] - 2026-05-30
### Added
- Multi-Agent Council for strategy verification.

## [1.0.0-alpha.12] - 2026-05-30
### Added
- SQLite persistence layer for tiered memory storage.
- Real file export logic for research reports.
- AnthropicProvider skeleton.

## [1.0.0-alpha.11] - 2026-05-30
### Added
- Integrated Revenue, Expenses, and Profit into `STATUS.json` reporting.
- Pluggable `LLMProvider` interface.
- Parameterized execution via `-params` flag.

## [1.0.0-alpha.10] - 2026-05-30
### Added
- Financial Ledger implementation.
- CLI flags (`-hustle`, `-sync`) for the Hustle Orchestrator.
- Social Media module provider interface.

## [1.0.0-alpha.9] - 2026-05-30
### Added
- Functional Healer loop with retry logic.
- Memory Search with keyword filtering.
- Initial Social Media hustle module scaffolding.

## [1.0.0-alpha.8] - 2026-05-30
### Added
- JSON persistence for tiered memory.
- Standalone entry points for Orchestrator and Research modules.
- Automatic integration between research results and orchestrator memory.

## [1.0.0-alpha.7] - 2026-05-30
### Added
- Utility-based heat scoring for memories with temporal decay.
- Enhanced Research Hustle pipeline with multi-provider support.

## [1.0.0-alpha.6] - 2026-05-30
### Added
- Real Git rollback implementation.
- Failure trap in `sync.sh`.

## [1.0.0-alpha.5] - 2026-05-30
### Added
- GitHub Actions CI pipeline.
- End-to-End (E2E) tests for the research hustle pipeline.

## [1.0.0-alpha.4] - 2026-05-29
### Added
- Enhanced `sync.sh` with branch reconciliation logic.
- Real git status and submodule monitoring.

## [1.0.0-alpha.3] - 2026-05-29
### Added
- Automated rollback handling skeleton.
- Real-time status reporting to `STATUS.json`.

## [1.0.0-alpha.2] - 2026-05-29
### Added
- Repository synchronization orchestration scripts.
- Sync monitoring and health check logic.

## [1.0.0-alpha.1] - 2026-05-29
### Added
- Repository initialization with knowledge base.
- Basic project structure scaffolding.
