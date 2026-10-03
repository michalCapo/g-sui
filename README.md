# g-sui

Go server-rendered UI framework with real-time WebSocket patches.

g-sui compiles Go node trees into pure JavaScript. The runtime sends versioned messages containing JavaScript that performs `document.createElement()` calls directly -- no HTML templates, no client-side framework. SVG elements use `document.createElementNS()` with proper namespace handling. User interactions trigger server actions via WebSocket, which return typed `Result` effects.

## Documentation

Full API documentation: [`docs/documentation.md`](docs/documentation.md)

For SPA-style applications written in Go, start with the
[server-driven applications documentation](docs/documentation.md#server-driven-applications) and `make example`.
It covers live links, typed actions/forms, region refresh, server-owned views,
URL-backed tables, shared authorization and cancellable subscriptions.

WebSocket connections are same-origin by default; set `App.AllowedOrigins` for additional trusted origins (or `"*"` to opt out). Tracked calls use reply envelopes so loading state clears even when an action returns no JavaScript.

## Install

```bash
go get github.com/michalCapo/g-sui
```

Requires Go 1.26+ (see `go.mod`).

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
PageHandler → *Node → JS        →   Minimal HTML + <script>
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
- Typed WebSocket actions with data payloads and field collection (`ref.Call(data).Collect(ids...)`)
- Typed `Result` effects: `Morph`, `Replace`, `Append`, `Prepend`, `Remove`, `Refresh`, `Run`
- Real-time server push via `ctx.Push()` and broadcast via `app.Broadcast()`
- Custom HTTP routes: `app.GET()`, `app.POST()`, `app.DELETE()`
- Layout system via `app.Layout()` and custom `Handler()` for embedding
- SEO metadata: `app.Title`, `app.Description`, `app.HTMLHead`
- Conditional rendering helpers: `If`, `Or`, `Map`
- Toast notifications: success, error, error-reload, info
- Local UI actions without JavaScript: `Show`, `Hide`, `Toggle`, `ToggleClass`, `SetAttr`, `SetValue`, `Focus`, `OpenDialog`, `CopyText`, `Navigate`, `Seq`, `Confirm`, `Delay` and more
- Node behaviors: `OnKey`, `Shortcut`, `OnOutsideClick`, `DragToScroll`, `ActiveClass`
- `App.Widget` for third-party JS libraries with mount and cleanup

## Writing UI with an LLM

See [`AGENTS.md`](AGENTS.md). It maps common UI needs to g-sui APIs, so agents
do not write raw JavaScript. Raw JS is only available through `Unsafe*` APIs.

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
make example
# Open http://localhost:1424
```

The example app includes pages demonstrating components, forms, tables, data panels, real-time updates, navigation, and more.

## Development commands

Run `make` for help. Commands use GNU Make and Bash from the repository root:

- `make run`: example with Air live reload, using the existing Go watch settings.
- `make example`: example without live reload, at `http://localhost:1424`.
- `make check`: go fix diagnostics, formatting, vet, staticcheck, optional gopls
  and golangci-lint (including revive rules), deadcode, build, and tests.
  Checks continue after failures and return a failing status if any check fails.
- `make build`: build all packages.
- `make test`: Go tests; install Node to also run the client runtime tests.
- `make test-race`: Go tests with the race detector.
- `make test-browser`: browser regression checks against an already running example.
  Requires Node and Playwright on `NODE_PATH`, plus Chromium. Set `GSUI_URL` to
  select the example address and `GSUI_BROWSER` to select a Chromium executable.
- `make tidy`: update module metadata.
- `make release`: check tracked changes, tidy modules, create the next
  `v1.MINOR.PATCH` annotated tag, push it and the current branch, and register
  the version with the Go module proxy. This command publishes a release.

`make check` requires `staticcheck` and `deadcode` on `PATH`; `gopls` and
`golangci-lint` are optional. `make run` requires Air. Existing Go environment
variables and tool configuration still apply. Temporary files use `TMPDIR`
(default `/tmp`) and are removed on success, failure, or a handled signal.
Air's temporary logs/state use a separate `.gsui-air.*` directory, also removed
on exit. Dependencies, tool caches, module changes, and user files are kept.

Configure application launchers with `make run`. In Libro, use its application controls
with `make run` or `make example` configured as the project start command.

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

Use `ui.RegisterSubscription(app, name, func(ctx *ui.Context, in T) error)` to
read data passed as `node.Subscribe(name, data)`.

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
