package orchestrator

import (
	"sync"
	"testing"
	"time"
)

// Concurrency contract for the parallel scheduler (v1.9.15): a slow task
// must not starve a fast one, shared state (ledger, memory tiers, chain
// registry) must survive concurrent hammering under -race, and a task
// must never overlap itself.
func TestConcurrentStateAccess(t *testing.T) {
	orch := NewOrchestrator()
	mgr := NewChainManager(orch, NewHustleProtocol())
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				orch.L1.Add(MemoryEntry{ID: "x", Content: "hello world", Tags: []string{"t"}})
				orch.L2.Add(MemoryEntry{ID: "y", Content: "vault entry"})
				_ = orch.L1.Search("hello")
				orch.Ledger.Add(Transaction{Amount: 1, Type: Revenue, Hustle: "test"})
				_ = orch.Ledger.Profit()
				_ = orch.Ledger.AnalyzeProfitability()
				mgr.Register(&Chain{Name: "c", Steps: []string{"hustle://research"}})
				_ = mgr.Snapshot()
				_ = orch.L1.Len()
			}
		}(g)
	}
	wg.Wait()
	if got := orch.L1.Len(); got != 8*50 {
		t.Errorf("L1 entries = %d, want %d", got, 8*50)
	}
	if got := orch.Ledger.TotalRevenue(); got != 8*50 {
		t.Errorf("revenue = %.0f, want %d", got, 8*50)
	}
}

func TestSchedulerRunsTasksConcurrently(t *testing.T) {
	orch := NewOrchestrator()
	s := NewScheduler(orch)
	s.BootStagger = time.Millisecond
	s.MaxConcurrent = 3

	slowDone := make(chan struct{})
	fastDone := make(chan struct{})
	var onceSlow, onceFast sync.Once
	s.Register("SlowTask", time.Hour, func(*Orchestrator) error {
		time.Sleep(2500 * time.Millisecond)
		onceSlow.Do(func() { close(slowDone) })
		return nil
	})
	s.Register("FastTask", time.Hour, func(*Orchestrator) error {
		onceFast.Do(func() { close(fastDone) })
		return nil
	})
	// Both are due immediately. The dispatch loop ticks every 1s and
	// FastTask's 1ms stagger slot lands on the second tick (~1s), so
	// the proof is: fast completes then, while slow (2.5s) still runs.
	go s.Start()
	defer s.Stop()

	select {
	case <-fastDone:
	case <-time.After(1800 * time.Millisecond):
		t.Fatal("fast task starved behind slow task (scheduler still serial?)")
	}
	select {
	case <-slowDone:
		t.Fatal("slow task finished before fast - no overlap happened")
	default:
	}
	select {
	case <-slowDone:
	case <-time.After(3 * time.Second):
		t.Fatal("slow task never completed")
	}
}
