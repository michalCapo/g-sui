# g-sui JS Mode - Usage Guide

## Overview

g-sui now supports **JavaScript Code Generation Mode** where Go components generate JavaScript code instead of HTML strings. This allows for direct DOM manipulation similar to React, while keeping server-side state management.

## Enabling JS Mode

```go
app := ui.MakeApp("en")
app.UseJS(true)  // Enable JS mode
```

## Architecture

### Current (HTML Mode)
```
Go Handler → HTML String → Client innerHTML → DOM
```

### New (JS Mode)
```
Go Handler → JS Code String → Client execute → Direct DOM Creation
```

## Component API

Each component now has two rendering methods:

- `Render(text string) string` - Generates HTML (existing)
- `JRender(text string) string` - Generates JavaScript code (new)

### Example: Button Component

```go
// HTML Mode
btn := ui.Button().Color(ui.Blue).Render("Click me")
// Output: <div class="..." id="...">Click me</div>

// JS Mode  
btn := ui.Button().Color(ui.Blue).JRender("Click me")
// Output: __e('div',{class:'...',id:'...'},['Click me'])
```

### Example: Input Component

```go
// HTML Mode
input := ui.IText("name").Placeholder("Enter name").Render("Name")
// Output: <div>...</div> with HTML

// JS Mode
input := ui.IText("name").Placeholder("Enter name").JRender("Name")
// Output: __e('div',{},[...]) with JS code
```

## DOM Helper Function

The `__e()` function is automatically included in all pages:

```javascript
function __e(tag, attrs, children) {
    var el = document.createElement(tag);
    // ... set attributes
    // ... append children
    return el;
}
```

## Server Actions

Server actions automatically detect the mode:

```go
func MyHandler(ctx *ui.Context) string {
    if ctx.App.UseJSMode {
        return ui.Button().JRender("JS Button")
    } else {
        return ui.Button().Render("HTML Button")
    }
}
```

## WebSocket Patches

WebSocket patches work in both modes:

**HTML Mode:**
```json
{"type":"patch", "id":"target123", "swap":"inline", "html":"<div>...</div>"}
```

**JS Mode:**
```json
{"type":"patch", "js":"(function(){var t=document.getElementById('target123');...})();"}
```

## Converted Components

The following components have JRender methods:

- ✅ Button (`ui.Button`)
- ✅ Label (`ui.Label`)
- ✅ Input Text (`ui.IText`)
- ✅ Input Area (`ui.IArea`)
- ✅ Input Password (`ui.IPassword`)
- ✅ Select (`ui.ISelect`)

## Adding JRender to Other Components

To add JS mode support to a component:

1. Add `JRender()` method that mirrors `Render()`
2. Use J-prefix constructors (`JDiv`, `JSpan`, etc.)
3. Wrap text with `JText()`
4. Join children with commas

Example pattern:

```go
func (c *MyComponent) Render(text string) string {
    return Div("my-class")(
        Span("")(text),
    )
}

func (c *MyComponent) JRender(text string) string {
    return JDiv("my-class")(
        JSpan("")(JText(text)),
    )
}
```

## Benefits

1. **Direct DOM manipulation** - No HTML parsing overhead
2. **Type-safe** - JS code is generated programmatically
3. **Server-side state** - Business logic stays on server
4. **Backward compatible** - HTML mode still works
5. **Gradual migration** - Convert components incrementally

## Performance Comparison

| Aspect | HTML Mode | JS Mode |
|--------|-----------|---------|
| Output | HTML string | JS code string |
| Client Processing | Parse HTML → DOM | Execute JS → Direct DOM |
| Script Handling | Extract & re-insert | N/A |
| Performance | Parse overhead | Direct creation |

## Migration Path

1. **Enable JS mode**: `app.UseJS(true)`
2. **Test existing pages** - They should still work
3. **Convert components** - Add JRender methods gradually
4. **Update handlers** - Use JRender in new code
5. **Remove HTML mode** - Once all components converted (optional)

## Example: Complete Page

```go
func HomePage(ctx *ui.Context) string {
    if ctx.App.UseJSMode {
        return ctx.App.HTML("Home", "bg-gray-100",
            ui.JDiv("container mx-auto p-4")(
                ui.JDiv("text-3xl font-bold")(ui.JText("Welcome")),
                ui.Button().Color(ui.Blue).JRender("Get Started"),
            ),
        )
    } else {
        return ctx.App.HTML("Home", "bg-gray-100",
            ui.Div("container mx-auto p-4")(
                ui.Div("text-3xl font-bold")("Welcome"),
                ui.Button().Color(ui.Blue).Render("Get Started"),
            ),
        )
    }
}
```

## Notes

- Both modes can coexist in the same application
- The client automatically detects which mode is used
- WebSocket patches work seamlessly in both modes
- No changes required to existing HTML-based code
