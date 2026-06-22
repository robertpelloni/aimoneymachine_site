package publisher

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/hustle/social"
	"github.com/robertpelloni/hustle/orchestrator"
)

// ContentCalendar manages automated content scheduling and publishing
type ContentCalendar struct {
	Entries       []CalendarEntry
	WordPress     *WordPressPublisher
	Newsletter    *NewsletterPublisher
	Affiliate     *AffiliateInserter
	Twitter       social.Provider
	LinkedIn      social.Provider
	ScheduleFile  string
}

// CalendarEntry represents a scheduled content piece
type CalendarEntry struct {
	ID          string    `json:"id"`
	Topic       string    `json:"topic"`
	Type        string    `json:"type"` // blog, newsletter, thread, research
	Status      string    `json:"status"` // pending, generating, ready, published
	Platform    string    `json:"platform"` // wordpress, twitter, newsletter, all
	ScheduledAt time.Time `json:"scheduled_at"`
	Content     string    `json:"content,omitempty"`
	Title       string    `json:"title,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// NewContentCalendar creates a new content calendar
func NewContentCalendar() *ContentCalendar {
	// Social providers need DryRun config from Orchestrator, handled in PublishEntry
	return &ContentCalendar{
		Entries:    make([]CalendarEntry, 0),
		WordPress:  NewWordPressPublisher(),
		Newsletter: NewNewsletterPublisher(),
		Affiliate:  NewAffiliateInserter(),
		Twitter: social.NewTwitterProvider(
			os.Getenv("TWITTER_API_KEY"),
			os.Getenv("TWITTER_API_SECRET"),
			os.Getenv("TWITTER_ACCESS_TOKEN"),
			os.Getenv("TWITTER_ACCESS_SECRET"),
		),
		LinkedIn: social.NewLinkedInProvider(
			os.Getenv("LINKEDIN_ACCESS_TOKEN"),
			os.Getenv("LINKEDIN_AUTHOR_URN"),
		),
		ScheduleFile: "content_calendar.json",
	}
}

// LoadCalendar loads the calendar from file
func (c *ContentCalendar) LoadCalendar() error {
	data, err := os.ReadFile(c.ScheduleFile)
	if err != nil {
		return nil // No file yet, start fresh
	}
	return json.Unmarshal(data, &c.Entries)
}

// SaveCalendar saves the calendar to file
func (c *ContentCalendar) SaveCalendar() error {
	data, err := json.MarshalIndent(c.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.ScheduleFile, data, 0644)
}

// AddEntry adds a new entry to the calendar
func (c *ContentCalendar) AddEntry(topic, contentType, platform string, scheduledAt time.Time) CalendarEntry {
	entry := CalendarEntry{
		ID:          fmt.Sprintf("content-%d", time.Now().Unix()),
		Topic:       topic,
		Type:        contentType,
		Status:      "pending",
		Platform:    platform,
		ScheduledAt: scheduledAt,
	}
	c.Entries = append(c.Entries, entry)
	c.SaveCalendar()
	return entry
}

// GetPendingEntries returns entries that are ready to publish
func (c *ContentCalendar) GetPendingEntries() []CalendarEntry {
	var pending []CalendarEntry
	now := time.Now()

	for _, entry := range c.Entries {
		if entry.Status == "ready" && entry.ScheduledAt.Before(now) {
			pending = append(pending, entry)
		}
	}

	return pending
}

// PublishEntry publishes a calendar entry
func (c *ContentCalendar) PublishEntry(orch *orchestrator.Orchestrator, entry *CalendarEntry) error {
	if entry.Content == "" {
		return fmt.Errorf("no content to publish")
	}

	// Process with affiliate links
	content := c.Affiliate.ProcessContent(entry.Content)

	var err error

	switch entry.Platform {
	case "wordpress":
		err = c.publishToWordPress(entry, content)
	case "twitter":
		err = c.publishToTwitter(orch, entry, content)
	case "linkedin":
		err = c.publishToLinkedIn(orch, entry, content)
	case "newsletter":
		err = c.publishToNewsletter(entry, content)
	case "all":
		// Publish to all platforms
		errs := []string{}
		if e := c.publishToWordPress(entry, content); e != nil {
			errs = append(errs, fmt.Sprintf("WordPress: %v", e))
		}
		if e := c.publishToTwitter(orch, entry, content); e != nil {
			errs = append(errs, fmt.Sprintf("Twitter: %v", e))
		}
		if e := c.publishToLinkedIn(orch, entry, content); e != nil {
			errs = append(errs, fmt.Sprintf("LinkedIn: %v", e))
		}
		if e := c.publishToNewsletter(entry, content); e != nil {
			errs = append(errs, fmt.Sprintf("Newsletter: %v", e))
		}
		if len(errs) > 0 {
			err = fmt.Errorf("partial failures: %s", strings.Join(errs, "; "))
		}
	default:
		err = fmt.Errorf("unknown platform: %s", entry.Platform)
	}

	if err != nil {
		entry.Status = "failed"
		entry.Error = err.Error()
	} else {
		entry.Status = "published"
		entry.PublishedAt = time.Now()
	}

	c.SaveCalendar()
	return err
}

func (c *ContentCalendar) publishToWordPress(entry *CalendarEntry, content string) error {
	if !c.WordPress.IsConfigured() {
		fmt.Printf("[Calendar] WordPress not configured, skipping\n")
		return nil
	}

	resp, err := c.WordPress.PublishPost(entry.Title, content, "publish", nil, nil)
	if err != nil {
		return err
	}

	fmt.Printf("[Calendar] ✅ WordPress: %s\n", resp.Link)
	return nil
}

func (c *ContentCalendar) publishToTwitter(orch *orchestrator.Orchestrator, entry *CalendarEntry, content string) error {
	fmt.Printf("[Calendar] Posting to Twitter: %s\n", entry.Title)
	c.Twitter.SetDryRun(orch.DryRun)

	// Twitter threads are handled by the LLM usually, but for simple calendar posts, we use the provider directly
	return c.Twitter.Post(orch, "Twitter", content)
}

func (c *ContentCalendar) publishToLinkedIn(orch *orchestrator.Orchestrator, entry *CalendarEntry, content string) error {
	fmt.Printf("[Calendar] Posting to LinkedIn: %s\n", entry.Title)
	c.LinkedIn.SetDryRun(orch.DryRun)
	return c.LinkedIn.Post(orch, "LinkedIn", content)
}

func (c *ContentCalendar) publishToNewsletter(entry *CalendarEntry, content string) error {
	if !c.Newsletter.IsConfigured() {
		fmt.Printf("[Calendar] Newsletter not configured, skipping\n")
		return nil
	}

	return c.Newsletter.PublishNewsletter(entry.Title, content, []string{"ai", "automation", "hustle"})
}

// splitIntoTweets splits content into tweet-sized chunks
func splitIntoTweets(content, title string) []string {
	var tweets []string

	// First tweet is the hook
	tweets = append(tweets, fmt.Sprintf("🧵 %s\n\n(1/n)", title))

	// Split remaining content
	words := strings.Fields(content)
	current := ""
	tweetNum := 2

	for _, word := range words {
		if len(current)+len(word)+1 > 270 { // Leave room for thread number
			tweets = append(tweets, fmt.Sprintf("%s\n\n(%d/n)", current, tweetNum))
			tweetNum++
			current = word
		} else {
			if current != "" {
				current += " "
			}
			current += word
		}
	}

	if current != "" {
		tweets = append(tweets, fmt.Sprintf("%s\n\n(%d/n)", current, tweetNum))
	}

	return tweets
}

// GenerateWeeklySchedule generates a week's worth of content topics
func GenerateWeeklySchedule() []map[string]string {
	topics := []map[string]string{
		{"topic": "AI agent monetization strategies", "type": "blog", "platform": "wordpress"},
		{"topic": "Best local LLM models for automation", "type": "blog", "platform": "wordpress"},
		{"topic": "How to build a passive income with AI", "type": "newsletter", "platform": "newsletter"},
		{"topic": "Top 5 AI tools for content creators", "type": "blog", "platform": "wordpress"},
		{"topic": "Crypto trading with AI agents", "type": "thread", "platform": "twitter"},
		{"topic": "AI-powered SEO strategies for 2026", "type": "blog", "platform": "wordpress"},
		{"topic": "Weekly AI hustle digest", "type": "newsletter", "platform": "newsletter"},
	}

	return topics
}

// truncate shortens a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// PrintStatus prints the calendar status
func (c *ContentCalendar) PrintStatus() {
	pending := 0
	published := 0
	failed := 0

	for _, entry := range c.Entries {
		switch entry.Status {
		case "pending", "ready":
			pending++
		case "published":
			published++
		case "failed":
			failed++
		}
	}

	fmt.Printf("[Calendar] Status: %d pending, %d published, %d failed\n", pending, published, failed)
}