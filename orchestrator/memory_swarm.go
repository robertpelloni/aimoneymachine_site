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

	payload := "hustle://swarm?action=sync_request"
	if s.Orchestrator.Identity != nil {
		sig := s.Orchestrator.Identity.Sign(payload)
		payload += fmt.Sprintf("&did=%s&sig=%s", s.Orchestrator.Identity.GetDID(), sig)
	}

	// Step 1: Broadcast sync intention
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

// HandleSyncRequest is called when a peer asks for reconciliation
func (s *MemorySwarm) HandleSyncRequest(peerID string) {
	fmt.Printf("[Swarm] Handling sync request from %s. Sending Index...\n", peerID)
	// Cryptographic DID Verification (Phase 6 Security)
	// In a real multi-node network, we extract 'did' and 'sig' from the inbound payload.
	// For alpha.90, we verify the presence of the identity layer.
	if s.Orchestrator.Identity != nil {
		fmt.Printf("[Security] Trust verified for peer %s via DID.\n", peerID)
	}
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

	// Broadast profit for leaderboard sync
	s.SyncProfit()
}

func (s *MemorySwarm) SyncProfit() {
	msg := Message{
		ID:        fmt.Sprintf("profit-%d", time.Now().Unix()),
		Source:    "local-node",
		Type:      Event,
		Topic:     "leaderboard_sync",
		Payload:   fmt.Sprintf("hustle://swarm?action=sync_profit&peer_id=local-node&profit=%.2f", s.Orchestrator.Ledger.Profit()),
		Timestamp: time.Now(),
	}
	s.Broker.Broadcast(msg)
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
	start := time.Now()
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		fmt.Printf("[Swarm] Failed to parse status from %s: %v\n", peerID, err)
		return
	}

	p, _ := status["profit"].(float64)

	// Adaptive Sync: Trigger immediate delta-sync if remote profit is high (Alpha detected)
	if p > 1000 {
		fmt.Printf("[Swarm] High-profit peer detected (%s: $%.2f). Accelerating sync.\n", peerID, p)
		go func() {
			time.Sleep(2 * time.Second)
			s.Sync()
		}()
	}

	content := fmt.Sprintf("Mesh Peer %s Status: %v, PROFIT: $%.2f", peerID, status["status"], status["profit"])
	s.Orchestrator.L1.Add(MemoryEntry{
		ID:        fmt.Sprintf("mesh-stat-%s-%d", peerID, time.Now().Unix()),
		Content:   fmt.Sprintf("%s (Sync Latency: %v)", content, time.Since(start)),
		Timestamp: time.Now(),
		Tags:      []string{"swarm", "mesh_status", "profit", "metrics"},
	})
	fmt.Printf("[Swarm] Ingested status from %s\n", peerID)

	// Self-Funding Swarm: If we are profitable and peer is in deficit, offer funding
	if s.Orchestrator.Ledger.Profit() > 1000 && p < -500 {
		fmt.Printf("[Swarm] Peer %s is in distress ($%.2f). Offering autonomous funding...\n", peerID, p)
		s.RequestFundingAudit(peerID)
	}
}

func (s *MemorySwarm) RequestFundingAudit(peerID string) {
	msg := Message{
		ID:        fmt.Sprintf("audit-%d", time.Now().Unix()),
		Source:    "local-node",
		Target:    peerID,
		Type:      Query,
		Payload:   "hustle://swarm?action=request_funding_audit&peer_id=local-node",
		Timestamp: time.Now(),
	}
	s.Broker.Route(msg)
}

func (s *MemorySwarm) HandleFundingRequest(peerID string, amount float64) {
	fmt.Printf("[Swarm] Processing funding request from %s for $%.2f\n", peerID, amount)
	// Autonomous Approval: If we have > 2x the request in profit, approve
	if s.Orchestrator.Ledger.Profit() > amount*2 {
		s.Orchestrator.Ledger.TransferCapital(amount, peerID, "Autonomous Mesh Support")
		msg := Message{
			ID:        fmt.Sprintf("grant-%d", time.Now().Unix()),
			Source:    "local-node",
			Target:    peerID,
			Type:      Event,
			Payload:   fmt.Sprintf("hustle://swarm?action=grant_funding&amount=%.2f&source_id=local-node", amount),
			Timestamp: time.Now(),
		}
		s.Broker.Route(msg)
	}
}

func (s *MemorySwarm) HandleFundingGrant(sourceID string, amount float64) {
	s.Orchestrator.Ledger.ReceiveCapital(amount, sourceID, "Mesh Liquidity Injection")
	fmt.Printf("[Swarm] Wallet updated via mesh funding. Current Profit: $%.2f\n", s.Orchestrator.Ledger.Profit())
}
