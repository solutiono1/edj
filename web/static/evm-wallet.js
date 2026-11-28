// EVM wallet adapter for injected wallets (MetaMask, Rabby, Coinbase Wallet,
// and any other EIP-1193 provider). Pagor never holds keys or funds. The
// wallet prompts for every signature and every transaction.

const ADDRESS_KEY = 'pagor_evm_address';
const CHAIN_KEY = 'pagor_evm_chain';

export const EVM_CHAINS = Object.freeze({
  arc: {
    id: 'arc',
    name: 'Arc',
    chainId: 5042,
    hexChainId: '0x13b2',
    rpcUrl: 'https://rpc.mainnet.arc.io',
    explorer: 'https://explorer.arc.io',
    nativeCurrency: { name: 'USDC', symbol: 'USDC', decimals: 18 },
    stable: { symbol: 'USDC', decimals: 18, native: true }
  },
  base: {
    id: 'base',
    name: 'Base',
    chainId: 8453,
    hexChainId: '0x2105',
    rpcUrl: 'https://mainnet.base.org',
    explorer: 'https://basescan.org',
    nativeCurrency: { name: 'Ether', symbol: 'ETH', decimals: 18 },
    stable: {
      symbol: 'USDC',
      decimals: 6,
      address: '0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913'
    }
  }
});

export function getInjectedProvider() {
  if (typeof window === 'undefined') return null;
  return window.ethereum || null;
}

export function formatUnits(value, decimals) {
  const raw = BigInt(value);
  const base = 10n ** BigInt(decimals);
  const whole = raw / base;
  const fraction = (raw % base).toString().padStart(decimals, '0').replace(/0+$/, '');
  return fraction ? whole.toString() + '.' + fraction : whole.toString();
}

function hexToBigInt(hex) {
  return BigInt(hex);
}

function balanceOfData(address) {
  const clean = address.toLowerCase().replace(/^0x/, '');
  return '0x70a08231' + clean.padStart(64, '0');
}

function utf8Hex(text) {
  return '0x' + Array.from(new TextEncoder().encode(text))
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('');
}

function remember(address, chain) {
  try {
    localStorage.setItem(ADDRESS_KEY, address);
    localStorage.setItem(CHAIN_KEY, chain);
  } catch (error) {
    // Private mode can block storage. The in-memory session still works.
  }
}

function forget() {
  try {
    localStorage.removeItem(ADDRESS_KEY);
    localStorage.removeItem(CHAIN_KEY);
  } catch (error) {}
}

export class EvmWallet {
  constructor(chainId) {
    const config = EVM_CHAINS[chainId];
    if (!config) throw new Error('Unknown EVM chain: ' + chainId);
    this.config = config;
    this.address = null;
    this.connected = false;
    this.currentChainId = null;
    this.provider = null;
    this.bound = false;
  }

  get available() {
    return Boolean(getInjectedProvider());
  }

  get onChain() {
    return this.connected && this.currentChainId === this.config.chainId;
  }

  _provider() {
    const provider = this.provider || getInjectedProvider();
    if (!provider) {
      throw new Error('No EVM wallet found. Install MetaMask, Rabby, or another browser wallet.');
    }
    this.provider = provider;
    return provider;
  }

  async _request(method, params) {
    return this._provider().request({ method, params });
  }

  _bind() {
    const provider = this.provider;
    if (this.bound || !provider || typeof provider.on !== 'function') return;
    this.bound = true;
    provider.on('accountsChanged', (accounts) => {
      if (!Array.isArray(accounts) || accounts.length === 0) {
        this.address = null;
        this.connected = false;
        forget();
        this._emit('pagor-evm-disconnect');
        return;
      }
      this.address = accounts[0];
      this.connected = true;
      remember(this.address, this.config.id);
      this._emit('pagor-evm-ready');
    });
    provider.on('chainChanged', (hex) => {
      this.currentChainId = Number.parseInt(hex, 16);
      this._emit('pagor-evm-ready');
    });
  }

  _emit(name, extra) {
    if (typeof window === 'undefined') return;
    window.dispatchEvent(new CustomEvent(name, {
      detail: {
        chain: this.config.id,
        address: this.address,
        chainId: this.currentChainId,
        ...extra
      }
    }));
  }

  async _chainId() {
    const hex = await this._request('eth_chainId');
    return Number.parseInt(hex, 16);
  }

