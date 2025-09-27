// Package ui, holds components for web application
package ui

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/yuin/goldmark"
)

const (
	XS = " p-1"
	SM = " p-2"
	MD = " p-3"
	ST = " p-4"
	LG = " p-5"
	XL = " p-6"
)

const (
	// DISABLED      = " cursor-text bg-gray-100 pointer-events-none"
	AREA          = " cursor-pointer bg-white border border-gray-300 hover:border-blue-500 rounded-lg block w-full"
	INPUT         = " cursor-pointer bg-white border border-gray-300 hover:border-blue-500 rounded-lg block w-full h-12"
	VALUE         = " bg-white border border-gray-300 hover:border-blue-500 rounded-lg block h-12"
	BTN           = " cursor-pointer font-bold text-center select-none"
	DISABLED      = " cursor-text pointer-events-none bg-gray-50"
	Yellow        = " bg-yellow-400 text-gray-800 hover:text-gray-200 hover:bg-yellow-600 font-bold border-gray-300 flex items-center justify-center"
	YellowOutline = " border border-yellow-500 text-yellow-600 hover:text-gray-700 hover:bg-yellow-500 flex items-center justify-center"
	Green         = " bg-green-600 text-white hover:bg-green-700 checked:bg-green-600 border-gray-300 flex items-center justify-center"
	GreenOutline  = " border border-green-500 text-green-500 hover:text-white hover:bg-green-600 flex items-center justify-center"
	Purple        = " bg-purple-500 text-white hover:bg-purple-700 border-purple-500 flex items-center justify-center"
	PurpleOutline = " border border-purple-500 text-purple-500 hover:text-white hover:bg-purple-600 flex items-center justify-center"
	Blue          = " bg-blue-800 text-white hover:bg-blue-700 border-gray-300 flex items-center justify-center"
	BlueOutline   = " border border-blue-500 text-blue-600 hover:text-white hover:bg-blue-700 checked:bg-blue-700 flex items-center justify-center"
	Red           = " bg-red-600 text-white hover:bg-red-800 border-gray-300 flex items-center justify-center"
	RedOutline    = " border border-red-500 text-red-600 hover:text-white hover:bg-red-700 flex items-center justify-center"
	Gray          = " bg-gray-600 text-white hover:bg-gray-800 focus:bg-gray-800 border-gray-300 flex items-center justify-center"
	GrayOutline   = " border border-gray-300 text-black hover:text-white hover:bg-gray-700 flex items-center justify-center"
	White         = " bg-white text-black hover:bg-gray-200 border-gray-200 flex items-center justify-center"
	WhiteOutline  = " border border-white text-balck hover:text-black hover:bg-white flex items-center justify-center"
)

type Attr struct {
	OnClick      string
	Step         string
	ID           string
	Href         string
	Title        string
	Alt          string
	Type         string
	Class        string
	Style        string
	Name         string
	Value        string
	Checked      string
	OnSubmit     string
	For          string
	Src          string
	Selected     string
	Pattern      string
	Placeholder  string
	Autocomplete string
	OnChange     string
	Max          string
	Min          string
	Target       string
	Rows         uint8
	Cols         uint8
	Width        uint8
	Height       uint8
	Disabled     bool
	Required     bool
	Readonly     bool
}

// TargetSwap pairs a target id with a swap strategy for convenient patch calls.
type TargetSwap struct {
    ID   string
    Swap Swap
}

type AOption struct {
    ID    string
    Value string
}

func MakeOptions(options []string) []AOption {
	var result []AOption

	for _, option := range options {
		result = append(result, AOption{ID: option, Value: option})
	}

	return result
}

// escapeAttr escapes HTML attribute values to prevent XSS attacks
func escapeAttr(s string) string {
	return html.EscapeString(s)
}

// escapeJS escapes JavaScript string literals to prevent XSS attacks
func escapeJS(s string) string {
	// Replace dangerous characters for JavaScript string literals
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	s = strings.ReplaceAll(s, "</script>", `<\/script>`)
	return s
}

