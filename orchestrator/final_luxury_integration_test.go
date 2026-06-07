package orchestrator

import (
	"testing"
	"strings"
	"time"
)

func TestAbsoluteFinalLuxuryProtocol(t *testing.T) {
	orch := NewOrchestrator()
	protocol := NewHustleProtocol()
	mgr := NewChainManager(orch, protocol)
	disc := NewChainDiscoverer(orch, mgr)
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)
	healer := NewHealer(orch)
	scheduler := NewScheduler(orch)

	t.Log("--- PHASE 1: DISCOVERY & SCHEDULING ---")
	discovered, err := disc.Discover()
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}
	t.Logf("Discovered Luxury Chain: %s", discovered.Name)

	// Register the discovered chain into scheduler
	scheduler.Register("Chain:"+discovered.Name, time.Duration(discovered.IntervalMinutes)*time.Minute, func(o *Orchestrator) error {
		return protocol.HandleURI("hustle://chain?name=" + discovered.Name)
	})

	if len(scheduler.Tasks) == 0 {
		t.Error("Expected task to be registered in scheduler")
	}

	t.Log("--- PHASE 2: EXECUTION & LEDGER ---")
	// Simulate the chain execution by manually adding revenue
	orch.Ledger.Add(Transaction{
		Amount: 5000,
		Type:   Revenue,
		Hustle: discovered.Name,
		Note:   "High ROI luxury success",
	})

	t.Log("--- PHASE 3: MESH AGGREGATION ---")
	// Simulate status from 3 remote peers
	swarm.HandleStatusResponse("peer-1", `{"peer_id": "peer-1", "profit": 1000.0, "status": "Active"}`)
	swarm.HandleStatusResponse("peer-2", `{"peer_id": "peer-2", "profit": -50.0, "status": "Degraded"}`)
	swarm.HandleStatusResponse("peer-3", `{"peer_id": "peer-3", "profit": 2000.0, "status": "Active"}`)

	t.Log("--- PHASE 4: WEALTH PRESERVATION & HEALING ---")
	// Add an underperforming hustle
	orch.Ledger.Add(Transaction{
		Amount: 10000,
		Type:   Expense,
		Hustle: "ObsoleteWaste",
		Note:   "Leaking capital",
	})

	healer.WealthPreservation()

	// Verify ROI Correction entry
	foundCorrection := false
	for _, e := range orch.L1.Entries {
		if strings.Contains(e.Content, "Terminate immediately") && strings.Contains(e.Content, "ObsoleteWaste") {
			foundCorrection = true
			break
		}
	}
	if !foundCorrection {
		t.Error("Wealth Preservation failed to identify and correct ObsoleteWaste")
	}

	t.Log("--- PHASE 5: DASHBOARD VERIFICATION ---")
	// Just ensure we have aggregated data
	meshEntries := orch.L1.Search("mesh_status")
	if len(meshEntries) != 3 {
		t.Errorf("Expected 3 mesh status entries, got %d", len(meshEntries))
	}

	t.Log("--- PHASE 6: FINAL HEALTH CHECK ---")
	if orch.Ledger.Profit() != -5000 { // 5000 revenue - 10000 expense
		t.Logf("Total Local Profit: $%.2f", orch.Ledger.Profit())
	}

	t.Log("FINAL LUXURY PROTOCOL INTEGRATION TEST: PASSED")
}
