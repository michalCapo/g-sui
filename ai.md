# g-sui Agent Guide

## Mission & Scope
`g-sui` (Go Server-Rendered UI) is a comprehensive framework for building interactive, server-rendered web applications in Go. It provides an ergonomic HTML DSL, server actions, WebSocket-based real-time updates, and powerful data management helpers. The framework enables building modern, interactive dashboards and UIs without requiring client-side JavaScript frameworks or SPAs. This repository contains the core `ui` package plus a feature-rich example app demonstrating all components, patterns, and capabilities.

## Core Philosophy
- **Server-first**: All logic runs on the server; minimal client JavaScript
- **Progressive enhancement**: Works without JavaScript, enhanced with WebSockets
- **Type-safe**: Leverages Go's type system for safety and IDE support
- **Performance**: Server-rendered HTML achieves excellent Lighthouse scores (97 Performance, 100 Accessibility)
- **Developer experience**: Hot reload, clear APIs, comprehensive examples

## Architecture Snapshot

### Application Lifecycle
- **Initialization**: `ui.MakeApp(locale)` constructs an `*ui.App` instance
- **Routing**: Register pages with `app.Page(path, layout(title, body), handler)` where handlers are `Callable` functions
- **Server**: Start HTTP/WebSocket server with `app.Listen(addr)` - automatically sets up `/__ws` endpoint
- **Development**: `app.AutoRestart(true)` enables file watching (fsnotify) that rebuilds and restarts on changes
- **Assets**: Serve static files with `app.Assets(fs, path)` and favicons with `app.Favicon(fs, path, cacheDuration)`

### Request Context
Each request receives a `*ui.Context` providing:
- **HTTP**: `ctx.Request` and `ctx.Response` for raw HTTP access
- **Session**: `ctx.Session(db, name)` returns GORM-backed JSON session storage with `.Load()`/`.Save()` methods
- **User feedback**: `ctx.Success(msg)`, `ctx.Error(msg)`, `ctx.Info(msg)`, `ctx.ErrorReload(msg)` for toast notifications
- **Navigation**: `ctx.Load(href)` for SPA-like navigation, `ctx.Reload()`, `ctx.Redirect(url)` for page changes
- **Security**: `ctx.SetDefaultCSP()` or `ctx.SetCSP(policy)` for Content Security Policy headers
- **IP**: `ctx.IP()` returns client IP address
- **Body parsing**: `ctx.Body(&struct)` hydrates structs from form/JSON submissions with automatic validation

### Rendering Model
- **Callables**: UI is built through `type Callable func(*ui.Context) string` functions that return HTML strings
- **Component DSL**: Chainable builders like `ui.Div(class)(children...)`, `ui.Button().Color(ui.Blue).Click(...)`
- **Styling**: Components emit Tailwind CSS-friendly classes; CSS constants (`ui.Blue`, `ui.INPUT`, etc.) ensure consistency
- **Markdown**: `ui.Markdown(classes...)(content)` renders markdown via Goldmark
- **Control flow**: `ui.Map`, `ui.For`, `ui.If`, `ui.Iff`, `ui.Or` for templating logic

### Interactivity & Server Actions
- **Action binding**: `ctx.Call(handler, payload?)` creates action descriptors
- **Swap strategies**: `.Render(target)` replaces inner HTML, `.Replace(target)` replaces element, `.Append(target)` adds to end, `.Prepend(target)` adds to start, `.None()` fire-and-forget
- **Event types**: Actions attach to `Click`, `Submit`, `Send` events
- **Shortcuts**: `ctx.Click`, `ctx.Submit`, `ctx.Send`, `ctx.Post` provide convenient wrappers

### Real-Time Updates (WebSocket)
- **Endpoint**: Built-in WebSocket server at `/__ws` (auto-initialized)
- **Server-initiated**: `ctx.Patch(targetSwap, html, clear?)` pushes updates to all connected clients
- **Client features**: Automatic reconnect, offline banner, auto-reload on reconnect after server restart
- **Heartbeat**: 25-second ping/pong keeps connections alive; stale connections closed after 75 seconds
- **Deferred fragments**: Return skeleton immediately, then `ctx.Patch` real content when ready (from goroutines)

