package orchestrator

import (
	"fmt"
	"testing"
	"time"
)

func TestCachingLLM(t *testing.T) {
	mockCount := 0
	mock := &MockLLM{
		GenerateFunc: func(prompt string) (string, error) {
			mockCount++
			return fmt.Sprintf("Response %d", mockCount), nil
		},
	}

	cache := NewCachingLLM(mock)

	// Test basic cache hit
	res1, err := cache.Generate("test prompt")
	if err != nil {
		t.Fatal(err)
	}
	if res1 != "Response 1" {
		t.Fatalf("expected Response 1, got %v", res1)
	}

	res2, err := cache.Generate("test prompt")
	if err != nil {
		t.Fatal(err)
	}
	if res2 != "Response 1" {
		t.Fatalf("expected cached Response 1, got %v", res2)
	}

	if mockCount != 1 {
		t.Fatalf("expected mock to be called once, called %d times", mockCount)
	}

	// Test cache miss on different prompt
	res3, err := cache.Generate("test prompt 2")
	if err != nil {
		t.Fatal(err)
	}
	if res3 != "Response 2" {
		t.Fatalf("expected Response 2, got %v", res3)
	}

	if mockCount != 2 {
		t.Fatalf("expected mock to be called twice, called %d times", mockCount)
	}

	// Test expiration
	cache.SetMaxAge(1 * time.Millisecond)
	time.Sleep(2 * time.Millisecond)

	res4, err := cache.Generate("test prompt")
	if err != nil {
		t.Fatal(err)
	}
	if res4 != "Response 3" {
		t.Fatalf("expected Response 3 after expiration, got %v", res4)
	}

	if mockCount != 3 {
		t.Fatalf("expected mock to be called thrice, called %d times", mockCount)
	}

	cache.ClearExpired()
}
