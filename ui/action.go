package ui

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

//go:embed local.js
var localJS string

// Action is what a node does on an event: a server call, a local UI step, or
// a sequence of both. Build it with typed helpers, not JavaScript:
//
//	save := ui.RegisterAction(app, "todo.save", handleSave)
//	ui.Button().Text("Save").OnClick(save.Call(input))
//	ui.Button().Text("Menu").OnClick(ui.Toggle("menu"))
//	ui.Button().Text("Delete").OnClick(ui.Confirm("Delete it?", remove.Call(input)))
type Action struct{ steps []step }

type step struct {
	call    string // server action name
	data    any
	collect []string
	js      string // local code; may use `event` and `el` (the element that handled the event)
	server  bool   // js wraps a server call
}

// UnsafeJS runs raw JavaScript as an action. `event` and `el` (the element
// that handled the event) are in scope. Prefer the typed helpers in this
// package; use this only for browser APIs they do not cover.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
func UnsafeJS(code string) *Action { return &Action{steps: []step{{js: code}}} }

func call(name string, data any) *Action {
	return &Action{steps: []step{{call: name, data: data}}}
}

// op builds a call to a local step in local.js. Arguments are JSON encoded.
func op(name string, args ...any) *Action {
	var b strings.Builder
	b.WriteString("__gsui.")
	b.WriteString(name)
	b.WriteString("(el")
	for _, arg := range args {
		data, err := json.Marshal(arg)
		if err != nil {
			panic(fmt.Errorf("gsui: %s argument: %w", name, err))
		}
		b.WriteByte(',')
		b.Write(data)
	}
	b.WriteString(");")
	return &Action{steps: []step{{js: b.String()}}}
}

// wrap puts an action's code inside a JS block, e.g. a condition or a timer.
func wrap(a *Action, before, after string) *Action {
	if a == nil {
		return nil
	}
	return &Action{steps: []step{{js: before + a.code() + after, server: a.server()}}}
}

// Collect reads the values of these element IDs before the server call and
// sends them by their name attribute. It applies to the server calls in a.
func (a *Action) Collect(ids ...string) *Action {
	if a == nil {
		return nil
	}
	out := &Action{steps: append([]step(nil), a.steps...)}
	for i := range out.steps {
		if out.steps[i].call != "" {
			out.steps[i].collect = append(append([]string(nil), out.steps[i].collect...), ids...)
		}
	}
	return out
}

func (a *Action) server() bool {
	for _, s := range a.steps {
		if s.call != "" || s.server {
			return true
		}
	}
	return false
}

// code returns JS statements. They expect `event` and `el` in scope.
func (a *Action) code() string {
	var b strings.Builder
	for _, s := range a.steps {
		if s.call == "" {
			b.WriteString(s.js)
			continue
		}
		name, _ := json.Marshal(s.call)
		data, err := json.Marshal(s.data)
		if err != nil {
			log.Printf("gsui: marshal action data: %v", err)
			data = []byte("{}")
		}
		collect := []byte("null")
		if len(s.collect) > 0 {
			collect, _ = json.Marshal(s.collect)
		}
		fmt.Fprintf(&b, "__ws.call(%s,%s,%s,el);", name, data, collect)
	}
	return b.String()
}

// script runs the action outside of an event, e.g. from a server Result.
func (a *Action) script() string { return "(function(event,el){" + a.code() + "})(null,null);" }

// ---------------------------------------------------------------------------
// Composition
// ---------------------------------------------------------------------------

// Seq runs actions in order. Nil actions are skipped.
func Seq(actions ...*Action) *Action {
	out := &Action{}
	for _, a := range actions {
		if a != nil {
			out.steps = append(out.steps, a.steps...)
		}
	}
	return out
}

// Confirm asks the user with a native confirm dialog and runs action on OK.
func Confirm(message string, action *Action) *Action {
	m, _ := json.Marshal(message)
	return wrap(action, "if(confirm("+string(m)+")){", "}")
}

