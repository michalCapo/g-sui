# g-sui Documentation

This document combines the LLM Reference Guide and Architecture Documentation for g-sui.

## Table of Contents

### Part I: LLM Reference Guide
1. [Quick Start](#quick-start)
2. [Core Types](#core-types)
3. [App Setup](#app-setup)
4. [HTML DSL](#html-dsl)
5. [Targets & Actions](#targets--actions)
6. [Click Examples](#click-examples)
7. [Submit Examples](#submit-examples)
8. [Buttons](#buttons)
9. [Inputs](#inputs)
10. [Forms](#forms)
11. [FormInstance (Disconnected Forms)](#forminstance-disconnected-forms)
12. [Context API](#context-api)
13. [WebSocket Patches (Real-time Updates)](#websocket-patches-real-time-updates)
14. [Skeletons (Loading States)](#skeletons-loading-states)
15. [Tables](#tables)
16. [Data Collation (Search, Sort, Filter, Pagination, Excel Export)](#data-collation-search-sort-filter-pagination-excel-export)
17. [Stateful Components](#stateful-components)
18. [Labels](#labels)
19. [Icons (Material Icons)](#icons-material-icons)
20. [CAPTCHA Components](#captcha-components)
21. [Theme Switcher](#theme-switcher)
22. [Hidden Fields](#hidden-fields)
23. [CSS Constants](#css-constants)
24. [Common Patterns](#common-patterns)
25. [Sessions](#sessions)
26. [Security](#security)
27. [Validation Tags (go-playground/validator)](#validation-tags-go-playgroundvalidator)
28. [Project Structure](#project-structure)
29. [Dependencies](#dependencies)
30. [Quick Reference](#quick-reference)
30. [UI Components](#ui-components)

### Part II: Architecture Documentation
31. [Architecture Overview](#architecture-overview)
32. [Architecture Principles](#architecture-principles)
33. [Package Structure](#package-structure)
34. [Component Reference](#component-reference)
35. [Form System](#form-system)
36. [Core Components](#core-components)
37. [Context API Reference](#context-api-reference)
38. [Smooth Navigation](#smooth-navigation)
39. [Request Lifecycle](#request-lifecycle)
40. [State Management](#state-management)
41. [Action System](#action-system)
42. [WebSocket Communication](#websocket-communication)
43. [Data Collation (TCollate)](#data-collation-tcollate)
44. [CAPTCHA Components](#captcha-components)
45. [Security Model](#security-model)
46. [Extension Points](#extension-points)
47. [SPA (Single Page Application)](#spa-single-page-application)
48. [PWA (Progressive Web App)](#pwa-progressive-web-app)
49. [Performance Considerations](#performance-considerations)
50. [Testing Patterns](#testing-patterns)
51. [Future Considerations](#future-considerations)

---

# Part I: LLM Reference Guide

> Server-rendered UI framework for Go. All logic runs server-side; minimal client JS.

## Quick Start

```go
package main

import (
    "github.com/michalCapo/g-sui/ui"
)

func main() {
    app := ui.MakeApp("en")
    
    app.Page("/", func(ctx *ui.Context) string {
        return app.HTML("Home", "bg-gray-100", 
            ui.Div("p-8")(
                ui.Div("text-2xl font-bold")("Hello World"),
            ),
        )
    })
    
    app.Listen(":8080")
}
```

---

## Core Types

```go
type Callable = func(*ui.Context) string  // All handlers return HTML strings
type Attr struct { ID, Href, Class, Value, OnClick, OnSubmit string; ... }
type AOption struct { ID, Value string }
```

---

## App Setup

```go
app := ui.MakeApp("en")                           // Create app with locale
app.Page("/path", "Page Title", handler)          // Register page route
app.Favicon(embedFS, "assets/favicon.svg", 24*time.Hour)
app.Assets(embedFS, "assets/", 24*time.Hour)      // Serve static files
app.AutoRestart(true)                             // Dev: rebuild on file changes
app.Listen(":8080")                               // Start server (also starts WS at /__ws)
```

### Custom HTTP Handlers (REST APIs)
```go
// Register custom HTTP handlers (checked before g-sui routes)
app.Custom("GET", "/api/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status": "ok"}`))
})

app.Custom("POST", "/api/users", createUserHandler)

// Shorthand methods
app.GET("/api/data", getDataHandler)
app.POST("/api/data", createDataHandler)
app.PUT("/api/data/:id", updateDataHandler)
app.DELETE("/api/data/:id", deleteDataHandler)
app.PATCH("/api/data/:id", patchDataHandler)
```

### Custom Server Configuration
```go
// Get http.Handler for custom server setups
app := ui.MakeApp("en")
app.Page("/", "Home", homeHandler)
app.StartSweeper()  // Manually start session sweeper

// Wrap with custom middleware
handler := myLoggingMiddleware(app.Handler())

// Use with custom server
server := &http.Server{
    Addr:    ":8080",
    Handler: handler,
}
server.ListenAndServe()
```

### Testing Handler
```go
handler := app.TestHandler()                      // Get http.Handler for testing
server := httptest.NewServer(handler)             // Create test server
resp, _ := http.Get(server.URL + "/path")         // Make test requests
```

### HTML Wrapper
```go
app.HTML(title, bodyClass, content) string        // Full HTML document with Tailwind
app.HTMLHead = append(app.HTMLHead, `<link ...>`) // Add to <head>
```

---

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


```go
type Callable = func(*ui.Context) string  // All handlers return HTML strings
type Attr struct { ID, Href, Class, Value, OnClick, OnSubmit string; ... }
type AOption struct { ID, Value string }
```

---

## App Setup

```go
app := ui.MakeApp("en")                           // Create app with locale
app.Page("/path", "Page Title", handler)          // Register page route
app.Favicon(embedFS, "assets/favicon.svg", 24*time.Hour)
app.Assets(embedFS, "assets/", 24*time.Hour)      // Serve static files
app.AutoRestart(true)                             // Dev: rebuild on file changes
app.Listen(":8080")                               // Start server (also starts WS at /__ws)
```

### Custom HTTP Handlers (REST APIs)
```go
// Register custom HTTP handlers (checked before g-sui routes)
app.Custom("GET", "/api/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status": "ok"}`))
})

app.Custom("POST", "/api/users", createUserHandler)

// Shorthand methods
app.GET("/api/data", getDataHandler)
app.POST("/api/data", createDataHandler)
app.PUT("/api/data/:id", updateDataHandler)
app.DELETE("/api/data/:id", deleteDataHandler)
app.PATCH("/api/data/:id", patchDataHandler)
```

### Custom Server Configuration
```go
// Get http.Handler for custom server setups
app := ui.MakeApp("en")
app.Page("/", "Home", homeHandler)
app.StartSweeper()  // Manually start session sweeper

// Wrap with custom middleware
handler := myLoggingMiddleware(app.Handler())

// Use with custom server
server := &http.Server{
    Addr:    ":8080",
    Handler: handler,
}
server.ListenAndServe()
```

### Testing Handler
```go
handler := app.TestHandler()                      // Get http.Handler for testing
server := httptest.NewServer(handler)             // Create test server
resp, _ := http.Get(server.URL + "/path")         // Make test requests
```

### HTML Wrapper
```go
app.HTML(title, bodyClass, content) string        // Full HTML document with Tailwind
app.HTMLHead = append(app.HTMLHead, `<link ...>`) // Add to <head>
```

---

## HTML DSL

### Elements
```go
ui.Div(class, attr...)(children...)    // <div>
ui.Span(class, attr...)(children...)   // <span>
ui.P(class, attr...)(children...)      // <p>
ui.A(class, attr...)(children...)      // <a>
ui.Form(class, attr...)(children...)   // <form>
ui.H1(class, attr...)(children...)     // <h1>
ui.H2(class, attr...)(children...)     // <h2>
ui.H3(class, attr...)(children...)     // <h3>
ui.List(class, attr...)(children...)   // <ul>
ui.ListItem(class, attr...)(children...)// <li>
ui.Img(class, attr...)                 // <img /> (self-closing)
ui.Input(class, attr...)               // <input /> (self-closing)
```

**Example:**
```go
// Heading with Tailwind classes
ui.H1("text-4xl font-bold text-blue-900")("Welcome")

// Nested headings and paragraphs
ui.Div("max-w-2xl")(
    ui.H1("text-3xl font-bold mb-4")("Article Title"),
    ui.H2("text-xl text-gray-600 mb-2")("Subtitle"),
    ui.P("text-gray-700 leading-relaxed")("Article content goes here..."),
)

// With custom attributes
ui.H3("", ui.Attr{ID: "section-1"})(
    ui.Span("text-sm text-gray-500")("[1.5]"),
    " Introduction",
)
```

### Attributes
```go
ui.Attr{
    ID: "myid",
    Class: "extra-classes",
    Href: "/path",
    Value: "val",
    OnClick: "jsCode()",
    OnSubmit: "jsCode()",
    Required: true,
    Disabled: true,
}

// Shorthand helpers
ui.Href("/path")           // Attr{Href: "/path"}
ui.ID("myid")              // Attr{ID: "myid"}
ui.Title("tooltip")        // Attr{Title: "tooltip"}
```

### Control Flow
```go
ui.Map(items, func(item *T, index int) string { return ... })
ui.For(0, 10, func(i int) string { return ... })
ui.If(condition, func() string { return ... })
ui.Iff(condition)("content shown if true")
ui.Or(condition, trueFunc, falseFunc)
```

### Utilities
```go
ui.Flex1           // Div that grows (flex-grow: 1)
ui.Space           // &nbsp;
ui.Classes(a, b)   // Join CSS classes
ui.Markdown(css)(md) // Render markdown to HTML
ui.Script(js...)   // Inline <script>
```

---

## Targets & Actions

### Creating Targets
```go
target := ui.Target()  // Returns Attr{ID: "i<random>"}

// Use in elements:
ui.Div("class", target)("content")

// Use in actions (where to put response):
ctx.Call(handler).Replace(target)
```

### Server Actions Overview
```go
// ctx.Call - returns JS string for onclick/onchange handlers
ctx.Call(handler, payloadStruct...).Render(target)   // Replace innerHTML
ctx.Call(handler, payloadStruct...).Replace(target)  // Replace entire element
ctx.Call(handler, payloadStruct...).Append(target)   // Append to element
ctx.Call(handler, payloadStruct...).Prepend(target)  // Prepend to element
ctx.Call(handler, payloadStruct...).None()           // Fire-and-forget (no DOM update)

// ctx.Submit - returns Attr{OnSubmit: ...} for forms
ctx.Submit(handler, payloadStruct...).Render(target)
ctx.Submit(handler, payloadStruct...).Replace(target)
ctx.Submit(handler, payloadStruct...).Append(target)
ctx.Submit(handler, payloadStruct...).Prepend(target)
ctx.Submit(handler, payloadStruct...).None()

// ctx.Click - returns Attr{OnClick: ...} for elements
ctx.Click(handler, payloadStruct...).Render(target)
ctx.Click(handler, payloadStruct...).Replace(target)
ctx.Click(handler, payloadStruct...).Append(target)
ctx.Click(handler, payloadStruct...).Prepend(target)
ctx.Click(handler, payloadStruct...).None()

// ctx.Send - returns JS string (similar to Call but for form-style submission)
ctx.Send(handler, payloadStruct...).Render(target)
```

---

## Click Examples

### Basic Click - Replace Target
```go
func MyPage(ctx *ui.Context) string {
    target := ui.Target()
    
    sayHello := func(ctx *ui.Context) string {
        ctx.Success("Hello!")
        return ui.Div("text-green-500", target)("Clicked!")
    }
    
    return ui.Div("", target)(
        ui.Button().
            Color(ui.Blue).
            Click(ctx.Call(sayHello).Replace(target)).
            Render("Click me"),
    )
}
```

### Click with State - Counter Example
```go
type Counter struct {
    Count int
}

func (c *Counter) Increment(ctx *ui.Context) string {
    ctx.Body(c)  // Restore state from previous call
    c.Count++
    return c.Render(ctx)
}

func (c *Counter) Decrement(ctx *ui.Context) string {
    ctx.Body(c)
    c.Count--
    if c.Count < 0 {
        c.Count = 0
    }
    return c.Render(ctx)
}

func (c *Counter) Render(ctx *ui.Context) string {
    target := ui.Target()
    return ui.Div("flex gap-2 items-center bg-purple-500 rounded text-white p-1", target)(
        ui.Button().
            Click(ctx.Call(c.Decrement, c).Replace(target)).  // Pass state as payload
            Class("rounded-l px-5").
            Render("-"),
        ui.Div("text-2xl")(fmt.Sprintf("%d", c.Count)),
        ui.Button().
            Click(ctx.Call(c.Increment, c).Replace(target)).
            Class("rounded-r px-5").
            Render("+"),
    )
}
```

### Click - Append Items
```go
func AppendDemo(ctx *ui.Context) string {
    target := ui.Target()

    addItem := func(ctx *ui.Context) string {
        now := time.Now().Format("15:04:05")
        return ui.Div("p-2 rounded border bg-white")(
            fmt.Sprintf("Added at %s", now),
        )
    }

    return ui.Div("")(
        ui.Button().
            Color(ui.Blue).
            Click(ctx.Call(addItem).Append(target)).
            Render("Add item at end"),
        ui.Button().
            Color(ui.Green).
            Click(ctx.Call(addItem).Prepend(target)).
            Render("Add item at start"),
        ui.Div("space-y-2 mt-4", target)(
            ui.Div("p-2 border bg-white")("Initial item"),
        ),
    )
}
```

### Click - Fire and Forget (None)
```go
func FireAndForget(ctx *ui.Context) string {
    logAction := func(ctx *ui.Context) string {
        log.Println("Button clicked, but no UI update needed")
        ctx.Success("Action logged!")
        return ""  // Return value ignored when using .None()
    }

    return ui.Button().
        Color(ui.Blue).
        Click(ctx.Call(logAction).None()).
        Render("Log Action")
}
```

---

## Submit Examples

### Basic Form Submit
```go
type LoginForm struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=6"`
}

func (f *LoginForm) OnSubmit(ctx *ui.Context) string {
    if err := ctx.Body(f); err != nil {
        ctx.Error("Invalid form data")
        return f.Render(ctx, &err)
    }
    
    v := validator.New()
    if err := v.Struct(f); err != nil {
        return f.Render(ctx, &err)
    }
    
    ctx.Success("Login successful!")
    return f.Render(ctx, nil)
}

func (f *LoginForm) Render(ctx *ui.Context, err *error) string {
    target := ui.Target()
    
    // Form with Submit - the form replaces itself on submit
    return ui.Form("bg-white p-6 rounded shadow", target, ctx.Submit(f.OnSubmit).Replace(target))(
        ui.ErrorForm(err, nil),
        ui.IEmail("Email", f).Required().Error(err).Render("Email"),
        ui.IPassword("Password").Required().Error(err).Render("Password"),
        ui.Button().Submit().Color(ui.Blue).Class("rounded").Render("Login"),
    )
}
```

### Form with Reset Button
```go
func (f *MyForm) OnReset(ctx *ui.Context) string {
    f.Name = ""
    f.Description = ""
    return f.Render(ctx)
}

func (f *MyForm) Render(ctx *ui.Context) string {
    target := ui.Target()
    
    return ui.Form("flex flex-col gap-4", target, ctx.Submit(f.OnSubmit).Replace(target))(
        ui.IText("Name", f).Render("Name"),
        ui.IArea("Description", f).Rows(4).Render("Description"),
        
        ui.Div("flex gap-4 justify-end")(
            // Reset button uses Click, not Submit
            ui.Button().
                Click(ctx.Call(f.OnReset).Replace(target)).
                Color(ui.Gray).
                Render("Reset"),
            ui.Button().Submit().Color(ui.Blue).Render("Submit"),
        ),
    )
}
```

### Form Submission with Validation Translations
```go
var translations = map[string]string{
    "Name":              "User name",
    "Email":             "Email address",
    "has invalid value": "is invalid",
}

func (f *Form) Render(ctx *ui.Context, err *error) string {
    target := ui.Target()
    return ui.Form("p-4", target, ctx.Submit(f.OnSubmit).Replace(target))(
        ui.ErrorForm(err, &translations),  // Pass translations map
        ui.IText("Name", f).Required().Error(err).Render("Name"),
        ui.IEmail("Email", f).Required().Error(err).Render("Email"),
        ui.Button().Submit().Color(ui.Blue).Render("Submit"),
    )
}
```

---

## Buttons

```go
ui.Button().
    Color(ui.Blue).           // Blue, Green, Red, Yellow, Purple, Gray, White + *Outline variants
    Size(ui.MD).              // XS, SM, MD, ST, LG, XL
    Class("rounded px-4").    // Custom classes
    Click(ctx.Call(...)).     // Click handler (JS string)
    Href("/path").            // Make it a link
    Submit().                 // type="submit" for forms
    Reset().                  // type="reset" for forms
    Disabled(true).           // Disable button
    If(condition).            // Conditional rendering
    Render("Button Text")
```

---

## Inputs

All inputs use fluent API: `IXxx(fieldName, dataPtr...).Method().Render("Label")`

### Text Inputs
```go
ui.IText("FieldName", &data).Required().Placeholder("hint").Render("Label")
ui.IEmail("Email", &data).Required().Render("Email")
ui.IPhone("Phone", &data).Render("Phone")       // +XXX pattern
ui.IPassword("Password").Required().Render("Password")
ui.IArea("Bio", &data).Rows(5).Render("Biography")
```

### Numbers & Dates
```go
ui.INumber("Age", &data).Numbers(0, 120, 1).Render("Age")
ui.INumber("Price", &data).Format("%.2f").Render("Price")
ui.IDate("BirthDate", &data).Dates(minTime, maxTime).Render("Birth Date")
ui.ITime("Alarm", &data).Render("Alarm Time")
ui.IDateTime("Meeting", &data).Render("Meeting")
```

### Selection
```go
// Dropdown select
options := ui.MakeOptions([]string{"A", "B", "C"})  // []AOption from strings
options := []ui.AOption{{ID: "val", Value: "Display"}, ...}
ui.ISelect("Country", &data).Options(options).Render("Country")

// Checkbox
ui.ICheckbox("Agree", &data).Required().Render("I agree")

// Radio buttons
ui.IRadio("Gender", &data).Value("male").Render("Male")
ui.IRadio("Gender", &data).Value("female").Render("Female")
// OR radio button group:
ui.IRadioButtons("Gender", &data).Options(genderOptions).Render("Gender")

// RadioDiv - Card-based radio selection (custom HTML options)
cardOptions := []ui.AOption{
    {ID: "1", Value: ui.Div("p-4 border rounded")("Card 1")},
    {ID: "2", Value: ui.Div("p-4 border rounded")("Card 2")},
}
ui.IRadioDiv("Plan", &data).Options(cardOptions).Render("Choose a plan")
```

### Common Input Methods
```go
.Required()           // Mark required
.Disabled()           // Disable input
.Readonly()           // Read-only
.Placeholder("hint")  // Placeholder text
.Class("classes")     // Wrapper classes
.ClassInput("cls")    // Input element classes
.ClassLabel("cls")    // Label classes
.Value("default")     // Default value
.Pattern("regex")     // HTML pattern
.Autocomplete("email")// Autocomplete hint
.Change(action)       // OnChange handler (JS string)
.Click(action)        // OnClick handler (JS string)
.Error(&err)          // Show validation error
.If(condition)        // Conditional render
.Render("Label")      // Render with label
```

---

## Forms

### Form Data Parsing with Automatic Type Inference

`ctx.Body(&struct)` automatically parses form data into Go structs with type inference:

```go
type UserForm struct {
    Name      string    // String fields
    Age       int       // Automatically parsed as int
    Height    float64   // Automatically parsed as float64
    Active    bool      // Automatically parsed as bool
    BirthDate time.Time // Automatically parsed from date/datetime-local inputs
    CreatedAt time.Time // Handles time.Time with multiple format support
}

func (f *UserForm) Submit(ctx *ui.Context) string {
    if err := ctx.Body(f); err != nil {  // Auto-parses all types
        ctx.Error("Invalid form data: " + err.Error())
        return f.Render(ctx, &err)
    }

    // f.Age is now an int, f.Active is a bool, f.BirthDate is time.Time
    ctx.Success(fmt.Sprintf("User %s, age %d, active: %v", f.Name, f.Age, f.Active))
    return f.Render(ctx, nil)
}
```

**Supported Types:**
- `string` - Direct assignment
- `int`, `int8`, `int16`, `int32`, `int64` - Parsed from numeric strings
- `uint`, `uint8`, `uint16`, `uint32`, `uint64` - Parsed from numeric strings
- `float32`, `float64` - Parsed from decimal strings
- `bool` - Parses "true", "false", "1", "0" (case-insensitive)
- `time.Time` - Parses multiple date/time formats:
  - `2006-01-02` (HTML date input)
  - `2006-01-02T15:04` (HTML datetime-local)
  - `15:04` (HTML time input)
  - RFC3339 and RFC3339Nano
- `gorm.DeletedAt` - Special handling for GORM's soft delete type
- **Type aliases** - Automatically handled (e.g., `type MyString string`)

### Basic Form Example

```go
type LoginForm struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

func (f *LoginForm) Submit(ctx *ui.Context) string {
    if err := ctx.Body(f); err != nil {  // Parse form data into struct
        return f.Render(ctx, &err)
    }

    v := validator.New()
    if err := v.Struct(f); err != nil {  // Validate struct
        return f.Render(ctx, &err)
    }

    ctx.Success("Login successful!")
    return ctx.Redirect("/dashboard")
}

func (f *LoginForm) Render(ctx *ui.Context, err *error) string {
    target := ui.Target()

    return ui.Form("bg-white p-6 rounded", target, ctx.Submit(f.Submit).Replace(target))(
        ui.ErrorForm(err, nil),  // Show validation errors

        ui.IEmail("Email", f).Required().Error(err).Render("Email"),
        ui.IPassword("Password").Required().Error(err).Render("Password"),

        ui.Button().Submit().Color(ui.Blue).Render("Login"),
    )
}
```

---

## FormInstance (Disconnected Forms)

The `FormInstance` allows creating forms where inputs and buttons are placed outside the HTML form element. All fields are automatically associated with the form via the `form` attribute.

### Basic Usage

```go
func Submit(ctx *ui.Context) string {
    ctx.Success("Form submitted!")
    return ""
}

func FormContent(ctx *ui.Context) string {
    target := ui.Target()

    // Create form instance with submit handler
    form := ui.FormNew(ctx.Submit(Submit).Replace(target))

    return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
        form.Render(),                         // Hidden form element
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
| `.Text(name, data...)` | `*TInput` | Text input |
| `.Area(name, data...)` | `*TInput` | Textarea |
| `.Password(name, data...)` | `*TInput` | Password input |
| `.Number(name, data...)` | `*TInput` | Number input |
| `.Phone(name, data...)` | `*TInput` | Phone input (tel) |
| `.Email(name, data...)` | `*TInput` | Email input |
| `.Date(name, data...)` | `*TInput` | Date picker |
| `.Time(name, data...)` | `*TInput` | Time picker |
| `.DateTime(name, data...)` | `*TInput` | DateTime picker |
| `.Select(name, data...)` | `*ASelect` | Dropdown select |
| `.Checkbox(name, data...)` | `*TInput` | Checkbox |
| `.Radio(name, data...)` | `*TInput` | Radio button |
| `.RadioButtons(name, data...)` | `*ARadio` | Radio button group |
| `.RadioDiv(name, data...)` | `*ARadio` | Card-based radio (custom HTML) |
| `.File(name)` | `*TFile` | File input |
| `.ImageUpload(name)` | `*TImageUpload` | Image upload with inline preview |
| `.Button()` | `*button` | Submit button |
| `.Render()` | `string` | Hidden form element |

### Reusable Form Pattern

```go
// Define a reusable form component
type UserForm struct {
    form *ui.FormInstance
    data *UserData
}

func NewUserForm(data *UserData, submitHandler ui.Callable) *UserForm {
    target := ui.Target()
    return &UserForm{
        form: ui.FormNew(ctx.Submit(submitHandler).Replace(target)),
        data: data,
    }
}

func (f *UserForm) Render(ctx *ui.Context) string {
    return ui.Div("bg-white p-6 rounded-lg")(
        f.form.Render(),
        f.form.Text("Name", f.data).Required().Render("Name"),
        f.form.Email("Email", f.data).Required().Render("Email"),
        f.form.Button().Color(ui.Blue).Submit().Render("Save"),
    )
}
```

---

## Context API

### Request Data
```go
ctx.Request          // *http.Request
ctx.Response         // http.ResponseWriter  
ctx.IP()             // Client IP
ctx.Body(&struct)    // Parse form/JSON into struct
```

### User Feedback (Toasts)
```go
ctx.Success("Operation completed")  // Green toast
ctx.Error("Something went wrong")   // Red toast
ctx.Info("FYI message")             // Blue toast
ctx.ErrorReload("Error - click to reload")  // Red toast with reload button
```

### Page Title
```go
ctx.Title("New Page Title")  // Update the page title dynamically
```

### Navigation
```go
ctx.Load("/path")    // Returns Attr for SPA-like navigation (no full reload)
ctx.Reload()         // Returns JS to reload page
ctx.Redirect("/url") // Returns JS to navigate away
```

### Sessions (requires GORM DB)
```go
session := ctx.Session(db, "session_name")
session.Load(&data)   // Load data from session
session.Save(&data)   // Save data to session
```

### File Downloads
```go
ctx.DownloadAs(&reader, "application/xlsx", "export.xlsx")
```

### File Uploads

**Single File Upload:**
```go
file, err := ctx.File("image")
if err != nil {
    ctx.Error("Failed to process file: " + err.Error())
    return renderForm(ctx)
}
if file == nil {
    ctx.Error("No file uploaded")
    return renderForm(ctx)
}

// File object properties:
// file.Name        - Original filename
// file.Data        - File contents as []byte
// file.ContentType - MIME type (e.g., "image/png")
// file.Size        - File size in bytes

// Save file
os.WriteFile("uploads/"+file.Name, file.Data, 0644)
```

**Multiple File Upload:**
```go
files, err := ctx.Files("images")  // Returns []*FileUpload
if err != nil {
    ctx.Error("Failed to process files: " + err.Error())
    return renderForm(ctx)
}
if len(files) == 0 {
    ctx.Error("No files uploaded")
    return renderForm(ctx)
}

// Process each file
for _, file := range files {
    // Validate and save each file
    if strings.HasPrefix(file.ContentType, "image/") {
        os.WriteFile("uploads/"+file.Name, file.Data, 0644)
    }
}
```

### Security Headers
```go
ctx.SetDefaultCSP()
ctx.SetSecurityHeaders()
```

### WebSocket Patches
```go
ctx.Render(target, html)   // Render HTML inside target (innerHTML)
ctx.Replace(target, html)  // Replace target element (outerHTML)
ctx.Patch(target.Render(), html)   // Full API: replace innerHTML
ctx.Patch(target.Replace(), html)  // Full API: replace element
ctx.Patch(target.Append(), html)   // Full API: append to target
ctx.Patch(target.Prepend(), html)  // Full API: prepend to target
```

---

## WebSocket Patches (Real-time Updates)

WebSocket patches push HTML updates from server to all connected clients. The WebSocket endpoint is automatically created at `/__ws`.

### Patch Methods
```go
// Convenience methods (recommended)
ctx.Render(target, html)   // Replace innerHTML of target
ctx.Replace(target, html)  // Replace entire target element

// Full Patch API with all swap strategies
ctx.Patch(target.Render(), html)   // Replace innerHTML of target
ctx.Patch(target.Replace(), html)  // Replace entire target element
ctx.Patch(target.Append(), html)   // Append HTML to target
ctx.Patch(target.Prepend(), html)  // Prepend HTML to target

// With cleanup callback (called when target no longer exists in DOM)
ctx.Patch(target.Replace(), html, func() {
    // Called when client reports target is invalid
    // Use to stop tickers, cleanup resources
})
```

### Live Clock Example
```go
func Clock(ctx *ui.Context) string {
    target := ui.Target()

    clockUI := func() string {
        t := time.Now()
        return ui.Div("font-mono text-3xl bg-white p-4 border rounded", target)(
            fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second()),
        )
    }

    // Start background updates
    stop := make(chan struct{})
    
    // First patch with cleanup callback (requires full Patch API)
    ctx.Patch(target.Replace(), clockUI(), func() {
        close(stop)  // Stop ticker when target disappears
    })

    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()
        for {
            select {
            case <-stop:
                return
            case <-ticker.C:
                // Using convenience method
                ctx.Replace(target, clockUI())
            }
        }
    }()

    return clockUI()
}
```

### Deferred Loading with Skeleton
```go
func DeferredComponent(ctx *ui.Context) string {
    target := ui.Target()

    // Start async data fetch
    go func() {
        defer func() { recover() }()  // Handle panics gracefully
        
        time.Sleep(2 * time.Second)  // Simulate slow API call
        
        html := ui.Div("space-y-4", target)(
            ui.Div("bg-white p-4 rounded shadow")(
                ui.Div("text-lg font-semibold")("Content loaded!"),
                ui.Div("text-gray-600")("This replaced the skeleton via WebSocket."),
            ),
        )
        // Using convenience method
        ctx.Replace(target, html)
    }()

    // Return skeleton immediately
    return target.Skeleton(ui.SkeletonComponent)
}
```

### Multiple Patches (Replace + Append)
```go
func DeferredWithButtons(ctx *ui.Context) string {
    target := ui.Target()

    // First patch: replace skeleton with content
    go func() {
        defer func() { recover() }()
        time.Sleep(2 * time.Second)
        
        content := ui.Div("bg-white p-4 rounded", target)(
            ui.Div("font-bold")("Main content loaded"),
        )
        ctx.Patch(target.Replace(), content)
    }()

    // Second patch: append controls after content
    go func() {
        defer func() { recover() }()
        time.Sleep(2100 * time.Millisecond)  // Slightly after first patch
        
        controls := ui.Div("flex gap-2 mt-4")(
            ui.Button().Color(ui.Blue).Class("rounded").
                Click(ctx.Call(DeferredWithButtons).Replace(target)).
                Render("Reload"),
        )
        ctx.Patch(target.Append(), controls)
    }()

    return target.Skeleton(ui.SkeletonList)
}
```

### Broadcast to All Clients
```go
// Patches are broadcast to ALL connected WebSocket clients
// This enables real-time collaboration features

func NotifyAll(ctx *ui.Context) string {
    notificationTarget := ui.ID("global-notifications")  // Fixed ID all clients have
    
    message := ui.Div("bg-blue-100 p-2 rounded")(
        fmt.Sprintf("New message at %s", time.Now().Format("15:04:05")),
    )
    
    // For append/prepend, use full Patch API
    ctx.Patch(notificationTarget.Append(), message)
    ctx.Success("Notification sent to all clients!")
    return ""
}

// Example using Render convenience method
func UpdateStatus(ctx *ui.Context) string {
    statusTarget := ui.ID("status-indicator")
    
    html := ui.Div("bg-green-500 text-white p-2 rounded")("Online")
    ctx.Render(statusTarget, html)  // Update innerHTML
    return ""
}
```

---

## Skeletons (Loading States)

```go
target := ui.Target()

// Return skeleton immediately, patch real content later
go func() {
    time.Sleep(2 * time.Second)
    html := ui.Div("", target)("Content loaded!")
    // Using convenience method
    ctx.Replace(target, html)
}()

return target.Skeleton(ui.SkeletonComponent)  // Or SkeletonList, SkeletonPage, SkeletonForm
```

### Skeleton Types
```go
target.Skeleton()                    // Default (3 text lines)
target.Skeleton(ui.SkeletonList)     // List items
target.Skeleton(ui.SkeletonComponent)// Component block
target.Skeleton(ui.SkeletonPage)     // Full page
target.Skeleton(ui.SkeletonForm)     // Form layout
```

---

## Tables

### Simple Table (No Headers)
```go
table := ui.SimpleTable(3, "w-full bg-white rounded")  // 3 columns
table.Field("Name", "font-bold")
table.Field("Age", "text-center")
table.Field("Email", "text-gray-600")
// New row starts automatically after 3 fields
table.Field("John", "font-bold")
table.Field("30", "text-center")
table.Field("john@example.com", "text-gray-600")
table.Render()
```

### Typed Table with Headers
```go
type Person struct { Name string; Age int; Email string }

persons := []*Person{
    {Name: "Alice", Age: 25, Email: "alice@example.com"},
    {Name: "Bob", Age: 30, Email: "bob@example.com"},
}

table := ui.Table[Person]("w-full bg-white rounded")
table.Head("Name", "font-bold")
table.Head("Age", "text-center")
table.Head("Email", "")

// FieldText escapes HTML (safe for user input)
table.FieldText(func(p *Person) string { return p.Name }, "font-bold")
table.FieldText(func(p *Person) string { return fmt.Sprintf("%d", p.Age) }, "text-center")
table.FieldText(func(p *Person) string { return p.Email }, "text-gray-600")

// Field allows raw HTML (use only for trusted content)
table.Field(func(p *Person) string {
    return ui.Button().Color(ui.Blue).Class("rounded").Render("Edit")
}, "")

table.Render(persons)
```

### Table with Colspan
```go
table := ui.SimpleTable(4, "w-full")
table.Field("Spanning 2 columns").Attr(`colspan="2"`)
table.Field("Col 3")
table.Field("Col 4")
table.Render()
```

---

## Data Collation (Search, Sort, Filter, Pagination, Excel Export)

Full-featured data management UI backed by GORM.

### Complete Example
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
    LastLogin time.Time
}

func PeopleList(ctx *ui.Context) string {
    // Define fields
    name := ui.TField{DB: "name", Field: "Name", Text: "Name"}
    surname := ui.TField{DB: "surname", Field: "Surname", Text: "Surname"}
    email := ui.TField{DB: "email", Field: "Email", Text: "Email"}
    country := ui.TField{
        DB: "country", Field: "Country", Text: "Country",
        As: ui.SELECT, Options: ui.MakeOptions([]string{"USA", "UK", "Germany", "France"}),
    }
    status := ui.TField{
        DB: "status", Field: "Status", Text: "Status",
        As: ui.SELECT, Options: ui.MakeOptions([]string{"new", "active", "blocked"}),
    }
    active := ui.TField{DB: "active", Field: "Active", Text: "Active", As: ui.BOOL}
    hasLoggedIn := ui.TField{DB: "last_login", Field: "LastLogin", Text: "Has logged in", As: ui.NOT_ZERO_DATE}
    neverLoggedIn := ui.TField{DB: "last_login", Field: "LastLogin", Text: "Never logged in", As: ui.ZERO_DATE}
    createdAt := ui.TField{DB: "created_at", Field: "CreatedAt", Text: "Created between", As: ui.DATES}

    // Initialize collate
    collate := ui.Collate[Person](&ui.TQuery{
        Limit: 10,
        Order: "surname asc",
    })

    // Configure features
    collate.Search(name, surname, email, country)  // Searchable fields
    collate.Sort(surname, name, email)              // Sortable fields
    collate.Filter(active, hasLoggedIn, neverLoggedIn, createdAt, country, status)  // Filter panel
    collate.Excel(surname, name, email, country, status, active, createdAt)  // Excel export columns

    // Define row rendering
    collate.Row(func(p *Person, idx int) string {
        statusBadge := ui.Span("px-2 py-0.5 rounded text-xs bg-blue-100 text-blue-700")(p.Status)
        activeBadge := ui.Iff(p.Active)(
            ui.Span("px-2 py-0.5 rounded text-xs bg-green-100 text-green-700")("active"),
        )

        return ui.Div("bg-white rounded-lg border p-3 mb-2")(
            ui.Div("flex items-center justify-between")(
                ui.Div("font-semibold")(fmt.Sprintf("%s %s", p.Surname, p.Name)),
                ui.Div("text-gray-500 text-sm")(p.Email),
                ui.Div("flex gap-1")(statusBadge, activeBadge),
            ),
            ui.Div("text-sm text-gray-600 mt-1")(
                fmt.Sprintf("Country: %s | Created: %s", p.Country, p.CreatedAt.Format("2006-01-02")),
            ),
        )
    })

    // Render with database connection
    return collate.Render(ctx, db)
}
```

### TField Configuration
```go
ui.TField{
    DB:        "column_name",      // Database column name
    Field:     "StructField",      // Go struct field name
    Text:      "Display Label",    // UI label
    As:        ui.SELECT,          // Filter type
    Options:   ui.MakeOptions([]string{"A", "B"}),  // For SELECT filters
    Bool:      false,              // Default bool value for BOOL filters
    Condition: " = 1",             // Custom SQL condition for BOOL
}
```

### Filter Types
```go
ui.BOOL           // Checkbox filter (column = 1)
ui.SELECT         // Dropdown filter (requires Options)
ui.DATES          // Date range picker (From/To)
ui.ZERO_DATE      // "Has no date" checkbox (column IS NULL or zero)
ui.NOT_ZERO_DATE  // "Has date" checkbox (column IS NOT NULL and not zero)
```

### Custom Excel Export
```go
collate.OnExcel = func(data *[]Person) (string, io.Reader, error) {
    f := excelize.NewFile()
    // ... custom Excel generation
    filename := fmt.Sprintf("export_%s.xlsx", time.Now().Format("20060102"))
    buffer, _ := f.WriteToBuffer()
    return filename, bytes.NewReader(buffer.Bytes()), nil
}
```

### SQLite Search Normalization
```go
// Register custom normalize function for accent-insensitive search
db, _ := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
ui.RegisterSQLiteNormalize(db)  // Enables searching "café" with "cafe"
```

---

## Stateful Components

```go
// Define component with state
type Counter struct {
    Count int
}

func (c *Counter) Increment(ctx *ui.Context) string {
    ctx.Body(c)  // Restore state from form
    c.Count++
    return c.render(ctx)
}

func (c *Counter) render(ctx *ui.Context) string {
    target := ui.Target()
    return ui.Div("flex gap-2", target)(
        ui.Button().
            Click(ctx.Call(c.Increment, c).Replace(target)).  // Pass state
            Render(fmt.Sprintf("Count: %d", c.Count)),
    )
}

// Usage
func Page(ctx *ui.Context) string {
    counter := &Counter{Count: 0}
    return counter.render(ctx)
}
```

---

## Labels

```go
target := ui.Target()

ui.Label(&target).Render("Field Label")
ui.Label(&target).Required(true).Render("Required Field")
ui.Label(&target).Class("text-lg font-bold").Render("Styled Label")
ui.Label(&target).ClassLabel("text-blue-600").Render("Custom Label Style")
ui.Label(nil).Render("Label without target")  // No `for` attribute
```

---

## Icons (Material Icons)

```go
// Include Material Icons and Google Fonts in HTMLHead
app.HTMLHead = append(app.HTMLHead,
    `<link rel="preconnect" href="https://fonts.googleapis.com">`,
    `<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>`,
    `<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">`,
    `<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">`,
)

// Icon helpers
ui.Icon("check")                             // <span class="material-icons">check</span>
ui.Icon2("check", "text-green-500")         // Icon with extra classes
ui.IconLeft("arrow_back", "Back")            // Icon + text (icon on left)
ui.IconRight("Next", "arrow_forward")        // Text + icon (icon on right)
ui.IconStart("download", "Download")         // Icon at start with gap

// Common Material Icons:
// check, arrow_back, arrow_forward, add, close, search, sort, arrow_upward, 
// arrow_downward, undo, download, tune, inbox, image, home, person, etc.
```

---

## CAPTCHA Components

### Captcha2 - Image-based Challenge
```go
func validated(ctx *ui.Context) string {
    return ui.Div("text-green-600")("CAPTCHA validated!")
}

func FormWithCaptcha(ctx *ui.Context) string {
    return ui.Div("")(
        ui.Captcha2(validated).Render(ctx),
    )
}
```

### Captcha3 - Draggable Tile Challenge
```go
func FormWithCaptcha3(ctx *ui.Context) string {
    onSuccess := func(ctx *ui.Context) string {
        ctx.Success("CAPTCHA passed!")
        return showProtectedContent(ctx)
    }
    
    return ui.Captcha3(onSuccess).
        Count(4).  // Number of tiles to arrange
        Render(ctx)
}
```

### CAPTCHA Configuration
```go
captcha := ui.Captcha2(onValidated)
captcha.SessionField("captcha_session")
captcha.ClientVerifiedField("captcha_verified")
captcha.AnswerField("captcha_answer")
captcha.Attempts(5)                    // Max attempts
captcha.Lifetime(5 * time.Minute)      // Session lifetime
captcha.Render(ctx)

// Server-side validation
if err := captcha.ValidateRequest(ctx.Request); err != nil {
    ctx.Error("Invalid CAPTCHA")
}
```

---

## Theme Switcher

```go
// Renders a button that cycles: System → Light → Dark
ui.ThemeSwitcher("")                    // Default styling
ui.ThemeSwitcher("fixed bottom-4 right-4")  // Custom positioning
```

---

## Hidden Fields

```go
// Preserve state in forms
ui.Hidden("UserID", "uint", 123)
ui.Hidden("Mode", "string", "edit")
ui.Hidden("Filter[0].Field", "string", "name")
```

---

## CSS Constants

### Colors (for buttons/components)
```go
ui.Blue, ui.BlueOutline
ui.Green, ui.GreenOutline
ui.Red, ui.RedOutline
ui.Yellow, ui.YellowOutline
ui.Purple, ui.PurpleOutline
ui.Gray, ui.GrayOutline
ui.White, ui.WhiteOutline
```

### Sizes
```go
ui.XS  // p-1
ui.SM  // p-2
ui.MD  // p-3 (default)
ui.ST  // p-4
ui.LG  // p-5
ui.XL  // p-6
```

### Input Styles
```go
ui.INPUT    // Standard input styling
ui.AREA     // Textarea styling
ui.BTN      // Button base styling
ui.DISABLED // Disabled state styling
```

---

## Common Patterns

### Page with Layout
```go
func main() {
    app := ui.MakeApp("en")
    
    layout := func(title string, body ui.Callable) ui.Callable {
        return func(ctx *ui.Context) string {
            nav := ui.Div("bg-white shadow p-4 flex items-center gap-4")(
                ui.A("hover:text-blue-600", ctx.Load("/"))("Home"),
                ui.A("hover:text-blue-600", ctx.Load("/about"))("About"),
                ui.Flex1,
                ui.ThemeSwitcher(""),
            )
            return app.HTML(title, "bg-gray-100 min-h-screen", 
                nav + ui.Div("max-w-4xl mx-auto p-4")(body(ctx)),
            )
        }
    }
    
    app.Page("/", layout("Home", HomePage))
    app.Page("/about", layout("About", AboutPage))
    app.Listen(":8080")
}
```

### Reusable Form Component
```go
// Reusable form that can be used in multiple places with different callbacks
type TemplateForm struct {
    target      ui.Attr
    Title       string
    Description string
    onSubmit    func(*ui.Context) string
}

func NewTemplateForm(title, description string) *TemplateForm {
    return &TemplateForm{
        target:      ui.Target(),
        Title:       title,
        Description: description,
    }
}

func (f *TemplateForm) OnReset(ctx *ui.Context) string {
    f.Title = ""
    f.Description = ""
    return f.Render(ctx)
}

func (f *TemplateForm) Render(ctx *ui.Context) string {
    return ui.Form("flex flex-col gap-4", f.target, ctx.Submit(f.onSubmit).Replace(f.target))(
        ui.IText("Title", f).Placeholder("Enter title").Render("Title"),
        ui.IArea("Description", f).Rows(4).Placeholder("Enter description").Render("Description"),
        ui.Div("flex gap-4 justify-end")(
            ui.Button().Click(ctx.Call(f.OnReset).Replace(f.target)).Color(ui.Gray).Render("Reset"),
            ui.Button().Submit().Color(ui.Blue).Render("Submit"),
        ),
    )
}

// Usage
func MyPage(ctx *ui.Context) string {
    form1 := NewTemplateForm("Hello", "Description here")
    form1.onSubmit = func(ctx *ui.Context) string {
        ctx.Body(form1)
        ctx.Success("Form 1 submitted!")
        return form1.Render(ctx)
    }
    
    form2 := NewTemplateForm("Another", "Different form")
    form2.onSubmit = func(ctx *ui.Context) string {
        ctx.Body(form2)
        db.Create(&MyModel{Title: form2.Title})
        ctx.Success("Saved to database!")
        return form2.Render(ctx)
    }
    
    return ui.Div("grid grid-cols-2 gap-4")(
        ui.Div("bg-white p-4 rounded")(form1.Render(ctx)),
        ui.Div("bg-white p-4 rounded")(form2.Render(ctx)),
    )
}
```

### CRUD Form Pattern
```go
type Item struct {
    ID   uint
    Name string `validate:"required"`
}

func (item *Item) Save(ctx *ui.Context) string {
    if err := ctx.Body(item); err != nil {
        return item.Form(ctx, &err)
    }
    
    v := validator.New()
    if err := v.Struct(item); err != nil {
        return item.Form(ctx, &err)
    }
    
    db.Save(item)
    ctx.Success("Saved!")
    return item.Form(ctx, nil)
}

func (item *Item) Form(ctx *ui.Context, err *error) string {
    target := ui.Target()
    return ui.Form("space-y-4", target, ctx.Submit(item.Save).Replace(target))(
        ui.ErrorForm(err, nil),
        ui.IText("Name", item).Required().Error(err).Render("Name"),
        ui.Button().Submit().Color(ui.Blue).Render("Save"),
    )
}
```

### Deferred Loading
```go
func LoadData(ctx *ui.Context) string {
    target := ui.Target()
    
    go func() {
        defer func() { recover() }()
        
        time.Sleep(2 * time.Second)
        data := fetchExpensiveData()
        
        html := ui.Div("bg-white p-4 rounded", target)(renderData(data))
        ctx.Patch(target.Replace(), html)
    }()
    
    return target.Skeleton(ui.SkeletonList)
}
```

### SPA-like Navigation
```go
func NavBar(ctx *ui.Context, currentPath string) string {
    routes := []struct{ Path, Title string }{
        {"/", "Home"},
        {"/users", "Users"},
        {"/settings", "Settings"},
    }
    
    return ui.Div("flex gap-2")(
        ui.Map(routes, func(r *struct{ Path, Title string }, _ int) string {
            isActive := r.Path == currentPath
            cls := "px-3 py-2 rounded hover:bg-gray-200"
            if isActive {
                cls = "px-3 py-2 rounded bg-blue-700 text-white"
            }
            // ctx.Load() enables SPA navigation without full page reload
            return ui.A(cls, ctx.Load(r.Path))(r.Title)
        }),
    )
}
```

### Modal/Dialog Pattern
```go
func ConfirmDelete(ctx *ui.Context) string {
    target := ui.Target()
    
    onConfirm := func(ctx *ui.Context) string {
        // Perform delete
        db.Delete(&Item{}, itemID)
        ctx.Success("Deleted!")
        return ""  // Close modal by returning empty
    }
    
    onCancel := func(ctx *ui.Context) string {
        return ""  // Close modal
    }
    
    return ui.Div("fixed inset-0 bg-black/50 flex items-center justify-center", target)(
        ui.Div("bg-white p-6 rounded-lg shadow-xl max-w-md")(
            ui.Div("text-lg font-bold mb-4")("Confirm Delete"),
            ui.Div("text-gray-600 mb-6")("Are you sure you want to delete this item?"),
            ui.Div("flex gap-4 justify-end")(
                ui.Button().Click(ctx.Call(onCancel).Replace(target)).Color(ui.Gray).Render("Cancel"),
                ui.Button().Click(ctx.Call(onConfirm).Replace(target)).Color(ui.Red).Render("Delete"),
            ),
        ),
    )
}
```

### Conditional Rendering
```go
func UserCard(ctx *ui.Context, user *User) string {
    return ui.Div("p-4 bg-white rounded")(
        ui.Div("font-bold")(user.Name),
        
        // Show admin badge only for admins
        ui.Iff(user.IsAdmin)(
            ui.Span("bg-red-100 text-red-700 px-2 py-1 rounded text-xs")("Admin"),
        ),
        
        // Show different buttons based on status
        ui.Or(user.IsActive,
            func() string {
                return ui.Button().Color(ui.Red).Render("Deactivate")
            },
            func() string {
                return ui.Button().Color(ui.Green).Render("Activate")
            },
        ),
        
        // Conditionally show with If
        ui.If(user.HasProfilePic, func() string {
            return ui.Img("rounded-full", ui.Attr{Src: user.ProfilePicURL})
        }),
    )
}
```

### List with Map
```go
func UserList(ctx *ui.Context, users []User) string {
    return ui.Div("space-y-2")(
        ui.Map(users, func(u *User, idx int) string {
            return ui.Div("flex items-center gap-4 p-4 bg-white rounded")(
                ui.Div("text-gray-500 w-8")(fmt.Sprintf("%d.", idx+1)),
                ui.Div("font-bold flex-1")(u.Name),
                ui.Div("text-gray-600")(u.Email),
                ui.Button().Color(ui.BlueOutline).Class("rounded").Render("View"),
            )
        }),
    )
}
```

### Error Handling
```go
func SafeHandler(ctx *ui.Context) string {
    defer func() {
        if r := recover(); r != nil {
            ctx.Error(fmt.Sprintf("An error occurred: %v", r))
        }
    }()
    
    // Your code here
    return doSomethingRisky()
}

// Display form errors
func FormWithErrors(ctx *ui.Context, err *error) string {
    target := ui.Target()
    
    return ui.Form("space-y-4", target, ctx.Submit(handleSubmit).Replace(target))(
        // Show all validation errors at top
        ui.ErrorForm(err, nil),
        
        // Individual field with error highlighting
        ui.IText("Name", &data).Required().Error(err).Render("Name"),
        ui.IEmail("Email", &data).Required().Error(err).Render("Email"),
        
        ui.Button().Submit().Color(ui.Blue).Render("Submit"),
    )
}
```

---

## Sessions

```go
// Sessions require GORM database connection
type SessionData struct {
    UserID   uint
    Username string
    Role     string
}

func LoginHandler(ctx *ui.Context) string {
    session := ctx.Session(db, "auth")
    
    data := &SessionData{
        UserID:   123,
        Username: "john",
        Role:     "admin",
    }
    session.Save(data)  // Save to _session table
    
    ctx.Success("Logged in!")
    return ctx.Redirect("/dashboard")
}

func DashboardHandler(ctx *ui.Context) string {
    session := ctx.Session(db, "auth")
    
    var data SessionData
    session.Load(&data)  // Load from _session table
    
    if data.UserID == 0 {
        return ctx.Redirect("/login")
    }
    
    return ui.Div("")("Welcome, " + data.Username)
}

// Session sweeper (cleanup expired sessions)
app.StartSweeper(24 * time.Hour)  // Clean up daily
```

---

## Security

### Content Security Policy
```go
func SecurePage(ctx *ui.Context) string {
    ctx.SetDefaultCSP()  // Restrictive default policy
    // Or custom:
    ctx.SetCSP("default-src 'self'; script-src 'self' 'unsafe-inline';")
    
    return renderPage()
}
```

### Security Headers
```go
func SecureHandler(ctx *ui.Context) string {
    ctx.SetSecurityHeaders()  // Sets all security headers:
    // - Strict-Transport-Security
    // - X-Frame-Options: DENY
    // - X-Content-Type-Options: nosniff
    // - X-XSS-Protection: 1; mode=block
    // - Referrer-Policy
    // - Permissions-Policy
    
    return renderPage()
}

// Custom security headers
ctx.SetCustomSecurityHeaders(ui.SecurityHeaderOptions{
    CSP:                "default-src 'self'",
    EnableHSTS:         true,
    FrameOptions:       "SAMEORIGIN",
    ContentTypeOptions: true,
    ReferrerPolicy:     "strict-origin",
})
```

### Safe Output
```go
// Always use FieldText/HeadText for user-provided content
table.FieldText(func(p *Person) string { return p.Name }, "")  // Escapes HTML
table.Head("User Input", "")  // Escapes HTML

// Field/HeadHTML allow raw HTML - only for trusted content
table.Field(func(p *Person) string { return trustedHTML }, "")  // No escaping
```

---

## Validation Tags (go-playground/validator)

```go
type Form struct {
    Email     string `validate:"required,email"`
    Age       int    `validate:"gte=0,lte=120"`
    Password  string `validate:"required,min=8"`
    Role      string `validate:"oneof=admin user guest"`
    Website   string `validate:"url"`
    Phone     string `validate:"e164"`
    Agree     bool   `validate:"eq=true"`
}
```

---

## Project Structure

```
myapp/
├── main.go          # App setup, routes, layout
├── pages/
│   ├── home.go      # Page handlers
│   └── users.go
├── components/
│   └── navbar.go    # Reusable components
├── assets/
│   └── favicon.svg
├── go.mod
└── go.sum
```

---

## Dependencies

```go
import (
    "github.com/michalCapo/g-sui/ui"          // Core UI library
    "github.com/go-playground/validator/v10"  // Struct validation
    "gorm.io/gorm"                            // Database ORM (sessions/collate)
    "gorm.io/driver/sqlite"                   // SQLite driver (or postgres/mysql)
)
```

---

## Quick Reference

### Actions Summary
| Method | Returns | Use For |
|--------|---------|---------|
| `ctx.Call(fn).Replace(target)` | JS string | Button onclick |
| `ctx.Click(fn).Replace(target)` | Attr{OnClick} | Element onclick attr |
| `ctx.Submit(fn).Replace(target)` | Attr{OnSubmit} | Form onsubmit attr |
| `ctx.Render(target, html)` | void | WebSocket push (innerHTML) |
| `ctx.Replace(target, html)` | void | WebSocket push (outerHTML) |
| `ctx.Patch(target.Replace(), html)` | void | WebSocket push (full API) |

### Swap Strategies
| Strategy | Effect |
|----------|--------|
| `.Render(target)` | Replace innerHTML |
| `.Replace(target)` | Replace entire element |
| `.Append(target)` | Add to end |
| `.Prepend(target)` | Add to start |
| `.None()` | No DOM update |

### Input Types
| Function | HTML Type |
|----------|-----------|
| `ui.IText()` | text |
| `ui.IEmail()` | email |
| `ui.IPhone()` | tel |
| `ui.IPassword()` | password |
| `ui.INumber()` | number |
| `ui.IDate()` | date |
| `ui.ITime()` | time |
| `ui.IDateTime()` | datetime-local |
| `ui.IArea()` | textarea |
| `ui.ISelect()` | select |
| `ui.ICheckbox()` | checkbox |
| `ui.IRadio()` | radio |
| `ui.IRadioButtons()` | radio group |
| `ui.IRadioDiv()` | card-based radio (custom HTML) |

### Skeleton Types
| Type | Description |
|------|-------------|
| `ui.SkeletonList` | List items with avatars |
| `ui.SkeletonComponent` | Card/component block |
| `ui.SkeletonPage` | Full page with header |
| `ui.SkeletonForm` | Form with inputs |
| (default) | 3 text lines |
# Part II: Architecture Documentation

## Architecture Overview

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
tmpl.Execute(&buf, map[string]any{
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
├── ui.go           # Core HTML DSL, colors, utilities
├── ui.server.go    # App, Context, HTTP server, WebSocket
├── ui.input.go     # Input components and validation
├── ui.form.go      # Form instance with automatic form association
├── ui.data.go      # Data collation (search/sort/filter/paging)
├── ui.button.go    # Button component
├── ui.table.go     # Simple table component
├── ui.label.go     # Form labels
├── ui.check.go     # Checkbox component
├── ui.radio.go     # Radio button component
├── ui.select.go    # Select dropdown component
├── ui.icon.go      # Icon helpers
├── ui.captcha.go   # Google reCAPTCHA integration
├── ui.captcha2.go  # Image-based CAPTCHA
├── ui.captcha3.go  # Tile-based CAPTCHA
├── ui.alert.go     # Alert notification banners
├── ui.badge.go     # Badge status indicators
├── ui.card.go      # Card content containers
├── ui.progress.go  # Progress bar indicators
├── ui.tooltip.go   # Hover tooltips
├── ui.tabs.go      # Tabbed content panels
├── ui.accordion.go # Collapsible sections
└── ui.dropdown.go  # Dropdown menus
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
| `ui.icon.go` | ~47 | Icon helpers (Material Icons) |
| `ui.captcha.go` | ~108 | Google reCAPTCHA integration |
| `ui.captcha2.go` | ~487 | Image CAPTCHA generation and validation |
| `ui.captcha3.go` | ~455 | Tile puzzle CAPTCHA |
| `ui.alert.go` | ~190 | Dismissible alert banners with dark mode |
| `ui.badge.go` | ~90 | Status indicators and notification counts |
| `ui.card.go` | ~120 | Card containers with header/body/footer |
| `ui.progress.go` | ~100 | Progress bars with striped animation |
| `ui.tooltip.go` | ~190 | Hover tooltips with positioning |
| `ui.tabs.go` | ~250 | Tabbed navigation with client-side state |
| `ui.accordion.go` | ~200 | Collapsible sections with toggle |
| `ui.dropdown.go` | ~200 | Context menus with click-outside-to-close |

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
| `IFile` | file | File input |
| `IImageUpload` | image | Image upload with inline preview |
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

### File Input Component

```go
// Basic file input
ui.IFile("image").Accept("image/*").Required().Render("Image")

// With custom ID for linking with ImagePreview
id := ui.RandomString(10)
ui.IFile("image").
    ID(id).                    // Set custom ID
    Accept("image/*").          // MIME types: "image/*", ".pdf,.doc", etc.
    Multiple().                 // Allow multiple files
    Required().
    Disabled(false).
    Class("custom-wrapper").    // Wrapper classes
    ClassInput("custom-input"). // Input element classes
    ClassLabel("custom-label"). // Label classes
    Change("console.log('changed')").  // OnChange handler
    If(condition).              // Conditional render
    Render("Upload Image")
```

**File Input Methods:**
- `.ID(id)` - Set custom ID (useful for linking with ImagePreview)
- `.GetID()` - Get the file input's ID
- `.Accept(types)` - Allowed file types (e.g., `"image/*"`, `".pdf,.doc"`)
- `.Multiple()` - Allow selecting multiple files
- `.Required()`, `.Disabled()` - State modifiers
- `.Class(s)`, `.ClassInput(s)`, `.ClassLabel(s)` - Custom classes
- `.Change(action)` - OnChange handler (JavaScript string)
- `.If(condition)` - Conditional render
- `.Form(id)` - Associate with form by ID

**File Upload Handler (Single File):**

```go
func uploadHandler(ctx *ui.Context) string {
    file, err := ctx.File("image")
    if err != nil {
        ctx.Error("Failed to process file: " + err.Error())
        return renderUploadForm(ctx)
    }
    if file == nil {
        ctx.Error("No file uploaded")
        return renderUploadForm(ctx)
    }

    // Validate file type
    if !strings.HasPrefix(file.ContentType, "image/") {
        ctx.Error("File must be an image")
        return renderUploadForm(ctx)
    }

    // Validate file size (max 5MB)
    if file.Size > 5*1024*1024 {
        ctx.Error("Image size must be less than 5MB")
        return renderUploadForm(ctx)
    }

    // Save file
    os.WriteFile("uploads/"+file.Name, file.Data, 0644)

    ctx.Success("File uploaded successfully!")
    return renderUploadForm(ctx)
}
```

**File Upload Handler (Multiple Files):**

```go
func uploadMultipleHandler(ctx *ui.Context) string {
    files, err := ctx.Files("images")  // Use .Multiple() on file input
    if err != nil {
        ctx.Error("Failed to process files: " + err.Error())
        return renderUploadForm(ctx)
    }
    if len(files) == 0 {
        ctx.Error("No files uploaded")
        return renderUploadForm(ctx)
    }

    var savedCount int
    for _, file := range files {
        // Validate file type
        if !strings.HasPrefix(file.ContentType, "image/") {
            continue
        }

        // Validate file size (max 5MB)
        if file.Size > 5*1024*1024 {
            continue
        }

        // Save file
        os.WriteFile("uploads/"+file.Name, file.Data, 0644)
        savedCount++
    }

    ctx.Success(fmt.Sprintf("%d files uploaded successfully!", savedCount))
    return renderUploadForm(ctx)
}
```

**File Object Properties:**
- `file.Name` - Original filename
- `file.Data` - File contents as `[]byte`
- `file.ContentType` - MIME type (e.g., `"image/png"`)
- `file.Size` - File size in bytes

### Image Preview Component

```go
// Basic image preview (single file)
id := ui.RandomString(10)
ui.IFile("image").ID(id).Accept("image/*").Render("Image")
ui.ImagePreview(id).
    MaxSize("320px").          // Max width/height
    Render()

// Multiple file preview (grid layout)
ui.ImagePreview(id).
    Multiple().                // Enable grid layout for multiple images
    MaxSize("200px").
    Class("my-4").             // Custom wrapper classes
    Render()

// With conditional rendering
ui.ImagePreview(id).
    MaxSize("320px").
    If(showPreview).
    Render()
```

**ImagePreview Methods:**
- `.Multiple()` - Enable grid layout for multiple images (default: single centered preview)
- `.MaxSize(size)` - Maximum image dimensions (e.g., `"320px"`, `"200px"`)
- `.Class(classes...)` - Custom wrapper classes
- `.If(condition)` - Conditional render
- `.Render()` - Generate HTML and JavaScript

**Usage Pattern:**

```go
id := ui.RandomString(10)

fileInput := form.File("image").
    ID(id).                    // Set ID
    Accept("image/*").
    Required()

fileInput.Render("Image")

// Image preview component (reuses the file input ID)
ui.ImagePreview(id).
    MaxSize("320px").
    Render()
```

The ImagePreview component automatically:
- Listens to the file input's `change` event
- Uses `FileReader` to preview images client-side
- Only shows previews for image files
- Clears previous previews on new selection
- Supports both single and multiple file selection

### Image Upload Component (Combined File + Preview)

The `ImageUpload` component combines file input and image preview into a single unified component with inline preview:

```go
// Basic image upload with inline preview
form.ImageUpload("image").
    Zone("Add Image", "Click to upload").
    MaxSize("320px").
    Required().
    Render("Image")

// With custom zone styling and icon
form.ImageUpload("image").
    Zone("Add Vehicle Photo", "Click to take or upload").
    ZoneIcon(ui.Icon("image", ui.Attr{Class: "text-5xl"})).  // Using Icon() component
    MaxSize("320px").
    ClassPreview("mt-4").
    Required().
    Render("VEHICLE PHOTO")

// Or with CSS classes directly
form.ImageUpload("image").
    Zone("Add Image", "Click to upload").
    ZoneIcon("w-10 h-10 bg-gray-500 rounded-full p-2 flex items-center justify-center").
    MaxSize("320px").
    Render("Image")
```

**Key Features:**
- **Inline Preview**: Selected image appears inside the upload zone (replacing the upload UI)
- **Change Button**: Built-in "Change Image" button to re-select images
- **Unified Experience**: Single component instead of separate File + ImagePreview
- **Zone Mode**: Uses dropzone-style UI by default for better UX
- **Auto-accept**: Defaults to `accept="image/*"` for images

**ImageUpload Methods:**
- `.Zone(title, hint)` - Enable dropzone mode with title and hint text
- `.ZoneIcon(html)` - Custom icon HTML for zone mode (e.g., `ui.Icon("image")` or CSS classes)
- `.ZoneContent(html)` - Completely custom HTML content for zone (overrides icon/title/hint)
- `.ClassZone(classes...)` - Zone container CSS classes
- `.MaxSize(size)` - Maximum image dimensions for preview (e.g., `"320px"`)
- `.ClassPreview(classes...)` - Preview container CSS classes
- `.Accept(types)` - Override default `"image/*"` (e.g., `"image/png,image/jpeg"`)
- `.Multiple()` - Allow selecting multiple files (preview shows first image)
- `.Required()`, `.Disabled()` - State modifiers
- `.Class(s)`, `.ClassInput(s)`, `.ClassLabel(s)` - Custom classes
- `.Change(action)` - OnChange handler (JavaScript string)
- `.If(condition)` - Conditional render
- `.Form(id)` - Associate with form by ID
- `.ID(id)`, `.GetID()` - Custom ID management

**Usage Pattern:**

```go
// Single component (recommended for images)
form.ImageUpload("image").
    Zone("Add Photo", "Click to upload").
    MaxSize("320px").
    Render("Photo")

// Alternative: Separate File + ImagePreview (still available)
id := ui.RandomString(10)
form.File("image").ID(id).Accept("image/*").Render("Image")
ui.ImagePreview(id).MaxSize("320px").Render()
```

The ImageUpload component automatically:
- Shows upload zone initially (with icon/title/hint)
- On file selection: hides upload zone, shows preview with image and "Change Image" button
- On "Change Image" click: triggers file input again
- Handles empty/no file cases by showing upload zone again
- Only accepts image files (defaults to `accept="image/*"`)

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

// RadioDiv - Card-based selection with custom HTML
cardOptions := []ui.AOption{
    {ID: "1", Value: ui.Div("h-20 w-full bg-blue-100 flex items-center justify-center")("Option 1")},
    {ID: "2", Value: ui.Div("h-20 w-full bg-green-100 flex items-center justify-center")("Option 2")},
    {ID: "3", Value: ui.Div("h-20 w-full bg-purple-100 flex items-center justify-center")("Option 3")},
}
ui.IRadioDiv("Plan", &data).
    Options(cardOptions).
    Render("Choose a plan")
```

**RadioDiv** renders radio buttons as selectable cards with custom HTML content. Each option's `Value` can be any HTML string, allowing rich visual selections like pricing cards, color swatches, or image thumbnails.

### Select Component

```go
ui.TSelect("Country", &data).
    Options(options).
    Required().
    Render("Country")
```

### Icon Helpers

```go
ui.Icon("check")                    // <span class="material-icons">check</span>
ui.Icon2("check", "text-green-500") // Icon with classes
ui.IconLeft("arrow_back", "Back")    // Icon + text
ui.IconRight("Next", "arrow_forward")  // Text + icon
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

## UI Components

### Alert - Notification Banners

Dismissible notification banners for info, success, warning, and error messages with dark mode support, optional titles, and localStorage persistence.

```go
ui.Alert().Message("Important information").Variant("info").Dismissible(true).Render()
ui.Alert().Title("Heads up!").Message("Changes saved!").Variant("success").Dismissible(true).Render()
ui.Alert().Message("Please review carefully").Variant("warning-outline").Dismissible(true).Render()
ui.Alert().Message("Something went wrong").Variant("error").Persist("error-alert").Dismissible(true).Render()
```

**Variants**: `"info"` (default), `"success"`, `"warning"`, `"error"`, `"info-outline"`, `"success-outline"`, `"warning-outline"`, `"error-outline"`

**Methods**:
- `.Message(text string)` - Set alert message
- `.Title(text string)` - Set optional alert title
- `.Variant(variant string)` - Set alert type (info/success/warning/error with optional -outline suffix)
- `.Dismissible(bool)` - Show close button (removes from DOM)
- `.Persist(key string)` - Use localStorage to remember dismissal ("don't show again")
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### Badge - Status Indicators

Small status indicators, notification counts, and labels with dark mode support, icons, and size variants.

```go
// Dot variant (notification indicator)
ui.Badge().Color("red").Dot().Size("lg").Render()

// Text/number badge with icon
icon := `<svg>...</svg>`
ui.Badge().Color("blue").Text("3").Icon(icon).Size("md").Render()

// Label badge with soft variant
ui.Badge().Color("green-soft").Text("Online").Square().Render()
```

**Colors**: `"red"`, `"green"`, `"blue"`, `"yellow"`, `"purple"`, `"gray"` (or use constants like `ui.Red`). Add `-soft` or `-outline` suffix for variants.

**Methods**:
- `.Text(text string)` - Set badge text
- `.Color(color string)` - Set badge color (supports -soft and -outline variants)
- `.Dot()` - Use dot variant (small circle, no text)
- `.Icon(html string)` - Add icon HTML before text
- `.Size(value string)` - Set size: `"sm"`, `"md"` (default), `"lg"`
- `.Square()` - Use square corners instead of rounded
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### Card - Content Containers

Consistent containers with optional header, body, footer, images, and hover effects.

```go
ui.Card().
    Header("<h3 class='font-bold'>Card Title</h3>").
    Body("<p class='text-gray-600'>Card content goes here.</p>").
    Footer("<div class='text-sm text-gray-500'>Card footer</div>").
    Variant(ui.CardBordered).
    Render()

// Card with image and hover effect
ui.Card().
    Image("https://example.com/image.jpg", "Alt text").
    Header("<h3>Card with Image</h3>").
    Body("<p>Content</p>").
    Hover(true).
    Render()

// Glass variant
ui.Card().
    Variant(ui.CardGlass).
    Header("<h3>Glass Card</h3>").
    Body("<p>Glassmorphism effect</p>").
    Render()
```

**Variants**: `ui.CardShadowed` (default), `ui.CardBordered`, `ui.CardFlat`, `ui.CardGlass`

**Methods**:
- `.Header(html string)` - Set header HTML
- `.Body(html string)` - Set body HTML
- `.Footer(html string)` - Set footer HTML
- `.Image(src string, alt string)` - Add image at top of card
- `.Variant(variant string)` - Set card style
- `.Hover(bool)` - Enable hover effect (shadow lift)
- `.Compact(bool)` - Use compact padding
- `.Padding(value string)` - Custom padding class
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### Progress Bar - Progress Indicators

Visual progress indicators with gradients, labels, indeterminate mode, and striped animation.

```go
ui.ProgressBar().Value(75).Render()
ui.ProgressBar().Value(50).Striped(true).Animated(true).Render()
ui.ProgressBar().Value(75).Gradient("#3b82f6", "#8b5cf6").Render()
ui.ProgressBar().Value(45).Label("Loading...").LabelPosition("outside").Render()
ui.ProgressBar().Indeterminate(true).Color("bg-blue-600").Render()
```

**Methods**:
- `.Value(percent int)` - Set progress value (0-100)
- `.Color(cssClass string)` - Custom color class (default: `bg-blue-600`)
- `.Gradient(colors ...string)` - Use gradient colors (overrides Color)
- `.Striped(bool)` - Show striped pattern
- `.Animated(bool)` - Animate stripes (requires Striped)
- `.Indeterminate(bool)` - Show indeterminate progress (animated bar)
- `.Size(value string)` - Set height: `"xs"`, `"sm"`, `"md"` (default), `"lg"`, `"xl"`
- `.Label(text string)` - Optional label text
- `.LabelPosition(value string)` - `"inside"` (default) or `"outside"`
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### StepProgress - Step Progress Indicator

A step progress indicator that shows the current step out of total steps with a visual progress bar. Useful for multi-step forms, wizards, and onboarding flows.

```go
ui.StepProgress(1, 4).Render()                    // Step 1 of 4
ui.StepProgress(2, 5).Color("bg-purple-500").Render()  // Step 2 of 5 with custom color
ui.StepProgress(3, 10).Size("lg").Render()        // Step 3 of 10 with large size
ui.StepProgress(4, 4).Color("bg-green-500").Render()  // Complete
```

**Methods**:
- `.Current(value int)` - Set the current step
- `.Total(value int)` - Set the total number of steps
- `.Color(cssClass string)` - Custom color class (default: `bg-blue-500`)
- `.Size(value string)` - Set height: `xs`, `sm`, `md` (default), `lg`, `xl`
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

**Sizes**:
- `xs` - Extra thin (h-0.5)
- `sm` - Small (h-1)
- `md` - Medium (h-1.5)
- `lg` - Large (h-2)
- `xl` - Extra large (h-3)

### Tooltip - Hover Tooltips

Hover tooltips with positioning, multiple variants, and configurable delay.

```go
ui.Tooltip().Content("Tooltip text").Position("top").Render(
    ui.Button().Color(ui.Blue).Render("Hover me"),
)

ui.Tooltip().Content("Delayed tooltip").Delay(500).Variant("green").Render(
    ui.Button().Color(ui.Green).Render("Hover me"),
)
```

**Positions**: `"top"` (default), `"bottom"`, `"left"`, `"right"`

**Variants**: `"dark"` (default), `"light"`, `"blue"`, `"green"`, `"red"`, `"yellow"`

**Methods**:
- `.Content(text string)` - Set tooltip text
- `.Position(position string)` - Set tooltip position
- `.Variant(variant string)` - Set appearance style
- `.Delay(ms int)` - Set show/hide delay in milliseconds (default: 200)
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### Tabs - Tabbed Content

Tabbed navigation for organizing content into switchable panels with icons and multiple styles. Client-side state management.

```go
ui.Tabs().
    Tab("Overview", "<div class='p-4'>Overview content</div>").
    Tab("Features", "<div class='p-4'>Features content</div>", iconHTML).
    Tab("Settings", "<div class='p-4'>Settings content</div>").
    Active(0).                      // Set initially active tab (0-indexed)
    Style("underline").             // "underline", "pills", "boxed", or "vertical"
    Render()
```

**Styles**: `"underline"` (default), `"pills"`, `"boxed"`, `"vertical"`

**Methods**:
- `.Tab(label string, content string, icon ...string)` - Add a tab panel with optional icon HTML
- `.Active(index int)` - Set initially active tab (0-indexed)
- `.Style(style string)` - Set tab style
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### Accordion - Collapsible Sections

Vertically stacked collapsible content sections with client-side toggle and multiple variants.

```go
// Single section open at a time (bordered variant)
ui.Accordion().
    Item("Section 1", "Content for section 1", true).  // Third param: initially open
    Item("Section 2", "Content for section 2").
    Item("Section 3", "Content for section 3").
    Render()

// Multiple sections can be open (separated variant)
ui.Accordion().
    Item("Section 1", "Content for section 1").
    Item("Section 2", "Content for section 2").
    Variant(ui.AccordionSeparated).
    Multiple(true).
    Render()

// Ghost variant (minimal styling)
ui.Accordion().
    Variant(ui.AccordionGhost).
    Item("Section 1", "Content").
    Render()
```

**Variants**: `ui.AccordionBordered` (default), `ui.AccordionGhost`, `ui.AccordionSeparated`

**Methods**:
- `.Item(title string, content string, open ...bool)` - Add collapsible section (third param: initially open)
- `.Multiple(bool)` - Allow multiple sections open simultaneously
- `.Variant(variant string)` - Set accordion style
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

### Dropdown Menu - Context Menus

Contextual action menus that appear on click with headers, dividers, danger items, icons, and multiple positions.

```go
iconEdit := `<svg>...</svg>`
iconDelete := `<svg>...</svg>`

ui.Dropdown().
    Trigger(ui.Button().Color(ui.Blue).Render("Options ▼")).
    Header("General").
    Item("Edit", "alert('Edit')", iconEdit).
    Item("Duplicate", "alert('Duplicate')").
    Divider().
    Header("Danger Zone").
    Danger("Delete", "alert('Delete')", iconDelete).
    Position("bottom-left").
    Render()
```

**Positions**: `"bottom-left"` (default), `"bottom-right"`, `"top-left"`, `"top-right"`

**Methods**:
- `.Trigger(html string)` - Set trigger element HTML
- `.Item(label string, onclick string, icon ...string)` - Add menu item with optional icon
- `.Danger(label string, onclick string, icon ...string)` - Add danger-variant item (red styling)
- `.Header(label string)` - Add non-interactive header label
- `.Divider()` - Add visual separator
- `.Position(value string)` - Set dropdown position
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

**Positions**: `"bottom-left"`, `"bottom-right"`, `"top-left"`, `"top-right"`

**Methods**:
- `.Trigger(html string)` - Set the trigger button/element
- `.Item(label string, action string)` - Add menu item (action is inline JS)
- `.Divider()` - Add separator line
- `.Position(position string)` - Set dropdown position
- `.If(condition)` - Conditional rendering
- `.Class(...classes)` - Custom CSS classes

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
| `.RadioDiv(name, data...)` | `*ARadio` | Card-based radio (custom HTML) |
| `.File(name)` | `*TFile` | File input |
| `.ImageUpload(name)` | `*TImageUpload` | Image upload with inline preview |
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
- `Page(path string, title string, handler Callable)` - Register page route with title
  - Supports parameterized routes: `app.Page("/user/{id}", "Title", handler)`
  - Path parameters use curly braces: `/user/{id}`, `/user/{userId}/post/{postId}`
- `Action(path string, handler Callable)` - Register server action
- `Listen(addr string)` - Start HTTP and WebSocket server
- `Favicon(fs embed.FS, path string, maxAge time.Duration)` - Serve favicon
- `Assets(fs embed.FS, path string, maxAge time.Duration)` - Serve static assets
- `StartSweeper(interval time.Duration)` - Start session cleanup goroutine
- `TestHandler() http.Handler` - Get HTTP handler for testing (returns handler without starting server)
- `PWA(config PWAConfig)` - Enable Progressive Web App support

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
- `Body(data any) error` - Parse form/JSON into struct

#### Route Parameters
- `PathParam(name string) string` - Get path parameter from route pattern (e.g., `/user/{id}`)
  - Returns empty string if parameter doesn't exist
  - Example: `userID := ctx.PathParam("id")` for route `/user/{id}`
- `QueryParam(name string) string` - Get query parameter from URL (e.g., `?name=Smith`)
  - Returns first value for multi-value params, empty string if not found
  - Works with both SPA navigation (`ctx.Load`) and direct requests
  - Example: `name := ctx.QueryParam("name")` for URL `/search?name=Smith`
- `QueryParams(name string) []string` - Get all values for a query parameter
  - Returns `nil` if parameter doesn't exist
  - Example: `tags := ctx.QueryParams("tag")` for URL `/tags?tag=a&tag=b`
- `AllQueryParams() map[string][]string` - Get all query parameters as a map
  - Returns `nil` if no query parameters exist
  - Falls back to `Request.URL.Query()` for direct requests

**Example: Using Path and Query Parameters Together**

```go
// Route: app.Page("/user/{id}", userDetailHandler)
// URL: /user/123?tab=profile&view=detailed&sort=name&order=asc

func userDetailHandler(ctx *ui.Context) string {
    // Get path parameters from route pattern
    userID := ctx.PathParam("id")
    
    // Get query parameters (if any) using ctx.QueryParam() - works with SPA navigation
    tab := ctx.QueryParam("tab")
    view := ctx.QueryParam("view")
    sort := ctx.QueryParam("sort")
    order := ctx.QueryParam("order")
    
    // Use the parameters...
    return ui.Div("")(fmt.Sprintf("User %s, tab: %s, view: %s", userID, tab, view))
}
```

**Important:** `QueryParam()` works seamlessly with both SPA navigation (via `ctx.Load()`) and direct HTTP requests, so query parameters are preserved when navigating with `ctx.Load("/user/123?tab=profile")`.

#### User Feedback (Toasts)
- `Success(msg string)` - Green toast notification
- `Error(msg string)` - Red toast notification
- `Info(msg string)` - Blue toast notification
- `ErrorReload(msg string)` - Red toast with reload button

#### Page Title
- `Title(title string)` - Update the page title dynamically

#### Navigation
- `Load(href string) Attr` - SPA-like navigation with background loading (returns Attr with href and onclick)
  - Sets both `href` attribute and `onclick` handler for accessibility and SPA navigation
  - Fetches page content in the background
  - Shows loader only if fetch takes longer than 50ms
  - Replaces page content seamlessly without full reload
  - Updates browser history and title
  - Supports right-click "Open in new tab", middle-click, and other native browser behaviors
- `Reload() string` - JavaScript to reload page
- `Redirect(url string) string` - JavaScript to navigate to URL


#### Session Management
- `Session(db *gorm.DB, name string) *Session` - Get session by name
- `Session.Load(data any)` - Load session data into struct
- `Session.Save(data any)` - Save struct to session

#### File Downloads
- `DownloadAs(reader io.Reader, contentType, filename string)` - Send file as download

#### Security Headers
- `SetDefaultCSP()` - Set default Content Security Policy
- `SetCSP(policy string)` - Set custom CSP
- `SetSecurityHeaders()` - Set all security headers
- `SetCustomSecurityHeaders(opts SecurityHeaderOptions)` - Custom security headers

#### WebSocket Patches
- `Render(target Attr, html string)` - Render HTML inside target (innerHTML)
- `Replace(target Attr, html string)` - Replace target element (outerHTML)
- `Patch(swap TargetSwap, html string)` - Server-initiated DOM update (full API)
- `Patch(swap TargetSwap, html string, cleanup func())` - With cleanup callback

---

## Smooth Navigation

g-sui provides seamless, SPA-like navigation with background loading and intelligent loader display.

### Background Loading

When using `ctx.Load()`, navigation works as follows:

1. **Immediate Fetch**: On link click, the fetch starts immediately in the background
2. **Delayed Loader**: A loader overlay appears only if the fetch takes longer than 50ms
3. **Instant Rendering**: If the fetch completes quickly (< 50ms), content renders without showing a loader
4. **Seamless Replacement**: Page content is replaced without a full page reload
5. **History Management**: Browser history and page title are updated automatically

### Navigation with `ctx.Load()`

Use `ctx.Load()` to enable smooth navigation on specific links:

```go
ui.A("px-2 py-1 rounded", ctx.Load("/about"))("About")
```

This creates an `<a>` element with both `href` attribute and an onclick handler that calls `__load()` for smooth navigation. The `href` attribute ensures accessibility (right-click menu, middle-click, status bar preview) while the onclick handler provides SPA navigation.

### Implementation Details

The smooth navigation system uses:

**`__load(href)` JavaScript Function**: Handles the background fetch and DOM replacement
- Starts fetch immediately
- Sets a 50ms timer for loader display
- Cancels loader if fetch completes quickly
- Replaces `document.body.innerHTML` with fetched content
- Executes scripts from the new page
- Updates browser history

### Example: Navigation Bar

```go
func NavBar(ctx *ui.Context) string {
    routes := []struct{ Path, Title string }{
        {"/", "Home"},
        {"/about", "About"},
        {"/contact", "Contact"},
    }
    
    return ui.Div("flex gap-2")(
        ui.Map(routes, func(r *struct{ Path, Title string }, _ int) string {
            isActive := ctx.Request.URL.Path == r.Path
            cls := "px-3 py-2 rounded"
            if isActive {
                cls += " bg-blue-700 text-white"
            } else {
                cls += " hover:bg-gray-200"
            }
            
            return ui.A(cls, ctx.Load(r.Path))(r.Title)
        }),
    )
}
```

---

## PWA (Progressive Web App)

g-sui has built-in support for Progressive Web App capabilities, allowing your app to be "installed" on mobile and desktop devices.

### Enabling PWA

```go
app.PWA(ui.PWAConfig{
    Name:                  "My Application",
    ShortName:             "MyApp",
    ID:                    "/",                              // App ID (defaults to StartURL if empty)
    Description:           "A full-featured g-sui application",
    ThemeColor:            "#1d4ed8",
    BackgroundColor:       "#ffffff",
    Display:               "standalone",
    StartURL:              "/",
    GenerateServiceWorker: true,
    CacheAssets:           []string{"/assets/app.css", "/assets/app.js"}, // Assets to pre-cache
    OfflinePage:           "/offline",                       // Fallback when offline
    Icons: []ui.PWAIcon{
        {Src: "/favicon.ico", Sizes: "any", Type: "image/x-icon"},
        {Src: "/icon-192.png", Sizes: "192x192", Type: "image/png", Purpose: "any"},
        {Src: "/icon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "any maskable"},
    },
})
```

### Manifest Generation
The framework automatically generates a `manifest.webmanifest` file and serves it at `/manifest.webmanifest`. The manifest includes:
- App name, short name, and description
- **App ID** - Unique identifier for the app (defaults to `StartURL`)
- Theme color and background color
- Display mode (standalone, fullscreen, etc.)
- Icons configuration with **purpose attribute** for adaptive/maskable icons
- Start URL

The framework also automatically adds the necessary `<link>` and `<meta>` tags to the HTML head:
- `<link rel="manifest" href="/manifest.webmanifest">`
- Mobile web app capable meta tags for iOS and Android
- Theme color meta tag

### Service Worker
When `GenerateServiceWorker` is `true`, g-sui:
- Generates and serves a service worker at `/sw.js`
- Automatically registers the service worker in the client via inline script
- Uses **network-first** strategy for pages (always fresh content from server)
- Uses **cache-first** strategy for assets in `CacheAssets` (fast loading)
- Generates a unique cache key on each server restart (auto-invalidation)
- Cleans up old caches automatically on activation
- Uses `skipWaiting()` and `clients.claim()` for immediate activation

The service worker ensures:
- **Pages always get fresh content** from the server on new deployments
- **Offline fallback** to `OfflinePage` (or `/`) when network fails
- **Fast asset loading** from cache

### PWA Configuration Options

The `PWAConfig` struct supports the following fields:

```go
type PWAConfig struct {
    Name                  string    // Full application name
    ShortName             string    // Short name for app launcher
    ID                    string    // App ID (defaults to StartURL if empty)
    Description           string    // App description
    ThemeColor            string    // Theme color (hex format, e.g., "#1d4ed8")
    BackgroundColor       string    // Background color (hex format)
    Display               string    // Display mode: "standalone", "fullscreen", "minimal-ui", "browser"
    StartURL              string    // Start URL (defaults to "/")
    GenerateServiceWorker bool      // Enable service worker generation
    CacheAssets           []string  // Asset URLs to pre-cache (cache-first strategy)
    OfflinePage           string    // Fallback page when offline (e.g., "/offline")
    Icons                 []PWAIcon // Array of app icons
}

type PWAIcon struct {
    Src     string // Icon source path
    Sizes   string // Icon sizes (e.g., "192x192", "512x512", "any")
    Type    string // MIME type (e.g., "image/png", "image/x-icon")
    Purpose string // Icon purpose: "any", "maskable", or "any maskable"
}
```

### Icon Purpose Values

The `Purpose` field specifies how an icon can be used by the browser:
- `"any"` - Icon can be used as a standard app icon
- `"maskable"` - Icon can be masked/safely cropped for adaptive icons (Android)
- `"any maskable"` - Icon serves both purposes

Using maskable icons ensures your app displays correctly on Android devices with adaptive icon layouts.

### Benefits of PWA Support

- **Installable**: Users can install your app on their devices
- **Offline Support**: Basic offline functionality with service worker caching
- **App-like Experience**: Standalone display mode removes browser UI
- **Branding**: Custom icons and theme colors for a native app feel
- **Cross-platform**: Works on iOS, Android, and desktop browsers
- **Modern Standards**: Complies with latest PWA specifications (eliminates Chrome DevTools warnings)

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
// Convenience methods (recommended)
ctx.Render(target, html)   // Replace innerHTML
ctx.Replace(target, html)  // Replace element

// Full Patch API with all swap strategies
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
// For append/prepend, use full Patch API
ctx.Patch(notificationTarget.Append(), notificationHTML)

// For render/replace, use convenience methods
statusTarget := ui.ID("status-indicator")
ctx.Render(statusTarget, "<div>Online</div>")  // Update innerHTML
ctx.Replace(statusTarget, "<div id='status-indicator'>Offline</div>")  // Replace element
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

## Releases and Versioning

### Creating a New Release

To create and push a new version tag, use the `deploy` script:

```bash
./deploy
```

The script automatically:
- Starts at version `v0.100` if no tags exist
- Increments the minor version by 1 (e.g., `v0.100` → `v0.101` → `v0.102`)
- Ensures your working tree is clean (no uncommitted changes)
- Creates an annotated git tag with a release message
- Pushes the tag to the remote repository

### Version Numbering Scheme

The project uses semantic versioning with the format `v0.XXX`:
- Major version: Fixed at `0` (pre-1.0 release)
- Minor version: Auto-incremented starting from `100` (`v0.100`, `v0.101`, `v0.102`, ...)

### After Deployment

After running `./deploy`, you can:
1. Create a GitHub release at https://github.com/michalCapo/g-sui/releases/new
2. Select the newly created tag
3. Add release notes describing the changes
4. Publish the release

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
