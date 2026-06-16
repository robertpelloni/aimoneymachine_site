// Package binance implements the ExchangeProvider interface for Binance.
package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/hustle/hustle/trading/exchange"
)

// Defaults
const (
	defaultBaseURL    = "https://api.binance.com"
	apiVersion        = "/api/v3"
	defaultRecvWindow = 5000 // milliseconds
	httpTimeout       = 15 * time.Second
	wsBaseURL         = "wss://stream.binance.com:9443/ws"
)

// Provider implements exchange.ExchangeProvider for Binance.
type Provider struct {
	apiKey     string
	secretKey  string
	baseURL    string
	client     *http.Client
	dryRun     bool
	mu         sync.Mutex
	wsCancel   func()
}

// NewProvider creates a Binance exchange provider from environment variables.
// Reads BINANCE_API_KEY and BINANCE_SECRET_KEY.
func NewProvider() *Provider {
	return &Provider{
		apiKey:    os.Getenv("BINANCE_API_KEY"),
		secretKey: os.Getenv("BINANCE_SECRET_KEY"),
		baseURL:   getEnvDefault("BINANCE_API_URL", defaultBaseURL),
		client:    &http.Client{Timeout: httpTimeout},
	}
}

// NewProviderWithClient creates a provider with a custom HTTP client (for testing).
func NewProviderWithClient(apiKey, secretKey, baseURL string, client *http.Client) *Provider {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if client == nil {
		client = &http.Client{Timeout: httpTimeout}
	}
	return &Provider{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		client:    client,
	}
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Name returns "binance".
func (p *Provider) Name() string { return "binance" }

// DryRun returns whether dry-run mode is active.
func (p *Provider) DryRun() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dryRun
}

// SetDryRun enables or disables dry-run mode.
func (p *Provider) SetDryRun(dry bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dryRun = dry
}

// ── REST Helpers ──

