# API - living

JSON API served by the Go process. Base path: `/api`. All responses are JSON.

## Auth and config

`GET /api/config`
```json
{
  "googleClientId": "146518974987-...apps.googleusercontent.com",
  "receiveAddress": "bitcoincash:q...",
  "tokenAddress": "bitcoincash:z...",
  "derivedAddress": true,
  "tokenDerivation": false,
  "feeRate": "1",
  "network": "mainnet"
}
```

`derivedAddress` means `EDJ_XPUB` is configured and links get a fresh `m/0/index` address. `feeRate` is the satoshi-per-virtual-byte rate used by the review step.

`POST /api/links` derives a new address for every link when an xpub is configured. The response includes `derivationIndex`, which is `0` when the configured fallback address was used instead. Without an xpub the old single-address behaviour remains.

`POST /api/auth/signup`
```json
{ "name": "Ada Lovelace", "email": "ada@example.com", "password": "at least 10 characters" }
```
Creates an email account. The email is lowercased and trimmed. Passwords are 10 to 72 characters, hashed with bcrypt, and never returned by any endpoint. Duplicate emails return `409`.

`POST /api/auth/login`
```json
{ "email": "ada@example.com", "password": "at least 10 characters" }
```
Returns the same shape as signup. Wrong credentials return `401` with a generic message that does not reveal whether the account exists. Both email endpoints are rate limited per client IP.

`POST /api/auth/google`
```json
{ "idToken": "<Google credential>" }
```
The server verifies the token with Google, checks `aud`, `email_verified` and `sub`, then returns:
```json
{ "token": "<pagor session>", "user": { "sub": "...", "email": "...", "name": "..." }, "isNew": true }
```
Google sign-in reuses an existing account when the verified email already belongs to an email signup, so a person has one account across both methods. No cookie is set. The browser stores the session token in localStorage.

`GET /api/keys` is public and returns the current signing key and its SHA-256 fingerprint. A verifier must decide separately whether to trust that key.

`GET /api/me` and `POST /api/logout` require `Authorization: Bearer <token>`.

Account lifecycle:
- `GET /api/account` returns status and the scheduled deletion date.
- `POST /api/account/disable` revokes all sessions immediately.
- `POST /api/account/delete` marks the account deleted, revokes sessions, and sets a 30-day recovery date.
- `POST /api/account/restore` works after signing in again, inside the 30-day window. It returns a new session token.

A background job purges users, sessions and their records after the window closes.

## Health

`GET /api/health` returns `{"status":"ok","engine":"pagor"}`.

## Links

`POST /api/links` requires the bearer token.
```json
{ "amount": 25000, "token": null, "note": "Invoice 42" }
```
Amounts are positive integers in the smallest unit: satoshis for BCH, cents for PUSD. A truthy `token` selects the CashTokens receiving address.

`201` response:
```json
{ "id": "L_...", "claimUrl": "/l/L_...", "address": "bitcoincash:...", "status": "open" }
```

`GET /api/links/{id}` returns the link state. It is public because the claim URL is the secret the payer receives.

## Payroll

`POST /api/private-payroll` requires the bearer token and accepts an opaque blob encrypted on the device:
```json
{ "salt": "<base64>", "iv": "<base64>", "ciphertext": "<base64>", "rowCount": 5 }
```
The server never receives names, addresses or amounts in this flow. It stores only the blob. `GET /api/private-payroll/{id}` requires the bearer token and returns the blob to its owner only. The browser decrypts it with the user's passphrase.

`POST /api/payroll/upload` is the older plaintext flow and is kept for compatibility. Prefer the private endpoint.

`POST /api/payroll/upload` requires the bearer token and is multipart with field `csv`. The CSV header is exactly `name,address,amount`, and amounts are integers.

`201` response:
```json
{ "id": "PR_...", "rows": 5, "total": 21500, "reviewUrl": "/payroll/PR_..." }
```

Each row starts with a `pending` result. `GET /api/payroll/{id}` returns the run with its `results`.

`POST /api/payroll/{id}/result` requires the bearer token and records a wallet-reported outcome:
```json
{ "index": 0, "status": "broadcast", "txid": "<wallet txid>", "error": "" }
```
Valid statuses are `signed`, `broadcast` and `failed`. A result already marked `broadcast` cannot be overwritten, so retrying after a crash cannot double-record a payout. The run becomes `complete` when every result is `broadcast`. This endpoint does not sign or broadcast anything.

`GET /api/payroll/{id}/export` requires the bearer token and returns CSV with columns `name,address,amount,status,txid,error`.

`GET /api/payroll/{id}` requires the bearer token and returns the draft run and payout plan. Uploading does not send funds.

## Disclosure

`POST /api/disclose` requires the bearer token.
```json
{
  "subject": "Lender X",
  "threshold": 1000000,
  "expiresInDays": 30,
  "anchor": { "txid": "<confirmed txid>", "vout": 0, "amount": 25000 }
}
```
The `anchor` is optional. When present, the txid, output index and amount are inside the signed claim, so changing them invalidates the signature. `201` response:
```json
{ "id": "DP_...", "proof": "<Ed25519 signature>", "verifyUrl": "/proof/DP_..." }
```

`GET /api/proofs/{id}` returns the statement, subject, threshold, expiry, signature, public key, signer fingerprint and `anchorCommitment`. It never returns a balance or the raw anchor before disclosure. The signed claim contains the commitment, not the txid.

`POST /api/proofs/{id}/reveal` requires the bearer token and must be called by the proof owner. It discloses the raw anchor and marks the proof `revealed`, so a verifier can hash it and compare with `anchorCommitment`. This is the dispute path, not the default.

The public verification page checks the signature in the browser. A valid signature only proves who signed it; the verifier must still trust that key and check the anchor. This is a signed statement, not a zero-knowledge proof.

## Guide

`POST /api/assistant` with `{"message":"how do I create a link?"}` returns `{"answer":"...","link":"/links"}`. It is rules-based over real app routes and real data, with no LLM.

## Errors

```json
{ "error": { "code": "ERROR", "message": "invalid amount" } }
```

Payroll uploads are capped at 8 MB. Records are stored in Supabase when configured, otherwise in a local JSON file.

## Source client

`web/static/sdk/pagor.js` is a dependency-free ES module over this API. It is served at `/static/sdk/pagor.js` and is not published to a package registry.

```js
import { PagorClient } from '/static/sdk/pagor.js'
const pagor = new PagorClient({ baseUrl: location.origin, token })
const link = await pagor.createLink({ amount: 25000, note: 'Invoice 42' })
```

Methods: `createLink`, `getLink`, `submitPayroll`, `getPayroll`, `createProof`, `getProof`, `me`, `portfolio`, `logout`.

## Portfolio

`GET /api/portfolio` requires the bearer token. It returns counts and totals from saved records. It never invents a balance:

```json
{
  "links": 2,
  "payrollRuns": 1,
  "proofs": 3,
  "linkTotal": 50000,
  "payrollTotal": 98000
}
```

## Persistence

Records and the signing key are stored in `data/pagor.json` by default. The path can be changed with `EDJ_DATA_FILE`. The file is ignored by git. Prototype data survives a restart; it is not a substitute for an audited database.

## Security note

Write endpoints and account reads require the bearer token. Public reads are limited to a link by its unguessable id and a disclosure statement by its id. The prototype is not audited for custody of funds.
