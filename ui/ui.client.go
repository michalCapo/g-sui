package ui

import (
	"encoding/json"
	"fmt"
	"time"
)

// ClientOpts is a map of options passed to client-side components as JSON.
type ClientOpts map[string]any

// Additional skeleton types for client zones.
const (
	SkeletonTable Skeleton = "table"
	SkeletonCards Skeleton = "cards"
)

// ChartType identifies the kind of chart for the chart component sugar.
type ChartType string

const (
	BarChart   ChartType = "bar"
	AreaChart  ChartType = "area"
	HBarChart  ChartType = "hbar"
	DonutChart ChartType = "donut"
)

// ---------------------------------------------------------------------------
// ClientColumn – fluent builder for column definitions
// ---------------------------------------------------------------------------

// ClientColumn describes a single column in a client-zone table.
type ClientColumn struct {
	key         string
	label       string
	class       string
	cellClass   string
	colType     string // "text"|"number"|"date"|"bool"|"enum"|"custom"
	format      string // "amount"|"date"|"number" or custom
	sortable    bool
	filterable  bool
	render      string // custom JS expression
	enumOptions []ClientOption
}

// ClientOption is a value/label pair used by enum-type columns.
type ClientOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ClientCol starts building a new column with the given data key.
func ClientCol(key string) *ClientColumn { return &ClientColumn{key: key} }

func (c *ClientColumn) Label(l string) *ClientColumn      { c.label = l; return c }
func (c *ClientColumn) Class(cl string) *ClientColumn     { c.class = cl; return c }
func (c *ClientColumn) CellClass(cl string) *ClientColumn { c.cellClass = cl; return c }
func (c *ClientColumn) Type(t string) *ClientColumn       { c.colType = t; return c }
func (c *ClientColumn) Format(f string) *ClientColumn     { c.format = f; return c }
func (c *ClientColumn) Sortable(s bool) *ClientColumn     { c.sortable = s; return c }
func (c *ClientColumn) Filterable(f bool) *ClientColumn   { c.filterable = f; return c }
func (c *ClientColumn) Render(r string) *ClientColumn     { c.render = r; return c }
func (c *ClientColumn) EnumOptions(opts ...ClientOption) *ClientColumn {
	c.enumOptions = opts
	return c
}

// toMap converts the column definition to a JSON-friendly map.
func (c *ClientColumn) toMap() map[string]any {
	m := map[string]any{"key": c.key}
	if c.label != "" {
		m["label"] = c.label
	}
	if c.class != "" {
		m["class"] = c.class
	}
	if c.cellClass != "" {
		m["cellClass"] = c.cellClass
	}
	if c.colType != "" {
		m["type"] = c.colType
	}
	if c.format != "" {
		m["format"] = c.format
	}
	if c.sortable {
		m["sortable"] = true
	}
	if c.filterable {
		m["filterable"] = true
	}
	if c.render != "" {
		m["render"] = c.render
	}
	if len(c.enumOptions) > 0 {
		m["enumOptions"] = c.enumOptions
	}
	return m
}

// ---------------------------------------------------------------------------
// ClientBuilder – fluent builder for client-side rendered zones
// ---------------------------------------------------------------------------

// ClientBuilder constructs a client-side rendered zone that boots via __client().
type ClientBuilder struct {
	ctx       *Context
	id        string
	source    string
	params    map[string]string
	component string
	opts      ClientOpts
	loading   Skeleton
	emptyIcon string
	emptyMsg  string
	showError bool
	autoLoad  bool
	poll      int // milliseconds
}

// Client creates a new ClientBuilder bound to the given request context.
func Client(ctx *Context) *ClientBuilder {
	return &ClientBuilder{
		ctx:       ctx,
		id:        "cl_" + RandomString(12),
		showError: true,
		autoLoad:  true,
	}
}

func (b *ClientBuilder) Source(url string) *ClientBuilder          { b.source = url; return b }
func (b *ClientBuilder) Params(p map[string]string) *ClientBuilder { b.params = p; return b }
func (b *ClientBuilder) Loading(s Skeleton) *ClientBuilder         { b.loading = s; return b }
func (b *ClientBuilder) Error(show bool) *ClientBuilder            { b.showError = show; return b }
func (b *ClientBuilder) AutoLoad(auto bool) *ClientBuilder         { b.autoLoad = auto; return b }

// Empty sets the icon and message shown when the data source returns no rows.
func (b *ClientBuilder) Empty(icon, message string) *ClientBuilder {
	b.emptyIcon = icon
	b.emptyMsg = message
	return b
}

// Poll sets the automatic refresh interval.
func (b *ClientBuilder) Poll(d time.Duration) *ClientBuilder {
	b.poll = int(d.Milliseconds())
	return b
}

// Component sets the registered JS component name and its options.
func (b *ClientBuilder) Component(name string, opts ClientOpts) *ClientBuilder {
	b.component = name
	b.opts = opts
	return b
}

// Table is sugar for .Component("table", …) with pre-compiled column definitions.
func (b *ClientBuilder) Table(columns ...*ClientColumn) *ClientBuilder {
	if b.opts == nil {
		b.opts = ClientOpts{}
	}
	cols := make([]map[string]any, len(columns))
	for i, col := range columns {
		cols[i] = col.toMap()
	}
	b.opts["columns"] = cols
	b.component = "table"
	return b
}