func attributes(attrs ...Attr) string {
	var result []string

	for _, attr := range attrs {

		if attr.ID != "" {
			result = append(result, fmt.Sprintf(`id="%s"`, escapeAttr(attr.ID)))
		}

		if attr.Href != "" {
			result = append(result, fmt.Sprintf(`href="%s"`, escapeAttr(attr.Href)))
		}

		if attr.Alt != "" {
			result = append(result, fmt.Sprintf(`alt="%s"`, escapeAttr(attr.Alt)))
		}

		if attr.Title != "" {
			result = append(result, fmt.Sprintf(`title="%s"`, escapeAttr(attr.Title)))
		}

		if attr.Src != "" {
			result = append(result, fmt.Sprintf(`src="%s"`, escapeAttr(attr.Src)))
		}

		if attr.For != "" {
			result = append(result, fmt.Sprintf(`for="%s"`, escapeAttr(attr.For)))
		}

		if attr.Type != "" {
			result = append(result, fmt.Sprintf(`type="%s"`, escapeAttr(attr.Type)))
		}

		if attr.Class != "" {
			result = append(result, fmt.Sprintf(`class="%s"`, escapeAttr(attr.Class)))
		}

		if attr.Style != "" {
			result = append(result, fmt.Sprintf(`style="%s"`, escapeAttr(attr.Style)))
		}

		if attr.OnClick != "" {
			result = append(result, fmt.Sprintf(`onclick="%s"`, escapeAttr(attr.OnClick)))
		}

		if attr.OnChange != "" {
			result = append(result, fmt.Sprintf(`onchange="%s"`, escapeAttr(attr.OnChange)))
		}

		if attr.OnSubmit != "" {
			result = append(result, fmt.Sprintf(`onsubmit="%s"`, escapeAttr(attr.OnSubmit)))
		}

		if attr.Value != "" {
			result = append(result, fmt.Sprintf(`value="%s"`, escapeAttr(attr.Value)))
		}

		if attr.Checked != "" {
			result = append(result, fmt.Sprintf(`checked="%s"`, escapeAttr(attr.Checked)))
		}

		if attr.Selected != "" {
			result = append(result, fmt.Sprintf(`selected="%s"`, escapeAttr(attr.Selected)))
		}

		if attr.Name != "" {
			result = append(result, fmt.Sprintf(`name="%s"`, escapeAttr(attr.Name)))
		}

		if attr.Placeholder != "" {
			result = append(result, fmt.Sprintf(`placeholder="%s"`, escapeAttr(attr.Placeholder)))
		}

		if attr.Autocomplete != "" {
			result = append(result, fmt.Sprintf(`autocomplete="%s"`, escapeAttr(attr.Autocomplete)))
		}

		if attr.Pattern != "" {
			result = append(result, fmt.Sprintf(`pattern="%s"`, escapeAttr(attr.Pattern)))
		}

		if attr.Cols != 0 {
			result = append(result, fmt.Sprintf(`cols="%d"`, attr.Cols))
		}

		if attr.Width != 0 {
			result = append(result, fmt.Sprintf(`width="%d"`, attr.Width))
		}

		if attr.Height != 0 {
			result = append(result, fmt.Sprintf(`height="%d"`, attr.Height))
		}

		if attr.Width != 0 {
			result = append(result, fmt.Sprintf(`width="%d"`, attr.Width))
		}

		if attr.Height != 0 {
			result = append(result, fmt.Sprintf(`height="%d"`, attr.Height))
		}

		if attr.Rows != 0 {
			result = append(result, fmt.Sprintf(`rows="%d"`, attr.Rows))
		}

		if attr.Step != "" {
			result = append(result, fmt.Sprintf(`step="%s"`, escapeAttr(attr.Step)))
		}

		if attr.Min != "" {
			result = append(result, fmt.Sprintf(`min="%s"`, escapeAttr(attr.Min)))
		}

		if attr.Max != "" {
			result = append(result, fmt.Sprintf(`max="%s"`, escapeAttr(attr.Max)))
		}

		if attr.Target != "" {
			result = append(result, fmt.Sprintf(`target="%s"`, escapeAttr(attr.Target)))
		}

		if attr.Required {
			result = append(result, `required="required"`)
		}

		if attr.Disabled {
			result = append(result, `disabled="disabled"`)
		}

		if attr.Readonly {
			result = append(result, `readonly="readonly"`)
		}
	}

	return strings.Join(result, " ")
}

func open(tag string) func(class string, attr ...Attr) func(elements ...string) string {
	return func(class string, attr ...Attr) func(elements ...string) string {
		return func(elements ...string) string {
			attr = append(attr, Attr{Class: class})
			return fmt.Sprintf(`<%s %s>%s</%s>`, tag, attributes(attr...), strings.Join(elements, " "), tag)
		}
	}
}

