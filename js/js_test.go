package js

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// parseClientJSON extracts the JSON config object from the Render() output.
// It looks for the __client({...}) call and unmarshals the JSON payload.
func parseClientJSON(t *testing.T, html string) map[string]any {
	t.Helper()
	const prefix = `<script>__client(`
	const suffix = `)</script>`
	i := strings.Index(html, prefix)
	if i < 0 {
		t.Fatalf("expected __client() script block, got: %s", html)
	}
	j := strings.Index(html[i:], suffix)
	if j < 0 {
		t.Fatalf("expected closing </script>, got: %s", html)
	}
	payload := html[i+len(prefix) : i+j]
	var m map[string]any
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		t.Fatalf("invalid JSON in __client(): %v\npayload: %s", err, payload)
	}
	return m
}

func assertContains(t *testing.T, html, expected string) {
	t.Helper()
	if !strings.Contains(html, expected) {
		t.Errorf("Expected to find %q in HTML, but it was not found.\nHTML: %s", expected, html)
	}
}

// ============================================================================
// Builder.Render() – output format
// ============================================================================

func TestBuilder_RenderFormat(t *testing.T) {
	html := Client(nil).
		Source("/api/invoices").
		Component("table", Opts{"columns": []any{}}).
		Render()

	assertContains(t, html, `<div id="cl_`)
	assertContains(t, html, `"></div>`)
	assertContains(t, html, `<script>__client(`)
	assertContains(t, html, `)</script>`)
}

func TestBuilder_RenderJSON(t *testing.T) {
	html := Client(nil).
		Source("/api/invoices").
		Component("table", Opts{"columns": []any{}}).
		Render()

	cfg := parseClientJSON(t, html)

	if src, ok := cfg["source"].(string); !ok || src != "/api/invoices" {
		t.Errorf("expected source=/api/invoices, got %v", cfg["source"])
	}
	if comp, ok := cfg["component"].(string); !ok || comp != "table" {
		t.Errorf("expected component=table, got %v", cfg["component"])
	}
	// id must start with cl_
	if id, ok := cfg["id"].(string); !ok || !strings.HasPrefix(id, "cl_") {
		t.Errorf("expected id starting with cl_, got %v", cfg["id"])
	}
}

// ============================================================================
// Defaults
// ============================================================================

func TestBuilder_Defaults(t *testing.T) {
	html := Client(nil).
		Source("/api/data").
		Component("widget", nil).
		Render()

	cfg := parseClientJSON(t, html)

	// showError defaults to true → should NOT appear in config
	if _, ok := cfg["error"]; ok {
		t.Error("error field should be omitted when showError is true (default)")
	}

	// autoLoad defaults to true → should NOT appear in config
	if _, ok := cfg["autoLoad"]; ok {
		t.Error("autoLoad field should be omitted when autoLoad is true (default)")
	}

	// id starts with cl_
	id := cfg["id"].(string)
	if !strings.HasPrefix(id, "cl_") {
		t.Errorf("expected id to start with cl_, got %q", id)
	}

	// nil opts → opts key should be absent
	if _, ok := cfg["opts"]; ok {
		t.Error("opts should be omitted when nil")
	}
}

func TestBuilder_UniqueIDs(t *testing.T) {
	a := Client(nil).Source("/a").Component("x", nil).Render()
	b := Client(nil).Source("/b").Component("x", nil).Render()

	cfgA := parseClientJSON(t, a)
	cfgB := parseClientJSON(t, b)

	if cfgA["id"] == cfgB["id"] {
		t.Errorf("expected unique IDs, both got %v", cfgA["id"])
	}
}

// ============================================================================
// Builder methods – each sets the correct JSON field
// ============================================================================

func TestBuilder_Source(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).Source("/api/users").Component("list", nil).Render())
	if cfg["source"] != "/api/users" {
		t.Errorf("source = %v, want /api/users", cfg["source"])
	}
}

func TestBuilder_Params(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Params(map[string]string{"status": "active", "page": "1"}).
		Component("list", nil).
		Render(),
	)

	params, ok := cfg["params"].(map[string]any)
	if !ok {
		t.Fatalf("params missing or wrong type: %v", cfg["params"])
	}
	if params["status"] != "active" {
		t.Errorf("params.status = %v, want active", params["status"])
	}
	if params["page"] != "1" {
		t.Errorf("params.page = %v, want 1", params["page"])
	}
}

