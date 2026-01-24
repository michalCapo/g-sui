# g-sui UI Components

## Buttons

```go
ui.Button().
    Color(ui.Blue).           // Blue, Green, Red, Yellow, Purple, Gray, White + *Outline
    Size(ui.MD).              // XS, SM, MD, ST, LG, XL
    Class("rounded px-4").    // Custom classes
    Click(ctx.Call(...)).     // Click handler (returns JS string)
    Href("/path").            // Make link (<a> element)
    Submit().                 // type="submit"
    Reset().                  // type="reset"
    Disabled(true).           // Disable button
    Form("form-id").          // Associate with form by ID
    Name("btn-name").         // Button name attribute
    Val("value").             // Button value attribute
    If(condition).            // Conditional render
    Render("Button Text")
```

### Colors

```go
ui.Blue, ui.BlueOutline
ui.Green, ui.GreenOutline
ui.Red, ui.RedOutline
ui.Yellow, ui.YellowOutline
ui.Purple, ui.PurpleOutline
ui.Gray, ui.GrayOutline
ui.White, ui.WhiteOutline
```

## Inputs

All inputs use fluent API: `ui.IType("Field", &data).Method().Render("Label")`

### Text Inputs

```go
ui.IText("Name", &data).Required().Placeholder("hint").Render("Name")
ui.IEmail("Email", &data).Required().Render("Email")
ui.IPhone("Phone", &data).Render("Phone")       // With pattern
ui.IPassword("Password").Required().Render("Password")
ui.IArea("Bio", &data).Rows(5).Render("Bio")
```

### Numbers & Dates

```go
ui.INumber("Age", &data).Numbers(0, 120, 1).Render("Age")
ui.INumber("Price", &data).Format("%.2f").Render("Price")
ui.IDate("BirthDate", &data).Dates(min, max).Render("Birth Date")
ui.ITime("Alarm", &data).Render("Alarm Time")
ui.IDateTime("Meeting", &data).Render("Meeting")
```

### Selection

```go
// Dropdown
options := ui.MakeOptions([]string{"A", "B", "C"})
ui.ISelect("Country", &data).Options(options).Render("Country")

// Checkbox
ui.ICheckbox("Agree", &data).Required().Render("I agree")

// Radio buttons
ui.IRadio("Gender", &data).Value("male").Render("Male")
ui.IRadio("Gender", &data).Value("female").Render("Female")

// Radio group
ui.IRadioButtons("Plan", &data).Options(planOptions).Render("Plan")

// Card-based radio (custom HTML)
cardOptions := []ui.AOption{
    {ID: "1", Value: ui.Div("p-4 border")("Card 1")},
    {ID: "2", Value: ui.Div("p-4 border")("Card 2")},
}
ui.IRadioDiv("Plan", &data).Options(cardOptions).Render("Plan")
```

### File Input

```go
// Basic file input
ui.IFile("image").Accept("image/*").Required().Render("Image")

// With custom ID for ImagePreview
id := ui.RandomString(10)
ui.IFile("image").
    ID(id).                    // Set custom ID
    Accept("image/*").          // MIME types: "image/*", ".pdf,.doc"
    Multiple().                // Allow multiple files
    Required().
    Disabled(false).
    Class("custom-wrapper").   // Wrapper classes
    ClassInput("custom-input"). // Input classes
    ClassLabel("custom-label"). // Label classes
    Change("console.log('changed')").  // OnChange handler
    If(condition).             // Conditional render
    Render("Upload Image")

// Get the file input ID
fileInputID := fileInput.GetID()
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

### Image Preview

```go
// Single file preview
id := ui.RandomString(10)
ui.IFile("image").ID(id).Accept("image/*").Render("Image")
ui.ImagePreview(id).
    MaxSize("320px").          // Max width/height
    Render()

// Multiple file preview (grid layout)
ui.ImagePreview(id).
    Multiple().                // Enable grid layout
    MaxSize("200px").
    Class("my-4").             // Custom wrapper classes
    Render()

// Usage pattern
id := ui.RandomString(10)
fileInput := form.File("image").ID(id).Accept("image/*").Render("Image")
ui.ImagePreview(id).MaxSize("320px").Render()
```

**ImagePreview Methods:**
- `.Multiple()` - Enable grid layout for multiple images (default: single centered)
- `.MaxSize(size)` - Maximum image dimensions (e.g., `"320px"`)
- `.Class(classes...)` - Custom wrapper classes
- `.If(condition)` - Conditional render
- `.Render()` - Generate HTML and JavaScript

### Image Upload Component (Combined File + Preview)

The `ImageUpload` component combines file input and image preview into a single unified component with inline preview:

```go
// Basic image upload with inline preview
form.ImageUpload("image").
    Zone("Add Image", "Click to upload").
    MaxSize("320px").
    Required().
    Render("Image")

// With custom zone styling
form.ImageUpload("image").
    Zone("Add Vehicle Photo", "Click to take or upload").
    ZoneIcon("w-10 h-10 bg-gray-500 rounded-full p-2 flex items-center justify-center").
    MaxSize("320px").
    ClassPreview("mt-4").
    Required().
    Render("VEHICLE PHOTO")
