# IDEA BANK

Captured ideas, integrations, routes and tools that are planned or parked. This is the shared memory for future work and future hackathons.

## V2 product

- Real BCH-native private CashTokens. BCH Script covenants already compile. Next step is chipnet deployment, then a privacy pool with real proofs.
- Selective disclosure: exact balance, at least a threshold, a payment was made, and payroll total. Each with a named reader and expiry. Currently only signed statements exist.
- Private payroll engine: multi-output, per-recipient one-time addresses, crash-safe resume, results export.
- Scheduled and recurring payroll with caps and end dates.
- Payment requests with the details in the fragment so nothing reaches the server.
- Claim links with a fresh note so the sender cannot spend it back.

## Integrations

- Wallet connection: WizardConnect or XO for BCH.
- CashFusion for the BCH privacy leg.
- Google Identity Services for sign-in.
- Supabase for persistence.
- Railway for the API.
- Cloudflare Pages for the frontend.
- MCP server for agent payments, read-only first, then signed with spending limits.
- x402 style agent payments over BCH.
- PUSD and other BCH CashToken stablecoins.

## Routes

App shell: portfolio, links, payroll, proofs, transactions, wallet, notifications, profile, settings.

Public pages: claim link, proof verification, docs, legal, support, faq.

## Tools

- SDK source client at `sdk/js`, served at `/static/sdk/pagor.js`.
- CLI at `tools/cli/pagor.mjs`.
- Contract compile and test at `tools/contract`.
- Chipnet deploy script at `tools/contract/deploy-chipnet.mjs`.
- Deployment helpers: `.railway/railway.ts`, `wrangler.toml`, `web/_redirects`, `scripts/deploy-cloudflare.sh`.

## API surface

Links, payroll, disclosure, portfolio, notifications, account, session, assistant. Full list in `API.md`.

## Research and open questions

- Whether BCH can verify ZK proofs inside CashVM covenants. This decides if pagor can reach real zero-knowledge privacy or must stay with signed statements and covenants.
- CashToken liquidity and depth for PUSD.
- BCH wallet support for CashTokens and payment URIs.
- Jurisdiction and compliance for private business payments.

## Next hackathons

- Shortlist events that fit: BCH grants, stablecoin B2B payments, privacy, and fintech infrastructure.
- Keep the demo, deck, and video reusable. Update the public link and txid before each submission.
