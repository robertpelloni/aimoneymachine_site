package media

import (
	"fmt"
	"strings"

	"github.com/robertpelloni/hustle/orchestrator"
)

// ShortsFactory manages the YouTube Shorts generation pipeline
type ShortsFactory struct {
	Orch *orchestrator.Orchestrator
}

// NewShortsFactory creates a new instance
func NewShortsFactory(orch *orchestrator.Orchestrator) *ShortsFactory {
	return &ShortsFactory{
		Orch: orch,
	}
}

// Run executes the full Shorts pipeline
func (f *ShortsFactory) Run(topic string) error {
	fmt.Printf("[Media] Running YouTube Shorts Factory for topic: %s\n", topic)

	// Step 1: Generate Script
	script, err := f.GenerateScript(topic)
	if err != nil {
		return err
	}
	fmt.Printf("[Media] Generated Script: %s\n", script)

	// Step 2: Generate Prompts for Images
	prompts, err := f.GenerateImagePrompts(script)
	if err != nil {
		return err
	}
	fmt.Printf("[Media] Generated %d Image Prompts\n", len(prompts))

	// Future Steps:
	// - TTS Generation (e.g. via ElevenLabs)
	// - Image Generation (e.g. via Stable Diffusion / Midjourney API)
	// - Video Assembly (FFmpeg)
	// - Upload to YouTube

	// Log to memory
	f.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("shorts-script-%s", strings.ReplaceAll(topic, " ", "-")),
		Content:   fmt.Sprintf("Generated Shorts script for %s: %s", topic, script),
		Tags:      []string{"media", "youtube", "shorts"},
	})

	fmt.Println("[Media] ✅ YouTube Shorts script and prompts generated successfully.")
	return nil
}

// GenerateScript creates a 60-second video script
func (f *ShortsFactory) GenerateScript(topic string) (string, error) {
	prompt := fmt.Sprintf(`Write a fast-paced, highly engaging 60-second YouTube Shorts script about "%s".
The script should have:
1. A strong hook (first 3 seconds).
2. Fast, value-dense body content.
3. A clear Call to Action (CTA) at the end.
Do not include stage directions, just the spoken text.`, topic)

	script, err := f.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate script: %w", err)
	}
	return strings.TrimSpace(script), nil
}

func truncateRunes(s string, l int) string {
	runes := []rune(s)
	if len(runes) > l {
		return string(runes[:l])
	}
	return s
}

// GenerateImagePrompts extracts visual scene descriptions from the script
func (f *ShortsFactory) GenerateImagePrompts(script string) ([]string, error) {
	prompt := fmt.Sprintf(`Given the following YouTube Shorts script, generate exactly 5 distinct image generation prompts (for Midjourney or Stable Diffusion) that would serve as the background visuals.
Format the output as a simple JSON array of strings.

Script:
%s`, script)

	var prompts []string
	if err := f.Orch.LLM.GenerateJSON(prompt, &prompts); err != nil {
		// Fallback mock prompts if JSON parsing fails
		return []string{
			"Cinematic establishing shot of " + truncateRunes(script, 20) + "...",
			"Dynamic close up representing " + truncateRunes(script, 20) + "...",
			"Abstract visual representing the core concept",
			"Infographic style representation of data",
			"High-energy conclusion graphic with glowing elements",
		}, nil
	}

	return prompts, nil
}
