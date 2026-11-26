// Portfolio balance is computed in the browser from the user's own UTXOs.
// The amount is never sent to the Pagor API.

import { adapterFor } from './wallet-adapter.js';
import { aggregateBalance } from './balance.js';

const balanceEl = document.getElementById('wallet-balance');
const connectBtn = document.getElementById('connect-balance');
if (balanceEl) {
  const adapter = adapterFor('bch');

  const render = async () => {
    if (!adapter.connected) {
      balanceEl.textContent = 'Connect your BCH wallet to see your balance.';
      return;
    }
    try {
      const utxos = await adapter.getUtxos();
      const balance = aggregateBalance(utxos);
      balanceEl.textContent = balance.display + ' BCH (' + balance.utxoCount + ' UTXOs)';
    } catch (error) {
      balanceEl.textContent = error?.message || 'Could not read your wallet.';
    }
  };

  if (connectBtn) {
    connectBtn.addEventListener('click', async () => {
      try {
        const { uri } = await adapter.connect();
        balanceEl.textContent = 'Scan this in your wallet: ' + uri;
      } catch (error) {
        balanceEl.textContent = error?.message || 'Could not start wallet connection.';
      }
    });
  }

  window.addEventListener('pagor-wallet-ready', render);
  window.addEventListener('pagor-wallet-disconnect', render);
  render();
}
