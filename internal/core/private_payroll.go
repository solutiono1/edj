package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// PrivatePayroll holds an opaque blob encrypted on the user's device. The
// server never receives names, addresses or amounts for this flow.
type PrivatePayroll struct {
	ID         string    `json:"id"`
	OwnerSub   string    `json:"-"`
	Salt       string    `json:"salt"`
	IV         string    `json:"iv"`
	Ciphertext string    `json:"ciphertext"`
	RowCount   int       `json:"rowCount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

func NewPrivatePayroll(salt, iv, ciphertext string, rowCount int) *PrivatePayroll {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return &PrivatePayroll{
		ID:         "PP_" + hex.EncodeToString(b),
		Salt:       salt,
		IV:         iv,
		Ciphertext: ciphertext,
		RowCount:   rowCount,
		Status:     "encrypted",
		CreatedAt:  time.Now().UTC(),
	}
}
