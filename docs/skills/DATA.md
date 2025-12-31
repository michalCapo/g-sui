# g-sui Data Collation

Full-featured data management UI with search, sort, filter, pagination, and Excel export backed by GORM.

## Complete Example

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

func PeopleList(ctx *ui.Context, db *gorm.DB) string {
    // Define fields
    name := ui.TField{DB: "name", Field: "Name", Text: "Name"}
    surname := ui.TField{DB: "surname", Field: "Surname", Text: "Surname"}
    email := ui.TField{DB: "email", Field: "Email", Text: "Email"}

    country := ui.TField{
        DB: "country", Field: "Country", Text: "Country",
        As: ui.SELECT, Options: ui.MakeOptions([]string{"USA", "UK", "Germany"}),
    }

    status := ui.TField{
        DB: "status", Field: "Status", Text: "Status",
        As: ui.SELECT, Options: ui.MakeOptions([]string{"new", "active", "blocked"}),
    }

    active := ui.TField{DB: "active", Field: "Active", Text: "Active", As: ui.BOOL}

    createdAt := ui.TField{
        DB: "created_at", Field: "CreatedAt", Text: "Created between",
        As: ui.DATES,
    }

    // Initialize collate
    collate := ui.Collate[Person](&ui.TQuery{
        Limit: 10,
        Order: "surname asc",
    })

    // Configure features
    collate.Search(name, surname, email)       // Searchable
    collate.Sort(surname, name, email)         // Sortable
    collate.Filter(active, createdAt, country, status)  // Filter panel
    collate.Excel(surname, name, email, country, status, active, createdAt)

    // Define row rendering
    collate.Row(func(p *Person, idx int) string {
        return ui.Div("bg-white rounded-lg border p-3 mb-2")(
            ui.Div("flex justify-between")(
                ui.Div("font-semibold")(p.Name + " " + p.Surname),
                ui.Div("text-gray-500")(p.Email),
            ),
        )
    })

    return collate.Render(ctx, db)
}
```

## TField Configuration

```go
ui.TField{
    DB:        "column_name",      // Database column
    Field:     "StructField",      // Go struct field
    Text:      "Display Label",    // UI label
    As:        ui.SELECT,          // Filter type
    Options:   ui.MakeOptions([]string{"A", "B"}),  // For SELECT
    Bool:      false,              // Default for BOOL filters
    Condition: " = 1",             // Custom SQL for BOOL
}
```

## Filter Types

```go
ui.BOOL           // Checkbox (column = 1)
ui.SELECT         // Dropdown (requires Options)
ui.DATES          // Date range (From/To)
ui.ZERO_DATE      // "Has no date" (IS NULL or zero)
ui.NOT_ZERO_DATE  // "Has date" (IS NOT NULL and not zero)
```

## Filter Examples

### Boolean Filter

```go
active := ui.TField{
    DB: "active", Field: "Active", Text: "Active only",
    As: ui.BOOL, Bool: false,
}
```

### Select Filter

```go
country := ui.TField{
    DB: "country", Field: "Country", Text: "Country",
    As: ui.SELECT,
    Options: ui.MakeOptions([]string{"USA", "UK", "Germany"}),
}
```

### Date Range Filter

```go
createdAt := ui.TField{
    DB: "created_at", Field: "CreatedAt", Text: "Created between",
    As: ui.DATES,
}
```

### Date Presence Filters

```go
hasLoggedIn := ui.TField{
    DB: "last_login", Field: "LastLogin", Text: "Has logged in",
    As: ui.NOT_ZERO_DATE,
}

neverLoggedIn := ui.TField{
    DB: "last_login", Field: "LastLogin", Text: "Never logged in",
    As: ui.ZERO_DATE,
}
```

## Excel Export

### Built-in Export

```go
collate.Excel(field1, field2, field3)
```

### Custom Export Handler

```go
collate.OnExcel = func(data *[]Person) (string, io.Reader, error) {
    f := excelize.NewFile()
    // Custom Excel generation
    filename := fmt.Sprintf("export_%s.xlsx", time.Now().Format("20060102"))
    buffer, _ := f.WriteToBuffer()
    return filename, bytes.NewReader(buffer.Bytes()), nil
}
```

## TQuery Configuration

```go
collate := ui.Collate[Person](&ui.TQuery{
    Limit:  20,           // Items per page
    Order:  "name asc",   // Default sort
})
```

## Search Configuration

```go
collate.Search(nameField, emailField, countryField)
```

- Adds search box
- Searches across all specified fields
- Accent-insensitive with `ui.RegisterSQLiteNormalize(db)`

## Sort Configuration

```go
collate.Sort(nameField, ageField, createdAtField)
```

- Adds clickable column headers
- Toggles asc/desc on click

## Filter Configuration

```go
collate.Filter(boolField, selectField, dateField)
```

- Renders filter panel
- Different input types based on `As` value

## Row Rendering

```go
collate.Row(func(item *Person, index int) string {
    // Return HTML for each row
    return ui.Div("...")(...)
})
```

## Custom Row Actions

```go
collate.Row(func(p *Person, idx int) string {
    return ui.Div("flex justify-between")(
        ui.Div()(p.Name),
        ui.Div("flex gap-2")(
            ui.Button().Color(ui.Blue).Class("text-sm").
                Click(ctx.Call(editHandler, p).Replace(target)).
                Render("Edit"),
            ui.Button().Color(ui.Red).Class("text-sm").
                Click(ctx.Call(deleteHandler, p).Replace(target)).
                Render("Delete"),
        ),
    )
})
```

## SQLite Search Normalization

Enable accent-insensitive search:

```go
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/michalCapo/g-sui/ui"
)

db, _ := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
ui.RegisterSQLiteNormalize(db)  // Search "cafe" finds "café"
```

## Accessing Query State

```go
// Inside your handler, TQuery is populated from request:
tq := &ui.TQuery{}
ctx.Body(tq)

// tq.Limit, tq.Offset, tq.Search, tq.Order, tq.Filters available
```
