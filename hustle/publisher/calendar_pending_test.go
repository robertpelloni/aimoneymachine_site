package publisher

import (
	"path/filepath"
	"testing"
	"time"
)

// Content-less shells (what schedule_week creates) must be skipped
// silently, not failed every cycle with "no content to publish".
func TestProcessPendingSkipsContentlessShells(t *testing.T) {
	t.Chdir(t.TempDir())
	cal := NewContentCalendar()
	cal.ScheduleFile = filepath.Join(t.TempDir(), "cal.json")
	cal.WordPress = &WordPressPublisher{} // unconfigured -> no-op success
	past := time.Now().Add(-time.Hour)
	cal.Entries = []CalendarEntry{
		{ID: "shell", Topic: "t", Status: "pending", Platform: "wordpress", ScheduledAt: past},
		{ID: "real", Topic: "t", Status: "pending", Platform: "wordpress", ScheduledAt: past, Content: "body", Title: "T"},
	}
	n, err := cal.ProcessPending()
	if err != nil {
		t.Fatalf("ProcessPending: %v", err)
	}
	if n != 1 {
		t.Fatalf("attempted = %d, want 1 (shell skipped)", n)
	}
	if cal.Entries[0].Status != "pending" {
		t.Errorf("shell status = %q, want untouched pending", cal.Entries[0].Status)
	}
	if cal.Entries[1].Status != "published" {
		t.Errorf("real entry status = %q, want published", cal.Entries[1].Status)
	}
}

// ProcessPending must publish due entries in place: looping over the copies
// from GetPendingEntries would discard the published status before
// SaveCalendar persists it. WordPress is unconfigured in tests, so a
// wordpress entry with content publishes as a no-op success.
func TestProcessPendingPublishesDueEntryInPlace(t *testing.T) {
	t.Chdir(t.TempDir()) // calendar/affiliate files are cwd-relative; keep them out of the repo
	cal := NewContentCalendar()
	cal.ScheduleFile = filepath.Join(t.TempDir(), "cal.json")
	cal.WordPress = &WordPressPublisher{} // unconfigured -> skip, no error
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	cal.Entries = []CalendarEntry{
		{ID: "due", Topic: "t", Status: "ready", Platform: "wordpress", ScheduledAt: past, Content: "body", Title: "T"},
		{ID: "later", Topic: "t", Status: "ready", Platform: "wordpress", ScheduledAt: future, Content: "body", Title: "T"},
	}
	n, err := cal.ProcessPending()
	if err != nil {
		t.Fatalf("ProcessPending error: %v", err)
	}
	if n != 1 {
		t.Fatalf("attempted = %d, want 1", n)
	}
	if cal.Entries[0].Status != "published" {
		t.Fatalf("due entry status = %q, want published", cal.Entries[0].Status)
	}
	if cal.Entries[1].Status != "ready" {
		t.Fatalf("future entry status = %q, want ready (untouched)", cal.Entries[1].Status)
	}
}
