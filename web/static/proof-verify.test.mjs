import test from 'node:test';
import assert from 'node:assert/strict';
import { fingerprintOf, verifyProof, anchorSummary } from './proof-verify.js';

const keyPair = await globalThis.crypto.subtle.generateKey({ name: 'Ed25519' }, false, ['sign', 'verify']);
const raw = new Uint8Array(await globalThis.crypto.subtle.exportKey('raw', keyPair.publicKey));
const publicKey = typeof btoa === 'function'
  ? btoa(String.fromCharCode(...raw))
  : Buffer.from(raw).toString('base64');
const claim = 'holds at least 3000 for Lender X until 2026-12-31T00:00:00Z';
const signatureBytes = await globalThis.crypto.subtle.sign({ name: 'Ed25519' }, keyPair.privateKey, new TextEncoder().encode(claim));
const signature = typeof btoa === 'function'
  ? btoa(String.fromCharCode(...new Uint8Array(signatureBytes)))
  : Buffer.from(signatureBytes).toString('base64');
const fingerprint = await fingerprintOf(publicKey);

test('valid proof verifies', async () => {
  const proof = { publicKey, signature, claim, signerFingerprint: fingerprint };
  assert.equal(await verifyProof(proof), true);
});

test('tampered claim fails', async () => {
  const proof = { publicKey, signature, claim: claim + ' tampered', signerFingerprint: fingerprint };
  assert.equal(await verifyProof(proof), false);
});

test('fingerprint mismatch fails', async () => {
  const proof = { publicKey, signature, claim, signerFingerprint: '00'.repeat(32) };
  assert.equal(await verifyProof(proof), false);
});

test('anchor text hides the txid until disclosed', () => {
  assert.match(anchorSummary({}), /No on-chain anchor/);
  const hidden = anchorSummary({ anchorCommitment: 'deadbeef' });
  assert.match(hidden, /commits to anchor deadbeef/);
  assert.doesNotMatch(hidden, /abc/);
  const shown = anchorSummary({ anchorCommitment: 'deadbeef', revealed: true, anchor: { txid: 'abc', vout: 1, amount: 100 } });
  assert.match(shown, /abc output 1/);
});
