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
	js := n.toJS()

	expect(t, js, "document.createElement('div')")
	expect(t, js, ".id='root'")
	expect(t, js, ".className='flex gap-4'")
	expect(t, js, ".textContent='hello'")
	expect(t, js, "document.body.appendChild(")
}

func TestElNoClasses(t *testing.T) {
	n := Div().ID("empty")
	js := n.toJS()

	expect(t, js, ".id='empty'")
	notExpect(t, js, ".className=")
}

func TestElWithRender(t *testing.T) {
	n := Div("parent").ID("parent").Render(
		Span("child-1").Text("first"),
		Span("child-2").Text("second"),
	)
	js := n.toJS()

	count := strings.Count(js, "document.createElement")
	if count != 3 {
		t.Errorf("expected 3 createElement calls, got %d", count)
	}
	expect(t, js, "appendChild(e1)")
	expect(t, js, "appendChild(e2)")
}

func TestClassAppend(t *testing.T) {
	n := Div("flex").Class("gap-4")
	js := n.toJS()

	expect(t, js, ".className='flex gap-4'")
}

func TestElWithAttributes(t *testing.T) {
	n := Input().Attr("type", "text").Attr("name", "username").Attr("placeholder", "Enter name")
	js := n.toJS()

	expect(t, js, "setAttribute('type','text')")
	expect(t, js, "setAttribute('name','username')")
	expect(t, js, "setAttribute('placeholder','Enter name')")
}

func TestElWithStyles(t *testing.T) {
	n := Div().Style("color", "red").Style("font-size", "16px")
	js := n.toJS()

	expect(t, js, ".style['color']='red'")
	expect(t, js, ".style['font-size']='16px'")
}

func TestElWithOnClick(t *testing.T) {
	n := Button("px-4 py-2").Text("Click me").OnClick(call("counter.increment", map[string]any{"count": 0}))
	js := n.toJS()

	expect(t, js, "_bind(e0,'click'")
	expect(t, js, `__ws.call("counter.increment"`)
	expect(t, js, "event.preventDefault()")
	expect(t, js, `"count":0`)
}

func TestElWithCollect(t *testing.T) {
	n := Button().Text("Submit").OnClick(call("form.submit", map[string]any{}).Collect("f-name", "f-email"))
	js := n.toJS()

	expect(t, js, `__ws.call("form.submit"`)
	expect(t, js, `["f-name","f-email"]`)
}

func TestNilChildrenSkipped(t *testing.T) {
	n := Div().Render(
		Span().Text("visible"),
		nil,
		Span().Text("also visible"),
	)
	js := n.toJS()

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
	js := n.toJS()

	expect(t, js, "'shown'")
	notExpect(t, js, "'hidden'")
}

func TestOrHelper(t *testing.T) {
	active := false
	n := Or(active,
		Span("text-green-500").Text("active"),
		Span("text-red-500").Text("inactive"),
	)
	js := n.toJS()

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
	js := parent.toJS()

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
	js := n.toJSReplace("old-counter")

	expect(t, js, "getElementById('old-counter')")
	expect(t, js, "replaceWith(")
}

func TestToJSAppend(t *testing.T) {
	n := Li().Text("new item")
	js := n.toJSAppend("list")

	expect(t, js, "getElementById('list')")
	expect(t, js, "_p.appendChild(")
}

func TestToJSPrepend(t *testing.T) {
	n := Li().Text("first item")
	js := n.toJSPrepend("list")

	expect(t, js, "getElementById('list')")
	expect(t, js, "_p.prepend(")
}

func TestToJSInner(t *testing.T) {
	n := Span().Text("replaced content")
	js := n.toJSInner("container")

	expect(t, js, "getElementById('container')")
	expect(t, js, "_t.innerHTML=''")
	expect(t, js, "_t.appendChild(")
}

// ---------------------------------------------------------------------------
// Helper function tests
// ---------------------------------------------------------------------------

