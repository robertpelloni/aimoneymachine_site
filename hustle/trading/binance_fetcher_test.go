package trading

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/hustle/hustle/trading/exchange/binance"
)

func binanceTestFetcher(url string) *BinancePriceFetcher {
	return NewBinancePriceFetcherWithProvider(
		binance.NewProviderWithClient("", "", url, &http.Client{Timeout: 5 * time.Second}))
}

func TestBinancePriceFetcher_SuccessAndCache(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if !strings.Contains(r.URL.Path, "/ticker/24hr") {
			t.Errorf("path = %s, want /ticker/24hr", r.URL.Path)
		}
		if got := r.URL.Query().Get("symbol"); got != "BTCUSDT" {
			t.Errorf("symbol param = %q, want BTCUSDT", got)
		}
		fmt.Fprint(w, `{"symbol":"BTCUSDT","bidPrice":"67000","askPrice":"67100","lastPrice":"67050.5","volume":"1","highPrice":"68000","lowPrice":"66000","priceChange":"10"}`)
	}))
	defer ts.Close()

	f := binanceTestFetcher(ts.URL)
	p, err := f.GetPrice("BTC")
	if err != nil || p != 67050.5 {
		t.Fatalf("GetPrice = %.2f, %v; want 67050.5, nil", p, err)
	}
	if _, err := f.GetPrice("BTC"); err != nil {
		t.Fatalf("cached GetPrice: %v", err)
	}
	if calls != 1 {
		t.Errorf("API calls = %d, want 1 (cache)", calls)
	}
}

func TestBinancePriceFetcher_StaleFallback(t *testing.T) {
	down := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if down {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"code":-1003,"msg":"down"}`)
			return
		}
		fmt.Fprint(w, `{"symbol":"ETHUSDT","lastPrice":"3000"}`)
	}))
	defer ts.Close()

	f := binanceTestFetcher(ts.URL)
	if _, err := f.GetPrice("ETH"); err != nil {
		t.Fatalf("prime: %v", err)
	}
	f.mu.Lock()
	e := f.cache["ETH"]
	e.expiresAt = time.Now().Add(-time.Hour)
	f.cache["ETH"] = e
	f.mu.Unlock()
	down = true

	p, err := f.GetPrice("ETH")
	if err != nil || p != 3000 {
		t.Fatalf("stale = %.2f, %v; want 3000, nil", p, err)
	}
	// No known price at all -> error, like CoinGecko.
	f2 := binanceTestFetcher(ts.URL)
	if _, err := f2.GetPrice("SOL"); err == nil {
		t.Fatal("expected error with no cached price and API down")
	}
}
