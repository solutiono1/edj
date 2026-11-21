(() => {
  'use strict';

  const API = window.EDJ_API_BASE || '';
  const host = document.getElementById('gbtn');
  const status = document.getElementById('gstatus');
  if (!host) return;

  const mode = document.body.dataset.auth || 'login';
  const setStatus = (text, tone) => {
    if (!status) return;
    status.textContent = text;
    status.dataset.tone = tone || '';
  };

  const loadGoogle = () => new Promise((resolve, reject) => {
    if (window.google?.accounts?.id) {
      resolve(window.google);
      return;
    }
    const existing = document.querySelector('script[data-edj-google]');
    if (existing) {
      existing.addEventListener('load', () => resolve(window.google));
      existing.addEventListener('error', () => reject(new Error('Google script failed to load')));
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.defer = true;
    script.dataset.edjGoogle = 'true';
    script.onload = () => resolve(window.google);
    script.onerror = () => reject(new Error('Google script failed to load'));
    document.head.appendChild(script);
  });

  const finish = async (response) => {
    if (!response?.credential) {
      setStatus('Google did not return a credential. Try again.', 'error');
      return;
    }
    setStatus('Verifying your Google account...');
    try {
      const res = await fetch(API+'/api/auth/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ idToken: response.credential })
      });
      const data = await res.json();
      if (!res.ok || !data.token) {
        setStatus(data?.error?.message || 'Sign-in failed. Try again.', 'error');
        return;
      }
      localStorage.setItem('edj_session', data.token);
      if (data.user?.name) localStorage.setItem('edj_name', data.user.name);
      const hasLocalKey = Boolean(localStorage.getItem('edj_key_fp'));
      const destination = mode === 'login' && hasLocalKey ? '/portfolio' : '/onboarding';
      setStatus('Signed in. Opening your account...', 'ok');
      location.href = destination;
    } catch (error) {
      setStatus('Could not reach the server. Check your connection and try again.', 'error');
    }
  };

  const start = async () => {
    setStatus('Loading Google sign-in...');
    let config;
    try {
      const res = await fetch(API+'/api/config', { cache: 'no-store' });
      config = await res.json();
    } catch (error) {
      setStatus('Could not load the sign-in configuration.', 'error');
      return;
    }
    if (!config.googleClientId) {
      setStatus('Google sign-in is not configured yet. Add GOOGLE_CLIENT_ID to the server.', 'error');
      return;
    }
    try {
      await loadGoogle();
      window.google.accounts.id.initialize({
        client_id: config.googleClientId,
        callback: finish,
        auto_select: false,
        cancel_on_tap_outside: true,
        ux_mode: 'popup'
      });
      window.google.accounts.id.renderButton(host, {
        type: 'standard',
        theme: 'outline',
        size: 'large',
        text: 'continue_with',
        shape: 'rectangular',
        logo_alignment: 'left',
        width: 336
      });
      setStatus('');
      window.setTimeout(() => {
        if (!host.querySelector('iframe') && !host.querySelector('button')) {
          setStatus('Google did not render. Make sure this origin is authorized for the client ID.', 'error');
        }
      }, 4000);
    } catch (error) {
      setStatus('Google sign-in could not load. Check your connection and try again.', 'error');
    }
  };

  start();
})();
