package careers

import (
	"fmt"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type CareerModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewCareerModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *CareerModule {
	return &CareerModule{
		Orch:   orch,
		Broker: broker,
	}
}

func firstLineSummary(s string, max int) string {
	line := strings.TrimSpace(s)
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	r := []rune(line)
	if len(r) > max {
		return string(r[:max]) + "..."
	}
	return line
}

// TailorResume optimizes a resume for a specific job description.
func (c *CareerModule) TailorResume(originalResume, jobDesc string) (string, error) {
	fmt.Printf("[Careers] Tailoring resume for job description\n")

	prompt := fmt.Sprintf(`Act as a professional career coach. Tailor the following resume to better match the job description provided.

RESUME:
%s

JOB DESCRIPTION:
%s

Requirements:
- Highlight relevant skills and experiences
- Use keywords from the job description
- Maintain a professional and clean format
- Focus on quantifiable achievements

Format: Professional Markdown Resume.`, originalResume, jobDesc)

	tailored, err := c.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", err
	}

	c.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("resume-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Tailored resume for job: %s\n\n%s", firstLineSummary(jobDesc, 80), tailored),
		Timestamp: time.Now(),
		Tags:      []string{"careers", "resume", "tailoring"},
	})

	return tailored, nil
}

// SearchJobs finds relevant job postings for a title and location.
func (c *CareerModule) SearchJobs(title, location string) (string, error) {
	fmt.Printf("[Careers] Searching for jobs: %s in %s\n", title, location)

	prompt := fmt.Sprintf(`Generate a list of 5 realistic, currently trending job titles and descriptions for: "%s" in "%s".

Requirements:
- Role name
- Company name (realistic)
- Key responsibilities
- Required skills
- Estimated salary range

Format: Professional Markdown list.`, title, location)

	results, err := c.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", err
	}

	c.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("jobs-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Job search completed for %s in %s:\n\n%s", title, location, results),
		Timestamp: time.Now(),
		Tags:      []string{"careers", "jobs", title},
	})

	return results, nil
}
