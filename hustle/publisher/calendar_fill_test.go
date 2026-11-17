package publisher

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// schedule_week creates topic-only shells (status pending, no content).
// ProcessPending skips them forever, so without a fill step they pile up
// as permanently overdue "pending" entries (56 observed on prod). FillPending
// generates content for due shells, capped per call, and supersedes shells
// whose topic+platform already carry content so repeated weekly schedules
// do not publish the same post over and over.
func TestFillPendingFillsCapsAndSupersedes(t *testing.T) {
	t.Chdir(t.TempDir())
	cal := NewContentCalendar()
	cal.ScheduleFile = filepath.Join(t.TempDir(), "cal.json")
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	cal.Entries = []CalendarEntry{
		{ID: "a", Topic: "AI agents", Type: "blog", Status: "pending", Platform: "wordpress", ScheduledAt: past},
		{ID: "b", Topic: "Local LLMs", Type: "blog", Status: "pending", Platform: "wordpress", ScheduledAt: past},
		{ID: "a-dup", Topic: "AI agents", Type: "blog", Status: "pending", Platform: "wordpress", ScheduledAt: past},
		{ID: "later", Topic: "Future topic", Type: "blog", Status: "pending", Platform: "wordpress", ScheduledAt: future},
		{ID: "done", Topic: "Old post", Type: "blog", Status: "published", Platform: "wordpress", ScheduledAt: past, Content: "x"},
		{ID: "old-dup", Topic: "Old post", Type: "blog", Status: "pending", Platform: "wordpress", ScheduledAt: past},
	}
	gen := func(e CalendarEntry) (string, string, error) {
		return "Title: " + e.Topic, "body about " + e.Topic, nil
	}
	filled, err := cal.FillPending(1, gen)
	if err != nil {
		t.Fatalf("FillPending: %v", err)
	}
	if filled != 1 {
		t.Fatalf("filled = %d, want 1 (cap)", filled)
	}
	byID := map[string]CalendarEntry{}
	for _, e := range cal.Entries {
		byID[e.ID] = e
	}
	if byID["a"].Status != "ready" || byID["a"].Content == "" || byID["a"].Title == "" {
		t.Errorf("a = %+v, want ready with title+content", byID["a"])
	}
	if byID["b"].Status != "pending" {
		t.Errorf("b status = %q, want pending (over cap)", byID["b"].Status)
	}
	if byID["a-dup"].Status != "superseded" {
		t.Errorf("a-dup status = %q, want superseded (same topic+platform filled)", byID["a-dup"].Status)
	}
	if byID["old-dup"].Status != "superseded" {
		t.Errorf("old-dup status = %q, want superseded (topic already published)", byID["old-dup"].Status)
	}
	if byID["later"].Status != "pending" {
		t.Errorf("later status = %q, want pending (not due)", byID["later"].Status)
	}
}

// A generator failure must leave the shell pending for the next cycle and
// surface the error, never mark it ready or failed.
func TestFillPendingGeneratorErrorLeavesShellPending(t *testing.T) {
	t.Chdir(t.TempDir())
	cal := NewContentCalendar()
	cal.ScheduleFile = filepath.Join(t.TempDir(), "cal.json")
	past := time.Now().Add(-time.Hour)
	cal.Entries = []CalendarEntry{
		{ID: "a", Topic: "AI agents", Type: "blog", Status: "pending", Platform: "wordpress", ScheduledAt: past},
	}
	gen := func(e CalendarEntry) (string, string, error) {
		return "", "", errors.New("llm down")
	}
	filled, err := cal.FillPending(2, gen)
	if err == nil {
		t.Fatal("FillPending error = nil, want generator error surfaced")
	}
	if filled != 0 {
		t.Fatalf("filled = %d, want 0", filled)
	}
	if cal.Entries[0].Status != "pending" || cal.Entries[0].Content != "" {
		t.Errorf("entry = %+v, want untouched pending shell", cal.Entries[0])
	}
}
