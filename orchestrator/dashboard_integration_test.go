package orchestrator

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDashboardSocialStatus(t *testing.T) {
	orch := &Orchestrator{
		L1: L1Memory{Entries: []MemoryEntry{}},
		L2: L2Memory{Entries: []MemoryEntry{}},
		L3: L3Memory{Entries: []MemoryEntry{}},
		Ledger: Ledger{Transactions: []Transaction{}},
		WealthGoal: 1000.0,
	}

	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

	// Test Offline Status
	os.Unsetenv("TWITTER_API_KEY")
	os.Unsetenv("TWITTER_ACCESS_TOKEN")
	os.Unsetenv("LINKEDIN_ACCESS_TOKEN")
	os.Unsetenv("LINKEDIN_AUTHOR_URN")

	output := captureOutput(func() {
		ShowDashboard(orch)
	})

	if !strings.Contains(output, "[✗ OFFLINE]") {
		t.Errorf("Expected OFFLINE status, got \n%s", output)
	}

	// Test Online Status
	os.Setenv("TWITTER_API_KEY", "test")
	os.Setenv("TWITTER_ACCESS_TOKEN", "test")
	os.Setenv("LINKEDIN_ACCESS_TOKEN", "test")
	os.Setenv("LINKEDIN_AUTHOR_URN", "test")
	defer func() {
		os.Unsetenv("TWITTER_API_KEY")
		os.Unsetenv("TWITTER_ACCESS_TOKEN")
		os.Unsetenv("LINKEDIN_ACCESS_TOKEN")
		os.Unsetenv("LINKEDIN_AUTHOR_URN")
	}()

	output = captureOutput(func() {
		ShowDashboard(orch)
	})

	if !strings.Contains(output, "[✓ ONLINE]") {
		t.Errorf("Expected ONLINE status, got \n%s", output)
	}
}
