package ui

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// connState tracks per-connection cancellation so that server-side Push
// goroutines stop when the client navigates away or reports a missing
// target element.
type connState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// ---------------------------------------------------------------------------
// App: routes, actions, server
// ---------------------------------------------------------------------------

// httpRoute stores a deferred HTTP handler registration (method + pattern).
type httpRoute struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

// LayoutHandler builds a shared layout wrapping every page. The returned
// *Node tree must contain exactly one element with ID("__content__") where
// the page content will be injected.
type LayoutHandler func(ctx *Context) *Node

// App is the top-level application container. It holds page routes (GET)
// and named actions (WS). Pages return a *Node tree that compiles to JS
// for the initial render. Actions return raw JS strings for DOM mutations.
type App struct {
	mu         sync.RWMutex
	pages      map[string]PageHandler
	actions    map[string]ActionHandler
	clients    map[*websocket.Conn]bool
	connStates map[*websocket.Conn]*connState
	mux        *http.ServeMux
	httpRoutes []httpRoute
	layout     LayoutHandler

	// Favicon is the URL path for the site favicon (e.g. "/assets/favicon.svg").
	// When set, a <link rel="icon"> tag is emitted in the HTML shell.
	Favicon string

	// Title sets the <title> element in the HTML shell.
	// Improves accessibility (screen readers) and SEO.
	Title string

	// Description sets the <meta name="description"> content in the HTML shell.
	// Used by search engines to summarize page content.
	Description string

	// HTMLHead contains raw HTML strings injected into the <head> section
	// after the built-in Tailwind/Material Icons/WS script tags.
	// Each entry is emitted as-is (e.g. "<style>body{margin:0}</style>",
	// "<script src=\"...\"></script>", "<link ...>").
	HTMLHead []string
}

// PageHandler builds the initial DOM for a GET route.
type PageHandler func(ctx *Context) *Node

// ActionHandler processes a WS action call and returns a JS string
// to execute on the client.
type ActionHandler func(ctx *Context) string

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		pages:      make(map[string]PageHandler),
		actions:    make(map[string]ActionHandler),
		clients:    make(map[*websocket.Conn]bool),
		connStates: make(map[*websocket.Conn]*connState),
		mux:        http.NewServeMux(),
	}
}

// Page registers a GET route. The handler returns a *Node tree which is
// compiled to JS and served inside a minimal HTML shell.
func (app *App) Page(path string, handler PageHandler) {
	app.mu.Lock()
	app.pages[path] = handler
	app.mu.Unlock()
}

// CSS registers external stylesheets and/or inline CSS rules that apply
// to every page in the application. The tags are injected into the HTML
// <head> server-side, so they load immediately without JavaScript.
//
//	app.CSS(
//	    []string{"https://fonts.googleapis.com/css2?family=Oswald&display=swap"},
//	    `body { font-family: 'Oswald', sans-serif; }`,
//	)
//
// Pass nil for urls if you only need inline CSS, or "" for css if you
// only need external links.
func (app *App) CSS(urls []string, css string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	for _, u := range urls {
		app.HTMLHead = append(app.HTMLHead,
			fmt.Sprintf(`<link rel="stylesheet" href="%s">`, u))
	}
	if css != "" {
		app.HTMLHead = append(app.HTMLHead,
			fmt.Sprintf("<style>%s</style>", css))
	}
}

// Action registers a named server action callable via WebSocket.
// The handler receives a Context (with .Body() for payload) and returns
// a raw JS string that the client executes directly.
func (app *App) Action(name string, handler ActionHandler) {
	app.mu.Lock()
	app.actions[name] = handler
	app.mu.Unlock()
}

// GET registers a standard HTTP GET handler on the internal mux.
// Use this for REST API endpoints that return JSON, files, etc.
func (app *App) GET(path string, handler http.HandlerFunc) {
	app.mu.Lock()
	app.httpRoutes = append(app.httpRoutes, httpRoute{"GET", path, handler})
	app.mu.Unlock()
}

