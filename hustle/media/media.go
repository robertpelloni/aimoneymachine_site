package media

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type MediaModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewMediaModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *MediaModule {
	return &MediaModule{
		Orch:   orch,
		Broker: broker,
	}
}

// PlanProduction converts a script into a list of required assets
func (m *MediaModule) PlanProduction(scriptTitle, scriptType string) (string, error) {
	fmt.Printf("[Media] Planning production for: %s (%s)\n", scriptTitle, scriptType)

	prompt := fmt.Sprintf(`Act as a video producer. Analyze the following script title and type, and generate a production asset list.

TITLE: %s
TYPE: %s

Requirements:
- List of 5-10 specific B-roll/stock footage descriptions
- Suggestions for background music (mood, tempo)
- SFX requirements
- Text overlay requirements
- Voiceover (TTS) style recommendation

Format: Professional Markdown checklist.`, scriptTitle, scriptType)

	plan, err := m.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	m.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("media-plan-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Media Production Plan for %s", scriptTitle),
		Timestamp: time.Now(),
		Tags:      []string{"media", "production", scriptType},
	})

	return plan, nil
}
