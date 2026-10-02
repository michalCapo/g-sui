// Local UI steps for Go Actions. Inlined in <head> so mount code can use it
// before /__ws.js loads. Every step receives the element that handled the event.
(function () {
  if (window.__gsui) return;
  function $(id) {
    var e = document.getElementById(id);
    if (!e) { console.warn('[g-sui] element #' + id + ' not found'); if (window.__ws) __ws.notfound(id); }
    return e;
  }
  function cleanup(el, fn) { (el.__gsuiCleanup || (el.__gsuiCleanup = [])).push(fn); }
  function words(s) { return String(s || '').split(/\s+/).filter(Boolean); }
  function hidden(e) { return e.hidden || e.classList.contains('hidden'); }
  function visible(el, e, show) {
    e.hidden = !show; e.classList.toggle('hidden', !show);
    if (!e.id) return;
    if (el && el !== e && el.setAttribute && !e.contains(el)) el.setAttribute('aria-controls', e.id);
    document.querySelectorAll('[aria-controls="' + CSS.escape(e.id) + '"]').forEach(function (t) {
      t.setAttribute('aria-expanded', String(show));
    });
  }
  function classes(id, cls, fn) { var e = $(id); if (e) words(cls).forEach(function (c) { fn(e.classList, c); }); }
  function listen(el, target, event, fn, capture) {
    target.addEventListener(event, fn, capture);
    cleanup(el, function () { target.removeEventListener(event, fn, capture); });
  }
  var g = window.__gsui = {
    widgets: {},
    show: function (el, id) { var e = $(id); if (e) visible(el, e, true); },
    hide: function (el, id) { var e = $(id); if (e) visible(el, e, false); },
    toggle: function (el, id) { var e = $(id); if (e) visible(el, e, hidden(e)); },
    addClass: function (el, id, c) { classes(id, c, function (l, x) { l.add(x); }); },
    removeClass: function (el, id, c) { classes(id, c, function (l, x) { l.remove(x); }); },
    toggleClass: function (el, id, c) { classes(id, c, function (l, x) { l.toggle(x); }); },
    setAttr: function (el, id, k, v) { var e = $(id); if (e) e.setAttribute(k, v); },
    removeAttr: function (el, id, k) { var e = $(id); if (e) e.removeAttribute(k); },
    toggleAttr: function (el, id, k) { var e = $(id); if (e) e.toggleAttribute(k); },
    setText: function (el, id, t) { var e = $(id); if (e) e.textContent = t; },
    setValue: function (el, id, v) {
      var e = $(id); if (!e) return;
      if (e.type === 'checkbox' || e.type === 'radio') e.checked = !!v; else e.value = v;
      e.dispatchEvent(new Event('input', {bubbles: true}));
      e.dispatchEvent(new Event('change', {bubbles: true}));
    },
    remove: function (el, id) { var e = $(id); if (e) { if (window.__gsuiDispose) __gsuiDispose(e); e.remove(); } },
    focus: function (el, id) { var e = $(id); if (e) { e.focus(); if (e.select) e.select(); } },
    scroll: function (el, id) { var e = $(id); if (e) e.scrollIntoView({behavior: 'smooth', block: 'start'}); },
    scrollTop: function () { window.scrollTo({top: 0, behavior: 'smooth'}); },
    open: function (el, id) { var e = $(id); if (!e) return; if (e.showModal) { if (!e.open) e.showModal(); } else visible(el, e, true); },
    close: function (el, id) { var e = $(id); if (!e) return; if (e.close) e.close(); else visible(el, e, false); },
    copy: function (el, text, id) {
      if (id) { var e = $(id); if (!e) return; text = 'value' in e ? e.value : e.textContent; }
      navigator.clipboard.writeText(text).then(function () { g.notify(el, 'success', 'Copied to clipboard'); },
        function () { g.notify(el, 'error', 'Copy failed'); });
    },
    nav: function (el, path, patch, replace) { __ws.navigate(path, {patch: patch, replace: replace}); },
    redirect: function (el, url) {
      document.body.style.visibility = 'hidden';
      setTimeout(function () { window.location.href = url; }, 200);
    },
    reload: function () { location.reload(); },
    back: function () { history.back(); },
    print: function () { window.print(); },
    reset: function (el, id) { var e = $(id); if (e) e.reset(); },
    submit: function (el, id) { var e = $(id); if (e) e.requestSubmit(); },
    password: function (el, id) {
      var e = $(id); if (!e) return;
      e.type = e.type === 'password' ? 'text' : 'password';
      if (el && el !== e && el.setAttribute) el.setAttribute('aria-pressed', String(e.type === 'text'));
    },
    theme: function (el, mode) { if (mode) setTheme(mode); else toggleTheme(); },
    title: function (el, t) { document.title = t; },
    download: function (el, name, mime, data) {
      var a = document.createElement('a'); a.href = 'data:' + mime + ';base64,' + data; a.download = name;
      document.body.appendChild(a); a.click(); a.remove();
    },
    debounce: function (el, ms, fn) {
      clearTimeout(el.__gsuiTimer);
      if (!el.__gsuiTimerCleanup) { el.__gsuiTimerCleanup = true; cleanup(el, function () { clearTimeout(el.__gsuiTimer); }); }
      el.__gsuiTimer = setTimeout(fn, ms);
    },
    // key matches "Enter", "Escape", "ctrl+k", "shift+/", "mod+s" (mod = Ctrl or Cmd).
    key: function (event, spec) {
      var parts = spec.toLowerCase().split('+'), key = parts.pop(), want = {ctrl: false, shift: false, alt: false, meta: false};
      parts.forEach(function (p) { if (p === 'mod') want[/Mac|iPhone|iPad/.test(navigator.platform) ? 'meta' : 'ctrl'] = true; else want[p] = true; });
      var k = String(event.key || '').toLowerCase();
      if (key === 'esc') key = 'escape';
      if (key === 'space') key = ' ';
      return k === key && event.ctrlKey === want.ctrl && event.altKey === want.alt && event.metaKey === want.meta &&
        (key.length !== 1 || /[a-z0-9]/.test(key) ? event.shiftKey === want.shift : true);
    },
    shortcut: function (el, spec, fn) {
      listen(el, document, 'keydown', function (event) {
        if (!g.key(event, spec) || event.defaultPrevented) return;
        var t = event.target, typing = t && (t.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName));
        if (typing && !(event.ctrlKey || event.metaKey || event.altKey) && event.key !== 'Escape') return;
        event.preventDefault(); fn(event, el);
      });
    },
    outside: function (el, fn) {
      listen(el, document, 'click', function (event) {
        var t = event.target;
        if (!el.isConnected || el.contains(t) || hidden(el)) return;
        if (el.id && t.closest && t.closest('[aria-controls="' + CSS.escape(el.id) + '"]')) return;
        fn(event, el);
      }, true);
    },
    drag: function (el) {
      var down = false, sx = 0, sl = 0;
      el.style.cursor = 'grab';
      listen(el, el, 'mousedown', function (e) {
        if (e.target.closest('input,select,button,a,.dt-filter-dropdown')) return;
        down = true; sx = e.pageX; sl = el.scrollLeft; el.style.cursor = 'grabbing'; el.style.userSelect = 'none';
      });
      listen(el, document, 'mouseup', function () {
        if (!down) return; down = false; el.style.cursor = 'grab'; el.style.removeProperty('user-select');
      });
      listen(el, document, 'mousemove', function (e) {
        if (!down) return; e.preventDefault(); el.scrollLeft = sl - (e.pageX - sx);
      });
    },
    active: function (el) {
      var path = new URL(el.getAttribute('href') || '', location.href).pathname.replace(/\/$/, '') || '/';
      var here = location.pathname.replace(/\/$/, '') || '/';
      var on = here === path || (path !== '/' && el.hasAttribute('data-gsui-prefix') && here.indexOf(path + '/') === 0);
      words(el.getAttribute('data-gsui-active')).forEach(function (c) { el.classList.toggle(c, on); });
      words(el.getAttribute('data-gsui-inactive')).forEach(function (c) { el.classList.toggle(c, !on); });
      if (on) el.setAttribute('aria-current', 'page'); else el.removeAttribute('aria-current');
    },
    mount: function (el, name, props) {
      var fn = g.widgets[name];
      if (!fn) { console.warn('[g-sui] widget ' + name + ' is not registered'); return; }
      var stop = fn(el, props);
      if (typeof stop === 'function') cleanup(el, stop);
    },
    notify: function (el, v, message) {
      var box = document.getElementById('__messages__');
      if (!box) {
        box = document.createElement('div'); box.id = '__messages__';
        box.style.cssText = 'position:fixed;top:0;right:0;padding:8px;z-index:9999;pointer-events:none';
        document.body.appendChild(box);
      }
      var n = document.createElement('div'), accent = '#4f46e5', timeout = 5000, dk = document.documentElement.classList.contains('dark');
      n.setAttribute('role', v === 'error' || v === 'error-reload' ? 'alert' : 'status');
      n.style.cssText = 'display:flex;align-items:center;gap:10px;padding:12px 16px;margin:8px;border-radius:12px;min-height:44px;width:calc(100vw - 32px);max-width:380px;box-shadow:0 6px 18px rgba(0,0,0,0.08);border:1px solid;font-weight:600;font-family:inherit;font-size:14px;opacity:0;transform:translateX(20px);transition:opacity 200ms,transform 200ms;pointer-events:auto';
      var colors = dk ? ['#1e1b4b', '#a5b4fc', '#312e81'] : ['#eef2ff', '#3730a3', '#e0e7ff'];
      if (v === 'success') { accent = '#16a34a'; colors = dk ? ['#052e16', '#86efac', '#14532d'] : ['#dcfce7', '#166534', '#bbf7d0']; }
      if (v === 'error' || v === 'error-reload') {
        accent = '#dc2626'; colors = dk ? ['#450a0a', '#fca5a5', '#7f1d1d'] : ['#fee2e2', '#991b1b', '#fecaca'];
        if (v === 'error-reload') timeout = 88000;
      }
      n.style.background = colors[0]; n.style.color = colors[1]; n.style.borderColor = colors[2];
      n.style.borderLeft = '4px solid ' + accent;
      var dot = document.createElement('span');
      dot.style.cssText = 'width:10px;height:10px;border-radius:9999px;flex-shrink:0;background:' + accent;
      var t = document.createElement('span'); t.style.flex = '1'; t.textContent = message;
      var close = document.createElement('button'); close.textContent = '×'; close.setAttribute('aria-label', 'Dismiss notification');
      close.style.cssText = 'border:0;background:transparent;color:inherit;font-size:20px;line-height:1;cursor:pointer;border-radius:6px';
      close.className = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-current';
      close.onclick = function () { n.remove(); };
      n.appendChild(dot); n.appendChild(t); n.appendChild(close);
      if (v === 'error-reload') {
        var btn = document.createElement('button'); btn.textContent = 'Reload';
        btn.style.cssText = 'background:#991b1b;color:#fff;border:none;padding:6px 10px;border-radius:8px;cursor:pointer;font-weight:700;font-size:13px';
        btn.onclick = function () { location.reload(); }; n.appendChild(btn);
      }
      box.appendChild(n);
      requestAnimationFrame(function () { n.style.opacity = '1'; n.style.transform = 'translateX(0)'; });
      setTimeout(function () {
        n.style.opacity = '0'; n.style.transform = 'translateX(20px)';
        setTimeout(function () { n.remove(); }, 200);
      }, timeout);
    }
  };
  window.addEventListener('gsui:navigated', function () {
    document.querySelectorAll('[data-gsui-active]').forEach(g.active);
  });
})();
