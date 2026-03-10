# g-sui Client-Side Rendering

Hybrid client-side rendering for g-sui. The server renders the page shell (layout, headers, empty containers), then a JS runtime fetches data from APIs and renders content via `__engine.create()`. Go builds JSON config, JS owns all visual decisions.

**Core loop:** Go emits `<div>` + `<script>__client(config)</script>` → JS shows skeleton → fetches API → calls registered component → `__engine.create()` → DOM.

**Package separation:** Client-side builder types live in the `js` package (`github.com/michalCapo/g-sui/js`). Server-rendered UI stays in `ui`. This makes it clear which code will be converted to JavaScript on the client.

## Quick Start

```go
import (
    "github.com/michalCapo/g-sui/js"
    "github.com/michalCapo/g-sui/ui"
)

func invoicesPage(ctx *ui.Context) string {
    return app.HTML("Invoices", "bg-gray-100",
        ui.Div("p-6")(
            ui.H1("text-2xl font-bold mb-4")("Invoices"),
            js.Client(ctx).
                Source("/api/invoices").
                Loading(ui.SkeletonTable).
                Empty("inbox", "No invoices found").
                Table(
                    js.Col("Firma").Label("Company").Sortable(true),
                    js.Col("Castka").Label("Amount").Type("number").Format("amount"),
                    js.Col("DatVyst").Label("Date").Type("date").Sortable(true),
                    js.Col("Uhrazeno").Label("Paid").Type("bool"),
                ).
                Search(true).
                Pagination(25).
                Render(),
        ),
    )
}
```

The API endpoint returns JSON:
```go
app.Handle("/api/invoices", func(ctx *ui.Context) string {
    rows := db.GetInvoices()
    ctx.JSON(rows)  // []map[string]any or []struct
    return ""
})
```

## Client Builder API Reference

### `js.Client(ctx *ui.Context) *Builder`

Creates a new client zone. Generates a unique `id` (`cl_<random>`), defaults to `autoLoad: true` and `showError: true`.

### `.Source(url string)`

**Required.** The API endpoint that returns JSON data.

```go
.Source("/api/invoices")
.Source("/api/dashboard/revenue?year=2025")
```

### `.Params(map[string]string)`

Extra query params appended to the source URL on fetch.

```go
.Params(map[string]string{"year": "2025", "type": "issued"})
```

### `.Loading(skeleton ui.Skeleton)`

Skeleton shown while fetching. Options: `ui.SkeletonTable`, `ui.SkeletonCards`, or default component skeleton.

```go
.Loading(ui.SkeletonTable)
.Loading(ui.SkeletonCards)
```

### `.Empty(icon, message string)`

What to show when API returns `null`, `[]`, or `{}`. Icon is a Material Icons name.

```go
.Empty("inbox", "No invoices found")
.Empty("search_off", "No results match your filters")
```

### `.Error(show bool)`

Show/hide error state on fetch failure. Default: `true`.

```go
.Error(false)  // silently fail
```

### `.AutoLoad(auto bool)`

Whether to fetch data on mount. Default: `true`. Set `false` for zones triggered by user action.

```go
.AutoLoad(false)  // call __clients["cl_xxx"].reload() manually
```

### `.Poll(d time.Duration)`

Auto-refresh interval. Re-fetches from source and re-renders.

```go
.Poll(30 * time.Second)
.Poll(5 * time.Minute)
```

### `.Component(name string, opts js.Opts)`

Sets the registered JS component to render data. `opts` is `map[string]any`, serialized to JSON.

```go
.Component("card-grid", js.Opts{
    "cols":      3,
    "primary":   "Firma",
    "secondary": "ICO",
})
```

### `.Table(columns ...*js.Column)`

Sugar for `.Component("table", ...)`. Compiles column definitions to JSON.

```go
.Table(
    js.Col("Firma").Label("Company").Sortable(true),
    js.Col("Castka").Label("Amount").Type("number").Format("amount"),
)
// equivalent to:
.Component("table", js.Opts{
    "columns": []map[string]any{
        {"key": "Firma", "label": "Company", "sortable": true},
        {"key": "Castka", "label": "Amount", "type": "number", "format": "amount"},
    },
})
```

### `.Filter(enabled bool)`

Adds `"filter": true` to table opts. Enables filter bar above table.

### `.Pagination(pageSize int)`

Adds `"pageSize": N` to table opts. Client-side pagination with prev/next and page numbers.

### `.Search(enabled bool)`

Adds `"search": true` to table opts. Renders search input that filters across all columns.

### `.Chart(chartType js.ChartType)`

Sugar for `.Component("chart", ...)`. Sets chart type.

```go
.Chart(js.BarChart)
.Chart(js.AreaChart)
.Chart(js.HBarChart)
.Chart(js.DonutChart)
```

### `.ChartOptions(opts js.Opts)`

Merges additional options into chart config.

```go
.Chart(js.BarChart).
ChartOptions(js.Opts{
    "width":       600,
    "height":      300,
    "colors":      []string{"#3b82f6", "#8b5cf6"},
    "valueFormat": "amount",
    "showValues":  true,
    "showLabels":  true,
})
```

### `.Render() string`

Outputs HTML: `<div id="cl_xxx"></div><script>__client({...})</script>`.

## Column Definition (js.Column)

Fluent builder for table column definitions. Start with `js.Col(key)`.

```go
js.Col("Firma").Label("Company").Sortable(true).Filterable(true)
js.Col("Castka").Label("Amount").Type("number").Format("amount").Sortable(true)
js.Col("DatVyst").Label("Date").Type("date").Sortable(true)
js.Col("Uhrazeno").Label("Paid").Type("bool")
js.Col("ICO").Label("ID").Class("w-32").CellClass("font-mono")
js.Col("Status").Label("Status").Type("enum").EnumOptions(
    js.Option{Value: "active", Label: "Active"},
    js.Option{Value: "closed", Label: "Closed"},
)
js.Col("Actions").Label("").Type("custom").Render(`return '<button onclick="alert(item.ID)">Edit</button>'`)
```

