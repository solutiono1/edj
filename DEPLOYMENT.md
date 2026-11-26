# DEPLOYMENT

Three services, one flow:

- Cloudflare Pages serves the static frontend from `web/`.
- Railway runs the Go API.
- Supabase stores records if the API runs on a stateless host.

## 1. Supabase

Create a Supabase project. Run `supabase/schema.sql` in the SQL editor. It creates tables for links, payroll runs, proofs, users, sessions, notifications, and the signing key.

Copy the project URL and the service role key. The service role key stays on Render. It is never sent to the browser.

If Supabase is not configured, the Go server falls back to `data/pagor.json` on a local disk. That is fine for local testing and not safe for a stateless host.

## 2. Railway

Connect the GitHub repo in Railway as a web service. `railway.json` sets the Nixpacks build and start commands. Railway supplies `PORT`, and the server reads it.

Set these environment variables:

| Variable | Purpose |
|---|---|
| `GOOGLE_CLIENT_ID` | Optional. Google sign-in client ID. Email sign-in works without it |
| `EDJ_RECEIVE_ADDRESS` | BCH address for BCH links |
| `EDJ_TOKEN_ADDRESS` | CashTokens address for token links |
| `EDJ_PUBLIC_ORIGIN` | Cloudflare Pages origin, for CORS |
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key |

Railway supplies `PORT`. The server reads it automatically.

After deployment, check:

```bash
curl https://<railway-service>/api/health
curl https://<railway-service>/api/config
```

## 3. Cloudflare Pages

Connect the GitHub repo as a Pages project. Set the build output directory to `web`. There is no build command.

Before deploying, set the API base URL in `web/static/config.js`:

```js
window.EDJ_API_BASE = 'https://<railway-service>';
```

Or use the helper script:

```bash
EDJ_API_BASE=https://<railway-service> ./scripts/deploy-cloudflare.sh
```

Every page loads `config.js` before the app scripts. All fetches prepend `window.EDJ_API_BASE`, so the frontend can talk to Render when hosted on Cloudflare.

For a single-origin deploy, leave `EDJ_API_BASE` empty and serve everything from the Go server.

## 4. Sign-in

Email sign-up and login need no extra configuration. They use the `users` and `sessions` tables in Supabase. Apply `supabase/schema.sql` so `users.password_hash` exists. Passwords are hashed with bcrypt on the server.

Google sign-in is optional. If you use it, add these origins in Google Cloud under Authorized JavaScript origins:

- `http://localhost:8080`
- `https://<your Cloudflare Pages domain>`

Do not add a raw IP origin. Google rejects it. If `GOOGLE_CLIENT_ID` is unset, the Google button is hidden and email sign-in still works.

## 5. Chipnet contract

A funded chipnet wallet is required. The WIF is passed by environment variable and is never committed.

```bash
cd tools/contract
npm install
CHIPNET_WIF=<funded-chipnet-wif> \
PAYEE_PKH=<20-byte-hex> \
PAYER_PKH=<20-byte-hex> \
CONTRACT_SATS=25000 \
REFUND_AFTER=<block-height> \
npm run deploy:chipnet
```

Set `DRY_RUN=1` to print the transaction without broadcasting. On a real run, record the txid and contract address in `SMART-CONTRACT.md`.

## 6. Verify

- Landing, login, signup and docs load from Cloudflare.
- Google button renders and sign-in works from the Pages origin.
- Portfolio reads real counts from `/api/portfolio`.
- A private link can be created and opened.
- A disclosure proof can be created and viewed.
- Chipnet txid is recorded for the covenant.

## Current deployment wiring

| Service | Value |
|---|---|
| Supabase project ID | `ibabwmpllpcckxhdeyvz` |
| Supabase URL | `https://ibabwmpllpcckxhdeyvz.supabase.co` |
| Supabase schema | Applied with migration `edj_initial_schema` |
| Railway project ID | `c93645db-f49e-4f71-9ca2-fa5b29c95b8f` |
| Railway service ID | `8e2b2810-9d80-45e9-8efa-13db2467cdfa` |
| Railway API URL | `https://pagor-api-v3-production.up.railway.app` |
| Cloudflare Pages project | `pagor` |
| Cloudflare Pages URL | `https://pagor-b48.pages.dev` |

The service role key is not stored in this repository. Add it as `SUPABASE_SERVICE_ROLE_KEY` in Railway.
