package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryEntry represents a single unit of information
type MemoryEntry struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	BaseScore float64   `json:"base_score"`
	Timestamp time.Time `json:"timestamp"`
	Tags      []string  `json:"tags"`
	Vector    []float32 `json:"vector,omitempty"`
}

// Score calculates the current "heat" of the memory with temporal decay
func (e MemoryEntry) Score() float64 {
	elapsed := time.Since(e.Timestamp).Hours()
	return e.BaseScore * math.Exp(-0.1*elapsed)
}

// Checksum returns a unique hash of the entry's state
func (e MemoryEntry) Checksum() string {
	h := sha256.New()
	h.Write([]byte(e.ID + e.Content))
	return hex.EncodeToString(h.Sum(nil))
}

// L1Memory (Scratchpad)
type L1Memory struct {
	Entries []MemoryEntry `json:"entries"`
	// mu guards Entries for concurrent scheduler/API/agent access.
	mu sync.RWMutex
}

func (m *L1Memory) Add(entry MemoryEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, entry)
}

// EntriesSnapshot returns a copy of the entries under lock.
func (m *L1Memory) EntriesSnapshot() []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]MemoryEntry, len(m.Entries))
	copy(out, m.Entries)
	return out
}

// Len returns the entry count under lock.
func (m *L1Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Entries)
}

func (m *L1Memory) Get(id string) (MemoryEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return MemoryEntry{}, false
}

func (m *L1Memory) Search(query string) []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.searchLocked(query)
}

func (m *L1Memory) searchLocked(query string) []MemoryEntry {
	var results []MemoryEntry
	query = strings.ToLower(query)
	for _, entry := range m.Entries {
		match := strings.Contains(strings.ToLower(entry.Content), query)
		if !match {
			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					match = true
					break
				}
			}
		}
		if match {
			results = append(results, entry)
		}
	}
	return results
}

func (m *L1Memory) RankedSearch(query string, embedder EmbeddingProvider) []MemoryEntry {
	m.mu.RLock()
	results := m.searchLocked(query)
	m.mu.RUnlock()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})
	return results
}

func (m *L1Memory) Checksum() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var combined string
	for _, e := range m.Entries {
		combined += e.Checksum()
	}
	h := sha256.New()
	h.Write([]byte(combined))
	return hex.EncodeToString(h.Sum(nil))
}

// L2Memory (Vault)
type L2Memory struct {
	Entries []MemoryEntry `json:"entries"`
	// mu guards Entries for concurrent scheduler/API/agent access.
	mu sync.RWMutex
}

func (m *L2Memory) Add(entry MemoryEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, entry)
}

// EntriesSnapshot returns a copy of the entries under lock.
func (m *L2Memory) EntriesSnapshot() []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]MemoryEntry, len(m.Entries))
	copy(out, m.Entries)
	return out
}

// Len returns the entry count under lock.
func (m *L2Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Entries)
}

func (m *L2Memory) Get(id string) (MemoryEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return MemoryEntry{}, false
}

func (m *L2Memory) Search(query string) []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.searchLocked(query)
}

func (m *L2Memory) searchLocked(query string) []MemoryEntry {
	var results []MemoryEntry
	query = strings.ToLower(query)
	for _, entry := range m.Entries {
		match := strings.Contains(strings.ToLower(entry.Content), query)
		if !match {
			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					match = true
					break
				}
			}
		}
		if match {
			results = append(results, entry)
		}
	}
	return results
}

func (m *L2Memory) RankedSearch(query string, embedder EmbeddingProvider) []MemoryEntry {
	m.mu.RLock()
	results := m.searchLocked(query)
	m.mu.RUnlock()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})
	return results
}

func (m *L2Memory) Checksum() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var combined string
	for _, e := range m.Entries {
		combined += e.Checksum()
	}
	h := sha256.New()
	h.Write([]byte(combined))
	return hex.EncodeToString(h.Sum(nil))
}

// L3Memory (Archive)
type L3Memory struct {
	Entries []MemoryEntry `json:"entries"`
	// mu guards Entries for concurrent scheduler/API/agent access.
	mu sync.RWMutex
}

func (m *L3Memory) Add(entry MemoryEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Entries = append(m.Entries, entry)
}

// EntriesSnapshot returns a copy of the entries under lock.
func (m *L3Memory) EntriesSnapshot() []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]MemoryEntry, len(m.Entries))
	copy(out, m.Entries)
	return out
}

