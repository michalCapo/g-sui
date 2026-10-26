package ui

import (
	"cmp"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"net/url"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// connState tracks per-connection cancellation so that server-side Push
// goroutines stop when the client navigates away or reports a missing
// target element.
type connState struct {
	ctx              context.Context
	connectionCtx    context.Context
	connectionCancel context.CancelFunc
	cancel           context.CancelFunc
	writeMu          sync.Mutex
	request          *http.Request
	version          int64
	view             View
	viewPattern      string
	session          map[string]any
	subscriptions    map[string]subscription
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
// for the initial render. Typed actions return Result effects.
type App struct {
	mu                  sync.RWMutex
	actions             map[string]actionHandler
	subscriptionActions map[string]bool
	clients             map[*socket]bool
	connStates          map[*socket]*connState
	mux                 *http.ServeMux
	pageMux             *http.ServeMux
	layout              LayoutHandler
	setupOnce           sync.Once
	middleware          []func(http.Handler) http.Handler
	widgets             map[string]string
	pageHandler         http.Handler
	// Authorize runs for initial pages, live navigation and every user action.
	// An empty action name denotes a page render. Check object permissions in
	// action handlers as well. Configure before serving requests.
	// Return Deny(result) to send custom effects when rejecting a WS action.
	Authorize func(*Context, string) error
	// Identity resolves the current user on every request/action. Cookies on an
	// existing socket are the handshake cookies; validate expiry in this hook.
	Identity func(*http.Request) (any, error)

	// MaxSubscriptions limits subscription IDs retained per connection.
	// Values <= 0 use the default of 64. Configure before serving requests.
	// New IDs are rejected at the limit; existing IDs may be restarted.
	MaxSubscriptions int

	// MessageLimit returns the inbound WebSocket message limit for a
	// connection from its handshake request. Nil, or values <= 0, use the
	// default of 1 MiB. Grant larger limits only to authenticated requests.
	// A larger message closes the connection; the client reports it as lost.
	// Configure before serving requests.
	MessageLimit func(*http.Request) int64

	// Favicon is the URL path for the site favicon (e.g. "/assets/favicon.svg").
	// When set, a <link rel="icon"> tag is emitted in the HTML shell.
	Favicon string

	// Lang sets the page language, <html lang>. Default "en".
	Lang string

	// Texts replaces the English texts the client shows by itself:
	// connection badges, connection errors and the loading badge. Empty
	// fields keep English.
	Texts Texts

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

	// OfflineGraceMs is how long reconnecting may keep failing before the
	// "Offline" indicator appears. It shows at once when the browser reports
	// no network. Short blips (tab wake-ups, proxy hiccups, congested links
	// during a large upload) reconnect within the grace window. A page coming
	// back from the background starts a fresh wait, so a socket that died
	// while hidden never flashes "Offline". Default 12000. Use -1 to never
	// show the indicator.
	OfflineGraceMs int

	// ConnectingGraceMs is how long a reconnect may take before a quiet
	// "Connecting…" badge appears. Default 1500. Use -1 to never show it.
	// The client emits a "gsui:connection" event with detail.state
	// "connecting", "connected" or "offline", and sets the same value as
	// the data-gsui-connection attribute of <html>.
	ConnectingGraceMs int

	// HeartbeatTimeoutMs is how long the client waits for the answer to a
	// keep-alive heartbeat before it replaces the socket as dead. The server
	// answers heartbeats before queued actions, but after a frame it is
	// writing, which may take up to its 10s write limit on a slow link; a
	// page shown again waits 4s. Default 12000.
	HeartbeatTimeoutMs int

	// HandshakeTimeoutMs gives up a connection attempt that has not been
	// greeted by the server in time and tries again; each timed out attempt
	// waits twice as long, up to 30s. Default 8000. Use -1 to wait for the
	// browser.
	HandshakeTimeoutMs int

	// ReconnectReloadAfterMs controls the page reload that happens after the
	// socket comes back. The reload only runs when the outage lasted at least
	// this long AND no client-side hold is active (see __ws.hold()). Shorter
	// outages resync silently: subscriptions restart and a
	// "gsui:reconnected" event is dispatched so the page can refresh itself
	// without losing in-page state such as an upload in progress.
	// Default 15000. Use -1 to never reload.
	ReconnectReloadAfterMs int

	// KeepAliveMs is the interval of the client-side ping that keeps idle
	// WebSockets alive through proxies and load balancers that close silent
	// connections (the common cause of "offline" during long uploads).
	// Default 25000. Use -1 to disable.
	KeepAliveMs int

	// InstanceID identifies the process rendering pages. Element ids
	// (ui.Target()) are random per process, so a page rendered by one process
	// cannot be patched by another: the client compares this value with the one
	// announced on every connection and reloads when it changes, which is what
	// makes a server restart recover instead of leaving a dead-looking UI.
	//
	// Defaults to a random per-process id, which is correct for a single
	// server. Run several replicas behind a load balancer and every reconnect
	// to a different replica would reload; set the same value (a build or
	// release id) on all of them so only a real deploy triggers a reload.
	InstanceID string

	// Loader sets the overlay shown while an action runs. LoaderOverlay (the
	// default) blurs the page after 120ms; LoaderNone never creates it.
	// Use Action.NoLoader to skip it for one call.
	Loader LoaderMode

	// SkeletonDelayMs is how long a page navigation may take before its
	// skeleton (see App.Skeleton) appears. Default 100. Use -1 to show it at once.
	SkeletonDelayMs int

	// SkeletonTarget is the id of the element that shows navigation
	// skeletons. Default: #__content__, else <main>, else <body>.
	SkeletonTarget string

	// AllowedOrigins adds browser WebSocket origins allowed to connect to /__ws.
	// By default, only same-origin requests are accepted; requests without an
	// Origin header are allowed for non-browser clients. Use "*" to disable
	// origin validation entirely.
	AllowedOrigins []string

	skeletons []skeletonRoute

	assets      map[string]*assetFile
	tailwindURL string
	iconFontURL string
}

// LoaderMode selects the loading overlay. See App.Loader.
type LoaderMode string

const (
	// LoaderOverlay blurs the page while an action runs. It is the default.
	LoaderOverlay LoaderMode = "overlay"
	// LoaderNone shows no overlay. Request tracking, reconnect and the
	// offline badge keep working.
	LoaderNone LoaderMode = "none"
)

type skeletonRoute struct {
	prefix string
	js     string // JS function that builds the skeleton and returns it
}

// PageHandler builds the initial DOM for a GET route.
type PageHandler func(ctx *Context) *Node

// actionHandler processes a WS action call and returns a JS string
// to execute on the client.
type actionHandler func(ctx *Context) string

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{
		actions:             make(map[string]actionHandler),
		subscriptionActions: make(map[string]bool),
		clients:             make(map[*socket]bool),
		connStates:          make(map[*socket]*connState),
		mux:                 http.NewServeMux(),
		pageMux:             http.NewServeMux(),
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

// action registers an internal dispatcher. Applications use RegisterAction.
func (app *App) action(name string, handler actionHandler) {
	app.mu.Lock()
	app.actions[name] = handler
	delete(app.subscriptionActions, name)
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
	files := http.FileServerFS(sub)
	var tags sync.Map // name -> ETag of files without a modification time
	app.mux.Handle(prefix, http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Embedded files have no modification time, so the browser cannot
		// revalidate them. A content ETag lets it get 304 instead of the file.
		name := strings.TrimPrefix(r.URL.Path, "/")
		if info, err := fs.Stat(sub, name); err == nil && !info.IsDir() && info.ModTime().IsZero() {
			tag, ok := tags.Load(name)
			if !ok {
				if b, err := fs.ReadFile(sub, name); err == nil {
					sum := sha256.Sum256(b)
					tag, _ = tags.LoadOrStore(name, `"`+hex.EncodeToString(sum[:12])+`"`)
				}
			}
			if tag != nil {
				w.Header().Set("ETag", tag.(string))
				w.Header().Set("Cache-Control", "no-cache")
			}
		}
		files.ServeHTTP(w, r)
	})))
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
	app.pageHandler = app.pageMux
	for _, v := range slices.Backward(app.middleware) {
		app.pageHandler = v(app.pageHandler)
	}
	// Built-in __nav action: handles popstate (browser back/forward).
	// Looks up the page handler for the URL and replaces content.
	// If a layout is registered, only the __content__ container is replaced
	// (the layout shell stays). Otherwise the full body is cleared and rebuilt.
	// Also cancels any outstanding Push goroutines for the connection since
	// their target elements no longer exist after navigation.
	app.action("__nav", app.navigate)
	app.action("__live", app.liveEvent)
	app.action("__unsubscribe", func(ctx *Context) string {
		var req struct {
			Name string `json:"name"`
			Key  string `json:"key"`
		}
		if ctx.body(&req) == nil {
			app.cancelSubscription(ctx.wsConn, req.Name, req.Key)
		}
		return ""
	})

	// Built-in __notfound action: the client sends this when a WS patch
	// targets a DOM element that no longer exists. Cancel push context so
	// server-side goroutines calling ctx.Push() get an error and stop.
	app.action("__notfound", func(ctx *Context) string {
		var req struct {
			Subscription string `json:"subscription"`
		}
		if ctx.body(&req) == nil && req.Subscription != "" {
			app.cancelSubscription(ctx.wsConn, "", req.Subscription)
		}
		return ""
	})

	// Built-in __ping action: keep-alive traffic from the client so idle
	// connections are not dropped by proxies. Sent without an id, so the
	// empty reply carries no DOM changes.
	app.action("__ping", func(_ *Context) string { return "" })

	// Serve the tiny WS client script
	app.mux.HandleFunc("GET /__ws.js", app.serveWSClient)
	app.mux.HandleFunc("GET "+assetPrefix, app.serveAsset)

	// WebSocket endpoint
	app.mux.HandleFunc("/__ws", app.acceptWS)

	// Page routes (catch-all). The nested mux provides ServeMux pattern
	// matching and populates Request.PathValue for Page handlers.
	app.mux.Handle("/", app.pageHandler)
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
		match.request = r
		return
	}
	route.app.renderPage(w, r, route.handler)
}

