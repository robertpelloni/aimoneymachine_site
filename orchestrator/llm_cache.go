package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type cacheEntry struct {
	response  string
	timestamp time.Time
}

// CacheStats reports runtime prompt-deduplication metrics for /api/status and the dashboard.
type CacheStats struct {
	Hits       uint64  `json:"hits"`
	Misses     uint64  `json:"misses"`
	HitRatePct float64 `json:"hit_rate_pct"`
	Entries    int     `json:"entries"`
	TTLSeconds int64   `json:"ttl_seconds"`
}

// CachingLLM wraps any LLMProvider with thread-safe SHA-256 prompt response caching
// so identical research/indicator/template prompts within maxAge avoid redundant proxy round-trips.
type CachingLLM struct {
	provider LLMProvider
	cache    map[string]cacheEntry
	mu       sync.RWMutex
	maxAge   time.Duration
	hits     atomic.Uint64
	misses   atomic.Uint64
}

func NewCachingLLM(p LLMProvider) *CachingLLM {
	return &CachingLLM{
		provider: p,
		cache:    make(map[string]cacheEntry),
		maxAge:   15 * time.Minute,
	}
}

// SetMaxAge configures the TTL for cached responses.
func (c *CachingLLM) SetMaxAge(age time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxAge = age
}

// ClearExpired removes items older than maxAge.
func (c *CachingLLM) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.cache {
		if now.Sub(v.timestamp) > c.maxAge {
			delete(c.cache, k)
		}
	}
}

// Stats returns a snapshot of cache performance.
func (c *CachingLLM) Stats() CacheStats {
	c.mu.RLock()
	entries := len(c.cache)
	ttl := int64(c.maxAge.Seconds())
	c.mu.RUnlock()

	hits := c.hits.Load()
	misses := c.misses.Load()
	var rate float64
	if total := hits + misses; total > 0 {
		rate = float64(int((float64(hits)/float64(total))*1000)) / 10
	}
	return CacheStats{
		Hits:       hits,
		Misses:     misses,
		HitRatePct: rate,
		Entries:    entries,
		TTLSeconds: ttl,
	}
}

// Generate implements LLMProvider with SHA-256 prompt caching.
func (c *CachingLLM) Generate(prompt string) (string, error) {
	hash := sha256.Sum256([]byte(prompt))
	key := hex.EncodeToString(hash[:])

	c.mu.RLock()
	maxAge := c.maxAge
	if entry, ok := c.cache[key]; ok && time.Since(entry.timestamp) <= maxAge {
		c.mu.RUnlock()
		c.hits.Add(1)
		return entry.response, nil
	}
	c.mu.RUnlock()
	c.misses.Add(1)

	val, err := c.provider.Generate(prompt)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.cache[key] = cacheEntry{
		response:  val,
		timestamp: time.Now(),
	}
	c.mu.Unlock()

	return val, nil
}

// GenerateJSON caches successful JSON responses keyed by "json:"+prompt and unmarshals into target.
func (c *CachingLLM) GenerateJSON(prompt string, target interface{}) error {
	hash := sha256.Sum256([]byte("json:" + prompt))
	key := hex.EncodeToString(hash[:])

	c.mu.RLock()
	maxAge := c.maxAge
	if entry, ok := c.cache[key]; ok && time.Since(entry.timestamp) <= maxAge {
		c.mu.RUnlock()
		if err := json.Unmarshal([]byte(entry.response), target); err == nil {
			c.hits.Add(1)
			return nil
		}
	} else {
		c.mu.RUnlock()
	}
	c.misses.Add(1)

	if err := c.provider.GenerateJSON(prompt, target); err != nil {
		// Fallback via raw Generate if underlying provider struggled
		raw, gerr := c.provider.Generate(prompt + "\n\nRespond ONLY with valid JSON.")
		if gerr != nil {
			return fmt.Errorf("GenerateJSON failed: %w", err)
		}
		start := strings.IndexAny(raw, "{[")
		endObj := strings.LastIndex(raw, "}")
		endArr := strings.LastIndex(raw, "]")
		end := endObj
		if endArr > end {
			end = endArr
		}
		if start < 0 || end <= start {
			return err
		}
		if uerr := json.Unmarshal([]byte(raw[start:end+1]), target); uerr != nil {
			return err
		}
	}

	if marshaled, merr := json.Marshal(target); merr == nil {
		c.mu.Lock()
		c.cache[key] = cacheEntry{
			response:  string(marshaled),
			timestamp: time.Now(),
		}
		c.mu.Unlock()
	}
	return nil
}
