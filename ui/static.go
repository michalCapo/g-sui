package ui

import (
	"html"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// What a full page load says before any script runs: its title, its
// description and, for public pages, its text. Search engines and link
// previews read this first.

// Description sets the page's <meta name="description">: in the HTML of a
// full load and after live navigation. App.Description stays the default for
// pages without one.
func (n *Node) Description(text string) *Node {
	n.Attr("data-gsui-description", text)
	return n.mount("description", text)
}

// StaticHTML makes a full load of this page also carry the page as plain
// HTML, so search engines, link previews and readers without JavaScript get
// its text. The client drops that copy and builds the live page as usual.
// Call it from the handlers of public pages: the page then goes out twice.
func (ctx *Context) StaticHTML() { ctx.staticHTML = true }

// firstAttr is the value of the first node in the tree, depth first, that
// has the attribute.
func firstAttr(n *Node, key string) string {
	if n == nil {
		return ""
	}
	if v, ok := n.attrs[key]; ok {
		return v
	}
	for _, child := range n.children {
		if v := firstAttr(child, key); v != "" {
			return v
		}
	}
	return ""
}

// staticRevealCSS shows the HTML copy to readers without JavaScript, as the
// page stays hidden until the client starts.
const staticRevealCSS = `<noscript><style>html.gsui-booting body{visibility:visible!important;opacity:1!important}</style></noscript>`

// staticID wraps the HTML copy so the client can drop it before it builds
// the page.
const staticID = "__gsui_static"

// staticHTML is the node as HTML: the same elements, text and attributes as
// the live page, without its behavior. Scripts, styles (the head has the
// page's styles; inside SVG a style's text would be read as markup), event
// attributes and names that are not plain element or attribute names are
// left out. Form fields are disabled: without the client a form would send
// its fields to the page's address.
func (n *Node) staticHTML() string {
	var b strings.Builder
	n.writeHTML(&b)
	return b.String()
}

func (n *Node) writeHTML(b *strings.Builder) {
	tag := strings.ToLower(n.tag)
	if tag == "script" || tag == "style" || !plainName(n.tag) {
		return
	}
	b.WriteByte('<')
	b.WriteString(n.tag)
	if n.id != "" {
		writeAttr(b, "id", n.id)
	}
	if n.class != "" {
		writeAttr(b, "class", n.class)
	}
	for _, k := range slices.Sorted(maps.Keys(n.attrs)) {
		if plainName(k) && !strings.HasPrefix(strings.ToLower(k), "on") {
			writeAttr(b, k, n.attrs[k])
		}
	}
	if len(n.styles) > 0 {
		var style strings.Builder
		for _, k := range slices.Sorted(maps.Keys(n.styles)) {
			style.WriteString(cssName(k))
			style.WriteByte(':')
			style.WriteString(n.styles[k])
			style.WriteByte(';')
		}
		writeAttr(b, "style", style.String())
	}
	if _, ok := n.attrs["disabled"]; !ok && formField[tag] {
		writeAttr(b, "disabled", "")
	}
	b.WriteByte('>')
	if n.void {
		return
	}
	b.WriteString(html.EscapeString(n.text))
	for _, child := range n.children {
		child.writeHTML(b)
	}
	b.WriteString("</")
	b.WriteString(n.tag)
	b.WriteByte('>')
}

// formField are the elements that send or start sending a form.
var formField = map[string]bool{"input": true, "textarea": true, "select": true, "button": true}

func writeAttr(b *strings.Builder, name, value string) {
	b.WriteByte(' ')
	b.WriteString(name)
	b.WriteString(`="`)
	b.WriteString(html.EscapeString(value))
	b.WriteByte('"')
}

// plainName reports whether s is safe as an element or attribute name.
func plainName(s string) bool {
	if s == "" || !isASCIILetter(rune(s[0])) {
		return false
	}
	for _, r := range s {
		if !isASCIILetter(r) && (r < '0' || r > '9') && !strings.ContainsRune("-_.:", r) {
			return false
		}
	}
	return true
}

func isASCIILetter(r rune) bool { return r < 128 && unicode.IsLetter(r) }

// cssName turns the JavaScript name of a style (backgroundColor) into its
// CSS name (background-color).
func cssName(k string) string {
	if strings.HasPrefix(k, "--") {
		return k
	}
	var b strings.Builder
	for _, r := range k {
		if unicode.IsUpper(r) {
			b.WriteByte('-')
			r = unicode.ToLower(r)
		}
		b.WriteRune(r)
	}
	return b.String()
}
