package ui

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

//go:embed runtime.js
var runtimeClientJS string

// Use installs page middleware shared by HTTP rendering, live navigation,
// region refresh and actions on that page. Register before Handler/Listen.
// Middleware must authorize without relying on writing cookies during WS
// actions; use a full HTTP redirect when session cookies need to change.
func (app *App) Use(middleware ...func(http.Handler) http.Handler) {
	app.middleware = append(app.middleware, middleware...)
}

// Context returns the page/subscription lifetime for connected actions, and
// the HTTP request lifetime for an initial render. Pass it to database work.
func (ctx *Context) Context() context.Context {
	if ctx.pushCtx != nil {
		return ctx.pushCtx
	}
	if ctx.Request != nil {
		return ctx.Request.Context()
	}
	return context.Background()
}

// User is the identity returned by App.Identity. It is never read from an
// action payload. Identity should resolve and validate a server-side session.
func (ctx *Context) User() any { return ctx.user }

func (app *App) authorize(ctx *Context, action string) error {
	if app.Identity != nil {
		user, err := app.Identity(ctx.Request)
		if err != nil {
			return err
		}
		ctx.user = user
	}
	if app.Authorize != nil {
		return app.Authorize(ctx, action)
	}
	return nil
}

func pageRequest(base *http.Request, path string) (*http.Request, error) {
	u, err := url.Parse(path)
	if err != nil || u.IsAbs() || u.Host != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
		return nil, errors.New("expected a local absolute path")
	}
	r := base.Clone(base.Context())
	r.Method = http.MethodGet
	r.URL = u
	r.RequestURI = u.RequestURI()
	r.Pattern = ""
	return r, nil
}

func (ctx *Context) setRequest(r *http.Request) {
	ctx.Request = r
	ctx.PathParams = requestPathParams(r)
	ctx.Query = make(map[string]string)
	for k, values := range r.URL.Query() {
		if len(values) > 0 {
			ctx.Query[k] = values[0]
		}
	}
}

func (app *App) prepareAction(ctx *Context, msg wsMessage) error {
	app.mu.RLock()
	st := app.connStates[ctx.wsConn]
	previous, version := st.request, st.version
	ctx.Session = st.session
	app.mu.RUnlock()
	if msg.Act == "__nav" {
		return nil
	} // destination authorization is in navigate
	if previous != nil && msg.Version != 0 && msg.Version != version {
		return errors.New("stale page")
	}
	path := msg.Page
	if path == "" && previous != nil {
		path = previous.URL.RequestURI()
	}
	if path == "" && len(app.middleware) > 0 {
		return errors.New("page required")
	}
	if previous != nil && st.view != nil && path != previous.URL.RequestURI() {
		return errors.New("view page changed")
	}
	if path != "" {
		r, err := pageRequest(ctx.Request, path)
		if err != nil {
			return err
		}
		_, matched, ok := app.matchPage(r)
		if !ok {
			return errors.New("page unavailable")
		}
		ctx.setRequest(matched)
	}
	if err := app.authorize(ctx, msg.Act); err != nil {
		return err
	}
	app.mu.Lock()
	if previous == nil {
		st.version = msg.Version
	}
	if path != "" {
		st.request = ctx.Request
	}
	if msg.Sub != "" {
		if old, ok := st.subscriptions[msg.Sub]; ok {
			old.cancel()
		}
		life, cancel := context.WithCancel(st.ctx)
		st.subscriptions[msg.Sub] = subscription{name: msg.Act, cancel: cancel}
		ctx.pushCtx = life
	}
	app.mu.Unlock()
	return nil
}

