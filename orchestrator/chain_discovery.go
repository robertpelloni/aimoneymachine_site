package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ChainDiscoverer utilizes the LLM to evolve system workflows based on data
type ChainDiscoverer struct {
	Orchestrator *Orchestrator
	Manager      *ChainManager
}

func NewChainDiscoverer(orch *Orchestrator, mgr *ChainManager) *ChainDiscoverer {
	return &ChainDiscoverer{
		Orchestrator: orch,
		Manager:      mgr,
	}
}

// Discover analyzes the ledger and memory to propose and register a new workflow
func (d *ChainDiscoverer) Discover() (*Chain, error) {
	fmt.Println("[ChainDiscoverer] Analyzing system state for workflow evolution...")

	// 1. Get Financial Context
	profitAnalysis := d.Orchestrator.Ledger.AnalyzeProfitability()

	// 2. Get Recent High-Value Memories
	recentSuccesses := d.Orchestrator.L1.Search("success")
	successContext := ""
	for _, e := range recentSuccesses {
		successContext += fmt.Sprintf("- %s (Tags: %v)\n", e.Content, e.Tags)
	}

	// 3. Construct Prompt for LLM
	prompt := fmt.Sprintf(`As the Swarm Evolution Engine, analyze the following system state and propose a NEW sequential workflow (Chain) of protocol URIs.
Financial Context: %s
Recent Successes:
%s

Existing Protocol Modules: research (query, action=analyze), curation (topic, feed), social (platform, topic, action=post), trading (symbol, action=all), healer (issue), ledger

Output a JSON object for the NEW Chain:
{
  "name": "unique_chain_name",
  "description": "brief description",
  "steps": ["hustle://step1", "hustle://step2"]
}`, profitAnalysis, successContext)

	response, err := d.Orchestrator.LLM.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate discovery: %v", err)
	}

	// 4. Parse JSON from response
	// Find the first { and last }
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("invalid LLM response format: %s", response)
	}
	jsonStr := response[start : end+1]

	var newChain Chain
	if err := json.Unmarshal([]byte(jsonStr), &newChain); err != nil {
		return nil, fmt.Errorf("failed to unmarshal discovered chain: %v", err)
	}

	// 5. Register the discovery
	d.Manager.Register(&newChain)

	// Log to L2 Vault
	d.Orchestrator.L2.Add(MemoryEntry{
		ID:        fmt.Sprintf("chain-disc-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Discovered and registered new workflow: %s (%s)", newChain.Name, newChain.Description),
		Timestamp: time.Now(),
		Tags:      []string{"discovery", "chain", "evolution"},
	})

	return &newChain, nil
}
