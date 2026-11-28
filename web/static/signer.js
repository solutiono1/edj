// Review-first BCH payment path. Signing is gated by PAGOR_SIGNER_ENABLED;
// broadcasting is always a separate, explicit call. Wallets approve the
// signature in their own UI. Nothing is broadcast automatically.

import { planPayment, paymentUri } from './payment-math.js';

const API = () => window.EDJ_API_BASE || '';
let cachedConfig = null;

async function config() {
  if (!cachedConfig) {
    const res = await fetch(API() + '/api/config', { cache: 'no-store' });
    if (!res.ok) throw new Error('Could not load payment configuration');
    cachedConfig = await res.json();
  }
  return cachedConfig;
}

function reverseHex32(hexString) {
  const bytes = String(hexString).replace(/^0x/, '');
  let out = '';
  for (let i = bytes.length - 2; i >= 0; i -= 2) out += bytes.slice(i, i + 2);
  return out;
}

function requireWallet() {
  if (!window.PagorWallet || !window.PagorWallet.connected) {
    throw new Error('Connect a wallet first');
  }
  return { mod: window.PagorWallet.module(), mgr: window.PagorWallet.manager() };
}

async function fetchUtxos(address) {
  const res = await fetch(API() + '/api/wallet/utxos?address=' + encodeURIComponent(address));
  if (!res.ok) throw new Error('Could not load wallet UTXOs');
  const utxos = await res.json();
  if (!Array.isArray(utxos)) throw new Error('Unexpected UTXO response');
  return utxos;
}

function findInputPaths(mod, mgr, address) {
  const paths = {};
  for (let chain = 0; chain <= 1; chain++) {
    for (let index = 0; index < 100; index++) {
      try {
        const pubkey = mgr.getPubkey(chain, BigInt(index));
        if (!pubkey) continue;
        const derived = mod.pubkeyToCashAddress(pubkey);
        if (derived === address) paths[address] = [chain, index];
      } catch (error) {
        // Wallets may not expose every derived key. Skip it.
      }
    }
  }
  return paths;
}

function makeSourceOutputs(mod, plan, address) {
  const lockingBytecode = mod.cashAddressToLockingBytecode(address).bytecode;
  return plan.inputs.map((utxo) => {
    const outpointTransactionHash = mod.hexToBin(reverseHex32(utxo.tx_hash));
    return {
      outpointIndex: utxo.tx_pos,
      outpointTransactionHash,
      sequenceNumber: 0xfffffffe,
      unlockingBytecode: new Uint8Array(),
      valueSatoshis: BigInt(utxo.value),
      lockingBytecode
    };
  });
}

window.PagorSigner = {
  async review({ to, amountSats }) {
    const { address } = requireWallet();
    if (!to) throw new Error('Missing destination address');
    const [settings, utxos] = await Promise.all([
      config(),
      fetchUtxos(address)
    ]);
    const feeRate = settings.feeRate || window.EDJ_FEE_RATE || 1;
    const plan = planPayment(utxos, amountSats, feeRate);
    const bch = (Number(plan.amount) / 100000000).toFixed(8);
    return {
      to,
      address,
      feeRate,
      ...plan,
      summary: `Pay ${bch} BCH to ${to}. Network fee ${plan.fee.toString()} sats. Change ${plan.change.toString()} sats.`
    };
  },

  async build(review) {
    const { mod, mgr } = requireWallet();
    const toLock = mod.cashAddressToLockingBytecode(review.to).bytecode;
    const changeLock = mod.cashAddressToLockingBytecode(review.address).bytecode;
    const sourceOutputs = makeSourceOutputs(mod, review, review.address);
    const inputs = sourceOutputs.map((source, i) => ({
      outpointIndex: source.outpointIndex,
      outpointTransactionHash: source.outpointTransactionHash,
      sequenceNumber: 0xfffffffe,
      unlockingBytecode: new Uint8Array()
    }));
    const paths = findInputPaths(mod, mgr, review.address);
    const path = paths[review.address];
    const inputPaths = inputs.map((_, i) => [
      i,
      path ? (path[0] === 0 ? 'receive' : 'change') : 'receive',
      path ? path[1] : 0
    ]);
    const outputs = [{ lockingBytecode: toLock, valueSatoshis: review.amount }];
    if (review.change > 0n) outputs.push({ lockingBytecode: changeLock, valueSatoshis: review.change });
    const transactionHex = mod.binToHex(mod.encodeTransaction({
      version: 2, inputs, outputs, locktime: 0
    }));
    return { transactionHex, sourceOutputs, inputPaths };
  },

  async sign(review, built) {
    if (window.PAGOR_SIGNER_ENABLED !== true) {
      throw new Error('Signer is not enabled. Use your wallet payment URI instead.');
    }
    const { mgr } = requireWallet();
    const bch = (Number(review.amount) / 100000000).toFixed(8);
    const response = await mgr.signTransaction({
      transaction: {
        transaction: built.transactionHex,
        sourceOutputs: built.sourceOutputs,
        userPrompt: `Review and sign: ${bch} BCH to ${review.to}. No broadcast yet.`,
        broadcast: false
      },
      inputPaths: built.inputPaths
    });
    return response.signedTransaction;
  },

  // Convenience path for callers that want review, build and sign together.
  // It still stops before broadcast.
  async pay({ to, amountSats }) {
    const review = await this.review({ to, amountSats });
    const built = await this.build(review);
    const signedTransaction = await this.sign(review, built);
    return { review, built, signedTransaction };
  },

  async broadcast(rawTx) {
    if (!rawTx) throw new Error('No signed transaction to broadcast');
    const res = await fetch(API() + '/api/wallet/broadcast', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rawTx })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error?.message || 'Broadcast failed');
    return data.txid;
  },

  uri(review) {
    return paymentUri(review.to, review.amount);
  }
};
