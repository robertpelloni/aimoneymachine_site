package affiliate

import (
	"fmt"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

// Module handles finding and injecting affiliate links
type Module struct {
	Orch *orchestrator.Orchestrator
}

// NewModule creates an affiliate monetization hustle
func NewModule(orch *orchestrator.Orchestrator) *Module {
	return &Module{
		Orch: orch,
	}
}

// GenerateAffiliateLink uses LLM to select an appropriate dummy affiliate product based on content
func (m *Module) GenerateAffiliateLink(topic string) string {
	prompt := fmt.Sprintf(`Given the topic "%s", suggest a highly relevant digital product, tool, or service to promote as an affiliate.
Respond ONLY with the name of the product and a dummy affiliate link in this format: Product Name - https://affiliate.link/dummy`, topic)

	suggestion, err := m.Orch.LLM.Generate(prompt)
	if err != nil {
		return "Premium Tech Tool - https://affiliate.link/fallback"
	}

	suggestion = strings.TrimSpace(suggestion)
	if suggestion == "" {
		return "Premium Tech Tool - https://affiliate.link/fallback"
	}

	return suggestion
}

// InjectAffiliateLink modifies social/blog content to include a monetization link
func (m *Module) InjectAffiliateLink(content string, topic string) string {
	link := m.GenerateAffiliateLink(topic)

	// Simply append for now
	injected := fmt.Sprintf("%s\n\nSponsored/Affiliate: %s", content, link)

	m.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("affiliate-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Injected affiliate link into content for topic: %s", topic),
		Timestamp: time.Now(),
		Tags:      []string{"affiliate", "monetization", topic},
	})

	m.Orch.Ledger.Add(orchestrator.Transaction{
		Amount: 0.10, // dummy potential revenue
		Type:   orchestrator.Revenue,
		Hustle: "Affiliate",
		Note:   fmt.Sprintf("Affiliate link generation for: %s", topic),
	})

	return injected
}
