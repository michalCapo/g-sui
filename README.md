<div align="center">

# g-sui — Go Server‑Rendered UI

Build interactive, component‑styled pages in Go with server actions, simple partial updates, and no client framework.

</div>

---

## Highlights

- Server‑rendered HTML components with a small helper DSL
- Lightweight interactivity via server actions (`Click`, `Submit`, `Send`)
- Partial updates: re-render or replace only the target element
- Deferred fragments with skeletons via WebSocket patches (`ctx.Patch` + skeleton helpers)
- Form helpers with validation (uses `go-playground/validator`)
- A small set of UI inputs (text, password, number, date/time, select, checkbox, radio, textarea), buttons, tables, icons
- Toast messages: `Success`, `Error`, `Info`, and an error toast with a Reload button
- Built-in live status via WebSocket (`/__ws`) with an offline banner, automatic reconnect, and auto-reload on reconnect
- Optional dev autorestart (`app.AutoRestart(true)`) to rebuild and restart on changes

## Install

```bash
go get github.com/michalCapo/g-sui
```

Go 1.21+ recommended (module currently targets Go 1.23 toolchain).

## Quick start

Create a minimal app with a single page and a server action.

```go
package main

import (
    "github.com/michalCapo/g-sui/ui"
)

func main() {
    app := ui.MakeApp("en")
    app.AutoRestart(true) // optional during development (rebuild + restart on file changes)

    hello := func(ctx *ui.Context) string { ctx.Success("Hello from g-sui!"); return "" }

    layout := func(title string, body func(*ui.Context) string) ui.Callable {
        return func(ctx *ui.Context) string {
            nav := ui.Div("bg-white shadow mb-6")(
                ui.Div("max-w-5xl mx-auto px-4 py-2 flex items-center")(
                    ui.A("px-2 py-1 rounded text-sm whitespace-nowrap bg-blue-700 text-white hover:bg-blue-600",
                        ui.Href("/"), ctx.Load("/"),
                    )("Home"),
                ),
            )
            content := body(ctx)
            return app.HTML(title, "bg-gray-100 min-h-screen", nav+ui.Div("max-w-5xl mx-auto px-2")(content))
        }
    }

    app.Page("/", layout("Home", func(ctx *ui.Context) string {
        return ui.Div("p-4")(
            ui.Button().Color(ui.Blue).Class("rounded").Click(ctx.Call(hello).None()).Render("Say hello"),
        )
    }))

    app.Listen(":1422")
}
```

Run and open http://localhost:1422

### Favicon (optional)

If you have a favicon, serve it with caching and add a link tag:

```go
// embed your assets
//go:embed assets/*
var assets embed.FS

app.Favicon(assets, "assets/favicon.svg", 24*time.Hour)
app.HTMLHead = append(app.HTMLHead, `<link rel="icon" href="/favicon.ico" type="image/svg+xml">`)
```

The server sets the proper Content-Type for common favicon formats (including `image/svg+xml`).

## Examples

This repo ships an example app showcasing most components and patterns.

Run the examples server:

```bash
go run examples/main.go
```

The examples include:
- Showcase form with validations
- Inputs: text/password/number/date/time/datetime/textarea/select/checkbox/radio
- Buttons and color variants (solid/outline)
- Tables with simple helpers (including colspan and empty cells)
- Icons helpers and Hello demo (success/info/error/crash)
- Markdown rendering and a CAPTCHA demo
- Navigation bar that highlights the current page based on the URL path

### Active navigation highlight

The examples' top navigation highlights the last selected page (the current route) without any extra state. Compare each route path to `ctx.Request.URL.Path` and choose classes accordingly:

```go
// inside a layout callable
ui.Map(routes, func(r *route, _ int) string {
    base := "px-2 py-1 rounded text-sm whitespace-nowrap"
    cls := base + " hover:bg-gray-200"
    if ctx != nil && ctx.Request != nil && r.Path == ctx.Request.URL.Path {
        cls = base + " bg-blue-700 text-white hover:bg-blue-600"
    }
    return ui.A(cls, ui.Href(r.Path), ctx.Load(r.Path))(r.Title)
})
```

## Server actions and partial updates

Attach server actions via:

- `ctx.Call(fn).Render(target)` – replace inner HTML of `target`
- `ctx.Call(fn).Replace(target)` – replace the element itself
- `ctx.Call(fn).None()` – fire and forget (no swap)

On the server, an action has the signature `func(*ui.Context) string` and returns HTML (or an empty string if not swapping anything).

Forms can use `ctx.Submit(fn).Render/Replace/None()` and `ctx.Body(out)` to bind values to a struct.

## Messages and errors

