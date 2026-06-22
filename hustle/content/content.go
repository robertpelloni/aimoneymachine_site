package content

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/hustle/publisher"
	"github.com/robertpelloni/hustle/orchestrator"
)

// ContentType defines what kind of content to generate
type ContentType string

const (
	BlogPost    ContentType = "blog"
	Newsletter  ContentType = "newsletter"
	SEOArticle  ContentType = "seo"
	SocialThread ContentType = "thread"
	MicroTool   ContentType = "tool"
	VideoScript ContentType = "video"
	ReelScript  ContentType = "reel"
)

// ContentRequest specifies what content to generate
type ContentRequest struct {
	Topic       string      `json:"topic"`
	Type        ContentType `json:"type"`
	Keywords    []string    `json:"keywords"`
	TargetWords int         `json:"target_words"`
	Niche       string      `json:"niche"`
}

// ContentResult holds the generated content
type ContentResult struct {
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Excerpt  string    `json:"excerpt"`
	Keywords []string  `json:"keywords"`
	Type     ContentType `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Filepath string    `json:"filepath"`
}

// ContentModule generates various types of monetizable content
type ContentModule struct {
	Orch       *orchestrator.Orchestrator
	OutputDir  string // directory to save generated content
}

// NewContentModule creates a content generation hustle
func NewContentModule(orch *orchestrator.Orchestrator, outputDir string) *ContentModule {
	if outputDir == "" {
		outputDir = "output/content"
	}
	return &ContentModule{
		Orch:      orch,
		OutputDir: outputDir,
	}
}

// Generate produces content based on the request using the LLM
func (c *ContentModule) Generate(req ContentRequest) (*ContentResult, error) {
	fmt.Printf("[Content] Generating %s about: %s\n", req.Type, req.Topic)

	if req.TargetWords == 0 {
		req.TargetWords = 800
	}
	if req.Niche == "" {
		req.Niche = "AI & automation"
	}

	prompt := c.buildPrompt(req)

	body, err := c.Orch.LLM.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM content generation failed: %v", err)
	}

	// Process content with affiliate links and disclosure
	inserter := publisher.NewAffiliateInserter()
	body = inserter.ProcessContent(body)

	// Extract title from the first heading if present
	title := req.Topic
	lines := strings.Split(body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Create excerpt (first 200 chars of body, no markdown)
	cleanBody := body
	cleanBody = strings.ReplaceAll(cleanBody, "#", "")
	cleanBody = strings.ReplaceAll(cleanBody, "*", "")
	cleanBody = strings.ReplaceAll(cleanBody, "-", "")
	excerpt := cleanBody
	if len(excerpt) > 200 {
		excerpt = excerpt[:200] + "..."
	}

	result := &ContentResult{
		Title:     title,
		Body:      body,
		Excerpt:   excerpt,
		Keywords:  req.Keywords,
		Type:      req.Type,
		CreatedAt: time.Now(),
	}

	// Save to file
	if err := c.saveToFile(result); err != nil {
		fmt.Printf("[Content] Warning: failed to save file: %v\n", err)
	}

	// Record in memory
	c.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("content-%s-%d", req.Type, time.Now().Unix()),
		Content:   fmt.Sprintf("Generated %s: %s (%d words)", req.Type, title, len(strings.Fields(body))),
		BaseScore: 60.0,
		Timestamp: time.Now(),
		Tags:      []string{"content", string(req.Type), req.Niche},
	})

	// Record potential revenue
	c.Orch.Ledger.Add(orchestrator.Transaction{
		Amount: 0.50, // estimated value per content piece (ads/affiliates)
		Type:   orchestrator.Revenue,
		Hustle: "ContentGeneration",
		Note:   fmt.Sprintf("%s: %s", req.Type, title),
	})

	fmt.Printf("[Content] ✅ Generated %s: %s (%d words)\n", req.Type, title, len(strings.Fields(body)))
	return result, nil
}

// buildPrompt creates an LLM prompt optimized for the content type
func (c *ContentModule) buildPrompt(req ContentRequest) string {
	// Check for optimized prompt in L3 memory
	if entries := c.Orch.L3.Search("optimized-prompt-content"); len(entries) > 0 {
		optimized := entries[len(entries)-1].Content
		fmt.Printf("[Content] Using SELF-OPTIMIZED PROMPT for %s piece\n", req.Type)
		return strings.ReplaceAll(optimized, "[topic]", req.Topic)
	}

	keywordsStr := strings.Join(req.Keywords, ", ")

	switch req.Type {
	case BlogPost:
		return fmt.Sprintf(`Write a comprehensive blog post about "%s" for the niche: %s.

Requirements:
- Target ~%d words
- Include an engaging H1 title
- Use H2 and H3 subheadings for structure
- Include these SEO keywords naturally: %s
- End with a compelling call-to-action
- Write in an authoritative but approachable tone
- Include specific data points, examples, or case studies where possible

Format: Markdown`, req.Topic, req.Niche, req.TargetWords, keywordsStr)

	case Newsletter:
		return fmt.Sprintf(`Write a curated newsletter issue about "%s" for the niche: %s.

Requirements:
- Target ~%d words
- Start with a personal, engaging intro (2-3 sentences)
- Cover 3-5 key developments or insights
- For each item: headline + 2-3 sentence analysis
- Include actionable takeaways
- End with a "What to watch next" section
- Tone: insider knowledge, concise, high-signal

Format: Markdown`, req.Topic, req.Niche, req.TargetWords)

	case SEOArticle:
		return fmt.Sprintf(`Write an SEO-optimized article targeting the keyword: "%s".

Requirements:
- Target ~%d words
- Primary keyword: %s
- Secondary keywords: %s
- Include the primary keyword in: title, first paragraph, at least 2 H2 headings, and conclusion
- Use short paragraphs (2-3 sentences max) for readability
- Include a "Key Takeaways" bulleted section
- Add meta description (155 chars max) at the top as: <!-- meta: Your meta description -->
- Structure: Introduction → Problem → Solution → How-To → Takeaways → Conclusion
- Write for search intent: informational/comparison

Format: Markdown with HTML meta comment`, req.Topic, req.TargetWords, req.Topic, keywordsStr)

	case SocialThread:
		return fmt.Sprintf(`Write a viral Twitter/X thread about "%s" for the niche: %s.

Requirements:
- 5-8 tweets in the thread
- First tweet must be a hook (controversial or surprising take)
- Each tweet: max 280 characters
- Use line breaks for readability
- Include relevant hashtags (2-3 per tweet)
- End with a CTA tweet (follow, subscribe, link)
- Tone: thought-provoking, insider perspective, no fluff

Format: Number each tweet like 1/ 2/ etc.`, req.Topic, req.Niche)

	case VideoScript:
		return fmt.Sprintf(`Write a high-retention YouTube video script for: "%s".

Requirements:
- Niche: %s
- Structure: Hook (0-15s) → Intro → Core Content (3-5 points) → Summary → CTA
- Include visual cues in [brackets] for the editor
- Include timestamps for a 8-10 minute video
- Write an SEO-optimized title, description (with timestamps), and tags
- Tone: engaging, educational, energetic

Format: Markdown with sections for Title, Script, Description, and Tags.`, req.Topic, req.Niche)

	case ReelScript:
		return fmt.Sprintf(`Write a high-impact TikTok/Reel script for: "%s".

Requirements:
- Niche: %s
- Duration: 30-60 seconds
- Structure: Immediate Hook → 3 quick tips/points → Call to action
- Include on-screen text overlays instructions
- Include music/SFX suggestions
- Write a catchy caption with trending hashtags
- Tone: fast-paced, entertaining, relatable

Format: Markdown with sections for Hook, On-Screen Text, SFX, Caption.`, req.Topic, req.Niche)

	case MicroTool:
		return fmt.Sprintf(`Generate a self-contained, single-file HTML/JavaScript utility tool for: "%s".

Requirements:
- Niche: %s
- The tool must be useful and solve a specific problem (e.g. calculator, converter, generator)
- Include CSS for a modern, dark-themed UI
- All logic must be in the same file
- Include SEO-friendly text description above and below the tool
- Add a footer with an affiliate link call-to-action

Format: Complete HTML file`, req.Topic, req.Niche)

	default:
		return fmt.Sprintf(`Write an informative article about "%s" (~%d words). Keywords: %s. Format: Markdown.`,
			req.Topic, req.TargetWords, keywordsStr)
	}
}

// saveToFile writes generated content to a markdown file
func (c *ContentModule) saveToFile(result *ContentResult) error {
	// Ensure output directory exists
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return err
	}

	// Create filename from title
	slug := strings.ToLower(result.Title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "")
	slug = strings.ReplaceAll(slug, ":", "")
	// Keep only alphanumeric and hyphens
	var cleanSlug strings.Builder
	for _, ch := range slug {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			cleanSlug.WriteRune(ch)
		}
	}
	slug = cleanSlug.String()
	if len(slug) > 60 {
		slug = slug[:60]
	}

	filename := fmt.Sprintf("%s-%s-%s.md", result.Type, slug, time.Now().Format("2006-01-02"))
	fullPath := filepath.Join(c.OutputDir, filename)

	// Add frontmatter
	frontmatter := fmt.Sprintf("---\ntitle: %q\ntype: %s\ndate: %s\nkeywords: [%s]\nexcerpt: %q\n---\n\n",
		result.Title, result.Type, result.CreatedAt.Format("2006-01-02"), strings.Join(result.Keywords, ", "), result.Excerpt)

	content := frontmatter + result.Body

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return err
	}

	result.Filepath = fullPath
	fmt.Printf("[Content] Saved to: %s\n", fullPath)
	return nil
}

// GenerateTopicIdeas uses the LLM to brainstorm content topics for a niche
func (c *ContentModule) GenerateTopicIdeas(niche string, count int) ([]string, error) {
	prompt := fmt.Sprintf(`Generate %d specific, trending content topic ideas for the niche: "%s".

Requirements:
- Each topic should target high-search-volume keywords
- Topics should be timely and relevant in 2026
- Mix of how-to, listicle, opinion, and news analysis formats
- Each topic should have clear monetization potential (ads, affiliates, or premium content)

Respond with a JSON array of strings. Example: ["topic 1", "topic 2", ...]

Respond ONLY with the JSON array, no other text.`, count, niche)

	topics, err := c.Orch.LLM.Generate(prompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON array
	topics = strings.TrimSpace(topics)
	if strings.HasPrefix(topics, "```") {
		start := strings.Index(topics, "[")
		end := strings.LastIndex(topics, "]")
		if start >= 0 && end > start {
			topics = topics[start : end+1]
		}
	}

	var topicList []string
	// Simple parsing: split by commas if JSON parsing fails
	if err := parseJSONStrings(topics, &topicList); err != nil {
		// Fallback: split by newlines
		lines := strings.Split(topics, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			line = strings.Trim(line, `"-[],`)
			if line != "" && len(line) > 5 {
				topicList = append(topicList, line)
			}
		}
	}

	return topicList, nil
}

