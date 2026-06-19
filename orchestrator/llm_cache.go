package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type cacheEntry struct {
	response  string
	timestamp time.Time
}

type CachingLLM struct {
	provider LLMProvider
	cache    map[string]cacheEntry
	mu       sync.RWMutex
	maxAge   time.Duration
}

func NewCachingLLM(p LLMProvider) *CachingLLM {
	return &CachingLLM{
		provider: p,
		cache:    make(map[string]cacheEntry),
		maxAge:   24 * time.Hour, // Default TTL of 24 hours
	}
}

// SetMaxAge configures the TTL for cached responses
func (c *CachingLLM) SetMaxAge(age time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxAge = age
}

// ClearExpired removes items older than maxAge
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

// Generate implements the LLMProvider interface with caching
func (c *CachingLLM) Generate(prompt string) (string, error) {
	hash := sha256.Sum256([]byte(prompt))
	key := hex.EncodeToString(hash[:])

	c.mu.RLock()
	if entry, ok := c.cache[key]; ok {
		if time.Since(entry.timestamp) <= c.maxAge {
			c.mu.RUnlock()
			return entry.response, nil
		}
	}
	c.mu.RUnlock()

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
