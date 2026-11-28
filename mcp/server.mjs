// Pagor MCP server. JSON-RPC over stdio. Exposes scoped payment primitives to
// an agent without giving it custody or keys. A hard amount ceiling is applied
// server-side and cannot be raised by the agent.

const TOOLS = [
  {
    name: 'createLink',
    description: 'Create a private BCH payment link. Amount is in satoshis and is capped by PAGOR_MCP_MAX_AMOUNT.',
    inputSchema: {
      type: 'object',
      properties: {
        amount: { type: 'integer', minimum: 1 },
        note: { type: 'string' },
        token: { type: ['string', 'null'] }
      },
      required: ['amount']
    }
  },
  {
    name: 'submitPayroll',
    description: 'Encrypt a payroll on the device and submit only the ciphertext. Rows are name,address,amount.',
    inputSchema: {
      type: 'object',
      properties: {
        rows: { type: 'array', items: { type: 'object' } },
        passphrase: { type: 'string', minLength: 10 }
      },
      required: ['rows', 'passphrase']
    }
  },
  {
    name: 'createProof',
    description: 'Create a signed threshold statement with an optional on-chain anchor commitment.',
    inputSchema: {
      type: 'object',
      properties: {
        subject: { type: 'string' },
        threshold: { type: 'integer', minimum: 1 },
        expiresInDays: { type: 'integer', minimum: 1 },
        anchor: { type: 'object' }
      },
      required: ['subject', 'threshold']
    }
  },
  {
    name: 'getLink',
    description: 'Read a private payment link by id.',
    inputSchema: { type: 'object', properties: { id: { type: 'string' } }, required: ['id'] }
  },
  {
    name: 'getProof',
    description: 'Read a public signed statement by id. The raw anchor stays hidden until disclosed by its owner.',
    inputSchema: { type: 'object', properties: { id: { type: 'string' } }, required: ['id'] }
  }
];

function cryptoHelpers() {
  const subtle = globalThis.crypto?.subtle;
  if (!subtle) throw new Error('Web Crypto is not available');
  return subtle;
}

function bytesToBase64(bytes) {
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return typeof btoa === 'function' ? btoa(binary) : Buffer.from(bytes).toString('base64');
}

async function encryptRows(rows, passphrase) {
  const subtle = cryptoHelpers();
  if (!Array.isArray(rows) || rows.length === 0) throw new Error('rows must be a non-empty array');
  if (!passphrase || passphrase.length < 10) throw new Error('passphrase must be at least 10 characters');
  const salt = globalThis.crypto.getRandomValues(new Uint8Array(16));
  const iv = globalThis.crypto.getRandomValues(new Uint8Array(12));
  const material = await subtle.importKey('raw', new TextEncoder().encode(passphrase), 'PBKDF2', false, ['deriveKey']);
  const key = await subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 150000, hash: 'SHA-256' },
    material,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt']
  );
  const encrypted = await subtle.encrypt({ name: 'AES-GCM', iv }, key, new TextEncoder().encode(JSON.stringify(rows)));
  return {
    salt: bytesToBase64(salt),
    iv: bytesToBase64(iv),
    ciphertext: bytesToBase64(new Uint8Array(encrypted)),
    rowCount: rows.length
  };
}

export function createHandler(options = {}) {
  const baseUrl = String(options.baseUrl || process.env.PAGOR_API_BASE || 'https://api.usepagor.xyz').replace(/\/$/, '');
  const token = options.token || process.env.PAGOR_API_TOKEN || '';
  const maxAmount = Number(options.maxAmount || process.env.PAGOR_MCP_MAX_AMOUNT || 100000);
  const fetchImpl = options.fetchImpl || globalThis.fetch;

  async function call(path, method, body) {
    const response = await fetchImpl(baseUrl + path, {
      method,
      headers: {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + token
      },
      body: body ? JSON.stringify(body) : undefined
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data?.error?.message || 'API request failed');
    return data;
  }

  async function callTool(name, args = {}) {
    if (name === 'createLink') {
      if (!Number.isInteger(args.amount) || args.amount <= 0) throw new Error('amount must be a positive integer');
      if (args.amount > maxAmount) throw new Error('amount exceeds the agent spending ceiling');
      return call('/api/links', 'POST', { amount: args.amount, note: args.note || '', token: args.token || null });
    }
    if (name === 'submitPayroll') {
      const payload = await encryptRows(args.rows, args.passphrase);
      return call('/api/private-payroll', 'POST', payload);
    }
    if (name === 'createProof') {
      if (!args.subject || !Number.isInteger(args.threshold) || args.threshold <= 0) {
        throw new Error('subject and a positive integer threshold are required');
      }
      return call('/api/disclose', 'POST', {
        subject: args.subject,
        threshold: args.threshold,
        expiresInDays: args.expiresInDays || 30,
        anchor: args.anchor || null
      });
    }
    if (name === 'getLink') {
      if (!args.id) throw new Error('id is required');
      return call('/api/links/' + encodeURIComponent(args.id), 'GET');
    }
    if (name === 'getProof') {
      if (!args.id) throw new Error('id is required');
      return call('/api/proofs/' + encodeURIComponent(args.id), 'GET');
    }
    throw new Error('Unknown tool: ' + name);
  }

  return async function handle(message) {
    const { id, method, params = {} } = message || {};
    if (method === 'initialize') {
      return { jsonrpc: '2.0', id, result: { protocolVersion: '2024-11-05', serverInfo: { name: 'pagor', version: '0.1.0' }, capabilities: { tools: {} } } };
    }
    if (method === 'tools/list') {
      return { jsonrpc: '2.0', id, result: { tools: TOOLS } };
    }
    if (method === 'tools/call') {
      try {
        const result = await callTool(params.name, params.arguments || {});
        return { jsonrpc: '2.0', id, result: { content: [{ type: 'text', text: JSON.stringify(result) }] } };
      } catch (error) {
        return { jsonrpc: '2.0', id, result: { content: [{ type: 'text', text: error.message }], isError: true } };
      }
    }
    return { jsonrpc: '2.0', id, error: { code: -32601, message: 'Method not found' } };
  };
}

if (process.argv[1] && import.meta.url.endsWith(process.argv[1].split('/').pop())) {
  const handle = createHandler();
  process.stdin.setEncoding('utf8');
  let buffer = '';
  process.stdin.on('data', (chunk) => {
    buffer += chunk;
    let index;
    while ((index = buffer.indexOf('\n')) >= 0) {
      const line = buffer.slice(0, index).trim();
      buffer = buffer.slice(index + 1);
      if (!line) continue;
      Promise.resolve(handle(JSON.parse(line)))
        .then((response) => process.stdout.write(JSON.stringify(response) + '\n'))
        .catch((error) => process.stdout.write(JSON.stringify({ jsonrpc: '2.0', error: { code: -32700, message: error.message } }) + '\n'));
    }
  });
}
