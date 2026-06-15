package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

type LLMProvider interface {
	Generate(prompt string) (string, error)
	GenerateJSON(prompt string, target interface{}) error
}

type EmbeddingProvider interface {
	Embed(text string) ([]float32, error)
}

type MockEmbedder struct{}

func (m *MockEmbedder) Embed(text string) ([]float32, error) {
	// Return a dummy 128-dim vector
	vec := make([]float32, 128)
	for i := range vec {
		vec[i] = float32(i) / 128.0
	}
	return vec, nil
}

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

type AnthropicProvider struct {
	APIKey string
}

func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}
}

func (p *AnthropicProvider) Generate(prompt string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}
	fmt.Printf("Anthropic LLM generating for prompt: %s\n", prompt)
	// Real API call logic would go here
	return "Anthropic synthesized content for: " + prompt, nil
}

func (p *AnthropicProvider) GenerateJSON(prompt string, target interface{}) error {
	return fmt.Errorf("GenerateJSON not implemented for Anthropic")
}

type CachingLLM struct {
	Provider LLMProvider
	Store    *SQLiteStore
}

func (c *CachingLLM) Generate(prompt string) (string, error) {
	if c.Store == nil {
		return c.Provider.Generate(prompt)
	}

	// Hash the prompt
	hash := sha256.Sum256([]byte(prompt))
	hashStr := hex.EncodeToString(hash[:])

	// Check cache
	if resp, ok := c.Store.GetCachedResponse(hashStr); ok {
		fmt.Printf("[LLM Cache] HIT: %s...\n", hashStr[:8])
		return resp, nil
	}

	fmt.Printf("[LLM Cache] MISS: Generating fresh response\n")
	resp, err := c.Provider.Generate(prompt)
	if err != nil {
		return "", err
	}

	// Cache result
	model := "unknown"
	if p, ok := c.Provider.(*OpenAICompatProvider); ok {
		model = p.Model
	}
	c.Store.CacheResponse(hashStr, resp, model)

	return resp, nil
}

func (c *CachingLLM) GenerateJSON(prompt string, target interface{}) error {
	// Add instruction to prompt for caching uniqueness
	jsonPrompt := prompt + "\n\nRespond ONLY with valid JSON, no markdown fences."

	// Hash the prompt
	hash := sha256.Sum256([]byte(jsonPrompt))
	hashStr := hex.EncodeToString(hash[:])

	// Check cache
	if resp, ok := c.Store.GetCachedResponse(hashStr); ok {
		fmt.Printf("[LLM Cache JSON] HIT: %s...\n", hashStr[:8])
		return json.Unmarshal([]byte(resp), target)
	}

	fmt.Printf("[LLM Cache JSON] MISS: Generating fresh JSON response\n")
	err := c.Provider.GenerateJSON(prompt, target)
	if err != nil {
		return err
	}

	// We need the raw string to cache it. Since GenerateJSON already unmarshaled into target,
	// we could re-marshal it, or refactor to get the raw string.
	// For simplicity and to ensure we cache exactly what the provider produced:
	raw, _ := json.Marshal(target)
	model := "unknown"
	if p, ok := c.Provider.(*OpenAICompatProvider); ok {
		model = p.Model
	}
	c.Store.CacheResponse(hashStr, string(raw), model)

	return nil
}
