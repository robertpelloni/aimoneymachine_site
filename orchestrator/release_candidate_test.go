package orchestrator

import (
	"testing"
)

func TestReleaseCandidateLuxuryProtocol(t *testing.T) {
	orch := NewOrchestrator()
	protocol := NewHustleProtocol()
	mgr := NewChainManager(orch, protocol)
	disc := NewChainDiscoverer(orch, mgr)
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)
	healer := NewHealer(orch)
	scheduler := NewScheduler(orch)

	t.Log("1. Testing Discovery...")
	discovered, err := disc.Discover()
	if err != nil {
		t.Fatalf("Chain discovery failed: %v", err)
	}
	t.Logf("Discovered: %s", discovered.Name)

	t.Log("2. Testing ROI Audit & Preservation Loop...")
	scheduler.Register("BadHustle", 1, func(o *Orchestrator) error { return nil })
	orch.Ledger.Add(Transaction{Amount: 5000, Type: Expense, Hustle: "BadHustle", Note: "Loss"})

	healer.WealthPreservation()
	scheduler.checkROICorrections()

	for _, task := range scheduler.Tasks {
		if task.Name == "BadHustle" {
			t.Error("Wealth preservation failed: BadHustle still active in scheduler")
		}
	}

	t.Log("3. Testing Mesh Scaling Logic...")
	swarm.HandleStatusResponse("test-peer", `{"peer_id": "test-peer", "profit": 1000.0, "status": "Active"}`)
	meshEntries := orch.L1.Search("Mesh Peer")
	if len(meshEntries) == 0 {
		t.Error("Mesh status aggregation failed: No entries found in memory")
	}

	t.Log("--- RELEASE CANDIDATE VERIFIED ---")
}
