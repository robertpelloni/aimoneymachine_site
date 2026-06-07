# Deployment & Environment Setup

## Prerequisites
- **Go 1.24.3** (Strict toolchain requirement)
- **SQLite 3.45+** (With extension loading support)
- **External APIs**:
  - `ANTHROPIC_API_KEY` (Primary LLM)
  - `OPENAI_API_KEY` (Secondary LLM / Waterfall fallback)
  - `TAVILY_API_KEY` (Mandatory for Research module)

## Standard Installation
1.  **Repository Setup**:
    ```bash
    git clone https://github.com/robertpelloni/fully_automated_gay_luxury_space_communism
    cd fully_automated_gay_luxury_space_communism
    ```
2.  **Binary Compilation**:
    ```bash
    ./build.sh
    ```

## High-ROI Mesh Configuration
For optimal autonomous performance, deploy as a federated mesh:
1.  **Primary Ledger Node**:
    ```bash
    ./bin/orchestrator -daemon -api 8080 -real-prices
    ```
2.  **Worker/Peer Nodes**:
    ```bash
    ./bin/orchestrator -daemon -api 8081 -seed http://primary-ip:8080 -real-prices
    ```

## Operational Monitoring
- **Real-time Visualization**: `./bin/orchestrator -dashboard` provides a live view of mesh-wide status and profits.
- **Financial Audit**: All nodes maintain a local `ledger.json` which is automatically synchronized during mesh updates.
- **Capital Protection**: The `Healer` module executes a "Wealth Preservation" audit every 12 hours, automatically unregistering hustles that generate negative ROI.

## Troubleshooting
If the system becomes unstable or branches drift:
```bash
./bin/orchestrator -uri "hustle://sync"
```
