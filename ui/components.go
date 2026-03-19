package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
)

// ---------------------------------------------------------------------------
// 1. Skeleton Loaders
// ---------------------------------------------------------------------------

// SkeletonTable returns an animated pulse skeleton representing a 4-column
// table with a header row and 5 data rows.
func SkeletonTable() *Node {
	bar := func(w string) *Node {
		return Div("h-4 bg-gray-200 dark:bg-gray-700 rounded " + w)
	}

	headerCells := make([]*Node, 4)
	widths := []string{"w-20", "w-28", "w-24", "w-16"}
	for i := 0; i < 4; i++ {
		headerCells[i] = Th("px-4 py-3").Render(bar(widths[i]))
	}
	header := Thead().Render(Tr("border-b border-gray-200 dark:border-gray-600").Render(headerCells...))

	rowWidths := [][]string{
		{"w-24", "w-32", "w-20", "w-16"},
		{"w-20", "w-28", "w-24", "w-12"},
		{"w-28", "w-20", "w-16", "w-24"},
		{"w-16", "w-24", "w-28", "w-20"},
		{"w-24", "w-16", "w-32", "w-12"},
	}
	rows := make([]*Node, 5)
	for i := 0; i < 5; i++ {
		cells := make([]*Node, 4)
		for j := 0; j < 4; j++ {
			cells[j] = Td("px-4 py-3").Render(bar(rowWidths[i][j]))
		}
		rows[i] = Tr("border-b border-gray-100 dark:border-gray-700").Render(cells...)
	}
	body := Tbody().Render(rows...)

	return Div("animate-pulse").Render(
		Table("w-full").Render(header, body),
	)
}

// SkeletonCards returns an animated pulse skeleton of a 6-card responsive grid.
func SkeletonCards() *Node {
	cards := make([]*Node, 6)
	barWidths := []string{"w-3/4", "w-1/2", "w-2/3", "w-5/6", "w-3/5", "w-4/5"}
	for i := 0; i < 6; i++ {
		cards[i] = Div("bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 flex flex-col gap-3").Render(
			Div("h-32 bg-gray-200 dark:bg-gray-700 rounded"),
			Div("h-4 bg-gray-200 dark:bg-gray-700 rounded "+barWidths[i]),
			Div("h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2"),
			Div("h-3 bg-gray-200 dark:bg-gray-700 rounded w-2/3"),
		)
	}
	return Div("animate-pulse grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4").Render(cards...)
}

// SkeletonList returns an animated pulse skeleton of 5 list rows, each with
// a circle avatar placeholder and two text line placeholders.
func SkeletonList() *Node {
	textWidths := [][]string{
		{"w-3/4", "w-1/2"},
		{"w-2/3", "w-2/5"},
		{"w-4/5", "w-1/3"},
		{"w-1/2", "w-3/5"},
		{"w-3/5", "w-2/5"},
	}
	rows := make([]*Node, 5)
	for i := 0; i < 5; i++ {
		rows[i] = Div("flex items-center gap-4 py-3").Render(
			Div("w-10 h-10 bg-gray-200 dark:bg-gray-700 rounded-full flex-shrink-0"),
			Div("flex-1 flex flex-col gap-2").Render(
				Div("h-4 bg-gray-200 dark:bg-gray-700 rounded "+textWidths[i][0]),
				Div("h-3 bg-gray-200 dark:bg-gray-700 rounded "+textWidths[i][1]),
			),
		)
	}
	return Div("animate-pulse flex flex-col divide-y divide-gray-200 dark:divide-gray-700").Render(rows...)
}

// SkeletonComponent returns an animated pulse skeleton of a single card with
// a title, 3 text lines, and a button placeholder.
func SkeletonComponent() *Node {
	return Div("animate-pulse bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 flex flex-col gap-4").Render(
		Div("h-5 bg-gray-200 dark:bg-gray-700 rounded w-1/3"),
		Div("h-3 bg-gray-200 dark:bg-gray-700 rounded w-full"),
		Div("h-3 bg-gray-200 dark:bg-gray-700 rounded w-5/6"),
		Div("h-3 bg-gray-200 dark:bg-gray-700 rounded w-2/3"),
		Div("h-9 bg-gray-200 dark:bg-gray-700 rounded w-28 mt-2"),
	)
}

// SkeletonPage returns an animated pulse skeleton of a full page layout with
// a header bar, sidebar column, and main content area.
func SkeletonPage() *Node {
	sidebarItems := make([]*Node, 5)
	sw := []string{"w-3/4", "w-1/2", "w-2/3", "w-4/5", "w-3/5"}
	for i := 0; i < 5; i++ {
		sidebarItems[i] = Div("h-4 bg-gray-200 dark:bg-gray-700 rounded " + sw[i])
	}

	return Div("animate-pulse flex flex-col gap-0 min-h-[400px]").Render(
		Div("h-14 bg-gray-200 dark:bg-gray-700 rounded-t-lg"),
		Div("flex flex-1 gap-0").Render(
			Div("w-48 border-r border-gray-200 dark:border-gray-700 p-4 flex flex-col gap-4").Render(sidebarItems...),
			Div("flex-1 p-6 flex flex-col gap-4").Render(
				Div("h-6 bg-gray-200 dark:bg-gray-700 rounded w-1/3"),
				Div("h-4 bg-gray-200 dark:bg-gray-700 rounded w-full"),
				Div("h-4 bg-gray-200 dark:bg-gray-700 rounded w-5/6"),
				Div("h-4 bg-gray-200 dark:bg-gray-700 rounded w-2/3"),
				Div("h-32 bg-gray-200 dark:bg-gray-700 rounded w-full mt-4"),
			),
		),
	)
}

// SkeletonForm returns an animated pulse skeleton of a form with 4 label+input
// pairs and a submit button placeholder.
func SkeletonForm() *Node {
	labelWidths := []string{"w-20", "w-24", "w-16", "w-28"}
	fields := make([]*Node, 4)
	for i := 0; i < 4; i++ {
		fields[i] = Div("flex flex-col gap-2").Render(
			Div("h-3 bg-gray-200 dark:bg-gray-700 rounded "+labelWidths[i]),
			Div("h-10 bg-gray-200 dark:bg-gray-700 rounded w-full"),
		)
	}
	return Div("animate-pulse flex flex-col gap-5").Render(
		fields[0], fields[1], fields[2], fields[3],
		Div("h-10 bg-gray-200 dark:bg-gray-700 rounded w-32 mt-2"),
	)
}

// ---------------------------------------------------------------------------
// 2. Markdown Rendering
// ---------------------------------------------------------------------------

// Markdown converts a markdown string to HTML and renders it inside a Div.
// Since the framework produces only document.createElement JS (no innerHTML
// setter on nodes), this uses .JS() to set innerHTML after the node mounts.
// The class parameter is applied to the container div.
func Markdown(class, content string) *Node {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return Div(class).Text(content)
	}

	id := Target()
	html := buf.String()

	return Div(class).ID(id).JS(
		fmt.Sprintf("document.getElementById('%s').innerHTML='%s';", escJS(id), escJS(html)),
	)
}

// ---------------------------------------------------------------------------
// 3. Captcha V3 (client-side only)
// ---------------------------------------------------------------------------

// CaptchaV3Builder configures a reCAPTCHA v3 widget that loads the Google
// script, executes the challenge, and stores the token in a hidden input.
type CaptchaV3Builder struct {
	siteKey    string
	formAction string
	tokenField string
}

// NewCaptchaV3 creates a new reCAPTCHA v3 builder with the given site key.
func NewCaptchaV3(siteKey string) *CaptchaV3Builder {
	return &CaptchaV3Builder{
		siteKey:    siteKey,
		tokenField: "g-recaptcha-response",
	}
}

// FormAction sets the action name sent to reCAPTCHA for scoring context.
func (c *CaptchaV3Builder) FormAction(name string) *CaptchaV3Builder {
	c.formAction = name
	return c
}

// TokenField sets the hidden input name where the token will be stored.
// Defaults to "g-recaptcha-response".
func (c *CaptchaV3Builder) TokenField(name string) *CaptchaV3Builder {
	c.tokenField = name
	return c
}

