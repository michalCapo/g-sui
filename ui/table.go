package ui

import "fmt"

// ---------------------------------------------------------------------------
// DataTable: generic, configurable table with search, sort, pagination, export
// and advanced per-column filtering
// ---------------------------------------------------------------------------

// FilterType defines the type of filter for a column
type FilterType string

const (
	FilterTypeText   FilterType = "text"
	FilterTypeDate   FilterType = "date"
	FilterTypeNumber FilterType = "number"
	FilterTypeSelect FilterType = "select"
)

// FilterOperator defines the operator for text/number filters
type FilterOperator string

const (
	OpContains   FilterOperator = "contains"
	OpStartsWith FilterOperator = "startswith"
	OpEquals     FilterOperator = "equals"
	OpRange      FilterOperator = "range"
	OpGTE        FilterOperator = "gte"
	OpLTE        FilterOperator = "lte"
	OpGT         FilterOperator = "gt"
	OpLT         FilterOperator = "lt"
)

// FilterValue represents a filter value for a column
type FilterValue struct {
	Operator string   `json:"op"`
	Value    string   `json:"val"`
	Values   []string `json:"vals"` // for multi-select
	From     string   `json:"from"` // for date/number range
	To       string   `json:"to"`   // for date/number range
}

// ColumnFilter defines filter configuration for a column
type ColumnFilter struct {
	Type    FilterType
	Options []string // for select type
	Label   string   // column label for the filter popup
}

// DataTable is a generic, configurable table component with built-in
// search, pagination, sorting indicators, export support, and per-column filtering.
// Data fetching is delegated to user-defined WS actions.
type DataTable[T any] struct {
	id          string
	heads       []tableHead
	fields      []tableField[T]
	action      string // WS action name for all operations (search, sort, page, export)
	page        int    // current page (1-based)
	totalPages  int    // total pages
	totalItems  int    // total item count (optional, for display)
	sortCol     int    // currently sorted column index (-1 = none)
	sortDir     string // "asc" or "desc"
	searchValue string // current search query
	pageSize    int    // items per page (default 10)
	sortable    []int  // which column indices are sortable
	cls         string // wrapper class
	tableCls    string // <table> class
	emptyText   string // text when no data
	rowOffset   int    // offset for alternating row colors (used when appending)
	hasMore     bool   // whether there are more items to load

	// Filtering
	filters      map[int]*ColumnFilter // column index -> filter config
	filterValues map[int]*FilterValue  // column index -> current filter value
	filterLabels []FilterBadge         // active filter badges to display
}

// FilterBadge represents an active filter badge
type FilterBadge struct {
	Label    string
	Value    string
	Column   int
	OnRemove string // JS to remove this filter
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
		id:           id,
		page:         1,
		totalPages:   1,
		pageSize:     10,
		sortCol:      -1,
		sortDir:      "asc",
		emptyText:    "No data",
		tableCls:     "w-full table-auto text-sm",
		filters:      make(map[int]*ColumnFilter),
		filterValues: make(map[int]*FilterValue),
	}
}

// ---------------------------------------------------------------------------
// Column definition methods
// ---------------------------------------------------------------------------

// ColOpt configures a column added via Col.
type ColOpt[T any] struct {
	Text          func(*T) *Node // cell renderer
	Sortable      bool           // whether the column is sortable
	Filter        FilterType     // filter type (NumFilter, TxtFilter, DateFilter, SelectFilter); empty = no filter
	FilterOptions []string       // options for SelectFilter
	HeadCls       string         // CSS class for the <th>
	CellCls       string         // CSS class for the <td>
}

const (
	NumFilter    FilterType = FilterTypeNumber
	TxtFilter    FilterType = FilterTypeText
	DateFilter   FilterType = FilterTypeDate
	SelectFilter FilterType = FilterTypeSelect
)

// Col adds a column with header label and options including the render function.
// The label is also used as the filter label when a Filter type is set.
func (dt *DataTable[T]) Col(label string, opt ColOpt[T]) *DataTable[T] {
	dt.heads = append(dt.heads, tableHead{label: label, cls: opt.HeadCls})
	dt.fields = append(dt.fields, tableField[T]{render: opt.Text, cls: opt.CellCls})
	colIdx := len(dt.heads) - 1
	if opt.Sortable {
		dt.sortable = append(dt.sortable, colIdx)
	}
	if opt.Filter != "" {
		dt.filters[colIdx] = &ColumnFilter{Type: opt.Filter, Label: label, Options: opt.FilterOptions}
	}
	return dt
}

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

// Action sets the single WS action name for all table operations (search, sort, page, export).
// The action will receive: {operation, search, page, sort, dir}
// where operation is one of: "search", "sort", "page", "export"
func (dt *DataTable[T]) Action(actionName string) *DataTable[T] {
	dt.action = actionName
	return dt
}

