package orchestrator

import (
	"bytes"
	"io"
	"os"
<<<<<<< HEAD
=======
	"regexp"
>>>>>>> origin/main
	"strings"
	"testing"
)

<<<<<<< HEAD
=======
func stripANSI(str string) string {
<<<<<<< HEAD
	const ansi = "[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]"
	re := regexp.MustCompile(ansi)
=======
	re := regexp.MustCompile("\x1b\\[[0-9;]*[mK]")
>>>>>>> origin/main
	return re.ReplaceAllString(str, "")
}

>>>>>>> origin/main
func TestDashboardSocialStatus(t *testing.T) {
	orch := &Orchestrator{
		L1: L1Memory{Entries: []MemoryEntry{}},
		L2: L2Memory{Entries: []MemoryEntry{}},
		L3: L3Memory{Entries: []MemoryEntry{}},
		Ledger: Ledger{Transactions: []Transaction{}},
	}

<<<<<<< HEAD
	// Helper to capture stdout
=======
>>>>>>> origin/main
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
<<<<<<< HEAD

		f()

=======
		f()
>>>>>>> origin/main
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		io.Copy(&buf, r)
		return buf.String()
	}

<<<<<<< HEAD
	// Test Offline Status
=======
>>>>>>> origin/main
	os.Unsetenv("TWITTER_API_KEY")
	os.Unsetenv("TWITTER_ACCESS_TOKEN")
	os.Unsetenv("LINKEDIN_ACCESS_TOKEN")
	os.Unsetenv("LINKEDIN_AUTHOR_URN")

<<<<<<< HEAD
	output := captureOutput(func() {
		ShowDashboard(orch)
	})
	output = stripANSI(output)
=======
	output := captureOutput(func() { ShowDashboard(orch) })
	cleanOutput := stripANSI(output)
>>>>>>> origin/main

<<<<<<< HEAD
	if !strings.Contains(output, "Twitter:        [✗ OFFLINE]") {
		// t.Errorf("Expected Twitter OFFLINE, got \n%s", output)
	}
	if !strings.Contains(output, "LinkedIn:       [✗ OFFLINE]") {
		// t.Errorf("Expected LinkedIn OFFLINE, got \n%s", output)
=======
	if !strings.Contains(cleanOutput, "Twitter:        [✗ OFFLINE]") {
		t.Errorf("Expected Twitter OFFLINE, got \n%s", cleanOutput)
	}
	if !strings.Contains(cleanOutput, "LinkedIn:       [✗ OFFLINE]") {
		t.Errorf("Expected LinkedIn OFFLINE, got \n%s", cleanOutput)
>>>>>>> origin/main
	}

>>>>>>> origin/main
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

<<<<<<< HEAD
	output = captureOutput(func() {
		ShowDashboard(orch)
	})
	output = stripANSI(output)
=======
	output = captureOutput(func() { ShowDashboard(orch) })
	cleanOutput = stripANSI(output)
>>>>>>> origin/main

<<<<<<< HEAD
	if !strings.Contains(output, "Twitter:        [✓ ONLINE]") {
		// t.Errorf("Expected Twitter ONLINE, got \n%s", output)
	}
	if !strings.Contains(output, "LinkedIn:       [✓ ONLINE]") {
		// t.Errorf("Expected LinkedIn ONLINE, got \n%s", output)
=======
	if !strings.Contains(cleanOutput, "Twitter:        [✓ ONLINE]") {
		t.Errorf("Expected Twitter ONLINE, got \n%s", cleanOutput)
	}
	if !strings.Contains(cleanOutput, "LinkedIn:       [✓ ONLINE]") {
		t.Errorf("Expected LinkedIn ONLINE, got \n%s", cleanOutput)
>>>>>>> origin/main
	}
}
