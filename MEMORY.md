[PROJECT_MEMORY]

# AI Hustle Machine — Project Memory & Architecture Summary

## 1. Core Vision & Philosophy
The **AI Hustle Machine** (also referred to as the "Fully Automated Luxury Protocol") is a self-orchestrating, LLM-driven autonomous agent system designed to run revenue-generating "hustles" using local or free LLM models. The system operates on continuous autonomous execution: Observe → Think → Act → Learn → Evaluate. It is designed to require zero human intervention once deployed, aggressively targeting high-ROI, low-maintenance workflows.

## 2. Monorepo Architecture
The project is built in **Go 1.25.0** using a `go.work` monorepo structure. It moved away from submodules to a unified repository to prevent sync issues.
*   **Orchestrator (`/orchestrator`)**: The core engine. It manages the Agent Loop, multi-agent Council, Memory tiers, Ledger, Scheduler, and Healer.
*   **Hustle Modules (`/hustle/*`)**: Specialized domain execution environments:
    *   `content`: Markdown CMS, static site generation, blogs, newsletters.
    *   `curation`: RSS aggregation and summarizing.
    *   `research`: Web search (Tavily/Brave) and alpha discovery.
    *   `social`: Twitter/LinkedIn automated posting and engagement.
    *   `trading`: Crypto trading and TA using CoinGecko.

## Discovered Optimizations
- **LLM Response Caching**: Content-addressable SQLite cache significantly reduces redundant API calls and costs.
- **LLM Sentiment Confluence**: Filtering signals through LLM-extracted sentiment reduces noise.
- **Mesh Aggregation**: Centralizing status in L1 memory allows the Dashboard to remain stateless.
- **Content Generation is Highest-ROI**: Zero marginal compute cost for directly monetizable output.
- **Automated Affiliate Inflow**: Every generated asset is a monetization vector via integrated affiliate insertion.
- **Lead Discovery Loop**: Research now feeds directly into lead generation, creating high-value business intelligence.
- **Engagement-Driven Growth**: Social module now includes a Search & Reply loop to autonomously grow audience reach.
- **Confluence 2.0**: Trading decisions are now gated by both technical indicators (Bollinger/MACD) and real-time market sentiment analysis.
- **Multi-Exchange Execution**: Abstracted trade execution allows for real-world capital growth on Binance and Kraken.

## Known Technical Debt
- Social posting providers (Twitter, LinkedIn) have retry logic but still require full OAuth2 flow verification in production environments.
