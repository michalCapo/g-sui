# g-sui

Go server-rendered UI framework with real-time WebSocket patches.

g-sui renders HTML on the server, sends actions over WebSocket, and updates specific DOM targets without full page reloads.

## Documentation

- Full API docs: [`docs/DOCUMENTATION.md`](docs/DOCUMENTATION.md)
- Assistant skill docs: [`docs/skills/SKILL.md`](docs/skills/SKILL.md)

## Features

- Server-rendered HTML components with a small helper DSL
- Lightweight interactivity via server actions (`Click`, `Submit`, `Send`)
- Partial updates: re-render, replace, append, or prepend only the target
- Smooth navigation with background loading and delayed loader (50ms threshold) via `ctx.Load()`
- **Parameterized routes** with path parameters (`/user/{id}`) and query parameters (`?name=Smith`) - works seamlessly with SPA navigation
- **Custom HTTP handlers** for REST APIs (`app.GET()`, `app.POST()`, etc.) - mix g-sui pages with standard HTTP endpoints
- **Custom server configuration** via `app.Handler()` - wrap with middleware or integrate with existing HTTP servers
- Built-in PWA support with manifest and service worker generation (`app.PWA()`)
- Deferred fragments with skeletons via WebSocket patches (`ctx.Patch` + skeleton helpers)
- Query/Collate helper for data UIs: search, sort, filters, paging, and XLS export (works with `gorm`)
- Form helpers with validation (uses `go-playground/validator`)
- A small set of UI inputs (text, email, phone, password, number, date/time/datetime, select, checkbox, radio, radio cards, textarea), buttons, tables, icons
- Toast messages: `Success`, `Error`, `Info`, and an error toast with a Reload button
- Built-in live status via WebSocket (`/__ws`) with an offline banner, automatic reconnect, and auto-reload on reconnect
- Built-in dark mode with a tiny theme switcher (`ui.ThemeSwitcher`) cycling System → Light → Dark with proper icon alignment
- Reverse proxy package (`proxy`) for HTTP and WebSocket forwarding with automatic URL rewriting
- Optional dev autorestart (`app.AutoRestart(true)`) to rebuild and restart on changes

## Lighthouse snapshot

![Lighthouse scores for the example app](docs/lighthouse-scores.png)

The bundled example app scores 97 for Performance, 100 for Accessibility, 100 for Best Practices, and 90 for SEO when audited with Lighthouse. These scores come from a local run against the default demo and showcase how the server-rendered approach keeps the experience fast and accessible out of the box.

## Install

```bash
go get github.com/michalCapo/g-sui
```

Go 1.21+ recommended (module currently targets Go 1.23 toolchain).

## Quick start

```go
package main

import (
    "github.com/michalCapo/g-sui/ui"
)

func main() {
    app := ui.MakeApp("en")
    
    app.Page("/", "Home", func(ctx *ui.Context) string {
        return app.HTML("Hello World", "bg-gray-100", 
            ui.Div("p-8")(
                ui.Div("text-2xl font-bold")("Hello World"),
            ),
        )
    })
    
    app.Listen(":8080")
}
```

Run and open http://localhost:8080

## Examples

This repo ships an example app showcasing most components and patterns.

Run the examples server:

```bash
go run examples/main.go
```

The examples include:
- Core UI components (alerts, badges, cards, tabs, accordion, dropdowns, tooltips)
- Forms and inputs with validation
- Tables, icons, and markdown rendering
- Route params and navigation examples
- WebSocket patches, deferred loading, and reverse proxy demo

## Progressive Web App (PWA)

g-sui provides built-in support for Progressive Web Apps, allowing your application to be installed on mobile and desktop devices. Enable it by providing a config:

