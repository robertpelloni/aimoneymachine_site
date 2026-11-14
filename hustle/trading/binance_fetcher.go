package trading

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/hustle/hustle/trading/exchange/binance"
)

// BinancePriceFetcher adapts the full Binance ExchangeProvider (which
// already implements signed orders/account/websockets under
// exchange/binance) to the PriceFetcher interface the TradingModule
// actually consumes: last price for SYMBOL/USDT. Price reads use the
// public ticker endpoint, so no API keys are required; keys in the
// environment are only used by the provider's signed calls.
//
// Select it with TRADING_PRICE_SOURCE=binance (default: coingecko).
// Same graceful-degradation contract as CoinGeckoFetcher: 60s cache,
// and a stale last-known price instead of an error when the API fails.
type BinancePriceFetcher struct {
	provider *binance.Provider
	cache    map[string]cachedPrice
	mu       sync.Mutex
}

// NewBinancePriceFetcher builds a fetcher from BINANCE_API_URL /
// BINANCE_API_KEY / BINANCE_SECRET_KEY (see binance.NewProvider).
func NewBinancePriceFetcher() *BinancePriceFetcher {
	return &BinancePriceFetcher{
		provider: binance.NewProvider(),
		cache:    make(map[string]cachedPrice),
	}
}

// NewBinancePriceFetcherWithProvider is the test seam.
func NewBinancePriceFetcherWithProvider(p *binance.Provider) *BinancePriceFetcher {
	return &BinancePriceFetcher{provider: p, cache: make(map[string]cachedPrice)}
}

func (f *BinancePriceFetcher) GetPrice(symbol string) (float64, error) {
	symbol = strings.ToUpper(symbol)
	if price, ok := f.getFromCache(symbol); ok {
		return price, nil
	}
	pair := symbol + "USDT"
	if symbol == "USDT" {
		pair = "USDTUSD"
	}
	ticker, err := f.provider.GetTicker(pair)
	if err == nil && ticker.LastPrice > 0 {
		f.addToCache(symbol, ticker.LastPrice)
		return ticker.LastPrice, nil
	}
	if err == nil {
		err = fmt.Errorf("binance ticker for %s has no last price", pair)
	}
	if stale, ok := f.getStale(symbol); ok {
		log.Printf("[Binance] Serving stale price for %s ($%.2f); fetch failed: %v", symbol, stale, err)
		return stale, nil
	}
	return 0, fmt.Errorf("binance: %s: %w", symbol, err)
}

func (f *BinancePriceFetcher) getFromCache(symbol string) (float64, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.cache[symbol]
	if !ok || time.Now().After(entry.expiresAt) {
		return 0, false
	}
	return entry.price, true
}

func (f *BinancePriceFetcher) getStale(symbol string) (float64, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.cache[symbol]
	if !ok {
		return 0, false
	}
	return entry.price, true
}

func (f *BinancePriceFetcher) addToCache(symbol string, price float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache[symbol] = cachedPrice{price: price, expiresAt: time.Now().Add(cacheTTL)}
}
