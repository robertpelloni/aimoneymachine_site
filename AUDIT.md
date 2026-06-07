# Performance Audit & Optimization Plan

## Audit Findings (v1.0.0-alpha.49)
- **Swarm Ingestion**: High performance (100 peers in ~1.7ms).
- **Memory Search**: High performance (1000 entries in ~300µs).
- **Bottleneck**: LLM inference remains the primary latency factor, which is acceptable for strategic orchestration but may impact high-frequency trading.

## Optimization Opportunities

### 1. Strategic Caching
- **LLM Cache**: Implement a simple content-addressable cache for LLM requests (e.g., sentiment extraction from identical search snippets).
- **Mesh Index Caching**: Cache the local memory index (checksums) to avoid full recalculation during every `Sync()` request.

### 2. High-Frequency Mesh Tuning
- **Batch Status Aggregation**: Instead of individual `hustle://swarm?action=provide_status` responses, nodes could batch multiple peer statuses into a single mesh broadcast.
- **Delta-Sync Frequency**: Implement adaptive sync intervals—sync more frequently during high volatility (identified via Trading/Research) and less frequently during stable periods.

### 3. Database & Memory
- **Sqlite-vec Indexing**: As the L3 archive grows, explicit indexing on the vector table will be required to maintain <10ms search latency.
- **L1 Pruning**: Implement a TTL (Time-To-Live) for L1 memory entries to prevent unbounded memory growth in long-running nodes.

### 4. Code Path Optimization
- **Concurrent Search**: Parallelize keyword search across L1, L2, and L3 tiers for massive memory stores.
- **Zero-Copy Serialization**: Use more efficient JSON parsing or binary formats for high-volume mesh messages.
