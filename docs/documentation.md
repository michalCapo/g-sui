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
8. [JS Compilation & DOM Swaps](#js-compilation--dom-swaps)
9. [JS Helper Functions](#js-helper-functions)
10. [Conditional Helpers](#conditional-helpers)
11. [Response Builder](#response-builder)
12. [Components](#components)
13. [Form Builder](#form-builder)
14. [Data Tables](#data-tables)
15. [Theme & Dark Mode](#theme--dark-mode)
16. [Security](#security)
17. [Examples](#examples)
18. [Deployment](#deployment)
19. [API Reference](#api-reference)

---

## Architecture

g-sui compiles Go node trees into **pure JavaScript** strings. The browser receives raw JS that performs `document.createElement()` calls directly -- no HTML templates, no JSON intermediate format, no client-side framework.

```
┌──────────────────────────────────────────┐
│  Server (Go)                             │
│                                          │
│  PageHandler → *Node tree → .ToJS()      │
│        ↓                                 │
│  Minimal HTML shell + <script> body      │
│                                          │
│  ActionHandler → JS string (DOM patch)   │
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
│  Offline overlay + auto-reconnect        │
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

### Page Routes

```go
app.Page("/path", func(ctx *ui.Context) *ui.Node {
    return ui.Div("...")
})
```

Registers a GET route. The handler returns a `*Node` tree that compiles to JS and is served inside a minimal HTML shell with Tailwind CSS, Material Icons, and the WebSocket client.

### Action Handlers

```go
app.Action("counter.inc", func(ctx *ui.Context) string {
    count++
    return ui.NewResponse().
        Replace("counter", ui.Span().ID("counter").Text(fmt.Sprintf("%d", count))).
        Build()
})
```

Registers a named server action callable via WebSocket. The handler receives a `Context` and returns a raw JS string that the client executes.

### Static Assets

```go
//go:embed assets/*
var assets embed.FS

app.Assets(assets, "assets", "/assets/")
app.Favicon = "/assets/favicon.svg"
```

Serves static files from an embedded or on-disk filesystem. The `Favicon` field adds a `<link rel="icon">` tag to the HTML shell.

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
| `Session` | `map[string]any` | Session data store |
| `PathParams` | `map[string]string` | URL path parameters |
| `Query` | `map[string]string` | URL query parameters |

### Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `WsData` | `() map[string]any` | Returns raw WebSocket data map |
| `Body` | `(target any) error` | Unmarshals WS data into a struct |
| `Push` | `(js string) error` | Sends JS to THIS client immediately |
| `Broadcast` | `(js string)` | Sends JS to ALL connected clients |

### Body Example

```go
app.Action("form.submit", func(ctx *ui.Context) string {
    var data struct {
        Name  string `json:"Name"`
        Email string `json:"Email"`
    }
    ctx.Body(&data)
    // use data.Name, data.Email
    return ui.Notify("success", "Saved!")
})
```

### Push (Real-time Updates)

```go
app.Action("clock.start", func(ctx *ui.Context) string {
    go func() {
        for {
            time.Sleep(time.Second)
            err := ctx.Push(ui.SetText("clock", time.Now().Format("15:04:05")))
            if err != nil {
                return // client navigated away
            }
        }
    }()
    return ""
})
```

`Push` returns an error when the client navigates away or the connection drops, allowing goroutines to clean up.

### Broadcast

```go
ctx.Broadcast(ui.Notify("info", "System maintenance in 5 minutes"))
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
| `JS` | `(raw string) *Node` | Raw JS executed after mount |

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
// Define action
app.Action("counter.inc", func(ctx *ui.Context) string {
    count++
    return ui.Span().ID("count").Text(fmt.Sprintf("%d", count)).ToJSReplace("count")
})

// Attach to element
ui.Button("...").OnClick(&ui.Action{Name: "counter.inc"})
```

### Actions with Data

```go
ui.Button("...").OnClick(&ui.Action{
    Name: "invoice.view",
    Data: map[string]any{"id": invoice.ID},
})
```

### Actions with Collect (Form Values)

```go
ui.Button("...").OnClick(&ui.Action{
    Name:    "search.run",
    Collect: []string{"search-input", "filter-select"},
})
```

`Collect` reads `.value` from DOM elements by ID and sends them with the action call.

### Client-Side Actions

```go
// Raw JS instead of WS call
ui.Button("...").OnClick(ui.JS("history.back()"))
ui.Button("...").OnClick(ui.JS("alert('Hello!')"))
```

---

## JS Compilation & DOM Swaps

Every `*Node` compiles to JavaScript. Five compilation strategies exist:

| Method | Description |
|--------|-------------|
| `ToJS()` | Appends node to `document.body` |
| `ToJSReplace(targetID)` | Replaces element with matching ID |
| `ToJSAppend(parentID)` | Appends as child of parent element |
| `ToJSPrepend(parentID)` | Prepends as first child |
| `ToJSInner(targetID)` | Replaces innerHTML of target |

All methods produce self-executing IIFEs. If the target element is not found, a warning is logged and `__ws.notfound` is called (which cancels any active Push goroutines for that connection).

### Example

```go
app.Action("item.add", func(ctx *ui.Context) string {
    newItem := ui.Li("py-2").Text("New Item")
    return newItem.ToJSAppend("item-list")
})
```

---

## JS Helper Functions

These return JS strings for common DOM operations. Use them in action handlers.

| Function | Signature | Description |
|----------|-----------|-------------|
| `Notify` | `(variant, message string) string` | Toast notification (success/error/error-reload/info) |
| `Redirect` | `(url string) string` | Full page navigation (`window.location.href`) |
| `SetLocation` | `(url string) string` | URL update without reload (`history.pushState`) |
| `Back` | `() *Action` | Browser back (`history.back()`) |
| `SetTitle` | `(title string) string` | Update document title |
| `RemoveEl` | `(id string) string` | Remove element by ID |
| `SetText` | `(id, text string) string` | Set textContent by ID |
| `SetAttr` | `(id, attr, value string) string` | Set attribute by ID |
| `AddClass` | `(id, cls string) string` | Add CSS class by ID |
| `RemoveClass` | `(id, cls string) string` | Remove CSS class by ID |
| `Show` | `(id string) string` | Remove `hidden` class |
| `Hide` | `(id string) string` | Add `hidden` class |
| `Download` | `(filename, mimeType, base64Data string) string` | Trigger file download |
| `DragToScroll` | `(id string) string` | Enable drag-to-scroll on element |

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

## Response Builder

For action handlers that need multiple operations in a single response:

```go
app.Action("invoice.delete", func(ctx *ui.Context) string {
    return ui.NewResponse().
        Remove("row-" + id).
        Toast("success", "Invoice deleted").
        Navigate("/invoices").
        Build()
})
```

### Methods

| Method | Description |
|--------|-------------|
| `Add(js string)` | Append raw JS |
| `Replace(targetID, node)` | Replace element |
| `Inner(targetID, node)` | Replace innerHTML |
| `Append(parentID, node)` | Append child |
| `Remove(id)` | Remove element |
| `Toast(variant, message)` | Show notification |
| `Navigate(url)` | Update URL (pushState) |
| `Back()` | Browser back |
| `Build() string` | Join all parts into JS string |

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
    CardVariant("shadowed").   // shadowed, bordered, flat, glass
    CardHover(true).
    CardCompact(true).
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
    Build()
```

Auto-closes on outside click and Escape key.

### Tooltip

```go
ui.NewTooltip("Helpful hint").
    TooltipPosition("top").     // top, bottom, left, right
    TooltipVariant("dark").     // dark, light, blue, green, red, yellow
    Delay(200).                 // ms, 0 = instant (CSS only)
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
    Build()
```

### Step Progress

```go
ui.NewStepProgress(2, 5).     // current step, total steps
    StepColor("bg-blue-500").
    StepSize("md").
    Build()
```

### Confirm Dialog

```go
ui.ConfirmDialog(
    "Delete Invoice",
    "Are you sure? This cannot be undone.",
    &ui.Action{Name: "invoice.delete", Data: map[string]any{"id": id}},
    // optional cancel action (defaults to removing the dialog)
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

Renders markdown to HTML using goldmark. Uses `.JS()` to set innerHTML after mount.

### Icon

```go
ui.Icon("home")                         // Material Icons Round
ui.Icon("settings", "text-lg text-blue-600")
ui.IconText("check_circle", "Verified", "text-green-600")
```

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
    Action("contact.submit").
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
app.Action("contact.submit", func(ctx *ui.Context) string {
    errs := form.Validate(ctx.WsData())
    if errs.HasErrors() {
        // re-render form with errors
        return renderFormWithErrors(errs)
    }

    var data ContactForm
    ctx.Body(&data)
    // process data...

    return ui.Notify("success", "Form submitted!")
})
```

`FormErrors` methods:
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
            OnClick(&ui.Action{Name: "invoice.view", Data: map[string]any{"id": inv.ID}})
    }).
    Searchable("invoice.search").
    Sortable(0, 1, 2).
    SortAction("invoice.sort").
    Paginated("invoice.page", 1, 10).
    TotalItems(42).
    Export("invoice.export").
    Sort(0, "asc").
    Search("").
    Empty("No invoices found").
    Render(invoices)
```

### DataTable Configuration

| Method | Description |
|--------|-------------|
| `Head(label, cls...)` | Add text header column |
| `HeadHTML(label, cls...)` | Add raw content header |
| `Field(fn, cls...)` | Column with `*Node` content |
| `FieldText(fn, cls...)` | Column with plain text (auto-escaped) |
| `Searchable(action)` | Enable search with debounced input |
| `Sortable(cols...)` | Mark columns as sortable |
| `SortAction(action)` | WS action for sort clicks |
| `Paginated(action, page, totalPages)` | Enable pagination |
| `TotalItems(count)` | Show item count in pagination |
| `Export(action)` | Enable export button |
| `Sort(col, dir)` | Current sort state |
| `Search(val)` | Current search value |
| `Empty(text)` | Text when no rows |
| `DataTableClass(cls)` | Wrapper div class |
| `TableClass(cls)` | `<table>` element class |

### SimpleTable

For quick tables without generics or data binding:

```go
ui.NewSimpleTable(3, "w-full").
    SimpleHeader("Name", "Email", "Role").
    CellText("John Doe").
    CellText("john@example.com").
    CellText("Admin").
    CellText("Jane Smith").
    CellText("jane@example.com").
    CellText("User").
    Build()
```

---

## Theme & Dark Mode

g-sui includes built-in dark mode with three states: System, Light, Dark.

### How It Works

1. A synchronous `<head>` script reads `localStorage("theme")` and applies the `dark` class on `<html>` before render (prevents FOUC)
2. CSS overrides in `<style>` provide dark mode fallbacks for common Tailwind classes
3. `ThemeSwitcher` component provides a UI toggle

### Theme Switcher

```go
ui.ThemeSwitcher()  // Cycles: System -> Light -> Dark
```

### Manual Theme Control

The client exposes two globals:
- `setTheme(mode)` -- "system", "light", or "dark"
- `toggleTheme()` -- toggles between light and dark

### Using Dark Mode in Components

Use Tailwind's `dark:` variant:

```go
ui.Div("bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100")
```

---

## Security

### Server-Side

- **JS String Escaping**: All strings embedded in JS are escaped (backslash, single quote, newlines, tabs) via `escJS()`
- **XSS Prevention**: `Text()` uses `textContent` (not `innerHTML`), preventing script injection
- **Safe Table Methods**: `FieldText()` for auto-escaped text, `Field()` for controlled `*Node` content
- **Panic Recovery**: Server panics in action handlers are recovered and surface as error toasts

### Client-Side

- **WebSocket-only**: No form submissions or XHR -- all interaction goes through the WS protocol
- **Auto-reconnect**: Dropped connections are automatically retried with an offline overlay
- **Not-found handling**: Missing DOM targets cancel Push goroutines and notify the server

---

## Examples

The `example/` directory contains a full working application demonstrating all features:

```bash
go run example/main.go
# Open http://localhost:1423
```

### Example Pages

| Page | Description |
|------|-------------|
| `/` | Component showcase (alerts, badges, cards, tabs, accordion, dropdowns, tooltips, progress) |
| `/counter` | Stateful counter with increment/decrement via WebSocket |
| `/hello` | Action responses: success, error, delayed, panic recovery |
| `/clock` | Live clock using `ctx.Push()` goroutine |
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
| `/others` | Miscellaneous component demos |

---

## Deployment

### Deploy Script

The `deploy` script creates and pushes version tags:

```bash
./deploy
```

- Versioning format: `v1.XXX` (e.g., `v1.001`, `v1.002`, `v1.003`)
- Auto-increments by `0.001` from the latest tag
- Ensures clean working tree before tagging
- Runs `go mod tidy`
- Creates annotated git tag and pushes to remote

### Using as a Dependency

```bash
go get github.com/michalCapo/g-sui@v1.001
```

---

## API Reference

### Package `ui`

#### Types

| Type | Description |
|------|-------------|
| `Node` | DOM element that compiles to JavaScript |
| `Action` | Server-side handler descriptor (or client-side JS) |
| `App` | Application container (routes, actions, WS clients) |
| `PageHandler` | `func(ctx *Context) *Node` |
| `ActionHandler` | `func(ctx *Context) string` |
| `Context` | Request data for pages and WS actions |
| `Response` | Multi-action response builder |
| `FormBuilder` | Declarative form builder |
| `FieldBuilder` | Single field configuration |
| `Field` | Field definition struct |
| `FieldType` | Field type enum |
| `FieldOption` | Value/label pair for select/radio |
| `FormErrors` | `map[string]string` of validation errors |
| `DataTable[T]` | Generic configurable table |
| `SimpleTable` | Non-generic quick table |
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

#### Constants

**Button Colors:** `BtnBlue`, `BtnRed`, `BtnGreen`, `BtnYellow`, `BtnPurple`, `BtnGray`, `BtnWhite`, `BtnBlueOutline`, `BtnRedOutline`, `BtnGreenOutline`

**Button Sizes:** `BtnXS`, `BtnSM`, `BtnMD`, `BtnLG`, `BtnXL`

**Field Types:** `FieldText`, `FieldPassword`, `FieldEmail`, `FieldNumber`, `FieldPhone`, `FieldDate`, `FieldTime`, `FieldDatetime`, `FieldUrl`, `FieldSearch`, `FieldTextarea`, `FieldSelect`, `FieldRadio`, `FieldRadioBtn`, `FieldRadioCard`, `FieldCheckbox`, `FieldHidden`

**Radio Styles:** `RadioInline`, `RadioButton`, `RadioCard`

#### Global Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `NewApp()` | `*App` | Create application |
| `El(tag, class...)` | `*Node` | Create element |
| `Target()` | `string` | Generate random DOM ID |
| `JS(code)` | `*Action` | Client-side-only action |
| `If(cond, node)` | `*Node` | Conditional render |
| `Or(cond, yes, no)` | `*Node` | Binary conditional |
| `Map[T](items, fn)` | `[]*Node` | Slice iteration |
| `Notify(variant, msg)` | `string` | Toast JS |
| `Redirect(url)` | `string` | Full redirect JS |
| `SetLocation(url)` | `string` | pushState JS |
| `Back()` | `*Action` | history.back() action |
| `SetTitle(title)` | `string` | Document title JS |
| `RemoveEl(id)` | `string` | Remove element JS |
| `SetText(id, text)` | `string` | Set text JS |
| `SetAttr(id, attr, val)` | `string` | Set attribute JS |
| `AddClass(id, cls)` | `string` | Add class JS |
| `RemoveClass(id, cls)` | `string` | Remove class JS |
| `Show(id)` | `string` | Show element JS |
| `Hide(id)` | `string` | Hide element JS |
| `Download(name, mime, b64)` | `string` | File download JS |
| `DragToScroll(id)` | `string` | Drag scroll JS |
| `NewResponse()` | `*Response` | Multi-action builder |
| `NewForm(id)` | `*FormBuilder` | Form builder |
| `NewDataTable[T](id)` | `*DataTable[T]` | Generic table |
| `NewSimpleTable(cols, cls...)` | `*SimpleTable` | Quick table |
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
