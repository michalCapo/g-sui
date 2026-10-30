// Notifications for App.PWA. Inlined in <head> after pwa.js and called with
// the PWA's defaults: {scope, title, url, icon, badge} and, with PWA.Push,
// the VAPID public key.
//
// <html data-gsui-notify> is the state ShowIfNotifications reads: "off",
// "on", "blocked", "install" (iPhone and iPad allow notifications only in
// the installed app), or absent where the browser has none. With PWA.Push,
// "on" also needs this device's push subscription, which is saved with a
// POST to /__gsui/push.
(function (config) {
  var g = window.__gsui, d = document.documentElement, state = null, busy = false, synced = true, runs = 0, saved = '';
  var saves = Promise.resolve(), steps = Promise.resolve(), lateAfter = false;
  // Time limits are in seconds; tests set a shorter unit.
  var second = config.unit || 1000;
  var push = !!config.key && window.isSecureContext && 'serviceWorker' in navigator && 'PushManager' in window;
  // A key that does not decode would break every check, so push stays off.
  try { if (push) key(); } catch (_) { push = false; console.warn('[g-sui] PWA.Push.PublicKey is not base64url'); }
  function ua() { return navigator.userAgent || ''; }
  // iPadOS reports a Mac user agent; touch support tells them apart.
  function apple() { return /iphone|ipad|ipod/i.test(ua()) || (/macintosh/i.test(ua()) && navigator.maxTouchPoints > 1); }
  function installed() { return (window.matchMedia && matchMedia('(display-mode: standalone)').matches) || navigator.standalone === true; }
  function supported() { return window.isSecureContext && 'Notification' in window; }
  function text() { return (window.__gsuiText || {}).notificationsFailed || 'Notifications could not be changed. Try again.'; }
  function failed() { if (g.notify) g.notify(null, 'error', text()); }
  // within rejects when p takes longer than ms, e.g. a service worker that
  // never gets ready or a stalled push service.
  function within(p, ms) {
    return new Promise(function (resolve, reject) {
      var timer = setTimeout(function () { reject(new Error('timeout')); }, ms);
      Promise.resolve(p).then(function (v) { clearTimeout(timer); resolve(v); }, function (e) { clearTimeout(timer); reject(e); });
    });
  }

  function set(next) {
    if (next === state) { render(); return; }
    state = next;
    if (next) d.setAttribute('data-gsui-notify', next); else d.removeAttribute('data-gsui-notify');
    render();
    try { window.dispatchEvent(new CustomEvent('gsui:notifications', {detail: {state: next}})); } catch (_) {}
  }
  function key() {
    var s = config.key.replace(/-/g, '+').replace(/_/g, '/');
    s += '='.repeat((4 - s.length % 4) % 4);
    return Uint8Array.from(atob(s), function (c) { return c.charCodeAt(0); });
  }
  // A subscription made with another key cannot get this server's pushes.
  function sameKey(sub) {
    var current = sub.options && sub.options.applicationServerKey;
    if (!current) return true;
    var a = new Uint8Array(current), b = key();
    return a.length === b.length && a.every(function (v, i) { return v === b[i]; });
  }
  function worker(r) { try { return new URL(r.active.scriptURL).pathname === '/sw.js'; } catch (_) { return false; } }
  // own is the app's registration of /sw.js, once /sw.js is its active
  // worker; a previous worker at the scope is waited out. Another service
  // worker that covers the page is never used.
  function own(ms) {
    var sw = navigator.serviceWorker;
    return within(sw.register('/sw.js', {scope: config.scope, updateViaCache: 'none'}).then(function (reg) {
      if (worker(reg)) return reg;
      return new Promise(function (resolve) {
        function watch(w) {
          if (w) w.addEventListener('statechange', function () { if (w.state === 'activated' && worker(reg)) resolve(reg); });
        }
        watch(reg.installing || reg.waiting);
        reg.addEventListener('updatefound', function () { watch(reg.installing); });
      });
    }), ms);
  }
  // subscription finds this device's subscription: reg is the app's
  // registration, sub its subscription. Without one there, a subscription
  // with this key in an earlier registration of /sw.js at another scope
  // counts too, so apps that move to PWA.Push keep their devices. Other
  // service workers of the origin are never touched.
  function subscription() {
    var sw = navigator.serviceWorker;
    return own(20 * second).then(function (reg) {
      return reg.pushManager.getSubscription().then(function (sub) {
        if (sub || !sw.getRegistrations) return {reg: reg, sub: sub};
        return sw.getRegistrations().then(function (list) {
          list = list.filter(function (r) { return r.scope !== reg.scope && worker(r); });
          var i = 0;
          function next() {
            if (i >= list.length) return {reg: reg, sub: null};
            return list[i++].pushManager.getSubscription().then(function (old) { return old && sameKey(old) ? {reg: reg, sub: old} : next(); });
          }
          return next();
        });
      });
    });
  }
  // save sends the subscription to PWA.Push.Save and resolves with its
  // outcome within 20 seconds. Each save is its own request, so live
  // navigation and reconnects do not lose the answer. Saves run in turn: the
  // next one waits for the request before it, so a slow save on page load
  // cannot undo a later removal. A save whose caller gave up before its turn
  // is not sent. A save of a run (reconcile) that ends after its caller gave
  // up calls late, so the server follows the browser again; saves of check
  // never do, so slow answers cannot repeat forever. A request still open
  // after a minute is aborted.
  function save(sub, remove, reconcile) {
    var data = sub.toJSON(), expired = false;
    data.remove = remove;
    var sent = saves.then(function () {
      if (expired) return false;
      var abort = window.AbortController ? new AbortController() : null;
      var timer = setTimeout(function () { if (abort) abort.abort(); }, 60 * second);
      var end = function (ok) { clearTimeout(timer); if (expired && reconcile) late(); return ok; };
      return fetch('/__gsui/push', {
        method: 'POST', credentials: 'same-origin', headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data), signal: abort ? abort.signal : undefined
      }).then(function (res) { return end(res.ok); }, function () { return end(false); });
    });
    saves = sent;
    return within(sent, 20 * second).catch(function () { expired = true; return false; });
  }
  // late reads the state again and saves the device if it is on, after a
  // step that ended when nobody waited for it any more. During a run it
  // waits for the run to end.
  function late() {
    if (busy) { lateAfter = true; return; }
    check(true);
  }

  // check reads the state from the browser: the browser's subscription is
  // what counts. With sync, or when its endpoint is not the one saved last
  // (the browser renewed it), it also saves the subscription, so the server
  // follows it and it follows the user who is signed in now. A failed
  // sync, e.g. on a sign-in page, is tried again after the next live
  // navigation.
  function check(sync) {
    if (!supported()) { set(apple() && !installed() && window.isSecureContext ? 'install' : ''); return Promise.resolve(); }
    var permission = Notification.permission;
    if (permission === 'denied') { set('blocked'); return Promise.resolve(); }
    if (permission !== 'granted') { set('off'); return Promise.resolve(); }
    if (!push) { set('on'); return Promise.resolve(); }
    return within(subscription(), 30 * second).then(function (s) {
      var on = !!s.sub && sameKey(s.sub);
      set(on ? 'on' : 'off');
      if (!on) saved = '';
      if (on && (sync || s.sub.endpoint !== saved)) {
        return save(s.sub, false, false).then(function (ok) { synced = ok; saved = ok ? s.sub.endpoint : ''; });
      }
    }, function () { set('off'); });
  }
  function permission() {
    if (Notification.permission !== 'default') return Promise.resolve(Notification.permission);
    // Older Safari takes a callback and returns nothing.
    return new Promise(function (resolve) {
      var p = Notification.requestPermission(resolve);
      if (p && p.then) p.then(resolve, function () { resolve(Notification.permission); });
    });
  }
  // enable and disable take live(), which turns false once their run timed
  // out; they check it before every change, and a step that answers later
  // leaves the result to late.
  function enable(live) {
    return permission().then(function (result) {
      if (result !== 'granted' || !push) return;
      var made = false;
      return subscription().then(function (s) {
        if (!live()) throw new Error('late');
        if (s.sub && sameKey(s.sub)) return s.sub;
        return (s.sub ? s.sub.unsubscribe() : Promise.resolve()).then(function () {
          if (!live()) throw new Error('late');
          made = true;
          return s.reg.pushManager.subscribe({userVisibleOnly: true, applicationServerKey: key()});
        });
      }).then(function (sub) {
        if (!live()) throw new Error('late');
        return save(sub, false, true).then(function (ok) {
          if (ok) { synced = true; saved = sub.endpoint; return; }
          // A subscription the server does not know gets no pushes, so it
          // must not count as on.
          return (made && live() ? sub.unsubscribe() : Promise.resolve()).catch(function () {}).then(function () { throw new Error('save'); });
        });
      });
    });
  }
  // disable removes the subscription on the server first: its endpoint is
  // gone after unsubscribe. When either step fails the run's final check
  // saves the subscription that is left, so the user can retry.
  function disable(live) {
    return subscription().then(function (s) {
      if (!s.sub || !live()) return;
      return save(s.sub, true, true).then(function (ok) {
        if (!ok) throw new Error('save');
        if (!live()) throw new Error('late');
        return s.sub.unsubscribe().then(function (gone) { if (gone === false) throw new Error('unsubscribe'); });
      });
    });
  }
  // run turns notifications on or off. A step that never answers ends the
  // run after a minute and a later run may start. Afterwards the state is
  // read again; after a failed task also saved, so the server follows what
  // the browser kept. A task that ends after its run did calls late.
  function run(task) {
    if (busy || !supported()) return;
    busy = true;
    render();
    var id = ++runs;
    var live = function () { return id === runs; };
    var done = function () {
      busy = false;
      render();
      if (lateAfter) { lateAfter = false; check(true); }
    };
    var ended = false;
    var end = function () { ended = true; if (!live()) late(); };
    var job = steps.then(function () { return task(live); });
    job.then(end, end);
    var work = within(job, 60 * second);
    steps = work.catch(function () {});
    work.then(function () { return false; }, function () {
      if (Notification.permission !== 'denied') failed();
      return true;
    // A task still going will call late when it ends; saving before that
    // could save a subscription it is removing.
    }).then(function (fail) { runs++; return check(fail && ended); }).then(done, done);
  }
  g.notifyEnable = function () { if (state !== 'on') run(enable); };
  // Without PWA.Push only the browser settings turn notifications off.
  g.notifyToggle = function () { if (state !== 'on') run(enable); else if (push) run(disable); };

  g.notification = function (el, n) {
    n = n || {};
    if (typeof n.count === 'number') { badge(n.count); badgeCount = n.count; }
    if (!supported() || Notification.permission !== 'granted') return;
    var url = n.url || config.url;
    // gsui marks notifications whose clicks the g-sui service worker handles.
    var options = {body: n.body || '', tag: n.tag || undefined, icon: n.icon || config.icon || undefined, badge: config.badge || undefined, data: {url: url, gsui: true}};
    var title = n.title || config.title;
    // Phones show notifications only through a service worker; new
    // Notification throws there. It is the fallback where none gets ready.
    if (!navigator.serviceWorker) { page(title, options); return; }
    own(10 * second).then(function (reg) { return reg.showNotification(title, options); }).catch(function () { page(title, options); });
  };
  function page(title, options) {
    try {
      var n = new Notification(title, options);
      n.onclick = function () {
        window.focus();
        n.close();
        var url = new URL(options.data.url, location.href);
        if (url.origin === location.origin && window.__ws) __ws.navigate(url.pathname + url.search + url.hash);
      };
    } catch (_) {}
  }
  function badge(count) {
    if (!('setAppBadge' in navigator)) return;
    (count > 0 ? navigator.setAppBadge(count) : navigator.clearAppBadge()).catch(function () {});
  }

  // render keeps switches and the app icon in step with the page. A newly
  // rendered AppBadge node sets the icon even to the same number, since a
  // push may have changed it meanwhile; a page without one leaves it.
  var badgeNode = null, badgeCount = -1, scheduled = false;
  function render() {
    var on = String(state === 'on'), off = busy || (state !== 'on' && state !== 'off');
    document.querySelectorAll('[data-gsui-notify-switch]').forEach(function (el) {
      if (el.getAttribute('aria-checked') !== on) el.setAttribute('aria-checked', on);
      if (el.disabled !== off) el.disabled = off;
    });
    var node = document.querySelector('[data-gsui-badge]');
    var count = node ? parseInt(node.getAttribute('data-gsui-badge'), 10) || 0 : 0;
    if (node && (node !== badgeNode || count !== badgeCount)) badge(count);
    badgeNode = node;
    badgeCount = node ? count : -1;
  }
  function schedule() {
    if (scheduled) return;
    scheduled = true;
    Promise.resolve().then(function () { scheduled = false; render(); });
  }
  // Morphs drop attributes the server did not render, so the observer also
  // watches the ones render sets.
  if (window.MutationObserver) {
    new window.MutationObserver(schedule).observe(d, {childList: true, subtree: true, attributes: true,
      attributeFilter: ['data-gsui-badge', 'data-gsui-notify-switch', 'aria-checked', 'disabled']});
  }
  document.addEventListener('DOMContentLoaded', render);
  // A navigation, and coming back to the page, read the state again: another
  // tab or the browser settings may have changed it. A navigation also sets
  // the page's count again. A failed sync is retried.
  function recheck() { if (!busy) check(!synced); }
  window.addEventListener('gsui:navigated', function () {
    badgeNode = null;
    render();
    recheck();
  });
  document.addEventListener('visibilitychange', function () { if (document.visibilityState === 'visible') recheck(); });
  if (navigator.permissions && navigator.permissions.query) {
    navigator.permissions.query({name: 'notifications'}).then(function (status) {
      status.onchange = function () { if (!busy) check(false); };
    }, function () {});
  }
  check(true);
})
