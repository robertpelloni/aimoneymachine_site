package trading

import (
	"encoding/json"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"math/rand"
	"net/http"
	"strings"
	"sync"
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

type CoinGeckoFetcher struct {
	APIKey string
}

func (c *CoinGeckoFetcher) GetPrice(symbol string) (float64, error) {
	// Map common symbols to CoinGecko IDs
	ids := map[string]string{
		"BTC":  "bitcoin",
		"ETH":  "ethereum",
		"SOL":  "solana",
		"DOGE": "dogecoin",
	}

	id, ok := ids[strings.ToUpper(symbol)]
	if !ok {
		id = strings.ToLower(symbol)
	}

	baseURL := "https://api.coingecko.com/api/v3"
	if c.APIKey != "" {
		baseURL = "https://pro-api.coingecko.com/api/v3"
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", baseURL, id)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	if c.APIKey != "" {
		req.Header.Set("x-cg-pro-api-key", c.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, fmt.Errorf("CoinGecko rate limit exceeded (HTTP 429)")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("CoinGecko API returned status %d", resp.StatusCode)
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	price, ok := result[id]["usd"]
	if !ok {
		return 0, fmt.Errorf("price for %s not found in response", symbol)
	}

	return price, nil
}

type TradingModule struct {
	Orchestrator *orchestrator.Orchestrator
	Broker       *orchestrator.A2ABroker
	Symbol       string
	Fetcher      PriceFetcher
	History      []float64
	RSIHistory   []float64
	Watchlist    []string
	mu           sync.Mutex
}

func (t *TradingModule) ExecuteStrategy() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	fmt.Printf("[Trading] Executing strategy for Symbol: %s\n", t.Symbol)

	price, err := t.Fetcher.GetPrice(t.Symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %v", err)
	}
	t.History = append(t.History, price)

	source := "MOCK"
	if f, ok := t.Fetcher.(*CoinGeckoFetcher); ok {
		source = "COINGECKO_PUBLIC"
		if f.APIKey != "" {
			source = "COINGECKO_PRO"
		}
	}
	fmt.Printf("[Trading] Current Price for %s: $%.2f (Source: %s)\n", t.Symbol, price, source)

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
		Content:   fmt.Sprintf("Trade Decision for %s: %s at $%.2f (Source: %s) (SMA: $%.2f, RSI: %.2f, Div: %s)", t.Symbol, decision, price, source, sma, rsi, divergence),
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
		if t.Broker != nil {
			msg := orchestrator.Message{
				ID:        fmt.Sprintf("trade-evt-%d", time.Now().Unix()),
				Source:    "trading-module",
				Type:      orchestrator.Event,
				Topic:     "trade_execution",
				Payload:   fmt.Sprintf("Symbol: %s, Action: %s, Price: %.2f", t.Symbol, decision, price),
				Timestamp: time.Now(),
			}
			t.Broker.Publish(msg)
			fmt.Printf("[Trading] Broadcasted %s decision to mesh topic: trade_execution\n", decision)
		}

		t.Orchestrator.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("event-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("ALERT: Strategy executed %s for %s", decision, t.Symbol),
			Timestamp: time.Now(),
			Tags:      []string{"a2a", "event", "trading"},
		})
	}

	return nil
}

func (t *TradingModule) AddToWatchlist(symbol string) {
	t.mu.Lock()
	defer t.mu.Unlock()
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
