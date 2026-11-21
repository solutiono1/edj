# ARCHITECTURE - living

## Shape

A small Go HTTP server serves server-rendered HTMX pages (templ templates + Tailwind CSS). The same server exposes a JSON API used by the SDK and by the HTMX fragments. All money movement happens through a chain adapter; core logic never speaks BCH RPC directly.

```
edj/
  cmd/edj/                  # server entrypoint
  internal/
    core/                   # chain-agnostic logic (links, payroll, proofs)
      link.go               # one-time address + claim link generation
      payroll.go            # CSV parse, batch plan, unlinkable output plan
      disclosure.go         # signed threshold proof (non-ZK)
    chain/
      adapter.go            # interface: Balance, Send, Route, Confirm
      bch.go                # BCH mainnet implementation (v1 only adapter)
    privacy/
      router.go             # token -> BCH -> CashFusion -> BCH -> token plan
      cashfusion.go         # CashFusion adapter (external tool invocation)
    api/
      http.go               # JSON API + HTMX partial handlers
    store/
      memory.go             # v1: in-memory/ephemeral job state, no database
  web/                      # FRONTEND BUILDER OWNS THIS DIRECTORY ONLY
    templates/              # .templ files
    static/                 # compiled Tailwind CSS, small assets
  handoffs/                 # incoming patches from frontend builder
  REVIEWS/                  # planner's reviews of each TASK
  TASKS/                    # copy-paste task envelopes
  *.md                      # living docs
```

## Chain adapter (interface)

```go
type Chain interface {
    NewOneTimeAddress(ctx context.Context) (Address, error)
    Send(ctx context.Context, from, to Address, amount int64, token *TokenID) (TxID, error)
    RoutePrivacy(ctx context.Context, amount int64, token TokenID) (TxIDs []string, err error)
    Confirm(ctx context.Context, txid TxID) (bool, error)
}
```

v1 implements only `bch.go`. Future chains are new implementations of the same interface - no core rewrite. **No multi-chain features are built in v1.**

## Core objects

```go
type Link struct {
    ID        string    // random, unguessable
    Address   string    // one-time cashaddr (token-enabled if PUSD)
    Amount    int64     // satoshis or token units
    Token     *TokenID  // nil = BCH
    CreatedAt time.Time
    Status    string    // open | claimed | expired
}

type PayrollRun struct {
    ID        string
    Rows      []PayrollRow // name/address/amount from CSV
    Total     int64
    Plan      []PayoutPlan // unlinkable outputs + privacy route steps
    Status    string       // draft | reviewed | executing | done
}

type DisclosureProof struct {
    ID        string
    Subject   string    // who it is for
    ExpiresAt time.Time
    Threshold int64
    Signature string    // signed by holder's key; verifiable offline
    Claim     string    // e.g. "holds at least X"
}
```

## Privacy flow (v1, honest guarantees)

1. BCH payment: one-time address per link; no address reuse; CashFusion can break BCH UTXO linkability when the recipient chooses to mix.
2. PUSD payment: automated route - PUSD → BCH (swap) → CashFusion-mixed BCH → PUSD (swap) → recipient's one-time token-enabled address. Public chain sees swaps and standard transfers; the amounts and the link between payer and payee are not published by edj.
3. Payroll: one batch, per-recipient one-time outputs; the payer's main wallet is not the visible source of each payout.
4. Disclosure: holder signs a threshold statement on their device; a verifier checks the signature. This is **not** a ZK proof and reveals the named verifier and the threshold.

## Trust boundaries

- No committee key, no custodian, no admin withdrawal.
- CashFusion is an external tool; we invoke it and report what it actually does.
- The chain adapter is the only component that touches the network/chain.
- Core logic must be unit-testable with a mock `Chain`.

## Security rules

- Never log seed phrases, private keys, or raw claim-link secrets.
- Link IDs are cryptographically random; no sequential IDs.
- CSV uploads are parsed strictly; no formula evaluation, no path traversal.
- All external commands are invoked with fixed argv, no shell interpolation.
