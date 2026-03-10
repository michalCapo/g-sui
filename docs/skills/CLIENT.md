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
