package research

import (
	"net/smtp"
	"strings"
	"testing"

	"github.com/robertpelloni/hustle/orchestrator"
)

type fakeSearcher struct{ results []SearchResult }

func (f fakeSearcher) Query(string) ([]SearchResult, error) { return f.results, nil }

func TestFindLeadsFromSearchMapsWithoutInventingEmails(t *testing.T) {
	lg := NewLeadGenerator(orchestrator.NewOrchestrator())
	leads, err := lg.FindLeadsFromSearch(fakeSearcher{results: []SearchResult{
		{Title: "Real Biz", URL: "https://real.example", Snippet: "does things"},
		{Title: "", URL: "https://skip.example", Snippet: "no title"},
	}}, "plumbing")
	if err != nil {
		t.Fatalf("FindLeadsFromSearch: %v", err)
	}
	if len(leads) != 1 {
		t.Fatalf("leads = %d, want 1", len(leads))
	}
	if leads[0].Email != "" {
		t.Errorf("email invented: %q — discovery must not fabricate addresses", leads[0].Email)
	}
	if leads[0].Source != "web-search" || leads[0].Website != "https://real.example" {
		t.Errorf("lead mapping wrong: %+v", leads[0])
	}
}

func TestFindLeadsAreLabeledDemo(t *testing.T) {
	lg := NewLeadGenerator(orchestrator.NewOrchestrator())
	leads, _ := lg.FindLeads("anything")
	for _, l := range leads {
		if l.Source != "demo-seed" {
			t.Errorf("seed lead %q source = %q, want demo-seed", l.Name, l.Source)
		}
	}
}

func TestSendEmailDryRunChargesNothing(t *testing.T) {
	t.Setenv("SMTP_HOST", "") // force dry-run path regardless of host env
	orch := orchestrator.NewOrchestrator()
	oc := NewOutreachCampaign(orch)
	before := orch.Ledger.TotalExpenses()
	if err := oc.SendEmail(Lead{Name: "X", Email: "a@b.test", Source: "demo-seed"}, "s", "b"); err != nil {
		t.Fatalf("dry-run SendEmail: %v", err)
	}
	if got := orch.Ledger.TotalExpenses(); got != before {
		t.Errorf("dry-run charged ledger: before %v after %v", before, got)
	}
}

func TestSendEmailRealPathDeliversThenCharges(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.test")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_FROM", "me@example.test")
	t.Setenv("SMTP_USER", "")
	t.Setenv("SMTP_PASS", "")

	var gotAddr, gotFrom string
	var gotTo []string
	var gotMsg []byte
	old := deliverSMTP
	deliverSMTP = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		gotAddr, gotFrom, gotTo, gotMsg = addr, from, to, msg
		return nil
	}
	defer func() { deliverSMTP = old }()

	orch := orchestrator.NewOrchestrator()
	oc := NewOutreachCampaign(orch)
	if err := oc.SendEmail(Lead{Name: "X", Email: "lead@biz.test", Source: "web-search"}, "Hi", "Body text"); err != nil {
		t.Fatalf("SendEmail: %v", err)
	}
	if gotAddr != "smtp.example.test:2525" || gotFrom != "me@example.test" || len(gotTo) != 1 || gotTo[0] != "lead@biz.test" {
		t.Errorf("delivery args: addr=%q from=%q to=%v", gotAddr, gotFrom, gotTo)
	}
	if !strings.Contains(string(gotMsg), "Subject: Hi") || !strings.Contains(string(gotMsg), "Body text") {
		t.Errorf("message malformed:\n%s", gotMsg)
	}
	if orch.Ledger.TotalExpenses() <= 0 {
		t.Error("successful delivery should charge the ledger")
	}
}

func TestSendEmailRefusesMissingAddressAndFailedSMTP(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.test")
	t.Setenv("SMTP_FROM", "me@example.test")
	orch := orchestrator.NewOrchestrator()
	oc := NewOutreachCampaign(orch)
	if err := oc.SendEmail(Lead{Name: "NoEmail", Source: "web-search"}, "s", "b"); err == nil {
		t.Error("missing email must be refused")
	}

	old := deliverSMTP
	deliverSMTP = func(string, smtp.Auth, string, []string, []byte) error {
		return errSMTPDown
	}
	defer func() { deliverSMTP = old }()
	before := orch.Ledger.TotalExpenses()
	if err := oc.SendEmail(Lead{Name: "X", Email: "a@b.test"}, "s", "b"); err == nil {
		t.Error("SMTP failure must return an error")
	}
	if orch.Ledger.TotalExpenses() != before {
		t.Error("failed delivery must not charge the ledger")
	}
}

var errSMTPDown = &smtpDownError{}

type smtpDownError struct{}

func (*smtpDownError) Error() string { return "smtp down" }
