package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"sync"
)

type Task struct {
	Name     string        `json:"name"`
	Interval time.Duration `json:"interval"`
	LastRun  time.Time     `json:"last_run"`
	Execute  func(orch *Orchestrator) error `json:"-"`
}

type Scheduler struct {
	Orchestrator *Orchestrator
	Tasks        []*Task
	mu           sync.Mutex
}

func NewScheduler(orch *Orchestrator) *Scheduler {
	return &Scheduler{
		Orchestrator: orch,
		Tasks:        make([]*Task, 0),
	}
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
	return os.WriteFile(filepath, data, 0644)
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

func (s *Scheduler) Start() {
	fmt.Println("Starting Task Scheduler...")
	for {
		// Check for ROI corrections in memory
		s.checkROICorrections()

		s.mu.Lock()
		for _, task := range s.Tasks {
			if time.Since(task.LastRun) >= task.Interval {
				fmt.Printf("Running task: %s\n", task.Name)
				s.mu.Unlock()
				err := task.Execute(s.Orchestrator)
				s.mu.Lock()

				if err != nil {
					fmt.Printf("Task %s failed: %v\n", task.Name, err)
				}
				task.LastRun = time.Now()
				s.saveStateNoLock("tasks.json")
			}
		}
		s.mu.Unlock()
		time.Sleep(1 * time.Second)
	}
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
