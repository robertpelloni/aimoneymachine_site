// Package coinbase implements the ExchangeProvider interface for Coinbase.

package coinbase

import (
	"encoding/json"
	"errors"
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
	defaultBaseURL    = "https://api.pro.coinbase.com"  // Coinbase Pro (Advanced Trade) v3
	defaultAPIVersion = "/api/v3"
	httpTimeout       = 15 * time.Second

	// Retry and rate-limit configuration
	maxRetries        = 3
	retryBackoff      = 1 * time.Second
	tickerCacheTTL    = 5 * time.Second
	orderBookCacheTTL = 2 * time.Second
)

// apiError carries the HTTP status code from Coinbase for retry decision-making.
type apiError struct {
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("coinbase API error (status %d): %s", e.StatusCode, e.Body)
}

// cachedTicker stores a ticker result with its expiration.
type cachedTicker struct {
	ticker     *exchange.Ticker
	expiresAt  time.Time
}

// cachedOrderBook stores an order book result with its expiration.
type cachedOrderBook struct {
	book      *exchange.OrderBook
	expiresAt time.Time
}

// Provider implements exchange.ExchangeProvider for Coinbase.
type Provider struct {
	apiKey         string
	apiSecret      string  // base64 encoded string
	baseURL        string
	client         *http.Client
	mu             sync.RWMutex
	
	// Cached data
	tickerCache    map[string]cachedTicker
	orderBookCache map[string]cachedOrderBook
}

// NewProvider creates a Coinbase exchange provider from environment variables.
// Reads:
//   COINBASE_API_KEY      - API key
//   COINBASE_API_SECRET  - Secret key (base64 encoded)
func NewProvider() *Provider {
	return &Provider{
		apiKey:          os.Getenv("COINBASE_API_KEY"),
		apiSecret:       os.Getenv("COINBASE_API_SECRET"),
		baseURL:         getEnvDefault("COINBASE_API_URL", defaultBaseURL),
		client:          &http.Client{Timeout: httpTimeout},
		tickerCache:     make(map[string]cachedTicker),
		orderBookCache:  make(map[string]cachedOrderBook),
	}
}

// NewProviderWithClient creates a provider with a custom HTTP client (for testing).
func NewProviderWithClient(apiKey, apiSecret, baseURL string, client *http.Client) *Provider {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if client == nil {
		client = &http.Client{Timeout: httpTimeout}
	}
	return &Provider{
		apiKey:         apiKey,
		apiSecret:      apiSecret,
		baseURL:        baseURL,
		client:         client,
		tickerCache:    make(map[string]cachedTicker),
		orderBookCache: make(map[string]cachedOrderBook),
	}
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Name returns "coinbase".
func (p *Provider) Name() string {
	return "coinbase"
}

// isRetryable returns true if the error is a transient failure that can be retried.
func (p *Provider) isRetryable(err error) bool {
	if err == nil {
		return false
	}
	// API errors with status codes
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		// 429 (rate limited) and 5xx (server errors) are retryable
		return apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500
	}
	// Configuration errors have "coinbase:" prefix
	if strings.HasPrefix(err.Error(), "coinbase: ") {
		return false
	}
	// Network errors (connection refused, timeout, DNS failures) are retryable
	return true
}

// doPublicWithRetry wraps publicRequest with retry logic and backoff.
func (p *Provider) doPublicWithRetry(endpoint string, params url.Values) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryBackoff * time.Duration(attempt))
		}
		body, err := p.publicRequest(endpoint, params)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !p.isRetryable(err) {
			log.Printf("[Coinbase] publicRequest %s: non-retryable error: %v", endpoint, err)
			break
		}
		log.Printf("[Coinbase] publicRequest %s: retryable error (attempt %d/%d): %v", endpoint, attempt+1, maxRetries, err)
	}
	return nil, fmt.Errorf("coinbase: %s after %d retries: %w", endpoint, maxRetries, lastErr)
}

// publicRequest performs an unsigned GET request to a public endpoint.
func (p *Provider) publicRequest(endpoint string, params url.Values) ([]byte, error) {
	u, _ := url.Parse(p.baseURL + defaultAPIVersion + endpoint)
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Coinbase requires CB-ACCESS-SIGN and timestamp for private endpoints,
	// but public products endpoint doesn't for /products/ticker and /products/book
	// however we include here for uniformity

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
		return nil, &apiError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	return body, nil
}

