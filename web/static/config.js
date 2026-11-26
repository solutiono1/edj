// Local dev uses the Go server directly. Cloudflare Pages uses the Railway API.
window.EDJ_API_BASE = location.hostname.endsWith('.pages.dev')
  ? 'https://edj-api-v3-production.up.railway.app'
  : '';
// In-app signing is off until the signer is verified on chipnet.
window.EDJ_SIGNER_ENABLED = false;
