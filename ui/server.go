package ui

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// connState tracks per-connection cancellation so that server-side Push
// goroutines stop when the client navigates away or reports a missing
// target element.
type connState struct {
	ctx     context.Context
	cancel  context.CancelFunc
	writeMu sync.Mutex
}

// ---------------------------------------------------------------------------
// App: routes, actions, server
// ---------------------------------------------------------------------------

// LayoutHandler builds a shared layout wrapping every page. The returned
// *Node tree must contain exactly one element with ID("__content__") where
// the page content will be injected.
type LayoutHandler func(ctx *Context) *Node

// App is the top-level application container. It holds page routes (GET)
// and named actions (WS). Pages return a *Node tree that compiles to JS
// for the initial render. Actions return raw JS strings for DOM mutations.
type App struct {
	mu         sync.RWMutex
	actions    map[string]ActionHandler
	clients    map[*websocket.Conn]bool
	connStates map[*websocket.Conn]*connState
	mux        *http.ServeMux
	pageMux    *http.ServeMux
	layout     LayoutHandler
	setupOnce  sync.Once

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
	//
	// This is a trusted raw API: never pass untrusted/user-controlled input to it.
	HTMLHead []string

	// AllowedOrigins adds browser WebSocket origins allowed to connect to /__ws.
	// By default, only same-origin requests are accepted; requests without an
	// Origin header are allowed for non-browser clients. Use "*" to disable
	// origin validation entirely.
	AllowedOrigins []string
}

// PageHandler builds the initial DOM for a GET route.
type PageHandler func(ctx *Context) *Node

// ActionHandler processes a WS action call and returns a JS string
// to execute on the client.
type ActionHandler func(ctx *Context) string

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		actions:    make(map[string]ActionHandler),
		clients:    make(map[*websocket.Conn]bool),
		connStates: make(map[*websocket.Conn]*connState),
		mux:        http.NewServeMux(),
		pageMux:    http.NewServeMux(),
	}
}

// Page registers a GET route using Go's http.ServeMux pattern syntax. Named
// wildcards are available through ctx.Request.PathValue and ctx.PathParams.
// The handler returns a *Node tree which is compiled to JS and served inside
// the standard HTML shell.
//
//	app.Page("/dp/{token}", func(ctx *ui.Context) *ui.Node {
//		token := ctx.Request.PathValue("token")
//		return ui.Div().Text(token)
//	})
func (app *App) Page(pattern string, handler PageHandler) {
	serveMuxPattern := pattern
	// Preserve the historical exact-match behavior of static routes ending in
	// a slash. ServeMux otherwise treats them as subtree routes. Callers that
	// want a subtree can register an explicit {name...} wildcard.
	if strings.HasSuffix(serveMuxPattern, "/") && !strings.Contains(serveMuxPattern, "{") {
		serveMuxPattern += "{$}"
	}
	app.pageMux.Handle("GET "+serveMuxPattern, pageRoute{app: app, handler: handler})
}

// CSS registers external stylesheets and/or inline CSS rules that apply
// to every page in the application. The tags are injected into the HTML
// <head> server-side, so they load immediately without JavaScript.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
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
	app.mux.HandleFunc("GET "+path, handler)
}

// POST registers a standard HTTP POST handler on the internal mux.
func (app *App) POST(path string, handler http.HandlerFunc) {
	app.mux.HandleFunc("POST "+path, handler)
}

// DELETE registers a standard HTTP DELETE handler on the internal mux.
func (app *App) DELETE(path string, handler http.HandlerFunc) {
	app.mux.HandleFunc("DELETE "+path, handler)
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
	app.setupOnce.Do(app.setupRoutes)
}