// Build produces the Node tree: a container with a hidden input for the
// token and JS that loads the reCAPTCHA v3 script, executes the challenge,
// and auto-refreshes the token every 110 seconds.
func (c *CaptchaV3Builder) Build() *Node {
	containerID := Target()
	inputID := Target()

	action := c.formAction
	if action == "" {
		action = "submit"
	}

	js := fmt.Sprintf(
		`(function(){`+
			`var s=document.createElement('script');`+
			`s.src='https://www.google.com/recaptcha/api.js?render=%s';`+
			`s.onload=function(){`+
			`grecaptcha.ready(function(){`+
			`function exec(){grecaptcha.execute('%s',{action:'%s'}).then(function(token){`+
			`var el=document.getElementById('%s');if(el)el.value=token;`+
			`})}`+
			`exec();`+
			`setInterval(exec,110000);`+
			`})`+
			`};`+
			`document.head.appendChild(s);`+
			`})();`,
		escJS(c.siteKey),
		escJS(c.siteKey),
		escJS(action),
		escJS(inputID),
	)

	return Div().ID(containerID).Render(
		IHidden().ID(inputID).Attr("name", c.tokenField),
	).JS(js)
}

// ---------------------------------------------------------------------------
// 4. Icon Helpers
// ---------------------------------------------------------------------------

// Icon renders a Material Symbols icon using Material Icons Round font.
// The name parameter is the icon ligature (e.g. "home", "settings").
// Optional class strings are appended to the base icon class.
func Icon(name string, class ...string) *Node {
	cls := "material-icons-round"
	if len(class) > 0 && class[0] != "" {
		cls += " " + class[0]
	}
	return I(cls).Text(name)
}

// IconText renders a flex row containing a Material icon and text label.
// Optional class strings are applied to the outer wrapper.
func IconText(icon, text string, class ...string) *Node {
	cls := "inline-flex items-center gap-2"
	if len(class) > 0 && class[0] != "" {
		cls += " " + class[0]
	}
	return Span(cls).Render(
		Icon(icon),
		Span().Text(text),
	)
}

// ---------------------------------------------------------------------------
// 5. Accordion Builder
// ---------------------------------------------------------------------------

type accordionItem struct {
	title   string
	content *Node
	open    bool
}

// AccordionBuilder builds an accordion component using native <details>/<summary>.
type AccordionBuilder struct {
	items    []accordionItem
	multiple bool
	variant  string
	class    string
}

// NewAccordion creates a new AccordionBuilder with default "bordered" variant.
func NewAccordion() *AccordionBuilder {
	return &AccordionBuilder{variant: "bordered"}
}

// Item adds an accordion item. The optional open parameter controls whether
// the item starts expanded (defaults to false).
func (a *AccordionBuilder) Item(title string, content *Node, open ...bool) *AccordionBuilder {
	isOpen := false
	if len(open) > 0 {
		isOpen = open[0]
	}
	a.items = append(a.items, accordionItem{title: title, content: content, open: isOpen})
	return a
}

// Multiple controls whether multiple items can be open simultaneously.
// When false (default), opening one item closes others.
func (a *AccordionBuilder) Multiple(m bool) *AccordionBuilder {
	a.multiple = m
	return a
}

// Variant sets the visual style: "bordered" (default), "ghost", or "separated".
func (a *AccordionBuilder) Variant(v string) *AccordionBuilder {
	a.variant = v
	return a
}

// AccordionClass sets additional CSS classes on the accordion wrapper.
func (a *AccordionBuilder) AccordionClass(cls string) *AccordionBuilder {
	a.class = cls
	return a
}

// Build produces the complete accordion Node tree.
func (a *AccordionBuilder) Build() *Node {
	groupID := Target()

	var wrapperClass, itemClass, summaryClass, contentClass string
	switch a.variant {
	case "ghost":
		wrapperClass = "flex flex-col"
		itemClass = ""
		summaryClass = "px-4 py-3 font-medium text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 list-none flex items-center justify-between"
		contentClass = "px-4 py-3 text-sm text-gray-600 dark:text-gray-400"
	case "separated":
		wrapperClass = "flex flex-col gap-2"
		itemClass = "bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden"
		summaryClass = "px-4 py-3 font-medium text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 list-none flex items-center justify-between"
		contentClass = "px-4 py-3 text-sm text-gray-600 dark:text-gray-400 border-t border-gray-100 dark:border-gray-700"
	default: // "bordered"
		wrapperClass = "border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden"
		itemClass = "border-b border-gray-200 dark:border-gray-700 last:border-b-0"
		summaryClass = "px-4 py-3 font-medium text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 list-none flex items-center justify-between"
		contentClass = "px-4 py-3 text-sm text-gray-600 dark:text-gray-400 border-t border-gray-100 dark:border-gray-700"
	}

	if a.class != "" {
		wrapperClass += " " + a.class
	}

	styleNode := El("style").Text(
		fmt.Sprintf(`#%s details[open] .chevron{transform:rotate(180deg)}`, groupID),
	)

	items := make([]*Node, len(a.items))
	for i, item := range a.items {
		details := Details(itemClass)
		if item.open {
			details.Attr("open", "true")
		}

		chevron := Icon("expand_more", "text-sm text-gray-400 dark:text-gray-500 transition-transform chevron")

		summary := Summary(summaryClass).Render(
			Span("dark:text-gray-200").Text(item.title),
			chevron,
		)

		details.Render(summary, Div(contentClass).Render(item.content))
		items[i] = details
	}

	wrapper := Div(wrapperClass).ID(groupID).Render(styleNode)
	wrapper.Render(items...)

	if !a.multiple {
		js := fmt.Sprintf(
			`(function(){`+
				`var g=document.getElementById('%s');if(!g)return;`+
				`var ds=g.querySelectorAll('details');`+
				`ds.forEach(function(d){`+
				`d.addEventListener('toggle',function(){`+
				`if(d.open){ds.forEach(function(o){if(o!==d)o.removeAttribute('open')})}`+
				`})});`+
				`})();`,
			escJS(groupID),
		)
		wrapper.JS(js)
	}

	return wrapper
}

// ---------------------------------------------------------------------------
// 6. Alert Builder
// ---------------------------------------------------------------------------

// AlertBuilder configures a styled alert/notification banner.
type AlertBuilder struct {
	message     string
	title       string
	variant     string
	dismissible bool
	persistKey  string
	class       string
}

// NewAlert creates a new AlertBuilder with "info" variant.
func NewAlert() *AlertBuilder {
	return &AlertBuilder{variant: "info"}
}

// Message sets the alert body text.
func (a *AlertBuilder) Message(m string) *AlertBuilder { a.message = m; return a }

// Title sets the alert heading.
func (a *AlertBuilder) Title(t string) *AlertBuilder { a.title = t; return a }

// Variant sets the visual variant: "info", "success", "warning", "error",
// or any of those with a "-outline" suffix for bordered style.
func (a *AlertBuilder) Variant(v string) *AlertBuilder { a.variant = v; return a }

// Dismissible enables a close button on the alert.
func (a *AlertBuilder) Dismissible(d bool) *AlertBuilder { a.dismissible = d; return a }

// Persist sets a localStorage key. When dismissed, the alert stays hidden
// on subsequent page loads until the key is cleared.
func (a *AlertBuilder) Persist(key string) *AlertBuilder { a.persistKey = key; return a }

// AlertClass sets additional CSS classes on the alert container.
func (a *AlertBuilder) AlertClass(cls string) *AlertBuilder { a.class = cls; return a }

