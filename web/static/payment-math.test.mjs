import test from 'node:test';
import assert from 'node:assert/strict';
import { planPayment, selectInputs, estimateVBytes, paymentUri, DUST_LIMIT } from './payment-math.js';

const utxos = [
  { tx_hash: 'aa', tx_pos: 0, value: 50000 },
  { tx_hash: 'bb', tx_pos: 1, value: 100000 }
];

test('coin selection prefers the largest UTXO', () => {
  const { picked, sum } = selectInputs(utxos, 60000);
  assert.equal(picked.length, 1);
  assert.equal(picked[0].tx_hash, 'bb');
  assert.equal(sum, 100000n);
});

test('fee uses P2PKH sizing and the requested rate', () => {
  assert.equal(estimateVBytes(1, 2), 226);
  const plan = planPayment(utxos, 40000, 1);
  assert.equal(plan.fee, 226n);
  const fast = planPayment(utxos, 40000, 2);
  assert.equal(fast.fee, 452n);
});

test('change output is kept when it is above dust', () => {
  const plan = planPayment(utxos, 40000, 1);
  assert.equal(plan.change, 59774n);
  assert.ok(plan.change >= DUST_LIMIT);
  assert.equal(plan.inputs.length, 1);
});

test('dust change is folded into the fee', () => {
  const custom = [{ tx_hash: 'cc', tx_pos: 0, value: 100000 }];
  const amount = 100000 - Number(estimateVBytes(1, 2)) - 500;
  const plan = planPayment(custom, amount, 1);
  assert.equal(plan.change, 0n);
  assert.equal(plan.fee, 726n);
});

test('insufficient funds throws', () => {
  assert.throws(() => planPayment(utxos, 200000, 1), /Insufficient funds/);
});

test('invalid inputs throw', () => {
  assert.throws(() => planPayment([], 1000), /No spendable funds/);
  assert.throws(() => planPayment(utxos, 0), /positive/);
  assert.throws(() => planPayment(utxos, 1000, 0), /positive/);
});

test('payment URI strips the prefix and formats eight decimals', () => {
  assert.equal(
    paymentUri('bitcoincash:qqexample', 25000),
    'bitcoincash:qqexample?amount=0.00025000'
  );
  assert.throws(() => paymentUri('', 1), /Missing receiving address/);
});