func closed(tag string) func(class string, attr ...Attr) string {
	return func(class string, attr ...Attr) string {
		attr = append(attr, Attr{Class: class})
		return fmt.Sprintf(`<%s %s/>`, tag, attributes(attr...))
	}
}

var (
	I        = open("i")
	A        = open("a")
	P        = open("p")
	Div      = open("div")
	Span     = open("span")
	Form     = open("form")
	Textarea = open("textarea")
	Select   = open("select")
	Option   = open("option")
	List     = open("ul")
	ListItem = open("li")
	Canvas   = open("canvas")

	Img   = closed("img")
	Input = closed("input")

	Flex1 = Div("flex-1")()

	Space = "&nbsp;"
	W35   = Attr{Style: "max-width: 35rem;"}
	W30   = Attr{Style: "max-width: 30rem;"}
	W25   = Attr{Style: "max-width: 25rem;"}
	W20   = Attr{Style: "max-width: 20rem;"}
)

func mdToHTML(md []byte) string {
	var html bytes.Buffer
	if err := goldmark.Convert(md, &html); err != nil {
		log.Fatal(err)
	}

	return html.String()
}

var Markdown = func(css string) func(elements ...string) string {
	return func(elements ...string) string {
		md := []byte(strings.Join(elements, " "))
		md = bytes.ReplaceAll(md, []byte("\t"), []byte(""))
		html := mdToHTML(md)

		return fmt.Sprintf(`<div class="markdown %s">%s</div>`, css, html)
	}
}

var Script = func(value ...string) string {
    return Trim(fmt.Sprintf(`<script>%s</script>`, strings.Join(value, " ")))
}

var Target = func() Attr {
    return Attr{ID: "i" + RandomString(15)}
}

// Convenience helpers on targets to produce swap descriptors compatible with ctx.PatchTo.
func (a Attr) Replace() TargetSwap { return TargetSwap{ID: a.ID, Swap: OUTLINE} }
func (a Attr) Render() TargetSwap  { return TargetSwap{ID: a.ID, Swap: INLINE} }
func (a Attr) Append() TargetSwap  { return TargetSwap{ID: a.ID, Swap: APPEND} }
func (a Attr) Prepend() TargetSwap { return TargetSwap{ID: a.ID, Swap: PREPEND} }

// ThemeSwitcher renders a small button that cycles System → Light → Dark.
// It relies on the global setTheme(mode) provided by the server (__theme script).
func ThemeSwitcher(css string) string {
	id := "tsui_theme_" + RandomString(8)
	sun := `<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6.76 4.84l-1.8-1.79-1.41 1.41 1.79 1.8 1.42-1.42zm10.48 14.32l1.79 1.8 1.41-1.41-1.8-1.79-1.4 1.4zM12 4V1h-0 0 0 0v3zm0 19v-3h0 0 0 0v3zM4 12H1v0 0 0 0h3zm19 0h-3v0 0 0 0h3zM6.76 19.16l-1.79 1.8 1.41 1.41 1.8-1.79-1.42-1.42zM19.16 6.76l1.8-1.79-1.41-1.41-1.8 1.79 1.41 1.41zM12 8a4 4 0 100 8 4 4 0 000-8z"/></svg>`
	moon := `<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/></svg>`
	desktop := `<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24"><path d="M3 4h18v12H3z"/><path d="M8 20h8v-2H8z"/></svg>`

	btn := Div("")(
		fmt.Sprintf(`<button id="%s" type="button" class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-gray-300 bg-white text-gray-700 hover:bg-gray-100 dark:bg-gray-800 dark:text-gray-200 dark:border-gray-600 dark:hover:bg-gray-700 shadow-sm %s">`, escapeAttr(id), escapeAttr(css)),
		`<span class="icon">`+desktop+`</span>`,
		`<span class="label">Auto</span>`,
		`</button>`,
	)

	// Escape quotes for JavaScript strings
	moonJS := strings.ReplaceAll(moon, `"`, `\"`)
	sunJS := strings.ReplaceAll(sun, `"`, `\"`)
	desktopJS := strings.ReplaceAll(desktop, `"`, `\"`)

	script := Script(
		Trim(fmt.Sprintf(`(function(){
            var btn=document.getElementById('%s'); if(!btn) return;
            var modes=['system','light','dark'];
            function getPref(){ try { return localStorage.getItem('theme')||'system'; } catch(_) { return 'system'; } }
            function resolve(mode){ if(mode==='system'){ try { return (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches)?'dark':'light'; } catch(_) { return 'light'; } } return mode; }
            function setMode(mode){ try { if (typeof setTheme==='function') setTheme(mode); } catch(_){} }
            function labelFor(mode){ return mode==='system'?'Auto':(mode.charAt(0).toUpperCase()+mode.slice(1)); }
            function iconFor(eff){ if(eff==='dark'){ return "%s"; } if(eff==='light'){ return "%s"; } return "%s"; }
            function render(){ var pref=getPref(); var eff=resolve(pref); var i=btn.querySelector('.icon'); if(i){ i.innerHTML=iconFor(eff); } var l=btn.querySelector('.label'); if(l){ l.textContent=labelFor(pref); } }
            render();
            btn.addEventListener('click', function(){ var pref=getPref(); var idx=modes.indexOf(pref); var next=modes[(idx+1)%%modes.length]; setMode(next); render(); });
            try { if (window.matchMedia){ window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(){ if(getPref()==='system'){ render(); } }); } } catch(_){ }
        })();`, id, moonJS, sunJS, desktopJS)),
	)

	return btn + script
}

