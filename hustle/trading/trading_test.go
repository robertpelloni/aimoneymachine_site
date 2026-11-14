package trading

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

// ── Existing: Strategy + MockPriceFetcher ──

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

// The daemon shares ONE TradingModule across symbols. Alternating symbols
// must build independent series: BTC prices around 100 must never leak into
// ETH's SMA/RSI (which trade around 10).
type fixedFetcher struct{ prices map[string]float64 }

func (f fixedFetcher) GetPrice(symbol string) (float64, error) {
	if p, ok := f.prices[strings.ToUpper(symbol)]; ok {
		return p, nil
	}
	return 0, fmt.Errorf("unknown symbol %s", symbol)
}

func TestTradingPerSymbolHistory(t *testing.T) {
	trader := &TradingModule{
		Orchestrator: orchestrator.NewOrchestrator(),
		Fetcher:      fixedFetcher{prices: map[string]float64{"BTC": 100, "ETH": 10}},
	}
	for i := 0; i < 10; i++ {
		if err := trader.ExecuteFor("BTC"); err != nil {
			t.Fatalf("BTC cycle %d: %v", i, err)
		}
		if err := trader.ExecuteFor("ETH"); err != nil {
			t.Fatalf("ETH cycle %d: %v", i, err)
		}
	}
	btc, eth := trader.HistoryFor("BTC"), trader.HistoryFor("ETH")
	if len(btc) != 10 || len(eth) != 10 {
		t.Fatalf("history lengths BTC=%d ETH=%d, want 10/10", len(btc), len(eth))
	}
	for _, p := range btc {
		if p != 100 {
			t.Fatalf("BTC series contaminated: got %v", btc)
		}
	}
	for _, p := range eth {
		if p != 10 {
			t.Fatalf("ETH series contaminated: got %v", eth)
		}
	}
	// SMA over each series must equal that symbol's own price level.
	if got := calculateSMA(btc, 5); got != 100 {
		t.Errorf("BTC SMA = %.2f, want 100", got)
	}
	if got := calculateSMA(eth, 5); got != 10 {
		t.Errorf("ETH SMA = %.2f, want 10", got)
	}
}

func TestMockPriceFetcher_DifferentSymbols(t *testing.T) {
	m := &MockPriceFetcher{}
	tests := []struct {
		symbol string
		min    float64
		max    float64
	}{
		{"BTC", 65000, 70000},
		{"ETH", 3000, 4000},
		{"SOL", 100, 200},
		{"DOGE", 0.1, 0.15},
		{"UNKNOWN", 95, 105}, // unknown symbol defaults to 100
	}

	for _, tt := range tests {
		price, err := m.GetPrice(tt.symbol)
		if err != nil {
			t.Errorf("GetPrice(%q) unexpected error: %v", tt.symbol, err)
		}
		if price < tt.min || price > tt.max {
			t.Errorf("GetPrice(%q) = %.2f, want between %.2f and %.2f", tt.symbol, price, tt.min, tt.max)
		}
	}
}

func TestMockPriceFetcher_ThreadSafety(t *testing.T) {
	m := &MockPriceFetcher{}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := m.GetPrice("BTC")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
}

// ── CoinGeckoFetcher Tests (mock HTTP server) ──

// newTestFetcher creates a CoinGeckoFetcher pointing at a test server.
func newTestFetcher(serverURL string) *CoinGeckoFetcher {
	return &CoinGeckoFetcher{
		apiKey:  "test-key",
		baseURL: serverURL,
		client:  &http.Client{Timeout: 5 * time.Second},
		cache:   make(map[string]cachedPrice),
	}
}

func TestCoinGeckoFetcher_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key header is passed
		if r.Header.Get("CG-API-Key") != "test-key" {
			t.Error("API key header not sent")
		}
		// Verify correct URL path
		if !strings.Contains(r.URL.String(), "bitcoin") {
			t.Errorf("Expected bitcoin in URL, got %s", r.URL.String())
		}
		fmt.Fprint(w, `{"bitcoin":{"usd":67500.42}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	price, err := f.GetPrice("BTC")
	if err != nil {
		t.Fatalf("GetPrice failed: %v", err)
	}
	if price != 67500.42 {
		t.Errorf("Expected 67500.42, got %.2f", price)
	}
}

func TestCoinGeckoFetcher_Cache(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprint(w, `{"bitcoin":{"usd":67500.42}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)

	// First call should hit API
	price1, err := f.GetPrice("BTC")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}

	// Second call should use cache
	price2, err := f.GetPrice("BTC")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected cache hit (still 1 API call), got %d", callCount)
	}
	if price1 != price2 {
		t.Errorf("Cached price %.2f != original %.2f", price2, price1)
	}
}

