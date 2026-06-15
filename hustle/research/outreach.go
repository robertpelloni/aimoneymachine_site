package research

import (
	"fmt"
	"net/smtp"
	"os"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

// OutreachPitch represents a personalized outreach message
type OutreachPitch struct {
	RecipientName string `json:"recipient_name"`
	RecipientEmail string `json:"recipient_email,omitempty"`
	Subject       string `json:"subject"`
	Message       string `json:"message"`
	Platform      string `json:"platform"` // email, linkedin, twitter
}

// SendEmail autonomously delivers a pitch via SMTP
func SendEmail(pitch OutreachPitch) error {
	from := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	to := pitch.RecipientEmail
	if to == "" {
		to = os.Getenv("OUTREACH_TARGET_EMAIL") // Fallback
	}

	if from == "" || password == "" || host == "" || to == "" {
		return fmt.Errorf("SMTP configuration incomplete (from=%s, host=%s, to=%s)", from, host, to)
	}

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + pitch.Subject + "\r\n" +
		"\r\n" +
		pitch.Message + "\r\n")

	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)
	if err != nil {
		return err
	}

	fmt.Printf("[Outreach] ✅ Successfully sent email to %s\n", to)
	return nil
}

// GenerateOutreach uses the LLM to create personalized pitches for discovered leads
func GenerateOutreach(orch *orchestrator.Orchestrator, topic string) ([]OutreachPitch, error) {
	fmt.Printf("[Outreach] Generating personalized pitches for: %s\n", topic)

	// 1. Fetch leads from memory (L2 vault)
	leads := orch.L2.Search("leadgen")
	if len(leads) == 0 {
		return nil, fmt.Errorf("no discovered leads found in memory for outreach")
	}

	// Use the last 3 leads for outreach
	if len(leads) > 3 {
		leads = leads[len(leads)-3:]
	}

	var context string
	for _, l := range leads {
		context += fmt.Sprintf("- LEAD INFO: %s\n", l.Content)
	}

	prompt := fmt.Sprintf(`Based on the following discovered leads, generate a highly personalized, professional outreach pitch for each one.
The goal is to offer AI automation services that solve their specific needs.

LEAD CONTEXT:
%s

Respond with a JSON array of objects:
[
  {
    "recipient_name": "Name of company or contact",
    "recipient_email": "Email address if known from context",
    "subject": "Compelling subject line",
    "message": "The personalized pitch (2-3 paragraphs)",
    "platform": "linkedin|email"
  }
]

Respond ONLY with valid JSON.`, context)

	var pitches []OutreachPitch
	if err := orch.LLM.GenerateJSON(prompt, &pitches); err != nil {
		return nil, fmt.Errorf("outreach generation failed: %v", err)
	}

	// 2. Record in memory
	for _, p := range pitches {
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("outreach-%d", time.Now().UnixNano()),
			Content:   fmt.Sprintf("PREPARED OUTREACH for %s via %s: %s", p.RecipientName, p.Platform, p.Subject),
			BaseScore: 80.0,
			Timestamp: time.Now(),
			Tags:      []string{"outreach", topic, p.RecipientName, p.Platform},
		})
	}

	return pitches, nil
}
