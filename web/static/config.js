// Local dev talks to the Go server directly. Every deployed host (pages.dev or
// a custom domain) talks to the Railway API.
window.EDJ_API_BASE = (location.hostname === 'localhost' || location.hostname === '127.0.0.1')
  ? ''
  : 'https://edj-api-v3-production.up.railway.app';

// In-app signing is off until the signer is verified on chipnet.
window.EDJ_SIGNER_ENABLED = false;
