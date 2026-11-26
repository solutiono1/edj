const state = { mod: null, mgr: null, relay: null, connected: false, address: null };

async function ensureModule() {
  if (!state.mod) state.mod = await import('/static/vendor/wizardconnect.js');
  return state.mod;
}

window.PagorWallet = {
  async connect() {
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
        window.dispatchEvent(new CustomEvent('pagor-wallet-ready', {
          detail: { name: msg.wallet_name, address: state.address }
        }));
      });
      state.mgr.on('disconnect', () => {
        state.connected = false;
        state.address = null;
        window.dispatchEvent(new CustomEvent('pagor-wallet-disconnect'));
      });
    }
    state.relay = mod.initiateDappRelay((payload) => state.mgr.updateConnection(payload.client, payload.status));
    state.mgr.attachRelay(state.relay);
    const qr = await mod.renderQR(state.relay.uri);
    return { uri: state.relay.uri, qr };
  },

  get connected() { return state.connected; },
  get address() { return state.address; },
  manager() { return state.mgr; },
  module() { return state.mod; },

  async disconnect() {
    if (state.mgr) {
      try { state.mgr.clearStoredSession(); } catch (error) {}
      try { await state.mgr.sendDisconnect('user closed'); } catch (error) {}
    }
    if (state.relay) {
      try { state.relay.cleanup(); } catch (error) {}
    }
    state.connected = false;
    state.address = null;
  }
};
