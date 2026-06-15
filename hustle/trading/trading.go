package trading

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

// Defaults
const (
	defaultCoinGeckoURL = "https://api.coingecko.com/api/v3"
	cacheTTL            = 60 * time.Second // cache prices for 60s
	maxRetries          = 3
	retryBackoff        = 2 * time.Second
)

// PriceFetcher defines the interface for fetching asset prices.
type PriceFetcher interface {
	GetPrice(symbol string) (float64, error)
}

// MockPriceFetcher generates realistic mock prices.
type MockPriceFetcher struct {
	basePrices map[string]float64
	mu         sync.Mutex
}

func (m *MockPriceFetcher) GetPrice(symbol string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.basePrices == nil {
		m.basePrices = map[string]float64{
			"BTC":  67500.0,
			"ETH":  3450.0,
			"SOL":  145.0,
			"DOGE": 0.12,
		}
	}

	base, ok := m.basePrices[strings.ToUpper(symbol)]
	if !ok {
		base = 100.0
	}

	// Add realistic noise (±2%)
	noise := 1.0 + (rand.Float64()-0.5)*0.04
	return base * noise, nil
}

// CoinGeckoFetcher fetches live prices from the CoinGecko API with caching,
// retry logic, and rate-limit awareness.
type CoinGeckoFetcher struct {
	apiKey  string
	baseURL string
	client  *http.Client
	cache   map[string]cachedPrice
	mu      sync.Mutex
}

type cachedPrice struct {
	price     float64
	expiresAt time.Time
}

// NewCoinGeckoFetcher creates a fetcher configured from environment variables.
// Reads COINGECKO_API_KEY (optional, for higher rate limits) and
// COINGECKO_API_URL (optional, defaults to api.coingecko.com).
func NewCoinGeckoFetcher() *CoinGeckoFetcher {
	return &CoinGeckoFetcher{
		apiKey:  os.Getenv("COINGECKO_API_KEY"),
		baseURL: getEnvDefault("COINGECKO_API_URL", defaultCoinGeckoURL),
		client:  &http.Client{Timeout: 15 * time.Second},
		cache:   make(map[string]cachedPrice),
	}
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// symbolToCoinGeckoID maps common tickers to CoinGecko API IDs.
var symbolToCoinGeckoID = map[string]string{
	"BTC":  "bitcoin",
	"ETH":  "ethereum",
	"SOL":  "solana",
	"DOGE": "dogecoin",
	"XRP":  "ripple",
	"ADA":  "cardano",
	"DOT":  "polkadot",
	"AVAX": "avalanche-2",
	"LINK": "chainlink",
	"MATIC": "matic-network",
	"UNI":  "uniswap",
	"ATOM": "cosmos",
	"LTC":  "litecoin",
	"BCH":  "bitcoin-cash",
	"FIL":  "filecoin",
	"APT":  "aptos",
	"SUI":  "sui",
	"ARB":  "arbitrum",
	"OP":   "optimism",
	"NEAR": "near",
}

// GetPrice fetches the current USD price for a symbol.
// Uses an in-memory cache with 60s TTL to avoid hitting rate limits.
// Retries up to 3 times with exponential backoff on transient errors.
func (c *CoinGeckoFetcher) GetPrice(symbol string) (float64, error) {
	symbol = strings.ToUpper(symbol)

	// Check cache first
	if price, ok := c.getFromCache(symbol); ok {
		return price, nil
	}

	// Resolve CoinGecko ID
	id := symbolToCoinGeckoID[symbol]
	if id == "" {
		id = strings.ToLower(symbol)
	}

	// Attempt fetch with retries
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryBackoff * time.Duration(attempt))
		}

		price, err := c.fetchPrice(id)
		if err == nil {
			c.addToCache(symbol, price)
			return price, nil
		}
		lastErr = err

		// Don't retry on "not found" — it's permanent
		if strings.Contains(err.Error(), "not found in response") {
			break
		}
	}

	log.Printf("[CoinGecko] All %d retries exhausted for %s: %v", maxRetries, symbol, lastErr)
	return 0, fmt.Errorf("coingecko: %s after %d retries: %w", symbol, maxRetries, lastErr)
}

func (c *CoinGeckoFetcher) fetchPrice(id string) (float64, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.baseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	// Add API key header if configured (for higher rate limits)
	if c.apiKey != "" {
		req.Header.Set("CG-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		log.Printf("[CoinGecko] Rate limited (429). Consider setting COINGECKO_API_KEY.")
		return 0, fmt.Errorf("rate limited (429)")
	}

	// Check for other non-200 status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	priceMap, ok := result[id]
	if !ok {
		return 0, fmt.Errorf("id %q not found in response", id)
	}

	price, ok := priceMap["usd"]
	if !ok {
		return 0, fmt.Errorf("usd price for %q not found in response", id)
	}

	return price, nil
}

func (c *CoinGeckoFetcher) getFromCache(symbol string) (float64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.cache[symbol]
	if !ok || time.Now().After(entry.expiresAt) {
		return 0, false
	}
	return entry.price, true
}

func (c *CoinGeckoFetcher) addToCache(symbol string, price float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[symbol] = cachedPrice{
		price:     price,
		expiresAt: time.Now().Add(cacheTTL),
	}
}

// ClearCache forces cache refresh on next call.
func (c *CoinGeckoFetcher) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]cachedPrice)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type TradeExecutor interface {
	ExecuteOrder(symbol, side, orderType string, quantity float64) error
}

