package realestate

import (
	"fmt"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
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

// CalculateSTRProfit evaluates a short-term rental (Airbnb) arbitrage opportunity.
func (r *RealEstateModule) CalculateSTRProfit(monthlyRent, nightlyRate float64, location string) (float64, string, error) {
	if monthlyRent <= 0 {
		monthlyRent = 2000.0
	}
	if nightlyRate <= 0 {
		nightlyRate = 165.0
	}
	fmt.Printf("[RealEstate] Calculating STR Profit for %s: Rent=$%.2f, Rate=$%.2f\n", location, monthlyRent, nightlyRate)

	revenue := nightlyRate * 30 * 0.65
	profit := revenue - monthlyRent - (revenue * 0.15)
	margin := (profit / monthlyRent) * 100.0

	prompt := fmt.Sprintf(`Act as an Airbnb Arbitrage expert. Analyze this opportunity:
Location: %s
Monthly Rent: $%.2f
Target Nightly Rate: $%.2f
Estimated Monthly Profit: $%.2f
Profit Margin: %.1f%%

Provide a 1-sentence risk assessment regarding seasonal demand and local regulations.`, location, monthlyRent, nightlyRate, profit, margin)

	assessment, err := r.Orch.LLM.Generate(prompt)
	if err != nil {
		return profit, "Error during assessment", err
	}

	r.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("realestate-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("STR Arb Opportunity: %s (Profit: $%.2f, Margin: %.1f%%) - %s", location, profit, margin, assessment),
		Timestamp: time.Now(),
		Tags:      []string{"realestate", "airbnb", location},
	})

	return profit, assessment, nil
}

// ManageListing handles guest communication and dynamic pricing for a property.
func (r *RealEstateModule) ManageListing(location, guestMessage string) (string, error) {
	fmt.Printf("[RealEstate] Managing STR for %s\n", location)

	prompt := fmt.Sprintf(`Act as an automated Airbnb property manager for a listing in: "%s".

GUEST MESSAGE:
%s

Requirements:
- Empathetic and helpful response to the guest
- Provide check-in/out instructions placeholder
- Recommend a dynamic price adjustment (increase/decrease/hold) based on demand context
- Professional and welcoming tone

Format: Professional Markdown reply.`, location, guestMessage)

	reply, err := r.Orch.LLM.Generate(prompt)
	if err == nil {
		r.Orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("realestate-msg-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("STR Guest Reply (%s):\n%s", location, reply),
			Timestamp: time.Now(),
			Tags:      []string{"realestate", "guest", location},
		})
	}
	return reply, err
}