// Delay runs action after d.
func Delay(d time.Duration, action *Action) *Action {
	return wrap(action, "setTimeout(function(){", fmt.Sprintf("},%d);", max(d.Milliseconds(), 0)))
}

// ---------------------------------------------------------------------------
// Local UI primitives. They run in the browser without a server round trip.
// ---------------------------------------------------------------------------

// Show reveals an element hidden with the hidden attribute or class.
func Show(id string) *Action { return op("show", id) }

// Hide hides an element with the hidden attribute and class.
func Hide(id string) *Action { return op("hide", id) }

// Toggle shows or hides an element and sets aria-expanded on the trigger.
func Toggle(id string) *Action { return op("toggle", id) }

// AddClass adds space-separated classes.
func AddClass(id, classes string) *Action { return op("addClass", id, classes) }

// RemoveClass removes space-separated classes.
func RemoveClass(id, classes string) *Action { return op("removeClass", id, classes) }

// ToggleClass toggles space-separated classes.
func ToggleClass(id, classes string) *Action { return op("toggleClass", id, classes) }

// SetAttr sets an attribute.
func SetAttr(id, name, value string) *Action { return op("setAttr", id, name, value) }

// RemoveAttr removes an attribute.
func RemoveAttr(id, name string) *Action { return op("removeAttr", id, name) }

// ToggleAttr toggles a boolean attribute such as disabled or open.
func ToggleAttr(id, name string) *Action { return op("toggleAttr", id, name) }

// SetText sets the text content.
func SetText(id, text string) *Action { return op("setText", id, text) }

// SetValue sets an input value and fires input and change events. For a
// checkbox or radio, a non-empty value checks it.
func SetValue(id, value string) *Action { return op("setValue", id, value) }

// Remove removes an element and runs its cleanup.
func Remove(id string) *Action { return op("remove", id) }

// Focus focuses an element and selects its text when possible.
func Focus(id string) *Action { return op("focus", id) }

// ScrollTo smoothly scrolls an element into view.
func ScrollTo(id string) *Action { return op("scroll", id) }

// ScrollTop smoothly scrolls the page to the top.
func ScrollTop() *Action { return op("scrollTop") }

// OpenDialog opens a <dialog> as a modal. Other elements are shown.
func OpenDialog(id string) *Action { return op("open", id) }

// CloseDialog closes a <dialog>. Other elements are hidden.
func CloseDialog(id string) *Action { return op("close", id) }

// CopyText copies text to the clipboard and shows a toast.
func CopyText(text string) *Action { return op("copy", text) }

// CopyFrom copies an input value or element text to the clipboard.
func CopyFrom(id string) *Action { return op("copy", "", id) }

// Notify shows a toast. Variants: "success", "error", "error-reload", "info".
func Notify(variant, message string) *Action { return op("notify", variant, message) }

// Toast shows a success toast.
func Toast(message string) *Action { return Notify("success", message) }

// Navigate opens a local page with live navigation.
func Navigate(path string) *Action { return op("nav", path, false, false) }

// PatchURL changes the query of the current page and keeps its live view.
// replace=true replaces the current history entry (useful for search input).
func PatchURL(path string, replace bool) *Action { return op("nav", path, true, replace) }

// Redirect loads a URL with a full page load.
func Redirect(url string) *Action { return op("redirect", url) }

// Reload reloads the page.
func Reload() *Action { return op("reload") }

// Back goes back in browser history.
func Back() *Action { return op("back") }

// Print opens the browser print dialog.
func Print() *Action { return op("print") }

// ResetForm resets a form to its initial values.
func ResetForm(id string) *Action { return op("reset", id) }

// SubmitForm submits a form like a submit button would, with validation.
func SubmitForm(id string) *Action { return op("submit", id) }

// TogglePassword switches a password input between hidden and visible text.
func TogglePassword(id string) *Action { return op("password", id) }

// ToggleTheme switches between light and dark theme.
func ToggleTheme() *Action { return op("theme", "") }

// SetTheme sets "light", "dark" or "system".
func SetTheme(mode string) *Action { return op("theme", mode) }

// SetTitle sets the document title.
func SetTitle(title string) *Action { return op("title", title) }