type pageMatchContextKey struct{}

type pageMatch struct {
	handler PageHandler
	request *http.Request
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
	match := &pageMatch{}
	r2 := r.WithContext(context.WithValue(r.Context(), pageMatchContextKey{}, match))
	h := app.pageHandler
	if h == nil {
		h = app.pageMux
	}
	h.ServeHTTP(&discardResponseWriter{header: make(http.Header)}, r2)
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
	for segment := range strings.SplitSeq(pattern, "/") {
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
	ctx.Session = make(map[string]any)
	ctx.setRequest(r)
	if err := app.authorize(ctx, ""); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
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
	jsBody := root.toJS()

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
	if acceptsGzip(r) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()
		writer = gz
	}

	_, _ = fmt.Fprintf(writer, `<!DOCTYPE html>
<html lang="%s" class="gsui-booting">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
%s
%s
%s
<script>%s
%s
%s
%s
%s</script>
%s<style>%s</style>
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
</html>`, html.EscapeString(cmp.Or(app.Lang, "en")), faviconTag, titleTag, descTag, themeInitJS, wsStubJS, app.wsConfigJS()+app.Texts.js(), localJS, app.widgetJS(), app.styleHeadHTML(), darkOverrideCSS, customHead, wsClientVersion, loadingCSS, bootInitJS, jsBody)
}

// ---------------------------------------------------------------------------
// WebSocket client script
// ---------------------------------------------------------------------------

func (app *App) serveWSClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=86400")

	// Gzip compress if supported
	if acceptsGzip(r) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()
		_, _ = gz.Write([]byte(wsClientJS))
		return
	}
	_, _ = w.Write([]byte(wsClientJS))
}

// wsClientVersion is a content hash of the embedded client script, appended
// to the /__ws.js URL so browsers refetch it when the client code changes
// despite the long-lived Cache-Control header.
var wsClientVersion = func() string {
	h := sha256.Sum256([]byte(wsClientJS))
	return hex.EncodeToString(h[:4])
}()

// serverInstanceID identifies this server process. Target() IDs are random per
// process, so a DOM rendered by one process cannot be patched by another: after
// a restart every generated id in the loaded page is unknown to the new
// process and its responses silently patch nothing. The client compares the id
// baked into the page with the one announced on each connection and reloads
// when they differ, which is the only way to resync a stale DOM.
var serverInstanceID = Target()

// instance returns the configured instance id, or this process's random one.
func (app *App) instance() string {
	app.mu.RLock()
	id := app.InstanceID
	app.mu.RUnlock()
	if id == "" {
		return serverInstanceID
	}
	return id
}

// helloFrame is the first frame sent on every connection. It is JSON so the
// client can tell it apart from executable JS.
func (app *App) helloFrame() string {
	return fmt.Sprintf(`{"__hello":%q}`, app.instance())
}

// wsConfigJS emits the connection-resilience settings read by the client
// script. It is inlined in <head> (the client script itself stays static and
// long-cached). Zero values fall back to the documented defaults.
func (app *App) wsConfigJS() string {
	app.mu.RLock()
	grace, reloadAfter, keepAlive := app.OfflineGraceMs, app.ReconnectReloadAfterMs, app.KeepAliveMs
	connecting, hbTimeout, handshake := app.ConnectingGraceMs, app.HeartbeatTimeoutMs, app.HandshakeTimeoutMs
	loader, skelDelay, skelTarget := app.Loader, app.SkeletonDelayMs, app.SkeletonTarget
	skeletons := app.skeletons
	app.mu.RUnlock()

	if grace == 0 {
		grace = 12000
	}
	if reloadAfter == 0 {
		reloadAfter = 15000
	}
	if keepAlive == 0 {
		keepAlive = 25000
	}
	var b strings.Builder
	fmt.Fprintf(&b, `window.__gsuiCfg={grace:%d,reloadAfter:%d,keepAlive:%d,inst:'%s'`, grace, reloadAfter, keepAlive, escJS(app.instance()))
	// Defaults live in the client; only set values are sent.
	for _, opt := range []struct {
		name  string
		value int
	}{{"connectingGrace", connecting}, {"hbTimeout", hbTimeout}, {"handshakeTimeout", handshake}} {
		if opt.value != 0 {
			fmt.Fprintf(&b, `,%s:%d`, opt.name, opt.value)
		}
	}
	if loader == LoaderNone {
		b.WriteString(`,loader:'none'`)
	}
	if len(skeletons) > 0 {
		if skelDelay == 0 {
			skelDelay = 100
		}
		fmt.Fprintf(&b, `,skelDelay:%d,skelTarget:'%s',skeletons:[`, max(skelDelay, 0), escJS(skelTarget))
		for i, s := range skeletons {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, `['%s',%s]`, escJS(s.prefix), s.js)
		}
		b.WriteByte(']')
	}
	b.WriteString(`};`)
	return b.String()
}

