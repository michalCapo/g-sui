// Harness for the embedded WS client (ui/server.go: wsStubJS + wsClientJS).
// Runs the real client source in a fake DOM against a scriptable WebSocket so
// the reconnect state machine is exercised, not just pattern-matched.
//
// Usage: node client_harness.js <stub.js> <client.js> <cfg.js>
// cfg.js is the __gsuiCfg script of an app with navigation skeletons.
const fs = require('fs');
const vm = require('vm');

const stubSrc = fs.readFileSync(process.argv[2], 'utf8');
const clientSrc = fs.readFileSync(process.argv[3], 'utf8');
const cfgSrc = fs.readFileSync(process.argv[4], 'utf8');

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
  const attrs = {}, classes = new Set();
  const el = {
    tagName: (tag || 'div').toUpperCase(), style: {cssText: '', opacity: '', display: ''},
    dataset: {}, children: [], parentNode: null, className: '', textContent: '', id: '',
    get firstChild() { return el.children[0] || null; },
    classList: {
      add(...names) { names.forEach(name => classes.add(name)); },
      remove(...names) { names.forEach(name => classes.delete(name)); },
      contains: name => classes.has(name),
    },
    appendChild(c) { c.parentNode = el; el.children.push(c); return c; },
    insertBefore(c, ref) {
      c.parentNode = el; const i = el.children.indexOf(ref);
      if (i < 0) el.children.push(c); else el.children.splice(i, 0, c);
      return c;
    },
    hasAttribute: name => name in attrs, focus() {},
    removeChild(c) { el.children = el.children.filter(x => x !== c); c.parentNode = null; },
    remove() { if (el.parentNode) el.parentNode.removeChild(el); },
    querySelector: () => null, querySelectorAll: () => [],
    addEventListener() {}, getAttribute: name => name in attrs ? attrs[name] : null,
    setAttribute(name, value) { attrs[name] = value; }, removeAttribute(name) { delete attrs[name]; },
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
    getElementById: id => {
      const find = el => el.children.reduce((hit, c) => hit || (c.id === id ? c : find(c)), null);
      return find(body);
    },
    querySelector: () => null, querySelectorAll: () => [],
    addEventListener(type, fn) { (listeners[type] = listeners[type] || []).push(fn); },
    fonts: null,
  };
  const window = {
    addEventListener(type, fn) { (listeners[type] = listeners[type] || []).push(fn); },
    dispatchEvent(e) { (listeners[e.type] || []).forEach(fn => fn(e)); return true; },
    matchMedia: () => ({matches: false, addEventListener() {}}),
    scrollTo() {}, scrollX: 0, scrollY: 0,
  };
  const location = {
    protocol: 'http:', host: 'x', origin: 'http://x', pathname: '/', search: '', hash: '', reload() { reloads++; },
    get href() { return 'http://x' + location.pathname + location.search + location.hash; },
    assign(url) { location.assigned = url; },
  };

  const history = {
    pushed: [], state: null,
    pushState(state, _title, url) { history.pushed.push(url); history.state = state; location.pathname = url; },
    replaceState(state) { history.state = state; },
  };
  const sandbox = {
    window, document, location, history, WebSocket: FakeWebSocket, console, URL,
    setTimeout, clearTimeout, setInterval, clearInterval, JSON, Object, Date, Math,
    requestAnimationFrame: fn => setTimeout(fn, 0),
    CustomEvent: class { constructor(t, o) { this.type = t; this.detail = o && o.detail; } },
    Event: class { constructor(t) { this.type = t; } },
  };
  // As in a browser, window is the global object.
  Object.assign(sandbox, window);
  sandbox.window = sandbox;
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
      sandbox.window.__gsuiVersion = 1;
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
  async plainReconnectKeepsLiveSocket() {
    const env = makeEnv();
    const ws = env.run({grace: 0, reloadAfter: 0, keepAlive: -1});
    const a = env.latest();
    ws.reconnect();
    check('plain reconnect keeps CONNECTING socket', env.sockets.length === 1 && a.readyState === 0);
    a.accept();
    ws.reconnect();
    check('plain reconnect keeps OPEN socket', env.sockets.length === 1 && a.readyState === 1 && ws.connected());
    await sleep(20);
    check('plain reconnect never schedules an outage for a live socket', !env.badge() && env.reloads() === 0);
  },

  async restartReplacesOpenSocketWithoutChangingPage() {
    const env = makeEnv();
    const ws = env.run({grace: 10, reloadAfter: 10, keepAlive: -1, inst: 'srv-1'});
    const a = env.latest();
    a.accept(); hello(a, 'srv-1');
    const draft = makeEl('textarea');
    draft.id = 'draft'; draft.value = 'Unsaved compose text';
    env.body.appendChild(draft);
    const button = makeEl('button'); button.form = makeEl('form');
    ws.subscribe('clock.start', {id: 'compose'});
    ws.call('save', {}, null, button);
    await sleep(150); // let the pending call's loader appear
    check('pending call is busy with a loader', button.disabled && button.form.getAttribute('aria-busy') &&
      env.body.children.some(c => c.id === '__ws-loader'));
    let reconnected = 0, resolved = false;
    env.sandbox.window.addEventListener('gsui:reconnected', () => reconnected++);
    const pending = ws.restart().then(() => { resolved = true; });
    const b = env.latest();
    check('restart immediately closes OPEN socket and creates a fresh one',
      a.readyState === 2 && b !== a && b.readyState === 0 && env.sockets.length === 2);
    check('restart waits for the replacement socket to open', !resolved && !ws.connected());
    check('restart releases button, form and busy classes', button.disabled === false &&
      !button.getAttribute('aria-busy') && !button.form.getAttribute('aria-busy') && !button.classList.contains('gsui-busy'));
    await sleep(180); // beyond both the outage threshold and loader fade
    check('restart does not flash offline or leave a loader', !env.badge() && !ws.offline() &&
      !env.body.children.some(c => c.id === '__ws-loader'));
    b.accept();
    await pending;
    hello(b, 'srv-2'); // even an instance change must not reload this restart
    const sent = b.sent.map(m => JSON.parse(m));
    check('restart resolves on open', resolved && ws.connected());
    check('restart re-arms managed subscriptions exactly once',
      sent.length === 1 && sent[0].act === 'clock.start' && sent[0].sub === 'clock.start|{"id":"compose"}');
    check('restart never replays in-flight calls', !sent.some(m => m.act === 'save'));
    check('restart preserves the same DOM node and form value', env.sandbox.document.getElementById('draft') === draft &&
      draft.value === 'Unsaved compose text');
    check('restart does not navigate, reload, change page version or emit an outage event',
      env.history.pushed.length === 0 && env.reloads() === 0 && env.sandbox.window.__gsuiVersion === 1 && reconnected === 0);

    const currentID = ws.call('current', {}, null, button);
    a.onclose({}); a.onerror({}); a.onopen({});
    a.deliver(JSON.stringify({__r: 1, id: currentID, version: 1, js: 'window.staleRestartPatch=true'}));
    check('retired socket handlers cannot release new calls, patch the DOM or close the new socket',
      button.disabled === true && !env.sandbox.window.staleRestartPatch && ws.connected() && b.readyState === 1);
    b.deliver(JSON.stringify({__r: 1, id: currentID, version: 1, js: ''}));
    await sleep(600); // a stale close must not schedule another connection
    check('retired socket does not create parallel connections', env.sockets.length === 2 && ws.connected());
  },

  async restartReplacesConnectingSocketAndSupersedesEarlierRestart() {
    const env = makeEnv();
    const ws = env.run({grace: 0, reloadAfter: 0, keepAlive: -1});
    const a = env.latest();
    const first = ws.restart().then(() => null, error => error);
    const b = env.latest();
    check('restart replaces a CONNECTING socket immediately', a.readyState === 2 && b !== a);
    const second = ws.restart();
    const c = env.latest();
    const error = await first;
    check('superseded restart rejects its own Promise', error && /superseded/.test(error.message));
    b.onopen({}); b.onclose({}); b.onerror({});
    c.accept();
    await second;
    await sleep(20);
    check('latest restart owns the connection', b.readyState !== 1 && c.readyState === 1 && ws.connected() &&
      env.sockets.length === 3 && !env.badge() && env.reloads() === 0);
  },

  async restartDoesNotDisableFutureServerRestartDetection() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: 15000, keepAlive: -1, inst: 'srv-1'});
    env.latest().accept();
    const pending = ws.restart();
    const b = env.latest();
    b.accept(); hello(b, 'srv-2');
    await pending;
    check('forced restart preserves page on server change', env.reloads() === 0);
    b.drop();
    await waitFor(() => env.latest() !== b, 2000);
    env.latest().accept(); hello(env.latest(), 'srv-2');
    check('later ordinary reconnect still detects a changed server', env.reloads() === 1);
  },

  async restartRejectsOnHandshakeFailure() {
    for (const fail of ['error', 'close', 'constructor']) {
      const env = makeEnv();
      const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
      env.latest().accept();
      if (fail === 'constructor') env.sandbox.WebSocket = class { constructor() { throw new Error('handshake construction failed'); } };
      const pending = ws.restart().then(() => null, error => error);
      if (fail === 'error') env.latest().onerror({});
      if (fail === 'close') env.latest().drop();
      const error = await pending;
      check('restart rejects on ' + fail, error && /failed/.test(error.message));
      // Restore the constructor and recover using the normal retry path.
      env.sandbox.WebSocket = env.sockets[0].constructor;
      const failed = env.latest();
      await waitFor(() => env.latest() !== failed, 2000);
      env.latest().accept();
      check('normal recovery remains available after ' + fail, ws.connected());
    }
  },

  async restartRejectsOnTimeoutAndRetiresHungSocket() {
    const env = makeEnv();
    // Advance only the restart deadline; do not wait ten seconds in the suite.
    env.sandbox.setTimeout = (fn, ms) => setTimeout(fn, ms === 10000 ? 25 : ms);
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    env.latest().accept();
    const pending = ws.restart().then(() => null, error => error);
    const hung = env.latest();
    const error = await pending;
    check('restart rejects after its timeout', error && /timed out/.test(error.message));
    check('timeout closes the hung attempt', hung.readyState !== 0 && !ws.connected());
    hung.onopen({});
    check('timed-out attempt cannot become ready', !ws.connected());
  },

  async preClientRestartPromiseSurvivesHandover() {
    for (const fail of [false, true]) {
      const env = makeEnv();
      let pending, settled = false;
      const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1}, sandbox => {
        sandbox.window.__ws.subscribe('clock.start');
        pending = sandbox.window.__ws.restart().then(() => { settled = true; return null; }, error => { settled = true; return error; });
      });
      check('stub restart creates a Promise and waits for open', pending && !settled && env.sockets.length === 2);
      if (fail) env.latest().drop(); else env.latest().accept();
      const error = await pending;
      check('stub Promise forwards ' + (fail ? 'failure' : 'success'), settled && (fail ? !!error : error === null && ws.connected()));
      if (!fail) check('stub subscription survives restart handover',
        env.latest().sent.filter(m => JSON.parse(m).act === 'clock.start').length === 1);
    }
  },

  async repliesOnlyReleaseTheirOwnButton() {
    const env=makeEnv(), ws=env.run({grace:-1,reloadAfter:-1,keepAlive:-1});
    env.latest().accept();
    const a=makeEl('button'),b=makeEl('button');
    const first=ws.call('a',{},null,a), second=ws.call('b',{},null,b);
    env.latest().deliver(JSON.stringify({__r:1,id:first,version:1,js:''}));
    check('first button released',a.disabled===false);
    check('second stays disabled',b.disabled===true);
    env.latest().deliver(JSON.stringify({__r:1,id:second,version:1,js:''}));
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

  // A dead socket must reject mutations without throwing or replaying them.
  async sendOnDeadSocketRejectsWithoutThrowing() {
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
    check('offline call is not replayed after reconnect',
      !b.sent.some(m => JSON.parse(m).act === 'doThing'), JSON.stringify(b.sent));
  },

  // Rejected offline calls must not leave orphan inflight ids (stuck loader).
  async offlineCallsDoNotLeakInflight() {
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
      if (p.id) b.deliver(JSON.stringify({__r: 1, id: p.id, version:1, js: ''}));
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
    ws.call('analytics.hit');
    check('subscribe sent once', a.sent.filter(m => JSON.parse(m).act === 'clock.start').length === 1);
    a.drop();
    await sleep(1200);
    env.fire('online');
    const b = env.latest();
    b.accept();
    await sleep(10);
    const acts = b.sent.map(m => JSON.parse(m).act);
    check('subscription replayed', acts.includes('clock.start'), JSON.stringify(acts));
    check('one-shot action not replayed', !acts.includes('analytics.hit'), JSON.stringify(acts));
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

  // Built-in navigation must invalidate
  // the previous page's subscriptions, but keep the new page's.
  async serverNavDropsStaleSubscriptions() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept();
    ws.subscribe('clock.start');     // page /clock
    // Navigation disposes old subscriptions before rendering the new page.
    a.deliver(JSON.stringify({__r: 1, id: 1, version:1, js:
      "__ws.beginNavigation(false);__ws.subscribe('table.live');"}));
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

  // Offline calls are rejected; reconnect must not send them or show a loader.
  async offlineCallIsNotReplayed() {
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
    check('no loader after reconnect', !env.body.children.find(c => c.id === '__ws-loader'));
    const sent = b.sent.map(m => JSON.parse(m)).find(m => m.act === 'save');
    check('offline call was not delivered', !sent);
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

  // Keep-alive heartbeats must flow on an open socket and stop on close.
  async keepAlivePings() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: -1, keepAlive: 40, hbTimeout: 1000});
    const a = env.sockets[0];
    a.accept();
    const beats = () => a.sent.filter(m => JSON.parse(m).act === '__hb');
    await sleep(60);
    check('heartbeat sent while open', beats().length === 1, 'beats=' + beats().length);
    a.deliver(JSON.stringify({__hb: JSON.parse(beats()[0]).hb}));
    await sleep(60);
    check('next heartbeat after the answer', beats().length >= 2, 'beats=' + beats().length);
    a.drop();
    const after = a.sent.length;
    await sleep(120);
    check('heartbeats stop after close', a.sent.length === after);
  },

  // A socket that looks open but never answers is replaced quickly.
  async unansweredHeartbeatReplacesSocket() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: 30, hbTimeout: 40});
    const a = env.sockets[0];
    a.accept(); hello(a, 'x');
    await waitFor(() => env.sockets.length > 1, 500);
    check('dead socket replaced without a close event', env.sockets.length === 2 && a.readyState >= 2,
      'sockets=' + env.sockets.length + ' state=' + a.readyState);
    check('client is not connected while replacing', !ws.connected());
    const b = env.latest(); b.accept(); hello(b, 'x');
    check('replacement connects', ws.connected() && ws.state() === 'connected');
  },

  // Coming back from the background probes an open socket at once.
  async resumeProbesOpenSocket() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: -1, keepAlive: -1, resumeTimeout: 40});
    const a = env.sockets[0];
    a.accept(); hello(a, 'x');
    env.sandbox.document.hidden = true;
    env.fire('visibilitychange');
    check('hidden page sends nothing', a.sent.length === 0);
    env.sandbox.document.hidden = false;
    env.fire('visibilitychange');
    check('resume sends a heartbeat', a.sent.some(m => JSON.parse(m).act === '__hb'));
    await sleep(80);
    check('unanswered resume heartbeat replaces the socket', env.sockets.length === 2);
  },

  // A socket that died in the background reconnects on resume without
  // flashing "Offline", however long it was down.
  async resumeAfterLongOutageDoesNotFlashOffline() {
    const env = makeEnv();
    // The grace outlasts the one-second reconnect cooldown.
    const ws = env.run({grace: 1200, connectingGrace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept(); hello(a, 'x');
    env.sandbox.document.hidden = true;
    a.drop();
    await sleep(1300);
    check('no badge while hidden', !env.badge());
    check('state reports offline while hidden', ws.state() === 'offline', ws.state());
    env.sandbox.document.hidden = false;
    env.fire('visibilitychange');
    await sleep(20);
    check('no offline flash right after resume', !env.badge());
    check('resume reports connecting', ws.state() === 'connecting', ws.state());
    await waitFor(() => env.latest().readyState === 0, 1100);
    env.latest().accept(); hello(env.latest(), 'x');
    await sleep(80);
    check('no badge after a quick reconnect', !env.badge() && ws.state() === 'connected');
  },

  // A short reconnect shows the quiet "Connecting…" badge, not "Offline",
  // and every state change is reported.
  async connectingBadgeAndStateEvents() {
    const env = makeEnv();
    const states = [];
    const ws = env.run({grace: 2000, connectingGrace: 30, reloadAfter: -1, keepAlive: -1});
    env.sandbox.window.addEventListener('gsui:connection', e => states.push(e.detail.state));
    const a = env.sockets[0];
    a.accept(); hello(a, 'x');
    a.drop();
    await sleep(80);
    const connecting = env.body.children.find(c => c.id === '__connecting__');
    check('connecting badge shows', !!connecting && !env.badge());
    check('connecting badge does not block input', connecting && /pointer-events:none/.test(connecting.style.cssText));
    await waitFor(() => env.latest() !== a, 1500);
    env.latest().accept(); hello(env.latest(), 'x');
    await sleep(200);
    check('connecting badge is gone after reconnect', !env.body.children.find(c => c.id === '__connecting__'));
    check('states reported in order', states.join(',') === 'connected,connecting,connected', states.join(','));
    check('html carries the state', env.sandbox.document.documentElement.getAttribute('data-gsui-connection') === 'connected');
    check('state() matches', ws.state() === 'connected');
  },

  // A browser that reports no network shows "Offline" at once.
  async browserOfflineShowsOfflineAtOnce() {
    const env = makeEnv();
    env.sandbox.navigator = {onLine: true};
    env.run({grace: 5000, reloadAfter: -1, keepAlive: -1});
    const a = env.sockets[0];
    a.accept(); hello(a, 'x');
    env.sandbox.navigator.onLine = false;
    a.drop();
    await sleep(20);
    check('offline badge at once without network', !!env.badge());
  },

  // A handshake that hangs is given up and tried again.
  async hangingHandshakeIsRetried() {
    const env = makeEnv();
    env.run({grace: -1, reloadAfter: -1, keepAlive: -1, handshakeTimeout: 40});
    const a = env.sockets[0];
    await waitFor(() => env.sockets.length > 1, 1500);
    check('hanging handshake retried', env.sockets.length === 2 && a.readyState >= 2, 'sockets=' + env.sockets.length);
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
  async rawFramesAreRejected() {
    const env = makeEnv();
    env.run({grace:-1,reloadAfter:-1,keepAlive:-1});
    const a=env.sockets[0];a.accept();
    a.deliver('window.rawFrameExecuted=true');
    a.deliver(JSON.stringify({js:'window.rawFrameExecuted=true'}));
    a.deliver(JSON.stringify({__r:1,id:1,js:'window.rawFrameExecuted=true'}));
    check('raw and unframed messages are ignored', !env.sandbox.window.rawFrameExecuted);
  },

  async helloFrameIsNotExecutedAsJs() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1, inst: 'srv-1'});
    const a = env.sockets[0];
    a.accept();
    ws.subscribe('clock.start', {});
    hello(a, 'srv-1');
    await sleep(20);
    a.drop();
    await waitFor(() => env.latest() !== a, 3000);
    const b = env.latest();
    b.accept(); hello(b, 'srv-1');
    await sleep(20);
    const subs = b.sent.filter(m => JSON.parse(m).act === 'clock.start');
    check('subscription survives the hello frame', subs.length === 1, 'sent=' + subs.length);
  },

  async skeletonShownAfterDelayAndRemovedOnSuccess() {
    const nav = navEnv({skelDelay: 30});
    nav.ws.navigate('/slow/7');
    const req = nav.lastNav();
    check('skeleton nav is sent', req && req.act === '__nav' && req.data.url === '/slow/7');
    check('region is busy at once, skeleton not yet shown', nav.content.getAttribute('aria-busy') === 'true' && !nav.skeleton());
    check('loading-start carries the url', nav.events.length === 1 && nav.events[0].type === 'gsui:loading-start' &&
      nav.events[0].detail.url === '/slow/7');
    await sleep(60);
    check('skeleton shown after delay and old content hidden', nav.skeleton() === nav.content.children[0] &&
      nav.old.style.display === 'none');
    await sleep(150);
    check('skeleton nav shows no loader overlay', !nav.loader());
    nav.reply(req);
    check('success removes skeleton and busy state', !nav.skeleton() && !nav.content.getAttribute('aria-busy') &&
      nav.content.children.length === 1 && nav.content.children[0].tagName === 'SECTION');
    check('success pushes history', nav.env.history.pushed[0] === '/slow/7');
    const end = nav.events[nav.events.length - 1];
    check('loading-end reports success', end.type === 'gsui:loading-end' && end.detail.url === '/slow/7' && end.detail.ok === true);
  },

  async fastNavigationShowsNoSkeleton() {
    const nav = navEnv({skelDelay: 50});
    nav.ws.navigate('/slow');
    await sleep(10);
    nav.reply(nav.lastNav());
    await sleep(80);
    check('fast navigation never builds a skeleton', nav.built() === 0 && !nav.skeleton());
    check('fast navigation ends busy state', !nav.content.getAttribute('aria-busy'));
  },

  async failedNavigationRestoresOldContent() {
    const nav = navEnv({skelDelay: -1});
    nav.old.style.display = 'flex';
    nav.ws.navigate('/slow');
    check('negative delay shows skeleton at once', !!nav.skeleton());
    const req = nav.lastNav();
    nav.socket.deliver(JSON.stringify({__r: 1, id: req.id, version: req.version, js: 'window.notified=true'}));
    check('error reply ran', nav.env.sandbox.window.notified === true);
    check('error reply restores old content', !nav.skeleton() && nav.content.children[0] === nav.old &&
      nav.old.style.display === 'flex' && !nav.content.getAttribute('aria-busy'));
    const end = nav.events[nav.events.length - 1];
    check('loading-end reports failure', end.type === 'gsui:loading-end' && end.detail.ok === false);
    check('failed navigation keeps the url', nav.env.history.pushed.length === 0);
  },

  async socketDropAndReconnectRestoreOldContent() {
    const nav = navEnv({skelDelay: -1});
    nav.ws.navigate('/slow');
    check('skeleton visible before drop', !!nav.skeleton());
    nav.socket.drop();
    check('socket drop restores old content', !nav.skeleton() && nav.old.style.display === '' &&
      !nav.content.getAttribute('aria-busy'));
    await waitFor(() => nav.env.latest() !== nav.socket, 3000);
    nav.env.latest().accept();
    await sleep(20);
    check('reconnect does not bring the skeleton back', !nav.skeleton() && nav.old.style.display === '');
  },

  async restartRestoresOldContent() {
    const nav = navEnv({skelDelay: -1});
    nav.ws.navigate('/slow');
    nav.ws.restart().catch(() => {});
    check('restart restores old content', !nav.skeleton() && nav.old.style.display === '');
  },

  async backForwardAndPrefixMatching() {
    const nav = navEnv({skelDelay: -1});
    nav.env.sandbox.location.pathname = '/slow/3';
    nav.env.sandbox.window.__gsuiPop({state: null});
    const req = nav.lastNav();
    check('back/forward shows the skeleton', !!nav.skeleton() && req.data.history === 'none');
    nav.reply(req);
    nav.ws.navigate('/slowpoke');
    check('prefix matches whole segments only', !nav.skeleton() && !nav.content.getAttribute('aria-busy'));
    nav.reply(nav.lastNav());
    nav.ws.navigate('/slow/off/1');
    check('nil skeleton turns a sub-path off', !nav.skeleton());
    await sleep(200);
    check('navigation without skeleton keeps the loader', !!nav.loader());
    nav.reply(nav.lastNav());
  },

  async failedSendDoesNotLeaveSkeleton() {
    const nav = navEnv({skelDelay: -1});
    nav.ws.navigate('/slow/1');
    check('first navigation shows skeleton', !!nav.skeleton());
    nav.socket.send = () => { throw new Error('send failed'); };
    nav.ws.navigate('/slow/2');
    check('unsent navigation restores old content', !nav.skeleton() && nav.old.style.display === '' &&
      !nav.content.getAttribute('aria-busy'));
    check('unsent call returns 0', nav.ws.call('x', {}) === 0);
  },

  async loadingEndListenerCanStartNavigation() {
    const nav = navEnv({skelDelay: 20});
    let once = true;
    nav.env.sandbox.window.addEventListener('gsui:loading-end', () => {
      if (once) { once = false; nav.ws.navigate('/slow/c'); }
    });
    nav.ws.navigate('/slow/a');
    nav.ws.navigate('/slow/b');
    await sleep(60);
    check('reentrant navigation shows one skeleton', nav.content.children.filter(c => c.getAttribute('data-gsui-skeleton') !== null).length === 1);
    nav.socket.drop();
    check('drop after reentrant navigation restores old content', !nav.skeleton() && nav.old.style.display === '');
    await waitFor(() => nav.env.latest() !== nav.socket, 3000);
    nav.env.latest().accept();
  },

  async loadingEndListenerCanRestart() {
    const nav = navEnv({skelDelay: -1});
    let once = true, done;
    nav.env.sandbox.window.addEventListener('gsui:loading-end', () => {
      if (once) { once = false; done = nav.ws.restart(); }
    });
    nav.ws.navigate('/slow/a');
    nav.ws.navigate('/slow/b');
    check('restart from loading-end restores old content', !nav.skeleton() && nav.old.style.display === '' &&
      !nav.content.getAttribute('aria-busy'));
    nav.env.latest().accept();
    await done;
    check('content stays visible after the restart', !nav.skeleton() && nav.old.style.display === '');
  },

  async pingReplyKeepsSkeleton() {
    const nav = navEnv({skelDelay: -1});
    const version = nav.env.sandbox.__gsuiVersion;
    nav.ws.navigate('/slow');
    nav.socket.deliver(JSON.stringify({__r: 1, id: 0, version, js: ''}));
    nav.socket.deliver(JSON.stringify({__r: 1, id: 0, version: version + 1, js: ''}));
    check('ping reply keeps the navigation skeleton', !!nav.skeleton() && nav.old.style.display === 'none' &&
      nav.content.getAttribute('aria-busy') === 'true');
  },

  async patchNavigationShowsNoSkeleton() {
    const nav = navEnv({skelDelay: -1});
    nav.ws.navigate('/slow?q=1', {patch: true});
    check('patch navigation is sent', nav.lastNav().data.patch === true);
    check('patch navigation shows no skeleton or events', !nav.skeleton() && nav.events.length === 0 &&
      !nav.content.getAttribute('aria-busy'));
  },

  async supersededNavigationReplacesSkeleton() {
    const nav = navEnv({skelDelay: -1});
    nav.ws.navigate('/slow/1');
    const first = nav.lastNav();
    nav.ws.navigate('/slow/2');
    const second = nav.lastNav();
    check('second navigation shows one skeleton', nav.content.children.filter(c => c.getAttribute('data-gsui-skeleton') !== null).length === 1 &&
      nav.old.style.display === 'none');
    nav.reply(first);
    check('stale reply does not end the newer navigation', !!nav.skeleton());
    nav.reply(second);
    check('newer reply finishes it', !nav.skeleton() && nav.env.history.pushed.join() === '/slow/2');
  },

  async loaderNoneCreatesNoOverlay() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1, loader: 'none'});
    env.latest().accept();
    const id = ws.call('slow.action', {});
    await sleep(200);
    check('loader none creates no overlay', id > 0 && !env.body.children.some(c => c.id === '__ws-loader'));
  },

  async quietCallSkipsLoader() {
    const env = makeEnv();
    const ws = env.run({grace: -1, reloadAfter: -1, keepAlive: -1});
    const a = env.latest(); a.accept();
    const quiet = ws.call('autosave', {}, null, null, {loader: false});
    await sleep(200);
    const loader = () => env.body.children.find(c => c.id === '__ws-loader');
    check('quiet call shows no loader', quiet > 0 && !loader());
    const loud = ws.call('save', {});
    await sleep(200);
    check('normal call still shows the loader', !!loader());
    a.deliver(JSON.stringify({__r: 1, id: loud, version: 1, js: ''}));
    await sleep(200);
    check('loader hides while only a quiet call is pending', !loader());
  },

  // The skeleton factory comes from Go (App.Skeleton) through cfg.js.
  async goSkeletonConfigBuildsSkeleton() {
    const env = makeEnv();
    vm.runInContext(cfgSrc, env.sandbox);
    const cfg = env.sandbox.window.__gsuiCfg;
    check('go cfg sets loader none and default delay', cfg.loader === 'none' && cfg.skelDelay === 100);
    check('go cfg lists the longest prefix first', cfg.skeletons[0][0] === '/mail/drafts' && cfg.skeletons[0][1] === null);
    const content = makeEl('main'); content.id = '__content__';
    env.body.appendChild(content);
    env.run(Object.assign({}, cfg, {grace: -1, reloadAfter: -1, keepAlive: -1, skelDelay: -1}));
    env.latest().accept();
    env.sandbox.__ws.navigate('/mail/inbox');
    const sk = content.children[0];
    check('go skeleton is built with status role and label', sk && sk.getAttribute('role') === 'status' &&
      sk.getAttribute('aria-label') === 'Loading' && sk.getAttribute('data-gsui-skeleton') === '' &&
      sk.className.includes('motion-reduce:animate-none') && sk.children.length === 5);
  },
};

