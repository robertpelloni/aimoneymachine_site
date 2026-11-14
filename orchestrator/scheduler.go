package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
	LastRun  time.Time     `json:"last_run"`
	Execute  func(orch *Orchestrator) error `json:"-"`
	// notBefore is the boot-stagger deadline: the task is not due before
	// this time, however old its LastRun is. Set once by Start.
	notBefore time.Time
}

type Scheduler struct {
	Orchestrator *Orchestrator
	Tasks        []*Task
	// KnownChain, when set, reports whether a chain name exists in the
	// ChainManager. LoadState uses it to drop persisted DiscoveredChain
	// tasks whose chain is gone (e.g. chains.json was replaced), instead
	// of re-registering orphans that fail on every cycle forever.
	KnownChain func(name string) bool
	// BootStagger spreads first executions after Start: task i is not
	// due until Start + i*BootStagger. Zero means the default (20s).
	// Without it, every restart (i.e. every deploy) fires all overdue
	// tasks in the same second — an LLM/API stampede that contributed
	// to the 2026-11-14 overload timeouts.
	BootStagger time.Duration
	// MaxConcurrent caps how many tasks execute at once. Zero means
	// the default (3). Before v1.9.15 the loop was strictly serial, so
	// one 7-step LLM chain starved every minute-interval trading task
	// behind it; shared state (Ledger, L1/L2/L3) is now mutex-guarded.
	MaxConcurrent int
	mu            sync.Mutex
	running       map[string]bool
	done          chan struct{}
	stopOnce      sync.Once
}

func NewScheduler(orch *Orchestrator) *Scheduler {
	return &Scheduler{
		Orchestrator: orch,
		Tasks:        make([]*Task, 0),
		running:      make(map[string]bool),
		done:         make(chan struct{}),
	}
}

// Stop ends the Start loop (in-flight tasks finish on their own).
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() { close(s.done) })
}

func (s *Scheduler) maxConcurrent() int {
	if s.MaxConcurrent > 0 {
		return s.MaxConcurrent
	}
	return 3
}

func (s *Scheduler) Register(name string, interval time.Duration, fn func(orch *Orchestrator) error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already registered, update if so
	for _, t := range s.Tasks {
		if t.Name == name {
			t.Interval = interval
			t.Execute = fn
			fmt.Printf("Task updated: %s (Interval: %v)\n", name, interval)
			return
		}
	}

	s.Tasks = append(s.Tasks, &Task{
		Name:     name,
		Interval: interval,
		Execute:  fn,
	})
	fmt.Printf("Task registered: %s (Interval: %v)\n", name, interval)
}

func (s *Scheduler) Unregister(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unregisterNoLock(name)
}

func (s *Scheduler) unregisterNoLock(name string) {
	newTasks := make([]*Task, 0)
	for _, t := range s.Tasks {
		if t.Name != name {
			newTasks = append(newTasks, t)
		}
	}
	if len(newTasks) != len(s.Tasks) {
		fmt.Printf("Task unregistered: %s\n", name)
		s.Tasks = newTasks
		s.saveStateNoLock("tasks.json")
	}
}

func (s *Scheduler) SaveState(filepath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveStateNoLock(filepath)
}

func (s *Scheduler) saveStateNoLock(filepath string) error {
	data, err := json.MarshalIndent(s.Tasks, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(filepath, data, 0644)
}

func (s *Scheduler) LoadState(filepath string, protocol *HustleProtocol) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var loadedTasks []*Task
	if err := json.Unmarshal(data, &loadedTasks); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, loaded := range loadedTasks {
		found := false
		for _, existing := range s.Tasks {
			if existing.Name == loaded.Name {
				existing.LastRun = loaded.LastRun
				existing.Interval = loaded.Interval
				found = true
				break
			}
		}
		// Re-register discovered chains
		if !found && strings.HasPrefix(loaded.Name, "DiscoveredChain:") {
			chainName := strings.TrimPrefix(loaded.Name, "DiscoveredChain:")
			if s.KnownChain != nil && !s.KnownChain(chainName) {
				fmt.Printf("[Scheduler] Dropping orphaned task %s: chain %q not found\n", loaded.Name, chainName)
				continue
			}
			s.Tasks = append(s.Tasks, &Task{
				Name:     loaded.Name,
				Interval: loaded.Interval,
				LastRun:  loaded.LastRun,
				Execute: func(o *Orchestrator) error {
					return protocol.HandleURI(fmt.Sprintf("hustle://chain?name=%s", chainName))
				},
			})
		}
	}
	return nil
}

