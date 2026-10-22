# aimoneymachine_site — "AI Money Machine"

A Go system for running autonomous, LLM-driven online revenue experiments ("hustles") — self-orchestrating agents that plan, execute, and iterate on money-making workflows — together with the website and the large research corpus the project accumulated. Version **1.0.0-alpha.81**; Go 1.25.

Despite the repo name, this is not just a website: the site is one output of the machine. The tree contains:

- **The agent system** — Go code for the orchestration loop (agents that think, act, and record what worked)
- **The research corpus** — dozens of standing research documents at the root (`AGENT_AUTOMATABLE_SIDE_HUSTLES.md`, `AI_AGENTS_FRAMEWORKS.md`, `CONNECTIVITY_MCP_A2A.md`, `CODING_HARNESS_TOOLS.md`, `INFRASTRUCTURE_PROXY_LAYERS.md`, `MASTER_IDEAS.md`, and many more) covering agent frameworks, automatable business models, MCP/A2A connectivity, and tooling
- **Ops/process state** — `AUDIT.md`, `CHANGELOG.md`, `HANDOFF.md`, `MEMORY.md`, per-assistant instruction files, and session directories from the bot-assisted development process. The domain `aimoneymachine.site` is served from the same production server as Robert's other projects (its admin mail forwards to him).

## Status — the honest part

Alpha, experimental, single-developer. "Fully automated revenue" is the *research goal*, not a measured result: no income figures are published or claimed here, because there are none worth claiming yet. Read the corpus as a well-organized research notebook attached to a working agent harness — that is what this repo verifiably is.

## Related

- [`HyperNexusLLC/HyperNexus`](https://github.com/HyperNexusLLC/HyperNexus) — the control plane this work feeds into
- [`robertpelloni/marketing_agent`](https://github.com/robertpelloni/marketing_agent) — the Go outreach service that came out of the same line of work
