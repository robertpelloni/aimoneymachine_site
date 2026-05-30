# Changelog

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
- **Curation Module (EXPERIMENTAL)**: Scaffolded a new module for automated content curation.
- **Task Persistence (BETA)**: Implemented `tasks.json` state persistence for the curation module.
- **Validation**: Added E2E validation for curation and task history.

## [1.0.0-alpha.20] - 2026-05-30
### Added
- **Autonomous Scheduling (BETA)**: Implemented `orchestrator/scheduler.go` for periodic hustle execution.
- **Daemon Mode (BETA)**: Added `-daemon` flag to orchestrator CLI for background operation.

## [1.0.0-alpha.19] - 2026-05-30
### Added
- **Professional Output (BETA)**: Integrated `gofpdf` for high-quality intelligence report generation.

## [1.0.0-alpha.18] - 2026-05-30
### Added
- **Intelligent Merge Engine (BETA)**: `sync.sh` now automates branch reconciliation across all local feature branches.
- **Hardened Rollback (BETA)**: Implementation of full Git recovery logic (abort, reset, clean).

## [1.0.0-alpha.17] - 2026-05-30
### Added
- **Anthropic Integration (BETA)**: Functional HTTP client for Claude 3.5 Sonnet.
- **Tavily Integration (BETA)**: Functional HTTP client for real-time web research.

## [1.0.0-alpha.16] - 2026-05-30
### Added
- **LLM Failover (BETA)**: Implemented `WaterfallLLM` for automatic sequential failover across multiple providers.

## [1.0.0-alpha.15] - 2026-05-30
### Added
- **Multi-Agent Governance (BETA)**: Implemented role-based weights and automated consensus scoring for the council.
- **Smart Memory (BETA)**: Ranked search combining relevance and temporal heat.

## [1.0.0-alpha.14] - 2026-05-30
### Added
- **Terminal Dashboard (BETA)**: Live visualization of machine health, finances, and events.
- **Council-Led Healing (BETA)**: Integrated agent debates into the self-healing loop.

## [1.0.0-alpha.13] - 2026-05-30
### Added
- **Multi-Agent Council (BETA)**: Consensus engine for strategic validation (Bull, Bear, Critic).

## [1.0.0-alpha.12] - 2026-05-30
### Added
- **SQLite Persistence (BETA)**: Robust memory storage backend.
- **Dual-Storage (BETA)**: JSON + SQLite synchronization.

## [1.0.0-alpha.11] - 2026-05-30
### Added
- **Financial Status (BETA)**: Integrated Revenue/Expenses/Profit into `STATUS.json`.
- **LLM Orchestration (BETA)**: Pluggable provider interface.

## [1.0.0-alpha.10] - 2026-05-30
### Added
- **Ledger System (BETA)**: Transactional tracking for all hustles.
- **Social Scaffolding (EXPERIMENTAL)**: LinkedIn and Twitter provider interfaces.

## [1.0.0-alpha.9] - 2026-05-30
### Added
- **State-Aware Healer (BETA)**: Enhanced loop with memory logging and retry limits.

## [1.0.0-alpha.8] - 2026-05-30
### Added
- **Memory Persistence (BETA)**: L1/L2/L3 tiers now save to `memory.json`.
- **Runnable CLI**: Dedicated entry points in `bin/`.

## [1.0.0-alpha.7] - 2026-05-30
### Added
- **Biological Memory (BETA)**: Heat scoring and exponential temporal decay.
- **Intelligence Synthesis (BETA)**: Research engine supporting multiple providers.

## [1.0.0-alpha.6] - 2026-05-30
### Added
- **Sync Failure Trap (BETA)**: Automated rollback on git merge failures.

## [1.0.0-alpha.5] - 2026-05-30
### Added
- **CI Pipeline (STABLE)**: GitHub Actions automation.
- **E2E Tests (STABLE)**: Pipeline validation for research hustles.

## [1.0.0-alpha.4] - 2026-05-30
### Added
- **Executive Sync (BETA)**: Branch reconciliation protocol in `sync.sh`.

## [1.0.0-alpha.3] - 2026-05-30
### Added
- **Health Reporting (BETA)**: `STATUS.json` real-time updates.

## [1.0.0-alpha.2] - 2026-05-30
### Added
- **Autonomous Scripts (STABLE)**: `sync.sh` and `build.sh`.

## [1.0.0-alpha.1] - 2026-05-30
### Added
- **Core Architecture (EXPERIMENTAL)**: Go-based tiered memory and self-healing loop.
- **Rebranding**: Complete pivot from Hypercode to AI Hustle Machine.
