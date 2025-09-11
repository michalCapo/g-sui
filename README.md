<div align="center">

# g-sui — Go Server‑Rendered UI

Build interactive, component‑styled pages in Go with server actions, simple partial updates, and no client framework.

</div>

---

## Highlights

- Server‑rendered HTML components with a small helper DSL
- Lightweight interactivity via server actions (`Click`, `Submit`, `Send`)
- Partial updates: re-render, replace, append, or prepend only the target
- Deferred fragments with skeletons via WebSocket patches (`ctx.Patch` + skeleton helpers)
- Query/Collate helper for data UIs: search, sort, filters, paging, and XLS export (works with `gorm`)
- Form helpers with validation (uses `go-playground/validator`)
- A small set of UI inputs (text, email, phone, password, number, date/time/datetime, select, checkbox, radio, textarea), buttons, tables, icons
- Toast messages: `Success`, `Error`, `Info`, and an error toast with a Reload button
- Built-in live status via WebSocket (`/__ws`) with an offline banner, automatic reconnect, and auto-reload on reconnect
- Built-in dark mode with a tiny theme switcher (`ui.ThemeSwitcher`) cycling System → Light → Dark
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
- Inputs: text/email/phone/password/number/date/time/datetime/textarea/select/checkbox/radio
- Buttons and color variants (solid/outline)
- Tables with simple helpers (including colspan and empty cells)
- Icons helpers and Hello demo (success/info/error/crash)
- Markdown rendering and a CAPTCHA demo
- Query demo: in-memory SQLite + GORM with `ui.TCollate` (search, sort, filters, paging, XLS export)
- Append/Prepend demo for list updates
- Clock demo and deferred fragments (skeleton → WS patch)
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
- `ctx.Call(fn).Append(target)` – insert HTML at the end of the target
- `ctx.Call(fn).Prepend(target)` – insert HTML at the beginning of the target
- `ctx.Call(fn).None()` – fire and forget (no swap)

On the server, an action has the signature `func(*ui.Context) string` and returns HTML (or an empty string if not swapping anything).

Forms can use `ctx.Submit(fn).Render/Replace/None()` and `ctx.Body(out)` to bind values to a struct.

## Query/Collate (search, sort, filters, paging, XLS)

Build data-centric pages quickly using `ui.TCollate`. Define fields, choose which ones participate in search/sort/filter, optionally enable Excel export, and provide a row renderer. Works with `gorm` and SQLite (and other DBs).

Minimal example (excerpt):

```go
type Person struct {
    ID        uint `gorm:"primaryKey"`
    Name      string
    Surname   string
    Email     string
    Country   string
    Status    string
    Active    bool
    CreatedAt time.Time
    LastLogin time.Time // zero means "never"
}

// Recommended for SQLite: install normalize() for accent-insensitive search
_ = ui.RegisterSQLiteNormalize(db)

// Define fields
name := ui.TField{DB: "name", Field: "Name", Text: "Name"}
surname := ui.TField{DB: "surname", Field: "Surname", Text: "Surname"}
email := ui.TField{DB: "email", Field: "Email", Text: "Email"}
status := ui.TField{DB: "status", Field: "Status", Text: "Status", As: ui.SELECT, Options: ui.MakeOptions([]string{"new","active","blocked"})}
active := ui.TField{DB: "active", Field: "Active", Text: "Active", As: ui.BOOL}
created := ui.TField{DB: "created_at", Field: "CreatedAt", Text: "Created", As: ui.DATES}
lastLogin := ui.TField{DB: "last_login", Field: "LastLogin", Text: "Has logged in", As: ui.NOT_ZERO_DATE}
neverLogin := ui.TField{DB: "last_login", Field: "LastLogin", Text: "Never logged in", As: ui.ZERO_DATE}

init := &ui.TQuery{Limit: 8, Order: "surname asc"}

collate := ui.Collate[Person](init)
collate.Search(surname, name, email, status)
collate.Sort(surname, email, lastLogin)
collate.Filter(active, lastLogin, neverLogin, created)
collate.Excel(surname, name, email, status, active, created, lastLogin)

// How each row renders
collate.Row(func(p *Person, _ int) string {
    badge := ui.Span("px-2 py-0.5 rounded text-xs border "+
        map[bool]string{true: "bg-green-100 text-green-700 border-green-200", false: "bg-gray-100 text-gray-700 border-gray-200"}[p.Active])(
        map[bool]string{true: "active", false: "inactive"}[p.Active],
    )
    return ui.Div("bg-white rounded border p-3 flex items-center justify-between")(
        ui.Div("font-semibold")(fmt.Sprintf("%s %s", p.Surname, p.Name))+
        ui.Div("text-sm text-gray-500 ml-2")(p.Email)+
        badge,
    )
})

// Render the full UI (search/sort/filters/paging/export)
content := collate.Render(ctx, db)
```

Notes:

- Use `ui.MakeOptions(slice)` to build options for `SELECT` fields quickly.
- `ZERO_DATE` / `NOT_ZERO_DATE` are convenient filters for nullable timestamps or "never" semantics.
- Excel export is enabled by calling `collate.Excel(...)` with the fields to include.
- For SQLite, `ui.RegisterSQLiteNormalize(db)` installs a `normalize()` function to improve diacritics/accents-insensitive search.

## Messages and errors

- Toasts: `ctx.Success(msg)`, `ctx.Error(msg)`, `ctx.Info(msg)`
- Error toast with Reload button: `ctx.ErrorReload(msg)`
- Built‑in client handlers display a compact error panel for failed fetches (HTTP 500 etc.) with a Reload button.
- Server panics are recovered and surface as an error toast with Reload.

## Components (selection)

- Buttons: `ui.Button().Color(...).Size(...).Class(...).Href(...).Submit().Reset().Click(...)`
- Inputs: `ui.IText`, `ui.IEmail`, `ui.IPhone`, `ui.IPassword`, `ui.INumber`, `ui.IDate`, `ui.ITime`, `ui.IDateTime`, `ui.IArea`, `ui.ISelect`, `ui.ICheckbox`, `ui.IRadio`, `ui.IRadioButtons`
- Table: `ui.SimpleTable(cols, classes...)` with `Field`, `Empty`, `Class`, `Attr` (supports `colspan`)
- Icons: `ui.Icon`, `ui.Icon2`, `ui.Icon3`, `ui.Icon4`
- Markdown: `ui.Markdown(classes...)(content)`

Refer to the `examples/` directory for practical usage and composition patterns.

## Theme & Dark Mode

- Built-in dark theme overrides load with `ui.MakeApp`. Use `ui.ThemeSwitcher("")` to render a compact toggle that cycles System → Light → Dark.
- Typical placement is in your layout’s top bar:

```go
nav := ui.Div("bg-white shadow mb-6")(
    ui.Div("max-w-5xl mx-auto px-4 py-2 flex items-center gap-2")(
        // ... your nav links ...
        ui.Flex1,
        ui.ThemeSwitcher(""),
    ),
)
```

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