// POST registers a standard HTTP POST handler on the internal mux.
func (app *App) POST(path string, handler http.HandlerFunc) {
	app.mu.Lock()
	app.httpRoutes = append(app.httpRoutes, httpRoute{"POST", path, handler})
	app.mu.Unlock()
}

// DELETE registers a standard HTTP DELETE handler on the internal mux.
func (app *App) DELETE(path string, handler http.HandlerFunc) {
	app.mu.Lock()
	app.httpRoutes = append(app.httpRoutes, httpRoute{"DELETE", path, handler})
	app.mu.Unlock()
}

// Assets serves static files from an embedded or on-disk filesystem.
// The dir is stripped from fsys (via fs.Sub) and the result is served
// under the given URL prefix. Example:
//
//	//go:embed assets/*
//	var assets embed.FS
//	app.Assets(assets, "assets", "/assets/")
//
// This makes assets/favicon.svg available at /assets/favicon.svg.
func (app *App) Assets(fsys fs.FS, dir, prefix string) {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		log.Fatalf("gsui: assets: fs.Sub(%q): %v", dir, err)
	}
	app.mux.Handle(prefix, http.StripPrefix(prefix, http.FileServerFS(sub)))
}

// Layout registers a layout handler that wraps every page render.
// The handler returns a *Node tree that must contain exactly one element
// with ID("__content__"). The page handler's output is injected there.
func (app *App) Layout(handler LayoutHandler) {
	app.mu.Lock()
	app.layout = handler
	app.mu.Unlock()
}

// Handler sets up routes and returns the internal http.Handler (mux).
// Use this when you need a custom http.Server for graceful shutdown, TLS, etc.
//
//	srv := &http.Server{Addr: ":8080", Handler: app.Handler()}
//	srv.ListenAndServe()
func (app *App) Handler() http.Handler {
	app.setup()
	return app.mux
}

// Listen sets up HTTP handlers and starts the server.
func (app *App) Listen(addr string) error {
	app.setup()
	log.Printf("gsui: listening on %s", addr)
	return http.ListenAndServe(addr, app.mux)
}

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------

func (app *App) setup() {
	// Built-in __nav action: handles popstate (browser back/forward).
	// Looks up the page handler for the URL and replaces content.
	// If a layout is registered, only the __content__ container is replaced
	// (the layout shell stays). Otherwise the full body is cleared and rebuilt.
	// Also cancels any outstanding Push goroutines for the connection since
	// their target elements no longer exist after navigation.
	app.Action("__nav", func(ctx *Context) string {
		// Cancel active pushes — the old page's elements are gone.
		app.cancelConn(ctx.wsConn)

		var req struct {
			URL string `json:"url"`
		}
		ctx.Body(&req)

		app.mu.RLock()
		handler, ok := app.pages[req.URL]
		layoutFn := app.layout
		app.mu.RUnlock()

		if !ok {
			return ""
		}

		pageNode := handler(ctx)
		if pageNode == nil {
			return ""
		}

		// With a layout: replace only __content__ inner content.
		// Without a layout: clear body and append the full page tree.
		if layoutFn != nil {
			return pageNode.ToJSInner("__content__")
		}
		return "(function(){document.body.innerHTML=''})();" + pageNode.ToJS()
	})

	// Built-in __notfound action: the client sends this when a WS patch
	// targets a DOM element that no longer exists. Cancel push context so
	// server-side goroutines calling ctx.Push() get an error and stop.
	app.Action("__notfound", func(ctx *Context) string {
		app.cancelConn(ctx.wsConn)
		return ""
	})

	// Serve the tiny WS client script
	app.mux.HandleFunc("GET /__ws.js", app.serveWSClient)

	// WebSocket endpoint
	app.mux.Handle("/__ws", websocket.Handler(app.handleWS))

	// Register user-defined HTTP routes (GET/POST/DELETE) before the
	// catch-all page handler so they take precedence.
	for _, rt := range app.httpRoutes {
		pattern := rt.method + " " + rt.pattern
		app.mux.HandleFunc(pattern, rt.handler)
	}

	// Page routes (catch-all)
	app.mux.HandleFunc("/", app.handlePage)
}

