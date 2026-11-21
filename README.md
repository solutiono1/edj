<p align="center">
  <img src="web/static/hero-preview.png" alt="edj landing page" width="720">
</p>

# edj

Private payments and payroll on Bitcoin Cash. Amounts stay out of the public story of a payment, and signed statements let someone check a threshold when needed.

[Local app](web/index.html) · [API](API.md) · [SDK](DEV-TOOLS.md) · [Contract](SMART-CONTRACT.md)

**Prototype running locally. No public deployment yet. No real ZK proofs. No custody of funds.**

## What works today

Everything below runs on localhost with a real Go server and a local JSON store.

| Feature | Status | What you can do |
|---|---|---|
| Private payment links | Local prototype | Create a link with an operator-configured BCH or CashTokens address. The payer opens the link and pays from any BCH wallet. |
| Payroll drafts | Local prototype | Upload a CSV with name, address and amount. Review recipients and total before any payout. Payout execution stays with the wallet. |
| Disclosure statements | Local prototype | Create an Ed25519 signed statement with a subject, threshold and expiry. It is not a zero-knowledge proof. The verification page shows the signature and public key. |
| Google sign-in | Working in local testing | Sign in with Google Identity Services. The server verifies the audience and verified email. The session token stays in localStorage. |
| Persistent store | Local file | Links, payroll drafts, proofs, users, sessions and the signing key survive a restart in `data/edj.json`. |
| Portfolio | Local API | The app shows real counts of saved links, payroll runs and proofs from `/api/portfolio`. No invented balance. |
| SDK | Source package | The client is served at `/static/sdk/edj.js` and packaged at `sdk/js`. It wraps the live HTTP API. |
| CLI | Local tool | `tools/cli/edj.mjs` calls the same API. Tested with `me`, `portfolio` and `links create`. |
| BCH covenant | Compiled and mock-tested | `tools/contract/PrivatePayroll.cash` constrains a spend to one output, a fixed amount and a fixed address. Chipnet steps are documented, not run here. |
| MCP | Planned | Scope is documented in `DEV-TOOLS.md`. No MCP server is shipped yet. |

## What edj does not do

- It does not make a payment anonymous. Bitcoin Cash settlement is public.
- It does not provide zero-knowledge proofs. Disclosure statements are signed statements with a named reader and an expiry.
- It does not hold funds or sign payments for the user.
- It does not deploy a mainnet contract in this repository.
- It does not use a production database. The local JSON store is a prototype store, not an audited system.

## Run it locally

Requirements:

- Go 1.27
- A Google Cloud OAuth client ID
- BCH addresses you control

```bash
cd /home/solution/edj
cp .env.example .env
./start.sh
```

Open `http://localhost:8080`.

`.env` values:

| Variable | Default | Purpose |
|---|---|---|
| `GOOGLE_CLIENT_ID` | none | Google client ID for sign-in |
| `EDJ_RECEIVE_ADDRESS` | none | BCH address for BCH links |
| `EDJ_TOKEN_ADDRESS` | none | CashTokens address for token links |
| `ADDR` | `:8080` | Listen address |
| `EDJ_DATA_FILE` | `data/edj.json` | Store path |

For Google sign-in, add `http://localhost:8080` and `http://127.0.0.1:8080` as authorized JavaScript origins. See `GOOGLE-OAUTH.md`.

## Tests

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test ./...

cd tools/contract
npm install
npm test
```

The contract test compiles the CashScript source and evaluates the covenant with a mock network provider. Current results:

- pay covenant evaluates and constrains the output
- refund covenant evaluates and constrains the output
- pay covenant rejects a different recipient

## Layout

```
cmd/edj/          server entrypoint and routes
internal/api/     handlers, session checks, portfolio
internal/core/    links, payroll parsing, signed statements
internal/chain/   BCH address adapter
internal/store/   persistent JSON store
web/              pages, CSS, scripts, icons, SDK
sdk/js/           packaged JavaScript client
tools/cli/        command line client
tools/contract/   BCH covenant source, artifact and tests
```

## SDK

```js
import { EdjClient } from '/static/sdk/edj.js'

const edj = new EdjClient({
  baseUrl: location.origin,
  token: localStorage.getItem('edj_session')
})

const link = await edj.createLink({ amount: 25000, note: 'Invoice 42' })
console.log(link.claimUrl)
```

Methods: `createLink`, `getLink`, `submitPayroll`, `getPayroll`, `createProof`, `getProof`, `me`, `portfolio`, `logout`.

## CLI

```bash
EDJ_TOKEN=<token> node tools/cli/edj.mjs me
EDJ_TOKEN=<token> node tools/cli/edj.mjs portfolio
EDJ_TOKEN=<token> node tools/cli/edj.mjs links create 5000 - 'Invoice 42'
```

`EDJ_URL` defaults to `http://localhost:8080`.

## Contract

`PrivatePayroll` is a BCH Script covenant, not an EVM contract. It restricts a locked payment to one output, a fixed amount and a fixed address, with a refund after a fixed block height.

Source: `tools/contract/PrivatePayroll.cash`
Design and chipnet checklist: `SMART-CONTRACT.md`

It is compiled and locally evaluated. It has not been deployed to chipnet or mainnet here.

## Trust and limits

- Session tokens live in localStorage. No cookies and no tracking scripts.
- Write endpoints and account reads require a bearer token. The UI is not the security boundary.
- The local JSON store is ignored by git.
- The signing key is generated locally and stored with the records.
- The prototype has no independent security audit.

## Docs

| File | What it covers |
|---|---|
| `SPEC.md` | Scope, honesty rules and hackathon requirements |
| `API.md` | Endpoints and request shapes |
| `ARCHITECTURE.md` | Current system boundaries |
| `SMART-CONTRACT.md` | BCH covenant design, tests and audit scope |
| `DEV-TOOLS.md` | SDK, CLI and MCP plan |
| `BRAND.md` | Colors, type and motion tokens |
| `UIUX.md` | Page order and screen behavior |
| `MOTION.md` | Animation rules |
| `ASSETS.md` | Source and license register |
| `GOOGLE-OAUTH.md` | OAuth setup and authorized origins |
| `DECISIONS.md` | Decision log |
| `ROADMAP.md` | Remaining work |
| `HANDOFF.md` | Repository ownership rules |

## Contributing

The human is the only author and approver. AI agents may write code and documentation, but they do not stage, commit, push, branch, merge or open pull requests.

Do not add references to or credits for other products in code, docs, UI copy or commit messages.

## License

MIT. See the license file in the repository. Third-party media and font licenses are recorded in `ASSETS.md`.
