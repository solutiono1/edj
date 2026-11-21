package chain

import "fmt"

const cashCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var cashGen = [5]uint64{0x98f2bc8e61, 0x79b76d99e2, 0xf33e5fb3c4, 0xae2eabe2a8, 0x1e4f43e470}

func polymod(values []byte) uint64 {
	c := uint64(1)
	for _, v := range values {
		c0 := c >> 35
		c = ((c & 0x07ffffffff) << 5) ^ uint64(v)
		for i, g := range cashGen {
			if (c0>>uint(i))&1 == 1 {
				c ^= g
			}
		}
	}
	return c
}

func prefixExpand(prefix string) []byte {
	out := make([]byte, 0, len(prefix)+1)
	for i := 0; i < len(prefix); i++ {
		out = append(out, prefix[i]&31)
	}
	return append(out, 0)
}

func convertBits(data []byte, from, to uint, pad bool) ([]byte, error) {
	var acc uint32
	var bits uint
	out := []byte{}
	maxv := uint32(1<<to) - 1
	for _, v := range data {
		if uint32(v)>>from != 0 {
			return nil, fmt.Errorf("invalid data for %d-bit conversion", from)
		}
		acc = (acc << from) | uint32(v)
		bits += from
		for bits >= to {
			bits -= to
			out = append(out, byte((acc>>bits)&maxv))
		}
	}
	if pad {
		if bits > 0 {
			out = append(out, byte((acc<<(to-bits))&maxv))
		}
	} else if bits >= from || ((acc<<(to-bits))&maxv) != 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	return out, nil
}

// EncodeCashaddr builds a mainnet P2PKH cashaddr from a 20-byte pubkey hash.
// version 0 = P2PKH; token-aware addresses use version 8 prefix bits variant.
func EncodeCashaddr(hash20 [20]byte, tokenAware bool) (string, error) {
	prefix := "bitcoincash"
	version := byte(0)
	if tokenAware {
		version = 16 // 0x10 -> first 5-bit group = 2 ('z'), token-aware P2PKH
	}
	withVersion := append([]byte{version}, hash20[:]...)
	data5, err := convertBits(withVersion, 8, 5, true)
	if err != nil {
		return "", err
	}
	payload := append([]byte{}, data5...)
	expanded := append(prefixExpand(prefix), payload...)
	expanded = append(expanded, make([]byte, 8)...)
	mod := polymod(expanded) ^ 1
	for i := 0; i < 8; i++ {
		payload = append(payload, byte((mod>>uint(5*(7-i)))&31))
	}
	addr := prefix + ":"
	for _, v := range payload {
		if int(v) >= len(cashCharset) {
			return "", fmt.Errorf("bad charset index")
		}
		addr += string(cashCharset[v])
	}
	return addr, nil
}
