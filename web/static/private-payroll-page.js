// Private payroll page flow: CSV is parsed and encrypted on this device. The
// server receives only the ciphertext blob, never names, addresses or amounts.

import { encryptRows, decryptRows, parsePayrollCsv } from './private-payroll.js';

const form = document.getElementById('pay-form');
if (form) {
  const API = window.EDJ_API_BASE || '';
  const token = localStorage.getItem('edj_session');
  const result = document.getElementById('pay-result');
  const summary = document.getElementById('pay-summary');

  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const file = form.elements.namedItem('csv')?.files?.[0];
    const passphrase = document.getElementById('pay-passphrase')?.value || '';
    if (!file) return;
    try {
      const rows = parsePayrollCsv(await file.text());
      const payload = await encryptRows(rows, passphrase);
      const res = await fetch(API + '/api/private-payroll', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error?.message || 'Upload failed');

      // Decrypt locally to show the user what they stored. The server cannot.
      const saved = await fetch(API + '/api/private-payroll/' + data.id, {
        headers: { Authorization: 'Bearer ' + token }
      });
      const blob = await saved.json();
      const plain = await decryptRows(blob, passphrase);
      const total = plain.reduce((sum, row) => sum + row.amount, 0);
      result.style.display = 'block';
      summary.textContent = plain.length + ' recipients · total ' + total + ' sats · encrypted id ' + data.id;
    } catch (error) {
      summary.textContent = error?.message || 'Could not prepare the payroll.';
      result.style.display = 'block';
    }
  });
}
