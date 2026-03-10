# Todo: g-sui Client-Side Rendering Support

## Overview

Add hybrid client-side rendering to the g-sui framework. The server renders the page shell (layout, headers, empty containers), then a client-side JS runtime fetches data from APIs and renders content using the existing `__engine.create()` JSElement JSON DOM protocol.

**Target:** `github.com/michalCapo/g-sui/ui` package (framework-level, reusable across projects)
**Events:** Client-side only (no server roundtrip for filter/sort/expand interactions)

### Core Idea

**`Client` is thin plumbing.** It does exactly 4 things:
1. Emits a `<div>` target container
2. Emits a `<script>` boot tag that fetches data from an API
3. Shows loading/error/empty states
4. Calls a **registered JS component** with the fetched data

**Components are JS functions.** All rendering logic lives in JS files. You register them with `__cregister("name", fn)` and reference them from Go by name + JSON opts. The framework ships `table` and `chart` as built-in registered components. Projects register their own (card grids, KPI bars, whatever).

**Go sugar for common cases.** `.Table(columns...)` on the builder is just a convenience that compiles column definitions to JSON and calls `.Component("table", compiledOpts)` internally. You never *need* it — you could always write `.Component("table", opts)` directly.

---

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│ SERVER (Go)                                              │
│                                                          │
│  ui.Client(ctx).                                    │
│      Source("/api/invoices").       ← zone plumbing      │
│      Loading(ui.SkeletonTable).    ← zone plumbing      │
│      Empty("inbox", "No data").    ← zone plumbing      │
│      Component("table", opts).     ← what to render     │
│      Render()                                            │
│                                                          │
│  Sugar (compiles to Component("table", ...)):            │
│  ui.Client(ctx).                                    │
│      Source("/api/invoices").                             │
│      Table(columns...).                                  │
│      Render()                                            │
│                                                          │
│  Output: <div id="cl_xxx"></div>                         │
│          <script>__client({                              │
│              id: "cl_xxx",                               │
│              source: "/api/invoices",                    │
│              component: "table",                         │
│              opts: { columns: [...], ... },              │
│              loading: "table",                           │
│              ...                                         │
│          })</script>                                     │
└───────────────────────┬──────────────────────────────────┘
                        │ HTML response
                        ▼