### Forms & Validation
- **Integration**: Uses `go-playground/validator` for struct validation
- **Input safety**: Automatic validation of field names, values, lengths (MaxBodySize: 10MB, MaxFieldNameLen: 256, MaxFieldValueLen: 1MB, MaxFieldCount: 1000)
- **Error binding**: Inputs support `.Error(&err)` to display validation errors
- **Type-specific**: Inputs have type-specific validation (dates, numbers, emails, etc.)

### Data Management (Query/Collate)
- **Purpose**: Build data-centric UIs with search, sort, filters, pagination, and Excel export
- **GORM integration**: Works seamlessly with GORM and any SQL database
- **Features**: 
  - Search across multiple fields with accent-insensitive matching
  - Sort by multiple fields
  - Filter by various field types (text, select, boolean, dates, zero/not-zero dates)
  - Pagination with configurable page size
  - Excel export via `excelize` library
- **SQLite support**: `ui.RegisterSQLiteNormalize(db)` adds `normalize()` function for better search

### Security Features
- **XSS prevention**: All HTML attributes automatically escaped via `html.EscapeString`
- **JS injection prevention**: JavaScript strings properly escaped in generated code
- **CSP**: Content Security Policy support via `ctx.SetDefaultCSP()` or custom policies
- **Input validation**: Comprehensive bounds checking, character validation, length limits
- **Field access safety**: `PathValue()` validates field paths to prevent unsafe reflection access
- **CAPTCHA**: Three variants (`ui.Captcha`, `ui.Captcha2`, `ui.Captcha3`) with rate-limiting and session management

### Theming & Styling
- **Dark mode**: Built-in dark theme support with automatic system preference detection
- **Theme switcher**: `ui.ThemeSwitcher(css)` renders compact toggle cycling System → Light → Dark
- **CSS constants**: Predefined color schemes (`ui.Blue`, `ui.Green`, `ui.Red`, etc.) and component styles
- **Responsive**: Tailwind CSS classes enable responsive design patterns

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

### Layout & HTML Primitives
- **Containers**: `ui.Div`, `ui.Span`, `ui.P` for block/inline elements
- **Links**: `ui.A(class, attrs...)(content)` with `ui.Href(url)` attribute helper
- **Forms**: `ui.Form(method, action)(children...)` with automatic CSRF handling
- **Lists**: `ui.List`, `ui.ListItem` for semantic lists
- **Media**: `ui.Img(attrs...)`, `ui.Canvas(attrs...)` for images and canvas
- **Control flow**: `ui.Map(slice, func)`, `ui.For(from, to, func)`, `ui.If(cond, func)`, `ui.Iff(cond)`, `ui.Or(cond, trueFunc, falseFunc)`
- **Content**: `ui.Markdown(classes...)(content)` for markdown rendering, `ui.Script(code...)` for inline scripts
- **Utilities**: `ui.Flex1` for flex-grow spacer, `ui.Space` for non-breaking space

### Buttons
- **Builder**: `ui.Button()` returns fluent builder
- **Types**: `.Submit()`, `.Reset()` for form buttons
- **Colors**: `.Color(ui.Blue|ui.Green|ui.Red|ui.Yellow|ui.Purple|ui.Gray|ui.White)` with outline variants (`ui.BlueOutline`, etc.)
- **Sizes**: `.Size(ui.XS|ui.SM|ui.MD|ui.ST|ui.LG|ui.XL)`
- **Actions**: `.Click(ctx.Call(...))` for click handlers
- **Navigation**: `.Href(url)` for link-style buttons
- **Styling**: `.Class(classes...)` for custom CSS classes

### Labels
- **Builder**: `ui.Label(target).Required(true).Class(...)` 
- **Features**: Automatic required asterisk, accessible `for` attribute binding
- **Styling**: Customizable classes for label and required indicator

### Input Components (All support fluent API)
- **Text inputs**: 
  - `ui.IText(name, data?)` - standard text input
  - `ui.IEmail(name, data?)` - email with validation and autocomplete
  - `ui.IPhone(name, data?)` - phone with pattern validation (+country code)
  - `ui.IPassword(name, data?)` - password input
