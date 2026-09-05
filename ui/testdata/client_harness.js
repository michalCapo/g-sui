// Harness for the embedded WS client (ui/server.go: wsStubJS + wsClientJS).
// Runs the real client source in a fake DOM against a scriptable WebSocket so
// the reconnect state machine is exercised, not just pattern-matched.
//
// Usage: node client_harness.js <stub.js> <client.js>
const fs = require('fs');
const vm = require('vm');

const stubSrc = fs.readFileSync(process.argv[2], 'utf8');
const clientSrc = fs.readFileSync(process.argv[3], 'utf8');

let failures = 0;
function check(name, cond, extra) {
  if (cond) return;
  failures++;
  console.log('FAIL: ' + name + (extra ? ' -- ' + extra : ''));
}
const sleep = ms => new Promise(r => setTimeout(r, ms));
async function waitFor(cond, ms) {
  const end = Date.now() + ms;
  while (Date.now() < end) { if (cond()) return true; await sleep(10); }
  return false;
}

function makeEl(tag) {
  const el = {
    tagName: (tag || 'div').toUpperCase(), style: {cssText: '', opacity: ''},
    dataset: {}, children: [], parentNode: null, className: '', textContent: '',
    classList: {add() {}, remove() {}, contains: () => false},
    appendChild(c) { c.parentNode = el; el.children.push(c); return c; },
    removeChild(c) { el.children = el.children.filter(x => x !== c); c.parentNode = null; },
    querySelector: () => null, querySelectorAll: () => [],
    addEventListener() {}, getAttribute: () => null, setAttribute() {}, removeAttribute() {},
  };
  return el;
}

function makeEnv() {
  const sockets = [];
  const listeners = {};
  const body = makeEl('body');
  let reloads = 0;

  // Models browser semantics: state transitions are one-way (CONNECTING ->
  // OPEN -> CLOSING -> CLOSED, or CONNECTING -> CLOSED), a closed socket can
  // never reopen, and close events are delivered asynchronously.
  class FakeWebSocket {
    constructor(url) {
      this.url = url; this.readyState = 0; this.sent = [];
      this.onopen = this.onclose = this.onmessage = this.onerror = null;
      sockets.push(this);
    }
    send(m) {
      if (this.readyState === 0) throw new Error('InvalidStateError');
      if (this.readyState !== 1) return;   // CLOSING/CLOSED: silently dropped
      this.sent.push(m);
    }
    close() {
      if (this.readyState === 3 || this.readyState === 2) return;
      this.readyState = 2;
      setTimeout(() => this.drop(), 0);
    }
    // test controls
    accept() {
      if (this.readyState !== 0) throw new Error('cannot open a ' + this.readyState + ' socket');
      this.readyState = 1;
      if (this.onopen) this.onopen({});
    }
    drop() {
      if (this.readyState === 3) return;
      this.readyState = 3;
      if (this.onclose) this.onclose({});
    }
    // CLOSED without a delivered close event: a backgrounded mobile tab
    dieSilently() { this.readyState = 3; }
    deliver(data) { if (this.onmessage) this.onmessage({data}); }
  }

  const documentEl = makeEl('html');
  const document = {
    documentElement: documentEl, body, hidden: false, readyState: 'complete',
    createElement: makeEl,
    getElementById: id => body.children.find(c => c.id === id) || null,
    querySelector: () => null, querySelectorAll: () => [],
    addEventListener(type, fn) { (listeners[type] = listeners[type] || []).push(fn); },
    fonts: null,
  };
  const window = {
    addEventListener(type, fn) { (listeners[type] = listeners[type] || []).push(fn); },
    dispatchEvent(e) { (listeners[e.type] || []).forEach(fn => fn(e)); return true; },
    matchMedia: () => ({matches: false, addEventListener() {}}),
  };
  const location = {protocol: 'http:', host: 'x', pathname: '/', search: '', reload() { reloads++; }};

  const history = {pushed: [], pushState(_state, _title, url) { history.pushed.push(url); location.pathname = url; }};
  const sandbox = {
    window, document, location, history, WebSocket: FakeWebSocket, console,
    setTimeout, clearTimeout, setInterval, clearInterval, JSON, Object, Date, Math,
    requestAnimationFrame: fn => setTimeout(fn, 0),
    CustomEvent: class { constructor(t, o) { this.type = t; this.detail = o && o.detail; } },
    Event: class { constructor(t) { this.type = t; } },
  };
  sandbox.window.window = sandbox.window;
  sandbox.window.history = history;
  vm.createContext(sandbox);
  // Use the sandbox's own Function so that `new Function(js)` in the client
  // (how server responses are executed) sees the fake window/document.
  sandbox.Function = vm.runInContext('Function', sandbox);
  return {
    sandbox, sockets, listeners, body, history,
    latest: () => sockets[sockets.length - 1],
    fire: (type, ev) => (listeners[type] || []).forEach(fn => fn(ev || {})),
    reloads: () => reloads,
    badge: () => body.children.find(c => c.id === '__offline__') || null,
    run(cfg, extra) {
      sandbox.window.__gsuiCfg = cfg;
      vm.runInContext(stubSrc, sandbox);
      if (extra) extra(sandbox);
      vm.runInContext(clientSrc, sandbox);
      // the client assigns `var __ws` in the sandbox global scope
      sandbox.window.__ws = sandbox.__ws;
      return sandbox.__ws;
    },
  };
}

