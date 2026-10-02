// Package ui provides a server-rendered UI framework where Go builds
// typed DOM node trees that compile to pure JavaScript strings.
// The runtime sends versioned messages with JavaScript DOM operations.
// Applications return typed Result effects without a client-side framework.
//
// SVG elements (svg, path, circle, rect, etc.) are automatically created
// with document.createElementNS using the SVG namespace. Child elements
// of an SVG root inherit the namespace. Classes on SVG elements use
// setAttribute('class', ...) instead of .className for compatibility.
package ui

import (
	"crypto/rand"
	"encoding/hex"
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
func IMonth(class ...string) *Node    { return Input(class...).Attr("type", "month") }
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

// OnClick runs action on click.
func (n *Node) OnClick(action *Action) *Node { return n.On("click", action) }

// OnSubmit runs action on submit. Native form submission is prevented.
func (n *Node) OnSubmit(action *Action) *Node { return n.On("submit", action) }

// On runs action on a DOM event. It replaces an earlier action for the event.
func (n *Node) On(event string, action *Action) *Node {
	if action == nil {
		return n
	}
	if n.events == nil {
		n.events = make(map[string]*Action)
	}
	n.events[event] = action
	return n
}

// UnsafeJS runs raw JavaScript after this node is inserted; `this` is the
// element. Prefer Node behaviors (OnKey, Shortcut, OnOutsideClick, ...) and
// Widget; use this only for browser APIs they do not cover.
//
// This is a trusted raw API: never pass untrusted/user-controlled input to it.
func (n *Node) UnsafeJS(raw string) *Node { n.rawJS += raw; return n }

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

// toJS compiles the node tree into a self-executing JavaScript function
// that builds and appends the entire tree to document.body.
func (n *Node) toJS() string {
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

// toJSReplace compiles JS that replaces an existing DOM element by its ID.
// The old element is found by ID, the new tree is built, and replaceWith() is called.
func (n *Node) toJSReplace(targetID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	fmt.Fprintf(&b, "var _t=document.getElementById('%s');", escJS(targetID))
	fmt.Fprintf(&b, "if(!_t){console.warn('[g-sui] replaceWith: element #%s not found');__ws.notfound('%s');return;}", escJS(targetID), escJS(targetID))
	counter := 0
	var postJS []string
	root := n.compile(&b, &counter, &postJS)
	fmt.Fprintf(&b, "if(window.__gsuiDispose)__gsuiDispose(_t);_t.replaceWith(%s);", root)
	for _, js := range postJS {
		b.WriteString(js)
	}
	b.WriteString("})();")
	return b.String()
}

// toJSAppend compiles JS that appends this node as a child of the target element.
func (n *Node) toJSAppend(parentID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	fmt.Fprintf(&b, "var _p=document.getElementById('%s');", escJS(parentID))
	fmt.Fprintf(&b, "if(!_p){console.warn('[g-sui] appendChild: element #%s not found');__ws.notfound('%s');return;}", escJS(parentID), escJS(parentID))
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

// toJSPrepend compiles JS that prepends this node as the first child.
func (n *Node) toJSPrepend(parentID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	fmt.Fprintf(&b, "var _p=document.getElementById('%s');", escJS(parentID))
	fmt.Fprintf(&b, "if(!_p){console.warn('[g-sui] prepend: element #%s not found');__ws.notfound('%s');return;}", escJS(parentID), escJS(parentID))
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

// toJSInner compiles JS that replaces the innerHTML of a target element
// with this node (sets target's children to just this node).
func (n *Node) toJSInner(targetID string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	fmt.Fprintf(&b, "var _t=document.getElementById('%s');", escJS(targetID))
	fmt.Fprintf(&b, "if(!_t){console.warn('[g-sui] innerHTML: element #%s not found');__ws.notfound('%s');return;}", escJS(targetID), escJS(targetID))
	b.WriteString("if(window.__gsuiDispose)Array.from(_t.children).forEach(__gsuiDispose);_t.innerHTML='';")
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

const svgNS = "http://www.w3.org/2000/svg"

// compile recursively emits JS statements to build a DOM element tree.
// Returns the variable name assigned to this node.
// Raw JS blocks (node.rawJS) are collected into postJS and deferred until
// after the root node is inserted into the DOM so that getElementById works.
// The inSVG flag propagates SVG namespace context to descendants.
func (n *Node) compile(b *strings.Builder, counter *int, postJS *[]string, inSVG ...bool) string {
	if *counter == 0 {
		b.WriteString("function _bind(el,event,fn){(el.__gsuiHandlers||(el.__gsuiHandlers={}))[event]=fn;el.addEventListener(event,fn)}")
	}
	varName := fmt.Sprintf("e%d", *counter)
	*counter++

	parentIsSVG := len(inSVG) > 0 && inSVG[0]
	// Only the <svg> root opens the SVG namespace; descendants inherit it via
	// parentIsSVG. A tag-name lookup would wrongly namespace HTML elements that
	// share a name with SVG (notably <a>, plus <title>, <text>, <image>,
	// <switch>) when they are used outside an <svg>.
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
		if action == nil {
			continue
		}
		prevent := ""
		if event == "submit" || (event == "click" && action.server()) {
			prevent = "event.preventDefault();"
		}
		fmt.Fprintf(b, "_bind(%s,'%s',function(event){var el=event.currentTarget;%s%s});", varName, escJS(event), prevent, action.code())
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
// Internal
// ---------------------------------------------------------------------------

// escJS escapes a string for safe embedding inside JS single-quoted strings
// that may themselves be embedded in an HTML <script> tag.
func escJS(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '<':
			b.WriteString(`\u003c`)
		case '>':
			b.WriteString(`\u003e`)
		case '&':
			b.WriteString(`\u0026`)
		case '=':
			b.WriteString(`\u003d`)
		case '\u2028':
			b.WriteString(`\u2028`)
		case '\u2029':
			b.WriteString(`\u2029`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}

	return b.String()
}
