(() => {
  'use strict';

  const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

  // Session timeout configuration (minutes until session expires)
  // Highly recommend 12 hours = 720 minutes. Minimum 6 hours = 360.
  const SESSION_TIMEOUT_MINUTES = 720;

  // Session-aware: check login status on page load
  const apiBase = window.EDJ_API_BASE || '';
  const session = localStorage.getItem('edj_session');
  
  // If user is signed in, show portfolio section immediately
  if (session) {
    fetch(apiBase + '/api/me', { headers: { Authorization: 'Bearer ' + session } })
      .then((r) => r.json())
      .then((u) => {
        if (u && u.sub) {
          // Valid session: show portfolio section, hide landing
          document.querySelectorAll('[data-view]').forEach(s => s.hidden = s.dataset.view !== '/portfolio');
          document.querySelectorAll('.nav-item').forEach(a => a.classList.add('active'));
          // Update avatar/name
          const ue = document.getElementById('profile-email');
          if (u.email) ue.textContent = u.email;
          const an = document.getElementById('profile-name');
          if (u.name) an.textContent = u.name;
          const av = document.getElementById('avatar');
          if (u.name) av.textContent = u.name.slice(0,2).toUpperCase();
        } else {
          // Invalid session: clear and show landing
          localStorage.removeItem('edj_session');
        }
      })
      .catch(() => {
        localStorage.removeItem('edj_session');
      });
  }

  // "Get started" button handler (for unsigned-in users)
  const gsBtn = document.getElementById('get-started');
  if (gsBtn) {
    gsBtn.addEventListener('click', async () => {
      // Final session check
      const token = localStorage.getItem('edj_session');
      if (!token) {
        // No session: show login state
        document.querySelectorAll('[data-view]').forEach(s => s.hidden = s.dataset.view !== '/login');
        return;
      }
      try {
        const r = await fetch(apiBase + '/api/me', { headers: { Authorization: 'Bearer ' + token } });
        const u = await r.json();
        if (u && u.sub) {
          // Valid session: show portfolio
          document.querySelectorAll('[data-view]').forEach(s => s.hidden = s.dataset.view !== '/portfolio');
          document.querySelectorAll('.nav-item').forEach(a => a.classList.add('active'));
        } else {
          // Invalid session: clear and show login
          localStorage.removeItem('edj_session');
          document.querySelectorAll('[data-view]').forEach(s => s.hidden = s.dataset.view !== '/login');
        }
      } catch (error) {
        localStorage.removeItem('edj_session');
        document.querySelectorAll('[data-view]').forEach(s => s.hidden = s.dataset.view !== '/login');
      }
    });
  }

  // Session timeout manager: clear session after minutes of inactivity
  function startSessionTimeout() {
    let timeoutId = null;
    let lastActivity = Date.now();
    
    function checkTimeout() {
      const elapsed = (Date.now() - lastActivity) / 60000; // minutes
      if (elapsed >= SESSION_TIMEOUT_MINUTES) {
        // Session timed out: clear and redirect to login
        localStorage.removeItem('edj_session');
        localStorage.removeItem('edj_name');
        // Show login view
        document.querySelectorAll('[data-view]').forEach(s => s.hidden = s.dataset.view !== '/login');
      }
    }
    
    function onActivity() {
      lastActivity = Date.now();
      // Restart the timeout check
      clearTimeout(timeoutId);
      timeoutId = setTimeout(checkTimeout, 20000); // check every 20 seconds
    }
    
    // Listen for user activity events
    ['mousemove', 'keydown', 'click', 'scroll', 'touchstart'].forEach(event => {
      document.addEventListener(event, onActivity, { passive: true });
    });
    
    // Initial check
    checkTimeout();
    
    return () => {
      // Cleanup function to remove event listeners
      ['mousemove', 'keydown', 'click', 'scroll', 'touchstart'].forEach(event => {
        document.removeEventListener(event, onActivity, { passive: true });
      });
      clearTimeout(timeoutId);
    };
  }

  // Initialize session timeout when page loads
  const cleanupTimeout = startSessionTimeout();

  // Nav: transparent over the hero at the top, compact glossy island after scrolling.
  // It returns to the transparent state only when the page is back at the top.
  const nav = document.getElementById('nav');
  if (nav) {
    let ticking = false;
    const updateNav = () => {
      nav.classList.toggle('island', window.scrollY > 24);
      ticking = false;
    };
    window.addEventListener('scroll', () => {
      if (!ticking) {
        window.requestAnimationFrame(() => {
          updateNav();
          ticking = true;
        });
      }
    });
  }
})();
