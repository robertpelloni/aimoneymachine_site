package orchestrator

import (
	"errors"
	"strings"
	"testing"
)

// A stalling/failing primary must not fail the generation when a
// fallback provider can answer (the FreeLLM 480s-timeout scenario).
func TestWaterfallFallsBackAfterPrimaryFailure(t *testing.T) {
	failing := &MockLLM{GenerateFunc: func(string) (string, error) {
		return "", errors.New("Post http://localhost:4000: context deadline exceeded (Client.Timeout)")
	}}
	ok := &MockLLM{GenerateFunc: func(string) (string, error) { return "fallback answer", nil }}
	w := &WaterfallLLM{Providers: []LLMProvider{failing, ok}}
	got, err := w.Generate("hello")
	if err != nil || got != "fallback answer" {
		t.Fatalf("Generate = %q, %v; want fallback answer", got, err)
	}
}

func TestWaterfallReturnsErrorWhenAllFail(t *testing.T) {
	failing := &MockLLM{GenerateFunc: func(string) (string, error) {
		return "", errors.New("boom")
	}}
	w := &WaterfallLLM{Providers: []LLMProvider{failing, failing}}
	if _, err := w.Generate("hello"); err == nil || !strings.Contains(err.Error(), "all LLM providers failed") {
		t.Fatalf("err = %v, want all-providers-failed", err)
	}
}
