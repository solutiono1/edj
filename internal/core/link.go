package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Link is a one-time private payment request.
type Link struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Amount    int64     `json:"amount"`
	Token     *string   `json:"token,omitempty"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	Status    string    `json:"status"`
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// NewLink creates a link with an unguessable ID and a fresh one-time address.
// The address comes from the chain adapter; callers supply it.
func NewLink(address string, amount int64, token *string, note string) *Link {
	return &Link{
		ID:        "L_" + randomHex(16),
		Address:   address,
		Amount:    amount,
		Token:     token,
		Note:      note,
		CreatedAt: time.Now().UTC(),
		Status:    "open",
	}
}