```go
app.PWA(ui.PWAConfig{
    Name:                  "My App",
    ShortName:             "App",
    Description:           "My awesome g-sui app",
    ThemeColor:            "#1d4ed8",
    BackgroundColor:       "#ffffff",
    GenerateServiceWorker: true,
    CacheAssets:           []string{"/assets/app.css", "/assets/app.js"}, // Assets to pre-cache
    OfflinePage:           "/offline",                                     // Fallback when offline
})
```

This automatically:
- Generates and serves `/manifest.webmanifest` with proper JSON formatting
- Generates and serves `/sw.js` (Service Worker) with smart caching
- Adds necessary meta tags and manifest link to the `<head>`
- Registers the service worker in the browser automatically

The service worker provides:
- **Network-first for pages**: Always fresh content from server, cache as offline fallback
- **Cache-first for assets**: Fast loading for files in `CacheAssets`
- **Auto-versioning**: New cache on each server restart, old caches auto-cleaned
- **Immediate activation**: `skipWaiting()` + `clients.claim()` for instant updates

## Server actions and partial updates

Attach server actions via:

- `ctx.Call(fn).Render(target)` – replace inner HTML of `target`
- `ctx.Call(fn).Replace(target)` – replace the element itself
- `ctx.Call(fn).Append(target)` – insert HTML at the end of the target
- `ctx.Call(fn).Prepend(target)` – insert HTML at the beginning of the target
- `ctx.Call(fn).None()` – fire and forget (no swap)

On the server, an action has the signature `func(*ui.Context) string` and returns HTML (or an empty string if not swapping anything).

Forms can use `ctx.Submit(fn).Render/Replace/None()` and `ctx.Body(out)` to bind values to a struct.

## Messages and errors

- Toasts: `ctx.Success(msg)`, `ctx.Error(msg)`, `ctx.Info(msg)`
- Error toast with Reload button: `ctx.ErrorReload(msg)`
- Page title: `ctx.Title(title)` - Update the page title dynamically
- Built‑in client handlers display a compact error panel for failed fetches (HTTP 500 etc.) with a Reload button.
- Server panics are recovered and surface as an error toast with Reload.

## Components (selection)

### UI Components
- **Alert**: `ui.Alert().Message(text).Variant("success").Title("Title").Dismissible(true).Persist("key").Render()` - Dismissible notification banners with dark mode, optional title, and localStorage persistence
- **Badge**: `ui.Badge().Text("3").Color("red").Dot().Size("lg").Icon(html).Square().Render()` - Status indicators with dot, icon, and size variants
- **Card**: `ui.Card().Header(h).Body(b).Footer(f).Image(src,alt).Variant(ui.CardGlass).Hover(true).Compact(true).Render()` - Content containers with 4 variants (shadowed, bordered, flat, glass), images, and hover effects
- **Progress**: `ui.ProgressBar().Value(75).Gradient("#3b82f6","#8b5cf6").Striped(true).Animated(true).Indeterminate(true).Label("Loading").LabelPosition("outside").Render()` - Progress indicators with gradients, labels, and indeterminate mode
- **Step Progress**: `ui.StepProgress(2, 5).Color("bg-blue-500").Size("md").Render()` - Step progress indicator showing "Step X of Y" with progress bar
- **Tooltip**: `ui.Tooltip().Content(text).Position("top").Variant("dark").Delay(500).Render(element)` - Hover tooltips with 4 positions, 6 variants, and configurable delay
- **Tabs**: `ui.Tabs().Tab(label, content, icon).Active(0).Style("boxed").Render()` - Tabbed content with 4 styles (underline, pills, boxed, vertical) and icon support
- **Accordion**: `ui.Accordion().Item(title, content, open).Variant(ui.AccordionSeparated).Multiple(true).Render()` - Collapsible sections with 3 variants (bordered, ghost, separated) and multiple open support
- **Dropdown**: `ui.Dropdown().Trigger(html).Item(label, onclick, icon).Header("Group").Divider().Danger("Delete", onclick).Position("bottom-right").Render()` - Context menus with headers, dividers, danger items, and 4 positions

