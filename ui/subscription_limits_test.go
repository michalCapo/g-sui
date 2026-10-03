package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/net/websocket"
)

func subscriptionConnection(t *testing.T, app *App) (*connState, func() *Context) {
	t.Helper()
	life, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	ws := &websocket.Conn{}
	st := &connState{ctx: life, session: make(map[string]any), subscriptions: make(map[string]subscription)}
	app.connStates[ws] = st
	return st, func() *Context {
		return &Context{app: app, wsConn: ws, Request: httptest.NewRequest("GET", "/", nil), pushCtx: life}
	}
}

func TestNonSubscriptionIgnoresSubscriptionID(t *testing.T) {
	for _, overwrite := range []bool{false, true} {
		t.Run(fmt.Sprintf("overwrite=%v", overwrite), func(t *testing.T) {
			app := NewApp()
			app.Page("/", func(*Context) *Node { return Div() })
			if overwrite {
				app.Subscription("save", func(*Context) error { return nil })
			}
			RegisterAction(app, "save", func(*Context, struct{}) (Result, error) { return Result{}, nil })
			st, newContext := subscriptionConnection(t, app)
			for i := range 100 {
				ctx := newContext()
				msg := wsMessage{Act: "save", Page: "/", Version: 1, Sub: fmt.Sprint(i) + strings.Repeat("x", 300)}
				if err := app.prepareAction(ctx, msg); err != nil {
					t.Fatal(err)
				}
				if len(st.subscriptions) != 0 || ctx.pushCtx != st.ctx || ctx.subscription != "" {
					t.Fatal("ordinary action retained subscription state")
				}
			}
		})
	}
}

func TestSubscriptionIDLength(t *testing.T) {
	for _, id := range []string{strings.Repeat("x", 256), strings.Repeat("é", 128), strings.Repeat("x", 257), strings.Repeat("é", 129)} {
		t.Run(fmt.Sprintf("bytes=%d", len(id)), func(t *testing.T) {
			app := NewApp()
			app.Page("/", func(*Context) *Node { return Div() })
			RegisterSubscription(app, "feed", func(*Context, struct{}) error { return nil })
			st, newContext := subscriptionConnection(t, app)
			ctx := newContext()
			err := app.prepareAction(ctx, wsMessage{Act: "feed", Page: "/", Version: 1, Sub: id})
			if len(id) > 256 {
				if err == nil {
					t.Fatal("over-long ID accepted")
				}
				if len(st.subscriptions) != 0 || ctx.pushCtx != st.ctx || ctx.subscription != "" || st.request != nil || st.version != 0 {
					t.Fatal("rejected frame retained state")
				}
			} else if err != nil || len(st.subscriptions) != 1 || ctx.pushCtx == st.ctx || ctx.subscription != id {
				t.Fatalf("valid ID rejected: %v", err)
			}
		})
	}
}

func TestSubscriptionConnectionLimit(t *testing.T) {
	for _, limit := range []int{0, -1, 2} {
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			app := NewApp()
			app.MaxSubscriptions = limit
			app.Page("/", func(*Context) *Node { return Div() })
			app.Subscription("feed", func(*Context) error { return nil })
			st, newContext := subscriptionConnection(t, app)
			if limit <= 0 {
				limit = 64
			}
			prepare := func(ctx *Context, id string) error {
				return app.prepareAction(ctx, wsMessage{Act: "feed", Page: "/", Version: 1, Sub: id})
			}
			var contexts []*Context
			for i := range limit {
				ctx := newContext()
				if err := prepare(ctx, fmt.Sprint(i)); err != nil {
					t.Fatal(err)
				}
				contexts = append(contexts, ctx)
			}
			for i := range 3 {
				ctx := newContext()
				if err := prepare(ctx, fmt.Sprintf("extra-%d", i)); err == nil {
					t.Fatal("subscription beyond limit accepted")
				}
				if len(st.subscriptions) != limit || ctx.pushCtx != st.ctx || ctx.subscription != "" {
					t.Fatal("rejected frame retained state")
				}
				for _, old := range contexts {
					if old.Context().Err() != nil {
						t.Fatal("limit evicted an existing subscription")
					}
				}
			}
			ctx := newContext()
			if err := prepare(ctx, "0"); err != nil || len(st.subscriptions) != limit {
				t.Fatalf("restart at capacity failed: %v", err)
			}
			if contexts[0].Context().Err() == nil || ctx.Context().Err() != nil {
				t.Fatal("restart did not replace the old context")
			}
			_, otherContext := subscriptionConnection(t, app)
			if err := prepare(otherContext(), "independent"); err != nil {
				t.Fatalf("another connection shared the limit: %v", err)
			}
			app.cancelSubscription(ctx.wsConn, "", "0")
			if ctx.Context().Err() == nil {
				t.Fatal("unsubscribe did not cancel")
			}
			if err := prepare(newContext(), "after-unsubscribe"); err != nil || len(st.subscriptions) != limit {
				t.Fatalf("unsubscribe did not free capacity: %v", err)
			}
		})
	}
}

