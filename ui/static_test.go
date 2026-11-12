package ui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func renderShell(t *testing.T, app *App, path string) string {
	t.Helper()
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
	return rr.Body.String()
}

func TestPageTitleAndDescriptionInShell(t *testing.T) {
	app := NewApp()
	app.Title = "App title"
	app.Description = "App description"
	app.Page("/", func(*Context) *Node { return Div().Text("home") })
	app.Page("/about", func(*Context) *Node {
		return Div().Title("About <us>").Description(`Who "we" are`).Render(P().Text("text"))
	})

	home := renderShell(t, app, "/")
	if !strings.Contains(home, "<title>App title</title>") || !strings.Contains(home, `content="App description"`) {
		t.Fatalf("app defaults missing: %s", truncate(home, 800))
	}
	about := renderShell(t, app, "/about")
	if !strings.Contains(about, "<title>About &lt;us&gt;</title>") {
		t.Fatalf("page title missing: %s", truncate(about, 800))
	}
	if !strings.Contains(about, `<meta name="description" content="Who &#34;we&#34; are">`) || strings.Contains(about, "App description") {
		t.Fatalf("page description missing: %s", truncate(about, 800))
	}
	if strings.Contains(about, staticID) {
		t.Fatal("page without StaticHTML has an HTML copy")
	}
}

func TestStaticHTMLCopy(t *testing.T) {
	app := NewApp()
	app.Page("/", func(ctx *Context) *Node {
		ctx.StaticHTML()
		return Div("p-4").ID("page").Render(
			H1().Text("Hello <world> & you"),
			Button().Attr("onclick", "alert(1)").Attr("aria-label", `a"b`).OnClick(Toast("hi")).Text("Go"),
			El("script").Text("alert(2)"),
			Img().Attr("src", "/a.png").Style("maxWidth", "10px"),
			El("bad tag").Text("hidden"),
			SVG().Render(El("style").Text("</svg><script>alert(3)</script>")),
			Form().Render(Input().Attr("name", "email")),
		)
	})
	body := renderShell(t, app, "/")
	start := strings.Index(body, `<div id="`+staticID+`">`)
	if start < 0 {
		t.Fatalf("no HTML copy: %s", truncate(body, 800))
	}
	copyHTML := body[start : strings.Index(body[start:], "\n")+start]
	want := `<div id="__gsui_static"><div id="page" class="p-4"><h1>Hello &lt;world&gt; &amp; you</h1><button aria-label="a&#34;b" disabled="">Go</button><img src="/a.png" style="max-width:10px;"><svg></svg><form><input name="email" disabled=""></form></div></div>`
	if copyHTML != want {
		t.Fatalf("HTML copy\n got %s\nwant %s", copyHTML, want)
	}
	if !strings.Contains(body, "getElementById('"+staticID+"');if(s)s.remove();") {
		t.Fatal("client does not drop the HTML copy")
	}
	if !strings.Contains(body, staticRevealCSS) {
		t.Fatal("readers without JavaScript do not see the copy")
	}
}
