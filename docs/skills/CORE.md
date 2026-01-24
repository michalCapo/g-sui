# g-sui Core Concepts

## Context API

The `*ui.Context` carries request-scoped data and provides methods for handling actions, responses, and state.

### Request Data

```go
ctx.Request   // *http.Request
ctx.Response  // http.ResponseWriter
ctx.IP()      // Client IP address
ctx.Body(&struct)  // Parse form/JSON into struct with automatic type inference
```

### Route Parameters

```go
// Path parameters from route patterns: /user/{id}
userID := ctx.PathParam("id")  // Returns empty string if not found

// Query parameters from URL: /search?name=Smith&age=30
name := ctx.QueryParam("name")        // Returns first value or empty string
age := ctx.QueryParam("age")          // Works with both SPA and direct navigation

// Multi-value query parameters: /tags?tag=a&tag=b
tags := ctx.QueryParams("tag")        // Returns []string or nil

// All query parameters
allParams := ctx.AllQueryParams()     // Returns map[string][]string
```

**Example:**
```go
// Route: app.Page("/user/{id}", handler)
// URL: /user/123?tab=profile&view=detailed

func handler(ctx *ui.Context) string {
    userID := ctx.PathParam("id")      // "123"
    tab := ctx.QueryParam("tab")       // "profile"
    view := ctx.QueryParam("view")     // "detailed"
    
    return ui.Div("")(fmt.Sprintf("User %s, tab: %s, view: %s", userID, tab, view))
}
```

### Type Inference in ctx.Body

Form data is automatically parsed into Go structs:

```go
type UserForm struct {
    Name      string    // String fields
    Age       int       // Auto-parsed as int
    Height    float64   // Auto-parsed as float64
    Active    bool      // Auto-parsed as bool
    BirthDate time.Time // Auto-parsed from date/datetime-local inputs
}

func (f *UserForm) Submit(ctx *ui.Context) string {
    ctx.Body(f)  // All types parsed automatically
    // f.Age is int, f.Active is bool, f.BirthDate is time.Time
}
```

### User Feedback (Toasts)

```go
ctx.Success("Operation completed")  // Green toast
ctx.Error("Something went wrong")   // Red toast
ctx.Info("FYI message")             // Blue toast
ctx.ErrorReload("Error - click to reload")  // Red toast with reload button
```

### Navigation

```go
ctx.Load("/path")    // SPA-like navigation (no full reload) - returns Attr
ctx.Reload()         // Reload current page - returns ""
ctx.Redirect("/url") // Navigate to different URL - returns ""
ctx.Title("New Title") // Update page title dynamically
```

### Security Headers

```go
ctx.SetDefaultCSP()  // Set secure default CSP headers
ctx.SetCSP("default-src 'self'; ...")  // Set custom CSP policy
ctx.SetSecurityHeaders()  // Set comprehensive security headers
ctx.SetCustomSecurityHeaders(options)  // Set custom security headers with options
```

### File Downloads

```go
ctx.DownloadAs(reader, "application/pdf", "document.pdf")  // Trigger file download
```

### Translations

```go
ctx.Translate("Hello %s", name)  // Translate message (requires app locale)
```

### Sessions (requires GORM)

```go
session := ctx.Session(db, "session_name")
session.Load(&data)   // Load from session
session.Save(&data)   // Save to session
```

## Targets & Actions

### Creating Targets

```go
target := ui.Target()  // Returns Attr{ID: "i<random>"}

// Use in elements
ui.Div("class", target)("content")

// Use in actions
ctx.Call(handler).Replace(target)
```

### Swap Strategies

```go
target.Render()   // Swap innerHTML
target.Replace()  // Replace entire element
target.Append()   // Append to element
target.Prepend()  // Prepend to element
```

### Action Methods

**ctx.Call** - Returns JS string for onclick/onchange:
```go
ctx.Call(handler, payload).Render(target)   // innerHTML
ctx.Call(handler, payload).Replace(target)  // outerHTML
ctx.Call(handler, payload).Append(target)   // Append
ctx.Call(handler, payload).Prepend(target)  // Prepend
ctx.Call(handler, payload).None()           // Fire-and-forget
```

