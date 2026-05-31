package orchestrator

import (
	"fmt"
	"strings"
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

	fmt.Printf("[Swarm] Local L2 Checksum: %s\n", localL2)

	// If checksums mismatch, we would normally request a list of IDs.
	// For alpha, we simulate asking for a specific missing entry.
	if peerChecksums != "" && !strings.Contains(peerChecksums, localL2) {
		fmt.Println("[Swarm] Checksum mismatch detected. Requesting missing context...")
		s.RequestEntry(peerID, "latest-research-context")
	}

	resp := Message{
		ID:        fmt.Sprintf("resp-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Response,
		Payload:   "Checksum comparison complete.",
		Timestamp: time.Now(),
	}

	s.Broker.Route(resp)
}

// RequestEntry asks a peer for a specific memory entry
func (s *MemorySwarm) RequestEntry(peerID, entryID string) {
	fmt.Printf("[Swarm] Requesting entry %s from %s\n", entryID, peerID)

	msg := Message{
		ID:        fmt.Sprintf("req-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Query,
		Payload:   fmt.Sprintf("hustle://swarm?action=request_entry&id=%s", entryID),
		Timestamp: time.Now(),
	}

	s.Broker.Route(msg)
}

// ProvideEntry sends a specific entry to a peer
func (s *MemorySwarm) ProvideEntry(peerID, entryID string) {
	fmt.Printf("[Swarm] Providing entry %s to %s\n", entryID, peerID)

	// Find entry in L2/L3
	entry, ok := s.Orchestrator.L2.Get(entryID)
	if !ok {
		entry, ok = s.Orchestrator.L3.Get(entryID)
	}

	if !ok { return }

	msg := Message{
		ID:        fmt.Sprintf("prov-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Response,
		Payload:   fmt.Sprintf("hustle://swarm?action=provide_entry&id=%s&content=%s", entry.ID, entry.Content),
		Timestamp: time.Now(),
	}

	s.Broker.Route(msg)
}