func TestLocalActions(t *testing.T) {
	cases := map[string]struct {
		action *Action
		want   string
	}{
		"Notify":      {Notify("success", "Saved!"), `__gsui.notify(el,"success","Saved!");`},
		"Redirect":    {Redirect("/dashboard"), `__gsui.redirect(el,"/dashboard");`},
		"SetTitle":    {SetTitle("New Title"), `__gsui.title(el,"New Title");`},
		"Remove":      {Remove("old-item"), `__gsui.remove(el,"old-item");`},
		"SetText":     {SetText("counter", "42"), `__gsui.setText(el,"counter","42");`},
		"AddClass":    {AddClass("btn", "active"), `__gsui.addClass(el,"btn","active");`},
		"RemoveClass": {RemoveClass("btn", "active"), `__gsui.removeClass(el,"btn","active");`},
		"Show":        {Show("panel"), `__gsui.show(el,"panel");`},
		"Hide":        {Hide("panel"), `__gsui.hide(el,"panel");`},
		"Toggle":      {Toggle("menu"), `__gsui.toggle(el,"menu");`},
		"PatchURL":    {PatchURL("/q?x=1", true), `__gsui.nav(el,"/q?x=1",true,true);`},
	}
	for name, c := range cases {
		if got := c.action.code(); got != c.want {
			t.Errorf("%s: got %s, want %s", name, got, c.want)
		}
	}
}

func TestLocalClickDoesNotPreventDefault(t *testing.T) {
	js := NavLink("/x").OnClick(Hide("menu")).toJS()
	notExpect(t, js, "preventDefault")
}

func TestActionComposition(t *testing.T) {
	save := call("item.save", map[string]any{"id": 1})
	js := Button().OnClick(Seq(Hide("menu"), Confirm("Sure?", save.Collect("name")))).toJS()
	expect(t, js, `__gsui.hide(el,"menu");if(confirm("Sure?")){__ws.call("item.save",{"id":1},["name"],el);}`)
	expect(t, js, "event.preventDefault()")

	js = Input().OnKey("Enter", save).OnKey("Escape", Hide("menu")).toJS()
	expect(t, js, `if(__gsui.key(event,"Enter"))`)
	expect(t, js, `if(__gsui.key(event,"Escape"))`)

	js = Input().OnInput(save).toJS()
	expect(t, js, "__gsui.debounce(el,200,function(){")
}

func TestNodeBehaviors(t *testing.T) {
	js := Div().ID("menu").OnOutsideClick(Hide("menu")).Shortcut("Escape", Hide("menu")).DragToScroll().toJS()
	expect(t, js, `__gsui.outside(this,function(event,el){__gsui.hide(el,"menu");});`)
	expect(t, js, `__gsui.shortcut(this,"Escape",function(event,el){__gsui.hide(el,"menu");});`)
	expect(t, js, `__gsui.drag(this);`)

	js = NavLink("/users", "px-2").ActiveClass("font-bold", "text-gray-500").ActivePrefix().toJS()
	expect(t, js, "setAttribute('data-gsui-active','font-bold')")
	expect(t, js, "setAttribute('data-gsui-inactive','text-gray-500')")
	expect(t, js, "__gsui.active(this);")

	js = Widget("chart", map[string]any{"points": []int{1, 2}}).toJS()
	expect(t, js, `__gsui.mount(this,"chart",{"points":[1,2]});`)
	expect(t, js, "data-gsui-preserve")
}

func TestResultRun(t *testing.T) {
	js, err := Result{}.Run(CloseDialog("edit"), Focus("name")).build(nil)
	if err != nil {
		t.Fatal(err)
	}
	expect(t, js, `(function(event,el){__gsui.close(el,"edit");__gsui.focus(el,"name");})(null,null);`)
}

// ---------------------------------------------------------------------------
// Result tests
// ---------------------------------------------------------------------------

func TestResultEffects(t *testing.T) {
	node := Div().ID("new-content").Text("Updated")

	js, err := (Result{}).Replace("content", node).Toast("Done!").build(nil)
	if err != nil {
		t.Fatal(err)
	}

	expect(t, js, "replaceWith(")
	expect(t, js, `__gsui.notify(el,"success","Done!")`)
}

// ---------------------------------------------------------------------------
// Escaping tests
// ---------------------------------------------------------------------------

func TestEscapeQuotes(t *testing.T) {
	n := Span().Text("it's a test")
	js := n.toJS()
	expect(t, js, `it\'s a test`)
}