### Methods

| Method | Description |
|--------|-------------|
| `Col(key)` | Start builder with data key matching JSON field |
| `.Label(s)` | Header text. Defaults to key if empty |
| `.Type(s)` | `"text"`, `"number"`, `"date"`, `"bool"`, `"enum"`, `"custom"` |
| `.Format(s)` | `"amount"` (thousand-separated with decimals), `"date"`, `"number"` |
| `.Sortable(bool)` | Enable click-to-sort on header (asc → desc → none) |
| `.Filterable(bool)` | Include in filter bar |
| `.Class(s)` | Extra CSS class on `<th>` |
| `.CellClass(s)` | Extra CSS class on `<td>` |
| `.Render(s)` | Custom JS expression. Receives `item` and `i`. Return HTML string |
| `.EnumOptions(...)` | Value/label pairs for enum-type columns |

## Table Component (Built-in)

Registered as `"table"`. Auto-registered by the framework. Renders a styled `<table>` with headers, rows, optional search, pagination, sorting, and expandable rows.

### Go Sugar

```go
js.Client(ctx).
    Source("/api/invoices").
    Loading(ui.SkeletonTable).
    Table(
        js.Col("Firma").Label("Company").Sortable(true),
        js.Col("Castka").Label("Amount").Format("amount").Sortable(true),
        js.Col("DatVyst").Label("Date").Type("date").Sortable(true),
    ).
    Search(true).
    Pagination(25).
    Render()
```

### Direct Component Call (equivalent)

```go
js.Client(ctx).
    Source("/api/invoices").
    Loading(ui.SkeletonTable).
    Component("table", js.Opts{
        "columns": []map[string]any{
            {"key": "Firma", "label": "Company", "sortable": true},
            {"key": "Castka", "label": "Amount", "format": "amount", "sortable": true},
            {"key": "DatVyst", "label": "Date", "type": "date", "sortable": true},
        },
        "search":   true,
        "pageSize": 25,
    }).
    Render()
```

### Table Options

| Key | Type | Description |
|-----|------|-------------|
| `columns` | `[]column` | Column definitions (see js.Column) |
| `search` | `bool` | Show search input |
| `pageSize` | `int` | Rows per page (0 = no pagination) |
| `expandable` | `bool` | Click row to expand detail |
| `renderDetail` | `string` | JS function body for expanded row. Receives `item`. Return JSElement |
| `onRowClick` | `string` | JS function called with row data on click |
| `filter` | `bool` | Show filter bar |

### Sorting

Click sortable column headers to cycle: ascending → descending → none. State managed via `setState({ sort: { col, dir } })`. No re-fetch — sorts cached data client-side.

### Expandable Rows

```go
.Component("table", js.Opts{
    "columns":    columns,
    "expandable": true,
    "renderDetail": `
        var pairs = [];
        for (var k in item) {
            pairs.push(__cel.div("mb-1", [
                __cel("span", {class: "font-medium text-gray-500 mr-2"}, [__cel.text(k + ":")]),
                __cel("span", null, [__cel.text(String(item[k] || ""))])
            ]));
        }
        return __cel.div("text-sm", pairs);
    `,
})
```

If `renderDetail` is omitted, the table auto-generates a key:value detail view from all fields.

## Chart Component (Built-in)

Registered as `"chart"`. Renders pure SVG charts. Four types: `bar`, `area`, `hbar`, `donut`.

### Data Format

API should return an array of `{label, value}` objects:
```json
[
    {"label": "Jan", "value": 15000},
    {"label": "Feb", "value": 22000},
    {"label": "Mar", "value": 18500}
]
```

Plain number arrays also work — labels default to index.

### Bar Chart

```go
js.Client(ctx).
    Source("/api/dashboard/revenue").
    Chart(js.BarChart).
    ChartOptions(js.Opts{
        "width": 600, "height": 300,
        "colors": []string{"#3b82f6"},
        "valueFormat": "amount",
    }).
    Render()
```

### Area Chart

```go
js.Client(ctx).
    Source("/api/dashboard/trend").
    Chart(js.AreaChart).
    ChartOptions(js.Opts{"height": 250}).
    Render()
```

### Horizontal Bar Chart

```go
js.Client(ctx).
    Source("/api/dashboard/categories").
    Chart(js.HBarChart).
    ChartOptions(js.Opts{
        "barWidth": 24,
        "valueFormat": "amount",
    }).
    Render()
```

### Donut Chart

```go
js.Client(ctx).
    Source("/api/dashboard/breakdown").
    Chart(js.DonutChart).
    ChartOptions(js.Opts{
        "width": 300, "height": 300,
        "innerRadius": 80,
        "valueFormat": "amount",
    }).
    Render()
```

### Chart Options

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `type` | `string` | `"bar"` | `"bar"`, `"area"`, `"hbar"`, `"donut"` |
| `width` | `int` | `600` | SVG width (donut: `300`) |
| `height` | `int` | `300` | SVG height |
| `colors` | `[]string` | 8-color palette | Fill colors, cycled |
| `valueFormat` | `string` | raw | `"amount"` or `"number"` |
| `showValues` | `bool` | `true` | Show value labels on bars/points |
| `showLabels` | `bool` | `true` | Show axis labels |
| `barWidth` | `int` | auto | Bar width in pixels |
| `gap` | `int` | auto | Gap between bars |
| `innerRadius` | `int` | `0.6*r` | Donut hole radius |

## Writing Custom Components

