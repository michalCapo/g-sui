# g-sui Server Setup

## App Initialization

```go
package main

import "github.com/michalCapo/g-sui/ui"

func main() {
    app := ui.MakeApp("en")  // Locale for translations

    // Register pages
    app.Page("/", "Home", homeHandler)
    app.Page("/about", "About", aboutHandler)

    // Serve static assets
    app.Assets(embedFS, "assets/", 24*time.Hour)
    app.Favicon(embedFS, "assets/favicon.svg", 24*time.Hour)

    // Register custom HTTP handlers (REST APIs)
    app.Custom("GET", "/api/health", healthHandler)
    app.GET("/api/users", getUsersHandler)
    app.POST("/api/users", createUserHandler)
    app.PUT("/api/users/:id", updateUserHandler)
    app.DELETE("/api/users/:id", deleteUserHandler)

    // Development options
    app.AutoRestart(true)         // Rebuild on file changes

    // Start server
    app.Listen(":8080")  // Also starts WebSocket at /__ws
}
```

## Route Registration

### g-sui Page Routes

```go
app.Page("/path", "Title", handler)  // Register page route with title
```

### Parameterized Routes

Routes can include path parameters using curly braces:

```go
// Single parameter
app.Page("/user/{id}", "User Detail", userDetailHandler)

// Multiple parameters
app.Page("/user/{userId}/post/{postId}", "Post Detail", postDetailHandler)

// Nested parameters
app.Page("/category/{category}/product/{product}", "Product Detail", productHandler)
```

**Accessing Path Parameters:**
```go
func userDetailHandler(ctx *ui.Context) string {
    userID := ctx.PathParam("id")  // Extract from /user/{id}
    // Use userID...
}
```

**Query Parameters:**
Query parameters work with any route (with or without path parameters):