func TestBuilder_Loading(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("list", nil).
		Loading(ui.SkeletonTable).
		Render(),
	)
	if cfg["loading"] != "table" {
		t.Errorf("loading = %v, want table", cfg["loading"])
	}
}

func TestBuilder_ErrorDisabled(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("list", nil).
		Error(false).
		Render(),
	)
	if cfg["error"] != false {
		t.Errorf("error = %v, want false", cfg["error"])
	}
}

func TestBuilder_AutoLoadDisabled(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("list", nil).
		AutoLoad(false).
		Render(),
	)
	if cfg["autoLoad"] != false {
		t.Errorf("autoLoad = %v, want false", cfg["autoLoad"])
	}
}

func TestBuilder_Poll(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("list", nil).
		Poll(5*time.Second).
		Render(),
	)
	// JSON numbers unmarshal as float64
	if cfg["poll"] != float64(5000) {
		t.Errorf("poll = %v, want 5000", cfg["poll"])
	}
}

func TestBuilder_EmptyState(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/data").
		Component("test", nil).
		Empty("inbox", "No data found").
		Render(),
	)

	empty, ok := cfg["empty"].(map[string]any)
	if !ok {
		t.Fatalf("empty missing or wrong type: %v", cfg["empty"])
	}
	if empty["icon"] != "inbox" {
		t.Errorf("empty.icon = %v, want inbox", empty["icon"])
	}
	if empty["message"] != "No data found" {
		t.Errorf("empty.message = %v, want No data found", empty["message"])
	}
}

func TestBuilder_EmptyIconOnly(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/data").
		Component("test", nil).
		Empty("alert", "").
		Render(),
	)

	empty, ok := cfg["empty"].(map[string]any)
	if !ok {
		t.Fatalf("empty missing or wrong type: %v", cfg["empty"])
	}
	if empty["icon"] != "alert" {
		t.Errorf("empty.icon = %v, want alert", empty["icon"])
	}
	if _, hasMsg := empty["message"]; hasMsg {
		t.Error("empty.message should be omitted when empty string")
	}
}

func TestBuilder_EmptyMessageOnly(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/data").
		Component("test", nil).
		Empty("", "Nothing here").
		Render(),
	)

	empty, ok := cfg["empty"].(map[string]any)
	if !ok {
		t.Fatalf("empty missing or wrong type: %v", cfg["empty"])
	}
	if _, hasIcon := empty["icon"]; hasIcon {
		t.Error("empty.icon should be omitted when empty string")
	}
	if empty["message"] != "Nothing here" {
		t.Errorf("empty.message = %v, want Nothing here", empty["message"])
	}
}

// ============================================================================
// Component + opts
// ============================================================================

func TestBuilder_ComponentWithOpts(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("widget", Opts{"color": "blue", "size": 42}).
		Render(),
	)

	if cfg["component"] != "widget" {
		t.Errorf("component = %v, want widget", cfg["component"])
	}
	opts, ok := cfg["opts"].(map[string]any)
	if !ok {
		t.Fatalf("opts missing or wrong type: %v", cfg["opts"])
	}
	if opts["color"] != "blue" {
		t.Errorf("opts.color = %v, want blue", opts["color"])
	}
	if opts["size"] != float64(42) {
		t.Errorf("opts.size = %v, want 42", opts["size"])
	}
}

func TestBuilder_ComponentNoOpts(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("bare", nil).
		Render(),
	)
	if cfg["component"] != "bare" {
		t.Errorf("component = %v, want bare", cfg["component"])
	}
	if _, ok := cfg["opts"]; ok {
		t.Error("opts should be omitted when nil")
	}
}

// ============================================================================
// Col builder
// ============================================================================

