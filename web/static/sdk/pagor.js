/**
 * pagor v1 source client.
 * Thin fetch wrapper over the live HTTP API. No dependencies, no build step.
 * This file ships with the app; it is not published to a package registry.
 */
const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

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

async function derivePayrollKey(subtle, passphrase, salt) {
  if (!passphrase || passphrase.length < 10) throw new TypeError('passphrase must be at least 10 characters');
  const material = await subtle.importKey('raw', textEncoder.encode(passphrase), 'PBKDF2', false, ['deriveKey']);
  return subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 150000, hash: 'SHA-256' },
    material,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
}

export class PagorClient {
  constructor({ baseUrl = (typeof window !== 'undefined' && window.EDJ_API_BASE) || '', token = null } = {}) {
    this.baseUrl = String(baseUrl).replace(/\/$/, '');
    this.token = token;
  }

  #headers(json) {
    const headers = {};
    if (json) headers['Content-Type'] = 'application/json';
    if (this.token) headers.Authorization = `Bearer ${this.token}`;
    return headers;
  }

  async #request(path, options = {}) {
    const response = await fetch(this.baseUrl + path, options);
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = new Error(data?.error?.message || `Request failed with ${response.status}`);
      error.code = data?.error?.code || 'ERROR';
      error.status = response.status;
      throw error;
    }
    return data;
  }

  createLink({ amount, token = null, note = '' }) {
    if (!Number.isInteger(amount) || amount <= 0) {
      throw new TypeError('amount must be a positive integer');
    }
    return this.#request('/api/links', {
      method: 'POST',
      headers: this.#headers(true),
      body: JSON.stringify({ amount, token, note })
    });
  }

  getLink(id) {
    return this.#request(`/api/links/${encodeURIComponent(id)}`);
  }

  /** Deprecated: sends plaintext CSV. Use createPrivatePayroll instead. */
  submitPayroll(csv, filename = 'payroll.csv') {
    const form = new FormData();
    form.append('csv', new Blob([csv], { type: 'text/csv' }), filename);
    return this.#request('/api/payroll/upload', {
      method: 'POST',
      headers: this.#headers(false),
      body: form
    });
  }

// Encrypted payroll. Names, addresses and amounts never leave the device.
  async createPrivatePayroll(rows, passphrase) {
    if (!Array.isArray(rows) || rows.length === 0) throw new TypeError('rows must be a non-empty array');
    const subtle = globalThis.crypto?.subtle;
    if (!subtle) throw new Error('Web Crypto is not available');
    const salt = globalThis.crypto.getRandomValues(new Uint8Array(16));
    const iv = globalThis.crypto.getRandomValues(new Uint8Array(12));
    const key = await derivePayrollKey(subtle, passphrase, salt);
    const encrypted = await subtle.encrypt(
      { name: 'AES-GCM', iv },
      key,
      textEncoder.encode(JSON.stringify(rows))
    );
    return this.#request('/api/private-payroll', {
      method: 'POST',
      headers: this.#headers(true),
      body: JSON.stringify({
        salt: bytesToBase64(salt),
        iv: bytesToBase64(iv),
        ciphertext: bytesToBase64(new Uint8Array(encrypted)),
        rowCount: rows.length
      })
    });
  }

  getPrivatePayroll(id) {
    return this.#request(`/api/private-payroll/${encodeURIComponent(id)}`);
  }

  async decryptPayroll(payload, passphrase) {
    const subtle = globalThis.crypto?.subtle;
    if (!subtle) throw new Error('Web Crypto is not available');
    const key = await derivePayrollKey(subtle, passphrase, base64ToBytes(payload.salt));
    const plain = await subtle.decrypt(
      { name: 'AES-GCM', iv: base64ToBytes(payload.iv) },
      key,
      base64ToBytes(payload.ciphertext)
    );
    return JSON.parse(textDecoder.decode(plain));
  }

  getPayroll(id) {
    return this.#request(`/api/payroll/${encodeURIComponent(id)}`);
  }

  createProof({ subject, threshold, expiresInDays = 30 }) {
    if (!subject || !Number.isInteger(threshold) || threshold <= 0) {
      throw new TypeError('subject and a positive integer threshold are required');
    }
    return this.#request('/api/disclose', {
      method: 'POST',
      headers: this.#headers(true),
      body: JSON.stringify({ subject, threshold, expiresInDays })
    });
  }

  getProof(id) {
    return this.#request(`/api/proofs/${encodeURIComponent(id)}`);
  }

  me() {
    return this.#request('/api/me', { headers: this.#headers(false) });
  }

  portfolio() {
    return this.#request('/api/portfolio', { headers: this.#headers(false) });
  }

  logout() {
    return this.#request('/api/logout', { method: 'POST', headers: this.#headers(false) });
  }
}

export const createClient = (options) => new PagorClient(options);
