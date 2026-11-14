package media

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitScenesGroupsAndDurations(t *testing.T) {
	script := "Hook line here! First point is fast. Second point is faster. Third point lands. CTA follows now."
	scenes := SplitScenes(script, []string{"p1", "p2", "p3", "p4", "p5"})
	if len(scenes) != 5 {
		t.Fatalf("scenes = %d, want 5", len(scenes))
	}
	for i, sc := range scenes {
		if sc.Text == "" {
			t.Errorf("scene %d empty", i)
		}
		if sc.Duration < 2 || sc.Duration > 8 {
			t.Errorf("scene %d duration %.2f outside 2–8s", i, sc.Duration)
		}
	}
	if scenes[0].Prompt != "p1" {
		t.Errorf("scene 0 prompt = %q, want p1", scenes[0].Prompt)
	}
	if got := SplitScenes("", nil); got != nil {
		t.Errorf("empty script scenes = %v, want nil", got)
	}
}

func TestFFmpegArgsShape(t *testing.T) {
	scenes := []Scene{{Text: "It's 100%: go!", Duration: 3}, {Text: "Second card", Duration: 2}}
	args := strings.Join(ffmpegArgs(scenes, "", "out.mp4", AssembleOptions{}), " ")
	for _, want := range []string{"1080x1920", "concat=n=2", "libx264", "anullsrc", "out.mp4", `It’s 100\%\: go!`} {
		if !strings.Contains(args, want) {
			t.Errorf("args missing %q\nargs: %s", want, args)
		}
	}
	withAudio := strings.Join(ffmpegArgs(scenes, "audio.mp3", "out.mp4", AssembleOptions{}), " ")
	if !strings.Contains(withAudio, "audio.mp3") || strings.Contains(withAudio, "anullsrc") {
		t.Errorf("audio variant wrong: %s", withAudio)
	}
}

// Real render at card size, kept tiny (2 scenes x min duration, small frame).
func TestAssembleRendersMP4(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	out := filepath.Join(t.TempDir(), "short.mp4")
	scenes := []Scene{{Text: "Hello Shorts", Duration: 2}, {Text: "Second card here", Duration: 2}}
	if _, err := Assemble(scenes, "", out, AssembleOptions{Width: 320, Height: 568, FPS: 15}); err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	info, err := os.Stat(out)
	if err != nil || info.Size() == 0 {
		t.Fatalf("output missing/empty: %v", err)
	}
}
