# ARCHITECTURE

## Current shape

pagor is one Go binary. It serves static HTML pages, hand-written CSS, a small amount of vanilla JavaScript, and a JSON API. The frontend has no build step. htmx is vendored locally but the pages currently use ordinary fetch calls.

```
cmd/pagor/                  server entrypoint, routes
internal/
  api/                    HTTP handlers, session checks
  core/                   links, payroll parsing, signed statements
  chain/                  Bitcoin Cash address adapter
  store/                  persistent JSON record store
web/                      pages, CSS, scripts, icons, SDK
data/pagor.json             local records and signing key, ignored by git
```

## Request flow

1. The browser loads a static page from `web/`.
2. Sign-in uses email and password, or optionally Google Identity Services.
3. Email credentials are verified on the server with bcrypt. Google ID tokens are verified with Google. Both paths return a session token that the browser stores in localStorage.
4. Authenticated requests send `Authorization: Bearer <token>`.
5. Write endpoints and account reads require that header. Public reads are limited to a link id or a proof id.
6. Chain addresses come from an operator xpub when configured. Each link derives a fresh `m/0/index` P2PKH address and records the index. Without an xpub the app uses the configured fallback address. It never invents a random address.

## Chain model

Bitcoin Cash settlement is public. pagor does not hold funds and does not broadcast transactions. The claim page opens a BCH payment URI in the user's wallet. Payout execution is handled by the wallet and an external privacy step.

The chain adapter in `internal/chain` keeps chain details out of core logic. BCH is the only adapter. It can derive addresses from a mainnet xpub or use operator-configured fallback addresses. There is no multi-chain support.

Payments are planned in the browser by `web/static/payment-math.js`: coin selection, fee estimation and dust handling are pure functions with Node tests. `web/static/signer.js` builds a transaction, asks the connected WizardConnect wallet to sign it with `broadcast: false`, and only broadcasts when the user presses a separate button. In-app signing stays disabled until chipnet verification.

Privacy claims are limited:

- BCH payments use a configured receiving address and an unguessable link id.
- CashFusion can be used by the recipient on the BCH leg.
- PUSD flows are described as a planned route. They are not presented as an implemented on-chain contract.
- Disclosure proofs are Ed25519 signed statements with a subject, threshold, and expiry. They are not zero-knowledge proofs.

## Store

Records and the signing key are written to `data/pagor.json` using an atomic rename. The default path can be changed with `EDJ_DATA_FILE`. The file is ignored by git.

This is a prototype store, not an audited database. It gives restart survival without adding a database dependency.

## Trust boundaries

- No custodian, no admin withdrawal, no committee key. Pagor never holds funds.
- The browser holds the session token and the derived key fingerprint.
- The server holds links, proofs, users, sessions, opaque encrypted payroll blobs, and the disclosure signing key. It does not hold payroll amounts in the private flow and never holds a user's wallet balance.
- The wallet holds funds and signs payments.
- External privacy tooling is reported for what it actually does.

## Security rules

- Never log session tokens, signing keys, or link secrets.
- Link and proof ids are cryptographically random.
- CSV uploads are parsed strictly with a fixed header.
- No shell interpolation for external commands.
- Write endpoints must keep server-side session checks. The UI is not a security boundary.

## Contract scope

BCH does not use EVM-style smart contracts. v1 keeps the payment route in the chain adapter and the wallet, not in a fabricated contract. Any future covenant or CashTokens contract must be written against BCH Script, tested on chipnet, and audited before it is described in user-facing copy.
