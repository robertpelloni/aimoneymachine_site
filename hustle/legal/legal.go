package legal

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type LegalModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewLegalModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *LegalModule {
	return &LegalModule{
		Orch:   orch,
		Broker: broker,
	}
}

// GenerateDocument creates legal boilerplate for a specific product or service
func (l *LegalModule) GenerateDocument(docType, businessName string) (string, error) {
	fmt.Printf("[Legal] Generating %s for: %s\n", docType, businessName)

	prompt := fmt.Sprintf(`Act as a legal consultant for an autonomous AI business. Generate a professional %s for the business named: "%s".

Requirements:
- Compliant with modern data privacy standards (GDPR/CCPA)
- Includes liability disclaimers for AI-generated content
- Clear and professional language

Format: Professional Markdown.`, docType, businessName)

	doc, err := l.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	l.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("legal-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Legal Document Generated: %s (%s)", docType, businessName),
		Timestamp: time.Now(),
		Tags:      []string{"legal", "compliance", docType},
	})

	return doc, nil
}
