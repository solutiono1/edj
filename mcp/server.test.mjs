import test from 'node:test';
import assert from 'node:assert/strict';
import { createHandler } from './server.mjs';

function mockFetch() {
  const calls = [];
  const fetchImpl = async (url, options) => {
    calls.push({ url, options });
    return { ok: true, status: 200, json: async () => ({ id: 'L_test', ok: true }) };
  };
  return { calls, fetchImpl };
}

test('initialize and tools/list expose the payment primitives', async () => {
  const { fetchImpl } = mockFetch();
  const handle = createHandler({ token: 't', fetchImpl });
  const init = await handle({ id: 1, method: 'initialize' });
  assert.equal(init.result.serverInfo.name, 'pagor');
  const list = await handle({ id: 2, method: 'tools/list' });
  const names = list.result.tools.map((tool) => tool.name);
  for (const name of ['createLink', 'submitPayroll', 'createProof', 'getLink', 'getProof']) {
    assert.ok(names.includes(name), name + ' missing');
  }
});

test('agent spending ceiling cannot be raised by arguments', async () => {
  const { fetchImpl } = mockFetch();
  const handle = createHandler({ token: 't', maxAmount: 5000, fetchImpl });
  const response = await handle({
    id: 1,
    method: 'tools/call',
    params: { name: 'createLink', arguments: { amount: 999999, token: null } }
  });
  assert.equal(response.result.isError, true);
  assert.match(response.result.content[0].text, /spending ceiling/);
});

test('createLink sends a capped amount to the API', async () => {
  const { calls, fetchImpl } = mockFetch();
  const handle = createHandler({ token: 't', baseUrl: 'http://example.test', maxAmount: 50000, fetchImpl });
  await handle({ id: 1, method: 'tools/call', params: { name: 'createLink', arguments: { amount: 25000 } } });
  const body = JSON.parse(calls[0].options.body);
  assert.equal(body.amount, 25000);
  assert.match(calls[0].url, /\/api\/links$/);
});

test('submitPayroll sends only ciphertext', async () => {
  const { calls, fetchImpl } = mockFetch();
  const handle = createHandler({ token: 't', baseUrl: 'http://example.test', fetchImpl });
  const rows = [{ name: 'Ada', address: 'bitcoincash:qaaa', amount: 125000 }];
  const response = await handle({
    id: 1,
    method: 'tools/call',
    params: { name: 'submitPayroll', arguments: { rows, passphrase: 'a long mcp passphrase' } }
  });
  assert.equal(response.result.isError, undefined);
  assert.match(calls[0].url, /\/api\/private-payroll$/);
  assert.ok(!calls[0].options.body.includes('Ada'));
  assert.ok(!calls[0].options.body.includes('125000'));
});

test('getProof passes the id through', async () => {
  const { calls, fetchImpl } = mockFetch();
  const handle = createHandler({ token: 't', baseUrl: 'http://example.test', fetchImpl });
  await handle({ id: 1, method: 'tools/call', params: { name: 'getProof', arguments: { id: 'DP_abc' } } });
  assert.match(calls[0].url, /\/api\/proofs\/DP_abc$/);
});
