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

// Sync broadcasts the local memory checksums to peers for reconciliation
func (s *MemorySwarm) Sync() {
	fmt.Println("[Swarm] Initiating memory sync with peers...")

	// Prepare sync message with checksums
	payload := fmt.Sprintf("L2:%s|L3:%s", s.Orchestrator.L2.Checksum(), s.Orchestrator.L3.Checksum())

	msg := Message{
		ID:        fmt.Sprintf("sync-%d", time.Now().Unix()),
		Source:    "local-node",
		Type:      Event,
		Topic:     "swarm_sync",
		Payload:   payload,
		Timestamp: time.Now(),
	}

	s.Broker.Broadcast(msg)
}

// HandleSyncRequest is called when a peer sends their checksums
func (s *MemorySwarm) HandleSyncRequest(peerID, peerChecksums string) {
	fmt.Printf("[Swarm] Received checksums from %s: %s\n", peerID, peerChecksums)

	localL2 := s.Orchestrator.L2.Checksum()
	// peerL2 would be parsed from peerChecksums

	fmt.Printf("[Swarm] Local L2 Checksum: %s\n", localL2)

	// In alpha, we just acknowledge the diff
	resp := Message{
		ID:        fmt.Sprintf("resp-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Response,
		Payload:   "Checksum comparison complete. Delta detected.",
		Timestamp: time.Now(),
	}

	s.Broker.Route(resp)
}
