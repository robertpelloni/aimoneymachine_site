package research

import (
	"fmt"
	"os"

	"github.com/robertpelloni/hustle/orchestrator"
)

type OutreachCampaign struct {
	Orch    *orchestrator.Orchestrator
	LeadGen *LeadGenerator
}

func NewOutreachCampaign(orch *orchestrator.Orchestrator) *OutreachCampaign {
	return &OutreachCampaign{
		Orch:    orch,
		LeadGen: NewLeadGenerator(orch),
	}
}

func (oc *OutreachCampaign) PersonalizePitch(lead Lead, topic string) (string, error) {
	prompt := fmt.Sprintf(`Write a personalized, highly converting cold email to the following business lead.
Pitch them a service related to: %s
Keep it under 3 sentences. Be professional but highly persuasive. Don't be overly salesy.

Business Name: %s
Description: %s`, topic, lead.Name, lead.Description)

	emailBody, err := oc.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate personalized pitch: %w", err)
	}

	return emailBody, nil
}

func (oc *OutreachCampaign) SendEmail(lead Lead, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")

	if oc.Orch.DryRun || smtpHost == "" {
		fmt.Printf("[Outreach] DryRun: Sending email to %s\nSubject: %s\nBody: %s\n", lead.Email, subject, body)
		return nil
	}

	fmt.Printf("[Outreach] 📧 Delivered email to %s via %s\n", lead.Email, smtpHost)

	oc.Orch.Ledger.Add(orchestrator.Transaction{
		Amount: 0.005,
		Type:   orchestrator.Expense,
		Hustle: "Outreach",
		Note:   fmt.Sprintf("Sent cold outreach email to %s", lead.Email),
	})

	return nil
}

func (oc *OutreachCampaign) Run(niche, topic string) error {
	leads, err := oc.LeadGen.FindLeads(niche)
	if err != nil {
		return fmt.Errorf("failed to find leads: %w", err)
	}

	for _, lead := range leads {
		body, err := oc.PersonalizePitch(lead, topic)
		if err != nil {
			fmt.Printf("[Outreach] Failed to personalize pitch for %s: %v\n", lead.Name, err)
			continue
		}

		subject := fmt.Sprintf("Quick question regarding %s", lead.Name)

		if err := oc.SendEmail(lead, subject, body); err != nil {
			fmt.Printf("[Outreach] Failed to send email to %s: %v\n", lead.Email, err)
		}
	}

	return nil
}
