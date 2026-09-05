package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

func runtimeSocket(t *testing.T, app *App) *websocket.Conn {
	t.Helper()
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	t.Cleanup(func() { _ = app.Close() })
	ws, err := websocket.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/__ws", "", srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ws.Close() })
	_ = ws.SetDeadline(time.Now().Add(5 * time.Second))
	var hello string
	if err := websocket.Message.Receive(ws, &hello); err != nil {
		t.Fatal(err)
	}
	return ws
}
func runtimeCall(t *testing.T, ws *websocket.Conn, act, page string, version int64, data map[string]any) string {
	t.Helper()
	msg := wsMessage{Act: act, Data: data, ID: 1, Page: page, Version: version}
	b, _ := json.Marshal(msg)
	if err := websocket.Message.Send(ws, string(b)); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatal(err)
	}
	var reply struct {
		JS string `json:"js"`
	}
	if err := json.Unmarshal([]byte(raw), &reply); err != nil {
		t.Fatalf("reply %s: %v", raw, err)
	}
	return reply.JS
}

func TestRuntimeSharedMiddleware(t *testing.T) {
	app := NewApp()
	var renders atomic.Int32
	type contextKey struct{}
	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/private" {
				http.Error(w, "denied", 401)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, "identity")))
		})
	})
	app.Page("/private", func(*Context) *Node { renders.Add(1); return Div().Text("PRIVATE") })
	app.Page("/public/{id}", func(*Context) *Node { return Div() })
	app.action("inspect", func(ctx *Context) string {
		return fmt.Sprintf("%s:%s:%s:%v", ctx.Request.URL.Path, ctx.PathParams["id"], ctx.Query["q"], ctx.Request.Context().Value(contextKey{}))
	})
	ws := runtimeSocket(t, app)
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/private", nil))
	if rr.Code != 401 {
		t.Fatal(rr.Code)
	}
	js := runtimeCall(t, ws, "__nav", "", 1, map[string]any{"url": "/private"})
	if renders.Load() != 0 || !strings.Contains(js, "window.location.href") {
		t.Fatal(js)
	}
	js = runtimeCall(t, ws, "inspect", "/public/42?q=yes", 1, nil)
	if js != "/public/42:42:yes:identity" {
		t.Fatal(js)
	}
}

func TestRuntimeNavigationContextAndCancellation(t *testing.T) {
	app := NewApp()
	old := make(chan context.Context, 1)
	app.Page("/a", func(*Context) *Node { return Div() })
	app.Page("/b", func(ctx *Context) *Node {
		if err := ctx.Context().Err(); err != nil {
			t.Error(err)
		}
		return Div().Text("new page")
	})
	app.action("start", func(ctx *Context) string { old <- ctx.Context(); return "" })
	ws := runtimeSocket(t, app)
	runtimeCall(t, ws, "start", "/a", 1, nil)
	previous := <-old
	js := runtimeCall(t, ws, "__nav", "/a", 2, map[string]any{"url": "/b", "history": "push"})
	if !strings.Contains(js, "new page") {
		t.Fatal(js)
	}
	select {
	case <-previous.Done():
	case <-time.After(time.Second):
		t.Fatal("old page not cancelled")
	}
	js = runtimeCall(t, ws, "__nav", "/b", 3, map[string]any{"url": "/missing"})
	if !strings.Contains(js, "window.location.href") {
		t.Fatal("missing route must use HTTP", js)
	}
}

func TestRuntimeAuthorizationRechecked(t *testing.T) {
	app := NewApp()
	var allowed atomic.Bool
	allowed.Store(true)
	app.Page("/", func(*Context) *Node { return Div() })
	app.Identity = func(*http.Request) (any, error) {
		if !allowed.Load() {
			return nil, errors.New("expired")
		}
		return "user", nil
	}
	var runs atomic.Int32
	app.action("save", func(ctx *Context) string {
		if ctx.User() != "user" {
			t.Error("missing identity")
		}
		runs.Add(1)
		return ""
	})
	ws := runtimeSocket(t, app)
	runtimeCall(t, ws, "save", "/", 1, nil)
	allowed.Store(false)
	runtimeCall(t, ws, "save", "/", 1, nil)
	if runs.Load() != 1 {
		t.Fatal("expired identity was accepted")
	}
}

func TestRuntimeDisconnectCancelsRunningAction(t *testing.T) {
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	started, stopped := make(chan struct{}), make(chan struct{})
	app.action("wait", func(ctx *Context) string {
		close(started)
		<-ctx.Context().Done()
		close(stopped)
		return ""
	})
	ws := runtimeSocket(t, app)
	if err := websocket.Message.Send(ws, `{"act":"wait","page":"/","version":1}`); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("action did not start")
	}
	ws.Close()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("disconnect did not cancel action")
	}
}

