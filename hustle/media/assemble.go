package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Scene is one card of the Short: spoken text, its image prompt, and how
// long it stays on screen.
type Scene struct {
	Text     string  `json:"text"`
	Prompt   string  `json:"prompt,omitempty"`
	Duration float64 `json:"duration_sec"`
}

// SplitScenes breaks a script into at most len(prompts) scenes (5 when no
// prompts), grouping sentences evenly. Duration assumes ~2.5 spoken
// words/sec, clamped to 2–8s per card.
func SplitScenes(script string, prompts []string) []Scene {
	maxScenes := len(prompts)
	if maxScenes == 0 {
		maxScenes = 5
	}
	sentences := splitSentences(script)
	if len(sentences) == 0 {
		return nil
	}
	if len(sentences) > maxScenes {
		// Merge overflow sentences into the last scenes evenly.
		grouped := make([]string, maxScenes)
		for i, s := range sentences {
			idx := i * maxScenes / len(sentences)
			if grouped[idx] != "" {
				grouped[idx] += " "
			}
			grouped[idx] += s
		}
		sentences = grouped
	}
	scenes := make([]Scene, len(sentences))
	for i, text := range sentences {
		prompt := ""
		if i < len(prompts) {
			prompt = prompts[i]
		}
		words := len(strings.Fields(text))
		dur := float64(words) / 2.5
		if dur < 2 {
			dur = 2
		}
		if dur > 8 {
			dur = 8
		}
		scenes[i] = Scene{Text: text, Prompt: prompt, Duration: dur}
	}
	return scenes
}

func splitSentences(s string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range s {
		cur.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			if t := strings.TrimSpace(cur.String()); t != "" {
				out = append(out, t)
			}
			cur.Reset()
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" {
		out = append(out, t)
	}
	return out
}

// escapeDrawtext escapes text for an ffmpeg drawtext filter argument.
func escapeDrawtext(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`'`, `\'`,
		`:`, `\:`,
		`%`, `\%`,
		"\n", " ",
	)
	// Keep cards readable: drawtext has no wrapping, so cap the text.
	if runes := []rune(s); len(runes) > 90 {
		s = string(runes[:87]) + "..."
	}
	return r.Replace(s)
}

// AssembleOptions controls the render. Zero values get Shorts defaults.
type AssembleOptions struct {
	Width  int
	Height int
	FPS    int
}

func (o AssembleOptions) withDefaults() AssembleOptions {
	if o.Width == 0 {
		o.Width = 1080
	}
	if o.Height == 0 {
		o.Height = 1920
	}
	if o.FPS == 0 {
		o.FPS = 30
	}
	return o
}

// ffmpegArgs builds the ffmpeg invocation: one lavfi colour card per scene
// with centred drawtext, concatenated, with the TTS track when present or
// a silent track otherwise (players/YouTube expect an audio stream).
func ffmpegArgs(scenes []Scene, audioPath, outPath string, opts AssembleOptions) []string {
	opts = opts.withDefaults()
	args := []string{"-y"}
	for _, sc := range scenes {
		args = append(args, "-f", "lavfi", "-i",
			fmt.Sprintf("color=c=0x0f172a:s=%dx%d:d=%.2f:r=%d", opts.Width, opts.Height, sc.Duration, opts.FPS))
	}
	audioIdx := len(scenes)
	if audioPath != "" {
		args = append(args, "-i", audioPath)
	} else {
		args = append(args, "-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo")
	}
	var filters []string
	for i, sc := range scenes {
		filters = append(filters, fmt.Sprintf(
			"[%d:v]drawtext=text='%s':fontcolor=white:fontsize=52:x=(w-text_w)/2:y=(h-text_h)/2:box=1:boxcolor=black@0.45:boxborderw=24[v%d]",
			i, escapeDrawtext(sc.Text), i))
	}
	concatIn := ""
	for i := range scenes {
		concatIn += fmt.Sprintf("[v%d]", i)
	}
	filters = append(filters, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vout]", concatIn, len(scenes)))
	args = append(args,
		"-filter_complex", strings.Join(filters, ";"),
		"-map", "[vout]", "-map", fmt.Sprintf("%d:a", audioIdx),
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac",
		"-shortest", outPath)
	return args
}