A component is a JS file that calls `__cregister(name, fn)`. The function receives `(data, state, opts)` and returns a JSElement tree.

### 1. Write the JS component

```js
// File: server/assets/js/components/card-grid.js
__cregister("card-grid", function(data, state, opts) {
    // data  = API response (array or object)
    // state = {zoneId, page, sort, filters, search, ...}
    // opts  = JSON from Go js.Opts

    var items = Array.isArray(data) ? data : [];

    // Client-side search
    if (state.search) {
        var q = state.search.toLowerCase();
        items = items.filter(function(item) {
            return (item[opts.primary] || "").toLowerCase().indexOf(q) >= 0;
        });
    }

    return __cel.div("space-y-4", [
        // Search input
        opts.searchable ? __cel("input", {
            class: "input w-64",
            placeholder: "Search...",
            value: state.search || ""
        }, null, __cel.on("input",
            "__clients['" + state.zoneId + "'].setState({search:this.value})"
        )) : null,

        // Card grid
        __cel.div("grid grid-cols-" + (opts.cols || 3) + " gap-4",
            __cel.map(items, function(item) {
                return __cel.div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow", [
                    __cel.div("font-bold text-lg", [
                        __cel.text(item[opts.primary] || "")
                    ]),
                    opts.secondary
                        ? __cel.div("text-sm text-gray-500", [
                            __cel.text(item[opts.secondary] || "")
                        ])
                        : null,
                    opts.valueKey
                        ? __cel.div("text-lg font-semibold mt-2", [
                            __cel.text(__cfmt.amount(item[opts.valueKey]))
                        ])
                        : null,
                ]);
            })
        )
    ]);
});
```

### 2. Load it via `<script src>`

The JS file is served as a static asset. It loads after framework JS but before boot scripts:

```go
// In your app setup:
app.Static("/js/", "server/assets/js/")
```

Include in your HTML head or layout:
```go
app.HTMLHead = append(app.HTMLHead, `<script src="/js/components/card-grid.js"></script>`)
```

### 3. Reference from Go

```go
js.Client(ctx).
    Source("/api/companies").
    Loading(ui.SkeletonCards).
    Empty("business", "No companies found").
    Component("card-grid", js.Opts{
        "cols":       3,
        "primary":    "Firma",
        "secondary":  "ICO",
        "valueKey":   "Obrat",
        "searchable": true,
    }).
    Render()
```

### Component Function Signature

```js
function(data, state, opts) → JSElement | null
```

| Param | Type | Description |
|-------|------|-------------|
| `data` | `any` | Raw JSON from API (usually array or object) |
| `state` | `object` | `{zoneId, page, sort, filters, search, ...}` + any custom keys |
| `opts` | `object` | JSON from `js.Opts` passed in Go |

Return a JSElement tree (built with `__cel`). Return `null` to show empty state.

### Available Tools Inside Components

- `__cel` — build JSElement trees
- `__cfmt` — format dates, amounts, numbers, booleans
- `__capi` — fetch additional data
- `__caction` — trigger server actions
- `__clients[state.zoneId].setState({...})` — re-render with new state (no re-fetch)
- `__clients[state.zoneId].reload()` — re-fetch from API + re-render
- `__cfilter` — filter/sort/paginate data arrays
- `__cfilterbar` — render filter bar with chips
- `__cpagination` — render pagination controls

## JS Module Reference

### `__cel(tag, attrs, children, events)`

Creates a JSElement JSON object consumed by `__engine.create()`.

```js
// Full form
__cel("div", {class: "p-4", id: "myDiv"}, [children...], {click: {act: "raw", js: "alert(1)"}})

// Returns: {t: "div", a: {class: "p-4", id: "myDiv"}, c: [...], e: {...}}
```

**Helpers:**

```js
__cel.div("classes", [children])     // shorthand for __cel("div", {class: ...}, [...])
__cel.span("classes", [children])    // shorthand for __cel("span", {class: ...}, [...])
__cel.text("string")                 // plain text node
__cel.icon("icon_name", "classes")   // <span class="material-icons ...">icon_name</span>
__cel.on("event", "jsFn")           // returns event object: {event: {act: "raw", js: jsFn}}
__cel.if(cond, thenFn, elseFn)      // conditional rendering
__cel.map(array, fn)                // maps array to JSElement array, filters nulls
```

**Nulls are safe.** `null`, `undefined`, and `false` children are silently skipped.

### `__cfmt`

Client-side formatters. Locale from `window.__locale` (set by server via `ui.MakeApp("sk")`).

```js
__cfmt.date(val)           // "2025-01-15" → "15.01.2025"
__cfmt.amount(val)         // 1234567.89 → "1 234 567,89"
__cfmt.number(val)         // 1234567 → "1 234 567" (no forced decimals)
__cfmt.bool(val)           // true → "check_circle", false → "cancel" (icon names)
__cfmt.truncate(val, max)  // "long string..." truncated to max chars
__cfmt.escape(val)         // HTML-escape &, <, >, "
```

### `__capi`

HTTP client with request deduplication for GET.

```js
__capi.get(url, params)    // GET, returns Promise<JSON>. Deduplicates concurrent identical requests
__capi.post(url, body)     // POST JSON, returns Promise<JSON>
__capi.buildUrl(url, params) // Build URL with query string (internal helper)
```

```js
// Example: fetch sub-detail inside a component
__capi.get("/api/invoice/" + item.ID + "/items").then(function(items) {
    // render sub-table
});
```

### `__caction(path, payload)`

Triggers a server action (POST via `__post`) from client-rendered content. Bridges client-side rendering with server-side action handlers.

```js
// In a component's event handler:
__caction("/api/invoice/delete", {id: item.ID})

// In a __cel event binding:
__cel("button", {class: "btn"}, [__cel.text("Delete")],
    __cel.on("click", "__caction('/api/invoice/delete', {id: " + item.ID + "})"))
```

