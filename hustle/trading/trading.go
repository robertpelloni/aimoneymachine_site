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
	"BTC":   "bitcoin",
	"ETH":   "ethereum",
	"SOL":   "solana",
	"DOGE":  "dogecoin",
	"XRP":   "ripple",
	"ADA":   "cardano",
	"DOT":   "polkadot",
	"AVAX":  "avalanche-2",
	"LINK":  "chainlink",
	"MATIC": "matic-network",
	"UNI":   "uniswap",
	"ATOM":  "cosmos",
	"LTC":   "litecoin",
	"BCH":   "bitcoin-cash",
	"FIL":   "filecoin",
	"APT":   "aptos",
	"SUI":   "sui",
	"ARB":   "arbitrum",
	"OP":    "optimism",
	"NEAR":  "near",
}

// batchSymbols are the symbols the daemon trades. On any cache miss the
// fetcher requests all of them in ONE CoinGecko call and primes the cache
// for the rest, instead of one call per symbol per cycle (7x fewer calls —
// the unauthenticated tier rate-limits at roughly this order of magnitude).
var batchSymbols = []string{"BTC", "ETH", "SOL", "DOGE", "XRP", "ADA", "LINK"}

// idsFor returns the CoinGecko IDs to request when symbol misses the cache:
// the whole batch for daemon symbols, just itself otherwise.
func idsFor(symbol, id string) []string {
	for _, s := range batchSymbols {
		if s == symbol {
			ids := make([]string, 0, len(batchSymbols))
			for _, b := range batchSymbols {
				ids = append(ids, symbolToCoinGeckoID[b])
			}
			return ids
		}
	}
	return []string{id}
}

// GetPrice fetches the current USD price for a symbol.
// Uses an in-memory cache with 60s TTL to avoid hitting rate limits.
// Retries up to 3 times with exponential backoff on transient errors.
// When every retry fails (e.g. a 429 storm) and any earlier price is known,
// it returns that stale price instead of failing the trading cycle: a
// slightly old price yields a HOLD-grade decision, a hard error yields a
// failed daemon task and log spam.
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

		prices, err := c.fetchPrices(idsFor(symbol, id))
		if err == nil {
			// Cache everything returned (keyed by CoinGecko ID); the
			// requested symbol may still be absent, which is the
			// permanent "not found" case.
			for pid, p := range prices {
				if sym, ok := idToSymbol[pid]; ok {
					c.addToCache(sym, p)
				}
			}
			if price, ok := prices[id]; ok {
				return price, nil
			}
			lastErr = fmt.Errorf("id %q not found in response", id)
			break
		}
		lastErr = err

		// Don't retry on "not found" — it's permanent
		if strings.Contains(err.Error(), "not found in response") {
			break
		}
	}

	// Graceful degradation: serve the last known price, however old.
	if stale, ok := c.getStale(symbol); ok {
		log.Printf("[CoinGecko] Serving stale price for %s ($%.2f); fetch failed: %v", symbol, stale, lastErr)
		return stale, nil
	}

	log.Printf("[CoinGecko] All %d retries exhausted for %s: %v", maxRetries, symbol, lastErr)
	return 0, fmt.Errorf("coingecko: %s after %d retries: %w", symbol, maxRetries, lastErr)
}

// idToSymbol reverses symbolToCoinGeckoID for cache keying.
var idToSymbol = func() map[string]string {
	m := make(map[string]string, len(symbolToCoinGeckoID))
	for sym, id := range symbolToCoinGeckoID {
		m[id] = sym
	}
	return m
}()

// fetchPrice fetches a single ID (kept for compatibility).
func (c *CoinGeckoFetcher) fetchPrice(id string) (float64, error) {
	prices, err := c.fetchPrices([]string{id})
	if err != nil {
		return 0, err
	}
	price, ok := prices[id]
	if !ok {
		return 0, fmt.Errorf("id %q not found in response", id)
	}
	return price, nil
}

