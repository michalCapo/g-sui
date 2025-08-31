![go-srui](https://github.com/user-attachments/assets/ec51e978-8e53-4cb1-8e24-bb5e0494453a)

# Go-SRUI (Server-Rendered UI)

A lightweight, server-rendered UI framework for Go that enables building interactive web applications from your Go code. Define your UI in Go and let the framework handle the rest.

## Features

- No JavaScript required
- No build step required
- No need to learn new syntax for writing HTML
- No plugins required for writing HTML
- Server-side rendering with client-side interactivity
- Built-in form handling and validation
- Automatic WebSocket-based live reload for development
- Tailwind CSS integration
- Session management
- File upload and download support
- Customizable HTML components
- Built-in form validation using go-playground/validator
- Support for various input types (text, email, phone, number, date, time, etc.)

## Installation

```bash
go get github.com/michalCapo/go-srui
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/michalCapo/go-srui/ui"
)

func main() {
    // Create a new app instance
    app := ui.MakeApp("en")
    
    // Enable live reload for development
    app.Autoreload(true)

    // Define a page handler
    page := func(ctx *ui.Context) string {
        return app.HTML("My App", "p-8 bg-gray-200",
            ui.Div("flex flex-row gap-4")(
                ui.Div("flex justify-start gap-4 items-center")(
                    "Hello, World!",
                ),
            ),
        )
    }

    // Register the page
    app.Page("/", page)
    
    // Start the server
    app.Listen(":8080")
}
```

## How it works

Basically as HTMX, but with Go. You define your components or actions in Go. Register these actions so framework can use them. Component is a function that returns HTML. Action is a function that returns HTML or calls another action. When you call action, it will be executed on server and result will be rendered in HTML document.

There are several ways to call actions:
- `ctx.Call(action, values...)` - call action with values
- `ctx.Submit(action)` - submit form with action
- `ctx.Click(action)` - click button with action
- `ctx.Send(action)` - send form with action

Depending on your needs you can render result of action in different places:
- `ctx.Replace(target)` - replace target with result
- `ctx.Render(target)` - render result inside target
- `ctx.None()` - do not render result

Code base is very simple, so please check examples and source code to see how it works.

## Components

### Basic Components

- `Div` - Container element
- `Span` - Inline text element
- `Form` - Form container
- `Button` - Button element
- `Select` - Dropdown select
- `Textarea` - Multi-line text input
- `A` - Link element
- `Table` - Table element
- `Label` - Label element

### Input Types

- `IText` - Text input
- `IPassword` - Password input
- `INumber` - Numeric input
- `IDate` - Date input
- `ITime` - Time input
- `IDateTime` - DateTime input
- `IArea` - Textarea input
- `IValue` - Display-only value field

### Form Handling

```go
type LoginForm struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required"`
}

func (form *LoginForm) Login(ctx *ui.Context) string {
    if err := ctx.Body(form); err != nil {
        return form.Render(ctx, &err)
    }
    
    // Handle login logic
    return form.Success(ctx)
}
```

## Features

### Live Reload

Enable live reload during development:

```go
app.Autoreload(true)
```

### Session Management

```go
session := ctx.Session(db, "user")
session.Load(&userData)
session.Save(&userData)
```

### File Handling

```go
// Download file
ctx.DownloadAs(&fileReader, "application/pdf", "document.pdf")
```

### Form Validation

Built-in support for go-playground/validator:

```go
type User struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=18"`
}
```

## Styling

The framework integrates with Tailwind CSS by default. You can add custom styles through the `HTMLHead` field:

```go
app.HTMLHead = append(app.HTMLHead, `<link rel="stylesheet" href="custom.css">`)
```

## Examples

### Run Examples

Single aggregated server with all examples:

```bash
go run examples/main.go
```

### Watch Mode (auto-reload on file changes)

Requires `reflex` (file watcher):

```bash
# Install once (choose one):
go install github.com/cespare/reflex@latest
# or on macOS with Homebrew
brew install reflex
```

Start the examples server in watch mode:

```bash
./watch.sh
```

This runs:

```bash
reflex -s -r '\\ .go$' -- sh -c 'go run ./examples'
```
so any `.go` change restarts the server automatically.

Then open http://localhost:1422. Use the top navigation menu to switch between pages.

Available routes:
- `/` (home)
- `/hello`
- `/button`
- `/counter`
- `/login`
- `/showcase`

### Showcase (All Components)

A comprehensive demo is part of the aggregated server at `/showcase`.

This example includes:
- Form inputs (`IText`, `IEmail`, `IPhone`, `IPassword`, `INumber`, `IArea`, `IDate`, `ITime`, `IDateTime`)
- Choices (`IRadioButtons`, `ISelect`, `ICheckbox`)
- Buttons with colors (`ui.Blue`, `ui.Green`, `ui.Red`, `ui.Purple`, `ui.Yellow`, `ui.Gray`)
- Actions with server callbacks (counter component)
- `SimpleTable` rendering
- `Markdown` rendering
- `Captcha2` (client-side demo; add server-side validation in real usage)

### Counter Component

A simple counter with increment and decrement buttons:

```go
package main

