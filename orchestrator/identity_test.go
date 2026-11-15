package orchestrator

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestIdentityGenerateLoadAndVerify(t *testing.T) {
	dir := t.TempDir()
	pub := filepath.Join(dir, "node.pub")
	priv := filepath.Join(dir, "node.key")

	id1, err := NewIdentityAt(pub, priv)
	if err != nil {
		t.Fatalf("NewIdentityAt: %v", err)
	}
	if !strings.HasPrefix(id1.GetDID(), "did:hustle:") {
		t.Fatalf("unexpected DID: %s", id1.GetDID())
	}

	sig := id1.Sign("hello-mesh")
	if !Verify(id1.PublicKey, "hello-mesh", sig) {
		t.Fatal("Verify rejected valid signature")
	}
	if Verify(id1.PublicKey, "tampered", sig) {
		t.Fatal("Verify accepted tampered message")
	}

	id2, err := NewIdentityAt(pub, priv)
	if err != nil {
		t.Fatalf("reload NewIdentityAt: %v", err)
	}
	if id2.PublicKey != id1.PublicKey {
		t.Fatalf("reloaded pubkey mismatch: %s != %s", id2.PublicKey, id1.PublicKey)
	}
}
