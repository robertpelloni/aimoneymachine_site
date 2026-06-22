package binance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/hustle/hustle/trading/exchange"
)

// ── Helpers ──

// mockServer creates a test HTTP server that returns predefined responses.
func mockServer(responseMap map[string]mockResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Find matching response by path (ignore query params for matching)
		path := r.URL.Path
		// Binance API uses /api/v3/...
		if !strings.HasPrefix(path, "/api/v3") {
			path = "/api/v3" + path
		}

		for pattern, resp := range responseMap {
			if strings.HasSuffix(path, pattern) || path == pattern {
				if resp.statusCode != 0 {
					w.WriteHeader(resp.statusCode)
				}
				_, _ = w.Write([]byte(resp.body))
				return
			}
		}

		// Fallback: return 404
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":-1121,"msg":"Invalid symbol."}`))
	}))
}

type mockResponse struct {
	body       string
	statusCode int
}

func newTestProvider(serverURL string) *Provider {
	return NewProviderWithClient("test-api-key", "test-secret-key", serverURL, &http.Client{Timeout: 5 * time.Second})
}

// ── Tests ──

func TestName(t *testing.T) {
	p := NewProvider()
	if name := p.Name(); name != "binance" {
		t.Errorf("expected 'binance', got %q", name)
	}
}

func TestDryRunDefault(t *testing.T) {
	p := NewProvider()
	if p.DryRun() {
		t.Error("expected DryRun to be false by default")
	}

	p.SetDryRun(true)
	if !p.DryRun() {
		t.Error("expected DryRun to be true after setting")
	}
}

func TestGetTicker(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/ticker/24hr": {
			body: `{
				"symbol": "BTCUSDT",
				"bidPrice": "67500.00",
				"askPrice": "67501.00",
				"lastPrice": "67500.50",
				"volume": "12345.678",
				"highPrice": "68000.00",
				"lowPrice": "67000.00",
				"priceChange": "150.25"
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	ticker, err := p.GetTicker("BTCUSDT")
	if err != nil {
		t.Fatalf("GetTicker failed: %v", err)
	}

	if ticker.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", ticker.Symbol)
	}
	if ticker.BidPrice != 67500.00 {
		t.Errorf("expected bid 67500, got %.2f", ticker.BidPrice)
	}
	if ticker.AskPrice != 67501.00 {
		t.Errorf("expected ask 67501, got %.2f", ticker.AskPrice)
	}
	if ticker.LastPrice != 67500.50 {
		t.Errorf("expected last 67500.50, got %.2f", ticker.LastPrice)
	}
	if ticker.Volume != 12345.678 {
		t.Errorf("expected volume 12345.678, got %f", ticker.Volume)
	}
	if ticker.HighPrice != 68000.00 {
		t.Errorf("expected high 68000, got %.2f", ticker.HighPrice)
	}
	if ticker.LowPrice != 67000.00 {
		t.Errorf("expected low 67000, got %.2f", ticker.LowPrice)
	}
	if ticker.PriceChange != 150.25 {
		t.Errorf("expected change 150.25, got %.2f", ticker.PriceChange)
	}
}

func TestGetOrderBook(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/depth": {
			body: `{
				"lastUpdateId": 123456,
				"bids": [
					["67500.00", "1.500"],
					["67499.50", "2.000"],
					["67499.00", "0.750"]
				],
				"asks": [
					["67501.00", "1.200"],
					["67502.00", "3.000"],
					["67503.00", "0.500"]
				]
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	book, err := p.GetOrderBook("BTCUSDT", 3)
	if err != nil {
		t.Fatalf("GetOrderBook failed: %v", err)
	}

	if book.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", book.Symbol)
	}
	if len(book.Bids) != 3 {
		t.Errorf("expected 3 bids, got %d", len(book.Bids))
	}
	if len(book.Asks) != 3 {
		t.Errorf("expected 3 asks, got %d", len(book.Asks))
	}

	// Check first bid
	if book.Bids[0].Price != 67500.00 {
		t.Errorf("expected bid price 67500, got %.2f", book.Bids[0].Price)
	}
	if book.Bids[0].Quantity != 1.500 {
		t.Errorf("expected bid qty 1.500, got %.3f", book.Bids[0].Quantity)
	}

	// Check first ask
	if book.Asks[0].Price != 67501.00 {
		t.Errorf("expected ask price 67501, got %.2f", book.Asks[0].Price)
	}
	if book.Asks[0].Quantity != 1.200 {
		t.Errorf("expected ask qty 1.200, got %.3f", book.Asks[0].Quantity)
	}
}