// ---------------------------------------------------------------------------
// Page handler: GET requests return minimal HTML + JS
// ---------------------------------------------------------------------------

// injectContent recursively walks the node tree and appends pageNode as a
// child of the first node with id="__content__". Returns true if found.
func injectContent(node, pageNode *Node) bool {
	if node.id == "__content__" {
		node.children = append(node.children, pageNode)
		return true
	}
	for _, child := range node.children {
		if injectContent(child, pageNode) {
			return true
		}
	}
	return false
}

func (app *App) handlePage(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	handler, ok := app.pages[r.URL.Path]
	app.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	ctx := &Context{
		Request:    r,
		PathParams: make(map[string]string),
		Query:      make(map[string]string),
	}

	// Parse query params
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			ctx.Query[k] = v[0]
		}
	}

	// Build the node tree
	pageNode := handler(ctx)
	if pageNode == nil {
		http.Error(w, "page handler returned nil", 500)
		return
	}

	// If a layout is registered, wrap the page content inside it.
	// The layout tree must contain a node with id="__content__" where
	// the page content gets injected as children.
	var root *Node
	app.mu.RLock()
	layoutFn := app.layout
	app.mu.RUnlock()

	if layoutFn != nil {
		layoutNode := layoutFn(ctx)
		if layoutNode != nil {
			injectContent(layoutNode, pageNode)
			root = layoutNode
		} else {
			root = pageNode
		}
	} else {
		root = pageNode
	}

	// Compile to JS
	jsBody := root.ToJS()

	// Respond with minimal HTML shell
	faviconTag := ""
	if app.Favicon != "" {
		faviconTag = fmt.Sprintf(`<link rel="icon" href="%s">`, app.Favicon)
	}
	titleTag := "<title>App</title>"
	if app.Title != "" {
		titleTag = fmt.Sprintf(`<title>%s</title>`, app.Title)
	}
	descTag := ""
	if app.Description != "" {
		descTag = fmt.Sprintf(`<meta name="description" content="%s">`, app.Description)
	}

	// Build custom head HTML
	customHead := strings.Join(app.HTMLHead, "\n")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Use gzip compression if the client supports it
	var writer io.Writer = w
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		writer = gz
	}

	fmt.Fprintf(writer, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
%s
%s
%s
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<script>%s</script>
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4" async></script>
<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons+Round" media="print" onload="this.media='all'">
<noscript><link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons+Round"></noscript>
<style type="text/tailwindcss">
@custom-variant dark (&:where(.dark, .dark *));
@theme{--font-sans:ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif;}
</style>
<style>%s</style>
<script src="/__ws.js" defer></script>
%s
</head>
<body>
<script>
%s
</script>
</body>
</html>`, faviconTag, titleTag, descTag, themeInitJS, darkOverrideCSS, customHead, jsBody)
}

// ---------------------------------------------------------------------------
// WebSocket client script
// ---------------------------------------------------------------------------

func (app *App) serveWSClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	// Gzip compress if supported
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gz.Write([]byte(wsClientJS))
		return
	}
	w.Write([]byte(wsClientJS))
}

// themeInitJS runs synchronously in <head> before the body renders to
// prevent FOUC. It reads the stored theme from localStorage, applies the
// "dark" class on <html>, and exposes setTheme(mode) / toggleTheme() globals.
const themeInitJS = `(function(){
if(window.__gsuiThemeInit)return;window.__gsuiThemeInit=true;
var d=document.documentElement;
function apply(m){if(m==='system')m=(window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches)?'dark':'light';if(m==='dark'){d.classList.add('dark');d.style.colorScheme='dark'}else{d.classList.remove('dark');d.style.colorScheme='light'}}
function set(m){localStorage.setItem('theme',m);apply(m)}
window.setTheme=set;window.toggleTheme=function(){set(d.classList.contains('dark')?'light':'dark')};
apply(localStorage.getItem('theme')||'system');
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change',function(){var s=localStorage.getItem('theme')||'';if(!s||s==='system')apply('system')});
})();`

// darkOverrideCSS provides fallback dark mode overrides for elements that
// may not carry explicit dark: Tailwind classes (e.g. plain body, form UA styles).
// These use .dark selectors without !important so explicit dark: variants win.
// Also applies a local font stack to avoid late webfont swaps and CLS.
const darkOverrideCSS = `html.dark{color-scheme:dark}
.dark body{color:#e5e7eb}
.dark .bg-white,.dark .bg-gray-50,.dark .bg-gray-100,.dark .bg-gray-200{background-color:#111827}
.dark .text-black,.dark .text-gray-900,.dark .text-gray-800,.dark .text-gray-700,.dark .text-gray-600,.dark .text-gray-500{color:#e5e7eb}
.dark .text-gray-400,.dark .text-gray-300{color:#d1d5db}
.dark .border-gray-100,.dark .border-gray-200,.dark .border-gray-300{border-color:#374151}
.dark input,.dark select,.dark textarea{color:#e5e7eb!important;background-color:#1f2937!important}
.dark input::placeholder,.dark textarea::placeholder{color:#9ca3af!important}
.dark .hover\:bg-gray-200:hover{background-color:#374151}
.dark .hover\:bg-gray-100:hover{background-color:#1f2937}
.dark .hover\:bg-gray-50:hover{background-color:#1f2937}
body{font-family:ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif}`

// wsClientJS is the entire client-side framework.
// It connects to the WS endpoint, sends action calls, executes
// whatever JS string the server sends back, and shows an offline
// overlay when the WebSocket disconnects.
const wsClientJS = `var __offline=(function(){
  var el=null;
  function show(){
    if(document.getElementById('__offline__')){el=document.getElementById('__offline__');return;}
    try{document.body.classList.add('pointer-events-none');}catch(_){}
    var o=document.createElement('div');o.id='__offline__';
    o.style.cssText='position:fixed;inset:0;z-index:60;pointer-events:none;opacity:0;transition:opacity 160ms ease-out;backdrop-filter:blur(2px);-webkit-backdrop-filter:blur(2px);background:'+(document.documentElement.classList.contains('dark')?'rgba(0,0,0,0.3)':'rgba(255,255,255,0.18)');
    var b=document.createElement('div');
    b.className='absolute top-3 left-3 flex items-center gap-2 rounded-full px-3 py-1 text-white shadow-lg ring-1 ring-white/30';
    b.style.background='linear-gradient(135deg,#ef4444,#ec4899)';
    var dot=document.createElement('span');dot.className='inline-block h-2.5 w-2.5 rounded-full bg-white/95 animate-pulse';
    var lbl=document.createElement('span');lbl.className='font-semibold tracking-wide';lbl.style.color='#fff';lbl.textContent='Offline';
    var sub=document.createElement('span');sub.className='ml-1 text-xs';sub.style.color='rgba(255,255,255,0.9)';sub.textContent='Trying to reconnect\u2026';
    b.appendChild(dot);b.appendChild(lbl);b.appendChild(sub);o.appendChild(b);
    document.body.appendChild(o);
    requestAnimationFrame(function(){o.style.opacity='1';});
    el=o;
  }
  function hide(){
    try{document.body.classList.remove('pointer-events-none');}catch(_){}
    var o=document.getElementById('__offline__');if(!o){el=null;return;}
    try{o.style.opacity='0';}catch(_){}
    setTimeout(function(){try{if(o&&o.parentNode){o.parentNode.removeChild(o);}}catch(_){}},150);
    el=null;
  }
  return{show:show,hide:hide};
})();
var __ws=(function(){
  var ws,q=[],ready=false,pending=0,loaderEl=null,loaderTimer=0,hadClose=false;
  function showLoader(){
    pending++;
    if(loaderEl||loaderTimer)return;
    loaderTimer=setTimeout(function(){
      loaderTimer=0;
      var o=document.createElement('div');
      o.id='__ws-loader';
      o.className='fixed inset-0 z-50 flex items-center justify-center transition-opacity opacity-0';
      o.style.cssText='backdrop-filter:blur(3px);-webkit-backdrop-filter:blur(3px);background:'+(document.documentElement.classList.contains('dark')?'rgba(0,0,0,0.35)':'rgba(255,255,255,0.28)')+';pointer-events:auto';
      var b=document.createElement('div');
      b.className='absolute top-3 left-3 flex items-center gap-2 rounded-full px-3 py-1 text-white shadow-lg ring-1 ring-white/30';
      b.style.background='linear-gradient(135deg,#6366f1,#22d3ee)';
      var dot=document.createElement('span');dot.className='inline-block h-2.5 w-2.5 rounded-full bg-white/95 animate-pulse';
      var lbl=document.createElement('span');lbl.className='font-semibold tracking-wide';lbl.textContent='Loading\u2026';
      var sub=document.createElement('span');sub.className='ml-1 text-white/85 text-xs';sub.style.color='rgba(255,255,255,0.9)';sub.textContent='Please wait';
      b.appendChild(dot);b.appendChild(lbl);b.appendChild(sub);o.appendChild(b);
      document.body.appendChild(o);
      requestAnimationFrame(function(){o.style.opacity='1';});
      loaderEl=o;
    },120);
  }
  function hideLoader(){
    pending--;
    if(pending>0)return;
    pending=0;
    if(loaderTimer){clearTimeout(loaderTimer);loaderTimer=0;}
    if(loaderEl){loaderEl.style.opacity='0';var el=loaderEl;loaderEl=null;setTimeout(function(){try{if(el&&el.parentNode)el.parentNode.removeChild(el)}catch(_){}},160);}
  }
  function connect(){
    ws=new WebSocket((location.protocol==='https:'?'wss://':'ws://')+location.host+'/__ws');
    ws.onopen=function(){
      ready=true;
      __offline.hide();
      q.forEach(function(m){ws.send(m)});q=[];
      if(hadClose){hadClose=false;try{location.reload();return;}catch(_){}}
    };
    ws.onmessage=function(e){hideLoader();try{new Function(e.data)()}catch(err){console.error('ws exec error:',err,e.data)}};
    ws.onclose=function(){ready=false;pending=0;hideLoader();__offline.show();hadClose=true;setTimeout(connect,1500)};
    ws.onerror=function(){ws.close()};
  }
  connect();
  window.addEventListener('popstate',function(){
    var msg=JSON.stringify({act:'__nav',data:{url:location.pathname}});
    showLoader();
    if(ready)ws.send(msg);else q.push(msg);
  });
  function collectValue(id,d){
    var el=document.getElementById(id);
    if(!el)return;
    var name=el.getAttribute('name')||id;
    var tag=el.tagName.toLowerCase();
    var type=(el.getAttribute('type')||'').toLowerCase();
    if(type==='radio'){
      var checked=document.querySelector('input[type=radio][name="'+name+'"]:checked');
      d[name]=checked?checked.value:'';
    }else if(type==='checkbox'){
      d[name]=el.checked;
    }else if(tag==='select'){
      d[name]=el.value;
    }else{
      d[name]=el.value;
    }
  }
  // Global Enter-key-to-submit: when Enter is pressed on a text-like
  // <input> (not textarea), find the nearest ancestor container that has
  // a <button> and click it. Works for both FormBuilder and manual forms.
  document.addEventListener('keydown',function(e){
    if(e.key!=='Enter')return;
    var t=e.target;
    if(!t||!t.tagName)return;
    var tag=t.tagName.toLowerCase();
    if(tag!=='input')return;
    var type=(t.getAttribute('type')||'text').toLowerCase();
    // Skip non-text input types
    if(type==='checkbox'||type==='radio'||type==='file'||type==='range'||type==='color'||type==='hidden'||type==='submit'||type==='reset'||type==='button'||type==='image')return;
    // Walk up to find a container with a <button>
    var p=t.parentElement;
    while(p&&p!==document.body){
      var btn=p.querySelector('button');
      if(btn){e.preventDefault();btn.click();return;}
      p=p.parentElement;
    }
  });
  return{
    call:function(act,data,collect){
      var d=Object.assign({},data||{});
      if(collect&&collect.length){
        collect.forEach(function(id){collectValue(id,d)});
      }
      var msg=JSON.stringify({act:act,data:d});
      showLoader();
      if(ready)ws.send(msg);else q.push(msg);
    },
    callSilent:function(act,data){
      var d=Object.assign({},data||{});
      var msg=JSON.stringify({act:act,data:d});
      if(ready)ws.send(msg);else q.push(msg);
    },
    notfound:function(id){
      var msg=JSON.stringify({act:'__notfound',data:{id:id}});
      if(ready)ws.send(msg);else q.push(msg);
    }
  };
})();`

// ---------------------------------------------------------------------------
// WebSocket handler
// ---------------------------------------------------------------------------

// wsMessage is what the client sends.
type wsMessage struct {
	Act  string         `json:"act"`
	Data map[string]any `json:"data"`
}

// cancelConn cancels the push context for a connection and creates a fresh
// one. Any goroutine holding the old context will see ctx.pushCtx.Err() != nil.
func (app *App) cancelConn(ws *websocket.Conn) {
	app.mu.Lock()
	defer app.mu.Unlock()
	if st, ok := app.connStates[ws]; ok {
		st.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	app.connStates[ws] = &connState{ctx: ctx, cancel: cancel}
}

// pushCtxForConn returns the current push context for a connection.
func (app *App) pushCtxForConn(ws *websocket.Conn) context.Context {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if st, ok := app.connStates[ws]; ok {
		return st.ctx
	}
	return context.Background()
}

func (app *App) handleWS(ws *websocket.Conn) {
	// Register client and create initial push context
	pushCtx, pushCancel := context.WithCancel(context.Background())
	app.mu.Lock()
	app.clients[ws] = true
	app.connStates[ws] = &connState{ctx: pushCtx, cancel: pushCancel}
	app.mu.Unlock()

	defer func() {
		app.mu.Lock()
		if st, ok := app.connStates[ws]; ok {
			st.cancel()
			delete(app.connStates, ws)
		}
		delete(app.clients, ws)
		app.mu.Unlock()
		ws.Close()
	}()

	for {
		var raw string
		err := websocket.Message.Receive(ws, &raw)
		if err != nil {
			if err != io.EOF {
				log.Printf("rework ws: read error: %v", err)
			}
			return
		}

		// Parse the incoming message
		var msg wsMessage
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			log.Printf("rework ws: invalid message: %s", raw)
			continue
		}

		// Look up the action handler
		app.mu.RLock()
		handler, ok := app.actions[msg.Act]
		app.mu.RUnlock()

		if !ok {
			errJS := Notify("error", fmt.Sprintf("Unknown action: %s", msg.Act))
			websocket.Message.Send(ws, errJS)
			continue
		}

		// Build context with current push context
		ctx := &Context{
			PathParams: make(map[string]string),
			Query:      make(map[string]string),
			wsConn:     ws,
			wsData:     msg.Data,
			app:        app,
			pushCtx:    app.pushCtxForConn(ws),
		}

		// Execute handler -> get JS string (recover from panics)
		jsResponse := func() (resp string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("rework ws: panic in action %q: %v", msg.Act, r)
					resp = Notify("error", fmt.Sprintf("Server error: %v", r))
				}
			}()
			return handler(ctx)
		}()

		// Send the JS back to the client for immediate execution
		if jsResponse != "" {
			if err := websocket.Message.Send(ws, jsResponse); err != nil {
				log.Printf("rework ws: send error: %v", err)
				return
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

// Context carries request data for both page renders and WS action calls.
type Context struct {
	Request    *http.Request
	Session    map[string]any
	PathParams map[string]string
	Query      map[string]string
	wsConn     *websocket.Conn
	wsData     map[string]any
	app        *App
	pushCtx    context.Context // cancelled when client navigates away or reports element not found
}

// WsData returns the raw WebSocket data map. Useful for passing to
// form validation (FormBuilder.Validate) before deserializing into a struct.
func (ctx *Context) WsData() map[string]any {
	return ctx.wsData
}

func (ctx *Context) Body(target any) error {
	if ctx.wsData == nil {
		return nil
	}
	b, err := json.Marshal(ctx.wsData)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

// Push sends a JS string to THIS client connection immediately.
// Useful for sending additional updates outside the normal request/response.
// Returns an error if the connection's push context has been cancelled
// (e.g. the client navigated away or reported a missing target element).
func (ctx *Context) Push(js string) error {
	if ctx.wsConn == nil {
		return fmt.Errorf("no websocket connection")
	}
	// Check the push context captured at action dispatch time.
	if ctx.pushCtx != nil && ctx.pushCtx.Err() != nil {
		return fmt.Errorf("push cancelled: %w", ctx.pushCtx.Err())
	}
	// Also check the live per-connection context (may have been cancelled
	// after this handler started but before this Push call).
	if ctx.app != nil {
		live := ctx.app.pushCtxForConn(ctx.wsConn)
		if live.Err() != nil {
			return fmt.Errorf("push cancelled: %w", live.Err())
		}
	}
	return websocket.Message.Send(ctx.wsConn, js)
}

// Broadcast sends a JS string to ALL connected WebSocket clients.
func (ctx *Context) Broadcast(js string) {
	ctx.app.mu.RLock()
	defer ctx.app.mu.RUnlock()
	for conn := range ctx.app.clients {
		websocket.Message.Send(conn, js)
	}
}

// Broadcast sends a JS string to ALL connected WebSocket clients.
// Can be called from anywhere (background goroutines, HTTP handlers, etc.)
// without needing a Context.
func (app *App) Broadcast(js string) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	for conn := range app.clients {
		websocket.Message.Send(conn, js)
	}
}

// ---------------------------------------------------------------------------
// Multi-action response builder
// ---------------------------------------------------------------------------

// Response collects multiple JS statements into a single response string.
type Response struct {
	parts []string
}

// NewResponse creates an empty response builder.
func NewResponse() *Response {
	return &Response{}
}

// Add appends a JS string to the response.
func (r *Response) Add(js string) *Response {
	r.parts = append(r.parts, js)
	return r
}

// Replace adds a node replacement operation to the response.
func (r *Response) Replace(targetID string, node *Node) *Response {
	r.parts = append(r.parts, node.ToJSReplace(targetID))
	return r
}

// Render adds a node innerHTML replacement operation to the response.
func (r *Response) Inner(targetID string, node *Node) *Response {
	r.parts = append(r.parts, node.ToJSInner(targetID))
	return r
}

// Append adds a node append operation.
func (r *Response) Append(parentID string, node *Node) *Response {
	r.parts = append(r.parts, node.ToJSAppend(parentID))
	return r
}

// Remove adds an element removal.
func (r *Response) Remove(id string) *Response {
	r.parts = append(r.parts, RemoveEl(id))
	return r
}

// Toast adds a notification.
func (r *Response) Toast(variant, message string) *Response {
	r.parts = append(r.parts, Notify(variant, message))
	return r
}

// Navigate updates the browser URL via pushState without a page reload.
func (r *Response) Navigate(url string) *Response {
	r.parts = append(r.parts, SetLocation(url))
	return r
}

// Back navigates back in browser history (history.back()).
func (r *Response) Back() *Response {
	r.parts = append(r.parts, "history.back();")
	return r
}

// Build joins all parts into a single JS string.
func (r *Response) Build() string {
	return strings.Join(r.parts, "")
}
