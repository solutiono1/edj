// Wallet page: one row per chain, with real connect, switch, balance, and
// disconnect actions. Balances are read in the browser and never sent to the
// Pagor API. Chains without an adapter say so instead of pretending to work.

import { CHAINS } from './chains.js';
import { adapterFor, supportedChains } from './wallet-adapter.js';
import { aggregateBalance } from './balance.js';

const list = document.getElementById('chain-list');
if (list) {
  const supported = new Set(supportedChains());
  const adapters = new Map();
  const rows = new Map();
  let renderId = 0;

  for (const chain of CHAINS) adapters.set(chain.id, adapterFor(chain.id));

  const build = () => {
    list.textContent = '';
    for (const chain of CHAINS) {
      const row = document.createElement('div');
      row.className = 'chain-row';
      row.dataset.chain = chain.id;

      const main = document.createElement('div');
      main.className = 'chain-main';
      const name = document.createElement('b');
      name.textContent = chain.name;
      const meta = document.createElement('span');
      meta.className = 'chain-meta';
      meta.textContent = chain.stables.join(', ') + ' · ' + chain.kind.toUpperCase();
      const address = document.createElement('div');
      address.className = 'chain-address';
      const balance = document.createElement('div');
      balance.className = 'chain-balance';
      main.append(name, meta, address, balance);

      const actions = document.createElement('div');
      actions.className = 'chain-actions';
      const status = document.createElement('span');
      status.className = 'chain-status';
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'btn btn-line btn-sm';
      actions.append(status, button);

      row.append(main, actions);
      list.appendChild(row);
      rows.set(chain.id, { row, status, address, balance, button });
    }
  };

  const setButton = (refs, label, handler, disabled = false) => {
    refs.button.textContent = label;
    refs.button.disabled = disabled;
    refs.button.onclick = handler;
  };

  const refreshBalance = async (chain, refs, adapter, id) => {
    try {
      if (chain.id === 'bch') {
        const utxos = await adapter.getUtxos();
        const total = aggregateBalance(utxos);
        if (id !== renderId) return;
        refs.balance.textContent = total.display + ' BCH · ' + total.utxoCount + ' UTXOs';
      } else {
        const total = await adapter.getBalance();
        if (id !== renderId) return;
        const parts = [];
        if (total.stable) parts.push(total.stable.display + ' ' + total.stable.symbol);
        if (total.native.symbol !== (total.stable && total.stable.symbol)) {
          parts.push(total.native.display + ' ' + total.native.symbol);
        }
        refs.balance.textContent = parts.join(' · ');
      }
    } catch (error) {
      if (id !== renderId) return;
      refs.balance.textContent = error?.message || 'Could not read balance.';
    }
  };

  const render = async () => {
    const id = ++renderId;
    for (const chain of CHAINS) {
      const refs = rows.get(chain.id);
      const adapter = adapters.get(chain.id);
      const isSupported = supported.has(chain.id);
      refs.row.dataset.supported = String(isSupported);
      refs.address.textContent = '';
      refs.balance.textContent = '';
      refs.status.textContent = '';

      if (!isSupported) {
        refs.status.textContent = chain.status === 'watchlist' ? 'Watchlist' : 'Planned';
        setButton(refs, 'Coming soon', null, true);
        continue;
      }

      if (!adapter.connected) {
        refs.status.textContent = 'Not connected';
        if (chain.id === 'bch') {
          setButton(refs, 'Connect', () => {
            if (window.PagorConnectBch) {
              window.PagorConnectBch();
            } else {
              adapter.connect().catch((error) => {
                refs.status.textContent = error?.message || 'Could not connect.';
              });
            }
          });
        } else {
          setButton(refs, 'Connect', async () => {
            refs.status.textContent = 'Opening wallet...';
            try {
              await adapter.connect();
              render();
            } catch (error) {
              refs.status.textContent = error?.message || 'Could not connect.';
            }
          });
        }
        continue;
      }

      refs.address.textContent = adapter.address || '';
      const onChain = chain.kind === 'evm' ? adapter.onChain : true;
      if (!onChain) {
        refs.status.textContent = 'Wallet is on another network';
        setButton(refs, 'Switch to ' + chain.name, async () => {
          refs.status.textContent = 'Switching...';
          try {
            await adapter.connect();
            render();
          } catch (error) {
            refs.status.textContent = error?.message || 'Could not switch.';
          }
        });
        continue;
      }

      refs.status.textContent = 'Connected';
      setButton(refs, 'Disconnect', async () => {
        try {
          await adapter.disconnect();
          render();
        } catch (error) {
          refs.status.textContent = error?.message || 'Could not disconnect.';
        }
      });
      refreshBalance(chain, refs, adapter, id);
    }
  };

  build();
  for (const chain of CHAINS) {
    if (!supported.has(chain.id)) continue;
    const adapter = adapters.get(chain.id);
    if (adapter.restore) adapter.restore().catch(() => {});
  }
  window.addEventListener('pagor-wallet-ready', render);
  window.addEventListener('pagor-wallet-disconnect', render);
  window.addEventListener('pagor-evm-ready', render);
  window.addEventListener('pagor-evm-disconnect', render);
  render();
}