func TestPlaceOrder(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/order": {
			body: `{
				"symbol": "BTCUSDT",
				"orderId": 123456789,
				"clientOrderId": "abc123",
				"transactTime": 1718380800000,
				"price": "67500.00",
				"origQty": "0.01000000",
				"executedQty": "0.01000000",
				"cummulativeQuoteQty": "675.00000000",
				"status": "FILLED",
				"type": "MARKET",
				"side": "BUY",
				"updateTime": 1718380800000
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	p.SetDryRun(false)

	order, err := p.PlaceOrder("BTCUSDT", exchange.Buy, exchange.Market, 0.01, 0)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if order.ID != "123456789" {
		t.Errorf("expected order ID 123456789, got %s", order.ID)
	}
	if order.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", order.Symbol)
	}
	if order.Side != exchange.Buy {
		t.Errorf("expected BUY, got %s", order.Side)
	}
	if order.Status != exchange.OrderFilled {
		t.Errorf("expected FILLED, got %s", order.Status)
	}
	if order.Quantity != 0.01 {
		t.Errorf("expected qty 0.01, got %f", order.Quantity)
	}
	if order.Price != 67500.00 {
		t.Errorf("expected price 67500, got %.2f", order.Price)
	}
}

func TestPlaceOrderDryRun(t *testing.T) {
	// In dry-run mode, no HTTP request should be made.
	// If it does, the server will return 404 and the test should fail.
	server := mockServer(map[string]mockResponse{})
	defer server.Close()

	p := newTestProvider(server.URL)
	p.SetDryRun(true)

	order, err := p.PlaceOrder("BTCUSDT", exchange.Sell, exchange.Limit, 1.0, 68000.00)
	if err != nil {
		t.Fatalf("PlaceOrder dry-run failed: %v", err)
	}

	if order.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", order.Symbol)
	}
	if order.Side != exchange.Sell {
		t.Errorf("expected SELL, got %s", order.Side)
	}
	if order.Quantity != 1.0 {
		t.Errorf("expected qty 1.0, got %f", order.Quantity)
	}
	// Dry-run orders should be marked FILLED
	if order.Status != exchange.OrderFilled {
		t.Errorf("expected FILLED in dry-run, got %s", order.Status)
	}
}

func TestGetAccount(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/account": {
			body: `{
				"canTrade": true,
				"canWithdraw": true,
				"canDeposit": true,
				"balances": [
					{"asset": "BTC", "free": "0.50000000", "locked": "0.10000000"},
					{"asset": "ETH", "free": "10.00000000", "locked": "0.00000000"},
					{"asset": "USDT", "free": "50000.00000000", "locked": "1000.00000000"},
					{"asset": "SOL", "free": "0.00000000", "locked": "0.00000000"}
				]
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	acc, err := p.GetAccount()
	if err != nil {
		t.Fatalf("GetAccount failed: %v", err)
	}

	if !acc.CanTrade {
		t.Error("expected CanTrade true")
	}
	if !acc.CanWithdraw {
		t.Error("expected CanWithdraw true")
	}
	if !acc.CanDeposit {
		t.Error("expected CanDeposit true")
	}

	// Should be 3 (SOL is zero and should be filtered out)
	if len(acc.Balances) != 3 {
		t.Errorf("expected 3 balances, got %d: %+v", len(acc.Balances), acc.Balances)
	}

	// Check BTC
	foundBTC := false
	for _, b := range acc.Balances {
		if b.Asset == "BTC" {
			foundBTC = true
			if b.Free != 0.5 {
				t.Errorf("expected BTC free 0.5, got %f", b.Free)
			}
			if b.Locked != 0.1 {
				t.Errorf("expected BTC locked 0.1, got %f", b.Locked)
			}
			if b.Total != 0.6 {
				t.Errorf("expected BTC total 0.6, got %f", b.Total)
			}
		}
	}
	if !foundBTC {
		t.Error("BTC balance not found")
	}

	// Check USDT
	foundUSDT := false
	for _, b := range acc.Balances {
		if b.Asset == "USDT" {
			foundUSDT = true
			if b.Free != 50000.0 {
				t.Errorf("expected USDT free 50000, got %f", b.Free)
			}
			if b.Total != 51000.0 {
				t.Errorf("expected USDT total 51000, got %f", b.Total)
			}
		}
	}
	if !foundUSDT {
		t.Error("USDT balance not found")
	}
}

func TestGetBalance(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/account": {
			body: `{
				"canTrade": true,
				"canWithdraw": true,
				"canDeposit": true,
				"balances": [
					{"asset": "BTC", "free": "0.50000000", "locked": "0.10000000"},
					{"asset": "ETH", "free": "10.00000000", "locked": "0.00000000"},
					{"asset": "USDT", "free": "50000.00000000", "locked": "1000.00000000"}
				]
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)

	// Test existing asset
	bal, err := p.GetBalance("BTC")
	if err != nil {
		t.Fatalf("GetBalance(BTC) failed: %v", err)
	}
	if bal.Asset != "BTC" {
		t.Errorf("expected BTC, got %s", bal.Asset)
	}
	if bal.Free != 0.5 {
		t.Errorf("expected free 0.5, got %f", bal.Free)
	}
	if bal.Total != 0.6 {
		t.Errorf("expected total 0.6, got %f", bal.Total)
	}

	// Test non-existent asset
	bal, err = p.GetBalance("DOGE")
	if err != nil {
		t.Fatalf("GetBalance(DOGE) failed: %v", err)
	}
	if bal.Asset != "DOGE" {
		t.Errorf("expected DOGE, got %s", bal.Asset)
	}
	if bal.Free != 0 || bal.Locked != 0 || bal.Total != 0 {
		t.Errorf("expected zero balance for DOGE, got %+v", bal)
	}
}

func TestGetOrderStatus(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/order": {
			body: `{
				"symbol": "BTCUSDT",
				"orderId": 987654321,
				"clientOrderId": "xyz789",
				"price": "67000.00",
				"origQty": "0.50000000",
				"executedQty": "0.25000000",
				"cummulativeQuoteQty": "16750.00000000",
				"status": "PARTIALLY_FILLED",
				"type": "LIMIT",
				"side": "BUY",
				"transactTime": 1718380800000,
				"updateTime": 1718380900000
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	order, err := p.GetOrderStatus("BTCUSDT", "987654321")
	if err != nil {
		t.Fatalf("GetOrderStatus failed: %v", err)
	}

	if order.ID != "987654321" {
		t.Errorf("expected order ID 987654321, got %s", order.ID)
	}
	if order.Status != exchange.OrderPartiallyFiel {
		t.Errorf("expected PARTIALLY_FILLED, got %s", order.Status)
	}
	if order.Side != exchange.Buy {
		t.Errorf("expected BUY, got %s", order.Side)
	}
	if order.Type != exchange.Limit {
		t.Errorf("expected LIMIT, got %s", order.Type)
	}
	if order.ExecutedQty != 0.25 {
		t.Errorf("expected executed 0.25, got %f", order.ExecutedQty)
	}
	if order.QuoteOrderQty != 16750.0 {
		t.Errorf("expected quote qty 16750, got %f", order.QuoteOrderQty)
	}
}

func TestCancelOrder(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/order": {
			body:   `{"symbol":"BTCUSDT","orderId":12345,"status":"CANCELED"}`,
			statusCode: http.StatusOK,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	err := p.CancelOrder("BTCUSDT", "12345")
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
}

func TestCancelOrderDryRun(t *testing.T) {
	// In dry-run, no request should be made.
	server := mockServer(map[string]mockResponse{})
	defer server.Close()

	p := newTestProvider(server.URL)
	p.SetDryRun(true)

	err := p.CancelOrder("BTCUSDT", "12345")
	if err != nil {
		t.Fatalf("CancelOrder dry-run failed: %v", err)
	}
}

func TestGetTickerAPIError(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/ticker/24hr": {
			body:       `{"code":-1121,"msg":"Invalid symbol."}`,
			statusCode: http.StatusBadRequest,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)
	_, err := p.GetTicker("INVALID")
	if err == nil {
		t.Fatal("expected error for invalid symbol, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected error to contain HTTP status, got: %v", err)
	}
}

func TestMissingCredentials(t *testing.T) {
	p := NewProviderWithClient("", "", "http://localhost", &http.Client{Timeout: time.Second})
	_, err := p.GetAccount()
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if !strings.Contains(err.Error(), "API_KEY") {
		t.Errorf("expected error about missing API key, got: %v", err)
	}
}

func TestSignParams(t *testing.T) {
	// Verify HMAC-SHA256 signing
	p := NewProviderWithClient("api-key", "secret-key", "http://localhost", nil)
	sig := p.signParams("symbol=BTCUSDT&timestamp=1234567890")
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
	if len(sig) != 64 {
		t.Errorf("expected 64-char hex signature, got %d chars", len(sig))
	}
}

func TestGetOrderBookInvalidLimit(t *testing.T) {
	server := mockServer(map[string]mockResponse{
		"/depth": {
			body: `{
				"lastUpdateId": 123456,
				"bids": [],
				"asks": []
			}`,
		},
	})
	defer server.Close()

	p := newTestProvider(server.URL)

	// Test with negative limit — should default to 100 and still work
	book, err := p.GetOrderBook("BTCUSDT", -1)
	if err != nil {
		t.Fatalf("GetOrderBook with invalid limit failed: %v", err)
	}
	if book.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", book.Symbol)
	}
}

