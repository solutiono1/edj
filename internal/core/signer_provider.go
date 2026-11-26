package core

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ProofSigner is the boundary for disclosure signing. A local Ed25519 key or
// a remote KMS/HSM can implement it. The private key never has to live in the
// application process when a remote signer is configured.
type ProofSigner interface {
	PublicKeyBase64() string
	Fingerprint() string
	Sign(claim string) (string, error)
}

func (s *Signer) Sign(claim string) (string, error) {
	return base64.RawURLEncoding.EncodeToString(ed25519.Sign(s.priv, []byte(claim))), nil
}

// NewProof builds a proof from any signer provider.
func NewProof(signer ProofSigner, subject string, threshold int64, days int, anchor *ProofAnchor) (*DisclosureProof, error) {
	if signer == nil {
		return nil, fmt.Errorf("proof signer not configured")
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	exp := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	claim := "holds at least " + strconv.FormatInt(threshold, 10) + " for " + subject + " until " + exp.Format(time.RFC3339) + " " + anchorText(anchor)
	signature, err := signer.Sign(claim)
	if err != nil {
		return nil, err
	}
	return &DisclosureProof{
		ID:                "DP_" + base64.RawURLEncoding.EncodeToString(b),
		Subject:           subject,
		Threshold:         threshold,
		ExpiresAt:         exp,
		Signature:         signature,
		PublicKey:         signer.PublicKeyBase64(),
		SignerFingerprint: signer.Fingerprint(),
		AnchorCommitment:  AnchorCommitment(anchor),
		Anchor:            anchor,
		Claim:             claim,
	}, nil
}

// KMSClient signs claims through an external HTTP signer. Configure it with
// EDJ_KMS_SIGN_URL and optionally EDJ_KMS_TOKEN. The endpoint must accept
// {"claim": "..."} and return {"signature": "<base64url>", "publicKey": "<base64url>"}.
type KMSClient struct {
	URL       string
	Token     string
	client    *http.Client
	publicKey string
}

func NewKMSClient(url, token string) *KMSClient {
	return &KMSClient{URL: url, Token: token, client: &http.Client{Timeout: 10 * time.Second}}
}

func (k *KMSClient) getPublicKey() (string, error) {
	if k.publicKey != "" {
		return k.publicKey, nil
	}
	body, err := k.call(map[string]string{"claim": "__public_key__"})
	if err != nil {
		return "", err
	}
	if body.PublicKey == "" {
		return "", fmt.Errorf("KMS did not return a public key")
	}
	k.publicKey = body.PublicKey
	return k.publicKey, nil
}

func (k *KMSClient) PublicKeyBase64() string {
	key, err := k.getPublicKey()
	if err != nil {
		return ""
	}
	return key
}

func (k *KMSClient) Fingerprint() string {
	key, err := k.getPublicKey()
	if err != nil {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return ""
	}
	return Fingerprint(raw)
}

type kmsResponse struct {
	Signature string `json:"signature"`
	PublicKey string `json:"publicKey"`
}

func (k *KMSClient) call(payload any) (*kmsResponse, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, k.URL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if k.Token != "" {
		req.Header.Set("Authorization", "Bearer "+k.Token)
	}
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("KMS request: %w", err)
	}
	defer resp.Body.Close()
	var out kmsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("KMS response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("KMS status %d", resp.StatusCode)
	}
	return &out, nil
}

func (k *KMSClient) Sign(claim string) (string, error) {
	out, err := k.call(map[string]string{"claim": claim})
	if err != nil {
		return "", err
	}
	if out.Signature == "" {
		return "", fmt.Errorf("KMS did not return a signature")
	}
	return out.Signature, nil
}
