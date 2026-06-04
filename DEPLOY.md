# Deployment Protocol

## 1. Prerequisites
- Go 1.24.x
- Git 2.x
- Access to LLM APIs (Anthropic, Google, etc.)

## 2. Local Setup
1. Clone the repository.
2. Run `./sync.sh` to ensure all submodules and branches are aligned.
3. Run `./build.sh` to compile the orchestrator and hustle modules.

### External API Keys
- `TAVILY_API_KEY`: Required for Research module.
- `ANTHROPIC_API_KEY`: Required for LLM-driven orchestration.
- `COINGECKO_API_KEY`: (Optional) For Pro-tier CoinGecko access.

## 3. Automated Execution
The system is designed for continuous autonomous execution. Ensure environment variables for API keys are set before launching the orchestrator.
