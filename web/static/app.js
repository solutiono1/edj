(() => {
  'use strict';

  const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

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
        ticking = true;
        requestAnimationFrame(updateNav);
      }
    }, { passive: true });
    updateNav();
  }

  // One white dropdown panel whose contents swap per section.
  const navPanel = document.getElementById('nav-panel');
  const navGroups = $$('.nav-group');
  if (navPanel && navGroups.length && nav) {
    let closeTimer;
    let activeGroup = null;

    const positionPanel = (group) => {
      const navRect = nav.getBoundingClientRect();
      const triggerRect = group.getBoundingClientRect();
      const panelWidth = navPanel.offsetWidth || 420;
      const half = panelWidth / 2 + 10;
      let left = triggerRect.left - navRect.left + triggerRect.width / 2;
      left = Math.max(half, Math.min(navRect.width - half, left));
      navPanel.style.setProperty('--panel-left', `${left}px`);
    };

    const openPanel = (group) => {
      clearTimeout(closeTimer);
      activeGroup = group;
      const name = group.dataset.menu;
      $$('.panel-set', navPanel).forEach((set) => {
        set.hidden = set.dataset.set !== name;
      });
      navGroups.forEach((item) => {
        const trigger = $('.nav-trigger', item);
        trigger.setAttribute('aria-expanded', String(item === group));
      });
      navPanel.classList.add('open');
      positionPanel(group);
      requestAnimationFrame(() => positionPanel(group));
    };

    const closePanel = (delay = 140) => {
      clearTimeout(closeTimer);
      closeTimer = setTimeout(() => {
        navPanel.classList.remove('open');
        navGroups.forEach((item) => $('.nav-trigger', item).setAttribute('aria-expanded', 'false'));
        activeGroup = null;
      }, delay);
    };

    navGroups.forEach((group) => {
      const trigger = $('.nav-trigger', group);
      group.addEventListener('mouseenter', () => openPanel(group));
      group.addEventListener('mouseleave', () => closePanel());
      trigger.addEventListener('focus', () => openPanel(group));
      trigger.addEventListener('click', (event) => {
        event.stopPropagation();
        if (activeGroup === group && navPanel.classList.contains('open')) closePanel(0);
        else openPanel(group);
      });
    });

    navPanel.addEventListener('mouseenter', () => clearTimeout(closeTimer));
    navPanel.addEventListener('mouseleave', () => closePanel());

    document.addEventListener('keydown', (event) => {
      if (event.key === 'Escape') closePanel(0);
    });
    document.addEventListener('click', (event) => {
      if (!nav.contains(event.target)) closePanel(0);
    });
    window.addEventListener('resize', () => {
      if (activeGroup) positionPanel(activeGroup);
    });
    window.addEventListener('scroll', () => {
      if (navPanel.classList.contains('open')) closePanel(0);
    }, { passive: true });
  }

  // Reveal sections and cards as they enter view. The class is added by JS so
  // content stays visible if scripting is unavailable.
  const revealEls = $$([
    '.section .eyebrow',
    '.section .h2',
    '.section .sub',
    '.step-card',
    '.photo-card',
    '.dev-card',
    '.dev-grid > div:first-child > *',
    '.cta h2',
    '.cta .btn',
    '.get-copy > *'
  ].join(','));
  if (revealEls.length) {
    if (reduce || !('IntersectionObserver' in window)) {
      revealEls.forEach((el) => el.classList.add('in'));
    } else {
      revealEls.forEach((el, index) => {
        if (el.classList.contains('reveal')) return;
        el.classList.add('reveal');
        const siblings = Array.from(el.parentElement.children).filter((child) => child.classList.contains('reveal'));
        const order = siblings.indexOf(el);
        el.style.setProperty('--d', `${Math.max(0, order) * 55}ms`);
        if (index === 0) el.style.setProperty('--d', '0ms');
      });
      const io = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (!entry.isIntersecting) return;
          entry.target.classList.add('in');
          io.unobserve(entry.target);
        });
      }, { threshold: 0.12, rootMargin: '0px 0px -40px 0px' });
      revealEls.forEach((el) => io.observe(el));
    }
  }

  // Avatar dropdown (logout etc).
  const avatar = document.getElementById('avatar');
  const avatarMenu = document.getElementById('avatar-menu');
  if (avatar && avatarMenu) {
    avatar.addEventListener('click', (event) => {
      event.stopPropagation();
      avatarMenu.hidden = !avatarMenu.hidden;
    });
    document.addEventListener('click', () => {
      avatarMenu.hidden = true;
    });
  }

  // Ed assistant panel toggle.
  const assistantFab = document.getElementById('ed-fab');
  const assistantPanel = document.getElementById('ed-assistant');
  const assistantClose = document.getElementById('ed-close');
  if (assistantFab && assistantPanel) {
    assistantFab.addEventListener('click', () => assistantPanel.classList.toggle('open'));
    if (assistantClose) assistantClose.addEventListener('click', () => assistantPanel.classList.remove('open'));
  }

  // Logout: clears the local session token and returns home.
  const logout = document.getElementById('logout');
  if (logout) logout.addEventListener('click', () => {
    localStorage.removeItem('edj_session');
    location.href = '/';
  });

  // Phone rising into the CTA band.
  const phone = $('.get-phone');
  if (phone) {
    if (reduce || !('IntersectionObserver' in window)) phone.classList.add('in');
    else {
      const phoneObserver = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (!entry.isIntersecting) return;
          entry.target.classList.add('in');
          phoneObserver.unobserve(entry.target);
        });
      }, { threshold: 0.15 });
      phoneObserver.observe(phone);
    }
  }

  // Footer wordmark reveal.
  const wordmark = $('.foot-word');
  if (wordmark) {
    if (reduce || !('IntersectionObserver' in window)) wordmark.classList.add('in');
    else {
      const wordObserver = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (!entry.isIntersecting) return;
          entry.target.classList.add('in');
          wordObserver.unobserve(entry.target);
        });
      }, { threshold: 0.15 });
      wordObserver.observe(wordmark);
    }
  }

  // Public/Private reveal on the landing preview.
  const toggle = document.getElementById('toggle-view');
  const chainView = document.getElementById('chain-view');
  if (toggle && chainView) {
    toggle.addEventListener('click', () => {
      const on = toggle.getAttribute('aria-pressed') === 'true';
      toggle.setAttribute('aria-pressed', String(!on));
      chainView.classList.toggle('blurred', on);
      toggle.textContent = on ? 'Show what the chain sees' : 'Hide what the chain sees';
    });
  }

  // Scroll-driven phone: moves down through the blue section, eases back on scroll up.
  const scrollPhone = $('.get-phone');
  const getSection = $('.get-edj');
  if (scrollPhone && getSection && !reduce) {
    let target = 0;
    let current = 0;
    let running = false;
    const frame = () => {
      current += (target - current) * 0.12;
      if (Math.abs(target - current) < 0.4) {
        current = target;
        running = false;
      }
      scrollPhone.style.setProperty('--py', `${current.toFixed(2)}px`);
      if (running) requestAnimationFrame(frame);
    };
    const measure = () => {
      const top = getSection.offsetTop;
      const height = getSection.offsetHeight;
      const vh = window.innerHeight;
      const start = top - vh;
      const end = top + height - vh;
      const progress = Math.max(0, Math.min(1, (window.scrollY - start) / Math.max(1, end - start)));
      target = progress * Math.max(64, Math.min(150, height * 0.16));
      if (!running) {
        running = true;
        requestAnimationFrame(frame);
      }
    };
    window.addEventListener('scroll', measure, { passive: true });
    window.addEventListener('resize', measure);
    measure();
  }

  // Footer sits behind the page. The section above slides over it on scroll up.
  const footer = $('.footer');
  if (footer && document.body.classList.contains('home')) {
    const fit = () => { document.body.style.paddingBottom = `${footer.offsetHeight}px`; };
    fit();
    window.addEventListener('resize', fit);
    const parts = $$('.foot-top, .foot-bottom, .foot-word', footer);
    if (reduce) {
      parts.forEach((el) => el.classList.add('in'));
      footer.style.setProperty('--fp', '1');
    } else {
      const update = () => {
        const max = document.documentElement.scrollHeight - window.innerHeight;
        const start = max - footer.offsetHeight;
        const p = Math.max(0, Math.min(1, (window.scrollY - start) / Math.max(1, footer.offsetHeight)));
        footer.style.setProperty('--fp', p.toFixed(3));
        if (p > 0.02) parts.forEach((el) => el.classList.add('in'));
      };
      window.addEventListener('scroll', update, { passive: true });
      window.addEventListener('resize', update);
      update();
    }
  }

  // Inner anchors served from their own route.
  if (['/faq', '/support'].includes(location.pathname)) {
    const section = location.pathname === '/faq' ? $('.faq-wrap') : $('.support-section');
    if (section) requestAnimationFrame(() => {
      const y = section.getBoundingClientRect().top + window.scrollY - 90;
      window.scrollTo({ top: y, behavior: 'instant' });
    });
  }

  // Lightbox for the framed screenshots.
  const shotImages = $$('.shot img');
  if (shotImages.length) {
    const lb = document.createElement('div');
    lb.className = 'lightbox';
    lb.innerHTML = '<button class="lb-close" type="button" aria-label="Close">×</button><img alt="">';
    document.body.appendChild(lb);
    const lbImg = lb.querySelector('img');
    const close = () => { lb.classList.remove('open'); document.body.style.overflow = ''; };
    shotImages.forEach((img) => {
      img.addEventListener('click', () => {
        lbImg.src = img.src;
        lb.classList.add('open');
        document.body.style.overflow = 'hidden';
      });
    });
    lb.addEventListener('click', (event) => {
      if (event.target === lb || event.target.classList.contains('lb-close')) close();
    });
    document.addEventListener('keydown', (event) => {
      if (event.key === 'Escape') close();
    });
  }

  // WizardConnect wallet modal.
  const connectWalletBtn = document.getElementById('connect-wallet');
  if (connectWalletBtn && window.EdjWallet) {
    const modal = document.createElement('div');
    modal.className = 'wallet-modal';
    modal.innerHTML = '<div class="wallet-card"><button class="wallet-close" type="button" aria-label="Close">×</button><h3>Connect your wallet</h3><p class="wallet-sub">Scan with a WizardConnect wallet to link your BCH address.</p><img class="wallet-qr" alt="QR code"><p class="wallet-uri"></p></div>';
    document.body.appendChild(modal);
    const qr = modal.querySelector('.wallet-qr');
    const close = modal.querySelector('.wallet-close');
    const open = async () => {
      modal.classList.add('open');
      try {
        const res = await window.EdjWallet.connect();
        qr.src = res.qr;
      } catch (error) {
        modal.classList.remove('open');
      }
    };
    connectWalletBtn.addEventListener('click', open);
    close.addEventListener('click', () => modal.classList.remove('open'));
    modal.addEventListener('click', (event) => {
      if (event.target === modal) modal.classList.remove('open');
    });
    window.addEventListener('edj-wallet-ready', (event) => {
      modal.classList.remove('open');
      const address = event.detail.address || '';
      connectWalletBtn.textContent = address ? address.slice(0, 10) + '...' + address.slice(-6) : 'Wallet connected';
      connectWalletBtn.title = address;
      connectWalletBtn.classList.add('connected');
    });
    window.addEventListener('edj-wallet-disconnect', () => {
      connectWalletBtn.textContent = 'Connect wallet';
      connectWalletBtn.classList.remove('connected');
    });
  }

  // Back to top.
  $$('[data-top]').forEach((el) => {
    el.addEventListener('click', (event) => {
      event.preventDefault();
      window.scrollTo({ top: 0, behavior: reduce ? 'auto' : 'smooth' });
    });
  });

  // Do not autoplay heavy media for people who ask for less motion.
  if (reduce) {
    $$('video[autoplay]').forEach((video) => {
      video.removeAttribute('autoplay');
      video.pause();
    });
  }
})();
