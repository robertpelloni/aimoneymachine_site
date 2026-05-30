package curation

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type CurationModule struct {
	Orchestrator *orchestrator.Orchestrator
}

func (c *CurationModule) Curate(topic string) error {
	fmt.Printf("Curating content for topic: %s\n", topic)

	// Mock content sourcing
	content := fmt.Sprintf("Top stories on %s: Item 1, Item 2, Item 3.", topic)

	// Use Orchestrator LLM for summarization
	summary, err := c.Orchestrator.LLM.Generate("Summarize this content into a catchy newsletter blurb: " + content)
	if err != nil {
		return err
	}

	fmt.Printf("Newsletter blurb: %s\n", summary)

	// Persist to memory
	c.Orchestrator.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("curation-%d", time.Now().Unix()),
		Content:   summary,
		Timestamp: time.Now(),
		Tags:      []string{"curation", topic},
	})

	return nil
}
