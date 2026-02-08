# g-sui Best Practices

## Testing

### Handler Testing

```go
func TestHomePage(t *testing.T) {
    app := ui.MakeApp("en")
    app.Page("/", "Home", func(ctx *ui.Context) string {
        return app.HTML("Test", "bg-white",
            ui.Div("p-4")("Hello"),
        )
    })

    handler := app.TestHandler()
    server := httptest.NewServer(handler)
    defer server.Close()

    resp, err := http.Get(server.URL + "/")
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)

    body, _ := io.ReadAll(resp.Body)
    assert.Contains(t, string(body), "Hello")
}
```

### Component Testing

```go
func TestButton(t *testing.T) {
    btn := ui.Button().Color(ui.Blue).Render("Click me")
    assert.Contains(t, btn, "Click me")
    assert.Contains(t, btn, "bg-blue-800")
}
```

### Form Testing

```go
func TestFormSubmission(t *testing.T) {
    app := ui.MakeApp("en")
    app.Page("/form", "Form", formHandler)

    handler := app.TestHandler()
    server := httptest.NewServer(handler)
    defer server.Close()

    // Submit form
    form := url.Values{}
    form.Set("Email", "test@example.com")
    form.Set("Password", "password123")

    resp, err := http.PostForm(server.URL+"/form", form)
    assert.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
}
```

## Validation

### go-playground/validator

```go
import "github.com/go-playground/validator/v10"

type UserForm struct {
    Name     string `validate:"required,min=3,max=50"`
    Email    string `validate:"required,email"`
    Age      int    `validate:"required,gte=0,lte=120"`
    Password string `validate:"required,min=8"`
    Website  string `validate:"url"`
}

func (f *UserForm) Submit(ctx *ui.Context) string {
    ctx.Body(f)

    v := validator.New()
    if err := v.Struct(f); err != nil {
        return f.Render(ctx, &err)
    }

    ctx.Success("Form valid!")
    return f.Render(ctx, nil)
}
```

### Common Validation Tags

| Tag | Description |
|-----|-------------|
| `required` | Field must be non-empty |
| `email` | Valid email format |
| `min=X` | Minimum length (strings) or value (numbers) |
| `max=X` | Maximum length or value |
| `gte=X` | Greater than or equal |
| `lte=X` | Less than or equal |
| `url` | Valid URL |
| `numeric` | String must be numeric |

### Error Display

```go
func (f *Form) Render(ctx *ui.Context, err *error) string {
    target := ui.Target()

    return ui.Form("p-4", target, ctx.Submit(f.Submit).Replace(target))(
        ui.ErrorForm(err, nil),  // Show validation errors at top
        ui.IText("Name", f).Required().Error(err).Render("Name"),
        ui.IEmail("Email", f).Required().Error(err).Render("Email"),
        ui.Button().Submit().Color(ui.Blue).Render("Submit"),
    )
}
```

### Custom Error Messages

```go
translations := map[string]string{
    "Name": "User name",
    "Email": "Email address",
    "has invalid value": "is not valid",
}

ui.ErrorForm(err, &translations)
```

## Security

### XSS Protection

g-sui automatically escapes HTML attributes:

```go
// All attributes are escaped via escapeAttr()
ui.Div().Class(userInput)  // Safe

// For JavaScript, use escapeJS()
ctx.Script("var x = '%s'", escapeJS(userInput))
```

### CSP Headers

```go
// Default CSP
ctx.SetDefaultCSP()

// Custom CSP
ctx.SetCSP("default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")
```

### Input Validation Limits

Built-in limits prevent excessive input:

```go
const (
    MaxBodySize      = 10 * 1024 * 1024  // 10MB
    MaxFieldNameLen  = 256
    MaxFieldValueLen = 1024 * 1024       // 1MB
    MaxFieldCount    = 1000
)
```

### Safe Field Names

Only safe characters allowed in field names:

```go
// Allowed: a-z, A-Z, 0-9, ., [, ], _
// Blocks: SQL injection attempts with unsafe characters
```

## State Management Patterns

### Page-Level State

```go
type PageState struct {
    Filter string
    Sort   string
}

func (s *PageState) Render(ctx *ui.Context) string {
    ctx.Body(s)  // Restore state
    // ... render UI
}
```

### Component State

```go
type Counter struct {
    Count int
}

func (c *Counter) Increment(ctx *ui.Context) string {
    ctx.Body(c)
    c.Count++
    return c.render(ctx)
}
```

### Session State (requires GORM)

```go
func handler(ctx *ui.Context, db *gorm.DB) string {
    session := ctx.Session(db, "user_prefs")

    var prefs UserPrefs
    session.Load(&prefs)

    prefs.VisitCount++
    session.Save(&prefs)

    return renderUI(prefs)
}
```

## Common Patterns

### Layout with Navigation

```go
func main() {
    app := ui.MakeApp("en")

    app.Layout(func(ctx *ui.Context) string {
        return ui.Div("bg-gray-100 min-h-screen")(
            ui.Div("bg-white shadow p-4 flex gap-3")(
                ui.A("", ctx.Load("/"))("Home"),
                ui.A("", ctx.Load("/users"))("Users"),
            ),
            ui.Div("p-8", ui.Attr{ID: "__content__"})(),
        )
    })

    app.Page("/", "Home", homeHandler)
    app.Page("/users", "Users", usersHandler)
    app.Listen(":8080")
}
```

### Delete Confirmation

