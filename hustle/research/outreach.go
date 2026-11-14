package research

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/robertpelloni/hustle/orchestrator"
)

type OutreachCampaign struct {
	Orch    *orchestrator.Orchestrator
	LeadGen *LeadGenerator
	// WebSearch, when set, discovers real leads from web results instead
	// of the demo seed list (see LeadGenerator.FindLeadsFromSearch).
	WebSearch SearchInterface
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

// buildMessage renders a minimal RFC 5322 message.
func buildMessage(from, to, subject, body string) []byte {
	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n",
		from, to, subject, body))
}

// deliverSMTP is smtp.SendMail behind a variable so tests can capture the
// delivery instead of touching a network.
var deliverSMTP = smtp.SendMail

// SendEmail delivers one pitch. Contract, previously violated: it must
// never report delivery that did not happen. Dry-run (Orch.DryRun or no
// SMTP_HOST configured) logs a clearly-labeled preview, sends nothing,
// and charges nothing. A real send charges the ledger only after the
// SMTP server accepts the message.
func (oc *OutreachCampaign) SendEmail(lead Lead, subject, body string) error {
	if strings.TrimSpace(lead.Email) == "" {
		return fmt.Errorf("lead %q has no email address (source: %s)", lead.Name, lead.Source)
	}

	smtpHost := os.Getenv("SMTP_HOST")
	if oc.Orch.DryRun || smtpHost == "" {
		fmt.Printf("[Outreach][DRY-RUN] Would send to %s\nSubject: %s\nBody: %s\n", lead.Email, subject, body)
		return nil
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = os.Getenv("SMTP_USER")
	}
	if from == "" {
		return fmt.Errorf("SMTP_FROM (or SMTP_USER) must be set to send")
	}

	var auth smtp.Auth
	if user, pass := os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"); user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, smtpHost)
	}

	addr := smtpHost + ":" + port
	if err := deliverSMTP(addr, auth, from, []string{lead.Email}, buildMessage(from, lead.Email, subject, body)); err != nil {
		return fmt.Errorf("smtp delivery to %s via %s: %w", lead.Email, addr, err)
	}

	fmt.Printf("[Outreach] Delivered email to %s via %s\n", lead.Email, addr)
	oc.Orch.Ledger.Add(orchestrator.Transaction{
		Amount: 0.005,
		Type:   orchestrator.Expense,
		Hustle: "Outreach",
		Note:   fmt.Sprintf("Sent cold outreach email to %s", lead.Email),
	})
	return nil
}

func (oc *OutreachCampaign) Run(niche, topic string) error {
	var leads []Lead
	var err error
	if oc.WebSearch != nil {
		leads, err = oc.LeadGen.FindLeadsFromSearch(oc.WebSearch, niche)
		if err != nil || len(leads) == 0 {
			fmt.Printf("[Outreach] Web lead discovery yielded nothing (%v); falling back to demo seed leads\n", err)
			leads, err = oc.LeadGen.FindLeads(niche)
		}
	} else {
		leads, err = oc.LeadGen.FindLeads(niche)
	}
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
			fmt.Printf("[Outreach] Not delivered to %s: %v\n", lead.Name, err)
		}
	}

	return nil
}
