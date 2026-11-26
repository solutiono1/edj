# SMART CONTRACT

## What this is

pagor uses Bitcoin Cash. BCH does not use EVM contracts. This contract is BCH Script using native introspection opcodes, often called a covenant. It restricts where locked funds can be sent. It does not provide anonymity, and it is not a zero-knowledge proof.

The contract is a prototype. It is not audited and is not used by the payment link flow yet.

## Contract: PrivatePayroll

Source: `tools/contract/PrivatePayroll.cash`
Compiled artifact: `tools/contract/PrivatePayroll.json`

Constructor arguments:

| Argument | Type | Meaning |
|---|---|---|
| `payeePkh` | bytes20 | Public key hash allowed to receive the payout |
| `payerPkh` | bytes20 | Public key hash allowed to receive a refund |
| `paymentAmount` | int | Exact payout amount in satoshis |
| `refundAfter` | int | Block height after which a refund is allowed |

Functions:

- `pay` requires exactly one output, the exact payment amount, and a P2PKH output to `payeePkh`.
- `refund` requires the refund height to be reached, then applies the same output rules to `payerPkh`.

Both functions are permissionless. That is safe here because the output address and amount are fixed in the contract. Anyone can call the function, but nobody can redirect the funds.

## What it proves

The covenant proves that a locked BCH amount can only leave the contract as a single payment of the fixed amount to the fixed recipient, or after the fixed height as a refund to the payer. This is a real on-chain constraint, not a promise in the UI.

## What it does not prove

- It does not hide the amount or the addresses. BCH settlement is public.
- It does not mix funds. CashFusion remains a separate user-chosen step.
- It does not have a multi-output payroll engine yet.
- It does not handle CashTokens yet.
- It has been deployed to chipnet only. It has not been deployed to mainnet.

## Local test

Commands:

```bash
cd tools/contract
npm install
npm test
```

`npm test` compiles the contract with `cashc` and evaluates it with CashScript's mock network provider. Current result:

- PASS pay covenant evaluates and constrains the output
- PASS refund covenant evaluates and constrains the output
- PASS pay covenant rejects a different recipient


## Contract: PrivatePayrollBatch

Source: `tools/contract/PrivatePayrollBatch.cash`
Compiled artifact: `tools/contract/PrivatePayrollBatch.json`

CashScript 0.14 has no array parameters, so capacity is four recipients passed as individual arguments. Constructor arguments:

| Argument | Type | Meaning |
|---|---|---|
| `payee0` to `payee3` | bytes20 | P2PKH hashes allowed to receive payouts, in order |
| `amount0` to `amount3` | int | Exact payout amount for each slot |
| `count` | int | Number of active slots, 1 to 4 |
| `total` | int | Full locked amount, used by the refund path |
| `payerPkh` | bytes20 | Hash allowed to receive the refund |
| `refundAfter` | int | Block height after which refund is allowed |

Functions:

- `pay` requires `count` outputs, each with the exact amount and P2PKH of its slot.
- `refund` requires the refund height, one output of exactly `total`, and the payer P2PKH.

Both functions are permissionless because every destination and amount is fixed in the script. Tests in `tools/contract/test-batch.mjs` cover two and four recipients, wrong recipient, wrong amount, wrong output count, early refund, and refund to the wrong address. `npm test` compiles both contracts and runs both suites.

This version has not been deployed to chipnet or mainnet.

## Deployed (chipnet)

- Transaction: `d34a9391d4ac6eb5ff695bfb9aafa8f4e0aa8aa4cac3702a9c7de957fce0d18a`
- Contract address: `bchtest:pv5mgtpn73gp9sl7qaxll5augr28rlvuwkqh0cw3n0nemxxcz95c5rxdu0zyl`
- Owner (funded) address: `bchtest:qrt3gw5d64zwk9hqmxujekeufcj2a3z90sywczl3v5`
- Date: 23 November 2026

This is a chipnet deployment. It is not mainnet.
## Chipnet test checklist

The deploy script is in `tools/contract/deploy-chipnet.mjs`. It reads a funded WIF from `CHIPNET_WIF` and never writes it to a file. See `DEPLOYMENT.md` for the exact command.

This part needs a funded chipnet wallet and has not been run here yet.

1. Fund the contract address shown by the test output with the exact payment amount plus a fee.
2. Build a `pay` spend with one output to the payee P2PKH address and no change output.
3. Broadcast on chipnet and confirm the transaction.
4. Repeat with the `refund` function after the refund height.
5. Try a spend to a different address and confirm the node rejects it.
6. Record the txids and the rejection reason in a review file.

## Linkability note

A fresh address per link does not hide on-chain amounts. Funding many payouts from one treasury UTXO links those payouts publicly. Prefer one funding UTXO per recipient, or a covenant per recipient, when recipient privacy matters. pagor does not claim this is anonymity.

## Audit scope

Before any mainnet use, review:

- Output count and exact amount checks.
- Change handling. The contract intentionally rejects change.
- Timelock behavior across block reorganizations.
- P2SH32 address type and script size.
- Fee sniping and transaction replacement considerations.
- Token handling if a CashTokens version is added.
- Key custody for payer and payee public key hashes.
- Independent script review and a chipnet test report.