// Len returns the entry count under lock.
func (m *L3Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Entries)
}

func (m *L3Memory) Get(id string) (MemoryEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return MemoryEntry{}, false
}

func (m *L3Memory) Search(query string) []MemoryEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.searchLocked(query)
}

func (m *L3Memory) searchLocked(query string) []MemoryEntry {
	var results []MemoryEntry
	query = strings.ToLower(query)
	for _, entry := range m.Entries {
		match := strings.Contains(strings.ToLower(entry.Content), query)
		if !match {
			for _, tag := range entry.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					match = true
					break
				}
			}
		}
		if match {
			results = append(results, entry)
		}
	}
	return results
}

func (m *L3Memory) RankedSearch(query string, embedder EmbeddingProvider) []MemoryEntry {
	m.mu.RLock()
	results := m.searchLocked(query)
	m.mu.RUnlock()
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})
	return results
}

func (m *L3Memory) Checksum() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var combined string
	for _, e := range m.Entries {
		combined += e.Checksum()
	}
	h := sha256.New()
	h.Write([]byte(combined))
	return hex.EncodeToString(h.Sum(nil))
}

// Orchestrator handles tiered memory orchestration, financial tracking, and LLM access
type Orchestrator struct {
	Version    string            `json:"version"`
	DryRun     bool              `json:"dry_run"`
	RSSFeeds   []string          `json:"rss_feeds"`
	WealthGoal float64           `json:"wealth_goal"`
	TaskQueue  []string          `json:"task_queue"`
	L1         L1Memory          `json:"l1"`
	L2         L2Memory          `json:"l2"`
	L3         L3Memory          `json:"l3"`
	Ledger     Ledger            `json:"ledger"`
	LLM        LLMProvider       `json:"-"`
	Embedder   EmbeddingProvider `json:"-"`
	DB         *SQLiteStore      `json:"-"`
	Identity   *Identity         `json:"-"`
	Cache      *CachingLLM       `json:"-"`
	Optimizer  *PromptOptimizer  `json:"-"`
}

func NewOrchestrator() *Orchestrator {
	o := &Orchestrator{
		WealthGoal: 10000.0,
		L1:         L1Memory{Entries: make([]MemoryEntry, 0)},
		L2:         L2Memory{Entries: make([]MemoryEntry, 0)},
		L3:         L3Memory{Entries: make([]MemoryEntry, 0)},
		Ledger:     Ledger{Transactions: make([]Transaction, 0)},
		LLM:        &MockLLM{},
		Embedder:   &MockEmbedder{},
		Optimizer:  NewPromptOptimizer(),
	}

	// Attempt to load existing ledger
	o.Ledger.Load("ledger.json")

	return o
}

func (o *Orchestrator) Save(filepath string) error {
	if o.DB != nil {
		for _, e := range o.L1.EntriesSnapshot() {
			o.DB.SaveMemory("L1", e)
		}
		for _, e := range o.L2.EntriesSnapshot() {
			o.DB.SaveMemory("L2", e)
		}
		for _, e := range o.L3.EntriesSnapshot() {
			o.DB.SaveMemory("L3", e)
		}
	}

	// Also persist ledger separately for modularity
	o.Ledger.Save("ledger.json")

	// Hold the tier + ledger read locks for the marshal: concurrent
	// scheduler tasks may be Adding while shutdown saves.
	o.L1.mu.RLock()
	defer o.L1.mu.RUnlock()
	o.L2.mu.RLock()
	defer o.L2.mu.RUnlock()
	o.L3.mu.RLock()
	defer o.L3.mu.RUnlock()
	o.Ledger.mu.RLock()
	defer o.Ledger.mu.RUnlock()
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(filepath, data, 0644)
}

func (o *Orchestrator) Load(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, o)
}

func (o *Orchestrator) Shutdown() {
	fmt.Println("[Orchestrator] Initiating graceful shutdown...")

	// Persist final state
	if err := o.Save("memory.json"); err != nil {
		fmt.Printf("[Orchestrator] Error during state persistence: %v\n", err)
	} else {
		fmt.Println("[Orchestrator] System state persisted to memory.json")
	}

	if o.DB != nil {
		o.DB.Close()
	}
	fmt.Println("[Orchestrator] Shutdown complete.")
}
