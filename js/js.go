// Package js holds client-side rendered component builders.
// Types in this package describe zones that will be converted to JavaScript
// on the client (tables, charts, custom components fetched from JSON APIs).
package js

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// Opts is a map of options passed to client-side components as JSON.
type Opts map[string]any

// ChartType identifies the kind of chart for the chart component sugar.
type ChartType string

const (
	BarChart   ChartType = "bar"
	AreaChart  ChartType = "area"
	HBarChart  ChartType = "hbar"
	DonutChart ChartType = "donut"
)

// ---------------------------------------------------------------------------
// Column – fluent builder for column definitions
// ---------------------------------------------------------------------------

// Column describes a single column in a client-zone table.
type Column struct {
	key         string
	label       string
	class       string
	cellClass   string
	colType     string // "text"|"number"|"date"|"bool"|"enum"|"custom"
	format      string // "amount"|"date"|"number" or custom
	sortable    bool
	filterable  bool
	render      string // custom JS expression
	enumOptions []Option
}

// Option is a value/label pair used by enum-type columns.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Col starts building a new column with the given data key.
func Col(key string) *Column { return &Column{key: key} }

func (c *Column) Label(l string) *Column      { c.label = l; return c }
func (c *Column) Class(cl string) *Column     { c.class = cl; return c }
func (c *Column) CellClass(cl string) *Column { c.cellClass = cl; return c }
func (c *Column) Type(t string) *Column       { c.colType = t; return c }
func (c *Column) Format(f string) *Column     { c.format = f; return c }
func (c *Column) Sortable(s bool) *Column     { c.sortable = s; return c }
func (c *Column) Filterable(f bool) *Column   { c.filterable = f; return c }
func (c *Column) Render(r string) *Column     { c.render = r; return c }
func (c *Column) EnumOptions(opts ...Option) *Column {
	c.enumOptions = opts
	return c
}

// ToMap converts the column definition to a JSON-friendly map.
func (c *Column) ToMap() map[string]any {
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
// Builder – fluent builder for client-side rendered zones
// ---------------------------------------------------------------------------

// Builder constructs a client-side rendered zone that boots via __client().
type Builder struct {
	ctx       *ui.Context
	id        string
	source    string
	params    map[string]string
	component string
	opts      Opts
	loading   ui.Skeleton
	emptyIcon string
	emptyMsg  string
	showError bool
	autoLoad  bool
	poll      int // milliseconds
}

// Client creates a new Builder bound to the given request context.
func Client(ctx *ui.Context) *Builder {
	return &Builder{
		ctx:       ctx,
		id:        "cl_" + ui.RandomString(12),
		showError: true,
		autoLoad:  true,
	}
}

func (b *Builder) Source(url string) *Builder          { b.source = url; return b }
func (b *Builder) Params(p map[string]string) *Builder { b.params = p; return b }
func (b *Builder) Loading(s ui.Skeleton) *Builder      { b.loading = s; return b }
func (b *Builder) Error(show bool) *Builder            { b.showError = show; return b }
func (b *Builder) AutoLoad(auto bool) *Builder         { b.autoLoad = auto; return b }

// Empty sets the icon and message shown when the data source returns no rows.
func (b *Builder) Empty(icon, message string) *Builder {
	b.emptyIcon = icon
	b.emptyMsg = message
	return b
}

// Poll sets the automatic refresh interval.
func (b *Builder) Poll(d time.Duration) *Builder {
	b.poll = int(d.Milliseconds())
	return b
}

// Component sets the registered JS component name and its options.
func (b *Builder) Component(name string, opts Opts) *Builder {
	b.component = name
	b.opts = opts
	return b
}

// Table is sugar for .Component("table", …) with pre-compiled column definitions.
func (b *Builder) Table(columns ...*Column) *Builder {
	if b.opts == nil {
		b.opts = Opts{}
	}
	cols := make([]map[string]any, len(columns))
	for i, col := range columns {
		cols[i] = col.ToMap()
	}
	b.opts["columns"] = cols
	b.component = "table"
	return b
}

// Filter enables or disables the filter bar on the table component.
func (b *Builder) Filter(enabled bool) *Builder {
	if b.opts == nil {
		b.opts = Opts{}
	}
	b.opts["filter"] = enabled
	return b
}

// Pagination sets the page size for the table component.
func (b *Builder) Pagination(pageSize int) *Builder {
	if b.opts == nil {
		b.opts = Opts{}
	}
	b.opts["pageSize"] = pageSize
	return b
}

// Search enables the search input on the table component.
func (b *Builder) Search(enabled bool) *Builder {
	if b.opts == nil {
		b.opts = Opts{}
	}
	b.opts["search"] = enabled
	return b
}

// Chart is sugar for .Component("chart", …) with the given chart type.
func (b *Builder) Chart(chartType ChartType) *Builder {
	if b.opts == nil {
		b.opts = Opts{}
	}
	b.opts["type"] = string(chartType)
	b.component = "chart"
	return b
}

// ChartOptions merges additional chart options into the builder.
func (b *Builder) ChartOptions(opts Opts) *Builder {
	if b.opts == nil {
		b.opts = Opts{}
	}
	for k, v := range opts {
		b.opts[k] = v
	}
	return b
}

// Render outputs the HTML: a target div + a script tag that boots __client().
func (b *Builder) Render() string {
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
