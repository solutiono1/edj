# V3 PLAN

Build decisions recorded from the product owner on 26 Nov 2026. These are user decisions, not inferred features.

## Production storage answer

Supabase **is** PostgreSQL. Production already uses it through the service-role key, so `data/pagor.json` is local development only. The remaining part of the 90-day upgrade was key custody: `ProofSigner` now supports an HTTP KMS/HSM endpoint via `EDJ_KMS_SIGN_URL`. When set, the app does not generate or store a local signing key. The actual KMS endpoint and credentials must be supplied by the operator.

## Custody decision

**LOCKED: Path A.** "Balance" means the user's own addresses aggregated into a displayed balance. Pagor never holds funds. Any off-ramp is a handoff to a regulated partner, not an in-app conversion run by Pagor. We do not become a custodian, do not hold keys, and do not claim to.

## Onboarding decision

**LOCKED: one path.** The user creates a Pagor account and gets a Pagor-managed non-custodial wallet through the Pagor adapter. The earlier "bring your own WizardConnect account" option is removed. Importing an existing key may be added later as an explicit user choice, but it is not the default path. WizardConnect remains the BCH transport *inside* the Pagor adapter.

## Privacy decision

**LOCKED.** Pagor must not be able to read a user's balance or what they were paid in payroll. Balances are computed in the browser. Payroll amounts are encrypted on the device before upload. The server stores opaque records. This strengthens the existing non-custodial claim; it does not add an anonymity claim.

## Account lifecycle decision

**LOCKED.** Users can disable or delete their account. Deletion has a 30-day recovery window before records are purged. Sessions are revoked immediately on disable or delete.

## Chain order

Decision: BCH first, then Arc, then Base, then Solana, then Stellar and XRPL. Robinhood Chain stays on the watchlist because it is newer and its activity is concentrated. Tempo is a watchlist item from current research.

Selection rule: only chains that are real payment rails with a stablecoin meeting all five tests: regulated issuer, 1:1 redemption, deep liquidity, transparent reserves, no sustained depeg.

| Chain | Status | Stablecoins |
|---|---|---|
| Bitcoin Cash | Live. First adapter. | PUSD |
| Arc | Next. Circle L1, USDC gas, sub-second finality. Opt-in privacy still in development. | USDC |
| Base | Planned. EVM L2 with deep USDC liquidity. Public ledger. | USDC |
| Solana | Planned. Different signing model, needs a new adapter. | USDC, PYUSD |
| Stellar | Planned. Payments-first, strong remittance fit. | USDC |
| XRPL | Planned. Payments-first, fast finality. | RLUSD |
| Robinhood Chain | Watchlist. Arbitrum Orbit L2, USDG canonical, RLUSD added. | USDG, RLUSD |
| Tempo | Watchlist. Stripe and Paradigm payments L1. | USD-backed native |

Every public-ledger chain must be described honestly. We do not claim private payments where the chain itself is fully public.

## Build order

1. Pagor wallet adapter interface with a BCH implementation and a chain registry.
2. Client-side balance aggregation and a portfolio that does not ask the server for amounts.
3. Account disable and delete with 30-day recovery.
4. Client-side encrypted payroll amounts.
5. Chain adapters in the order above, each with chipnet or testnet verification before any mainnet claim.

## Build status - 26 Nov 2026

- Chain registry and stablecoin tests: done. BCH is the only live chain.
- Pagor wallet adapter: done. One interface, BCH implementation over WizardConnect, other chains fail loudly.
- Client-side balance: done. Portfolio computes from the user's own UTXOs and the server returns counts only.
- Account lifecycle: done. Disable, delete with 30-day recovery, restore, session revocation, and background purge of user records.
- Payroll privacy: done. CSV is parsed and encrypted in the browser; the server stores only an opaque blob it cannot read.
- Chain adapters for Arc, Base, Solana, Stellar and XRPL: not started.
- Remote signing: done. `ProofSigner` interface with local Ed25519 and HTTP KMS implementations. KMS mode avoids generating or storing a local signing key.
- Selective disclosure: done. Proofs carry an anchor commitment inside the signed claim; raw txid is visible only after the owner calls the reveal endpoint.
- MCP server: done in `mcp/`, with a hard per-agent spending ceiling.
- Batch covenant tooling: done. `tools/contract/payroll-covenant.mjs` builds up to four recipients from CSV.
- SDK encryption: done. `createPrivatePayroll` and `decryptPayroll` added to `@pagor/sdk` and the served source client.

- Wallet UX: WizardConnect sessions are restored from local storage on page load, so an already linked wallet is reflected immediately. Fresh links still need the wallet relay handshake.
- Landing preview: README now uses a current screenshot from `usepagor.xyz`.

Still open: key import, chipnet signature verification with a real wallet, mainnet transaction, independent audit, and chain adapters after BCH.

## Explicitly out of scope

- Custody, hosted balances, in-app fiat conversion, money transmission.
- Anonymity or zero-knowledge claims.
- Key import in the first release.
