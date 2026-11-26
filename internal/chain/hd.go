package chain

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

const maxDerivationIndex = uint32(0x7fffffff)

// DeriveAddress derives a P2PKH cashaddr from an xpub at m/0/index.
// The xpub must be a mainnet extended public key. No private key is ever
// handled by the server, and no address is invented at random.
func DeriveAddress(xpub string, index uint32, tokenAware bool) (string, error) {
	xpub = strings.TrimSpace(xpub)
	if xpub == "" {
		return "", fmt.Errorf("xpub not configured")
	}
	if index > maxDerivationIndex {
		return "", fmt.Errorf("derivation index out of range")
	}
	master, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return "", fmt.Errorf("invalid xpub: %w", err)
	}
	if master.IsPrivate() {
		return "", fmt.Errorf("expected a public xpub, got a private key")
	}
	if !master.IsForNet(&chaincfg.MainNetParams) {
		return "", fmt.Errorf("xpub is not for mainnet")
	}
	external, err := master.Derive(0)
	if err != nil {
		return "", fmt.Errorf("derive external chain: %w", err)
	}
	child, err := external.Derive(index)
	if err != nil {
		return "", fmt.Errorf("derive child %d: %w", index, err)
	}
	pub, err := child.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("child public key: %w", err)
	}
	var hash [20]byte
	copy(hash[:], btcutil.Hash160(pub.SerializeCompressed()))
	return EncodeCashaddrForNetwork(hash, "bitcoincash", tokenAware)
}

// RandomDerivationIndex returns a non-hardened BIP32 child index.
func RandomDerivationIndex() (uint32, error) {
	var raw [4]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(raw[:]) & maxDerivationIndex, nil
}
