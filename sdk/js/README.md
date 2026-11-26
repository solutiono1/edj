# @pagorpay/sdk

A dependency-free JavaScript client for the pagor HTTP API.

The same source is served by the app at `/static/sdk/pagor.js` and published here as a package for local tooling. API shapes are defined in `API.md`.

## Install

```bash
npm install @pagorpay/sdk
```

## Use

```js
import { PagorClient } from '@pagorpay/sdk'

const pagor = new PagorClient({
  baseUrl: 'https://pagor.finance',
  token: localStorage.getItem('edj_session')
})

const link = await pagor.createLink({ amount: 25000, note: 'Invoice 42' })
console.log(link.claimUrl)
```

## Methods

| Method | Purpose |
|---|---|
| `createLink({ amount, token, note })` | Create a private payment link |
| `getLink(id)` | Read a link |
| `submitPayroll(csv, filename)` | Upload a payroll CSV |
| `getPayroll(id)` | Read a payroll run |
| `createProof({ subject, threshold, expiresInDays })` | Create a signed statement |
| `getProof(id)` | Read a statement |
| `me()` | Read the current account |
| `logout()` | Delete the server session |

Errors are thrown as `Error` with `status` and `code` properties from the API response.
