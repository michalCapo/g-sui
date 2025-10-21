# g-sui Agent Guide

## Mission & Scope
`g-sui` delivers server-rendered UI primitives for Go applications. The library pairs an ergonomic HTML DSL with server actions, WebSocket patching, and optional data helpers so you can build interactive dashboards without adopting a client SPA framework. This repository contains the core `ui` package plus a feature-rich example app that exercises the components, skeletons, and data patterns.

## Architecture Snapshot
- **App lifecycle**: `ui.MakeApp(locale)` constructs an `*ui.App`. Register handlers with `app.Page(path, layout, callable)` and start the HTTP/WebSocket server via `app.Listen(addr)`. `app.AutoRestart(true)` adds file watching (fsnotify) that rebuilds/restarts the server during local development.
- **Context**: Each request receives a `*ui.Context` holding the request/response, session helpers (`ctx.Session`), toast emitters (`ctx.Success/Error/Info`), redirects, CSP configuration, and HTML builders (`ctx.HTML`).
- **Rendering model**: UI is composed through lightweight string-returning functions (`type Callable func(*ui.Context) string`). Components (e.g., `ui.Div`, `ui.Button`, `ui.Input`) emit HTML with Tailwind-friendly classes. Markdown is supported through Goldmark.
- **Interactivity**: `ctx.Call(fn, payload)` wires server actions to events (`Click`, `Submit`, `Send`). Actions can `Render`, `Replace`, `Append`, `Prepend`, or do nothing.
- **Partial updates**: `ui.Target()` pairs an element ID with a swap strategy. Server-initiated updates use `ctx.Patch(target.Swap(), html)` over a built-in WebSocket (`/__ws`), enabling deferred fragments and streaming updates.
- **Forms & validation**: Form helpers integrate with `go-playground/validator`. Inputs expose type-specific attributes, error messages, and sanitation utilities (`validateNumericInput`, etc.).
- **Data tooling**: The Query/Collate helpers (see `examples/pages/collate.go`) combine filters, sort, pagination, XLS export (via `excelize`), and GORM integration for CRUD-style dashboards.
- **Security**: HTML attributes and JS strings are escaped centrally; `ctx.SetDefaultCSP()` provides strict CSP defaults. Input length/character caps guard against oversized payloads. Captcha helpers (`ui.Captcha`, `ui.Captcha2/3`) implement rate-limiting and shared state (Redis/DB ready).
- **Theming**: Dark mode and theme switching ship by default (`ui.ThemeSwitcher`). CSS presets (e.g., `ui.Blue`, `ui.Button`) simplify consistent styling.

## Repository Layout
- `ui/` — Core library: rendering DSL, component definitions (buttons, inputs, tables, icons), server runtime (`ui.server.go`), captcha variants, query helpers, toast utilities, and swap constants.
- `examples/` — Showcase application:
  - `main.go` bootstraps routes, navigation, assets, autoreload, and theme switcher.
  - `pages/` contains focused demos (forms, inputs, append/prepend, deferred patches, collate/XLS, captcha, shared state, etc.).
  - `assets/` holds example static files (favicon, etc.).
  - `examples` (binary) is a previously built example executable; safe to ignore/remove when packaging.
- `docs/` — Ancillary docs (e.g., Lighthouse screenshot).
- `README.md` — In-depth introduction, usage samples, security guidelines, and component walkthrough.
- `go.mod`/`go.sum` — Go 1.23 module, listing dependencies like `fsnotify`, `validator`, `gorm`, `excelize`, `goldmark`, and `golang.org/x/net/websocket`.

## Key APIs & Patterns
- **Page registration**: `app.Page(path, layout(title, body), handler)`; layouts typically build nav bars and include `ui.ThemeSwitcher`.
- **Component DSL**: Chainable builders such as `ui.Div(class)(children...)`, `ui.Button().Color(ui.Blue).Click(...)`, `ui.Input().Type("email")`.
- **Actions**: `ctx.Call(handler, payload?).Render(target)` for synchronous responses; `.Replace/.Append/.Prepend/.None` adjust swap behavior. For long-running work, return a skeleton immediately, then `ctx.Patch` new HTML asynchronously.
- **Targets**: `target := ui.Target()` yields a unique ID. Use `target.Render()`, `target.Replace()`, `target.Append()`, `target.Skeleton(kind)` to manage placeholder and updates.
- **Forms**: Compose with `ui.Form`, `ui.Input`, `ui.Checkbox`, etc. Tie validation errors to fields via returned error structures. Examples in `examples/pages/login.go` and `showcase.go`.
- **Tables**: `ui.Table` helpers render headers, rows, empty states, and allow `FieldText` vs `FieldHTML` for safe output.
- **Data flows**: `ui.TCollate` orchestrates search/sort/paging. Combine with SQLite or real DB using GORM. Export spreadsheets via `ui.XLS` helpers (see `collate.go`).
- **Sessions**: `ctx.Session(db, name)` persists JSON blobs keyed by session ID in `_session` table. Works with any GORM-backed database.
- **Captcha**: `ui.Captcha` (in-memory) and `ui.Captcha2/3` (shared store) rate-limit and validate forms; integrate with `ui.Form` actions.