### Form Components
- Buttons: `ui.Button().Color(...).Size(...).Class(...).Href(...).Submit().Reset().Click(...)`
- Inputs: `ui.IText`, `ui.IEmail`, `ui.IPhone`, `ui.IPassword`, `ui.INumber`, `ui.IDate`, `ui.ITime`, `ui.IDateTime`, `ui.IArea`, `ui.ISelect`, `ui.ICheckbox`, `ui.IRadio`, `ui.IRadioButtons`, `ui.IRadioDiv`
- Table: `ui.SimpleTable(cols, classes...)` with `Field`, `Empty`, `Class`, `Attr` (supports `colspan`)
- Icons: `ui.Icon`, `ui.Icon2`, `ui.Icon3`, `ui.Icon4`
- Markdown: `ui.Markdown(classes...)(content)`

Refer to the `examples/` directory for practical usage and composition patterns.

## Theme & Dark Mode

- Built-in dark theme overrides load with `ui.MakeApp`. Use `ui.ThemeSwitcher("")` to render a compact toggle that cycles System → Light → Dark.
- The theme switcher includes properly aligned icons and smooth transitions between system, light, and dark modes
- Typical placement is in your layout's top bar:

```go
nav := ui.Div("bg-white shadow mb-6")(
    ui.Div("max-w-5xl mx-auto px-4 py-2 flex items-center gap-2")(
        // ... your nav links ...
        ui.Flex1,
        ui.ThemeSwitcher(""),  // Shows auto/light/dark icons with proper alignment
    ),
)
```

## Security

g-sui includes built-in security measures to prevent Cross-Site Scripting (XSS) and other attacks:

### Server-side Protections

- **HTML Attribute Escaping**: All HTML attributes (values, classes, IDs, etc.) are automatically escaped using `html.EscapeString`
- **JavaScript Escaping**: JavaScript code generation (URLs, IDs in event handlers) uses proper escaping to prevent injection
- **Safe Table Methods**: Use `HeadHTML()` and `FieldText()` for explicit control over HTML vs. text content

### Client-side Protections

- **Content Security Policy**: Use `ctx.SetDefaultCSP()` or `ctx.SetCSP(policy)` to set restrictive CSP headers:

```go
func handler(ctx *ui.Context) string {
    ctx.SetDefaultCSP() // Sets secure defaults
    // Your page content...
    return ui.Div("")("Your content")
}
```

## Claude Code Skills

[g-sui](https://github.com/michalCapo/g-sui) includes **Claude Code skills** to help Claude (and other LLMs) understand the framework better. These skills provide comprehensive documentation that Claude can reference when answering questions or generating code.

### Quick Install

Available across all your projects
```bash
mkdir -p ~/.claude/skills/g-sui && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/SKILL.md -o ~/.claude/skills/g-sui/SKILL.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/CORE.md -o ~/.claude/skills/g-sui/CORE.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/COMPONENTS.md -o ~/.claude/skills/g-sui/COMPONENTS.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/DATA.md -o ~/.claude/skills/g-sui/DATA.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/SERVER.md -o ~/.claude/skills/g-sui/SERVER.md && curl -sL https://raw.githubusercontent.com/michalCapo/g-sui/main/docs/skills/PATTERNS.md -o ~/.claude/skills/g-sui/PATTERNS.md
```

Then restart Claude Code to load the skills.

### What's Included

| Skill | Description |
|-------|-------------|
| **SKILL.md** | Main entry point with quick start and navigation |
| **CORE.md** | Architecture, Context API, Actions, Targets, WebSocket patches |
| **COMPONENTS.md** | Buttons, inputs, forms, tables, alerts, cards, tabs, dropdowns, etc. |
| **DATA.md** | Data collation (TQuery/TCollate), search, sort, filter, pagination, Excel |
| **SERVER.md** | App initialization, routes, WebSocket, PWA, assets |
| **PATTERNS.md** | Testing, validation, security, state management patterns |

## License

MIT
