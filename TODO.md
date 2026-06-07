# Todo List

## 🔴 Blockers (Must Fix Before Anything Else)
- [ ] **Fix Windows CGO build** — `go-sqlite3` fails with gcc 15.2.0 on Windows (`cgo: cannot parse gcc output as ELF`). Investigate: try `go-sqlite3` pure-Go alternative (`modernc.org/sqlite`) or fix gcc toolchain.
- [x] **Add `content` module to `go.work` uses** — verify `go work sync` resolves all module dependencies
- [ ] **Verify `orchestrator/go.mod`** has no stale/missing requires after new files added

## 🟠 High Priority (Phase 3 Completion)
- [ ] Write `orchestrator/agent_loop_test.go` — test the observe/think/act/learn/evaluate cycle
- [ ] Write `orchestrator/openai_compat_test.go` — test real LLM connection, model detection, embedding
- [ ] Write `hustle/content/content_test.go` — test all 4 content types with mock LLM
- [x] Create `.env.example` — document all environment variables
- [x] Add `hustle://content` to agent loop's available actions list in `agent_loop.go` think prompt
- [x] Add `hustle://swarm?action=aggregate` to agent loop's available actions list
- [ ] Test end-to-end: `./bin/orchestrator -agent -agent-iterations 3` with LM Studio running
- [ ] Test: `./bin/orchestrator -autoplan` with LM Studio running
- [ ] Test: `./bin/orchestrator -daemon -real-prices` runs for 10 minutes without crash

## 🟡 Medium Priority (Phase 4)
- [ ] Wire Tavily API into research module — real web search, not mock results
- [ ] Make RSS feed list configurable (env var or config file, not hardcoded HN)
- [ ] Add `--dry-run` flag to social module — show what would be posted without posting
- [ ] Add human approval gate before social posts in agent mode
- [ ] Implement content topic queue — LLM generates 10 topics, agent works through them sequentially
- [ ] Fix `hustle/research/report.go` — gofpdf dependency may not compile on Windows
- [ ] Add graceful shutdown (SIGINT handler) to daemon/agent modes — persist state on exit
- [ ] Enhance dashboard to show agent loop status (iterations, successes, errors per agent)
- [x] Add `output/` directory tracking to `.gitignore`

## 🟢 Lower Priority (Phase 5+)
- [ ] Implement closed-loop healer: `heal_and_verify(error, file)` with execute-fix-verify-retry
- [ ] Implement stop hooks — intercept before session ends, verify promises fulfilled
- [ ] Stress test multi-node clustering with 10+ peers
- [ ] Implement `hustle://chain?action=optimize` to refactor existing chains via LLM
- [ ] Multi-exchange support (Binance, Kraken) for the Trading module
- [ ] Browser extension for remote "Hustle Monitoring"
- [ ] Integration with decentralized identity (DID) for peer verification
- [ ] LLM response cache — content-addressable, avoid re-generating identical prompts

## ✅ Completed
- [x] Integrate `sqlite-vec` for hyper-fast, local-first context matching. (v1.0.0-alpha.35)
- [x] Implement A2A Mesh for cross-host collaboration. (v1.0.0-alpha.34)
- [x] Implement autonomous Luxury discovery. (v1.0.0-alpha.42)
- [x] Wire WealthPreservation to Scheduler — auto-unregisters failing tasks. (v1.0.0-alpha.44)
- [x] OpenAI-compatible LLM provider (LM Studio / Ollama). (v1.0.0-alpha.63)
- [x] Agent Loop (Observe → Think → Act → Learn → Evaluate). (v1.0.0-alpha.63)
- [x] Content Hustle module (blog, newsletter, SEO, social thread). (v1.0.0-alpha.63)
- [x] Agent loop strategy guidelines updated — content generation prioritized as highest ROI
