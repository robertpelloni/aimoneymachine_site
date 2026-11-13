# aimoneymachine_site — "AI Money Machine"

A Go system for running autonomous, LLM-driven online revenue experiments ("hustles") — self-orchestrating agents that plan, execute, and iterate on money-making workflows — together with the website and the large research corpus the project accumulated.

Despite the repo name, this is not just a website: the site is one output of the machine. The tree contains:

- **The agent system** — Go code for the orchestration loop (agents that think, act, and record what worked), plus `orchestrator/`, `hustle/`, and a 19 MB compiled `orchestrator_linux_new` binary
- **The research corpus** — dozens of standing research documents at the root (`AGENT_AUTOMATABLE_SIDE_HUSTLES.md`, `AI_AGENTS_FRAMEWORKS.md`, `CONNECTIVITY_MCP_A2A.md`, `CODING_HARNESS_TOOLS.md`, `INFRASTRUCTURE_PROXY_LAYERS.md`, `MASTER_IDEAS.md`, and many more) covering agent frameworks, automatable business models, MCP/A2A connectivity, and tooling
- **Ops/process state** — `AUDIT.md`, `CHANGELOG.md`, `HANDOFF.md`, `MEMORY.md`, per-assistant instruction files, and session directories from the bot-assisted development process. The domain `aimoneymachine.site` is served from the same production server as Robert's other projects.

## Status — the honest part

**Alpha, experimental, single-developer, parked.** Last substantive work: 2026-08-03 (markdown-to-HTML conversion across 382 blog posts); last commit: 2026-10-22 (README rewrite).

"Fully automated revenue" is the *research goal*, not a measured result. The only income figures recorded in the tree are in `STATUS.json` (2026-07-03): revenue $171.60, expenses $1,005, profit **−$833.40**. Read the corpus as a well-organized research notebook attached to a working agent harness — that is what this repo verifiably is.

**Note:** `VERSION.md` and `STATUS.json` both still contain unresolved merge-conflict markers (two competing versions, 1.0.0-alpha.83 vs v1.1.0-alpha.92). The version is unsettled until someone resolves them.

## Related

- [`HyperNexusLLC/HyperNexus`](https://github.com/HyperNexusLLC/HyperNexus) — the control plane this work feeds into
- [`robertpelloni/marketing_agent`](https://github.com/robertpelloni/marketing_agent) — the Go outreach service that came out of the same line of work
