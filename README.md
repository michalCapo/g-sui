<div align="center">

# g-sui — Go Server‑Rendered UI

Build interactive, component‑styled pages in Go with server actions, simple partial updates, and no client framework.

</div>

---

## Highlights

- Server‑rendered HTML components with a small helper DSL
- Lightweight interactivity via server actions (`Click`, `Submit`, `Send`)
- Partial updates: re-render or replace only the target element
- Deferred fragments with skeletons and WebSocket patching (`ctx.Defer(...)`)
- Form helpers with validation (uses `go-playground/validator`)
- A small set of UI inputs (text, password, number, date/time, select, checkbox, radio, textarea), buttons, tables, icons
- Toast messages: `Success`, `Error`, `Info`, and an error toast with a Reload button
- Optional dev autoreload overlay (`app.Autoreload(true)`) for quick feedback
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
    app.Autoreload(true)  // optional during development (browser auto-reload + offline banner)

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

Show a skeleton immediately, then replace it with the real content when the server finishes rendering. Use `ctx.Defer(fn)` to run a callable in the background and push a WebSocket patch upon completion.

```go
// inside a page handler
target := ui.Target()

heavy := func(c *ui.Context) string {
    time.Sleep(800 * time.Millisecond) // simulate work
    return ui.Div("bg-white p-4 rounded shadow border", target)(
        ui.Div("font-semibold")("Deferred content loaded"),
        ui.Div("text-gray-500 text-sm")("Replaced via WS patch"),
    )
}

// Show a component skeleton now; replace `target` via WS when ready
placeholder := ctx.Defer(heavy).SkeletonComponent().Replace(target)

return app.HTML(
    "Deferred Demo",
    "bg-gray-100 min-h-screen",
    ui.Div("max-w-5xl mx-auto p-6")(
        ui.Div("text-xl font-bold mb-2")("Deferred fragment"),
        placeholder,
    ),
)
```

Notes:

- `Skeleton(...)`, `SkeletonList(n)`, `SkeletonComponent()`, `SkeletonPage()`, and `SkeletonForm()` set the placeholder.
- Use `.Render(target)` to swap `innerHTML`, or `.Replace(target)` to swap the whole element.
- `.None()` runs the callable for side-effects and returns a minimal placeholder (`<!-- -->`).

## Development notes

- Autoreload: `app.Autoreload(true)` injects a tiny WS client that shows an offline banner and reloads on reconnect (handy during dev).
- Autorestart: `app.AutoRestart(true)` watches your main package for file changes and rebuilds + restarts the app process. Combine with Autoreload for a smooth local DX.
- The library favors simple strings for HTML; helpers build class names and attributes for you.
- Validation uses `go-playground/validator`; see the login and showcase examples.

## License

MIT
