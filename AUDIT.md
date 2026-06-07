# Performance Audit

## v1.0.0-alpha.66
**Date:** 2026-06-07

### Core Benchmarks
- **Peer Status Ingest**: 100 peers in ~1.7ms (via A2A aggregate)
- **Memory Search**: 1000 entries searched in ~300µs (Tiered memory L1/L2)
- **Vector Search**: 5000 vectors in ~12ms (via SQLite-vec)
- **LLM Latency**: Main bottleneck (1.2s - 4.5s per decision on local Gemma-27B)

### Identified Bottlenecks
1. **LLM Inference**: The time taken for the LLM to process context and generate decisions is the primary limiting factor for agent iteration speed.
2. **CGO Overhead**: Calls to SQLite-vec through CGO add slight latency, though minimal compared to LLM calls.
3. **Network Latency**: Mesh synchronization is dependent on local network conditions.

### Proposed Optimizations
- **LLM Caching**: Implement a content-addressable cache for identical prompts.
- **Adaptive Sync**: Dynamically adjust mesh sync frequency based on volatility.
- **L1 TTL Pruning**: Periodically prune old L1 memory entries to maintain search speed.