### `__clients[id]`

Instance methods for each client zone. Access via `__clients[state.zoneId]` inside components.

```js
__clients[id].setState(partial)  // Merge into state, re-render (no re-fetch)
__clients[id].reload()           // Re-fetch from source + re-render
__clients[id].refetch()          // Alias for reload()
__clients[id].destroy()          // Clean up interval, clear DOM
__clients[id].getState()         // Returns current state object
__clients[id].getData()          // Returns cached data
```

**`setState` is the primary interaction model.** Components call it to update page, sort, search, filters, or any custom state key. This triggers a full re-render with cached data (no network request).

```js
// Sort
__clients[zoneId].setState({ sort: { col: "Firma", dir: "asc" }, page: 0 })

// Search
__clients[zoneId].setState({ search: "Smith", page: 0 })

// Custom state
__clients[zoneId].setState({ selectedTab: "details" })
```

### `__cfilter`

Shared data processing utilities. Used by the table component internally, available to custom components.

```js
__cfilter.applyToData(data, filterState, columns)  // Apply search/filter/sort to array
__cfilter.paginate(data, page, pageSize)            // Returns {items, totalPages, page, total}
__cfilter.toQueryString(filterState)                // Convert filter state to URL params
__cfilter.cycleSort(currentSort, colKey)            // Cycle: none → asc → desc → none
```

### `__cfilterbar(zoneId, state, columns)`

Renders a filter bar with search input and active filter chips. Returns JSElement or `null`.

### `__cpagination(zoneId, page, totalPages, pageSize, total)`

Renders pagination controls with prev/next, page numbers with ellipsis, and record count. Returns JSElement or `null`.

## Loading Order

```
1. Framework JS (embedded in Go, emitted by Script() in <head>)
   __cfmt, __capi, __cel, __cregister, __cfilter, __cfilterbar, __cpagination, __caction

2. Framework components (embedded, auto-registered)
   __cregister("table", ...)
   __cregister("chart", ...)

3. __client() boot function (embedded)

4. Project JS (static assets, <script src="...">)
   components/card-grid.js  → __cregister("card-grid", ...)
   components/kpi-bar.js    → __cregister("kpi-bar", ...)

5. Boot scripts (inline in HTML, per js.Client(...).Render() call)
   <script>__client({id:"cl_1", source:"/api/...", component:"table", ...})</script>
   <script>__client({id:"cl_2", source:"/api/...", component:"kpi-bar", ...})</script>
```

Framework JS and components are always available. Project component files must load before boot scripts. Boot scripts execute inline where `.Render()` is called.

## Examples

### Simple Table Page

```go
func invoicesPage(ctx *ui.Context) string {
    return app.HTML("Invoices", "bg-gray-100",
        ui.Div("p-6")(
            ui.H1("text-2xl font-bold mb-6")("Invoices"),
            js.Client(ctx).
                Source("/api/invoices").
                Loading(ui.SkeletonTable).
                Empty("receipt_long", "No invoices").
                Table(
                    js.Col("Firma").Label("Company").Sortable(true),
                    js.Col("Castka").Label("Amount").Format("amount").Sortable(true),
                    js.Col("DatVyst").Label("Issued").Type("date").Sortable(true),
                    js.Col("Uhrazeno").Label("Paid").Type("bool"),
                ).
                Search(true).
                Pagination(20).
                Render(),
        ),
    )
}
```

### Dashboard with KPIs + Chart + Table

```go
func dashboardPage(ctx *ui.Context) string {
    return app.HTML("Dashboard", "bg-gray-100",
        ui.Div("p-6 space-y-6")(
            ui.H1("text-2xl font-bold")("Dashboard"),

            // KPI bar (custom component)
            js.Client(ctx).
                Source("/api/dashboard/kpis").
                Component("kpi-bar", js.Opts{
                    "items": []string{"revenue", "expenses", "profit", "invoices"},
                }).
                Render(),

            // Revenue chart
            ui.Div("grid grid-cols-2 gap-6")(
                js.Client(ctx).
                    Source("/api/dashboard/revenue").
                    Chart(js.BarChart).
                    ChartOptions(js.Opts{
                        "height":      280,
                        "valueFormat": "amount",
                        "colors":      []string{"#3b82f6"},
                    }).
                    Render(),

                js.Client(ctx).
                    Source("/api/dashboard/breakdown").
                    Chart(js.DonutChart).
                    ChartOptions(js.Opts{
                        "width":       280,
                        "height":      280,
                        "valueFormat": "amount",
                    }).
                    Render(),
            ),

            // Recent invoices table
            js.Client(ctx).
                Source("/api/dashboard/recent-invoices").
                Loading(ui.SkeletonTable).
                Table(
                    js.Col("Firma").Label("Company").Sortable(true),
                    js.Col("Castka").Label("Amount").Format("amount"),
                    js.Col("DatVyst").Label("Date").Type("date"),
                ).
                Pagination(10).
                Render(),
        ),
    )
}
```

### Custom Card Grid Component

**JS component** (`server/assets/js/components/card-grid.js`):

