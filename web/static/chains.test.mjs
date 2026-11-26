import test from 'node:test';
import assert from 'node:assert/strict';
import { CHAINS, chainById, stablesForChain, passesStableTests, isLive } from './chains.js';
import { sumSats, formatBch, aggregateBalance } from './balance.js';
import { PagorWalletAdapter, BchAdapter, adapterFor, supportedChains } from './wallet-adapter.js';

test('chain registry has one live chain and explicit future chains', () => {
  assert.equal(isLive('bch'), true);
  for (const id of ['arc', 'base', 'solana', 'stellar', 'xrpl']) {
    assert.equal(isLive(id), false, id + ' should not be marked live');
  }
  assert.ok(CHAINS.length >= 8);
  assert.equal(chainById('unknown'), null);
});

test('stables per chain are limited and meet the five tests', () => {
  assert.deepEqual(stablesForChain('arc').map((s) => s.symbol), ['USDC']);
  assert.deepEqual(stablesForChain('solana').map((s) => s.symbol), ['USDC', 'PYUSD']);
  assert.equal(passesStableTests('USDC'), true);
  assert.equal(passesStableTests('USDG'), true);
  assert.equal(passesStableTests('PUSD'), false, 'PUSD has limited liquidity');
  assert.equal(passesStableTests('NOPE'), false);
});

test('unsupported adapters fail loudly', async () => {
  const adapter = adapterFor('arc');
  assert.equal(adapter.supported, false);
  assert.ok(adapter instanceof PagorWalletAdapter);
  await assert.rejects(() => adapter.connect(), /not available yet/);
  await assert.rejects(() => adapter.signTransaction(), /not available/);
  await assert.rejects(() => adapter.broadcast(), /not available/);
});

test('BCH adapter is registered', () => {
  assert.ok(adapterFor('bch') instanceof BchAdapter);
  assert.deepEqual(supportedChains(), ['bch']);
});

test('balance is summed locally and formatted', () => {
  const utxos = [
    { value: 150000 },
    { value: 250000 },
    { value: 100000, token: 'abc' }
  ];
  assert.equal(sumSats(utxos), 500000n);
  assert.equal(formatBch(123456789n), '1.23456789');
  assert.equal(formatBch(100000000n), '1');
  const balance = aggregateBalance(utxos);
  assert.equal(balance.sats, 400000n);
  assert.equal(balance.display, '0.004');
  assert.equal(balance.tokenUtxoCount, 1);
  assert.equal(balance.source, 'local-wallet');
});

test('balance rejects malformed input', () => {
  assert.throws(() => sumSats(null), /array/);
});
