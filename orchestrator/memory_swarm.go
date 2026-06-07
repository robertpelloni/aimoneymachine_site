package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/url"
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

// Sync initiates the multi-step reconciliation process with peers
func (s *MemorySwarm) Sync() {
	fmt.Println("[Swarm] Initiating automated delta-sync with peers...")

	// Step 1: Broadcast sync intention
	msg := Message{
		ID:        fmt.Sprintf("sync-%d", time.Now().Unix()),
		Source:    "local-node",
		Type:      Event,
		Topic:     "swarm_sync",
		Payload:   "hustle://swarm?action=sync_request",
		Timestamp: time.Now(),
	}

	s.Broker.Broadcast(msg)
}

// HandleSyncRequest is called when a peer asks for reconciliation
func (s *MemorySwarm) HandleSyncRequest(peerID string) {
	fmt.Printf("[Swarm] Handling sync request from %s. Sending Index...\n", peerID)
	s.ProvideIndex(peerID)
}

// ProvideIndex sends a list of ID:Checksum pairs to a peer
func (s *MemorySwarm) ProvideIndex(peerID string) {
	var indexParts []string
	// Index L2 and L3
	for _, e := range s.Orchestrator.L2.Entries {
		indexParts = append(indexParts, fmt.Sprintf("%s:%s", e.ID, e.Checksum()))
	}
	for _, e := range s.Orchestrator.L3.Entries {
		indexParts = append(indexParts, fmt.Sprintf("%s:%s", e.ID, e.Checksum()))
	}

	payload := strings.Join(indexParts, ",")
	msg := Message{
		ID:        fmt.Sprintf("idx-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Response,
		Payload:   fmt.Sprintf("hustle://swarm?action=provide_index&peer_id=local-node&data=%s", url.QueryEscape(payload)),
		Timestamp: time.Now(),
	}
	s.Broker.Route(msg)
}

// ReconcileIndex compares a peer's index with local and requests missing entries
func (s *MemorySwarm) ReconcileIndex(peerID, indexData string) {
	fmt.Printf("[Swarm] Reconciling index from %s...\n", peerID)

	remoteEntries := strings.Split(indexData, ",")
	for _, part := range remoteEntries {
		kv := strings.Split(part, ":")
		if len(kv) != 2 { continue }

		id := kv[0]
		checksum := kv[1]

		// Check L2/L3 for existence and checksum match
		localEntry, found := s.Orchestrator.L2.Get(id)
		if !found {
			localEntry, found = s.Orchestrator.L3.Get(id)
		}

		if !found || localEntry.Checksum() != checksum {
			fmt.Printf("[Swarm] Delta detected for %s. Requesting entry...\n", id)
			s.RequestEntry(peerID, id)
		}
	}
}

// RequestEntry asks a peer for a specific memory entry
func (s *MemorySwarm) RequestEntry(peerID, entryID string) {
	msg := Message{
		ID:        fmt.Sprintf("req-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Query,
		Payload:   fmt.Sprintf("hustle://swarm?action=request_entry&peer_id=local-node&id=%s", url.QueryEscape(entryID)),
		Timestamp: time.Now(),
	}
	s.Broker.Route(msg)
}

// ProvideEntry sends a specific entry to a peer
func (s *MemorySwarm) ProvideEntry(peerID, entryID string) {
	fmt.Printf("[Swarm] Providing entry %s to %s\n", entryID, peerID)

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
		Payload:   fmt.Sprintf("hustle://swarm?action=provide_entry&peer_id=local-node&id=%s&content=%s", url.QueryEscape(entry.ID), url.QueryEscape(entry.Content)),
		Timestamp: time.Now(),
	}

	s.Broker.Route(msg)
}

// AggregateStatus requests status from all peers
func (s *MemorySwarm) AggregateStatus() {
	fmt.Println("[Swarm] Aggregating mesh-wide status and profit...")
	msg := Message{
		ID:        fmt.Sprintf("agg-%d", time.Now().Unix()),
		Source:    "local-node",
		Type:      Query,
		Topic:     "swarm_aggregate",
		Payload:   "hustle://swarm?action=status_request",
		Timestamp: time.Now(),
	}
	s.Broker.Broadcast(msg)
}

// ProvideStatus sends local status to a peer
func (s *MemorySwarm) ProvideStatus(peerID string) {
	status := map[string]interface{}{
		"peer_id": "local-node",
		"profit":  s.Orchestrator.Ledger.Profit(),
		"revenue": s.Orchestrator.Ledger.TotalRevenue(),
		"status":  "Active",
	}
	data, _ := json.Marshal(status)

	msg := Message{
		ID:        fmt.Sprintf("stat-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Response,
		Payload:   fmt.Sprintf("hustle://swarm?action=provide_status&peer_id=local-node&data=%s", url.QueryEscape(string(data))),
		Timestamp: time.Now(),
	}
	s.Broker.Route(msg)
}

// HandleStatusResponse logs the received peer status
func (s *MemorySwarm) HandleStatusResponse(peerID, data string) {
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		fmt.Printf("[Swarm] Failed to parse status from %s: %v\n", peerID, err)
		return
	}

	content := fmt.Sprintf("Mesh Peer %s Status: %v, Profit: $%.2f", peerID, status["status"], status["profit"])
	s.Orchestrator.L1.Add(MemoryEntry{
		ID:        fmt.Sprintf("mesh-stat-%s-%d", peerID, time.Now().Unix()),
		Content:   content,
		Timestamp: time.Now(),
		Tags:      []string{"swarm", "mesh_status", "profit"},
	})
	fmt.Printf("[Swarm] Ingested status from %s\n", peerID)
}
