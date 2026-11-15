package marketing

import (
	"fmt"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type Campaign struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Platform string  `json:"platform"`
	Status   string  `json:"status"`
	Budget   float64 `json:"budget"`
	Spend    float64 `json:"spend"`
	Revenue  float64 `json:"revenue"`
}

type MarketingModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewMarketingModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *MarketingModule {
	return &MarketingModule{
		Orch:   orch,
		Broker: broker,
	}
}

func truncMkt(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "..."
}

// AnalyzeROI determines which marketing platforms are performing best.
func (m *MarketingModule) AnalyzeROI() (string, error) {
	fmt.Printf("[Marketing] Analyzing multi-platform campaign ROI\n")

	entries := m.Orch.L1.Search("campaign")
	var history string
	for _, e := range entries {
		history += fmt.Sprintf("- %s\n", e.Content)
	}
	if history == "" {
		history = "- Baseline organic & social campaigns active across Bluesky, WordPress, and Affiliate feeds."
	}

	prompt := fmt.Sprintf(`Act as a marketing analytics expert. Analyze the following campaign history and recommend which platform to increase budget for and which to pause.

CAMPAIGN DATA:
%s

Requirements:
- Calculate ROAS (Return on Ad Spend) per platform
- Identify top performer
- Suggest 1 optimization for the underperformer

Format: Professional Markdown.`, history)

	recommendation, err := m.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", err
	}

	m.Orch.L2.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("mkt-roi-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Marketing ROI Analysis:\n%s", truncMkt(recommendation, 600)),
		Timestamp: time.Now(),
		Tags:      []string{"marketing", "roi", "optimization"},
	})

	return recommendation, nil
}
