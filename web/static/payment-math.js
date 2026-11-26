// Pure payment planning helpers. No DOM, no network, no wallet dependency.
// Kept in its own module so the same logic can be tested in Node and used in
// the browser. P2PKH-only sizing; CashToken inputs are not supported here.

export const DUST_LIMIT = 546n;

export function estimateVBytes(inputCount, outputCount) {
  return inputCount * 148 + outputCount * 34 + 10;
}

export function selectInputs(utxos, needed) {
  const sorted = [...utxos].sort((a, b) => {
    const av = BigInt(a.value);
    const bv = BigInt(b.value);
    return av < bv ? 1 : av > bv ? -1 : 0;
  });
  const picked = [];
  let sum = 0n;
  for (const utxo of sorted) {
    picked.push(utxo);
    sum += BigInt(utxo.value);
    if (sum >= needed) break;
  }
  return { picked, sum };
}

export function planPayment(utxos, amountSats, feeRate = 1) {
  const amount = BigInt(amountSats);
  const rate = BigInt(feeRate);
  if (amount <= 0n) throw new Error('Amount must be positive');
  if (rate <= 0n) throw new Error('Fee rate must be positive');
  if (!Array.isArray(utxos) || utxos.length === 0) throw new Error('No spendable funds');

  for (let count = 1; count <= utxos.length; count++) {
    const vbytes = estimateVBytes(count, 2);
    const fee = BigInt(vbytes) * rate;
    const { picked, sum } = selectInputs(utxos, amount + fee);
    if (picked.length !== count || sum < amount + fee) continue;
    const change = sum - amount - fee;
    if (change >= DUST_LIMIT) {
      return { inputs: picked, amount, fee, change, totalInput: sum, vbytes };
    }
    if (change >= 0n) {
      // Dust change would be uneconomical. Give it to the miner rather than
      // creating an output that nodes may reject.
      return { inputs: picked, amount, fee: fee + change, change: 0n, totalInput: sum, vbytes };
    }
  }
  throw new Error('Insufficient funds for amount and fee');
}

export function paymentUri(address, amountSats) {
  const clean = String(address || '').replace(/^bitcoincash:/, '');
  if (!clean) throw new Error('Missing receiving address');
  const bch = (Number(amountSats) / 100000000).toFixed(8);
  return 'bitcoincash:' + clean + '?amount=' + bch;
}
