package core

import (
	"crypto/ed25519"
	"fmt"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"
)

// DisclosureProof is a signed statement, NOT a zero-knowledge proof.
type DisclosureProof struct {
	ID        string    `json:"id"`
	Subject   string    `json:"subject"`
	Threshold int64     `json:"threshold"`
	ExpiresAt time.Time `json:"expiresAt"`
	Signature string    `json:"signature"`
	PublicKey string    `json:"publicKey"`
	Claim     string    `json:"claim"`
}

type Signer struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func NewSigner() *Signer {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return &Signer{pub: pub, priv: priv}
}

// PrivateKey returns a copy of the signing key so it can be persisted across restarts.
func (s *Signer) PrivateKey() ed25519.PrivateKey {
	return append(ed25519.PrivateKey(nil), s.priv...)
}

// NewSignerFromKey rebuilds the signer from a persisted Ed25519 private key.
func NewSignerFromKey(priv ed25519.PrivateKey) (*Signer, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid signing key length")
	}
	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("invalid signing key")
	}
	return &Signer{pub: pub, priv: append(ed25519.PrivateKey(nil), priv...)}, nil
}

// NewProof signs "holds at least X for subject until expiry".
func (s *Signer) NewProof(subject string, threshold int64, days int) *DisclosureProof {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	exp := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	claim := "holds at least " + strconvI64(threshold) + " for " + subject + " until " + exp.Format(time.RFC3339)
	sig := ed25519.Sign(s.priv, []byte(claim))
	return &DisclosureProof{
		ID:        "DP_" + base64.RawURLEncoding.EncodeToString(b),
		Subject:   subject,
		Threshold: threshold,
		ExpiresAt: exp,
		Signature: base64.RawURLEncoding.EncodeToString(sig),
		PublicKey: base64.RawURLEncoding.EncodeToString(s.pub),
		Claim:     claim,
	}
}

func (s *Signer) Verify(p *DisclosureProof) bool {
	sig, err := base64.RawURLEncoding.DecodeString(p.Signature)
	if err != nil {
		return false
	}
	return ed25519.Verify(s.pub, []byte(p.Claim), sig)
}

func strconvI64(v int64) string {
	return strconv.FormatInt(v, 10)
}
