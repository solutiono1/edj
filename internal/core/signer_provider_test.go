package core

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func kmsTestServer(t *testing.T) (*httptest.Server, *Signer) {
	t.Helper()
	signer := NewSigner()
	mux := http.NewServeMux()
	mux.HandleFunc("/sign", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Claim string `json:"claim"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.Claim == "__public_key__" {
			_ = json.NewEncoder(w).Encode(map[string]string{"publicKey": signer.PublicKeyBase64()})
			return
		}
		sig := ed25519.Sign(signer.priv, []byte(req.Claim))
		_ = json.NewEncoder(w).Encode(map[string]string{
			"signature": base64.RawURLEncoding.EncodeToString(sig),
			"publicKey": signer.PublicKeyBase64(),
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, signer
}

func TestKMSClientSignsVerifiableProof(t *testing.T) {
	server, signer := kmsTestServer(t)
	client := NewKMSClient(server.URL+"/sign", "")
	proof, err := NewProof(client, "Lender X", 5000, 30, &ProofAnchor{Txid: "abc", Vout: 0, Amount: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyProof(proof) {
		t.Fatal("KMS-signed proof did not verify")
	}
	if proof.SignerFingerprint != signer.Fingerprint() {
		t.Fatal("fingerprint does not match KMS key")
	}
}

func TestKMSClientRejectsBadResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"signature":""}`))
	}))
	defer server.Close()
	client := NewKMSClient(server.URL, "")
	if _, err := NewProof(client, "Lender X", 1000, 30, nil); err == nil {
		t.Fatal("empty KMS signature was accepted")
	}
}
