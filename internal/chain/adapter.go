package chain

import "context"

// Chain is the chain-agnostic boundary. v1 implements only BCH.
type Chain interface {
	AddressFor(ctx context.Context, token *string) (string, error)
	Send(ctx context.Context, from, to string, amount int64, token *string) (string, error)
	RoutePrivacy(ctx context.Context, amount int64, token *string) ([]string, error)
}
