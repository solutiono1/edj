# V2 PLAN

Goal: turn the pagor demo into a defensible BCH-native private payments product, ready for a future hackathon or grant.

## Build status - 26 Nov 2026

Phases 1 and 2 are in progress.

- Phase 1: xpub derivation for fresh link addresses, tested coin selection and fee planning, review-first signer with an explicit separate broadcast step, URI fallback retained. In-app signing is still gated by `EDJ_SIGNER_ENABLED = false` until chipnet verification.
- Phase 2: four-recipient covenant compiled and tested on mock chipnet, with rejection cases recorded. Payroll runs now carry per-row results with idempotent updates and CSV export.

Not started: selective disclosure, SDK/CLI/MCP parity, transparency page, independent audit, growth assets.

## The one thing that wins

BCH-native private CashTokens. The differentiation is not copying EVM privacy pools. It is making BCH stablecoins private using covenants, CashFusion for the BCH leg, and eventually zero-knowledge proofs. Everything else supports that.

## M0 - Close v1

- One real BCH mainnet transaction through the app.
- Chipnet covenant transaction and txid.
- Wallet connection.
- EMBER banner.
- Demo video and deck.
- Publish build-in-public posts and Telegram note.

## M1 - Privacy core

- Covenant pool for CashTokens and BCH on chipnet, then mainnet.
- Selective disclosure proofs: exact balance, at least a threshold, a payment was made, payroll total. Each with a named reader and expiry.
- Keep signed statements as the fallback while ZK is researched.

## M2 - Privacy engine features

- Private send inside the pool.
- Claim links with a fresh note so the sender cannot spend it back.
- Payroll engine: per-recipient one-time outputs, crash-safe resume, results export.
- Scheduled and recurring payroll with caps.
- Payment requests with details in the fragment.

## M3 - Integrations and tools

- Wallet connection.
- CashFusion for the BCH privacy leg.
- MCP server: read-only first, then signed agent payments with spending limits.
- SDK and CLI parity with the API.
- Sanctions screening and issuer policy checks where required.

## M4 - Compliance, honesty, audit

- Transparency page showing what is public.
- Independent audit of the covenant and privacy claims.
- Plain-language limits on every feature. No anonymity claim and no false ZK claim.

## M5 - Growth

- Reusable demo, deck and video for multiple hackathons and grants.
- Record the target event and deadline in `DECISIONS.md` before each build.

## Next hackathon targets

Shortlist events that fit: BCH grants, stablecoin B2B payments, privacy, and fintech infrastructure. Prefer events where a BCH-native privacy demo is a differentiator, not events dominated by EVM projects.