func TestCol_ToMap(t *testing.T) {
	tests := []struct {
		name   string
		col    *Column
		expect map[string]any
	}{
		{
			name:   "key only",
			col:    Col("id"),
			expect: map[string]any{"key": "id"},
		},
		{
			name: "all fields",
			col: Col("amount").
				Label("Amount").
				Class("text-right").
				CellClass("font-mono").
				Type("number").
				Format("amount").
				Sortable(true).
				Filterable(true).
				Render("row => row.amount.toFixed(2)"),
			expect: map[string]any{
				"key":        "amount",
				"label":      "Amount",
				"class":      "text-right",
				"cellClass":  "font-mono",
				"type":       "number",
				"format":     "amount",
				"sortable":   true,
				"filterable": true,
				"render":     "row => row.amount.toFixed(2)",
			},
		},
		{
			name: "with enum options",
			col: Col("status").
				Type("enum").
				EnumOptions(
					Option{Value: "active", Label: "Active"},
					Option{Value: "inactive", Label: "Inactive"},
				),
			expect: map[string]any{
				"key":  "status",
				"type": "enum",
				"enumOptions": []Option{
					{Value: "active", Label: "Active"},
					{Value: "inactive", Label: "Inactive"},
				},
			},
		},
		{
			name: "false booleans omitted",
			col:  Col("name").Label("Name").Sortable(false).Filterable(false),
			expect: map[string]any{
				"key":   "name",
				"label": "Name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.col.ToMap()

			for k, want := range tt.expect {
				got, ok := m[k]
				if !ok {
					t.Errorf("missing key %q in column map", k)
					continue
				}
				// Special handling for enumOptions (slice comparison)
				if k == "enumOptions" {
					gotJSON, _ := json.Marshal(got)
					wantJSON, _ := json.Marshal(want)
					if string(gotJSON) != string(wantJSON) {
						t.Errorf("key %q = %s, want %s", k, gotJSON, wantJSON)
					}
					continue
				}
				if got != want {
					t.Errorf("key %q = %v, want %v", k, got, want)
				}
			}

			// Ensure no extra keys beyond expected
			for k := range m {
				if _, ok := tt.expect[k]; !ok {
					t.Errorf("unexpected key %q in column map (value: %v)", k, m[k])
				}
			}
		})
	}
}

// ============================================================================
// HTMLMap helper
// ============================================================================

func TestHTMLMap(t *testing.T) {
	expr := HTMLMap("status", map[string]string{
		"paid":  `<span class="badge-green">Paid</span>`,
		"draft": `<span class="badge-gray">Draft</span>`,
	})

	// Must reference the correct key
	if !strings.Contains(expr, `item["status"]`) {
		t.Errorf("expected item[\"status\"] reference, got: %s", expr)
	}

	// Must contain the HTML for each value
	if !strings.Contains(expr, "badge-green") {
		t.Error("expected badge-green HTML in output")
	}
	if !strings.Contains(expr, "badge-gray") {
		t.Error("expected badge-gray HTML in output")
	}

	// Must be valid JS body (contains return statement)
	if !strings.Contains(expr, "return") {
		t.Error("expected return statement in JS expression")
	}
}

// ============================================================================
// Table sugar
// ============================================================================

func TestBuilder_TableSugar(t *testing.T) {
	html := Client(nil).
		Source("/api/test").
		Table(
			Col("name").Label("Name").Sortable(true),
			Col("age").Label("Age").Type("number"),
		).
		Render()

	cfg := parseClientJSON(t, html)

	if cfg["component"] != "table" {
		t.Errorf("component = %v, want table", cfg["component"])
	}

	opts, ok := cfg["opts"].(map[string]any)
	if !ok {
		t.Fatalf("opts missing or wrong type: %v", cfg["opts"])
	}

	cols, ok := opts["columns"].([]any)
	if !ok {
		t.Fatalf("columns missing or wrong type: %v", opts["columns"])
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}

	col0 := cols[0].(map[string]any)
	if col0["key"] != "name" {
		t.Errorf("col[0].key = %v, want name", col0["key"])
	}
	if col0["label"] != "Name" {
		t.Errorf("col[0].label = %v, want Name", col0["label"])
	}
	if col0["sortable"] != true {
		t.Errorf("col[0].sortable = %v, want true", col0["sortable"])
	}

	col1 := cols[1].(map[string]any)
	if col1["key"] != "age" {
		t.Errorf("col[1].key = %v, want age", col1["key"])
	}
	if col1["type"] != "number" {
		t.Errorf("col[1].type = %v, want number", col1["type"])
	}
}