func (app *App) setupRoutes() {
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

		u, err := url.Parse(req.URL)
		if err != nil || req.URL == "" {
			return ""
		}

		// Patch the request so page handlers see the navigated URL instead of
		// the WebSocket upgrade path (/__ws). Server-side logic that keys on
		// the current path (active nav state, breadcrumbs) depends on this.
		if ctx.Request != nil {
			r2 := ctx.Request.Clone(ctx.Request.Context())
			r2.URL.Path = u.Path
			r2.URL.RawQuery = u.RawQuery
			ctx.Request = r2
		}
		ctx.Query = make(map[string]string)
		for k, v := range u.Query() {
			if len(v) > 0 {
				ctx.Query[k] = v[0]
			}
		}

		handler, matchedRequest, ok := app.matchPage(ctx.Request)

		if !ok {
			return ""
		}
		ctx.Request = matchedRequest
		ctx.PathParams = requestPathParams(matchedRequest)

		app.mu.RLock()
		layoutFn := app.layout
		app.mu.RUnlock()

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
	app.mux.Handle("/__ws", websocket.Server{Handshake: app.wsHandshake, Handler: app.handleWS})

	// Page routes (catch-all). The nested mux provides ServeMux pattern
	// matching and populates Request.PathValue for Page handlers.
	app.mux.Handle("/", app.pageMux)
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

// pageRoute renders normally for HTTP requests. matchPage adds pageMatch to
// the request context to resolve a route without rendering it, which is used
// by WebSocket navigation.
type pageRoute struct {
	app     *App
	handler PageHandler
}

func (route pageRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if match, ok := r.Context().Value(pageMatchContextKey{}).(*pageMatch); ok {
		match.handler = route.handler
		match.request = r.WithContext(match.requestContext)
		return
	}
	route.app.renderPage(w, r, route.handler)
}

type pageMatchContextKey struct{}

type pageMatch struct {
	handler        PageHandler
	request        *http.Request
	requestContext context.Context
}

// discardResponseWriter absorbs redirects and not-found responses produced by
// ServeMux while matchPage probes for a matching Page route.
type discardResponseWriter struct {
	header http.Header
}

func (w *discardResponseWriter) Header() http.Header         { return w.header }
func (w *discardResponseWriter) WriteHeader(_ int)           {}
func (w *discardResponseWriter) Write(p []byte) (int, error) { return len(p), nil }

func (app *App) matchPage(r *http.Request) (PageHandler, *http.Request, bool) {
	if r == nil {
		return nil, nil, false
	}
	match := &pageMatch{requestContext: r.Context()}
	r2 := r.WithContext(context.WithValue(r.Context(), pageMatchContextKey{}, match))
	app.pageMux.ServeHTTP(&discardResponseWriter{header: make(http.Header)}, r2)
	if match.handler == nil || match.request == nil {
		return nil, nil, false
	}
	return match.handler, match.request, true
}

func requestPathParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	if r == nil || r.Pattern == "" {
		return params
	}
	pattern := r.Pattern
	if i := strings.IndexByte(pattern, ' '); i >= 0 {
		pattern = pattern[i+1:]
	}
	if i := strings.IndexByte(pattern, '/'); i >= 0 {
		pattern = pattern[i:]
	}
	for _, segment := range strings.Split(pattern, "/") {
		if len(segment) < 3 || segment[0] != '{' || segment[len(segment)-1] != '}' {
			continue
		}
		name := strings.TrimSuffix(segment[1:len(segment)-1], "...")
		if name != "" && name != "$" {
			params[name] = r.PathValue(name)
		}
	}
	return params
}

