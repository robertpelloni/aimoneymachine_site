package content

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SiteConfig controls how the static site is generated
type SiteConfig struct {
	Title        string
	Description  string
	BaseURL      string
	Author       string
	Theme        string // "light" or "dark"
	OutputDir    string // where to write HTML files
	DeployTarget string // "wordpress" or "github" — optional deployment after generation
}

// DefaultSiteConfig returns sensible defaults
func DefaultSiteConfig() SiteConfig {
	return SiteConfig{
		Title:       "AI Money Machine — Content Library",
		Description: "Autonomously generated content by the AI Hustle Machine",
		BaseURL:     "https://aimoneymachine.site",
		Author:      "AI Money Machine",
		Theme:       "dark",
		OutputDir:   "output/site",
		DeployTarget: "", // no deployment by default
	}
}

// SitePage represents a rendered HTML page
type SitePage struct {
	Title       string
	Description string
	Body        string // HTML body content (without layout)
	URL         string
	Date        time.Time
	Type        ContentType
	Keywords    []string
}

// GenerateSite reads all markdown from the content output dir and builds a static HTML site
func GenerateSite(sourceDir string, config SiteConfig) error {
	if config.OutputDir == "" {
		config.OutputDir = "output/site"
	}

	fmt.Printf("[Deploy] Generating static site from: %s\n", sourceDir)
	fmt.Printf("[Deploy] Output to: %s\n", config.OutputDir)

	// Ensure output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Create assets directory
	assetsDir := filepath.Join(config.OutputDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create assets dir: %w", err)
	}

	// Scan for markdown files
	pattern := filepath.Join(sourceDir, "*.md")
	markdownFiles, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan markdown files: %w", err)
	}

	if len(markdownFiles) == 0 {
		return fmt.Errorf("no markdown files found in %s", sourceDir)
	}

	fmt.Printf("[Deploy] Found %d content files\n", len(markdownFiles))

	// Parse all files into pages
	var pages []SitePage
	for _, mf := range markdownFiles {
		page, err := parseMarkdownFile(mf)
		if err != nil {
			fmt.Printf("[Deploy] Warning: skipping %s: %v\n", mf, err)
			continue
		}
		pages = append(pages, *page)
	}

	// Sort by date descending
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Date.After(pages[j].Date)
	})

	// Generate CSS
	css := generateCSS(config.Theme)
	if err := os.WriteFile(filepath.Join(assetsDir, "style.css"), []byte(css), 0644); err != nil {
		return fmt.Errorf("failed to write CSS: %w", err)
	}

	// Generate index page
	indexHTML := generateIndexHTML(pages, config)
	if err := os.WriteFile(filepath.Join(config.OutputDir, "index.html"), []byte(indexHTML), 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	// Generate individual article pages
	for _, page := range pages {
		articleHTML := generateArticleHTML(page, config)
		slug := slugify(page.Title)
		articlePath := filepath.Join(config.OutputDir, slug+".html")

		if err := os.WriteFile(articlePath, []byte(articleHTML), 0644); err != nil {
			fmt.Printf("[Deploy] Warning: failed to write %s: %v\n", articlePath, err)
			continue
		}
	}

	// Generate RSS feed
	rss := generateRSSFeed(pages, config)
	if err := os.WriteFile(filepath.Join(config.OutputDir, "feed.xml"), []byte(rss), 0644); err != nil {
		return fmt.Errorf("failed to write RSS: %w", err)
	}

	fmt.Printf("[Deploy] ✅ Site generated: %d pages to %s\n", len(pages), config.OutputDir)
	return nil
}

