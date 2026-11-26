<div align="center">

# pagor

Private payments and payroll on Bitcoin Cash. Keep amounts out of the public story of a payment, and show a signed threshold when someone needs to check.

[Live app](https://usepagor.xyz) · [Developer docs](web/docs.html) · [API](API.md) · [SDK](DEV-TOOLS.md) · [Contract](SMART-CONTRACT.md) · [MCP](MCP.md)

**pagor never holds funds or keys. Settlement is public. Details are private. This is not anonymity.**

</div>

![pagor landing page](web/static/hero-preview.png)

---

## What pagor is

pagor is a non-custodial software layer for payment links, team payroll and signed disclosure statements. The wallet holds funds and signs payments. pagor holds records, not money.

It is built around three promises:

1. **No custody.** Pagor cannot hold a balance, move funds or sign a payment for the user.
2. **No amount in the public story.** A link carries the amount to the payer, but pagor does not publish it. On-chain amounts remain public, as Bitcoin Cash requires.
3. **No false claims.** Disclosure statements are Ed25519 signed statements, not zero-knowledge proofs. Public-ledger chains are described as public.

The product direction and decisions live in [V3-PLAN.md](V3-PLAN.md) and [DECISIONS.md](DECISIONS.md).

## Status

| Area | State |
|---|---|
| Domain | `usepagor.xyz` acquired. Cloudflare custom domain wiring is pending. |
| API | Go server. Railway service `edj-api-v3` is deployed but still serving an older build until a redeploy is triggered. |
| Frontend | Cloudflare Pages serves the current build. |
| Database | Supabase PostgreSQL in production. Local JSON file for development only. |
| Signing | Local Ed25519 or an external KMS/HSM endpoint. In-app wallet signing is gated by `EDJ_SIGNER_ENABLED = false` until chipnet verification. |
| Mainnet payment | Not yet performed through the app. |
| Audit | Not yet performed. |

## Features

### Accounts and onboarding

- Email sign-up and login with bcrypt hashing, 10 to 72 character passwords and a per-IP rate limit.
- Optional Google sign-in. Google accounts reuse an existing email account when the verified email matches.
- One onboarding path: create a pagor account and get a pagor-managed non-custodial wallet through the Pagor adapter.
- Account disable, delete with a 30-day recovery window, restore after signing in again, and immediate session revocation. A background job purges the user, sessions and their records after the window closes.
- Sessions are bearer tokens in localStorage. No cookies and no tracking scripts.

### Payments

- Private links with a fresh `m/0/index` P2PKH address per link when `EDJ_XPUB` is configured. A link never reuses an address. Without an xpub the operator-configured address is used.
- Review-first signer: select inputs, estimate the fee, handle change and dust, build the transaction, ask the connected WizardConnect wallet to sign with `broadcast: false`, then require a separate explicit broadcast action.
- `bitcoincash:` URI fallback when in-app signing is off or the wallet cannot connect.
- Electrum client for UTXO reads and broadcast.

### Payroll

- Strict CSV upload with header `name,address,amount`.
- Private payroll flow: the CSV is parsed and encrypted in the browser with AES-GCM and a user passphrase. The server stores only salt, IV and ciphertext. It never receives names, addresses or amounts.
- Per-row payout results with idempotent updates, so retrying after a crash cannot double-record a payout.
- CSV export of payout outcomes.
- On-chain reality is stated plainly: funding many payouts from one UTXO links them publicly. Fund each payout from a separate UTXO when recipient privacy matters.

### Disclosure statements

- Ed25519 signed statements with a subject, threshold, expiry and named reader. Not zero-knowledge proofs.
- Every proof carries a signer fingerprint and an optional on-chain anchor commitment. The raw transaction id is inside the commitment, not the public proof.
- `POST /api/proofs/{id}/reveal` is owner-only and discloses the raw anchor on dispute.
- The public verification page checks the signature in the browser, shows the fingerprint and commitment, and lets a verifier pin a key on their device. No central registry.
- `GET /api/keys` publishes the current signing key and fingerprint.

### Developer surfaces

- Go JSON API documented in [API.md](API.md).
- `@pagor/sdk` published on npm, with a served source client at `/static/sdk/pagor.js`. Includes encrypted payroll methods.
- CLI at `tools/cli/pagor.mjs`.
- MCP server at `mcp/server.mjs` exposing `createLink`, `submitPayroll`, `createProof`, `getLink` and `getProof` over JSON-RPC stdio, with a hard per-agent spending ceiling. See [MCP.md](MCP.md).

### Bitcoin Cash covenant

- `PrivatePayroll` constrains one payout to one output, a fixed amount and a fixed address, with a refund after a height.
- `PrivatePayrollBatch` supports four fixed payouts with a count guard and a refund path. `tools/contract/payroll-covenant.mjs` builds a batch from a CSV.
- Both are compiled with `cashc` and evaluated with CashScript's mock network provider. The chipnet deployment of `PrivatePayroll` is recorded in [SMART-CONTRACT.md](SMART-CONTRACT.md).
- The batch covenant is mock tested, not deployed, not audited and not yet wired into the payroll UI.

## Chain direction

BCH is the only live chain. The registry in `web/static/chains.js` records the planned order and the stablecoin rules.

| Chain | Status | Stablecoins |
|---|---|---|
| Bitcoin Cash | Live | PUSD |
| Arc | Next | USDC |
| Base | Planned | USDC |
| Solana | Planned | USDC, PYUSD |
| Stellar | Planned | USDC |
| XRP Ledger | Planned | RLUSD |
| Robinhood Chain | Watchlist | USDG, RLUSD |
| Tempo | Watchlist | USDC |

A chain is added only when its stablecoin passes five tests: regulated issuer, 1:1 redemption, deep liquidity, transparent reserves and no sustained depeg. Public-ledger chains are never described as private payment rails.

## Run locally

Requirements:

- Go 1.23 or newer
- Node.js for contract, SDK and MCP tests
- PostgreSQL through Supabase, or a local JSON file for development
- BCH addresses you control, or a mainnet xpub

```bash
cp .env.example .env
./start.sh
```

Open `http://localhost:8080`.

### Environment

| Variable | Default | Purpose |
|---|---|---|
| `ADDR` or `PORT` | `:8080` | Listen address |
| `GOOGLE_CLIENT_ID` | none | Optional Google sign-in client ID |
| `EDJ_RECEIVE_ADDRESS` | none | Fallback BCH address for links |
| `EDJ_TOKEN_ADDRESS` | none | Fallback CashTokens address |
| `EDJ_XPUB` | none | Mainnet xpub for fresh link addresses |
| `EDJ_TOKEN_XPUB` | none | Token-aware mainnet xpub |
| `EDJ_FEE_RATE` | `1` | Satoshis per virtual byte for review |
| `EDJ_ELECTRUM_HOST` | none | Electrum host and port for UTXO reads and broadcast |
| `EDJ_SIGNER_ENABLED` | `false` | In-app signing gate. Keep false until chipnet verification |
| `EDJ_KMS_SIGN_URL` | none | Remote proof signer endpoint. Disables local signing key generation |
| `EDJ_KMS_TOKEN` | none | Bearer token for the KMS endpoint |
| `EDJ_PUBLIC_ORIGIN` | `*` | Allowed CORS origin for the deployed API |
| `SUPABASE_URL` | none | Supabase project URL |
| `SUPABASE_SERVICE_ROLE_KEY` | none | Backend-only Supabase key |
| `EDJ_DATA_FILE` | `data/edj.json` | Local store path when Supabase is not configured |

For Google sign-in, add `http://localhost:8080` and `http://127.0.0.1:8080` as authorized JavaScript origins. See [GOOGLE-OAUTH.md](GOOGLE-OAUTH.md).

## Tests

```bash
export PATH=$HOME/.local/go/bin:$PATH
go build ./...
go test ./...

cd web/static
node --test payment-math.test.mjs chains.test.mjs private-payroll.test.mjs proof-verify.test.mjs sdk/pagor.test.mjs

cd ../../mcp
node --test

cd ../tools/contract
npm test
```

Coverage:

- Go: email auth, account lifecycle, links, xpub derivation, payroll resume and export, private payroll blobs, KMS signing, proof verification and selective disclosure.
- Browser: coin selection, fee and dust handling, chain registry, stablecoin tests, payroll encryption, signature verification and SDK wire privacy.
- Contract: both covenants, rejection cases and batch construction from CSV.
- MCP: tool list, spending ceiling and ciphertext-only payroll submission.

## Layout

```
cmd/edj/              server entrypoint and routes
internal/api/         HTTP handlers, session checks, auth and account lifecycle
internal/core/        links, payroll, disclosure proofs, proof signers
internal/chain/       BCH adapter, cashaddr encoding, HD derivation, Electrum
internal/store/       Supabase store and local JSON store
web/                  pages, CSS, scripts, icons, SDK
web/static/           wallet adapter, signer, payment math, balance, payroll encryption, proof verification
sdk/js/               packaged JavaScript client
tools/cli/            command line client
tools/contract/       covenant sources, artifacts, deploy script and tests
mcp/                  MCP server and tests
supabase/             schema
```

## SDK

```js
import { PagorClient } from '@pagor/sdk'

const pagor = new PagorClient({
  baseUrl: 'https://edj-api-v3-production.up.railway.app',
  token: localStorage.getItem('edj_session')
})

const link = await pagor.createLink({ amount: 25000, note: 'Invoice 42' })
const run = await pagor.createPrivatePayroll(rows, passphrase)
const plain = await pagor.decryptPayroll(await pagor.getPrivatePayroll(run.id), passphrase)
```

Methods: `createLink`, `getLink`, `submitPayroll` (deprecated, plaintext), `createPrivatePayroll`, `getPrivatePayroll`, `decryptPayroll`, `getPayroll`, `createProof`, `getProof`, `me`, `portfolio`, `logout`.

The deprecated `submitPayroll` sends a plaintext CSV and will be removed in 0.2.0. Use `createPrivatePayroll`.

## CLI

```bash
EDJ_TOKEN=<token> node tools/cli/pagor.mjs me
EDJ_TOKEN=<token> node tools/cli/pagor.mjs portfolio
EDJ_TOKEN=<token> node tools/cli/pagor.mjs links create 5000 - 'Invoice 42'
```

`EDJ_URL` defaults to `http://localhost:8080`.

## MCP

```bash
cd mcp
PAGOR_API_TOKEN=<token> PAGOR_MCP_MAX_AMOUNT=50000 npm start
```

The agent can create links up to the ceiling and submit encrypted payroll. It cannot sign, broadcast or see a balance.

## Documentation

- [API.md](API.md) is the HTTP reference.
- [SMART-CONTRACT.md](SMART-CONTRACT.md) covers the covenants and chipnet checklist.
- [MCP.md](MCP.md) covers the agent server.
- [V3-PLAN.md](V3-PLAN.md) records the product decisions and build order.
- [REVIEW-RESPONSE.md](REVIEW-RESPONSE.md) maps external review findings to what shipped.
- [DEPLOYMENT.md](DEPLOYMENT.md) covers Railway, Cloudflare Pages and Supabase.
- [ARCHITECTURE.md](ARCHITECTURE.md) covers trust boundaries.
- [ASSETS.md](ASSETS.md) records third-party media and font licenses.

## What pagor does not do

- It does not make a payment anonymous. Bitcoin Cash settlement is public.
- It does not claim zero-knowledge proofs.
- It does not hold funds, hold keys or sign payments for the user.
- It does not verify the on-chain anchor for a proof verifier. The verifier checks it themselves.
- It does not have an independent security audit yet.
- It does not support key import yet.
- It does not have chain adapters beyond BCH yet.

## License

MIT. See the license file in the repository. Third-party media and font licenses are recorded in [ASSETS.md](ASSETS.md).