// ── WebSocket Frame Decoder Tests ──

func TestWSDecoderTextFrame(t *testing.T) {
	// A valid unmasked text frame with payload "hello"
	// Byte 0: 0x81 (FIN=1, opcode=1=text)
	// Byte 1: 0x05 (payload length = 5)
	// Bytes 2-6: "hello"
	frame := []byte{0x81, 0x05, 'h', 'e', 'l', 'l', 'o'}

	decoder := newWSDecoder(strings.NewReader(string(frame)))
	msg, err := decoder.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}
	if string(msg) != "hello" {
		t.Errorf("expected 'hello', got %q", string(msg))
	}
}

func TestWSDecoderMaskedFrame(t *testing.T) {
	// A masked text frame with payload "test"
	// Mask key: [0x01, 0x02, 0x03, 0x04]
	// Payload "test" XORed with mask key
	maskKey := []byte{0x01, 0x02, 0x03, 0x04}
	payload := []byte("test")
	masked := make([]byte, len(payload))
	for i := range payload {
		masked[i] = payload[i] ^ maskKey[i%4]
	}

	frame := append([]byte{0x81, 0x84, 0x01, 0x02, 0x03, 0x04}, masked...)

	decoder := newWSDecoder(strings.NewReader(string(frame)))
	msg, err := decoder.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}
	if string(msg) != "test" {
		t.Errorf("expected 'test', got %q", string(msg))
	}
}