## Component Toolkit
- **Layout & HTML**: `ui.Div`, `ui.Span`, `ui.A`, `ui.Form`, `ui.Textarea`, `ui.Select`, `ui.Option`, `ui.List`, `ui.ListItem`, `ui.Canvas`, `ui.Img`, `ui.Input` wire raw HTML. Use `ui.Map`, `ui.For`, `ui.If/Iff/Or` for templating control flow. `ui.Markdown` renders Goldmark markdown blocks and `ui.Script` injects inline scripts. `ui.ThemeSwitcher` adds the built-in dark/light toggle. Skeleton helpers (`target.Skeleton*`, `ui.Skeleton*Block`) provide loading placeholders.
- **Buttons & Labels**: `ui.Button()` returns a fluent builder supporting `Submit`, `Reset`, `.Color(ui.Blue|ui.Green|...)`, `.Size(ui.SM|ui.MD)` and `.Click(ctx.Call(...))`. Label helpers (`ui.Label(target).Required(true).Class(...)`) render accessible labels with optional required asterisks.
- **Input suite**: `ui.IText`, `ui.IPassword`, `ui.IArea` (textarea), `ui.IDate`, `ui.ITime`, `ui.IDateTime`, `ui.INumber`, `ui.ISelect`, `ui.ICheckbox`, `ui.IRadio`, `ui.IRadioButtons`, and `ui.IValue` share fluent setters for `Placeholder`, `Required`, `Disabled`, `Change`, `Click`, numeric/date ranges, and validator error binding via `.Error(&err)`.
- **Data display**: `ui.Table` and `ui.SimpleTable` offer table composition with typed rows and safe text fields. `ui.Icon*` helpers embed FontAwesome-compatible icon markup with position variants. Toasts, cards, and layout primitives can reuse CSS constants like `ui.Blue`, `ui.GrayOutline`, `ui.INPUT`.
- **Captcha & Security**: `ui.Captcha` renders a simple challenge, while `ui.Captcha2`/`ui.Captcha3` add shared store backing (Redis/DB-ready) with methods to configure attempts, lifetime, and auto-validation callbacks.
- **Query tooling**: `ui.Collate` composes filter/search/sort forms, ties to `ctx.Call` actions, and exposes Excel export hooks via `collate.OnExcel`. Normalization helpers (`ui.NormalizeForSearch`, `ui.RegisterSQLiteNormalize`) smooth search input and register SQLite functions when needed.

## Development Workflow
- **Run the showcase**: `go run examples/main.go` → visit `http://localhost:1422`.
- **Autoreload**: Uncomment `app.AutoRestart(true)` in the example or your app to enable rebuild-on-change (uses `fsnotify` + `go build` under the hood; requires exec permissions).
- **Build**: `go build ./...` validates compilation (library + examples).
- **Tests**: There are no dedicated test packages yet. `go test ./...` is the default entry point once tests are added; ensure the Go build cache directory is writable in your environment.
- **Lint/format**: The project relies on `gofmt` (standard gofmt on save). No extra linters configured.
- **Dependencies**: Managed with Go modules; `go mod tidy` keeps the tree clean. No additional toolchain beyond Go 1.23+, though examples rely on SQLite for the collate demo (the code seeds an in-memory DB).

## Contributing Tips
- Favor composition-friendly callables (`func(ctx *ui.Context) string`) and keep side effects inside actions or goroutines guarded by `ctx.Patch`.
- Reuse shared CSS constants (e.g., `ui.Blue`, `ui.INPUT`) to maintain visual consistency.
- Escape or sanitize user-provided data unless explicitly rendered via trusted helpers (`HeadHTML`, etc.).
- When adding new real-time sections, wrap long-running work in goroutines and `ctx.Patch` results instead of blocking handlers.
- Update the example app to demonstrate new primitives; it doubles as documentation and regression coverage.

## Context Helpers
- **Request/session**: `ctx.Request`/`ctx.Response` expose raw HTTP types. `ctx.IP()` gives the remote address, and `ctx.Session(db, name)` returns a GORM-backed JSON session wrapper with `.Load`/`.Save`.
- **Body & parsing**: `ctx.Body(&payload)` hydrates structs from form submissions or JSON. Input guards (`validateInputSafety`, `validateNumericInput`) run automatically during action decoding.
- **Actions & events**: `ctx.Call(handler, values...).Render|Replace|Append|Prepend|None()` attaches to `Click`, `Submit`, or `Send`. Shortcuts `ctx.Click`, `ctx.Submit`, `ctx.Send`, and `ctx.Post` wire swap strategy + payloads. `ctx.Load(href)` prepares SPA-lite navigation attributes; `ctx.Reload()`/`ctx.Redirect(url)` produce JS commands.
- **Realtime updates**: `ctx.Patch(targetSwap, html, clear...)` pushes WebSocket updates to connected clients (render/replace/append/prepend).
- **Feedback**: `ctx.Success`, `ctx.Error`, `ctx.ErrorReload`, `ctx.Info` enqueue toast notifications; `ctx.DownloadAs` streams files; `ctx.Translate` resolves i18n strings.
- **Security**: `ctx.SetDefaultCSP()` or `ctx.SetCSP(policy)` manage CSP headers; `ctx.SetSecurityHeaders`/`ctx.SetCustomSecurityHeaders` set broader security policies.
- **App-level**: `app.Page`, `app.Action`, `app.Callable`, `app.Assets`, `app.Favicon`, `app.AutoRestart`, `app.StartSweeper`, and `app.Listen` control routing, static assets, hot reload, session sweeps, and server lifecycle.

## References
- README (`README.md`) contains canonical walkthroughs, security notes, and deferred fragment examples. Start there for narrative guidance, then inspect `examples/pages/*.go` for concrete patterns.
- Lighthouse screenshot (`docs/lighthouse-scores.png`) showcases the performance profile of the sample app.