```

**Key Features:**
- **Inline Preview**: Selected image appears inside the upload zone (replacing the upload UI)
- **Change Button**: Built-in "Change Image" button to re-select images
- **Unified Experience**: Single component instead of separate File + ImagePreview
- **Zone Mode**: Uses dropzone-style UI by default for better UX
- **Auto-accept**: Defaults to `accept="image/*"` for images

**ImageUpload Methods:**
- `.Zone(title, hint)` - Enable dropzone mode with title and hint text
- `.ZoneIcon(classes)` - Custom icon CSS classes for zone mode
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

### Common Input Methods

```go
.Required()              // Mark required
.Disabled()              // Disable
.Readonly()              // Read-only
.Placeholder("hint")     // Placeholder text
.Class("cls")            // Wrapper classes
.ClassInput("cls")       // Input classes
.ClassLabel("cls")       // Label classes
.Value("default")        // Default value
.Pattern("regex")        // HTML pattern
.Autocomplete("email")   // Autocomplete hint
.Change(action)          // OnChange handler
.Click(action)           // OnClick handler
.Error(&err)             // Show validation error
.If(condition)           // Conditional render
.Render("Label")         // Render with label
```

## Forms

### Basic Form

```go
type LoginForm struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

func (f *LoginForm) Submit(ctx *ui.Context) string {
    if err := ctx.Body(f); err != nil {
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

    return ui.Form("bg-white p-6 rounded", target,
        ctx.Submit(f.Submit).Replace(target))(
        ui.ErrorForm(err, nil),
        ui.IEmail("Email", f).Required().Error(err).Render("Email"),
        ui.IPassword("Password").Required().Error(err).Render("Password"),
        ui.Button().Submit().Color(ui.Blue).Render("Login"),
    )
}
```

### FormInstance (Disconnected Forms)

Place inputs outside the form element:

```go
form := ui.FormNew(ctx.Submit(handler).Replace(target))

return ui.Div("max-w-5xl")(
    form.Render(),                      // Hidden form element
    form.Text("Title").Required().Render("Title"),
    form.Email("Email").Required().Render("Email"),
    form.Button().Color(ui.Blue).Submit().Render("Submit"),
)
```

### Validation Translations

```go
translations := map[string]string{
    "Name": "User name",
    "Email": "Email address",
}

ui.ErrorForm(err, &translations)
```

## Tables

### Simple Table

```go
table := ui.SimpleTable(3, "w-full bg-white")  // 3 columns
table.Field("Name", "font-bold")
table.Field("Age", "text-center")
table.Field("Email", "")
// New row starts after 3 fields
table.Render()
```

### Typed Table with Headers

```go
type Person struct { Name string; Age int; Email string }

table := ui.Table[Person]("w-full bg-white")
table.Head("Name", "font-bold")
table.Head("Age", "text-center")
table.Head("Email", "")

table.FieldText(func(p *Person) string { return p.Name }, "font-bold")
table.FieldText(func(p *Person) string {
    return fmt.Sprintf("%d", p.Age)
}, "text-center")
table.FieldText(func(p *Person) string { return p.Email }, "")

table.Render(persons)
```

### Colspan

```go
table := ui.SimpleTable(4, "w-full")
table.Field("Spans 2 columns").Attr(`colspan="2"`)
table.Field("Col 3")
table.Field("Col 4")
table.Render()
```

## Other Components

### Alert

```go
ui.Alert().
    Variant("info").              // "info", "success", "warning", "error", or *-outline variants
    Title("Heads up!").           // Optional title
    Message("Info message").      // Alert message
    Dismissible(true).            // Show dismiss button
    Persist("alert-key").         // localStorage key for "don't show again"
    Class("custom-class").        // Additional classes
    If(condition).                // Conditional render
    Render()
```

**Variants:** `"info"`, `"success"`, `"warning"`, `"error"`, `"info-outline"`, `"success-outline"`, `"warning-outline"`, `"error-outline"`

### Badge

```go
ui.Badge().
    Text("New").                  // Badge text
    Color("blue").                // "blue", "green", "red", "yellow", "purple", "gray" (+ "-soft", "-outline")
    Size("md").                   // "sm", "md", "lg"
    Dot().                        // Show as dot indicator (no text)
    Icon(iconHTML).               // Optional icon HTML
    Square().                     // Square corners (default: rounded)
    Class("custom").              // Additional classes
    If(condition).                // Conditional render
    Render()
```

**Color variants:** `"blue"`, `"green"`, `"red"`, `"yellow"`, `"purple"`, `"gray"`, `"blue-soft"`, `"green-soft"`, etc.

### Card

```go
ui.Card().
    Variant(ui.CardShadowed).    // CardShadowed, CardBordered, CardFlat, CardGlass
    Header(html).                 // Header HTML
    Body(html).                   // Body HTML
    Footer(html).                 // Footer HTML
    Image(src, alt).              // Optional image at top
    Hover(true).                  // Enable hover effect
    Compact(true).                // Compact padding
    Padding("p-4").               // Custom padding
    Class("custom").              // Additional classes
    If(condition).                // Conditional render
    Render()
```

**Variants:** `ui.CardShadowed` (default), `ui.CardBordered`, `ui.CardFlat`, `ui.CardGlass`

### Tabs

```go
tabs := ui.Tabs()

tabs.Tab("Tab 1", contentHTML, iconHTML).  // Label, content, optional icon
tabs.Tab("Tab 2", contentHTML).
tabs.Active(0).                            // Initially active tab (0-based)
tabs.Style("underline").                   // "underline", "pills", "boxed", "vertical"
tabs.Class("custom").                      // Additional classes
tabs.If(condition).                        // Conditional render
tabs.Render()
```

**Styles:** `"underline"` (default), `"pills"`, `"boxed"`, `"vertical"`

### Accordion

```go
acc := ui.Accordion()

acc.Item("Section 1", contentHTML, true).  // Title, content, optional open state
acc.Item("Section 2", contentHTML).
acc.Multiple(true).                       // Allow multiple sections open
acc.Variant(ui.AccordionBordered).        // AccordionBordered, AccordionGhost, AccordionSeparated
acc.Class("custom").                      // Additional classes
acc.If(condition).                        // Conditional render
acc.Render()
```

**Variants:** `ui.AccordionBordered` (default), `ui.AccordionGhost`, `ui.AccordionSeparated`

### Dropdown

```go
dropdown := ui.Dropdown()

dropdown.Trigger(html).                   // Trigger element HTML
dropdown.Item("Option 1", onclickJS, iconHTML).  // Label, onclick JS, optional icon
dropdown.Item("Option 2", onclickJS).
dropdown.Header("Group Name").            // Non-interactive header
dropdown.Divider().                       // Visual separator
dropdown.Danger("Delete", onclickJS, iconHTML).  // Danger variant item
dropdown.Position("bottom-left").         // "bottom-left", "bottom-right", "top-left", "top-right"
dropdown.Class("custom").                 // Additional classes
dropdown.If(condition).                   // Conditional render
dropdown.Render()
```

**Positions:** `"bottom-left"` (default), `"bottom-right"`, `"top-left"`, `"top-right"`

### Progress Bar

```go
ui.ProgressBar().
    Value(75).                            // Percentage (0-100)
    Color("bg-blue-600").                // Color class
    Gradient("#3b82f6", "#8b5cf6").      // Gradient colors (overrides Color)
    Striped(true).                       // Striped pattern
    Animated(true).                      // Animate stripes
    Indeterminate(true).                 // Indeterminate progress
    Size("md").                          // "xs", "sm", "md", "lg", "xl"
    Label("Loading...").                 // Optional label text
    LabelPosition("outside").            // "inside" (default) or "outside"
    Class("custom").                     // Additional classes
    If(condition).                       // Conditional render
    Render()
```

### Step Progress

```go
ui.StepProgress(2, 5).                    // Current step, total steps
    Current(3).                          // Update current step
    Total(6).                           // Update total steps
    Color("bg-blue-500").               // Progress bar color
    Size("md").                         // "xs", "sm", "md", "lg", "xl"
    Class("custom").                    // Additional classes
    If(condition).                      // Conditional render
    Render()
```

Shows "Step X of Y" label with progress bar.

### Tooltip

```go
ui.Tooltip().
    Content("Help text").                // Tooltip text
    Position("top").                     // "top", "bottom", "left", "right"
    Variant("dark").                     // "dark", "light", "blue", "green", "red", "yellow"
    Delay(500).                          // Show delay in milliseconds
    Class("custom").                     // Additional classes
    If(condition).                       // Conditional render
    Render(elementHTML)                  // Wrap element with tooltip
```

**Positions:** `"top"` (default), `"bottom"`, `"left"`, `"right"`  
**Variants:** `"dark"` (default), `"light"`, `"blue"`, `"green"`, `"red"`, `"yellow"`

## Labels & Icons

### Labels

```go
target := ui.Target()
ui.Label(&target).Render("Field Label")
ui.Label(&target).Required(true).Render("Required")
ui.Label(&target).Class("text-lg").Render("Styled")
```

### Icons (FontAwesome)

```go
// Include in app.HTMLHead
app.HTMLHead = append(app.HTMLHead,
    `<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">`,
)

ui.Icon("fa fa-check")                       // <i class="fa fa-check"></i>
ui.Icon2("fa fa-check", "text-green-500")    // With classes
ui.IconLeft("fa fa-arrow-left", "Back")      // Icon + text
ui.IconRight("Next", "fa fa-arrow-right")    // Text + icon
```

## Theme Switcher

```go
// Cycles: System → Light → Dark
ui.ThemeSwitcher("")                         // Default
ui.ThemeSwitcher("fixed bottom-4 right-4")   // Positioned
```

## Hidden Fields

```go
ui.Hidden("UserID", "uint", 123)
ui.Hidden("Mode", "string", "edit")
ui.Hidden("Filter[0].Field", "string", "name")
```