import (
    "fmt"
    "github.com/michalCapo/go-srui/ui"
)

// Create counter with initial count value
func Counter(count int) *TCounter {
    return &TCounter{Count: count}
}

// Struct definition
type TCounter struct {
    Count int
}

func (counter *TCounter) Increment(ctx *ui.Context) string {
    // Scan request body and update counter struct
    ctx.Body(counter)
    
    // Increase count by 1
    counter.Count++
    
    // Render component
    return counter.Render(ctx)
}

func (counter *TCounter) Decrement(ctx *ui.Context) string {
    // Scan body
    ctx.Body(counter)
    
    counter.Count--
    if counter.Count < 0 {
        counter.Count = 0
    }
    
    // Render component as result of this action
    return counter.Render(ctx)
}

func (counter *TCounter) Render(ctx *ui.Context) string {
    // Temporary id
    target := ui.Target()
    
    // Render HTML, see target (placeholder) at the end, this is place where action result will be rendered
    return ui.Div("flex gap-2 items-center bg-purple-500 rounded text-white p-px", target)(
        ui.Button().
            // Click will call decrement action with counter variable as values sent to this action and result will be rendered at target place
            Click(ctx.Call(counter.Decrement, counter).Replace(target)).
            Class("rounded-l px-5").
            Render("-"),
        
        // Display current count
        ui.Div("text-2xl")(fmt.Sprintf("%d", counter.Count)),
        
        ui.Button().
            // With action result you can replace (overwrite the target component) or render (inline into target component)
            Click(ctx.Call(counter.Increment, counter).Replace(target)).
            Class("rounded-r px-5").
            Render("+"),
    )
}
```

### Login Form with Validation

A complete login form with validation and error handling:

```go
package main

import (
    "github.com/go-playground/validator/v10"
    "github.com/michalCapo/go-srui/ui"
)

func LoginForm(name string) *TLoginForm {
    return &TLoginForm{Name: name}
}

// Defining login form with validations for given fields
type TLoginForm struct {
    Name     string `validate:"required,oneof=user"`
    Password string `validate:"required,oneof=password"`
}

// We want to display success message
func (form *TLoginForm) Success(ctx *ui.Context) string {
    return ui.Div("text-green-600 max-w-md p-8 text-center font-bold rounded-lg bg-white shadow-xl")("Success")
}

// Login action
func (form *TLoginForm) Login(ctx *ui.Context) string {
    // Scan request body, if there is an error render with using render method of this component
    if err := ctx.Body(form); err != nil {
        return form.Render(ctx, &err)
    }

    v := validator.New()
    // Let's validate our input, and display error if any
    if err := v.Struct(form); err != nil {
        return form.Render(ctx, &err)
    }

    // Great a successful login
    return form.Success(ctx)
}

// Translations for login form
var translations = map[string]string{
    "Name":              "User name",
    "has invalid value": "is invalid",
}

// Temporary id
var target = ui.Target()

