// Browser mechanics for Go-owned pages. No application state lives here.
(function () {
  window.__gsuiVersion = window.__gsuiVersion || 1;
  window.__gsuiPage = location.pathname + location.search;
  window.__gsuiPushSubscription = '';
  window.__gsuiReport = function(message) {
    var old=document.getElementById('__gsui-action-error');if(old)old.remove();
    var el=document.createElement('div');el.id='__gsui-action-error';el.setAttribute('role','alert');
    el.className='fixed top-4 right-4 z-50 rounded bg-red-600 text-white px-4 py-3 shadow-lg';
    el.textContent=message;document.body.appendChild(el);setTimeout(function(){el.remove()},6000);
  };

  window.__gsuiDispose = function dispose(node) {
    if (!node) return;
    if (node.__gsuiCleanup) {
      node.__gsuiCleanup.splice(0).forEach(function (fn) { try { fn(); } catch (e) { console.error(e); } });
    }
    Array.from(node.children || []).forEach(dispose);
  };

  function key(node) { return node.nodeType === 1 ? node.getAttribute('data-gsui-key') || node.id : ''; }
  function compatible(a, b) { return a.nodeType === b.nodeType && a.nodeName === b.nodeName && a.namespaceURI === b.namespaceURI; }
  function bind(old, fresh) {
    Object.keys(old.__gsuiHandlers || {}).forEach(function (event) {
      old.removeEventListener(event, old.__gsuiHandlers[event]);
    });
    old.__gsuiHandlers = fresh.__gsuiHandlers || {};
    Object.keys(old.__gsuiHandlers).forEach(function (event) {
      old.addEventListener(event, old.__gsuiHandlers[event]);
    });
  }
  function morph(old, fresh) {
    if (!compatible(old, fresh)) {
      __gsuiDispose(old); old.replaceWith(fresh); return fresh;
    }
    fresh.__gsuiMounted = old;
    if (old.nodeType !== 1) {
      if (old.nodeValue !== fresh.nodeValue) old.nodeValue = fresh.nodeValue;
      return old;
    }
    if (old.hasAttribute('data-gsui-preserve')) {
      (function skip(node){node.__gsuiMounted=true;Array.from(node.children||[]).forEach(skip);})(fresh);
      return old;
    }
    var active = old === document.activeElement;
    var busy = old.getAttribute('aria-busy') === 'true';
    var input = /^(INPUT|TEXTAREA|SELECT)$/.test(old.tagName);
    var dirty = input && (active || (old.tagName==='SELECT'
      ? Array.from(old.options).some(function(o){return o.selected!==o.defaultSelected})
      : /^(checkbox|radio)$/.test(old.type) ? old.checked!==old.defaultChecked : old.value!==old.defaultValue));
    var keep = input && dirty && !fresh.hasAttribute('data-gsui-reset');
    var value = old.value, checked = old.checked, freshValue = fresh.value, freshChecked = fresh.checked;
    var freshSelected = fresh.tagName === 'SELECT' ? Array.from(fresh.selectedOptions).map(function (o) { return o.value; }) : [];
    var selected = old.tagName === 'SELECT' ? Array.from(old.selectedOptions).map(function (o) { return o.value; }) : [];
    var start = active ? old.selectionStart : null, end = active ? old.selectionEnd : null;
    // An open dialog is user state, like a typed value: a morph never closes
    // it. The page closes it when it is done (dialog.close()).
    var openDialog = old.tagName === 'DIALOG' && old.open;
    Array.from(old.attributes).forEach(function (attr) {
      if (!fresh.hasAttribute(attr.name) && !(openDialog && attr.name === 'open')) old.removeAttribute(attr.name);
    });
    Array.from(fresh.attributes).forEach(function (attr) {
      if (old.getAttribute(attr.name) !== attr.value) old.setAttribute(attr.name, attr.value);
    });
    bind(old, fresh);
    if (old.hasAttribute('data-gsui-active') && window.__gsui) __gsui.active(old);
    if(busy){old.setAttribute('aria-busy','true');if(old.tagName==='BUTTON')old.disabled=true;old.classList.add('gsui-busy','opacity-60','cursor-wait')}
    children(old, fresh);
    if (input) {
      if (old.type !== 'file') old.value = keep ? value : freshValue;
      if (old.tagName === 'INPUT') old.checked = keep ? checked : freshChecked;
      if (old.tagName === 'SELECT' && old.multiple) {
        var wanted = keep ? selected : freshSelected;
        Array.from(old.options).forEach(function (o) { o.selected = wanted.indexOf(o.value) >= 0; });
      }
      if (keep && start !== null && old.setSelectionRange) {
        try { old.setSelectionRange(start, end); } catch (_) {}
      }
    }
    return old;
  }
  function openDialogIn(node) {
    return node.nodeType === 1 && (node.matches('dialog[open]') || !!node.querySelector('dialog[open]'));
  }
  function children(old, fresh) {
    var existing = Array.from(old.childNodes), keyed = new Map();
    existing.forEach(function (n) { if (key(n)) keyed.set(key(n), n); });
    var used = new Set(), cursor = old.firstChild;
    Array.from(fresh.childNodes).forEach(function (next) {
      var id = key(next), match = id ? keyed.get(id) : null;
      if (!id && cursor && !key(cursor) && !used.has(cursor) && compatible(cursor, next)) match = cursor;
      if (match && used.has(match)) match = null;
      var mounted = match ? morph(match, next) : next;
      if(match && match === cursor) cursor = mounted;
      if (mounted !== cursor && match && mounted.parentNode === old && openDialogIn(mounted)) {
        // Moving an open modal dialog would take it out of the top layer:
        // the nodes before it move behind it instead, in order, unless one
        // of them holds an open dialog too.
        var between = [];
        for (var n = cursor; n && n !== mounted; n = n.nextSibling) between.push(n);
        if (!between.some(openDialogIn)) {
          var tail = mounted.nextSibling;
          between.forEach(function (node) { old.insertBefore(node, tail); });
          cursor = mounted;
        }
      }
      if (mounted !== cursor) old.insertBefore(mounted, cursor);
      used.add(mounted);
      cursor = mounted.nextSibling;
    });
    Array.from(old.childNodes).forEach(function (n) {
      if (!used.has(n)) { __gsuiDispose(n); n.remove(); }
    });
  }
  window.__gsuiMorph = morph;
  window.__gsuiMorphChildren = children;

  var scrolls = new Map(), historyKey = 0;
  function rememberScroll() {
    var state = history.state || {};
    if (!state.__gsuiKey && history.replaceState) {
      state = Object.assign({}, state, {__gsuiKey: ++historyKey});
      history.replaceState(state, '', location.href);
    }
    scrolls.set(state.__gsuiKey, {x: window.scrollX || 0, y: window.scrollY || 0});
  }
  window.__gsuiNavigate = function (path, options) {
    options = options || {};
    var url = new URL(path, location.href);
    // Before the client runs, or when the socket is reported offline, a full
    // load is the only way. While connecting, the call waits for the socket.
    if (url.origin !== location.origin || !__ws.beginNavigation || (!__ws.connected() && __ws.state() === 'offline')) { location.assign(url.href); return; }
    // A full queue refuses the navigation before the page changes anything.
    if (!__ws.accepts(true)) { window.__gsuiReport('Connection lost before sending. Please try again.'); return; }
    if (options.history !== 'none') rememberScroll();
    if (!options.patch) window.__gsuiVersion++;
    __ws.beginNavigation(!!options.patch);
    var target = url.pathname + url.search + url.hash;
    var build = options.patch ? null : skeletonFor(url.pathname);
    var id = __ws.call('__nav', {
      url: target,
      history: options.history || (options.replace ? 'replace' : 'push'),
      patch: !!options.patch,
      regions: options.patch && options.regions || undefined,
      x: options.x || 0, y: options.y || 0
    }, null, null, build ? {loader:false} : undefined);
    if (options.patch) return;
    // Refused or waiting for a connection: the old content stays meanwhile.
    if (id && __ws.connected()) window.__gsuiLoadStart(id, target, build);
    else window.__gsuiLoadEnd(0, false);
  };

  // Navigation skeletons (App.Skeleton). A slow page load hides the content
  // and shows a skeleton; a failed one shows the old content again.
  var load = null;
  function skeletonFor(path) {
    var list = (window.__gsuiCfg || {}).skeletons || [];
    for (var i = 0; i < list.length; i++) {
      var p = list[i][0];
      if (p === '/' || path === p || path.indexOf(p + '/') === 0) return list[i][1];
    }
    return null;
  }
  function skeletonRoot() {
    var id = (window.__gsuiCfg || {}).skelTarget;
    return (id && document.getElementById(id)) || document.getElementById('__content__') || document.querySelector('main') || document.body;
  }
  function showSkeleton(l) {
    l.timer = 0;
    try { l.el = l.build(); } catch (e) { console.error('gsui: skeleton failed:', e); return; }
    l.hidden = Array.from(l.root.children).map(function (c) { var d = c.style.display; c.style.display = 'none'; return [c, d]; });
    l.root.insertBefore(l.el, l.root.firstChild);
  }
  function loadEvent(name, detail) {
    try { window.dispatchEvent(new CustomEvent(name, {detail: detail})); } catch (_) {}
  }
  // finish clears load l. Events are sent by the callers once the state is
  // final, so listeners that navigate or restart the socket see it settled.
  function finish(l) {
    if (load === l) load = null;
    if (l.timer) clearTimeout(l.timer);
    if (l.el) {
      l.el.remove();
      l.hidden.forEach(function (h) { h[0].style.display = h[1]; });
    }
    if (l.root) l.root.removeAttribute('aria-busy');
  }
  window.__gsuiLoadStart = function (id, url, build) {
    var prev = load, shown = !!(prev && prev.el);
    if (prev) finish(prev);
    var l = load = {id: id, url: url, build: build, timer: 0};
    if (build) {
      l.root = skeletonRoot();
      l.root.setAttribute('aria-busy', 'true');
      var delay = (window.__gsuiCfg || {}).skelDelay;
      if (typeof delay !== 'number') delay = 100;
      // A skeleton already on screen is replaced at once to avoid a flash.
      if (shown || delay <= 0) showSkeleton(l);
      else l.timer = setTimeout(function () { if (load === l) showSkeleton(l); }, delay);
    }
    if (prev) loadEvent('gsui:loading-end', {url: prev.url, ok: false});
    if (load === l) loadEvent('gsui:loading-start', {url: url});
  };
  // __gsuiLoadEnd ends the navigation with request id (0: any). It runs on
  // success, on an error reply and when the socket drops.
  window.__gsuiLoadEnd = function (id, ok) {
    var l = load;
    if (!l || (id && l.id !== id)) return;
    finish(l);
    loadEvent('gsui:loading-end', {url: l.url, ok: !!ok});
  };
  window.__gsuiNavigationDone = function (request) {
    if (!request.patch) window.__gsuiLoadEnd(0, true);
    if (request.history === 'push') history.pushState({__gsuiKey: ++historyKey}, '', request.url);
    if (request.history === 'replace') history.replaceState(history.state, '', request.url);
    var root = document.getElementById('__content__') || document.querySelector('main') || document.body;
    var title = root.querySelector('[data-gsui-title]');
    if (title) document.title = title.getAttribute('data-gsui-title');
    if (request.history === 'none') window.scrollTo(request.x, request.y);
    else if (!request.patch) {
      var hash = new URL(request.url, location.href).hash;
      var target = hash ? document.getElementById(decodeURIComponent(hash.slice(1))) : null;
      if (target) target.scrollIntoView(); else window.scrollTo(0, 0);
    }
    if (!request.patch) {
      var focus = root.querySelector('[autofocus],h1') || root;
      if (!focus.hasAttribute('tabindex')) focus.setAttribute('tabindex', '-1');
      focus.focus({preventScroll:true});
    }
    window.dispatchEvent(new CustomEvent('gsui:navigated', {detail:request}));
  };
  window.__gsuiQueryDone = function (path, replace) {
    rememberScroll();
    if(replace)history.replaceState(history.state,'',path);
    else history.pushState({__gsuiKey:++historyKey},'',path);
    window.__gsuiPage=location.pathname+location.search;
    window.dispatchEvent(new CustomEvent('gsui:navigated',{detail:{url:path,patch:true}}));
  };
  window.__gsuiPop = function (event) {
    var pos = scrolls.get(event && event.state && event.state.__gsuiKey) || {x:0,y:0};
    __gsuiNavigate(location.pathname + location.search + location.hash, {history:'none',x:pos.x,y:pos.y});
  };
  window.addEventListener('scroll', function () {
    if (history.state && history.state.__gsuiKey) scrolls.set(history.state.__gsuiKey,{x:window.scrollX,y:window.scrollY});
  }, {passive:true});
  document.addEventListener('click', function (event) {
    if (event.defaultPrevented || event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
    var a = event.target.closest && event.target.closest('a[data-gsui-nav]');
    if (!a || a.hasAttribute('download') || (a.target && a.target !== '_self')) return;
    var url = new URL(a.href, location.href);
    if (url.origin !== location.origin) return;
    if (url.pathname === location.pathname && url.search === location.search && url.hash) return;
    event.preventDefault();
    __gsuiNavigate(a.href, {patch:a.hasAttribute('data-gsui-patch'),replace:a.hasAttribute('data-gsui-replace')});
  });

  window.__gsuiFormErrors = function (id, errors) {
    var form = document.getElementById(id);
    if (!form) return;
    form.querySelectorAll('[data-gsui-error]').forEach(function (el) { el.remove(); });
    Array.from(form.elements).forEach(function (el) { el.removeAttribute('aria-invalid');el.removeAttribute('aria-errormessage'); });
    var first;
    Object.keys(errors).forEach(function (name) {
      var input = Array.from(form.elements).find(function (el) { return el.name === name; });
      if (!input) return;
      var error = document.createElement('span');
      error.id = id + '-error-' + name;
      error.setAttribute('data-gsui-error',''); error.setAttribute('role','alert');
      error.textContent = errors[name];
      input.setAttribute('aria-invalid','true'); input.setAttribute('aria-errormessage',error.id);
      input.after(error); first = first || input;
    });
    if (first) first.focus();
  };
  window.__gsuiSubmit = function (event, action, values) {
    var form = event.currentTarget;
    if(form.getAttribute('aria-busy')==='true')return;
    if (!form.reportValidity()) return;
    __gsuiFormErrors(form.id,{});
    var data = {__form:form.id};
    if (values) {
      // FormValues: the entry list a native post would send.
      // The submit button goes last, as in the typed mode.
      var entries = new FormData(form), submitter = event.submitter;
      if (submitter && submitter.name) entries.append(submitter.name, submitter.value);
      data.__values = [];
      entries.forEach(function (value, name) { if (typeof value === 'string') data.__values.push([name, value]); });
      __ws.call(action,data,null,event.submitter || form);
      return;
    }
    Array.from(form.elements).forEach(function (el) {
      if (!el.name || el.disabled || /^(submit|button|reset|file)$/.test(el.type)) return;
      if (el.type === 'radio') { if (el.checked) data[el.name] = el.value; }
      else if (el.type === 'checkbox') data[el.name] = el.checked;
      else if (el.type === 'number' || el.type === 'range') data[el.name] = el.value === '' ? null : el.valueAsNumber;
      else if (el.multiple) data[el.name] = Array.from(el.selectedOptions).map(function (o) { return o.value; });
      else data[el.name] = el.value;
    });
    var button = event.submitter;
    if (button && button.name) data[button.name] = button.value;
    __ws.call(action,data,null,button || form);
  };
})();