func TestCoinGeckoFetcher_CacheExpiry(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprint(w, `{"bitcoin":{"usd":67500.42}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)

	// First call — API hit
	_, _ = f.GetPrice("BTC")
	if callCount != 1 {
		t.Fatalf("Expected 1 API call, got %d", callCount)
	}

	// Manually expire the cache
	f.ClearCache()

	// Call again — should hit API
	_, _ = f.GetPrice("BTC")
	if callCount != 2 {
		t.Errorf("Expected 2 API calls after cache clear, got %d", callCount)
	}
}

func TestCoinGeckoFetcher_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"error":"rate limited"}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	_, err := f.GetPrice("BTC")
	if err == nil {
		t.Fatal("Expected error for rate limit, got nil")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("Expected rate limit error, got: %v", err)
	}
}

func TestCoinGeckoFetcher_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":"internal error"}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	_, err := f.GetPrice("BTC")
	if err == nil {
		t.Fatal("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected 500 error, got: %v", err)
	}
}

func TestCoinGeckoFetcher_UnknownSymbol(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	_, err := f.GetPrice("NONEXISTENT")
	if err == nil {
		t.Fatal("Expected error for unknown symbol, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestCoinGeckoFetcher_MissingUSD(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"bitcoin":{}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	_, err := f.GetPrice("BTC")
	if err == nil {
		t.Fatal("Expected error for missing USD price, got nil")
	}
	if !strings.Contains(err.Error(), "usd price") {
		t.Errorf("Expected 'usd price' error, got: %v", err)
	}
}

func TestCoinGeckoFetcher_DifferentSymbols(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The fetcher batches daemon symbols into one ids= call (v1.3.1),
		// so answer like the real API: one entry per requested ID.
		prices := map[string]float64{"ethereum": 3450.50, "solana": 145.30, "dogecoin": 0.125}
		parts := []string{}
		for _, id := range strings.Split(r.URL.Query().Get("ids"), ",") {
			if p, ok := prices[id]; ok {
				parts = append(parts, fmt.Sprintf(`"%s":{"usd":%v}`, id, p))
			}
		}
		fmt.Fprintf(w, "{%s}", strings.Join(parts, ","))
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)

	tests := []struct {
		symbol string
		want   float64
	}{
		{"ETH", 3450.50},
		{"SOL", 145.30},
		{"DOGE", 0.125},
	}

	for _, tt := range tests {
		price, err := f.GetPrice(tt.symbol)
		if err != nil {
			t.Errorf("GetPrice(%q) unexpected error: %v", tt.symbol, err)
			continue
		}
		if price != tt.want {
			t.Errorf("GetPrice(%q) = %.2f, want %.2f", tt.symbol, price, tt.want)
		}
	}
}

func TestCoinGeckoFetcher_ThreadSafety(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"bitcoin":{"usd":67500.42}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.GetPrice("BTC")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestCoinGeckoFetcher_RetryThenSucceed(t *testing.T) {
	attempt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"error":"temporary"}`)
			return
		}
		fmt.Fprint(w, `{"bitcoin":{"usd":67500.42}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	price, err := f.GetPrice("BTC")
	if err != nil {
		t.Fatalf("GetPrice failed after retry: %v", err)
	}
	if attempt != 3 {
		t.Errorf("Expected 3 attempts (2 fail + 1 succeed), got %d", attempt)
	}
	if price != 67500.42 {
		t.Errorf("Expected 67500.42, got %.2f", price)
	}
}

func TestCoinGeckoFetcher_AllRetriesFail(t *testing.T) {
	attempt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"error":"temporary"}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	_, err := f.GetPrice("BTC")
	if err == nil {
		t.Fatal("Expected error after all retries exhausted, got nil")
	}
	if attempt != maxRetries {
		t.Errorf("Expected %d retries, got %d", maxRetries, attempt)
	}
}

// One cache miss for a daemon symbol must batch-fetch all of them: the
// second symbol is then served from cache with no extra HTTP call.
func TestCoinGeckoFetcher_BatchPrimesCache(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if !strings.Contains(r.URL.String(), "ethereum") {
			t.Errorf("batch request should include ethereum, got %s", r.URL.String())
		}
		fmt.Fprint(w, `{"bitcoin":{"usd":100},"ethereum":{"usd":200},"solana":{"usd":3},"dogecoin":{"usd":0.1},"ripple":{"usd":0.5},"cardano":{"usd":0.4},"chainlink":{"usd":15}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	if p, err := f.GetPrice("BTC"); err != nil || p != 100 {
		t.Fatalf("BTC = %.2f, %v; want 100, nil", p, err)
	}
	if p, err := f.GetPrice("ETH"); err != nil || p != 200 {
		t.Fatalf("ETH = %.2f, %v; want 200, nil (cache)", p, err)
	}
	if callCount != 1 {
		t.Errorf("API calls = %d, want 1 (batched)", callCount)
	}
}

// With a known-but-expired price and the API rate-limiting, GetPrice must
// degrade to the stale price instead of failing the trading cycle. With no
// known price it must still error (covered by TestCoinGeckoFetcher_RateLimit).
func TestCoinGeckoFetcher_StaleFallbackOnRateLimit(t *testing.T) {
	limited := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limited {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":"rate limited"}`)
			return
		}
		fmt.Fprint(w, `{"bitcoin":{"usd":67500.42},"ethereum":{"usd":1},"solana":{"usd":1},"dogecoin":{"usd":1},"ripple":{"usd":1},"cardano":{"usd":1},"chainlink":{"usd":1}}`)
	}))
	defer ts.Close()

	f := newTestFetcher(ts.URL)
	if _, err := f.GetPrice("BTC"); err != nil {
		t.Fatalf("priming GetPrice failed: %v", err)
	}
	// Expire the entry, then rate-limit: stale price must be served.
	f.mu.Lock()
	e := f.cache["BTC"]
	e.expiresAt = e.expiresAt.Add(-2 * cacheTTL)
	f.cache["BTC"] = e
	f.mu.Unlock()
	limited = true

	price, err := f.GetPrice("BTC")
	if err != nil {
		t.Fatalf("stale fallback: err = %v, want nil", err)
	}
	if price != 67500.42 {
		t.Errorf("stale price = %.2f, want 67500.42", price)
	}
}
