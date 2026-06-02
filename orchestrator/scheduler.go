package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
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
}

func NewScheduler(orch *Orchestrator) *Scheduler {
	return &Scheduler{
		Orchestrator: orch,
		Tasks:        make([]*Task, 0),
	}
}

func (s *Scheduler) ReevaluateStrategy(recommendation string) {
	fmt.Printf("[Scheduler] Self-Evolving: Re-evaluating strategy based on: %s\n", recommendation)

	for _, task := range s.Tasks {
		// Example: If Trading is the best performer, increase its frequency
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

func (s *Scheduler) Register(name string, interval time.Duration, fn func(orch *Orchestrator) error) {
	s.Tasks = append(s.Tasks, &Task{
		Name:     name,
		Interval: interval,
		Execute:  fn,
	})
	fmt.Printf("Task registered: %s (Interval: %v)\n", name, interval)
}

func (s *Scheduler) SaveState(filepath string) error {
	data, err := json.MarshalIndent(s.Tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

func (s *Scheduler) LoadState(filepath string) error {
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

	// Update existing registered tasks with their LastRun time
	for _, loaded := range loadedTasks {
		for _, existing := range s.Tasks {
			if existing.Name == loaded.Name {
				existing.LastRun = loaded.LastRun
			}
		}
	}
	return nil
}

func (s *Scheduler) Start() {
	fmt.Println("Starting Task Scheduler...")
	for {
		for _, task := range s.Tasks {
			if time.Since(task.LastRun) >= task.Interval {
				fmt.Printf("Running task: %s\n", task.Name)
				err := task.Execute(s.Orchestrator)
				if err != nil {
					fmt.Printf("Task %s failed: %v\n", task.Name, err)
				}
				task.LastRun = time.Now()
				// Auto-save state after each task run
				s.SaveState("tasks.json")
			}
		}
		time.Sleep(1 * time.Second)
	}
}