// Skeleton shows node in place of the page content while a navigation to a
// path under prefix loads. It covers links, Navigate and back/forward.
// Prefixes match whole path segments and the longest one wins; "/" matches
// every page. A nil node turns skeletons off for prefix.
// Fast navigations (see SkeletonDelayMs) show nothing. The old content comes
// back when navigation fails. Register skeletons before serving requests.
//
//	app.Skeleton("/", ui.SkeletonList())
//	app.Skeleton("/settings", ui.SkeletonForm())
//
// The node is rendered in the browser on every slow navigation, so keep it
// static: no actions, ids or UnsafeJS. It gets role="status" and
// aria-label="Loading" unless it sets its own.
func (app *App) Skeleton(prefix string, node *Node) {
	prefix = "/" + strings.Trim(prefix, "/")
	js := "null"
	if node != nil {
		root := *node
		root.attrs = maps.Clone(node.attrs)
		root.Attr("role", cmp.Or(root.attrs["role"], "status"))
		root.Attr("aria-label", cmp.Or(root.attrs["aria-label"], app.Texts.Loading, "Loading"))
		root.Attr("data-gsui-skeleton", "")
		var b strings.Builder
		b.WriteString("function(){")
		counter := 0
		var postJS []string
		v := root.compile(&b, &counter, &postJS)
		b.WriteString(strings.Join(postJS, ""))
		fmt.Fprintf(&b, "return %s;}", v)
		js = b.String()
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	app.skeletons = slices.DeleteFunc(app.skeletons, func(s skeletonRoute) bool { return s.prefix == prefix })
	app.skeletons = append(app.skeletons, skeletonRoute{prefix: prefix, js: js})
	// Longest prefix first, so the client takes the first match.
	slices.SortStableFunc(app.skeletons, func(a, b skeletonRoute) int { return len(b.prefix) - len(a.prefix) })
}

// wsStubJS installs a queuing stub for __ws synchronously in <head>. The real
// client (/__ws.js) loads with defer and therefore runs AFTER the inline body
// script, so page-load JS (Node.UnsafeJS blocks, ctx.UnsafeHeadJS) that calls __ws would
// otherwise hit "__ws is not defined". The stub queues those calls; the real
// client replays and replaces it when it initializes.
// Sending methods queue and are replayed; state methods answer immediately so
// they are never replayed twice. restart() queues with its Promise callbacks.
// hold() uses the shared window.__gsuiHolds
// counter, so a hold taken before the client loads still suppresses the
// reconnect reload and its release function keeps working afterwards.
const wsStubJS = `window.__gsuiHolds=window.__gsuiHolds||0;
window.__ws||(window.__ws={__q:[],navigate:function(){this.__q.push(['navigate',arguments])},call:function(){this.__q.push(['call',arguments])},subscribe:function(){this.__q.push(['subscribe',arguments])},unsubscribe:function(){this.__q.push(['unsubscribe',arguments])},notfound:function(){this.__q.push(['notfound',arguments])},hold:function(){window.__gsuiHolds++;var r=false;return function(){if(r)return;r=true;window.__gsuiHolds=Math.max(0,window.__gsuiHolds-1)}},holds:function(){return window.__gsuiHolds},connected:function(){return false},offline:function(){return false},state:function(){return 'connecting'},reconnect:function(){},restart:function(){var q=this.__q;return new Promise(function(resolve,reject){q.push(['restart',[],resolve,reject])})}});`

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

// bootInitJS keeps the application hidden until its initial DOM, stylesheets
// and Tailwind-generated CSS are ready to paint. Web fonts are not awaited:
// text shows in a fallback font and swaps. Waiting on these resources instead
// of window.load avoids holding the UI for images. The timeout is a fail-safe
// for stalled third-party resources.
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
 if(preparing||done)return;preparing=true;paintReveal();
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
// The in-browser Tailwind engine styles the DOM once it is complete. A
// compiled stylesheet styles it at once, so the page shows when parsing
// ends, without waiting for deferred scripts.
if(!document.querySelector('script[data-gsui-style-engine]')){domReady=document.readyState!=='loading';if(!domReady)document.addEventListener('readystatechange',function(){if(document.readyState!=='loading'){domReady=true;check()}})}
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
// whatever JS string the server sends back, and shows a non-blocking
// offline badge when the WebSocket stays down past the grace window.
var wsClientJS = runtimeClientJS + `var __gsuiText=window.__gsuiText||` + Texts{}.json() + `;
var __wsPre=window.__ws;
var __gsuiCfg=window.__gsuiCfg||{};
function __gsuiOpt(k,d){var v=__gsuiCfg[k];return typeof v==='number'?v:d}
var __offline=(function(){
  // Two quiet badges: "Connecting…" while a reconnect is under way and
  // "Offline" once reconnecting keeps failing or the browser reports no
  // network. A page coming back from the background starts a fresh wait, so
  // a socket that died while hidden never flashes "Offline" on resume.
  var connTimer=0,offTimer=0,removeTimers={},els={},shown='',state='connecting';
  function online(){return typeof navigator==='undefined'||navigator.onLine!==false}
  function setState(next){
    if(state===next)return;state=next;
    try{document.documentElement.setAttribute('data-gsui-connection',next)}catch(_){}
    try{window.dispatchEvent(new CustomEvent('gsui:connection',{detail:{state:next}}))}catch(_){}
  }
  function build(kind){
    // A badge from a previous outage may still be fading out; cancel its
    // pending removal and fade it back in instead of leaving a dead node.
    if(removeTimers[kind]){clearTimeout(removeTimers[kind]);delete removeTimers[kind];}
    var el=els[kind];
    if(el&&el.parentNode){el.style.opacity='1';return;}
    var off=kind==='offline';
    var o=document.createElement('div');o.id=off?'__offline__':'__connecting__';
    o.setAttribute('role','status');
    // Badge only: never covers the page and never blocks input, so work in
    // progress (uploads, typing, drag & drop) keeps running while offline.
    o.style.cssText='position:fixed;top:12px;left:12px;z-index:2147483647;pointer-events:none;opacity:0;transition:opacity 160ms ease-out';
    var b=document.createElement('div');
    b.className='flex items-center gap-2 rounded-full px-3 py-1 text-white shadow-lg ring-1 ring-white/30';
    b.style.background=off?'linear-gradient(135deg,#ef4444,#ec4899)':'rgba(63,63,70,0.88)';
    var dot=document.createElement('span');dot.className='inline-block h-2.5 w-2.5 rounded-full bg-white/95 animate-pulse';
    var lbl=document.createElement('span');lbl.className=off?'font-semibold tracking-wide':'text-xs tracking-wide';lbl.style.color='#fff';lbl.textContent=off?__gsuiText.offline:__gsuiText.connecting;
    b.appendChild(dot);b.appendChild(lbl);
    if(off){var sub=document.createElement('span');sub.className='ml-1 text-xs';sub.style.color='rgba(255,255,255,0.9)';sub.textContent=__gsuiText.reconnecting;b.appendChild(sub)}
    o.appendChild(b);
    document.body.appendChild(o);
    requestAnimationFrame(function(){o.style.opacity='1';});
    els[kind]=o;
  }
  function fade(kind){
    var o=els[kind];if(!o)return;
    try{o.style.opacity='0';}catch(_){}
    if(removeTimers[kind])clearTimeout(removeTimers[kind]);
    // The element is kept until the node is actually gone so a new outage
    // inside the fade window reuses it instead of orphaning it.
    removeTimers[kind]=setTimeout(function(){
      delete removeTimers[kind];
      try{if(o&&o.parentNode){o.parentNode.removeChild(o);}}catch(_){}
      if(els[kind]===o)delete els[kind];
    },150);
  }
  function show(kind){
    if(kind==='offline')setState('offline');
    // A hidden page shows nothing; it starts over when it is shown again.
    if(document.hidden||shown===kind)return;
    if(shown)fade(shown);
    shown=kind;build(kind);
  }
  function clearTimers(){
    if(connTimer){clearTimeout(connTimer);connTimer=0}
    if(offTimer){clearTimeout(offTimer);offTimer=0}
  }
  // down runs when the socket is gone. The badges wait: first the quiet
  // "Connecting…", then "Offline" when the outage outlives the grace window.
  function down(){
    var grace=__gsuiOpt('grace',12000),conn=__gsuiOpt('connectingGrace',1500);
    if(state!=='offline')setState('connecting');
    if(grace>=0&&!online()){clearTimers();show('offline');return}
    if(conn>=0&&!connTimer&&!shown)connTimer=setTimeout(function(){connTimer=0;if(!shown)show('connecting')},conn);
    if(grace>=0&&!offTimer&&shown!=='offline')offTimer=setTimeout(function(){offTimer=0;show('offline')},grace);
  }
  // resume restarts the waits of an outage that began while hidden.
  function resume(){
    clearTimers();
    if(shown){fade(shown);shown=''}
    if(state==='offline')state='connecting';
  }
  function hide(){
    clearTimers();
    if(shown){fade(shown);shown=''}
    setState('connected');
  }
  return{down:down,resume:resume,hide:hide,visible:function(){return shown==='offline'},state:function(){return state}};
})();
var __ws=(function(){
  var ws,ready=false,seq=0,inflight={},loaderEl=null,loaderTimer=0,hadClose=false,backoff=500;
  var downSince=0,retryTimer=0,pingTimer=0,lastAttempt=0,subs=[],api,busy={},restartPending=null;
  var hbSeq=0,hbTimer=0,hsTimer=0,hsFails=0;
  // calls holds every call until its reply. A call that was never written to
  // a socket waits, in order, for a greeted connection: while a socket that
  // was idle in the background proves it is alive (verifying), while
  // connecting and until the server greeting. Only a sent navigation is sent
  // again when its socket dies, because it only reads. A sent action may have
  // run, so it is reported and never repeated.
  var calls={},verifying=false,greeted=false,callTimer=0,reloadAfterFlush=false;
  function message(act,data,id,sub){return JSON.stringify({act:act,data:data||{},id:id||0,sub:sub||'',page:window.__gsuiPage||location.pathname+location.search,version:window.__gsuiVersion})}
  function release(id){var el=busy[id];if(el){el.disabled=false;el.removeAttribute('aria-busy');if(el.form)el.form.removeAttribute('aria-busy');el.classList.remove('gsui-busy','opacity-60','cursor-wait');delete busy[id]}}
  var inst=typeof __gsuiCfg.inst==='string'?__gsuiCfg.inst:'',cur=inst,pendingReload='';
  // Holds live on window so the pre-client stub and the real client share one
  // counter across the stub handover.
  if(typeof window.__gsuiHolds!=='number')window.__gsuiHolds=0;
  // Reload the page, unless in-progress work (uploads, wizards, unsaved forms)
  // holds the connection: then wait and reload when the last hold is released.
  // reloadAfter<0 disables page reloads entirely.
  function requestReload(reason){
    if(__gsuiOpt('reloadAfter',15000)<0)return false;
    if(window.__gsuiHolds>0){
      if(pendingReload)return false;
      pendingReload=reason;
      try{window.dispatchEvent(new CustomEvent('gsui:reloadpending',{detail:{reason:reason}}))}catch(_){}
      return false;
    }
    try{location.reload();return true}catch(_){return false}
  }
  // Every connection starts with the server announcing its process id. Element
  // ids (ui.Target()) are random per process, so a page rendered by a previous
  // process holds ids the new one never generated: its responses would patch
  // elements that do not exist and the UI would look dead. Reloading is the
  // only way to resync, so a restarted server is treated like a long outage.
  // An explicit socket restart reports the change but preserves the page.
  // Queued calls are sent once the server is known. Calls queued for an
  // earlier process are dropped and reported; later calls go to the new one,
  // also when the reload is deferred or disabled.
  function onHello(id,restarting){
    if(!inst||!id||id===cur){greeted=true;flush();return}
    greeted=true;
    try{window.dispatchEvent(new CustomEvent('gsui:serverchanged',{detail:{was:cur,now:id}}))}catch(_){}
    // A restart keeps the page, so a later ordinary reconnect still sees the
    // change.
    if(restarting){flush();return}
    cur=id;reloadAfterFlush=false;expire(true);requestReload('server-restart');
  }
  function loud(){for(var id in inflight)if(inflight[id])return true;return false}
  function showLoader(){
    if(__gsuiCfg.loader==='none')return;
    // While disconnected there is no pending request; the offline badge already
    // says so, and a blocking loader would lock the page for the whole outage.
    if(!ready||loaderEl||loaderTimer)return;
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
      var lbl=document.createElement('span');lbl.className='font-semibold tracking-wide';lbl.textContent=__gsuiText.loading;
      var sub=document.createElement('span');sub.className='ml-1 text-white/85 text-xs';sub.style.color='rgba(255,255,255,0.9)';sub.textContent=__gsuiText.pleaseWait;
      b.appendChild(dot);b.appendChild(lbl);b.appendChild(sub);o.appendChild(b);
      document.body.appendChild(o);
      requestAnimationFrame(function(){o.style.opacity='1';});
      loaderEl=o;
    },120);
  }
  function hideLoader(){
    if(loud())return;
    if(loaderTimer){clearTimeout(loaderTimer);loaderTimer=0;}
    if(loaderEl){loaderEl.style.opacity='0';var el=loaderEl;loaderEl=null;setTimeout(function(){try{if(el&&el.parentNode)el.parentNode.removeChild(el)}catch(_){}},160);}
  }
  // A heartbeat is answered by the server's reader at once, even while an
  // action runs. No answer within hbTimeout means a socket that looks open
  // but is dead (a proxy or a suspended phone dropped it): replace it now
  // instead of waiting for TCP to notice. Any frame from the server counts.
  function probe(timeout){
    if(!ready||!ws||ws.readyState!==1||hbTimer)return;
    var sock=ws,id=++hbSeq;
    if(!rawSend('{"act":"__hb","hb":'+id+'}')){replaceDead(sock);return}
    hbTimer=setTimeout(function(){hbTimer=0;if(ws===sock)replaceDead(sock)},timeout);
  }
  function clearHB(){if(hbTimer){clearTimeout(hbTimer);hbTimer=0}}
  // alive runs for every frame from the server. A verified socket sends the
  // calls that waited for it.
  function alive(){clearHB();if(verifying){verifying=false;flush()}}
  function replaceDead(sock){
    ws=null;try{sock.close()}catch(_){}
    markDown();
    if(retryTimer){clearTimeout(retryTimer);retryTimer=0}
    connect();
  }
  function startPing(){
    stopPing();
    var iv=__gsuiOpt('keepAlive',25000);
    if(iv<0)return;
    // Idle WebSockets get closed by proxies; a periodic heartbeat keeps the
    // connection alive during long client-side work such as file uploads,
    // and finds a dead one.
    pingTimer=setInterval(function(){probe(__gsuiOpt('hbTimeout',12000))},iv);
  }
  function stopPing(){if(pingTimer){clearInterval(pingTimer);pingTimer=0}clearHB()}
  // rawSend never throws: readyState can already be CLOSING/CLOSED before the
  // close event fires, and send() would then throw and lose the message.
  function rawSend(msg){
    if(!ws||ws.readyState!==1)return false;
    try{ws.send(msg);return true}catch(_){return false}
  }
  // Queued calls keep their busy buttons. A navigation shows the old content
  // until its reply arrives.
  function releaseCalls(){
    inflight={};hideLoader();
    Object.keys(busy).forEach(function(id){if(!calls[id])release(id)});
    window.__gsuiLoadEnd(0,false);
  }
  function hasCalls(){for(var id in calls)return true;return false}
  // send puts one queued call on the wire, never ahead of an older one.
  function send(id){
    var c=calls[id];
    if(!c||c.sent||verifying||!greeted||!api.connected())return false;
    for(var k in calls)if(Number(k)<Number(id)&&!calls[k].sent)return false;
    // An open socket that refuses a send is broken; replace it.
    if(!rawSend(c.msg)){if(ws&&ws.readyState===1)replaceDead(ws);return false}
    c.sent=true;inflight[id]=c.loader;if(c.loader)showLoader();
    return true;
  }
  // flush sends the queued calls in order, then checks the socket: a call on
  // a dead socket is found within the heartbeat timeout.
  function flush(){
    expire(false);
    var any=false;
    Object.keys(calls).map(Number).sort(function(a,b){return a-b}).forEach(function(id){if(send(id))any=true});
    if(any)probe(__gsuiOpt('hbTimeout',12000));
  }
  // lost handles the sent calls whose replies can no longer arrive.
  function lost(report){
    var now=Date.now(),n=0;
    Object.keys(calls).forEach(function(id){
      var c=calls[id];if(!c.sent)return;
      if(c.act==='__nav'){c.sent=false;c.at=now;return}
      delete calls[id];release(id);n++;
    });
    if(n&&report)window.__gsuiReport(__gsuiText.lostMaybeDone);
    expireLater();
  }
  // A queued call waits at most queueTimeout for a connection.
  function expireLater(){
    if(callTimer)return;
    for(var id in calls)if(!calls[id].sent&&calls[id].act!=='__nav'){callTimer=setTimeout(function(){callTimer=0;expire(false)},1000);return}
  }
  // A queued navigation only reads, so it waits for the connection however
  // long it takes; a newer navigation replaces it. Only all=true drops it.
  function expire(all){
    var limit=__gsuiOpt('queueTimeout',20000),now=Date.now(),n=0;
    Object.keys(calls).forEach(function(id){
      var c=calls[id];
      if(c.sent||(!all&&(c.act==='__nav'||now-c.at<limit)))return;
      delete calls[id];delete inflight[id];release(id);window.__gsuiLoadEnd(Number(id),false);
      if(c.act!=='__nav')n++;
    });
    if(!loud())hideLoader();
    if(n)window.__gsuiReport(__gsuiText.lostNotSent);
    expireLater();
    reloadIfDone();
  }
  // reloadIfDone runs the reload a long outage deferred for queued calls,
  // unless a restart cancelled it meanwhile.
  function reloadIfDone(){if(reloadAfterFlush&&!hasCalls()){reloadAfterFlush=false;requestReload('long-outage')}}
  // At most one navigation waits (a newer one replaces it), so only actions
  // count against the queue limit.
  function unsent(){var n=0;for(var id in calls)if(!calls[id].sent&&calls[id].act!=='__nav')n++;return n}
  // accepts says whether a call can be sent or queued now. A navigation
  // replaces a waiting one, so the limit is for actions.
  function accepts(nav){return(api.connected()||__offline.state()!=='offline')&&(nav||unsent()<__gsuiOpt('queueLimit',8))}
  // markDown is the single place that records "the connection is gone". It
  // runs from onclose and when a silently dead socket is discovered, so the
  // bookkeeping is identical either way.
  function markDown(){
    if(!ready&&hadClose)return;
    ready=false;verifying=false;greeted=false;stopPing();
    // Replies for messages that were already on the wire will never arrive.
    lost(true);
    releaseCalls();
    if(!downSince)downSince=Date.now();
    __offline.down();hadClose=true;
  }
  function reconnectNow(){
    // readyState is authoritative: a socket can be dead while the ready flag
    // is still true because the close event has not been delivered yet
    // (backgrounded mobile tabs do this routinely).
    if(ws&&(ws.readyState===0||ws.readyState===1))return;
    markDown();
    // Do not reset the backoff here: repeated tab switches or online events
    // would otherwise hammer a server that is actually down. Just bring the
    // next scheduled attempt forward, at most once per second.
    var wait=1000-(Date.now()-lastAttempt);
    if(wait>0){
      // Still inside the cooldown. Schedule the attempt for when it expires
      // rather than dropping it, otherwise a socket that died within a second
      // of connecting waits for an outside event to recover.
      if(retryTimer)return;
      retryTimer=setTimeout(function(){retryTimer=0;connect()},wait);
      return;
    }
    if(retryTimer){clearTimeout(retryTimer);retryTimer=0}
    connect();
  }
  function finishRestart(err){
    var pending=restartPending;if(!pending)return;
    restartPending=null;clearTimeout(pending.timer);
    if(err)pending.reject(err);else pending.resolve();
  }
  function restartNow(){
    finishRestart(new Error('WebSocket restart superseded'));
    // Retire the socket before closing it so even a synchronous or late event
    // cannot turn this intentional replacement into an outage.
    var old=ws;ws=null;ready=false;verifying=false;greeted=false;stopPing();
    // Queued calls and navigations go out on the new socket. As before, an
    // action already sent is released without a report.
    lost(false);releaseCalls();reloadAfterFlush=false;hadClose=false;downSince=0;pendingReload='';
    if(retryTimer){clearTimeout(retryTimer);retryTimer=0}
    __offline.hide();
    if(old){try{old.close()}catch(_){}}
    return new Promise(function(resolve,reject){
      var sock;
      restartPending={resolve:resolve,reject:reject,timer:setTimeout(function(){
        finishRestart(new Error('WebSocket restart timed out'));
        if(ws===sock){ws=null;try{sock.close()}catch(_){}retry()}
      },10000)};
      // A fresh native WebSocket handshake uses the browser's current cookies.
      // Intentional restarts bypass reconnectNow's one-second cooldown.
      sock=connect(true);
    });
  }
  function retry(){
    markDown();
    var d=Math.min(10000,backoff)*(0.75+Math.random()*0.5);backoff=Math.min(10000,backoff*2);
    if(retryTimer)clearTimeout(retryTimer);
    retryTimer=setTimeout(function(){retryTimer=0;connect()},d);
  }
  function connect(restarting){
    retryTimer=0;
    if(ws&&(ws.readyState===0||ws.readyState===1))return;
    lastAttempt=Date.now();
    // sock is captured per attempt: handlers of an abandoned socket must not
    // touch the state of a newer one (that spawned parallel connections).
    var sock;
    try{sock=new WebSocket((location.protocol==='https:'?'wss://':'ws://')+location.host+'/__ws')}
    catch(err){if(restarting)finishRestart(err);retry();return;}
    ws=sock;greeted=false;
    // A handshake that hangs (common right after a phone wakes up) is given
    // up early; the server greets every connection at once.
    // Each timed out attempt waits twice as long, up to 30s, so a very slow
    // network still gets through.
    if(hsTimer)clearTimeout(hsTimer);
    var hs=__gsuiOpt('handshakeTimeout',8000);
    if(hs>=0)hsTimer=setTimeout(function(){hsTimer=0;if(ws===sock&&!sock.__gsuiHello){hsFails++;ws=null;try{sock.close()}catch(_){}if(restarting)finishRestart(new Error('WebSocket restart timed out'));retry()}},Math.min(Math.max(hs,30000),hs*Math.pow(2,hsFails)));
    sock.onopen=function(){
      if(ws!==sock){try{sock.close()}catch(_){}return;}
      // Without a known server id there is no greeting to wait for.
      ready=true;backoff=500;if(!inst)greeted=true;
      if(retryTimer){clearTimeout(retryTimer);retryTimer=0}
      __offline.hide();
      startPing();
      // Re-arm server-side subscriptions (Push loops) that died with the old
      // connection. Without this a silent
      // reconnect leaves live regions frozen.
      subs.forEach(function(s){rawSend(message(s.act,s.data,0,s.key))});
      if(loud())showLoader();
      if(restarting)finishRestart();
      if(hadClose){
        hadClose=false;
        var outage=downSince?Date.now()-downSince:0;downSince=0;
        var after=__gsuiOpt('reloadAfter',15000),holds=window.__gsuiHolds;
        // Reload only after a long outage. While the page holds the connection
        // the reload is deferred until the work finishes, never dropped.
        // Pending calls are answered first; the reload follows their replies.
        if(after>=0&&outage>=after){if(hasCalls())reloadAfterFlush=true;else if(requestReload('long-outage'))return}
        try{window.dispatchEvent(new CustomEvent('gsui:reconnected',{detail:{downMs:outage,holds:holds}}))}catch(_){
          try{window.dispatchEvent(new Event('gsui:reconnected'))}catch(__){}
        }
      }
      if(greeted)flush();
    };
    sock.onmessage=function(e){
      if(ws!==sock)return;
      alive();
      var m;try{m=JSON.parse(e.data)}catch(_){return}
      if(m&&m.__hb)return;
      if(m&&m.__hello){sock.__gsuiHello=true;hsFails=0;if(hsTimer){clearTimeout(hsTimer);hsTimer=0}onHello(m.__hello,restarting);return}
      if(!m||(!m.__r&&!m.__push))return;
      if(!m.broadcast&&(!Number.isInteger(m.version)||m.version<1))return;
      if(m&&m.__r){
        delete inflight[m.id];delete calls[m.id];release(m.id);hideLoader();
        // The reload deferred by a long outage runs after this reply is applied.
        if(reloadAfterFlush&&!hasCalls())setTimeout(reloadIfDone,0);
      }
      if(m&&m.version&&m.version!==window.__gsuiVersion){if(m.__r&&m.id)window.__gsuiLoadEnd(m.id,false);return}
      if(m&&m.__push&&m.subscription&&!subs.some(function(s){return s.key===m.subscription}))return;
      window.__gsuiPushSubscription=m&&m.subscription||'';
      var js=m.js;
      try{if(js)new Function(js)()}catch(err){console.error('ws exec error:',err)}
      finally{window.__gsuiPushSubscription='';if(m.__r&&m.id)window.__gsuiLoadEnd(m.id,false)}
      try{window.dispatchEvent(new Event('gsui:updated'))}catch(_){}
    };
    sock.onclose=function(){
      if(ws!==sock)return;
      if(restarting)finishRestart(new Error('WebSocket restart failed'));
      retry();
    };
    sock.onerror=function(){
      if(ws!==sock)return;
      if(restarting)finishRestart(new Error('WebSocket restart failed'));
      try{sock.close()}catch(_){}
    };
    return sock;
  }
  // resume runs when the page is shown again or the network comes back: an
  // open socket must prove it is alive before calls use it, a closed one
  // reconnects at once.
  // The badges wait from now, not from when the socket died in the
  // background.
  function resume(){
    if(document.hidden)return;
    __offline.resume();
    if(ws&&ws.readyState===1&&ready){verifying=true;probe(__gsuiOpt('resumeTimeout',4000));return}
    if(!api.connected()&&hadClose)__offline.down();
    reconnectNow();
  }
  connect();
  window.addEventListener('online',resume);
  window.addEventListener('offline',function(){if(!api.connected())__offline.down();else probe(__gsuiOpt('resumeTimeout',4000))});
  window.addEventListener('pageshow',function(e){if(e&&e.persisted)resume()});
  document.addEventListener('visibilitychange',resume);
  window.addEventListener('popstate',function(event){window.__gsuiPop(event)});
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
    if(e.defaultPrevented||e.isComposing||e.key!=='Enter')return;
    var t=e.target;
    if(!t||!t.tagName)return;
    var tag=t.tagName.toLowerCase();
    if(tag!=='input')return;
    var type=(t.getAttribute('type')||'text').toLowerCase();
    // Skip non-text input types
    if(type==='checkbox'||type==='radio'||type==='file'||type==='range'||type==='color'||type==='hidden'||type==='submit'||type==='reset'||type==='button'||type==='image')return;
    var form=t.form||t.closest('form'),btn;
    if(form&&form.hasAttribute('data-gsui-form'))return;
    if(form){btn=form.querySelector('button[type=submit]')||form.querySelector('button:not([type])')||form.querySelector('button');if(!btn&&form.id)btn=document.querySelector('button[form="'+form.id+'"]')}else{var p=t.parentElement;while(p&&p!==document.body){btn=p.querySelector('button');if(btn)break;p=p.parentElement}}
    if(btn){e.preventDefault();btn.click()}
  });
  api={
    navigate:function(path,options){window.__gsuiNavigate(path,options)},
    accepts:accepts,
    beginNavigation:function(patch){
      if(!patch){
        // Calls still on their way keep their busy controls.
        subs=[];inflight={};Object.keys(busy).forEach(function(id){if(!calls[id])release(id)});hideLoader();
        // Earlier navigations are superseded by this one.
        Object.keys(calls).forEach(function(id){if(calls[id].act==='__nav')delete calls[id]});
      }
    },
    // hold() marks critical in-page work (file upload, multi-step form).
    // While at least one hold is active, a reconnect never reloads the page,
    // so the work is not destroyed. Call the returned function when done.
    hold:function(){
      window.__gsuiHolds++;var released=false;
      return function(){
        if(released)return;released=true;window.__gsuiHolds=Math.max(0,window.__gsuiHolds-1);
        // A reload deferred by this hold runs as soon as the last one is gone.
        if(window.__gsuiHolds===0&&pendingReload)requestReload(pendingReload);
      };
    },
    holds:function(){return window.__gsuiHolds},
    connected:function(){return ready&&!!ws&&ws.readyState===1},
    offline:function(){return __offline.visible()},
    // state is "connecting", "connected" or "offline"; a gsui:connection
    // event reports each change.
    state:function(){return __offline.state()},
    reconnect:reconnectNow,
    restart:restartNow,
    // Managed subscriptions restart once per connection and stop on navigation.
    subscribe:function(act,data){
      var d=Object.assign({},data||{}),key=act+'|'+JSON.stringify(d);
      subs=subs.filter(function(s){return s.key!==key});
      subs.push({act:act,data:d,key:key});
      rawSend(message(act,d,0,key));
    },
    unsubscribe:function(act,data){
      var key=data===undefined?'':act+'|'+JSON.stringify(data||{});
      subs=subs.filter(function(s){return key?s.key!==key:s.act!==act});
      rawSend(message('__unsubscribe',{name:act,key:key}));
    },
    // opts.loader=false skips the loading overlay for this call.
    call:function(act,data,collect,source,opts){
      // Only a connection reported offline refuses calls. While it is
      // connecting or being verified, calls wait in the queue.
      if(!api.connected()&&__offline.state()==='offline'){window.__gsuiReport(__gsuiText.offlineNotSent);return 0}
      // The queue is bounded, below the server's inbound buffer of 16.
      if(!accepts(act==='__nav')){window.__gsuiReport(__gsuiText.lostNotSent);return 0}
      var d=Object.assign({},data||{});
      if(collect&&collect.length){
        collect.forEach(function(id){collectValue(id,d)});
      }
      var id=++seq;
      if(source&&/^(BUTTON|FORM)$/.test(source.tagName)){busy[id]=source;source.disabled=true;source.setAttribute('aria-busy','true');if(source.form)source.form.setAttribute('aria-busy','true');source.classList.add('gsui-busy','opacity-60','cursor-wait')}
      calls[id]={act:act,msg:message(act,d,id),at:Date.now(),sent:false,loader:!(opts&&opts.loader===false)};
      if(send(id))probe(__gsuiOpt('hbTimeout',12000));
      else{expireLater();if(!api.connected())reconnectNow()}
      return id;
    },
    notfound:function(id){
      rawSend(message('__notfound',{id:id,subscription:window.__gsuiPushSubscription||''}));
    }
  };
  return api;
})();
if(__wsPre&&__wsPre.__q){__wsPre.__q.forEach(function(it){try{var result=__ws[it[0]].apply(__ws,it[1]);if(it[2])result.then(it[2],it[3])}catch(e){if(it[3])it[3](e);else console.error('gsui: queued ws call failed:',e)}})}`

// ---------------------------------------------------------------------------
// WebSocket handler
// ---------------------------------------------------------------------------

// wsMessage is what the client sends.
type wsMessage struct {
	Act     string         `json:"act"`
	Data    map[string]any `json:"data"`
	ID      int64          `json:"id"`
	Page    string         `json:"page"`
	Version int64          `json:"version"`
	Sub     string         `json:"sub"`
}

// socket is one WebSocket and the request that opened it.
type socket struct {
	*websocket.Conn
	req *http.Request
}

// acceptWS opens a WebSocket. Replies are large, repetitive JS, so they are
// compressed (permessage-deflate) when the browser supports it.
func (app *App) acceptWS(w http.ResponseWriter, r *http.Request) {
	if err := app.wsHandshake(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // wsHandshake checked the origin
		CompressionMode:    websocket.CompressionNoContextTakeover,
	})
	if err != nil {
		return
	}
	app.handleWS(&socket{Conn: c, req: r})
}