// Build produces the alert Node tree.
func (a *AlertBuilder) Build() *Node {
	id := Target()

	base := a.variant
	outline := false
	if strings.HasSuffix(base, "-outline") {
		outline = true
		base = strings.TrimSuffix(base, "-outline")
	}

	var bgClass, borderClass, textClass, iconName string
	switch base {
	case "success":
		bgClass = "bg-green-50 dark:bg-green-900/20"
		borderClass = "border-green-200 dark:border-green-800"
		textClass = "text-green-800 dark:text-green-300"
		iconName = "check_circle"
	case "warning":
		bgClass = "bg-amber-50 dark:bg-amber-900/20"
		borderClass = "border-amber-200 dark:border-amber-800"
		textClass = "text-amber-800 dark:text-amber-300"
		iconName = "warning"
	case "error":
		bgClass = "bg-red-50 dark:bg-red-900/20"
		borderClass = "border-red-200 dark:border-red-800"
		textClass = "text-red-800 dark:text-red-300"
		iconName = "error"
	default:
		bgClass = "bg-blue-50 dark:bg-blue-900/20"
		borderClass = "border-blue-200 dark:border-blue-800"
		textClass = "text-blue-800 dark:text-blue-300"
		iconName = "info"
	}

	if outline {
		bgClass = "bg-white dark:bg-gray-900"
	}

	cls := fmt.Sprintf("border rounded-lg p-4 flex items-start gap-3 %s %s %s", bgClass, borderClass, textClass)
	if a.class != "" {
		cls += " " + a.class
	}

	container := Div(cls).ID(id)
	container.Render(Icon(iconName, "text-lg flex-shrink-0 mt-0.5"))

	textWrapper := Div("flex-1 min-w-0")
	if a.title != "" {
		textWrapper.Render(Div("font-semibold text-sm").Text(a.title))
	}
	if a.message != "" {
		textWrapper.Render(Div("text-sm mt-0.5").Text(a.message))
	}
	container.Render(textWrapper)

	if a.dismissible {
		dismissJS := fmt.Sprintf("document.getElementById('%s').remove()", escJS(id))
		if a.persistKey != "" {
			dismissJS = fmt.Sprintf(
				"localStorage.setItem('%s','1');document.getElementById('%s').remove()",
				escJS(a.persistKey), escJS(id),
			)
		}
		container.Render(
			Button("text-lg leading-none opacity-50 hover:opacity-100 cursor-pointer flex-shrink-0 -mt-1").
				Text("\u00d7").
				OnClick(JS(dismissJS)),
		)
	}

	if a.persistKey != "" {
		container.JS(
			fmt.Sprintf(
				"if(localStorage.getItem('%s')==='1'){document.getElementById('%s').style.display='none';}",
				escJS(a.persistKey), escJS(id),
			),
		)
	}

	return container
}

// ---------------------------------------------------------------------------
// 7. Badge Builder
// ---------------------------------------------------------------------------

// BadgeBuilder configures an inline badge / pill label.
type BadgeBuilder struct {
	text   string
	color  string
	size   string
	dot    bool
	icon   string
	square bool
	class  string
}

// NewBadge creates a new BadgeBuilder with the given text.
func NewBadge(text string) *BadgeBuilder {
	return &BadgeBuilder{text: text, color: "gray", size: "md"}
}

// Color sets the color scheme. Supports: "gray", "red", "green", "blue",
// "yellow", "purple", each optionally with "-outline" or "-soft" suffix.
func (b *BadgeBuilder) Color(c string) *BadgeBuilder { b.color = c; return b }

// BadgeSize sets the badge size: "sm", "md" (default), or "lg".
func (b *BadgeBuilder) BadgeSize(s string) *BadgeBuilder { b.size = s; return b }

// Dot renders a small dot indicator instead of text content.
func (b *BadgeBuilder) Dot() *BadgeBuilder { b.dot = true; return b }

// BadgeIcon sets a Material icon name to prepend inside the badge.
func (b *BadgeBuilder) BadgeIcon(name string) *BadgeBuilder { b.icon = name; return b }

// Square uses rounded-md corners instead of fully rounded (pill shape).
func (b *BadgeBuilder) Square() *BadgeBuilder { b.square = true; return b }

// BadgeClass sets additional CSS classes on the badge.
func (b *BadgeBuilder) BadgeClass(cls string) *BadgeBuilder { b.class = cls; return b }

// Build produces the badge Node.
func (b *BadgeBuilder) Build() *Node {
	baseColor := b.color
	variant := ""
	if strings.HasSuffix(baseColor, "-outline") {
		variant = "outline"
		baseColor = strings.TrimSuffix(baseColor, "-outline")
	} else if strings.HasSuffix(baseColor, "-soft") {
		variant = "soft"
		baseColor = strings.TrimSuffix(baseColor, "-soft")
	}

	type colorSet struct{ solid, outline, soft string }
	colors := map[string]colorSet{
		"gray":   {"bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300", "bg-transparent border border-gray-400 text-gray-700 dark:border-gray-500 dark:text-gray-400", "bg-gray-50 text-gray-700 dark:bg-gray-800 dark:text-gray-400"},
		"red":    {"bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400", "bg-transparent border border-red-400 text-red-700 dark:border-red-500 dark:text-red-400", "bg-red-50 text-red-700 dark:bg-red-900/20 dark:text-red-400"},
		"green":  {"bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400", "bg-transparent border border-green-400 text-green-700 dark:border-green-500 dark:text-green-400", "bg-green-50 text-green-800 dark:bg-green-900/20 dark:text-green-400"},
		"blue":   {"bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400", "bg-transparent border border-blue-400 text-blue-700 dark:border-blue-500 dark:text-blue-400", "bg-blue-50 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400"},
		"yellow": {"bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400", "bg-transparent border border-yellow-500 text-yellow-800 dark:border-yellow-500 dark:text-yellow-400", "bg-yellow-50 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400"},
		"purple": {"bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400", "bg-transparent border border-purple-400 text-purple-700 dark:border-purple-500 dark:text-purple-400", "bg-purple-50 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400"},
	}

	cs, ok := colors[baseColor]
	if !ok {
		cs = colors["gray"]
	}
	var colorCls string
	switch variant {
	case "outline":
		colorCls = cs.outline
	case "soft":
		colorCls = cs.soft
	default:
		colorCls = cs.solid
	}

	if b.dot {
		dotColors := map[string]string{
			"gray": "bg-gray-500", "red": "bg-red-500", "green": "bg-green-500",
			"blue": "bg-blue-500", "yellow": "bg-yellow-500", "purple": "bg-purple-500",
		}
		dc := dotColors[baseColor]
		if dc == "" {
			dc = "bg-gray-500"
		}
		var dotSize string
		switch b.size {
		case "sm":
			dotSize = "w-2 h-2"
		case "lg":
			dotSize = "w-3.5 h-3.5"
		default:
			dotSize = "w-2.5 h-2.5"
		}
		cls := fmt.Sprintf("inline-block rounded-full %s %s", dotSize, dc)
		if b.class != "" {
			cls += " " + b.class
		}
		return Span(cls)
	}

	var sizeClass, iconSize string
	switch b.size {
	case "sm":
		sizeClass = "px-1.5 py-0.5 text-[10px]"
		iconSize = "text-[10px]"
	case "lg":
		sizeClass = "px-3 py-1 text-sm"
		iconSize = "text-sm"
	default:
		sizeClass = "px-2 py-0.5 text-xs"
		iconSize = "text-xs"
	}

	shape := "rounded-full"
	if b.square {
		shape = "rounded-md"
	}

	cls := fmt.Sprintf("inline-flex items-center gap-1 font-medium %s %s %s", sizeClass, shape, colorCls)
	if b.class != "" {
		cls += " " + b.class
	}

	node := Span(cls)
	if b.icon != "" {
		node.Render(Icon(b.icon, iconSize))
	}
	node.Render(Span().Text(b.text))
	return node
}

// ---------------------------------------------------------------------------
// 8. Button (High-Level) Builder
// ---------------------------------------------------------------------------

