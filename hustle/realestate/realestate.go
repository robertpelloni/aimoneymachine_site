package realestate

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type RealEstateModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewRealEstateModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *RealEstateModule {
	return &RealEstateModule{
		Orch:   orch,
		Broker: broker,
	}
}

// CalculateSTRProfit evaluates an Airbnb arbitrage opportunity
func (r *RealEstateModule) CalculateSTRProfit(monthlyRent, nightlyRate float64, location string) (float64, string, error) {
	fmt.Printf("[RealEstate] Calculating STR Profit for %s: Rent=$%.2f, Rate=$%.2f\n", location, monthlyRent, nightlyRate)

	// Naive calculation: 65% occupancy
	revenue := nightlyRate * 30 * 0.65
	profit := revenue - monthlyRent - (revenue * 0.15) // 15% for cleaning/supplies
	margin := (profit / monthlyRent) * 100.0

	prompt := fmt.Sprintf(`Act as an Airbnb Arbitrage expert. Analyze this opportunity:
Location: %s
Monthly Rent: $%.2f
Target Nightly Rate: $%.2f
Estimated Monthly Profit: $%.2f
Profit Margin: %.1f%%

Provide a 1-sentence risk assessment regarding seasonal demand and local regulations.`, location, monthlyRent, nightlyRate, profit, margin)

	assessment, err := r.Orch.LLM.Generate(prompt)
	if err != nil { return profit, "Error during assessment", err }

	// Store in memory
	r.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("realestate-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("STR Arb Opportunity: %s (Profit: $%.2f) - %s", location, profit, assessment),
		Timestamp: time.Now(),
		Tags:      []string{"realestate", "airbnb", location},
	})

	return profit, assessment, nil
}
