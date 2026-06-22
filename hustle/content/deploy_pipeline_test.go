package content

import "testing"

func TestDeploySite_UnknownTarget(t *testing.T) {
	if err := DeploySite("./doesnotmatter", "invalid"); err == nil {
		t.Fatalf("expected error for unknown target, got nil")
	}
}
