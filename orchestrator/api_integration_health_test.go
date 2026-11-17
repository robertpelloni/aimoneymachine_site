package orchestrator

import "testing"

// A configured integration whose last observed outcome was an
// auth/payment rejection must not report OK (the Twitter dot stayed
// green through weeks of X 402 responses).
func TestIntegrationOK(t *testing.T) {
	cases := []struct {
		configured bool
		health     string
		want       bool
	}{
		{true, "", true},
		{true, "ok", true},
		{true, "auth_failed", false},
		{false, "", false},
		{false, "ok", false},
	}
	for _, c := range cases {
		if got := integrationOK(c.configured, c.health); got != c.want {
			t.Errorf("integrationOK(%v, %q) = %v, want %v", c.configured, c.health, got, c.want)
		}
	}

	o := NewOrchestrator()
	if got := o.IntegrationHealth("twitter"); got != "" {
		t.Errorf("unrecorded health = %q, want empty", got)
	}
	o.SetIntegrationHealth("Twitter", "auth_failed")
	if got := o.IntegrationHealth("twitter"); got != "auth_failed" {
		t.Errorf("health = %q, want auth_failed (case-insensitive)", got)
	}
	o.SetIntegrationHealth("twitter", "ok")
	if got := o.IntegrationHealth("TWITTER"); got != "ok" {
		t.Errorf("health after ok = %q, want ok", got)
	}
}