func TestWSDecoderExtendedLength(t *testing.T) {
	// Test a text frame with payload length 126 (uses 2-byte extended length)
	payload := strings.Repeat("A", 126)
	frame := append([]byte{0x81, 126, byte(len(payload) >> 8), byte(len(payload))}, []byte(payload)...)

	decoder := newWSDecoder(strings.NewReader(string(frame)))
	msg, err := decoder.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}
	if string(msg) != payload {
		t.Errorf("expected %d chars, got %d", len(payload), len(msg))
	}
}

func TestWSDecoderExtendedLength64(t *testing.T) {
	// Test a text frame with payload length 127 (uses 8-byte extended length)
	payload := strings.Repeat("B", 300)
	frame := []byte{0x81, 127}
	// 8-byte big-endian length
	lenBytes := make([]byte, 8)
	lenBytes[6] = byte(len(payload) >> 8)
	lenBytes[7] = byte(len(payload))
	frame = append(frame, lenBytes...)
	frame = append(frame, []byte(payload)...)

	decoder := newWSDecoder(strings.NewReader(string(frame)))
	msg, err := decoder.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}
	if string(msg) != payload {
		t.Errorf("expected %d chars, got %d", len(payload), len(msg))
	}
}

func TestWSDecoderCloseFrame(t *testing.T) {
	// Close frame (opcode 0x8)
	frame := []byte{0x88, 0x00}

	decoder := newWSDecoder(strings.NewReader(string(frame)))
	_, err := decoder.ReadMessage()
	if err == nil {
		t.Fatal("expected error for close frame, got nil")
	}
	if !strings.Contains(err.Error(), "close frame") {
		t.Errorf("expected 'close frame' error, got: %v", err)
	}
}