// Button color presets.
const (
	BtnBlue         = "bg-blue-600 hover:bg-blue-700 text-white"
	BtnRed          = "bg-red-600 hover:bg-red-700 text-white"
	BtnGreen        = "bg-green-600 hover:bg-green-700 text-white"
	BtnYellow       = "bg-yellow-500 hover:bg-yellow-600 text-gray-900"
	BtnPurple       = "bg-purple-600 hover:bg-purple-700 text-white"
	BtnGray         = "bg-gray-600 hover:bg-gray-700 text-white"
	BtnWhite        = "bg-white hover:bg-gray-50 text-gray-700 border border-gray-300 dark:bg-gray-800 dark:hover:bg-gray-700 dark:text-gray-200 dark:border-gray-600"
	BtnBlueOutline  = "bg-transparent hover:bg-blue-50 text-blue-600 border border-blue-600 dark:hover:bg-blue-950 dark:text-blue-400 dark:border-blue-500"
	BtnRedOutline   = "bg-transparent hover:bg-red-50 text-red-600 border border-red-600 dark:hover:bg-red-950 dark:text-red-400 dark:border-red-500"
	BtnGreenOutline = "bg-transparent hover:bg-green-50 text-green-600 border border-green-600 dark:hover:bg-green-950 dark:text-green-400 dark:border-green-500"
)

// Button size presets.
const (
	BtnXS = "px-2 py-1 text-xs"
	BtnSM = "px-3 py-1.5 text-sm"
	BtnMD = "px-4 py-2 text-sm"
	BtnLG = "px-5 py-2.5 text-base"
	BtnXL = "px-6 py-3 text-lg"
)

// ButtonBuilder configures a high-level button component.
type ButtonBuilder struct {
	label    string
	color    string
	size     string
	icon     string
	href     string
	disabled bool
	submit   bool
	reset    bool
	formID   string
	action   *Action
	class    string
}

// NewButton creates a new ButtonBuilder with the given label.
// Defaults to BtnBlue color and BtnMD size.
func NewButton(label string) *ButtonBuilder {
	return &ButtonBuilder{label: label, color: BtnBlue, size: BtnMD}
}

// BtnColor sets the button color preset (use Btn* constants or custom classes).
func (b *ButtonBuilder) BtnColor(c string) *ButtonBuilder { b.color = c; return b }

// BtnSize sets the button size preset (use Btn* size constants or custom classes).
func (b *ButtonBuilder) BtnSize(s string) *ButtonBuilder { b.size = s; return b }

// BtnIcon sets a Material icon name to display before the label.
func (b *ButtonBuilder) BtnIcon(name string) *ButtonBuilder { b.icon = name; return b }

// Href converts the button to an <a> tag linking to the given URL.
func (b *ButtonBuilder) Href(url string) *ButtonBuilder { b.href = url; return b }

// Disabled marks the button as visually and functionally disabled.
func (b *ButtonBuilder) Disabled(d bool) *ButtonBuilder { b.disabled = d; return b }

// Submit makes the button a submit button. Optionally pass a form ID.
func (b *ButtonBuilder) Submit(formID ...string) *ButtonBuilder {
	b.submit = true
	if len(formID) > 0 {
		b.formID = formID[0]
	}
	return b
}

// Reset makes the button a reset button.
func (b *ButtonBuilder) Reset() *ButtonBuilder { b.reset = true; return b }

// OnBtnClick sets the action to execute when the button is clicked.
func (b *ButtonBuilder) OnBtnClick(a *Action) *ButtonBuilder { b.action = a; return b }

// BtnClass sets additional CSS classes on the button element.
func (b *ButtonBuilder) BtnClass(cls string) *ButtonBuilder { b.class = cls; return b }

// Build produces the button Node.
func (b *ButtonBuilder) Build() *Node {
	baseCls := "inline-flex items-center justify-center gap-2 rounded-lg font-medium cursor-pointer transition-colors"
	cls := fmt.Sprintf("%s %s %s", baseCls, b.size, b.color)

	if b.disabled {
		cls += " opacity-50 cursor-not-allowed pointer-events-none"
	}
	if b.class != "" {
		cls += " " + b.class
	}

	var node *Node
	if b.href != "" {
		node = A(cls).Attr("href", b.href)
	} else {
		node = Button(cls)
		if b.submit {
			node.Attr("type", "submit")
		} else if b.reset {
			node.Attr("type", "reset")
		} else {
			node.Attr("type", "button")
		}
		if b.formID != "" {
			node.Attr("form", b.formID)
		}
		if b.disabled {
			node.Attr("disabled", "true")
		}
	}

	if b.icon != "" {
		node.Render(Icon(b.icon, "text-[1.1em]"))
	}
	node.Render(Span().Text(b.label))

	if b.action != nil && !b.disabled {
		node.OnClick(b.action)
	}

	return node
}

// ---------------------------------------------------------------------------
// 9. Card Builder
// ---------------------------------------------------------------------------

// CardBuilder constructs a card component with optional header, body, footer,
// image, and variant styling.
type CardBuilder struct {
	header        *Node
	body          *Node
	footer        *Node
	image         string
	imageAlt      string
	imageWidth    string
	imageHeight   string
	imagePriority bool
	variant       string
	hover         bool
	compact       bool
	padding       string
	class         string
}

// NewCard creates a new CardBuilder with default settings.
func NewCard() *CardBuilder {
	return &CardBuilder{variant: "shadowed", padding: "p-6"}
}

// CardHeader sets the card header node.
func (c *CardBuilder) CardHeader(n *Node) *CardBuilder { c.header = n; return c }

// CardBody sets the card body node.
func (c *CardBuilder) CardBody(n *Node) *CardBuilder { c.body = n; return c }

// CardFooter sets the card footer node.
func (c *CardBuilder) CardFooter(n *Node) *CardBuilder { c.footer = n; return c }

// CardImage sets the card image source URL and alt text.
func (c *CardBuilder) CardImage(src, alt string) *CardBuilder {
	c.image = src
	c.imageAlt = alt
	return c
}

// CardImageSize sets explicit width and height on the card image to prevent
// layout shifts (CLS). Values should be the intrinsic pixel dimensions.
func (c *CardBuilder) CardImageSize(width, height string) *CardBuilder {
	c.imageWidth = width
	c.imageHeight = height
	return c
}

// CardImagePriority marks the card image with fetchpriority="high" for LCP
// optimization. Use on the most important above-the-fold card image.
func (c *CardBuilder) CardImagePriority(p bool) *CardBuilder {
	c.imagePriority = p
	return c
}

// CardVariant sets the card variant: "shadowed", "bordered", "flat", or "glass".
func (c *CardBuilder) CardVariant(v string) *CardBuilder { c.variant = v; return c }

// CardHover enables a lift effect on hover.
func (c *CardBuilder) CardHover(h bool) *CardBuilder { c.hover = h; return c }

// CardCompact reduces padding and image height.
func (c *CardBuilder) CardCompact(cp bool) *CardBuilder { c.compact = cp; return c }

// CardPadding overrides the default padding class.
func (c *CardBuilder) CardPadding(p string) *CardBuilder { c.padding = p; return c }

// CardClass appends additional CSS classes to the card container.
func (c *CardBuilder) CardClass(cls string) *CardBuilder { c.class = cls; return c }

// Build compiles the card into a *Node.
func (c *CardBuilder) Build() *Node {
	pad := c.padding
	if c.compact {
		pad = "p-4"
	}

	cls := "bg-white dark:bg-gray-900 rounded-xl overflow-hidden"
	switch c.variant {
	case "bordered":
		cls += " border border-gray-200 dark:border-gray-700"
	case "flat":
		// no border or shadow
	case "glass":
		cls += " backdrop-blur-lg bg-white/70 dark:bg-gray-900/70 border border-white/20"
	default:
		cls += " shadow-md border border-gray-100 dark:border-gray-800"
	}

	if c.hover {
		cls += " hover:shadow-lg hover:-translate-y-1 transition-all duration-200"
	}
	if c.class != "" {
		cls += " " + c.class
	}

	card := Div(cls)

	if c.image != "" {
		imgH := "h-48"
		if c.compact {
			imgH = "h-32"
		}
		img := Img("w-full object-cover "+imgH).Attr("src", c.image).Attr("alt", c.imageAlt).Attr("decoding", "async")
		if c.imageWidth != "" {
			img.Attr("width", c.imageWidth)
		}
		if c.imageHeight != "" {
			img.Attr("height", c.imageHeight)
		}
		if c.imagePriority {
			img.Attr("fetchpriority", "high")
			img.Attr("loading", "eager")
		} else {
			img.Attr("loading", "lazy")
		}
		card.Render(img)
	}
	if c.header != nil {
		card.Render(
			Div("border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50 " + pad).Render(c.header),
		)
	}
	if c.body != nil {
		card.Render(Div(pad).Render(c.body))
	}
	if c.footer != nil {
		card.Render(
			Div("border-t border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50 " + pad).Render(c.footer),
		)
	}

	return card
}