// signParams generates an HMAC-SHA256 signature for the given query string.
func (p *Provider) signParams(query string) string {
	mac := hmac.New(sha256.New, []byte(p.secretKey))
	mac.Write([]byte(query))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

// publicRequest performs an unsigned GET request to a public endpoint.
func (p *Provider) publicRequest(endpoint string, params url.Values) ([]byte, error) {
	u, _ := url.Parse(p.baseURL + apiVersion + endpoint)
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// signedRequest performs a SIGNED GET request with HMAC-SHA256 authentication.
func (p *Provider) signedRequest(method, endpoint string, params url.Values) ([]byte, error) {
	if p.apiKey == "" || p.secretKey == "" {
		return nil, fmt.Errorf("binance: BINANCE_API_KEY and BINANCE_SECRET_KEY must be set")
	}

	if params == nil {
		params = url.Values{}
	}
	params.Set("timestamp", fmt.Sprintf("%d", time.Now().UnixMilli()))
	params.Set("recvWindow", strconv.Itoa(defaultRecvWindow))

	queryStr := params.Encode()
	signature := p.signParams(queryStr)

	u, _ := url.Parse(p.baseURL + apiVersion + endpoint)
	u.RawQuery = queryStr + "&signature=" + signature

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ── Interface Implementation ──

// GetTicker fetches 24hr ticker for a symbol.
func (p *Provider) GetTicker(symbol string) (*exchange.Ticker, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))

	body, err := p.publicRequest("/ticker/24hr", params)
	if err != nil {
		return nil, fmt.Errorf("get ticker: %w", err)
	}

	var raw struct {
		Symbol      string `json:"symbol"`
		BidPrice    string `json:"bidPrice"`
		AskPrice    string `json:"askPrice"`
		LastPrice   string `json:"lastPrice"`
		Volume      string `json:"volume"`
		HighPrice   string `json:"highPrice"`
		LowPrice    string `json:"lowPrice"`
		PriceChange string `json:"priceChange"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing ticker: %w", err)
	}

	ticker := &exchange.Ticker{
		Symbol:    raw.Symbol,
		Timestamp: time.Now(),
	}
	ticker.BidPrice, _ = strconv.ParseFloat(raw.BidPrice, 64)
	ticker.AskPrice, _ = strconv.ParseFloat(raw.AskPrice, 64)
	ticker.LastPrice, _ = strconv.ParseFloat(raw.LastPrice, 64)
	ticker.Volume, _ = strconv.ParseFloat(raw.Volume, 64)
	ticker.HighPrice, _ = strconv.ParseFloat(raw.HighPrice, 64)
	ticker.LowPrice, _ = strconv.ParseFloat(raw.LowPrice, 64)
	ticker.PriceChange, _ = strconv.ParseFloat(raw.PriceChange, 64)

	return ticker, nil
}

// GetOrderBook fetches the order book for a symbol.
func (p *Provider) GetOrderBook(symbol string, limit int) (*exchange.OrderBook, error) {
	if limit <= 0 || limit > 5000 {
		limit = 100 // default
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("limit", strconv.Itoa(limit))

	body, err := p.publicRequest("/depth", params)
	if err != nil {
		return nil, fmt.Errorf("get order book: %w", err)
	}

	var raw struct {
		LastUpdateID int             `json:"lastUpdateId"`
		Bids         [][]interface{} `json:"bids"`
		Asks         [][]interface{} `json:"asks"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing order book: %w", err)
	}

	book := &exchange.OrderBook{
		Symbol:    strings.ToUpper(symbol),
		Timestamp: time.Now(),
	}

	for _, b := range raw.Bids {
		if len(b) >= 2 {
			price, _ := strconv.ParseFloat(fmt.Sprintf("%v", b[0]), 64)
			qty, _ := strconv.ParseFloat(fmt.Sprintf("%v", b[1]), 64)
			book.Bids = append(book.Bids, exchange.OrderBookEntry{Price: price, Quantity: qty})
		}
	}
	for _, a := range raw.Asks {
		if len(a) >= 2 {
			price, _ := strconv.ParseFloat(fmt.Sprintf("%v", a[0]), 64)
			qty, _ := strconv.ParseFloat(fmt.Sprintf("%v", a[1]), 64)
			book.Asks = append(book.Asks, exchange.OrderBookEntry{Price: price, Quantity: qty})
		}
	}

	return book, nil
}

// PlaceOrder places an order on Binance.
func (p *Provider) PlaceOrder(symbol string, side exchange.OrderSide, orderType exchange.OrderType, quantity, price float64) (*exchange.Order, error) {
	if p.DryRun() {
		log.Printf("[Binance] DRY RUN: Would place %s %s order for %s qty=%.4f price=%.2f", side, orderType, symbol, quantity, price)
		return &exchange.Order{
			ID:        "dry-run-" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Symbol:    strings.ToUpper(symbol),
			Side:      side,
			Type:      orderType,
			Status:    exchange.OrderFilled,
			Price:     price,
			Quantity:  quantity,
			CreatedAt: time.Now(),
		}, nil
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("side", string(side))
	params.Set("type", string(orderType))
	params.Set("quantity", formatFloat(quantity, 8))

	if orderType == exchange.Limit {
		params.Set("price", formatFloat(price, 8))
		params.Set("timeInForce", "GTC") // Good til cancelled
	}
	// For MARKET orders with quote quantity
	if orderType == exchange.Market && price > 0 {
		params.Set("quoteOrderQty", formatFloat(price*quantity, 8))
	}

	body, err := p.signedRequest("POST", "/order", params)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	return parseOrderResponse(body)
}

// CancelOrder cancels an open order.
func (p *Provider) CancelOrder(symbol, orderID string) error {
	if p.DryRun() {
		log.Printf("[Binance] DRY RUN: Would cancel order %s for %s", orderID, symbol)
		return nil
	}

	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("orderId", orderID)

	_, err := p.signedRequest("DELETE", "/order", params)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	return nil
}

// GetOrderStatus retrieves the status of an order.
func (p *Provider) GetOrderStatus(symbol, orderID string) (*exchange.Order, error) {
	params := url.Values{}
	params.Set("symbol", strings.ToUpper(symbol))
	params.Set("orderId", orderID)

	body, err := p.signedRequest("GET", "/order", params)
	if err != nil {
		return nil, fmt.Errorf("get order status: %w", err)
	}

	return parseOrderResponse(body)
}

// GetAccount fetches account balances and permissions.
func (p *Provider) GetAccount() (*exchange.Account, error) {
	body, err := p.signedRequest("GET", "/account", nil)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	var raw struct {
		CanTrade    bool `json:"canTrade"`
		CanWithdraw bool `json:"canWithdraw"`
		CanDeposit  bool `json:"canDeposit"`
		Balances    []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing account: %w", err)
	}

	acc := &exchange.Account{
		CanTrade:    raw.CanTrade,
		CanWithdraw: raw.CanWithdraw,
		CanDeposit:  raw.CanDeposit,
	}
	for _, b := range raw.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		// Skip zero balances to keep response concise
		if free == 0 && locked == 0 {
			continue
		}
		acc.Balances = append(acc.Balances, exchange.Balance{
			Asset:  b.Asset,
			Free:   free,
			Locked: locked,
			Total:  free + locked,
		})
	}

	return acc, nil
}

// GetBalance fetches balance for a single asset.
func (p *Provider) GetBalance(asset string) (*exchange.Balance, error) {
	acc, err := p.GetAccount()
	if err != nil {
		return nil, err
	}

	asset = strings.ToUpper(asset)
	for _, b := range acc.Balances {
		if b.Asset == asset {
			return &b, nil
		}
	}

	return &exchange.Balance{Asset: asset, Free: 0, Locked: 0, Total: 0}, nil
}

// SubscribeTickers streams real-time ticker data via WebSocket.
func (p *Provider) SubscribeTickers(symbols ...string) (<-chan *exchange.Ticker, func(), error) {
	if len(symbols) == 0 {
		return nil, nil, fmt.Errorf("binance: at least one symbol required")
	}

	// Build stream names: btcusdt@ticker, ethusdt@ticker, etc.
	var streams []string
	for _, s := range symbols {
		streams = append(streams, strings.ToLower(s)+"@ticker")
	}

	streamName := strings.Join(streams, "/")
	wsURL := fmt.Sprintf("%s/%s", wsBaseURL, streamName)

	ch := make(chan *exchange.Ticker, 100)

	// Use gorilla/websocket via a simple raw HTTP upgrade
	// Since we want to minimize deps, we do a simple net/http upgrade
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(ch)

		// Try up to 3 reconnects
		for attempt := 0; attempt < 3; attempt++ {
			select {
			case <-stopChan:
				return
			default:
			}

			err := p.connectWebSocket(wsURL, ch, stopChan)
			if err != nil {
				log.Printf("[Binance] WebSocket error (attempt %d/3): %v", attempt+1, err)
				select {
				case <-stopChan:
					return
				case <-time.After(time.Duration(attempt+1) * time.Second):
				}
			}
		}
	}()

	cancel := func() {
		close(stopChan)
		p.mu.Lock()
		if p.wsCancel != nil {
			p.wsCancel()
		}
		p.mu.Unlock()
	}

	p.mu.Lock()
	p.wsCancel = cancel
	p.mu.Unlock()

	return ch, cancel, nil
}

