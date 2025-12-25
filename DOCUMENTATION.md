# g-sui Architecture Documentation

This document describes the internal architecture, design patterns, and implementation details of g-sui.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Principles](#architecture-principles)
3. [Package Structure](#package-structure)
4. [Component Reference](#component-reference)
5. [Form System](#form-system)
6. [Core Components](#core-components)
7. [Context API Reference](#context-api-reference)
8. [Request Lifecycle](#request-lifecycle)
9. [State Management](#state-management)
10. [Action System](#action-system)
11. [WebSocket Communication](#websocket-communication)
12. [Data Collation (TCollate)](#data-collation-tcollate)
13. [CAPTCHA Components](#captcha-components)
14. [Security Model](#security-model)
15. [Extension Points](#extension-points)
16. [Performance Considerations](#performance-considerations)
17. [Testing Patterns](#testing-patterns)
18. [Future Considerations](#future-considerations)

---

## Overview

g-sui is a server-rendered UI framework for Go that enables building interactive web applications without client-side JavaScript frameworks. The architecture follows these key principles:

- **Server-Centric**: All HTML generation, business logic, and state management occur on the server
- **String-Based Rendering**: Components are plain Go functions that return HTML strings
- **HTPX-Inspired Actions**: User interactions trigger server actions that return partial HTML updates
- **WebSocket-Enhanced**: Real-time updates and server-initiated DOM patches via WebSocket
- **Security-First**: Built-in XSS protection, CSP headers, and input validation

### Technology Stack

- **Go 1.21+**: Core language and standard library
- **Tailwind CSS**: Utility-first CSS (loaded via CDN in dev)
- **go-playground/validator**: Struct validation
- **GORM**: Optional database ORM for sessions and data collation

---

## Architecture Principles

### 1. String Composition Over Template Languages

Instead of using Go's `html/template` or third-party templating languages, g-sui uses function composition:

```go
// Traditional template approach
tmpl.Execute(&buf, map[string]interface{}{
    "class": "p-4 bg-white",
    "content": "Hello",
})

// g-sui approach
ui.Div("p-4 bg-white")("Hello")
```

**Benefits:**
- Type-safe through Go's type system
- IDE autocomplete and refactoring support
- No template syntax to learn
- Easy to test and debug

### 2. Callable Pattern

All page handlers and components use the `Callable` type:

```go
type Callable = func(*Context) string
```

This unifies the interface for:
- Page handlers
- Server actions
- Component renderers
- Middleware

### 3. Target-Based Updates

DOM updates use `Target` attributes with unique IDs:

```go
target := ui.Target()  // Attr{ID: "i<random>"}
// Later: ctx.Call(handler).Replace(target)
```

The framework generates unique IDs and handles the patching logic transparently.

---

## Package Structure

```
ui/
├── ui.go          # Core HTML DSL, colors, utilities
├── ui.server.go   # App, Context, HTTP server, WebSocket
├── ui.input.go    # Input components and validation
├── ui.form.go     # Form instance with automatic form association
├── ui.data.go     # Data collation (search/sort/filter/paging)
├── ui.button.go   # Button component
├── ui.table.go    # Simple table component
├── ui.label.go    # Form labels
├── ui.check.go    # Checkbox component
├── ui.radio.go    # Radio button component
├── ui.select.go   # Select dropdown component
├── ui.icon.go     # Icon helpers
├── ui.captcha.go  # Google reCAPTCHA integration
├── ui.captcha2.go # Image-based CAPTCHA
└── ui.captcha3.go # Tile-based CAPTCHA
```

### File Responsibilities

| File | Lines | Responsibility |
|------|-------|----------------|
| `ui.go` | ~927 | HTML element helpers, color constants, utility functions |
| `ui.server.go` | ~2,188 | App setup, routing, WebSocket, request handling |
| `ui.input.go` | ~877 | All input types with validation binding |
| `ui.form.go` | ~72 | Form instance for automatic form attribute association |
| `ui.data.go` | ~824 | Data table with search, sort, filter, pagination, Excel export |
| `ui.button.go` | ~130 | Button component with fluent API |
| `ui.table.go` | ~251 | Simple table with column definitions |
| `ui.label.go` | ~54 | Form label component |
| `ui.check.go` | ~95 | Checkbox component |
| `ui.radio.go` | ~330 | Radio button and radio button group |
| `ui.select.go` | ~163 | Select dropdown component |
| `ui.icon.go` | ~47 | Icon helpers (FontAwesome) |
| `ui.captcha.go` | ~108 | Google reCAPTCHA integration |
| `ui.captcha2.go` | ~487 | Image CAPTCHA generation and validation |
| `ui.captcha3.go` | ~455 | Tile puzzle CAPTCHA |

---

## Component Reference

### HTML Elements

All HTML elements follow the pattern: `ElementName(class string, attr ...Attr) func(...string) string`

| Element | Description |
|---------|-------------|
| `Div` | `<div>` container |
| `Span` | `<span>` inline container |
| `P` | `<p>` paragraph |
| `H1`, `H2`, `H3`, `H4`, `H5`, `H6` | Headings |
| `A` | `<a>` link |
| `Form` | `<form>` with method/action |
| `Textarea` | `<textarea>` for multi-line input |
| `Select` | `<select>` dropdown |
| `Option` | `<option>` for select |
| `List` | `<ul>` unordered list |
| `ListItem` | `<li>` list item |
| `Canvas` | `<canvas>` element |
| `Img` | `<img>` self-closing image |
| `Input` | `<input>` self-closing input |
| `Script` | `<script>` inline JavaScript |
| `ButtonRaw` | Raw `<button>` element |
| `Nav`, `Main`, `Header`, `Footer`, `Section`, `Article` | Semantic HTML5 elements |

### Attribute Helpers

| Helper | Description |
|--------|-------------|
| `ID(name string)` | Set element ID |
| `Href(url string)` | Set href attribute |
| `Title(text string)` | Set title attribute |
| `Target(name string)` | Generate unique target ID |
| `Attributes(class string, attr ...Attr)` | Build attribute string |

### Color Constants

Solid colors: `Blue`, `Green`, `Red`, `Purple`, `Yellow`, `Gray`, `White`

Outline variants: `BlueOutline`, `GreenOutline`, `RedOutline`, `PurpleOutline`, `YellowOutline`, `GrayOutline`, `WhiteOutline`

### Size Constants

| Size | CSS Class |
|------|-----------|
| `XS` | p-1 |
| `SM` | p-2 |
| `MD` | p-3 (default) |
| `ST` | p-4 |
| `LG` | p-5 |
| `XL` | p-6 |

### Utility Constants

| Constant | Description |
|----------|-------------|
| `Flex1` | Div with `flex-grow: 1` |
| `Space` | Non-breaking space (`&nbsp;`) |
| `INPUT` | Standard input styling class |
| `AREA` | Textarea styling class |
| `BTN` | Button base styling class |
| `DISABLED` | Disabled state styling |

### Utility Functions

| Function | Description |
|----------|-------------|
| `Classes(...string)` | Join CSS classes |
| `Map(items, fn)` | Map over slice |
| `For(start, end, fn)` | Loop with index |
| `If(condition, fn)` | Conditional render |
| `Iff(condition, html)` | Inline conditional |
| `Or(condition, trueFn, falseFn)` | Binary conditional |
| `MakeOptions(...string)` | Create option array from strings |
| `Markdown(css)(content)` | Render Markdown to HTML |

### Input Components

All inputs use fluent API: `IXxx(fieldName, dataPtr...).Method().Render("Label")`

| Input | HTML Type | Description |
|-------|-----------|-------------|
| `IText` | text | Text input |
| `IEmail` | email | Email input with validation |
| `IPhone` | tel | Phone input with +XXX pattern |
| `IPassword` | password | Password input |
| `INumber` | number | Number input |
| `IDate` | date | Date picker |
| `ITime` | time | Time picker |
| `IDateTime` | datetime-local | DateTime picker |
| `IArea` | textarea | Multi-line text area |
| `ISelect` | select | Dropdown select |
| `ICheckbox` | checkbox | Single checkbox |
| `IRadio` | radio | Single radio button |
| `IRadioButtons` | radio | Radio button group |
| `IValue` | - | Display-only value |
| `ILabel` | - | Standalone label |

#### Input Methods

- `.Class(s)`, `.ClassInput(s)`, `.ClassLabel(s)` - Custom classes
- `.Size(Size)` - Padding size
- `.Placeholder(s)` - Placeholder text
- `.Value(v)` - Default value
- `.Pattern(s)` - HTML pattern
- `.Autocomplete(s)` - Autocomplete hint
- `.Required()`, `.Disabled()`, `.Readonly()` - State modifiers
- `.Change(action)` - OnChange handler
- `.Click(action)` - OnClick handler
- `.Error(errPtr)` - Show validation error
- `.If(condition)` - Conditional render
- `.Dates(min, max)` - Min/max dates for date inputs
- `.Numbers(min, max, step)` - Min/max/step for numbers
- `.Format(s)` - Number format string
- `.Rows(n)` - Textarea rows
- `.Options(opts)` - Select options (for ISelect, IRadioButtons)
- `.Href(url)` - Make clickable link
- `.Submit()`, `.Reset()` - Form button types

### Button Component

```go
ui.Button().
    Color(ui.Blue).           // Color constant
    Size(ui.MD).              // Size constant
    Class("rounded").         // Custom classes
    Click(action).            // Click handler (JS string)
    Href("/path").            // Make it a link
    Submit().                 // type="submit"
    Reset().                  // type="reset"
    Disabled(true).           // Disable button
    If(condition).            // Conditional render
    Render("Button Text")     // Render with text
```

### Table Component

#### SimpleTable

```go
table := ui.SimpleTable(columns, classes...)
table.Field(text, class)           // Add field (auto-wraps to new row)
table.Field(text, class).Attr(`colspan="2"`)  // With custom attributes
table.Empty(cols, class)            // Empty cell with colspan
table.Render()
```

#### Generic Table

```go
table := ui.Table[T]("classes...")
table.Head(text, class)             // Header (text is escaped)
table.HeadHTML(html, class)         // Header (raw HTML)
table.Field(fn, class)              // Cell with raw HTML
table.FieldText(fn, class)          // Cell with escaped text
table.Row(fn)                       // Set row renderer
table.Render(items)
```

### Label Component

```go
ui.Label(&target).Render("Field Label")
ui.Label(&target).Required(true).Class("text-lg").Render("Label")
ui.Label(nil).Render("Label without for attribute")
```

### Checkbox Component

```go
ui.TCheckbox("FieldName", &data).
    Checked(true).
    Required().
    Render("Checkbox Label")
```

### Radio Components

```go
// Single radio button
ui.IRadio("Gender", &data).
    Value("male").
    Render("Male")

// Radio button group
ui.IRadioButtons("Gender", &data).
    Options(options).
    Render("Gender")
```

### Select Component

```go
ui.TSelect("Country", &data).
    Options(options).
    Required().
    Render("Country")
```

### Icon Helpers

```go
ui.Icon("fa fa-check")                    // <i class="fa fa-check"></i>
ui.Icon2("fa fa-check", "text-green-500") // Icon with classes
ui.IconLeft("fa fa-arrow-left", "Back")    // Icon + text
ui.IconRight("Next", "fa fa-arrow-right")  // Text + icon
ui.IconStart("fa fa-download", "Download") // Icon with gap
```

### Hidden Fields

```go
ui.Hidden("UserID", "uint", 123)
ui.Hidden("Mode", "string", "edit")
ui.Hidden("Filter[0].Field", "string", "name")
```

### Theme Switcher

```go
ui.ThemeSwitcher("")                    // Default styling
ui.ThemeSwitcher("fixed bottom-4")       // Custom positioning
// Cycles: System → Light → Dark
```

### Error Display

```go
ui.ErrorForm(&err, nil)                  // Show validation errors
ui.ErrorForm(&err, &translations)        // With translated messages
```

### Skeleton Types

| Type | Description |
|------|-------------|
| `SkeletonList` | List items with avatars |
| `SkeletonComponent` | Card/component block |
| `SkeletonPage` | Full page with header |
| `SkeletonForm` | Form with inputs |
| default | 3 text lines |

```go
target.Skeleton()                       // Default
target.Skeleton(SkeletonList)           // List skeleton
target.SkeletonList(n)                  // List with n items
target.SkeletonComponent()              // Component skeleton
target.SkeletonPage()                   // Page skeleton
target.SkeletonForm()                   // Form skeleton
```

---

## Form System

The Form system provides a convenient way to create reusable forms where input fields and submit buttons are defined outside the HTML form element. This is particularly useful when you want to reuse the same form in multiple places or separate the form structure from its content.

### FormInstance

The `FormInstance` manages form creation and automatically associates all inputs and buttons with the form via the `form` attribute.

```go
type FormInstance struct {
    FormId   string   // Unique form identifier
    OnSubmit Attr     // OnSubmit action handler
}
```

### Creating a Form

```go
func Submit(ctx *ui.Context) string {
    return "Form submitted successfully!"
}

func FormContent(ctx *ui.Context) string {
    target := ui.Target()

    // Create form with submit handler
    form := ui.FormNew(ctx.Submit(Submit).Replace(target))

    return ui.Div("max-w-5xl mx-auto")(
        form.Render(),                    // Hidden form element
        form.Text("Title").Required().Render("Title"),
        form.Email("Email").Required().Render("Email"),
        form.Phone("Phone").Render("Phone"),
        form.Number("Age").Render("Age"),
        form.Area("Address").Render("Address"),
        form.Password("Password").Render("Password"),
        form.Date("BirthDate").Render("Birth Date"),
        form.Time("AppointmentTime").Render("Time"),
        form.DateTime("CreatedAt").Render("Created At"),
        form.Select("Country").Options(options).Render("Country"),
        form.Checkbox("Agree").Required().Render("I agree"),
        form.Radio("Gender", data).Value("male").Render("Male"),
        form.RadioButtons("Plan").Options(planOptions).Render("Plan"),
        form.Button().Color(ui.Blue).Submit().Render("Submit"),
    )
}
```

### FormInstance Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `.Text(name, data...)` | `*TInput` | Text input field |
| `.Area(name, data...)` | `*TInput` | Textarea field |
| `.Password(name, data...)` | `*TInput` | Password input |
| `.Number(name, data...)` | `*TInput` | Number input |
| `.Phone(name, data...)` | `*TInput` | Phone input (tel type) |
| `.Email(name, data...)` | `*TInput` | Email input |
| `.Date(name, data...)` | `*TInput` | Date picker |
| `.Time(name, data...)` | `*TInput` | Time picker |
| `.DateTime(name, data...)` | `*TInput` | DateTime picker |
| `.Select(name, data...)` | `*ASelect` | Dropdown select |
| `.Checkbox(name, data...)` | `*TInput` | Checkbox |
| `.Radio(name, data...)` | `*TInput` | Radio button |
| `.RadioButtons(name, data...)` | `*ARadio` | Radio button group |
| `.Button()` | `*button` | Submit button |
| `.Render()` | `string` | Hidden form element |

### How It Works

1. **Form Creation**: `FormNew()` generates a unique form ID and stores the submit handler
2. **Input Association**: Each input method automatically adds the `form` attribute with the form ID
3. **Button Association**: Buttons created via `.Button()` also get the `form` attribute
4. **Hidden Form**: `.Render()` outputs a hidden `<form>` element that handles the submit event

### Benefits

- **Separation of Concerns**: Form structure is separate from field layout
- **Reusability**: Same form definition can be used in multiple contexts
- **Flexibility**: Fields can be placed anywhere in the DOM, not just inside the form element
- **Automatic Association**: No need to manually set `form` attributes on each field

---

## Core Components

### App

The `App` struct is the central container for the entire application:

```go
type App struct {
    Lanugage  string                       // Default locale
    HTMLBody  func(string, string) string   // Custom body wrapper
    HTMLHead  []string                     // Additional head elements
    sessions  *sync.Map                    // Session storage
    wsClients *sync.Map                    // WebSocket connections
    // ... internal fields
}
```

**Key Methods:**
- `MakeApp(lang string) *App` - Create new app instance
- `Page(path string, handler Callable)` - Register page route
- `Action(path string, handler Callable)` - Register server action
- `Listen(addr string)` - Start HTTP and WebSocket server
- `Favicon(fs embed.FS, path string, maxAge time.Duration)` - Serve favicon
- `Assets(fs embed.FS, path string, maxAge time.Duration)` - Serve static assets
- `StartSweeper(interval time.Duration)` - Start session cleanup goroutine
- `Handler() http.Handler` - Get HTTP handler for testing/custom servers

**HTML Generation:**
- `HTML(title, bodyClass, content) string` - Full HTML document with head, scripts, styles
- `HTMLHead []string` - Additional `<head>` elements

### Context

The `Context` struct holds request-scoped data:

```go
type Context struct {
    App       *App
    Request   *http.Request
    Response  http.ResponseWriter
    SessionID string
    // ... internal tracking fields
}
```

**Lifecycle:**
1. Created per HTTP request
2. Passed to handlers via `Callable` signature
3. Not reused across requests
4. WebSocket patches use a reference to the original context

### Context API Reference

#### Request/Response Access
- `Request *http.Request` - HTTP request
- `Response http.ResponseWriter` - HTTP response writer
- `IP() string` - Client IP address
- `Body(data interface{}) error` - Parse form/JSON into struct

#### User Feedback (Toasts)
- `Success(msg string)` - Green toast notification
- `Error(msg string)` - Red toast notification
- `Info(msg string)` - Blue toast notification
- `ErrorReload(msg string)` - Red toast with reload button

#### Page Title
- `Title(title string)` - Update the page title dynamically

#### Navigation
- `Load(href string) Attr` - SPA-like navigation (returns Attr for onclick)
- `Reload() string` - JavaScript to reload page
- `Redirect(url string) string` - JavaScript to navigate to URL

#### Session Management
- `Session(db *gorm.DB, name string) *Session` - Get session by name
- `Session.Load(data interface{})` - Load session data into struct
- `Session.Save(data interface{})` - Save struct to session

#### File Downloads
- `DownloadAs(reader io.Reader, contentType, filename string)` - Send file as download

#### Security Headers
- `SetDefaultCSP()` - Set default Content Security Policy
- `SetCSP(policy string)` - Set custom CSP
- `SetSecurityHeaders()` - Set all security headers
- `SetCustomSecurityHeaders(opts SecurityHeaderOptions)` - Custom security headers

#### WebSocket Patches
- `Patch(swap TargetSwap, html string)` - Server-initiated DOM update
- `Patch(swap TargetSwap, html string, cleanup func())` - With cleanup callback

---

## Request Lifecycle

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Request                          │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     HTTP Server (net/http)                      │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                ┌───────────────┴───────────────┐
                │                               │
                ▼                               ▼
        ┌───────────────┐               ┌──────────────┐
        │ Page Route    │               │ /__action/*  │
        │ (e.g. /)      │               │              │
        └───────┬───────┘               └──────┬───────┘
                │                               │
                ▼                               ▼
        ┌───────────────┐               ┌──────────────┐
        │ Page Handler  │               │ Action       │
        │ (Callable)    │               │ Handler      │
        └───────┬───────┘               │ (Callable)   │
                │                       └──────┬───────┘
                │                               │
                ▼                               ▼
        ┌───────────────┐               ┌──────────────┐
        │ HTML String   │               │ HTML String  │
        └───────┬───────┘               └──────┬───────┘
                │                               │
                └───────────────┬───────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │ Wrap in app.HTML()    │
                    │ (if full page)        │
                    └───────────┬───────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │ Write to Response     │
                    └───────────┬───────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Client Response                          │
└─────────────────────────────────────────────────────────────────┘
```

### Page Request Flow

1. **HTTP Request** arrives at server
2. **Router** matches to registered page path
3. **Handler Callable** is invoked with new `Context`
4. **HTML Generation** - handler returns HTML string
5. **Full Page Wrap** - `app.HTML()` wraps content with `<html>`, `<head>`, scripts
6. **Response Written** to client

### Action Request Flow

1. **Client Action** triggered (click, form submit, change)
2. **XHR Request** sent to `/__action/{id}` with JSON payload
3. **Router** matches to action handler
4. **Handler Callable** invoked
5. **HTML Response** returned (partial or empty)
6. **Client swaps DOM** based on swap method (innerHTML, outerHTML, append, etc.)

---

## State Management

### Stateless by Design

g-sui is primarily stateless - each request is independent. State is maintained through:

1. **Form Payloads**: State passed from client via form submission
2. **URL Parameters**: State in query string
3. **Sessions**: Optional server-side session storage (GORM-backed)

### Form State Pattern

For components that need to maintain state across actions:

```go
type Counter struct {
    Count int
}

func (c *Counter) Increment(ctx *ui.Context) string {
    ctx.Body(c)  // Restore state from request payload
    c.Count++
    return c.Render(ctx)
}

func (c *Counter) Render(ctx *ui.Context) string {
    target := ui.Target()
    return ui.Div("", target)(
        ui.Button().
            Click(ctx.Call(c.Increment, c).Replace(target)).
            Render(fmt.Sprintf("Count: %d", c.Count)),
    )
}
```

The state (`c`) is passed as payload to `ctx.Call()`, sent to client in hidden fields, then restored via `ctx.Body()` on the next action.

### Session Storage

For persistent state across requests:

```go
type SessionData struct {
    UserID   uint
    Username string
}

func Handler(ctx *ui.Context) string {
    session := ctx.Session(db, "auth")

    // Save
    session.Save(&SessionData{UserID: 123})

    // Load
    var data SessionData
    session.Load(&data)

    return ui.Div("")("Welcome, " + data.Username)
}
```

**Implementation:**
- Uses GORM `_session` table
- Keyed by session ID + session name
- Manual cleanup via `StartSweeper()`

---

## Action System

The action system enables interactive behavior without client-side frameworks.

### Action Types

| Action | Trigger | Returns |
|--------|---------|---------|
| `ctx.Call(fn)` | Generic (onclick, onchange) | JS string |
| `ctx.Click(fn)` | Click event | `Attr{OnClick}` |
| `ctx.Submit(fn)` | Form submit | `Attr{OnSubmit}` |
| `ctx.Change(fn)` | Input change | `Attr{OnChange}` |
| `ctx.Send(fn)` | Form-style send | JS string |

### Swap Strategies

| Strategy | Client Effect | Use Case |
|----------|---------------|----------|
| `.Render(target)` | `target.innerHTML = html` | Update content |
| `.Replace(target)` | `target.outerHTML = html` | Replace element |
| `.Append(target)` | `target.insertAdjacentHTML('beforeend', html)` | Add to end |
| `.Prepend(target)` | `target.insertAdjacentHTML('afterbegin', html)` | Add to start |
| `.None()` | No DOM swap | Fire-and-forget |

### Generated JavaScript

Actions generate inline JavaScript that:

```javascript
fetch('/__action/{actionID}', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(payload)
})
.then(r => r.json())
.then(data => {
    // Apply swap based on strategy
    if (data.swap === 'innerHTML') {
        target.innerHTML = data.html;
    } else if (data.swap === 'outerHTML') {
        target.outerHTML = data.html;
    }
    // ... handle messages
})
```

---

## WebSocket Communication

The WebSocket system at `/__ws` enables:

1. **Server-initiated DOM updates** - `ctx.Patch()`
2. **Connection monitoring** - Offline detection
3. **Auto-reconnect** - With page reload on reconnect
4. **Real-time collaboration** - Broadcast to all clients

### WebSocket Message Flow

```
┌─────────────────────┐         ┌─────────────────────┐
│   Server (Go)       │         │   Client (JS)       │
└──────────┬──────────┘         └──────────┬──────────┘
           │                                │
           │  1. WS Connection             │
           ├───────────────────────────────>│
           │                                │
           │  2. Connected                  │
           │<───────────────────────────────┤
           │                                │
           │  3. Patch Message              │
           ├───────────────────────────────>│
           │  {id: "i123",                  │
           │   swap: "innerHTML",           │
           │   html: "..."}                 │
           │                                │
           │  4. Target Invalid (cleanup)   │
           │<───────────────────────────────┤
           │  {id: "i123", invalid: true}   │
           │                                │
           │  5. Stop Sending               │
           │  (call cleanup callback)       │
           │                                │
```

### Patch Methods

```go
ctx.Patch(target.Render(), html)   // Replace innerHTML
ctx.Patch(target.Replace(), html)  // Replace element
ctx.Patch(target.Append(), html)   // Append child
ctx.Patch(target.Prepend(), html)  // Prepend child

// With cleanup callback
ctx.Patch(target.Replace(), html, func() {
    // Called when target no longer exists in DOM
    // Use to stop tickers, close channels, etc.
})
```

### Broadcast Pattern

Patches are sent to **all connected clients**. Use fixed IDs for shared elements:

```go
notificationTarget := ui.ID("global-notifications")
ctx.Patch(notificationTarget.Append(), notificationHTML)
```

---

## Data Collation (TCollate)

The `TCollate` system provides a full-featured data table with search, sort, filter, pagination, and Excel export.

### TField Configuration

```go
type TField struct {
    DB        string        // Database column name
    Field     string        // Go struct field name
    Text      string        // Display label
    As        string        // Filter type (BOOL, SELECT, DATES, etc.)
    Options   []AOption     // Options for SELECT filters
    Bool      bool          // Default value for BOOL filters
    Condition string        // Custom SQL condition
}
```

### Filter Types

| Type | Description | Example |
|------|-------------|---------|
| `BOOL` | Checkbox filter (column = 1) | Active/Inactive toggle |
| `SELECT` | Dropdown filter | Status, Category, Country |
| `DATES` | Date range picker | Created between X and Y |
| `ZERO_DATE` | "Has no date" checkbox | Never logged in |
| `NOT_ZERO_DATE` | "Has date" checkbox | Has logged in |

### TQuery Configuration

```go
type TQuery struct {
    Limit  int    // Items per page
    Order  string // Default sort (e.g., "surname asc")
    Search string // Search query
    Sort   string // Sort field
    Filter string // Filter JSON
    Page   int    // Current page
}
```

### Collate Setup

```go
// Define fields
name := ui.TField{DB: "name", Field: "Name", Text: "Name"}
status := ui.TField{
    DB: "status", Field: "Status", Text: "Status",
    As: ui.SELECT, Options: ui.MakeOptions([]string{"new", "active", "blocked"}),
}
active := ui.TField{DB: "active", Field: "Active", Text: "Active", As: ui.BOOL}

// Create collate
collate := ui.Collate[Person](&ui.TQuery{Limit: 10, Order: "name asc"})

// Configure features
collate.Search(name, status)           // Searchable fields
collate.Sort(name, status)             // Sortable fields
collate.Filter(active, status)         // Filterable fields
collate.Excel(name, status, active)    // Excel export columns

// Set row renderer
collate.Row(func(p *Person, idx int) string {
    return ui.Div("p-4 bg-white")(
        ui.Div("font-bold")(p.Name),
        ui.Div("text-sm")(p.Status),
    )
})

// Render full UI
return collate.Render(ctx, db)
```

### SQLite Normalization (Optional)

For accent-insensitive search in SQLite:

```go
ui.RegisterSQLiteNormalize(db)
// Enables searching "café" with "cafe"
```

---

## CAPTCHA Components

### Google reCAPTCHA

```go
ui.Captcha(siteKey, solvedHTML)
```

### Captcha2 (Image-based)

```go
captcha := ui.Captcha2(onValidated)
captcha.SessionField("captcha_session")
captcha.ClientVerifiedField("captcha_verified")
captcha.AnswerField("captcha_answer")
captcha.Attempts(5)
captcha.Lifetime(5 * time.Minute)
captcha.Render(ctx)

// Validation
ok, err := captcha.ValidateRequest(ctx.Request)
```

### Captcha3 (Tile Puzzle)

```go
captcha := ui.Captcha3(onValidated)
captcha.Count(5)  // Number of tiles to arrange
captcha.ArrangementField("arrangement")
captcha.ClientVerifiedField("verified")
captcha.Attempts(3)
captcha.Lifetime(10 * time.Minute)
captcha.Render(ctx)

// Validation
ok, err := captcha.ValidateRequest(ctx.Request)
```

---

## Security Model

### Server-Side Protections

#### 1. HTML Attribute Escaping

All HTML attributes are escaped using `html.EscapeString`:

```go
func escapeAttr(s string) string {
    return html.EscapeString(s)
}
```

Applied to:
- ID
- Class
- Value
- Href
- All custom attributes

#### 2. JavaScript String Escaping

For code generation (URLs in fetch, IDs):

```go
func escapeJS(s string) string {
    // JSON encoding provides safe JS string escaping
    b, _ := json.Marshal(s)
    return string(b[1 : len(b)-1]) // Remove quotes
}
```

#### 3. Safe Table Methods

- `Head(text, class)` - Text is escaped
- `HeadHTML(html, class)` - Raw HTML (use for trusted content)
- `FieldText(fn, class)` - Function return is escaped
- `Field(fn, class)` - Raw HTML (use for trusted content)

### Client-Side Protections

#### Content Security Policy

```go
ctx.SetDefaultCSP()
// Sets: default-src 'self'; script-src 'self' 'unsafe-inline'; ...
```

Or custom:

```go
ctx.SetCSP("default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline';")
```

#### Security Headers

```go
ctx.SetSecurityHeaders()
// Sets:
// - Strict-Transport-Security
// - X-Frame-Options: DENY
// - X-Content-Type-Options: nosniff
// - X-XSS-Protection: 1; mode=block
// - Referrer-Policy
// - Permissions-Policy
```

### Input Validation

Integrates with `go-playground/validator`:

```go
type Form struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

v := validator.New()
if err := v.Struct(&form); err != nil {
    // Handle validation errors
}
```

---

## Extension Points

### Custom HTML Elements

Add new HTML helpers in `ui.go`:

```go
func Article(class string, attr ...Attr) func(...string) string {
    return func(children ...string) string {
        return fmt.Sprintf("<article%s>%s</article>",
            attributes(class, attr...),
            strings.Join(children, ""))
    }
}
```

### Custom Input Types

Add new input types in `ui.input.go` by extending `TInput`:

```go
func IColor(field string, data ...any) *TInput {
    return &TInput{
        type_: "color",
        field: field,
        data:  data,
    }
}
```

### Custom Skeleton Types

Add skeleton variations in `ui.server.go`:

```go
const SkeletonCustom Skeleton = "custom"

func (a Attr) SkeletonCustom(n int) string {
    // Generate custom skeleton HTML
}
```

### Middleware

Implement middleware by wrapping handlers:

```go
func withAuth(handler ui.Callable) ui.Callable {
    return func(ctx *ui.Context) string {
        session := ctx.Session(db, "auth")
        var data SessionData
        session.Load(&data)
        if data.UserID == 0 {
            return ctx.Redirect("/login")
        }
        return handler(ctx)
    }
}

app.Page("/dashboard", withAuth(dashboardHandler))
```

---

## Performance Considerations

### HTML Generation

- **String concatenation** over `fmt.Sprintf` for simple cases
- **Avoid unnecessary allocations** - reuse buffers where possible
- **Lazy evaluation** - defer expensive work until needed

### Session Storage

- **Use sync.Map** for concurrent access without locks
- **Implement session sweeper** to prevent memory leaks
- **Consider Redis** for production deployments

### WebSocket

- **Limit broadcast recipients** by checking client subscriptions
- **Implement rate limiting** for patch messages
- **Use cleanup callbacks** to stop unnecessary goroutines

---

## Testing Patterns

### Unit Testing Handlers

```go
func TestButtonAction(t *testing.T) {
    app := ui.MakeApp("en")
    ctx := &ui.Context{App: app}

    clicked := false
    handler := func(c *ui.Context) string {
        clicked = true
        return ui.Div("")("Clicked!")
    }

    result := handler(ctx)

    if !clicked {
        t.Error("Handler was not called")
    }
    if result == "" {
        t.Error("Handler should return HTML")
    }
}
```

### Integration Testing

```go
func TestPageRender(t *testing.T) {
    app := ui.MakeApp("en")
    app.Page("/test", func(ctx *ui.Context) string {
        return ui.Div("test-class")("Test Content")
    })

    // Start test server
    server := httptest.NewServer(app.Handler())
    defer server.Close()

    resp, _ := http.Get(server.URL + "/test")
    body, _ := io.ReadAll(resp.Body)

    if !strings.Contains(string(body), "Test Content") {
        t.Error("Expected content not found")
    }
}
```

---

## Future Considerations

### Potential Enhancements

1. **Redis Session Backend** - Production-ready session storage
2. **Middleware Chain** - Formal middleware API
3. **Component Library** - Pre-built common components
4. **Plugin System** - Third-party extensions
5. **Hot Reload** - Smoother development experience
6. **Static Type Generation** - TypeScript types from Go structs
7. **SSR Optimization** - Streaming HTML for large pages

### Scalability

- **Horizontal scaling**: Session storage must be externalized (Redis)
- **WebSocket scaling**: Use Redis Pub/Sub for cross-server broadcasts
- **Caching**: Add response caching for static pages
- **CDN**: Serve Tailwind CSS and assets from CDN

---

## License

MIT
