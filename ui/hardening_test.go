package ui

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"golang.org/x/net/websocket"
)

func TestURLAttributesRefuseScriptSchemes(t *testing.T) {
	for _, tc := range []struct{ k, v, want string }{
		{"href", "/orders?id=1", "/orders?id=1"},
		{"href", "https://example.com/x", "https://example.com/x"},
		{"href", "mailto:a@b.c", "mailto:a@b.c"},
		{"href", "page:1", blockedURL},
		{"href", "javascript:alert(1)", blockedURL},
		{"href", " \tJava\nScript:alert(1)", blockedURL},
		{"src", "data:image/png;base64,AAAA", "data:image/png;base64,AAAA"},
		{"href", "data:text/html,<script>", blockedURL},
		{"src", "data:text/html,<script>", blockedURL},
		{"formaction", "vbscript:x", blockedURL},
		{"title", "javascript:alert(1)", "javascript:alert(1)"},
	} {
		if got := attrValue("img", tc.k, tc.v); got != tc.want {
			t.Errorf("attrValue(%q, %q) = %q, want %q", tc.k, tc.v, got, tc.want)
		}
	}
	if attrValue("iframe", "src", "data:image/svg+xml,x") != blockedURL || attrValue("iframe", "srcdoc", "<script>") != "" {
		t.Fatal("iframe data: URL or srcdoc kept")
	}
	js := A().Attr("href", "javascript:alert(1)").toJS()
	html := A().Attr("href", "javascript:alert(1)").staticHTML()
	if strings.Contains(js, "javascript") || strings.Contains(html, "javascript") {
		t.Fatalf("javascript: URL kept: %s / %s", js, html)
	}
	if strings.Contains(SetAttr("x", "href", "javascript:alert(1)").code(), "javascript") {
		t.Fatal("SetAttr kept a javascript: URL")
	}
}

func TestStaticHTMLStyleCannotAddDeclarations(t *testing.T) {
	html := Div().Style("color", "red;position:fixed").Style("fontFamily", `"Inter"`).staticHTML()
	if strings.Contains(html, "position") || !strings.Contains(html, "font-family:&#34;Inter&#34;") {
		t.Fatalf("style = %s", html)
	}
}

func TestPageRequestRejectsBackslash(t *testing.T) {
	base := httptest.NewRequest("GET", "/", nil)
	if _, err := pageRequest(base, `/\evil.example/x`); err == nil {
		t.Fatal(`/\evil.example must be rejected`)
	}
	if _, err := pageRequest(base, `/files?path=C:\docs#a\b`); err != nil {
		t.Fatalf("backslash in the query rejected: %v", err)
	}
}

func TestAssetsServeOnlyFiles(t *testing.T) {
	app := NewApp()
	app.Assets(fstest.MapFS{
		"assets/app.css":          {Data: []byte("body{}")},
		"assets/.env":             {Data: []byte("SECRET=1")},
		"assets/.git/config":      {Data: []byte("x")},
		"assets/docs/index.html":  {Data: []byte("<p>docs</p>")},
		"assets/private/a.txt":    {Data: []byte("a")},
		"assets/odd/index.html/x": {Data: []byte("x")},
	}, "assets", "/assets/")
	for path, want := range map[string]int{
		"/assets/app.css":     200,
		"/assets/":            404,
		"/assets/private/":    404,
		"/assets/.env":        404,
		"/assets/.git/config": 404,
		"/assets/docs/":       200,
		"/assets/odd/":        404,
	} {
		rr := httptest.NewRecorder()
		app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		if rr.Code != want {
			t.Errorf("GET %s = %d, want %d", path, rr.Code, want)
		}
	}
}

func TestFormValidateDoesNotTrustTheBrowser(t *testing.T) {
	f := NewForm("f")
	f.Text("Name", "Name").Required()
	f.SelectField("Role", "Role").Opts("user", "User", "editor", "Editor")
	f.Email("Mail", "Mail").PatternValidation(`[^@]+@example\.com`)
	type input struct {
		Name string `json:"name"`
		Role string `json:"role"`
		Mail string `json:"mail"`
	}
	errs := f.Validate(input{Name: "   ", Role: "admin", Mail: "evil@attacker.test"})
	for _, name := range []string{"Name", "Role", "Mail"} {
		if _, ok := errs[name]; !ok {
			t.Errorf("no error for %s: %v", name, errs)
		}
	}
	if errs := f.Validate(input{Name: "Ann", Role: "user", Mail: "ann@example.com"}); len(errs) != 0 {
		t.Fatalf("valid input rejected: %v", errs)
	}
}

func TestCollateClean(t *testing.T) {
	c := NewCollate[struct{}]("c").
		Sort(CollateSortField{Field: "name"}).
		Filter(CollateFilterField{Field: "status", Type: CollateSelect, Options: []CollateOption{{Value: "open"}}})
	order, limit, filters := c.Clean("name desc; DROP TABLE x", 1_000_000, []CollateFilterValue{
		{Field: "status", Type: "select", Value: "open"},
		{Field: "status", Type: "select", Value: "x' OR 1=1"},
		{Field: "secret", Type: "bool", Bool: true},
	})
	if order != "name asc" || limit != 100 || len(filters) != 1 {
		t.Fatalf("order %q limit %d filters %v", order, limit, filters)
	}
	if order, _, _ := c.Clean("salary desc", 0, nil); order != "" {
		t.Fatalf("unknown sort field kept: %q", order)
	}
	if order, _, _ := c.Clean("name desc", 0, nil); order != "name desc" {
		t.Fatalf("order %q", order)
	}
}

