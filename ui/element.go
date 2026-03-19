// Package ui provides a server-rendered UI framework where Go builds
// typed DOM node trees that compile to pure JavaScript strings.
// The browser receives raw JS that performs document.createElement calls
// directly -- no HTML, no JSON intermediate, no client-side framework.
//
// SVG elements (svg, path, circle, rect, etc.) are automatically created
// with document.createElementNS using the SVG namespace. Child elements
// of an SVG root inherit the namespace. Classes on SVG elements use
// setAttribute('class', ...) instead of .className for compatibility.
package ui

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// Node: the core building block
// ---------------------------------------------------------------------------

// Node represents a DOM element built in Go that compiles to JavaScript.
type Node struct {
	tag      string
	id       string
	class    string
	text     string
	attrs    map[string]string
	styles   map[string]string
	children []*Node
	events   map[string]*Action
	rawJS    string // arbitrary JS executed after this node is mounted
	void     bool   // self-closing element (input, img, br, hr)
}

// Action describes a server-side handler invoked via WebSocket,
// or a client-side JS snippet when created via JS().
type Action struct {
	Name    string         // e.g. "counter.increment"
	Data    map[string]any // state payload sent with the call
	Collect []string       // element IDs whose .value to collect before calling
	rawJS   string         // if set, execute client-side JS instead of WS call
}

