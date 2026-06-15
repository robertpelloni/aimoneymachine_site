package orchestrator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMeshClusterRegistrationAndAggregation(t *testing.T) {
	// Node 1: Seed
	orch1 := NewOrchestrator()
	broker1 := NewA2ABroker(orch1)
	swarm1 := NewMemorySwarm(orch1, broker1)

	// Mock Peer A Server
	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer serverA.Close()

	// Mock Peer B Server
	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer serverB.Close()

	// 1. Simulation of registration
	broker1.RegisterPeer("node-a", serverA.URL)
	broker1.RegisterPeer("node-b", serverB.URL)

	if len(broker1.Peers) != 2 {
		t.Fatalf("Seed failed to register 2 peers, got %d", len(broker1.Peers))
	}

	// 2. Simulation of status aggregation
	swarm1.AggregateStatus()

	// 3. Simulate Peer A providing status
	statusA := map[string]interface{}{"peer_id": "node-a", "profit": 500.0, "status": "Active"}
	dataA, _ := json.Marshal(statusA)
	swarm1.HandleStatusResponse("node-a", string(dataA))

	// 4. Simulate Peer B providing status
	statusB := map[string]interface{}{"peer_id": "node-b", "profit": -100.0, "status": "Degraded"}
	dataB, _ := json.Marshal(statusB)
	swarm1.HandleStatusResponse("node-b", string(dataB))

	// 5. Verify Seed Memory
	entries := orch1.L1.Search("Mesh Peer")
	if len(entries) != 2 {
		t.Errorf("Expected 2 aggregated mesh status entries in Seed memory, got %d", len(entries))
	}

	foundA := false
	foundB := false
	for _, e := range entries {
		hasMeshTag := false
		for _, tag := range e.Tags {
			if tag == "mesh_status" {
				hasMeshTag = true
				break
			}
		}
		if hasMeshTag {
			if strings.HasPrefix(e.Content, "Mesh Peer node-a Status: Active, PROFIT: $500.00") { foundA = true }
			if strings.HasPrefix(e.Content, "Mesh Peer node-b Status: Degraded, PROFIT: $-100.00") { foundB = true }
		}
	}

	if !foundA { t.Error("Node A status not found in Seed memory") }
	if !foundB { t.Error("Node B status not found in Seed memory") }

	t.Log("3-Node Mesh Cluster Simulation: SUCCESS")
}
