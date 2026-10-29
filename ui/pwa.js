// Install support and service worker registration for App.PWA. Inlined in
// <head> after local.js and called with the service worker scope.
//
// <html data-gsui-install> is "prompt" when the browser offered its install
// prompt, "manual" where the user installs from a browser menu (iPhone, iPad,
// macOS Safari 17+), and absent when the app runs installed or cannot be
// installed. A CSS rule shows ShowIfInstallable nodes only while it is set, so
// nodes added later need no observer.
(function (scope) {
  var g = window.__gsui, d = document.documentElement, saved = null, installed = false, state = '';
  var standalone = window.matchMedia && matchMedia('(display-mode: standalone)');
  function ua() { return navigator.userAgent || ''; }
  // iPadOS reports a Mac user agent; touch support tells them apart.
  function apple() { return /iphone|ipad|ipod/i.test(ua()) || (/macintosh/i.test(ua()) && navigator.maxTouchPoints > 1); }
  // Safari 17+ on macOS installs web apps with File > Add to Dock.
  function macSafari() {
    var m = ua().match(/version\/(\d+)[\d.]* safari/i);
    return /macintosh/i.test(ua()) && !/chrome|chromium|edg|firefox|opr/i.test(ua()) && !!m && +m[1] >= 17;
  }
  function sync() {
    var next = '';
    if (!installed && window.isSecureContext && !(standalone && standalone.matches) && navigator.standalone !== true) {
      next = saved ? 'prompt' : apple() || macSafari() ? 'manual' : '';
    }
    if (next === state) return;
    state = next;
    if (next) d.setAttribute('data-gsui-install', next); else d.removeAttribute('data-gsui-install');
    try { window.dispatchEvent(new CustomEvent('gsui:install', {detail: {state: next}})); } catch (_) {}
  }
  // install shows the browser's prompt when there is one, else the app's
  // dialog with install steps.
  g.install = function (el, id) {
    var p = saved;
    function help() { if (id) g.open(el, id); }
    if (!p) { help(); return; }
    saved = null;
    try {
      Promise.resolve(p.prompt()).then(function () { return p.userChoice; }).then(function (choice) {
        if (choice && choice.outcome === 'accepted') installed = true;
        sync();
      }, help);
    } catch (_) { help(); }
    sync();
  };
  window.addEventListener('beforeinstallprompt', function (event) { event.preventDefault(); saved = event; sync(); });
  window.addEventListener('appinstalled', function () { saved = null; installed = true; sync(); });
  if (standalone && standalone.addEventListener) standalone.addEventListener('change', sync);
  sync();
  if ('serviceWorker' in navigator && window.isSecureContext) {
    window.addEventListener('load', function () {
      navigator.serviceWorker.register('/sw.js', {scope: scope, updateViaCache: 'none'}).catch(function () {});
    });
  }
})