// JS creates a client-side-only Action that executes raw JavaScript
// instead of calling the server via WebSocket.
//
//	r.Button("...").OnClick(r.JS("history.back()"))
func JS(code string) *Action {
	return &Action{rawJS: code}
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// El creates a node with an arbitrary tag and optional CSS class string.
//
//	El("section", "max-w-5xl mx-auto")
func El(tag string, class ...string) *Node {
	c := ""
	if len(class) > 0 {
		c = class[0]
	}
	return &Node{tag: tag, class: c}
}

// Convenience constructors. All accept an optional single class string.
//
//	Div("flex gap-4 items-center")
//	Button("px-4 py-2 bg-blue-600 text-white")
//	Span()  // no classes
func Div(class ...string) *Node      { return El("div", class...) }
func Span(class ...string) *Node     { return El("span", class...) }
func Button(class ...string) *Node   { return El("button", class...) }
func H1(class ...string) *Node       { return El("h1", class...) }
func H2(class ...string) *Node       { return El("h2", class...) }
func H3(class ...string) *Node       { return El("h3", class...) }
func H4(class ...string) *Node       { return El("h4", class...) }
func H5(class ...string) *Node       { return El("h5", class...) }
func H6(class ...string) *Node       { return El("h6", class...) }
func P(class ...string) *Node        { return El("p", class...) }
func A(class ...string) *Node        { return El("a", class...) }
func Nav(class ...string) *Node      { return El("nav", class...) }
func Main(class ...string) *Node     { return El("main", class...) }
func Header(class ...string) *Node   { return El("header", class...) }
func Footer(class ...string) *Node   { return El("footer", class...) }
func Section(class ...string) *Node  { return El("section", class...) }
func Article(class ...string) *Node  { return El("article", class...) }
func Aside(class ...string) *Node    { return El("aside", class...) }
func Form(class ...string) *Node     { return El("form", class...) }
func Pre(class ...string) *Node      { return El("pre", class...) }
func Code(class ...string) *Node     { return El("code", class...) }
func Ul(class ...string) *Node       { return El("ul", class...) }
func Ol(class ...string) *Node       { return El("ol", class...) }
func Li(class ...string) *Node       { return El("li", class...) }
func Label(class ...string) *Node    { return El("label", class...) }
func Textarea(class ...string) *Node { return El("textarea", class...) }
func Select(class ...string) *Node   { return El("select", class...) }
func Option(class ...string) *Node   { return El("option", class...) }
func SVG(class ...string) *Node      { return El("svg", class...) }

// Table elements
func Table(class ...string) *Node { return El("table", class...) }
func Thead(class ...string) *Node { return El("thead", class...) }
func Tbody(class ...string) *Node { return El("tbody", class...) }
func Tfoot(class ...string) *Node { return El("tfoot", class...) }
func Tr(class ...string) *Node    { return El("tr", class...) }
func Th(class ...string) *Node    { return El("th", class...) }
func Td(class ...string) *Node    { return El("td", class...) }

// Media / embed
func Video(class ...string) *Node  { return El("video", class...) }
func Audio(class ...string) *Node  { return El("audio", class...) }
func Canvas(class ...string) *Node { return El("canvas", class...) }

// Inline text
func Strong(class ...string) *Node { return El("strong", class...) }
func Em(class ...string) *Node     { return El("em", class...) }
func Small(class ...string) *Node  { return El("small", class...) }
func B(class ...string) *Node      { return El("b", class...) }
func I(class ...string) *Node      { return El("i", class...) }
func U(class ...string) *Node      { return El("u", class...) }
func Sub(class ...string) *Node    { return El("sub", class...) }
func Sup(class ...string) *Node    { return El("sup", class...) }
func Mark(class ...string) *Node   { return El("mark", class...) }
func Abbr(class ...string) *Node   { return El("abbr", class...) }
func Time(class ...string) *Node   { return El("time", class...) }

// Block content
func Blockquote(class ...string) *Node { return El("blockquote", class...) }
func Figure(class ...string) *Node     { return El("figure", class...) }
func Figcaption(class ...string) *Node { return El("figcaption", class...) }
func Dl(class ...string) *Node         { return El("dl", class...) }
func Dt(class ...string) *Node         { return El("dt", class...) }
func Dd(class ...string) *Node         { return El("dd", class...) }

// Forms (extended)
func Fieldset(class ...string) *Node { return El("fieldset", class...) }
func Legend(class ...string) *Node   { return El("legend", class...) }
func Optgroup(class ...string) *Node { return El("optgroup", class...) }
func Datalist(class ...string) *Node { return El("datalist", class...) }
func Output(class ...string) *Node   { return El("output", class...) }
func Progress(class ...string) *Node { return El("progress", class...) }
func Meter(class ...string) *Node    { return El("meter", class...) }

// Interactive
func Details(class ...string) *Node { return El("details", class...) }
func Summary(class ...string) *Node { return El("summary", class...) }
func Dialog(class ...string) *Node  { return El("dialog", class...) }

// Embed
func Iframe(class ...string) *Node  { return El("iframe", class...) }
func Object(class ...string) *Node  { return El("object", class...) }
func Picture(class ...string) *Node { return El("picture", class...) }

// Table (extended)
func Caption(class ...string) *Node  { return El("caption", class...) }
func Colgroup(class ...string) *Node { return El("colgroup", class...) }

// Void elements (self-closing)
func Input(class ...string) *Node  { return voidEl("input", class...) }
func Img(class ...string) *Node    { return voidEl("img", class...) }
func Br() *Node                    { return &Node{tag: "br", void: true} }
func Hr() *Node                    { return &Node{tag: "hr", void: true} }
func Source(class ...string) *Node { return voidEl("source", class...) }
func Embed(class ...string) *Node  { return voidEl("embed", class...) }
func Col(class ...string) *Node    { return voidEl("col", class...) }
func Wbr() *Node                   { return &Node{tag: "wbr", void: true} }
func Link() *Node                  { return &Node{tag: "link", void: true} }
func Meta() *Node                  { return &Node{tag: "meta", void: true} }

// Typed input constructors — shorthand for Input(<class>).Attr("type", "<type>").
func IText(class ...string) *Node     { return Input(class...).Attr("type", "text") }
func IPassword(class ...string) *Node { return Input(class...).Attr("type", "password") }
func IEmail(class ...string) *Node    { return Input(class...).Attr("type", "email") }
func IPhone(class ...string) *Node    { return Input(class...).Attr("type", "tel") }
func INumber(class ...string) *Node   { return Input(class...).Attr("type", "number") }
func ISearch(class ...string) *Node   { return Input(class...).Attr("type", "search") }
func IUrl(class ...string) *Node      { return Input(class...).Attr("type", "url") }
func IDate(class ...string) *Node     { return Input(class...).Attr("type", "date") }
func ITime(class ...string) *Node     { return Input(class...).Attr("type", "time") }
func IDatetime(class ...string) *Node { return Input(class...).Attr("type", "datetime-local") }
func IFile(class ...string) *Node     { return Input(class...).Attr("type", "file") }
func ICheckbox(class ...string) *Node { return Input(class...).Attr("type", "checkbox") }
func IRadio(class ...string) *Node    { return Input(class...).Attr("type", "radio") }
func IRange(class ...string) *Node    { return Input(class...).Attr("type", "range") }
func IColor(class ...string) *Node    { return Input(class...).Attr("type", "color") }
func IHidden(class ...string) *Node   { return Input(class...).Attr("type", "hidden") }
func ISubmit(class ...string) *Node   { return Input(class...).Attr("type", "submit") }
func IReset(class ...string) *Node    { return Input(class...).Attr("type", "reset") }
func IArea(class ...string) *Node     { return Textarea(class...) }

// voidEl creates a void (self-closing) element with optional class.
func voidEl(tag string, class ...string) *Node {
	c := ""
	if len(class) > 0 {
		c = class[0]
	}
	return &Node{tag: tag, void: true, class: c}
}

// ---------------------------------------------------------------------------
// Builder (chainable)
// ---------------------------------------------------------------------------

// ID sets the element id attribute.
func (n *Node) ID(id string) *Node { n.id = id; return n }

// Class appends additional CSS classes (space-separated) to any classes
// already set via the constructor. Useful for conditional class additions.
func (n *Node) Class(cls string) *Node {
	if n.class == "" {
		n.class = cls
	} else {
		n.class = n.class + " " + cls
	}
	return n
}

// Text sets the textContent.
func (n *Node) Text(t string) *Node { n.text = t; return n }

// Attr sets an arbitrary HTML attribute.
func (n *Node) Attr(key, val string) *Node {
	if n.attrs == nil {
		n.attrs = make(map[string]string)
	}
	n.attrs[key] = val
	return n
}

// Style sets an inline style property.
func (n *Node) Style(key, val string) *Node {
	if n.styles == nil {
		n.styles = make(map[string]string)
	}
	n.styles[key] = val
	return n
}

// Render appends child nodes and returns this node. This is the primary
// way to compose a node tree. Nil children are silently skipped.
//
//	Div("flex", "gap-4").Render(
//	    H1("text-3xl").Text("Title"),
//	    P("text-gray-600").Text("Description"),
//	)
func (n *Node) Render(children ...*Node) *Node {
	for _, c := range children {
		if c != nil {
			n.children = append(n.children, c)
		}
	}
	return n
}

// OnClick attaches a click event that calls a server action via WS.
func (n *Node) OnClick(action *Action) *Node { return n.On("click", action) }

// OnSubmit attaches a submit event action.
func (n *Node) OnSubmit(action *Action) *Node { return n.On("submit", action) }

// On attaches a named event to a server action.
func (n *Node) On(event string, action *Action) *Node {
	if n.events == nil {
		n.events = make(map[string]*Action)
	}
	n.events[event] = action
	return n
}

// JS sets raw JavaScript to execute after this node is appended to the DOM.
func (n *Node) JS(raw string) *Node { n.rawJS = raw; return n }

// ---------------------------------------------------------------------------
// Conditional helpers
// ---------------------------------------------------------------------------

// If returns the node only when cond is true, otherwise nil.
func If(cond bool, node *Node) *Node {
	if cond {
		return node
	}
	return nil
}

// Or returns yes when cond is true, no otherwise.
func Or(cond bool, yes, no *Node) *Node {
	if cond {
		return yes
	}
	return no
}

// Map iterates a slice, calls fn for each item, and returns a parent with
// the results as children. Useful for rendering lists.
func Map[T any](items []T, fn func(T, int) *Node) []*Node {
	out := make([]*Node, 0, len(items))
	for i, item := range items {
		if node := fn(item, i); node != nil {
			out = append(out, node)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// ID generation
// ---------------------------------------------------------------------------

// Target generates a random, collision-resistant ID string suitable for use
// as an HTML id attribute (e.g. "t-a1b2c3d4e5f6").
func Target() string {
	b := make([]byte, 6) // 6 bytes → 12 hex chars
	if _, err := rand.Read(b); err != nil {
		panic("gsui: crypto/rand failed: " + err.Error())
	}
	return "t-" + hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// JS Compilation
// ---------------------------------------------------------------------------

// ToJS compiles the node tree into a self-executing JavaScript function
// that builds and appends the entire tree to document.body.
func (n *Node) ToJS() string {
	var b strings.Builder
	b.WriteString("(function(){")
	counter := 0
	var postJS []string
	root := n.compile(&b, &counter, &postJS)
	fmt.Fprintf(&b, "document.body.appendChild(%s);", root)
	for _, js := range postJS {
		b.WriteString(js)
	}
	b.WriteString("})();")
	return b.String()
}

// ToJSReplace compiles JS that replaces an existing DOM element by its ID.
// The old element is found by ID, the new tree is built, and replaceWith() is called.
func (n *Node) ToJSReplace(targetID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	b.WriteString(fmt.Sprintf("var _t=document.getElementById('%s');", escJS(targetID)))
	b.WriteString(fmt.Sprintf("if(!_t){console.warn('[g-sui] replaceWith: element #%s not found');__ws.notfound('%s');return;}", escJS(targetID), escJS(targetID)))
	counter := 0
	var postJS []string
	root := n.compile(&b, &counter, &postJS)
	fmt.Fprintf(&b, "_t.replaceWith(%s);", root)
	for _, js := range postJS {
		b.WriteString(js)
	}
	b.WriteString("})();")
	return b.String()
}

// ToJSAppend compiles JS that appends this node as a child of the target element.
func (n *Node) ToJSAppend(parentID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	b.WriteString(fmt.Sprintf("var _p=document.getElementById('%s');", escJS(parentID)))
	b.WriteString(fmt.Sprintf("if(!_p){console.warn('[g-sui] appendChild: element #%s not found');__ws.notfound('%s');return;}", escJS(parentID), escJS(parentID)))
	counter := 0
	var postJS []string
	root := n.compile(&b, &counter, &postJS)
	fmt.Fprintf(&b, "_p.appendChild(%s);", root)
	for _, js := range postJS {
		b.WriteString(js)
	}
	b.WriteString("})();")
	return b.String()
}

// ToJSPrepend compiles JS that prepends this node as the first child.
func (n *Node) ToJSPrepend(parentID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	b.WriteString(fmt.Sprintf("var _p=document.getElementById('%s');", escJS(parentID)))
	b.WriteString(fmt.Sprintf("if(!_p){console.warn('[g-sui] prepend: element #%s not found');__ws.notfound('%s');return;}", escJS(parentID), escJS(parentID)))
	counter := 0
	var postJS []string
	root := n.compile(&b, &counter, &postJS)
	fmt.Fprintf(&b, "_p.prepend(%s);", root)
	for _, js := range postJS {
		b.WriteString(js)
	}
	b.WriteString("})();")
	return b.String()
}

// ToJSInner compiles JS that replaces the innerHTML of a target element
// with this node (sets target's children to just this node).
func (n *Node) ToJSInner(targetID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	b.WriteString(fmt.Sprintf("var _t=document.getElementById('%s');", escJS(targetID)))
	b.WriteString(fmt.Sprintf("if(!_t){console.warn('[g-sui] innerHTML: element #%s not found');__ws.notfound('%s');return;}", escJS(targetID), escJS(targetID)))
	b.WriteString("_t.innerHTML='';")
	counter := 0
	var postJS []string
	root := n.compile(&b, &counter, &postJS)
	fmt.Fprintf(&b, "_t.appendChild(%s);", root)
	for _, js := range postJS {
		b.WriteString(js)
	}
	b.WriteString("})();")
	return b.String()
}

// svgTags lists element names that belong to the SVG namespace.
// When any of these is encountered (or is a descendant of one),
// compile emits createElementNS instead of createElement.
var svgTags = map[string]bool{
	"svg": true, "g": true, "path": true, "circle": true, "ellipse": true,
	"line": true, "polyline": true, "polygon": true, "rect": true, "text": true,
	"tspan": true, "defs": true, "symbol": true, "use": true, "image": true,
	"clipPath": true, "mask": true, "pattern": true, "linearGradient": true,
	"radialGradient": true, "stop": true, "filter": true, "feBlend": true,
	"feColorMatrix": true, "feComponentTransfer": true, "feComposite": true,
	"feConvolveMatrix": true, "feDiffuseLighting": true, "feDisplacementMap": true,
	"feFlood": true, "feGaussianBlur": true, "feImage": true, "feMerge": true,
	"feMergeNode": true, "feMorphology": true, "feOffset": true,
	"feSpecularLighting": true, "feTile": true, "feTurbulence": true,
	"marker": true, "title": true, "desc": true, "metadata": true,
	"foreignObject": true, "switch": true, "a": true, "animate": true,
	"animateMotion": true, "animateTransform": true, "set": true,
	"textPath": true,
}

const svgNS = "http://www.w3.org/2000/svg"

// compile recursively emits JS statements to build a DOM element tree.
// Returns the variable name assigned to this node.
// Raw JS blocks (node.rawJS) are collected into postJS and deferred until
// after the root node is inserted into the DOM so that getElementById works.
// The inSVG flag propagates SVG namespace context to descendants.
func (n *Node) compile(b *strings.Builder, counter *int, postJS *[]string, inSVG ...bool) string {
	varName := fmt.Sprintf("e%d", *counter)
	*counter++

	parentIsSVG := len(inSVG) > 0 && inSVG[0]
	useSVGNS := parentIsSVG || n.tag == "svg"

	if useSVGNS {
		fmt.Fprintf(b, "var %s=document.createElementNS('%s','%s');", varName, svgNS, escJS(n.tag))
	} else {
		fmt.Fprintf(b, "var %s=document.createElement('%s');", varName, escJS(n.tag))
	}

	if n.id != "" {
		fmt.Fprintf(b, "%s.id='%s';", varName, escJS(n.id))
	}
	if n.class != "" {
		if useSVGNS {
			// SVG elements have className as SVGAnimatedString; use setAttribute.
			fmt.Fprintf(b, "%s.setAttribute('class','%s');", varName, escJS(n.class))
		} else {
			fmt.Fprintf(b, "%s.className='%s';", varName, escJS(n.class))
		}
	}
	if n.text != "" {
		fmt.Fprintf(b, "%s.textContent='%s';", varName, escJS(n.text))
	}

	// Attributes
	for k, v := range n.attrs {
		fmt.Fprintf(b, "%s.setAttribute('%s','%s');", varName, escJS(k), escJS(v))
	}

	// Inline styles
	for k, v := range n.styles {
		fmt.Fprintf(b, "%s.style['%s']='%s';", varName, escJS(k), escJS(v))
	}

	// Events
	for event, action := range n.events {
		if action.rawJS != "" {
			// Client-side only: raw JS, no WS call
			fmt.Fprintf(b,
				"%s.addEventListener('%s',function(event){%s});",
				varName, escJS(event), action.rawJS,
			)
		} else if len(action.Collect) > 0 {
			collectJSON, _ := json.Marshal(action.Collect)
			dataJSON, _ := json.Marshal(action.Data)
			fmt.Fprintf(b,
				"%s.addEventListener('%s',function(){__ws.call('%s',%s,%s)});",
				varName, escJS(event), escJS(action.Name), string(dataJSON), string(collectJSON),
			)
		} else {
			dataJSON, _ := json.Marshal(action.Data)
			fmt.Fprintf(b,
				"%s.addEventListener('%s',function(){__ws.call('%s',%s)});",
				varName, escJS(event), escJS(action.Name), string(dataJSON),
			)
		}
	}

	// Children
	for _, child := range n.children {
		childVar := child.compile(b, counter, postJS, useSVGNS)
		fmt.Fprintf(b, "%s.appendChild(%s);", varName, childVar)
	}

	// Collect raw JS for deferred execution (after DOM insertion).
	// The snippet is wrapped in .call(eN) so that `this` refers to
	// the DOM element — no manual ID bookkeeping needed.
	if n.rawJS != "" {
		*postJS = append(*postJS, fmt.Sprintf("(function(){%s}).call(%s);", n.rawJS, varName))
	}

	return varName
}

// ---------------------------------------------------------------------------
// JS Helper Functions (return JS strings for common DOM operations)
// ---------------------------------------------------------------------------

// Notify returns JS that shows a toast notification styled with a left
// accent border, colored dot, and auto-dismiss. Supports "success", "error",
// "error-reload" (persistent with Reload button), and "info" (default) variants.
func Notify(variant, message string) string {
	return fmt.Sprintf(
		`(function(){`+
			// Ensure __messages__ container exists
			`var box=document.getElementById('__messages__');`+
			`if(!box){box=document.createElement('div');box.id='__messages__';`+
			`box.style.cssText='position:fixed;top:0;right:0;padding:8px;z-index:9999;pointer-events:none';`+
			`document.body.appendChild(box);}`+
			// Create notification element
			`var n=document.createElement('div');`+
			`n.style.cssText='display:flex;align-items:center;gap:10px;padding:12px 16px;margin:8px;border-radius:12px;min-height:44px;min-width:340px;max-width:340px;box-shadow:0 6px 18px rgba(0,0,0,0.08);border:1px solid;font-weight:600;font-family:Inter,system-ui,sans-serif;font-size:14px;opacity:0;transform:translateX(20px);transition:opacity 200ms,transform 200ms;pointer-events:auto';`+
			// Variant-specific colors (dark-mode aware)
			`var v='%s',accent='#4f46e5',timeout=5000,dk=document.documentElement.classList.contains('dark');`+
			`if(v==='success'){accent='#16a34a';if(dk){n.style.background='#052e16';n.style.color='#86efac';n.style.borderColor='#14532d'}else{n.style.background='#dcfce7';n.style.color='#166534';n.style.borderColor='#bbf7d0'}}`+
			`else if(v==='error'||v==='error-reload'){accent='#dc2626';if(dk){n.style.background='#450a0a';n.style.color='#fca5a5';n.style.borderColor='#7f1d1d'}else{n.style.background='#fee2e2';n.style.color='#991b1b';n.style.borderColor='#fecaca'};if(v==='error-reload')timeout=88000}`+
			`else{if(dk){n.style.background='#1e1b4b';n.style.color='#a5b4fc';n.style.borderColor='#312e81'}else{n.style.background='#eef2ff';n.style.color='#3730a3';n.style.borderColor='#e0e7ff'}}`+
			`n.style.borderLeft='4px solid '+accent;`+
			// Dot indicator
			`var dot=document.createElement('span');dot.style.cssText='width:10px;height:10px;border-radius:9999px;flex-shrink:0;background:'+accent;`+
			// Message text
			`var t=document.createElement('span');t.style.flex='1';t.textContent='%s';`+
			`n.appendChild(dot);n.appendChild(t);`+
			// Reload button for error-reload
			`if(v==='error-reload'){var btn=document.createElement('button');btn.textContent='Reload';`+
			`btn.style.cssText='background:#991b1b;color:#fff;border:none;padding:6px 10px;border-radius:8px;cursor:pointer;font-weight:700;font-size:13px';`+
			`btn.onclick=function(){try{location.reload()}catch(_){}};n.appendChild(btn)}`+
			// Mount and animate in
			`box.appendChild(n);`+
			`requestAnimationFrame(function(){n.style.opacity='1';n.style.transform='translateX(0)'});`+
			// Auto-dismiss with fade out
			`setTimeout(function(){try{n.style.opacity='0';n.style.transform='translateX(20px)';`+
			`setTimeout(function(){try{if(n&&n.parentNode)n.parentNode.removeChild(n)}catch(_){}},200)}catch(_){}},timeout)`+
			`})();`,
		escJS(variant), escJS(message),
	)
}

// Redirect returns JS that navigates to a new URL (full page reload).
func Redirect(url string) string {
	return fmt.Sprintf("window.location.href='%s';", escJS(url))
}

// SetLocation returns JS that updates the browser URL without a page reload
// using history.pushState. Use this in WS actions to keep the address bar
// in sync with the visible content.
func SetLocation(url string) string {
	return fmt.Sprintf("history.pushState(null,'','%s');", escJS(url))
}

// Back returns JS that navigates back in browser history (history.back()).
func Back() *Action {
	// return "history.back();"
	return JS("history.back()")
}

// SetTitle returns JS that updates the document title.
func SetTitle(title string) string {
	return fmt.Sprintf("document.title='%s';", escJS(title))
}

// RemoveEl returns JS that removes an element by ID.
func RemoveEl(id string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] remove: element #%s not found');__ws.notfound('%s');return;}e.remove()})();", escJS(id), escJS(id), escJS(id))
}

// SetText returns JS that sets the textContent of an element by ID.
func SetText(id, text string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] setText: element #%s not found');__ws.notfound('%s');return;}e.textContent='%s'})();", escJS(id), escJS(id), escJS(id), escJS(text))
}

// SetAttr returns JS that sets an attribute on an element by ID.
func SetAttr(id, attr, value string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] setAttr: element #%s not found');__ws.notfound('%s');return;}e.setAttribute('%s','%s')})();", escJS(id), escJS(id), escJS(id), escJS(attr), escJS(value))
}

// AddClass returns JS that adds a CSS class to an element.
func AddClass(id, cls string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] addClass: element #%s not found');__ws.notfound('%s');return;}e.classList.add('%s')})();", escJS(id), escJS(id), escJS(id), escJS(cls))
}

// RemoveClass returns JS that removes a CSS class from an element.
func RemoveClass(id, cls string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] removeClass: element #%s not found');__ws.notfound('%s');return;}e.classList.remove('%s')})();", escJS(id), escJS(id), escJS(id), escJS(cls))
}