```js
__cregister("card-grid", function(data, state, opts) {
    var items = Array.isArray(data) ? data : [];

    if (state.search) {
        var q = state.search.toLowerCase();
        items = items.filter(function(item) {
            return (item[opts.primary] || "").toLowerCase().indexOf(q) >= 0;
        });
    }

    return __cel.div("space-y-4", [
        opts.searchable ? __cel("input", {
            class: "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-64",
            placeholder: "Search...",
            value: state.search || ""
        }, null, __cel.on("input",
            "__clients['" + state.zoneId + "'].setState({search:this.value})"
        )) : null,

        __cel.div("grid grid-cols-" + (opts.cols || 3) + " gap-4",
            __cel.map(items, function(item) {
                return __cel.div("bg-white dark:bg-gray-900 rounded-lg p-4 shadow hover:shadow-md cursor-pointer", [
                    __cel.div("font-bold text-lg mb-1", [__cel.text(item[opts.primary] || "")]),
                    opts.secondary
                        ? __cel.div("text-sm text-gray-500", [__cel.text(item[opts.secondary] || "")])
                        : null,
                    opts.valueKey
                        ? __cel.div("text-xl font-semibold mt-3 text-blue-600", [
                            __cel.text(__cfmt.amount(item[opts.valueKey]))
                        ])
                        : null,
                ]);
            })
        )
    ]);
});
```

**Go usage:**

```go
js.Client(ctx).
    Source("/api/companies").
    Loading(ui.SkeletonCards).
    Empty("business", "No companies found").
    Component("card-grid", js.Opts{
        "cols":       3,
        "primary":    "Firma",
        "secondary":  "ICO",
        "valueKey":   "Obrat",
        "searchable": true,
    }).
    Render()
```

---

## New Features

Everything below was added in the framework expansion (Sections A-J of the backlog).

---

## Core JS Utilities

### `__debounce(fn, delay)`

Debounces a function call. Returns a new function that delays invoking `fn` until `delay` ms have passed since the last call. The returned function has a `.cancel()` method.

```js
// In a component event handler:
var debouncedSearch = __debounce(function(q) {
    __clients[zoneId].setState({ search: q });
}, 300);
// Usage in __cel event:
__cel.on("input", "debouncedSearch(this.value)")
```

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `fn` | `function` | required | Function to debounce |
| `delay` | `int` | `250` | Delay in milliseconds |

### `__clipboard(text)`

Copies text to clipboard using `navigator.clipboard.writeText` with `document.execCommand("copy")` fallback. Shows a toast notification on success.

```js
__clipboard("Hello World")
// In a __cel event:
__cel.on("click", "__clipboard('some text')")
```

Returns a `Promise` that resolves when text is copied.

---

## Formatter Additions (`__cfmt`)

### `__cfmt.relativeTime(dateStr)`

Converts a date string or Date object to a human-readable relative time string.

```js
__cfmt.relativeTime("2025-12-01T10:00:00Z")  // "3 days ago"
__cfmt.relativeTime(new Date())               // "just now"
__cfmt.relativeTime("2026-01-15T00:00:00Z")   // "in 5d" (future dates)
```

Output examples: `"just now"`, `"30s ago"`, `"5 min ago"`, `"3 hours ago"`, `"2 days ago"`, `"in 3d"`. Falls back to `__cfmt.date()` for dates older than 7 days.

### `__cfmt.datePreset(name)`

Returns a `{from, to}` object with ISO date strings for common date range presets.

```js
var range = __cfmt.datePreset("thisMonth");
// { from: "2026-03-01", to: "2026-03-10" }

var q1 = __cfmt.datePreset("thisQuarter");
// { from: "2026-01-01", to: "2026-03-10" }
```

| Preset | Description |
|--------|-------------|
| `"today"` | Today only |
| `"thisWeek"` | Monday to today |
| `"thisMonth"` | 1st of month to today |
| `"thisQuarter"` | 1st of quarter to today |
| `"thisYear"` | Jan 1st to today |
| `"lastMonth"` | Full previous month |
| `"lastYear"` | Full previous year |

---

## Client API Additions (`__capi`)

### `__capi.setBase(url)`

Sets a base URL for all subsequent API calls. Enables CORS mode and `credentials: "include"` automatically.

```js
__capi.setBase("https://api.example.com")
// Now all __capi.get/post/upload/download calls prefix this base URL
```

### `__capi.upload(url, formData, onProgress)`

Uploads files via `FormData`. Uses `XMLHttpRequest` when `onProgress` is provided (for progress tracking), falls back to `fetch` otherwise.

```js
var fd = new FormData();
fd.append("files", fileInput.files[0]);

__capi.upload("/api/upload", fd, function(progress) {
    console.log(Math.round(progress * 100) + "%");  // 0.0 to 1.0
}).then(function(result) {
    console.log("Upload complete:", result);
});
```

| Param | Type | Description |
|-------|------|-------------|
| `url` | `string` | Upload endpoint |
| `formData` | `FormData` | Files + metadata |
| `onProgress` | `function(ratio)` | Optional. Called with `0.0` to `1.0` progress |

Returns `Promise<JSON>`.

### `__capi.download(url, filename)`

Downloads a file by fetching a URL and triggering a browser save dialog.

```js
__capi.download("/api/export/invoices.xlsx", "invoices.xlsx")
```

Creates a temporary `<a>` element with `blob:` URL and clicks it programmatically.

### `__capi.healthCheck(url, opts)`

Checks if a URL is reachable with configurable retries and timeout.

