# g-sui UI Components

## Buttons

```go
ui.Button().
    Color(ui.Blue).           // Blue, Green, Red, Yellow, Purple, Gray, White + *Outline
    Size(ui.MD).              // XS, SM, MD, ST, LG, XL
    Class("rounded px-4").    // Custom classes
    Click(ctx.Call(...)).     // Click handler
    Href("/path").            // Make link
    Submit().                 // type="submit"
    Reset().                  // type="reset"
    Disabled(true).           // Disable
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
ui.Alert("info").Message("Info message").Render()
ui.Alert("success").Message("Success!").Render()
ui.Alert("warning").Message("Warning!").Render()
ui.Alert("error").Message("Error!").Render()
```

### Badge

```go
ui.Badge("blue").Text("New").Render()
ui.Badge("green").Text("Active").Render()
ui.Badge("red").Text("Deleted").Render()
```

### Card

```go
ui.Card().
    Title("Card Title").
    Subtitle("Subtitle").
    Action(ui.Button().Color(ui.Blue).Render("Action")).
    Body(ui.P("")("Card content")).
    Render()
```

### Tabs

```go
tabs := ui.Tabs("tabs-id")

tabs.Tab(ui.TabItem{
    ID:      "tab1",
    Label:   "Tab 1",
    Content: ui.Div("")("Content 1"),
    Active:  true,
})

tabs.Tab(ui.TabItem{
    ID:    "tab2",
    Label: "Tab 2",
    Content: ui.Div("")("Content 2"),
})

tabs.Render()
```

### Accordion

```go
acc := ui.Accordion("acc-id")

acc.Item(ui.AccordionItem{
    ID:      "item1",
    Title:   "Section 1",
    Content: ui.Div("")("Content 1"),
    Open:    true,
})

acc.Item(ui.AccordionItem{
    ID:      "item2",
    Title:   "Section 2",
    Content: ui.Div("")("Content 2"),
})

acc.Render()
```

### Dropdown

```go
dropdown := ui.Dropdown("dropdown-id")

dropdown.Item(ui.DropdownItem{
    Label: "Option 1",
    Click: ctx.Call(handler1).Replace(target),
})

dropdown.Item(ui.DropdownItem{
    Label: "Option 2",
    Click: ctx.Call(handler2).Replace(target),
})

dropdown.Trigger(ui.Button().Color(ui.Blue).Render("Menu"))
dropdown.Render()
```

### Progress Bar

```go
ui.Progress().Value(50).Max(100).Color(ui.Blue).Render()
```

### Step/Wizard

```go
steps := ui.Step("steps-id")

steps.StepItem(ui.StepItem{
    Number: 1,
    Title:  "Step 1",
    Status: "completed",  // completed, active, pending
})

steps.StepItem(ui.StepItem{
    Number: 2,
    Title:  "Step 2",
    Status: "active",
})

steps.StepItem(ui.StepItem{
    Number: 3,
    Title:  "Step 3",
    Status: "pending",
})

steps.Render()
```

### Tooltip

```go
ui.Button().
    Tooltip(ui.TooltipInfo{Text: "Help text"}).
    Render("Button with tooltip")
```

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
