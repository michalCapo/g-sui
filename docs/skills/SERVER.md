# g-sui Server Setup

## App Initialization

```go
package main

import "github.com/michalCapo/g-sui/ui"

func main() {
    app := ui.MakeApp("en")  // Locale for translations

    // Register pages
    app.Page("/", homeHandler)
    app.Page("/about", aboutHandler)

    // Serve static assets
    app.Assets(embedFS, "assets/", 24*time.Hour)
    app.Favicon(embedFS, "assets/favicon.svg", 24*time.Hour)

    // Development options
    app.AutoRestart(true)         // Rebuild on file changes
    app.SmoothNavigation(true)    // SPA-like navigation

    // Start server
    app.Listen(":8080")  // Also starts WebSocket at /__ws
}
```

## Route Registration

```go
app.Page("/path", handler)           // GET route
app.Page("/path", handler, "POST")   // POST route
```

## HTML Wrapper

```go
// Full HTML document with Tailwind CSS
app.HTML(title, bodyClass, content) string

// Example
func homeHandler(ctx *ui.Context) string {
    return app.HTML("Home", "bg-gray-100",
        ui.Div("p-8")(
            ui.Div("text-2xl font-bold")("Hello World"),
        ),
    )
}
```

## Custom Head Content

```go
app.HTMLHead = []string{
    `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">`,
    `<meta name="description" content="My app">`,
}
```

## PWA Configuration

```go
app.PWA(ui.PWAConfig{
    Name:        "My App",
    ShortName:   "App",
    Description: "My Progressive Web App",
    Background:  "#ffffff",
    Theme:       "#000000",
    Icon:        "assets/icon-192.png",
    Icon512:     "assets/icon-512.png",
    Display:     "standalone",  // standalone, fullscreen, minimal-ui
})
```

## Testing Handler

```go
handler := app.TestHandler()              // Get http.Handler
server := httptest.NewServer(handler)     // Create test server
resp, _ := http.Get(server.URL + "/path") // Make requests
```

## WebSocket

WebSocket endpoint is automatically created at `/__ws`.

### Manual WebSocket Connection (for external clients)

```javascript
const ws = new WebSocket('ws://localhost:8080/__ws');

ws.onmessage = (event) => {
    const patch = JSON.parse(event.data);
    // patch.Target: element ID
    // patch.HTML: content to insert
    // patch.Swap: "inline", "outline", "append", "prepend"
};

// Send patches from Go:
ctx.Patch(target.Replace(), html)
```

## Smooth Navigation

SPA-like navigation without full page reload:

```go
app.SmoothNavigation(true)

// In your code:
ctx.Load("/path")  // Returns Attr for smooth navigation
```

### Navigation Example

```go
ui.Button().
    Color(ui.Blue).
    Click(ctx.Call(navigateHandler).Load("/about")).
    Render("Go to About")

// Or directly:
ui.A("", ui.Href("/page"), ctx.Load("/page"))("Link")
```

## Auto Restart (Development)

```go
app.AutoRestart(true)  // Rebuild on file changes
```

When enabled, the app watches for file changes and rebuilds automatically.

## Assets

```go
// Serve directory
app.Assets(embedFS, "assets/", 24*time.Hour)

// Serve single file (favicon)
app.Favicon(embedFS, "assets/favicon.svg", 24*time.Hour)
```

## Context Properties

```go
type Context struct {
    App       *ui.App
    Request   *http.Request
    Response  http.ResponseWriter
    SessionID string
}
```

## Complete Example

```go
//go:embed assets/*
var assets embed.FS

func main() {
    app := ui.MakeApp("en")

    // Custom head
    app.HTMLHead = []string{
        `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">`,
    }

    // Routes
    app.Page("/", homePage)
    app.Page("/users", usersPage)
    app.Page("/users/:id", userDetailPage)

    // Assets
    app.Assets(assets, "assets", 24*time.Hour)

    // PWA
    app.PWA(ui.PWAConfig{
        Name:      "My App",
        ShortName: "App",
        Icon:      "assets/icon-192.png",
        Icon512:   "assets/icon-512.png",
    })

    // Dev mode
    app.AutoRestart(true)
    app.SmoothNavigation(true)

    app.Listen(":8080")
}

func homePage(ctx *ui.Context) string {
    return ctx.App.HTML("Home", "bg-gray-100",
        ui.Div("p-8")(
            ui.Div("text-2xl font-bold")("Welcome"),
            ui.Button().Color(ui.Blue).
                Click(ctx.Call(func(c *ui.Context) string {
                    return c.Load("/users")
                }).None()).
                Render("View Users"),
        ),
    )
}
```

## File Structure Recommendation

```
project/
├── main.go
├── assets/
│   ├── favicon.svg
│   ├── icon-192.png
│   └── icon-512.png
├── embed.go       // //go:embed assets/*
├── handlers/
│   ├── home.go
│   └── users.go
└── models/
    └── user.go
```