```go
// URL: /search?name=Smith&age=30
func searchHandler(ctx *ui.Context) string {
    name := ctx.QueryParam("name")  // "Smith"
    age := ctx.QueryParam("age")   // "30"
    // Works with both SPA navigation (ctx.Load) and direct requests
}

// URL: /user/123?tab=profile&view=detailed
func userHandler(ctx *ui.Context) string {
    // Get path parameters from route pattern
    userID := ctx.PathParam("id")   // "123" (path param)
    
    // Get query parameters (if any) using ctx.QueryParam() - works with SPA navigation
    tab := ctx.QueryParam("tab")    // "profile" (query param)
    view := ctx.QueryParam("view")  // "detailed" (query param)
    sort := ctx.QueryParam("sort")  // Empty string if not present
    order := ctx.QueryParam("order") // Empty string if not present
}

// Multi-value query parameters: /tags?tag=a&tag=b
func tagsHandler(ctx *ui.Context) string {
    tags := ctx.QueryParams("tag")  // []string{"a", "b"}
    allParams := ctx.AllQueryParams()  // map[string][]string
}
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
    `<link rel="preconnect" href="https://fonts.googleapis.com">`,
    `<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>`,
    `<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">`,
    `<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">`,
    `<meta name="description" content="My app">`,
}
```

## PWA Configuration

```go
app.PWA(ui.PWAConfig{
    Name:                  "My App",
    ShortName:             "App",
    ID:                    "/",                              // App ID (defaults to StartURL)
    Description:           "My Progressive Web App",
    ThemeColor:            "#ffffff",
    BackgroundColor:       "#000000",
    Display:               "standalone",                     // standalone, fullscreen, minimal-ui
    StartURL:              "/",
    GenerateServiceWorker: true,
    CacheAssets:           []string{"/assets/app.css", "/assets/app.js"}, // Assets to pre-cache
    OfflinePage:           "/offline",                       // Fallback when offline
    Icons: []ui.PWAIcon{
        {Src: "/icon-192.png", Sizes: "192x192", Type: "image/png", Purpose: "any"},
        {Src: "/icon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "any maskable"},
    },
})
```

**PWAConfig Fields:**
- `Name` - Full application name
- `ShortName` - Short name for home screen
- `ID` - App identity (defaults to `StartURL` if empty)
- `Description` - App description
- `ThemeColor` - Theme color (hex)
- `BackgroundColor` - Splash screen background (hex)
- `Display` - Display mode: `standalone`, `fullscreen`, `minimal-ui`, `browser`
- `StartURL` - Launch URL (defaults to `/`)
- `GenerateServiceWorker` - Generate service worker (network-first pages, cache-first assets)
- `CacheAssets` - Asset URLs to pre-cache (e.g., `["/assets/app.css"]`)
- `OfflinePage` - Fallback page when offline (e.g., `"/offline"`)
- `Icons` - Array of app icons

**Service Worker Behavior:**
- **Pages**: Network-first (always fresh content, cache only when offline)
- **Assets**: Cache-first (fast loading from cache)
- **Cache versioning**: New cache on each server restart, old caches auto-cleaned

**PWAIcon Fields:**
- `Src` - Icon path
- `Sizes` - Size (e.g., `192x192`, `512x512`, `any`)
- `Type` - MIME type (e.g., `image/png`, `image/x-icon`)
- `Purpose` - Icon purpose: `any`, `maskable`, or `any maskable`

## Custom HTTP Handlers (REST APIs)

Custom handlers are checked **before** g-sui page routes, so they take priority. This allows you to mix server-rendered pages with traditional REST API endpoints.

```go
// Full method signature
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

**Example REST API Handler:**
```go
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Save user...
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}
```

**Handler Priority:**
1. Asset handlers (e.g., `/assets/*`)
2. Custom HTTP handlers (e.g., `/api/*`)
3. g-sui Page routes (e.g., `/`, `/about`)

## Custom Server Configuration

You can retrieve the `http.Handler` and use it with a custom server setup or wrap it with middleware:

```go
app := ui.MakeApp("en")
app.Page("/", "Home", homeHandler)
app.StartSweeper()  // Manually start session sweeper when using Handler()

// Get the handler
handler := app.Handler()

// Wrap with custom middleware
handler = loggingMiddleware(handler)
handler = corsMiddleware(handler)

// Use with custom server
server := &http.Server{
    Addr:         ":8080",
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
}

log.Fatal(server.ListenAndServe())
```

**Note:** When using `app.Handler()`:
- You must manually call `app.StartSweeper()` to enable session cleanup
- WebSocket endpoint at `/__ws` is automatically registered
- Call `app.initWS()` is **not** needed (handled internally)

## Testing Handler

```go
handler := app.TestHandler()              // Get http.Handler for testing (minimal setup)
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
// Use ctx.Load() for smooth navigation
ctx.Load("/path")  // Returns Attr with href and onclick for smooth SPA navigation
```

### Navigation Example

```go
ui.Button().
    Color(ui.Blue).
    Click(ctx.Call(navigateHandler).Load("/about")).
    Render("Go to About")

// Or directly:
ui.A("", ctx.Load("/page"))("Link")
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
        `<link rel="preconnect" href="https://fonts.googleapis.com">`,
        `<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>`,
        `<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">`,
        `<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">`,
    }

    // Routes
    app.Page("/", homePage)
    app.Page("/users", usersPage)
    app.Page("/users/:id", userDetailPage)

    // Assets
    app.Assets(assets, "assets", 24*time.Hour)

    // PWA
    app.PWA(ui.PWAConfig{
        Name:                  "My App",
        ShortName:             "App",
        ID:                    "/",
        Description:           "My Application",
        ThemeColor:            "#1d4ed8",
        BackgroundColor:       "#ffffff",
        GenerateServiceWorker: true,
        Icons: []ui.PWAIcon{
            {Src: "/icon-192.png", Sizes: "192x192", Type: "image/png", Purpose: "any"},
            {Src: "/icon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "any maskable"},
        },
    })

    // Dev mode
    app.AutoRestart(true)

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

## Releases and Versioning

### Creating a New Release

To create and push a new version tag:

```bash
./deploy
```

The `deploy` script:
- Starts at `v0.100` if no tags exist
- Auto-increments minor version (`v0.100` → `v0.101` → `v0.102`, etc.)
- Validates clean working tree before tagging
- Creates annotated git tag and pushes to remote

### Version Numbering

Format: `v0.XXX`
- Major version: Fixed at `0` (pre-1.0 release)
- Minor version: Auto-incremented starting from `100`

After running `./deploy`, create a GitHub release at the repository's releases page.
