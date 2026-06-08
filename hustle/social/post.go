package social

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type SocialPost struct {
	Platform    string
	Content     string
	ScheduledAt time.Time
}

type Provider interface {
	Post(orch *orchestrator.Orchestrator, platform, content string) error
	SetDryRun(enabled bool)
}

type TwitterProvider struct {
	DryRun bool
}

func (p *TwitterProvider) SetDryRun(enabled bool) {
	p.DryRun = enabled
}

func (p *TwitterProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[Twitter][DRY-RUN] Would post to %s: %s\n", platform, content)
		return nil
	}
	fmt.Printf("[Twitter] Posting to %s: %s\n", platform, content)
	return nil
}

func NewTwitterProvider() *TwitterProvider {
	return &TwitterProvider{}
}

type LinkedInProvider struct {
	DryRun bool
}

func (p *LinkedInProvider) SetDryRun(enabled bool) {
	p.DryRun = enabled
}

func (p *LinkedInProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[LinkedIn][DRY-RUN] Would post to %s: %s\n", platform, content)
		return nil
	}
	fmt.Printf("[LinkedIn] Posting to %s: %s\n", platform, content)
	return nil
}

func NewLinkedInProvider() *LinkedInProvider {
	return &LinkedInProvider{}
}

func GenerateContent(orch *orchestrator.Orchestrator, topic string) string {
	prompt := fmt.Sprintf("Create a short, revolutionary social media post about %s with hashtags.", topic)
	content, err := orch.LLM.Generate(prompt)
	if err != nil {
		return fmt.Sprintf("Revolutionary insights on %s! #hustle #ai", topic)
	}
	return content
}

func SchedulePost(orch *orchestrator.Orchestrator, provider Provider, platform, topic string) {
	content := GenerateContent(orch, topic)
	fmt.Printf("Scheduling post for %s: %s\n", platform, content)

	err := provider.Post(orch, platform, content)
	if err == nil {
		orch.L1.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("social-%s-%d", platform, time.Now().Unix()),
			Content:   fmt.Sprintf("Posted to %s: %s", platform, content),
			Timestamp: time.Now(),
			Tags:      []string{"social", platform},
		})

		orch.Ledger.Add(orchestrator.Transaction{
			Amount: 0.01,
			Type:   orchestrator.Expense,
			Hustle: "SocialMedia",
			Note:   fmt.Sprintf("API post to %s", platform),
		})
	}
}
