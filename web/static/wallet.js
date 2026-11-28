const state = {
  mod: null, mgr: null, relay: null, connected: false, address: null,
  uri: null, qr: null, explicitDisconnect: false
};

const ADDRESS_KEY = 'pagor_wallet_address';
const CONNECTED_KEY = 'pagor_wallet_connected';

function rememberConnection(address) {
  try {
    localStorage.setItem(CONNECTED_KEY, '1');
    if (address) localStorage.setItem(ADDRESS_KEY, address);
  } catch (error) {
    // Private mode can block localStorage. The in-memory session still works.
  }
}

function forgetConnection() {
  try {
    localStorage.removeItem(CONNECTED_KEY);
    localStorage.removeItem(ADDRESS_KEY);
  } catch (error) {}
}

async function ensureModule() {
  if (!state.mod) state.mod = await import('/static/vendor/wizardconnect.js');
  return state.mod;
}

window.PagorWallet = {
  async connect() {
    if (state.connected && state.uri) return { uri: state.uri, qr: state.qr };
    // Fast path: our own flag is enough to keep the wallet connected across a
    // refresh. Do not wait for the large WizardConnect module to load.
    try {
      if (localStorage.getItem(CONNECTED_KEY) === '1') {
        state.address = localStorage.getItem(ADDRESS_KEY) || state.address;
        state.connected = true;
        window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
          detail: { name: 'wallet', address: state.address, restored: true }
        }));
        return { uri: state.uri, qr: state.qr };
      }
    } catch (error) {}

    const mod = await ensureModule();
    if (!state.mgr) {
      state.mgr = new mod.DappConnectionManager('pagor');
      state.mgr.on('walletready', (msg) => {
        state.connected = true;
        try {
          state.address = mod.pubkeyToCashAddress(state.mgr.getPubkey(0, 0n));
        } catch (error) {
          state.address = null;
        }
        rememberConnection(state.address);
        window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
          detail: { name: msg.wallet_name, address: state.address }
        }));
      });
      state.mgr.on('disconnect', () => {
        if (!state.explicitDisconnect) return;
        state.connected = false;
        state.address = null;
        forgetConnection();
        state.uri = null;
        state.qr = null;
        window.dispatchEvent(new CustomEvent('pagor-wallet-disconnect'));
      });
    }
    state.relay = mod.initiateDappRelay((payload) => state.mgr.updateConnection(payload.client, payload.status));
    state.mgr.attachRelay(state.relay);
    const qr = await mod.renderQR(state.relay.uri);
    state.uri = state.relay.uri;
    state.qr = qr;
    return { uri: state.uri, qr };
  },

  // Restore a previously linked wallet from WizardConnect's local session.
  // This runs before any QR scan and makes an existing connection visible
  // immediately, instead of waiting for the wallet relay handshake.
  async restore() {
    // Fast path: our own flag is enough to keep the wallet connected across a
    // refresh. Do not wait for the large WizardConnect module to load.
    try {
      if (localStorage.getItem(CONNECTED_KEY) === '1') {
        state.address = localStorage.getItem(ADDRESS_KEY) || state.address;
        state.connected = true;
        window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
          detail: { name: 'wallet', address: state.address, restored: true }
        }));
        return true;
      }
    } catch (error) {}

    const mod = await ensureModule();
    if (!state.mgr) {
      state.mgr = new mod.DappConnectionManager('pagor');
      state.mgr.on('walletready', (msg) => {
        state.connected = true;
        try {
          state.address = mod.pubkeyToCashAddress(state.mgr.getPubkey(0, 0n));
        } catch (error) {
          state.address = null;
        }
        rememberConnection(state.address);
        window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
          detail: { name: msg.wallet_name, address: state.address }
        }));
      });
      state.mgr.on('disconnect', () => {
        if (!state.explicitDisconnect) return;
        state.connected = false;
        state.address = null;
        forgetConnection();
        state.uri = null;
        state.qr = null;
        window.dispatchEvent(new CustomEvent('pagor-wallet-disconnect'));
      });
    }
    const stored = state.mgr.loadStoredSession?.();
    if (!stored?.paths?.length) {
      // WizardConnect sometimes loses its own session on reload. Pagor keeps
      // its own flag so a wallet the user linked stays linked across refreshes.
      try {
        if (localStorage.getItem(CONNECTED_KEY) !== '1') return false;
      } catch (error) {
        return false;
      }
      state.address = localStorage.getItem(ADDRESS_KEY) || state.address;
      state.connected = true;
      window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
        detail: { name: 'wallet', address: state.address, restored: true }
      }));
      return true;
    }
    try {
      state.address = mod.pubkeyToCashAddress(state.mgr.getPubkey(0, 0n));
      state.connected = true;
      rememberConnection(state.address);
      window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
        detail: { name: stored.walletName || 'wallet', address: state.address, restored: true }
      }));
      return true;
    } catch (error) {
      try {
        if (localStorage.getItem(CONNECTED_KEY) !== '1') return false;
      } catch (inner) {
        return false;
      }
      state.connected = true;
      state.address = localStorage.getItem(ADDRESS_KEY) || state.address;
      window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
        detail: { name: 'wallet', address: state.address, restored: true }
      }));
      return true;
    }
  },

  get connected() { return state.connected; },
  get address() { return state.address; },
  manager() { return state.mgr; },
  module() { return state.mod; },

  async disconnect() {
    state.explicitDisconnect = true;
    forgetConnection();
    if (state.mgr) {
      try { state.mgr.clearStoredSession(); } catch (error) {}
      try { await state.mgr.sendDisconnect('user closed'); } catch (error) {}
    }
    if (state.relay) {
      try { state.relay.cleanup(); } catch (error) {}
    }
    state.connected = false;
    state.address = null;
    state.uri = null;
    state.qr = null;
  }
};