func (form *TLoginForm) Render(ctx *ui.Context, err *error) string {
    // Submitting form will call login action and result will be rendered to target id
    return ui.Form("flex flex-col gap-4 max-w-md bg-white p-8 rounded-lg shadow-xl", target, ctx.Submit(form.Login).Replace(target))(
        // Display all error in one place
        ui.ErrorForm(err, &translations),

        // Text component
        ui.IText("Name", form).
            // Is required
            Required().
            // If there is specific error for this field display it
            Error(err).
            Render("Name"),

        // Password component
        ui.IPassword("Password").
            Required().
            Error(err).
            Render("Password"),

        // Submit button, see submit part on form several lines above
        ui.Button().
            Submit().
            Color(ui.Blue).
            Class("rounded").
            Render("Login"),
    )
}
```

### Toggle Button

A button that toggles between two states:

```go
package main

import (
    "github.com/michalCapo/go-srui/ui"
)

func main() {
    app := ui.MakeApp("en")
    app.Autoreload(true)

    buttonId := ui.Target()
    var show ui.Callable

    // Primary button that toggles to the "hide" view
    button := func(ctx *ui.Context) string {
        return ui.Button(buttonId).
            Click(ctx.Call(show).Replace(buttonId)).
            Class("rounded").
            Color(ui.Blue).
            Render("Click me")
    }

    page := func(ctx *ui.Context) string {
        // Define the alternate state shown after clicking
        show = func(ctx *ui.Context) string {
            return ui.Div("flex gap-2 items-center bg-red-500 rounded text-white p-px pl-4", buttonId)(
                "Clicked",
                ui.Button().
                    Click(ctx.Call(button).Replace(buttonId)).
                    Class("rounded").
                    Color(ui.Red).
                    Render("Hide me"),
            )
        }

        return app.HTML("Test", "p-8 bg-gray-200",
            ui.Div("flex flex-row gap-4")(
                ui.Div("flex justify-start gap-4 items-center")(
                    button(ctx),
                ),
            ),
        )
    }

    app.Page("/", page)
    app.Listen(":1422")
}
```

## API Reference

### Context Methods

- `ctx.Body(output any)` - Parse request body into struct
- `ctx.Call(method Callable, values ...any)` - Call action with values, returns Actions struct
- `ctx.Submit(method Callable, values ...any)` - Submit form with action, returns Submits struct
- `ctx.Click(method Callable, values ...any)` - Click button with action, returns Submits struct
- `ctx.Send(method Callable, values ...any)` - Send form with action, returns Actions struct
- `ctx.Success(message string)` - Display success message
- `ctx.Error(message string)` - Display error message
- `ctx.Redirect(href string)` - Redirect to URL
- `ctx.Reload()` - Reload current page
- `ctx.DownloadAs(file *io.Reader, content_type string, name string)` - Download file
- `ctx.Callable(action Callable)` - Create callable reference for an action
- `ctx.Action(uid string, action Callable)` - Register action with custom UID

### Action Methods

- `.Render(target Attr)` - Render result inside target
- `.Replace(target Attr)` - Replace target with result
- `.None()` - Do not render result

### Input Components

- `ui.IText(name string, data ...any)` - Text input
- `ui.IPassword(name string, data ...any)` - Password input
- `ui.INumber(name string, data ...any)` - Number input
- `ui.IDate(name string, data ...any)` - Date input
- `ui.ITime(name string, data ...any)` - Time input
- `ui.IDateTime(name string, data ...any)` - DateTime input
- `ui.IArea(name string, data ...any)` - Textarea input

### Button Component

- `ui.Button(attr ...Attr)` - Create button
- `.Click(action string)` - Set click action
- `.Submit()` - Make submit button
- `.Color(color string)` - Set button color
- `.Class(class ...string)` - Set CSS classes
- `.Disabled(value bool)` - Disable button
- `.Render(text string)` - Render button

### Utility Functions

- `ui.Target()` - Generate unique target ID
- `ui.ErrorForm(err *error, translations *map[string]string)` - Display form errors
- `ui.Trim(s string)` - Trim whitespace
- `ui.Normalize(s string)` - Normalize string for HTML
- `ui.Classes(classes ...string)` - Join CSS classes
- `ui.If(cond bool, value func() string)` - Conditional rendering
- `ui.Map(values []T, iter func(*T, int) string)` - Map over slice
- `ui.For(from, to int, iter func(int) string)` - For loop rendering

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 