// GetTicker fetches price ticker for a symbol with caching and retry.
func (p *Provider) GetTicker(symbol string) (*exchange.Ticker, error) {
	symbolKey := strings.ToUpper(symbol)

	// Check cache first
	p.mu.RLock()
	if entry, ok := p.tickerCache[symbolKey]; ok && time.Now().Before(entry.expiresAt) {
		p.mu.RUnlock()
		return entry.ticker, nil
	}
	p.mu.RUnlock()

	// Fetch with retry
	params := url.Values{}
	params.Set("product_id", symbolKey)

	body, err := p.doPublicWithRetry(fmt.Sprintf("/products/%s/ticker", symbolKey), params)
	if err != nil {
		return nil, fmt.Errorf("get ticker: %w", err)
	}

	var raw struct {
		TradeID   int    `json:"trade_id"`
		Price     string `json:"price"`
		Bid       string `json:"bid"`
		Ask       string `json:"ask"`
		Volume    string `json:"volume"`
		Time      string `json:"time"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing ticker: %w", err)
	}

	price, _ := strconv.ParseFloat(raw.Price, 64)
	bid, _ := strconv.ParseFloat(raw.Bid, 64)
	ask, _ := strconv.ParseFloat(raw.Ask, 64)
	volume, _ := strconv.ParseFloat(raw.Volume, 64)
	timeParsed, _ := time.Parse(time.RFC3339, raw.Time)

	if timeParsed.IsZero() {
		timeParsed = time.Now()
	}

	ticker := &exchange.Ticker{
		Symbol:    symbolKey,
		BidPrice:  bid,
		AskPrice:  ask,
		LastPrice: price,
		Volume:    volume,
		Timestamp: timeParsed,
	}

	// Store in cache
	p.mu.Lock()
	p.tickerCache[symbolKey] = cachedTicker{
		ticker:    ticker,
		expiresAt: time.Now().Add(tickerCacheTTL),
	}
	p.mu.Unlock()

	return ticker, nil
}

// GetOrderBook fetches the order book for a symbol with caching and retry.
func (p *Provider) GetOrderBook(symbol string, limit int) (*exchange.OrderBook, error) {
	symbolKey := strings.ToUpper(symbol)
	cacheKey := fmt.Sprintf("%s:%d", symbolKey, limit)

	p.mu.RLock()
	if cached, ok := p.orderBookCache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		p.mu.RUnlock()
		return cached.book, nil
	}
	p.mu.RUnlock()

	params := url.Values{}
	params.Set("level", "2")
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	body, err := p.doPublicWithRetry(fmt.Sprintf("/products/%s/book", symbolKey), params)
	if err != nil {
		return nil, fmt.Errorf("get order book: %w", err)
	}

	var raw struct {
		Sequence   string          `json:"sequence"`
		Bids       [][]interface{} `json:"bids"`
		Asks       [][]interface{} `json:"asks"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing order book: %w", err)
	}

	book := &exchange.OrderBook{
		Symbol:    symbolKey,
		Timestamp: time.Now(),
	}

	// Parse bids
	for _, entry := range raw.Bids {
		if len(entry) >= 2 {
			price, err1 := strconv.ParseFloat(fmt.Sprintf("%v", entry[0]), 64)
			quantity, err2 := strconv.ParseFloat(fmt.Sprintf("%v", entry[1]), 64)
			if err1 == nil && err2 == nil {
				book.Bids = append(book.Bids, exchange.OrderBookEntry{
					Price:    price,
					Quantity: quantity,
				})
			}
		}
	}

	// Parse asks
	for _, entry := range raw.Asks {
		if len(entry) >= 2 {
			price, err1 := strconv.ParseFloat(fmt.Sprintf("%v", entry[0]), 64)
			quantity, err2 := strconv.ParseFloat(fmt.Sprintf("%v", entry[1]), 64)
			if err1 == nil && err2 == nil {
				book.Asks = append(book.Asks, exchange.OrderBookEntry{
					Price:    price,
					Quantity: quantity,
				})
			}
		}
	}

	// Store in cache
	p.mu.Lock()
	defer p.mu.Unlock()
	p.orderBookCache[cacheKey] = cachedOrderBook{
		book:      book,
		expiresAt: time.Now().Add(orderBookCacheTTL),
	}

	return book, nil
}

// SetDryRun enables or disables dry-run mode (not implemented for Coinbase).
func (p *Provider) SetDryRun(dry bool) {
	// Coinbase does not support dry-run for order placement
	// But we could mock it for testing if needed
}

// DryRun returns whether dry-run mode is active (not implemented for Coinbase).
func (p *Provider) DryRun() bool {
	return false
}

// PlaceOrder and CancelOrder are NOT IMPLEMENTED for public Coinbase API
// These require OAuth2 or HMAC authentication and are not covered by public/read-only API
func (p *Provider) PlaceOrder(symbol string, side exchange.OrderSide, orderType exchange.OrderType, quantity, price float64) (*exchange.Order, error) {
	return nil, fmt.Errorf("coinbase: PlaceOrder not supported for read-only public API")
}

func (p *Provider) CancelOrder(symbol, orderID string) error {
	return fmt.Errorf("coinbase: CancelOrder not supported for read-only public API")
}

func (p *Provider) GetOrderStatus(symbol, orderID string) (*exchange.Order, error) {
	return nil, fmt.Errorf("coinbase: GetOrderStatus not supported for read-only public API")
}

func (p *Provider) GetAccount() (*exchange.Account, error) {
	return nil, fmt.Errorf("coinbase: GetAccount not supported for read-only public API")
}

func (p *Provider) GetBalance(asset string) (*exchange.Balance, error) {
	return nil, fmt.Errorf("coinbase: GetBalance not supported for read-only public API")
}

// SubscribeTickers is not implemented — Coinbase WebSocket API is more complex and may require connection pooling.
func (p *Provider) SubscribeTickers(symbols ...string) (<-chan *exchange.Ticker, func(), error) {
	return nil, nil, fmt.Errorf("coinbase: SubscribeTickers not implemented — use REST API with polling")
}

// ClearCache clears all cached data (useful for testing or admin).
func (p *Provider) ClearCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tickerCache = make(map[string]cachedTicker)
	p.orderBookCache = make(map[string]cachedOrderBook)
}