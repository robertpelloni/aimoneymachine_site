package orchestrator

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type PromptVariation struct {
	Template string
	Wins     int
	Uses     int
}

type PromptOptimizer struct {
	mu         sync.RWMutex
	Variations map[string][]PromptVariation
	rnd        *rand.Rand
}

func NewPromptOptimizer() *PromptOptimizer {
	return &PromptOptimizer{
		Variations: make(map[string][]PromptVariation),
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (po *PromptOptimizer) AddVariation(category, template string) {
	po.mu.Lock()
	defer po.mu.Unlock()

	po.Variations[category] = append(po.Variations[category], PromptVariation{
		Template: template,
		Wins:     0,
		Uses:     0,
	})
}

func (po *PromptOptimizer) GetPrompt(category string) (string, int, error) {
	po.mu.Lock()
	defer po.mu.Unlock()

	vars, ok := po.Variations[category]
	if !ok || len(vars) == 0 {
		return "", -1, fmt.Errorf("no variations for category %s", category)
	}

	// Simple epsilon-greedy strategy
	epsilon := 0.2
	var bestIdx int
	var bestWinRate float64 = -1.0

	// Force bestIdx = 0 if uses == 0 so we always try them first in order
	allTested := true
	for i, v := range vars {
		if v.Uses == 0 {
			bestIdx = i
			allTested = false
			break
		}
	}

	if allTested && po.rnd.Float64() < epsilon {
		bestIdx = po.rnd.Intn(len(vars))
	} else if allTested {
		for i, v := range vars {
			winRate := float64(v.Wins) / float64(v.Uses)
			if winRate > bestWinRate {
				bestWinRate = winRate
				bestIdx = i
			}
		}
	}

	po.Variations[category][bestIdx].Uses++
	return po.Variations[category][bestIdx].Template, bestIdx, nil
}

func (po *PromptOptimizer) RecordWin(category string, index int) {
	po.mu.Lock()
	defer po.mu.Unlock()

	if vars, ok := po.Variations[category]; ok && index >= 0 && index < len(vars) {
		po.Variations[category][index].Wins++
	}
}

func (po *PromptOptimizer) RecordLoss(category string, index int) {
	// A loss is just a use without a win, we already incremented Uses
}
