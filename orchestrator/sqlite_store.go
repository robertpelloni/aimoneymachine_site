package orchestrator

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(filepath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
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

	return &SQLiteStore{db: db}, nil
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
	return err
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

// VectorSearch finds similar memories using cosine similarity (mocked for now until sqlite-vec is loaded)
func (s *SQLiteStore) VectorSearch(tier string, target []float32, limit int) ([]MemoryEntry, error) {
	// In a real sqlite-vec setup, we would use:
	// SELECT id, content, distance FROM memories WHERE tier = ? ORDER BY vec_distance_cosine(embedding, ?) LIMIT ?

	// For alpha, we load and calculate manually
	memories, err := s.LoadMemories(tier)
	if err != nil {
		return nil, err
	}

	// Simple mock: just return first N for now
	if len(memories) > limit {
		return memories[:limit], nil
	}
	return memories, nil
}