- **Numeric**: `ui.INumber(name, data?)` with `.Numbers(min, max, step)` for range validation
- **Dates**: 
  - `ui.IDate(name, data?)` - date picker with `.Dates(min, max)` for range
  - `ui.ITime(name, data?)` - time picker
  - `ui.IDateTime(name, data?)` - datetime picker
- **Text area**: `ui.IArea(name, data?)` with `.Rows(count)` for multi-line text
- **Select**: `ui.ISelect(name, options, data?)` with `ui.MakeOptions([]string)` helper
- **Checkbox**: `ui.ICheckbox(name, data?)` for boolean inputs
- **Radio**: 
  - `ui.IRadio(name, value, data?)` for single radio button
  - `ui.IRadioButtons(name, options, data?)` for radio group
- **Display**: `ui.IValue(attrs...)` for read-only value display

**Common input methods** (available on all inputs):
- `.Placeholder(text)`, `.Required(bool?)`, `.Disabled(bool?)`, `.Readonly(bool?)`
- `.Class(classes...)`, `.ClassInput(classes...)`, `.ClassLabel(classes...)`
- `.Value(text)`, `.Pattern(regex)`, `.Autocomplete(value)`
- `.Change(action)`, `.Click(action)` for event handlers
- `.Error(&err)` for validation error binding
- `.If(visible)` for conditional rendering
- `.Format(format)` for value formatting

### Tables
- **Simple table**: `ui.SimpleTable(cols, classes...)` with methods:
  - `.Field(func(item) string, classes...)` for data cells
  - `.Head(text, classes...)` for header cells (auto-escaped)
  - `.HeadHTML(html, classes...)` for HTML headers
  - `.FieldText(func(item) string, classes...)` for safe text cells
  - `.Empty(message)` for empty state
  - `.Class(rowClasses...)` for row styling
  - `.Attr(attrs...)` for row attributes (supports `colspan`)
- **Advanced**: `ui.Table` for more complex scenarios with typed rows

### Icons
- **FontAwesome**: `ui.Icon`, `ui.Icon2`, `ui.Icon3`, `ui.Icon4` helpers for FontAwesome icon markup
- **Positioning**: Variants support different positioning classes

### Skeletons (Loading Placeholders)
- **Target-based**: `target.Skeleton(kind)` where `kind` is `ui.SkeletonList`, `ui.SkeletonComponent`, `ui.SkeletonPage`, `ui.SkeletonForm`, or default
- **Standalone**: `ui.SkeletonDefault()`, `ui.SkeletonListN(count)`, `ui.SkeletonComponentBlock()`, `ui.SkeletonPageBlock()`, `ui.SkeletonFormBlock()`
- **Usage**: Return skeleton immediately, then replace with `ctx.Patch` when data ready

### CAPTCHA Components
- **ui.Captcha**: Google reCAPTCHA integration - `ui.Captcha(siteKey, solvedHTML)` renders widget, swaps in HTML when solved
- **ui.Captcha2**: Image-based challenge entirely in Go - `ui.Captcha2(onValidated)` with configurable:
  - `.SessionField(name)`, `.ClientVerifiedField(name)`, `.AnswerField(name)`
  - `.Attempts(max)`, `.Lifetime(duration)`
  - `.ValidateRequest(req)` for server-side validation
- **ui.Captcha3**: Draggable tile ordering challenge - `ui.Captcha3(onValidated)` with:
  - `.Count(tiles)` for number of tiles
  - `.ArrangementField(name)` for submission field
  - Same session/validation config as Captcha2
- **Session storage**: In-memory by default; can be swapped for Redis/DB backing

### Query/Collate (Data Management)
- **Initialization**: `collate := ui.Collate[Type](&ui.TQuery{Limit: 10, Order: "field asc"})`
- **Field definition**: `ui.TField{DB: "column", Field: "StructField", Text: "Label", As: ui.SELECT|ui.BOOL|ui.DATES|ui.ZERO_DATE|ui.NOT_ZERO_DATE, Options: ui.MakeOptions([]string)}`
- **Configuration**:
  - `.Search(fields...)` - fields searchable via search box
  - `.Sort(fields...)` - fields available for sorting
  - `.Filter(fields...)` - fields available as filters
  - `.Excel(fields...)` - fields included in Excel export
