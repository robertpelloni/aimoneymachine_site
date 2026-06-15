package devagency

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DevAgencyModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewDevAgencyModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *DevAgencyModule {
	return &DevAgencyModule{
		Orch:   orch,
		Broker: broker,
	}
}

// AuditCode scans a directory and proposes improvements
func (d *DevAgencyModule) AuditCode(path string) (string, error) {
	fmt.Printf("[DevAgency] Auditing code at: %s\n", path)

	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if !info.IsDir() && (strings.HasSuffix(p, ".go") || strings.HasSuffix(p, ".js") || strings.HasSuffix(p, ".py")) {
			// Read first 500 bytes for context
			content, _ := os.ReadFile(p)
			snippet := string(content)
			if len(snippet) > 500 { snippet = snippet[:500] }
			files = append(files, fmt.Sprintf("File: %s\nContent:\n%s\n", p, snippet))
		}
		if len(files) > 5 { return filepath.SkipDir } // Don't scan too much
		return nil
	})

	if err != nil { return "", err }

	prompt := fmt.Sprintf(`Act as a senior software engineer. Audit the following code snippets and propose 3 specific improvements for performance, security, or readability.

CODE SNIPPETS:
%s

Respond with a professional markdown report.`, strings.Join(files, "\n---\n"))

	report, err := d.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	d.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("audit-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Code Audit Completed for %s", path),
		Timestamp: time.Now(),
		Tags:      []string{"devagency", "audit", path},
	})

	return report, nil
}

// ProposeService generates a freelance proposal for a discovered project
func (d *DevAgencyModule) ProposeService(projectDesc string) (string, error) {
	prompt := fmt.Sprintf(`Generate a professional freelance coding proposal for the following project description: "%s".

Requirements:
- Professional introduction
- Understanding of the problem
- Proposed solution architecture
- Estimated timeline and milestones
- Call to action

Format: Professional Markdown.`, projectDesc)

	return d.Orch.LLM.Generate(prompt)
}
