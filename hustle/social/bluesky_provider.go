package social

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/robertpelloni/hustle/orchestrator"
)

// BlueskyProvider posts to Bluesky via the external Python script.
type BlueskyProvider struct {
	DryRun   bool
	Handle   string
	Password string
}

func (p *BlueskyProvider) SetDryRun(enabled bool) {
	p.DryRun = enabled
}

func (p *BlueskyProvider) Search(query string) ([]SocialPost, error) {
	return nil, fmt.Errorf("Bluesky search not implemented")
}

func (p *BlueskyProvider) Reply(postID, content string) error {
	return fmt.Errorf("Bluesky reply not implemented")
}

func (p *BlueskyProvider) Post(orch *orchestrator.Orchestrator, platform, content string) error {
	if p.DryRun {
		fmt.Printf("[Bluesky] DryRun: Posting to %s: %s\n", platform, content)
		return nil
	}

	if p.Handle == "" || p.Password == "" {
		return fmt.Errorf("missing BLUESKY_HANDLE or BLUESKY_APP_PASSWORD")
	}

	cmd := exec.Command("python3", "/opt/aimoneymachine/bin/bluesky_poster.py")
	cmd.Env = append(os.Environ(),
		"BLUESKY_HANDLE="+p.Handle,
		"BLUESKY_APP_PASSWORD="+p.Password,
	)

	// Pass content via stdin
	cmd.Stdin = strings.NewReader(content)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("bluesky post failed: %w\noutput: %s", err, string(out))
	}

	fmt.Printf("[Bluesky] Post result: %s\n", string(out))
	return nil
}

func NewBlueskyProvider(handle, password string) *BlueskyProvider {
	hand := handle
	if hand == "" {
		hand = os.Getenv("BLUESKY_HANDLE")
	}
	pass := password
	if pass == "" {
		pass = os.Getenv("BLUESKY_APP_PASSWORD")
	}
	return &BlueskyProvider{
		DryRun:   os.Getenv("BLUESKY_DRY_RUN") == "true" || os.Getenv("DRY_RUN") == "true",
		Handle:   hand,
		Password: pass,
	}
}