- **Rendering**: `collate.Render(ctx, db)` returns complete UI (search, sort, filters, pagination, export button)
- **Row rendering**: `.Row(func(item *Type, index int) string)` defines how each row displays
- **Results**: `collate.Load(query)` returns `*TCollateResult[Type]` with `.Items`, `.Total`, `.Page`, `.Pages`
- **Helpers**: `ui.Empty(result)`, `ui.Filtering(ctx, collate, query)`, `ui.Searching(ctx, collate, query)`, `ui.Sorting(ctx, collate, query)`, `ui.Paging(ctx, collate, result)`
- **Search normalization**: `ui.NormalizeForSearch(text)` and `ui.RegisterSQLiteNormalize(db)` for accent-insensitive search

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

## Context Helpers (Detailed API)

### Request & Session Management
- **HTTP access**: `ctx.Request` (*http.Request), `ctx.Response` (http.ResponseWriter)
- **Client info**: `ctx.IP()` returns client IP address
- **Session storage**: `ctx.Session(db *gorm.DB, name string) *TSession` returns session helper:
  - `.Load(data any)` - loads JSON data from session into struct
  - `.Save(output any)` - saves struct as JSON to session
  - Sessions stored in `_session` table (auto-created by GORM)

### Body Parsing & Validation
- **Hydration**: `ctx.Body(&payload)` parses form/JSON into struct with automatic validation
- **Safety**: Automatic validation of:
  - Field count (max 1000)
  - Field name length (max 256 chars) and character safety
  - Field value length (max 1MB)
  - Body size (max 10MB)
  - Numeric bounds checking
- **Path access**: `ui.PathValue(obj, "field.path")` safely accesses nested struct fields with validation

### Server Actions & Events
- **Action creation**: `ctx.Call(handler Callable, payload ...any)` creates action descriptor
- **Swap strategies**:
  - `.Render(target)` - replaces innerHTML of target element
  - `.Replace(target)` - replaces entire target element
  - `.Append(target)` - appends HTML to end of target
  - `.Prepend(target)` - prepends HTML to start of target
  - `.None()` - executes without DOM update (fire-and-forget)
- **Shortcuts**: 
  - `ctx.Click(handler, payload?)` - click action with default swap
  - `ctx.Submit(handler, payload?)` - form submit action
  - `ctx.Send(handler, payload?)` - generic send action
  - `ctx.Post(handler, payload?)` - POST action

### Navigation
- **SPA-like**: `ctx.Load(href)` returns attributes for client-side navigation (no full page reload)
- **Reload**: `ctx.Reload()` returns JavaScript to reload current page
- **Redirect**: `ctx.Redirect(url)` returns JavaScript to navigate to URL

### Real-Time Updates (WebSocket)
- **Patch**: `ctx.Patch(targetSwap TargetSwap, html string, clear ...func())` pushes update to all connected clients
- **Target swap**: Use `target.Render()`, `target.Replace()`, `target.Append()`, `target.Prepend()` to create TargetSwap
- **Clear callbacks**: Optional `clear` functions called if target element not found in DOM
- **Connection management**: Automatic reconnect, heartbeat (25s ping), stale connection cleanup (75s timeout)

### User Feedback (Toasts)
- **Success**: `ctx.Success(message string)` - green success toast
- **Error**: `ctx.Error(message string)` - red error toast
- **Info**: `ctx.Info(message string)` - blue info toast
- **Error with reload**: `ctx.ErrorReload(message string)` - error toast with reload button

### File Downloads
- **Stream**: `ctx.DownloadAs(filename string, contentType string, content []byte)` streams file to client

### Internationalization
- **Translation**: `ctx.Translate(key string) string` resolves i18n strings (locale set in `ui.MakeApp(locale)`)

### Security Headers
- **CSP**: `ctx.SetDefaultCSP()` sets strict Content Security Policy defaults
- **Custom CSP**: `ctx.SetCSP(policy string)` sets custom CSP policy
- **Security headers**: `ctx.SetSecurityHeaders()` or `ctx.SetCustomSecurityHeaders(headers map[string]string)` for broader security

