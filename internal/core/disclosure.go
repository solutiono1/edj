package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// DisclosureProof is a signed statement, NOT a zero-knowledge proof.
// It carries the signer's key fingerprint and an optional anchor to a
// confirmed on-chain output, so a verifier can check both who signed and
// which UTXO the statement claims to relate to. No central registry is
// implied: the verifier decides which keys they trust.
type DisclosureProof struct {
	ID                string       `json:"id"`
	OwnerSub          string       `json:"-"`
	Subject           string       `json:"subject"`
	Threshold         int64        `json:"threshold"`
	ExpiresAt         time.Time    `json:"expiresAt"`
	Signature         string       `json:"signature"`
	PublicKey         string       `json:"publicKey"`
	SignerFingerprint string       `json:"signerFingerprint"`
	AnchorCommitment  string       `json:"anchorCommitment,omitempty"`
	Anchor            *ProofAnchor `json:"anchor,omitempty"`
	Revealed          bool         `json:"revealed,omitempty"`
	Claim             string       `json:"claim"`
}

// ProofAnchor points to an on-chain output. The amount is in satoshis.
type ProofAnchor struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Amount int64  `json:"amount"`
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

// Fingerprint identifies a signing key without exposing a central authority.
func Fingerprint(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	return hex.EncodeToString(sum[:])
}

// PublicKeyBase64 returns the raw public key for /api/keys.
func (s *Signer) PublicKeyBase64() string {
	return base64.RawURLEncoding.EncodeToString(s.pub)
}

func (s *Signer) Fingerprint() string {
	return Fingerprint(s.pub)
}

func AnchorCommitment(anchor *ProofAnchor) string {
	if anchor == nil || anchor.Txid == "" {
		return ""
	}
	raw := anchor.Txid + ":" + strconv.Itoa(anchor.Vout) + ":" + strconv.FormatInt(anchor.Amount, 10)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func anchorText(anchor *ProofAnchor) string {
	commitment := AnchorCommitment(anchor)
	if commitment == "" {
		return "anchor:none"
	}
	return "anchor:" + commitment
}

// VerifyAnchor checks a revealed anchor against the commitment inside the
// signed claim. It does not check that the output exists; the verifier does
// that on a block explorer of their choice.
func VerifyAnchor(p *DisclosureProof, anchor *ProofAnchor) bool {
	if p == nil || p.AnchorCommitment == "" || anchor == nil {
		return false
	}
	return AnchorCommitment(anchor) == p.AnchorCommitment
}

// NewProof signs "holds at least X for subject until expiry" with an optional
// on-chain anchor. The anchor is inside the signed claim, so changing it
// invalidates the signature.
func (s *Signer) NewProof(subject string, threshold int64, days int, anchor *ProofAnchor) *DisclosureProof {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	exp := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	claim := "holds at least " + strconv.FormatInt(threshold, 10) + " for " + subject + " until " + exp.Format(time.RFC3339) + " " + anchorText(anchor)
	sig := ed25519.Sign(s.priv, []byte(claim))
	return &DisclosureProof{
		ID:                "DP_" + base64.RawURLEncoding.EncodeToString(b),
		Subject:           subject,
		Threshold:         threshold,
		ExpiresAt:         exp,
		Signature:         base64.RawURLEncoding.EncodeToString(sig),
		PublicKey:         base64.RawURLEncoding.EncodeToString(s.pub),
		SignerFingerprint: s.Fingerprint(),
		AnchorCommitment:  AnchorCommitment(anchor),
		Anchor:            anchor,
		Claim:             claim,
	}
}

// VerifyProof checks the signature against the key embedded in the proof and
// confirms the fingerprint matches that key. It does not decide whether the
// verifier trusts the key or the anchor.
func VerifyProof(p *DisclosureProof) bool {
	if p == nil {
		return false
	}
	pub, err := base64.RawURLEncoding.DecodeString(p.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return false
	}
	if Fingerprint(pub) != p.SignerFingerprint {
		return false
	}
	sig, err := base64.RawURLEncoding.DecodeString(p.Signature)
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, []byte(p.Claim), sig)
}

func (s *Signer) Verify(p *DisclosureProof) bool {
	return VerifyProof(p)
}
