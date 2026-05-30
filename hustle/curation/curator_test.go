package curation

import (
	"context"
	"github.com/robertpelloni/hustle/orchestrator"
	"testing"
	"time"
)

func TestCurationPipeline(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	curator := &CurationModule{
		Orchestrator: orch,
		Fetcher:      &MockFetcher{},
		Feeds:        []string{"https://example.com/rss"},
	}

	err := curator.Curate("Technology")
	if err != nil {
		t.Fatalf("Curation failed: %v", err)
	}

	if len(orch.L1.Entries) == 0 {
		t.Error("Curation summary not saved to L1 memory")
	}
}

type MockFetcher struct{}

func (m *MockFetcher) Fetch(ctx context.Context, url string) ([]FeedItem, error) {
	return []FeedItem{
		{
			Title:       "AI is taking over",
			Description: "A deep dive into AI advancements.",
			Link:        "https://example.com/ai",
			Published:   time.Now(),
		},
		{
			Title:       "New Go release",
			Description: "Go 1.24 is out now.",
			Link:        "https://example.com/go",
			Published:   time.Now(),
		},
	}, nil
}

func TestCurator(t *testing.T) {
	orc := orchestrator.NewOrchestrator()
	mockLLM := &orchestrator.MockLLM{}
	orc.LLM = mockLLM

	curator := &CurationModule{
		Orchestrator: orc,
		Fetcher:      &MockFetcher{},
		Feeds:        []string{"https://example.com/rss"},
	}

	err := curator.Curate("AI")
	if err != nil {
		t.Fatalf("Curate failed: %v", err)
	}

	memories := orc.L1.Search("")
	if len(memories) == 0 {
		t.Errorf("Expected summary to be saved in memory")
	}

	found := false
	for _, m := range memories {
		for _, tag := range m.Tags {
			if tag == "AI" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected tag 'AI' in memory")
	}
}