// hasDrawtext probes whether this ffmpeg build includes the drawtext
// filter (minimal/static builds without libfreetype do not).
func hasDrawtext(ffmpeg string) bool {
	out, err := exec.Command(ffmpeg, "-hide_banner", "-filters").Output()
	return err == nil && strings.Contains(string(out), "drawtext")
}

// plainArgs is ffmpegArgs without drawtext: blank colour cards.
func plainArgs(scenes []Scene, audioPath, outPath string, opts AssembleOptions) []string {
	args := ffmpegArgs(scenes, audioPath, outPath, opts)
	for i, a := range args {
		if a == "-filter_complex" {
			concatIn := ""
			for j := range scenes {
				concatIn += fmt.Sprintf("[%d:v]", j)
			}
			args[i+1] = fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vout]", concatIn, len(scenes))
		}
	}
	return args
}

// Assemble renders scenes to an MP4 with ffmpeg (must be on PATH). The
// returned note is non-empty when the render degraded (e.g. no drawtext
// in this ffmpeg build, so cards carry no text).
func Assemble(scenes []Scene, audioPath, outPath string, opts AssembleOptions) (string, error) {
	if len(scenes) == 0 {
		return "", fmt.Errorf("no scenes to assemble")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg not on PATH: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return "", err
	}
	note := ""
	args := ffmpegArgs(scenes, audioPath, outPath, opts)
	if !hasDrawtext(ffmpeg) {
		note = "drawtext unavailable in this ffmpeg build: rendered blank cards (no on-screen text)"
		args = plainArgs(scenes, audioPath, outPath, opts)
	}
	// Minimal ffmpeg builds may lack libx264; fall back to native mpeg4.
	if encOut, err := exec.Command(ffmpeg, "-hide_banner", "-encoders").Output(); err == nil && !strings.Contains(string(encOut), "libx264") {
		if strings.Contains(string(encOut), "mpeg4") {
			for i, a := range args {
				if a == "libx264" {
					args[i] = "mpeg4"
				}
			}
			if note != "" {
				note += "; "
			}
			note += "libx264 unavailable: encoded with mpeg4"
		}
	}
	cmd := exec.Command(ffmpeg, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// Report the tail: ffmpeg's banner head hides the actual error.
		msg := stderr.String()
		if len(msg) > 600 {
			msg = msg[len(msg)-600:]
		}
		return "", fmt.Errorf("ffmpeg failed: %w: %s", err, msg)
	}
	return note, nil
}

// TTSProvider synthesizes speech to an audio file.
type TTSProvider interface {
	Synthesize(text, outPath string) error
}

// ElevenLabsTTS calls the ElevenLabs API (requires ELEVENLABS_API_KEY).
type ElevenLabsTTS struct {
	APIKey  string
	VoiceID string // default: Rachel
	Client  *http.Client
}

func (e *ElevenLabsTTS) Synthesize(text, outPath string) error {
	voice := e.VoiceID
	if voice == "" {
		voice = "21m00Tcm4TlvDq8ikWAM"
	}
	client := e.Client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	body, _ := json.Marshal(map[string]interface{}{
		"text":     text,
		"model_id": "eleven_turbo_v2_5",
	})
	req, err := http.NewRequest("POST", "https://api.elevenlabs.io/v1/text-to-speech/"+voice, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("xi-api-key", e.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		return fmt.Errorf("elevenlabs status %d: %s", resp.StatusCode, string(b))
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// NewTTSFromEnv returns an ElevenLabs provider when ELEVENLABS_API_KEY is
// set, else nil (the pipeline then renders silent cards and says so).
func NewTTSFromEnv() TTSProvider {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		return nil
	}
	return &ElevenLabsTTS{APIKey: key, VoiceID: os.Getenv("ELEVENLABS_VOICE_ID")}
}

// Manifest records exactly which pipeline stages ran, so a silent or
// un-uploaded Short is never mistaken for a finished one.
type Manifest struct {
	Topic     string    `json:"topic"`
	Script    string    `json:"script_file"`
	Scenes    []Scene   `json:"scenes"`
	Audio     string    `json:"audio_file,omitempty"`
	AudioNote string    `json:"audio_note,omitempty"`
	Video     string    `json:"video_file,omitempty"`
	VideoNote string    `json:"video_note,omitempty"`
	Upload    string    `json:"upload"`
	CreatedAt time.Time `json:"created_at"`
}

func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