**ctx.Submit** - Returns Attr{OnSubmit: ...} for forms:
```go
ctx.Submit(handler, payload).Render(target)
ctx.Submit(handler, payload).Replace(target)
ctx.Submit(handler, payload).Append(target)
ctx.Submit(handler, payload).Prepend(target)
ctx.Submit(handler, payload).None()
```

**ctx.Click** - Returns Attr{OnClick: ...} for elements:
```go
ctx.Click(handler, payload).Render(target)   // Returns Attr
ctx.Click(handler, payload).Replace(target)   // Returns Attr
ctx.Click(handler, payload).Append(target)    // Returns Attr
ctx.Click(handler, payload).Prepend(target)   // Returns Attr
ctx.Click(handler, payload).None()            // Returns Attr
```

**ctx.Send** - Returns Actions (same as Call but uses FORM method):
```go
ctx.Send(handler, payload).Render(target)   // Returns string
ctx.Send(handler, payload).Replace(target)  // Returns string
ctx.Send(handler, payload).Append(target)   // Returns string
ctx.Send(handler, payload).Prepend(target)  // Returns string
ctx.Send(handler, payload).None()           // Returns string
```

## Stateful Components

Pass state through payload structs:

```go
type Counter struct { Count int }

func (c *Counter) Increment(ctx *ui.Context) string {
    ctx.Body(c)  // Restore state from request
    c.Count++
    return c.Render(ctx)
}

func (c *Counter) Render(ctx *ui.Context) string {
    target := ui.Target()
    return ui.Div("flex gap-2", target)(
        ui.Button().
            Click(ctx.Call(c.Increment, c).Replace(target)).
            Render(fmt.Sprintf("Count: %d", c.Count)),
    )
}
```

## WebSocket Patches (Real-time Updates)

Broadcast HTML updates to all connected clients:

```go
// Convenience methods
ctx.Render(target, html)   // Replace innerHTML
ctx.Replace(target, html)  // Replace entire element

// Full API
ctx.Patch(target.Render(), html)   // innerHTML
ctx.Patch(target.Replace(), html)  // outerHTML
ctx.Patch(target.Append(), html)   // Append
ctx.Patch(target.Prepend(), html)  // Prepend
```

### Live Updates Example

```go
func Clock(ctx *ui.Context) string {
    target := ui.Target()

    stop := make(chan struct{})

    // Cleanup when target disappears
    ctx.Patch(target.Replace(), clockHTML(), func() {
        close(stop)
    })

    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-stop:
                return
            case <-ticker.C:
                ctx.Replace(target, clockHTML())
            }
        }
    }()

    return clockHTML()
}
```

## HTML DSL

### Elements

```go
ui.Div(class, attr...)(children...)    // <div>
ui.Span(class, attr...)(children...)   // <span>
ui.P(class, attr...)(children...)      // <p>
ui.A(class, attr...)(children...)      // <a>
ui.Form(class, attr...)(children...)   // <form>
ui.Input(class, attr...)               // <input />
ui.Img(class, attr...)                 // <img />
```

### Attributes

```go
ui.Attr{
    ID: "myid",
    Class: "extra",
    Href: "/path",
    Value: "val",
    OnClick: "js()",
    Required: true,
    Disabled: true,
}

// Shorthands
ui.Href("/path")
ui.ID("myid")
ui.Title("tooltip")
```

### Control Flow

```go
ui.Map(items, func(item *T, i int) string { return ... })
ui.For(0, 10, func(i int) string { return ... })
ui.If(condition, func() string { return ... })
ui.Iff(condition)("content if true")
ui.Or(condition, trueFn, falseFn)
```

## Skeleton Loading States

```go
target.Skeleton()                    // Default (3 lines)
target.Skeleton(ui.SkeletonList)     // List items
target.Skeleton(ui.SkeletonComponent) // Component block
target.Skeleton(ui.SkeletonPage)     // Full page
target.Skeleton(ui.SkeletonForm)     // Form layout
```

### Deferred Loading Pattern

```go
func DeferredComponent(ctx *ui.Context) string {
    target := ui.Target()

    go func() {
        time.Sleep(2 * time.Second)  // Simulate slow fetch
        ctx.Replace(target, loadedContent())
    }()

    return target.Skeleton(ui.SkeletonComponent)
}
```