func (app *App) renderPage(w http.ResponseWriter, r *http.Request, handler PageHandler) {

	ctx := &Context{
		Request:    r,
		PathParams: requestPathParams(r),
		Query:      make(map[string]string),
		app:        app,
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
		faviconTag = fmt.Sprintf(`<link rel="icon" href="%s">`, html.EscapeString(app.Favicon))
	}
	titleTag := "<title>App</title>"
	if app.Title != "" {
		titleTag = fmt.Sprintf(`<title>%s</title>`, html.EscapeString(app.Title))
	}
	descTag := ""
	if app.Description != "" {
		descTag = fmt.Sprintf(`<meta name="description" content="%s">`, html.EscapeString(app.Description))
	}

	// Build custom head HTML (app-wide + per-page)
	customHead := strings.Join(app.HTMLHead, "\n")
	if pageCSS := ctx.cssHeadHTML(); pageCSS != "" {
		customHead += "\n" + pageCSS
	}
	if pageJS := ctx.jsHeadHTML(); pageJS != "" {
		customHead += "\n" + pageJS
	}

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
<html lang="en" class="gsui-booting">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
%s
%s
%s
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<script>%s
%s</script>

<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4" data-gsui-style-engine async onload="this.dataset.gsuiLoaded='true'" onerror="this.dataset.gsuiLoaded='error'"></script>
<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons+Round" media="print" onload="this.media='all';this.dataset.gsuiLoaded='true'" onerror="this.dataset.gsuiLoaded='error'">
<noscript><link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons+Round"></noscript>
<style type="text/tailwindcss">
@custom-variant dark (&:where(.dark, .dark *));
@theme{--font-sans:ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif;}
</style>
<style>%s</style>
%s
<script src="/__ws.js?v=%s" defer></script>
<style>%s</style>
<script>%s</script>
</head>
<body>
<script>
%s
</script>
</body>
</html>`, faviconTag, titleTag, descTag, themeInitJS, wsStubJS, darkOverrideCSS, customHead, wsClientVersion, loadingCSS, bootInitJS, jsBody)
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

// wsClientVersion is a content hash of the embedded client script, appended
// to the /__ws.js URL so browsers refetch it when the client code changes
// despite the long-lived Cache-Control header.
var wsClientVersion = func() string {
	h := sha256.Sum256([]byte(wsClientJS))
	return hex.EncodeToString(h[:4])
}()

// wsStubJS installs a queuing stub for __ws synchronously in <head>. The real
// client (/__ws.js) loads with defer and therefore runs AFTER the inline body
// script, so page-load JS (Node.JS blocks, ctx.HeadJS) that calls __ws would
// otherwise hit "__ws is not defined". The stub queues those calls; the real
// client replays and replaces it when it initializes.
const wsStubJS = `window.__ws||(window.__ws={__q:[],call:function(){this.__q.push(['call',arguments])},callSilent:function(){this.__q.push(['callSilent',arguments])},notfound:function(){this.__q.push(['notfound',arguments])}});`

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

// bootInitJS keeps the application hidden until its initial DOM, stylesheets,
// Tailwind-generated CSS, and active web fonts are ready to paint. Waiting on
// these resources instead of window.load avoids holding the UI for images. The
// timeout is a fail-safe for stalled third-party resources.
const bootInitJS = `(function(){
if(window.__gsuiBootInit)return;window.__gsuiBootInit=true;
var d=document.documentElement,done=false,preparing=false,pending=0,domReady=document.readyState!=='loading',timer;
d.classList.add('gsui-booting');
var scheme=d.style.colorScheme;try{if(!scheme)scheme=getComputedStyle(d).colorScheme}catch(_){}
if(d.classList.contains('dark')||(scheme!=='light'&&window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches))d.classList.add('gsui-loading-dark');
function paintReveal(){
 if(done)return;done=true;clearTimeout(timer);
 requestAnimationFrame(function(){requestAnimationFrame(function(){
  d.classList.remove('gsui-booting','gsui-loading-dark');d.classList.add('gsui-ready','gsui-revealing');
  setTimeout(function(){d.classList.remove('gsui-revealing')},180);
  try{window.dispatchEvent(new Event('gsui:ready'))}catch(_){}
 })});
}
function prepareReveal(){
 if(preparing||done)return;preparing=true;
 if(document.fonts&&document.fonts.ready){document.fonts.ready.then(paintReveal,paintReveal)}else{paintReveal()}
}
function check(){if(domReady&&pending===0)prepareReveal()}
function waitFor(el){
 var fetched=el.tagName==='SCRIPT'&&el.src&&window.performance&&performance.getEntriesByName&&performance.getEntriesByName(el.src).length>0;
 if(el.dataset.gsuiLoaded==='true'||el.dataset.gsuiLoaded==='error'||(el.tagName==='LINK'&&el.sheet)||fetched)return;
 pending++;
 var settled=false;
 function settle(){if(settled)return;settled=true;pending--;check()}
 el.addEventListener('load',settle,{once:true});el.addEventListener('error',settle,{once:true});
}
document.querySelectorAll('link[rel~="stylesheet"],script[data-gsui-style-engine],script[src*="@tailwindcss/browser"]').forEach(waitFor);
if(!domReady){document.addEventListener('DOMContentLoaded',function(){domReady=true;check()},{once:true})}
check();
timer=setTimeout(paintReveal,4000);
})();`

const loadingCSS = `html{background-color:var(--gsui-loading-bg,#fff)}
html.dark,html.gsui-loading-dark{color-scheme:dark;background-color:var(--gsui-loading-bg-dark,#0b1120)}
html.gsui-booting body{visibility:hidden;opacity:0}
html.gsui-revealing body{visibility:visible;transition:opacity 160ms cubic-bezier(.22,1,.36,1)}
@media (prefers-reduced-motion:reduce){html.gsui-revealing body{transition:none}}`

// darkOverrideCSS provides fallback dark mode overrides for elements that
// may not carry explicit dark: Tailwind classes (e.g. plain body, form UA styles).
// These use .dark selectors without !important so explicit dark: variants win.
// Also applies a local font stack to avoid late webfont swaps and CLS.
const darkOverrideCSS = `.dark body{color:#e5e7eb;background-color:#0b1120}
.dark input::placeholder,.dark textarea::placeholder{color:#9ca3af}
body{font-family:ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif}`

// wsClientJS is the entire client-side framework.
// It connects to the WS endpoint, sends action calls, executes
// whatever JS string the server sends back, and shows an offline
// overlay when the WebSocket disconnects.
const wsClientJS = `var __wsPre=window.__ws;
var __offline=(function(){
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
  var ws,q=[],ready=false,seq=0,inflight={},loaderEl=null,loaderTimer=0,hadClose=false,backoff=500;
  function showLoader(){
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
    if(Object.keys(inflight).length)return;
    if(loaderTimer){clearTimeout(loaderTimer);loaderTimer=0;}
    if(loaderEl){loaderEl.style.opacity='0';var el=loaderEl;loaderEl=null;setTimeout(function(){try{if(el&&el.parentNode)el.parentNode.removeChild(el)}catch(_){}},160);}
  }
  function connect(){
    ws=new WebSocket((location.protocol==='https:'?'wss://':'ws://')+location.host+'/__ws');
    ws.onopen=function(){
      ready=true;backoff=500;
      __offline.hide();
      q.forEach(function(m){ws.send(m)});q=[];
      if(hadClose){hadClose=false;try{location.reload();return;}catch(_){}}
    };
    ws.onmessage=function(e){var m;try{m=JSON.parse(e.data)}catch(_){}if(m&&typeof m==='object'&&m.__r){if(inflight[m.id]){delete inflight[m.id];hideLoader()}if(m.js){try{new Function(m.js)()}catch(err){console.error('ws exec error:',err,m.js)}}document.querySelectorAll('button.gsui-busy').forEach(function(b){b.disabled=false;b.classList.remove('gsui-busy','opacity-60','cursor-wait')})}else{try{new Function(e.data)()}catch(err){console.error('ws exec error:',err,e.data)}}try{window.dispatchEvent(new Event('gsui:updated'))}catch(_){}};
    ws.onclose=function(){ready=false;inflight={};hideLoader();document.querySelectorAll('button.gsui-busy').forEach(function(b){b.disabled=false;b.classList.remove('gsui-busy','opacity-60','cursor-wait')});__offline.show();hadClose=true;var d=Math.min(10000,backoff)*(0.75+Math.random()*0.5);backoff=Math.min(10000,backoff*2);setTimeout(connect,d)};
    ws.onerror=function(){ws.close()};
  }
  connect();
  window.addEventListener('popstate',function(){
    var id=++seq,msg=JSON.stringify({act:'__nav',data:{url:location.pathname+location.search},id:id});
    inflight[id]=true;
    showLoader();
    if(ready)ws.send(msg);else queue(msg);
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
    var form=t.form||t.closest('form'),btn;
    if(form){btn=form.querySelector('button[type=submit]')||form.querySelector('button:not([type])')||form.querySelector('button');if(!btn&&form.id)btn=document.querySelector('button[form="'+form.id+'"]')}else{var p=t.parentElement;while(p&&p!==document.body){btn=p.querySelector('button');if(btn)break;p=p.parentElement}}
    if(btn){e.preventDefault();btn.click()}
  });
  function queue(msg){if(q.length>=100){console.warn('gsui: WebSocket queue full; dropping message');return}q.push(msg)}
  return{
    call:function(act,data,collect){
      var d=Object.assign({},data||{});
      if(collect&&collect.length){
        collect.forEach(function(id){collectValue(id,d)});
      }
      var id=++seq,msg=JSON.stringify({act:act,data:d,id:id});
      inflight[id]=true;
      showLoader();
      if(ready)ws.send(msg);else queue(msg);
    },
    callSilent:function(act,data){
      var d=Object.assign({},data||{});
      var msg=JSON.stringify({act:act,data:d});
      if(ready)ws.send(msg);else queue(msg);
    },
    notfound:function(id){
      var msg=JSON.stringify({act:'__notfound',data:{id:id}});
      if(ready)ws.send(msg);else queue(msg);
    }
  };
})();
if(__wsPre&&__wsPre.__q){__wsPre.__q.forEach(function(it){try{__ws[it[0]].apply(__ws,it[1])}catch(e){console.error('gsui: queued ws call failed:',e)}})}`

// ---------------------------------------------------------------------------
// WebSocket handler
// ---------------------------------------------------------------------------

// wsMessage is what the client sends.
type wsMessage struct {
	Act  string         `json:"act"`
	Data map[string]any `json:"data"`
	ID   int64          `json:"id"`
}

// wsHandshake accepts same-origin browser requests, configured origins, and
// non-browser clients that do not send Origin.
func (app *App) wsHandshake(_ *websocket.Config, r *http.Request) error {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return nil
	}
	app.mu.RLock()
	allowed := append([]string(nil), app.AllowedOrigins...)
	app.mu.RUnlock()
	for _, v := range allowed {
		if v == "*" || v == origin {
			return nil
		}
	}
	u, err := url.Parse(origin)
	if err == nil && u.Scheme != "" && strings.EqualFold(u.Host, r.Host) {
		return nil
	}
	return fmt.Errorf("origin %q is not allowed", origin)
}

// send serializes outbound frames for a connection shared by handler replies,
// Push calls, and broadcasts.
func (app *App) send(ws *websocket.Conn, s string) error {
	app.mu.RLock()
	st, ok := app.connStates[ws]
	app.mu.RUnlock()
	if !ok {
		return fmt.Errorf("connection is closed")
	}
	st.writeMu.Lock()
	defer st.writeMu.Unlock()
	return websocket.Message.Send(ws, s)
}

func wsReply(id int64, js string) string {
	b, err := json.Marshal(struct {
		Reply int64  `json:"__r"`
		ID    int64  `json:"id"`
		JS    string `json:"js"`
	}{Reply: 1, ID: id, JS: js})
	if err != nil {
		log.Printf("gsui: marshal WebSocket reply: %v", err)
		return `{"__r":1,"id":0,"js":""}`
	}
	return string(b)
}

// cancelConn cancels the push context for a connection and creates a fresh
// one. Any goroutine holding the old context will see ctx.pushCtx.Err() != nil.
func (app *App) cancelConn(ws *websocket.Conn) {
	app.mu.Lock()
	defer app.mu.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	if st, ok := app.connStates[ws]; ok {
		st.cancel()
		st.ctx = ctx
		st.cancel = cancel
		return
	}
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
				log.Printf("gsui: ws read error: %v", err)
			}
			return
		}

		// Parse the incoming message
		var msg wsMessage
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			log.Printf("gsui: invalid WebSocket message: %s", raw)
			continue
		}

		// Look up the action handler
		app.mu.RLock()
		handler, ok := app.actions[msg.Act]
		app.mu.RUnlock()

		if !ok {
			errJS := Notify("error", fmt.Sprintf("Unknown action: %s", msg.Act))
			if msg.ID != 0 {
				errJS = wsReply(msg.ID, errJS)
			}
			if err := app.send(ws, errJS); err != nil {
				log.Printf("gsui: ws send error: %v", err)
				return
			}
			continue
		}

		// Build context with current push context
		ctx := &Context{
			Request:    ws.Request(),
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
					log.Printf("gsui: panic in action %q: %v\n%s", msg.Act, r, debug.Stack())
					resp = Notify("error", "Server error")
				}
			}()
			return handler(ctx)
		}()

		// Prepend any per-page CSS/JS injection from ctx.HeadCSS()/ctx.HeadJS()
		var prefix string
		if cssJS := ctx.cssInjectJS(); cssJS != "" {
			prefix += cssJS
		}
		if jsJS := ctx.jsInjectJS(); jsJS != "" {
			prefix += jsJS
		}
		if prefix != "" && jsResponse != "" {
			jsResponse = prefix + "\n" + jsResponse
		} else if prefix != "" {
			jsResponse = prefix
		}

		// Tracked requests always receive an envelope, including empty JS.
		if msg.ID != 0 {
			jsResponse = wsReply(msg.ID, jsResponse)
		}
		if jsResponse != "" {
			if err := app.send(ws, jsResponse); err != nil {
				log.Printf("gsui: ws send error: %v", err)
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
	headCSS    []string        // per-page <style>/<link> tags collected via ctx.HeadCSS()
	headJS     []string        // per-page <script> blocks collected via ctx.HeadJS()
}

// WsData returns the raw WebSocket data map. Useful for passing to
// form validation (FormBuilder.Validate) before deserializing into a struct.
func (ctx *Context) WsData() map[string]any {
	return ctx.wsData
}

// HeadCSS registers external stylesheets and/or inline CSS rules for the
// current page. On a full page load the tags are injected into the HTML
// <head> server-side (instant, no JS needed). On SPA navigations (WS
// actions) the same resources are injected into <head> via JS with
// deduplication so they are not loaded twice.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
//
//	ctx.HeadCSS(
//	    []string{"https://fonts.googleapis.com/css2?family=Oswald&display=swap"},
//	    `.hero { font-family: 'Oswald', sans-serif; }`,
//	)
//
// Pass nil for urls if you only need inline CSS, or "" for css if you
// only need external links.
func (ctx *Context) HeadCSS(urls []string, css string) {
	for _, u := range urls {
		ctx.headCSS = append(ctx.headCSS,
			fmt.Sprintf(`<link rel="stylesheet" href="%s">`, u))
	}
	if css != "" {
		ctx.headCSS = append(ctx.headCSS,
			fmt.Sprintf("<style>%s</style>", css))
	}
}

// HeadJS registers a JavaScript block that runs once when the page loads.
// On a full page load the script is emitted as a <script> tag in <head>.
// On SPA navigations the code is prepended to the WS response so it
// executes before the DOM swap.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
//
// Use this for page-level setup (global functions, event listeners, etc.)
// instead of the Div("").JS(`...`) pattern.
//
//	ctx.HeadJS(`
//	    window.toggleMobileNav = function() {
//	        var nav = document.getElementById('mobile-nav');
//	        if (nav) nav.classList.toggle('hidden');
//	    };
//	`)
func (ctx *Context) HeadJS(code string) {
	if code != "" {
		ctx.headJS = append(ctx.headJS, code)
	}
}

// cssHeadHTML returns the collected per-page CSS as raw HTML for <head>.
func (ctx *Context) cssHeadHTML() string {
	return strings.Join(ctx.headCSS, "\n")
}

// jsHeadHTML returns the collected per-page JS as a <script> block for <head>.
func (ctx *Context) jsHeadHTML() string {
	if len(ctx.headJS) == 0 {
		return ""
	}
	return "<script>" + strings.Join(ctx.headJS, "\n") + "</script>"
}

// cssInjectJS returns JS code that injects the per-page CSS into <head>
// at runtime (used during SPA/WS navigations). External links are
// deduplicated by href.
func (ctx *Context) cssInjectJS() string {
	if len(ctx.headCSS) == 0 {
		return ""
	}
	var js strings.Builder
	for _, tag := range ctx.headCSS {
		if strings.HasPrefix(tag, "<link") {
			// Extract href from <link rel="stylesheet" href="...">
			start := strings.Index(tag, `href="`)
			if start < 0 {
				continue
			}
			start += 6
			end := strings.Index(tag[start:], `"`)
			if end < 0 {
				continue
			}
			href := tag[start : start+end]
			eu := escJS(href)
			fmt.Fprintf(&js,
				"if(!document.querySelector('link[href=\\'%s\\']')){"+
					"var l=document.createElement('link');"+
					"l.rel='stylesheet';l.href='%s';"+
					"document.head.appendChild(l);}",
				eu, eu,
			)
		} else if after, ok := strings.CutPrefix(tag, "<style>"); ok {
			// Extract CSS content between <style> and </style>
			inner := after
			inner = strings.TrimSuffix(inner, "</style>")
			fmt.Fprintf(&js,
				"var _s=document.createElement('style');"+
					"_s.textContent='%s';"+
					"document.head.appendChild(_s);",
				escJS(inner),
			)
		}
	}
	return js.String()
}

// jsInjectJS returns the collected per-page JS as raw code to prepend
// to a WS response (used during SPA navigations).
func (ctx *Context) jsInjectJS() string {
	return strings.Join(ctx.headJS, "\n")
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
	if ctx.app == nil {
		return fmt.Errorf("no app context")
	}
	return ctx.app.send(ctx.wsConn, js)
}

// Broadcast sends a JS string to ALL connected WebSocket clients.
func (ctx *Context) Broadcast(js string) {
	if ctx.app != nil {
		ctx.app.Broadcast(js)
	}
}

// Broadcast sends a JS string to ALL connected WebSocket clients.
// Can be called from anywhere (background goroutines, HTTP handlers, etc.)
// without needing a Context.
func (app *App) Broadcast(js string) {
	app.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(app.clients))
	for conn := range app.clients {
		clients = append(clients, conn)
	}
	app.mu.RUnlock()
	for _, conn := range clients {
		if err := app.send(conn, js); err != nil {
			log.Printf("gsui: broadcast send error: %v", err)
		}
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
