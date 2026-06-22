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
		ids := r.URL.Query().Get("ids")
		switch ids {
		case "ethereum":
			fmt.Fprint(w, `{"ethereum":{"usd":3450.50}}`)
		case "solana":
			fmt.Fprint(w, `{"solana":{"usd":145.30}}`)
		case "dogecoin":
			fmt.Fprint(w, `{"dogecoin":{"usd":0.125}}`)
		default:
			fmt.Fprint(w, `{}`)
		}
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
