# Response to external review

Review received 26 Nov 2026. Each point is recorded with its status, not just dismissed or accepted.

## 1. Public UTXO linkability

**Valid in part.** A private link hides the amount from our UI, but Bitcoin Cash settlement is public. An observer can see amounts and addresses on a block explorer.

What already existed: `EDJ_XPUB` gives every link a fresh `m/0/index` P2PKH address, recorded as `derivationIndex`. Two links never reuse an address. This is a fresh single-use receiving address, though it is not the Cash Code / RPA standard.

What is still true: funding many payouts from one treasury UTXO links those payouts on-chain. We do not hide that. The payroll UI and `SMART-CONTRACT.md` now state it explicitly and recommend funding each payout from a separate UTXO. Cash Code / RPA remains research, not a shipped integration.

## 2. Self-hosted attestor spoofing

**Valid.** A signature from a locally generated key proves only that whoever held that key signed the statement. A verifier who trusts the wrong key can be fooled.

What changed:

- Every proof now carries `signerFingerprint`, the SHA-256 of the Ed25519 public key.
- An optional on-chain `anchor` (txid, output index, amount in satoshis) is inside the signed claim. Editing it invalidates the signature.
- `GET /api/keys` publishes the current public key and fingerprint.
- The public verification page verifies the signature in the browser with Web Crypto, shows the fingerprint and anchor, and offers key pinning on the verifier's device.

What is deliberately not claimed: there is no central registry of verified employer keys, and we do not verify the anchor for the verifier. That would just move the trust to us. The verifier decides which keys they trust and checks the anchor themselves.

## 3. Single-output covenant and JSON file store

**Outdated.** Both parts were already addressed before this review:

- `PrivatePayrollBatch.cash` supports four fixed payouts with a `count` guard and a refund path. `npm test` runs two-recipient and four-recipient cases plus rejection cases. The original single-payee `PrivatePayroll.cash` is retained as the simpler one-person version.
- Supabase is the production store. The JSON file is local development only.

What is still true: the batch covenant is chipnet-mock tested, not deployed, not audited, and not wired into the payroll UI. Record ownership columns now exist so account deletion can purge a user's records.

## Domain

`usepagor.xyz` has been acquired. All page metadata and example links use it. Railway `EDJ_PUBLIC_ORIGIN` and the Google OAuth origin still need to be set to `https://usepagor.xyz` before the custom domain is served.
