package content

import (
	"os"
	"strings"
)

// DeployMode represents the deployment target.
type DeployMode string

const (
	DeployModeWordPress DeployMode = "wordpress"
	DeployModeGitHub    DeployMode = "github"
)

// WordPressConfig holds WordPress REST API authentication.
type WordPressConfig struct {
	URL      string // e.g. "https://example.com/wp-json/wp/v2"
	User     string
	Password string // or application password
}

// GitHubPagesConfig holds GitHub Pages deployment configuration.
type GitHubPagesConfig struct {
	Token  string // GITHUB_TOKEN
	Repo   string // e.g. "user/repo"
	Branch string // target branch (default "gh-pages")
}

// DeployConfig is the consolidated configuration for content deployment.
type DeployConfig struct {
	Mode      DeployMode
	DryRun    bool
	WordPress WordPressConfig
	GitHub    GitHubPagesConfig
	SourceDir string // directory with generated HTML files
}

// DefaultDeployConfig reads deployment configuration from environment variables.
func DefaultDeployConfig() DeployConfig {
	return DeployConfig{
		Mode:   DeployMode(strings.ToLower(os.Getenv("DEPLOY_TARGET"))),
		DryRun: os.Getenv("DEPLOY_DRY_RUN") == "true",
		WordPress: WordPressConfig{
			URL:      os.Getenv("WORDPRESS_URL"),
			User:     os.Getenv("WORDPRESS_USER"),
			Password: os.Getenv("WORDPRESS_PASSWORD"),
		},
		GitHub: GitHubPagesConfig{
			Token:  os.Getenv("GITHUB_TOKEN"),
			Repo:   os.Getenv("GITHUB_REPO"),
			Branch: getEnvOrDefault("GITHUB_PAGES_BRANCH", "gh-pages"),
		},
		SourceDir: "output/site",
	}
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// IsConfigured returns true if the configuration is usable for the selected mode.
func (c DeployConfig) IsConfigured() bool {
	switch c.Mode {
	case DeployModeWordPress:
		return c.WordPress.URL != "" && c.WordPress.User != "" && c.WordPress.Password != ""
	case DeployModeGitHub:
		return c.GitHub.Token != "" && c.GitHub.Repo != ""
	default:
		return false
	}
}

// ModeLabel returns a human-readable label for the selected mode.
func (c DeployConfig) ModeLabel() string {
	switch c.Mode {
	case DeployModeWordPress:
		return "WordPress"
	case DeployModeGitHub:
		return "GitHub Pages"
	default:
		return "none"
	}
}
