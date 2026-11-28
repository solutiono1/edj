// Chain and stablecoin registry. A chain is only listed when it is a real
// payment rail. Status is explicit so the UI never implies a chain works
// before an adapter exists.

export const CHAINS = Object.freeze([
  { id: 'bch', name: 'Bitcoin Cash', kind: 'utxo', status: 'live', stables: ['PUSD'] },
  { id: 'arc', name: 'Arc', kind: 'evm', status: 'live', stables: ['USDC'] },
  { id: 'base', name: 'Base', kind: 'evm', status: 'live', stables: ['USDC'] },
  { id: 'solana', name: 'Solana', kind: 'svm', status: 'planned', stables: ['USDC', 'PYUSD'] },
  { id: 'stellar', name: 'Stellar', kind: 'account', status: 'planned', stables: ['USDC'] },
  { id: 'xrpl', name: 'XRP Ledger', kind: 'account', status: 'planned', stables: ['RLUSD'] },
  { id: 'robinhood', name: 'Robinhood Chain', kind: 'evm', status: 'watchlist', stables: ['USDG', 'RLUSD'] },
  { id: 'tempo', name: 'Tempo', kind: 'evm', status: 'watchlist', stables: ['USDC'] }
]);

export const STABLECOINS = Object.freeze({
  PUSD: { issuer: 'BCH ecosystem', redeemable: true, liquidity: 'limited', reserves: 'issuer reported', depeg: 'none observed', status: 'available' },
  USDC: { issuer: 'Circle', redeemable: true, liquidity: 'deep', reserves: 'attested', depeg: 'none sustained', status: 'available' },
  USDG: { issuer: 'Paxos', redeemable: true, liquidity: 'deep', reserves: 'attested', depeg: 'none sustained', status: 'available' },
  PYUSD: { issuer: 'Paxos', redeemable: true, liquidity: 'moderate', reserves: 'attested', depeg: 'none sustained', status: 'available' },
  RLUSD: { issuer: 'Ripple', redeemable: true, liquidity: 'growing', reserves: 'attested', depeg: 'none sustained', status: 'available' }
});

export function chainById(id) {
  return CHAINS.find((chain) => chain.id === id) || null;
}

export function stablesForChain(id) {
  const chain = chainById(id);
  return chain ? chain.stables.map((symbol) => ({ symbol, ...STABLECOINS[symbol] })) : [];
}

// Five tests from the product decision. A stablecoin is usable only if all pass.
export function passesStableTests(symbol) {
  const coin = STABLECOINS[symbol];
  if (!coin || coin.status !== 'available') return false;
  return coin.redeemable === true
    && coin.liquidity !== 'limited'
    && coin.depeg === 'none sustained'
    && typeof coin.reserves === 'string' && coin.reserves.length > 0;
}

export function isLive(chainId) {
  const chain = chainById(chainId);
  return Boolean(chain && chain.status === 'live');
}