// Sortable marks which column indices support click-to-sort.
func (dt *DataTable[T]) Sortable(columns ...int) *DataTable[T] {
	dt.sortable = columns
	return dt
}

// FilterText adds a text filter to a column with the given index.
// Deprecated: Use Col() with ColOpt{Filter: TxtFilter(label)} instead.
func (dt *DataTable[T]) FilterText(colIdx int, label string) *DataTable[T] {
	dt.filters[colIdx] = &ColumnFilter{
		Type:  FilterTypeText,
		Label: label,
	}
	return dt
}

// FilterDate adds a date filter to a column with the given index.
// Deprecated: Use Col() with ColOpt{Filter: DateFilter(label)} instead.
func (dt *DataTable[T]) FilterDate(colIdx int, label string) *DataTable[T] {
	dt.filters[colIdx] = &ColumnFilter{
		Type:  FilterTypeDate,
		Label: label,
	}
	return dt
}

// FilterNumber adds a number filter to a column with the given index.
// Deprecated: Use Col() with ColOpt{Filter: NumFilter(label)} instead.
func (dt *DataTable[T]) FilterNumber(colIdx int, label string) *DataTable[T] {
	dt.filters[colIdx] = &ColumnFilter{
		Type:  FilterTypeNumber,
		Label: label,
	}
	return dt
}

// FilterSelect adds a select/multi-select filter to a column with the given index.
// Deprecated: Use Col() with ColOpt{Filter: SelectFilter(label, options)} instead.
func (dt *DataTable[T]) FilterSelect(colIdx int, label string, options []string) *DataTable[T] {
	dt.filters[colIdx] = &ColumnFilter{
		Type:    FilterTypeSelect,
		Label:   label,
		Options: options,
	}
	return dt
}

// SetFilterValue sets a filter value for a column.
func (dt *DataTable[T]) SetFilterValue(colIdx int, value *FilterValue) *DataTable[T] {
	if value != nil {
		dt.filterValues[colIdx] = value
	} else {
		delete(dt.filterValues, colIdx)
	}
	return dt
}

// SetFilterLabels sets the active filter badges to display above the table.
func (dt *DataTable[T]) SetFilterLabels(badges []FilterBadge) *DataTable[T] {
	dt.filterLabels = badges
	return dt
}

// Page sets the current page number (1-based).
func (dt *DataTable[T]) Page(page int) *DataTable[T] {
	dt.page = page
	return dt
}

// PageSize sets the number of items per page (default 10).
func (dt *DataTable[T]) PageSize(size int) *DataTable[T] {
	dt.pageSize = size
	return dt
}

// TotalPages sets the total number of pages.
func (dt *DataTable[T]) TotalPages(total int) *DataTable[T] {
	dt.totalPages = total
	return dt
}

// TotalItems sets the total item count for display (e.g. "42 items").
func (dt *DataTable[T]) TotalItems(count int) *DataTable[T] {
	dt.totalItems = count
	return dt
}

// getAction returns the action name for table operations.
func (dt *DataTable[T]) getAction() string {
	return dt.action
}
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

// HasMore indicates there are more rows to load (shows "load more" button).
func (dt *DataTable[T]) HasMore(more bool) *DataTable[T] {
	dt.hasMore = more
	return dt
}

// RowOffset sets the starting offset for alternating row stripe colors.
// Used when appending rows to maintain correct striping.
func (dt *DataTable[T]) RowOffset(offset int) *DataTable[T] {
	dt.rowOffset = offset
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

	// 1. Toolbar (search + filter badges + reset)
	if toolbar := dt.renderToolbar(); toolbar != nil {
		children = append(children, toolbar)
	}

	// 2. Table with overflow wrapper (relative for filter popups)
	tableWrap := Div("overflow-x-auto relative").Render(dt.renderTable(data))
	children = append(children, tableWrap)

	// 3. Footer (export + item count)
	if dt.getAction() != "" {
		children = append(children, dt.renderFooter())
	}

	wrapCls := dt.cls
	if wrapCls == "" {
		wrapCls = "w-full"
	}

	return Div(wrapCls).ID(dt.id).Attr("data-page", fmt.Sprintf("%d", dt.page)).Render(children...)
}

// ---------------------------------------------------------------------------
// Toolbar: search input + filter count badge + filter badges + reset
// ---------------------------------------------------------------------------