// wsHandshake accepts same-origin browser requests, configured origins, and
// non-browser clients that do not send Origin.
func (app *App) wsHandshake(r *http.Request) error {
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
func (app *App) send(ws *socket, s string) error {
	app.mu.RLock()
	st, ok := app.connStates[ws]
	app.mu.RUnlock()
	if !ok {
		return fmt.Errorf("connection is closed")
	}
	st.writeMu.Lock()
	defer st.writeMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return ws.Write(ctx, websocket.MessageText, []byte(s))
}

// cancelConn cancels the push context for a connection and creates a fresh
// one. Any goroutine holding the old context will see ctx.pushCtx.Err() != nil.
func (app *App) cancelConn(ws *socket) {
	app.mu.Lock()
	defer app.mu.Unlock()
	if st, ok := app.connStates[ws]; ok {
		parent := st.connectionCtx
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithCancel(parent)
		st.cancel()
		for _, sub := range st.subscriptions {
			sub.cancel()
		}
		st.subscriptions = make(map[string]subscription)
		st.ctx = ctx
		st.cancel = cancel
		return
	}
}

// pushCtxForConn returns the current push context for a connection.
func (app *App) pushCtxForConn(ws *socket) context.Context {
	app.mu.RLock()
	defer app.mu.RUnlock()
	if st, ok := app.connStates[ws]; ok {
		return st.ctx
	}
	return context.Background()
}

func (app *App) handleWS(ws *socket) {
	// Register client and create initial push context
	connectionCtx, connectionCancel := context.WithCancel(ws.req.Context())
	defer connectionCancel()
	pushCtx, pushCancel := context.WithCancel(connectionCtx)
	var limit int64
	if app.MessageLimit != nil {
		limit = app.MessageLimit(ws.req)
	}
	if limit <= 0 {
		limit = 1 << 20
	}
	ws.SetReadLimit(limit)
	app.mu.Lock()
	app.clients[ws] = true
	app.connStates[ws] = &connState{ctx: pushCtx, connectionCtx: connectionCtx, connectionCancel: connectionCancel, cancel: pushCancel, session: make(map[string]any), subscriptions: make(map[string]subscription)}
	app.mu.Unlock()

	// Announce which process this is. A client whose page was rendered by a
	// different process holds a DOM full of ids this process never generated,
	// so it must reload instead of silently patching nothing.
	if err := app.send(ws, app.helloFrame()); err != nil {
		log.Printf("gsui: ws hello send error: %v", err)
	}

	defer func() {
		app.mu.Lock()
		if st, ok := app.connStates[ws]; ok {
			st.cancel()
			for _, sub := range st.subscriptions {
				sub.cancel()
			}
			delete(app.connStates, ws)
		}
		delete(app.clients, ws)
		app.mu.Unlock()
		_ = ws.CloseNow()
	}()

	// Read independently so disconnect cancels a running action. Dispatch is
	// serial and the inbound queue is bounded, in messages and in bytes.
	messages := make(chan string, 16)
	var queued atomic.Int64
	go func() {
		defer connectionCancel()
		defer close(messages)
		for {
			_, data, err := ws.Read(connectionCtx)
			if err != nil {
				return
			}
			raw := string(data)
			if id, ok := heartbeatID(raw); ok {
				// Answer here, not after queued actions, so a busy connection
				// is never mistaken for a dead one.
				if app.send(ws, `{"__hb":`+id+`}`) != nil {
					return
				}
				continue
			}
			if queued.Add(int64(len(raw))) > 2*limit {
				// Large messages queued behind a slow action would hold memory.
				return
			}
			select {
			case messages <- raw:
			case <-connectionCtx.Done():
				return
			default:
				// Backpressure must not strand the reader behind a blocked action.
				// Cancelling here also unblocks handlers using ctx.Context().
				return
			}
		}
	}()
	for raw := range messages {
		queued.Add(-int64(len(raw)))
		if connectionCtx.Err() != nil {
			return
		}

		// Parse the incoming message
		var msg wsMessage
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			log.Printf("gsui: invalid WebSocket message")
			continue
		}

		// Look up the action handler
		app.mu.RLock()
		handler, ok := app.actions[msg.Act]
		app.mu.RUnlock()

		if !ok {
			errJS := Notify("error", fmt.Sprintf("Unknown action: %s", msg.Act)).script()
			errJS = runtimeReply(msg.ID, msg.Version, errJS)
			if err := app.send(ws, errJS); err != nil {
				log.Printf("gsui: ws send error: %v", err)
				return
			}
			continue
		}

		// Build context with current push context
		ctx := &Context{
			Request:    ws.req,
			PathParams: make(map[string]string),
			Query:      make(map[string]string),
			wsConn:     ws,
			wsData:     msg.Data,
			app:        app,
			pushCtx:    app.pushCtxForConn(ws),
			version:    msg.Version,
		}

		// Execute handler -> get JS string (recover from panics)
		jsResponse := func() (resp string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("gsui: panic in action %q: %v\n%s", msg.Act, r, debug.Stack())
					resp = Notify("error", ctx.texts().ServerError).script()
				}
			}()
			if err := app.prepareAction(ctx, msg); err != nil {
				if denied, ok := errors.AsType[*denialError](err); ok {
					js, buildErr := denied.result.build(ctx)
					if buildErr != nil {
						return actionError(ctx, buildErr)
					}
					return js
				}
				return Notify("error", ctx.texts().Denied).script()
			}
			return handler(ctx)
		}()

		// Prepend any per-page CSS/JS injection from ctx.HeadCSS()/ctx.UnsafeHeadJS()
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
		jsResponse = runtimeReply(msg.ID, msg.Version, jsResponse)
		if jsResponse != "" {
			if err := app.send(ws, jsResponse); err != nil {
				log.Printf("gsui: ws send error: %v", err)
				return
			}
		}
		for _, start := range ctx.after {
			start()
		}
	}
}

