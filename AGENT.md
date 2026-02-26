# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**g-sui** is a server-rendered UI framework for Go that enables building interactive web applications without client-side JavaScript frameworks. All HTML generation, business logic, and state management occur on the server. Interactivity is achieved through server actions (HTPX-inspired) and WebSocket patches for real-time updates.

## Development Commands

```bash
# Run the example application
go run examples/main.go

# Build the project
go build

# Run all tests
go test ./...

# Run tests for a specific package
go test ./ui/...

# Run a specific test
go test -run TestName ./ui/

# Module tidy
go mod tidy

# Create and push a new version tag
./deploy
```

### Versioning and Releases

The project uses semantic versioning starting at `v0.100`:
- Run `./deploy` to automatically create and push a new version tag
- The script increments the minor version (e.g., `v0.100` → `v0.101` → `v0.102`)
- Requires a clean working tree (no uncommitted changes)
- Creates an annotated git tag and pushes it to the remote repository

## High-Level Architecture

### Core Philosophy
### Key Concepts

1. **Server-Centric Rendering**: All HTML is generated server-side as strings
2. **String-Based Components**: Components are Go functions returning HTML strings
3. **Action-Based Interactivity**: User interactions trigger server handlers that return partial HTML updates
4. **WebSocket-Enhanced**: Real-time updates via `/__ws` endpoint for server-initiated DOM patches
5. **Multitenant Support**: Multiple app instances can run with different URL prefixes
6. **Automatic Compression**: Gzip compression enabled by default for supported clients

### Key Concepts

1. **Callable Type**: All handlers have signature `func(*ui.Context) string` returning HTML
2. **Targets & Actions**: `ui.Target()` generates unique IDs for DOM elements that can be updated
3. **Swap Methods**: `Render()` (innerHTML), `Replace()` (entire element), `Append()`, `Prepend()`, `None()`
4. **Context**: Request-scoped context carrying request, response writer, and state

### Package Structure

```
ui/
├── ui.go              # Core types, HTML DSL, utility functions
├── ui.server.go       # HTTP server, WebSocket handler, app initialization
├── ui.button.go       # Button component
├── ui.form.go         # Form handling and validation
├── ui.input.go        # Input components (text, email, password, etc.)
├── ui.table.go        # Table component
├── ui.data.go         # Data collation (search, sort, filter, pagination, XLS export)
├── ui.captcha*.go     # CAPTCHA components (Captcha2, Captcha3)
├── ui.alert.go        # Alert component
├── ui.badge.go        # Badge component
├── ui.card.go         # Card component
├── ui.tabs.go         # Tabs component
├── ui.accordion.go    # Accordion component
├── ui.dropdown.go     # Dropdown component
├── ui.tooltip.go      # Tooltip component
├── ui.progress.go     # Progress bar component
├── ui.step.go         # Step/wizard component
└── *_test.go          # Comprehensive test coverage

proxy/
├── proxy.go           # Reverse proxy with HTTP/WebSocket forwarding and URL rewriting
└── *_test.go          # Comprehensive test coverage
```

### App Initialization Pattern

```go
app := ui.MakeApp("en")              // Create app with locale
app.Page("/path", "Title", handler)   // Register page route

// Optional: Define a global layout (wraps all pages)
// Omit this to render pages directly without a wrapper
app.Layout(func(ctx *ui.Context) string {
    return ui.Div("layout", ui.Attr{})(
        ui.Div("header", ui.Attr{})("Header"),
        ui.Div("", ui.Attr{ID: "__content__"})(),  // Content slot - uses app.ContentID
        ui.Div("footer", ui.Attr{})("Footer"),
    )
})

app.Favicon(embedFS, "path", 24*time.Hour)
app.Assets(embedFS, "assets/", 24*time.Hour)
app.AutoRestart(true)                 // Dev: rebuild on file changes
app.PWA(ui.PWAConfig{...})           // Enable PWA

// Custom HTTP handlers (REST APIs)
app.Custom("GET", "/api/health", healthHandler)  // Register custom handler
app.GET("/api/users", getUsersHandler)           // Shorthand for GET
app.POST("/api/users", createUserHandler)        // Shorthand for POST

app.Listen(":8080")                   // Start server (also starts WebSocket, enables gzip)

// OR use custom server configuration
handler := app.Handler()              // Get http.Handler for custom setups (auto-starts sweeper and WS)
server := &http.Server{Addr: ":8080", Handler: handler}

// OR mount multiple apps with different prefixes (multitenant)
mainMux := http.NewServeMux()
adminApp.Mount("/admin", mainMux)     // Admin app at /admin/*
publicApp.Mount("", mainMux)          // Public app at root
http.ListenAndServe(":8080", mainMux)
```

