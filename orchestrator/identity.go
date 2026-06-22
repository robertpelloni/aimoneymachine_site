package orchestrator

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

// Identity handles node identification and mesh security
type Identity struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"-"`
}

// NewIdentity generates or loads a persistent node identity
func NewIdentity() (*Identity, error) {
	pubFile := "node.pub"
	privFile := "node.key"

	if _, err := os.Stat(privFile); err == nil {
		// Load existing
		pubData, _ := os.ReadFile(pubFile)
		privData, _ := os.ReadFile(privFile)
		return &Identity{
			PublicKey:  string(pubData),
			PrivateKey: string(privData),
		}, nil
	}

	// Generate new Ed25519 keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pubHex := hex.EncodeToString(pub)
	privHex := hex.EncodeToString(priv)

	os.WriteFile(pubFile, []byte(pubHex), 0644)
	os.WriteFile(privFile, []byte(privHex), 0600)

	return &Identity{
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}, nil
}

// Sign generates a simple DID-like signature for a message
func (i *Identity) Sign(message string) string {
	privBytes, _ := hex.DecodeString(i.PrivateKey)
	sig := ed25519.Sign(privBytes, []byte(message))
	return hex.EncodeToString(sig)
}

// Verify checks a node signature
func Verify(pubKeyHex, message, sigHex string) bool {
	pubBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil { return false }
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil { return false }

	return ed25519.Verify(pubBytes, []byte(message), sigBytes)
}

// GetDID returns a simple DID string for this node
func (i *Identity) GetDID() string {
	return fmt.Sprintf("did:hustle:%s", i.PublicKey)
}
