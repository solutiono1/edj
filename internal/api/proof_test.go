package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"edj.cash/edj/internal/core"
	"edj.cash/edj/internal/store"
)

func TestSigningKeyEndpoint(t *testing.T) {
	s := &Server{Store: store.New(), Signer: core.NewSigner()}
	rec := httptest.NewRecorder()
	s.SigningKey(rec, httptest.NewRequest(http.MethodGet, "/api/keys", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	pub, _ := body["publicKey"].(string)
	fp, _ := body["fingerprint"].(string)
	if pub == "" || len(fp) != 64 {
		t.Fatalf("bad key response: %v", body)
	}
}

func TestAnchoredProofIsSignedAndVerifiable(t *testing.T) {
	s := &Server{Store: store.New(), Signer: core.NewSigner()}
	body := `{"subject":"Lender X","threshold":1000,"expiresInDays":30,"anchor":{"txid":"abc123","vout":1,"amount":25000}}`
	req := httptest.NewRequest(http.MethodPost, "/api/disclose", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.CreateProof(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id, _ := created["id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/api/proofs/"+id, nil)
	getReq.SetPathValue("id", id)
	getRec := httptest.NewRecorder()
	s.GetProof(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d", getRec.Code)
	}
	var proof core.DisclosureProof
	if err := json.Unmarshal(getRec.Body.Bytes(), &proof); err != nil {
		t.Fatal(err)
	}
	if !core.VerifyProof(&proof) {
		t.Fatal("public verification failed")
	}
	if proof.Anchor == nil || proof.Anchor.Txid != "abc123" || proof.Anchor.Amount != 25000 {
		t.Fatalf("anchor not stored: %+v", proof.Anchor)
	}
	if proof.SignerFingerprint != s.Signer.Fingerprint() {
		t.Fatal("fingerprint mismatch")
	}
}
