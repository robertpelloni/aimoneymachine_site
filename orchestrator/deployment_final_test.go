package orchestrator

import (
	"testing"
)

func TestDeploymentFinal(t *testing.T) {
	orch := NewOrchestrator()
	protocol := NewHustleProtocol()
	mgr := NewChainManager(orch, protocol)
	disc := NewChainDiscoverer(orch, mgr)
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)
	healer := NewHealer(orch)
	scheduler := NewScheduler(orch)

	t.Log("1. Verifying Monorepo Lifecycle Integration")

	t.Log("STEP A: Discovery")
	disc.Discover()
	if len(mgr.Chains) == 0 {
		t.Error("Discovery failed to register chains")
	}

	t.Log("STEP B: Mesh Swarm Connectivity")
	swarm.HandleStatusResponse("final-peer", `{"peer_id": "final-peer", "profit": 333.33, "status": "Active"}`)
	if len(orch.L1.Search("Mesh Peer")) == 0 {
		t.Error("Mesh status aggregation failed")
	}

	t.Log("STEP C: Capital Protection (Wealth Preservation)")
	scheduler.Register("LeakingHustle", 1, func(o *Orchestrator) error { return nil })
	orch.Ledger.Add(Transaction{Amount: 9999, Type: Expense, Hustle: "LeakingHustle"})
	healer.WealthPreservation()
	scheduler.checkROICorrections()

	leakingTaskFound := false
	for _, task := range scheduler.Tasks {
		if task.Name == "LeakingHustle" { leakingTaskFound = true }
	}
	if leakingTaskFound {
		t.Error("Wealth preservation loop failed to stop leaking hustle")
	}

	t.Log("--- DEPLOYMENT FINAL INTEGRATION PASSED ---")
}