func TestBuilder_TableSugarEquivalence(t *testing.T) {
	sugar := Client(nil).Source("/api/test").
		Table(Col("A").Label("AA").Sortable(true)).
		Render()

	explicit := Client(nil).Source("/api/test").
		Component("table", Opts{
			"columns": []map[string]any{
				{"key": "A", "label": "AA", "sortable": true},
			},
		}).
		Render()

	sugarCfg := parseClientJSON(t, sugar)
	explCfg := parseClientJSON(t, explicit)

	// Both set component=table
	if sugarCfg["component"] != "table" || explCfg["component"] != "table" {
		t.Error("both should have component=table")
	}

	// Both should have the same column structure
	sugarOpts := sugarCfg["opts"].(map[string]any)
	explOpts := explCfg["opts"].(map[string]any)

	sugarColsJSON, _ := json.Marshal(sugarOpts["columns"])
	explColsJSON, _ := json.Marshal(explOpts["columns"])

	if string(sugarColsJSON) != string(explColsJSON) {
		t.Errorf("column JSON mismatch:\n  sugar:    %s\n  explicit: %s", sugarColsJSON, explColsJSON)
	}
}

// ============================================================================
// Filter, Pagination, Search sugar
// ============================================================================

func TestBuilder_FilterSugar(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Table(Col("a")).
		Filter(true).
		Render(),
	)

	opts := cfg["opts"].(map[string]any)
	if opts["filter"] != true {
		t.Errorf("filter = %v, want true", opts["filter"])
	}
}

func TestBuilder_PaginationSugar(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Table(Col("a")).
		Pagination(25).
		Render(),
	)

	opts := cfg["opts"].(map[string]any)
	if opts["pageSize"] != float64(25) {
		t.Errorf("pageSize = %v, want 25", opts["pageSize"])
	}
}

func TestBuilder_SearchSugar(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Table(Col("a")).
		Search(true).
		Render(),
	)

	opts := cfg["opts"].(map[string]any)
	if opts["search"] != true {
		t.Errorf("search = %v, want true", opts["search"])
	}
}

func TestBuilder_FilterOnBareComponent(t *testing.T) {
	// Filter/Pagination/Search should work even without Table() sugar
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("custom", nil).
		Filter(true).
		Pagination(10).
		Search(true).
		Render(),
	)

	opts := cfg["opts"].(map[string]any)
	if opts["filter"] != true {
		t.Errorf("filter = %v, want true", opts["filter"])
	}
	if opts["pageSize"] != float64(10) {
		t.Errorf("pageSize = %v, want 10", opts["pageSize"])
	}
	if opts["search"] != true {
		t.Errorf("search = %v, want true", opts["search"])
	}
}

// ============================================================================
// Chart sugar
// ============================================================================

func TestBuilder_ChartSugar(t *testing.T) {
	tests := []struct {
		name      string
		chartType ChartType
		wantType  string
	}{
		{"bar chart", BarChart, "bar"},
		{"area chart", AreaChart, "area"},
		{"horizontal bar", HBarChart, "hbar"},
		{"donut chart", DonutChart, "donut"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := parseClientJSON(t, Client(nil).
				Source("/api/revenue").
				Chart(tt.chartType).
				Render(),
			)

			if cfg["component"] != "chart" {
				t.Errorf("component = %v, want chart", cfg["component"])
			}

			opts := cfg["opts"].(map[string]any)
			if opts["type"] != tt.wantType {
				t.Errorf("opts.type = %v, want %s", opts["type"], tt.wantType)
			}
		})
	}
}

func TestBuilder_ChartOptions(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/revenue").
		Chart(BarChart).
		ChartOptions(Opts{"width": 600, "height": 400, "stacked": true}).
		Render(),
	)

	opts := cfg["opts"].(map[string]any)
	if opts["type"] != "bar" {
		t.Errorf("opts.type = %v, want bar", opts["type"])
	}
	if opts["width"] != float64(600) {
		t.Errorf("opts.width = %v, want 600", opts["width"])
	}
	if opts["height"] != float64(400) {
		t.Errorf("opts.height = %v, want 400", opts["height"])
	}
	if opts["stacked"] != true {
		t.Errorf("opts.stacked = %v, want true", opts["stacked"])
	}
}

func TestBuilder_ChartOptionsMerge(t *testing.T) {
	// ChartOptions should merge, not replace
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Chart(AreaChart).
		ChartOptions(Opts{"colors": []string{"#ff0000"}}).
		ChartOptions(Opts{"legend": true}).
		Render(),
	)

	opts := cfg["opts"].(map[string]any)
	if opts["type"] != "area" {
		t.Error("type should survive multiple ChartOptions calls")
	}
	if opts["legend"] != true {
		t.Error("second ChartOptions should merge")
	}
	if opts["colors"] == nil {
		t.Error("first ChartOptions values should survive merge")
	}
}