- Toasts: `ctx.Success(msg)`, `ctx.Error(msg)`, `ctx.Info(msg)`
- Error toast with Reload button: `ctx.ErrorReload(msg)`
- Built‑in client handlers display a compact error panel for failed fetches (HTTP 500 etc.) with a Reload button.
- Server panics are recovered and surface as an error toast with Reload.

## Components (selection)

- Buttons: `ui.Button().Color(...).Size(...).Class(...).Href(...).Submit().Reset().Click(...)`
- Inputs: `ui.IText`, `ui.IPassword`, `ui.INumber`, `ui.IDate`, `ui.ITime`, `ui.IDateTime`, `ui.IArea`, `ui.ISelect`, `ui.ICheckbox`, `ui.IRadio`, `ui.IRadioButtons`
- Table: `ui.SimpleTable(cols, classes...)` with `Field`, `Empty`, `Class`, `Attr` (supports `colspan`)
- Icons: `ui.Icon`, `ui.Icon2`, `ui.Icon3`, `ui.Icon4`
- Markdown: `ui.Markdown(classes...)(content)`

Refer to the `examples/` directory for practical usage and composition patterns.

## Deferred fragments (WS + Skeleton)

Show a skeleton immediately, then push server patches when background work finishes. The example below mirrors the TS version: it returns a skeleton, then replaces it after ~2s and appends controls after ~3s.

```go
package pages

import (
    "time"
    "github.com/michalCapo/g-sui/ui"
)

// optional body payload to choose a skeleton kind
type deferForm struct { As ui.Skeleton }

// Deffered: return a skeleton now; push WS patches when ready
func Deffered(ctx *ui.Context) string {
    target := ui.Target()

    // read preferred skeleton type from the request (optional)
    form := deferForm{}
    _ = ctx.Body(&form)

    // replace the skeleton when the data is ready (~2s)
    go func() {
        time.Sleep(2 * time.Second)
        html := ui.Div("space-y-4", target)(
            ui.Div("bg-gray-50 dark:bg-gray-900 p-4 rounded shadow border rounded p-4")(
                ui.Div("text-lg font-semibold")("Deferred content loaded"),
                ui.Div("text-gray-600 text-sm")("This block replaced the skeleton via WebSocket patch."),
            ),
        )
        ctx.Patch(target.Replace(), html)
    }()

    // append more controls after more work (~3s)
    go func() {
        time.Sleep(3 * time.Second)
        controls := ui.Div("grid grid-cols-5 gap-4")(
            ui.Button().Color(ui.Blue).Class("rounded text-sm").
                Click(ctx.Call(Deffered, &deferForm{}).Replace(target)).Render("Default skeleton"),
            ui.Button().Color(ui.Blue).Class("rounded text-sm").
                Click(ctx.Call(Deffered, &deferForm{As: ui.SkeletonComponent}).Replace(target)).Render("Component skeleton"),
            ui.Button().Color(ui.Blue).Class("rounded text-sm").
                Click(ctx.Call(Deffered, &deferForm{As: ui.SkeletonList}).Replace(target)).Render("List skeleton"),
            ui.Button().Color(ui.Blue).Class("rounded text-sm").
                Click(ctx.Call(Deffered, &deferForm{As: ui.SkeletonPage}).Replace(target)).Render("Page skeleton"),
            ui.Button().Color(ui.Blue).Class("rounded text-sm").
                Click(ctx.Call(Deffered, &deferForm{As: ui.SkeletonForm}).Replace(target)).Render("Form skeleton"),
        )
        ctx.Patch(target.Append(), controls)
    }()

    // return chosen skeleton immediately
    return target.Skeleton(form.As)
}
```

Notes:

- `ctx.Patch(ts, html)` pushes server‑initiated patches to connected clients. Use `target.Render()`, `target.Replace()`, `target.Append()`, or `target.Prepend()` to describe the swap.
- Skeleton helpers: `target.Skeleton(kind)`, `target.SkeletonList(n)`, `target.SkeletonComponent()`, `target.SkeletonPage()`, `target.SkeletonForm()`.
- Actions: `ctx.Call(fn).Render/Replace/Append/Prepend/None()` for user‑initiated interactions.

## Development notes

- Live status: pages include a lightweight WS client bound to `/__ws` that shows an offline banner, reconnects automatically, and reloads the page on reconnect (useful when the server restarts). The panic fallback page also auto-reloads on reconnect.
- Autorestart: `app.AutoRestart(true)` watches your main package for file changes and rebuilds + restarts the app process. The built‑in WS client then reloads the page automatically on reconnect, so you get a smooth local DX without any extra setup.
- The library favors simple strings for HTML; helpers build class names and attributes for you.
- Validation uses `go-playground/validator`; see the login and showcase examples.

## License

MIT