// ── Order JSON Parsing Tests ──

func TestParseOrderResponse(t *testing.T) {
	jsonStr := `{
		"symbol": "ETHUSDT",
		"orderId": 555,
		"clientOrderId": "test123",
		"transactTime": 1718380800000,
		"price": "3500.00",
		"origQty": "2.00000000",
		"executedQty": "2.00000000",
		"cummulativeQuoteQty": "7000.00000000",
		"status": "FILLED",
		"type": "LIMIT",
		"side": "SELL",
		"updateTime": 1718380900000
	}`

	order, err := parseOrderResponse([]byte(jsonStr))
	if err != nil {
		t.Fatalf("parseOrderResponse failed: %v", err)
	}

	if order.ID != "555" {
		t.Errorf("expected ID 555, got %s", order.ID)
	}
	if order.Symbol != "ETHUSDT" {
		t.Errorf("expected ETHUSDT, got %s", order.Symbol)
	}
	if order.Side != exchange.Sell {
		t.Errorf("expected SELL, got %s", order.Side)
	}
	if order.Type != exchange.Limit {
		t.Errorf("expected LIMIT, got %s", order.Type)
	}
	if order.Price != 3500.00 {
		t.Errorf("expected price 3500, got %.2f", order.Price)
	}
	if order.Quantity != 2.0 {
		t.Errorf("expected qty 2.0, got %f", order.Quantity)
	}
	if order.ExecutedQty != 2.0 {
		t.Errorf("expected executed 2.0, got %f", order.ExecutedQty)
	}
	if order.QuoteOrderQty != 7000.0 {
		t.Errorf("expected quote qty 7000, got %f", order.QuoteOrderQty)
	}
}

func TestParseOrderStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected exchange.OrderStatus
	}{
		{"NEW", exchange.OrderNew},
		{"PARTIALLY_FILLED", exchange.OrderPartiallyFiel},
		{"FILLED", exchange.OrderFilled},
		{"CANCELED", exchange.OrderCanceled},
		{"REJECTED", exchange.OrderRejected},
		{"EXPIRED", exchange.OrderExpired},
		{"UNKNOWN", exchange.OrderStatus("UNKNOWN")},
	}

	for _, tt := range tests {
		got := parseOrderStatus(tt.input)
		if got != tt.expected {
			t.Errorf("parseOrderStatus(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSubscribeTickersNoSymbols(t *testing.T) {
	p := NewProvider()
	_, _, err := p.SubscribeTickers()
	if err == nil {
		t.Fatal("expected error for empty symbols, got nil")
	}
}

// ── Integration with Mock JSON ──

func TestBinanceOrderJSONRoundTrip(t *testing.T) {
	// Verify that the order parsing handles real Binance JSON correctly
	rawJSON := `{
		"symbol": "BTCUSDT",
		"orderId": 12345,
		"clientOrderId": "web_abc123",
		"transactTime": 1718380800000,
		"price": "67250.00000000",
		"origQty": "0.01000000",
		"executedQty": "0.01000000",
		"cummulativeQuoteQty": "672.50000000",
		"status": "FILLED",
		"type": "LIMIT",
		"side": "BUY",
		"updateTime": 1718380800000
	}`

	var raw struct {
		OrderID         int    `json:"orderId"`
		Symbol          string `json:"symbol"`
		Status          string `json:"status"`
		Type            string `json:"type"`
		Side            string `json:"side"`
		Price           string `json:"price"`
		OrigQty         string `json:"origQty"`
		ExecutedQty     string `json:"executedQty"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &raw); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if raw.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", raw.Symbol)
	}
	if raw.OrderID != 12345 {
		t.Errorf("expected order ID 12345, got %d", raw.OrderID)
	}
	if raw.Price != "67250.00000000" {
		t.Errorf("expected price 67250.00000000, got %s", raw.Price)
	}
}