func TestRuntimeCloseCancelsRunningActionWithQueuedMessages(t *testing.T) {
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	started, stopped := make(chan struct{}), make(chan struct{})
	app.action("wait", func(ctx *Context) string {
		close(started)
		<-ctx.Context().Done()
		close(stopped)
		return ""
	})
	ws := runtimeSocket(t, app)
	if err := websocket.Message.Send(ws, `{"act":"wait","page":"/","version":1}`); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("action did not start")
	}
	for i := 0; i < 8; i++ {
		if err := websocket.Message.Send(ws, `{"act":"__ping"}`); err != nil {
			t.Fatal(err)
		}
	}
	app.Close()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not cancel action")
	}
}

func TestRuntimeSubscriptionsAreIndependent(t *testing.T) {
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	started := make(chan string, 2)
	stopped := make(chan string, 2)
	app.Subscription("ticks", func(ctx *Context) error {
		id := ctx.wsData["id"].(string)
		started <- id
		<-ctx.Context().Done()
		stopped <- id
		return ctx.Context().Err()
	})
	ws := runtimeSocket(t, app)
	for _, id := range []string{"a", "b"} {
		b, _ := json.Marshal(wsMessage{Act: "ticks", Page: "/", Version: 1, Sub: id, Data: map[string]any{"id": id}})
		if err := websocket.Message.Send(ws, string(b)); err != nil {
			t.Fatal(err)
		}
		var reply string
		if err := websocket.Message.Receive(ws, &reply); err != nil {
			t.Fatal(err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("subscription not started")
		}
	}
	runtimeCall(t, ws, "__unsubscribe", "/", 1, map[string]any{"key": "a"})
	select {
	case id := <-stopped:
		if id != "a" {
			t.Fatal(id)
		}
	case <-time.After(time.Second):
		t.Fatal("unsubscribe did not cancel")
	}
	select {
	case id := <-stopped:
		t.Fatalf("unrelated subscription %s cancelled", id)
	default:
	}
	app.Close()
	select {
	case id := <-stopped:
		if id != "b" {
			t.Fatal(id)
		}
	case <-time.After(time.Second):
		t.Fatal("close did not cancel")
	}
}

type runtimeCounter struct{ count int }

func (v *runtimeCounter) Render(ctx *ViewContext) *Node {
	return Span().Text(fmt.Sprintf("count:%d", v.count))
}
func (v *runtimeCounter) Handle(ctx *ViewContext, event Event) error {
	if event.Name == "inc" {
		v.count++
	}
	return nil
}
func TestRuntimeViewsAreIsolatedAndPatchPreservesState(t *testing.T) {
	app := NewApp()
	app.Live("/counter", func() View { return &runtimeCounter{} })
	a := runtimeSocket(t, app)
	b := runtimeSocket(t, app)
	for i := 1; i <= 2; i++ {
		js := runtimeCall(t, a, "__live", "/counter", 1, map[string]any{"event": "inc"})
		if !strings.Contains(js, fmt.Sprintf("count:%d", i)) {
			t.Fatal(js)
		}
	}
	js := runtimeCall(t, b, "__live", "/counter", 1, map[string]any{"event": "inc"})
	if !strings.Contains(js, "count:1") {
		t.Fatal(js)
	}
	js = runtimeCall(t, a, "__nav", "/counter", 1, map[string]any{"url": "/counter?tab=other", "patch": true})
	if !strings.Contains(js, "count:2") {
		t.Fatal("patch lost state", js)
	}
	js = runtimeCall(t, a, "__nav", "/counter?tab=other", 2, map[string]any{"url": "/counter"})
	if !strings.Contains(js, "count:0") {
		t.Fatal("navigation did not reset", js)
	}
}

type runtimeInput struct {
	Name string `json:"name"`
}

func (input *runtimeInput) Validate() error {
	if input.Name == "" {
		return ValidationError{Fields: FormErrors{"name": "Required"}}
	}
	return nil
}
func TestRuntimeTypedActionValidationAndRefresh(t *testing.T) {
	app := NewApp()
	app.Page("/", func(ctx *Context) *Node { return Region("profile", Span().Text(fmt.Sprint(ctx.Session["name"]))) })
	ref := RegisterAction(app, "rename", func(ctx *Context, input runtimeInput) (Result, error) {
		ctx.Session["name"] = input.Name
		return Refresh("profile").Toast("Saved"), nil
	})
	if ref.Call(runtimeInput{Name: "test"}).Name != "rename" {
		t.Fatal("incorrect action reference")
	}
	ws := runtimeSocket(t, app)
	js := runtimeCall(t, ws, "rename", "/", 1, map[string]any{"__form": "rename-form"})
	if !strings.Contains(js, "__gsuiFormErrors('rename-form'") {
		t.Fatal(js)
	}
	js = runtimeCall(t, ws, "rename", "/", 1, map[string]any{"name": "Alice"})
	if !strings.Contains(js, "Alice") || !strings.Contains(js, "__gsuiMorph") {
		t.Fatal(js)
	}
	js = runtimeCall(t, ws, "rename", "/", 1, map[string]any{"name": 42})
	if !strings.Contains(js, "Invalid input") {
		t.Fatal(js)
	}
}

func TestRuntimeTableSourceKeepsStateInURL(t *testing.T) {
	app := NewApp()
	queries := make(chan TableQuery, 4)
	source := RegisterTable(app, "records", func(ctx *Context, q TableQuery) (TablePage[runtimeInput], error) {
		queries <- q
		return TablePage[runtimeInput]{Rows: []*runtimeInput{{Name: q.Search}}, Total: 1}, nil
	}, func(table *DataTable[runtimeInput]) {
		table.PageSize(5).RowKey(func(row *runtimeInput) string { return row.Name }).Col("Name", ColOpt[runtimeInput]{Sortable: true, Text: func(row *runtimeInput) *Node { return Span().Text(row.Name) }})
	})
	app.Page("/", func(ctx *Context) *Node {
		node, err := source.Render(ctx)
		if err != nil {
			t.Error(err)
		}
		return node
	})
	a := runtimeSocket(t, app)
	b := runtimeSocket(t, app)
	js := runtimeCall(t, a, "table.records", "/", 1, map[string]any{"operation": "search", "search": "Alice", "page": 100, "pageSize": 1000, "sort": 999})
	q := <-queries
	if q.Search != "Alice" || q.Limit() != 2000 || q.SortColumn != -1 {
		t.Fatalf("invalid normalized query: %+v", q)
	}
	if !strings.Contains(js, "__gsuiQueryDone") || !strings.Contains(js, "records.q=Alice") {
		t.Fatal(js)
	}
	runtimeCall(t, b, "table.records", "/", 1, map[string]any{"operation": "search", "search": "Bob", "page": 1, "pageSize": 5, "sort": 0})
	q = <-queries
	if q.Search != "Bob" || q.Limit() != 5 {
		t.Fatal(q)
	}
	runtimeCall(t, a, "__nav", "/", 2, map[string]any{"url": "/?records.q=Alice&records.page=2&records.size=5"})
	q = <-queries
	if q.Search != "Alice" || q.Limit() != 10 {
		t.Fatal("URL did not restore query", q)
	}
}

func TestRuntimeRejectsUnversionedAndPagelessActions(t *testing.T) {
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	var runs atomic.Int32
	RegisterAction(app, "save", func(*Context, struct{}) (Result, error) {
		runs.Add(1)
		return Result{}.Toast("Saved"), nil
	})
	ws := runtimeSocket(t, app)
	for _, request := range []struct {
		page    string
		version int64
	}{{"/", 0}, {"", 1}} {
		js := runtimeCall(t, ws, "save", request.page, request.version, nil)
		if !strings.Contains(js, "Request denied") {
			t.Fatal(js)
		}
	}
	if runs.Load() != 0 {
		t.Fatal("unsupported request reached handler")
	}
	runtimeCall(t, ws, "save", "/", 1, nil)
	if runs.Load() != 1 {
		t.Fatal("versioned action did not run")
	}
}

func TestRuntimeTypedPushAndBroadcastEnvelopes(t *testing.T) {
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	stopped := make(chan error, 1)
	app.Subscription("clock", func(ctx *Context) error {
		if err := ctx.Push(Result{}.SetText("clock", "tick")); err != nil {
			return err
		}
		<-ctx.Context().Done()
		err := ctx.Push(Result{}.SetText("clock", "late"))
		stopped <- err
		return err
	})
	ws := runtimeSocket(t, app)
	err := websocket.Message.Send(ws, `{"act":"clock","page":"/","version":1,"sub":"clock|{}","data":{}}`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		var raw string
		if err := websocket.Message.Receive(ws, &raw); err != nil {
			t.Fatal(err)
		}
		var frame struct {
			Reply        int    `json:"__r"`
			Push         bool   `json:"__push"`
			Version      int    `json:"version"`
			Subscription string `json:"subscription"`
			JS           string `json:"js"`
		}
		if err := json.Unmarshal([]byte(raw), &frame); err != nil {
			t.Fatal(err)
		}
		if frame.Version != 1 {
			t.Fatal("push/reply missing version", raw)
		}
		if i == 1 && (!frame.Push || frame.Subscription != "clock|{}" || !strings.Contains(frame.JS, "tick")) {
			t.Fatal(raw)
		}
	}
	if err := app.Broadcast(Result{}.Toast("Broadcast")); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatal(err)
	}
	var frame struct {
		Push bool   `json:"__push"`
		JS   string `json:"js"`
	}
	if err := json.Unmarshal([]byte(raw), &frame); err != nil || !frame.Push || !strings.Contains(frame.JS, "Broadcast") {
		t.Fatal("invalid broadcast", raw, err)
	}
	runtimeCall(t, ws, "__unsubscribe", "/", 1, map[string]any{"key": "clock|{}"})
	select {
	case err := <-stopped:
		if err == nil {
			t.Fatal("cancelled subscription allowed push")
		}
	case <-time.After(time.Second):
		t.Fatal("subscription did not stop")
	}
	if err := app.Broadcast(Refresh("page")); err == nil {
		t.Fatal("broadcast accepted page-dependent effect")
	}
}
