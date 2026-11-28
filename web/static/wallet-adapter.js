// Pagor wallet adapter. One interface for every chain. BCH is the only live
// implementation; other chains fail loudly instead of pretending to work.
// Pagor never holds keys or funds. Importing an existing key is a future,
// explicit user choice and is not part of this adapter.

import { chainById, isLive } from './chains.js';
import { EvmWallet, EVM_CHAINS } from './evm-wallet.js';

const API = () => (typeof window !== 'undefined' && window.EDJ_API_BASE) || '';

export class PagorWalletAdapter {
  constructor(chainId) {
    const chain = chainById(chainId);
    if (!chain) throw new Error('Unknown chain: ' + chainId);
    this.chainId = chainId;
    this.chain = chain;
  }

  get supported() {
    return isLive(this.chainId) && this.chainId === 'bch';
  }

  async connect() {
    throw new Error('Wallet adapter for ' + this.chain.name + ' is not available yet');
  }

  async disconnect() {
    throw new Error('Wallet adapter for ' + this.chain.name + ' is not available yet');
  }

  get address() {
    return null;
  }

  get connected() {
    return false;
  }

  async getUtxos() {
    throw new Error('UTXO reads are not available on ' + this.chain.name);
  }

  async signTransaction() {
    throw new Error('Signing is not available on ' + this.chain.name);
  }

  async broadcast() {
    throw new Error('Broadcast is not available on ' + this.chain.name);
  }
}

export class BchAdapter extends PagorWalletAdapter {
  constructor() {
    super('bch');
  }

  get supported() {
    return true;
  }

  get wallet() {
    if (typeof window === 'undefined' || !window.PagorWallet) {
      throw new Error('BCH wallet module is not loaded');
    }
    return window.PagorWallet;
  }

  async connect() {
    return this.wallet.connect();
  }

  async restore() {
    if (!this.wallet.restore) return false;
    return this.wallet.restore();
  }

  async disconnect() {
    return this.wallet.disconnect();
  }

  get address() {
    try { return this.wallet.address; } catch (error) { return null; }
  }

  get connected() {
    try { return this.wallet.connected; } catch (error) { return false; }
  }

  async getUtxos() {
    if (!this.address) throw new Error('Connect a wallet first');
    const res = await fetch(API() + '/api/wallet/utxos?address=' + encodeURIComponent(this.address));
    if (!res.ok) throw new Error('Could not load wallet UTXOs');
    const data = await res.json();
    if (!Array.isArray(data)) throw new Error('Unexpected UTXO response');
    return data;
  }

  async signTransaction(review, built) {
    if (!window.PagorSigner) throw new Error('Signer module is not loaded');
    return window.PagorSigner.sign(review, built);
  }

  async broadcast(rawTx) {
    if (!window.PagorSigner) throw new Error('Signer module is not loaded');
    return window.PagorSigner.broadcast(rawTx);
  }
}

export class EvmAdapter extends PagorWalletAdapter {
  constructor(chainId) {
    super(chainId);
    this.evm = new EvmWallet(chainId);
  }

  get supported() {
    return Boolean(EVM_CHAINS[this.chainId]);
  }

  async connect() {
    return this.evm.connect();
  }

  async restore() {
    return this.evm.restore();
  }

  async disconnect() {
    return this.evm.disconnect();
  }

  get address() {
    return this.evm.address;
  }

  get connected() {
    return this.evm.connected;
  }

  get currentChainId() {
    return this.evm.currentChainId;
  }

  get onChain() {
    return this.evm.onChain;
  }

  async getBalance() {
    return this.evm.getBalance();
  }

  async signMessage(message) {
    return this.evm.signMessage(message);
  }

  async signTransaction(tx) {
    return this.evm.sendTransaction(tx);
  }

  async broadcast() {
    throw new Error('EVM wallets broadcast when they sign. Call signTransaction with explicit user approval.');
  }
}

const AVAILABLE = Object.freeze({ bch: BchAdapter, arc: EvmAdapter, base: EvmAdapter });

export function adapterFor(chainId) {
  const Adapter = AVAILABLE[chainId];
  return Adapter ? new Adapter(chainId) : new PagorWalletAdapter(chainId);
}

export function supportedChains() {
  return Object.keys(AVAILABLE);
}
