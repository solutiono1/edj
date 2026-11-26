// Client-side balance aggregation. Nothing here calls the server with an
// amount: the browser reads UTXOs from the user's own addresses and sums them.
// Pagor cannot see this number unless the user chooses to share it.

export function sumSats(utxos) {
  if (!Array.isArray(utxos)) throw new Error('UTXOs must be an array');
  return utxos.reduce((sum, utxo) => sum + BigInt(utxo.value), 0n);
}

export function formatBch(sats) {
  const value = BigInt(sats);
  const whole = value / 100000000n;
  const fraction = (value % 100000000n).toString().padStart(8, '0').replace(/0+$/, '');
  return fraction ? whole.toString() + '.' + fraction : whole.toString();
}

export function isTokenUtxo(utxo) {
  return Boolean(utxo && (utxo.token || utxo.cashToken));
}

export function aggregateBalance(utxos) {
  const list = Array.isArray(utxos) ? utxos : [];
  const bchUtxos = list.filter((utxo) => !isTokenUtxo(utxo));
  const tokenUtxos = list.filter(isTokenUtxo);
  const sats = sumSats(bchUtxos);
  return {
    chain: 'bch',
    sats,
    display: formatBch(sats),
    utxoCount: bchUtxos.length,
    tokenUtxoCount: tokenUtxos.length,
    source: 'local-wallet'
  };
}