func TestEscapeNewlines(t *testing.T) {
	n := Span().Text("line1\nline2")
	js := n.toJS()
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
			OnClick(call("counter.dec", map[string]any{"Count": count})),
		Span("text-4xl font-mono w-20 text-center").
			ID("counter-val").
			Text(fmt.Sprintf("%d", count)),
		Button("w-10 h-10 rounded bg-blue-600 text-white text-xl font-bold").
			Text("+").
			OnClick(call("counter.inc", map[string]any{"Count": count})),
	)

	js := counterNode.toJS()

	expect(t, js, "document.createElement('div')")
	expect(t, js, "document.createElement('button')")
	expect(t, js, "document.createElement('span')")
	expect(t, js, `"counter.dec"`)
	expect(t, js, `"counter.inc"`)
	expect(t, js, "'5'")

	jsReplace := counterNode.toJSReplace("counter")
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
			OnClick(call("auth.login", map[string]any{}).Collect("f-email", "f-pass")),
	)

	js := formNode.toJS()

	expect(t, js, `"auth.login"`)
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
	js := n.toJS()

	expect(t, js, ".className='w-full border rounded'")
	expect(t, js, "setAttribute('type','text')")
}

// ---------------------------------------------------------------------------
// RawJS deferred execution: rawJS must come AFTER DOM insertion
// ---------------------------------------------------------------------------

func TestRawJSDeferredAfterAppend(t *testing.T) {
	n := Div().ID("container").UnsafeJS("document.getElementById('container').dataset.ready='1';")
	js := n.toJS()

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
	n := Div().ID("child").UnsafeJS("document.getElementById('child').style.color='red';")
	js := n.toJSInner("target")

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
// RawJS `this` binding: `.UnsafeJS()` code can use `this` to reference the element
// ---------------------------------------------------------------------------

func TestRawJSThisBinding(t *testing.T) {
	n := El("svg", "w-6 h-6").Attr("viewBox", "0 0 24 24").
		UnsafeJS("this.innerHTML='<circle cx=\"12\" cy=\"12\" r=\"10\"/>'")
	js := n.toJS()

	// Must use .call(eN) so `this` is the element
	expect(t, js, ".call(e0)")
	// The raw JS must still be deferred after appendChild
	appendIdx := strings.Index(js, "document.body.appendChild(")
	callIdx := strings.Index(js, ".call(e0)")
	if appendIdx < 0 || callIdx < 0 {
		t.Fatal("expected both appendChild and .call in output")
	}
	if callIdx < appendIdx {
		t.Errorf(".call (at %d) must come after appendChild (at %d)", callIdx, appendIdx)
	}
}

func TestRawJSThisBindingNested(t *testing.T) {
	// Child node with .UnsafeJS() — `this` should reference the child (e1), not parent (e0)
	n := Div("parent").Render(
		Span("child").UnsafeJS("this.dataset.init='1'"),
	)
	js := n.toJS()

	expect(t, js, ".call(e1)")
	notExpect(t, js, ".call(e0)")
}

// ---------------------------------------------------------------------------
// SVG namespace: svg elements must use createElementNS
// ---------------------------------------------------------------------------

func TestSVGUsesCreateElementNS(t *testing.T) {
	n := SVG("w-6 h-6").Attr("viewBox", "0 0 24 24")
	js := n.toJS()

	expect(t, js, "createElementNS('http://www.w3.org/2000/svg','svg')")
	notExpect(t, js, "createElement('svg')")
	// class must use setAttribute for SVG elements
	expect(t, js, "setAttribute('class','w-6 h-6')")
	notExpect(t, js, ".className=")
}

func TestSVGChildrenInheritNamespace(t *testing.T) {
	n := SVG("w-6 h-6").Render(
		El("path").Attr("d", "M0 0L10 10"),
		El("circle").Attr("cx", "5").Attr("cy", "5").Attr("r", "3"),
	)
	js := n.toJS()

	expect(t, js, "createElementNS('http://www.w3.org/2000/svg','svg')")
	expect(t, js, "createElementNS('http://www.w3.org/2000/svg','path')")
	expect(t, js, "createElementNS('http://www.w3.org/2000/svg','circle')")
	notExpect(t, js, "createElement('path')")
	notExpect(t, js, "createElement('circle')")
}

func TestNonSVGStillUsesCreateElement(t *testing.T) {
	n := Div("flex").Render(Span("text"))
	js := n.toJS()

	expect(t, js, "createElement('div')")
	expect(t, js, "createElement('span')")
	notExpect(t, js, "createElementNS")
}

func TestFormShowErrorsUsesFieldErrorIDs(t *testing.T) {
	f := NewForm("contact").Text("Email", "email").Required().Render().Checkbox("Terms", "terms").Required().Render()
	js, err := f.ShowErrors(FormErrors{"email": "already registered"}).build(nil)
	if err != nil {
		t.Fatal(err)
	}
	expect(t, js, "err-contact-email")
	expect(t, js, "err-contact-terms")
	expect(t, js, "already registered")
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
