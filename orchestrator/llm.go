package orchestrator

import (
	"fmt"
	"os"
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
	if m.Response != "" {
		return m.Response, nil
	}
	return "Generated content based on " + prompt, nil
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
