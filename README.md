# g-sui

Go server-rendered UI framework with real-time WebSocket patches.

g-sui compiles Go node trees into pure JavaScript. The browser receives raw JS that performs `document.createElement()` calls directly -- no HTML templates, no JSON intermediate, no client-side framework. SVG elements use `document.createElementNS()` with proper namespace handling. User interactions trigger server actions via WebSocket, which respond with JS strings for DOM mutations.

## Documentation

Full API documentation: [`docs/documentation.md`](docs/documentation.md)

## Install

```bash
go get github.com/michalCapo/g-sui
```

Requires Go 1.24+.

## Quick Start

```go
package main

import r "github.com/michalCapo/g-sui/ui"

func main() {
    app := r.NewApp()

    app.Page("/", func(ctx *r.Context) *r.Node {
        return r.Div("min-h-screen bg-gray-100 p-8").Render(
            r.H1("text-3xl font-bold").Text("Hello World"),
        )
    })

    app.Listen(":8080")
}
```

## Architecture

```
Server (Go)                          Browser
─────────────                        ───────
PageHandler → *Node → .ToJS()   →   Minimal HTML + <script>
ActionHandler → JS string       ←→  WebSocket (__ws)
```

- **Server-centric** -- all DOM trees built in Go, compiled to JavaScript
- **WebSocket-only interactivity** -- click/submit events call server handlers, responses are JS strings
- **Partial updates** -- replace, append, prepend, or innerHTML specific DOM targets
- **No client framework** -- the client is a ~120-line WS connector with offline overlay and auto-reconnect
- **Tailwind CSS** -- loaded via browser CDN (`@tailwindcss/browser@4`)
- **Dark mode** -- built-in theme system (System/Light/Dark) with `ThemeSwitcher` component

## Features

- Server-rendered UI with a Go DSL (60+ element constructors, SVG namespace support)
- WebSocket actions with data payloads and field collection (`Collect`)
- Five DOM swap strategies: `ToJS`, `ToJSReplace`, `ToJSAppend`, `ToJSPrepend`, `ToJSInner`
- Multi-action `Response` builder for complex updates
- Real-time server push via `ctx.Push()` and broadcast via `ctx.Broadcast()` / `app.Broadcast()`
- Custom HTTP routes: `app.GET()`, `app.POST()`, `app.DELETE()`
- Layout system via `app.Layout()` and custom `Handler()` for embedding
- SEO metadata: `app.Title`, `app.Description`, `app.HTMLHead`
- Conditional rendering helpers: `If`, `Or`, `Map`
- Toast notifications: success, error, error-reload, info
- JS helpers: `Redirect`, `SetLocation`, `SetTitle`, `RemoveEl`, `SetText`, `SetAttr`, `AddClass`, `RemoveClass`, `Show`, `Hide`, `Download`, `DragToScroll`

### Components

- **Alert** -- info/success/warning/error variants, dismissible, localStorage persistence
- **Badge** -- solid/outline/soft color variants, dot indicator, icon support
- **Button** -- color/size presets, icon, link, submit, disabled states
- **Card** -- header/body/footer, image, 4 variants (shadowed/bordered/flat/glass), hover effect
- **Accordion** -- bordered/ghost/separated variants, single/multiple open
- **Tabs** -- underline/pills/boxed/vertical styles, keyboard navigation, ARIA
- **Dropdown** -- items, headers, dividers, danger items, 4 positions, auto-close
- **Tooltip** -- 4 positions, 6 color variants, configurable delay
- **Progress** -- gradient, striped, animated, indeterminate, labels
- **Step Progress** -- step X of Y with progress bar
- **Confirm Dialog** -- overlay with confirm/cancel actions
- **Skeleton Loaders** -- table, cards, list, component, page, form
- **Markdown** -- goldmark renderer
- **Icon** -- Material Icons Round with `IconText` helper, inline SVG with automatic namespace
- **Theme Switcher** -- System/Light/Dark toggle
- **reCAPTCHA v3** -- auto-refresh token

### Forms

- Declarative `FormBuilder` with 17 field types
- Client-side validation (required, regex pattern)
- Server-side validation with `FormErrors`
- Multiple submit buttons with action identification
- Radio variants: inline, button-style, card-style
- Form-scoped radio names (multiple forms on same page)

### Data Tables

- Generic `DataTable[T]` with search, sort, pagination, column filters, export
- Column definitions with `*Node` content or plain text
- Per-column filters: text, date, number, select with operators
- Expandable row detail (accordion)
- Debounced search, click-to-sort headers, page range with ellipsis
- `SimpleTable` for quick non-generic tables

### Collate (Data Panel)

- Generic `Collate[T]` -- card/list-style data component with slide-out filter/sort panel
- Configurable sort fields and filter types: boolean, date range, select, multi-check
- Debounced search, load-more pagination, export action
- Expandable row detail
- Custom row rendering via callback
- Server-driven filter/sort/search with `CollateFilterValue` payloads

## Examples

```bash
go run example/main.go
# Open http://localhost:1423
```

The example app includes 23 pages demonstrating components, forms, tables, data panels, real-time updates, navigation, and more.

## Server Actions

```go
// Register action
app.Action("counter.inc", func(ctx *ui.Context) string {
    count++
    return ui.Span().ID("count").Text(fmt.Sprintf("%d", count)).ToJSReplace("count")
})

// Attach to element
ui.Button("...").Text("+1").OnClick(&ui.Action{Name: "counter.inc"})
```

### Multi-Action Response

```go
return ui.NewResponse().
    Replace("row-"+id, updatedRow).
    Toast("success", "Updated").
    Navigate("/items").
    Build()
```

### Real-Time Push

```go
go func() {
    for {
        time.Sleep(time.Second)
        if err := ctx.Push(ui.SetText("clock", time.Now().Format("15:04:05"))); err != nil {
            return
        }
    }
}()
```

## Theme & Dark Mode

```go
ui.ThemeSwitcher()  // System -> Light -> Dark toggle
```

Uses Tailwind `dark:` variants. Theme is persisted in localStorage and applied before render to prevent FOUC.

## Security

- **JS string escaping** -- all embedded strings escaped via `escJS()`
- **textContent** -- `Text()` uses `textContent`, not `innerHTML`, preventing XSS
- **Panic recovery** -- server panics surface as error toasts
- **WebSocket-only** -- no form submissions or XHR
- **Auto-reconnect** -- offline overlay with automatic retry

## Deploy

```bash
./deploy
```

Creates an annotated git tag and pushes to remote. Version format: `v1.XXX`, auto-incrementing by `0.001`.

## License

MIT
