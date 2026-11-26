// pagor signer: review-first BCH payment via WizardConnect and the Electrum backend.
// The actual signature path is disabled until EDJ_SIGNER_ENABLED is true and it
// has been verified on chipnet. Until then, review() works and pay() refuses.

const API = () => window.EDJ_API_BASE || '';

function reverseHex32(hexString) {
  const bytes = hexString.replace(/^0x/, '');
  let out = '';
  for (let i = bytes.length - 2; i >= 0; i -= 2) {
    out += bytes.slice(i, i + 2);
  }
  return out;
}

function estimateFee(inputCount, outputCount) {
  return inputCount * 148 + outputCount * 34 + 10;
}

async function fetchUtxos(address) {
  const res = await fetch(`${API()}/api/wallet/utxos?address=${encodeURIComponent(address)}`);
  if (!res.ok) throw new Error('Could not load wallet UTXOs');
  return await res.json();
}

function selectInputs(utxos, needed) {
  const sorted = [...utxos].sort((a, b) => b.value - a.value);
  const picked = [];
  let sum = 0n;
  for (const utxo of sorted) {
    picked.push(utxo);
    sum += BigInt(utxo.value);
    if (sum >= needed) break;
  }
  return { picked, sum };
}

window.PagorSigner = {
  async review({ to, amountSats }) {
    if (!window.PagorWallet || !window.PagorWallet.connected) {
      throw new Error('Connect a wallet first');
    }
    const address = window.PagorWallet.address;
    const utxos = await fetchUtxos(address);
    if (!utxos.length) {
      throw new Error('No spendable funds at this address');
    }

    const amount = BigInt(amountSats);
    let fee = 0n;
    let change = 0n;
    let totalInput = 0n;
    let picked = [];

    for (let i = 1; i <= utxos.length; i++) {
      const outputCount = change > 546n ? 2 : 1;
      fee = BigInt(estimateFee(i, outputCount));
      const sel = selectInputs(utxos, amount + fee);
      picked = sel.picked;
      totalInput = sel.sum;
      change = totalInput - amount - fee;
      if (totalInput >= amount + fee) break;
    }

    if (totalInput < amount + fee) {
      throw new Error('Insufficient funds');
    }
    if (change > 0n && change < 546n) {
      fee += change;
      change = 0n;
    }

    return { address, amount, fee, change, totalInput, inputs: picked };
  },

  async pay({ to, amountSats }) {
    if (window.EDJ_SIGNER_ENABLED !== true) {
      throw new Error('Signer is not enabled. Use your wallet payment URI instead.');
    }
    const mod = window.PagorWallet.module();
    const mgr = window.PagorWallet.manager();
    const review = await this.review({ to, amountSats });

    // Map each input address to its derivation path.
    const pathFor = {};
    for (let child = 0; child <= 1; child++) {
      for (let index = 0; index < 100; index++) {
        try {
          const pubkey = mgr.getPubkey(child, BigInt(index));
          if (!pubkey) continue;
          const addr = mod.pubkeyToCashAddress(pubkey);
          pathFor[addr] = [child, index];
        } catch (error) {}
      }
    }

    const toLock = mod.cashAddressToLockingBytecode(to).bytecode;
    const changeLock = mod.cashAddressToLockingBytecode(review.address).bytecode;

    const inputs = [];
    const sourceOutputs = [];
    const inputPaths = [];

    review.inputs.forEach((utxo, i) => {
      const reversed = mod.hexToBin(reverseHex32(utxo.tx_hash));
      const lockingBytecode = mod.cashAddressToLockingBytecode(review.address).bytecode;
      inputs.push({
        outpointIndex: utxo.tx_pos,
        outpointTransactionHash: reversed,
        sequenceNumber: 0xfffffffe,
        unlockingBytecode: new Uint8Array()
      });
      sourceOutputs.push({
        outpointIndex: utxo.tx_pos,
        outpointTransactionHash: reversed,
        sequenceNumber: 0xfffffffe,
        unlockingBytecode: new Uint8Array(),
        valueSatoshis: BigInt(utxo.value),
        lockingBytecode
      });
      const path = pathFor[review.address];
      inputPaths.push([i, path ? (path[0] === 0 ? 'receive' : 'change') : 'receive', path ? path[1] : 0]);
    });

    const outputs = [{ lockingBytecode: toLock, valueSatoshis: review.amount }];
    if (review.change > 0n) {
      outputs.push({ lockingBytecode: changeLock, valueSatoshis: review.change });
    }

    const common = { version: 2, inputs, outputs, locktime: 0 };
    const txHex = mod.binToHex(mod.encodeTransaction(common));

    const response = await mgr.signTransaction({
      transaction: {
        transaction: txHex,
        sourceOutputs,
        userPrompt: `Pay ${review.amount} sats to ${to}`,
        broadcast: false
      },
      inputPaths
    });

    const broadcastRes = await fetch(`${API()}/api/wallet/broadcast`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rawTx: response.signedTransaction })
    });
    const broadcast = await broadcastRes.json();
    if (!broadcastRes.ok) throw new Error(broadcast.error?.message || 'Broadcast failed');
    return broadcast.txid;
  }
};