// GenerateGallery creates a stylized HTML index of all generated content
func (c *ContentModule) GenerateGallery() (string, error) {
	files, err := os.ReadDir(c.OutputDir)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><title>AI Hustle Machine Content Gallery</title>")
	sb.WriteString("<style>body{font-family:sans-serif;line-height:1.6;max-width:800px;margin:2em auto;background:#f4f4f9;color:#333;}")
	sb.WriteString(".card{background:#fff;padding:1.5em;margin-bottom:1.5em;border-radius:8px;box-shadow:0 2px 5px rgba(0,0,0,0.1);}")
	sb.WriteString("h1{color:#2c3e50;} .tag{display:inline-block;background:#3498db;color:#fff;padding:0.2em 0.6em;border-radius:4px;font-size:0.8em;margin-right:0.5em;}")
	sb.WriteString(".meta{color:#7f8c8d;font-size:0.9em;margin-bottom:1em;}</style></head><body>")
	sb.WriteString("<h1>🚀 Generated Content Gallery</h1>")

	count := 0
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
			continue
		}

		path := filepath.Join(c.OutputDir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		contentStr := string(data)
		title := f.Name()
		createdAt := "Unknown"
		contentType := "article"
		excerpt := ""

		// Simple frontmatter parsing
		lines := strings.Split(contentStr, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "title: ") {
				title = strings.Trim(strings.TrimPrefix(line, "title: "), `"'`)
			} else if strings.HasPrefix(line, "date: ") {
				createdAt = strings.TrimSpace(strings.TrimPrefix(line, "date: "))
			} else if strings.HasPrefix(line, "type: ") {
				contentType = strings.TrimSpace(strings.TrimPrefix(line, "type: "))
			} else if strings.HasPrefix(line, "excerpt: ") {
				excerpt = strings.Trim(strings.TrimPrefix(line, "excerpt: "), `"'`)
			}
		}

		sb.WriteString("<div class='card'>")
		sb.WriteString(fmt.Sprintf("<h2>%s</h2>", title))
		sb.WriteString(fmt.Sprintf("<div class='meta'><span class='tag'>%s</span> %s</div>", contentType, createdAt))
		if excerpt != "" {
			sb.WriteString(fmt.Sprintf("<p>%s</p>", excerpt))
		}
		sb.WriteString(fmt.Sprintf("<a href='%s'>Read Markdown</a>", f.Name()))
		sb.WriteString("</div>")
		count++
	}

	if count == 0 {
		sb.WriteString("<p>No content generated yet. Launch a hustle to start!</p>")
	}

	sb.WriteString("</body></html>")

	galleryPath := filepath.Join(c.OutputDir, "index.html")
	if err := os.WriteFile(galleryPath, []byte(sb.String()), 0644); err != nil {
		return "", err
	}

	return galleryPath, nil
}

// Quick helper for simple JSON string array parsing
func parseJSONStrings(data string, target *[]string) error {
	data = strings.TrimSpace(data)
	if !strings.HasPrefix(data, "[") {
		return fmt.Errorf("not a JSON array")
	}
	
	// Remove brackets
	data = strings.TrimPrefix(data, "[")
	data = strings.TrimSuffix(data, "]")
	
	parts := strings.Split(data, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			*target = append(*target, part)
		}
	}
	return nil
}