type MockExecutor struct{}

func (m *MockExecutor) ExecuteOrder(symbol, side, orderType string, quantity float64) error {
	fmt.Printf("[MockExecutor] Simulating %s %s for %f\n", side, symbol, quantity)
	return nil
}

type TradingModule struct {
	Orchestrator *orchestrator.Orchestrator
	Broker       *orchestrator.A2ABroker
	Symbol       string
	Fetcher      PriceFetcher
	Executor     TradeExecutor
	History      []float64
	RSIHistory   []float64
	MACDHistory  []float64
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

	fmt.Printf("[Trading] Current Price for %s: $%.2f\n", t.Symbol, price)

	// Technical Indicator: Simple Moving Average (SMA)
	sma := t.calculateSMA(5)

	// Technical Indicator: RSI
	rsi := t.calculateRSI(14)
	t.RSIHistory = append(t.RSIHistory, rsi)

	fmt.Printf("[Trading] Indicators -> SMA(5): $%.2f | RSI(14): %.2f\n", sma, rsi)

	// Bollinger Bands (20, 2)
	bbUpper, bbLower := t.calculateBollingerBands(20, 2.0)

	// MACD
	macd, signal, hist := t.calculateMACD()
	t.MACDHistory = append(t.MACDHistory, macd)

	fmt.Printf("[Trading] Indicators -> BB: (U:%.2f, L:%.2f) | MACD: %.2f (S:%.2f, H:%.2f)\n", bbUpper, bbLower, macd, signal, hist)

	divergence := t.detectDivergence()
	if divergence != "" {
		fmt.Printf("[Trading] DIVERGENCE DETECTED: %s\n", divergence)
	}

	decision := "HOLD"
	if len(t.History) >= 26 {
		// Complex Decision Engine with Confluence
		if (rsi < 30 && price < sma && price <= bbLower) || divergence == "BULLISH" {
			decision = "BUY"
		} else if (rsi > 70 && price > sma && price >= bbUpper) || divergence == "BEARISH" {
			decision = "SELL"
		}
	} else {
		fmt.Println("[Trading] Insufficient history for complex indicators (need 26 for MACD), defaulting to HOLD.")
	}

	fmt.Printf("[Trading] Strategy Decision: %s\n", decision)

	// Persist to memory
	t.Orchestrator.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("trade-%s-%d", t.Symbol, time.Now().Unix()),
		Content:   fmt.Sprintf("Trade Decision for %s: %s at $%.2f (SMA: $%.2f, RSI: %.2f, BB-L: %.2f, MACD: %.2f)", t.Symbol, decision, price, sma, rsi, bbLower, macd),
		Timestamp: time.Now(),
		Tags:      []string{"trading", t.Symbol, decision},
	})

	if decision != "HOLD" {
		// Execution of real or mock trade
		if t.Executor != nil {
			// Naive quantity logic for autonomous engine (0.001 BTC equivalent)
			qty := 0.001
			if strings.Contains(t.Symbol, "ETH") {
				qty = 0.01
			} else if strings.Contains(t.Symbol, "SOL") {
				qty = 0.1
			}

			if err := t.Executor.ExecuteOrder(t.Symbol, decision, "MARKET", qty); err != nil {
				fmt.Printf("[Trading] ❌ Execution failed: %v\n", err)
			}
		}

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
			losses -= math.Abs(change)
		}
	}

	if losses == 0 {
		return 100.0
	}

	// RSI uses absolute values for average loss
	absLosses := math.Abs(losses)
	rs := (gains / float64(period)) / (absLosses / float64(period))
	return 100.0 - (100.0 / (1.0 + rs))
}

func (t *TradingModule) calculateBollingerBands(period int, stdDevMultiplier float64) (upper, lower float64) {
	if len(t.History) < period {
		return 0, 0
	}

	sma := t.calculateSMA(period)
	variance := 0.0
	for i := len(t.History) - period; i < len(t.History); i++ {
		diff := t.History[i] - sma
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(period))

	upper = sma + (stdDevMultiplier * stdDev)
	lower = sma - (stdDevMultiplier * stdDev)
	return upper, lower
}

func (t *TradingModule) calculateEMA(period int, data []float64) float64 {
	if len(data) < period {
		return 0
	}
	multiplier := 2.0 / (float64(period) + 1.0)
	ema := data[len(data)-period]
	for i := len(data) - period + 1; i < len(data); i++ {
		ema = (data[i]-ema)*multiplier + ema
	}
	return ema
}

func (t *TradingModule) calculateMACD() (macd, signal, histogram float64) {
	if len(t.History) < 26 {
		return 0, 0, 0
	}
	ema12 := t.calculateEMA(12, t.History)
	ema26 := t.calculateEMA(26, t.History)
	macd = ema12 - ema26

	signal = t.calculateEMA(9, t.MACDHistory)
	histogram = macd - signal

	return macd, signal, histogram
}
