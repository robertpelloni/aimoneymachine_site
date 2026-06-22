package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DiscoveredChain includes scheduling information
type DiscoveredChain struct {
	Chain
	IntervalMinutes int `json:"interval_minutes"`
}

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
func (d *ChainDiscoverer) Discover() (*DiscoveredChain, error) {
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
	prompt := fmt.Sprintf("As the Swarm Evolution Engine, analyze the following system state and propose a NEW sequential workflow (Chain) of protocol URIs. Focus on LUXURY HUSTLES: high-ROI, low-maintenance, and passive income generation.\nFinancial Context: %s\nRecent Successes:\n%s\n\nExisting Protocol Modules: research (query), curation (topic), social (platform, topic, content), trading (symbol), healer (issue), ledger\n\nOutput a JSON object for the NEW Chain including a suggested execution interval in minutes:\n{\n  \"name\": \"luxury_chain_name\",\n  \"description\": \"brief luxury-focused description\",\n  \"steps\": [\"hustle://step1\", \"hustle://step2\"],\n  \"interval_minutes\": 60\n}", profitAnalysis, successContext)

	response, err := d.Orchestrator.LLM.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate discovery: %v", err)
	}

	// 4. Parse JSON from response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("invalid LLM response format: %s", response)
	}
	jsonStr := response[start : end+1]

	var discovered DiscoveredChain
	if err := json.Unmarshal([]byte(jsonStr), &discovered); err != nil {
		return nil, fmt.Errorf("failed to unmarshal discovered chain: %v", err)
	}

	// 5. Register the discovery (skip if already exists)
	if _, exists := d.Manager.Chains[discovered.Name]; exists {
		fmt.Printf("[ChainDiscoverer] Chain '%s' already exists, skipping duplicate\n", discovered.Name)
		return nil, fmt.Errorf("chain '%s' already exists", discovered.Name)
	}
	d.Manager.Register(&discovered.Chain)

	// Log to L2 Vault
	d.Orchestrator.L2.Add(MemoryEntry{
		ID:        fmt.Sprintf("chain-disc-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Discovered and registered NEW LUXURY workflow: %s (%s) with interval %d min", discovered.Name, discovered.Description, discovered.IntervalMinutes),
		Timestamp: time.Now(),
		Tags:      []string{"discovery", "chain", "evolution", "luxury"},
	})

	return &discovered, nil
}
