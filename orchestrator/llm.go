package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type LLMProvider interface {
	Generate(prompt string) (string, error)
}

type MockLLM struct {
	Response string
	Fail     bool
}

func (m *MockLLM) Generate(prompt string) (string, error) {
	if m.Fail {
		return "", fmt.Errorf("mock provider failure")
	}
	fmt.Printf("Mock LLM generating for prompt: %s\n", prompt)
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

	url := "https://api.anthropic.com/v1/messages"

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	payload := struct {
		Model     string    `json:"model"`
		MaxTokens int       `json:"max_tokens"`
		Messages  []Message `json:"messages"`
	}{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages:  []Message{{Role: "user", Content: prompt}},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic error: %s", string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}
