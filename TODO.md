# Todo List

## 🔴 Blockers
- [ ] **Fix Windows CGO build** — Migrate `go-sqlite3` to `modernc.org/sqlite` (pure Go).
- [ ] **Real Social Posting** — Replace stubs in `hustle/social/post.go` with real API calls.
- [ ] **Real Web Search** — Replace mock in `hustle/research/search.go` with Tavily/Brave API.

## 🟠 High Priority
- [ ] Implement Git-based rollback in `orchestrator/rollback.go`.
- [x] Add graceful shutdown (SIGINT handler) to persist state on exit. (v1.0.0-alpha.67)
- [ ] Enhance dashboard to show multi-agent status and iterations.
- [ ] Add `--dry-run` flag to social module.

## 🟡 Medium Priority
- [ ] Make RSS feed list configurable via `.env`.
- [ ] Implement content topic queue for sequential generation.
- [ ] Fix `hustle/research/report.go` compilation issues on Windows.

## 🟢 Lower Priority
- [ ] Multi-exchange support (Binance, Kraken).
- [ ] LLM response cache.
- [ ] Browser extension for remote monitoring.

## ✅ Completed
- [x] Merge Phase 3 and v1.0.0-alpha.65 stable baseline.
- [x] OpenAI-compatible LLM provider.
- [x] Agent Loop (Observe → Think → Act → Learn → Evaluate).
- [x] Content Hustle module.
- [x] Refactored Healer with LLM verification.
- [x] Content module test coverage.
