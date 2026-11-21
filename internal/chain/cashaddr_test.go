package chain

import (
	"encoding/hex"
	"testing"
)

func TestSpecVector(t *testing.T) {
	b, err := hex.DecodeString("F5BF48B397DAE70BE82B3CCA4793F8EB2B6CDAC9")
	if err != nil {
		t.Fatal(err)
	}
	var h [20]byte
	copy(h[:], b)
	got, err := EncodeCashaddr(h, false)
	if err != nil {
		t.Fatal(err)
	}
	want := "bitcoincash:qr6m7j9njldwwzlg9v7v53unlr4jkmx6eylep8ekg2"
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestEncodeCashaddrChecksum(t *testing.T) {
	var h [20]byte
	for i := range h {
		h[i] = byte(i)
	}
	addr, err := EncodeCashaddr(h, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(addr) < 50 || addr[:11] != "bitcoincash" {
		t.Fatalf("bad address: %s", addr)
	}
	body := addr[12:]
	payload := make([]byte, 0, len(body))
	for _, c := range []byte(body) {
		idx := -1
		for j := 0; j < len(cashCharset); j++ {
			if cashCharset[j] == c {
				idx = j
				break
			}
		}
		if idx < 0 {
			t.Fatalf("bad char %q", c)
		}
		payload = append(payload, byte(idx))
	}
	check := append(prefixExpand("bitcoincash"), payload...)
	if polymod(check) != 1 {
		t.Fatalf("checksum invalid for %s", addr)
	}
}

func TestTokenAwareAddress(t *testing.T) {
	var h [20]byte
	h[0] = 1
	addr, err := EncodeCashaddr(h, true)
	if err != nil {
		t.Fatal(err)
	}
	if addr[12] != 'z' {
		t.Fatalf("token-aware address should start with z, got %s", addr)
	}
}
