/**
 * edj v1 source client.
 * Thin fetch wrapper over the live HTTP API. No dependencies, no build step.
 * This file ships with the app; it is not published to a package registry.
 */
export class EdjClient {
  constructor({ baseUrl = '', token = null } = {}) {
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

  submitPayroll(csv, filename = 'payroll.csv') {
    const form = new FormData();
    form.append('csv', new Blob([csv], { type: 'text/csv' }), filename);
    return this.#request('/api/payroll/upload', {
      method: 'POST',
      headers: this.#headers(false),
      body: form
    });
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

export const createClient = (options) => new EdjClient(options);