func (dt *DataTable[T]) renderToolbar() *Node {
	action := dt.getAction()
	if action == "" {
		return nil
	}

	filterBarItems := make([]*Node, 0)

	// Search input with magnifying glass icon
	searchID := dt.id + "-search"
	searchIcon := Span("text-gray-400 dark:text-gray-500 text-lg leading-none absolute left-3 top-1/2 -translate-y-1/2").
		Style("font-family", "Material Icons Round").
		Text("search")
	searchInput := ISearch(
		"w-64 border border-gray-300 dark:border-gray-600 rounded-full pl-10 pr-4 py-2 text-sm "+
			"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 "+
			"placeholder-gray-400 dark:placeholder-gray-500 "+
			"focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400",
	).ID(searchID).
		Attr("placeholder", "Hľadať...").
		Attr("value", dt.searchValue).
		On("keydown", JS(dt.searchEnterJS(searchID))).
		On("search", JS(dt.searchImmediateJS(searchID)))

	searchWrap := Div("relative inline-flex items-center").Render(searchIcon, searchInput)
	filterBarItems = append(filterBarItems, searchWrap)

	// Filter count badge (lime green circle)
	activeFilterCount := len(dt.filterValues)
	if activeFilterCount > 0 {
		countBadge := Span(
			"inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-bold " +
				"bg-lime-400 text-gray-900",
		).Text(fmt.Sprintf("%d", activeFilterCount))
		filterBarItems = append(filterBarItems, countBadge)
	}

	// Active filter badges (pill shaped with x to remove)
	for _, badge := range dt.filterLabels {
		badgeNode := dt.renderFilterBadge(badge)
		filterBarItems = append(filterBarItems, badgeNode)
	}

	// Spacer
	filterBarItems = append(filterBarItems, Div("flex-1"))

	// Reset button (right aligned)
	if activeFilterCount > 0 || dt.searchValue != "" {
		resetBtn := Button(
			"text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 " +
				"cursor-pointer transition-colors",
		).Text("Zrušiť").OnClick(JS(dt.resetFiltersJS()))
		filterBarItems = append(filterBarItems, resetBtn)
	}

	return Div("flex items-center gap-3 mb-4 flex-wrap").Render(filterBarItems...)
}

func (dt *DataTable[T]) renderFilterBadge(badge FilterBadge) *Node {
	// Badge: "Label: Value  ×" in a rounded pill
	return Div(
		"inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm "+
			"bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300",
	).Render(
		Span().Text(badge.Label+": "),
		Span("font-medium").Text(badge.Value),
		Button(
			"ml-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 cursor-pointer "+
				"focus:outline-none text-base leading-none",
		).Attr("type", "button").
			OnClick(JS(badge.OnRemove)).
			Text("×"),
	)
}

func (dt *DataTable[T]) resetFiltersJS() string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	return fmt.Sprintf(
		"__ws.call('%s',{operation:'filter',search:'',page:1,pageSize:%d,sort:%d,dir:'%s',filters:{}})",
		escJS(action), dt.pageSize, dt.sortCol, escJS(dt.sortDir),
	)
}

func (dt *DataTable[T]) searchEnterJS(searchID string) string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	return fmt.Sprintf(
		"if(event.key==='Enter'){event.preventDefault();"+
			"__ws.call('%s',{operation:'search',search:document.getElementById('%s').value,page:1,pageSize:%d,sort:%d,dir:'%s'})}",
		escJS(action), escJS(searchID),
		dt.pageSize, dt.sortCol, escJS(dt.sortDir),
	)
}

func (dt *DataTable[T]) searchImmediateJS(searchID string) string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	return fmt.Sprintf(
		"__ws.call('%s',{operation:'search',search:document.getElementById('%s').value,page:1,pageSize:%d,sort:%d,dir:'%s'})",
		escJS(action), escJS(searchID),
		dt.pageSize, dt.sortCol, escJS(dt.sortDir),
	)
}

func (dt *DataTable[T]) exportJS() string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	return fmt.Sprintf(
		"__ws.call('%s',{operation:'export',search:'%s',pageSize:%d,sort:%d,dir:'%s'})",
		escJS(action), escJS(dt.searchValue),
		dt.pageSize, dt.sortCol, escJS(dt.sortDir),
	)
}

// ---------------------------------------------------------------------------
// Table: thead + tbody
// ---------------------------------------------------------------------------

func (dt *DataTable[T]) renderTable(data []*T) *Node {
	headerCells := make([]*Node, len(dt.heads))
	for i, h := range dt.heads {
		baseCls := "text-left font-semibold p-2 border-b border-gray-200 dark:border-gray-700 " +
			"text-gray-700 dark:text-gray-300 text-xs uppercase tracking-wider relative"
		if h.cls != "" {
			baseCls = h.cls + " relative"
		}

		th := Th(baseCls)
		hasSortOrFilter := dt.isSortable(i) || dt.hasFilter(i)

		if !hasSortOrFilter {
			// Plain header - no sort, no filter
			th.Text(h.label)
		} else {
			// Build header: LABEL [sort_arrow] [filter_icon]
			var headerParts []*Node

			// Label text (clickable for sort if sortable)
			if dt.isSortable(i) {
				labelSpan := Span("cursor-pointer select-none").Text(h.label)
				headerParts = append(headerParts, labelSpan)

				// Sort arrow: ↑ green when asc active, ↓ when desc, ⇅ gray when inactive
				indicator := dt.sortIndicator(i)
				headerParts = append(headerParts, indicator)

				// Sort click on the whole th
				th.OnClick(JS(dt.sortClickJS(i)))
				th.Class(" cursor-pointer select-none")
			} else {
				headerParts = append(headerParts, Span().Text(h.label))
			}

			// Filter icon (tune icon) - opens popup on click, stops propagation
			if dt.hasFilter(i) {
				filterIcon := dt.renderFilterIcon(i)
				headerParts = append(headerParts, filterIcon)
			}

			th.Render(Div("inline-flex items-center gap-1").Render(headerParts...))

			// Render filter popup (hidden by default, toggled by JS)
			if dt.hasFilter(i) {
				popup := dt.renderFilterPopupInline(i)
				th.Render(popup)
			}
		}

		headerCells[i] = th
	}

	thead := Thead("bg-gray-50 dark:bg-gray-800/50").Render(
		Tr().Render(headerCells...),
	)

	tbodyID := dt.id + "-tbody"
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
		tbody = Tbody().ID(tbodyID).Render(emptyRow)
	} else {
		rows := dt.buildRows(data)
		tbody = Tbody().ID(tbodyID).Render(rows...)
	}

	return Table(dt.tableCls).Render(thead, tbody)
}