// navEnv starts a connected client with #__content__ holding one old child
// and skeletons for /slow (sub-path /slow/off disabled).
function navEnv(cfg) {
  const env = makeEnv();
  let built = 0;
  const events = [];
  const content = makeEl('main'); content.id = '__content__';
  const old = makeEl('div'); content.appendChild(old);
  env.body.appendChild(content);
  const build = () => { built++; const el = makeEl('div'); el.setAttribute('data-gsui-skeleton', ''); return el; };
  const ws = env.run(Object.assign({grace: -1, reloadAfter: -1, keepAlive: -1,
    skeletons: [['/slow/off', null], ['/slow', build]]}, cfg));
  const socket = env.latest(); socket.accept();
  ['gsui:loading-start', 'gsui:loading-end'].forEach(t => env.sandbox.window.addEventListener(t, e => events.push(e)));
  return {
    env, ws, socket, content, old, events, built: () => built,
    skeleton: () => content.children.find(c => c.getAttribute('data-gsui-skeleton') !== null) || null,
    loader: () => env.body.children.find(c => c.id === '__ws-loader') || null,
    lastNav: () => socket.sent.map(m => JSON.parse(m)).filter(m => m.act === '__nav').pop(),
    // reply answers a __nav request the way the server does: replace the
    // content, then finish the navigation.
    reply(req) {
      socket.deliver(JSON.stringify({__r: 1, id: req.id, version: req.version, js:
        "var c=document.getElementById('__content__');c.children.slice().forEach(function(x){c.removeChild(x)});" +
        "c.appendChild(document.createElement('section'));" +
        '__gsuiNavigationDone(' + JSON.stringify({url: req.data.url, history: req.data.history, patch: req.data.patch, x: 0, y: 0}) + ');'}));
    },
  };
}

function hello(sock, id) { sock.deliver(JSON.stringify({__hello: id})); }

(async () => {
  for (const [name, fn] of Object.entries(tests)) {
    try { await fn(); } catch (e) { failures++; console.log('ERROR: ' + name + ' -- ' + e.stack); }
  }
  if (failures) { console.log(failures + ' failure(s)'); process.exit(1); }
  console.log('ok - ' + Object.keys(tests).length + ' client scenarios');
  // Clients of finished scenarios keep retrying their sockets; stop them.
  process.exit(0);
})();
