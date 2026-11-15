package orchestrator

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// Identity handles decentralized Ed25519 node identification and mesh message signing.
type Identity struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"-"`
}

// NewIdentity generates or loads the persistent node identity from node.pub / node.key.
func NewIdentity() (*Identity, error) {
	return NewIdentityAt("node.pub", "node.key")
}

// NewIdentityAt generates or loads an Ed25519 keypair at explicit file paths.
func NewIdentityAt(pubFile, privFile string) (*Identity, error) {
	if _, err := os.Stat(privFile); err == nil {
		pubData, pubErr := os.ReadFile(pubFile)
		privData, privErr := os.ReadFile(privFile)
		if pubErr == nil && privErr == nil && len(strings.TrimSpace(string(privData))) > 0 {
			return &Identity{
				PublicKey:  strings.TrimSpace(string(pubData)),
				PrivateKey: strings.TrimSpace(string(privData)),
			}, nil
		}
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(pub)
	privHex := hex.EncodeToString(priv)

	_ = os.WriteFile(pubFile, []byte(pubHex), 0644)
	_ = os.WriteFile(privFile, []byte(privHex), 0600)

	return &Identity{
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}, nil
}

// Sign creates an Ed25519 hex signature for message.
func (i *Identity) Sign(message string) string {
	privBytes, err := hex.DecodeString(i.PrivateKey)
	if err != nil || len(privBytes) != ed25519.PrivateKeySize {
		return ""
	}
	sig := ed25519.Sign(privBytes, []byte(message))
	return hex.EncodeToString(sig)
}

// Verify checks an Ed25519 node signature.
func Verify(pubKeyHex, message, sigHex string) bool {
	pubBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		return false
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil || len(sigBytes) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(pubBytes, []byte(message), sigBytes)
}

// GetDID returns the W3C-style DID identifier for this mesh node.
func (i *Identity) GetDID() string {
	if i == nil || i.PublicKey == "" {
		return "did:hustle:ephemeral"
	}
	return fmt.Sprintf("did:hustle:%s", i.PublicKey)
}