// fetchPrices fetches several IDs in one /simple/price call and returns
// USD prices keyed by CoinGecko ID. IDs absent from the response are
// simply missing from the map (callers decide if that is an error).
func (c *CoinGeckoFetcher) fetchPrices(ids []string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.baseURL, strings.Join(ids, ","))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Add API key header if configured (for higher rate limits)
	if c.apiKey != "" {
		req.Header.Set("CG-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		log.Printf("[CoinGecko] Rate limited (429). Consider setting COINGECKO_API_KEY.")
		return nil, fmt.Errorf("rate limited (429)")
	}

	// Check for other non-200 status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Preserve the historical error distinction: a requested ID that is
	// present in the response but has no USD price is a distinct failure
	// from an absent ID (which callers treat as permanent "not found").
	for _, id := range ids {
		if priceMap, ok := result[id]; ok {
			if _, has := priceMap["usd"]; !has {
				return nil, fmt.Errorf("usd price for %q not found in response", id)
			}
		}
	}

	prices := make(map[string]float64, len(result))
	for pid, priceMap := range result {
		if price, ok := priceMap["usd"]; ok {
			prices[pid] = price
		}
	}
	return prices, nil
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

// getStale returns the last known price even after expiry (0, false if none).
func (c *CoinGeckoFetcher) getStale(symbol string) (float64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.cache[symbol]
	if !ok {
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

type TradingModule struct {
	Orchestrator *orchestrator.Orchestrator
	Broker       *orchestrator.A2ABroker
	Symbol       string
	Fetcher      PriceFetcher
	// History/RSIHistory mirror the most recently executed symbol's series
	// (kept for compatibility). The authoritative series are per-symbol in
	// histories/rsiHistories: the daemon reuses ONE module for every symbol,
	// so a single shared series mixed BTC/ETH/ADA prices and made SMA/RSI
	// meaningless (prod showed ADA SMA $0.58 vs its $0.16 price).
	History      []float64
	RSIHistory   []float64
	histories    map[string][]float64
	rsiHistories map[string][]float64
	Watchlist    []string
	mu           sync.Mutex
}

// HistoryFor returns a copy of the price series for one symbol.
func (t *TradingModule) HistoryFor(symbol string) []float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	src := t.histories[strings.ToUpper(symbol)]
	out := make([]float64, len(src))
	copy(out, src)
	return out
}

// ExecuteFor runs one strategy cycle for symbol, atomically: the shared
// module's Symbol field is only mutated under the module lock, so
// concurrent per-symbol daemon tasks cannot cross-contaminate each other.
func (t *TradingModule) ExecuteFor(symbol string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Symbol = strings.ToUpper(symbol)
	return t.executeLocked()
}

func (t *TradingModule) ExecuteStrategy() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.executeLocked()
}

func (t *TradingModule) executeLocked() error {
	symbol := strings.ToUpper(t.Symbol)
	if t.histories == nil {
		t.histories = make(map[string][]float64)
		t.rsiHistories = make(map[string][]float64)
	}

	fmt.Printf("[Trading] Executing strategy for Symbol: %s\n", symbol)

	price, err := t.Fetcher.GetPrice(symbol)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %v", err)
	}
	history := append(t.histories[symbol], price)
	t.histories[symbol] = history
	t.History = history

	fmt.Printf("[Trading] Current Price for %s: $%.2f\n", symbol, price)

	// Technical Indicators, computed on this symbol's series only
	sma := calculateSMA(history, 5)
	rsi := calculateRSI(history, 14)
	rsiHistory := append(t.rsiHistories[symbol], rsi)
	t.rsiHistories[symbol] = rsiHistory
	t.RSIHistory = rsiHistory

	macdLine, macdSignal, macdHist, prevHist := calculateMACD(history)
	bbUpper, bbMid, bbLower, bbPctB := calculateBollinger(history, 20, 2)

	fmt.Printf("[Trading] Indicators -> SMA(5): $%.2f | RSI(14): %.2f | MACD: %.4f (signal %.4f, hist %.4f) | BB: [%.2f | %.2f | %.2f] %%B %.2f\n",
		sma, rsi, macdLine, macdSignal, macdHist, bbLower, bbMid, bbUpper, bbPctB)

	divergence := detectDivergence(history, rsiHistory)
	if divergence != "" {
		fmt.Printf("[Trading] DIVERGENCE DETECTED: %s\n", divergence)
	}

	decision := "HOLD"
	if len(history) >= 14 {
		// Complex Decision Engine with Confluence
		if (rsi < 30 && price < sma) || divergence == "BULLISH" {
			decision = "BUY"
		} else if (rsi > 70 && price > sma) || divergence == "BEARISH" {
			decision = "SELL"
		}
	}
	// MACD + Bollinger confluence (needs the longer MACD warmup).
	if len(history) >= 35 && decision == "HOLD" {
		macdBullCross := prevHist <= 0 && macdHist > 0
		macdBearCross := prevHist >= 0 && macdHist < 0
		switch {
		case macdBullCross && bbPctB < 0.8:
			decision = "BUY"
			fmt.Println("[Trading] MACD bullish cross confirmed by Bollinger %B < 0.8")
		case macdBearCross && bbPctB > 0.2:
			decision = "SELL"
			fmt.Println("[Trading] MACD bearish cross confirmed by Bollinger %B > 0.2")
		case price < bbLower && rsi < 35:
			decision = "BUY"
			fmt.Println("[Trading] Price below lower Bollinger band with RSI < 35 (mean reversion)")
		case price > bbUpper && rsi > 65:
			decision = "SELL"
			fmt.Println("[Trading] Price above upper Bollinger band with RSI > 65 (mean reversion)")
		}
	}
	if len(history) < 14 {
		fmt.Println("[Trading] Insufficient history for complex indicators, defaulting to HOLD.")
	}

	fmt.Printf("[Trading] Strategy Decision: %s\n", decision)

	// Persist to memory
	t.Orchestrator.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("trade-%s-%d", symbol, time.Now().Unix()),
		Content:   fmt.Sprintf("Trade Decision for %s: %s at $%.2f (SMA: $%.2f, RSI: %.2f, MACD hist: %.4f, BB %%B: %.2f, Div: %s)", symbol, decision, price, sma, rsi, macdHist, bbPctB, divergence),
		Timestamp: time.Now(),
		Tags:      []string{"trading", symbol, decision},
	})

	if decision != "HOLD" {
		t.Orchestrator.Ledger.Add(orchestrator.Transaction{
			Amount: 0.10, // Simulating execution fee
			Type:   orchestrator.Expense,
			Hustle: "Trading",
			Note:   fmt.Sprintf("%s %s (RSI: %.2f, Div: %s)", decision, symbol, rsi, divergence),
		})

		// Broadcast trade event to mesh
		if t.Broker != nil {
			msg := orchestrator.Message{
				ID:        fmt.Sprintf("trade-evt-%d", time.Now().Unix()),
				Source:    "trading-module",
				Type:      orchestrator.Event,
				Topic:     "trade_execution",
				Payload:   fmt.Sprintf("Symbol: %s, Action: %s, Price: %.2f", symbol, decision, price),
				Timestamp: time.Now(),
			}
			t.Broker.Publish(msg)
			fmt.Printf("[Trading] Broadcasted %s decision to mesh topic: trade_execution\n", decision)
		}

		t.Orchestrator.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("event-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("ALERT: Strategy executed %s for %s", decision, symbol),
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

func detectDivergence(history, rsiHistory []float64) string {
	if len(history) < 5 || len(rsiHistory) < 5 {
		return ""
	}

	lastIdx := len(history) - 1
	prevIdx := lastIdx - 2

	priceFalling := history[lastIdx] < history[prevIdx]
	rsiRising := rsiHistory[lastIdx] > rsiHistory[prevIdx]

	priceRising := history[lastIdx] > history[prevIdx]
	rsiFalling := rsiHistory[lastIdx] < rsiHistory[prevIdx]

	if priceFalling && rsiRising && rsiHistory[lastIdx] < 40 {
		return "BULLISH"
	}
	if priceRising && rsiFalling && rsiHistory[lastIdx] > 60 {
		return "BEARISH"
	}

	return ""
}

func calculateSMA(history []float64, period int) float64 {
	if len(history) == 0 {
		return 0
	}

	count := period
	if len(history) < period {
		count = len(history)
	}

	sum := 0.0
	for i := len(history) - count; i < len(history); i++ {
		sum += history[i]
	}
	return sum / float64(count)
}

func calculateRSI(history []float64, period int) float64 {
	if len(history) <= period {
		return 50.0 // Neutral default
	}

	var gains, losses float64
	for i := len(history) - period; i < len(history); i++ {
		change := history[i] - history[i-1]
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

// emaSeries computes the exponential moving average series, seeded with
// the SMA of the first period values (standard MACD seeding).
func emaSeries(values []float64, period int) []float64 {
	if period <= 0 || len(values) < period {
		return nil
	}
	k := 2.0 / float64(period+1)
	out := make([]float64, 0, len(values)-period+1)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	ema := sum / float64(period)
	out = append(out, ema)
	for i := period; i < len(values); i++ {
		ema = values[i]*k + ema*(1-k)
		out = append(out, ema)
	}
	return out
}

// calculateMACD returns MACD line, signal line, histogram, and the
// previous histogram value (for cross detection) using 12/26/9 EMAs.
// All zeros when the series is too short (< 34 points).
func calculateMACD(history []float64) (macd, signal, hist, prevHist float64) {
	if len(history) < 34 {
		return 0, 0, 0, 0
	}
	fast := emaSeries(history, 12)
	slow := emaSeries(history, 26)
	// Align: slow starts 14 points later than fast.
	macdSeries := make([]float64, len(slow))
	for i := range slow {
		macdSeries[i] = fast[i+len(fast)-len(slow)] - slow[i]
	}
	signalSeries := emaSeries(macdSeries, 9)
	if len(signalSeries) == 0 {
		return 0, 0, 0, 0
	}
	macd = macdSeries[len(macdSeries)-1]
	signal = signalSeries[len(signalSeries)-1]
	hist = macd - signal
	if len(signalSeries) > 1 {
		prevHist = macdSeries[len(macdSeries)-2] - signalSeries[len(signalSeries)-2]
	}
	return macd, signal, hist, prevHist
}

// calculateBollinger returns upper band, middle (SMA), lower band, and
// %B ((price-lower)/(upper-lower); 0.5 at the mean, 0.5 when flat).
func calculateBollinger(history []float64, period int, mult float64) (upper, mid, lower, pctB float64) {
	if period <= 0 || len(history) < period {
		if len(history) > 0 {
			p := history[len(history)-1]
			return p, p, p, 0.5
		}
		return 0, 0, 0, 0.5
	}
	window := history[len(history)-period:]
	mid = calculateSMA(window, period)
	var variance float64
	for _, v := range window {
		d := v - mid
		variance += d * d
	}
	sd := math.Sqrt(variance / float64(period))
	upper = mid + mult*sd
	lower = mid - mult*sd
	price := history[len(history)-1]
	if upper == lower {
		pctB = 0.5
	} else {
		pctB = (price - lower) / (upper - lower)
	}
	return upper, mid, lower, pctB
}
