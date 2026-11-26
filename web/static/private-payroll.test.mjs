import test from 'node:test';
import assert from 'node:assert/strict';
import { encryptRows, decryptRows, parsePayrollCsv } from './private-payroll.js';

const rows = [
  { name: 'Ada', address: 'bitcoincash:qaaa', amount: 125000 },
  { name: 'Grace', address: 'bitcoincash:qbbb', amount: 98000 }
];
const passphrase = 'a long payroll passphrase';

test('payroll rows round-trip through encryption', async () => {
  const payload = await encryptRows(rows, passphrase);
  assert.equal(payload.rowCount, 2);
  assert.ok(payload.ciphertext.length > 0);
  assert.ok(!payload.ciphertext.includes('Ada'));
  assert.ok(!payload.ciphertext.includes('125000'));
  const plain = await decryptRows(payload, passphrase);
  assert.deepEqual(plain, rows);
});

test('wrong passphrase fails', async () => {
  const payload = await encryptRows(rows, passphrase);
  await assert.rejects(() => decryptRows(payload, 'a different long passphrase'));
});

test('short passphrase is rejected', async () => {
  await assert.rejects(() => encryptRows(rows, 'short'), /at least 10/);
});

test('CSV is parsed locally with strict header and amounts', () => {
  const csv = 'name,address,amount\nAda,bitcoincash:qaaa,125000\nGrace,bitcoincash:qbbb,98000';
  assert.deepEqual(parsePayrollCsv(csv), rows);
  assert.throws(() => parsePayrollCsv('name,addr,amount\nAda,x,1'), /Expected header/);
  assert.throws(() => parsePayrollCsv('name,address,amount\nAda,x,0'), /positive integer/);
});