// Skeletons
//
// Provides simple placeholder UIs similar to t-sui, usable while loading data.
// These do not require WebSockets; you can render them directly or swap later.

type Skeleton string

const (
    SkeletonList      Skeleton = "list"
    SkeletonComponent Skeleton = "component"
    SkeletonPage      Skeleton = "page"
    SkeletonForm      Skeleton = "form"
)

// Skeleton renders a skeleton for the given target.
// If kind is not provided or unknown, renders the Default variant
// (three text lines), matching the TS implementation.
func (a Attr) Skeleton(kind ...Skeleton) string {
    if len(kind) == 0 || string(kind[0]) == "" {
        return a.SkeletonDefault()
    }
    switch kind[0] {
    case SkeletonList:
        return a.SkeletonList(5)
    case SkeletonComponent:
        return a.SkeletonComponent()
    case SkeletonPage:
        return a.SkeletonPage()
    case SkeletonForm:
        return a.SkeletonForm()
    default:
        return a.SkeletonDefault()
    }
}

// SkeletonDefault renders three generic text lines.
func (a Attr) SkeletonDefault() string {
    return Div("animate-pulse", a)(
        Div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow")(
            Div("bg-gray-200 h-5 rounded w-5/6 mb-2")(),
            Div("bg-gray-200 h-5 rounded w-2/3 mb-2")(),
            Div("bg-gray-200 h-5 rounded w-4/6")(),
        ),
    )
}

// SkeletonList renders a vertical list of generic list items (avatar + text).
func (a Attr) SkeletonList(count int) string {
    if count <= 0 { count = 5 }
    items := make([]string, 0, count)
    for i := 0; i < count; i++ {
        row := Div("flex items-center gap-3 mb-3")(
            Div("bg-gray-200 rounded-full h-10 w-10")(),
            Div("flex-1")(
                Div("bg-gray-200 h-4 rounded w-5/6 mb-2")(),
                Div("bg-gray-200 h-4 rounded w-3/6")(),
            ),
        )
        items = append(items, row)
    }
    return Div("animate-pulse", a)(
        Div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow")(strings.Join(items, "")),
    )
}

// SkeletonComponent renders a component-sized content block.
func (a Attr) SkeletonComponent() string {
    return Div("animate-pulse", a)(
        Div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow")(
            Div("bg-gray-200 h-6 rounded w-2/5 mb-4")(),
            Div("bg-gray-200 h-4 rounded w-full mb-2")(),
            Div("bg-gray-200 h-4 rounded w-5/6 mb-2")(),
            Div("bg-gray-200 h-4 rounded w-4/6")(),
        ),
    )
}

// SkeletonPage renders a larger page-level skeleton with header and two cards.
func (a Attr) SkeletonPage() string {
    card := func() string {
        return Div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow mb-4")(
            Div("bg-gray-200 h-5 rounded w-2/5 mb-3")(),
            Div("bg-gray-200 h-4 rounded w-full mb-2")(),
            Div("bg-gray-200 h-4 rounded w-5/6 mb-2")(),
            Div("bg-gray-200 h-4 rounded w-4/6")(),
        )
    }
    return Div("animate-pulse", a)(
        Div("bg-gray-200 h-8 rounded w-1/3 mb-6")(),
        card(),
        card(),
    )
}

