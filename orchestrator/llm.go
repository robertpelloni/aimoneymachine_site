package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LLMProvider interface {
	Generate(prompt string) (string, error)
	GenerateJSON(prompt string, target interface{}) error
}

type EmbeddingProvider interface {
	Embed(text string) ([]float32, error)
}

// --- RetryWrapper wraps any LLMProvider with indefinite exponential-backoff retry ---
type RetryWrapper struct {
	Provider  LLMProvider
	MaxWait   time.Duration // max total wait before giving up (0 = wait forever)
	BaseDelay time.Duration // initial delay between retries (default: 2s)
	MaxDelay  time.Duration // max delay between retries (default: 5min)
	Verbose   bool
}

func NewRetryWrapper(provider LLMProvider) *RetryWrapper {
	return &RetryWrapper{
		Provider:  provider,
		MaxWait:   0, // wait forever
		BaseDelay: 2 * time.Second,
		MaxDelay:  5 * time.Minute,
		Verbose:   true,
	}
}

func (r *RetryWrapper) Generate(prompt string) (string, error) {
	delay := r.BaseDelay
	start := time.Now()
	attempt := 0
	for {
		attempt++
		resp, err := r.Provider.Generate(prompt)
		if err == nil {
			return resp, nil
		}
		if r.MaxWait > 0 && time.Since(start) > r.MaxWait {
			return "", fmt.Errorf("LLM retry exhausted after %v: %v", time.Since(start), err)
		}
		if r.Verbose {
			fmt.Printf("[LLM Retry] attempt %d failed (%v). Waiting %v before retry...\n", attempt, err, delay)
		}
		time.Sleep(delay)
		delay *= 2
		if delay > r.MaxDelay {
			delay = r.MaxDelay
		}
	}
}

func (r *RetryWrapper) GenerateJSON(prompt string, target interface{}) error {
	jsonPrompt := prompt + "\n\nRespond ONLY with valid JSON, no markdown fences."
	resp, err := r.Generate(jsonPrompt)
	if err != nil {
		return err
	}
	// Strip markdown code fences if present
	cleaned := resp
	if idx := findFence(cleaned); idx >= 0 {
		start := idx
		end := lastFence(cleaned)
		if end > start {
			fenced := cleaned[start+3 : end]
			if nlIdx := indexByte(fenced, '\n'); nlIdx >= 0 {
				fenced = fenced[nlIdx+1:]
			}
			cleaned = fenced
		}
	}
	cleaned = trimSpace(cleaned)
	return json.Unmarshal([]byte(cleaned), target)
}

func findFence(s string) int {
	for i := 0; i < len(s)-2; i++ {
		if s[i] == '`' && s[i+1] == '`' && s[i+2] == '`' {
			return i
		}
	}
	return -1
}

func lastFence(s string) int {
	for i := len(s) - 3; i >= 0; i-- {
		if s[i] == '`' && s[i+1] == '`' && s[i+2] == '`' {
			return i
		}
	}
	return -1
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	if start >= end {
		return ""
	}
	return s[start:end]
}

// --- MockEmbedder ---
type MockEmbedder struct{}

func (m *MockEmbedder) Embed(text string) ([]float32, error) {
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = float32(i) / 128.0
	}
	return vec, nil
}

// --- MockLLM ---
type MockLLM struct {
	Response     string
	GenerateFunc func(prompt string) (string, error)
}

func (m *MockLLM) Generate(prompt string) (string, error) {
	fmt.Printf("Mock LLM generating for prompt: %s\n", prompt)
	if m.GenerateFunc != nil {
		return m.GenerateFunc(prompt)
	}
	if m.Response != "" {
		return m.Response, nil
	}
	return "Generated content based on " + prompt, nil
}

func (m *MockLLM) GenerateJSON(prompt string, target interface{}) error {
	_, err := m.Generate(prompt)
	return err
}

// --- AnthropicProvider ---
type AnthropicProvider struct {
	APIKey string
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{APIKey: os.Getenv("ANTHROPIC_API_KEY")}
}

func (p *AnthropicProvider) Generate(prompt string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	fmt.Printf("Anthropic LLM generating for prompt: %s\n", prompt)
	return "Anthropic synthesized content for: " + prompt, nil
}

func (p *AnthropicProvider) GenerateJSON(prompt string, target interface{}) error {
	return fmt.Errorf("GenerateJSON not implemented for Anthropic")
}

// --- CachingLLM wraps any LLMProvider with prompt-response caching ---
type CachingLLM struct {
	Provider LLMProvider
	Store    *SQLiteStore
}

func (c *CachingLLM) Generate(prompt string) (string, error) {
	if c.Store == nil {
		return c.Provider.Generate(prompt)
	}

	hash := sha256.Sum256([]byte(prompt))
	hashStr := hex.EncodeToString(hash[:])

	if resp, ok := c.Store.GetCachedResponse(hashStr); ok {
		fmt.Printf("[LLM Cache] HIT: %s...\n", hashStr[:8])
		return resp, nil
	}

	fmt.Printf("[LLM Cache] MISS: Generating fresh response\n")
	resp, err := c.Provider.Generate(prompt)
	if err != nil {
		return "", err
	}

	c.Store.CacheResponse(hashStr, resp, "unknown")
	return resp, nil
}

func (c *CachingLLM) GenerateJSON(prompt string, target interface{}) error {
	jsonPrompt := prompt + "\n\nRespond ONLY with valid JSON, no markdown fences."

	hash := sha256.Sum256([]byte(jsonPrompt))
	hashStr := hex.EncodeToString(hash[:])

	if resp, ok := c.Store.GetCachedResponse(hashStr); ok {
		fmt.Printf("[LLM Cache JSON] HIT: %s...\n", hashStr[:8])
		return json.Unmarshal([]byte(resp), target)
	}

	fmt.Printf("[LLM Cache JSON] MISS: Generating fresh JSON response\n")
	if err := c.Provider.GenerateJSON(prompt, target); err != nil {
		return err
	}

	raw, _ := json.Marshal(target)
	c.Store.CacheResponse(hashStr, string(raw), "unknown")
	return nil
}
