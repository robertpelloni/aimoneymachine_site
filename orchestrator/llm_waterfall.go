package orchestrator

import (
	"fmt"
	"strings"
)

type WaterfallLLM struct {
	Providers []LLMProvider
}

func (w *WaterfallLLM) Generate(prompt string) (string, error) {
	var lastErr error
	for i, p := range w.Providers {
		res, err := p.Generate(prompt)
		if err == nil {
			// Successful generation
			return res, nil
		}

		// Intelligent error classification
		isRetryable := w.isRetryable(err)
		lastErr = err

		fmt.Printf("LLM Provider %d failed (Retryable: %v): %v\n", i, isRetryable, err)

		if !isRetryable {
			// If it's a fatal error (e.g. prompt too long), don't bother falling back
			return "", fmt.Errorf("fatal LLM error: %v", err)
		}
	}
	return "", fmt.Errorf("all LLM providers failed: %v", lastErr)
}

func (w *WaterfallLLM) isRetryable(err error) bool {
	msg := strings.ToLower(err.Error())
	// Non-retryable patterns
	if strings.Contains(msg, "too long") ||
	   strings.Contains(msg, "invalid_request_error") ||
	   strings.Contains(msg, "unauthorized") {
		return false
	}
	return true
}
