# Pagor MCP server

`mcp/server.mjs` is a JSON-RPC over stdio MCP server. It gives an agent scoped access to Pagor primitives without custody of funds or keys.

## Tools

- `createLink`: creates a private link. Amounts above `PAGOR_MCP_MAX_AMOUNT` are rejected by the server; the agent cannot raise the ceiling.
- `submitPayroll`: encrypts rows on the machine running the server and sends only the ciphertext to `/api/private-payroll`.
- `createProof`: creates a signed threshold statement with an optional on-chain anchor commitment.
- `getLink` and `getProof`: read-only lookups. Raw proof anchors stay hidden until the owner discloses them.

## Configuration

| Variable | Purpose |
|---|---|
| `PAGOR_API_BASE` | API base URL. Defaults to the Railway service. |
| `PAGOR_API_TOKEN` | Pagor session token for the agent's account. |
| `PAGOR_MCP_MAX_AMOUNT` | Hard spending ceiling in satoshis. Defaults to 100000. |

Set a low ceiling. The agent can create links but never holds keys and never sees a hosted balance.

## Run and test

```bash
cd mcp
npm test
PAGOR_API_TOKEN=<token> PAGOR_MCP_MAX_AMOUNT=50000 npm start
```

The server reads one JSON-RPC request per stdin line and writes one response per line. It implements `initialize`, `tools/list` and `tools/call`.

## Honest limits

- There is no auto-broadcast and no signing in the MCP server. Payments still require a wallet.
- The agent's token is a full Pagor session token. Limit its lifetime and scope at the account level.
- `submitPayroll` needs Web Crypto in the Node process running the server.
