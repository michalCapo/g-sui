# g-sui Documentation

> Server-rendered UI framework for Go. All HTML generation, business logic, and state management occur on the server. Interactivity achieved through WebSocket-delivered JavaScript patches.

**Module:** `github.com/michalCapo/g-sui`
**Go version:** 1.24+
**License:** MIT

---

## Table of Contents

1. [Architecture](#architecture)
2. [Getting Started](#getting-started)
3. [App & Server](#app--server)
4. [Context](#context)
5. [Node (DOM Builder)](#node-dom-builder)
6. [Element Constructors](#element-constructors)
7. [Actions & Events](#actions--events)
8. [DOM Swaps](#dom-swaps)
9. [Conditional Helpers](#conditional-helpers)
10. [Result Effects](#result-effects)
11. [Components](#components)
12. [Form Builder](#form-builder)
13. [Data Tables](#data-tables)
14. [Collate (Data Panel)](#collate-data-panel)
15. [Theme & Dark Mode](#theme--dark-mode)
16. [Localization](#localization)
17. [Page Loading Screen](#page-loading-screen)
18. [Security](#security)
19. [Examples](#examples)
20. [Release](#release)
21. [API Reference](#api-reference)
22. [Server-driven Applications](#server-driven-applications)

---

## Architecture

g-sui compiles Go node trees into **pure JavaScript** strings. The browser receives versioned messages containing JavaScript that performs `document.createElement()` calls directly -- no HTML templates, no client-side framework. SVG elements are created with `document.createElementNS()` using the proper SVG namespace, so inline SVG icons render correctly without workarounds.

```
┌──────────────────────────────────────────┐
│  Server (Go)                             │
│                                          │
│  PageHandler → *Node tree → JS           │
│        ↓                                 │
│  Minimal HTML shell + <script> body      │
│                                          │
│  RegisterAction → Result effects         │
│        ↕                                 │
│  WebSocket (__ws endpoint)               │
└──────────────────────────────────────────┘
         ↕  WS messages (JSON ↑, JS ↓)
┌──────────────────────────────────────────┐
│  Browser                                 │
│                                          │
│  __ws client auto-connects               │
│  Executes JS received from server        │
│  Sends action calls as JSON              │
│  Offline badge + auto-reconnect          │
└──────────────────────────────────────────┘
```

**Key principles:**

1. **Server-centric rendering** -- all DOM trees built in Go
2. **String-based compilation** -- nodes compile to JS, not HTML
3. **Action-based interactivity** -- click/submit events trigger server handlers via WebSocket
4. **Partial updates** -- replace, append, prepend, or innerHTML specific DOM targets
5. **No client framework** -- the client is a ~120-line WS connector script
6. **Tailwind CSS** -- loaded via browser CDN (`@tailwindcss/browser@4`)

---

## Getting Started

### Install

```bash
go get github.com/michalCapo/g-sui
```

### Minimal Application

```go
package main

import r "github.com/michalCapo/g-sui/ui"

func main() {
    app := r.NewApp()

    app.Page("/", func(ctx *r.Context) *r.Node {
        return r.Div("min-h-screen bg-gray-100 p-8").Render(
            r.H1("text-3xl font-bold").Text("Hello World"),
            r.P("text-gray-600 mt-2").Text("g-sui is running."),
        )
    })

    app.Listen(":8080")
}
```

Run and open `http://localhost:8080`.

---

## App & Server

### NewApp

```go
app := ui.NewApp()
```

Creates the application instance. Holds page routes, action handlers, WebSocket clients, and an HTTP mux.

`App.AllowedOrigins` permits additional exact WebSocket origins. Same-origin and non-browser requests without `Origin` are allowed by default; use `"*"` only when intentionally disabling origin validation. Client calls use an internal reply envelope, so the loader clears even for empty action responses; Push and Broadcast messages do not affect it.

#### App Fields

| Field | Type | Description |
|-------|------|-------------|
| `Favicon` | `string` | Path to favicon (adds `<link rel="icon">`) |
| `Title` | `string` | Default document title |
| `Description` | `string` | Meta description tag |
| `HTMLHead` | `[]string` | Additional raw HTML injected into `<head>`; trusted raw API, never pass untrusted input |

### Page Routes

```go
app.Page("/path", func(ctx *ui.Context) *ui.Node {
    return ui.Div("...")
})
```

Registers a GET route using Go's `http.ServeMux` pattern syntax. The handler returns a `*Node` tree that compiles to JS and is served inside the standard HTML shell with Tailwind CSS, Material Icons, the WebSocket client, and the initial loading gate.

Named path wildcards are available through both `ctx.Request.PathValue` and `ctx.PathParams`:

```go
app.Page("/dp/{token}", func(ctx *ui.Context) *ui.Node {
    token := ctx.Request.PathValue("token")
    // Equivalent: token := ctx.PathParams["token"]
    return ui.Div().Text(token)
})

app.Page("/files/{path...}", func(ctx *ui.Context) *ui.Node {
    return ui.Div().Text(ctx.Request.PathValue("path"))
})
```

Static routes remain exact matches, including `/` and routes ending in `/`. Use an explicit `{name...}` wildcard for a subtree. Static routes take precedence over wildcard routes according to `http.ServeMux` matching rules. Path values are also populated during built-in WebSocket navigation.

### Action Handlers

```go
type RenameInput struct { Name string `json:"name"` }
rename := ui.RegisterAction(app, "profile.rename", func(ctx *ui.Context, input RenameInput) (ui.Result, error) {
    return ui.Result{}.SetText("name", input.Name).Toast("Saved"), nil
})
ui.Button().Text("Rename").OnClick(rename.Call(RenameInput{Name: "Alice"}))
```

Actions decode and validate their input before invoking the handler. Return a
`Result` and an error. Use `App.Subscription` for cancellable background updates.

### Layout (Built-in)

```go
app.Layout(func(ctx *ui.Context) *ui.Node {
    return ui.Div("min-h-screen").Render(
        ui.Nav("bg-white shadow").Render(/* nav content */),
        ui.Main("max-w-5xl mx-auto").ID("__content__"),
    )
})
```

Sets a global layout handler. The layout wraps page content for all routes. The layout tree **must** contain exactly one element with `ID("__content__")` — the framework injects the page handler's output there on initial render, and swaps only its innerHTML on browser back/forward navigation.

> **Note:** The `"__content__"` ID is hardcoded in the framework. If you need a custom content ID, use the manual layout pattern below.

### Handler

```go
handler := app.Handler()
http.ListenAndServeTLS(":443", "cert.pem", "key.pem", handler)
```

Returns the `http.Handler` for custom server configurations (TLS, middleware wrapping, etc.).

### App-Level Broadcast

```go
err := app.Broadcast(ui.Result{}.Notify("info", "Server restarting in 5 minutes"))
```

Sends typed effects to all connected clients and returns any errors. Broadcast effects must not depend on a page context.

### Static Assets

```go
//go:embed assets/*
var assets embed.FS

app.Assets(assets, "assets", "/assets/")
app.Favicon = "/assets/favicon.svg"
```

Serves static files from an embedded or on-disk filesystem. The `Favicon` field adds a `<link rel="icon">` tag to the HTML shell.

### CSS (App-Level)

```go
app.CSS(
    []string{"https://fonts.googleapis.com/css2?family=Oswald&display=swap"},
    `body { font-family: 'Oswald', sans-serif; }`,
)
```

Registers external stylesheets and/or inline CSS rules that apply to every page. Tags are injected into the HTML `<head>` server-side, so they load immediately without JavaScript. Pass `nil` for `urls` if you only need inline CSS, or `""` for `css` if you only need external links. This is a trusted raw API; never pass untrusted input.

### HeadCSS (Per-Page via Context)

```go
app.Page("/about", func(ctx *ui.Context) *ui.Node {
    ctx.HeadCSS(
        []string{"https://cdn.example.com/lib.css"},
        `.hero { animation: fadeIn 0.3s ease-out; }
         @keyframes fadeIn { from { opacity:0 } to { opacity:1 } }`,
    )
    return ui.Div("hero").Text("About")
})
```

Registers external stylesheets and/or inline CSS rules for the current page only. On a full page load the tags are injected into the HTML `<head>` server-side (instant, no JS needed). On SPA navigations (WS actions) the same resources are injected into `<head>` via JS with deduplication so external links are not loaded twice. Pass `nil` for `urls` if you only need inline CSS, or `""` for `css` if you only need external links. This is a trusted raw API; never pass untrusted input.

### UnsafeHeadJS (Per-Page via Context)

```go
app.Page("/dashboard", func(ctx *ui.Context) *ui.Node {
    ctx.UnsafeHeadJS(`window.analytics && analytics.page('dashboard');`)
    return ui.Div("").Text("Dashboard")
})
```

Registers a JavaScript block that runs once when the page loads. On a full page load the script is emitted as a `<script>` tag in `<head>`. On SPA navigations the code is prepended to the WS response so it executes before the DOM swap. Prefer [local actions](#local-actions) and [widgets](#widgets-third-party-javascript); for example, a mobile menu is `OnClick(ui.Toggle("mobile-nav"))`. This is a trusted raw API; never pass untrusted input.

**When to use which:**

| Method | Scope | Injection | Deduplication |
|--------|-------|-----------|---------------|
| `app.CSS(urls, css)` | Global (all pages) | Server-side `<head>` | N/A (rendered once) |
| `ctx.HeadCSS(urls, css)` | Per-page | Server-side `<head>` on full load; JS injection on SPA nav | External links deduped by `href` |
| `ctx.UnsafeHeadJS(code)` | Per-page | `<script>` in `<head>` on full load; prepended JS on SPA nav | N/A |

### Listen

```go
app.Listen(":8080")
```

Sets up HTTP handlers (page routes, WebSocket endpoint at `/__ws`, client script at `/__ws.js`) and starts the server.

---

## Context

`Context` carries request data for both page renders (GET) and WS action calls.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Request` | `*http.Request` | The original HTTP request (nil for WS actions) |
| `Session` | `map[string]any` | Connection-local scratch data; resets on reconnect. Not an authentication store. |
| `PathParams` | `map[string]string` | URL path parameters |
| `Query` | `map[string]string` | URL query parameters |

### Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Push` | `(result Result) error` | Sends effects to THIS client immediately |

| `CSS` | `(urls []string, css string)` | Registers per-page CSS (stylesheets and/or inline rules) |
| `UnsafeHeadJS` | `(code string)` | Registers trusted per-page JavaScript for `<head>` |

### Typed Input Example

```go
type ContactInput struct {
    Name string `json:"Name"`
    Email string `json:"Email"`
}
ui.RegisterAction(app, "form.submit", func(ctx *ui.Context, input ContactInput) (ui.Result, error) {
    // Store input.Name and input.Email.
    return ui.Result{}.Toast("Saved!"), nil
})
```

### Push (Real-time Updates)

```go
app.Subscription("clock", func(ctx *ui.Context) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Context().Done():
            return ctx.Context().Err()
        case now := <-ticker.C:
            if err := ctx.Push(ui.Result{}.SetText("clock", now.Format("15:04:05"))); err != nil {
                return err
            }
        }
    }
})
ui.Span().ID("clock").Subscribe("clock")
```

Subscriptions stop on navigation, node removal or disconnect. Reconnect starts a fresh subscription.

### Broadcast

```go
err := app.Broadcast(ui.Result{}.Notify("info", "System maintenance in 5 minutes"))
```

Sends a JS string to every connected WebSocket client.

---

## Node (DOM Builder)

`Node` represents a DOM element built in Go that compiles to JavaScript.

### Creating Nodes

```go
// With class
ui.Div("flex gap-4 items-center")

// Without class
ui.Span()

// Generic element
ui.El("section", "max-w-5xl mx-auto")
```

### Chainable Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `ID` | `(id string) *Node` | Sets element ID |
| `Class` | `(cls string) *Node` | Appends CSS classes |
| `Text` | `(t string) *Node` | Sets textContent |
| `Attr` | `(key, val string) *Node` | Sets an HTML attribute |
| `Style` | `(key, val string) *Node` | Sets an inline style property |
| `Render` | `(children ...*Node) *Node` | Appends child nodes (nil children skipped) |
| `OnClick` | `(action *Action) *Node` | Attaches click event |
| `OnSubmit` | `(action *Action) *Node` | Attaches submit event |
| `On` | `(event string, action *Action) *Node` | Attaches any named event |
| `JS` | `(raw string) *Node` | Raw JS executed after mount (`this` refers to the element) |

### Composing Trees

```go
ui.Div("p-6").Render(
    ui.H1("text-2xl font-bold").Text("Dashboard"),
    ui.Div("grid grid-cols-3 gap-4").Render(
        card("Users", "1,234"),
        card("Revenue", "$56K"),
        card("Orders", "890"),
    ),
)
```

---

## Element Constructors

### Standard Elements

`Div`, `Span`, `Button`, `H1`-`H6`, `P`, `A`, `Nav`, `Main`, `Header`, `Footer`, `Section`, `Article`, `Aside`, `Form`, `Pre`, `Code`, `Ul`, `Ol`, `Li`, `Label`, `Textarea`, `Select`, `Option`

### SVG Elements

`SVG` creates the root `<svg>` element. All children of an SVG root are automatically created with `document.createElementNS('http://www.w3.org/2000/svg', tag)` -- no manual namespace handling needed. Use `El(tag)` for SVG child elements:

```go
// Inline SVG icon -- all elements use the correct SVG namespace
ui.SVG("w-6 h-6").
    Attr("viewBox", "0 0 24 24").Attr("fill", "none").
    Attr("stroke", "currentColor").Attr("stroke-width", "2").
    Render(
        ui.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10"),
        ui.El("path").Attr("d", "M9 12l2 2 4-4"),
    )
```

Supported SVG child tags (all created with `createElementNS` when inside an SVG root):

`g`, `path`, `circle`, `ellipse`, `line`, `polyline`, `polygon`, `rect`, `text`, `tspan`, `defs`, `symbol`, `use`, `image`, `clipPath`, `mask`, `pattern`, `linearGradient`, `radialGradient`, `stop`, `filter`, `marker`, `title`, `desc`, `foreignObject`, `animate`, `animateMotion`, `animateTransform`, `set`, `textPath`

SVG elements use `setAttribute('class', ...)` instead of `.className` for compatibility with the SVG DOM.

### Table Elements

`Table`, `Thead`, `Tbody`, `Tfoot`, `Tr`, `Th`, `Td`, `Caption`, `Colgroup`

### Media / Embed

`Video`, `Audio`, `Canvas`, `Iframe`, `Object`, `Picture`

### Inline Text

`Strong`, `Em`, `Small`, `B`, `I`, `U`, `Sub`, `Sup`, `Mark`, `Abbr`, `Time`

### Block Content

`Blockquote`, `Figure`, `Figcaption`, `Dl`, `Dt`, `Dd`

### Forms (Extended)

`Fieldset`, `Legend`, `Optgroup`, `Datalist`, `Output`, `Progress`, `Meter`

### Interactive

`Details`, `Summary`, `Dialog`

### Void Elements (Self-Closing)

`Input`, `Img`, `Br`, `Hr`, `Source`, `Embed`, `Col`, `Wbr`, `Link`, `Meta`

### Typed Input Constructors

Shorthand for `Input().Attr("type", "...")`:

| Function | HTML Type |
|----------|-----------|
| `IText` | text |
| `IPassword` | password |
| `IEmail` | email |
| `IPhone` | tel |
| `INumber` | number |
| `ISearch` | search |
| `IUrl` | url |
| `IDate` | date |
| `ITime` | time |
| `IDatetime` | datetime-local |
| `IFile` | file |
| `ICheckbox` | checkbox |
| `IRadio` | radio |
| `IRange` | range |
| `IColor` | color |
| `IHidden` | hidden |
| `ISubmit` | submit |
| `IReset` | reset |
| `IArea` | textarea (alias) |

All accept an optional class string: `ui.IText("w-full border rounded px-3 py-2")`

---

## Actions & Events

### Server Actions (WebSocket)

```go
refresh := ui.RegisterAction(app, "counter.refresh", func(ctx *ui.Context, _ struct{}) (ui.Result, error) {
    return ui.Refresh("counter"), nil
})
ui.Button().Text("Refresh").OnClick(refresh.Call(struct{}{}))
```

Use `App.Live` for per-tab counter state; see [live views](#live-views).

### Actions with Data

```go
type ViewInput struct{ ID int }

var viewInvoice ui.ActionRef[ViewInput]

func Register(app *ui.App) {
    viewInvoice = ui.RegisterAction(app, "invoice.view", handleView)
}

ui.Button("...").OnClick(viewInvoice.Call(ViewInput{ID: invoice.ID}))
```

Register actions before pages that use them. Keep the returned `ActionRef` in a
package variable and call it from render code.

### Actions with Collect (Form Values)

```go
ui.Button("...").OnClick(search.Call(SearchInput{}).Collect("search-input", "filter-select"))
```

`Collect` reads `.value` from DOM elements by ID and sends them with the action
call. The values are decoded into the action input by field name or JSON tag.

### Local Actions

Local actions change the page in the browser without a server round trip. Use
them for menus, tabs, dialogs, toggles, copy buttons and other UI state that
the server does not need to know about.

```go
ui.Button("...").Text("Menu").OnClick(ui.Toggle("menu"))
ui.Div("hidden ...").ID("menu").OnOutsideClick(ui.Hide("menu")).Render(...)

ui.Button("...").Text("Delete").OnClick(ui.Confirm("Delete item?", deleteItem.Call(DeleteInput{ID: id})))
ui.Button("...").Text("Copy").OnClick(ui.CopyFrom("api-key"))
ui.Input("...").OnKey("Enter", search.Call(SearchInput{}).Collect("q")).OnKey("Escape", ui.SetValue("q", ""))
ui.Div().Shortcut("mod+k", ui.Focus("q"))
```

| Function | Description |
| --- | --- |
| `Show(id)` / `Hide(id)` / `Toggle(id)` | Change visibility (`hidden` attribute and class). The trigger gets `aria-expanded` and `aria-controls` |
| `AddClass(id, cls)` / `RemoveClass(id, cls)` / `ToggleClass(id, cls)` | Change space-separated classes |
| `SetAttr(id, name, value)` / `RemoveAttr(id, name)` / `ToggleAttr(id, name)` | Change attributes |
| `SetText(id, text)` | Set text content |
| `SetValue(id, value)` | Set an input value and fire `input` and `change` |
| `Remove(id)` | Remove an element and run its cleanup |
| `Focus(id)` | Focus an element and select its text |
| `ScrollTo(id)` / `ScrollTop()` | Smooth scroll |
| `OpenDialog(id)` / `CloseDialog(id)` | Open or close a `<dialog>` as a modal; other elements are shown or hidden |
| `CopyText(text)` / `CopyFrom(id)` | Copy to the clipboard and show a toast |
| `Notify(variant, message)` / `Toast(message)` | Show a notification |
| `Navigate(path)` / `PatchURL(path, replace)` | Live navigation |
| `Redirect(url)` / `Reload()` / `Back()` / `Print()` | Browser navigation |
| `ResetForm(id)` / `SubmitForm(id)` | Reset or submit a form |
| `TogglePassword(id)` | Show or hide a password |
| `ToggleTheme()` / `SetTheme(mode)` | Change the theme (`light`, `dark`, `system`) |
| `SetTitle(title)` | Set the document title |

Compose actions:

| Function | Description |
| --- | --- |
| `Seq(actions...)` | Run actions in order; nil actions are skipped |
| `Confirm(message, action)` | Ask with a native confirm dialog, then run the action |
| `Delay(d, action)` | Run the action after a duration |
| `action.Collect(ids...)` | Send element values with the server calls in the action |

Server calls and local steps mix freely, for example
`ui.Seq(ui.Hide("menu"), save.Call(input))`. A click with a server call
prevents the default browser action. A local click does not, so `NavLink`
still navigates.

### Node Behaviors

| Method | Description |
| --- | --- |
| `OnClick(a)` / `OnSubmit(a)` / `OnChange(a)` / `On(event, a)` | Run an action on an event |
| `OnInput(a, delay...)` | Run an action on input; server calls are debounced (200 ms by default) |
| `OnKey(key, a)` | Run an action on a key press while focused. Keys: `Enter`, `Escape`, `ctrl+s`, `mod+k` (Ctrl or Cmd), `shift+/`, `space` |
| `Shortcut(key, a)` | Page-wide key press while the node is mounted. Plain keys are ignored while typing |
| `OnOutsideClick(a)` | Run an action on a click outside this visible node. Clicks on its toggle trigger are ignored |
| `DragToScroll()` | Drag with the mouse to scroll horizontally |
| `ActiveClass(active, inactive...)` | Classes for a link to the current page; sets `aria-current="page"` |
| `ActivePrefix()` | Also mark the link active on nested paths |

```go
ui.NavLink("/users", "px-3 py-1 rounded").
    ActiveClass("bg-blue-100 text-blue-700", "text-gray-700").
    ActivePrefix().
    Text("Users")
```

### Widgets (Third-Party JavaScript)

Use a widget for a JavaScript library such as a chart or an editor. The mount
function gets the element and the JSON props, and may return a cleanup
function. Morphs keep the widget's DOM.

```go
app.Widget("chart", `function(el, props) {
    var chart = new Chart(el, props);
    return function() { chart.destroy(); };
}`)

ui.Widget("chart", ChartConfig{Type: "bar", Data: data}, "h-64")
```

### Raw JavaScript

Prefer local actions, node behaviors and widgets. Raw JavaScript is a trusted
last resort and must never contain user input:

| API | Description |
| --- | --- |
| `UnsafeJS(code) *Action` | Raw JS action; `event` and `el` are in scope |
| `node.UnsafeJS(code)` | Raw JS that runs when the node is inserted; `this` is the element |
| `ctx.UnsafeHeadJS(code)` | Per-page JS in `<head>` |

In tests, `node.DebugJS()` returns the JavaScript a node compiles to, without
rendering a page. It is for inspection only; the output format may change.

---

## DOM Swaps

Server actions update the page with `Result` effects. Each node compiles to a
self-executing JavaScript function. If the target element is not found, a
warning is logged and `__ws.notfound` is called (which cancels any active Push
goroutines for that connection).

### SVG Namespace

When compiling, the framework detects SVG elements and emits `document.createElementNS('http://www.w3.org/2000/svg', tag)` instead of `document.createElement(tag)`. The SVG context propagates automatically to all descendants -- any `El("path")`, `El("circle")`, etc. nested inside an `SVG()` root will use the correct namespace. CSS classes on SVG elements are set via `setAttribute('class', ...)` since SVG's `.className` is an `SVGAnimatedString`.

### Example

```go
ui.RegisterAction(app, "item.add", func(ctx *ui.Context, _ struct{}) (ui.Result, error) {
    newItem := ui.Li("py-2").Text("New Item")
    return ui.Result{}.Append("item-list", newItem), nil
})
```

### Notification Variants

```go
ui.Notify("success", "Record saved")
ui.Notify("error", "Something went wrong")
ui.Notify("error-reload", "Connection lost")  // persistent, with Reload button
ui.Notify("info", "Processing...")
```

---

## Conditional Helpers

| Function | Signature | Description |
|----------|-----------|-------------|
| `If` | `(cond bool, node *Node) *Node` | Returns node if true, nil otherwise |
| `Or` | `(cond bool, yes, no *Node) *Node` | Binary conditional |
| `Map` | `[T](items []T, fn func(T, int) *Node) []*Node` | Iterate slice into nodes |

### Examples

```go
// Conditional rendering
ui.Div("...").Render(
    ui.If(user.IsAdmin, ui.Button("...").Text("Admin Panel")),
    ui.Or(loggedIn,
        ui.Span().Text("Welcome back"),
        ui.A("...").Text("Login"),
    ),
)

// List rendering
items := ui.Map(products, func(p Product, i int) *ui.Node {
    return ui.Li("py-2").Text(p.Name)
})
ui.Ul("...").Render(items...)
```

---

## Result Effects

```go
type DeleteInput struct { ID string `json:"id"` }
ui.RegisterAction(app, "invoice.delete", func(ctx *ui.Context, input DeleteInput) (ui.Result, error) {
    // Authorize and delete the invoice first.
    return ui.Result{}.Remove("row-" + input.ID).Toast("Invoice deleted").Navigate("/invoices"), nil
})
```

Effects are immutable: assign the returned value when building a result in a loop.

Combine results built by several helpers with `ui.Merge` or `Add`. Effects run
in the order they were added.

```go
return ui.Merge(refreshTopBar(), renderProjects(ctx), ui.Result{}.Run(ui.CloseDialog("edit"))), nil
// or: topBar.Add(projects, closeDialog)
```

| Method | Description |
| --- | --- |
| `Morph(id, node)` | Update a subtree while preserving drafts and focus |
| `Replace(id, node)` | Replace a subtree and reset its input state |
| `Append(id, node)` / `Prepend(id, node)` | Insert a child |
| `SetText(id, text)` | Update text |
| `Remove(id)` | Remove an element |
| `Toast(message)` / `Notify(variant, message)` | Show a notification |
| `Navigate(url)` / `PatchURL(url, replace)` | Navigate or rerender the current route |
| `Refresh(regions...)` | Rerender named regions |
| `Download(filename, mimeType, base64Data)` | Download generated data |
| `Add(others...)` | Append the effects of other results, in order |

Return the result directly. `Context.Push` accepts a result for subscription
updates. `App.Broadcast` accepts effects that do not require a page context and
returns any build or delivery errors.

---

## Components

### Alert

```go
ui.NewAlert().
    Message("Operation completed successfully.").
    Title("Success").
    Variant("success").         // info, success, warning, error (+ "-outline" suffix)
    Dismissible(true).
    Persist("alert-welcome").   // localStorage key
    AlertClass("mb-4").
    Build()
```

### Badge

```go
ui.NewBadge("Active").
    Color("green").           // gray, red, green, blue, yellow, purple (+ "-outline"/"-soft")
    BadgeSize("md").          // sm, md, lg
    BadgeIcon("check_circle").
    Square().                 // rounded-md instead of pill
    BadgeClass("ml-2").       // additional CSS classes
    Build()

// Dot variant
ui.NewBadge("").Dot().Color("red").Build()
```

### Button (High-Level)

```go
ui.NewButton("Save").
    BtnColor(ui.BtnBlue).     // BtnBlue, BtnRed, BtnGreen, BtnYellow, BtnPurple, BtnGray, BtnWhite
    BtnSize(ui.BtnMD).        // BtnXS, BtnSM, BtnMD, BtnLG, BtnXL
    BtnIcon("save").
    Disabled(false).
    Submit("formID").          // makes type="submit"
    Reset().                   // makes type="reset"
    BtnClass("mt-4").          // additional CSS classes
    OnBtnClick(action).
    Build()

// Outline variants: BtnBlueOutline, BtnRedOutline, BtnGreenOutline
// Link variant:
ui.NewButton("View").Href("/details").Build()
```

### Card

```go
ui.NewCard().
    CardHeader(ui.H3("font-semibold").Text("Title")).
    CardBody(ui.P("text-gray-600").Text("Content here.")).
    CardFooter(ui.Button("...").Text("Action")).
    CardImage("/img/photo.jpg", "Photo").
    CardImageSize("400", "300").       // width, height for CLS prevention
    CardImagePriority(true).           // fetchpriority="high" for LCP
    CardVariant("shadowed").           // shadowed, bordered, flat, glass
    CardHover(true).
    CardCompact(true).
    CardPadding("p-8").                // custom padding
    CardClass("custom-card-class").
    Build()
```

### Accordion

```go
ui.NewAccordion().
    Item("Section 1", content1, true).  // true = open by default
    Item("Section 2", content2).
    Item("Section 3", content3).
    Multiple(false).                     // one at a time
    Variant("bordered").                 // bordered, ghost, separated
    AccordionClass("mb-4").              // additional CSS classes
    Build()
```

### Tabs

```go
ui.NewTabs().
    Tab("Overview", overviewNode, "dashboard").  // optional icon
    Tab("Details", detailsNode).
    Tab("Settings", settingsNode, "settings").
    Active(0).                                    // 0-based index
    TabStyle("underline").                        // underline, pills, boxed, vertical
    TabsClass("mb-6").                            // additional CSS classes
    Build()
```

Includes keyboard navigation (Arrow keys) and ARIA attributes.

### Dropdown

```go
ui.NewDropdown(triggerButton).
    DropdownHeader("Actions").
    DropdownItem("Edit", editAction, "edit").
    DropdownItem("Duplicate", dupAction, "content_copy").
    DropdownDivider().
    DropdownDanger("Delete", deleteAction, "delete").
    DropdownPosition("bottom-left").  // bottom-left, bottom-right, top-left, top-right
    DropdownClass("ml-auto").         // additional CSS classes
    Build()
```

Auto-closes on outside click and Escape key.

### Tooltip

```go
ui.NewTooltip("Helpful hint").
    TooltipPosition("top").     // top, bottom, left, right
    TooltipVariant("dark").     // dark, light, blue, green, red, yellow
    Delay(200).                 // ms, 0 = instant (CSS only)
    TooltipClass("z-50").       // additional CSS classes
    Wrap(targetElement)
```

### Progress Bar

```go
ui.NewProgress().
    ProgressValue(75).
    ProgressColor("bg-blue-600").
    ProgressGradient("#3b82f6", "#8b5cf6").  // overrides solid color
    ProgressSize("md").                       // xs, sm, md, lg, xl
    Striped(true).
    Animated(true).
    Indeterminate(false).
    ProgressLabel("Loading...").
    LabelPosition("outside").                 // inside (lg/xl only), outside
    ProgressClass("mb-4").                    // additional CSS classes
    Build()
```

### Step Progress

```go
ui.NewStepProgress(2, 5).     // current step, total steps
    StepColor("bg-blue-500").
    StepSize("md").
    StepClass("mb-6").
    Locale(&ui.StepProgressLocale{
        StepOf: func(cur, total int) string { return fmt.Sprintf("Krok %d z %d", cur, total) },
    }).
    Build()
```

### Confirm Dialog

```go
ui.ConfirmDialog(
    "Delete Invoice",
    "Are you sure? This cannot be undone.",
    deleteInvoice.Call(DeleteInput{ID: id}), // both buttons close the dialog
    // optional: custom cancel action and/or locale
    ui.ConfirmOpt{
        Locale: &ui.ConfirmLocale{Cancel: "Zrusit", Confirm: "Potvrdit"},
    },
)
```

### Skeleton Loaders

```go
ui.SkeletonTable()       // 4-column table with header + 5 rows
ui.SkeletonCards()       // 6-card responsive grid
ui.SkeletonList()        // 5 rows with avatar + text lines
ui.SkeletonComponent()   // Single card with title + text + button
ui.SkeletonPage()        // Header + sidebar + main content area
ui.SkeletonForm()        // 4 label+input pairs + submit button
```

### Markdown

```go
ui.Markdown("prose dark:prose-invert", markdownContent)
```

Renders markdown to HTML using goldmark. Sets innerHTML after mount. Goldmark's default safe renderer omits raw HTML and unsafe links such as `javascript:` URLs; do not enable unsafe markdown rendering for untrusted input.

### Icon

**Material Icons** (font-based):

```go
ui.Icon("home")                         // Material Icons Round
ui.Icon("settings", "text-lg text-blue-600")
ui.IconText("check_circle", "Verified", "text-green-600")
```

**Inline SVG** (namespace-aware, no font dependency):

```go
// Stroke icon
ui.SVG("w-5 h-5").
    Attr("viewBox", "0 0 24 24").Attr("fill", "none").
    Attr("stroke", "currentColor").Attr("stroke-width", "2").
    Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").
    Render(
        ui.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10"),
        ui.El("path").Attr("d", "M9 12l2 2 4-4"),
    )

// Filled icon
ui.SVG("w-5 h-5 text-pink-500").
    Attr("viewBox", "0 0 24 24").Attr("fill", "currentColor").
    Render(
        ui.El("path").Attr("d", "M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78L12 21.23l8.84-8.84a5.5 5.5 0 000-7.78z"),
    )

// Animated spinner
ui.SVG("w-5 h-5 animate-spin").
    Attr("viewBox", "0 0 24 24").Attr("fill", "none").
    Render(
        ui.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10").
            Attr("stroke", "currentColor").Attr("stroke-width", "3").
            Style("opacity", "0.25"),
        ui.El("path").Attr("d", "M4 12a8 8 0 018-8").
            Attr("stroke", "currentColor").Attr("stroke-width", "3").
            Attr("stroke-linecap", "round"),
    )
```

All SVG child elements (`path`, `circle`, `line`, `polygon`, etc.) inherit the SVG namespace automatically. Tailwind classes like `w-5`, `h-5`, `text-pink-500`, `animate-spin` work on SVG elements via `setAttribute('class', ...)`.

### Theme Switcher

```go
ui.ThemeSwitcher()              // System -> Light -> Dark toggle
ui.ThemeSwitcher("ml-auto")    // with extra classes
```

### reCAPTCHA v3

```go
ui.NewCaptchaV3("your-site-key").
    FormAction("login").
    TokenField("captcha-token").
    Build()
```

---

## Form Builder

Declarative form builder with client-side and server-side validation.

### Creating a Form

```go
form := ui.NewForm("contact-form").
    Action(contactSubmit). // ui.ActionRef from RegisterAction
    Text("Full Name", "Name").Required().Placeholder("John Doe").Render().
    Email("Email Address", "Email").Required().Render().
    Phone("Phone", "Phone").Placeholder("+1 555-0100").Render().
    Area("Message", "Message").Required().Render().
    SelectField("Priority", "Priority").
        Opts(":Select...", "low:Low", "medium:Medium", "high:High").
        Required().Render().
    Radio("Gender", "Gender").
        Opts("male:Male", "female:Female", "other:Other").Render().
    Checkbox("Accept Terms", "Terms").Required().Render().
    Submit("send", "Send Message", "px-4 py-2 bg-blue-600 text-white rounded cursor-pointer")

node := form.Build()
```

### Field Types

| Method | Type | Description |
|--------|------|-------------|
| `Text` | text | Standard text input |
| `Password` | password | Password input |
| `Email` | email | Email input |
| `Number` | number | Numeric input |
| `Phone` | tel | Phone input |
| `DateField` | date | Date picker |
| `TimeField` | time | Time picker |
| `DatetimeField` | datetime-local | Date+time picker |
| `UrlField` | url | URL input |
| `SearchField` | search | Search input |
| `Area` | textarea | Multi-line text |
| `SelectField` | select | Dropdown select |
| `Radio` | radio | Inline radio buttons |
| `RadioBtn` | radio | Button-style radios with borders |
| `RadioCard` | radio | Card-style radios (peer-checked) |
| `Checkbox` | checkbox | Checkbox |
| `Hidden` | hidden | Hidden field |

### Field Configuration (Chaining)

```go
form.Text("Name", "name").
    Required().
    Placeholder("Enter name").
    Value("John").
    PatternValidation(`[A-Za-z ]+`).
    Err("Name must contain only letters").
    IsChecked(true).          // checkbox checked state
    Class("custom-input-class").
    WrapClass("custom-wrapper-class").
    Render()
```

### Select/Radio Options

```go
// Format: "value:Label" or just "Label" (value = lowercase)
.Opts(":Select...", "us:United States", "uk:United Kingdom")
```

### Multiple Submit Buttons

```go
form.
    Submit("save", "Save Draft", "bg-gray-500 text-white px-4 py-2 rounded cursor-pointer").
    Submit("publish", "Publish", "bg-blue-600 text-white px-4 py-2 rounded cursor-pointer")
```

The handler receives `Action` field in data to identify which button was clicked.

### Server-Side Validation

```go
func (input *ContactInput) Validate() error {
    if input.Name == "" {
        return ui.ValidationError{Fields: ui.FormErrors{"Name": "Name is required"}}
    }
    return nil
}
ui.RegisterAction(app, "contact.submit", func(ctx *ui.Context, input ContactInput) (ui.Result, error) {
    return ui.Result{}.Toast("Form submitted!"), nil
})
```

`RegisterAction` calls `Validate` before the handler. `FormFor[ContactInput]`
displays field errors for typed form submissions.

`FormErrors` methods:

Return `form.ShowErrors(errs)` (a `Result`) to fill the error nodes created by `Build`, set `aria-invalid`, and focus the first invalid field after `form.Validate(input)` returns server-side errors.

Pressing Enter in a text input submits its owning form's submit button first; buttons outside the form are considered only when no form exists. Toasts include a dismiss button. Dark mode styles only the document baseline, so explicit Tailwind `dark:` classes remain authoritative.
- `HasErrors() bool` -- true if any field has an error
- `Get(name string) string` -- error message for a field

### Form Configuration

| Method | Description |
|--------|-------------|
| `FormClass(cls)` | Override wrapper div class |
| `InputClass(cls)` | Default CSS for all text inputs |
| `ErrClass(cls)` | CSS for error messages |
| `Action(name)` | WS action name for submit |

---

## Data Tables

### DataTable (Generic)

```go
type Invoice struct {
    ID     int
    Number string
    Amount float64
    Status string
}

table := ui.NewDataTable[Invoice]("invoice-table").
    Action(invoiceData). // ui.ActionRef from RegisterAction
    Head("Number").
    Head("Amount", "text-right").
    Head("Status").
    Head("Actions").
    FieldText(func(inv *Invoice) string { return inv.Number }).
    FieldText(func(inv *Invoice) string { return fmt.Sprintf("$%.2f", inv.Amount) }, "text-right").
    Field(func(inv *Invoice) *ui.Node {
        return ui.NewBadge(inv.Status).Color("green").Build()
    }).
    Field(func(inv *Invoice) *ui.Node {
        return ui.Button("text-sm text-blue-600").Text("View").
            OnClick(viewInvoice.Call(ViewInput{ID: inv.ID}))
    }).
    Sortable(0, 1, 2).
    Sort(0, "asc").
    Page(1).
    PageSize(10).
    TotalItems(42).
    Search("").
    Empty("No invoices found").
    Render(invoices)
```

### Unified Column Definition

The `Col` method provides a single-call column definition combining header, cell renderer, sort, and filter:

```go
table := ui.NewDataTable[Invoice]("invoice-table").
    Action(invoiceData). // ui.ActionRef from RegisterAction
    Col("Number", ui.ColOpt[Invoice]{
        Text:    func(inv *Invoice) *Node { return ui.Span().Text(inv.Number) },
        Sortable: true,
    }).
    Col("Amount", ui.ColOpt[Invoice]{
        Text:     func(inv *Invoice) *Node { return ui.Span().Text(fmt.Sprintf("$%.2f", inv.Amount)) },
        Sortable: true,
        Filter:   ui.NumFilter,
        HeadCls:  "text-right",
        CellCls:  "text-right",
    }).
    Col("Department", ui.ColOpt[Invoice]{
        Text:          func(inv *Invoice) *Node { return ui.Span().Text(inv.Department) },
        Filter:        ui.SelectFilter,
        FilterOptions: []string{"Engineering", "Marketing", "Sales", "HR"},
    }).
    Render(invoices)
```

### Column Filters

Per-column filters are shown as popups triggered from the header. Four filter types are available:

| Type | Constant | Aliases | Description |
|------|----------|---------|-------------|
| `FilterTypeText` | `"text"` | `TxtFilter` | Contains, starts with, equals |
| `FilterTypeDate` | `"date"` | `DateFilter` | Date range (from/to) |
| `FilterTypeNumber` | `"number"` | `NumFilter` | Range, gte, lte, gt, lt, equals |
| `FilterTypeSelect` | `"select"` | `SelectFilter` | Select from predefined options |

Filter operators:

| Operator | Constant | Used By |
|----------|----------|---------|
| `"contains"` | `OpContains` | Text |
| `"startswith"` | `OpStartsWith` | Text |
| `"equals"` | `OpEquals` | Text, Number |
| `"range"` | `OpRange` | Date, Number |
| `"gte"` | `OpGTE` | Number |
| `"lte"` | `OpLTE` | Number |
| `"gt"` | `OpGT` | Number |
| `"lt"` | `OpLT` | Number |

### Expandable Row Detail

```go
table.Detail(func(inv *Invoice) *Node {
    return ui.Div("p-4 bg-gray-50").Render(
        ui.Span("text-sm").Text(fmt.Sprintf("Notes: %s", inv.Notes)),
    )
})
```

Clicking a row toggles an accordion-style detail panel below it.

### DataTable Configuration

| Method | Description |
|--------|-------------|
| `Head(label, cls...)` | Add text header column |
| `HeadHTML(label, cls...)` | Add raw content header |
| `Field(fn, cls...)` | Column with `*Node` content |
| `FieldText(fn, cls...)` | Column with plain text (auto-escaped) |
| `Col(label, ColOpt)` | Unified column definition (header + cell + sort + filter) |
| `Action(name)` | WS action name for all data operations |
| `Sortable(cols...)` | Mark columns as sortable |
| `Detail(fn)` | Expandable row detail renderer |
| `SetFilterValue(col, val)` | Set active filter value for column |
| `SetFilterLabels(badges)` | Set active filter badge labels |
| `Page(page)` | Current page number |
| `PageSize(size)` | Items per page |
| `TotalPages(total)` | Total number of pages |
| `TotalItems(count)` | Total item count |
| `HasMore(bool)` | Whether more items exist (load-more mode) |
| `RowOffset(offset)` | Row offset for alternating stripes |
| `Sort(col, dir)` | Current sort state |
| `Search(val)` | Current search value |
| `Empty(text)` | Text when no rows |
| `DataTableClass(cls)` | Wrapper div class |
| `TableClass(cls)` | `<table>` element class |
| `Locale(loc)` | Per-instance `*TableLocale`; nil = English |
| `Render(data)` | Full table render |
| `RenderRows(data)` | Render rows only (for append) |
| `RenderFooter()` | Render footer only |
| `TbodyID()` | ID of the tbody element |
| `FooterID()` | ID of the footer element |

### SimpleTable

For quick tables without generics or data binding:

```go
ui.NewSimpleTable(3, "w-full").
    SimpleHeader("Name", "Email", "Role").
    CellText("John Doe").
    CellText("john@example.com").
    CellText("Admin").
    Cell(ui.NewBadge("Active").Color("green").Build()).  // *Node cell
    CellText("jane@example.com").
    CellText("User").
    Build()
```

`CellText(text)` adds a plain text cell. `Cell(node)` adds a `*Node` cell for custom content (badges, buttons, etc.). Rows auto-flush when `numCols` is reached.

---

## Collate (Data Panel)

A generic data component with a slide-out filter/sort panel, search bar, load-more pagination, and export. Unlike `DataTable` (inline per-column filters and sort arrows), `Collate` uses a dedicated filter panel with an Apply button. All operations go through a single WS action.

### Creating a Collate

```go
type Employee struct {
    ID         int
    Name       string
    Department string
    Salary     float64
    HireDate   string
    Active     bool
    Role       string
}

collate := ui.NewCollate[Employee]("employees-collate").
    Action(employeesData). // ui.ActionRef from RegisterAction
    Limit(10).
    Sort(
        ui.CollateSortField{Field: "name", Label: "Name"},
        ui.CollateSortField{Field: "department", Label: "Department"},
        ui.CollateSortField{Field: "salary", Label: "Salary"},
        ui.CollateSortField{Field: "hire_date", Label: "Hire Date"},
    ).
    Filter(
        ui.CollateFilterField{Field: "active", Label: "Active Only", Type: ui.CollateBool},
        ui.CollateFilterField{Field: "hire_date", Label: "Hire Date", Type: ui.CollateDateRange},
        ui.CollateFilterField{
            Field: "department",
            Label: "Department",
            Type:  ui.CollateSelect,
            Options: []ui.CollateOption{
                {Value: "Engineering", Label: "Engineering"},
                {Value: "Marketing", Label: "Marketing"},
                {Value: "Sales", Label: "Sales"},
            },
        },
    ).
    Row(func(emp *Employee, idx int) *ui.Node {
        return ui.Div("p-4 border-b").Render(
            ui.Span("font-medium").Text(emp.Name),
            ui.Span("text-sm text-gray-500 ml-2").Text(emp.Department),
        )
    }).
    Detail(func(emp *Employee) *ui.Node {
        return ui.Div("p-4 bg-gray-50").Render(
            ui.Span("text-sm").Text(fmt.Sprintf("Salary: $%.2f", emp.Salary)),
        )
    }).
    Empty("No employees").
    EmptyIcon("group_off").
    Page(1).TotalItems(len(allEmployees)).HasMore(true).
    Render(data)
```

### Filter Types

| Constant | Control | Description |
|----------|---------|-------------|
| `CollateBool` | Checkbox | Boolean toggle (e.g., "Active Only") |
| `CollateDateRange` | Date pickers | From/to date range |
| `CollateSelect` | Dropdown | Single-value select from options |
| `CollateMultiCheck` | Checkboxes | Multiple values from options |

### CollateFilterValue

The action handler receives filter values as `[]CollateFilterValue`:

```go
type CollateFilterValue struct {
    Field string `json:"field"` // filter field name
    Type  string `json:"type"`  // "bool", "date", "select"
    Bool  bool   `json:"bool"`  // for CollateBool
    From  string `json:"from"`  // for CollateDateRange
    To    string `json:"to"`    // for CollateDateRange
    Value string `json:"value"` // for CollateSelect
}
```

### Action Request Format

The WS action receives a JSON payload with these fields:

| Field | Type | Description |
|-------|------|-------------|
| `operation` | string | `"search"`, `"filter"`, `"reset"`, `"loadmore"`, `"export"` |
| `search` | string | Current search query |
| `page` | int | Current page (1-based) |
| `limit` | int | Items per page |
| `order` | string | Sort order, e.g. `"name asc"` or `"salary desc"` |
| `filters` | array | Array of `CollateFilterValue` objects |

### Load More (Append Rows)

```go
// In the action handler, for "loadmore" operation:
dt := newCollateWithState(req.Search, req.Order).
    Page(req.Page).TotalItems(totalItems).HasMore(hasMore).
    RowOffset(start)

resp := ui.Result{}
rows := dt.RenderRows(pageData)
for _, row := range rows {
    resp = resp.Append(dt.BodyID(), row)
}
resp = resp.Replace(dt.FooterID(), dt.RenderFooter())
return resp, nil
```

### Collate Configuration

| Method | Description |
|--------|-------------|
| `NewCollate[T](id)` | Create with unique ID |
| `Action(name)` | WS action name for all operations |
| `Sort(fields...)` | Sortable fields shown in panel |
| `Filter(fields...)` | Filter fields shown in panel |
| `Row(fn)` | Row renderer `func(*T, int) *Node` |
| `Detail(fn)` | Expandable detail `func(*T) *Node` |
| `Limit(n)` | Items per page |
| `Page(p)` | Current page (1-based) |
| `TotalItems(n)` | Total matching items |
| `Search(val)` | Current search query |
| `Order(order)` | Current sort (e.g. `"name asc"`) |
| `HasMore(bool)` | Whether more items exist |
| `SetFilter(field, val)` | Set active filter value |
| `Empty(text)` | Empty state message |
| `EmptyIcon(icon)` | Material icon for empty state |
| `CollateClass(cls)` | Wrapper CSS class |
| `RowOffset(n)` | Row offset for alternating stripes |
| `Locale(loc)` | Per-instance `*CollateLocale`; nil = English |
| `Render(data)` | Full render with data |
| `RenderRows(data)` | Render rows only (for append) |
| `RenderFooter()` | Render footer only |
| `BodyID()` | ID of the row container |
| `FooterID()` | ID of the footer element |

---

## Theme & Dark Mode

g-sui includes built-in dark mode with three states: System, Light, Dark.

### How It Works

1. A synchronous `<head>` script reads `localStorage("theme")` and applies the `dark` class on `<html>` before render
2. The HTML shell keeps the body hidden until the initial DOM, stylesheets, Tailwind-generated CSS, and active fonts are ready to paint
3. CSS overrides in `<style>` provide dark mode fallbacks for common Tailwind classes
4. `ThemeSwitcher` component provides a UI toggle

The readiness gate reveals on the next paint boundary with a 160 ms ease-out fade and dispatches `gsui:ready`. The fade is disabled when the user prefers reduced motion. The gate does not wait for images. A four-second fail-safe prevents a stalled third-party resource from leaving the page blank indefinitely.

### Theme Switcher

```go
ui.ThemeSwitcher()  // Cycles: System -> Light -> Dark

// With locale:
ui.ThemeSwitcher(ui.ThemeSwitcherOpt{
    Locale: &ui.ThemeSwitcherLocale{ThemeAuto: "Auto", ThemeLight: "Svetly", ThemeDark: "Tmavy"},
})
```

### Manual Theme Control

The client exposes two globals:
- `setTheme(mode)` -- "system", "light", or "dark"
- `toggleTheme()` -- toggles between light and dark

From Go, use the `ui.SetTheme(mode)` and `ui.ToggleTheme()` actions.

### Using Dark Mode in Components

Use Tailwind's `dark:` variant:

```go
ui.Div("bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100")
```

---

## Localization

Components ship with English text by default. When no locale is set, English strings are used automatically -- no configuration needed for English-only apps.

To translate a component, create a locale struct with the fields you need and pass it via `.Locale()`. Each component has its own locale type containing only the strings it uses.

### Per-Component Locale Types

| Type | Component | Defined in |
|------|-----------|------------|
| `TableLocale` | `DataTable`, `FilterPopup`, `SimpleTable` | `table.go` |
| `CollateLocale` | `Collate` | `collate.go` |
| `ConfirmLocale` | `ConfirmDialog` | `components.go` |
| `ThemeSwitcherLocale` | `ThemeSwitcher` | `components.go` |
| `StepProgressLocale` | `StepProgress` | `components.go` |
| `FilterLocale` | Embedded by `TableLocale` and `CollateLocale` | `table.go` |

### DataTable

```go
table := ui.NewDataTable[Invoice]("t").
    Locale(&ui.TableLocale{
        FilterLocale: ui.FilterLocale{
            From: "Od", To: "Do",
            Today: "Dnes", ThisWeek: "Tento tyden",
            ThisMonth: "Tento mesic", ThisQuarter: "Tento kvartal",
            ThisYear: "Tento rok", LastMonth: "Minuly mesic", LastYear: "Minuly rok",
        },
        Search:      "Hledat...",
        Apply:       "Pouzit",
        Cancel:      "Zrusit",
        Reset:       "Obnovit",
        Excel:       "Excel",
        LoadMore:    "Nacist dalsi...",
        NoData:      "Zadna data",
        SearchText:  "Hledat text...",
        SelectAll:   "Vybrat vse",
        ClearSelect: "Zrusit vyber",
        Value:       "Hodnota",
        Contains:    "Obsahuje",
        StartsWith:  "Zacina na",
        Equals:      "Rovna se",
        Range:       "Rozsah",
        GreaterOrEq: ">=",
        LessOrEq:    "<=",
        GreaterThan: ">",
        LessThan:    "<",
        NumEquals:   "= Rovna se",
        ItemCount:   func(showing, total int) string {
            return fmt.Sprintf("%d z %d", showing, total)
        },
    }).
    Render(data)
```

### Collate

```go
collate := ui.NewCollate[Employee]("c").
    Locale(&ui.CollateLocale{
        FilterLocale: ui.FilterLocale{
            From: "Od", To: "Do",
            Today: "Dnes", ThisWeek: "Tento tyden",
            ThisMonth: "Tento mesic", ThisQuarter: "Tento kvartal",
            ThisYear: "Tento rok", LastMonth: "Minuly mesic", LastYear: "Minuly rok",
        },
        Search:            "Hledat...",
        Apply:             "Pouzit",
        Reset:             "Obnovit",
        Excel:             "Excel",
        Filter:            "Filtr",
        LoadMore:          "Nacist dalsi...",
        NoData:            "Zadna data",
        AllOption:         "-- Vse --",
        FiltersAndSorting: "Filtry a razeni",
        Filters:           "Filtry",
        SortBy:            "Radit dle",
        ItemCount: func(showing, total int) string {
            return fmt.Sprintf("%d z %d", showing, total)
        },
    }).
    Render(data)
```

### Confirm Dialog

```go
ui.ConfirmDialog("Smazat?", "Opravdu chcete smazat?", deleteAction, ui.ConfirmOpt{
    Locale: &ui.ConfirmLocale{Cancel: "Zrusit", Confirm: "Potvrdit"},
})
```

### Theme Switcher

```go
ui.ThemeSwitcher(ui.ThemeSwitcherOpt{
    Locale: &ui.ThemeSwitcherLocale{
        ThemeAuto: "Automaticky", ThemeLight: "Svetly", ThemeDark: "Tmavy",
    },
})
```

### Step Progress

```go
ui.NewStepProgress(2, 5).
    Locale(&ui.StepProgressLocale{
        StepOf: func(current, total int) string {
            return fmt.Sprintf("Krok %d z %d", current, total)
        },
    }).
    Build()
```

### FilterLocale Fields

`FilterLocale` is embedded in both `TableLocale` and `CollateLocale`. It holds shared date/range labels:

| Field | Default |
|-------|---------|
| `From` | `"From"` |
| `To` | `"To"` |
| `Today` | `"Today"` |
| `ThisWeek` | `"This week"` |
| `ThisMonth` | `"This month"` |
| `ThisQuarter` | `"This quarter"` |
| `ThisYear` | `"This year"` |
| `LastMonth` | `"Last month"` |
| `LastYear` | `"Last year"` |

### TableLocale Fields

| Field | Default |
|-------|---------|
| `Search` | `"Search..."` |
| `Apply` | `"Apply"` |
| `Cancel` | `"Cancel"` |
| `Reset` | `"Reset"` |
| `Excel` | `"Excel"` |
| `LoadMore` | `"Load more..."` |
| `NoData` | `"No data"` |
| `SearchText` | `"Search text..."` |
| `SelectAll` | `"Select all"` |
| `ClearSelect` | `"Clear selection"` |
| `Value` | `"Value"` |
| `Contains` | `"Contains"` |
| `StartsWith` | `"Starts with"` |
| `Equals` | `"Equals"` |
| `Range` | `"Range"` |
| `GreaterOrEq` | `"≥ Greater or equal"` |
| `LessOrEq` | `"≤ Less or equal"` |
| `GreaterThan` | `"> Greater than"` |
| `LessThan` | `"< Less than"` |
| `NumEquals` | `"= Equals"` |
| `ItemCount` | `func(showing, total int) string` -- `"X of Y"` |

### CollateLocale Fields

| Field | Default |
|-------|---------|
| `Search` | `"Search..."` |
| `Apply` | `"Apply"` |
| `Reset` | `"Reset"` |
| `Excel` | `"Excel"` |
| `Filter` | `"Filter"` |
| `LoadMore` | `"Load more..."` |
| `NoData` | `"No data"` |
| `AllOption` | `"— All —"` |
| `FiltersAndSorting` | `"Filters & Sorting"` |
| `Filters` | `"Filters"` |
| `SortBy` | `"Sort by"` |
| `ItemCount` | `func(showing, total int) string` -- `"X of Y"` |

---

---

## Security

### Server-Side

- **JS String Escaping**: All strings embedded in JS are escaped (backslash, single quote, newlines, tabs) via `escJS()`
- **XSS Prevention**: `Text()` uses `textContent` (not `innerHTML`), preventing script injection
- **Safe Table Methods**: `FieldText()` for auto-escaped text, `Field()` for controlled `*Node` content
- **Panic Recovery**: Server panics in action handlers are recovered and surface as error toasts

### Client-Side

- **WebSocket-only**: No form submissions or XHR -- all interaction goes through the WS protocol
- **Auto-reconnect**: Dropped connections are retried with exponential backoff (see [Connection resilience](#connection-resilience))
- **Not-found handling**: Missing DOM targets cancel Push goroutines and notify the server

### Connection resilience

Brief disconnects are invisible. The client keeps the page fully usable while the socket is down, so long-running in-page work (large file uploads, multi-step forms) survives a hiccup.

| Behavior | Detail |
| --- | --- |
| Keep-alive ping | Every `App.KeepAliveMs` (default 25000, `-1` disables). Prevents proxies and load balancers from dropping idle sockets, the usual cause of "offline" during a slow upload. |
| Offline badge | Appears only after the outage outlives `App.OfflineGraceMs` (default 3000, `-1` never shows). It is a small non-blocking badge -- it never dims the page or swallows clicks. |
| Reconnect reload | The page reloads after reconnect only when the outage lasted at least `App.ReconnectReloadAfterMs` (default 15000, `-1` never reloads) **and** no hold is active. |
| Holds | `window.__ws.hold()` registers critical work and returns a release function. While any hold is active a reconnect never reloads the page. |
| Fast retry | Reconnect is attempted immediately on the browser `online` event and when a hidden tab becomes visible again, instead of waiting out the backoff. |
| Offline calls | Actions are rejected while disconnected and never replayed. Retry explicitly after reconnect. |
| Subscriptions | `window.__ws.subscribe(act, data)` re-sends the call on every reconnect, so server-side `Push` loops are re-armed. Identical registrations are deduplicated, and the call is never queued (that would deliver it twice on reconnect). `__ws.unsubscribe(act)` drops one. |
| Server restart | Every connection starts with the server announcing its instance id. Element ids (`ui.Target()`) are random per process, so a page rendered by an earlier process can never be patched by a new one: the client detects the change and reloads, deferring the reload while a hold is active. Set `App.InstanceID` to a shared build id when running multiple replicas, otherwise reconnecting to a different replica reloads too. |
| Subscription lifetime | Live navigation cancels the old page's subscriptions. `Node.Subscribe` also cancels on removal. A missing target cancels only its sending subscription. See [server-driven apps](#server-driven-applications). |

Use `App.Subscription` and `Node.Subscribe` for background updates.

```go
app := ui.NewApp()
app.OfflineGraceMs = 5000          // tolerate 5s blips silently
app.ReconnectReloadAfterMs = -1    // never reload after reconnect
```

Set these before serving traffic. They are plain exported fields, so assigning them while requests are in flight is a data race; the values are read once per page render.

Wrap in-page work that must not be interrupted:

```js
const release = window.__ws.hold();
try {
  await uploadLargeFile(file);   // reconnects will not reload the page
} finally {
  release();
}
```

After a reconnect that did not reload, the client dispatches `gsui:reconnected` with `detail.downMs` so the page can resync itself:

```js
window.addEventListener('gsui:reconnected', e => {
  if (e.detail.downMs > 60000) __ws.call('refreshData', {});
});
```

Other client helpers: `__ws.connected()`, `__ws.offline()`, `__ws.holds()`, `__ws.reconnect()`. All of these -- including `hold()` -- are available on the pre-client stub, so they are safe to call from `Node.UnsafeJS` blocks and `ctx.UnsafeHeadJS` that run before `/__ws.js` loads.

A hold defers a reload rather than cancelling it: when the last hold is released, the pending reload runs. The client dispatches `gsui:reloadpending` with `detail.reason` (`"server-restart"` or `"long-outage"`) at the moment the reload is postponed, so the page can warn the user or finish up. Always release holds in a `finally` block -- a leaked hold postpones the reload for the lifetime of the page.

When the server process changes (a restart or a deploy), the client dispatches `gsui:serverchanged` with `detail.was`/`detail.now` before reloading:

```js
window.addEventListener('gsui:serverchanged', () => showToast('Updating…'));
```

With `ReconnectReloadAfterMs = -1` no reload ever happens, including after a restart. The page then stays connected but unpatchable -- responses target ids the new process never generated -- so handle `gsui:serverchanged` yourself if you disable reloads.

---

## Examples

The `example/` directory contains a full working application demonstrating all features:

```bash
make example
# Open http://localhost:1424
```

### Example Pages

| Page | Description |
|------|-------------|
| `/` | Component showcase (alerts, badges, cards, tabs, accordion, dropdowns, tooltips, progress) |
| `/counter` | Stateful counter with increment/decrement via WebSocket |
| `/hello` | Action responses: success, error, delayed, panic recovery |
| `/clock` | Live clock using a managed subscription |
| `/form` | FormBuilder with validation, multiple submit buttons |
| `/login` | Login form with server-side validation |
| `/shared` | Reusable form template pattern |
| `/invoices` | Full CRUD invoice management |
| `/routes` | Route parameters and query parameters |
| `/skeleton` | All skeleton loader variants |
| `/reload-redirect` | Client-side navigation and redirects |
| `/append` | Append/prepend DOM operations |
| `/button`, `/text`, `/password`, `/number`, `/date`, `/area`, `/select`, `/checkbox`, `/radio` | Individual input demos |
| `/icons` | Material Icons showcase |
| `/table` | Table component demo |
| `/collate` | Collate data panel with filter/sort, search, load-more, expandable detail |
| `/others` | Miscellaneous component demos |
| `/actions` | Local actions: menu, shortcuts, copy, dialog, theme, widget |

---

## Release

### Release Command

`make release` creates and pushes version tags:

```bash
make release
```

- Versioning format: `v1.MINOR.PATCH` (e.g., `v1.1.0`, `v1.1.1`)
- Increments the patch of the latest `v1.*` tag; starts at `v1.1.0`
- Ensures clean working tree before tagging
- Runs `make tidy`
- Creates annotated git tag and pushes to remote

### Using as a Dependency

```bash
go get github.com/michalCapo/g-sui@v1.1.0
```

---

## API Reference

### Package `ui`

#### Types

| Type | Description |
|------|-------------|
| `Node` | DOM element that compiles to JavaScript |
| `Action` | What a node does on an event: server calls and local UI steps |
| `ActionRef[T]` | Typed server action returned by `RegisterAction`; `Call(data)` builds an `Action` |
| `App` | Application container (routes, actions, WS clients) |
| `LayoutHandler` | `func(ctx *Context) *Node` |
| `PageHandler` | `func(ctx *Context) *Node` |
| `Context` | Request data for pages and WS actions |
| `FormBuilder` | Declarative form builder |
| `FieldBuilder` | Single field configuration |
| `Field` | Field definition struct |
| `FieldType` | Field type enum |
| `FieldOption` | Value/label pair for select/radio |
| `FormErrors` | `map[string]string` of validation errors |
| `DataTable[T]` | Generic configurable table |
| `ColOpt[T]` | Unified column definition for `DataTable` |
| `FilterType` | Column filter type (`"text"`, `"date"`, `"number"`, `"select"`) |
| `FilterOperator` | Filter operator (`"contains"`, `"equals"`, `"range"`, etc.) |
| `FilterValue` | Active filter value with operator and value(s) |
| `ColumnFilter` | Column filter configuration |
| `FilterBadge` | Active filter badge display |
| `SimpleTable` | Non-generic quick table |
| `Collate[T]` | Generic data panel with filter/sort panel |
| `CollateSortField` | Sort field definition for Collate |
| `CollateFilterType` | Filter control type for Collate |
| `CollateFilterField` | Filter field definition for Collate |
| `CollateOption` | Value/label pair for Collate filters |
| `CollateFilterValue` | Active filter value for Collate |
| `AlertBuilder` | Alert component builder |
| `BadgeBuilder` | Badge component builder |
| `ButtonBuilder` | High-level button builder |
| `CardBuilder` | Card component builder |
| `AccordionBuilder` | Accordion component builder |
| `TabsBuilder` | Tabs component builder |
| `DropdownBuilder` | Dropdown menu builder |
| `ProgressBuilder` | Progress bar builder |
| `StepProgressBuilder` | Step progress builder |
| `TooltipBuilder` | Tooltip builder |
| `CaptchaV3Builder` | reCAPTCHA v3 builder |
| `FilterLocale` | Shared date/range filter strings (embedded by `TableLocale`, `CollateLocale`) |
| `TableLocale` | Locale strings for `DataTable` and `FilterPopup` |
| `CollateLocale` | Locale strings for `Collate` |
| `ConfirmLocale` | Locale strings for `ConfirmDialog` |
| `ThemeSwitcherLocale` | Locale strings for `ThemeSwitcher` |
| `StepProgressLocale` | Locale strings for `StepProgress` |

#### Constants

**Button Colors:** `BtnBlue`, `BtnRed`, `BtnGreen`, `BtnYellow`, `BtnPurple`, `BtnGray`, `BtnWhite`, `BtnBlueOutline`, `BtnRedOutline`, `BtnGreenOutline`

**Button Sizes:** `BtnXS`, `BtnSM`, `BtnMD`, `BtnLG`, `BtnXL`

**Field Types:** `FieldText`, `FieldPassword`, `FieldEmail`, `FieldNumber`, `FieldPhone`, `FieldDate`, `FieldTime`, `FieldDatetime`, `FieldUrl`, `FieldSearch`, `FieldTextarea`, `FieldSelect`, `FieldRadio`, `FieldRadioBtn`, `FieldRadioCard`, `FieldCheckbox`, `FieldHidden`

**Radio Styles:** `RadioInline`, `RadioButton`, `RadioCard`

**Filter Types:** `FilterTypeText` (`TxtFilter`), `FilterTypeDate` (`DateFilter`), `FilterTypeNumber` (`NumFilter`), `FilterTypeSelect` (`SelectFilter`)

**Filter Operators:** `OpContains`, `OpStartsWith`, `OpEquals`, `OpRange`, `OpGTE`, `OpLTE`, `OpGT`, `OpLT`

**Collate Filter Types:** `CollateBool`, `CollateDateRange`, `CollateSelect`, `CollateMultiCheck`

#### App Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Page` | `(pattern string, handler PageHandler)` | Register GET page route using `http.ServeMux` patterns and path values |
| `Layout` | `(handler LayoutHandler)` | Set global layout (uses `__content__` ID) |
| `CSS` | `(urls []string, css string)` | Global stylesheets/inline CSS in `<head>` |
| `GET` | `(path string, handler http.HandlerFunc)` | Register HTTP GET handler |
| `POST` | `(path string, handler http.HandlerFunc)` | Register HTTP POST handler |
| `DELETE` | `(path string, handler http.HandlerFunc)` | Register HTTP DELETE handler |
| `Assets` | `(fsys fs.FS, dir, prefix string)` | Serve static files |
| `Handler` | `() http.Handler` | Returns mux for custom server setup |
| `Listen` | `(addr string) error` | Start HTTP server |
| `Broadcast` | `(result Result) error` | Send effects to all connected clients |
| `Widget` | `(name, mount string)` | Register a client widget mount function |
| `Subscription` | `(name string, run func(*Context) error)` | Register cancellable background work started by `node.Subscribe` |

#### Context Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Push` | `(result Result) error` | Sends effects to THIS client immediately |

| `CSS` | `(urls []string, css string)` | Per-page CSS; `<head>` on full load, JS injection on SPA nav (links deduped) |
| `UnsafeHeadJS` | `(code string)` | Trusted per-page JS; `<script>` in `<head>` on full load, prepended JS on SPA nav |

#### Global Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `NewApp()` | `*App` | Create application |
| `El(tag, class...)` | `*Node` | Create element |
| `Target()` | `string` | Generate random DOM ID |
| `RegisterSubscription[T](app, name, run)` | | Like `app.Subscription`; decodes `node.Subscribe` data into `T` |

| `UnsafeJS(code)` | `*Action` | Trusted raw JS action |
| `If(cond, node)` | `*Node` | Conditional render |
| `Or(cond, yes, no)` | `*Node` | Binary conditional |
| `Map[T](items, fn)` | `[]*Node` | Slice iteration |
| `Merge(results...)` | `Result` | Combine results in order; same as `Result{}.Add(results...)` |
| `Seq`, `Confirm`, `Delay` | `*Action` | Compose actions; see [Local Actions](#local-actions) |
| `Show`, `Hide`, `Toggle`, `Remove`, `SetText`, ... | `*Action` | Local UI actions; see [Local Actions](#local-actions) |
| `Widget(name, props, class...)` | `*Node` | Mount a widget registered with `App.Widget` |
| `NewForm(id)` | `*FormBuilder` | Form builder |
| `NewDataTable[T](id)` | `*DataTable[T]` | Generic table |
| `FilterPopup(col, label, type, opts, val)` | `*Node` | Standalone filter popup |
| `NewSimpleTable(cols, cls...)` | `*SimpleTable` | Quick table |
| `NewCollate[T](id)` | `*Collate[T]` | Collate data panel |
| `NewAlert()` | `*AlertBuilder` | Alert builder |
| `NewBadge(text)` | `*BadgeBuilder` | Badge builder |
| `NewButton(label)` | `*ButtonBuilder` | Button builder |
| `NewCard()` | `*CardBuilder` | Card builder |
| `NewAccordion()` | `*AccordionBuilder` | Accordion builder |
| `NewTabs()` | `*TabsBuilder` | Tabs builder |
| `NewDropdown(trigger)` | `*DropdownBuilder` | Dropdown builder |
| `NewProgress()` | `*ProgressBuilder` | Progress bar builder |
| `NewStepProgress(cur, total)` | `*StepProgressBuilder` | Step progress builder |
| `NewTooltip(content)` | `*TooltipBuilder` | Tooltip builder |
| `NewCaptchaV3(siteKey)` | `*CaptchaV3Builder` | reCAPTCHA builder |
| `ConfirmDialog(...)` | `*Node` | Confirmation dialog |
| `Markdown(class, content)` | `*Node` | Markdown renderer |
| `Icon(name, class...)` | `*Node` | Material icon |
| `IconText(icon, text, class...)` | `*Node` | Icon + text |
| `ThemeSwitcher(class...)` | `*Node` | Theme toggle |
| `SkeletonTable()` | `*Node` | Table skeleton |
| `SkeletonCards()` | `*Node` | Cards skeleton |
| `SkeletonList()` | `*Node` | List skeleton |
| `SkeletonComponent()` | `*Node` | Component skeleton |
| `SkeletonPage()` | `*Node` | Page skeleton |
| `SkeletonForm()` | `*Node` | Form skeleton |

---

## Server-driven applications

The Go server owns application state and renders nodes. The browser runtime
handles navigation, keyed DOM updates, forms, focus and subscriptions. Tailwind
loading, theme configuration and the existing Node/JavaScript renderer are unchanged.

Run the component showcase:

```sh
make example
## http://127.0.0.1:1424
```

### Pages and navigation

Register each page once. Use the built-in layout for a persistent application shell:

```go
app := ui.NewApp()
app.Layout(func(ctx *ui.Context) *ui.Node {
    return ui.Div().Render(
        ui.Nav().Render(ui.NavLink("/projects").Text("Projects")),
        ui.Main().ID("__content__"),
    )
})
app.Page("/projects", projectsPage)
```

`NavLink` produces a real anchor. Normal clicks perform live navigation;
modified clicks, downloads, external links and `target="_blank"` retain browser
behavior. Ordinary `A().Attr("href", ...)` links still use HTTP navigation.

| Operation | Behavior |
| --- | --- |
| `ui.Navigate(url)` / `Result.Navigate(url)` | Load a page, cancel its predecessor, then update history and focus. |
| `ui.PatchURL(url, replace)` | Rerender the current route with new URL parameters, preserving its live view and scroll. |
| `ui.Redirect(url)` | Full HTTP navigation, including cookies and HTTP middleware responses. |

On `NavLink`, `Attr("data-gsui-patch", "")` selects PatchURL behavior and
`Attr("data-gsui-replace", "")` replaces the history entry. A patch to another
route falls back to HTTP. Unknown, denied and redirecting live destinations also
fall back to HTTP so the address, status and visible page agree.

`Node.Title("Projects")` updates the document title. `App.Title` remains the
initial HTML title. Back/Forward rerenders the destination and restores its saved
scroll position. Normal navigation focuses a heading (or the content container).

### Authorization and identity

Install route middleware through `app.Use` **before** `Handler` or `Listen`:

```go
app.Use(requireSession) // func(http.Handler) http.Handler
app.Identity = currentUser // func(*http.Request) (any, error)
app.Authorize = func(ctx *ui.Context, action string) error {
    return checkAccess(ctx.User(), ctx.Request.URL.Path, action)
}
```

`Use` applies to initial pages and resolves the destination for live navigation,
region refresh and page actions. Middleware-added request context values survive
route resolution. `Authorize` runs with an empty action for page rendering and
the action name for user events. `Identity` is resolved again for every action:
validate expiry/revocation in your session store, not only at WebSocket upgrade.

Wrapping `app.Handler()` externally cannot make path-based middleware run again
inside a WebSocket. Move that protection to `Use`/`Authorize`. Actions are globally
registered: authorizing a public page alone does not authorize an unrelated action.
Check action and object permissions on the server. Payloads, route parameters,
DOM IDs and client page versions are not authentication credentials.

An open WebSocket has its handshake cookies. Login, logout or cookie changes
should use an HTTP endpoint and full redirect. Middleware must not depend on
setting a cookie during a live action.

### Typed actions and region refresh

```go
type RenameProject struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

rename := ui.RegisterAction(app, "project.rename",
    func(ctx *ui.Context, input RenameProject) (ui.Result, error) {
        if err := projects.Rename(ctx.Context(), ctx.User(), input); err != nil {
            return ui.Result{}, err
        }
        return ui.Refresh("project").Toast("Project renamed"), nil
    },
)

// In a page:
ui.Region("project", renderProject(project))
ui.Button().Text("Rename").OnClick(rename.Call(RenameProject{ID: "42", Name: "New name"}))
```

`Refresh` reruns the current page renderer and morphs just the named regions.
Keep rendering free of mutations and use stable region IDs. The zero `Result`
means success without a DOM update. Other effects include `Morph`, `Remove`,
`Navigate` and `PatchURL`. Errors produce a generic user message and a server log.
Handlers return `(Result, error)`; raw string handlers are not supported.

Typed actions and live events reject new calls while disconnected. Requests
already sent are not automatically replayed; a dropped connection can leave an
unknown outcome. For payments or other non-repeatable operations, implement
transaction IDs/idempotency in application storage. All actions use the same delivery policy; offline mutations are never queued.

### Typed forms

```go
func (input *RenameProject) Validate() error {
    if input.Name == "" {
        return ui.ValidationError{Fields: ui.FormErrors{"name": "Name is required"}}
    }
    return nil
}

ui.FormFor[RenameProject]("rename").Render(
    ui.IHidden().Attr("name", "id").Attr("value", project.ID),
    ui.Label().Attr("for", "name").Text("Name"),
    ui.IText().ID("name").Attr("name", "name").Attr("required", ""),
    ui.Button().Attr("type", "submit").Text("Save"),
).Submit(rename)
```

Input names match JSON fields. Text is sent unchanged, numbers become numbers,
checkboxes become booleans and multi-selects become string slices. Empty number
inputs send `null`; use pointers for optional numbers. File uploads require an
HTTP upload endpoint. Repeated text field names are not a slice binding API.

Native constraints run before sending. `RegisterAction` invokes `Validate() error`
on the input pointer when implemented. `ValidationError` displays field messages,
sets ARIA error state and focuses the first invalid field. Errors do not replace
the form or discard its values. Busy state belongs to the submitting control,
and another request's reply does not release it.

### Server-owned views

```go
app.Live("/counter", func() ui.View { return &Counter{} })

type Counter struct { Count int }

func (v *Counter) Render(ctx *ui.ViewContext) *ui.Node {
    return ui.Div().Render(
        ui.Span().Text(strconv.Itoa(v.Count)),
        ui.Button().Text("+1").OnClick(ctx.Event("increment")),
    )
}

func (v *Counter) Handle(ctx *ui.ViewContext, event ui.Event) error {
    if event.Name == "increment" { v.Count++ }
    return nil
}
```

Views are per connection/tab. Events run serially, then the framework rerenders
and morphs the view. `ctx.Event(name, payload)` carries optional data;
`event.Decode(&input)` decodes it. Validate event data in the handler.

Optional `Mount(*ViewContext) error` runs once on the connected view. The initial
HTTP render uses a temporary view without Mount, so Render must tolerate its zero
state. A reconnect creates a fresh view and calls Mount again. Durable state and
draft recovery belong in application storage; in-memory view state is not a
durable session. PatchURL preserves the connected view; navigation remounts it.

`Context.Session` is connection-local scratch data, shared by that connection's
serial actions and retained across page navigation. It resets on reconnect; the
HTTP render has a separate empty map. It is not an authentication/session store.
Do not access it or mutate a view from background goroutines without your own
synchronization. Use subscriptions to push independent updates instead.

### Keyed updates and local interaction

`Result.Morph` and typed refreshes preserve elements by `Key` or ID. Keys must be
unique among siblings. Dirty inputs keep their values; active inputs keep their
selection. Add `Attr("data-gsui-reset", "")` to explicitly replace an input value.
`Preserve()` leaves an external widget and its descendants under browser ownership.

`Replace` remains a destructive swap. `Node.UnsafeJS`, widgets and `Subscribe` setup
run only on newly mounted nodes during morphs. To recreate a widget or change a
subscription's captured parameters, change its key or explicitly replace it.

[Local actions](#local-actions) such as `Toggle(id)` and `OpenDialog(id)` are
authored in Go. `node.OnInput(action, 250*time.Millisecond)` debounces server
calls and cancels its timer on removal; `Action.Collect` supplies the field values. Ordinary server
events use the same request/reply API.

### Typed tables

```go
table := ui.RegisterTable(app, "projects",
    func(ctx *ui.Context, q ui.TableQuery) (ui.TablePage[Project], error) {
        return projects.Load(ctx.Context(), ctx.User(), q)
    },
    func(t *ui.DataTable[Project]) {
        t.PageSize(20).Col("Name", ui.ColOpt[Project]{
            Sortable: true,
            Text: func(p *Project) *ui.Node { return ui.Span().Text(p.Name) },
        })
    },
)

// Within a page handler:
node, err := table.Render(ctx)
```

The loader receives typed search, sorting and filter state. Return the first
`q.Limit()` matching rows and the total count; this follows DataTable's load-more
interface. Page size is bounded to 100 and loaded pages to 20. Search replaces
the URL entry; sort/filter/load-more add one. Query keys are namespaced by table
ID, so Back/Forward and copied URLs restore the table without global filter maps.
Only configured sort/filter columns reach the loader. Validate filter operators
and values against your domain and use parameterized database queries.
Configure `RowKey(func(*T) string)` with a record ID when rows contain inputs or
widgets, so sorting keeps browser state attached to the correct record.

This adapter omits export buttons; use the existing DataTable API for custom
export workflows. `Collate` retains its current lower-level API.

### Subscriptions and shutdown

```go
app.Subscription("clock", func(ctx *ui.Context) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Context().Done():
            return ctx.Context().Err()
        case now := <-ticker.C:
            if err := ctx.Push(ui.Result{}.SetText("clock", now.Format(time.TimeOnly))); err != nil {
                return err
            }
        }
    }
})

ui.Span().ID("clock").Subscribe("clock")
```

Read `Subscribe` data with `RegisterSubscription`. Data must be a JSON object.
Like `RegisterAction`, it calls an optional `Validate() error` on `*T`:

```go
type FolderFeed struct {
    Folder string `json:"folder"`
}

ui.RegisterSubscription(app, "folder.feed", func(ctx *ui.Context, in FolderFeed) error {
    // Watch in.Folder until ctx.Context() is done.
    return nil
})

ui.Div().ID("list").Subscribe("folder.feed", FolderFeed{Folder: "inbox"})
```

Subscriptions start after their response is sent, cancel on removal/navigation/
disconnect, and restart on reconnect. Name plus data identifies a subscription;
use different data for independent instances. Removing one does not cancel other
subscriptions. A missing target only cancels its sending subscription.

Pass `ctx.Context()` to blocking work and stop on cancellation. Reads continue
while an action runs so disconnect can cancel it. Actions remain serial per
connection: keep them short and use subscriptions/background jobs for streams.
The inbound message limit is 1 MiB; writes have a ten-second deadline.

Call `app.Close()` alongside `http.Server.Shutdown()` to close upgraded sockets.
The example uses `app.Listen` for simple startup. For graceful shutdown, use a
custom HTTP server with `app.Handler()`. Large broadcasts still use
the existing broadcast API; tenant-scoped fan-out belongs in the application.

### Verification

Run `make` for all commands; see [development commands](../README.md#development-commands)
for tools, environment variables, and cleanup behavior. In Libro, start the
example through its application controls.

```sh
make test-race
make example
## In another terminal, with Playwright installed:
make test-browser
```

The browser check covers typed forms, table search/exports, menu routes,
navigation/history, per-tab views, keyed DOM changes, drafts and subscription reconnect. Set `GSUI_URL` for a
different example address and `GSUI_BROWSER` for an installed Chromium executable.