// buildRows creates row nodes from data, using rowOffset for stripe coloring.
func (dt *DataTable[T]) buildRows(data []*T) []*Node {
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

		absIdx := dt.rowOffset + i
		rowCls := "hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors"
		if absIdx%2 == 1 {
			rowCls = "bg-gray-50/50 dark:bg-gray-800/20 hover:bg-gray-100/60 dark:hover:bg-gray-800/40 transition-colors"
		}
		rows[i] = Tr(rowCls).Render(cells...)
	}
	return rows
}

// RenderRows builds only the <tr> rows (no wrapper, no thead, no toolbar).
// Use with ToJSAppend to the tbody ID (dt.id + "-tbody") for "load more".
func (dt *DataTable[T]) RenderRows(data []*T) []*Node {
	return dt.buildRows(data)
}

func (dt *DataTable[T]) isSortable(colIdx int) bool {
	if dt.getAction() == "" {
		return false
	}
	for _, s := range dt.sortable {
		if s == colIdx {
			return true
		}
	}
	return false
}

func (dt *DataTable[T]) hasFilter(colIdx int) bool {
	_, exists := dt.filters[colIdx]
	return exists
}

func (dt *DataTable[T]) renderFilterIcon(colIdx int) *Node {
	isActive := dt.filterValues[colIdx] != nil
	iconColor := "text-gray-400 dark:text-gray-500"
	if isActive {
		iconColor = "text-lime-500 dark:text-lime-400"
	}

	popupID := fmt.Sprintf("%s-filter-popup-%d", dt.id, colIdx)

	// Filter icon (tune/sliders) - click toggles popup, stops propagation to prevent sort
	icon := Span("text-base leading-none "+iconColor+" cursor-pointer hover:text-lime-500 dark:hover:text-lime-400 transition-colors").
		Style("font-family", "Material Icons Round").
		Text("tune")

	btn := Button("inline-flex items-center focus:outline-none ml-0.5").
		Attr("type", "button").
		OnClick(JS(fmt.Sprintf(
			"event.stopPropagation();var p=document.getElementById('%s');"+
				"if(p.style.display==='none'||!p.style.display){"+
				"document.querySelectorAll('[id^=\"%s-filter-popup-\"]').forEach(function(el){el.style.display='none'});"+
				"var r=this.getBoundingClientRect();"+
				"p.style.left=r.left+'px';p.style.top=(r.bottom+4)+'px';"+
				"p.style.display='block'}else{p.style.display='none'}",
			escJS(popupID), escJS(dt.id),
		)))

	return btn.Render(icon)
}

// renderFilterPopupInline renders a filter popup dropdown (hidden by default)
// positioned absolutely below the column header
func (dt *DataTable[T]) renderFilterPopupInline(colIdx int) *Node {
	filter := dt.filters[colIdx]
	if filter == nil {
		return nil
	}

	popupID := fmt.Sprintf("%s-filter-popup-%d", dt.id, colIdx)
	currentValue := dt.filterValues[colIdx]

	popup := Div("fixed z-50 bg-white dark:bg-gray-800 "+
		"rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-2.5 w-[200px]").
		ID(popupID).
		Style("display", "none")

	// Header label
	header := Div("text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2").
		Text(filter.Label)

	// Content based on filter type
	var content *Node
	switch filter.Type {
	case FilterTypeDate:
		content = renderDateFilter(colIdx, currentValue)
	case FilterTypeNumber:
		content = renderNumberFilter(colIdx, currentValue)
	case FilterTypeSelect:
		content = renderSelectFilter(colIdx, filter.Options, currentValue)
	default:
		content = renderTextFilter(colIdx, currentValue)
	}

	// Action buttons: [Použiť]  Zrušiť
	actions := Div("flex items-center gap-2 mt-2.5").Render(
		Button(
			"px-3 py-1.5 text-xs font-medium rounded-md cursor-pointer "+
				"bg-gray-900 dark:bg-gray-700 text-white dark:text-gray-100 "+
				"hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors",
		).Text("Použiť").OnClick(JS(dt.applyFilterJS(colIdx))),
		Button(
			"text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 cursor-pointer",
		).Text("Zrušiť").OnClick(JS(fmt.Sprintf(
			"event.stopPropagation();document.getElementById('%s').style.display='none'",
			escJS(popupID),
		))),
	)

	// Stop click propagation on popup itself (prevent sort)
	popup.OnClick(JS("event.stopPropagation()"))

	return popup.Render(header, content, actions)
}

