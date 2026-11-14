package trading

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// KrakenPriceFetcher is the Kraken PriceFetcher (public Ticker endpoint,
// no keys required). Select with TRADING_PRICE_SOURCE=kraken.
// Same contract as the CoinGecko/Binance fetchers: 60s cache, stale
// last-known price instead of an error when the API fails.
type KrakenPriceFetcher struct {
	baseURL string
	client  *http.Client
	cache   map[string]cachedPrice
	mu      sync.Mutex
}

// krakenPair maps a ticker to Kraken's pair name (BTC is XBT there).
func krakenPair(symbol string) string {
	symbol = strings.ToUpper(symbol)
	if symbol == "BTC" {
		return "XBTUSD"
	}
	return symbol + "USD"
}

// NewKrakenPriceFetcher builds a fetcher from KRAKEN_API_URL
// (default https://api.kraken.com).
func NewKrakenPriceFetcher() *KrakenPriceFetcher {
	return &KrakenPriceFetcher{
		baseURL: getEnvDefault("KRAKEN_API_URL", "https://api.kraken.com"),
		client:  &http.Client{Timeout: 15 * time.Second},
		cache:   make(map[string]cachedPrice),
	}
}

// NewKrakenPriceFetcherWithClient is the test seam.
func NewKrakenPriceFetcherWithClient(baseURL string, client *http.Client) *KrakenPriceFetcher {
	return &KrakenPriceFetcher{baseURL: baseURL, client: client, cache: make(map[string]cachedPrice)}
}

type krakenTickerResponse struct {
	Error  []string `json:"error"`
	Result map[string]struct {
		C []string `json:"c"` // [last price, lot volume]
	} `json:"result"`
}

func (f *KrakenPriceFetcher) GetPrice(symbol string) (float64, error) {
	symbol = strings.ToUpper(symbol)
	if price, ok := f.getFromCache(symbol); ok {
		return price, nil
	}
	price, err := f.fetch(symbol)
	if err == nil {
		f.addToCache(symbol, price)
		return price, nil
	}
	if stale, ok := f.getStale(symbol); ok {
		log.Printf("[Kraken] Serving stale price for %s ($%.2f); fetch failed: %v", symbol, stale, err)
		return stale, nil
	}
	return 0, fmt.Errorf("kraken: %s: %w", symbol, err)
}

func (f *KrakenPriceFetcher) fetch(symbol string) (float64, error) {
	url := fmt.Sprintf("%s/0/public/Ticker?pair=%s", strings.TrimRight(f.baseURL, "/"), krakenPair(symbol))
	resp, err := f.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	var body krakenTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}
	if len(body.Error) > 0 {
		return 0, fmt.Errorf("kraken error: %s", strings.Join(body.Error, "; "))
	}
	// Result keys are Kraken-internal pair names (XXBTZUSD, XETHZUSD,
	// SOLUSD...); a single-pair query returns exactly one entry.
	for _, ticker := range body.Result {
		if len(ticker.C) == 0 {
			break
		}
		price, err := strconv.ParseFloat(ticker.C[0], 64)
		if err != nil {
			return 0, fmt.Errorf("parsing last price %q: %w", ticker.C[0], err)
		}
		if price <= 0 {
			return 0, fmt.Errorf("non-positive price %v", price)
		}
		return price, nil
	}
	return 0, fmt.Errorf("no ticker result for %s", symbol)
}

func (f *KrakenPriceFetcher) getFromCache(symbol string) (float64, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.cache[symbol]
	if !ok || time.Now().After(entry.expiresAt) {
		return 0, false
	}
	return entry.price, true
}

func (f *KrakenPriceFetcher) getStale(symbol string) (float64, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.cache[symbol]
	if !ok {
		return 0, false
	}
	return entry.price, true
}

func (f *KrakenPriceFetcher) addToCache(symbol string, price float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache[symbol] = cachedPrice{price: price, expiresAt: time.Now().Add(cacheTTL)}
}
