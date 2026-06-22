package coinbase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/hustle/hustle/trading/exchange"
)

// mockServer creates a test HTTP server that returns predefined responses.
func mockServer(responseMap map[string]mockResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		for pattern, resp := range responseMap {
			if strings.HasSuffix(path, pattern) || path == pattern {
				if resp.statusCode != 0 {
					w.WriteHeader(resp.statusCode)
				}
				_, _ = w.Write([]byte(resp.body))
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
}

type mockResponse struct {
	body       string
	statusCode int
}

func newTestProvider(serverURL string) *Provider {
	return NewProviderWithClient("test-key", "test-secret", serverURL, &http.Client{Timeout: 5 * time.Second})
}

func TestName(t *testing.T) {
	p := NewProvider()
	if name := p.Name(); name != "coinbase" {
		t.Errorf("expected 'coinbase', got %q", name)
	}
}

func TestDryRunDefault(t *testing.T) {
	p := NewProvider()
	if p.DryRun() {
		t.Error("expected DryRun to be false by default")
	}
}

func TestGetTicker(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/products/BTC-USD/ticker": {
			body: `{
				"trade_id": 12345,
				"price": "67500.00",
				"bid": "67499.50",
				"ask": "67500.50",
				"volume": "12345.678",
				"time": "2024-06-15T12:00:00Z"
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	ticker, err := p.GetTicker("BTC-USD")
	if err != nil {
		t.Fatalf("GetTicker failed: %v", err)
	}

	if ticker.Symbol != "BTC-USD" {
		t.Errorf("expected BTC-USD, got %s", ticker.Symbol)
	}
	if ticker.LastPrice != 67500.00 {
		t.Errorf("expected last 67500, got %.2f", ticker.LastPrice)
	}
	if ticker.BidPrice != 67499.50 {
		t.Errorf("expected bid 67499.50, got %.2f", ticker.BidPrice)
	}
	if ticker.AskPrice != 67500.50 {
		t.Errorf("expected ask 67500.50, got %.2f", ticker.AskPrice)
	}
	if ticker.Volume != 12345.678 {
		t.Errorf("expected volume 12345.678, got %f", ticker.Volume)
	}
}

func TestGetOrderBook(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/products/BTC-USD/book": {
			body: `{
				"sequence": "abc123",
				"bids": [
					["67500.00", "1.500", "order1"],
					["67499.50", "2.000", "order2"]
				],
				"asks": [
					["67501.00", "1.200", "order3"],
					["67502.00", "3.000", "order4"]
				]
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	book, err := p.GetOrderBook("BTC-USD", 2)
	if err != nil {
		t.Fatalf("GetOrderBook failed: %v", err)
	}

	if book.Symbol != "BTC-USD" {
		t.Errorf("expected BTC-USD, got %s", book.Symbol)
	}
	if len(book.Bids) != 2 {
		t.Errorf("expected 2 bids, got %d", len(book.Bids))
	}
	if len(book.Asks) != 2 {
		t.Errorf("expected 2 asks, got %d", len(book.Asks))
	}

	if book.Bids[0].Price != 67500.00 {
		t.Errorf("expected bid price 67500, got %.2f", book.Bids[0].Price)
	}
	if book.Asks[0].Price != 67501.00 {
		t.Errorf("expected ask price 67501, got %.2f", book.Asks[0].Price)
	}
}

func TestGetTickerAPIError(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/products/INVALID/ticker": {
			body:       `{"message":"NotFound"}`,
			statusCode: http.StatusNotFound,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	_, err := p.GetTicker("INVALID")
	if err == nil {
		t.Fatal("expected error for invalid symbol, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected error to contain HTTP status, got: %v", err)
	}
}

func TestGetTickerCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte(`{
			"trade_id": 1,
			"price": "100.00",
			"bid": "99.50",
			"ask": "100.50",
			"volume": "500",
			"time": "2024-06-15T12:00:00Z"
		}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL)

	// First call should hit the server
	_, err := p.GetTicker("BTC-USD")
	if err != nil {
		t.Fatalf("first GetTicker failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 server call, got %d", callCount)
	}

	// Second call should use cache
	_, err = p.GetTicker("BTC-USD")
	if err != nil {
		t.Fatalf("second GetTicker failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 server call (cache hit), got %d", callCount)
	}
}

func TestClearCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte(`{
			"trade_id": 1,
			"price": "100.00",
			"bid": "99.50",
			"ask": "100.50",
			"volume": "500",
			"time": "2024-06-15T12:00:00Z"
		}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL)

	// First call
	_, _ = p.GetTicker("BTC-USD")
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Clear cache
	p.ClearCache()

	// Should hit server again
	_, _ = p.GetTicker("BTC-USD")
	if callCount != 2 {
		t.Errorf("expected 2 calls after ClearCache, got %d", callCount)
	}
}

func TestPlaceOrderNotSupported(t *testing.T) {
	p := NewProvider()
	_, err := p.PlaceOrder("BTC-USD", exchange.Buy, exchange.Market, 1.0, 0)
	if err == nil {
		t.Fatal("expected error for PlaceOrder, got nil")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected 'not supported' error, got: %v", err)
	}
}

func TestCancelOrderNotSupported(t *testing.T) {
	p := NewProvider()
	err := p.CancelOrder("BTC-USD", "123")
	if err == nil {
		t.Fatal("expected error for CancelOrder, got nil")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected 'not supported' error, got: %v", err)
	}
}

func TestGetAccountNotSupported(t *testing.T) {
	p := NewProvider()
	_, err := p.GetAccount()
	if err == nil {
		t.Fatal("expected error for GetAccount, got nil")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected 'not supported' error, got: %v", err)
	}
}

func TestSubscribeTickersNotImplemented(t *testing.T) {
	p := NewProvider()
	_, _, err := p.SubscribeTickers("BTC-USD")
	if err == nil {
		t.Fatal("expected error for SubscribeTickers, got nil")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected 'not implemented' error, got: %v", err)
	}
}

func TestIsRetryable(t *testing.T) {
	p := NewProvider()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"429 rate limited", &apiError{StatusCode: 429, Body: "rate limited"}, true},
		{"500 server error", &apiError{StatusCode: 500, Body: "server error"}, true},
		{"503 service unavailable", &apiError{StatusCode: 503, Body: "unavailable"}, true},
		{"400 bad request", &apiError{StatusCode: 400, Body: "bad request"}, false},
		{"401 unauthorized", &apiError{StatusCode: 401, Body: "unauthorized"}, false},
		{"404 not found", &apiError{StatusCode: 404, Body: "not found"}, false},
		{"config error", fmt.Errorf("coinbase: API_KEY not set"), false},
		{"network error", fmt.Errorf("http request: connection refused"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.isRetryable(tt.err)
			if got != tt.expected {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}

func TestCoinbaseTickerJSONRoundTrip(t *testing.T) {
	rawJSON := `{
		"trade_id": 12345,
		"price": "67250.00",
		"bid": "67249.50",
		"ask": "67250.50",
		"volume": "5000.123",
		"time": "2024-06-15T12:00:00Z"
	}`

	var raw struct {
		TradeID int    `json:"trade_id"`
		Price   string `json:"price"`
		Bid     string `json:"bid"`
		Ask     string `json:"ask"`
		Volume  string `json:"volume"`
		Time    string `json:"time"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if raw.Price != "67250.00" {
		t.Errorf("expected price 67250.00, got %s", raw.Price)
	}
	if raw.TradeID != 12345 {
		t.Errorf("expected trade_id 12345, got %d", raw.TradeID)
	}
}
