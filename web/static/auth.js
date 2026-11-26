(() => {
  'use strict';

  const API = window.EDJ_API_BASE || '';
  const status = document.getElementById('auth-status') || document.getElementById('gstatus');
  const mode = document.body.dataset.auth || 'login';

  const setStatus = (text, tone) => {
    if (!status) return;
    status.textContent = text;
    status.dataset.tone = tone || '';
  };

  const signedIn = (data) => {
    localStorage.setItem('edj_session', data.token);
    if (data.user?.name) localStorage.setItem('edj_name', data.user.name);
    const hasLocalKey = Boolean(localStorage.getItem('edj_key_fp'));
    const destination = mode === 'login' && hasLocalKey ? '/portfolio' : '/onboarding';
    setStatus('Signed in. Opening your account...', 'ok');
    location.href = destination;
  };

  // Email sign-up and login work without Google. No third party required.
  const form = document.getElementById('email-form');
  if (form) {
    form.addEventListener('submit', async (event) => {
      event.preventDefault();
      const email = document.getElementById('email').value.trim();
      const password = document.getElementById('password').value;
      const nameField = document.getElementById('name');
      const button = form.querySelector('button[type="submit"]');
      if (!email || !password) {
        setStatus('Enter your email and password.', 'error');
        return;
      }
      button.disabled = true;
      setStatus(mode === 'signup' ? 'Creating your account...' : 'Signing in...');
      try {
        const body = { email, password };
        if (nameField) body.name = nameField.value.trim();
        const res = await fetch(API + (mode === 'signup' ? '/api/auth/signup' : '/api/auth/login'), {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });
        const data = await res.json();
        if (!res.ok || !data.token) {
          setStatus(data?.error?.message || 'Sign-in failed. Try again.', 'error');
          button.disabled = false;
          return;
        }
        signedIn(data);
      } catch (error) {
        setStatus('Could not reach the server. Check your connection and try again.', 'error');
        button.disabled = false;
      }
    });
  }

  const host = document.getElementById('gbtn');
  if (!host) return;

  const loadGoogle = () => new Promise((resolve, reject) => {
    if (window.google?.accounts?.id) {
      resolve(window.google);
      return;
    }
    const existing = document.querySelector('script[data-pagor-google]');
    if (existing) {
      existing.addEventListener('load', () => resolve(window.google));
      existing.addEventListener('error', () => reject(new Error('Google script failed to load')));
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://accounts.google.com/gsi/client';
    script.async = true;
    script.defer = true;
    script.dataset.pagorGoogle = 'true';
    script.onload = () => resolve(window.google);
    script.onerror = () => reject(new Error('Google script failed to load'));
    document.head.appendChild(script);
  });

  const finishGoogle = async (response) => {
    if (!response?.credential) {
      setStatus('Google did not return a credential. Try again.', 'error');
      return;
    }
    setStatus('Verifying your Google account...');
    try {
      const res = await fetch(API + '/api/auth/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ idToken: response.credential })
      });
      const data = await res.json();
      if (!res.ok || !data.token) {
        setStatus(data?.error?.message || 'Sign-in failed. Try again.', 'error');
        return;
      }
      signedIn(data);
    } catch (error) {
      setStatus('Could not reach the server. Check your connection and try again.', 'error');
    }
  };

  const startGoogle = async () => {
    let config;
    try {
      const res = await fetch(API + '/api/config', { cache: 'no-store' });
      config = await res.json();
    } catch (error) {
      setStatus('Could not load the sign-in configuration.', 'error');
      return;
    }
    if (!config.googleClientId) {
      host.hidden = true;
      const divider = document.querySelector('.divider');
      if (divider) divider.hidden = true;
      return;
    }
    try {
      await loadGoogle();
      window.google.accounts.id.initialize({
        client_id: config.googleClientId,
        callback: finishGoogle,
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
      window.setTimeout(() => {
        if (!host.querySelector('iframe') && !host.querySelector('button')) {
          setStatus('Google did not render. Email sign-in still works.', 'error');
        }
      }, 4000);
    } catch (error) {
      host.hidden = true;
    }
  };

  startGoogle();
})();