// heartbeatID returns the id of a client heartbeat, {"act":"__hb","hb":N}.
func heartbeatID(raw string) (string, bool) {
	rest, ok := strings.CutPrefix(raw, `{"act":"__hb","hb":`)
	if !ok {
		return "", false
	}
	id, ok := strings.CutSuffix(rest, "}")
	if !ok || id == "" || len(id) > 16 {
		return "", false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return id, true
}

// ---------------------------------------------------------------------------
// Context
// ---------------------------------------------------------------------------

// Context carries request data for both page renders and WS action calls.
type Context struct {
	Request      *http.Request
	Session      map[string]any // connection-local scratch state; reset on reconnect
	PathParams   map[string]string
	Query        map[string]string
	wsConn       *socket
	wsData       map[string]any
	app          *App
	pushCtx      context.Context // cancelled when client navigates away or reports element not found
	version      int64
	subscription string
	user         any
	after        []func()
	headCSS      []string // per-page <style>/<link> tags collected via ctx.HeadCSS()
	headJS       []string // per-page <script> blocks collected via ctx.UnsafeHeadJS()
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

// UnsafeHeadJS registers a JavaScript block that runs once when the page
// loads. On a full page load it is a <script> in <head>. On live navigation it
// runs before the DOM swap. Prefer Actions, Node behaviors and App.Widget;
// use this only for page setup they do not cover.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
func (ctx *Context) UnsafeHeadJS(code string) {
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

func (ctx *Context) body(target any) error {
	if ctx.wsData == nil {
		return nil
	}
	b, err := json.Marshal(ctx.wsData)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

// Push sends Result effects to this client connection immediately.
// Useful for sending additional updates outside the normal request/response.
// Returns an error if the connection's push context has been cancelled
// (e.g. the client navigated away or reported a missing target element).
func (ctx *Context) Push(result Result) error {
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
	js, err := result.build(ctx)
	if err != nil {
		return err
	}
	b, _ := json.Marshal(map[string]any{"__push": true, "js": js, "version": ctx.version, "subscription": ctx.subscription})
	return ctx.app.send(ctx.wsConn, string(b))
}

// Broadcast sends effects to all connected clients. Effects must not depend on a page context.
func (app *App) Broadcast(result Result) error {
	js, err := result.build(nil)
	if err != nil {
		return err
	}
	frame, _ := json.Marshal(map[string]any{"__push": true, "broadcast": true, "js": js})
	app.mu.RLock()
	clients := make([]*socket, 0, len(app.clients))
	for conn := range app.clients {
		clients = append(clients, conn)
	}
	app.mu.RUnlock()
	var failures []error
	for _, conn := range clients {
		if err := app.send(conn, string(frame)); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}