### Action System

Actions are attached via `ctx.Call()`, `ctx.Submit()`, `ctx.Click()`, or `ctx.Send()`:

```go
// Click handlers
ctx.Call(handler).Render(target)    // Replace innerHTML
ctx.Call(handler).Replace(target)   // Replace entire element
ctx.Call(handler).Append(target)    // Append to element
ctx.Call(handler).Prepend(target)   // Prepend to element
ctx.Call(handler).None()            // Fire-and-forget

// Form submission
ctx.Submit(handler).Replace(target)

// Direct element click
ui.Button().Click(ctx.Call(handler).Render(target)).Render("Click me")
```

### Multitenant / Path-Prefix Mounting

Run multiple isolated g-sui apps with different URL prefixes:

```go
// Create apps
adminApp := ui.MakeApp("en")
publicApp := ui.MakeApp("en")

// Register routes (relative to mount point)
adminApp.Page("/", "Dashboard", adminDashboard)      // Will be at /admin/
publicApp.Page("/", "Home", homeHandler)             // Will be at /

// Mount on shared mux
mainMux := http.NewServeMux()
adminApp.Mount("/admin", mainMux)
publicApp.Mount("", mainMux)

http.ListenAndServe(":8080", mainMux)
```

Each app has isolated sessions, routes, and state. WebSocket paths are automatically adjusted.

### State Management

State is passed through payload structs in actions:

```go
type Counter struct { Count int }

func (c *Counter) Increment(ctx *ui.Context) string {
    ctx.Body(c)  // Restore state from request
    c.Count++
    return c.Render(ctx)
}

// Usage: ctx.Call(c.Increment, c).Replace(target)
```

### Data Collation (TQuery/TCollate)

For data-centric pages, `ui.Collate[T]()` provides search, sort, filter, pagination, and Excel export:

```go
collate := ui.Collate[Model](init)
collate.Search(field1, field2)
collate.Sort(field1)
collate.Filter(field1)
collate.Excel(field1, field2)
collate.Row(func(item *Model, index int) string { return ... })

content := collate.Render(ctx, db)
```

### Security Model

- **XSS Protection**: All HTML attributes escaped via `escapeAttr()`; JavaScript via `escapeJS()`
- **CSP Headers**: Use `ctx.SetDefaultCSP()` or `ctx.SetCSP(policy)`
- **Validation**: `go-playground/validator` for form validation
- **Safe Methods**: Use `table.Head()` for escaped text, `table.HeadHTML()` for raw HTML

## Styling

- **Tailwind CSS**: Utility-first CSS loaded via CDN in development
- **Dark Mode**: Built-in dark theme overrides; use `ui.ThemeSwitcher("")` for toggle
  - Theme switcher displays brightness_auto icon for system/auto mode, light_mode for light, dark_mode for dark
  - Icons are properly centered with flexbox alignment
  - Includes cursor-pointer class for better UX
- **CSS Constants**: Pre-defined color constants (Blue, Green, Red, etc.) and size constants (XS, SM, MD, ST, LG, XL)
- **Icons**: Material Icons (with Google Fonts) are automatically included in all apps
  - Use `ui.Icon("icon_name")` for single icons (renders as `<div class="material-icons w-8">icon_name</div>`)
  - Font Awesome names are automatically converted (e.g., `fa-home` → `home`)
  - Customize size and style: `ui.Icon("home", ui.Attr{Class: "text-lg text-blue-500"})`
  - Helper functions: `ui.IconLeft()`, `ui.IconRight()`, `ui.IconStart()`, `ui.IconEnd()`

