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
	p := s.NewProof("Lender X", 1000, 30)
	if !s.Verify(p) {
		t.Fatal("signature failed")
	}
	p.Claim += " tampered"
	if s.Verify(p) {
		t.Fatal("tampered claim verified")
	}
}