const tests = {
  async repliesOnlyReleaseTheirOwnButton() {
    const env=makeEnv(), ws=env.run({grace:-1,reloadAfter:-1,keepAlive:-1});
    env.latest().accept();
    const a=makeEl('button'),b=makeEl('button');
    const first=ws.call('a',{},null,a), second=ws.call('b',{},null,b);
    env.latest().deliver(JSON.stringify({__r:1,id:first,js:''}));
    check('first button released',a.disabled===false);
    check('second stays disabled',b.disabled===true);
    env.latest().deliver(JSON.stringify({__r:1,id:second,js:''}));
    check('second released on its reply',b.disabled===false);
  },
  async stalePageRepliesAreIgnored() {
    const env=makeEnv(),ws=env.run({grace:-1,reloadAfter:-1,keepAlive:-1});
    env.latest().accept();
    const id=ws.call('old');
    env.sandbox.window.__gsuiVersion=2;
    env.latest().deliver(JSON.stringify({__r:1,id,version:1,js:'window.stalePatch=true'}));
    check('old page reply ignored',!env.sandbox.window.stalePatch);
  },
  async unsubscribeNotifiesServer() {
    const env=makeEnv(),ws=env.run({grace:-1,reloadAfter:-1,keepAlive:-1});
    env.latest().accept();
    ws.subscribe('clock',{id:'a'});ws.unsubscribe('clock',{id:'a'});
    const message=JSON.parse(env.latest().sent.at(-1));
    check('unsubscribe goes over wire',message.act==='__unsubscribe'&&message.data.key==='clock|{"id":"a"}');
  },
  // A socket that died silently must not let its late close event spawn a
  // second live connection.
  async staleSocketDoesNotSpawnParallelConnections() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    await sleep(1100);               // clear the reconnect-attempt cooldown
    a.dieSilently();                 // CLOSED, close event not delivered yet
    env.fire('online');              // triggers reconnectNow -> socket B
    check('B created', env.sockets.length === 2, 'sockets=' + env.sockets.length);
    const b = env.sockets[1];
    b.accept();
    a.drop();                        // stale close event arrives late
    await sleep(60);
    check('stale close does not open a third socket', env.sockets.length === 2,
      'sockets=' + env.sockets.length);
    check('stale close does not mark client offline', ws.connected() === true);
    // stale onerror must close its own socket, not the live one
    if (a.onerror) a.onerror({});
    check('stale error does not close live socket', b.readyState === 1);
  },

  // Repeated foreground/online events on a dead server must not defeat backoff.
  async reconnectStormIsThrottled() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    env.sockets[0].drop();
    const before = env.sockets.length;
    for (let i = 0; i < 20; i++) { env.fire('online'); env.fire('visibilitychange'); }
    check('reconnect storm throttled', env.sockets.length - before <= 1,
      'new sockets=' + (env.sockets.length - before));
  },

  // send() must not lose a message when readyState flipped before onclose.
  async sendOnDeadSocketQueuesInsteadOfThrowing() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    a.dieSilently();
    let threw = false;
    try { ws.call('doThing', {x: 1}); } catch (e) { threw = true; }
    check('call does not throw on a dead socket', !threw);
    a.drop();
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(10);
    check('queued call is delivered after reconnect',
      b.sent.some(m => JSON.parse(m).act === 'doThing'), JSON.stringify(b.sent));
  },

  // Overflowing the queue must not leave orphan inflight ids (stuck loader).
  async queueOverflowDoesNotLeakInflight() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    a.drop();
    for (let i = 0; i < 130; i++) ws.call('a' + i, {});
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(10);
    // reply to everything the client actually sent; loader must then clear
    b.sent.forEach(m => {
      const p = JSON.parse(m);
      if (p.id) b.deliver(JSON.stringify({__r: 1, id: p.id, js: ''}));
    });
    await sleep(220);
    const loader = env.body.children.find(c => c.id === '__ws-loader');
    check('loader cleared after all replies', !loader);
  },

  // Push subscriptions must be re-armed after a silent reconnect.
  async subscriptionsAreReplayedOnReconnect() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    ws.subscribe('clock.start');
    ws.callSilent('analytics.hit');
    check('subscribe sent once', a.sent.filter(m => JSON.parse(m).act === 'clock.start').length === 1);
    a.drop();
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(10);
    const acts = b.sent.map(m => JSON.parse(m).act);
    check('subscription replayed', acts.includes('clock.start'), JSON.stringify(acts));
    check('one-shot callSilent not replayed', !acts.includes('analytics.hit'), JSON.stringify(acts));
  },

  // The reload only fires on a long outage, and never while a hold is active.
  async reloadPolicy() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: 150, keepAlive: -1});
    let a = env.sockets[0];
    a.accept();
    a.drop();
    await sleep(1200);              // outage well past reloadAfter
    env.fire('online');
    env.latest().accept();
    check('long outage reloads', env.reloads() === 1, 'reloads=' + env.reloads());

    const env2 = makeEnv();
    const ws2 = env2.run({grace: -1, reloadAfter: 150, keepAlive: -1});
    env2.sockets[0].accept();
    const release = ws2.hold();
    env2.sockets[0].drop();
    await sleep(1200);
    env2.fire('online');
    env2.sockets[env2.sockets.length - 1].accept();
    check('hold suppresses reload', env2.reloads() === 0, 'reloads=' + env2.reloads());
    release(); release();           // double release must not underflow
    check('hold counter does not underflow', ws2.holds() === 0, 'holds=' + ws2.holds());
    void ws;
  },

  // A hold taken before /__ws.js loads must still suppress the reload.
  async preClientHoldSurvivesHandover() {
    const env = makeEnv();
    let release;
    const ws = env.run({grace: -1, reloadAfter: 150, keepAlive: -1}, sandbox => {
      release = sandbox.window.__ws.hold();          // stub API, before client
      sandbox.window.__ws.subscribe('clock.start');  // queued for replay
    });
    check('hold visible to client', ws.holds() === 1, 'holds=' + ws.holds());
    const a = env.sockets[0];
    a.accept();
    await sleep(5);
    check('queued subscribe replayed to server',
      a.sent.some(m => JSON.parse(m).act === 'clock.start'), JSON.stringify(a.sent));
    a.drop();
    await sleep(1200);
    env.fire('online');
    env.latest().accept();
    check('pre-client hold suppresses reload', env.reloads() === 0, 'reloads=' + env.reloads());
    release();
    check('pre-client release works', ws.holds() === 0, 'holds=' + ws.holds());
  },

  // Blips stay invisible; only outages past the grace window show the badge,
  // and the badge must never block interaction.
  async offlineBadgeGraceAndReuse() {
    const env = makeEnv();
    env.run({grace: 500, reloadAfter: -1, keepAlive: -1});
    const badges = () => env.body.children.filter(c => c.id === '__offline__');
    const a = env.sockets[0];
    a.accept();
    a.drop();
    await sleep(30);
    check('no badge inside grace window', !env.badge());
    // Recovery comes from the normal backoff retry, as in a real blip.
    await waitFor(() => env.latest() !== a, 2000);
    env.latest().accept();
    await sleep(400);
    check('badge never appears for a blip', !env.badge());

    env.latest().drop();
    await sleep(700);
    const badge = env.badge();
    check('badge appears after grace window', !!badge);
    check('badge does not block input', badge && /pointer-events:none/.test(badge.style.cssText),
      badge && badge.style.cssText);
    check('exactly one badge', badges().length === 1, 'count=' + badges().length);

    // Recover, then drop again inside the badge's 150ms fade-out: the element
    // must be reused, never duplicated or orphaned.
    const dead = env.latest();
    await waitFor(() => env.latest() !== dead, 3000);
    env.latest().accept();
    await sleep(5);
    check('badge fading, still attached', badges().length === 1);
    env.latest().drop();
    await sleep(800);
    check('badge visible again after fast recover/drop', !!env.badge());
    check('no orphaned badge left behind', badges().length === 1, 'count=' + badges().length);
    check('reused badge is opaque', env.badge().style.opacity === '1');
  },

  // A subscription registered while disconnected (or before the client loads)
  // must reach the server exactly once, not once per delivery path.
  async subscriptionIsSentExactlyOnce() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    ws.subscribe('clock.start');            // socket still CONNECTING
    env.sockets[0].accept();
    await sleep(10);
    const once = env.sockets[0].sent.filter(m => JSON.parse(m).act === 'clock.start').length;
    check('subscription sent exactly once on first open', once === 1, 'count=' + once);
    // and again after a reconnect
    env.sockets[0].drop();
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(10);
    const again = b.sent.filter(m => JSON.parse(m).act === 'clock.start').length;
    check('subscription sent exactly once per reconnect', again === 1, 'count=' + again);
    // duplicate registration of the same act+data must not double up
    ws.subscribe('clock.start');
    ws.subscribe('clock.start');
    b.drop();
    await sleep(1200);
    env.fire('online');
    const c = env.latest();
    c.accept();
    await sleep(10);
    const dedup = c.sent.filter(m => JSON.parse(m).act === 'clock.start').length;
    check('duplicate subscribe is deduplicated', dedup === 1, 'count=' + dedup);
  },

  // A socket that dies within the reconnect cooldown must still recover
  // without waiting for another outside event.
  async fastDeathStillReconnects() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    a.dieSilently();                 // dead <1s after the attempt
    env.fire('visibilitychange');    // inside cooldown: must schedule, not drop
    await sleep(1400);
    check('a new socket was created without further events', env.sockets.length >= 2,
      'sockets=' + env.sockets.length);
    check('the new socket is a distinct object', env.latest() !== a);
  },

  // Server-driven navigation (Response.Navigate -> pushState) must invalidate
  // the previous page's subscriptions, but keep the new page's.
  async serverNavDropsStaleSubscriptions() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    ws.subscribe('clock.start');     // page /clock
    // Server response: new content subscribes, then pushState + pageChanged
    // (the real ordering produced by Response.Inner(...).Navigate(...)).
    a.deliver(JSON.stringify({__r: 1, id: 1, js:
      "__ws.subscribe('table.live');history.pushState(null,'','/table');if(window.__ws&&__ws.pageChanged)__ws.pageChanged();"}));
    await sleep(10);
    a.drop();
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(10);
    const acts = b.sent.map(m => JSON.parse(m).act);
    check('stale subscription is not replayed', !acts.includes('clock.start'), JSON.stringify(acts));
    check('current page subscription is replayed', acts.includes('table.live'), JSON.stringify(acts));
  },

  // A call made while offline must still get a loader once the socket is back,
  // and that loader must clear when the reply arrives.
  async queuedCallTracksLoader() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    a.drop();
    await sleep(5);
    ws.call('save', {});
    check('no loader while offline', !env.body.children.find(c => c.id === '__ws-loader'));
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(200);
    check('loader appears once reconnected', !!env.body.children.find(c => c.id === '__ws-loader'));
    const sent = b.sent.map(m => JSON.parse(m)).find(m => m.act === 'save');
    check('queued call was delivered', !!sent);
    b.deliver(JSON.stringify({__r: 1, id: sent.id, js: ''}));
    await sleep(220);
    check('loader clears on reply', !env.body.children.find(c => c.id === '__ws-loader'));
  },

  // With a grace window shorter than the badge's 150ms fade-out, a new outage
  // lands while the old badge is still fading. It must be reused, not
  // duplicated, and the cancelled removal must not orphan a node.
  async badgeReuseInsideFadeWindow() {
    const env = makeEnv();
    env.run({grace: 40, reloadAfter: -1, keepAlive: -1});
    const badges = () => env.body.children.filter(c => c.id === '__offline__');
    const a = env.sockets[0];
    a.accept();
    a.drop();
    await sleep(120);
    check('badge shown', badges().length === 1, 'count=' + badges().length);
    await waitFor(() => env.latest() !== a, 3000);
    env.latest().accept();          // hide() starts the 150ms fade
    env.latest().drop();            // new outage immediately, inside the fade
    await sleep(120);               // build() runs at ~40ms, before removal
    check('badge reused inside fade window', badges().length === 1, 'count=' + badges().length);
    await sleep(200);               // the cancelled removal must not fire late
    check('badge survives the cancelled removal', badges().length === 1 && !!env.badge(),
      'count=' + badges().length);
    check('reused badge is visible', env.badge() && env.badge().style.opacity === '1');
  },

  // Keep-alive pings must flow on an open socket and stop on close.
  async keepAlivePings() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: -1, keepAlive: 40});
    const a = env.sockets[0];
    a.accept();
    await sleep(100);
    const pings = a.sent.filter(m => JSON.parse(m).act === '__ping').length;
    check('pings sent while open', pings >= 1, 'pings=' + pings);
    check('ping carries no id', a.sent.filter(m => JSON.parse(m).act === '__ping')
      .every(m => !JSON.parse(m).id));
    a.drop();
    const after = a.sent.length;
    await sleep(120);
    check('pings stop after close', a.sent.length === after);
  },

  // Reconnecting to the same process must never reload: the DOM is still valid.
  async sameServerInstanceDoesNotReload() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: 15000, keepAlive: -1, inst: 'srv-1'});
    const a = env.sockets[0];
    a.accept(); hello(a, 'srv-1');
    a.drop();
    await waitFor(() => env.latest() !== a, 3000);
    const b = env.latest();
    b.accept(); hello(b, 'srv-1');
    await sleep(50);
    check('same instance does not reload', env.reloads() === 0, 'reloads=' + env.reloads());
  },

  // A restarted server generates fresh ui.Target() ids, so the loaded DOM can
  // no longer be patched by it. The client must reload even after a short
  // outage, otherwise every click silently patches nothing.
  async restartedServerReloadsStaleDom() {
    const env = makeEnv();
    let changed = null;
    env.run({grace: -1, reloadAfter: 15000, keepAlive: -1, inst: 'srv-1'});
    env.sandbox.window.addEventListener('gsui:serverchanged', e => { changed = e.detail; });
    const a = env.sockets[0];
    a.accept(); hello(a, 'srv-1');
    a.drop();
    await waitFor(() => env.latest() !== a, 3000);
    const b = env.latest();
    b.accept(); hello(b, 'srv-2');   // different process
    await sleep(50);
    check('restart reloads despite short outage', env.reloads() === 1, 'reloads=' + env.reloads());
    check('gsui:serverchanged carries both ids',
      changed && changed.was === 'srv-1' && changed.now === 'srv-2', JSON.stringify(changed));
  },

  // In-progress work (a big upload) must not be destroyed by the restart
  // reload; it is deferred until the last hold is released.
  async restartReloadWaitsForHolds() {
    const env = makeEnv();
    let pending = null;
    const ws = env.run({grace: -1, reloadAfter: 15000, keepAlive: -1, inst: 'srv-1'});
    env.sandbox.window.addEventListener('gsui:reloadpending', e => { pending = e.detail; });
    const a = env.sockets[0];
    a.accept(); hello(a, 'srv-1');
    const release = ws.hold();
    a.drop();
    await waitFor(() => env.latest() !== a, 3000);
    const b = env.latest();
    b.accept(); hello(b, 'srv-2');
    await sleep(50);
    check('hold defers the restart reload', env.reloads() === 0, 'reloads=' + env.reloads());
    check('gsui:reloadpending fired', pending && pending.reason === 'server-restart',
      JSON.stringify(pending));
    release();
    check('reload runs when the last hold is released', env.reloads() === 1,
      'reloads=' + env.reloads());
    release();
    check('releasing twice does not reload twice', env.reloads() === 1);
  },

  // reloadAfter<0 is a hard opt-out: no reload, but the app still gets told.
  async reloadsDisabledSuppressRestartReload() {
    const env = makeEnv();
    let changed = false;
    env.run({grace: -1, reloadAfter: -1, keepAlive: -1, inst: 'srv-1'});
    env.sandbox.window.addEventListener('gsui:serverchanged', () => { changed = true; });
    const a = env.sockets[0];
    a.accept(); hello(a, 'srv-1');
    a.drop();
    await waitFor(() => env.latest() !== a, 3000);
    const b = env.latest();
    b.accept(); hello(b, 'srv-2');
    await sleep(50);
    check('reloads disabled means no reload', env.reloads() === 0, 'reloads=' + env.reloads());
    check('app is still notified', changed);
  },

  // The hello frame is protocol, not a page update: it must not run as JS nor
  // age out subscriptions registered before it.
  async helloFrameIsNotExecutedAsJs() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1, inst: 'srv-1'});
    const a = env.sockets[0];
    a.accept();
    ws.subscribe('clock.start', {});
    hello(a, 'srv-1');
    await sleep(20);
    ws.pageChanged();                // keeps only subscriptions of this epoch
    a.drop();
    await waitFor(() => env.latest() !== a, 3000);
    const b = env.latest();
    b.accept(); hello(b, 'srv-1');
    await sleep(20);
    const subs = b.sent.filter(m => JSON.parse(m).act === 'clock.start');
    check('subscription survives the hello frame', subs.length === 1, 'sent=' + subs.length);
  },
};

function hello(sock, id) { sock.deliver(JSON.stringify({__hello: id})); }

(async () => {
  for (const [name, fn] of Object.entries(tests)) {
    try { await fn(); } catch (e) { failures++; console.log('ERROR: ' + name + ' -- ' + e.stack); }
  }
  if (failures) { console.log(failures + ' failure(s)'); process.exit(1); }
  console.log('ok - ' + Object.keys(tests).length + ' client scenarios');
})();