// SkeletonForm renders a form-shaped skeleton: labels, inputs, actions.
func (a Attr) SkeletonForm() string {
    fieldShort := func() string {
        return Div("")(
            Div("bg-gray-200 h-4 rounded w-3/6 mb-2")(),
            Div("bg-gray-200 h-10 rounded w-full")(),
        )
    }
    fieldArea := func() string {
        return Div("")(
            Div("bg-gray-200 h-4 rounded w-2/6 mb-2")(),
            Div("bg-gray-200 h-24 rounded w-full")(),
        )
    }
    actions := func() string {
        return Div("flex justify-end gap-3 mt-6")(
            Div("bg-gray-200 h-10 rounded w-24")(),
            Div("bg-gray-200 h-10 rounded w-32")(),
        )
    }
    return Div("animate-pulse", a)(
        Div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow")(
            Div("bg-gray-200 h-6 rounded w-2/5 mb-5")(),
            Div("grid grid-cols-1 md:grid-cols-2 gap-4")(
                Div("")(fieldShort()),
                Div("")(fieldShort()),
                Div("")(fieldArea()),
                Div("")(fieldShort()),
            ),
            actions(),
        ),
    )
}

// Convenience globals (without binding to a specific target id)
func SkeletonDefault() string        { return Attr{}.SkeletonDefault() }
func SkeletonListN(count int) string { return Attr{}.SkeletonList(count) }
func SkeletonComponentBlock() string { return Attr{}.SkeletonComponent() }
func SkeletonPageBlock() string      { return Attr{}.SkeletonPage() }
func SkeletonFormBlock() string      { return Attr{}.SkeletonForm() }

func Variable[T any](getter func(*T) string) func(item *T) Attr {
	temp := Target()

	return func(item *T) Attr {
		value := getter(item)
		return Attr{ID: temp.ID + "_" + value}
	}
}

var ID = func(target string) Attr {
	return Attr{ID: target}
}

var Href = func(value string, target ...string) Attr {
	if len(target) > 0 {
		return Attr{Href: value, Target: target[0]}
	}

	return Attr{Href: value}
}

var Src = func(alt string, target string) Attr {
	return Attr{Src: target, Alt: alt}
}

var Title = func(value string) Attr {
	return Attr{Title: string(value)}
}

var Classes = func(values ...string) string {
	return Trim(strings.Join(values, " "))
}

// type Hook struct {
// 	Target string
// }

// func (h *Hook) Id() Attr {
// 	if len(h.Target) == 0 {
// 		h.Target = "i" + RandomString(15)
// 	}

// 	return Attr{Id: h.Target}
// }

// func (h *Hook) Submit(ctx *Context, action Action) Attr {
// 	action.Type = FORM

// 	if action.Target.Id == "" {
// 		action.Target = Attr{Id: h.Target}
// 	}

// 	return Attr{OnSubmit: ctx.Post(action)}
// }

// func (h *Hook) Click(ctx *Context, action Action) Attr {
// 	if action.Target.Id == "" {
// 		action.Target = Attr{Id: h.Target}
// 	}

// 	return Attr{OnClick: ctx.Post(action)}
// }

type Action struct {
	Method *Callable
	Target Attr
	Values []any
	// Type   ActionType
}

// func SendForm(Target *Attr, Method ComponentMethod, Values ...any) Action {
// 	return Action{
// 		Type:   FORM,
// 		Target: *Target,
// 		Method: Method,
// 		Values: Values,
// 	}
// }

var (
	re  = regexp.MustCompile(`\s{4,}`)
	re2 = regexp.MustCompile(`[\t\n]+`)
	re3 = regexp.MustCompile(`"`)
    // Remove comments before flattening to a single line
    reCommentBlock = regexp.MustCompile(`(?s)/\*.*?\*/`)
    reHtmlComment  = regexp.MustCompile(`(?s)<!--.*?-->`)
    reLineComment  = regexp.MustCompile(`(?m)^[ \t]*//.*$`)
)

func Trim(s string) string {
	// Strip JS/TS block comments and HTML comments first
	s = reCommentBlock.ReplaceAllString(s, "")
	s = reHtmlComment.ReplaceAllString(s, "")
	// Strip full-line // comments (avoid breaking inline URLs like http://)
	s = reLineComment.ReplaceAllString(s, "")
	// Collapse newlines/tabs, then long spaces
	s = re2.ReplaceAllString(s, "")
	return re.ReplaceAllString(s, " ")
}