// Show returns JS that removes the 'hidden' class (shows the element).
func Show(id string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] show: element #%s not found');__ws.notfound('%s');return;}e.classList.remove('hidden')})();", escJS(id), escJS(id), escJS(id))
}

// Hide returns JS that adds the 'hidden' class (hides the element).
func Hide(id string) string {
	return fmt.Sprintf("(function(){var e=document.getElementById('%s');if(!e){console.warn('[g-sui] hide: element #%s not found');__ws.notfound('%s');return;}e.classList.add('hidden')})();", escJS(id), escJS(id), escJS(id))
}

// Download returns JS that triggers a file download.
func Download(filename, mimeType, base64Data string) string {
	return fmt.Sprintf(
		"(function(){var a=document.createElement('a');a.href='data:%s;base64,%s';a.download='%s';document.body.appendChild(a);a.click();a.remove()})();",
		escJS(mimeType), base64Data, escJS(filename),
	)
}

// DragToScroll returns JS that enables mouse-drag horizontal scrolling
// on the element with the given ID. Interactive children (input, select,
// button, a, .dt-filter-dropdown) are excluded from triggering the drag.
func DragToScroll(id string) string {
	return fmt.Sprintf(
		"(function(){"+
			"var el=document.getElementById('%s');if(!el){console.warn('[g-sui] dragToScroll: element #%s not found');__ws.notfound('%s');return;}"+
			"var down=false,sx=0,sl=0;"+
			"el.addEventListener('mousedown',function(e){"+
			"if(e.target.closest('input,select,button,a,.dt-filter-dropdown'))return;"+
			"down=true;sx=e.pageX-el.offsetLeft;sl=el.scrollLeft;"+
			"el.style.cursor='grabbing';el.style.userSelect='none';});"+
			"document.addEventListener('mouseup',function(){"+
			"if(!down)return;down=false;el.style.cursor='grab';el.style.removeProperty('user-select');});"+
			"document.addEventListener('mousemove',function(e){"+
			"if(!down)return;e.preventDefault();"+
			"el.scrollLeft=sl-(e.pageX-el.offsetLeft-sx);});"+
			"})();",
		escJS(id), escJS(id), escJS(id),
	)
}

// ---------------------------------------------------------------------------
// Internal
// ---------------------------------------------------------------------------

// escJS escapes a string for safe embedding inside JS single-quoted strings.
func escJS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
