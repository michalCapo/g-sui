package ui

import "fmt"

// ---------------------------------------------------------------------------
// DataTable: generic, configurable table with search, sort, pagination, export
// ---------------------------------------------------------------------------

// DataTable is a generic, configurable table component with built-in
// search, pagination, sorting indicators, and export support.
// Data fetching is delegated to user-defined WS actions.
type DataTable[T any] struct {
	id           string
	heads        []tableHead
	fields       []tableField[T]
	searchAction string // WS action name for search (sends {search, page, sort, dir})
	sortAction   string // WS action name for sort (sends {search, page, sort, dir})
	pageAction   string // WS action name for pagination (sends {search, page, sort, dir})
	exportAction string // WS action name for export (sends {search, sort, dir})
	page         int    // current page (1-based)
	totalPages   int    // total pages
	totalItems   int    // total item count (optional, for display)
	sortCol      int    // currently sorted column index (-1 = none)
	sortDir      string // "asc" or "desc"
	searchValue  string // current search query
	sortable     []int  // which column indices are sortable
	cls          string // wrapper class
	tableCls     string // <table> class
	emptyText    string // text when no data
}

type tableHead struct {
	label string
	html  bool // if true, label is raw (used for HeadHTML)
	cls   string
}

type tableField[T any] struct {
	render func(*T) *Node  // returns a *Node for the cell content
	text   func(*T) string // returns escaped text (mutually exclusive with render)
	cls    string
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// NewDataTable creates a new DataTable with the given wrapper ID.
// The ID is used on the outer div, enabling Replace() for live updates.
func NewDataTable[T any](id string) *DataTable[T] {
	return &DataTable[T]{
		id:         id,
		page:       1,
		totalPages: 1,
		sortCol:    -1,
		sortDir:    "asc",
		emptyText:  "No data",
		tableCls:   "w-full table-auto text-sm",
	}
}

// ---------------------------------------------------------------------------
// Column definition methods
// ---------------------------------------------------------------------------

// Head adds a text header column.
func (dt *DataTable[T]) Head(label string, cls ...string) *DataTable[T] {
	c := ""
	if len(cls) > 0 {
		c = cls[0]
	}
	dt.heads = append(dt.heads, tableHead{label: label, cls: c})
	return dt
}

// HeadHTML adds a header column whose label is rendered as raw content.
func (dt *DataTable[T]) HeadHTML(label string, cls ...string) *DataTable[T] {
	c := ""
	if len(cls) > 0 {
		c = cls[0]
	}
	dt.heads = append(dt.heads, tableHead{label: label, html: true, cls: c})
	return dt
}

// Field adds a column whose cell content is a *Node returned by fn.
func (dt *DataTable[T]) Field(fn func(*T) *Node, cls ...string) *DataTable[T] {
	c := ""
	if len(cls) > 0 {
		c = cls[0]
	}
	dt.fields = append(dt.fields, tableField[T]{render: fn, cls: c})
	return dt
}

// FieldText adds a column whose cell content is plain text returned by fn.
// The text is set via .Text() on the <td>, which uses textContent (auto-escaped).
func (dt *DataTable[T]) FieldText(fn func(*T) string, cls ...string) *DataTable[T] {
	c := ""
	if len(cls) > 0 {
		c = cls[0]
	}
	dt.fields = append(dt.fields, tableField[T]{text: fn, cls: c})
	return dt
}

// ---------------------------------------------------------------------------
// Feature configuration methods
// ---------------------------------------------------------------------------

// Searchable enables the search input and sets the WS action to call on input.
func (dt *DataTable[T]) Searchable(actionName string) *DataTable[T] {
	dt.searchAction = actionName
	return dt
}

// Sortable marks which column indices support click-to-sort.
func (dt *DataTable[T]) Sortable(columns ...int) *DataTable[T] {
	dt.sortable = columns
	return dt
}

// SortAction sets the WS action name invoked when a sortable header is clicked.
func (dt *DataTable[T]) SortAction(actionName string) *DataTable[T] {
	dt.sortAction = actionName
	return dt
}

// Paginated enables pagination controls and sets the current page state.
func (dt *DataTable[T]) Paginated(actionName string, page, totalPages int) *DataTable[T] {
	dt.pageAction = actionName
	dt.page = page
	dt.totalPages = totalPages
	return dt
}

// TotalItems sets the total item count for display (e.g. "42 items").
func (dt *DataTable[T]) TotalItems(count int) *DataTable[T] {
	dt.totalItems = count
	return dt
}

// Export enables the export button with the given WS action name.
func (dt *DataTable[T]) Export(actionName string) *DataTable[T] {
	dt.exportAction = actionName
	return dt
}

// Sort sets the current sort column and direction ("asc" or "desc").
func (dt *DataTable[T]) Sort(col int, dir string) *DataTable[T] {
	dt.sortCol = col
	dt.sortDir = dir
	return dt
}

// Search sets the current search query value (for re-rendering).
func (dt *DataTable[T]) Search(val string) *DataTable[T] {
	dt.searchValue = val
	return dt
}

// Empty overrides the text shown when there are no rows.
func (dt *DataTable[T]) Empty(text string) *DataTable[T] {
	dt.emptyText = text
	return dt
}

// DataTableClass sets the wrapper div's CSS classes.
func (dt *DataTable[T]) DataTableClass(cls string) *DataTable[T] {
	dt.cls = cls
	return dt
}

// TableClass overrides the <table> element's CSS classes.
func (dt *DataTable[T]) TableClass(cls string) *DataTable[T] {
	dt.tableCls = cls
	return dt
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

// Render builds the full table Node tree from the provided data slice.
func (dt *DataTable[T]) Render(data []*T) *Node {
	children := make([]*Node, 0, 4)

	// 1. Toolbar (search + export)
	if toolbar := dt.renderToolbar(); toolbar != nil {
		children = append(children, toolbar)
	}

	// 2. Table with overflow wrapper
	tableWrap := Div("overflow-x-auto").Render(dt.renderTable(data))
	children = append(children, tableWrap)

	// 3. Pagination
	if dt.pageAction != "" {
		children = append(children, dt.renderPagination())
	}

	wrapCls := dt.cls
	if wrapCls == "" {
		wrapCls = "w-full"
	}

	return Div(wrapCls).ID(dt.id).Render(children...)
}

// ---------------------------------------------------------------------------
// Toolbar: search input + export button
// ---------------------------------------------------------------------------

func (dt *DataTable[T]) renderToolbar() *Node {
	if dt.searchAction == "" && dt.exportAction == "" {
		return nil
	}

	items := make([]*Node, 0, 2)

	if dt.searchAction != "" {
		searchID := dt.id + "-search"
		searchInput := ISearch(
			"w-64 border border-gray-300 dark:border-gray-600 rounded-md px-3 py-1.5 text-sm "+
				"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 "+
				"placeholder-gray-400 dark:placeholder-gray-500 "+
				"focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400",
		).ID(searchID).
			Attr("placeholder", "Search...").
			Attr("value", dt.searchValue).
			On("input", JS(dt.searchDebounceJS(searchID)))

		items = append(items, searchInput)
	} else {
		items = append(items, Div())
	}

	if dt.exportAction != "" {
		exportBtn := Button(
			"inline-flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-md cursor-pointer "+
				"border border-gray-300 dark:border-gray-600 "+
				"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 "+
				"hover:bg-gray-50 dark:hover:bg-gray-700",
		).OnClick(JS(dt.exportJS())).Render(
			Span("text-base leading-none").
				Style("font-family", "Material Icons Round").
				Text("download"),
			Span().Text("Export"),
		)
		items = append(items, exportBtn)
	}

	return Div("flex items-center justify-between gap-4 mb-4").Render(items...)
}

func (dt *DataTable[T]) searchDebounceJS(searchID string) string {
	return fmt.Sprintf(
		"clearTimeout(window['__dt_search_%s']);"+
			"window['__dt_search_%s']=setTimeout(function(){"+
			"__ws.call('%s',{search:document.getElementById('%s').value,page:1,sort:%d,dir:'%s'})"+
			"},300);",
		escJS(dt.id), escJS(dt.id),
		escJS(dt.searchAction), escJS(searchID),
		dt.sortCol, escJS(dt.sortDir),
	)
}

func (dt *DataTable[T]) exportJS() string {
	return fmt.Sprintf(
		"__ws.call('%s',{search:'%s',sort:%d,dir:'%s'})",
		escJS(dt.exportAction), escJS(dt.searchValue),
		dt.sortCol, escJS(dt.sortDir),
	)
}

// ---------------------------------------------------------------------------
// Table: thead + tbody
// ---------------------------------------------------------------------------

func (dt *DataTable[T]) renderTable(data []*T) *Node {
	headerCells := make([]*Node, len(dt.heads))
	for i, h := range dt.heads {
		baseCls := "text-left font-semibold p-2 border-b border-gray-200 dark:border-gray-700 " +
			"text-gray-700 dark:text-gray-300 text-xs uppercase tracking-wider"
		if h.cls != "" {
			baseCls = h.cls
		}

		th := Th(baseCls)

		if dt.isSortable(i) {
			th.Class(" cursor-pointer select-none hover:text-blue-600 dark:hover:text-blue-400")
			indicator := dt.sortIndicator(i)
			labelNode := Span().Text(h.label)
			inner := Div("inline-flex items-center gap-1").Render(labelNode, indicator)
			th.Render(inner)
			th.OnClick(JS(dt.sortClickJS(i)))
		} else {
			th.Text(h.label)
		}

		headerCells[i] = th
	}

	thead := Thead("bg-gray-50 dark:bg-gray-800/50").Render(
		Tr().Render(headerCells...),
	)

	var tbody *Node
	if len(data) == 0 {
		colSpan := len(dt.heads)
		if colSpan == 0 {
			colSpan = 1
		}
		emptyRow := Tr().Render(
			Td("text-center p-8 text-gray-400 dark:text-gray-500").
				Attr("colspan", fmt.Sprintf("%d", colSpan)).
				Text(dt.emptyText),
		)
		tbody = Tbody().Render(emptyRow)
	} else {
		rows := make([]*Node, len(data))
		for i, item := range data {
			cells := make([]*Node, len(dt.fields))
			for j, f := range dt.fields {
				cellCls := "p-2 border-b border-gray-100 dark:border-gray-700/50 text-gray-800 dark:text-gray-200"
				if f.cls != "" {
					cellCls = f.cls
				}
				td := Td(cellCls)

				if f.render != nil {
					content := f.render(item)
					if content != nil {
						td.Render(content)
					}
				} else if f.text != nil {
					td.Text(f.text(item))
				}

				cells[j] = td
			}

			rowCls := "hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors"
			if i%2 == 1 {
				rowCls = "bg-gray-50/50 dark:bg-gray-800/20 hover:bg-gray-100/60 dark:hover:bg-gray-800/40 transition-colors"
			}
			rows[i] = Tr(rowCls).Render(cells...)
		}
		tbody = Tbody().Render(rows...)
	}

	return Table(dt.tableCls).Render(thead, tbody)
}

func (dt *DataTable[T]) isSortable(colIdx int) bool {
	if dt.sortAction == "" {
		return false
	}
	for _, s := range dt.sortable {
		if s == colIdx {
			return true
		}
	}
	return false
}

func (dt *DataTable[T]) sortIndicator(colIdx int) *Node {
	if dt.sortCol == colIdx {
		if dt.sortDir == "desc" {
			return Span("text-blue-600 dark:text-blue-400 text-xs ml-0.5").Text("\u25bc")
		}
		return Span("text-blue-600 dark:text-blue-400 text-xs ml-0.5").Text("\u25b2")
	}
	return Span("text-gray-300 dark:text-gray-600 text-xs ml-0.5").Text("\u21c5")
}

func (dt *DataTable[T]) sortClickJS(colIdx int) string {
	newDir := "asc"
	if colIdx == dt.sortCol {
		if dt.sortDir == "asc" {
			newDir = "desc"
		} else {
			newDir = "asc"
		}
	}
	return fmt.Sprintf(
		"__ws.call('%s',{search:'%s',page:%d,sort:%d,dir:'%s'})",
		escJS(dt.sortAction), escJS(dt.searchValue),
		dt.page, colIdx, escJS(newDir),
	)
}

// ---------------------------------------------------------------------------
// Pagination
// ---------------------------------------------------------------------------

func (dt *DataTable[T]) renderPagination() *Node {
	pageInfo := fmt.Sprintf("Page %d of %d", dt.page, dt.totalPages)
	leftParts := make([]*Node, 0, 2)
	leftParts = append(leftParts,
		Span("text-sm text-gray-600 dark:text-gray-400").Text(pageInfo),
	)
	if dt.totalItems > 0 {
		leftParts = append(leftParts,
			Span("text-sm text-gray-400 dark:text-gray-500").Text(
				fmt.Sprintf("(%d items)", dt.totalItems),
			),
		)
	}
	left := Div("flex items-center gap-2").Render(leftParts...)

	buttons := make([]*Node, 0, 10)

	// Previous button
	prevCls := "px-2.5 py-1 text-xs font-medium rounded cursor-pointer " +
		"border border-gray-300 dark:border-gray-600 " +
		"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 " +
		"hover:bg-gray-50 dark:hover:bg-gray-700"
	if dt.page <= 1 {
		prevCls = "px-2.5 py-1 text-xs font-medium rounded " +
			"border border-gray-200 dark:border-gray-700 " +
			"bg-gray-100 dark:bg-gray-800 text-gray-300 dark:text-gray-600 " +
			"cursor-not-allowed"
	}
	prevBtn := Button(prevCls).Text("\u2190  Prev")
	if dt.page > 1 {
		prevBtn.OnClick(JS(dt.pageJS(dt.page - 1)))
	} else {
		prevBtn.Attr("disabled", "true")
	}
	buttons = append(buttons, prevBtn)

	// Page number buttons
	pageNums := dt.pageRange()
	for _, p := range pageNums {
		if p == -1 {
			buttons = append(buttons,
				Span("px-1 text-xs text-gray-400 dark:text-gray-500 self-end").Text("\u2026"),
			)
		} else {
			isActive := p == dt.page
			pgCls := "w-8 h-8 text-xs font-medium rounded cursor-pointer " +
				"border border-gray-300 dark:border-gray-600 " +
				"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 " +
				"hover:bg-gray-50 dark:hover:bg-gray-700"
			if isActive {
				pgCls = "w-8 h-8 text-xs font-bold rounded " +
					"bg-blue-600 dark:bg-blue-500 text-white " +
					"border border-blue-600 dark:border-blue-500"
			}
			btn := Button(pgCls).Text(fmt.Sprintf("%d", p))
			if !isActive {
				btn.OnClick(JS(dt.pageJS(p)))
				btn.Class(" cursor-pointer")
			}
			buttons = append(buttons, btn)
		}
	}

	// Next button
	nextCls := "px-2.5 py-1 text-xs font-medium rounded cursor-pointer " +
		"border border-gray-300 dark:border-gray-600 " +
		"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 " +
		"hover:bg-gray-50 dark:hover:bg-gray-700"
	if dt.page >= dt.totalPages {
		nextCls = "px-2.5 py-1 text-xs font-medium rounded " +
			"border border-gray-200 dark:border-gray-700 " +
			"bg-gray-100 dark:bg-gray-800 text-gray-300 dark:text-gray-600 " +
			"cursor-not-allowed"
	}
	nextBtn := Button(nextCls).Text("Next  \u2192")
	if dt.page < dt.totalPages {
		nextBtn.OnClick(JS(dt.pageJS(dt.page + 1)))
	} else {
		nextBtn.Attr("disabled", "true")
	}
	buttons = append(buttons, nextBtn)

	right := Div("flex items-center gap-1").Render(buttons...)

	return Div("flex items-center justify-between mt-4").Render(left, right)
}

func (dt *DataTable[T]) pageJS(page int) string {
	return fmt.Sprintf(
		"__ws.call('%s',{search:'%s',page:%d,sort:%d,dir:'%s'})",
		escJS(dt.pageAction), escJS(dt.searchValue),
		page, dt.sortCol, escJS(dt.sortDir),
	)
}

// pageRange computes which page numbers to display.
// Returns a slice of ints where -1 represents an ellipsis.
func (dt *DataTable[T]) pageRange() []int {
	total := dt.totalPages
	current := dt.page

	if total <= 7 {
		pages := make([]int, total)
		for i := range pages {
			pages[i] = i + 1
		}
		return pages
	}

	pages := make([]int, 0, 9)
	pages = append(pages, 1)

	if current > 3 {
		pages = append(pages, -1)
	}

	start := current - 1
	end := current + 1
	if start < 2 {
		start = 2
	}
	if end > total-1 {
		end = total - 1
	}

	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	if current < total-2 {
		pages = append(pages, -1)
	}

	if total > 1 {
		pages = append(pages, total)
	}

	return pages
}

// ---------------------------------------------------------------------------
// SimpleTable: non-generic quick table builder
// ---------------------------------------------------------------------------

// SimpleTable builds a basic table from manually added cells.
// Rows are auto-wrapped based on numCols.
type SimpleTable struct {
	numCols    int
	cls        string
	heads      []string
	rows       [][]*Node
	currentRow []*Node
}

// NewSimpleTable creates a new SimpleTable with the given number of columns.
func NewSimpleTable(numCols int, cls ...string) *SimpleTable {
	c := ""
	if len(cls) > 0 {
		c = cls[0]
	}
	return &SimpleTable{numCols: numCols, cls: c}
}

// SimpleHeader sets the header labels for the table.
func (t *SimpleTable) SimpleHeader(labels ...string) *SimpleTable {
	t.heads = labels
	return t
}

// Cell adds a *Node cell to the table. When the current row reaches
// numCols, it is flushed and a new row starts automatically.
func (t *SimpleTable) Cell(node *Node) *SimpleTable {
	t.currentRow = append(t.currentRow, node)
	if len(t.currentRow) == t.numCols {
		t.rows = append(t.rows, t.currentRow)
		t.currentRow = nil
	}
	return t
}

// CellText adds a plain text cell.
func (t *SimpleTable) CellText(text string) *SimpleTable {
	return t.Cell(Span().Text(text))
}

// Build renders the SimpleTable into a *Node.
func (t *SimpleTable) Build() *Node {
	if len(t.currentRow) > 0 {
		for len(t.currentRow) < t.numCols {
			t.currentRow = append(t.currentRow, Span())
		}
		t.rows = append(t.rows, t.currentRow)
		t.currentRow = nil
	}

	tableCls := "w-full table-auto text-sm"
	if t.cls != "" {
		tableCls = t.cls
	}

	parts := make([]*Node, 0, 2)

	if len(t.heads) > 0 {
		ths := make([]*Node, len(t.heads))
		for i, label := range t.heads {
			ths[i] = Th("text-left font-semibold p-2 border-b border-gray-200 dark:border-gray-700 " +
				"text-gray-700 dark:text-gray-300 text-xs uppercase tracking-wider").Text(label)
		}
		parts = append(parts, Thead("bg-gray-50 dark:bg-gray-800/50").Render(Tr().Render(ths...)))
	}

	if len(t.rows) > 0 {
		trs := make([]*Node, len(t.rows))
		for i, row := range t.rows {
			tds := make([]*Node, len(row))
			for j, cell := range row {
				tds[j] = Td("p-2 border-b border-gray-100 dark:border-gray-700/50 text-gray-800 dark:text-gray-200").
					Render(cell)
			}
			rowCls := "hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors"
			if i%2 == 1 {
				rowCls = "bg-gray-50/50 dark:bg-gray-800/20 hover:bg-gray-100/60 dark:hover:bg-gray-800/40 transition-colors"
			}
			trs[i] = Tr(rowCls).Render(tds...)
		}
		parts = append(parts, Tbody().Render(trs...))
	} else {
		colSpan := t.numCols
		if colSpan == 0 {
			colSpan = 1
		}
		emptyRow := Tr().Render(
			Td("text-center p-8 text-gray-400 dark:text-gray-500").
				Attr("colspan", fmt.Sprintf("%d", colSpan)).
				Text("No data"),
		)
		parts = append(parts, Tbody().Render(emptyRow))
	}

	return Table(tableCls).Render(parts...)
}
