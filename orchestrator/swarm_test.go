package orchestrator

import (
	"strings"
	"testing"
)

func TestSwarmAggregation(t *testing.T) {
	orch := NewOrchestrator()
	broker := NewA2ABroker(orch)
	swarm := NewMemorySwarm(orch, broker)

	// Simulate a status response from a peer
	testData := `{"peer_id": "test-peer", "profit": 500.50, "status": "Active"}`
	swarm.HandleStatusResponse("test-peer", testData)

	if len(orch.L1.Entries) == 0 {
		t.Fatal("Expected L1 entry for mesh status, got 0")
	}

	entry := orch.L1.Entries[0]
	found := false
	for _, tag := range entry.Tags {
		if tag == "mesh_status" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected tag 'mesh_status', got %v", entry.Tags)
	}

	expectedContent := "Mesh Peer test-peer Status: Active, PROFIT: $500.50"
	if !strings.HasPrefix(entry.Content, expectedContent) {
		t.Errorf("Expected content to start with '%s', got '%s'", expectedContent, entry.Content)
	}
}
