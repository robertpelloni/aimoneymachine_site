# Changelog

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
