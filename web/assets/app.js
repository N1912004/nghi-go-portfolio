(() => {
  'use strict';
  const root = document.documentElement;
  const toggle = document.querySelector('#theme-toggle');
  let savedTheme = null;
  try { savedTheme = localStorage.getItem('nghi-theme'); } catch (_) { /* Private browsing may block storage. */ }
  const theme = savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  function setTheme(value) {
    root.dataset.theme = value;
    toggle.setAttribute('aria-label', `Switch to ${value === 'dark' ? 'light' : 'dark'} theme`);
    const meta = document.querySelector('meta[name="theme-color"]');
    meta.content = value === 'dark' ? '#101b21' : '#f7f8fa';
  }
  setTheme(theme);
  toggle.hidden = false;
  toggle.addEventListener('click', () => {
    const value = root.dataset.theme === 'dark' ? 'light' : 'dark';
    setTheme(value);
    try { localStorage.setItem('nghi-theme', value); } catch (_) {}
  });

  const filters = document.querySelector('#project-filters');
  const cards = [...document.querySelectorAll('.project-card')];
  filters.hidden = false;
  filters.addEventListener('click', event => {
    const button = event.target.closest('button[data-filter]');
    if (!button) return;
    filters.querySelectorAll('button').forEach(item => {
      const active = item === button;
      item.classList.toggle('active', active);
      item.setAttribute('aria-pressed', String(active));
    });
    let count = 0;
    cards.forEach(card => {
      card.hidden = button.dataset.filter !== 'all' && card.dataset.category !== button.dataset.filter;
      if (!card.hidden) count++;
    });
    document.querySelector('#filter-status').textContent = `${count} projects shown`;
  });

  document.querySelector('.api-actions').hidden = false;
  const check = document.querySelector('#check-backend');
  check.addEventListener('click', async () => {
    const status = document.querySelector('#backend-status');
    const output = document.querySelector('#api-output');
    check.disabled = true;
    status.textContent = 'Checking the running backend…';
    const started = performance.now();
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 6000);
    try {
      const response = await fetch('/readyz', { signal: controller.signal, cache: 'no-store' });
      const body = await response.json();
      const elapsed = Math.round(performance.now() - started);
      output.textContent = `GET /readyz → HTTP ${response.status}\n\n${JSON.stringify(body, null, 2)}`;
      status.textContent = response.ok ? `Backend ready · ${elapsed} ms for this request (not a benchmark).` : 'Readiness check failed. Please try again later.';
    } catch (_) {
      output.textContent = 'The backend could not be reached. No response has been simulated.';
      status.textContent = 'Connection unavailable. Please try again.';
    } finally {
      clearTimeout(timeout);
      check.disabled = false;
    }
  });
})();
