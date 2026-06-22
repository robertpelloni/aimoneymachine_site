package content

import (
    "bytes"
    "fmt"
      "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/robertpelloni/hustle/hustle/publisher"
)

// DeploySite deploys the generated static site to the specified target.
// Supported targets: "wordpress", "github". Returns an error on failure.
func DeploySite(siteDir, target string) error {
    switch strings.ToLower(target) {
    case "wordpress":
        return deployToWordPress(siteDir)
    case "github":
        return deployToGitHub(siteDir)
    case "":
        // No deployment requested.
        return nil
    default:
        return fmt.Errorf("unknown deploy target %q", target)
    }
}

// deployToWordPress publishes each HTML file as a WordPress post (HTML content).
func deployToWordPress(siteDir string) error {
    wp := publisher.NewWordPressPublisher()
    if !wp.IsConfigured() {
        return fmt.Errorf("WordPress not configured via env vars")
    }

    // Walk HTML files
    err := filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        if strings.HasSuffix(info.Name(), ".html") {
            // Read file content
            data, err := os.ReadFile(path)
            if err != nil {
                return fmt.Errorf("read %s: %w", path, err)
            }
            // Derive title from filename (strip extension, replace dashes/underscores)
            base := strings.TrimSuffix(info.Name(), ".html")
            title := strings.ReplaceAll(strings.Title(strings.ReplaceAll(base, "-", " ")), " ", " ")
            // Publish as HTML post
            _, err = wp.PublishHTMLPost(title, string(data), "publish", nil, nil)
            if err != nil {
                return fmt.Errorf("publish %s: %w", path, err)
            }
            fmt.Printf("[Deploy] Published %s to WordPress\n", info.Name())
        }
        return nil
    })
    return err
}

// deployToGitHub pushes the site directory to a GitHub Pages repository.
func deployToGitHub(siteDir string) error {
    repo := os.Getenv("GITHUB_REPO") // e.g., "user/repo"
    token := os.Getenv("GITHUB_TOKEN")
    dryRun := os.Getenv("DEPLOY_DRY_RUN") == "true"

    if repo == "" || token == "" {
        return fmt.Errorf("GITHUB_REPO and GITHUB_TOKEN env vars must be set for GitHub deployment")
    }

    if dryRun {
        fmt.Printf("[Deploy] Dry run: would push %s to GitHub repo %s\n", siteDir, repo)
        return nil
    }

    // Prepare remote URL with token for authentication
    remoteURL := fmt.Sprintf("https://%s@github.com/%s.git", token, repo)

    // Ensure the site directory is a git repo (init if needed)
    if _, err := os.Stat(filepath.Join(siteDir, ".git")); os.IsNotExist(err) {
        if err := runCmd(siteDir, "git", "init"); err != nil {
            return err
        }
    }

    // Set remote (force replace)
    if err := runCmd(siteDir, "git", "remote", "remove", "origin"); err != nil {
        // ignore if remote not present
    }
    if err := runCmd(siteDir, "git", "remote", "add", "origin", remoteURL); err != nil {
        return err
    }

    // Add all files and commit
    if err := runCmd(siteDir, "git", "add", "."); err != nil {
        return err
    }
    commitMsg := fmt.Sprintf("Deploy site %s", time.Now().Format(time.RFC3339))
    if err := runCmd(siteDir, "git", "commit", "-m", commitMsg, "--allow-empty"); err != nil {
        return err
    }

    // Push to gh-pages branch (or master if repo is set up for GitHub Pages)
    // Here we push to "gh-pages" branch; adjust as needed.
    if err := runCmd(siteDir, "git", "push", "-u", "origin", "master:gh-pages", "--force"); err != nil {
        return err
    }

    fmt.Printf("[Deploy] Successfully pushed site to GitHub repo %s (gh-pages)\n", repo)
    return nil
}

// runCmd runs a command in the given directory and streams output.
func runCmd(dir, name string, args ...string) error {
    cmd := exec.Command(name, args...)
    cmd.Dir = dir
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("%s %v failed: %v (stdout: %s, stderr: %s)", name, args, err, stdout.String(), stderr.String())
    }
    return nil
}