func TestSubscriptionAuthorizationBeforeRegistration(t *testing.T) {
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	app.Subscription("feed", func(*Context) error { return nil })
	st, newContext := subscriptionConnection(t, app)
	allowed := true
	app.Authorize = func(ctx *Context, action string) error {
		if ctx.pushCtx != st.ctx || ctx.subscription != "" {
			t.Fatal("subscription context created before authorization")
		}
		if allowed {
			if len(st.subscriptions) != 0 {
				t.Fatal("subscription registered before authorization")
			}
			return nil
		}
		return Deny(Result{}.Toast("Expired"))
	}
	msg := wsMessage{Act: "feed", Page: "/", Version: 1, Sub: "existing"}
	old := newContext()
	if err := app.prepareAction(old, msg); err != nil {
		t.Fatal(err)
	}
	allowed = false
	for _, id := range []string{"existing", "new"} {
		msg.Sub = id
		ctx := newContext()
		if err := app.prepareAction(ctx, msg); err == nil {
			t.Fatal("unauthorized subscription accepted")
		}
		if len(st.subscriptions) != 1 || old.Context().Err() != nil || ctx.pushCtx != st.ctx || ctx.subscription != "" {
			t.Fatal("denial changed subscription state")
		}
	}
}

func TestSubscriptionLimitReplies(t *testing.T) {
	app := NewApp()
	app.MaxSubscriptions = 1
	app.Page("/", func(*Context) *Node { return Div() })
	app.Subscription("feed", func(*Context) error { return nil })
	ws := runtimeSocket(t, app)
	msg := wsMessage{Act: "feed", ID: 23, Page: "/", Version: 1, Sub: "valid"}
	if js := subscriptionFrame(t, ws, msg); js != "" {
		t.Fatalf("valid subscription rejected: %s", js)
	}
	for _, id := range []string{strings.Repeat("x", 257), "extra"} {
		msg.Sub = id
		if js := subscriptionFrame(t, ws, msg); js != Notify("error", "Request denied or page expired").script() {
			t.Fatalf("limit rejection missing default reply: %s", js)
		}
	}
	app.mu.RLock()
	defer app.mu.RUnlock()
	for _, st := range app.connStates {
		if len(st.subscriptions) != 1 {
			t.Fatal("rejected subscription retained an entry")
		}
		if _, ok := st.subscriptions["valid"]; !ok {
			t.Fatal("rejected subscription evicted an existing entry")
		}
	}
}

func subscriptionFrame(t *testing.T, ws *websocket.Conn, msg wsMessage) string {
	t.Helper()
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := websocket.Message.Send(ws, string(b)); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatal(err)
	}
	var reply struct {
		Reply   int    `json:"__r"`
		ID      int64  `json:"id"`
		Version int64  `json:"version"`
		JS      string `json:"js"`
	}
	if err := json.Unmarshal([]byte(raw), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Reply != 1 || reply.ID != msg.ID || reply.Version != msg.Version {
		t.Fatalf("invalid reply envelope: %s", raw)
	}
	return reply.JS
}

func TestAuthorizeDenialResult(t *testing.T) {
	for _, mode := range []string{"custom", "wrapped", "empty", "default", "failed-result"} {
		t.Run(mode, func(t *testing.T) {
			app := NewApp()
			app.Page("/", func(*Context) *Node { return Div() })
			var runs atomic.Int32
			app.Subscription("feed", func(*Context) error { runs.Add(1); return nil })
			result := Result{}.SetText("session", "Expired").Run(Redirect("/session/refresh"))
			app.Authorize = func(ctx *Context, action string) error {
				if action != "feed" || ctx.Request.URL.Path != "/" {
					t.Errorf("unexpected authorization context: %s %s", action, ctx.Request.URL.Path)
				}
				switch mode {
				case "wrapped":
					return fmt.Errorf("session: %w", Deny(result))
				case "empty":
					return Deny(Result{})
				case "default":
					return errors.New("private session error")
				case "failed-result":
					return Deny(Result{}.Morph("session", nil))
				default:
					return Deny(result)
				}
			}
			ws := runtimeSocket(t, app)
			js := subscriptionFrame(t, ws, wsMessage{Act: "feed", ID: 17, Page: "/", Version: 1, Sub: "denied"})
			switch mode {
			case "default":
				if js != Notify("error", "Request denied or page expired").script() {
					t.Fatal(js)
				}
			case "failed-result":
				if js != Notify("error", "Unable to complete the request").script() {
					t.Fatal(js)
				}
			case "empty":
				if js != "" {
					t.Fatalf("empty denial received effects: %s", js)
				}
			default:
				want, err := result.build(nil)
				if err != nil || js != want {
					t.Fatalf("custom denial missing: %s (%v)", js, err)
				}
			}
			if runs.Load() != 0 {
				t.Fatal("denied subscription started")
			}
			app.mu.RLock()
			defer app.mu.RUnlock()
			for _, st := range app.connStates {
				if len(st.subscriptions) != 0 || st.request != nil || st.version != 0 {
					t.Fatal("denial retained subscription or page state")
				}
			}
		})
	}
}