```js
__capi.healthCheck("/api/health", {
    timeout: 5000,    // ms per attempt
    retries: 3,       // number of attempts
    interval: 2000,   // ms between retries
    onChange: function(isUp) {
        console.log("Service is " + (isUp ? "UP" : "DOWN"));
    }
}).then(function(ok) {
    console.log("Final result:", ok);
});
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `timeout` | `int` | `5000` | Timeout per attempt in ms |
| `retries` | `int` | `3` | Number of retry attempts |
| `interval` | `int` | `2000` | Delay between retries in ms |
| `onChange` | `function(bool)` | `null` | Called when state changes |

Returns `Promise<boolean>`.

---

## Filter System

### URL Sync

Filter state (search, sort, page, column filters) is automatically persisted to URL query parameters via `history.replaceState`. On page load, state is restored from URL params first, then falls back to `localStorage`.

**URL parameter format:**
- `?search=query` — search term
- `?sort=Firma:asc` — sort column and direction
- `?page=2` — current page
- `?f.Status=eq:active` — column filter (prefix `f.`)

**localStorage key:** `__gsui_filter_<pathname>`

This is automatic for all client zones — no configuration needed.

### Column Filter Dropdowns

Enable per-column filtering by adding `.Filterable(true)` to column definitions and `.Filter(true)` to the table:

```go
js.Client(ctx).
    Source("/api/invoices").
    Table(
        js.Col("Firma").Label("Company").Sortable(true).Filterable(true),
        js.Col("Castka").Label("Amount").Type("number").Filterable(true),
        js.Col("DatVyst").Label("Date").Type("date").Filterable(true),
        js.Col("Status").Label("Status").Type("enum").Filterable(true).EnumOptions(
            js.Option{Value: "active", Label: "Active"},
            js.Option{Value: "closed", Label: "Closed"},
        ),
        js.Col("Uhrazeno").Label("Paid").Type("bool").Filterable(true),
    ).
    Filter(true).
    Render()
```

**Filter types by column type:**

| Column Type | Filter UI | Operators |
|-------------|-----------|-----------|
| `text` / `custom` | Text input | contains, startsWith, equals |
| `number` | Number input | eq, gt, lt, gte, lte, between |
| `date` | Date range picker + presets | today, thisWeek, thisMonth, thisYear |
| `enum` | Multi-checkbox list | All, None, individual values |
| `bool` | Radio buttons | Any, Yes, No |

Each dropdown has **Apply** and **Clear** buttons. Active filters show a blue filter icon on the column header.

---

## Table Enhancements

### Async Expandable Row Detail (`detailSource`)

Fetch detail data lazily when a row is expanded, instead of rendering from cached data:

```go
.Component("table", js.Opts{
    "columns":      columns,
    "expandable":   true,
    "detailSource": "/api/invoices/{id}/detail",
})
```

The `{id}` and `{key}` tokens in the URL are replaced with the row's `id` or `ID` field value. Data is cached per row in `state._detailCache`.

Shows a skeleton while loading. On error, shows an error message with retry button.

### Tabbed Detail Views (`detailTabs`)

Render a tabbed interface in the expanded row area. Each tab can lazily fetch its own data:

```go
.Component("table", js.Opts{
    "columns":    columns,
    "expandable": true,
    "detailTabs": []map[string]string{
        {"label": "Overview", "source": "/api/invoices/{id}/overview"},
        {"label": "Items",    "source": "/api/invoices/{id}/items"},
        {"label": "History",  "source": "/api/invoices/{id}/history"},
    },
})
```

- Tabs appear as a horizontal tab bar in the expanded row
- Each tab fetches data from its `source` URL on first activation
- Tab data is cached independently (key: `cacheKey_tabN`)
- `{id}` and `{key}` tokens are replaced with row data
- Active tab tracked via `state._activeTab`

---

## Chart Enhancements

### Two-Series Bar Charts

Support for comparing two data series side-by-side in bar charts.

**Using `value2`:**
```json
[
    {"label": "Jan", "value": 15000, "value2": 12000},
    {"label": "Feb", "value": 22000, "value2": 19000}
]
```

**Using `series` array:**
```json
[
    {"label": "Jan", "series": [{"name": "2025", "value": 15000}, {"name": "2024", "value": 12000}]},
    {"label": "Feb", "series": [{"name": "2025", "value": 22000}, {"name": "2024", "value": 19000}]}
]
```

```go
js.Client(ctx).
    Source("/api/revenue/compare").
    Chart(js.BarChart).
    ChartOptions(js.Opts{
        "colors":      []string{"#3b82f6", "#94a3b8"},
        "seriesName":  "2025",
        "series2Name": "2024",
    }).
    Render()
```

A legend row is rendered below the chart when multiple series are present.

### SVG Tooltips

All chart types now include `<title>` elements on interactive elements for native browser tooltips on hover:
- **Bar chart:** tooltip on each bar showing `"Label: FormattedValue"`
- **Area chart:** tooltip on each dot
- **Horizontal bar chart:** tooltip on each bar
- **Donut chart:** tooltip on each slice showing label, value, and percentage

No configuration needed — tooltips are always present.

---

## New Components

### File Upload (`file-upload`)

Drag & drop file upload component with progress tracking, file validation, and batch upload.

**Go Builder sugar:**
```go
js.Client(ctx).
    Source("/api/upload").
    Upload(33554432, ".pdf,.doc,.xlsx").  // 32MB max, specific types
    MaxFiles(5).
    Render()
```

**Direct component call:**
```go
js.Client(ctx).
    Source("/api/upload").
    Component("file-upload", js.Opts{
        "uploadUrl": "/api/upload",
        "maxSize":   33554432,
        "accept":    ".pdf,.doc,.xlsx",
        "multiple":  true,
        "maxFiles":  5,
    }).
    Render()
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `uploadUrl` | `string` | `""` | POST endpoint for upload |
| `maxSize` | `int` | `33554432` (32MB) | Max file size in bytes |
| `accept` | `string` | `""` | Accepted file types (MIME or extensions) |
| `multiple` | `bool` | `true` | Allow multiple files |
| `maxFiles` | `int` | `20` | Maximum number of files |

Features:
- Drag & drop zone with visual feedback
- Click to browse files
- File list with size display and remove buttons
- Progress bar during upload (uses `__capi.upload` with XHR for progress)
- Success/error notifications via `__notify`

### Confirm Dialog (`__confirm`)

Promise-based confirmation dialog. Creates a modal overlay dynamically.

