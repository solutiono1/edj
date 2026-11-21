package chain

import (
	"context"
	"fmt"
)

// BCH implements Chain for Bitcoin Cash mainnet.
// Addresses are supplied by the operator (merchant-controlled). v1 does not
// derive fresh HD addresses; that requires an xpub integration. We never
// generate random addresses, because paying one would lose funds.
type BCH struct {
	ReceiveAddress string // BCH cashaddr controlled by the operator
	TokenAddress   string // CashTokens-enabled cashaddr controlled by the operator
}

func (b BCH) AddressFor(ctx context.Context, token *string) (string, error) {
	if token != nil {
		if b.TokenAddress == "" {
			return "", fmt.Errorf("EDJ_TOKEN_ADDRESS not configured")
		}
		return b.TokenAddress, nil
	}
	if b.ReceiveAddress == "" {
		return "", fmt.Errorf("EDJ_RECEIVE_ADDRESS not configured")
	}
	return b.ReceiveAddress, nil
}

// Send is delegated to the connected wallet (payment URI) in the browser.
func (BCH) Send(ctx context.Context, from, to string, amount int64, token *string) (string, error) {
	return "", nil
}

func (BCH) RoutePrivacy(ctx context.Context, amount int64, token *string) ([]string, error) {
	return []string{"dedicated-address", "cashfusion", "one-time-output-plan"}, nil
}
