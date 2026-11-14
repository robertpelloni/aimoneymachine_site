package media

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Steps 3–5: scenes -> optional TTS -> FFmpeg render. Each stage that
	// cannot run (no TTS key, no ffmpeg, no YouTube OAuth) is recorded in
	// the manifest instead of failing or faking success.
	if err := f.render(topic, script, prompts); err != nil {
		return err
	}

	// Log to memory
	f.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("shorts-script-%s", strings.ReplaceAll(topic, " ", "-")),
		Content:   fmt.Sprintf("Generated Shorts script for %s: %s", topic, script),
		Tags:      []string{"media", "youtube", "shorts"},
	})

	fmt.Println("[Media] ✅ YouTube Shorts pipeline complete (see manifest.json in output dir).")
	return nil
}

// render writes script/scenes/manifest under output/shorts/<slug>-<ts>/,
// synthesizes TTS when a provider is configured, and renders the MP4 with
// ffmpeg when available. The manifest states precisely what happened.
func (f *ShortsFactory) render(topic, script string, prompts []string) error {
	scenes := SplitScenes(script, prompts)
	dir := filepath.Join("output", "shorts", fmt.Sprintf("%s-%d", slugify(topic), time.Now().Unix()))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	scriptPath := filepath.Join(dir, "script.txt")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return err
	}

	m := Manifest{
		Topic:     topic,
		Script:    scriptPath,
		Scenes:    scenes,
		Upload:    "not_configured: set YOUTUBE_CLIENT_ID/YOUTUBE_REFRESH_TOKEN to enable upload",
		CreatedAt: time.Now(),
	}

	// TTS (optional)
	audioPath := ""
	if tts := NewTTSFromEnv(); tts != nil {
		audioPath = filepath.Join(dir, "audio.mp3")
		if err := tts.Synthesize(script, audioPath); err != nil {
			m.AudioNote = fmt.Sprintf("tts failed: %v", err)
			audioPath = ""
		} else {
			m.Audio = audioPath
		}
	} else {
		m.AudioNote = "skipped: ELEVENLABS_API_KEY not set (silent cards)"
	}

	// FFmpeg render (optional but expected on prod)
	videoPath := filepath.Join(dir, "short.mp4")
	if note, err := Assemble(scenes, audioPath, videoPath, AssembleOptions{}); err != nil {
		m.VideoNote = fmt.Sprintf("render skipped/failed: %v", err)
		fmt.Printf("[Media] %s\n", m.VideoNote)
	} else {
		m.Video = videoPath
		m.VideoNote = note // e.g. blank cards when drawtext is unavailable
		fmt.Printf("[Media] Rendered %s (%d scenes)\n", videoPath, len(scenes))
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644)
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
	raw, gerr := f.Orch.LLM.Generate(prompt + "\n\nRespond ONLY with a valid JSON array of strings, no markdown fences.")
	if gerr != nil || json.Unmarshal([]byte(raw), &prompts) != nil {
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
