package orchestrator

import (
	"testing"
	"strings"
)

func TestGrandLuxuryProtocol(t *testing.T) {
	orch := NewOrchestrator()
	protocol := NewHustleProtocol()
	mgr := NewChainManager(orch, protocol)
	disc := NewChainDiscoverer(orch, mgr)
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)
	healer := NewHealer(orch)

	t.Log("1. Discovering Luxury Chain...")
	discovered, err := disc.Discover()
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	t.Log("2. Simulating Successful Execution & Revenue...")
	orch.Ledger.Add(Transaction{Amount: 1000, Type: Revenue, Hustle: discovered.Name, Note: "Luxury execution success"})

	t.Log("3. Simulating Mesh Aggregation...")
	swarm.HandleStatusResponse("peer-alpha", `{"peer_id": "peer-alpha", "profit": 250.0, "status": "Active"}`)

	t.Log("4. Running Wealth Preservation Audit...")
	healer.WealthPreservation()
	// Should be healthy initially

	t.Log("5. Simulating Failure & ROI Correction...")
	orch.Ledger.Add(Transaction{Amount: 2000, Type: Expense, Hustle: "FailingHustle", Note: "Heavy loss"})
	healer.WealthPreservation()

	foundCorrection := false
	for _, e := range orch.L1.Entries {
		if strings.Contains(e.Content, "Terminate immediately") {
			foundCorrection = true
			break
		}
	}
	if !foundCorrection {
		t.Error("Expected ROI correction (termination) for underperforming hustle, but none found.")
	}

	t.Log("6. Verifying Dashboard Aggregation...")
	// We need to check if the tag is correctly applied
	foundMesh := false
	for _, e := range orch.L1.Entries {
		for _, tag := range e.Tags {
			if tag == "mesh_status" {
				foundMesh = true
				break
			}
		}
	}
	if !foundMesh {
		t.Error("Expected mesh status entries in memory for dashboard.")
	}

	t.Log("Grand Luxury Protocol Verification: SUCCESS")
}
