# DEV TOOLS

## Source of truth

`API.md` defines endpoints, request shapes, and responses. The SDK and CLI wrap that API. If the API changes, update `API.md` first, then the tools.

## JavaScript SDK

Package: `sdk/js`
Browser copy: `web/static/sdk/edj.js`

Both files contain the same `EdjClient` source. Methods:

| Method | Endpoint |
|---|---|
| `createLink` | `POST /api/links` |
| `getLink` | `GET /api/links/{id}` |
| `submitPayroll` | `POST /api/payroll/upload` |
| `getPayroll` | `GET /api/payroll/{id}` |
| `createProof` | `POST /api/disclose` |
| `getProof` | `GET /api/proofs/{id}` |
| `me` | `GET /api/me` |
| `portfolio` | `GET /api/portfolio` |
| `logout` | `POST /api/logout` |

The client accepts a `baseUrl` and a session token. Errors carry `status` and `code` from the API response.

To use the package from another local project:

```bash
npm install /absolute/path/to/edj/sdk/js
```

## CLI

Path: `tools/cli/edj.mjs`

Environment:

| Variable | Default | Meaning |
|---|---|---|
| `EDJ_URL` | `http://localhost:8080` | API base URL |
| `EDJ_TOKEN` | none | Session token, required |

Commands:

```bash
EDJ_TOKEN=<token> node tools/cli/edj.mjs me
EDJ_TOKEN=<token> node tools/cli/edj.mjs portfolio
EDJ_TOKEN=<token> node tools/cli/edj.mjs links create 5000 - 'Invoice 42'
EDJ_TOKEN=<token> node tools/cli/edj.mjs payroll upload web/static/example-payroll.csv
EDJ_TOKEN=<token> node tools/cli/edj.mjs proofs create 'Lender review' 100000 30
```

All commands were tested against the running server with a seeded session. The seed and its data were removed after the test.

## Contract tools

Path: `tools/contract`

```bash
cd tools/contract
npm install
npm test
```

`npm test` compiles `PrivatePayroll.cash` and evaluates the covenant with a mock network provider. The design, limits, chipnet checklist, and audit scope are in `SMART-CONTRACT.md`.

## MCP plan

MCP is planned, not implemented. It should expose a narrow set of tools over the same API:

| Tool | Input | API call |
|---|---|---|
| `edj_create_link` | amount, token, note | `POST /api/links` |
| `edj_create_proof` | subject, threshold, expiry days | `POST /api/disclose` |
| `edj_get_portfolio` | none | `GET /api/portfolio` |

Rules for the future MCP server:

- Require a scoped token with only the operations the agent needs.
- Never expose logout or raw session creation.
- Keep an allowlist of destinations and amounts.
- Record every tool call with a request id and a redacted token reference.
- Do not claim anonymity or ZK behavior.