func (dt *DataTable[T]) applyFilterJS(colIdx int) string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	filter := dt.filters[colIdx]
	if filter == nil {
		return ""
	}

	popupID := fmt.Sprintf("%s-filter-popup-%d", dt.id, colIdx)

	switch filter.Type {
	case FilterTypeDate:
		return fmt.Sprintf(
			"event.stopPropagation();"+
				"var f=document.getElementById('filter-%d-from').value;"+
				"var t=document.getElementById('filter-%d-to').value;"+
				"document.getElementById('%s').style.display='none';"+
				"__ws.call('%s',{operation:'filter',col:%d,type:'date',from:f,to:t,search:'%s',page:1,pageSize:%d,sort:%d,dir:'%s'})",
			colIdx, colIdx, escJS(popupID),
			escJS(action), colIdx, escJS(dt.searchValue), dt.pageSize, dt.sortCol, escJS(dt.sortDir),
		)
	case FilterTypeNumber:
		return fmt.Sprintf(
			"event.stopPropagation();"+
				"var op=document.getElementById('filter-%d-op').value;"+
				"var f=document.getElementById('filter-%d-from').value;"+
				"var t=document.getElementById('filter-%d-to').value;"+
				"document.getElementById('%s').style.display='none';"+
				"__ws.call('%s',{operation:'filter',col:%d,type:'number',op:op,from:f,to:t,search:'%s',page:1,pageSize:%d,sort:%d,dir:'%s'})",
			colIdx, colIdx, colIdx, escJS(popupID),
			escJS(action), colIdx, escJS(dt.searchValue), dt.pageSize, dt.sortCol, escJS(dt.sortDir),
		)
	case FilterTypeSelect:
		return fmt.Sprintf(
			"event.stopPropagation();"+
				"var vals=[];document.querySelectorAll('[id^=\"filter-%d-opt-\"]').forEach(function(c){if(c.checked)vals.push(c.getAttribute('data-val'))});"+
				"document.getElementById('%s').style.display='none';"+
				"__ws.call('%s',{operation:'filter',col:%d,type:'select',vals:vals,search:'%s',page:1,pageSize:%d,sort:%d,dir:'%s'})",
			colIdx, escJS(popupID),
			escJS(action), colIdx, escJS(dt.searchValue), dt.pageSize, dt.sortCol, escJS(dt.sortDir),
		)
	default: // text
		return fmt.Sprintf(
			"event.stopPropagation();"+
				"var op=document.getElementById('filter-%d-op').value;"+
				"var v=document.getElementById('filter-%d-val').value;"+
				"document.getElementById('%s').style.display='none';"+
				"__ws.call('%s',{operation:'filter',col:%d,type:'text',op:op,val:v,search:'%s',page:1,pageSize:%d,sort:%d,dir:'%s'})",
			colIdx, colIdx, escJS(popupID),
			escJS(action), colIdx, escJS(dt.searchValue), dt.pageSize, dt.sortCol, escJS(dt.sortDir),
		)
	}
}

func (dt *DataTable[T]) sortIndicator(colIdx int) *Node {
	if dt.sortCol == colIdx {
		if dt.sortDir == "desc" {
			// Active desc: green down arrow
			return Span("text-lime-500 dark:text-lime-400 text-xs ml-0.5").Text("\u2193")
		}
		// Active asc: green up arrow
		return Span("text-lime-500 dark:text-lime-400 text-xs ml-0.5").Text("\u2191")
	}
	// Inactive: no visible arrow (empty span for spacing)
	return Span()
}

func (dt *DataTable[T]) sortClickJS(colIdx int) string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	newDir := "asc"
	if colIdx == dt.sortCol {
		if dt.sortDir == "asc" {
			newDir = "desc"
		} else {
			newDir = "asc"
		}
	}
	return fmt.Sprintf(
		"var cp=parseInt(document.getElementById('%s').getAttribute('data-page'))||%d;"+
			"__ws.call('%s',{operation:'sort',search:'%s',page:cp,pageSize:%d,sort:%d,dir:'%s'})",
		escJS(dt.id), dt.page,
		escJS(action), escJS(dt.searchValue),
		dt.pageSize, colIdx, escJS(newDir),
	)
}

