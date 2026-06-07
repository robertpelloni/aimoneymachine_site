package orchestrator

import (
	"testing"
	"net/http"
	"net/http/httptest"
)

func TestMeshAggregationProtocol(t *testing.T) {
	orch := NewOrchestrator()
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)

	// 1. Mock a peer server
	peerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer peerServer.Close()

	broker.RegisterPeer("peer-1", peerServer.URL)

	// 2. Trigger aggregation
	swarm.AggregateStatus()

	// 3. Simulate peer sending its status back
	statusData := `{"peer_id": "peer-1", "profit": 123.45, "status": "Active"}`
	swarm.HandleStatusResponse("peer-1", statusData)

	// 4. Verify entry in memory
	// Search uses Content, and Content contains "Mesh Peer"
	entries := orch.L1.Search("Mesh Peer")
	if len(entries) == 0 {
		t.Fatal("Expected aggregated status entry in L1 memory")
	}

	found := false
	for _, e := range entries {
		hasProfit := false
		hasSwarm := false
		for _, tag := range e.Tags {
			if tag == "profit" { hasProfit = true }
			if tag == "swarm" { hasSwarm = true }
		}
		if hasProfit && hasSwarm {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected tags 'profit' and 'swarm' on aggregated entry")
	}
}
