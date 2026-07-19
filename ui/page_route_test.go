package ui

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"
)

func TestPageSupportsServeMuxPathValues(t *testing.T) {
	app := NewApp()
	var gotPathValue, gotPathParam, gotPattern, gotQuery string
	app.Page("/dp/{token}", func(ctx *Context) *Node {
		gotPathValue = ctx.Request.PathValue("token")
		gotPathParam = ctx.PathParams["token"]
		gotPattern = ctx.Request.Pattern
		gotQuery = ctx.Query["stage"]
		return Div().Text("tax page")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.test/dp/abc-123?stage=documents", nil)
	app.Handler().ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if gotPathValue != "abc-123" {
		t.Fatalf("PathValue(token) = %q, want %q", gotPathValue, "abc-123")
	}
	if gotPathParam != "abc-123" {
		t.Fatalf("PathParams[token] = %q, want %q", gotPathParam, "abc-123")
	}
	if gotPattern != "GET /dp/{token}" {
		t.Fatalf("Request.Pattern = %q, want %q", gotPattern, "GET /dp/{token}")
	}
	if gotQuery != "documents" {
		t.Fatalf("Query[stage] = %q, want %q", gotQuery, "documents")
	}
}

func TestPageSupportsCatchAllPathValues(t *testing.T) {
	app := NewApp()
	var got string
	app.Page("/files/{path...}", func(ctx *Context) *Node {
		got = ctx.PathParams["path"]
		return Div().Text(got)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.test/files/reports/2026/tax.pdf", nil)
	app.Handler().ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if got != "reports/2026/tax.pdf" {
		t.Fatalf("catch-all path value = %q, want %q", got, "reports/2026/tax.pdf")
	}
}

func TestStaticRootPageRemainsExact(t *testing.T) {
	app := NewApp()
	app.Page("/", func(ctx *Context) *Node { return Div().Text("home") })

	home := httptest.NewRecorder()
	app.Handler().ServeHTTP(home, httptest.NewRequest("GET", "http://example.test/", nil))
	if home.Code != 200 {
		t.Fatalf("root page returned HTTP %d", home.Code)
	}

	missing := httptest.NewRecorder()
	app.Handler().ServeHTTP(missing, httptest.NewRequest("GET", "http://example.test/missing", nil))
	if missing.Code != 404 {
		t.Fatalf("unknown path returned HTTP %d, want 404", missing.Code)
	}
}

func TestWebSocketNavigationSupportsPagePathValues(t *testing.T) {
	app := NewApp()
	app.Page("/dp/{token}", func(ctx *Context) *Node {
		return Div().Text(ctx.Request.PathValue("token") + ":" + ctx.PathParams["token"])
	})

	server := httptest.NewServer(app.Handler())
	defer server.Close()

	config, err := websocket.NewConfig("ws"+strings.TrimPrefix(server.URL, "http")+"/__ws", server.URL)
	if err != nil {
		t.Fatalf("create WebSocket config: %v", err)
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		t.Fatalf("connect WebSocket: %v", err)
	}
	defer ws.Close()

	message := `{"act":"__nav","data":{"url":"/dp/ws-token"},"id":1}`
	if err := websocket.Message.Send(ws, message); err != nil {
		t.Fatalf("send navigation message: %v", err)
	}
	var raw string
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatalf("receive navigation response: %v", err)
	}
	var reply struct {
		Reply int    `json:"__r"`
		ID    int64  `json:"id"`
		JS    string `json:"js"`
	}
	if err := json.Unmarshal([]byte(raw), &reply); err != nil {
		t.Fatalf("decode navigation response %q: %v", raw, err)
	}

	if reply.Reply != 1 || reply.ID != 1 {
		t.Fatalf("unexpected navigation envelope: %+v", reply)
	}
	if !strings.Contains(reply.JS, "ws-token:ws-token") {
		t.Fatalf("navigation response does not contain both path values: %s", reply.JS)
	}
}
