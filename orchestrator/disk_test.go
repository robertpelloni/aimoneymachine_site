package orchestrator

import "testing"

// DiskUsage must return sane values for the current directory on
// platforms that implement it (linux/windows CI and prod included).
func TestDiskUsageSane(t *testing.T) {
	pct, freeGB, ok := DiskUsage(".")
	if !ok {
		t.Skip("DiskUsage unimplemented on this platform")
	}
	if pct < 0 || pct > 100 {
		t.Errorf("used pct = %.1f, want 0..100", pct)
	}
	if freeGB < 0 {
		t.Errorf("free GB = %.1f, want >= 0", freeGB)
	}
}