// Filter enables or disables the filter bar on the table component.
func (b *ClientBuilder) Filter(enabled bool) *ClientBuilder {
	if b.opts == nil {
		b.opts = ClientOpts{}
	}
	b.opts["filter"] = enabled
	return b
}

// Pagination sets the page size for the table component.
func (b *ClientBuilder) Pagination(pageSize int) *ClientBuilder {
	if b.opts == nil {
		b.opts = ClientOpts{}
	}
	b.opts["pageSize"] = pageSize
	return b
}

// Search enables the search input on the table component.
func (b *ClientBuilder) Search(enabled bool) *ClientBuilder {
	if b.opts == nil {
		b.opts = ClientOpts{}
	}
	b.opts["search"] = enabled
	return b
}

// Chart is sugar for .Component("chart", …) with the given chart type.
func (b *ClientBuilder) Chart(chartType ChartType) *ClientBuilder {
	if b.opts == nil {
		b.opts = ClientOpts{}
	}
	b.opts["type"] = string(chartType)
	b.component = "chart"
	return b
}

// ChartOptions merges additional chart options into the builder.
func (b *ClientBuilder) ChartOptions(opts ClientOpts) *ClientBuilder {
	if b.opts == nil {
		b.opts = ClientOpts{}
	}
	for k, v := range opts {
		b.opts[k] = v
	}
	return b
}

// Render outputs the HTML: a target div + a script tag that boots __client().
func (b *ClientBuilder) Render() string {
	config := map[string]any{
		"id":     b.id,
		"source": b.source,
	}
	if len(b.params) > 0 {
		config["params"] = b.params
	}
	if b.component != "" {
		config["component"] = b.component
	}
	if b.opts != nil && len(b.opts) > 0 {
		config["opts"] = b.opts
	}
	if b.loading != "" {
		config["loading"] = string(b.loading)
	}
	if b.emptyIcon != "" || b.emptyMsg != "" {
		empty := map[string]string{}
		if b.emptyIcon != "" {
			empty["icon"] = b.emptyIcon
		}
		if b.emptyMsg != "" {
			empty["message"] = b.emptyMsg
		}
		config["empty"] = empty
	}
	if !b.showError {
		config["error"] = false
	}
	if !b.autoLoad {
		config["autoLoad"] = false
	}
	if b.poll > 0 {
		config["poll"] = b.poll
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Sprintf(`<div id="%s"><!-- client config error: %s --></div>`, b.id, err.Error())
	}

	return fmt.Sprintf(`<div id="%s"></div><script>__client(%s)</script>`, b.id, string(jsonBytes))
}

// ---------------------------------------------------------------------------
// Skeleton renderers for SkeletonTable and SkeletonCards
// ---------------------------------------------------------------------------

// SkeletonTableBlock renders a table-shaped skeleton with default attributes.
func SkeletonTableBlock() string { return Attr{}.SkeletonTable() }

// SkeletonTable renders a table-shaped skeleton placeholder.
func (a Attr) SkeletonTable() string {
	headerCells := ""
	for i := 0; i < 4; i++ {
		headerCells += `<th class="p-3"><div class="bg-gray-200 dark:bg-gray-700 h-4 rounded w-20"></div></th>`
	}
	rows := ""
	for r := 0; r < 5; r++ {
		cells := ""
		for c := 0; c < 4; c++ {
			w := "w-24"
			if c == 0 {
				w = "w-32"
			}
			if c == 2 {
				w = "w-16"
			}
			cells += fmt.Sprintf(`<td class="p-3"><div class="bg-gray-200 dark:bg-gray-700 h-4 rounded %s"></div></td>`, w)
		}
		rows += `<tr class="border-t border-gray-100 dark:border-gray-800">` + cells + `</tr>`
	}
	return Div("animate-pulse", a)(
		`<div class="bg-white dark:bg-gray-900 rounded-lg shadow overflow-hidden">` +
			`<table class="w-full"><thead><tr class="border-b border-gray-200 dark:border-gray-700">` +
			headerCells + `</tr></thead><tbody>` + rows + `</tbody></table></div>`,
	)
}

// SkeletonCardsBlock renders a card grid skeleton with default attributes.
func SkeletonCardsBlock() string { return Attr{}.SkeletonCards() }

// SkeletonCards renders a card grid skeleton placeholder.
func (a Attr) SkeletonCards() string {
	cards := ""
	for i := 0; i < 6; i++ {
		cards += `<div class="bg-white dark:bg-gray-900 rounded-lg p-4 shadow">` +
			`<div class="bg-gray-200 dark:bg-gray-700 h-5 rounded w-3/4 mb-3"></div>` +
			`<div class="bg-gray-200 dark:bg-gray-700 h-4 rounded w-1/2 mb-2"></div>` +
			`<div class="bg-gray-200 dark:bg-gray-700 h-4 rounded w-2/3"></div></div>`
	}
	return Div("animate-pulse", a)(
		`<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">` + cards + `</div>`,
	)
}
