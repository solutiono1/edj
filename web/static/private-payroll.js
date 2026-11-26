// Client-side payroll encryption. Names, addresses and amounts are encrypted
// in the browser before upload. The server stores only an opaque blob and
// cannot read any payroll detail. The user must keep the passphrase.

const encoder = new TextEncoder();
const decoder = new TextDecoder();

function bytesToBase64(bytes) {
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return typeof btoa === 'function' ? btoa(binary) : Buffer.from(bytes).toString('base64');
}

function base64ToBytes(value) {
  if (typeof atob === 'function') {
    const binary = atob(value);
    const out = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i);
    return out;
  }
  return new Uint8Array(Buffer.from(value, 'base64'));
}

async function deriveKey(passphrase, salt) {
  if (!passphrase || passphrase.length < 10) throw new Error('Use a passphrase of at least 10 characters');
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) throw new Error('Web Crypto is not available');
  const material = await subtle.importKey('raw', encoder.encode(passphrase), 'PBKDF2', false, ['deriveKey']);
  return subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 150000, hash: 'SHA-256' },
    material,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
}

export async function encryptRows(rows, passphrase) {
  if (!Array.isArray(rows) || rows.length === 0) throw new Error('No payroll rows');
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) throw new Error('Web Crypto is not available');
  if (!globalThis.crypto?.getRandomValues) throw new Error('Web Crypto is not available');
  const salt = globalThis.crypto.getRandomValues(new Uint8Array(16));
  const iv = globalThis.crypto.getRandomValues(new Uint8Array(12));
  const key = await deriveKey(passphrase, salt);
  const encrypted = await subtle.encrypt({ name: 'AES-GCM', iv }, key, encoder.encode(JSON.stringify(rows)));
  return {
    salt: bytesToBase64(salt),
    iv: bytesToBase64(iv),
    ciphertext: bytesToBase64(new Uint8Array(encrypted)),
    rowCount: rows.length
  };
}

export async function decryptRows(payload, passphrase) {
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) throw new Error('Web Crypto is not available');
  const key = await deriveKey(passphrase, base64ToBytes(payload.salt));
  const plain = await subtle.decrypt(
    { name: 'AES-GCM', iv: base64ToBytes(payload.iv) },
    key,
    base64ToBytes(payload.ciphertext)
  );
  return JSON.parse(decoder.decode(plain));
}

export function parsePayrollCsv(text) {
  const lines = String(text || '').trim().split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  if (lines.length < 2) throw new Error('CSV needs a header and at least one row');
  const header = lines[0].split(',').map((part) => part.trim().toLowerCase());
  if (header.join(',') !== 'name,address,amount') throw new Error('Expected header name,address,amount');
  return lines.slice(1).map((line) => {
    const parts = line.split(',').map((part) => part.trim());
    if (parts.length !== 3) throw new Error('Each row must have name,address,amount');
    const amount = Number(parts[2]);
    if (!Number.isInteger(amount) || amount <= 0) throw new Error('Amount must be a positive integer');
    return { name: parts[0], address: parts[1], amount };
  });
}