// parseMarkdownFile reads a markdown file with YAML frontmatter and returns a SitePage
func parseMarkdownFile(path string) (*SitePage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	page := &SitePage{
		Date: time.Now(),
	}

	// Extract YAML frontmatter (--- ... ---)
	if strings.HasPrefix(content, "---") {
		end := strings.Index(content[3:], "---")
		if end >= 0 {
			frontmatter := content[3 : end+3]
			content = content[end+6:]

			// Parse simple YAML fields
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "title:") {
					page.Title = strings.Trim(strings.TrimPrefix(line, "title:"), ` "`)
				} else if strings.HasPrefix(line, "type:") {
					page.Type = ContentType(strings.TrimSpace(strings.TrimPrefix(line, "type:")))
				} else if strings.HasPrefix(line, "date:") {
					dateStr := strings.TrimSpace(strings.TrimPrefix(line, "date:"))
					if t, err := time.Parse("2006-01-02", dateStr); err == nil {
						page.Date = t
					}
				} else if strings.HasPrefix(line, "excerpt:") {
					page.Description = strings.Trim(strings.TrimPrefix(line, "excerpt:"), ` "`)
				} else if strings.HasPrefix(line, "keywords:") {
					kwStr := strings.Trim(strings.TrimPrefix(line, "keywords:"), ` []`)
					for _, k := range strings.Split(kwStr, ",") {
						k = strings.TrimSpace(k)
						if k != "" {
							page.Keywords = append(page.Keywords, k)
						}
					}
				}
			}
		}
	}

	if page.Title == "" {
		// Derive title from first heading if no frontmatter
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				page.Title = strings.TrimPrefix(line, "# ")
				break
			}
		}
	}

	// Convert markdown body to basic HTML
	page.Body = markdownToHTML(content)
	page.URL = slugify(page.Title) + ".html"

	return page, nil
}

// markdownToHTML converts basic markdown to HTML (no external dependencies)
func markdownToHTML(md string) string {
	var buf strings.Builder
	lines := strings.Split(md, "\n")
	inCodeBlock := false
	inList := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code blocks
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				buf.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				buf.WriteString("<pre><code>")
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			buf.WriteString(html.EscapeString(line) + "\n")
			continue
		}

		// Headings
		if strings.HasPrefix(trimmed, "### ") {
			closeList(&buf, &inList)
			buf.WriteString(fmt.Sprintf("<h3>%s</h3>\n", html.EscapeString(strings.TrimPrefix(trimmed, "### "))))
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			closeList(&buf, &inList)
			buf.WriteString(fmt.Sprintf("<h2>%s</h2>\n", html.EscapeString(strings.TrimPrefix(trimmed, "## "))))
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			closeList(&buf, &inList)
			buf.WriteString(fmt.Sprintf("<h1>%s</h1>\n", html.EscapeString(strings.TrimPrefix(trimmed, "# "))))
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			closeList(&buf, &inList)
			buf.WriteString("<hr>\n")
			continue
		}

		// Unordered list
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				buf.WriteString("<ul>\n")
				inList = true
			}
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.TrimPrefix(item, "* ")
			buf.WriteString(fmt.Sprintf("  <li>%s</li>\n", inlineMarkdown(item)))
			continue
		}

		// Ordered list
		if len(trimmed) > 0 && trimmed[0] >= '1' && trimmed[0] <= '9' && strings.Contains(trimmed, ". ") {
			if !inList {
				buf.WriteString("<ol>\n")
				inList = true
			}
			idx := strings.Index(trimmed, ". ")
			if idx > 0 && idx < 3 {
				item := trimmed[idx+2:]
				buf.WriteString(fmt.Sprintf("  <li>%s</li>\n", inlineMarkdown(item)))
				continue
			}
		}

		// Blockquote
		if strings.HasPrefix(trimmed, "> ") {
			closeList(&buf, &inList)
			buf.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", inlineMarkdown(strings.TrimPrefix(trimmed, "> "))))
			continue
		}

		// Paragraph (non-empty line)
		if trimmed != "" {
			closeList(&buf, &inList)
			buf.WriteString(fmt.Sprintf("<p>%s</p>\n", inlineMarkdown(trimmed)))
			continue
		}

		// Empty line
		closeList(&buf, &inList)
	}

	closeList(&buf, &inList)

	return buf.String()
}

func closeList(buf *strings.Builder, inList *bool) {
	if *inList {
		buf.WriteString("</ul>\n")
		*inList = false
	}
}

// inlineMarkdown handles bold, italic, links, and code within a line
func inlineMarkdown(text string) string {
	// Escape HTML first
	text = html.EscapeString(text)

	// Code: `code`
	for strings.Contains(text, "`") {
		start := strings.Index(text, "`")
		end := strings.Index(text[start+1:], "`")
		if end < 0 {
			break
		}
		code := text[start+1 : start+1+end]
		text = text[:start] + "<code>" + code + "</code>" + text[start+2+end:]
	}

	// Bold: **text**
	for strings.Contains(text, "**") {
		start := strings.Index(text, "**")
		end := strings.Index(text[start+2:], "**")
		if end < 0 {
			break
		}
		bold := text[start+2 : start+2+end]
		text = text[:start] + "<strong>" + bold + "</strong>" + text[start+4+end:]
	}

	// Links: [text](url)
	for strings.Contains(text, "](") {
		start := strings.Index(text, "[")
		if start < 0 {
			break
		}
		mid := strings.Index(text[start:], "](")
		if mid < 0 {
			break
		}
		mid = start + mid
		end := strings.Index(text[mid:], ")")
		if end < 0 {
			break
		}
		end = mid + end
		linkText := text[start+1 : mid]
		linkURL := text[mid+2 : end]
		text = text[:start] + fmt.Sprintf(`<a href="%s">%s</a>`, linkURL, linkText) + text[end+1:]
	}

	return text
}