// ---------------------------------------------------------------------------
// 10. Confirm Dialog
// ---------------------------------------------------------------------------

// ConfirmLocale holds translatable strings for ConfirmDialog.
type ConfirmLocale struct {
	Cancel  string
	Confirm string
}

// ConfirmDialog creates a fixed-overlay confirmation dialog with title,
// message, a confirm button (red), and a cancel button.
// ConfirmOpt configures optional ConfirmDialog settings.
type ConfirmOpt struct {
	CancelAction *Action        // custom cancel action; nil = dismiss overlay
	Locale       *ConfirmLocale // per-instance locale; nil = English default
}

func ConfirmDialog(title, message string, confirmAction *Action, opts ...ConfirmOpt) *Node {
	overlayID := Target()
	loc := &ConfirmLocale{Cancel: "Cancel", Confirm: "Confirm"}
	var cancel *Action
	if len(opts) > 0 {
		if opts[0].CancelAction != nil {
			cancel = opts[0].CancelAction
		}
		if opts[0].Locale != nil {
			loc = opts[0].Locale
		}
	}
	if cancel == nil {
		cancel = JS(RemoveEl(overlayID))
	}

	overlay := Div("fixed inset-0 z-[10000] flex items-center justify-center bg-black/50").ID(overlayID)

	inner := Div("bg-white dark:bg-gray-900 rounded-xl shadow-2xl p-6 max-w-md w-full").Render(
		H3("text-lg font-semibold text-gray-900 dark:text-white").Text(title),
		P("text-sm text-gray-600 dark:text-gray-400 mt-2").Text(message),
		Div("flex justify-end gap-3 mt-6").Render(
			Button("px-4 py-2 text-sm font-medium rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer").
				Text(loc.Cancel).OnClick(cancel),
			Button("px-4 py-2 text-sm font-medium rounded-lg bg-red-600 text-white hover:bg-red-700 cursor-pointer").
				Text(loc.Confirm).OnClick(confirmAction),
		),
	)

	return overlay.Render(inner)
}

// ---------------------------------------------------------------------------
// 11. Dropdown Builder
// ---------------------------------------------------------------------------

type dropdownItem struct {
	label   string
	action  *Action
	icon    string
	danger  bool
	divider bool
	header  bool
}

// DropdownBuilder constructs a dropdown menu attached to a trigger node.
type DropdownBuilder struct {
	trigger  *Node
	items    []dropdownItem
	position string
	class    string
}

// NewDropdown creates a new DropdownBuilder with the given trigger node.
func NewDropdown(trigger *Node) *DropdownBuilder {
	return &DropdownBuilder{trigger: trigger, position: "bottom-left"}
}

// DropdownItem adds a menu item with label, action, and optional icon.
func (d *DropdownBuilder) DropdownItem(label string, action *Action, icon ...string) *DropdownBuilder {
	it := dropdownItem{label: label, action: action}
	if len(icon) > 0 {
		it.icon = icon[0]
	}
	d.items = append(d.items, it)
	return d
}

// DropdownDanger adds a danger-styled menu item.
func (d *DropdownBuilder) DropdownDanger(label string, action *Action, icon ...string) *DropdownBuilder {
	it := dropdownItem{label: label, action: action, danger: true}
	if len(icon) > 0 {
		it.icon = icon[0]
	}
	d.items = append(d.items, it)
	return d
}

// DropdownHeader adds a non-interactive header label.
func (d *DropdownBuilder) DropdownHeader(label string) *DropdownBuilder {
	d.items = append(d.items, dropdownItem{label: label, header: true})
	return d
}

// DropdownDivider adds a visual separator.
func (d *DropdownBuilder) DropdownDivider() *DropdownBuilder {
	d.items = append(d.items, dropdownItem{divider: true})
	return d
}

// DropdownPosition sets the dropdown position relative to the trigger.
func (d *DropdownBuilder) DropdownPosition(p string) *DropdownBuilder { d.position = p; return d }

// DropdownClass appends additional CSS classes to the wrapper.
func (d *DropdownBuilder) DropdownClass(cls string) *DropdownBuilder { d.class = cls; return d }

// Build compiles the dropdown into a *Node.
func (d *DropdownBuilder) Build() *Node {
	wrapperID := Target()
	menuID := Target()

	var posClass string
	switch d.position {
	case "bottom-right":
		posClass = "top-full right-0 mt-1"
	case "top-left":
		posClass = "bottom-full left-0 mb-1"
	case "top-right":
		posClass = "bottom-full right-0 mb-1"
	default:
		posClass = "top-full left-0 mt-1"
	}

	wrapCls := "relative inline-block"
	if d.class != "" {
		wrapCls += " " + d.class
	}

	wrapper := Div(wrapCls).ID(wrapperID)

	d.trigger.OnClick(JS(fmt.Sprintf(
		"var m=document.getElementById('%s');m.classList.toggle('hidden');",
		escJS(menuID),
	)))
	wrapper.Render(d.trigger)

	menuCls := "absolute z-50 min-w-[12rem] bg-white dark:bg-gray-900 rounded-xl shadow-xl border border-gray-200 dark:border-gray-700 py-1.5 hidden " + posClass
	menu := Div(menuCls).ID(menuID)

	for _, item := range d.items {
		if item.divider {
			menu.Render(Div("my-1 border-t border-gray-100 dark:border-gray-800"))
			continue
		}
		if item.header {
			menu.Render(
				Div("px-3 py-1.5 text-xs font-semibold text-gray-400 uppercase tracking-wider").Text(item.label),
			)
			continue
		}

		btnCls := "w-full text-left px-3 py-2 text-sm flex items-center gap-2 hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer"
		if item.danger {
			btnCls = "w-full text-left px-3 py-2 text-sm flex items-center gap-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-950 cursor-pointer"
		}

		btn := Button(btnCls)
		if item.icon != "" {
			btn.Render(Span("material-icons-round text-base").Text(item.icon))
		}
		btn.Render(Span().Text(item.label))
		if item.action != nil {
			btn.OnClick(item.action)
		}
		menu.Render(btn)
	}

	wrapper.Render(menu)

	wrapper.JS(fmt.Sprintf(
		`(function(){`+
			`var w=document.getElementById('%s');`+
			`var m=document.getElementById('%s');`+
			`document.addEventListener('click',function(e){`+
			`if(!w.contains(e.target)){m.classList.add('hidden')}`+
			`},true);`+
			`document.addEventListener('keydown',function(e){`+
			`if(e.key==='Escape'){m.classList.add('hidden')}`+
			`});`+
			`})();`,
		escJS(wrapperID), escJS(menuID),
	))

	return wrapper
}

// ---------------------------------------------------------------------------
// 12. Progress Bar Builder
// ---------------------------------------------------------------------------

// ProgressBuilder constructs a progress bar component.
type ProgressBuilder struct {
	value         int
	color         string
	gradient      []string
	size          string
	striped       bool
	animated      bool
	indeterminate bool
	label         string
	labelPos      string
	class         string
}

// NewProgress creates a new ProgressBuilder with default settings.
func NewProgress() *ProgressBuilder {
	return &ProgressBuilder{color: "bg-blue-600", size: "md", labelPos: "inside"}
}

// ProgressValue sets the progress percentage, clamped to 0-100.
func (p *ProgressBuilder) ProgressValue(v int) *ProgressBuilder {
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	p.value = v
	return p
}

// ProgressColor sets the bar background color class.
func (p *ProgressBuilder) ProgressColor(c string) *ProgressBuilder { p.color = c; return p }

// ProgressGradient sets gradient colors (overrides solid color).
func (p *ProgressBuilder) ProgressGradient(colors ...string) *ProgressBuilder {
	p.gradient = colors
	return p
}