func (s *Scheduler) ReevaluateStrategy(recommendation string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("[Scheduler] Self-Evolving: Re-evaluating strategy based on: %s\n", recommendation)

	for _, task := range s.Tasks {
		if strings.Contains(strings.ToLower(recommendation), strings.ToLower(task.Name)) {
			oldInterval := task.Interval
			task.Interval = task.Interval / 2
			if task.Interval < 1*time.Minute {
				task.Interval = 1 * time.Minute
			}
			fmt.Printf("[Scheduler] Accelerated %s task: %v -> %v\n", task.Name, oldInterval, task.Interval)
		}
	}
}

// taskDue reports whether task may run at now: past its boot-stagger
// deadline and at least Interval since its last run.
func taskDue(task *Task, now time.Time) bool {
	if now.Before(task.notBefore) {
		return false
	}
	return now.Sub(task.LastRun) >= task.Interval
}

func (s *Scheduler) Start() {
	fmt.Println("Starting Task Scheduler...")
	// Boot stagger: assign each registered task a first-eligible time so
	// a restart does not fire everything at once (see BootStagger).
	stagger := s.BootStagger
	if stagger == 0 {
		stagger = 20 * time.Second
	}
	base := time.Now()
	s.mu.Lock()
	for i, task := range s.Tasks {
		task.notBefore = base.Add(time.Duration(i) * stagger)
	}
	if len(s.Tasks) > 1 {
		fmt.Printf("[Scheduler] Boot stagger: %d tasks spread over ~%v\n", len(s.Tasks), time.Duration(len(s.Tasks)-1)*stagger)
	}
	s.mu.Unlock()

	sem := make(chan struct{}, s.maxConcurrent())
	for {
		select {
		case <-s.done:
			fmt.Println("[Scheduler] Stopped.")
			return
		default:
		}

		// Check for ROI corrections in memory
		s.checkROICorrections()

		// Update Orchestrator metadata for UI
		s.updateMetadata()

		// Dispatch every due task concurrently (bounded by sem). A task
		// never overlaps itself: `running` is only cleared on completion,
		// when LastRun is also stamped.
		s.mu.Lock()
		for _, task := range s.Tasks {
			if s.running[task.Name] || !taskDue(task, time.Now()) {
				continue
			}
			select {
			case sem <- struct{}{}:
			default:
				continue // all slots busy; retry next tick
			}
			s.running[task.Name] = true
			fmt.Printf("Running task: %s\n", task.Name)
			go s.executeTask(task, sem)
		}
		s.mu.Unlock()

		select {
		case <-s.done:
			fmt.Println("[Scheduler] Stopped.")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

// executeTask runs one task and does completion bookkeeping under lock.
func (s *Scheduler) executeTask(task *Task, sem chan struct{}) {
	defer func() { <-sem }()
	start := time.Now()
	err := task.Execute(s.Orchestrator)
	duration := time.Since(start)

	status := "success"
	msg := ""
	if err != nil {
		fmt.Printf("Task %s failed: %v\n", task.Name, err)
		status = "failed"
		msg = err.Error()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Orchestrator.DB != nil {
		s.Orchestrator.DB.LogTaskExecution(task.Name, duration, status, msg)
	}
	task.LastRun = time.Now()
	delete(s.running, task.Name)
	s.saveStateNoLock("tasks.json")
}

func (s *Scheduler) updateMetadata() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var queue []string
	// Find next 3 tasks by time remaining
	type taskWithTime struct {
		name      string
		remaining time.Duration
	}
	var sorted []taskWithTime
	for _, t := range s.Tasks {
		rem := t.Interval - time.Since(t.LastRun)
		if rem < 0 {
			rem = 0
		}
		sorted = append(sorted, taskWithTime{t.Name, rem})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].remaining < sorted[j].remaining
	})

	for i := 0; i < len(sorted) && i < 3; i++ {
		queue = append(queue, fmt.Sprintf("%s (%v)", sorted[i].name, sorted[i].remaining.Round(time.Second)))
	}
	s.Orchestrator.TaskQueue = queue
}

func (s *Scheduler) checkROICorrections() {
	corrections := s.Orchestrator.L1.Search("Wealth Preservation Action")
	if len(corrections) == 0 { return }

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range corrections {
		// e.Content looks like: "Wealth Preservation Action: Requesting termination of underperforming hustles. Reason: ... Underperforming hustle detected: [Name] ..."
		// We need to be careful to only match the task name if it's explicitly marked for termination.
		for _, task := range s.Tasks {
			// Search for "Underperforming hustle detected: [TaskName]" or "[TaskName] (Deficit: ..."
			if strings.Contains(e.Content, "Underperforming hustle detected: "+task.Name) ||
			   strings.Contains(e.Content, "terminate "+task.Name) {
				fmt.Printf("[Scheduler] Executing Wealth Preservation: Unregistering %s\n", task.Name)
				s.unregisterNoLock(task.Name)
			}
		}
	}
}
