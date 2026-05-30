package orchestrator

import (
	"fmt"
	"time"
)

type Task struct {
	Name     string
	Interval time.Duration
	LastRun  time.Time
	Execute  func(orch *Orchestrator) error
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

func (s *Scheduler) Register(name string, interval time.Duration, fn func(orch *Orchestrator) error) {
	s.Tasks = append(s.Tasks, &Task{
		Name:     name,
		Interval: interval,
		Execute:  fn,
	})
	fmt.Printf("Task registered: %s (Interval: %v)\n", name, interval)
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
			}
		}
		time.Sleep(1 * time.Second)
	}
}
