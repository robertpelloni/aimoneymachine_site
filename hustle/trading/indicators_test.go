package trading

import (
	"math"
	"testing"
)

func series(start, step float64, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = start + step*float64(i)
	}
	return out
}

func TestMACDTrendDirection(t *testing.T) {
	// On a perfectly steady linear trend the MACD line is a nonzero
	// constant (the fast/slow EMA gap) and the histogram is ~0 because
	// the signal line converges onto it.
	up := series(100, 1, 60)
	macd, _, hist, _ := calculateMACD(up)
	if macd <= 0 {
		t.Errorf("steady uptrend: MACD=%.4f, want positive", macd)
	}
	if math.Abs(hist) > 0.01 {
		t.Errorf("steady uptrend: hist=%.4f, want ~0 (signal converged)", hist)
	}
	down := series(200, -1, 60)
	macd, _, hist, _ = calculateMACD(down)
	if macd >= 0 {
		t.Errorf("steady downtrend: MACD=%.4f, want negative", macd)
	}
	if math.Abs(hist) > 0.01 {
		t.Errorf("steady downtrend: hist=%.4f, want ~0 (signal converged)", hist)
	}
	flat := series(100, 0, 60)
	macd, signal, hist, _ := calculateMACD(flat)
	if math.Abs(macd) > 1e-9 || math.Abs(signal) > 1e-9 || math.Abs(hist) > 1e-9 {
		t.Errorf("flat series: MACD/signal/hist = %.6f/%.6f/%.6f, want 0", macd, signal, hist)
	}
	if m, s, h, _ := calculateMACD(series(100, 1, 20)); m != 0 || s != 0 || h != 0 {
		t.Errorf("short series must return zeros, got %.4f/%.4f/%.4f", m, s, h)
	}
}

func TestBollingerBands(t *testing.T) {
	flat := series(100, 0, 40)
	upper, mid, lower, pctB := calculateBollinger(flat, 20, 2)
	if upper != 100 || mid != 100 || lower != 100 || pctB != 0.5 {
		t.Errorf("flat: bands %.2f/%.2f/%.2f pctB %.2f, want 100/100/100 and 0.5", upper, mid, lower, pctB)
	}
	// Ramp ending at the window max: last price sits at/near upper band.
	ramp := series(100, 1, 40)
	upper, mid, lower, pctB = calculateBollinger(ramp, 20, 2)
	if !(lower < mid && mid < upper) {
		t.Fatalf("bands unordered: %.2f %.2f %.2f", lower, mid, upper)
	}
	if pctB < 0.8 {
		t.Errorf("ramp last price pctB = %.2f, want >= 0.8 (near upper band)", pctB)
	}
}
