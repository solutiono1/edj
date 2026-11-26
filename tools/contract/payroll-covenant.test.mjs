import test from 'node:test';
import assert from 'node:assert/strict';
import { Address, NETWORK } from 'cashaddr';
import { hex } from '@scure/base';
import { buildBatch, parsePkh } from './payroll-covenant.mjs';

const pkhAddress = (pkh) => Address(NETWORK).encode({ format: 'cashaddr', type: 'pkh', hash: hex.decode(pkh) });
const a = pkhAddress('f5bf48b397dae70be82b3cca4793f8eb2b6cdac9');
const b = pkhAddress('a1b2c3d4e5f60718293a4b5c6d7e8f9001a2b3c4');

test('batch covenant is built from two recipients', () => {
  const result = buildBatch([
    { address: a, amount: 25000 },
    { address: b, amount: 15000 }
  ]);
  assert.match(result.address, /^bchtest:/);
  assert.equal(result.count, 2);
  assert.equal(result.total, 40000n);
  assert.equal(result.args.length, 12);
});

test('batch covenant rejects unsupported sizes and bad addresses', () => {
  assert.throws(() => buildBatch([]), /1 to 4/);
  assert.throws(() => buildBatch(Array.from({ length: 5 }, () => ({ address: a, amount: 1 }))), /1 to 4/);
  assert.throws(() => buildBatch([{ address: 'not-an-address', amount: 1 }]), /Invalid cashaddr/);
  assert.throws(() => buildBatch([{ address: a, amount: 0 }]), /positive integer/);
});

test('P2PKH parsing round-trips', () => {
  const pkh = 'f5bf48b397dae70be82b3cca4793f8eb2b6cdac9';
  assert.equal(parsePkh(pkhAddress(pkh)), pkh);
});
