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

// LLMs invent parameter values too (platform=YouTube+Shorts observed
// live): social steps must name a real platform or never register.
func TestValidateChainStepsRejectsInventedPlatform(t *testing.T) {
	p := testProtocol()
	p.Register("social", func(url.Values) error { return nil })
	if err := ValidateChainSteps([]string{"hustle://social?platform=YouTube+Shorts&topic=x"}, p); err == nil {
		t.Error("invented platform YouTube+Shorts accepted")
	}
	if err := ValidateChainSteps([]string{"hustle://social?platform=Bluesky&topic=x"}, p); err != nil {
		t.Errorf("Bluesky rejected: %v", err)
	}
	if err := ValidateChainSteps([]string{"hustle://social?topic=x"}, p); err != nil {
		t.Errorf("platform-less social step rejected: %v", err)
	}
}

// Closed action sets get the same invented-value treatment, and a
// chain step naming a nonexistent chain must be rejected in Discover.
func TestValidateChainStepsRejectsInventedActions(t *testing.T) {
	p := testProtocol()
	for _, m := range []string{"calendar", "publisher", "affiliate", "swarm", "media", "chain"} {
		p.Register(m, func(url.Values) error { return nil })
	}
	bad := []string{
		"hustle://calendar?action=publish",
		"hustle://publisher?action=post",
		"hustle://affiliate?action=sell",
		"hustle://swarm?action=explode",
		"hustle://media?action=movie",
	}
	for _, s := range bad {
		if err := ValidateChainSteps([]string{s}, p); err == nil {
			t.Errorf("invented action accepted: %s", s)
		}
	}
	good := []string{
		"hustle://calendar?action=schedule",
		"hustle://calendar?action=schedule_week",
		"hustle://publisher?action=wordpress",
		"hustle://publisher?action=distribute",
		"hustle://affiliate?action=batch_insert",
		"hustle://swarm?action=set_goal",
		"hustle://media?action=shorts",
	}
	for _, s := range good {
		if err := ValidateChainSteps([]string{s}, p); err != nil {
			t.Errorf("valid step rejected: %s (%v)", s, err)
		}
	}
}

func TestDiscoverRejectsDanglingChainReference(t *testing.T) {
	t.Chdir(t.TempDir())
	orch := NewOrchestrator()
	orch.LLM = &MockLLM{GenerateFunc: func(string) (string, error) {
		return `{"name":"ref_chain","description":"x","steps":["hustle://chain?name=does_not_exist"],"interval_minutes":30}`, nil
	}}
	p := testProtocol()
	p.Register("chain", func(url.Values) error { return nil })
	mgr := NewChainManager(orch, p)
	d := NewChainDiscoverer(orch, mgr)
	if _, err := d.Discover(); err == nil || !strings.Contains(err.Error(), "unknown chain") {
		t.Fatalf("Discover err = %v, want unknown-chain rejection", err)
	}
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

// A well-formed chain JSON with zero steps must also get the strict
// retry (observed live: "Evergreen Content Distribution" with no steps).
func TestDiscoverRetriesWhenStepsEmpty(t *testing.T) {
	t.Chdir(t.TempDir())
	orch := NewOrchestrator()
	orch.LLM = &MockLLM{GenerateFunc: func(prompt string) (string, error) {
		if strings.Contains(prompt, "Respond ONLY with valid JSON") {
			return `{"name":"retry_chain","description":"x","steps":["hustle://research?query=AI"],"interval_minutes":30}`, nil
		}
		return `{"name":"empty_chain","description":"x","steps":[],"interval_minutes":30}`, nil
	}}
	mgr := NewChainManager(orch, testProtocol())
	d := NewChainDiscoverer(orch, mgr)
	got, err := d.Discover()
	if err != nil {
		t.Fatalf("Discover with empty-steps retry: %v", err)
	}
	if got.Name != "retry_chain" {
		t.Fatalf("name = %q, want retry_chain", got.Name)
	}
}

// Prod (2026-11-17): discovery re-proposed the existing static chain
// "luxury_passive_income_engine" and the run was logged as a task failure
// ("chain ... already exists"). A duplicate name must be a benign no-op,
// and the prompt must tell the model which names are taken so it does not
// burn an 18-minute LLM cycle re-proposing one.
func TestDiscoverDuplicateNameIsBenign(t *testing.T) {
	t.Chdir(t.TempDir())
	orch := NewOrchestrator()
	var gotPrompt string
	orch.LLM = &MockLLM{GenerateFunc: func(prompt string) (string, error) {
		gotPrompt = prompt
		return `{"name":"luxury_passive_income_engine","description":"x","steps":["hustle://research?query=AI"],"interval_minutes":30}`, nil
	}}
	mgr := NewChainManager(orch, testProtocol())
	mgr.Register(&Chain{Name: "luxury_passive_income_engine"})
	d := NewChainDiscoverer(orch, mgr)
	got, err := d.Discover()
	if err != nil {
		t.Fatalf("Discover duplicate err = %v, want nil (benign no-op)", err)
	}
	if got != nil {
		t.Fatalf("Discover duplicate = %+v, want nil chain", got)
	}
	if !strings.Contains(gotPrompt, "luxury_passive_income_engine") {
		t.Error("discovery prompt does not list existing chain names")
	}
	if len(mgr.Chains) != 1 {
		t.Fatalf("chains = %d, want 1 (duplicate not re-registered)", len(mgr.Chains))
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
