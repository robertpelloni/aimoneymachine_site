package crm

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type LeadStatus string

const (
	Discovered LeadStatus = "discovered"
	Contacted  LeadStatus = "contacted"
	Replied    LeadStatus = "replied"
	Closed     LeadStatus = "closed"
)

type ClientLead struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Status    LeadStatus `json:"status"`
	Niche     string     `json:"niche"`
	LastTouch time.Time  `json:"last_touch"`
}

type CRMModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewCRMModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *CRMModule {
	return &CRMModule{
		Orch:   orch,
		Broker: broker,
	}
}

// UpdateLeadStatus updates the status of a lead in memory
func (c *CRMModule) UpdateLeadStatus(email string, status LeadStatus) error {
	fmt.Printf("[CRM] Updating status for %s to: %s\n", email, status)

	// Store status update in L1
	c.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("crm-update-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("CRM Update: %s set to %s", email, status),
		Timestamp: time.Now(),
		Tags:      []string{"crm", "lead", email, string(status)},
	})

	return nil
}

// GenerateFollowUp creates a personalized follow-up message for a lead
func (c *CRMModule) GenerateFollowUp(lead ClientLead) (string, error) {
	prompt := fmt.Sprintf(`Act as a professional relationship manager. Generate a polite follow-up email for the lead: "%s" (%s) who was previously contacted about: "%s".

Requirements:
- Referral to the previous message
- Add a new value-proposition or insight
- Low-friction call to action

Format: Professional Markdown email.`, lead.Name, lead.Email, lead.Niche)

	return c.Orch.LLM.Generate(prompt)
}
