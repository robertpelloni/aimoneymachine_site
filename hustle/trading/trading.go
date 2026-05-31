package trading

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"math/rand"
	"time"
)

type PriceFetcher interface {
	GetPrice(symbol string) (float64, error)
}

type MockPriceFetcher struct{}

func (m *MockPriceFetcher) GetPrice(symbol string) (float64, error) {
	// Simple mock price generation
	return 50000.0 + rand.Float64()*1000.0, nil
}

type TradingModule struct {
	Orchestrator *orchestrator.Orchestrator
	Symbol       string
	Fetcher      PriceFetcher
	History      []float64
	RSIHistory   []float64
	Watchlist    []string
}

func (t *TradingModule) ExecuteStrategy() error {
	fmt.Printf("[Trading] Executing strategy for Symbol: %s\n", t.Symbol)

	price, err := t.Fetcher.GetPrice(t.Symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %v", err)
	}
	t.History = append(t.History, price)

	fmt.Printf("[Trading] Current Price for %s: $%.2f\n", t.Symbol, price)

	// Technical Indicator: Simple Moving Average (SMA)
	sma := t.calculateSMA(5)

	// Technical Indicator: RSI
	rsi := t.calculateRSI(14)
	t.RSIHistory = append(t.RSIHistory, rsi)

	fmt.Printf("[Trading] Indicators -> SMA(5): $%.2f | RSI(14): %.2f\n", sma, rsi)

	divergence := t.detectDivergence()
	if divergence != "" {
		fmt.Printf("[Trading] DIVERGENCE DETECTED: %s\n", divergence)
	}

	decision := "HOLD"
	if len(t.History) >= 14 {
		// Complex Decision Engine with Confluence
		if (rsi < 30 && price < sma) || divergence == "BULLISH" {
			decision = "BUY"
		} else if (rsi > 70 && price > sma) || divergence == "BEARISH" {
			decision = "SELL"
		}
	} else {
		fmt.Println("[Trading] Insufficient history for complex indicators, defaulting to HOLD.")
	}

	fmt.Printf("[Trading] Strategy Decision: %s\n", decision)

	// Persist to memory
	t.Orchestrator.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("trade-%s-%d", t.Symbol, time.Now().Unix()),
		Content:   fmt.Sprintf("Trade Decision for %s: %s at $%.2f (SMA: $%.2f, RSI: %.2f, Div: %s)", t.Symbol, decision, price, sma, rsi, divergence),
		Timestamp: time.Now(),
		Tags:      []string{"trading", t.Symbol, decision},
	})

	if decision != "HOLD" {
		t.Orchestrator.Ledger.Add(orchestrator.Transaction{
			Amount: 0.10, // Simulating execution fee
			Type:   orchestrator.Expense,
			Hustle: "Trading",
			Note:   fmt.Sprintf("%s %s (RSI: %.2f, Div: %s)", decision, t.Symbol, rsi, divergence),
		})

		// Broadcast trade event to mesh
		t.Orchestrator.L1.Add(orchestrator.MemoryEntry{
			ID: fmt.Sprintf("event-%d", time.Now().Unix()),
			Content: fmt.Sprintf("ALERt: Strategy executed %s for %s", decision, t.Symbol),
			Timestamp: time.Now(),
			Tags: []string{"a2a", "event", "trading"},
		})
	}

	return nil
}

func (t *TradingModule) AddToWatchlist(symbol string) {
	fmt.Printf("[Trading] Adding %s to watchlist\n", symbol)
	t.Watchlist = append(t.Watchlist, symbol)
}

func (t *TradingModule) detectDivergence() string {
	if len(t.History) < 5 || len(t.RSIHistory) < 5 {
		return ""
	}

	lastIdx := len(t.History) - 1
	prevIdx := lastIdx - 2

	priceFalling := t.History[lastIdx] < t.History[prevIdx]
	rsiRising := t.RSIHistory[lastIdx] > t.RSIHistory[prevIdx]

	priceRising := t.History[lastIdx] > t.History[prevIdx]
	rsiFalling := t.RSIHistory[lastIdx] < t.RSIHistory[prevIdx]

	if priceFalling && rsiRising && t.RSIHistory[lastIdx] < 40 {
		return "BULLISH"
	}
	if priceRising && rsiFalling && t.RSIHistory[lastIdx] > 60 {
		return "BEARISH"
	}

	return ""
}

func (t *TradingModule) calculateSMA(period int) float64 {
	if len(t.History) == 0 {
		return 0
	}

	count := period
	if len(t.History) < period {
		count = len(t.History)
	}

	sum := 0.0
	for i := len(t.History) - count; i < len(t.History); i++ {
		sum += t.History[i]
	}
	return sum / float64(count)
}

func (t *TradingModule) calculateRSI(period int) float64 {
	if len(t.History) <= period {
		return 50.0 // Neutral default
	}

	var gains, losses float64
	for i := len(t.History) - period; i < len(t.History); i++ {
		change := t.History[i] - t.History[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	if losses == 0 {
		return 100.0
	}

	rs := (gains / float64(period)) / (losses / float64(period))
	return 100.0 - (100.0 / (1.0 + rs))
}