┌──────────────────────────────────────────────────────────┐
│ CLIENT (Browser JS)                                      │
│                                                          │
│  Registry:                                               │
│    __cregister("table", fn)    ← framework built-in      │
│    __cregister("chart", fn)    ← framework built-in      │
│    __cregister("card-grid", fn) ← project component      │
│    __cregister("kpi-bar", fn)   ← project component      │
│    __cregister("my-thing", fn)  ← project component      │
│                                                          │
│  __client(config) boots:                                 │
│    1. Show skeleton in <div id="cl_xxx">                 │
│    2. fetch(config.source) → JSON data                   │
│    3. fn = __cget(config.component)                      │
│    4. jsElement = fn(data, state, config.opts)            │
│    5. __engine.create(jsElement) → DOM                   │
│    6. Replace div contents                               │
│                                                          │
│  State: { zoneId, page, sort, filters, search, ... }    │
│  setState(partial) → re-render                           │
│  reload() → re-fetch + re-render                         │
└──────────────────────────────────────────────────────────┘
```

**Every component — built-in or custom — goes through the same registry.** There is no special path for tables or charts. They're just components that happen to ship with the framework.

---

## Phase 1: Core Infrastructure

### 1.1 JSElement Client-Side Builder (`__cel`)

- [ ] Create a client-side JS module `__cel` (client element) that mirrors the server's `JSElement` builder
  - `__cel(tag, attrs, children)` → returns JSElement JSON object (not DOM node)
  - This is different from `__e()` which creates actual DOM nodes. `__cel` creates the JSON structure that `__engine.create()` consumes
  - Support: `{t: "div", a: {class: "...", id: "..."}, c: [...children], e: {click: {act: "raw", js: "..."}}}`
  - Helper shortcuts: `__cel.div(cls, ...children)`, `__cel.span(cls, ...children)`, `__cel.text(str)`
  - Event binding: `__cel.on(event, jsFn)` for client-side event handlers
  - Conditional: `__cel.if(cond, thenFn, elseFn)`
  - Loop: `__cel.map(arr, fn)` → array of JSElement JSON objects
  - Note: This is a lightweight JSON factory, NOT a reactive system. It produces static JSElement trees that `__engine.create()` materializes once
  - File: add to `ui/ui.server.go` as a new JS module string variable (same pattern as `__engine`, `__post`, etc.)
  - Size target: ~80-120 lines of JS

### 1.2 Component Registry (`__cregister` / `__cget`)

- [ ] Create the component registry JS module
  - `window.__cregistry = {}`
  - `__cregister(name, fn)` — stores render function. Signature: `function(data, state, opts) → JSElement`
  - `__cget(name)` — retrieves by name, throws if not found
  - Size target: ~15-20 lines of JS
  - File: add to `ui/ui.server.go` alongside `__cel`

### 1.3 Client Boot Function (`__client`)

- [ ] Create the `__client(config)` JS function — the zone orchestrator
  - Config shape (all fields set by Go builder, serialized as JSON):
    ```js
    {
      id: "cl_xxx",              // target div ID
      source: "/api/data",       // API endpoint URL
      params: {},                // extra query params
      component: "table",        // registered component name
      opts: {},                  // JSON config passed to component fn
      loading: "table",          // skeleton type: "table"|"cards"|"list"|"component"|"page"
      empty: { icon: "inbox", message: "No data" },
      error: true,               // show error state on fetch failure
      autoLoad: true,            // fetch on boot (default true)
      poll: 0,                   // polling interval in ms (0 = disabled)
    }
    ```
  - Lifecycle:
    1. Find `<div id="config.id">`
    2. Show skeleton loading state
    3. `fetch(config.source + queryString)` with error handling
    4. `var fn = __cget(config.component)`
    5. `var el = fn(data, state, config.opts)` → JSElement JSON
    6. `__engine.create(el)` → DOM node → replace div contents
    7. On error: render error state
    8. On empty data (`data` is `null`, empty array, or empty object): render empty state
    9. If `config.poll > 0`: set interval to re-fetch and re-render
  - State object passed to components:
    ```js
    {
      zoneId: "cl_xxx",     // for __clients[id].setState() / .reload()
      page: 0,
      sort: { col: "", dir: "" },
      filters: {},
      search: "",
      // + any custom keys set by component via setState()
    }
    ```
  - Methods on instance:
    - `reload()` — re-fetch from source + re-render
    - `destroy()` — clean up interval, remove listeners
    - `setState(partial)` — merge into state, re-render (no re-fetch, uses cached data)
    - `refetch()` — re-fetch from source + re-render (same as reload)
  - Store instances in `window.__clients[id]`
  - File: add to `ui/ui.server.go` as `var __clientScript string`
  - Size target: ~150-200 lines of JS

### 1.4 Go Builder (`Client`)

- [ ] Create `Client` Go struct and builder in `ui/ui.client.go`
  - **Builder handles zone plumbing only:**
    ```go
    ui.Client(ctx).
        Source("/api/invoices").             // required: data URL
        Params(map[string]string{"y": "25"}).// optional: query params
        Loading(ui.SkeletonTable).           // optional: skeleton type
        Empty("inbox", "No invoices").       // optional: empty state
        Error(true).                         // optional: show errors (default true)
        Poll(30 * time.Second).              // optional: re-fetch interval
        AutoLoad(true).                      // optional: fetch on mount (default true)
        Component("table", opts).            // required: what to render
        Render()                             // emit HTML
    ```
  - `.Component(name, opts)` — sets the registered component name + JSON opts
  - `.Render()` outputs:
    1. `<div id="cl_xxx"></div>`
    2. `<script>__client({...json config...})</script>`
  - `opts` type: `ui.CZOpts` which is `map[string]any` — serialized to JSON
  - Generate unique IDs via `ui.RandomString()` or `ui.Target()`

  **Go sugar for tables (convenience, not a separate path):**
  ```go
  // This:
  ui.Client(ctx).
      Source("/api/invoices").
      Table(
          ui.CZCol("Firma").Label("Company").Sortable(true),
          ui.CZCol("Castka").Label("Amount").Type("number").Format("amount"),
      ).
      Render()

  // Is equivalent to:
  ui.Client(ctx).
      Source("/api/invoices").
      Component("table", ui.CZOpts{
          "columns": []ui.CZOpts{
              {"key": "Firma", "label": "Company", "sortable": true},
              {"key": "Castka", "label": "Amount", "type": "number", "format": "amount"},
          },
      }).
      Render()
  ```
  - `.Table(columns...)` compiles `CZColumn` structs into the opts map and calls `.Component("table", opts)` internally
  - `.Filter(true)`, `.Pagination(25)`, `.Search(true)` are also sugar — they add keys to the opts map
  - `.Chart(type)` + `.ChartOptions(opts)` similarly compile to `.Component("chart", opts)`

### 1.5 Wire into g-sui Script() Output

- [ ] Add `__cel`, `__cregister`, and `__client` JS modules to the `Script()` function in `ui.server.go`
  - Append after `__engine` module (since `__client` depends on `__engine.create()`)
  - Also append built-in component registrations (`__cregister("table", ...)`, `__cregister("chart", ...)`)
  - Ensure all framework JS loads before project component JS and boot scripts
  - Only include when `ui.Client` is used (or always include — it's small)

---

## Phase 2: Built-in Table Component

The `table` component ships with the framework and is auto-registered. It's the most common use case so it gets Go sugar (`.Table(columns...)`), but it's still just a registered component like any other.

### 2.1 Column Definition API (Go Sugar)

- [ ] Create `CZColumn` struct for defining table columns in Go:
  ```go
  type CZColumn struct {
      Key        string
      Label      string
      Class      string
      CellClass  string
      Type       string     // "text"|"number"|"date"|"bool"|"enum"|"custom"
      Format     string     // "amount"|"date"|"number" or custom
      Sortable   bool
      Filterable bool
      Render     string     // custom JS expression
      EnumOptions []CZOption
  }
  ```
  - Fluent builder:
    ```go
    ui.CZCol("Firma").Label("Company").Sortable(true).Filterable(true)
    ui.CZCol("Castka").Label("Amount").Type("number").Format("amount")
    ui.CZCol("DatVyst").Label("Date").Type("date").Sortable(true)
    ```
  - These compile to JSON objects in the opts map — the JS table component reads them

### 2.2 Table Component JS

- [ ] Create the `table` component and register it:
  ```js
  __cregister("table", function(data, state, opts) {
      // opts.columns: [{key, label, type, format, sortable, ...}, ...]
      // opts.filter: true/false
      // opts.search: true/false
      // opts.pageSize: number
      // opts.expandable: bool
      // opts.onRowClick: string (JS expression)
      // ...builds <table> using __cel, applies formatters, sort, filter, pagination
  });
  ```
  - Builds JSElement JSON for `<table>` with `<thead>` and `<tbody>`
  - Uses `__cel` for element construction
  - Column-level formatting via `__cfmt`:
    - `"date"` → `__cfmt.date(val)`
    - `"amount"` → `__cfmt.amount(val)`
    - `"bool"` → checkmark/cross
    - `"enum"` → label lookup
    - `"custom"` → `new Function("item", "i", expr)`
  - Sort: click header → cycle asc/desc/none → `__clients[state.zoneId].setState({sort: ...})` → re-render
  - Filter (if `opts.filter`): render filter bar above table, use `__cfilter` to manage state
  - Pagination (if `opts.pageSize`): render pagination below table, slice data client-side
  - Search (if `opts.search`): render search input, filter data client-side
  - All state changes via `setState()` — no re-fetch, just re-render with cached data
  - Size target: ~250-300 lines of JS (includes filter bar + pagination inline, or delegates to `__cfilter`/`__cpagination`)

### 2.3 Expandable Rows

- [ ] Add expand/accordion support to the table component
  - When `opts.expandable` is true, clicking a row inserts a detail `<tr>` below
  - Detail content from `opts.renderDetail` — a JS function string or registered component name
  - Toggle: click again to collapse, click another to switch
  - CSS transitions for smooth expand/collapse
  - Go sugar:
    ```go
    .Table(columns...).OnRowExpand(func(cz *ui.CZDetail) {
        cz.Field("item.Firma", "Company")
        cz.SubTable("/api/invoices/{item.ID}/items", subColumns...)
    })
    ```
    Compiles to `opts.expandable: true, opts.renderDetail: "function(item) {...}"`

---

## Phase 3: Built-in Chart Component

Ships with framework, auto-registered as `"chart"`.

### 3.1 Chart Component JS

- [ ] Create chart renderers and register as component:
  ```js
  __cregister("chart", function(data, state, opts) {
      // opts.type: "bar"|"area"|"hbar"|"donut"
      // opts.width, opts.height, opts.colors, opts.labels, opts.valueFormat
      var renderer = __cchart[opts.type];
      return renderer(data, opts);
  });
  ```
  - `__cchart.bar(data, opts)` → bar chart JSElement (SVG)
  - `__cchart.area(data, opts)` → area/line chart JSElement (SVG)
  - `__cchart.hbar(data, opts)` → horizontal bar chart JSElement (SVG)
  - `__cchart.donut(data, opts)` → donut/ring chart JSElement (SVG)
  - Port from existing `render.js` chart functions, output JSElement instead of innerHTML
  - Size target: ~300-400 lines of JS

### 3.2 Go Chart Sugar

- [ ] Add `.Chart(type)` and `.ChartOptions(opts)` to `Client` builder
  ```go
  ui.Client(ctx).
      Source("/api/dashboard/revenue").
      Chart(ui.BarChart).
      ChartOptions(ui.ChartOpts{
          Width: 600, Height: 300,
          Colors: []string{"#3b82f6", "#8b5cf6"},
      }).
      Render()
  ```
  Compiles to `.Component("chart", ui.CZOpts{"type": "bar", "width": 600, ...})`

---

## Phase 4: Project Components (Custom JS)

Any UI that isn't a table or chart. You write the JS, register it, reference from Go.

### 4.1 Component Authoring Pattern

- [ ] Document the standard pattern for writing project components
  - A component is a JS file that calls `__cregister`:
    ```js
    // File: server/assets/js/components/card-grid.js
    __cregister("card-grid", function(data, state, opts) {
        // data = API response (array or object)
        // state = {zoneId, page, sort, filters, search, ...}
        // opts = JSON from Go: ui.CZOpts{"cols": 3, "primary": "Firma", ...}

        var items = data;
        if (state.search) {
            items = items.filter(function(c) {
                return c[opts.primary].toLowerCase().includes(state.search.toLowerCase());
            });
        }

        return __cel.div("space-y-4", [
            opts.searchable ? __cel("input", {
                class: "input w-64",
                placeholder: "Search...",
                value: state.search || "",
            }, [], {input: {act:"raw",
                js:"__clients['"+state.zoneId+"'].setState({search:this.value})"}})
            : null,

            __cel.div("grid grid-cols-"+(opts.cols||3)+" gap-4",
                items.map(function(item) {
                    return __cel.div("card p-4 cursor-pointer", [
                        __cel.div("font-bold text-lg", [__cel.text(item[opts.primary]||"")]),
                        opts.secondary
                            ? __cel.div("text-sm text-neutral-500", [__cel.text(item[opts.secondary]||"")])
                            : null,
                    ]);
                })
            )
        ]);
    });
    ```
  - **Go side is trivial:**
    ```go
    ui.Client(ctx).
        Source("/api/companies").
        Loading(ui.SkeletonCards).
        Component("card-grid", ui.CZOpts{
            "cols":       3,
            "primary":    "Firma",
            "secondary":  "ICO",
            "searchable": true,
        }).
        Render()
    ```
  - Tools available to components:
    - `__cel` — build JSElement trees
    - `__cfmt` — format dates, amounts, numbers, booleans
    - `__capi` — fetch additional data (for sub-detail expansion, etc.)
    - `__clients[state.zoneId].setState({...})` — trigger re-render with new state
    - `__clients[state.zoneId].reload()` — re-fetch data from API

### 4.2 Component Loading Order

- [ ] Ensure correct script loading order
  ```
  1. Framework JS (embedded in Go):  __cel, __cregister, __cfmt, __capi, __client
  2. Framework components (embedded): __cregister("table", ...), __cregister("chart", ...)
  3. Project JS files (static assets): components/*.js with __cregister() calls
  4. Boot scripts (inline in HTML):    <script>__client({...})</script>
  ```
  - Framework JS is emitted by `Script()` in `<head>` or before `</body>`
  - Project component files loaded via `<script src="/js/components/card-grid.js">`
  - Boot scripts are inline, emitted where `Client.Render()` is called

### 4.3 Example Project Components

- [ ] Create reference implementations (project-level, not framework):

  **`card-grid.js`** — Responsive card grid with search (~80 lines)
  **`kpi-bar.js`** — Horizontal stat tiles (~50 lines)
  **`detail-view.js`** — Read-only field:value layout (~60 lines)
  **`list.js`** — Styled vertical list with icons (~70 lines)

  These are starting points, not framework code. Each project copies/modifies as needed.

---

## Phase 5: Filter & Pagination (Shared Infrastructure)

These modules are used by the built-in `table` component but are also available to any project component.

### 5.1 Client Filter State (`__cfilter`)

- [ ] Create `__cfilter` JS module — client-side filter state manager
  - State: `{ search: "", filters: { col: { op, value, values, from, to } }, sort: { col, dir } }`
  - URL sync: read/write filter state to URL query params
  - localStorage persistence (optional)
  - Methods: `setFilter`, `removeFilter`, `setSort`, `setSearch`, `setPage`, `applyToData`, `toQueryString`
  - Any component can use this — not table-specific
  - Size target: ~200-250 lines of JS

### 5.2 Filter Bar Renderer (`__cfilterbar`)

- [ ] Create `__cfilterbar(zoneId, filterState, columns)` JS function
  - Renders: search input, active filter chips, clear all button
  - Filter dropdowns: text (contains/equals), number (range), date (range), bool (toggle), enum (multi-select)
  - Returns JSElement JSON
  - Size target: ~300 lines of JS

### 5.3 Pagination Renderer (`__cpagination`)

- [ ] Create `__cpagination(zoneId, page, totalPages, pageSize)` JS function
  - Prev/Next, page numbers with ellipsis, page size selector, record count
  - Returns JSElement JSON
  - Size target: ~80-100 lines of JS

---

## Phase 6: Utility Modules

### 6.1 Client-Side Formatters (`__cfmt`)

- [ ] Create `__cfmt` JS module
  - `__cfmt.date(val)`, `__cfmt.amount(val)`, `__cfmt.number(val)`, `__cfmt.bool(val)`, `__cfmt.truncate(val, max)`, `__cfmt.escape(val)`
  - Locale from `window.__locale` (set by server)
  - Available to all components (built-in and project)
  - Size target: ~60-80 lines of JS

### 6.2 Client-Side API Helper (`__capi`)

- [ ] Create `__capi` JS module
  - `__capi.get(url, params)`, `__capi.post(url, body)`
  - Error handling, timeout, request deduplication
  - Used by `__client` for data fetching + available to components for secondary fetches
  - Size target: ~40-60 lines of JS

---

## Phase 7: Integration

### 7.1 Server Action Bridge

- [ ] Enable server actions from within client-rendered content
  - `__caction(handlerPath, payload)` — triggers `__post()` roundtrip
  - Use case: Delete button in a table row, form submit from a detail view
  - Works with any component

### 7.2 WebSocket Patch Integration

- [ ] Ensure `ctx.Patch()` can target elements inside client-rendered zones
  - Elements created via `__engine.create()` have proper IDs
  - Existing `__engine.applyPatch()` already finds by ID — no special handling needed

### 7.3 Documentation

- [ ] Add `CLIENT.md` to g-sui skill documentation
  - `Client` API reference
  - Table/chart Go sugar reference
  - Component authoring guide with full examples
  - JS module reference: `__cel`, `__cregister`, `__client`, `__cfmt`, `__capi`

---

## Phase 8: Aximo Migration (Project-Specific)

### 8.1 Table Pages (Go sugar: `.Table(columns...)`)

- [ ] Migrate `addressesPage` — **first migration, validate API**
- [ ] Migrate `invoicesPage`
- [ ] Migrate `bankPage`
- [ ] Migrate `cashPage`, `journalPage`
- [ ] Migrate `chartOfAccountsPage` (grouped tables)
- [ ] Migrate remaining: `internalDocsPage`, `ordersPage`, `contractsPage`, `postingRulesPage`, `numberSeriesPage`, `generalLedgerPage`, `employeesPage`, `payrollPage`, `assetsPage`, `smallAssetsPage`, `paymentsPage`

### 8.2 Multi-Client Pages (multiple `ui.Client` calls on one page)

- [ ] Migrate `receivablesPage` — `Component("kpi-bar")` + `Table(...)` with expandable rows
- [ ] Migrate `payablesPage` — same pattern
- [ ] Migrate `costRevenuePage` — `Component("kpi-bar")` + two `Table(...)` instances
- [ ] Migrate `dashboardPage` — `Component("kpi-bar")` + `Chart(...)` + `Table(...)`

### 8.3 Custom Component Pages (write JS, `.Component(name, opts)`)

- [ ] Write + register `"card-grid"` component → migrate `companiesPage`
- [ ] Write + register `"job-list"` component → migrate `jobsPage` (with `Poll()`)
- [ ] Write + register `"import-manager"` component → migrate `importPage`

### 8.4 Remove Legacy JS

- [ ] Remove `server/assets/js/render.js` (replaced by `table` + `chart` components)
- [ ] Remove `server/assets/js/filter.js` (replaced by `__cfilter`)
- [ ] Remove `server/assets/js/api.js` (replaced by `__capi`)

---

## Implementation Notes

### File Locations

**Framework (g-sui package):**

| File | Purpose |
|---|---|
| `ui/ui.server.go` — new vars | `__cel`, `__cregister`, `__client`, `__cfmt`, `__capi`, `__ctable`, `__cchart`, `__cfilter`, `__cfilterbar`, `__cpagination` JS modules |
| `ui/ui.client.go` (new) | `Client` builder, `CZColumn`, `CZOpts` types |
| `ui/ui.client_test.go` (new) | Tests for Go builder JSON output |
| `.claude/skills/g-sui/CLIENT.md` (new) | Skill documentation |

**Project-level (e.g. Aximo):**

| File | Purpose |
|---|---|
| `server/assets/js/components/*.js` | Project-specific registered components |

### JS Load Order

```
── framework (embedded in Go, emitted by Script()) ──
__cfmt                                    (standalone)
__capi                                    (standalone)
__cel                                     (standalone)
__cregister                               (standalone, creates window.__cregistry)
__cfilter                                 (uses __cel)
__cfilterbar                              (uses __cel, __cfilter)
__cpagination                             (uses __cel)
__cregister("table", fn)                  (uses __cel, __cfmt, __cfilter, __cpagination)
__cregister("chart", fn)                  (uses __cel, __cfmt)
__client                                  (uses __engine, __cregister, __capi)
── project JS (static assets, <script src="...">) ──
components/card-grid.js                   (uses __cregister, __cel, __cfmt)
components/kpi-bar.js                     (uses __cregister, __cel, __cfmt)
components/...                            (any project components)
── boot scripts (inline in HTML, per ui.Client call) ──
<script>__client({id:"cl_1", source:"/api/...", component:"table", opts:{...}})</script>
<script>__client({id:"cl_2", source:"/api/...", component:"kpi-bar", opts:{...}})</script>
```

### Design Constraints

1. **No virtual DOM / reactivity** — One-shot render. State changes → full re-render via `__engine.create()`.
2. **No build step** — Framework JS embedded in Go strings. Project JS served as static files.
3. **No new dependencies** — Pure vanilla JS + existing `__engine.create()`.
4. **Locale from server** — `window.__locale` set by `ui.MakeApp("sk")`.
5. **Dark mode** — All components respect existing theme (`dark:` Tailwind prefixes).
6. **Framework JS budget** — Under 2KB gzipped, ~1500 lines total. Project component JS is outside this budget.
7. **One render path** — Every component goes through `__cregister` / `__cget`. No special codepath for tables vs custom. Tables are just a component that has Go sugar.
8. **Go is plumbing, JS is rendering** — Go builder never contains HTML/CSS. It emits JSON config. JS components own all visual decisions.
