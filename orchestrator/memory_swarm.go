package orchestrator

import (
	"fmt"
	"time"
)

// MemorySwarm manages federated memory synchronization across the mesh
type MemorySwarm struct {
	Orchestrator *Orchestrator
	Broker       *A2ABroker
}

func NewMemorySwarm(orch *Orchestrator, broker *A2ABroker) *MemorySwarm {
	return &MemorySwarm{
		Orchestrator: orch,
		Broker:       broker,
	}
}

// Sync broadcasts the local L2/L3 memory index to peers for reconciliation
func (s *MemorySwarm) Sync() {
	fmt.Println("[Swarm] Initiating memory sync with peers...")

	// Prepare sync message
	msg := Message{
		ID:        fmt.Sprintf("sync-%d", time.Now().Unix()),
		Source:    "local-node",
		Type:      Event,
		Payload:   "hustle://swarm?action=sync_request",
		Timestamp: time.Now(),
	}

	s.Broker.Broadcast(msg)
}

// HandleSyncRequest is called when a peer asks for our memory state
func (s *MemorySwarm) HandleSyncRequest(peerID string) {
	fmt.Printf("[Swarm] Received sync request from %s\n", peerID)

	// In a real swarm, we'd send hashes or a compressed index.
	// For alpha, we just acknowledge the request.
	resp := Message{
		ID:        fmt.Sprintf("resp-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Response,
		Payload:   "Memory index ready for reconciliation",
		Timestamp: time.Now(),
	}

	s.Broker.Route(resp)
}