func Normalize(s string) string {
	return re.ReplaceAllString(re2.ReplaceAllString(re3.ReplaceAllString(s, "&quot;"), ""), " ")
	// return re3.ReplaceAllString(s, "&quot;")
}

func If(cond bool, value func() string) string {
	if cond {
		return value()
	}

	return ""
}

func Iff(cond bool) func(value ...string) string {
	return func(value ...string) string {
		if cond {
			return strings.Join(value, " ")
		}

		return ""
	}
}

func Or(cond bool, value func() string, other func() string) string {
	if cond {
		return value()
	}

	return other()
}

// func Orr(cond bool) (func(...string) string, func(...string) string) {
// 	yes := func(value ...string) string {
// 		if cond {
// 			return strings.Join(value, " ")
// 		}

// 		return ""
// 	}

// 	no := func(value ...string) string {
// 		if !cond {
// 			return strings.Join(value, " ")
// 		}

// 		return ""
// 	}

// 	return yes, no
// }

func Map[T any](values []T, iter func(*T, int) string) string {
	var result []string
	for key, value := range values {
		result = append(result, iter(&value, key))
	}

	return strings.Join(result, " ")
}

func Map2[T any](values []T, iter func(T, int) []string) string {
	var result []string
	for key, value := range values {
		result = append(result, iter(value, key)...)
	}

	return strings.Join(result, " ")
}

func For[T any](from, to int, iter func(int) string) string {
	var result []string
	for i := from; i < to; i++ {
		result = append(result, iter(i))
	}

	return strings.Join(result, " ")
}

func Error(err error) string {
	stackErr := errors.WithStack(err)
	fmt.Printf("%+v\n", stackErr)

	return Div("bg-red-500 text-white font-bold p-8 text-center border border-red-800")(
		"Opps, something went wrong",
	)
}

func ErrorField(err validator.FieldError) string {
	if err == nil {
		return ""
	}

	return Div("text-red-600 p-2 rounded bg-white")(
		// Span("bg-red-800 text-white rounded-full")(err.Field()),
		"obsahuje nevalidnú hodnotu",
	)
}

func ErrorForm(errs *error, translations *map[string]string) string {
	if errs == nil || *errs == nil {
		return ""
	}

	temp := (*errs).(validator.ValidationErrors)

	// for _, err := range errs { fmt.Printf("%+v\n", err.Field()) }

	return Div("text-red-600 p-4 rounded text-center border-4 border-red-600 bg-white")(
		Map(temp, func(err *validator.FieldError, _ int) string {
			trans := (*err).Field()
			invalid := "has invalid value"

			if translations != nil && (*translations)[trans] != "" {
				trans = (*translations)[trans]
			}

			if translations != nil && (*translations)[invalid] != "" {
				invalid = (*translations)[invalid]
			}

			return Div("")(
				Span("font-bold uppercase")(trans),
				Space,
				invalid,
			)
		}),
	)
}

func Print(value any) string {
	return fmt.Sprintf("%+v", value)
}

func PathValue(obj any, path string) (*reflect.Value, error) {
	parts := strings.Split(path, ".")
	current := reflect.ValueOf(obj)

	for _, part := range parts {
		// Handle slice index notation
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			fieldName := part[:strings.Index(part, "[")]
			indexStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
			indexVal, err := strconv.Atoi(indexStr)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			current = current.Elem().FieldByName(fieldName)
			if !current.IsValid() || current.Kind() != reflect.Slice {
				fmt.Printf("invalid slice field: %s", fieldName)
				return nil, err
			}

			for current.Len() <= indexVal {
				elemType := current.Type().Elem()

				if elemType.Kind() == reflect.Pointer {
					newElem := reflect.New(elemType.Elem())
					current.Set(reflect.Append(current, newElem))
				} else {
					newElem := reflect.New(elemType).Elem()
					current.Set(reflect.Append(current, newElem))
				}
			}

			current = current.Index(indexVal)

			// fmt.Printf("current: %v\n", current)
		} else {
			if current.Kind() == reflect.Pointer {
				current = current.Elem()
			}

			current = current.FieldByName(part)
		}

		if !current.IsValid() {
			err := fmt.Errorf("invalid path segment: %s", part)

			fmt.Println(err)

			return nil, err
		}
	}

	return &current, nil
}
