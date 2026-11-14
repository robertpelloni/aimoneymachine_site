package trading

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

// Series must survive a module restart (fresh instance, same file) so
// MACD's 34-point warmup is not reset by every deploy.
func TestTradingHistoryPersistenceRoundtrip(t *testing.T) {
	file := filepath.Join(t.TempDir(), "trading_history.json")

	m1 := &TradingModule{
		Orchestrator: orchestrator.NewOrchestrator(),
		Fetcher:      fixedFetcher{prices: map[string]float64{"BTC": 100, "ETH": 10}},
		HistoryFile:  file,
	}
	for i := 0; i < 3; i++ {
		if err := m1.ExecuteFor("BTC"); err != nil {
			t.Fatalf("BTC cycle: %v", err)
		}
	}
	if err := m1.ExecuteFor("ETH"); err != nil {
		t.Fatalf("ETH cycle: %v", err)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("history file not written: %v", err)
	}

	m2 := &TradingModule{
		Orchestrator: orchestrator.NewOrchestrator(),
		Fetcher:      fixedFetcher{prices: map[string]float64{"BTC": 100}},
		HistoryFile:  file,
	}
	if err := m2.LoadHistory(); err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}
	if got := len(m2.HistoryFor("BTC")); got != 3 {
		t.Errorf("restored BTC points = %d, want 3", got)
	}
	if got := len(m2.HistoryFor("ETH")); got != 1 {
		t.Errorf("restored ETH points = %d, want 1", got)
	}

	// Missing file is a fresh start, not an error.
	m3 := &TradingModule{HistoryFile: filepath.Join(t.TempDir(), "nope.json")}
	if err := m3.LoadHistory(); err != nil {
		t.Errorf("missing file should be nil error, got %v", err)
	}
}

func TestTrimSeriesCapsLength(t *testing.T) {
	s := make([]float64, 600)
	for i := range s {
		s[i] = float64(i)
	}
	got := trimSeries(s, maxHistoryPoints)
	if len(got) != maxHistoryPoints {
		t.Fatalf("len = %d, want %d", len(got), maxHistoryPoints)
	}
	if got[0] != 100 {
		t.Errorf("kept wrong end: first = %.0f, want 100 (newest 500)", got[0])
	}
}
