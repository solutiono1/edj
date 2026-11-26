// Public verification of a signed statement. No server call is needed to
// check the signature: the key is embedded in the proof. The verifier must
// still decide whether they trust that key and its on-chain anchor.

function base64ToBytes(value) {
  if (typeof atob === 'function') {
    const binary = atob(value);
    const out = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i);
    return out;
  }
  return new Uint8Array(Buffer.from(value, 'base64'));
}

function bytesToHex(bytes) {
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

export async function fingerprintOf(publicKeyBase64) {
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) throw new Error('Web Crypto is not available');
  const digest = await subtle.digest('SHA-256', base64ToBytes(publicKeyBase64));
  return bytesToHex(new Uint8Array(digest));
}

export async function verifyProof(proof) {
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) throw new Error('Web Crypto is not available');
  if (!proof?.publicKey || !proof?.signature || !proof?.claim) return false;
  const fingerprint = await fingerprintOf(proof.publicKey);
  if (proof.signerFingerprint && fingerprint !== proof.signerFingerprint) return false;
  try {
    const key = await subtle.importKey('raw', base64ToBytes(proof.publicKey), { name: 'Ed25519' }, false, ['verify']);
    return await subtle.verify({ name: 'Ed25519' }, key, base64ToBytes(proof.signature), new TextEncoder().encode(proof.claim));
  } catch (error) {
    return false;
  }
}

export function anchorSummary(proof) {
  const commitment = proof?.anchorCommitment;
  if (!commitment) return 'No on-chain anchor was supplied. This statement is not backed by a checkable output.';
  if (proof.revealed && proof.anchor?.txid) {
    return 'Anchor disclosed after dispute: ' + proof.anchor.txid + ' output ' + proof.anchor.vout + ', ' + proof.anchor.amount + ' sats. Its hash must equal ' + commitment + '.';
  }
  return 'The signed claim commits to anchor ' + commitment + '. The raw transaction id is private until the owner discloses it on dispute.';
}