// ProgressSize sets the bar height: "xs", "sm", "md", "lg", "xl".
func (p *ProgressBuilder) ProgressSize(s string) *ProgressBuilder { p.size = s; return p }

// Striped enables diagonal stripe pattern on the bar.
func (p *ProgressBuilder) Striped(s bool) *ProgressBuilder { p.striped = s; return p }

// Animated enables animated stripe movement (implies striped).
func (p *ProgressBuilder) Animated(a bool) *ProgressBuilder { p.animated = a; return p }

// Indeterminate shows a bouncing bar with no specific value.
func (p *ProgressBuilder) Indeterminate(i bool) *ProgressBuilder { p.indeterminate = i; return p }

// ProgressLabel sets a text label to display on or above the bar.
func (p *ProgressBuilder) ProgressLabel(l string) *ProgressBuilder { p.label = l; return p }

// LabelPosition sets the label position: "inside" or "outside".
func (p *ProgressBuilder) LabelPosition(pos string) *ProgressBuilder { p.labelPos = pos; return p }

// ProgressClass appends additional CSS classes to the outer wrapper.
func (p *ProgressBuilder) ProgressClass(cls string) *ProgressBuilder { p.class = cls; return p }

func progressHeight(size string) string {
	switch size {
	case "xs":
		return "h-1"
	case "sm":
		return "h-1.5"
	case "lg":
		return "h-4"
	case "xl":
		return "h-6"
	default:
		return "h-2.5"
	}
}

// Build compiles the progress bar into a *Node.
func (p *ProgressBuilder) Build() *Node {
	height := progressHeight(p.size)

	wrapCls := "w-full"
	if p.class != "" {
		wrapCls += " " + p.class
	}
	wrapper := Div(wrapCls)

	if p.label != "" && p.labelPos == "outside" {
		wrapper.Render(
			Div("flex justify-between items-center mb-1").Render(
				Span("text-sm font-medium text-gray-700 dark:text-gray-300").Text(p.label),
				Span("text-sm font-medium text-gray-700 dark:text-gray-300").Text(fmt.Sprintf("%d%%", p.value)),
			),
		)
	}

	containerCls := "w-full overflow-hidden bg-gray-200 dark:bg-gray-800 rounded-full " + height
	container := Div(containerCls)

	barScale := fmt.Sprintf("%.4f", float64(p.value)/100)
	if p.indeterminate {
		barScale = "0.33"
	}

	barCls := "h-full rounded-full origin-left transform-gpu transition-transform duration-500 will-change-transform"
	if len(p.gradient) == 0 {
		barCls += " " + p.color
	}

	bar := Div(barCls).Style("transform", "scaleX("+barScale+")")

	if len(p.gradient) > 0 {
		bar.Style("background", "linear-gradient(90deg, "+strings.Join(p.gradient, ", ")+")")
	}

	if p.striped || p.animated {
		bar.Style("backgroundImage", "linear-gradient(45deg, rgba(255,255,255,.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,.15) 50%, rgba(255,255,255,.15) 75%, transparent 75%, transparent)")
		bar.Style("backgroundSize", "1rem 1rem")
	}

	if p.label != "" && p.labelPos == "inside" && (p.size == "lg" || p.size == "xl") {
		bar.Class("flex items-center justify-center")
		bar.Render(Span("text-xs font-medium text-white leading-none").Text(p.label))
	}

	container.Render(bar)

	if p.animated {
		animID := Target()
		css := fmt.Sprintf("@keyframes %s{from{background-position:1rem 0}to{background-position:0 0}}", animID)
		wrapper.Render(El("style").Text(css))
		bar.Style("animation", animID+" 1s linear infinite")
	}

	if p.indeterminate {
		animID := Target()
		css := fmt.Sprintf("@keyframes %s{0%%{transform:translateX(-100%%) scaleX(0.33)}50%%{transform:translateX(200%%) scaleX(0.33)}100%%{transform:translateX(-100%%) scaleX(0.33)}}", animID)
		wrapper.Render(El("style").Text(css))
		bar.Style("animation", animID+" 1.5s ease-in-out infinite")
	}

	wrapper.Render(container)
	return wrapper
}

// ---------------------------------------------------------------------------
// 13. Step Progress Builder
// ---------------------------------------------------------------------------

// StepProgressBuilder constructs a simple step progress indicator.
type StepProgressBuilder struct {
	current int
	total   int
	color   string
	size    string
	class   string
	locale  *StepProgressLocale
}

// StepProgressLocale holds translatable strings for StepProgress.
type StepProgressLocale struct {
	// StepOf formats "Step X of Y" — receives (current, total).
	StepOf func(current, total int) string
}

// NewStepProgress creates a new StepProgressBuilder.
func NewStepProgress(current, total int) *StepProgressBuilder {
	return &StepProgressBuilder{current: current, total: total, color: "bg-blue-500", size: "md"}
}

// StepColor sets the fill bar color class.
func (s *StepProgressBuilder) StepColor(c string) *StepProgressBuilder { s.color = c; return s }

// StepSize sets the bar height: "xs", "sm", "md", "lg", "xl".
func (s *StepProgressBuilder) StepSize(sz string) *StepProgressBuilder { s.size = sz; return s }

// StepClass appends additional CSS classes.
func (s *StepProgressBuilder) StepClass(cls string) *StepProgressBuilder { s.class = cls; return s }

// Locale sets a per-instance locale.
func (s *StepProgressBuilder) Locale(l *StepProgressLocale) *StepProgressBuilder {
	s.locale = l
	return s
}

func (s *StepProgressBuilder) loc() *StepProgressLocale {
	if s.locale != nil {
		return s.locale
	}
	return &StepProgressLocale{
		StepOf: func(current, total int) string { return fmt.Sprintf("Step %d of %d", current, total) },
	}
}

func stepHeight(size string) string {
	switch size {
	case "xs":
		return "h-0.5"
	case "sm":
		return "h-1"
	case "lg":
		return "h-2"
	case "xl":
		return "h-3"
	default:
		return "h-1.5"
	}
}

// Build compiles the step progress into a *Node.
func (s *StepProgressBuilder) Build() *Node {
	cur := s.current
	if cur < 0 {
		cur = 0
	}
	if cur > s.total {
		cur = s.total
	}

	pct := 0
	if s.total > 0 {
		pct = cur * 100 / s.total
	}

	height := stepHeight(s.size)

	wrapCls := "w-full"
	if s.class != "" {
		wrapCls += " " + s.class
	}

	wrapper := Div(wrapCls)
	wrapper.Render(
		Span("text-sm font-medium text-gray-500 dark:text-gray-400 mb-1").
			Text(s.loc().StepOf(cur, s.total)),
	)

	containerCls := "w-full bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden " + height
	container := Div(containerCls)
	barCls := "h-full " + s.color + " rounded-full origin-left transform-gpu transition-transform duration-500 will-change-transform"
	bar := Div(barCls).Style("transform", fmt.Sprintf("scaleX(%.4f)", float64(pct)/100))

	container.Render(bar)
	wrapper.Render(container)
	return wrapper
}

// ---------------------------------------------------------------------------
// 14. Tabs Builder
// ---------------------------------------------------------------------------

type tabItem struct {
	label   string
	content *Node
	icon    string
}

// TabsBuilder constructs a tabbed interface with multiple panels.
type TabsBuilder struct {
	tabs     []tabItem
	active   int
	tabStyle string
	class    string
}

// NewTabs creates a new TabsBuilder.
func NewTabs() *TabsBuilder {
	return &TabsBuilder{tabStyle: "underline"}
}

// Tab adds a tab with label, content panel, and optional icon.
func (t *TabsBuilder) Tab(label string, content *Node, icon ...string) *TabsBuilder {
	it := tabItem{label: label, content: content}
	if len(icon) > 0 {
		it.icon = icon[0]
	}
	t.tabs = append(t.tabs, it)
	return t
}

// Active sets the initially active tab index (0-based).
func (t *TabsBuilder) Active(index int) *TabsBuilder { t.active = index; return t }

