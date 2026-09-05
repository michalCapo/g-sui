package ui

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// TableQuery is URL-backed state for a load-more table. Fetch the first Limit()
// matching rows and the total count. SortColumn is a configured column index;
// translate it to an allowlisted database column, never interpolate raw input.
type TableQuery struct {
	Search     string
	Page       int
	PageSize   int
	SortColumn int
	Direction  string
	Filters    map[int]*FilterValue
}

func (q TableQuery) Limit() int { return min(max(q.Page, 1), 20) * min(max(q.PageSize, 1), 100) }

type TablePage[T any] struct {
	Rows  []*T
	Total int
}

// TableSource binds a typed loader to an existing DataTable. Query parameters
// are namespaced by table ID, so tables/tabs/users do not share filter state.
// Export is opt-in through the existing lower-level DataTable API.
type TableSource[T any] struct {
	id        string
	action    string
	load      func(*Context, TableQuery) (TablePage[T], error)
	configure func(*DataTable[T])
}

// RegisterTable registers once at startup. configure runs for every render and
// should configure columns, filters, page size and appearance without state.
func RegisterTable[T any](app *App, id string, load func(*Context, TableQuery) (TablePage[T], error), configure func(*DataTable[T])) *TableSource[T] {
	source := &TableSource[T]{id: id, action: "table." + id, load: load, configure: configure}
	app.action(source.action, func(ctx *Context) string {
		var req struct {
			Operation string `json:"operation"`
			Search    string `json:"search"`
			Page      int    `json:"page"`
			PageSize  int    `json:"pageSize"`
			Sort      int    `json:"sort"`
			Dir       string `json:"dir"`
			Col       int    `json:"col"`
			Type      string `json:"type"`
			FilterValue
		}
		if err := ctx.Body(&req); err != nil {
			return Notify("error", "Invalid table query")
		}
		table := source.table()
		query := source.query(ctx, table)
		query.Search, query.Page, query.PageSize, query.SortColumn, query.Direction = req.Search, req.Page, req.PageSize, req.Sort, req.Dir
		switch req.Operation {
		case "search", "sort", "loadmore":
		case "removeFilter":
			delete(query.Filters, req.Col)
		case "filter":
			if req.Type == "" {
				query.Filters = make(map[int]*FilterValue)
			} else {
				filter, ok := table.filters[req.Col]
				if !ok || string(filter.Type) != req.Type {
					return Notify("error", "Unknown filter")
				}
				value := req.FilterValue
				if value.Value == "" && len(value.Values) == 0 && value.From == "" && value.To == "" {
					delete(query.Filters, req.Col)
				} else {
					query.Filters[req.Col] = &value
				}
			}
		default:
			return Notify("error", "Unsupported table operation")
		}
		normalizeTableQuery(&query, table)
		node, err := source.render(ctx, table, query)
		if err != nil {
			return actionError(ctx, err)
		}
		u := *ctx.Request.URL
		params := u.Query()
		source.writeQuery(params, query)
		u.RawQuery = params.Encode()
		path, _ := json.Marshal(u.RequestURI())
		// Content and URL are committed in one reply. Browser Back uses __nav to
		// rebuild the full page from this same query, without a second registry.
		return node.ToJSMorph(source.id) + fmt.Sprintf("__gsuiQueryDone(%s,%t);", path, req.Operation == "search")
	})
	return source
}

func (source *TableSource[T]) table() *DataTable[T] {
	table := NewDataTable[T](source.id).Action(source.action)
	if source.configure != nil {
		source.configure(table)
	}
	table.hideExport = true
	return table
}
func (source *TableSource[T]) query(ctx *Context, table *DataTable[T]) TableQuery {
	values := ctx.Request.URL.Query()
	get := func(key string) string { return values.Get(source.id + "." + key) }
	page, _ := strconv.Atoi(get("page"))
	size, _ := strconv.Atoi(get("size"))
	sort := -1
	if get("sort") != "" {
		sort, _ = strconv.Atoi(get("sort"))
	}
	if size == 0 {
		size = table.pageSize
	}
	q := TableQuery{Search: get("q"), Page: page, PageSize: size, SortColumn: sort, Direction: get("dir"), Filters: make(map[int]*FilterValue)}
	_ = json.Unmarshal([]byte(get("filters")), &q.Filters)
	normalizeTableQuery(&q, table)
	return q
}
func normalizeTableQuery[T any](q *TableQuery, table *DataTable[T]) {
	q.Page = min(max(q.Page, 1), 20)
	q.PageSize = min(max(q.PageSize, 1), 100)
	if q.Direction != "desc" {
		q.Direction = "asc"
	}
	allowed := false
	for _, column := range table.sortable {
		if q.SortColumn == column {
			allowed = true
		}
	}
	if !allowed {
		q.SortColumn = -1
	}
	if q.Filters == nil {
		q.Filters = make(map[int]*FilterValue)
	}
	for column, value := range q.Filters {
		if _, ok := table.filters[column]; !ok || value == nil {
			delete(q.Filters, column)
		}
	}
}
func (source *TableSource[T]) writeQuery(values url.Values, q TableQuery) {
	put := func(k, v string) { values.Set(source.id+"."+k, v) }
	put("q", q.Search)
	put("page", strconv.Itoa(q.Page))
	put("size", strconv.Itoa(q.PageSize))
	put("sort", strconv.Itoa(q.SortColumn))
	put("dir", q.Direction)
	filters, _ := json.Marshal(q.Filters)
	put("filters", string(filters))
}
func (source *TableSource[T]) render(ctx *Context, table *DataTable[T], q TableQuery) (*Node, error) {
	page, err := source.load(ctx, q)
	if err != nil {
		return nil, err
	}
	rows := page.Rows
	if len(rows) > q.Limit() {
		rows = rows[:q.Limit()]
	}
	for column, value := range q.Filters {
		table.SetFilterValue(column, value)
	}
	return table.Page(q.Page).PageSize(q.PageSize).TotalItems(max(page.Total, 0)).
		HasMore(q.Page < 20 && len(rows) < page.Total).Search(q.Search).Sort(q.SortColumn, q.Direction).Render(rows), nil
}
func (source *TableSource[T]) Render(ctx *Context) (*Node, error) {
	table := source.table()
	return source.render(ctx, table, source.query(ctx, table))
}
