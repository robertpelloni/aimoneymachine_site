package orchestrator

import (
	"testing"
	"net/url"
)

func TestE2ELuxuryRelease(t *testing.T) {
	orch := NewOrchestrator()
	protocol := NewHustleProtocol()
	mgr := NewChainManager(orch, protocol)
	disc := NewChainDiscoverer(orch, mgr)
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)
	healer := NewHealer(orch)
	scheduler := NewScheduler(orch)

	// Mock protocol registrations
	protocol.Register("research", func(p url.Values) error { return nil })
	protocol.Register("trading", func(p url.Values) error { return nil })

	t.Log("STEP 1: Autonomous Discovery")
	discovered, err := disc.Discover()
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	t.Log("STEP 2: Ledger Simulation (High ROI)")
	orch.Ledger.Add(Transaction{Amount: 500, Type: Revenue, Hustle: discovered.Name})

	t.Log("STEP 3: Mesh Scaling (Status Aggregation)")
	swarm.HandleStatusResponse("peer-remote", `{"peer_id": "peer-remote", "profit": 1500.0, "status": "Active"}`)

	t.Log("STEP 4: Wealth Preservation Audit (Deficit Detection)")
	scheduler.Register("BadTask", 1, func(o *Orchestrator) error { return nil })
	orch.Ledger.Add(Transaction{Amount: 1000, Type: Expense, Hustle: "BadTask", Note: "Loss"})

	healer.WealthPreservation()

	t.Log("STEP 5: Closed-Loop Termination")
	scheduler.checkROICorrections()

	taskFound := false
	for _, task := range scheduler.Tasks {
		if task.Name == "BadTask" {
			taskFound = true
			break
		}
	}
	if taskFound {
		t.Error("Wealth Preservation loop failed: BadTask was not unregistered from Scheduler after deficit")
	}

	t.Log("STEP 6: System Visibility Verification")
	meshEntries := orch.L1.Search("Mesh Peer")
	if len(meshEntries) == 0 {
		t.Error("Mesh status missing from memory")
	}

	t.Log("--- E2E LUXURY PROTOCOL VERIFIED STABLE ---")
}
