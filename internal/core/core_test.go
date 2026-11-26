package core

import (
	"strings"
	"testing"
)

func TestParsePayroll(t *testing.T) {
	in := "name,address,amount\nAda,bitcoincash:qtest,1000\nBob,bitcoincash:qtest2,250\n"
	rows, err := ParsePayroll(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Amount != 1000 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestParsePayrollRejectsBadAmount(t *testing.T) {
	in := "name,address,amount\nAda,bitcoincash:qtest,abc\n"
	if _, err := ParsePayroll(strings.NewReader(in)); err == nil {
		t.Fatal("expected error")
	}
}

func TestProofSignVerify(t *testing.T) {
	s := NewSigner()
	anchor := &ProofAnchor{Txid: "d34a9391d4ac6eb5ff695bfb9aafa8f4e0aa8aa4cac3702a9c7de957fce0d18a", Vout: 0, Amount: 25000}
	p := s.NewProof("Lender X", 1000, 30, anchor)
	if !VerifyProof(p) {
		t.Fatal("signature failed")
	}
	if p.SignerFingerprint != s.Fingerprint() || len(p.SignerFingerprint) != 64 {
		t.Fatalf("bad fingerprint: %s", p.SignerFingerprint)
	}
	if !strings.Contains(p.Claim, anchor.Txid) {
		t.Fatal("anchor is not inside the signed claim")
	}
	p.Claim += " tampered"
	if VerifyProof(p) {
		t.Fatal("tampered claim verified")
	}
}

func TestProofFingerprintMismatchIsRejected(t *testing.T) {
	s := NewSigner()
	p := s.NewProof("Lender X", 1000, 30, nil)
	p.SignerFingerprint = "0000000000000000000000000000000000000000000000000000000000000000"
	if VerifyProof(p) {
		t.Fatal("fingerprint mismatch was accepted")
	}
}

func TestVerificationUsesEmbeddedKey(t *testing.T) {
	a := NewSigner()
	b := NewSigner()
	p := a.NewProof("Lender X", 1000, 30, nil)
	if !VerifyProof(p) {
		t.Fatal("public verification should use the embedded key")
	}
	if !b.Verify(p) {
		t.Fatal("a verifier without the private key should still verify the signature")
	}
}
