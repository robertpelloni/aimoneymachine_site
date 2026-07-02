package orchestrator

import (
	"fmt"
	"strings"
	"os"
	"time"
)

type LLMProvider interface {
	Generate(prompt string) (string, error)
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
	if strings.Contains(prompt, "Respond with ONLY the hustle:// URI") {
		return "hustle://synergy_leadgen", nil
	}

	if strings.Contains(prompt, "Respond with a JSON array") {
		return `[{"name": "mock_hustle", "description": "mock", "category": "Research", "steps": ["hustle://research"], "interval_min": 60, "priority": "high"}]`, nil
	}

	if m.Response != "" {
		return m.Response, nil
	}
	return "Generated content based on " + prompt, nil
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
	// Real API call logic would go here
	return "Anthropic synthesized content for: " + prompt, nil
}
