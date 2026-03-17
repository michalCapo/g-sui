package ui

import (
	"fmt"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Node builder tests
// ---------------------------------------------------------------------------

func TestElBasic(t *testing.T) {
	n := Div("flex gap-4").ID("root").Text("hello")
	js := n.ToJS()

	expect(t, js, "document.createElement('div')")
	expect(t, js, ".id='root'")
	expect(t, js, ".className='flex gap-4'")
	expect(t, js, ".textContent='hello'")
	expect(t, js, "document.body.appendChild(")
}

func TestElNoClasses(t *testing.T) {
	n := Div().ID("empty")
	js := n.ToJS()

	expect(t, js, ".id='empty'")
	notExpect(t, js, ".className=")
}

func TestElWithRender(t *testing.T) {
	n := Div("parent").ID("parent").Render(
		Span("child-1").Text("first"),
		Span("child-2").Text("second"),
	)
	js := n.ToJS()

	count := strings.Count(js, "document.createElement")
	if count != 3 {
		t.Errorf("expected 3 createElement calls, got %d", count)
	}
	expect(t, js, "appendChild(e1)")
	expect(t, js, "appendChild(e2)")
}

func TestClassAppend(t *testing.T) {
	n := Div("flex").Class("gap-4")
	js := n.ToJS()

	expect(t, js, ".className='flex gap-4'")
}

func TestElWithAttributes(t *testing.T) {
	n := Input().Attr("type", "text").Attr("name", "username").Attr("placeholder", "Enter name")
	js := n.ToJS()

	expect(t, js, "setAttribute('type','text')")
	expect(t, js, "setAttribute('name','username')")
	expect(t, js, "setAttribute('placeholder','Enter name')")
}

func TestElWithStyles(t *testing.T) {
	n := Div().Style("color", "red").Style("font-size", "16px")
	js := n.ToJS()

	expect(t, js, ".style['color']='red'")
	expect(t, js, ".style['font-size']='16px'")
}

func TestElWithOnClick(t *testing.T) {
	n := Button("px-4 py-2").Text("Click me").OnClick(&Action{
		Name: "counter.increment",
		Data: map[string]any{"count": 0},
	})
	js := n.ToJS()

	expect(t, js, "addEventListener('click'")
	expect(t, js, "__ws.call('counter.increment'")
	expect(t, js, `"count":0`)
}

func TestElWithCollect(t *testing.T) {
	n := Button().Text("Submit").OnClick(&Action{
		Name:    "form.submit",
		Data:    map[string]any{},
		Collect: []string{"f-name", "f-email"},
	})
	js := n.ToJS()

	expect(t, js, "__ws.call('form.submit'")
	expect(t, js, `["f-name","f-email"]`)
}

func TestNilChildrenSkipped(t *testing.T) {
	n := Div().Render(
		Span().Text("visible"),
		nil,
		Span().Text("also visible"),
	)
	js := n.ToJS()

	count := strings.Count(js, "document.createElement")
	if count != 3 {
		t.Errorf("expected 3 createElement calls, got %d", count)
	}
}

func TestIfHelper(t *testing.T) {
	show := true
	n := Div().Render(
		If(show, Span().Text("shown")),
		If(!show, Span().Text("hidden")),
	)
	js := n.ToJS()

	expect(t, js, "'shown'")
	notExpect(t, js, "'hidden'")
}

func TestOrHelper(t *testing.T) {
	active := false
	n := Or(active,
		Span("text-green-500").Text("active"),
		Span("text-red-500").Text("inactive"),
	)
	js := n.ToJS()

	expect(t, js, "'inactive'")
	expect(t, js, "text-red-500")
	notExpect(t, js, "'active'")
}

func TestMapHelper(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry"}
	nodes := Map(items, func(item string, i int) *Node {
		return Li().ID(fmt.Sprintf("item-%d", i)).Text(item)
	})

	parent := Ul().ID("list").Render(nodes...)
	js := parent.ToJS()

	expect(t, js, "'Apple'")
	expect(t, js, "'Banana'")
	expect(t, js, "'Cherry'")
	expect(t, js, "item-0")
	expect(t, js, "item-2")
}

// ---------------------------------------------------------------------------
// Swap strategy tests
// ---------------------------------------------------------------------------

func TestToJSReplace(t *testing.T) {
	n := Div().ID("new-counter").Text("42")
	js := n.ToJSReplace("old-counter")

	expect(t, js, "getElementById('old-counter')")
	expect(t, js, "replaceWith(")
}

func TestToJSAppend(t *testing.T) {
	n := Li().Text("new item")
	js := n.ToJSAppend("list")

	expect(t, js, "getElementById('list')")
	expect(t, js, "_p.appendChild(")
}

func TestToJSPrepend(t *testing.T) {
	n := Li().Text("first item")
	js := n.ToJSPrepend("list")

	expect(t, js, "getElementById('list')")
	expect(t, js, "_p.prepend(")
}

func TestToJSInner(t *testing.T) {
	n := Span().Text("replaced content")
	js := n.ToJSInner("container")

	expect(t, js, "getElementById('container')")
	expect(t, js, "_t.innerHTML=''")
	expect(t, js, "_t.appendChild(")
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestNotify(t *testing.T) {
	js := Notify("success", "Saved!")
	expect(t, js, "'Saved!'")
	expect(t, js, "#16a34a")      // success accent color
	expect(t, js, "#dcfce7")      // success background
	expect(t, js, "__messages__") // toast container
	expect(t, js, "setTimeout")
}

func TestRedirect(t *testing.T) {
	js := Redirect("/dashboard")
	expect(t, js, "window.location.href='/dashboard'")
}

func TestSetTitle(t *testing.T) {
	js := SetTitle("New Title")
	expect(t, js, "document.title='New Title'")
}

func TestRemoveEl(t *testing.T) {
	js := RemoveEl("old-item")
	expect(t, js, "getElementById('old-item')")
	expect(t, js, "e.remove()")
}

func TestSetText(t *testing.T) {
	js := SetText("counter", "42")
	expect(t, js, "getElementById('counter')")
	expect(t, js, "e.textContent='42'")
}

func TestAddClass(t *testing.T) {
	js := AddClass("btn", "active")
	expect(t, js, "classList.add('active')")
}

func TestRemoveClass(t *testing.T) {
	js := RemoveClass("btn", "active")
	expect(t, js, "classList.remove('active')")
}

func TestShowHide(t *testing.T) {
	js := Show("panel")
	expect(t, js, "classList.remove('hidden')")

	js = Hide("panel")
	expect(t, js, "classList.add('hidden')")
}

// ---------------------------------------------------------------------------
// Response builder tests
// ---------------------------------------------------------------------------

func TestResponseBuilder(t *testing.T) {
	node := Div().ID("new-content").Text("Updated")

	js := NewResponse().
		Replace("content", node).
		Toast("success", "Done!").
		Build()

	expect(t, js, "replaceWith(")
	expect(t, js, "'Done!'")
	expect(t, js, "__messages__")
}

// ---------------------------------------------------------------------------
// Escaping tests
// ---------------------------------------------------------------------------

func TestEscapeQuotes(t *testing.T) {
	n := Span().Text("it's a test")
	js := n.ToJS()
	expect(t, js, `it\'s a test`)
}

func TestEscapeNewlines(t *testing.T) {
	n := Span().Text("line1\nline2")
	js := n.ToJS()
	expect(t, js, `line1\nline2`)
}

// ---------------------------------------------------------------------------
// Full example: Counter component
// ---------------------------------------------------------------------------

func TestCounterExample(t *testing.T) {
	count := 5
	counterNode := Div("flex gap-4 items-center p-8").ID("counter").Render(
		Button("w-10 h-10 rounded bg-red-600 text-white text-xl font-bold").
			Text("-").
			OnClick(&Action{
				Name: "counter.dec",
				Data: map[string]any{"Count": count},
			}),
		Span("text-4xl font-mono w-20 text-center").
			ID("counter-val").
			Text(fmt.Sprintf("%d", count)),
		Button("w-10 h-10 rounded bg-blue-600 text-white text-xl font-bold").
			Text("+").
			OnClick(&Action{
				Name: "counter.inc",
				Data: map[string]any{"Count": count},
			}),
	)

	js := counterNode.ToJS()

	expect(t, js, "document.createElement('div')")
	expect(t, js, "document.createElement('button')")
	expect(t, js, "document.createElement('span')")
	expect(t, js, "'counter.dec'")
	expect(t, js, "'counter.inc'")
	expect(t, js, "'5'")

	jsReplace := counterNode.ToJSReplace("counter")
	expect(t, jsReplace, "getElementById('counter')")
	expect(t, jsReplace, "replaceWith(")

	t.Logf("Counter JS (%d bytes):\n%s", len(js), js)
}

// ---------------------------------------------------------------------------
// Full example: Form with field collection
// ---------------------------------------------------------------------------

func TestFormExample(t *testing.T) {
	formNode := Div("max-w-sm mx-auto p-8 space-y-4").ID("login-form").Render(
		H2("text-2xl font-bold").Text("Login"),
		Div("space-y-1").Render(
			Label().Text("Email"),
			Input("w-full border rounded px-3 py-2").ID("f-email").
				Attr("type", "email").Attr("name", "Email"),
		),
		Div("space-y-1").Render(
			Label().Text("Password"),
			Input("w-full border rounded px-3 py-2").ID("f-pass").
				Attr("type", "password").Attr("name", "Password"),
		),
		Button("w-full bg-blue-600 text-white rounded py-2 font-bold").
			Text("Sign In").
			OnClick(&Action{
				Name:    "auth.login",
				Data:    map[string]any{},
				Collect: []string{"f-email", "f-pass"},
			}),
	)

	js := formNode.ToJS()

	expect(t, js, "'auth.login'")
	expect(t, js, `["f-email","f-pass"]`)
	expect(t, js, "setAttribute('type','email')")
	expect(t, js, "setAttribute('type','password')")

	t.Logf("Form JS (%d bytes):\n%s", len(js), js)
}

// ---------------------------------------------------------------------------
// Input with classes in constructor
// ---------------------------------------------------------------------------

func TestInputWithClasses(t *testing.T) {
	n := Input("w-full border rounded").Attr("type", "text")
	js := n.ToJS()

	expect(t, js, ".className='w-full border rounded'")
	expect(t, js, "setAttribute('type','text')")
}

// ---------------------------------------------------------------------------
// RawJS deferred execution: rawJS must come AFTER DOM insertion
// ---------------------------------------------------------------------------

func TestRawJSDeferredAfterAppend(t *testing.T) {
	n := Div().ID("container").JS("document.getElementById('container').dataset.ready='1';")
	js := n.ToJS()

	// The rawJS must appear AFTER the appendChild call
	appendIdx := strings.Index(js, "document.body.appendChild(")
	rawIdx := strings.Index(js, "document.getElementById('container').dataset.ready")
	if appendIdx < 0 {
		t.Fatal("expected appendChild call in output")
	}
	if rawIdx < 0 {
		t.Fatal("expected rawJS in output")
	}
	if rawIdx < appendIdx {
		t.Errorf("rawJS (at %d) must come after appendChild (at %d)", rawIdx, appendIdx)
	}
}

func TestRawJSDeferredInner(t *testing.T) {
	n := Div().ID("child").JS("document.getElementById('child').style.color='red';")
	js := n.ToJSInner("target")

	// The rawJS must appear AFTER the _t.appendChild call
	appendIdx := strings.Index(js, "_t.appendChild(")
	rawIdx := strings.Index(js, "document.getElementById('child').style.color")
	if appendIdx < 0 {
		t.Fatal("expected _t.appendChild call in output")
	}
	if rawIdx < 0 {
		t.Fatal("expected rawJS in output")
	}
	if rawIdx < appendIdx {
		t.Errorf("rawJS (at %d) must come after appendChild (at %d)", rawIdx, appendIdx)
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func expect(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q\ngot: %s", want, truncate(got, 300))
	}
}

func notExpect(t *testing.T, got, notWant string) {
	t.Helper()
	if strings.Contains(got, notWant) {
		t.Errorf("expected output NOT to contain %q\ngot: %s", notWant, truncate(got, 300))
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
