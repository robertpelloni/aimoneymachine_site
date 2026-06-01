package curation

import (
	"context"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"strings"
	"time"
)

type CurationModule struct {
	Orchestrator *orchestrator.Orchestrator
	Broker       *orchestrator.A2ABroker
	Fetcher      FeedFetcher
	Feeds        []string
}

func (c *CurationModule) AddFeed(url string) {
	c.Feeds = append(c.Feeds, url)
}

func (c *CurationModule) Curate(topic string) error {
	fmt.Printf("Curating content for topic: %s\n", topic)

	var allItems []FeedItem
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, url := range c.Feeds {
		items, err := c.Fetcher.Fetch(ctx, url)
		if err != nil {
			fmt.Printf("Warning: failed to fetch feed %s: %v\n", url, err)
			continue
		}
		allItems = append(allItems, items...)
	}

	if len(allItems) == 0 {
		return fmt.Errorf("no content found for curation on topic: %s", topic)
	}

	// Filter and synthesis
	var relevantContent []string
	for _, item := range allItems {
		if strings.Contains(strings.ToLower(item.Title), strings.ToLower(topic)) ||
			strings.Contains(strings.ToLower(item.Description), strings.ToLower(topic)) {
			relevantContent = append(relevantContent, fmt.Sprintf("- %s: %s (%s)", item.Title, item.Description, item.Link))
		}
	}

	if len(relevantContent) == 0 {
		// Fallback: just use top 3 items
		for i := 0; i < len(allItems) && i < 3; i++ {
			item := allItems[i]
			relevantContent = append(relevantContent, fmt.Sprintf("- %s: %s (%s)", item.Title, item.Description, item.Link))
		}
	}

	contentToSummarize := strings.Join(relevantContent, "\n")

	// Use Orchestrator LLM for summarization
	summary, err := c.Orchestrator.LLM.Generate("Summarize this curated content into a catchy newsletter blurb for topic " + topic + ":\n\n" + contentToSummarize)
	if err != nil {
		return err
	}

	fmt.Printf("Newsletter blurb: %s\n", summary)

	// Persist to memory
	c.Orchestrator.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("curation-%s-%d", topic, time.Now().Unix()),
		Content:   summary,
		Timestamp: time.Now(),
		Tags:      []string{"curation", topic},
	})

	// Broadcast to mesh
	if c.Broker != nil {
		msg := orchestrator.Message{
			ID:        fmt.Sprintf("curation-%d", time.Now().Unix()),
			Source:    "curation-module",
			Type:      orchestrator.Event,
			Topic:     "content_curation",
			Payload:   summary,
			Timestamp: time.Now(),
		}
		c.Broker.Publish(msg)
		fmt.Println("[Curation] Broadcasted curated blurb to mesh topic: content_curation")
	}

	return nil
}