  async _ensureChain() {
    const current = await this._chainId();
    if (current === this.config.chainId) return;
    try {
      await this._request('wallet_switchEthereumChain', [{ chainId: this.config.hexChainId }]);
    } catch (error) {
      const code = error && error.code;
      const message = String((error && error.message) || '');
      if (code === 4902 || /unrecognized chain|not added|unknown chain/i.test(message)) {
        await this._request('wallet_addEthereumChain', [{
          chainId: this.config.hexChainId,
          chainName: this.config.name,
          nativeCurrency: this.config.nativeCurrency,
          rpcUrls: [this.config.rpcUrl],
          blockExplorerUrls: [this.config.explorer]
        }]);
        await this._request('wallet_switchEthereumChain', [{ chainId: this.config.hexChainId }]);
      } else {
        throw error;
      }
    }
  }

  async connect() {
    const accounts = await this._request('eth_requestAccounts');
    if (!Array.isArray(accounts) || accounts.length === 0) {
      throw new Error('Your wallet returned no account.');
    }
    await this._ensureChain();
    this.address = accounts[0];
    this.connected = true;
    this.currentChainId = await this._chainId();
    remember(this.address, this.config.id);
    this._bind();
    this._emit('pagor-evm-ready');
    return { address: this.address, chainId: this.currentChainId, chain: this.config };
  }

  async restore() {
    if (typeof window === 'undefined') return false;
    const provider = getInjectedProvider();
    if (!provider) return false;
    this.provider = provider;
    let accounts;
    try {
      accounts = await provider.request({ method: 'eth_accounts' });
    } catch (error) {
      return false;
    }
    if (!Array.isArray(accounts) || accounts.length === 0) return false;

    let storedChain = null;
    let storedAddress = null;
    try {
      storedChain = localStorage.getItem(CHAIN_KEY);
      storedAddress = localStorage.getItem(ADDRESS_KEY);
    } catch (error) {}

    this.currentChainId = await this._chainId();
    const matches = this.currentChainId === this.config.chainId;
    if (storedChain && storedChain !== this.config.id && !matches) return false;
    if (!storedChain && !matches) return false;

    this.address = storedAddress && accounts.includes(storedAddress) ? storedAddress : accounts[0];
    this.connected = true;
    this._bind();
    this._emit('pagor-evm-ready', { restored: true, matches });
    return { address: this.address, chainId: this.currentChainId, matches };
  }

  async disconnect() {
    try {
      await this._request('wallet_revokePermissions', [{ eth_accounts: {} }]);
    } catch (error) {
      // Not every wallet supports revoking permissions. Pagor still forgets
      // the link locally, and the wallet can revoke it from its own settings.
    }
    this.address = null;
    this.connected = false;
    this.currentChainId = null;
    forget();
    this._emit('pagor-evm-disconnect');
  }

  async getBalance() {
    if (!this.connected || !this.address) throw new Error('Connect a wallet first');
    if (!this.onChain) throw new Error('Switch your wallet to ' + this.config.name + ' first');

    const nativeHex = await this._request('eth_getBalance', [this.address, 'latest']);
    const nativeValue = hexToBigInt(nativeHex);
    const native = {
      symbol: this.config.nativeCurrency.symbol,
      decimals: this.config.nativeCurrency.decimals,
      value: nativeValue,
      display: formatUnits(nativeValue, this.config.nativeCurrency.decimals)
    };

    let stable = null;
    if (this.config.stable.native) {
      stable = {
        symbol: this.config.stable.symbol,
        decimals: this.config.stable.decimals,
        value: nativeValue,
        display: native.display
      };
    } else if (this.config.stable.address) {
      try {
        const result = await this._request('eth_call', [{
          to: this.config.stable.address,
          data: balanceOfData(this.address)
        }, 'latest']);
        const value = hexToBigInt(result);
        stable = {
          symbol: this.config.stable.symbol,
          decimals: this.config.stable.decimals,
          value,
          display: formatUnits(value, this.config.stable.decimals)
        };
      } catch (error) {
        stable = null;
      }
    }

    return { chain: this.config.id, address: this.address, native, stable };
  }

  async signMessage(message) {
    if (!this.connected || !this.address) throw new Error('Connect a wallet first');
    return this._request('personal_sign', [utf8Hex(String(message)), this.address]);
  }

  async sendTransaction(tx) {
    if (!this.connected || !this.address) throw new Error('Connect a wallet first');
    await this._ensureChain();
    return this._request('eth_sendTransaction', [{ from: this.address, ...tx }]);
  }
}
