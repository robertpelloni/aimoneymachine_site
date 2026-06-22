package localization

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type LocalizationModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewLocalizationModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *LocalizationModule {
	return &LocalizationModule{
		Orch:   orch,
		Broker: broker,
	}
}

// TranslateContent translates a string into the target language
func (l *LocalizationModule) TranslateContent(content, targetLang string) (string, error) {
	fmt.Printf("[Localization] Translating content to: %s\n", targetLang)

	prompt := fmt.Sprintf(`Translate the following text into %s. Maintain the original formatting and tone.

TEXT:
%s

Respond with ONLY the translated text.`, targetLang, content)

	translated, err := l.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	l.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("trans-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Translated content to %s", targetLang),
		Timestamp: time.Now(),
		Tags:      []string{"localization", "translation", targetLang},
	})

	return translated, nil
}

// LocalizeSEO provides language-specific SEO keywords for a niche
func (l *LocalizationModule) LocalizeSEO(niche, targetLang string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate 5 high-volume SEO keywords for the niche: "%s" in the %s language.

Respond with a JSON array of strings.`, niche, targetLang)

	var keywords []string
	if err := l.Orch.LLM.GenerateJSON(prompt, &keywords); err != nil {
		return nil, err
	}
	return keywords, nil
}
