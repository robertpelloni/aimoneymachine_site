package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"sort"
	"strings"
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
}

func (m *L1Memory) Add(entry MemoryEntry) {
	m.Entries = append(m.Entries, entry)
}

func (m *L1Memory) Get(id string) (MemoryEntry, bool) {
	for _, e := range m.Entries {
		if e.ID == id { return e, true }
	}
	return MemoryEntry{}, false
}

func (m *L1Memory) Search(query string) []MemoryEntry {
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
	results := m.Search(query)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})
	return results
}

func (m *L1Memory) Checksum() string {
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
}

func (m *L2Memory) Add(entry MemoryEntry) {
	m.Entries = append(m.Entries, entry)
}

func (m *L2Memory) Get(id string) (MemoryEntry, bool) {
	for _, e := range m.Entries {
		if e.ID == id { return e, true }
	}
	return MemoryEntry{}, false
}

func (m *L2Memory) Search(query string) []MemoryEntry {
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
	results := m.Search(query)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})
	return results
}

func (m *L2Memory) Checksum() string {
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
}

func (m *L3Memory) Add(entry MemoryEntry) {
	m.Entries = append(m.Entries, entry)
}

func (m *L3Memory) Get(id string) (MemoryEntry, bool) {
	for _, e := range m.Entries {
		if e.ID == id { return e, true }
	}
	return MemoryEntry{}, false
}

func (m *L3Memory) Search(query string) []MemoryEntry {
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
	results := m.Search(query)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score() > results[j].Score()
	})
	return results
}

func (m *L3Memory) Checksum() string {
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
	L1        L1Memory          `json:"l1"`
	L2        L2Memory          `json:"l2"`
	L3        L3Memory          `json:"l3"`
	Ledger    Ledger            `json:"ledger"`
	LLM       LLMProvider       `json:"-"`
	Embedder  EmbeddingProvider `json:"-"`
	DB        *SQLiteStore      `json:"-"`
}

func NewOrchestrator() *Orchestrator {
	return &Orchestrator{
		L1:       L1Memory{Entries: make([]MemoryEntry, 0)},
		L2:       L2Memory{Entries: make([]MemoryEntry, 0)},
		L3:       L3Memory{Entries: make([]MemoryEntry, 0)},
		Ledger:   Ledger{Transactions: make([]Transaction, 0)},
		LLM:      &MockLLM{},
		Embedder: &MockEmbedder{},
	}
}

func (o *Orchestrator) Save(filepath string) error {
	if o.DB != nil {
		for _, e := range o.L1.Entries { o.DB.SaveMemory("L1", e) }
		for _, e := range o.L2.Entries { o.DB.SaveMemory("L2", e) }
		for _, e := range o.L3.Entries { o.DB.SaveMemory("L3", e) }
	}

	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
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