```js
__confirm("Are you sure you want to delete this?", function() {
    __caction("/api/delete", { id: 123 });
}, {
    title: "Delete Invoice",
    confirmText: "Delete",
    cancelText: "Keep",
    variant: "danger"
}).then(function(confirmed) {
    console.log("User chose:", confirmed);  // true or false
});
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `title` | `string` | `"Confirm"` | Dialog title |
| `confirmText` | `string` | `"Confirm"` | Confirm button label |
| `cancelText` | `string` | `"Cancel"` | Cancel button label |
| `variant` | `string` | `"default"` | `"default"` (blue) or `"danger"` (red) |

Closes on: confirm click, cancel click, overlay click, Escape key.

### Toast API (`__toast`)

Convenience wrapper around `__notify` for showing toast notifications.

```js
__toast.show("Operation completed", "success")
__toast.success("Saved!")
__toast.error("Something went wrong")
__toast.info("Processing...")
```

### Modal Preview (`__cmodal`)

Multi-purpose modal for previewing images, text, or HTML content.

```js
// Image preview
__cmodal.open("/images/photo.jpg")
__cmodal.preview("/images/photo.jpg", { alt: "Photo" })

// HTML content
__cmodal.open("<h2>Invoice Details</h2><p>Amount: $1,234</p>")

// Plain text
__cmodal.open("Long text content here...")

// DOM element
__cmodal.open(someElement)

// Close
__cmodal.close()
```

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `maxWidth` | `string` | `"90vw"` | Maximum modal width |
| `alt` | `string` | `""` | Alt text for images |

Auto-detects content type: image URLs (`.jpg`, `.png`, etc.), HTML strings (starting with `<`), plain text, or DOM elements.

### Badge Counter (`badge`)

Client-rendered badge component. Shows a count with configurable appearance. Hides when count is zero (by default).

```go
js.Client(ctx).
    Source("/api/notifications/count").
    Component("badge", js.Opts{
        "bg":       "bg-red-500",
        "text":     "text-white",
        "size":     "min-w-[20px] h-5 text-xs",
        "hideZero": true,
    }).
    Poll(30 * time.Second).
    Render()
```

API should return `{"count": 5}`, `{"value": 5}`, or a plain number.

---

## Conditional Polling (`.PollWhile`)

### `Builder.PollWhile(condition string)`

Sets a JavaScript expression that is evaluated on each poll cycle. Polling stops when the expression returns a falsy value.

```go
js.Client(ctx).
    Source("/api/tasks/status").
    Component("task-status", js.Opts{}).
    Poll(3 * time.Second).
    PollWhile(`data.status !== "completed"`).
    Render()
```

The `condition` string is a JS expression with access to `data` (last API response) and `state` (current zone state). When the expression returns `false`, the interval is cleared and polling stops.

**Examples:**
```go
.PollWhile(`data.status === "processing"`)    // Stop when done
.PollWhile(`data.progress < 100`)             // Stop at 100%
.PollWhile(`state.page === 0`)                // Only poll on first page
```

---

## Go Helper Components (`js` package)

These are standalone Go functions that render self-contained HTML + inline `<script>` blocks. They don't use the `Builder` pattern — call them directly in your page functions.

### `js.LiveSearch(targetSelector, inputID, class)`

Renders a search input that filters server-rendered HTML elements by text content.

```go
// Page function
func listPage(ctx *ui.Context) string {
    return ui.Div("p-6")(
        js.LiveSearch(".list-item", "search_input", ""),
        ui.Div("mt-4")(
            ui.Div("list-item p-2")("Apple"),
            ui.Div("list-item p-2")("Banana"),
            ui.Div("list-item p-2")("Cherry"),
        ),
    )
}
```

| Param | Type | Description |
|-------|------|-------------|
| `targetSelector` | `string` | CSS selector for filterable items |
| `inputID` | `string` | HTML ID for the search input |
| `class` | `string` | CSS classes (uses default if empty) |

### `js.ContentSearch(containerSelector, triggerKey)`

Renders an in-page text search bar (like Ctrl+F) for a container. Uses `TreeWalker` to find and highlight text matches with `<mark>` elements.

```go
js.ContentSearch("#main-content", "/")
```

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `containerSelector` | `string` | required | CSS selector for search area |
| `triggerKey` | `string` | `"/"` | Key to open/close search bar |

Features:
- Enter: next match, Shift+Enter: previous match
- Escape: close search bar
- Active match highlighted in amber, others in yellow
- Match counter (e.g. "3/12")
- Auto-scroll to active match

### `js.Autocomplete(inputID, sourceURL, class)`

Renders an input with a browser-native `<datalist>` for autocomplete. Fetches options from an API endpoint on page load.

```go
js.Autocomplete("company_input", "/api/companies/names", "")
```

API should return `["Option1", "Option2"]` or `[{"value": "v1", "label": "Label 1"}, ...]`.

### `js.AsyncButton(label, url, resultID, class)`

Renders a button that POSTs to a URL on click, shows loading state, and displays the result.

```go
ui.Div("")(
    js.AsyncButton("Generate Report", "/api/reports/generate", "report_result", ""),
)
```

| Param | Type | Description |
|-------|------|-------------|
| `label` | `string` | Button text |
| `url` | `string` | POST endpoint |
| `resultID` | `string` | HTML ID for result display |
| `class` | `string` | CSS classes (uses default if empty) |

Collects form data from closest `<form>` if present. Shows green success or red error message.

### `js.AutoFill(selectID, mappings)`

Renders a script that auto-populates form fields when a `<select>` element changes.

```go
js.AutoFill("template_select", map[string]map[string]string{
    "invoice": {
        "subject_field": "Invoice",
        "amount_field":  "0.00",
    },
    "receipt": {
        "subject_field": "Receipt",
        "amount_field":  "0.00",
    },
})
```

| Param | Type | Description |
|-------|------|-------------|
| `selectID` | `string` | HTML ID of the trigger select |
| `mappings` | `map[string]map[string]string` | `optionValue → {fieldID → fieldValue}` |

Dispatches `input` events on populated fields for reactivity.

### `js.AjaxForm(formID, opts)`

Intercepts form submission and uses `fetch` instead of native form submit. Serializes fields as JSON.

```go
js.AjaxForm("invoice_form", map[string]string{
    "successRedirect": "/invoices",
})
// or
js.AjaxForm("settings_form", map[string]string{
    "onSuccess": "alert('Saved: ' + JSON.stringify(data))",
    "onError":   "alert('Error: ' + err.message)",
})
```

| Option | Type | Description |
|--------|------|-------------|
| `successRedirect` | `string` | URL to navigate to on success |
| `onSuccess` | `string` | JS function body, receives `data` param |
| `onError` | `string` | JS function body, receives `err` param |

Shows loading state on submit button. Falls back to `__notify` for feedback if no callbacks provided.

### `js.SPA(id, class)` + `js.SPALink(target, url, class, content)`

Client-side navigation using `fetch` + `pushState`. Loads HTML pages into a container without full page reloads.

```go
// Layout: create the SPA container
ui.Div("flex")(
    ui.Nav("w-64")(
        js.SPALink("main", "/dashboard", "", "Dashboard"),
        js.SPALink("main", "/invoices", "", "Invoices"),
        js.SPALink("main", "/settings", "", "Settings"),
    ),
    js.SPA("main", "flex-1 p-6"),
)
```

**`js.SPA(id, class)`** — creates the container div and registers the `window.__spa` navigation system. Handles `popstate` for back/forward buttons. Executes inline `<script>` tags in loaded content.

**`js.SPALink(target, url, class, content)`** — renders an `<a>` that triggers SPA navigation on click. Falls back to normal navigation if JS fails.

**Programmatic navigation:**
```js
window.__spa.load("main", "/new-page")        // with pushState
window.__spa.load("main", "/new-page", false)  // without pushState (for popstate)
```

### `js.ExternalLink(url, class, content)`

Renders a link that bypasses SPA interception by using `window.location.href` directly.

```go
js.ExternalLink("https://github.com/example", "", "GitHub")
```

### `js.Shortcuts()` + `js.RegisterShortcut(key, jsHandler, description)`

Keyboard shortcut framework with single-key, modifier, and sequence support.

```go
// In layout (once):
js.Shortcuts()