// ---------------------------------------------------------------------------
// Footer: Excel export + item count
// ---------------------------------------------------------------------------

func (dt *DataTable[T]) renderFooter() *Node {
	footerID := dt.id + "-footer"
	action := dt.getAction()

	footerItems := make([]*Node, 0, 4)

	// Excel export button (bottom-left)
	if action != "" {
		exportBtn := Button(
			"inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-md cursor-pointer "+
				"border border-gray-300 dark:border-gray-600 "+
				"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 "+
				"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
		).OnClick(JS(dt.exportJS())).Render(
			Span("text-base leading-none").
				Style("font-family", "Material Icons Round").
				Text("grid_on"),
			Span().Text("Excel"),
		)
		footerItems = append(footerItems, exportBtn)
	}

	// Spacer
	footerItems = append(footerItems, Div("flex-1"))

	// "X of Y" count + load more button (right aligned)
	if dt.totalItems > 0 {
		showing := dt.page * dt.pageSize
		if showing > dt.totalItems {
			showing = dt.totalItems
		}
		countText := Span("text-sm text-gray-500 dark:text-gray-400").
			Text(fmt.Sprintf("%d z %d", showing, dt.totalItems))
		footerItems = append(footerItems, countText)
	}

	// Reset paging button (when user has loaded more than first page)
	if dt.page > 1 && action != "" {
		resetBtn := Button(
			"inline-flex items-center justify-center w-8 h-8 text-sm font-medium rounded-md cursor-pointer "+
				"border border-gray-300 dark:border-gray-600 "+
				"bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400 "+
				"hover:bg-gray-50 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-gray-200 transition-colors",
		).Text("×").OnClick(JS(dt.resetPagingJS()))
		footerItems = append(footerItems, resetBtn)
	}

	if dt.hasMore && action != "" {
		loadMoreBtn := Button(
			"inline-flex items-center gap-1 px-3 py-1.5 text-sm font-medium rounded-md cursor-pointer "+
				"border border-gray-300 dark:border-gray-600 "+
				"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 "+
				"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
		).Text("Načítať ďalšie...").OnClick(JS(dt.loadMoreJS()))
		footerItems = append(footerItems, loadMoreBtn)
	}

	footer := Div("flex items-center gap-3 mt-3").ID(footerID).Render(footerItems...)

	return footer
}

func (dt *DataTable[T]) loadMoreJS() string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	return fmt.Sprintf(
		"var cp=parseInt(document.getElementById('%s').getAttribute('data-page'))||%d;"+
			"document.getElementById('%s').setAttribute('data-page',cp+1);"+
			"__ws.call('%s',{operation:'loadmore',search:'%s',page:cp+1,pageSize:%d,sort:%d,dir:'%s'})",
		escJS(dt.id), dt.page,
		escJS(dt.id),
		escJS(action), escJS(dt.searchValue),
		dt.pageSize, dt.sortCol, escJS(dt.sortDir),
	)
}

func (dt *DataTable[T]) resetPagingJS() string {
	action := dt.getAction()
	if action == "" {
		return ""
	}
	return fmt.Sprintf(
		"document.getElementById('%s').setAttribute('data-page','1');"+
			"__ws.call('%s',{operation:'sort',search:'%s',page:1,pageSize:%d,sort:%d,dir:'%s'})",
		escJS(dt.id),
		escJS(action), escJS(dt.searchValue),
		dt.pageSize, dt.sortCol, escJS(dt.sortDir),
	)
}

// RenderFooter builds just the footer node (for replacing after load-more).
func (dt *DataTable[T]) RenderFooter() *Node {
	return dt.renderFooter()
}

// TbodyID returns the ID of the tbody element for use with ToJSAppend.
func (dt *DataTable[T]) TbodyID() string {
	return dt.id + "-tbody"
}

// FooterID returns the ID of the footer element for use with ToJSReplace.
func (dt *DataTable[T]) FooterID() string {
	return dt.id + "-footer"
}

// ---------------------------------------------------------------------------
// Filter Popup Component
// ---------------------------------------------------------------------------

