import test from 'node:test';
import assert from 'node:assert/strict';
import { PagorClient } from './pagor.js';

const rows = [{ name: 'Ada', address: 'bitcoincash:qaaa', amount: 125000 }];
const passphrase = 'a long sdk passphrase';

test('SDK sends encrypted payroll and keeps amounts off the wire', async () => {
  let sent = null;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async (url, options) => {
    sent = { url, options };
    return {
      ok: true,
      status: 201,
      json: async () => ({ id: 'PP_test', rowCount: rows.length })
    };
  };
  try {
    const client = new PagorClient({ baseUrl: 'http://example.test', token: 'token' });
    const result = await client.createPrivatePayroll(rows, passphrase);
    assert.equal(result.id, 'PP_test');
    assert.equal(sent.url, 'http://example.test/api/private-payroll');
    const payload = JSON.parse(sent.options.body);
    assert.equal(payload.rowCount, 1);
    assert.ok(payload.ciphertext);
    assert.ok(!sent.options.body.includes('Ada'));
    assert.ok(!sent.options.body.includes('125000'));

    const plain = await client.decryptPayroll(payload, passphrase);
    assert.deepEqual(plain, rows);
    await assert.rejects(() => client.decryptPayroll(payload, 'a different long passphrase'));
  } finally {
    globalThis.fetch = originalFetch;
  }
});
