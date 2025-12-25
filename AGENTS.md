# g-sui LLM Reference

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
app.Page("/path", handler)                        // Register page route
app.Favicon(embedFS, "assets/favicon.svg", 24*time.Hour)
app.Assets(embedFS, "assets/", 24*time.Hour)      // Serve static files
app.AutoRestart(true)                             // Dev: rebuild on file changes
app.Listen(":8080")                               // Start server (also starts WS at /__ws)
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
ui.List(class, attr...)(children...)   // <ul>
ui.ListItem(class, attr...)(children...)// <li>
ui.Img(class, attr...)                 // <img /> (self-closing)
ui.Input(class, attr...)               // <input /> (self-closing)
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

## Icons (FontAwesome)

```go
// Include FontAwesome in HTMLHead
app.HTMLHead = append(app.HTMLHead,
    `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">`,
)

// Icon helpers
ui.Icon("fa fa-check")                       // <i class="fa fa-check"></i>
ui.Icon2("fa fa-check", "text-green-500")    // Icon with extra classes
ui.IconLeft("fa fa-arrow-left", "Back")      // Icon + text (icon on left)
ui.IconRight("Next", "fa fa-arrow-right")    // Text + icon (icon on right)
ui.IconStart("fa fa-download", "Download")   // Icon at start with gap
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
                ui.A("hover:text-blue-600", ui.Href("/"), ctx.Load("/"))("Home"),
                ui.A("hover:text-blue-600", ui.Href("/about"), ctx.Load("/about"))("About"),
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
            return ui.A(cls, ui.Href(r.Path), ctx.Load(r.Path))(r.Title)
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

### Skeleton Types
| Type | Description |
|------|-------------|
| `ui.SkeletonList` | List items with avatars |
| `ui.SkeletonComponent` | Card/component block |
| `ui.SkeletonPage` | Full page with header |
| `ui.SkeletonForm` | Form with inputs |
| (default) | 3 text lines |