func TestTableFilterValuesAreChecked(t *testing.T) {
	text := &ColumnFilter{Type: FilterTypeText}
	sel := &ColumnFilter{Type: FilterTypeSelect, Options: []string{"a", "b"}}
	if validFilter(text, &FilterValue{Operator: "= 1 OR 1=1 --"}) || !validFilter(text, &FilterValue{Operator: "contains", Value: "x"}) {
		t.Fatal("text operator not checked")
	}
	if validFilter(sel, &FilterValue{Values: []string{"a", "zzz"}}) || !validFilter(sel, &FilterValue{Values: []string{"a", "b"}}) {
		t.Fatal("select values not checked")
	}
}

func TestEventAttributesAreNotSet(t *testing.T) {
	if js := Div().Attr("onclick", "alert(1)").Attr("ONload", "alert(2)").toJS(); strings.Contains(js, "alert") {
		t.Fatalf("event attribute set: %s", js)
	}
	if code := SetAttr("x", "onclick", "alert(1)").code(); strings.Contains(code, "alert") {
		t.Fatalf("SetAttr set an event attribute: %s", code)
	}
}

func TestBroadcastReachesOnlySignedIn(t *testing.T) {
	type user struct{ Name string }
	var revoked atomic.Bool
	app := NewApp()
	app.Page("/", func(*Context) *Node { return Div() })
	app.Identity = func(r *http.Request) (any, error) {
		if c, err := r.Cookie("s"); err == nil && c.Value == "ann" && !revoked.Load() {
			return &user{Name: "ann"}, nil
		}
		if r.Header.Get("X-Fail") != "" {
			return nil, errors.New("no session")
		}
		var none *user
		return none, nil // typed nil: not signed in
	}
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	t.Cleanup(func() { _ = app.Close() })
	dial := func(header map[string]string) *websocket.Conn {
		cfg, _ := websocket.NewConfig("ws"+strings.TrimPrefix(srv.URL, "http")+"/__ws", srv.URL)
		for k, v := range header {
			cfg.Header.Set(k, v)
		}
		ws, err := websocket.DialConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = ws.Close() })
		var hello string
		_ = ws.SetDeadline(time.Now().Add(5 * time.Second))
		if err := websocket.Message.Receive(ws, &hello); err != nil {
			t.Fatal(err)
		}
		return ws
	}
	signed := dial(map[string]string{"Cookie": "s=ann"})
	anon := dial(nil)
	failed := dial(map[string]string{"X-Fail": "1"})
	receive := func(ws *websocket.Conn) string {
		_ = ws.SetDeadline(time.Now().Add(300 * time.Millisecond))
		var raw string
		_ = websocket.Message.Receive(ws, &raw)
		return raw
	}
	if err := app.Broadcast(Result{}.Toast("private")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(receive(signed), "private") {
		t.Fatal("signed-in socket missed the broadcast")
	}
	for _, ws := range []*websocket.Conn{anon, failed} {
		if raw := receive(ws); raw != "" {
			t.Fatalf("anonymous socket got %s", raw)
		}
	}
	// The session ends after the handshake: later broadcasts skip it.
	revoked.Store(true)
	if err := app.Broadcast(Result{}.Toast("late")); err != nil {
		t.Fatal(err)
	}
	if raw := receive(signed); raw != "" {
		t.Fatalf("revoked session got %s", raw)
	}
	if err := app.BroadcastAll(Result{}.Toast("public")); err != nil {
		t.Fatal(err)
	}
	for _, ws := range []*websocket.Conn{signed, anon, failed} {
		if !strings.Contains(receive(ws), "public") {
			t.Fatal("BroadcastAll missed a socket")
		}
	}
}

func TestCSPNoncesEveryScript(t *testing.T) {
	app := NewApp()
	app.CSP = true
	app.HTMLHead = []string{`<SCRIPT src="/x.js"></SCRIPT>`, `<template data-x="<scripts>"></template>`}
	app.Page("/", func(ctx *Context) *Node {
		ctx.UnsafeHeadJS("window.x=1")
		return Div().Text("home")
	})
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	csp := rr.Header().Get("Content-Security-Policy")
	m := regexp.MustCompile(`'nonce-([^']+)'`).FindStringSubmatch(csp)
	if m == nil || !strings.Contains(csp, "object-src 'none'") {
		t.Fatalf("csp = %q", csp)
	}
	body := rr.Body.String()
	all := len(regexp.MustCompile(`(?i)<script[\s>]`).FindAllString(body, -1))
	nonced := len(regexp.MustCompile(`(?i)<script nonce="`+regexp.QuoteMeta(m[1])+`"`).FindAllString(body, -1))
	if all == 0 || all != nonced {
		t.Fatalf("%d of %d scripts have the nonce:\n%s", nonced, all, body)
	}
	if strings.Contains(body, " onload=") || strings.Contains(body, " onerror=") {
		t.Fatal("inline event attributes would be blocked by the policy")
	}
	if strings.Contains(body, "<scripts nonce") {
		t.Fatal("nonce added to a non-script tag")
	}
	// A policy set by middleware stays next to the library's.
	rr = httptest.NewRecorder()
	h := app.Handler()
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
		h.ServeHTTP(w, r)
	}).ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Values("Content-Security-Policy"); len(got) != 2 {
		t.Fatalf("policies = %q", got)
	}
	rr = httptest.NewRecorder()
	app.CSP = false
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Header().Get("Content-Security-Policy") != "" {
		t.Fatal("CSP must be off by default")
	}
}