### App-Level Configuration
- **Routing**: `app.Page(path, layout, handler)` registers page route
- **Actions**: `app.Action(path, handler)` registers action endpoint
- **Callables**: `app.Callable(path, handler)` registers callable endpoint
- **Static assets**: `app.Assets(fs embed.FS, path string)` serves static files
- **Favicon**: `app.Favicon(fs embed.FS, path string, cacheDuration time.Duration)` serves favicon with caching
- **Auto-restart**: `app.AutoRestart(enabled bool)` enables file watching and auto-rebuild
- **Session sweeper**: `app.StartSweeper(interval time.Duration)` cleans up expired sessions
- **Server start**: `app.Listen(addr string)` starts HTTP/WebSocket server

## Common Patterns & Examples

### Basic Page with Server Action
```go
app.Page("/", layout("Home", func(ctx *ui.Context) string {
    hello := func(ctx *ui.Context) string {
        ctx.Success("Hello!")
        return ""
    }
    return ui.Div("p-4")(
        ui.Button().Color(ui.Blue).Click(ctx.Call(hello).None()).Render("Say hello"),
    )
}))
```

### Form with Validation
```go
type LoginForm struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

func Login(ctx *ui.Context) string {
    if ctx.Request.Method == http.MethodPost {
        var form LoginForm
        if err := ctx.Body(&form); err != nil {
            ctx.Error("Invalid form data")
            return ""
        }
        // Process login...
        ctx.Success("Logged in!")
    }
    var form LoginForm
    return ui.Form("post", "/login")(
        ui.IEmail("Email", &form).Required().Render("Email"),
        ui.IPassword("Password", &form).Required().Render("Password"),
        ui.Button().Color(ui.Blue).Submit().Render("Login"),
    )
}
```

### Deferred Fragment with Skeleton
```go
func Deferred(ctx *ui.Context) string {
    target := ui.Target()
    
    // Return skeleton immediately
    go func() {
        time.Sleep(2 * time.Second)
        html := ui.Div("")(target)("Content loaded!")
        ctx.Patch(target.Replace(), html)
    }()
    
    return target.Skeleton(ui.SkeletonComponent)
}
```

### Data Table with Collate
```go
type Person struct {
    ID   uint
    Name string
    Age  int
}

collate := ui.Collate[Person](&ui.TQuery{Limit: 10})
nameField := ui.TField{DB: "name", Field: "Name", Text: "Name"}
collate.Search(nameField)
collate.Sort(nameField)
collate.Row(func(p *Person, _ int) string {
    return ui.Div("")(fmt.Sprintf("%s (%d)", p.Name, p.Age))
})

content := collate.Render(ctx, db)
```

### WebSocket Real-Time Updates
```go
func Clock(ctx *ui.Context) string {
    target := ui.Target()
    
    go func() {
        ticker := time.NewTicker(time.Second)
        for {
            select {
            case <-ticker.C:
                html := ui.Div("")(target)(time.Now().Format(time.RFC3339))
                ctx.Patch(target.Render(), html)
            }
        }
    }()
    
    return ui.Div("")(target)("Loading...")
}
```

## Security Best Practices

1. **Always use text-safe methods** for user input:
   - Tables: `table.FieldText(func(item) string { return item.UserInput })`
   - Headers: `table.Head("User Text")` (auto-escaped)

2. **Set CSP headers** in handlers:
   ```go
   ctx.SetDefaultCSP()
   ```

3. **Validate input server-side** using `go-playground/validator` tags

4. **Use PathValue** for safe struct field access (prevents unsafe reflection)

5. **Avoid raw HTML** in user-controlled content; use component methods

6. **Rate limit** sensitive operations using CAPTCHA components

## Performance Considerations

- **Server-rendered**: All HTML generated on server; minimal client JS
- **Partial updates**: Only update changed DOM elements via WebSocket patches
- **Skeleton loading**: Show placeholders immediately, patch real content when ready
- **Lighthouse scores**: Example app achieves 97 Performance, 100 Accessibility, 100 Best Practices, 90 SEO
- **WebSocket efficiency**: Heartbeat keeps connections alive; stale connections auto-closed

## References
- **README.md**: Comprehensive walkthroughs, security notes, deferred fragment examples, component documentation
- **examples/main.go**: Complete example application demonstrating all features
- **examples/pages/**: Individual page examples for each component and pattern
- **docs/lighthouse-scores.png**: Performance profile visualization
- **ui/**: Core library source code with inline documentation