// Register shortcuts anywhere in pages:
js.RegisterShortcut("n", "window.location.href='/invoices/new'", "New invoice")
js.RegisterShortcut("ctrl+s", "document.getElementById('main-form').submit()", "Save")
js.RegisterShortcut("gi", "window.location.href='/invoices'", "Go to invoices")
```

**Key formats:**
- Single key: `"n"`, `"/"`, `"?"` (built-in: shows help overlay)
- Modifier: `"ctrl+s"`, `"shift+n"`, `"alt+d"`, `"meta+k"` / `"cmd+k"`
- Sequence: `"gi"` (type `g` then `i` within 800ms)

**Built-in `?` shortcut:** Opens a modal overlay listing all registered shortcuts with their descriptions and key bindings.

**JS API:**
```js
window.__shortcuts.register("x", function(e) { ... }, "Description")
window.__shortcuts.unregister("x")
window.__shortcuts.list()   // [{key, description}, ...]
window.__shortcuts.flash(el)  // Brief blue highlight on an element
```

Shortcuts are ignored when the user is typing in `<input>`, `<textarea>`, `<select>`, or `contentEditable` elements.

### `js.Toast(message, variant)`

Server-side convenience for triggering a toast notification on page load.

```go
// In a POST handler redirect response:
return app.HTML("Success", "",
    js.Toast("Invoice saved successfully", "success"),
    invoiceListContent(ctx),
)
```

| Variant | Description |
|---------|-------------|
| `"success"` | Green success toast |
| `"error"` | Red error toast |
| `"info"` | Blue info toast |

### `ui.ConfirmDialog(title, message, confirmAction, cancelURL, class)`

Server-rendered confirmation dialog overlay. Unlike `__confirm` (client-side), this renders a full HTML modal with a `<form>` that POSTs to `confirmAction`.

```go
ui.ConfirmDialog(
    "Delete Invoice",
    "Are you sure you want to delete invoice #1234?",
    "/api/invoices/1234/delete",
    "/invoices",  // cancel navigates here (empty = close overlay)
    "",           // default classes
)
```

| Param | Type | Description |
|-------|------|-------------|
| `title` | `string` | Dialog title |
| `message` | `string` | Dialog body text |
| `confirmAction` | `string` | POST URL for confirm button |
| `cancelURL` | `string` | Navigation URL for cancel (empty = JS close) |
| `class` | `string` | CSS classes for dialog (uses default if empty) |

---

## Updated Loading Order

```
1. Framework JS (embedded in Go, emitted by Script() in <head>)
   __cfmt, __capi, __cel, __cregister, __cfilter, __cfilterbar, __cpagination, __caction
   __debounce, __clipboard                    ← NEW
   __confirm, __toast, __cmodal, __cbadge     ← NEW

2. Framework components (embedded, auto-registered)
   __cregister("table", ...)     ← enhanced: column filters, detailSource, detailTabs
   __cregister("chart", ...)     ← enhanced: two-series, SVG tooltips
   __cregister("file-upload", ...) ← NEW

3. __client() boot function (embedded)
   ← enhanced: URL sync, localStorage fallback, pollWhile

4. Project JS (static assets, <script src="...">)
   components/card-grid.js  → __cregister("card-grid", ...)

5. Boot scripts (inline in HTML, per js.Client(...).Render() call)
   <script>__client({id:"cl_1", ...})</script>
```

Total: 32 JS modules in the framework bundle.