// ---------------------------------------------------------------------------
// Node behaviors
// ---------------------------------------------------------------------------

// OnChange runs action when the value of an input, select or textarea changes.
func (n *Node) OnChange(action *Action) *Node { return n.On("change", action) }

// OnKey runs action when key is pressed while the node has focus. Keys use
// KeyboardEvent.key names with modifiers: "Enter", "Escape", "ctrl+s",
// "mod+k" (Ctrl, or Cmd on Mac). Several OnKey calls can share a node.
func (n *Node) OnKey(key string, action *Action) *Node {
	k, _ := json.Marshal(key)
	return n.On("keydown", Seq(n.events["keydown"], wrap(action, "if(__gsui.key(event,"+string(k)+")){event.preventDefault();", "}")))
}

// Shortcut runs action on a page-wide key press while the node is mounted.
// Plain keys are ignored while the user types in a field. See OnKey for names.
func (n *Node) Shortcut(key string, action *Action) *Node {
	return n.mount("shortcut", key, action)
}

// OnOutsideClick runs action on a click outside this visible node, e.g. to
// close a menu. Clicks on the Toggle button for this node are ignored.
func (n *Node) OnOutsideClick(action *Action) *Node { return n.mount("outside", action) }

// DragToScroll lets the user drag with the mouse to scroll the node
// horizontally. Inputs, buttons and links inside keep working.
func (n *Node) DragToScroll() *Node { return n.mount("drag") }

// ActiveClass adds classes while the link's href is the current page, and
// sets aria-current. inactive classes are used otherwise. It updates after
// live navigation. Use with NavLink or A with an href.
func (n *Node) ActiveClass(active string, inactive ...string) *Node {
	n.Attr("data-gsui-active", active)
	if len(inactive) > 0 {
		n.Attr("data-gsui-inactive", inactive[0])
	}
	return n.mount("active")
}

// ActivePrefix also marks the link active on nested paths, e.g. "/users/7"
// for href "/users". Use with ActiveClass.
func (n *Node) ActivePrefix() *Node { return n.Attr("data-gsui-prefix", "") }

// mount runs a local step when the node is inserted. A trailing *Action
// argument becomes a callback with `event` and `el`.
func (n *Node) mount(name string, args ...any) *Node {
	var b strings.Builder
	b.WriteString("__gsui.")
	b.WriteString(name)
	b.WriteString("(this")
	for _, arg := range args {
		b.WriteByte(',')
		if a, ok := arg.(*Action); ok {
			if a == nil {
				return n
			}
			b.WriteString("function(event,el){")
			b.WriteString(a.code())
			b.WriteString("}")
			continue
		}
		data, err := json.Marshal(arg)
		if err != nil {
			panic(fmt.Errorf("gsui: %s argument: %w", name, err))
		}
		b.Write(data)
	}
	b.WriteString(");")
	n.rawJS += b.String()
	return n
}

// ---------------------------------------------------------------------------
// Widgets: third-party JavaScript with a mount and cleanup lifecycle
// ---------------------------------------------------------------------------

// Widget registers a client widget for third-party libraries such as charts
// or editors. mount is a JS function `function(el, props) { ...; return
// function cleanup() {} }`. It runs for every Widget node that is inserted,
// and cleanup runs when the node is removed.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
func (app *App) Widget(name, mount string) {
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.widgets == nil {
		app.widgets = make(map[string]string)
	}
	app.widgets[name] = mount
}

// Widget renders a node mounted by the client widget registered with
// App.Widget. props is JSON encoded. Morphs keep the widget's DOM.
func Widget(name string, props any, class ...string) *Node {
	return Div(class...).Preserve().mount("mount", name, props)
}

func (app *App) widgetJS() string {
	app.mu.RLock()
	defer app.mu.RUnlock()
	var b strings.Builder
	for name, mount := range app.widgets {
		k, _ := json.Marshal(name)
		fmt.Fprintf(&b, "__gsui.widgets[%s]=(%s);\n", k, mount)
	}
	return b.String()
}
