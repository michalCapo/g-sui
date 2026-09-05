# g-sui

Go server-rendered UI framework with real-time WebSocket patches.

g-sui compiles Go node trees into pure JavaScript. The runtime sends versioned messages containing JavaScript that performs `document.createElement()` calls directly -- no HTML templates, no client-side framework. SVG elements use `document.createElementNS()` with proper namespace handling. User interactions trigger server actions via WebSocket, which return typed `Result` effects.

## Documentation

Full API documentation: [`docs/documentation.md`](docs/documentation.md)

For SPA-style applications written in Go, start with the
[server-driven applications documentation](docs/documentation.md#server-driven-applications) and `go run ./example`.
It covers live links, typed actions/forms, region refresh, server-owned views,
URL-backed tables, shared authorization and cancellable subscriptions.

WebSocket connections are same-origin by default; set `App.AllowedOrigins` for additional trusted origins (or `"*"` to opt out). Tracked calls use reply envelopes so loading state clears even when an action returns no JavaScript.

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
RegisterAction → Result       ←→  WebSocket (__ws)
```

- **Server-centric** -- all DOM trees built in Go, compiled to JavaScript
- **WebSocket-only interactivity** -- click/submit events call server handlers, handlers return typed effects
- **Partial updates** -- replace, append, prepend, or innerHTML specific DOM targets
- **No client framework** -- the client is a small WS connector with keep-alive, silent reconnect, and a non-blocking offline badge
- **Tailwind CSS** -- loaded via browser CDN (`@tailwindcss/browser@4`)
- **Dark mode** -- built-in theme system (System/Light/Dark) with `ThemeSwitcher` component
- **Localization** -- per-component locale structs; English by default, override only what you need

## Features

- Server-rendered UI with a Go DSL (60+ element constructors, SVG namespace support)
- WebSocket actions with data payloads and field collection (`Collect`)
- Five DOM swap strategies: `ToJS`, `ToJSReplace`, `ToJSAppend`, `ToJSPrepend`, `ToJSInner`
- Typed `Result` effects for complex updates
- Real-time server push via `ctx.Push()` and broadcast via `app.Broadcast()`
- Custom HTTP routes: `app.GET()`, `app.POST()`, `app.DELETE()`
- Layout system via `app.Layout()` and custom `Handler()` for embedding
- SEO metadata: `app.Title`, `app.Description`, `app.HTMLHead`
- Conditional rendering helpers: `If`, `Or`, `Map`
- Toast notifications: success, error, error-reload, info
- JS helpers: `Redirect`, `SetTitle`, `RemoveEl`, `SetText`, `SetAttr`, `AddClass`, `RemoveClass`, `Show`, `Hide`, `Download`, `DragToScroll`

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
# Open http://localhost:1424
```

The example app includes pages demonstrating components, forms, tables, data panels, real-time updates, navigation, and more.

## Server Actions

```go
type RenameInput struct { Name string `json:"name"` }
rename := ui.RegisterAction(app, "profile.rename", func(ctx *ui.Context, input RenameInput) (ui.Result, error) {
    return ui.Result{}.SetText("name", input.Name).Toast("Saved"), nil
})
ui.Button().Text("Rename").OnClick(rename.Call(RenameInput{Name: "Alice"}))
```

### Multiple Effects

```go
return ui.Result{}.Morph("row-"+id, updatedRow).Toast("Updated").Navigate("/items"), nil
```

### Real-Time Push

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

## Theme & Dark Mode

```go
ui.ThemeSwitcher()  // System -> Light -> Dark toggle
```

Uses Tailwind `dark:` variants. Theme is persisted in localStorage and applied before render. The application stays hidden until its initial DOM, stylesheets, and active fonts are ready, preventing unstyled content and light-background flashes. It then fades in over 160 ms with a reduced-motion fallback. A four-second fail-safe reveals the page if a third-party resource stalls.

## Localization

Components use English text by default. Pass a locale struct only when you need non-English:

```go
// DataTable
table.Locale(&ui.TableLocale{Search: "Hledat...", Apply: "Pouzit", NoData: "Zadna data"})

// Collate
collate.Locale(&ui.CollateLocale{Filter: "Filtr", Reset: "Obnovit", SortBy: "Radit dle"})

// Confirm dialog
ui.ConfirmDialog("Smazat?", "Opravdu?", action, ui.ConfirmOpt{
    Locale: &ui.ConfirmLocale{Cancel: "Zrusit", Confirm: "Potvrdit"},
})
```

Each component has its own locale type (`TableLocale`, `CollateLocale`, `ConfirmLocale`, `ThemeSwitcherLocale`, `StepProgressLocale`) with only the fields it uses. See [`docs/documentation.md`](docs/documentation.md#localization) for all fields and defaults.

## Security

- **JS string escaping** -- all embedded strings escaped via `escJS()`
- **textContent** -- `Text()` uses `textContent`, not `innerHTML`, preventing XSS
- **Panic recovery** -- server panics surface as error toasts
- **WebSocket-only** -- no form submissions or XHR
- **Auto-reconnect** -- keep-alive ping, backoff retry, offline badge only after a grace window, and no page reload while `__ws.hold()` guards in-progress work; a restarted server is detected and the stale page resyncs

## License

MIT
