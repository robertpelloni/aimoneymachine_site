package retail

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type RetailModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewRetailModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *RetailModule {
	return &RetailModule{
		Orch:   orch,
		Broker: broker,
	}
}

// CalculateArbitrageROI determines if a retail arbitrage opportunity is profitable
func (r *RetailModule) CalculateArbitrageROI(buyPrice, sellPrice float64, productCategory string) (float64, string, error) {
	fmt.Printf("[Retail] Calculating ROI for %s: Buy=$%.2f, Sell=$%.2f\n", productCategory, buyPrice, sellPrice)

	// Naive fee calculation: 15% Amazon referral fee + $5 estimated FBA fulfillment fee
	fees := (sellPrice * 0.15) + 5.0
	profit := sellPrice - buyPrice - fees
	roi := (profit / buyPrice) * 100.0

	prompt := fmt.Sprintf(`Act as an Amazon FBA expert. Analyze this arbitrage opportunity:
Category: %s
Buy Price: $%.2f
Target Sell Price: $%.2f
Estimated Profit: $%.2f
ROI: %.1f%%

Provide a 1-sentence risk assessment and recommend 'BUY' or 'SKIP'.`, productCategory, buyPrice, sellPrice, profit, roi)

	assessment, err := r.Orch.LLM.Generate(prompt)
	if err != nil { return roi, "Error during assessment", err }

	// Store in memory
	r.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("retail-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Retail Arb Opportunity: %s (ROI: %.1f%%) - %s", productCategory, roi, assessment),
		Timestamp: time.Now(),
		Tags:      []string{"retail", "arbitrage", productCategory},
	})

	return roi, assessment, nil
}