// FilterPopup renders a filter popup for a specific column.
// This should be used by the server to render the popup content when
// operation='openFilter' is received.
func FilterPopup(colIdx int, colLabel string, filterType FilterType, options []string, currentValue *FilterValue) *Node {
	popupID := fmt.Sprintf("filter-popup-%d", colIdx)

	// Popup container
	popup := Div("absolute z-50 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-2.5 w-[200px]").
		ID(popupID)

	// Header
	header := Div("text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2").
		Text(colLabel)

	// Content based on filter type
	var content *Node
	switch filterType {
	case FilterTypeDate:
		content = renderDateFilter(colIdx, currentValue)
	case FilterTypeNumber:
		content = renderNumberFilter(colIdx, currentValue)
	case FilterTypeSelect:
		content = renderSelectFilter(colIdx, options, currentValue)
	default:
		content = renderTextFilter(colIdx, currentValue)
	}

	// Action buttons: Použiť (black) left, Zrušiť (text) right, no border
	actions := Div("flex items-center gap-2 mt-2.5").Render(
		Button(
			"px-3 py-1.5 text-xs font-medium rounded-md cursor-pointer "+
				"bg-gray-900 dark:bg-gray-700 text-white dark:text-gray-100 "+
				"hover:bg-gray-800 dark:hover:bg-gray-600 transition-colors",
		).Text("Použiť").OnClick(JS(fmt.Sprintf("applyFilter(%d)", colIdx))),
		Button(
			"text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 cursor-pointer",
		).Text("Zrušiť").OnClick(JS(fmt.Sprintf(
			"document.getElementById('filter-popup-%d').style.display='none'", colIdx,
		))),
	)

	return popup.Render(header, content, actions)
}

func renderTextFilter(colIdx int, currentValue *FilterValue) *Node {
	operator := OpContains
	value := ""
	if currentValue != nil {
		operator = FilterOperator(currentValue.Operator)
		value = currentValue.Value
	}

	// Operator dropdown
	opSelect := Select(
		"w-full mb-2 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded " +
			"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500",
	).ID(fmt.Sprintf("filter-%d-op", colIdx))

	opOptions := []struct {
		val   string
		label string
	}{
		{string(OpContains), "Obsahuje"},
		{string(OpStartsWith), "Začína na"},
		{string(OpEquals), "Rovná sa"},
	}

	for _, op := range opOptions {
		opt := Option().Attr("value", op.val).Text(op.label)
		if FilterOperator(op.val) == operator {
			opt.Attr("selected", "selected")
		}
		opSelect.Render(opt)
	}

	// Value input
	valInput := IText(
		"w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded "+
			"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 "+
			"focus:outline-none focus:ring-1 focus:ring-blue-500",
	).ID(fmt.Sprintf("filter-%d-val", colIdx)).
		Attr("placeholder", "Hľadaný text...").
		Attr("value", value)

	return Div().Render(opSelect, valInput)
}

func renderDateFilter(colIdx int, currentValue *FilterValue) *Node {
	from := ""
	to := ""
	if currentValue != nil {
		from = currentValue.From
		to = currentValue.To
	}

	inputCls := "flex-1 min-w-0 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded " +
		"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 " +
		"focus:outline-none focus:ring-1 focus:ring-blue-500"

	// Od (From) row
	fromRow := Div("flex items-center gap-2 mb-1.5").Render(
		Label("text-xs text-gray-500 dark:text-gray-400 w-6").Text("Od"),
		IDate(inputCls).ID(fmt.Sprintf("filter-%d-from", colIdx)).Attr("value", from),
	)

	// Do (To) row
	toRow := Div("flex items-center gap-2 mb-2").Render(
		Label("text-xs text-gray-500 dark:text-gray-400 w-6").Text("Do"),
		IDate(inputCls).ID(fmt.Sprintf("filter-%d-to", colIdx)).Attr("value", to),
	)

	// Quick select buttons
	quickButtons := Div("flex flex-wrap gap-1").Render(
		renderQuickDateBtn(colIdx, "Dnes", "today"),
		renderQuickDateBtn(colIdx, "Tento týždeň", "thisweek"),
		renderQuickDateBtn(colIdx, "Tento mesiac", "thismonth"),
		renderQuickDateBtn(colIdx, "Tento kvartál", "thisquarter"),
		renderQuickDateBtn(colIdx, "Tento rok", "thisyear"),
		renderQuickDateBtn(colIdx, "Minulý mesiac", "lastmonth"),
		renderQuickDateBtn(colIdx, "Minulý rok", "lastyear"),
	)

	return Div().Render(fromRow, toRow, quickButtons)
}

func renderQuickDateBtn(colIdx int, label, rangeType string) *Node {
	return Button(
		"px-2 py-1 text-[10px] rounded-full border border-gray-200 dark:border-gray-600 " +
			"bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-300 " +
			"hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors cursor-pointer",
	).Text(label).OnClick(JS(fmt.Sprintf(
		"(function(){"+
			"var d=new Date(),y=d.getFullYear(),m=d.getMonth(),day=d.getDate(),f,t;"+
			"function fmt(dt){return dt.toISOString().slice(0,10)}"+
			"switch('%s'){"+
			"case 'today':f=t=fmt(d);break;"+
			"case 'thisweek':var dow=d.getDay()||7;f=fmt(new Date(y,m,day-dow+1));t=fmt(new Date(y,m,day-dow+7));break;"+
			"case 'thismonth':f=fmt(new Date(y,m,1));t=fmt(new Date(y,m+1,0));break;"+
			"case 'thisquarter':var q=Math.floor(m/3)*3;f=fmt(new Date(y,q,1));t=fmt(new Date(y,q+3,0));break;"+
			"case 'thisyear':f=fmt(new Date(y,0,1));t=fmt(new Date(y,11,31));break;"+
			"case 'lastmonth':f=fmt(new Date(y,m-1,1));t=fmt(new Date(y,m,0));break;"+
			"case 'lastyear':f=fmt(new Date(y-1,0,1));t=fmt(new Date(y-1,11,31));break;"+
			"}"+
			"document.getElementById('filter-%d-from').value=f;"+
			"document.getElementById('filter-%d-to').value=t;"+
			"})()",
		escJS(rangeType), colIdx, colIdx,
	)))
}