// slugify converts a title to a URL-safe slug
func slugify(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "\"", "")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "!", "")

	var clean strings.Builder
	for _, ch := range slug {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			clean.WriteRune(ch)
		}
	}
	slug = clean.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) > 80 {
		slug = slug[:80]
	}
	return slug
}

// generateCSS creates the site stylesheet
func generateCSS(theme string) string {
	bg, text, accent, card, cardBg := "#0d1117", "#e6edf3", "#58a6ff", "#161b22", "#161b22"
	if theme == "light" {
		bg, text, accent, card, cardBg = "#ffffff", "#24292f", "#0969da", "#f6f8fa", "#f6f8fa"
	}

	return fmt.Sprintf(`/* AI Money Machine — Auto-generated site */
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  background: %s;
  color: %s;
  line-height: 1.6;
  max-width: 860px;
  margin: 0 auto;
  padding: 20px;
}

a { color: %s; text-decoration: none; }
a:hover { text-decoration: underline; }

header {
  text-align: center;
  padding: 40px 0 30px;
  border-bottom: 1px solid %s;
  margin-bottom: 30px;
}

header h1 { font-size: 1.8em; margin-bottom: 8px; }
header p { color: #8b949e; font-size: 0.95em; }

.articles { list-style: none; }

.article-card {
  background: %s;
  border: 1px solid %s;
  border-radius: 8px;
  padding: 20px 24px;
  margin-bottom: 16px;
  transition: border-color 0.2s;
}

.article-card:hover { border-color: %s; }

.article-card h2 { font-size: 1.2em; margin-bottom: 8px; }
.article-card .meta {
  font-size: 0.85em;
  color: #8b949e;
  margin-bottom: 8px;
}
.article-card .meta span { margin-right: 12px; }
.article-card p { font-size: 0.95em; color: #8b949e; }
.article-card .tags { margin-top: 10px; }
.article-card .tag {
  display: inline-block;
  background: %s;
  color: %s;
  font-size: 0.8em;
  padding: 2px 10px;
  border-radius: 12px;
  margin-right: 6px;
}

/* Article pages */
article { padding: 20px 0; }
article h1 { font-size: 2em; margin: 24px 0 12px; }
article h2 { font-size: 1.4em; margin: 20px 0 10px; padding-bottom: 6px; border-bottom: 1px solid %s; }
article h3 { font-size: 1.15em; margin: 16px 0 8px; }
article p { margin: 12px 0; }
article ul, article ol { margin: 12px 0; padding-left: 24px; }
article li { margin: 4px 0; }
article blockquote {
  border-left: 3px solid %s;
  padding: 8px 16px;
  margin: 12px 0;
  color: #8b949e;
  background: %s;
  border-radius: 0 6px 6px 0;
}
article pre {
  background: %s;
  border: 1px solid %s;
  border-radius: 6px;
  padding: 16px;
  overflow-x: auto;
  font-size: 0.9em;
  margin: 12px 0;
}
article code {
  background: %s;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
}
article pre code { background: none; padding: 0; }
article hr { border: none; border-top: 1px solid %s; margin: 24px 0; }

.back { display: inline-block; margin-bottom: 20px; font-size: 0.9em; }
.back::before { content: "← "; }

footer {
  text-align: center;
  padding: 30px 0;
  color: #8b949e;
  font-size: 0.85em;
  border-top: 1px solid %s;
  margin-top: 40px;
}

@media (max-width: 600px) {
  body { padding: 12px; }
  header h1 { font-size: 1.4em; }
}
`, bg, text, accent, card, cardBg, card, accent, card, accent, card, accent, card, card, card, card, card, card)
}

