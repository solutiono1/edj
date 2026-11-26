package chain

import (
	"strings"
	"testing"
)

// BIP32 test vector 1 master xpub, seed 000102030405060708090a0b0c0d0e0f.
const vectorXpub = "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"

func TestDeriveAddressVector(t *testing.T) {
	cases := []struct {
		index uint32
		want  string
	}{
		{0, "bitcoincash:qqx3e8qz57lfh29css5qfl4ev9ypeejkrvlz5vxrjz"},
		{1, "bitcoincash:qqdyc0gkgzwam3yezcprphyy5yvzk24n3cm835tnv6"},
		{2, "bitcoincash:qza3h2r3sq2trs5y62ud6av07g7jvjxdtv03klzcnv"},
	}
	for _, tc := range cases {
		got, err := DeriveAddress(vectorXpub, tc.index, false)
		if err != nil {
			t.Fatalf("index %d: %v", tc.index, err)
		}
		if got != tc.want {
			t.Fatalf("index %d: got %s, want %s", tc.index, got, tc.want)
		}
	}
}

func TestDeriveAddressIsStableAndDistinct(t *testing.T) {
	a, _ := DeriveAddress(vectorXpub, 7, false)
	b, _ := DeriveAddress(vectorXpub, 7, false)
	c, _ := DeriveAddress(vectorXpub, 8, false)
	if a != b {
		t.Fatal("same index produced different addresses")
	}
	if a == c {
		t.Fatal("different indexes produced the same address")
	}
}

func TestDeriveAddressRejectsBadInput(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"garbage":     "not-an-xpub",
		"private key": "xprv9s21ZrQH143K3QTDL4LXw2F7HEK3wJUD2nW2nRk4stbPy6cq3jPPqjiChkVvvNKmPGJxWUtg6LnF5kejMRNNU3TGtRBeJgk33yuGBxrMPHi",
		"testnet":     "tpubD6NzVbkrYhZ4WaWSyoBvQw1gpLqS9kugum5MLkTRVDoKUM2mKgqKzC4KYcV4TFA6ZmYcZqjBz6hDgZgqQ5A6pKtAFz5LzVhK9T5z6mYcYVp",
	}
	for name, key := range cases {
		if _, err := DeriveAddress(key, 0, false); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	if _, err := DeriveAddress(vectorXpub, 0x80000000, false); err == nil {
		t.Fatal("hardened index was accepted")
	}
}

func TestRandomDerivationIndex(t *testing.T) {
	for i := 0; i < 100; i++ {
		idx, err := RandomDerivationIndex()
		if err != nil {
			t.Fatal(err)
		}
		if idx > maxDerivationIndex {
			t.Fatalf("index %d outside non-hardened range", idx)
		}
	}
}

func TestEncodeCashaddrForNetwork(t *testing.T) {
	var hash [20]byte
	for i := range hash {
		hash[i] = byte(i)
	}
	main, err := EncodeCashaddrForNetwork(hash, "bitcoincash", false)
	if err != nil || !strings.HasPrefix(main, "bitcoincash:") {
		t.Fatalf("mainnet address: %s %v", main, err)
	}
	test, err := EncodeCashaddrForNetwork(hash, "bchtest", false)
	if err != nil || !strings.HasPrefix(test, "bchtest:") {
		t.Fatalf("chipnet address: %s %v", test, err)
	}
}