func renderNumberFilter(colIdx int, currentValue *FilterValue) *Node {
	operator := OpRange
	from := ""
	to := ""
	if currentValue != nil {
		operator = FilterOperator(currentValue.Operator)
		from = currentValue.From
		to = currentValue.To
	}

	// Operator dropdown
	opSelect := Select(
		"w-full mb-2 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded " +
			"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500",
	).ID(fmt.Sprintf("filter-%d-op", colIdx))

	opOptions := []struct {
		val   string
		label string
	}{
		{string(OpRange), "Rozsah"},
		{string(OpGTE), "≥ Väčšie alebo rovné"},
		{string(OpLTE), "≤ Menšie alebo rovné"},
		{string(OpGT), "> Väčšie ako"},
		{string(OpLT), "< Menšie ako"},
		{string(OpEquals), "= Rovná sa"},
	}

	for _, op := range opOptions {
		opt := Option().Attr("value", op.val).Text(op.label)
		if FilterOperator(op.val) == operator {
			opt.Attr("selected", "selected")
		}
		opSelect.Render(opt)
	}

	inputCls := "flex-1 min-w-0 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded " +
		"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 " +
		"focus:outline-none focus:ring-1 focus:ring-blue-500"

	fromID := fmt.Sprintf("filter-%d-from", colIdx)
	toID := fmt.Sprintf("filter-%d-to", colIdx)
	isRange := operator == OpRange

	fromPlaceholder := "Hodnota"
	if isRange {
		fromPlaceholder = "Od"
	}

	fromInput := INumber(inputCls).ID(fromID).
		Attr("placeholder", fromPlaceholder).
		Attr("value", from)

	toDisplay := "none"
	if isRange {
		toDisplay = "flex"
	}
	toWrap := Div("flex gap-1.5").ID(fmt.Sprintf("filter-%d-to-wrap", colIdx)).
		Style("display", toDisplay).
		Render(INumber(inputCls).ID(toID).Attr("placeholder", "Do").Attr("value", to))

	// Toggle "to" field visibility and "from" placeholder based on operator
	opSelect.On("change", JS(fmt.Sprintf(
		"var isRange=this.value==='range';"+
			"document.getElementById('%s').style.display=isRange?'flex':'none';"+
			"document.getElementById('%s').placeholder=isRange?'Od':'Hodnota';",
		escJS(fmt.Sprintf("filter-%d-to-wrap", colIdx)), escJS(fromID),
	)))

	inputs := Div("flex flex-col gap-1.5").Render(fromInput, toWrap)

	return Div().Render(opSelect, inputs)
}

func renderSelectFilter(colIdx int, options []string, currentValue *FilterValue) *Node {
	selected := make(map[string]bool)
	if currentValue != nil {
		for _, v := range currentValue.Values {
			selected[v] = true
		}
	}

	// Select all / Clear links
	header := Div("flex items-center justify-between mb-2").Render(
		Button(
			"text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 cursor-pointer",
		).Text("Vybrať všetko").OnClick(JS(fmt.Sprintf(
			"document.querySelectorAll('[id^=\"filter-%d-opt-\"]').forEach(function(c){c.checked=true})", colIdx,
		))),
		Button(
			"text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 cursor-pointer",
		).Text("Zrušiť výber").OnClick(JS(fmt.Sprintf(
			"document.querySelectorAll('[id^=\"filter-%d-opt-\"]').forEach(function(c){c.checked=false})", colIdx,
		))),
	)

	// Checkbox list
	checkboxes := Div("max-h-36 overflow-y-auto space-y-0.5")
	for idx, opt := range options {
		isChecked := selected[opt]
		chkID := fmt.Sprintf("filter-%d-opt-%d", colIdx, idx)
		chk := ICheckbox("mr-1.5 accent-gray-900 dark:accent-gray-300").
			ID(chkID).
			Attr("data-val", opt)
		if isChecked {
			chk.Attr("checked", "checked")
		}

		row := Div("flex items-center px-1.5 py-1 hover:bg-gray-50 dark:hover:bg-gray-700 rounded cursor-pointer").Render(
			chk,
			Label("text-xs text-gray-700 dark:text-gray-300 cursor-pointer").Attr("for", chkID).Text(opt),
		)
		checkboxes.Render(row)
	}

	return Div().Render(header, checkboxes)
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
