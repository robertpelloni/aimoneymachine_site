package trading

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

// The v2 dashboard reads samples/last_price/bb_percent_b/recent_prices/
// in_position; the backend emitted points/price/bb_pct_b/sparkline, so
// the UI showed Cycles 0, BB %B 0%, Last $0 on live data. SnapshotAll
// must populate both key sets consistently.
func TestSnapshotAllDashboardAliases(t *testing.T) {
	m := &TradingModule{
		Orchestrator: orchestrator.NewOrchestrator(),
		Fetcher:      fixedFetcher{prices: map[string]float64{"BTC": 100}},
	}
	for i := 0; i < 40; i++ {
		if err := m.ExecuteFor("BTC"); err != nil {
			t.Fatalf("cycle %d: %v", i, err)
		}
	}
	snaps := m.SnapshotAll()
	var btc *BotSnapshot
	for i := range snaps {
		if snaps[i].Symbol == "BTC" {
			btc = &snaps[i]
		}
	}
	if btc == nil {
		t.Fatal("BTC snapshot missing")
	}
	if btc.Samples != btc.Points || btc.Samples != 40 {
		t.Errorf("samples=%d points=%d, want both 40", btc.Samples, btc.Points)
	}
	if btc.LastPrice != 100 || btc.Price != 100 {
		t.Errorf("last_price=%v price=%v, want 100/100", btc.LastPrice, btc.Price)
	}
	if btc.BBPercentB != btc.BBPctB {
		t.Errorf("bb_percent_b=%v bb_pct_b=%v, want equal", btc.BBPercentB, btc.BBPctB)
	}
	if len(btc.RecentPrices) != len(btc.Sparkline) || len(btc.RecentPrices) == 0 {
		t.Errorf("recent_prices=%d sparkline=%d, want equal and non-empty", len(btc.RecentPrices), len(btc.Sparkline))
	}
	if btc.InPosition {
		t.Error("in_position must stay false (no portfolio engine)")
	}
	raw, err := json.Marshal(btc)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"samples":40`, `"last_price":100`, `"bb_percent_b"`, `"recent_prices"`, `"in_position":false`} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("snapshot JSON missing %s: %.120s", key, raw)
		}
	}
}
