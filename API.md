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

`GET /api/me` and `POST /api/logout` require `Authorization: Bearer <token>`.

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
{ "subject": "Lender X", "threshold": 1000000, "expiresInDays": 30 }
```
`201` response:
```json
{ "id": "DP_...", "proof": "<Ed25519 signature>", "verifyUrl": "/proof/DP_..." }
```

`GET /api/proofs/{id}` returns the statement, subject, threshold, expiry, signature and public key. It never returns a balance. This is a signed statement, not a zero-knowledge proof.

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