// TabStyle sets the tab style: "underline", "pills", "boxed", "vertical".
func (t *TabsBuilder) TabStyle(s string) *TabsBuilder { t.tabStyle = s; return t }

// TabsClass appends additional CSS classes.
func (t *TabsBuilder) TabsClass(cls string) *TabsBuilder { t.class = cls; return t }

// Build compiles the tabs into a *Node.
func (t *TabsBuilder) Build() *Node {
	containerID := Target()

	containerCls := ""
	if t.tabStyle == "vertical" {
		containerCls = "flex gap-4"
	}
	if t.class != "" {
		if containerCls != "" {
			containerCls += " " + t.class
		} else {
			containerCls = t.class
		}
	}

	container := Div(containerCls).ID(containerID).Attr("data-tabs-active", fmt.Sprintf("%d", t.active))

	var tabListCls string
	switch t.tabStyle {
	case "pills":
		tabListCls = "flex gap-1"
	case "boxed":
		tabListCls = "flex gap-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg"
	case "vertical":
		tabListCls = "flex flex-col gap-1 min-w-[10rem]"
	default:
		tabListCls = "flex border-b border-gray-200 dark:border-gray-700"
	}

	tabList := Div(tabListCls).Attr("role", "tablist")

	btnIDs := make([]string, len(t.tabs))
	panelIDs := make([]string, len(t.tabs))
	for i := range t.tabs {
		btnIDs[i] = Target()
		panelIDs[i] = Target()
	}

	for i, tab := range t.tabs {
		isActive := i == t.active

		var btnCls string
		switch t.tabStyle {
		case "pills":
			if isActive {
				btnCls = "px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded-lg cursor-pointer"
			} else {
				btnCls = "px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg cursor-pointer"
			}
		case "boxed":
			if isActive {
				btnCls = "px-4 py-2 text-sm font-medium bg-white dark:bg-gray-900 shadow-sm rounded-lg cursor-pointer"
			} else {
				btnCls = "px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 rounded-lg cursor-pointer"
			}
		case "vertical":
			if isActive {
				btnCls = "px-4 py-2 text-sm font-medium text-left border-l-2 border-blue-600 bg-blue-50 dark:bg-blue-950 text-blue-600 cursor-pointer"
			} else {
				btnCls = "px-4 py-2 text-sm font-medium text-left text-gray-600 dark:text-gray-400 border-l-2 border-transparent hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
			}
		default:
			if isActive {
				btnCls = "px-4 py-2 text-sm font-medium border-b-2 border-blue-600 text-blue-600 -mb-px cursor-pointer"
			} else {
				btnCls = "px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-400 border-b-2 border-transparent hover:text-gray-900 dark:hover:text-gray-200 -mb-px cursor-pointer"
			}
		}

		btn := Button(btnCls).ID(btnIDs[i]).
			Attr("role", "tab").
			Attr("data-tab-index", fmt.Sprintf("%d", i)).
			Attr("aria-selected", fmt.Sprintf("%t", isActive)).
			Attr("aria-controls", panelIDs[i])

		if tab.icon != "" {
			btn.Render(Span("material-icons-round text-base mr-1.5 align-middle").Text(tab.icon))
		}
		btn.Render(Span().Text(tab.label))
		tabList.Render(btn)
	}

	container.Render(tabList)

	panelWrapper := Div()
	if t.tabStyle == "vertical" {
		panelWrapper.Class("flex-1")
	}

	for i, tab := range t.tabs {
		isActive := i == t.active
		panelCls := "pt-4"
		if !isActive {
			panelCls += " hidden"
		}
		panel := Div(panelCls).ID(panelIDs[i]).
			Attr("role", "tabpanel").
			Attr("data-tab-index", fmt.Sprintf("%d", i))
		if tab.content != nil {
			panel.Render(tab.content)
		}
		panelWrapper.Render(panel)
	}

	container.Render(panelWrapper)

	// JS for tab switching and keyboard navigation
	var js strings.Builder
	js.WriteString("(function(){")
	js.WriteString(fmt.Sprintf("var c=document.getElementById('%s');", escJS(containerID)))
	js.WriteString("if(!c)return;")
	js.WriteString("var btns=[")
	for i, id := range btnIDs {
		if i > 0 {
			js.WriteString(",")
		}
		fmt.Fprintf(&js, "'%s'", escJS(id))
	}
	js.WriteString("];var pans=[")
	for i, id := range panelIDs {
		if i > 0 {
			js.WriteString(",")
		}
		fmt.Fprintf(&js, "'%s'", escJS(id))
	}
	js.WriteString("];")

	var activeBtn, inactiveBtn string
	switch t.tabStyle {
	case "pills":
		activeBtn = "bg-blue-600 text-white rounded-lg"
		inactiveBtn = "text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg"
	case "boxed":
		activeBtn = "bg-white dark:bg-gray-900 shadow-sm rounded-lg"
		inactiveBtn = "text-gray-600 dark:text-gray-400 rounded-lg"
	case "vertical":
		activeBtn = "border-l-2 border-blue-600 bg-blue-50 dark:bg-blue-950 text-blue-600"
		inactiveBtn = "border-l-2 border-transparent text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800"
	default:
		activeBtn = "border-b-2 border-blue-600 text-blue-600"
		inactiveBtn = "text-gray-600 dark:text-gray-400 border-b-2 border-transparent hover:text-gray-900 dark:hover:text-gray-200"
	}

	fmt.Fprintf(&js, "var aCls='%s'.split(' ');", escJS(activeBtn))
	fmt.Fprintf(&js, "var iCls='%s'.split(' ');", escJS(inactiveBtn))

	js.WriteString("function activate(idx){btns.forEach(function(id,i){var b=document.getElementById(id);if(!b)return;var p=document.getElementById(pans[i]);if(i===idx){iCls.forEach(function(cl){if(cl)b.classList.remove(cl)});aCls.forEach(function(cl){if(cl)b.classList.add(cl)});b.setAttribute('aria-selected','true');if(p)p.classList.remove('hidden')}else{aCls.forEach(function(cl){if(cl)b.classList.remove(cl)});iCls.forEach(function(cl){if(cl)b.classList.add(cl)});b.setAttribute('aria-selected','false');if(p)p.classList.add('hidden')}});c.setAttribute('data-tabs-active',idx)}")
	js.WriteString("btns.forEach(function(id,i){var b=document.getElementById(id);if(!b)return;b.addEventListener('click',function(){activate(i)})});")

	if t.tabStyle == "vertical" {
		js.WriteString("c.addEventListener('keydown',function(e){var cur=parseInt(c.getAttribute('data-tabs-active'))||0;if(e.key==='ArrowDown'){e.preventDefault();activate((cur+1)%btns.length);document.getElementById(btns[(cur+1)%btns.length]).focus()}else if(e.key==='ArrowUp'){e.preventDefault();activate((cur-1+btns.length)%btns.length);document.getElementById(btns[(cur-1+btns.length)%btns.length]).focus()}});")
	} else {
		js.WriteString("c.addEventListener('keydown',function(e){var cur=parseInt(c.getAttribute('data-tabs-active'))||0;if(e.key==='ArrowRight'){e.preventDefault();activate((cur+1)%btns.length);document.getElementById(btns[(cur+1)%btns.length]).focus()}else if(e.key==='ArrowLeft'){e.preventDefault();activate((cur-1+btns.length)%btns.length);document.getElementById(btns[(cur-1+btns.length)%btns.length]).focus()}});")
	}

	js.WriteString("})();")
	container.JS(js.String())

	return container
}

// ---------------------------------------------------------------------------
// 15. Tooltip Builder
// ---------------------------------------------------------------------------

// TooltipBuilder wraps a trigger node with a tooltip overlay shown on hover.
type TooltipBuilder struct {
	content  string
	position string
	variant  string
	delay    int
	class    string
}

// NewTooltip creates a new TooltipBuilder with the given text content.
func NewTooltip(content string) *TooltipBuilder {
	return &TooltipBuilder{content: content, position: "top", variant: "dark", delay: 200}
}

