package trading

import (
	"github.com/robertpelloni/hustle/orchestrator"
	"testing"
)

func TestTradingStrategy(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	fetcher := &MockPriceFetcher{}
	trader := &TradingModule{
		Orchestrator: orch,
		Symbol:       "BTC",
		Fetcher:      fetcher,
	}

	// Run 15 times to build history for RSI(14)
	for i := 0; i < 15; i++ {
		err := trader.ExecuteStrategy()
		if err != nil {
			t.Fatalf("Strategy execution failed on iteration %d: %v", i, err)
		}
	}

	if len(trader.History) != 15 {
		t.Errorf("Expected history length 15, got %d", len(trader.History))
	}
}
