package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Chain defines a sequential workflow of protocol-driven tasks
type Chain struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"` // Sequence of hustle:// URIs
}

// ChainManager handles registration and execution of complex workflows
type ChainManager struct {
	Orchestrator *Orchestrator
	Protocol     *HustleProtocol
	Chains       map[string]*Chain
}

func NewChainManager(orch *Orchestrator, protocol *HustleProtocol) *ChainManager {
	return &ChainManager{
		Orchestrator: orch,
		Protocol:     protocol,
		Chains:       make(map[string]*Chain),
	}
}

func (m *ChainManager) Register(c *Chain) {
	m.Chains[c.Name] = c
	fmt.Printf("[Chain] Registered workflow: %s (%d steps)\n", c.Name, len(c.Steps))
	m.SaveState("chains.json")
}

func (m *ChainManager) SaveState(filepath string) error {
	data, err := json.MarshalIndent(m.Chains, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

func (m *ChainManager) LoadState(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.Chains)
}

func (m *ChainManager) Execute(name string) error {
	chain, ok := m.Chains[name]
	if !ok {
		return fmt.Errorf("chain %s not found", name)
	}

	fmt.Printf("[Chain] Executing workflow: %s - %s\n", chain.Name, chain.Description)

	for i, stepURI := range chain.Steps {
		fmt.Printf("[Chain] Step %d/%d: %s\n", i+1, len(chain.Steps), stepURI)

		// Record start in L1
		m.Orchestrator.L1.Add(MemoryEntry{
			ID:        fmt.Sprintf("chain-%s-step-%d-%d", name, i, time.Now().Unix()),
			Content:   fmt.Sprintf("Executing chain %s step: %s", name, stepURI),
			Timestamp: time.Now(),
			Tags:      []string{"chain", name, "step"},
		})

		if err := m.Protocol.HandleURI(stepURI); err != nil {
			return fmt.Errorf("chain %s failed at step %d (%s): %v", name, i+1, stepURI, err)
		}
	}

	fmt.Printf("[Chain] Workflow %s completed successfully.\n", chain.Name)
	return nil
}
