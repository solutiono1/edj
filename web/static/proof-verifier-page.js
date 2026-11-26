// Fetches a public proof and verifies its signature in the browser. The
// verifier's trust decision stays with the verifier.

import { verifyProof, anchorSummary } from './proof-verify.js';

const claimEl = document.getElementById('claim');
if (claimEl) {
  const API = window.EDJ_API_BASE || '';
  const id = decodeURIComponent(location.pathname.split('/').pop() || '');
  const setText = (elementId, text) => {
    const el = document.getElementById(elementId);
    if (el) el.textContent = text;
  };

  fetch(API + '/api/proofs/' + encodeURIComponent(id))
    .then((r) => r.json())
    .then(async (proof) => {
      if (proof.error) {
        claimEl.textContent = 'Statement not found.';
        setText('verify-status', 'No statement to verify.');
        return;
      }
      claimEl.textContent = proof.claim || 'Signed statement';
      setText('subject', proof.subject || '-');
      setText('threshold', String(proof.threshold ?? '-'));
      setText('expires', proof.expiresAt ? new Date(proof.expiresAt).toLocaleString() : '-');
      setText('signature', proof.signature || '-');
      setText('publickey', proof.publicKey || '-');
      setText('fingerprint', proof.signerFingerprint || 'not supplied');
      setText('commitment', proof.anchorCommitment || 'none');
      setText('anchor-note', anchorSummary(proof));

      const valid = await verifyProof(proof);
      setText('verify-status', valid
        ? 'Signature valid for the embedded key.'
        : 'SIGNATURE INVALID. Do not rely on this statement.');
      document.getElementById('verify-status').dataset.tone = valid ? 'ok' : 'error';

      const pinBtn = document.getElementById('pin-key');
      if (pinBtn && proof.signerFingerprint) {
        pinBtn.hidden = false;
        const key = 'pagor_pinned_' + proof.signerFingerprint;
        if (localStorage.getItem(key)) {
          pinBtn.textContent = 'Key already pinned on this device';
          pinBtn.disabled = true;
        }
        pinBtn.onclick = () => {
          localStorage.setItem(key, proof.publicKey);
          pinBtn.textContent = 'Key pinned on this device';
          pinBtn.disabled = true;
          setText('trust-note', 'This key is pinned on this device. Future statements with the same fingerprint can be compared to it.');
        };
      }
    })
    .catch(() => {
      claimEl.textContent = 'The statement could not be loaded. Check your connection and reload.';
    });
}
