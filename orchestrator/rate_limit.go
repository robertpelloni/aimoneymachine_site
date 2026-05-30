package orchestrator

import (
	"context"
	"golang.org/x/time/rate"
)

// RateLimiter wraps golang.org/x/time/rate.Limiter
type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	return r.limiter.Wait(ctx)
}

// LimitedLLM wraps an LLMProvider with rate limiting
type LimitedLLM struct {
	Provider LLMProvider
	Limiter  *RateLimiter
}

func (l *LimitedLLM) Generate(prompt string) (string, error) {
	err := l.Limiter.Wait(context.Background())
	if err != nil {
		return "", err
	}
	return l.Provider.Generate(prompt)
}
