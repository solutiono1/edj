# SMART CONTRACT

## What this is

edj uses Bitcoin Cash. BCH does not use EVM contracts. This contract is BCH Script using native introspection opcodes, often called a covenant. It restricts where locked funds can be sent. It does not provide anonymity, and it is not a zero-knowledge proof.

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
- It has not been deployed to mainnet or chipnet in this repository.

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

## Chipnet test checklist

The deploy script is in `tools/contract/deploy-chipnet.mjs`. It reads a funded WIF from `CHIPNET_WIF` and never writes it to a file. See `DEPLOYMENT.md` for the exact command.

This part needs a funded chipnet wallet and has not been run here yet.

1. Fund the contract address shown by the test output with the exact payment amount plus a fee.
2. Build a `pay` spend with one output to the payee P2PKH address and no change output.
3. Broadcast on chipnet and confirm the transaction.
4. Repeat with the `refund` function after the refund height.
5. Try a spend to a different address and confirm the node rejects it.
6. Record the txids and the rejection reason in a review file.

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