// ============================================================================
// Full builder chain – complex scenario
// ============================================================================

func TestBuilder_FullChain(t *testing.T) {
	html := Client(nil).
		Source("/api/invoices").
		Params(map[string]string{"year": "2025"}).
		Table(
			Col("number").Label("#").Sortable(true),
			Col("company").Label("Firma").Type("text"),
			Col("total").Label("Celkom").Type("number").Format("amount").Sortable(true),
			Col("status").Label("Stav").Type("enum").
				EnumOptions(
					Option{Value: "paid", Label: "Zaplatené"},
					Option{Value: "unpaid", Label: "Nezaplatené"},
				),
		).
		Filter(true).
		Search(true).
		Pagination(50).
		Loading(ui.SkeletonTable).
		Empty("inbox", "Žiadne faktúry").
		Poll(30 * time.Second).
		Error(false).
		AutoLoad(false).
		Render()

	cfg := parseClientJSON(t, html)

	// Core fields
	if cfg["source"] != "/api/invoices" {
		t.Errorf("source = %v", cfg["source"])
	}
	if cfg["component"] != "table" {
		t.Errorf("component = %v", cfg["component"])
	}
	if cfg["loading"] != "table" {
		t.Errorf("loading = %v", cfg["loading"])
	}
	if cfg["poll"] != float64(30000) {
		t.Errorf("poll = %v", cfg["poll"])
	}
	if cfg["error"] != false {
		t.Errorf("error = %v", cfg["error"])
	}
	if cfg["autoLoad"] != false {
		t.Errorf("autoLoad = %v", cfg["autoLoad"])
	}

	// Params
	params := cfg["params"].(map[string]any)
	if params["year"] != "2025" {
		t.Errorf("params.year = %v", params["year"])
	}

	// Empty state
	empty := cfg["empty"].(map[string]any)
	if empty["icon"] != "inbox" || empty["message"] != "Žiadne faktúry" {
		t.Errorf("empty = %v", empty)
	}

	// Opts
	opts := cfg["opts"].(map[string]any)
	if opts["filter"] != true {
		t.Error("filter should be true")
	}
	if opts["search"] != true {
		t.Error("search should be true")
	}
	if opts["pageSize"] != float64(50) {
		t.Errorf("pageSize = %v", opts["pageSize"])
	}

	// Columns
	cols := opts["columns"].([]any)
	if len(cols) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(cols))
	}

	col2 := cols[2].(map[string]any)
	if col2["key"] != "total" || col2["format"] != "amount" {
		t.Errorf("col[2] = %v", col2)
	}

	col3 := cols[3].(map[string]any)
	if col3["type"] != "enum" {
		t.Errorf("col[3].type = %v, want enum", col3["type"])
	}
	enumOpts := col3["enumOptions"].([]any)
	if len(enumOpts) != 2 {
		t.Fatalf("expected 2 enum options, got %d", len(enumOpts))
	}
}

// ============================================================================
// Edge cases
// ============================================================================

func TestBuilder_EmptySource(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).Component("test", nil).Render())
	if cfg["source"] != "" {
		t.Errorf("source = %v, want empty string", cfg["source"])
	}
}

func TestBuilder_NoComponent(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).Source("/api/x").Render())
	if _, ok := cfg["component"]; ok {
		t.Error("component should be omitted when not set")
	}
}

func TestBuilder_EmptyParams(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Params(map[string]string{}).
		Component("test", nil).
		Render(),
	)
	if _, ok := cfg["params"]; ok {
		t.Error("params should be omitted when empty map")
	}
}

func TestBuilder_EmptyOpts(t *testing.T) {
	cfg := parseClientJSON(t, Client(nil).
		Source("/api/x").
		Component("test", Opts{}).
		Render(),
	)
	if _, ok := cfg["opts"]; ok {
		t.Error("opts should be omitted when empty Opts")
	}
}

func TestBuilder_DivIDMatchesConfigID(t *testing.T) {
	html := Client(nil).Source("/api/x").Component("w", nil).Render()
	cfg := parseClientJSON(t, html)
	id := cfg["id"].(string)

	expectedDiv := `<div id="` + id + `"></div>`
	assertContains(t, html, expectedDiv)
}
