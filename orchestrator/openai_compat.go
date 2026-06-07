package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAICompatProvider implements LLMProvider for any OpenAI-compatible API
// (LM Studio, Ollama, vLLM, LocalAI, etc.)
type OpenAICompatProvider struct {
	BaseURL    string
	APIKey     string
	Model      string
	MaxTokens  int
	HTTPClient *http.Client
}

type chatRequest struct {
	Model       string          `json:"model"`
	Messages    []chatMessage   `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewOpenAICompatProvider creates a provider from env vars or defaults
// Env: LLM_BASE_URL (default http://localhost:1234/v1)
// Env: LLM_API_KEY (default from LLM_API_KEY env or placeholder for local servers)
// Env: LLM_MODEL (auto-detected if empty)
func NewOpenAICompatProvider() *OpenAICompatProvider {
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1234/v1"
	}
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("LM_STUDIO_KEY") // LM Studio accepts any string, but read from env
	}
	model := os.Getenv("LLM_MODEL") // empty = auto-detect

	return &OpenAICompatProvider{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		APIKey:    apiKey,
		Model:     model,
		MaxTokens: 2048,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // Local models can be slow
		},
	}
}

// NewOllamaProvider creates a provider configured for Ollama's OpenAI compat endpoint
func NewOllamaProvider(model string) *OpenAICompatProvider {
	return &OpenAICompatProvider{
		BaseURL:   "http://localhost:11434/v1",
		APIKey:    os.Getenv("OLLAMA_API_KEY"), // typically empty for local Ollama
		Model:     model,
		MaxTokens: 2048,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// DetectModel queries the /models endpoint and returns the first available chat model
func (p *OpenAICompatProvider) DetectModel() (string, error) {
	req, _ := http.NewRequest("GET", p.BaseURL+"/models", nil)
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot reach LLM server at %s: %v", p.BaseURL, err)
	}
	defer resp.Body.Close()

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return "", fmt.Errorf("failed to parse models response: %v", err)
	}
	if len(models.Data) == 0 {
		return "", fmt.Errorf("no models available at %s", p.BaseURL)
	}
	// Prefer non-embedding models
	for _, m := range models.Data {
		if !strings.Contains(strings.ToLower(m.ID), "embed") {
			return m.ID, nil
		}
	}
	return models.Data[0].ID, nil
}

// Generate implements LLMProvider
func (p *OpenAICompatProvider) Generate(prompt string) (string, error) {
	model := p.Model
	if model == "" {
		detected, err := p.DetectModel()
		if err != nil {
			return "", fmt.Errorf("no model specified and auto-detect failed: %v", err)
		}
		model = detected
		fmt.Printf("[LLM] Auto-detected model: %s\n", model)
	}

	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are an AI hustle automation agent. Be concise, actionable, and output valid JSON when asked for structured data. Focus on high-ROI, low-maintenance strategies."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   p.MaxTokens,
		Temperature: 0.7,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse LLM response: %v (body: %s)", err, string(body[:min(len(body), 200)]))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM API error: %s (%s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	return content, nil
}

// GenerateJSON is a helper that requests JSON output and parses it
func (p *OpenAICompatProvider) GenerateJSON(prompt string, target interface{}) error {
	raw, err := p.Generate(prompt + "\n\nRespond ONLY with valid JSON, no markdown fences.")
	if err != nil {
		return err
	}
	// Strip markdown code fences if present
	cleaned := raw
	if strings.Contains(cleaned, "```") {
		start := strings.Index(cleaned, "```")
		end := strings.LastIndex(cleaned, "```")
		if start >= 0 && end > start {
			fenced := cleaned[start+3 : end]
			// Strip optional language tag on first line
			if nlIdx := strings.Index(fenced, "\n"); nlIdx >= 0 {
				fenced = fenced[nlIdx+1:]
			}
			cleaned = fenced
		}
	}
	cleaned = strings.TrimSpace(cleaned)
	return json.Unmarshal([]byte(cleaned), target)
}

// OpenAICompatEmbedder implements EmbeddingProvider via OpenAI-compatible /embeddings
type OpenAICompatEmbedder struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewOpenAICompatEmbedder creates an embedding provider from env or defaults
// Env: EMBED_BASE_URL (default http://localhost:1234/v1)
// Env: EMBED_MODEL (default nomic-embed-text — auto-detected if empty)
func NewOpenAICompatEmbedder() *OpenAICompatEmbedder {
	baseURL := os.Getenv("EMBED_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1234/v1"
	}
	apiKey := os.Getenv("EMBED_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("LM_STUDIO_KEY") // placeholder for local servers
	}
	model := os.Getenv("EMBED_MODEL")

	return &OpenAICompatEmbedder{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		APIKey:     apiKey,
		Model:      model,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// DetectModel finds the first embedding model
func (e *OpenAICompatEmbedder) DetectModel() (string, error) {
	req, _ := http.NewRequest("GET", e.BaseURL+"/models", nil)
	if e.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.APIKey)
	}
	resp, err := e.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot reach embedding server: %v", err)
	}
	defer resp.Body.Close()

	var models struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return "", fmt.Errorf("failed to parse models: %v", err)
	}
	// Prefer models with "embed" in the name
	for _, m := range models.Data {
		if strings.Contains(strings.ToLower(m.ID), "embed") || strings.Contains(strings.ToLower(m.ID), "nomic") {
			return m.ID, nil
		}
	}
	return "", fmt.Errorf("no embedding model found at %s", e.BaseURL)
}

// Embed implements EmbeddingProvider
func (e *OpenAICompatEmbedder) Embed(text string) ([]float32, error) {
	model := e.Model
	if model == "" {
		detected, err := e.DetectModel()
		if err != nil {
			// Fall back to mock if no embedding server
			mock := &MockEmbedder{}
			return mock.Embed(text)
		}
		model = detected
	}

	reqBody := embeddingRequest{
		Model: model,
		Input: []string{text},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", e.BaseURL+"/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if e.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.APIKey)
	}

	resp, err := e.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %v", err)
	}
	if embResp.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", embResp.Error.Message)
	}
	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embResp.Data[0].Embedding, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
