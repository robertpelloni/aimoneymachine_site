package support

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type SupportModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewSupportModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *SupportModule {
	return &SupportModule{
		Orch:   orch,
		Broker: broker,
	}
}

// ResolveTicket uses L3 memory and LLM to solve a customer issue
func (s *SupportModule) ResolveTicket(customerName, issue string) (string, error) {
	fmt.Printf("[Support] Resolving ticket for %s: %s\n", customerName, issue)

	// Search L3 archive for technical documentation or past resolutions
	knowledge := s.Orch.L3.Search(issue)
	var context string
	if len(knowledge) > 0 {
		context = "\nRELEVANT DOCUMENTATION:\n"
		for _, k := range knowledge {
			context += fmt.Sprintf("- %s\n", k.Content)
		}
	}

	prompt := fmt.Sprintf(`Act as a helpful customer support agent for the "AI Hustle Machine".
Solve the following issue for %s: "%s".

%s

Requirements:
- Empathetic and professional tone
- Specific, actionable steps to resolve the issue
- Reference relevant documentation if available
- Close with a polite offer for further help

Format: Professional Markdown reply.`, customerName, issue, context)

	reply, err := s.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	s.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("ticket-%s-%d", customerName, time.Now().Unix()),
		Content:   fmt.Sprintf("Support Ticket Resolved for %s: %s", customerName, issue),
		Timestamp: time.Now(),
		Tags:      []string{"support", "ticket", customerName},
	})

	return reply, nil
}

// GenerateFAQ updates the L3 memory with new FAQs based on recent tickets
func (s *SupportModule) GenerateFAQ() error {
	tickets := s.Orch.L1.Search("ticket")
	if len(tickets) < 3 { return nil }

	var ticketHistory string
	for _, t := range tickets {
		ticketHistory += fmt.Sprintf("- %s\n", t.Content)
	}

	prompt := fmt.Sprintf(`Analyze the following resolved support tickets and generate 3 Frequently Asked Questions (FAQ) entries.

TICKETS:
%s

Respond with a JSON array of objects:
[
  {
    "question": "The question?",
    "answer": "The helpful answer."
  }
]`, ticketHistory)

	var faqs []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}

	if err := s.Orch.LLM.GenerateJSON(prompt, &faqs); err == nil {
		for _, f := range faqs {
			s.Orch.L3.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("faq-%d", time.Now().UnixNano()),
				Content:   fmt.Sprintf("FAQ: %s\nA: %s", f.Question, f.Answer),
				Timestamp: time.Now(),
				Tags:      []string{"support", "faq"},
			})
		}
	}
	return nil
}