type navigationRequest struct {
	URL     string `json:"url"`
	History string `json:"history"` // push, replace, or none (popstate)
	Patch   bool   `json:"patch"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

func (app *App) navigate(ctx *Context) string {
	var req navigationRequest
	if err := ctx.Body(&req); err != nil {
		return Notify("error", "Invalid navigation")
	}
	r, err := pageRequest(ctx.Request, req.URL)
	if err != nil {
		return Notify("error", "Invalid navigation")
	}
	handler, matched, ok := app.matchPage(r)
	// HTTP owns redirects, 404s, and middleware responses, including Set-Cookie.
	if !ok {
		return Redirect(req.URL)
	}
	ctx.setRequest(matched)
	if err := app.authorize(ctx, ""); err != nil {
		return Redirect(req.URL)
	}
	app.mu.RLock()
	st := app.connStates[ctx.wsConn]
	previous := st.request
	app.mu.RUnlock()
	if req.Patch && (previous == nil || previous.Pattern != matched.Pattern) {
		return Redirect(req.URL)
	}
	if !req.Patch {
		app.cancelConn(ctx.wsConn)
		app.mu.Lock()
		st.view = nil
		st.viewPattern = ""
		app.mu.Unlock()
	}
	ctx.pushCtx = app.pushCtxForConn(ctx.wsConn)
	app.mu.Lock()
	st.request, st.version = matched, ctx.version
	app.mu.Unlock()
	node := handler(ctx)
	if node == nil {
		return Notify("error", "Page could not be rendered")
	}
	var js string
	if req.Patch {
		if app.layout != nil {
			js = node.ToJSMorphInner("__content__")
		} else {
			js = node.toJSMorphBody()
		}
	} else {
		if app.layout != nil {
			js = node.ToJSInner("__content__")
		} else {
			js = "if(window.__gsuiDispose)__gsuiDispose(document.body);document.body.replaceChildren();" + node.ToJS()
		}
	}
	data, _ := json.Marshal(req)
	page, _ := json.Marshal(matched.URL.RequestURI())
	return "window.__gsuiPage=" + string(page) + ";" + js + "__gsuiNavigationDone(" + string(data) + ");"
}

func runtimeReply(id, version int64, js string) string {
	if version == 0 {
		return wsReply(id, js)
	}
	b, _ := json.Marshal(map[string]any{"__r": 1, "id": id, "version": version, "js": js})
	return string(b)
}

type subscription struct {
	name   string
	cancel context.CancelFunc
}

func (app *App) cancelSubscription(wsConn *websocket.Conn, name, key string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	if st, ok := app.connStates[wsConn]; ok {
		for id, sub := range st.subscriptions {
			if (key != "" && id == key) || (key == "" && name != "" && sub.name == name) {
				sub.cancel()
				delete(st.subscriptions, id)
			}
		}
	}
}

// Close closes active WebSockets and cancels their page/subscription contexts.
// Call alongside http.Server.Shutdown, which does not close upgraded sockets.
func (app *App) Close() error {
	app.mu.RLock()
	var clients []*websocket.Conn
	var cancels []context.CancelFunc
	for conn := range app.clients {
		clients = append(clients, conn)
		if st := app.connStates[conn]; st != nil && st.connectionCancel != nil {
			cancels = append(cancels, st.connectionCancel)
		}
	}
	app.mu.RUnlock()
	for _, cancel := range cancels {
		cancel()
	}
	var errs []error
	for _, conn := range clients {
		if err := conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Subscription starts cancellable background work after the response has been
// sent. The callback must select on ctx.Context().Done() or use cancellable IO.
// Reconnect starts a fresh subscription; page changes/unsubscribe cancel it.
func (app *App) Subscription(name string, run func(*Context) error) {
	app.Action(name, func(ctx *Context) string {
		if ctx.subscription == "" {
			return Notify("error", "Subscription required")
		}
		ctx.after = append(ctx.after, func() {
			go func() {
				defer func() {
					if recover() != nil {
						log.Printf("gsui: subscription %q panicked", name)
					}
				}()
				if err := run(ctx); err != nil && ctx.Context().Err() == nil {
					log.Printf("gsui: subscription %q failed: %v", name, err)
				}
			}()
		})
		return ""
	})
}

// Subscribe attaches a subscription to this node's lifetime. Use different
// data to distinguish multiple subscriptions to the same action.
func (n *Node) Subscribe(name string, data ...any) *Node {
	var payload any = map[string]any{}
	if len(data) > 0 {
		payload = data[0]
	}
	b, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Errorf("subscription payload: %w", err))
	}
	n.rawJS += fmt.Sprintf("__ws.subscribe('%s',%s);(this.__gsuiCleanup||(this.__gsuiCleanup=[])).push(function(){__ws.unsubscribe('%s',%s)});", escJS(name), b, escJS(name), b)
	return n
}

// NavLink is a real anchor enhanced with live navigation. Modified clicks,
// downloads, external URLs and target=_blank keep native browser behavior.
func NavLink(path string, class ...string) *Node {
	return A(class...).Attr("href", path).Attr("data-gsui-nav", "")
}

// Navigate mounts a destination page and updates browser history after render.
func Navigate(path string) string { return fmt.Sprintf("__ws.navigate('%s');", escJS(path)) }

// PatchURL updates parameters within the same route, preserving the live view.
// replace=true replaces the current history entry (useful for search input).
func PatchURL(path string, replace bool) string {
	return fmt.Sprintf("__ws.navigate('%s',{patch:true,replace:%t});", escJS(path), replace)
}

// Key gives a node stable identity during a morph. Keys must be unique among
// siblings. IDs also act as keys. Preserve opts an external widget out of morphs.
func (n *Node) Key(key string) *Node { return n.Attr("data-gsui-key", key) }
func (n *Node) Preserve() *Node      { return n.Attr("data-gsui-preserve", "") }
func Region(name string, content *Node) *Node {
	return Div().ID(name).Attr("data-gsui-region", "").Render(content)
}

// OnInput sends a debounced server action (default 200ms). Collect field IDs
// through Action.Collect as usual. Removing the input cancels its timer.
func (n *Node) OnInput(action *Action, delay ...time.Duration) *Node {
	if action == nil {
		return n
	}
	if action.rawJS != "" {
		return n.On("input", action)
	}
	wait := 200 * time.Millisecond
	if len(delay) > 0 {
		wait = delay[0]
	}
	data, err := json.Marshal(action.Data)
	if err != nil {
		panic(err)
	}
	collect, err := json.Marshal(action.Collect)
	if err != nil {
		panic(err)
	}
	return n.On("input", JS(fmt.Sprintf("var el=event.currentTarget;clearTimeout(el.__gsuiInputTimer);if(!el.__gsuiInputCleanup){el.__gsuiInputCleanup=true;(el.__gsuiCleanup||(el.__gsuiCleanup=[])).push(function(){clearTimeout(el.__gsuiInputTimer)})}el.__gsuiInputTimer=setTimeout(function(){__ws.call('%s',%s,%s,null,{queue:%t})},%d);", escJS(action.Name), data, collect, !action.NoQueue, max(wait.Milliseconds(), 0))))
}

// Toggle and dialog helpers run locally; business actions still run in Go.
func Toggle(id string) *Action {
	return JS(fmt.Sprintf("var el=document.getElementById('%s');if(el){el.hidden=!el.hidden;event.currentTarget.setAttribute('aria-expanded',String(!el.hidden))}", escJS(id)))
}
func OpenDialog(id string) *Action {
	return JS(fmt.Sprintf("var el=document.getElementById('%s');if(el&&!el.open)el.showModal();", escJS(id)))
}
func CloseDialog(id string) *Action {
	return JS(fmt.Sprintf("var el=document.getElementById('%s');if(el)el.close();", escJS(id)))
}

// Result describes server effects without requiring handlers to build JS.
// Its zero value means success with no DOM changes.
type Result struct {
	effects []func(*Context) (string, error)
}

func (r Result) effect(fn func(*Context) (string, error)) Result {
	r.effects = append(append([]func(*Context) (string, error){}, r.effects...), fn)
	return r
}
func (r Result) Toast(message string) Result {
	return r.effect(func(*Context) (string, error) { return Notify("success", message), nil })
}
func (r Result) Navigate(path string) Result {
	return r.effect(func(*Context) (string, error) { return Navigate(path), nil })
}
func (r Result) PatchURL(path string, replace bool) Result {
	return r.effect(func(*Context) (string, error) { return PatchURL(path, replace), nil })
}
func (r Result) Morph(id string, node *Node) Result {
	return r.effect(func(*Context) (string, error) {
		if node == nil {
			return "", errors.New("nil node")
		}
		return node.ToJSMorph(id), nil
	})
}
func (r Result) Remove(id string) Result {
	return r.effect(func(*Context) (string, error) { return RemoveEl(id), nil })
}

// Refresh reruns the current page and morphs only the named regions. Page
// rendering should be free of mutations and return stable region IDs.
func Refresh(regions ...string) Result { return Result{}.Refresh(regions...) }
func (r Result) Refresh(regions ...string) Result {
	return r.effect(func(ctx *Context) (string, error) {
		handler, matched, ok := ctx.app.matchPage(ctx.Request)
		if !ok {
			return "", errors.New("page unavailable")
		}
		ctx.setRequest(matched)
		if err := ctx.app.authorize(ctx, ""); err != nil {
			return "", err
		}
		root := handler(ctx)
		var b strings.Builder
		for _, id := range regions {
			n := findRegion(root, id)
			if n == nil {
				return "", fmt.Errorf("region %q not found", id)
			}
			b.WriteString(n.ToJSMorph(id))
		}
		return b.String(), nil
	})
}
func findRegion(n *Node, id string) *Node {
	if n == nil {
		return nil
	}
	if n.id == id {
		if _, ok := n.attrs["data-gsui-region"]; ok {
			return n
		}
	}
	for _, child := range n.children {
		if found := findRegion(child, id); found != nil {
			return found
		}
	}
	return nil
}
func (r Result) build(ctx *Context) (string, error) {
	var b strings.Builder
	for _, effect := range r.effects {
		js, err := effect(ctx)
		if err != nil {
			return "", err
		}
		b.WriteString(js)
	}
	return b.String(), nil
}

// ActionRef is a typed action handle. Register once during application setup.
type ActionRef[T any] struct{ name string }

func (a ActionRef[T]) Call(input T) *Action {
	b, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		panic("action input must be a JSON object")
	}
	return &Action{Name: a.name, Data: data, NoQueue: true}
}

// RegisterAction decodes input, calls optional Validate() error on *T, and
// renders Result effects. Errors are logged server-side, not exposed to users.
func RegisterAction[T any](app *App, name string, handler func(*Context, T) (Result, error)) ActionRef[T] {
	app.Action(name, func(ctx *Context) string {
		var input T
		if err := ctx.Body(&input); err != nil {
			return Notify("error", "Invalid input")
		}
		if valid, ok := any(&input).(interface{ Validate() error }); ok {
			if err := valid.Validate(); err != nil {
				return actionError(ctx, err)
			}
		}
		result, err := handler(ctx, input)
		if err != nil {
			return actionError(ctx, err)
		}
		js, err := result.build(ctx)
		if err != nil {
			return actionError(ctx, err)
		}
		return js
	})
	return ActionRef[T]{name: name}
}

// ValidationError carries safe field messages for a form submission.
type ValidationError struct{ Fields FormErrors }

func (e ValidationError) Error() string { return "form validation failed" }
func actionError(ctx *Context, err error) string {
	var validation ValidationError
	if errors.As(err, &validation) {
		form, _ := ctx.wsData["__form"].(string)
		b, _ := json.Marshal(validation.Fields)
		return fmt.Sprintf("__gsuiFormErrors('%s',%s);", escJS(form), b)
	}
	log.Printf("gsui: action failed: %v", err)
	return Notify("error", "Unable to complete the request")
}

// TypedForm uses native form submission and input constraints. Supply ordinary
// input nodes with name attributes matching T's JSON fields. Number inputs
// decode as numbers, checkboxes as booleans, and multi-selects as string slices.
type TypedForm[T any] struct{ node *Node }

func FormFor[T any](id string, class ...string) *TypedForm[T] {
	return &TypedForm[T]{node: Form(class...).ID(id)}
}
func (f *TypedForm[T]) Render(children ...*Node) *TypedForm[T] { f.node.Render(children...); return f }
func (f *TypedForm[T]) Submit(action ActionRef[T]) *Node {
	return f.node.Attr("data-gsui-form", "").OnSubmit(JS(fmt.Sprintf("event.preventDefault();__gsuiSubmit(event,'%s');", escJS(action.name))))
}

// Title sets the title used after live navigation. For initial HTTP metadata,
// use App.Title; the marker also updates the initial document after mounting.
func (n *Node) Title(title string) *Node {
	n.Attr("data-gsui-title", title)
	n.rawJS += SetTitle(title)
	return n
}

// ToJSMorph updates an element while preserving keyed descendants and dirty
// inputs. Node.JS/Subscribe setup runs only for newly inserted nodes.
func (n *Node) ToJSMorph(id string) string {
	return n.morphJS("document.getElementById('"+escJS(id)+"')", false)
}
func (n *Node) ToJSMorphInner(id string) string {
	return n.morphJS("document.getElementById('"+escJS(id)+"')", true)
}
func (n *Node) toJSMorphBody() string { return n.morphJS("document.body", true) }
func (n *Node) morphJS(target string, inner bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "(function(){var target=%s;if(!target)return;", target)
	counter := 0
	var post []string
	root := n.compile(&b, &counter, &post)
	if inner {
		fmt.Fprintf(&b, "var wrapper=document.createElement('div');wrapper.appendChild(%s);__gsuiMorphChildren(target,wrapper);", root)
	} else {
		fmt.Fprintf(&b, "__gsuiMorph(target,%s);", root)
	}
	for _, js := range post {
		idx := strings.LastIndex(js, ".call(")
		if idx >= 0 {
			variable := strings.TrimSuffix(js[idx+6:], ");")
			fmt.Fprintf(&b, "if(!%s.__gsuiMounted){%s}", variable, js)
		}
	}
	b.WriteString("})();")
	return b.String()
}

// View is instantiated per connection, never shared between users/tabs. Its
// events run serially on the connection. Reconnect creates a fresh view: load
// durable state in Mount, or restore drafts explicitly from application storage.
type View interface {
	Render(*ViewContext) *Node
	Handle(*ViewContext, Event) error
}
type ViewContext = Context
type Event struct {
	Name string
	Data json.RawMessage
}

func (e Event) Decode(target any) error { return json.Unmarshal(e.Data, target) }
func (ctx *Context) Event(name string, data ...any) *Action {
	var payload any = map[string]any{}
	if len(data) > 0 {
		payload = data[0]
	}
	return &Action{Name: "__live", Data: map[string]any{"event": name, "data": payload}, NoQueue: true}
}

// Live registers a server-owned view. Mount(*ViewContext) error is optional and
// runs once when connected. HTTP render is a temporary view without Mount.
func (app *App) Live(pattern string, factory func() View) {
	app.Page(pattern, func(ctx *Context) *Node { return app.renderLive(ctx, factory) })
}
func (app *App) renderLive(ctx *Context, factory func() View) *Node {
	var view View
	if ctx.wsConn != nil {
		app.mu.RLock()
		st := app.connStates[ctx.wsConn]
		view = st.view
		pattern := st.viewPattern
		app.mu.RUnlock()
		if view != nil && pattern != ctx.Request.Pattern {
			return Div().Text("Navigate to open this view")
		}
		if view == nil {
			view = factory()
			if view == nil {
				return Div().Text("View unavailable")
			}
			if mounted, ok := view.(interface{ Mount(*ViewContext) error }); ok {
				if err := mounted.Mount(ctx); err != nil {
					log.Printf("gsui: mount failed: %v", err)
					return Div().Text("View unavailable")
				}
			}
			app.mu.Lock()
			st.view, st.viewPattern = view, ctx.Request.Pattern
			app.mu.Unlock()
		}
	} else {
		view = factory()
	}
	if view == nil {
		return Div().Text("View unavailable")
	}
	return Div().ID("__live__").Render(view.Render(ctx)).Subscribe("__live")
}
func (app *App) liveEvent(ctx *Context) string {
	handler, matched, ok := app.matchPage(ctx.Request)
	if !ok {
		return Notify("error", "Page unavailable")
	}
	ctx.setRequest(matched)
	var req struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if err := ctx.Body(&req); err != nil {
		return Notify("error", "Invalid event")
	}
	// Ensures Mount precedes the first event, including reconnect registration.
	root := handler(ctx)
	app.mu.RLock()
	view := app.connStates[ctx.wsConn].view
	app.mu.RUnlock()
	if view == nil {
		return Notify("error", "No live view mounted")
	}
	if req.Event != "" {
		if err := view.Handle(ctx, Event{Name: req.Event, Data: req.Data}); err != nil {
			return actionError(ctx, err)
		}
		root = handler(ctx)
	}
	return root.ToJSMorph("__live__")
}
