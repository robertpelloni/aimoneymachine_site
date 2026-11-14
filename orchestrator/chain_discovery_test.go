package orchestrator

import (
	"net/url"
	"strings"
	"testing"
)

func testProtocol() *HustleProtocol {
	p := NewHustleProtocol()
	p.Register("research", func(url.Values) error { return nil })
	p.Register("content", func(url.Values) error { return nil })
	return p
}

func TestValidateChainSteps(t *testing.T) {
	p := testProtocol()
	if err := ValidateChainSteps([]string{"hustle://research?query=x", "hustle://content?topic=y&type=blog"}, p); err != nil {
		t.Errorf("valid steps rejected: %v", err)
	}
	for _, bad := range [][]string{
		{},
		{"hustle://step1"},
		{"not-a-uri"},
		{"https://research?query=x"},
		{"hustle://research?query=x", "hustle://nope"},
	} {
		if err := ValidateChainSteps(bad, p); err == nil {
			t.Errorf("steps %v accepted, want rejection", bad)
		}
	}
}

// The historical failure: the LLM parrots the prompt's placeholder steps.
// Discovery must reject them instead of persisting a doomed chain.
func TestDiscoverRejectsPlaceholderSteps(t *testing.T) {
	orch := NewOrchestrator()
	orch.LLM = &MockLLM{GenerateFunc: func(string) (string, error) {
		return `{"name":"luxury_chain_name","description":"x","steps":["hustle://step1","hustle://step2"],"interval_minutes":60}`, nil
	}}
	mgr := NewChainManager(orch, testProtocol())
	d := NewChainDiscoverer(orch, mgr)
	if _, err := d.Discover(); err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("Discover err = %v, want rejection", err)
	}
	if len(mgr.Chains) != 0 {
		t.Fatalf("chains registered: %v, want none", mgr.Chains)
	}
}

// Reasoning-style replies with no JSON ("The user wants me to act as...")
// must trigger exactly one strict-JSON retry, not an outright failure —
// this is the WorkflowDiscovery failure seen on prod (2026-11-14).
func TestDiscoverRetriesWithStrictJSON(t *testing.T) {
	t.Chdir(t.TempDir())
	orch := NewOrchestrator()
	calls := 0
	orch.LLM = &MockLLM{GenerateFunc: func(prompt string) (string, error) {
		calls++
		if strings.Contains(prompt, "Respond ONLY with valid JSON") {
			return `{"name":"retry_chain","description":"x","steps":["hustle://research?query=AI"],"interval_minutes":30}`, nil
		}
		return "The user wants me to act as the Swarm Evolution Engine and propose a workflow of protocol URIs.", nil
	}}
	mgr := NewChainManager(orch, testProtocol())
	d := NewChainDiscoverer(orch, mgr)
	got, err := d.Discover()
	if err != nil {
		t.Fatalf("Discover with retry: %v", err)
	}
	if got.Name != "retry_chain" {
		t.Fatalf("name = %q, want retry_chain", got.Name)
	}
	if calls < 2 {
		t.Fatalf("LLM calls = %d, want >= 2 (initial + strict retry)", calls)
	}
}

func TestDiscoverRegistersValidChain(t *testing.T) {
	t.Chdir(t.TempDir()) // Register saves chains.json in cwd
	orch := NewOrchestrator()
	orch.LLM = &MockLLM{GenerateFunc: func(string) (string, error) {
		return `{"name":"good_chain","description":"x","steps":["hustle://research?query=AI"],"interval_minutes":30}`, nil
	}}
	mgr := NewChainManager(orch, testProtocol())
	d := NewChainDiscoverer(orch, mgr)
	got, err := d.Discover()
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got.Name != "good_chain" {
		t.Fatalf("name = %q", got.Name)
	}
	if _, ok := mgr.Chains["good_chain"]; !ok {
		t.Fatal("valid chain not registered")
	}
}