// TooltipPosition sets the tooltip position: "top", "bottom", "left", "right".
func (t *TooltipBuilder) TooltipPosition(p string) *TooltipBuilder { t.position = p; return t }

// TooltipVariant sets the tooltip color variant.
func (t *TooltipBuilder) TooltipVariant(v string) *TooltipBuilder { t.variant = v; return t }

// Delay sets the show delay in milliseconds. 0 = pure CSS (instant).
func (t *TooltipBuilder) Delay(ms int) *TooltipBuilder { t.delay = ms; return t }

// TooltipClass appends additional CSS classes.
func (t *TooltipBuilder) TooltipClass(cls string) *TooltipBuilder { t.class = cls; return t }

// Wrap wraps the given trigger node in a tooltip container and returns it.
func (t *TooltipBuilder) Wrap(trigger *Node) *Node {
	wrapperID := Target()
	tooltipID := Target()

	wrapper := Div("relative inline-block group").ID(wrapperID)
	wrapper.Render(trigger)

	var posClass string
	switch t.position {
	case "bottom":
		posClass = "top-full left-1/2 -translate-x-1/2 mt-2"
	case "left":
		posClass = "right-full top-1/2 -translate-y-1/2 mr-2"
	case "right":
		posClass = "left-full top-1/2 -translate-y-1/2 ml-2"
	default:
		posClass = "bottom-full left-1/2 -translate-x-1/2 mb-2"
	}

	var variantClass string
	switch t.variant {
	case "light":
		variantClass = "bg-white text-gray-900 border border-gray-200 shadow-sm"
	case "blue":
		variantClass = "bg-blue-600 text-white"
	case "green":
		variantClass = "bg-green-600 text-white"
	case "red":
		variantClass = "bg-red-600 text-white"
	case "yellow":
		variantClass = "bg-yellow-500 text-gray-900"
	default:
		variantClass = "bg-gray-900 text-white dark:bg-gray-100 dark:text-gray-900"
	}

	baseCls := "absolute z-[100] px-2.5 py-1.5 text-xs font-medium rounded-lg whitespace-nowrap pointer-events-none"
	if t.delay == 0 {
		baseCls += " opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-150"
	} else {
		baseCls += " opacity-0 invisible transition-all duration-150"
	}

	tooltipCls := baseCls + " " + posClass + " " + variantClass
	if t.class != "" {
		tooltipCls += " " + t.class
	}

	tooltip := Div(tooltipCls).ID(tooltipID).Text(t.content)
	tooltip.Render(tooltipArrow(t.position, t.variant))
	wrapper.Render(tooltip)

	if t.delay > 0 {
		wrapper.JS(fmt.Sprintf(
			`(function(){`+
				`var w=document.getElementById('%s');`+
				`var tip=document.getElementById('%s');`+
				`var timer=null;`+
				`w.addEventListener('mouseenter',function(){`+
				`timer=setTimeout(function(){`+
				`tip.classList.remove('opacity-0','invisible');`+
				`tip.classList.add('opacity-100','visible');`+
				`},%d);`+
				`});`+
				`w.addEventListener('mouseleave',function(){`+
				`if(timer){clearTimeout(timer);timer=null}`+
				`tip.classList.remove('opacity-100','visible');`+
				`tip.classList.add('opacity-0','invisible');`+
				`});`+
				`})();`,
			escJS(wrapperID), escJS(tooltipID), t.delay,
		))
	}

	return wrapper
}

// ---------------------------------------------------------------------------
// 16. Theme Switcher
// ---------------------------------------------------------------------------

// ThemeSwitcherLocale holds translatable strings for ThemeSwitcher.
type ThemeSwitcherLocale struct {
	ThemeAuto  string
	ThemeLight string
	ThemeDark  string
}

// ThemeSwitcherOpt configures optional ThemeSwitcher settings.
type ThemeSwitcherOpt struct {
	Class  string               // additional CSS class
	Locale *ThemeSwitcherLocale // per-instance locale; nil = English default
}

// ThemeSwitcher renders a tri-state toggle button that cycles through
// System → Light → Dark themes. It reads from localStorage("theme"),
// calls window.setTheme(mode), and updates its icon+label reactively.
func ThemeSwitcher(opts ...ThemeSwitcherOpt) *Node {
	btnID := Target()
	loc := &ThemeSwitcherLocale{ThemeAuto: "Auto", ThemeLight: "Light", ThemeDark: "Dark"}
	extraClass := ""
	if len(opts) > 0 {
		if opts[0].Locale != nil {
			loc = opts[0].Locale
		}
		extraClass = opts[0].Class
	}

	baseCls := "inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer " +
		"border border-gray-200 bg-white text-gray-700 hover:bg-gray-50 " +
		"dark:bg-gray-800 dark:text-gray-200 dark:border-gray-600 dark:hover:bg-gray-700 " +
		"transition-colors"
	if extraClass != "" {
		baseCls += " " + extraClass
	}

	btn := Button(baseCls).ID(btnID).Render(
		I("material-icons-round text-base").ID(btnID+"-icon").Text("brightness_auto"),
		Span().ID(btnID+"-label").Text(loc.ThemeAuto),
	)

	js := fmt.Sprintf(
		`(function(){`+
			`var btn=document.getElementById('%s');`+
			`var icon=document.getElementById('%s-icon');`+
			`var lbl=document.getElementById('%s-label');`+
			`if(!btn)return;`+
			`var modes=['system','light','dark'];`+
			`var icons={system:'brightness_auto',light:'light_mode',dark:'dark_mode'};`+
			`var labels={system:'%s',light:'%s',dark:'%s'};`+
			`function upd(){var m=localStorage.getItem('theme')||'system';icon.textContent=icons[m]||icons.system;lbl.textContent=labels[m]||labels.system}`+
			`btn.addEventListener('click',function(){`+
			`var cur=localStorage.getItem('theme')||'system';`+
			`var idx=(modes.indexOf(cur)+1)%%3;`+
			`setTheme(modes[idx]);upd()});`+
			`upd();`+
			`window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change',function(){upd()});`+
			`})();`,
		escJS(btnID), escJS(btnID), escJS(btnID),
		escJS(loc.ThemeAuto), escJS(loc.ThemeLight), escJS(loc.ThemeDark),
	)

	return btn.JS(js)
}

// tooltipArrow builds the CSS triangle arrow for a tooltip.
func tooltipArrow(position, variant string) *Node {
	arrow := Div()
	arrow.Style("position", "absolute")
	arrow.Style("width", "0")
	arrow.Style("height", "0")

	color := "#111827"
	switch variant {
	case "light":
		color = "white"
	case "blue":
		color = "#2563eb"
	case "green":
		color = "#16a34a"
	case "red":
		color = "#dc2626"
	case "yellow":
		color = "#eab308"
	}

	switch position {
	case "bottom":
		arrow.Style("top", "-4px")
		arrow.Style("left", "50%")
		arrow.Style("transform", "translateX(-50%)")
		arrow.Style("borderLeft", "4px solid transparent")
		arrow.Style("borderRight", "4px solid transparent")
		arrow.Style("borderBottom", "4px solid "+color)
	case "left":
		arrow.Style("right", "-4px")
		arrow.Style("top", "50%")
		arrow.Style("transform", "translateY(-50%)")
		arrow.Style("borderTop", "4px solid transparent")
		arrow.Style("borderBottom", "4px solid transparent")
		arrow.Style("borderLeft", "4px solid "+color)
	case "right":
		arrow.Style("left", "-4px")
		arrow.Style("top", "50%")
		arrow.Style("transform", "translateY(-50%)")
		arrow.Style("borderTop", "4px solid transparent")
		arrow.Style("borderBottom", "4px solid transparent")
		arrow.Style("borderRight", "4px solid "+color)
	default:
		arrow.Style("bottom", "-4px")
		arrow.Style("left", "50%")
		arrow.Style("transform", "translateX(-50%)")
		arrow.Style("borderLeft", "4px solid transparent")
		arrow.Style("borderRight", "4px solid transparent")
		arrow.Style("borderTop", "4px solid "+color)
	}

	return arrow
}