```go
func deletePage(ctx *ui.Context) string {
    target := ui.Target()

    confirm := func(ctx *ui.Context) string {
        // Actual delete logic
        return "Item deleted!"
    }

    return ui.Div("p-4")(
        ui.Div("mb-4")("Are you sure you want to delete?"),
        ui.Div("flex gap-2")(
            ui.Button().Color(ui.Gray).
                Click(ctx.Call(deletePage).Replace(target)).
                Render("Cancel"),
            ui.Button().Color(ui.Red).
                Click(ctx.Call(confirm).Replace(target)).
                Render("Delete"),
        ),
    )
}
```

### Loading Skeleton Pattern

```go
func loadData(ctx *ui.Context) string {
    target := ui.Target()

    // Start async fetch
    go func() {
        defer func() { recover() }()

        data := fetchFromAPI()  // Slow operation
        ctx.Replace(target, renderData(data))
    }()

    // Return skeleton immediately
    return target.Skeleton(ui.SkeletonList)
}
```

### Form with Reset

```go
func (f *MyForm) Render(ctx *ui.Context) string {
    target := ui.Target()

    return ui.Form("flex flex-col gap-4", target,
        ctx.Submit(f.Submit).Replace(target))(
        ui.IText("Name", f).Render("Name"),
        ui.Div("flex gap-4 justify-end")(
            ui.Button().Color(ui.Gray).
                Click(ctx.Call(f.Reset).Replace(target)).
                Render("Reset"),
            ui.Button().Submit().Color(ui.Blue).Render("Submit"),
        ),
    )
}

func (f *MyForm) Reset(ctx *ui.Context) string {
    f.Name = ""
    f.Description = ""
    return f.Render(ctx)
}
```

## File Upload

### File Input Component

```go
// Basic file input
form.File("image").
    Accept("image/*").
    Required().
    Render("Image")

// With custom ID for ImagePreview
id := ui.RandomString(10)
form.File("image").
    ID(id).                    // Set custom ID
    Accept("image/*").
    Multiple().                // Allow multiple files
    Required().
    Render("Image")

// Image preview component
ui.ImagePreview(id).
    MaxSize("320px").
    Render()
```

### File Upload Handler

**Single File:**
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

**Multiple Files:**
```go
func uploadMultipleHandler(ctx *ui.Context) string {
    files, err := ctx.Files("images")  // Use .Multiple() on file input
    if err != nil {
        ctx.Error("Failed to process files: " + err.Error())
        return renderUploadForm(ctx)
    }

    for _, file := range files {
        // Validate and save each file
        if strings.HasPrefix(file.ContentType, "image/") && file.Size <= 5*1024*1024 {
            os.WriteFile("uploads/"+file.Name, file.Data, 0644)
        }
    }

    ctx.Success("Files uploaded successfully!")
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
// Single file preview
id := ui.RandomString(10)
form.File("image").ID(id).Accept("image/*").Render("Image")
ui.ImagePreview(id).
    MaxSize("320px").
    Render()

// Multiple file preview (grid layout)
ui.ImagePreview(id).
    Multiple().                // Enable grid layout
    MaxSize("200px").
    Render()
```

The ImagePreview component automatically handles client-side image previews before upload.

### Image Upload Component (Combined)

For image uploads, use the unified `ImageUpload` component that combines file input and preview:

```go
// Single component with inline preview (recommended)
form.ImageUpload("image").
    Zone("Add Photo", "Click to upload").
    MaxSize("320px").
    Required().
    Render("Photo")
```

**Key Features:**
- Inline preview: Selected image appears inside the upload zone
- Change button: Built-in "Change Image" button to re-select
- Unified experience: Single component instead of File + ImagePreview
- Auto-accept: Defaults to `accept="image/*"`

**Alternative:** You can still use separate File + ImagePreview components if needed:

```go
id := ui.RandomString(10)
form.File("image").ID(id).Accept("image/*").Render("Image")
ui.ImagePreview(id).MaxSize("320px").Render()
```

## CAPTCHA

### Captcha2 (Image-based)

```go
func validated(ctx *ui.Context) string {
    return ui.Div("text-green-600")("CAPTCHA validated!")
}

func formWithCaptcha(ctx *ui.Context) string {
    return ui.Div("")(
        ui.Captcha2(validated).Render(ctx),
    )
}
```

### Captcha3 (Draggable tile)

```go
func formWithCaptcha3(ctx *ui.Context) string {
    onSuccess := func(ctx *ui.Context) string {
        ctx.Success("CAPTCHA passed!")
        return showProtectedContent(ctx)
    }

    return ui.Captcha3(onSuccess).
        Count(4).  // Number of tiles
        Render(ctx)
}
```

## CSS Constants

### Sizes

```go
ui.XS  // p-1
ui.SM  // p-2
ui.MD  // p-3
ui.ST  // p-4
ui.LG  // p-5
ui.XL  // p-6
```

### Input Styles

```go
ui.INPUT    // Standard input
ui.AREA     // Textarea
ui.BTN      // Button base
ui.DISABLED // Disabled state
```

### Utility Attributes

```go
ui.W35  // Attr{Style: "max-width: 35rem;"}
ui.W30  // Attr{Style: "max-width: 30rem;"}
ui.W25  // Attr{Style: "max-width: 25rem;"}
ui.W20  // Attr{Style: "max-width: 20rem;"}

ui.Flex1  // Div that grows (flex-grow: 1)
ui.Space  // &nbsp;
```