// generateIndexHTML creates the main listing page
func generateIndexHTML(pages []SitePage, config SiteConfig) string {
	var articlesHTML strings.Builder

	for _, page := range pages {
		dateStr := page.Date.Format("Jan 2, 2006")
		typeBadge := page.Type

		var tagsHTML string
		for _, kw := range page.Keywords {
			tagsHTML += fmt.Sprintf(`<span class="tag">%s</span> `, html.EscapeString(kw))
		}

		articlesHTML.WriteString(fmt.Sprintf(`
  <li class="article-card">
    <h2><a href="%s">%s</a></h2>
    <div class="meta">
      <span>📅 %s</span>
      <span>🏷 %s</span>
    </div>
    <p>%s</p>
    <div class="tags">%s</div>
  </li>`,
			page.URL, html.EscapeString(page.Title),
			dateStr, typeBadge,
			html.EscapeString(page.Description),
			tagsHTML,
		))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <meta name="description" content="%s">
  <link rel="alternate" type="application/rss+xml" title="%s" href="feed.xml">
  <link rel="stylesheet" href="assets/style.css">
</head>
<body>
  <header>
    <h1>🏭 %s</h1>
    <p>%s — %d articles</p>
    <p style="margin-top:8px"><a href="feed.xml">📡 RSS Feed</a></p>
  </header>

  <ul class="articles">
%s  </ul>

  <footer>
    <p>Generated autonomously on %s by %s</p>
    <p>All content CC0 — free to use, share, remix</p>
  </footer>
</body>
</html>`,
		html.EscapeString(config.Title),
		html.EscapeString(config.Description),
		html.EscapeString(config.Title),
		html.EscapeString(config.Title),
		html.EscapeString(config.Description),
		len(pages),
		articlesHTML.String(),
		time.Now().Format("January 2, 2006"),
		html.EscapeString(config.Author),
	)
}

// generateArticleHTML creates a single article page
func generateArticleHTML(page SitePage, config SiteConfig) string {
	dateStr := page.Date.Format("Jan 2, 2006")
	var keywordsStr string
	if len(page.Keywords) > 0 {
		keywordsStr = strings.Join(page.Keywords, ", ")
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s — %s</title>
  <meta name="description" content="%s">
  <meta name="keywords" content="%s">
  <link rel="stylesheet" href="assets/style.css">
</head>
<body>
  <a class="back" href="index.html">Back to Library</a>
  <article>
    <h1>%s</h1>
    <div class="meta" style="color:#8b949e;margin-bottom:20px">
      <span>📅 %s</span>
      <span>🏷 %s</span>
    </div>
    %s
  </article>

  <footer>
    <p>Generated autonomously on %s by %s</p>
    <p><a href="index.html">← Back to Library</a></p>
  </footer>
</body>
</html>`,
		html.EscapeString(page.Title), html.EscapeString(config.Title),
		html.EscapeString(page.Description),
		html.EscapeString(keywordsStr),
		html.EscapeString(page.Title),
		dateStr, page.Type,
		page.Body,
		time.Now().Format("January 2, 2006"),
		html.EscapeString(config.Author),
	)
}

// generateRSSFeed creates an RSS 2.0 feed
func generateRSSFeed(pages []SitePage, config SiteConfig) string {
	var items strings.Builder

	for _, page := range pages {
		url := config.BaseURL + "/" + page.URL
		dateStr := page.Date.Format(time.RFC1123Z)
		description := html.EscapeString(page.Description)

		items.WriteString(fmt.Sprintf(`    <item>
      <title>%s</title>
      <link>%s</link>
      <guid isPermaLink="true">%s</guid>
      <pubDate>%s</pubDate>
      <description>%s</description>
    </item>
`,
			html.EscapeString(page.Title), url, url, dateStr, description))
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>%s</title>
    <link>%s</link>
    <description>%s</description>
    <language>en-us</language>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/feed.xml" rel="self" type="application/rss+xml"/>
%s  </channel>
</rss>`,
		html.EscapeString(config.Title),
		config.BaseURL,
		html.EscapeString(config.Description),
		time.Now().Format(time.RFC1123Z),
		config.BaseURL,
		items.String(),
	)
}

// ServeSite starts a local HTTP server for the generated site
func ServeSite(siteDir string, port int) error {
	absDir, err := filepath.Abs(siteDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return fmt.Errorf("site directory does not exist: %s", absDir)
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("[Deploy] Serving site at http://localhost%s\n", addr)
	fmt.Printf("[Deploy] Directory: %s\n", absDir)

	handler := http.FileServer(http.Dir(absDir))
	return http.ListenAndServe(addr, handler)
}

// DeployToDir copies the generated site to a target directory
func DeployToDir(sourceDir, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(sourceDir, path)
		targetPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()

		target, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer target.Close()

		_, err = io.Copy(target, source)
		return err
	})
}
