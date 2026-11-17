package publisher

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ContentCalendar manages automated content scheduling and publishing.
// mu guards Entries: the concurrent scheduler can run the calendar job
// while chain steps and the API snapshot read the same slice.
type ContentCalendar struct {
	mu           sync.Mutex
	Entries      []CalendarEntry
	WordPress    *WordPressPublisher
	Newsletter   *NewsletterPublisher
	Affiliate    *AffiliateInserter
	Twitter      *TwitterPoster
	ScheduleFile string
}

// CalendarEntry represents a scheduled content piece
type CalendarEntry struct {
	ID          string    `json:"id"`
	Topic       string    `json:"topic"`
	Type        string    `json:"type"` // blog, newsletter, thread, research
	Status      string    `json:"status"` // pending, generating, ready, published, superseded
	Platform    string    `json:"platform"` // wordpress, twitter, newsletter, all
	ScheduledAt time.Time `json:"scheduled_at"`
	Content     string    `json:"content,omitempty"`
	Title       string    `json:"title,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// TwitterPoster wraps the social posting for the calendar
type TwitterPoster struct {
	BearerToken string
}

// NewContentCalendar creates a new content calendar
func NewContentCalendar() *ContentCalendar {
	return &ContentCalendar{
		Entries:      make([]CalendarEntry, 0),
		WordPress:    NewWordPressPublisher(),
		Newsletter:   NewNewsletterPublisher(),
		Affiliate:    NewAffiliateInserter(),
		Twitter:      &TwitterPoster{BearerToken: os.Getenv("TWITTER_BEARER_TOKEN")},
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
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := CalendarEntry{
		// UnixNano: second-resolution IDs collided when schedule_week
		// added 7 entries in the same second (prod calendar carried
		// 5-7 duplicates of three IDs, all content-less).
		ID:          fmt.Sprintf("content-%d", time.Now().UnixNano()),
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
func (c *ContentCalendar) PublishEntry(entry *CalendarEntry) error {
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
		err = c.publishToTwitter(entry, content)
	case "newsletter":
		err = c.publishToNewsletter(entry, content)
	case "all":
		// Publish to all platforms
		errs := []string{}
		if e := c.publishToWordPress(entry, content); e != nil {
			errs = append(errs, fmt.Sprintf("WordPress: %v", e))
		}
		if e := c.publishToTwitter(entry, content); e != nil {
			errs = append(errs, fmt.Sprintf("Twitter: %v", e))
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

func (c *ContentCalendar) publishToTwitter(entry *CalendarEntry, content string) error {
	if c.Twitter.BearerToken == "" {
		fmt.Printf("[Calendar] Twitter not configured, skipping\n")
		return nil
	}

	// Preview the thread in the log, then fail honestly: this path is
	// NOT wired to the X API (real posting lives in hustle/social), and
	// returning nil would mark the entry published — a post that never
	// happened, reported as real by the dashboard. The entry lands in
	// failed with this reason until the calendar is wired to the social
	// poster (X currently returns 402 credits-depleted regardless).
	tweets := splitIntoTweets(content, entry.Title)
	for i, tweet := range tweets {
		fmt.Printf("[Calendar] Tweet %d/%d: %s\n", i+1, len(tweets), truncate(tweet, 50))
	}

	return fmt.Errorf("calendar twitter posting is not wired to the X API (preview only); post via hustle://social?platform=Twitter")
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

// Snapshot returns a copy of the entries for read-only consumers (API).
func (c *ContentCalendar) Snapshot() []CalendarEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]CalendarEntry, len(c.Entries))
	copy(out, c.Entries)
	return out
}

// FillPending generates content for due topic-only shells (status pending,
// no content — what schedule_week creates). ProcessPending skips shells
// forever, so without this step they accumulate as permanently overdue
// pending entries (56 observed on prod, none publishable). gen is supplied
// by the caller (the content module) so this package stays LLM-free and
// testable. At most limit shells are generated per call — each fill is an
// LLM call, and the scheduler runs this every 90 minutes. A shell whose
// topic+platform already carries content (ready/published, or filled
// earlier in this same pass) is marked superseded instead of generating a
// duplicate post: schedule_week re-adds the same topics every run. A
// generator error leaves the shell pending for the next cycle and is
// returned. Returns the number of shells filled.
func (c *ContentCalendar) FillPending(limit int, gen func(CalendarEntry) (title, body string, err error)) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	covered := map[string]bool{}
	for _, e := range c.Entries {
		if strings.TrimSpace(e.Content) != "" || e.Status == "published" {
			covered[e.Topic+"|"+e.Platform] = true
		}
	}
	filled := 0
	changed := false
	var firstErr error
	for i := range c.Entries {
		e := &c.Entries[i]
		if e.Status != "pending" || strings.TrimSpace(e.Content) != "" || e.ScheduledAt.After(now) {
			continue
		}
		key := e.Topic + "|" + e.Platform
		if covered[key] {
			e.Status = "superseded"
			changed = true
			continue
		}
		if filled >= limit {
			continue
		}
		title, body, err := gen(*e)
		if err == nil && strings.TrimSpace(body) == "" {
			err = fmt.Errorf("generator returned empty content for %q", e.Topic)
		}
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if strings.TrimSpace(title) == "" {
			title = e.Topic
		}
		e.Title = title
		e.Content = body
		e.Status = "ready"
		covered[key] = true
		filled++
		changed = true
	}
	if changed {
		if err := c.SaveCalendar(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return filled, firstErr
}

// ProcessPending publishes every due entry (status pending or ready, scheduled
// at or before now) in place, so PublishEntry's SaveCalendar persists the
// published/failed status. GetPendingEntries returns copies, so looping over
// it would silently discard status updates — hence this index-based loop.
// Returns the number of entries attempted.
func (c *ContentCalendar) ProcessPending() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	attempted := 0
	var firstErr error
	for i := range c.Entries {
		e := &c.Entries[i]
		if (e.Status == "ready" || e.Status == "pending") && !e.ScheduledAt.After(now) {
			// Content-less entries are planned shells (schedule_week
			// creates them topic-only). They are not failures — skip
			// silently instead of failing every process cycle on
			// "no content to publish" (observed repeatedly on prod).
			if strings.TrimSpace(e.Content) == "" {
				continue
			}
			attempted++
			if err := c.PublishEntry(e); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return attempted, firstErr
}

// PrintStatus prints the calendar status
func (c *ContentCalendar) PrintStatus() {
	c.mu.Lock()
	defer c.mu.Unlock()
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