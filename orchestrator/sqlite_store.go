package orchestrator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"math"
	"sort"
	"time"
)

type SQLiteStore struct {
	db        *sql.DB
	extLoaded bool
}

func NewSQLiteStore(filepath string) (*SQLiteStore, error) {
	// Enable loading extensions if available in the environment
	db, err := sql.Open("sqlite", filepath+"?_load_extension=1")
	if err != nil {
		return nil, err
	}

	// Attempt to load sqlite-vec if it exists in standard paths
	extLoaded := false
	if _, err := db.Exec("SELECT load_extension('vec0')"); err == nil {
		fmt.Println("[SQLite] sqlite-vec extension loaded successfully.")
		extLoaded = true
	} else {
		fmt.Printf("[SQLite] Warning: failed to load sqlite-vec: %v. Using Go-level fallback.\n", err)
	}

	if extLoaded {
		_, err = db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS vec_memories USING vec0(
			embedding FLOAT[1536]
		);`)
		if err != nil {
			fmt.Printf("[SQLite] Warning: failed to create virtual table: %v\n", err)
			extLoaded = false
		}
	}

	// Standard table for metadata
	query := `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		tier TEXT,
		content TEXT,
		base_score REAL,
		timestamp DATETIME,
		tags TEXT,
		embedding BLOB
	);`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	// Task History table
	taskQuery := `
	CREATE TABLE IF NOT EXISTS task_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT,
		duration_ms INTEGER,
		status TEXT,
		message TEXT,
		timestamp DATETIME
	);`
	_, err = db.Exec(taskQuery)
	if err != nil {
		return nil, err
	}

	// LLM Cache table
	cacheQuery := `
	CREATE TABLE IF NOT EXISTS llm_cache (
		prompt_hash TEXT PRIMARY KEY,
		prompt TEXT,
		response TEXT,
		timestamp DATETIME
	);`
	_, err = db.Exec(cacheQuery)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db, extLoaded: extLoaded}, nil
}

func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteStore) SaveMemory(tier string, entry MemoryEntry) error {
	tagsData, _ := json.Marshal(entry.Tags)
	var embeddingData []byte
	if len(entry.Vector) > 0 {
		embeddingData, _ = json.Marshal(entry.Vector)
	}
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO memories (id, tier, content, base_score, timestamp, tags, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, tier, entry.Content, entry.BaseScore, entry.Timestamp, string(tagsData), embeddingData)
	if err != nil {
		return err
	}
	if s.extLoaded && len(entry.Vector) > 0 {
		_, err = s.db.Exec(`
			INSERT OR REPLACE INTO vec_memories(rowid, embedding)
			VALUES ((SELECT rowid FROM memories WHERE id = ?), ?)`,
			entry.ID, embeddingData)
		if err != nil {
			fmt.Printf("[SQLite] Warning: failed to update vec_memories: %v\n", err)
		}
	}
	return nil
}

func (s *SQLiteStore) LogTaskExecution(taskID string, duration time.Duration, status string, message string) error {
	_, err := s.db.Exec(`
		INSERT INTO task_history (task_id, duration_ms, status, message, timestamp)
		VALUES (?, ?, ?, ?, ?)`,
		taskID, duration.Milliseconds(), status, message, time.Now())
	return err
}

type TaskHistoryEntry struct {
	ID         int       `json:"id"`
	TaskID     string    `json:"task_id"`
	DurationMs int64     `json:"duration_ms"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

func (s *SQLiteStore) GetTaskHistory(limit int) ([]TaskHistoryEntry, error) {
	rows, err := s.db.Query("SELECT id, task_id, duration_ms, status, message, timestamp FROM task_history ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var history []TaskHistoryEntry
	for rows.Next() {
		var h TaskHistoryEntry
		err := rows.Scan(&h.ID, &h.TaskID, &h.DurationMs, &h.Status, &h.Message, &h.Timestamp)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (s *SQLiteStore) LoadMemories(tier string) ([]MemoryEntry, error) {
	rows, err := s.db.Query("SELECT id, content, base_score, timestamp, tags, embedding FROM memories WHERE tier = ?", tier)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		var tagsStr string
		var embeddingData []byte
		err := rows.Scan(&e.ID, &e.Content, &e.BaseScore, &e.Timestamp, &tagsStr, &embeddingData)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(tagsStr), &e.Tags)
		if len(embeddingData) > 0 {
			json.Unmarshal(embeddingData, &e.Vector)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *SQLiteStore) GetLLMCache(promptHash string) (string, error) {
	var response string
	err := s.db.QueryRow("SELECT response FROM llm_cache WHERE prompt_hash = ?", promptHash).Scan(&response)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return response, err
}

func (s *SQLiteStore) SaveLLMCache(promptHash, prompt, response string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO llm_cache (prompt_hash, prompt, response, timestamp)
		VALUES (?, ?, ?, ?)`,
		promptHash, prompt, response, time.Now())
	return err
}

func (s *SQLiteStore) ClearLLMCache() error {
	_, err := s.db.Exec("DELETE FROM llm_cache")
	return err
}

func (s *SQLiteStore) GetCacheCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM llm_cache").Scan(&count)
	return count, err
}

func (s *SQLiteStore) VectorSearch(tier string, target []float32, limit int) ([]MemoryEntry, error) {
	// Simple cosine fallback if extension not loaded
	memories, err := s.LoadMemories(tier)
	if err != nil {
		return nil, err
	}

	type rankedResult struct {
		entry MemoryEntry
		sim   float64
	}
	var ranked []rankedResult

	for _, m := range memories {
		if len(m.Vector) == 0 { continue }
		sim := cosineSimilarity(m.Vector, target)
		ranked = append(ranked, rankedResult{m, sim})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].sim > ranked[j].sim
	})

	var results []MemoryEntry
	for i := 0; i < len(ranked) && i < limit; i++ {
		results = append(results, ranked[i].entry)
	}
	return results, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 { return 0 }
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	if normA == 0 || normB == 0 { return 0 }
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
