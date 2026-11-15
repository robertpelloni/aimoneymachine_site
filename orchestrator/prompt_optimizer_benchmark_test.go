package orchestrator

import (
	"fmt"
	"testing"
)

func BenchmarkPromptOptimizer(b *testing.B) {
	po := NewPromptOptimizer()
	po.AddVariation("bench", "A")
	po.AddVariation("bench", "B")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, idx, _ := po.GetPrompt("bench")
		if i%3 == 0 {
			po.RecordWin("bench", idx)
		} else {
			po.RecordLoss("bench", idx)
		}
	}
}

func BenchmarkCachingLLM(b *testing.B) {
	mockCount := 0
	mock := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			mockCount++
			return fmt.Sprintf("Response %d", mockCount), nil
		},
	}
	cache := NewCachingLLM(mock)

	// Pre-populate cache to test hit performance
	cache.Generate("test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Generate("test")
	}
}
