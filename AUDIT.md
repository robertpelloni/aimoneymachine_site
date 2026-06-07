# Performance Audit & Optimization Plan

## Audit Findings (v1.0.0-alpha.63)

### What Works Well
- **Swarm Ingestion**: High performance (100 peers in ~1.7ms).
- **Memory Search**: High performance (1000 entries in ~300µs).
- **LLM Connection**: Auto-detects LM Studio models in <1s, graceful fallback to MockLLM.
- **Agent Loop**: 2-second pause between iterations prevents LLM hammering; wealth preservation auto-stop works.
- **CoinGecko Price Fetch**: Typically completes under 2s.

### Current Bottleneck
- **LLM Inference**: Local model inference is the primary latency factor. Gemma 27B takes 3-15s per generation depending on prompt length and GPU. This is acceptable for strategic orchestration but limits high-frequency actions.

### Build Blocker
- **CGO/sqlite3 on Windows**: `go-sqlite3` + gcc 15.2.0 fails with `cgo: cannot parse gcc output`. This blocks all compilation on Windows. Migration to `modernc.org/sqlite` (pure Go, zero CGO) is the highest-priority fix.

## Optimization Opportunities

### 1. Build System
- **Migrate to pure-Go SQLite**: Replace `mattn/go-sqlite3` with `modernc.org/sqlite`. Eliminates CGO dependency, enables cross-platform builds without gcc, fixes Windows compilation. This is the #1 blocker.
- **Remove gofpdf dependency**: The `hustle/research` module depends on `jung-kurt/gofpdf` which may have similar CGO/build issues. Consider alternatives or make it optional.

### 2. LLM Performance
- **LLM Cache**: Implement content-addressable cache for LLM requests. If the same prompt is sent twice (e.g., periodic research on the same query), return cached response. Saves GPU time and reduces latency from 10s to <1ms on cache hits.
- **Prompt Compression**: Agent loop observe() builds context from memory. As L1 grows, the prompt gets larger. Implement a summarization step that compresses recent history before sending to LLM.
- **Streaming Responses**: Use SSE streaming for long content generation (blog posts, newsletters). Display partial output in dashboard/interactive mode.

### 3. Memory & Database
- **Sqlite-vec Indexing**: As L3 archive grows, explicit indexing on the vector table required to maintain <10ms search latency.
- **L1 Pruning TTL**: Implement Time-To-Live for L1 memory entries to prevent unbounded growth in long-running nodes. Auto-promote high-score entries to L2, discard low-score entries after 24h.
- **Batch Memory Operations**: When agent loop learns, batch multiple memory adds into a single write transaction for SQLite performance.

### 4. Agent Loop
- **Vary Actions**: Agent sometimes repeats the same action. Add deduplication: track last 3 actions and explicitly tell the LLM "do NOT repeat these."
- **Rate-Limit Per Module**: Trading module hits CoinGecko every 30s in daemon mode — respect their free tier rate limits (10-30 calls/min).
- **Concurrent Agent Loops**: MultiAgentOrchestrator runs agents concurrently but shares a single Orchestrator (and its mutex-protected state). Consider per-agent L1 scratchpads for isolation.

### 5. High-Frequency Mesh Tuning
- **Batch Status Aggregation**: Instead of individual `hustle://swarm?action=provide_status` responses, batch multiple peer statuses into a single mesh broadcast.
- **Delta-Sync Frequency**: Implement adaptive sync intervals — sync more frequently during high volatility (identified via Trading) and less frequently during stable periods.

### Priority Order

1. 🔴 Fix Windows build (migrate to pure-Go SQLite)
2. 🔴 Write tests for new modules
3. 🟠 LLM cache (biggest performance win)
4. 🟠 L1 pruning TTL
5. 🟡 Agent action deduplication
6. 🟡 Streaming responses
7. 🟢 Adaptive sync intervals
