// Local dev uses the Go server directly. Cloudflare Pages uses the Railway API.
window.EDJ_API_BASE = location.hostname.endsWith('.pages.dev')
  ? 'https://edj-api-production.up.railway.app'
  : '';