// ── WebSocket Implementation ──

// wsTickerEvent mirrors the Binance WebSocket 24hr ticker event.
type wsTickerEvent struct {
	EventType  string `json:"e"`
	EventTime  int64  `json:"E"`
	Symbol     string `json:"s"`
	PriceChange string `json:"p"`
	LastPrice  string `json:"c"`
	BidPrice   string `json:"b"`
	AskPrice   string `json:"a"`
	HighPrice  string `json:"h"`
	LowPrice   string `json:"l"`
	Volume     string `json:"v"`
}

func (p *Provider) connectWebSocket(wsURL string, ch chan<- *exchange.Ticker, stop <-chan struct{}) error {
	// Use standard HTTP to upgrade to WebSocket
	req, err := http.NewRequest("GET", wsURL, nil)
	if err != nil {
		return fmt.Errorf("creating WS request: %w", err)
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==") // dummy base64 key

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("WS upgrade: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("WS upgrade failed: HTTP %d", resp.StatusCode)
	}

	// Read WebSocket frames using a simple framing decoder
	decoder := newWSDecoder(resp.Body)

	for {
		select {
		case <-stop:
			return nil
		default:
		}

		msg, err := decoder.ReadMessage()
		if err != nil {
			return fmt.Errorf("WS read: %w", err)
		}

		var evt wsTickerEvent
		if err := json.Unmarshal(msg, &evt); err != nil {
			continue // skip non-ticker messages
		}

		if evt.EventType != "24hrTicker" {
			continue
		}

		ticker := &exchange.Ticker{
			Symbol:    evt.Symbol,
			Timestamp: time.UnixMilli(evt.EventTime),
		}
		ticker.LastPrice, _ = strconv.ParseFloat(evt.LastPrice, 64)
		ticker.BidPrice, _ = strconv.ParseFloat(evt.BidPrice, 64)
		ticker.AskPrice, _ = strconv.ParseFloat(evt.AskPrice, 64)
		ticker.HighPrice, _ = strconv.ParseFloat(evt.HighPrice, 64)
		ticker.LowPrice, _ = strconv.ParseFloat(evt.LowPrice, 64)
		ticker.Volume, _ = strconv.ParseFloat(evt.Volume, 64)
		ticker.PriceChange, _ = strconv.ParseFloat(evt.PriceChange, 64)

		select {
		case ch <- ticker:
		default:
			// Drop if channel full (backpressure)
		}
	}
}

// ── Helpers ──

func parseOrderResponse(body []byte) (*exchange.Order, error) {
	var raw struct {
		OrderID         int    `json:"orderId"`
		ClientOrderID   string `json:"clientOrderId"`
		Symbol          string `json:"symbol"`
		Side            string `json:"side"`
		Type            string `json:"type"`
		Status          string `json:"status"`
		Price           string `json:"price"`
		OrigQty         string `json:"origQty"`
		ExecutedQty     string `json:"executedQty"`
		CummulativeQty  string `json:"cummulativeQuoteQty"`
		TransactTime    int64  `json:"transactTime"`
		UpdateTime      int64  `json:"updateTime"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing order: %w", err)
	}

	order := &exchange.Order{
		ID:            fmt.Sprintf("%d", raw.OrderID),
		ClientOrderID: raw.ClientOrderID,
		Symbol:        raw.Symbol,
		Side:          exchange.OrderSide(raw.Side),
		Type:          exchange.OrderType(raw.Type),
		Status:        parseOrderStatus(raw.Status),
		CreatedAt:     time.UnixMilli(raw.TransactTime),
		UpdatedAt:     time.UnixMilli(raw.UpdateTime),
	}
	order.Price, _ = strconv.ParseFloat(raw.Price, 64)
	order.Quantity, _ = strconv.ParseFloat(raw.OrigQty, 64)
	order.ExecutedQty, _ = strconv.ParseFloat(raw.ExecutedQty, 64)
	order.QuoteOrderQty, _ = strconv.ParseFloat(raw.CummulativeQty, 64)

	return order, nil
}

func parseOrderStatus(s string) exchange.OrderStatus {
	switch s {
	case "NEW":
		return exchange.OrderNew
	case "PARTIALLY_FILLED":
		return exchange.OrderPartiallyFiel
	case "FILLED":
		return exchange.OrderFilled
	case "CANCELED":
		return exchange.OrderCanceled
	case "REJECTED":
		return exchange.OrderRejected
	case "EXPIRED":
		return exchange.OrderExpired
	default:
		return exchange.OrderStatus(s)
	}
}

func formatFloat(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}

// ── Simple WebSocket Frame Decoder ──
// This implements a minimal WebSocket client without external dependencies.
// It reads unmasked frames from the server (as per the spec, server frames are unmasked).

type wsDecoder struct {
	r io.Reader
}

func newWSDecoder(r io.Reader) *wsDecoder {
	return &wsDecoder{r: r}
}

func (d *wsDecoder) ReadMessage() ([]byte, error) {
	// Read frame header (2 bytes minimum)
	header := make([]byte, 2)
	if _, err := io.ReadFull(d.r, header); err != nil {
		return nil, fmt.Errorf("reading WS header: %w", err)
	}

	// fin := header[0]&0x80 != 0
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := int64(header[1] & 0x7F)

	// Extended payload length
	if payloadLen == 126 {
		ext := make([]byte, 2)
		if _, err := io.ReadFull(d.r, ext); err != nil {
			return nil, fmt.Errorf("reading extended length: %w", err)
		}
		payloadLen = int64(ext[0])<<8 | int64(ext[1])
	} else if payloadLen == 127 {
		ext := make([]byte, 8)
		if _, err := io.ReadFull(d.r, ext); err != nil {
			return nil, fmt.Errorf("reading extended length 64: %w", err)
		}
		payloadLen = 0
		for i := 0; i < 8; i++ {
			payloadLen = payloadLen<<8 | int64(ext[i])
		}
	}

	// Read mask key (if present — server frames are unmasked but we handle both)
	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(d.r, maskKey[:]); err != nil {
			return nil, fmt.Errorf("reading mask key: %w", err)
		}
	}

	// Read payload
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(d.r, payload); err != nil {
		return nil, fmt.Errorf("reading payload: %w", err)
	}

	// Unmask if needed
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	// Handle control frames
	switch opcode {
	case 0x8: // Close
		return nil, fmt.Errorf("WS close frame received")
	case 0x9: // Ping — respond with Pong
		return nil, nil // ignored for now; production should send pong
	case 0xA: // Pong
		return nil, nil
	case 0x1: // Text
		return payload, nil
	case 0x2: // Binary
		return payload, nil
	default:
		return nil, nil
	}
}
