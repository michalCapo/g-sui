package ui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

const scriptBreakoutPayload = "</script><script>alert(1)</script>"

func TestTextEscapesScriptBreakoutForScriptContext(t *testing.T) {
	js := Div().Text(scriptBreakoutPayload).ToJS()

	assertNoRawScriptBreakout(t, js, scriptBreakoutPayload)
	expect(t, js, `\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e`)
}

func TestAttrEscapesScriptBreakoutForScriptContext(t *testing.T) {
	js := Div().Attr("title", scriptBreakoutPayload).ToJS()

	assertNoRawScriptBreakout(t, js, scriptBreakoutPayload)
}

func TestHelpersEscapeScriptBreakoutForScriptContext(t *testing.T) {
	helpers := map[string]string{
		"SetTitle":     SetTitle(scriptBreakoutPayload),
		"SetText":      SetText("target", scriptBreakoutPayload),
		"SetAttrValue": SetAttr("target", "title", scriptBreakoutPayload),
		"SetAttrName":  SetAttr("target", scriptBreakoutPayload, "value"),
		"Notify":       Notify("info", scriptBreakoutPayload),
	}

	for name, js := range helpers {
		t.Run(name, func(t *testing.T) {
			assertNoRawScriptBreakout(t, js, scriptBreakoutPayload)
		})
	}
}

func TestActionDataEscapesScriptBreakoutForScriptContext(t *testing.T) {
	js := Button().Text("go").OnClick(&Action{
		Name: "submit",
		Data: map[string]any{"payload": scriptBreakoutPayload},
	}).ToJS()

	assertNoRawScriptBreakout(t, js, scriptBreakoutPayload)
}

func TestAppHandlerEscapesRenderedTextInHTMLScript(t *testing.T) {
	app := NewApp()
	app.Page("/", func(ctx *Context) *Node {
		return Div().Text(scriptBreakoutPayload)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	app.Handler().ServeHTTP(rr, req)

	body := rr.Body.String()
	if rr.Code != 200 {
		t.Fatalf("expected HTTP 200, got %d: %s", rr.Code, body)
	}
	if strings.Contains(body, scriptBreakoutPayload) {
		t.Fatalf("HTML response contains raw script breakout payload: %s", truncate(body, 500))
	}
	expect(t, body, `\u003c/script\u003e\u003cscript\u003ealert(1)\u003c/script\u003e`)
}

func TestEscJSEscapesScriptContextSensitiveCharacters(t *testing.T) {
	got := escJS("\\'\n\r\t<>&=\u2028\u2029\x00\x1f")
	want := `\\\'\n\r\t\u003c\u003e\u0026\u003d\u2028\u2029\u0000\u001f`
	if got != want {
		t.Fatalf("escJS mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestAppShellMetadataIsHTMLEscaped(t *testing.T) {
	app := NewApp()
	app.Title = `</title><script>alert(1)</script>`
	app.Description = `desc"><script>alert(2)</script><meta name="x`
	app.Favicon = `/favicon.ico" onerror="alert(3)`
	app.Page("/", func(ctx *Context) *Node { return Div().Text("ok") })

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	app.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()

	if strings.Contains(body, app.Title) {
		t.Fatalf("title was not escaped: %s", truncate(body, 500))
	}
	if strings.Contains(body, app.Description) {
		t.Fatalf("description was not escaped: %s", truncate(body, 500))
	}
	if strings.Contains(body, app.Favicon) {
		t.Fatalf("favicon was not escaped: %s", truncate(body, 500))
	}
	notExpect(t, body, `</title><script>alert(1)</script>`)
	notExpect(t, body, `desc"><script>alert(2)</script>`)
	notExpect(t, body, `/favicon.ico" onerror="alert(3)`)
	expect(t, body, `&lt;/title&gt;&lt;script&gt;alert(1)&lt;/script&gt;`)
	expect(t, body, `/favicon.ico&#34; onerror=&#34;alert(3)`)
}

func TestMarkdownOmitsRawHTMLAndEscapesScriptBreakout(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"script-tag", `<script>alert(1)</script>`},
		{"javascript-link", `[x](javascript:alert(1))`},
		{"script-breakout", scriptBreakoutPayload},
		{"html-event-handler", `<img src=x onerror=alert(1)>`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			js := Markdown("prose", tc.content).ToJS()
			assertNoRawScriptBreakout(t, js, tc.content)
			notExpect(t, strings.ToLower(js), "javascript:alert")
			notExpect(t, strings.ToLower(js), "onerror=alert")
		})
	}
}

func assertNoRawScriptBreakout(t *testing.T, got, rawPayload string) {
	t.Helper()
	lower := strings.ToLower(got)
	if strings.Contains(lower, "</script") || strings.Contains(lower, "<script") {
		t.Fatalf("output contains raw script tag delimiter: %s", truncate(got, 500))
	}
	if rawPayload != "" && strings.Contains(got, rawPayload) {
		t.Fatalf("output contains raw payload %q: %s", rawPayload, truncate(got, 500))
	}
}
