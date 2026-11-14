package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
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
	// mu guards Chains: scheduler chain tasks, discovery registration
	// and API /chains reads now happen concurrently.
	mu sync.RWMutex
}

func NewChainManager(orch *Orchestrator, protocol *HustleProtocol) *ChainManager {
	return &ChainManager{
		Orchestrator: orch,
		Protocol:     protocol,
		Chains:       make(map[string]*Chain),
	}
}

func (m *ChainManager) Register(c *Chain) {
	m.mu.Lock()
	m.Chains[c.Name] = c
	m.mu.Unlock()
	fmt.Printf("[Chain] Registered workflow: %s (%d steps)\n", c.Name, len(c.Steps))
	m.SaveState("chains.json")
}

func (m *ChainManager) SaveState(filepath string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, err := json.MarshalIndent(m.Chains, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(filepath, data, 0644)
}

func (m *ChainManager) LoadState(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, &m.Chains)
}

// Snapshot returns a copy of the registry for safe reading (e.g. the
// API /chains handler marshaling while discovery registers).
func (m *ChainManager) Snapshot() map[string]*Chain {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]*Chain, len(m.Chains))
	for k, v := range m.Chains {
		out[k] = v
	}
	return out
}

// Has reports whether a chain is registered (used by scheduler orphan
// cleanup and discovery duplicate checks).
func (m *ChainManager) Has(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.Chains[name]
	return ok
}

func (m *ChainManager) Execute(name string) error {
	m.mu.RLock()
	chain, ok := m.Chains[name]
	m.mu.RUnlock()
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
