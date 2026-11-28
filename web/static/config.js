// Local dev talks to the Go server directly. Every deployed host (pages.dev or
// a custom domain) talks to the Railway API.
window.EDJ_API_BASE = (location.hostname === 'localhost' || location.hostname === '127.0.0.1')
  ? ''
  : 'https://api.usepagor.xyz';

// Default fee rate in satoshis per virtual byte. The server's /api/config
// value takes precedence when it is available.
window.EDJ_FEE_RATE = 1;

// In-app signing is off until the signer is verified on chipnet.
window.PAGOR_SIGNER_ENABLED = false;
