# edj

Private payments and payroll on Bitcoin Cash.

edj is a web app for freelancers and small teams who want to pay and get paid without publishing amounts on a public chain. It creates private payment links, prepares payroll runs from a CSV, and issues signed disclosure statements.

edj does not hold funds. It is not anonymous, and disclosure statements are not zero-knowledge proofs.

## What it does

- Creates a private payment link with a receiving address you control. The payer opens the link and pays from any BCH wallet.
- Parses a payroll CSV and shows the recipients and total before anything is sent.
- Creates a signed statement that a named reader can check, with a threshold and an expiry. The statement does not reveal a balance.
- Signs users in with Google and keeps a session token in the browser. No cookies are used.
- Provides a small JavaScript SDK over the same HTTP API used by the app.

## How it is built

- HTTP server in Go, using the standard library only.
- Static HTML pages and hand-written CSS in `web/`.
- htmx is vendored locally. There is no React, TypeScript, or frontend build step.
- Records are stored in a local JSON file so they survive a restart. The file is ignored by git.
- Chain addresses come from the environment. edj never generates a random receiving address.

The core packages are:

| Path | Purpose |
|---|---|
| `cmd/edj` | Routes and process startup |
| `internal/api` | HTTP handlers and session checks |
| `internal/core` | Links, payroll parsing, and signed statements |
| `internal/chain` | Bitcoin Cash address adapter |
| `internal/store` | Persistent record store |
| `web` | Pages, styles, scripts, icons, and SDK |

## Quick start

You need Go 1.27, a Google Cloud OAuth client ID, and BCH addresses that you control.

```bash
cd /home/solution/edj
cp .env.example .env
```

Edit `.env`:

| Variable | Purpose |
|---|---|
| `GOOGLE_CLIENT_ID` | Google client ID for sign-in |
| `EDJ_RECEIVE_ADDRESS` | BCH address for BCH payment links |
| `EDJ_TOKEN_ADDRESS` | CashTokens address for token payment links |
| `ADDR` | Listen address, default `:8080` |
| `EDJ_DATA_FILE` | Store path, default `data/edj.json` |

For Google sign-in, add `http://localhost:8080` and `http://127.0.0.1:8080` as authorized JavaScript origins. See `GOOGLE-OAUTH.md`.

Run the app:

```bash
./start.sh
```

Open `http://localhost:8080`.

## Development

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test ./...
go build -o /tmp/opencode/edj-server ./cmd/edj
```

The server is started by `start.sh`, which loads `.env` and exports the configured addresses.

Useful checks:

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/config
```

Unauthenticated writes return HTTP 401. A complete list of endpoints is in `API.md`.

## SDK

The source client is served at `/static/sdk/edj.js`.

```js
import { EdjClient } from '/static/sdk/edj.js'

const edj = new EdjClient({
  baseUrl: location.origin,
  token: localStorage.getItem('edj_session')
})

const link = await edj.createLink({ amount: 25000, note: 'Invoice 42' })
```

The client wraps link, payroll, proof, session, and logout calls.

## Deployment

Set the environment variables above, build the binary, and serve it behind HTTPS. Add the public domain as an authorized JavaScript origin for the Google client. Keep `.env` and `data/` out of git.

## Security and honesty

- Session tokens live in localStorage. There are no cookies and no tracking scripts.
- Disclosure statements are signed statements with a named reader, threshold, and expiry. They are not zero-knowledge proofs.
- Settlement on Bitcoin Cash is public. edj reduces what is published for specific flows; it does not make a payment anonymous.
- The app contains no mock records. The JSON store holds real records created through the app.
- The prototype has not been audited for custody of funds. Do not treat it as a wallet or exchange.

## Documentation

| File | What it covers |
|---|---|
| `SPEC.md` | Scope, honesty rules, and hackathon requirements |
| `ARCHITECTURE.md` | System boundaries and chain adapter |
| `API.md` | Endpoints and request shapes |
| `UIUX.md` | Page order and screen behavior |
| `BRAND.md` | Colors, type, and motion tokens |
| `MOTION.md` | Animation rules |
| `ASSETS.md` | Source and license register |
| `GOOGLE-OAUTH.md` | OAuth setup and authorized origins |
| `DECISIONS.md` | Decision log |
| `ROADMAP.md` | Remaining work |
| `HANDOFF.md` | Repository ownership rules |
| `SMART-CONTRACT.md` | BCH covenant design, tests, chipnet checklist, audit scope |
| `DEV-TOOLS.md` | SDK, CLI, and MCP plan |

## Contributing

The human is the only author and approver. AI agents may write code and documentation, but they do not stage, commit, push, branch, merge, or open pull requests.

Do not add references to or credits for other products in code, docs, UI copy, or commit messages.

## Assets

Third-party media and font licenses are recorded in `ASSETS.md`. Do not add analytics, advertising, or tracking SDKs.