## Reverse Proxy Package

The `proxy` package provides HTTP and WebSocket reverse proxying with automatic URL rewriting:

```go
import "github.com/michalCapo/g-sui/proxy"

p, err := proxy.New(proxy.Config{
    ProxyPort:  "8640",
    TargetPort: "8642",
    TargetHost: "localhost",
    Logger:     log.Default(),
})
if err != nil {
    log.Fatal(err)
}

// Start blocks while serving requests
if err := p.Start(); err != nil {
    log.Fatal(err)
}

// Later, stop gracefully
p.Stop()
```

**Features**:
- HTTP request forwarding with transparent proxying
- WebSocket connection tunneling
- Automatic URL rewriting in HTML, CSS, JavaScript, and JSON responses
- Port reference rewriting
- Optional debug logging for WebSocket messages
- Graceful shutdown support

**Use cases**:
- Development: frontend on one port, backend on another, accessed through single proxy
- Testing: swap proxy targets without changing frontend URLs
- Staging: transparent port forwarding for distributed services

See `examples/pages/proxy.go` for a complete UI example with start/stop controls and real-time status updates.

## Dependencies

- `github.com/fsnotify/fsnotify` - File watching for auto-restart
- `github.com/go-playground/validator/v10` - Form validation
- `github.com/mattn/go-sqlite3` - SQLite driver
- `gorm.io/gorm` + `gorm.io/driver/sqlite` - ORM for database operations
- `github.com/xuri/excelize/v2` - Excel export for data collation
- `github.com/yuin/goldmark` - Markdown rendering

## Documentation

Comprehensive documentation is available in `docs/DOCUMENTATION.md`, which includes:
- Complete API reference for all components
- Architecture documentation
- Security best practices
- Validation tags reference
- Testing patterns

The `examples/` directory contains working implementations of most components and patterns.

## Claude Code Skills

Project-specific **Claude Code skills** are available in `docs/skills/` to help Claude (and LLMs in general) understand and work with g-sui more effectively.

### Installing Skills

**Personal (recommended):** Available across all your projects
```bash
mkdir -p ~/.claude/skills/g-sui && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/SKILL.md -o ~/.claude/skills/g-sui/SKILL.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/CORE.md -o ~/.claude/skills/g-sui/CORE.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/COMPONENTS.md -o ~/.claude/skills/g-sui/COMPONENTS.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/DATA.md -o ~/.claude/skills/g-sui/DATA.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/SERVER.md -o ~/.claude/skills/g-sui/SERVER.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/PATTERNS.md -o ~/.claude/skills/g-sui/PATTERNS.md
```

**Project-local:** Shared with your team via git
```bash
mkdir -p .claude/skills/g-sui && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/SKILL.md -o .claude/skills/g-sui/SKILL.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/CORE.md -o .claude/skills/g-sui/CORE.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/COMPONENTS.md -o .claude/skills/g-sui/COMPONENTS.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/DATA.md -o .claude/skills/g-sui/DATA.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/SERVER.md -o .claude/skills/g-sui/SERVER.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/PATTERNS.md -o .claude/skills/g-sui/PATTERNS.md
```

Then restart Claude Code to load the skills.

### Available Skills

| File | Description |
|------|-------------|
| `SKILL.md` | Main entry point with navigation |
| `CORE.md` | Architecture, Context, Actions, Targets, WebSocket patches |
| `COMPONENTS.md` | Buttons, inputs, forms, tables, alerts, cards, tabs, etc. |
| `DATA.md` | Data collation, search, sort, filter, pagination, Excel export |
| `SERVER.md` | App setup, routes, WebSocket, PWA, assets |
| `PATTERNS.md` | Testing, validation, security, state management |

### Using Skills

Once installed, Claude will automatically use these skills when you:
- Mention "g-sui" in your prompts
- Ask about server-rendered UI, Go UI frameworks
- Work with forms, data tables, or WebSocket patches
