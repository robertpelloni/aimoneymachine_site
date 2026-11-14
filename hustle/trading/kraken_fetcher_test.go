package trading

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestKrakenPriceFetcher_SuccessCacheAndPairMapping(t *testing.T) {
	calls := 0
	var lastPair string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		lastPair = r.URL.Query().Get("pair")
		fmt.Fprint(w, `{"error":[],"result":{"XXBTZUSD":{"c":["67400.5","0.001"]}}}`)
	}))
	defer ts.Close()

	f := NewKrakenPriceFetcherWithClient(ts.URL, &http.Client{Timeout: 5 * time.Second})
	p, err := f.GetPrice("BTC")
	if err != nil || p != 67400.5 {
		t.Fatalf("GetPrice = %.2f, %v; want 67400.5, nil", p, err)
	}
	if lastPair != "XBTUSD" {
		t.Errorf("pair = %q, want XBTUSD (Kraken calls BTC XBT)", lastPair)
	}
	if _, err := f.GetPrice("BTC"); err != nil {
		t.Fatalf("cached: %v", err)
	}
	if calls != 1 {
		t.Errorf("API calls = %d, want 1 (cache)", calls)
	}
}

func TestKrakenPriceFetcher_ErrorAndStale(t *testing.T) {
	mode := "ok"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mode == "apierr" {
			fmt.Fprint(w, `{"error":["EQuery:Unknown asset pair"],"result":{}}`)
			return
		}
		fmt.Fprint(w, `{"error":[],"result":{"XETHZUSD":{"c":["3000.25","1"]}}}`)
	}))
	defer ts.Close()

	f := NewKrakenPriceFetcherWithClient(ts.URL, &http.Client{Timeout: 5 * time.Second})
	if _, err := f.GetPrice("ETH"); err != nil {
		t.Fatalf("prime: %v", err)
	}
	f.mu.Lock()
	e := f.cache["ETH"]
	e.expiresAt = time.Now().Add(-time.Hour)
	f.cache["ETH"] = e
	f.mu.Unlock()
	mode = "apierr"

	p, err := f.GetPrice("ETH")
	if err != nil || p != 3000.25 {
		t.Fatalf("stale = %.2f, %v; want 3000.25, nil", p, err)
	}
	f2 := NewKrakenPriceFetcherWithClient(ts.URL, &http.Client{Timeout: 5 * time.Second})
	if _, err := f2.GetPrice("ETH"); err == nil || !strings.Contains(err.Error(), "Unknown asset pair") {
		t.Fatalf("no-cache err = %v, want kraken error surfaced", err)
	}
}
