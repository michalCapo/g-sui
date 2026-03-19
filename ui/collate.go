package ui

import "fmt"

// ---------------------------------------------------------------------------
// Collate locale
// ---------------------------------------------------------------------------

// CollateLocale holds all translatable strings used by Collate.
// Create one only when you need non-English text; pass it via .Locale().
//
//	loc := &ui.CollateLocale{Search: "Hľadať...", Apply: "Použiť", ...}
//	collate := ui.NewCollate[T]("c").Locale(loc)
type CollateLocale struct {
	FilterLocale             // date/range labels shared with DataTable
	Search            string // search input placeholder
	Apply             string // apply button
	Reset             string // reset button
	Excel             string // export button
	Filter            string // filter toggle button
	LoadMore          string // load more button
	NoData            string // empty state
	AllOption         string // "— All —" select option
	FiltersAndSorting string // panel header
	Filters           string // filters section header
	SortBy            string // sort section header

	// ItemCount formats "X of Y" — receives (showing, total).
	ItemCount func(showing, total int) string
}

func defaultCollateLocale() *CollateLocale {
	return &CollateLocale{
		FilterLocale: defaultFilterLocale(),
		Search:       "Search...", Apply: "Apply", Reset: "Reset",
		Excel: "Excel", Filter: "Filter", LoadMore: "Load more...",
		NoData: "No data", AllOption: "— All —",
		FiltersAndSorting: "Filters & Sorting", Filters: "Filters", SortBy: "Sort by",
		ItemCount: func(showing, total int) string { return fmt.Sprintf("%d of %d", showing, total) },
	}
}

// ---------------------------------------------------------------------------
// Collate: data component with filter panel, sort buttons, search, pagination
// ---------------------------------------------------------------------------
// Unlike DataTable (inline per-column filters & sort arrows), Collate uses
// a slide-out filter/sort panel with an Apply button. All operations go
// through a single WS action.

// CollateSortField describes a sortable field shown as a button in the filter panel.
type CollateSortField struct {
	Field string // DB/sort field name
	Label string // display label
}

// CollateFilterType defines the kind of filter control.
type CollateFilterType int

const (
	CollateBool       CollateFilterType = iota // checkbox
	CollateDateRange                           // from/to date pickers
	CollateSelect                              // dropdown select
	CollateMultiCheck                          // multiple checkboxes (like old BOOL with condition)
)

// CollateFilterField describes a filterable field shown in the filter panel.
type CollateFilterField struct {
	Field   string            // DB field name
	Label   string            // display label
	Type    CollateFilterType // control type
	Options []CollateOption   // for CollateSelect / CollateMultiCheck
}

// CollateOption is a key/value pair for select and multi-check filters.
type CollateOption struct {
	Value string
	Label string
}

// CollateFilterValue holds the current value of a filter.
type CollateFilterValue struct {
	Field string `json:"field"`
	Type  string `json:"type"`  // "bool", "date", "select"
	Bool  bool   `json:"bool"`  // for CollateBool
	From  string `json:"from"`  // for CollateDateRange
	To    string `json:"to"`    // for CollateDateRange
	Value string `json:"value"` // for CollateSelect
}

// Collate is a generic data component with a filter/sort panel, search bar,
// load-more pagination, and Excel export. Data fetching is delegated to a
// user-defined WS action.
type Collate[T any] struct {
	id           string
	action       string // WS action name
	sortFields   []CollateSortField
	filterFields []CollateFilterField
	rowFn        func(*T, int) *Node // renders a single data row/card

	// State (set by server on each render)
	limit      int    // items per page (default 20)
	page       int    // current page (1-based)
	totalItems int    // total matching items
	search     string // current search query
	order      string // current sort e.g. "name asc"
	hasMore    bool   // more items to load

	// Filter state
	filterValues map[string]*CollateFilterValue

	// Row detail (accordion)
	detail func(*T) *Node // renders expandable detail content below a row

	// UI options
	cls       string // wrapper class
	emptyText string
	emptyIcon string
	rowOffset int // for alternating stripes when appending

	// Locale (per-instance override; nil = English default)
	locale *CollateLocale
}

// NewCollate creates a new Collate with sensible defaults.
func NewCollate[T any](id string) *Collate[T] {
	return &Collate[T]{
		id:           id,
		limit:        20,
		page:         1,
		emptyIcon:    "inbox",
		filterValues: make(map[string]*CollateFilterValue),
	}
}

// Locale sets a per-instance locale for this component's UI strings.
// When nil (default), English text is used.
//
//	collate.Locale(&ui.CollateLocale{Search: "Hľadať...", Apply: "Použiť"})
func (c *Collate[T]) Locale(l *CollateLocale) *Collate[T] {
	c.locale = l
	return c
}

// loc returns the effective locale for this component.
func (c *Collate[T]) loc() *CollateLocale {
	if c.locale != nil {
		return c.locale
	}
	return defaultCollateLocale()
}

// ---------------------------------------------------------------------------
// Builder methods
// ---------------------------------------------------------------------------

// Action sets the WS action name for all operations.
// The action receives: {operation, search, page, limit, order, filters}
func (c *Collate[T]) Action(name string) *Collate[T] {
	c.action = name
	return c
}

// Sort adds sortable fields shown as buttons in the filter panel.
func (c *Collate[T]) Sort(fields ...CollateSortField) *Collate[T] {
	c.sortFields = append(c.sortFields, fields...)
	return c
}

// Filter adds filterable fields shown in the filter panel.
func (c *Collate[T]) Filter(fields ...CollateFilterField) *Collate[T] {
	c.filterFields = append(c.filterFields, fields...)
	return c
}

// Row sets the function that renders each data item.
// The function receives a pointer to the item and its index.
func (c *Collate[T]) Row(fn func(*T, int) *Node) *Collate[T] {
	c.rowFn = fn
	return c
}

// Limit sets items per page (default 20).
func (c *Collate[T]) Limit(n int) *Collate[T] {
	c.limit = n
	return c
}

// Page sets the current page (1-based).
func (c *Collate[T]) Page(p int) *Collate[T] {
	c.page = p
	return c
}

// TotalItems sets the total matching item count for display.
func (c *Collate[T]) TotalItems(n int) *Collate[T] {
	c.totalItems = n
	return c
}

// Search sets the current search value.
func (c *Collate[T]) Search(val string) *Collate[T] {
	c.search = val
	return c
}

// Order sets the current sort order (e.g. "name asc").
func (c *Collate[T]) Order(order string) *Collate[T] {
	c.order = order
	return c
}

// HasMore indicates there are more rows to load.
func (c *Collate[T]) HasMore(more bool) *Collate[T] {
	c.hasMore = more
	return c
}

// SetFilter sets a filter value for a field.
func (c *Collate[T]) SetFilter(field string, val *CollateFilterValue) *Collate[T] {
	if val != nil {
		c.filterValues[field] = val
	} else {
		delete(c.filterValues, field)
	}
	return c
}

// CollateClass sets the wrapper CSS classes.
func (c *Collate[T]) CollateClass(cls string) *Collate[T] {
	c.cls = cls
	return c
}

// Empty sets the empty state text.
func (c *Collate[T]) Empty(text string) *Collate[T] {
	c.emptyText = text
	return c
}

func (c *Collate[T]) getEmptyText() string {
	if c.emptyText != "" {
		return c.emptyText
	}
	return c.loc().NoData
}

// EmptyIcon sets the Material Icon name for the empty state.
func (c *Collate[T]) EmptyIcon(icon string) *Collate[T] {
	c.emptyIcon = icon
	return c
}

// RowOffset sets the starting offset for alternating row stripes.
func (c *Collate[T]) RowOffset(offset int) *Collate[T] {
	c.rowOffset = offset
	return c
}

// Detail sets a function that renders expandable detail content for each row.
// When set, clicking a row toggles an accordion-style detail panel below it
// with a smooth expand/collapse animation.
func (c *Collate[T]) Detail(fn func(*T) *Node) *Collate[T] {
	c.detail = fn
	return c
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

// Render builds the full Collate Node tree from the provided data slice.
func (c *Collate[T]) Render(data []*T) *Node {
	children := make([]*Node, 0, 4)

	// 1. Header: search + filter button + filter panel (wrapped in relative for positioning)
	if c.action != "" {
		headerParts := []*Node{c.renderHeader()}
		if len(c.sortFields) > 0 || len(c.filterFields) > 0 {
			headerParts = append(headerParts, c.renderFilterPanel())
		}
		children = append(children, Div("relative").Render(headerParts...))
	}

	// 3. Data rows
	if len(data) == 0 {
		children = append(children, c.renderEmpty())
	} else {
		bodyID := c.id + "-body"
		rows := c.buildRows(data)
		body := Div().ID(bodyID).Render(rows...)
		children = append(children, body)
	}

	// 4. Footer: export + count + load more
	if c.action != "" {
		children = append(children, c.renderFooter())
	}

	wrapCls := c.cls
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-2 w-full"
	}

	return Div(wrapCls).ID(c.id).
		Attr("data-page", fmt.Sprintf("%d", c.page)).
		Attr("data-limit", fmt.Sprintf("%d", c.limit)).
		Render(children...)
}

// RenderRows builds only the data row nodes (for appending on "load more").
func (c *Collate[T]) RenderRows(data []*T) []*Node {
	return c.buildRows(data)
}

// BodyID returns the ID of the body container for use with ToJSAppend.
func (c *Collate[T]) BodyID() string {
	return c.id + "-body"
}

// FooterID returns the ID of the footer element for use with ToJSReplace.
func (c *Collate[T]) FooterID() string {
	return c.id + "-footer"
}

// ---------------------------------------------------------------------------
// Internal: build rows
// ---------------------------------------------------------------------------

func (c *Collate[T]) buildRows(data []*T) []*Node {
	hasDetail := c.detail != nil
	capacity := len(data)
	if hasDetail {
		capacity *= 2 // data row + detail row
	}
	rows := make([]*Node, 0, capacity)

	for i, item := range data {
		idx := c.rowOffset + i

		if hasDetail {
			detailID := fmt.Sprintf("%s-detail-%d", c.id, idx)

			// Build toggle JS for detail row
			toggleJS := fmt.Sprintf(
				"(function(){"+
					"var d=document.getElementById('%s');"+
					"var inner=d.querySelector('.collate-detail-inner');"+
					"var chevron=d.previousElementSibling.querySelector('[data-detail-chevron]');"+
					"if(d.style.display==='none'||!d.style.display){"+
					"d.style.display='block';inner.style.maxHeight=inner.scrollHeight+'px';inner.style.opacity='1';"+
					"if(chevron)chevron.classList.add('rotate-180');"+
					"d.previousElementSibling.classList.add('bg-gray-50','dark:bg-gray-800/30')"+
					"}else{"+
					"inner.style.maxHeight='0';inner.style.opacity='0';"+
					"if(chevron)chevron.classList.remove('rotate-180');"+
					"setTimeout(function(){d.style.display='none';"+
					"d.previousElementSibling.classList.remove('bg-gray-50','dark:bg-gray-800/30')"+
					"},200)"+
					"}"+
					"})()",
				escJS(detailID),
			)

			// Create the data row
			var rowNode *Node
			if c.rowFn != nil {
				rowNode = c.rowFn(item, idx)
			}
			if rowNode != nil {
				// Add click handler to the row
				rowNode.Class(" cursor-pointer group transition-colors")
				rowNode.OnClick(JS(toggleJS))

				rows = append(rows, rowNode)

				// Detail row (hidden by default)
				detailContent := c.detail(item)
				innerWrap := Div("collate-detail-inner overflow-hidden transition-all duration-200 ease-in-out").
					Style("max-height", "0").
					Style("opacity", "0").
					Render(
						Div("px-6 py-4 border-t border-gray-200 dark:border-gray-700/50").Render(detailContent),
					)
				detailRow := Div().ID(detailID).
					Style("display", "none").
					Class("border-b border-gray-200 dark:border-gray-700/50 bg-gray-50/50 dark:bg-gray-800/30").
					Render(innerWrap)
				rows = append(rows, detailRow)
			}
		} else {
			// No detail, just render the row
			if c.rowFn != nil {
				if node := c.rowFn(item, idx); node != nil {
					rows = append(rows, node)
				}
			}
		}
	}
	return rows
}

// ---------------------------------------------------------------------------
// Internal: header (search + filter toggle)
// ---------------------------------------------------------------------------

func (c *Collate[T]) renderHeader() *Node {
	items := make([]*Node, 0, 4)

	// Search input
	searchID := c.id + "-search"
	searchIcon := Span("text-gray-400 dark:text-gray-500 text-lg leading-none absolute left-3 top-1/2 -translate-y-1/2").
		Style("font-family", "Material Icons Round").
		Text("search")
	searchInput := ISearch(
		"w-64 border border-gray-300 dark:border-gray-600 rounded-full pl-10 pr-4 py-2 text-sm "+
			"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 "+
			"placeholder-gray-400 dark:placeholder-gray-500 "+
			"focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400",
	).ID(searchID).
		Attr("placeholder", c.loc().Search).
		Attr("value", c.search).
		On("keydown", JS(c.searchEnterJS(searchID))).
		On("search", JS(c.searchImmediateJS(searchID)))

	searchWrap := Div("relative inline-flex items-center").Render(searchIcon, searchInput)
	items = append(items, searchWrap)

	// Spacer
	items = append(items, Div("flex-1"))

	// Excel export button (right aligned, next to filter)
	exportBtn := Button(
		"inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md cursor-pointer "+
			"border border-gray-300 dark:border-gray-600 "+
			"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 "+
			"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
	).OnClick(JS(c.exportJS())).Render(
		Span("text-base leading-none").
			Style("font-family", "Material Icons Round").
			Text("grid_on"),
		Span().Text(c.loc().Excel),
	)
	items = append(items, exportBtn)

	// Filter/sort toggle button (right aligned)
	if len(c.sortFields) > 0 || len(c.filterFields) > 0 {
		panelID := c.id + "-panel"
		activeCount := c.activeFilterCount()

		btnParts := make([]*Node, 0, 3)
		btnParts = append(btnParts,
			Span("text-base leading-none").
				Style("font-family", "Material Icons Round").
				Text("tune"),
		)
		btnParts = append(btnParts, Span().Text(c.loc().Filter))

		if activeCount > 0 {
			badge := Span(
				"inline-flex items-center justify-center w-5 h-5 rounded-full text-[10px] font-bold " +
					"bg-lime-400 text-gray-900",
			).Text(fmt.Sprintf("%d", activeCount))
			btnParts = append(btnParts, badge)
		}

		filterBtn := Button(
			"inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md cursor-pointer " +
				"border border-gray-300 dark:border-gray-600 " +
				"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 " +
				"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
		).OnClick(JS(fmt.Sprintf(
			"var p=document.getElementById('%s');p.classList.toggle('hidden')",
			escJS(panelID),
		))).Render(btnParts...)

		items = append(items, filterBtn)
	}

	return Div("flex items-center gap-3 flex-wrap").Render(items...)
}

func (c *Collate[T]) activeFilterCount() int {
	count := 0
	for _, v := range c.filterValues {
		if v.Bool || v.Value != "" || v.From != "" || v.To != "" {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// Internal: filter panel
// ---------------------------------------------------------------------------

func (c *Collate[T]) renderFilterPanel() *Node {
	panelID := c.id + "-panel"

	parts := make([]*Node, 0, 6)

	// Header
	parts = append(parts,
		Div("flex items-center justify-between mb-3").Render(
			Span("text-sm font-semibold text-gray-700 dark:text-gray-300").Text(c.loc().FiltersAndSorting),
			Button(
				"w-8 h-8 rounded-full flex items-center justify-center "+
					"hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer transition-colors",
			).OnClick(JS(fmt.Sprintf(
				"document.getElementById('%s').classList.add('hidden')", escJS(panelID),
			))).Render(
				Span("text-base leading-none text-gray-400").
					Style("font-family", "Material Icons Round").
					Text("close"),
			),
		),
	)

	// Sort section
	if len(c.sortFields) > 0 {
		parts = append(parts, c.renderSortSection())
	}

	// Filters section
	if len(c.filterFields) > 0 {
		parts = append(parts, c.renderFiltersSection())
	}

	// Footer: Reset + Apply
	parts = append(parts, c.renderPanelFooter())

	return Div(
		"hidden absolute right-0 top-full mt-2 z-50 " +
			"bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 " +
			"shadow-2xl p-4 w-96",
	).ID(panelID).
		OnClick(JS("event.stopPropagation()")).
		Render(parts...)
}

func (c *Collate[T]) renderSortSection() *Node {
	buttons := make([]*Node, 0, len(c.sortFields))

	for _, sf := range c.sortFields {
		buttons = append(buttons, c.renderSortButton(sf))
	}

	return Div("flex flex-col gap-2 mb-3").Render(
		Div("text-xs font-bold text-gray-600 dark:text-gray-400 mb-1").Text(c.loc().SortBy),
		Div("flex flex-wrap gap-1").Render(buttons...),
	)
}

func (c *Collate[T]) renderSortButton(sf CollateSortField) *Node {
	btnID := fmt.Sprintf("%s-sort-%s", c.id, sf.Field)
	orderID := c.id + "-pending-order"

	// Determine current sort state
	direction := c.parseSortDirection(sf.Field)

	iconName := "sort"
	if direction == "asc" {
		iconName = "arrow_upward"
	} else if direction == "desc" {
		iconName = "arrow_downward"
	}

	activeCls := "rounded text-sm bg-gray-900 dark:bg-gray-600 text-white font-medium " +
		"cursor-pointer select-none px-3 py-2 flex items-center gap-2 transition-colors"
	inactiveCls := "rounded text-sm border border-gray-300 dark:border-gray-600 " +
		"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 " +
		"hover:bg-gray-50 dark:hover:bg-gray-700 " +
		"cursor-pointer select-none px-3 py-2 flex items-center gap-2 transition-colors"

	btnCls := inactiveCls
	if direction != "" {
		btnCls = activeCls
	}

	// JS: cycle none -> asc -> desc -> none, update hidden input, toggle visual state
	cycleJS := fmt.Sprintf(
		`(function(){`+
			`var h=document.getElementById('%s');`+
			`if(!h)return;`+
			`var field='%s';`+
			`var parts=(h.value||'').trim().split(/\s+/);`+
			`var cf=(parts[0]||'').toLowerCase();`+
			`var cd=(parts[1]||'').toLowerCase();`+
			`var nd='';`+
			`if(cf===field.toLowerCase()){`+
			`if(cd==='asc'){nd='desc';h.value=field+' desc'}`+
			`else if(cd==='desc'){nd='';h.value=''}`+
			`else{nd='asc';h.value=field+' asc'}`+
			`}else{nd='asc';h.value=field+' asc'}`+
			// Update all sort buttons visually
			`var wrap=event.currentTarget.closest('[id$="-panel"]');`+
			`if(!wrap)return;`+
			`wrap.querySelectorAll('[data-sort-field]').forEach(function(b){`+
			`var bf=b.getAttribute('data-sort-field');`+
			`var icon=b.querySelector('[data-sort-icon]');`+
			`if(bf.toLowerCase()===field.toLowerCase()&&nd!==''){`+
			`b.className='%s';`+
			`if(icon)icon.textContent=nd==='asc'?'arrow_upward':'arrow_downward'`+
			`}else{`+
			`b.className='%s';`+
			`if(icon)icon.textContent='sort'`+
			`}});`+
			`})()`,
		escJS(orderID), escJS(sf.Field),
		escJS(activeCls), escJS(inactiveCls),
	)

	return Button(btnCls).
		ID(btnID).
		Attr("data-sort-field", sf.Field).
		OnClick(JS(cycleJS)).
		Render(
			Span("text-base leading-none").
				Style("font-family", "Material Icons Round").
				Attr("data-sort-icon", "1").
				Text(iconName),
			Span().Text(sf.Label),
		)
}

func (c *Collate[T]) parseSortDirection(field string) string {
	if c.order == "" {
		return ""
	}
	// Simple parse: "field asc" or "field desc"
	if len(c.order) <= len(field) {
		return ""
	}
	// Case-insensitive prefix match
	orderLower := toLower(c.order)
	fieldLower := toLower(field)
	if len(orderLower) > len(fieldLower) && orderLower[:len(fieldLower)] == fieldLower {
		rest := orderLower[len(fieldLower):]
		if len(rest) > 1 && rest[0] == ' ' {
			dir := rest[1:]
			if dir == "asc" {
				return "asc"
			}
			if dir == "desc" {
				return "desc"
			}
		}
	}
	return ""
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func (c *Collate[T]) renderFiltersSection() *Node {
	items := make([]*Node, 0, len(c.filterFields))

	for _, ff := range c.filterFields {
		items = append(items, c.renderFilterControl(ff))
	}

	return Div("flex flex-col gap-2 mt-2 pt-3 border-t border-gray-200 dark:border-gray-700").Render(
		append([]*Node{
			Div("text-xs font-bold text-gray-600 dark:text-gray-400 mb-1").Text(c.loc().Filters),
		}, items...)...,
	)
}

func (c *Collate[T]) renderFilterControl(ff CollateFilterField) *Node {
	current := c.filterValues[ff.Field]

	switch ff.Type {
	case CollateBool:
		return c.renderBoolFilter(ff, current)
	case CollateDateRange:
		return c.renderCollateDateFilter(ff, current)
	case CollateSelect:
		return c.renderCollateSelectFilter(ff, current)
	default:
		return c.renderBoolFilter(ff, current)
	}
}

func (c *Collate[T]) renderBoolFilter(ff CollateFilterField, current *CollateFilterValue) *Node {
	chkID := fmt.Sprintf("%s-filter-%s", c.id, ff.Field)
	chk := ICheckbox("mr-2 accent-gray-900 dark:accent-gray-300").
		ID(chkID).
		Attr("data-filter-field", ff.Field).
		Attr("data-filter-type", "bool")
	if current != nil && current.Bool {
		chk.Attr("checked", "checked")
	}

	return Div("flex items-center py-1").Render(
		chk,
		Label("text-sm text-gray-700 dark:text-gray-300 cursor-pointer select-none").
			Attr("for", chkID).
			Text(ff.Label),
	)
}

func (c *Collate[T]) renderCollateDateFilter(ff CollateFilterField, current *CollateFilterValue) *Node {
	fromID := fmt.Sprintf("%s-filter-%s-from", c.id, ff.Field)
	toID := fmt.Sprintf("%s-filter-%s-to", c.id, ff.Field)

	fromVal := ""
	toVal := ""
	if current != nil {
		fromVal = current.From
		toVal = current.To
	}

	inputCls := "flex-1 min-w-0 px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded " +
		"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 " +
		"focus:outline-none focus:ring-1 focus:ring-blue-500"

	return Div("flex flex-col gap-1").Render(
		Label("text-xs font-medium text-gray-600 dark:text-gray-400").Text(ff.Label),
		Div("flex items-center gap-2").Render(
			Label("text-xs text-gray-500 dark:text-gray-400 w-6").Text(c.loc().From),
			IDate(inputCls).ID(fromID).
				Attr("value", fromVal).
				Attr("data-filter-field", ff.Field).
				Attr("data-filter-type", "date-from"),
		),
		Div("flex items-center gap-2").Render(
			Label("text-xs text-gray-500 dark:text-gray-400 w-6").Text(c.loc().To),
			IDate(inputCls).ID(toID).
				Attr("value", toVal).
				Attr("data-filter-type", "date-to"),
		),
		// Quick date buttons
		Div("flex flex-wrap gap-1 mt-1").Render(
			c.collateQuickDateBtn(fromID, toID, c.loc().Today, "today"),
			c.collateQuickDateBtn(fromID, toID, c.loc().ThisWeek, "thisweek"),
			c.collateQuickDateBtn(fromID, toID, c.loc().ThisMonth, "thismonth"),
			c.collateQuickDateBtn(fromID, toID, c.loc().ThisQuarter, "thisquarter"),
			c.collateQuickDateBtn(fromID, toID, c.loc().ThisYear, "thisyear"),
			c.collateQuickDateBtn(fromID, toID, c.loc().LastMonth, "lastmonth"),
			c.collateQuickDateBtn(fromID, toID, c.loc().LastYear, "lastyear"),
		),
	)
}

func (c *Collate[T]) collateQuickDateBtn(fromID, toID, label, rangeType string) *Node {
	return Button(
		"px-2 py-0.5 text-[10px] rounded-full border border-gray-200 dark:border-gray-600 "+
			"bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-300 "+
			"hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors cursor-pointer",
	).Attr("type", "button").Text(label).OnClick(JS(fmt.Sprintf(
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
			"document.getElementById('%s').value=f;"+
			"document.getElementById('%s').value=t;"+
			"})()",
		escJS(rangeType), escJS(fromID), escJS(toID),
	)))
}

func (c *Collate[T]) renderCollateSelectFilter(ff CollateFilterField, current *CollateFilterValue) *Node {
	selID := fmt.Sprintf("%s-filter-%s", c.id, ff.Field)

	currentVal := ""
	if current != nil {
		currentVal = current.Value
	}

	sel := Select(
		"w-full px-2 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded "+
			"bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 "+
			"focus:outline-none focus:ring-1 focus:ring-blue-500",
	).ID(selID).
		Attr("data-filter-field", ff.Field).
		Attr("data-filter-type", "select")

	// Empty option
	emptyOpt := Option().Attr("value", "").Text(c.loc().AllOption)
	if currentVal == "" {
		emptyOpt.Attr("selected", "selected")
	}
	sel.Render(emptyOpt)

	for _, opt := range ff.Options {
		o := Option().Attr("value", opt.Value).Text(opt.Label)
		if opt.Value == currentVal {
			o.Attr("selected", "selected")
		}
		sel.Render(o)
	}

	return Div("flex flex-col gap-1").Render(
		Label("text-xs font-medium text-gray-600 dark:text-gray-400").Text(ff.Label),
		sel,
	)
}

// ---------------------------------------------------------------------------
// Internal: panel footer (Reset + Apply)
// ---------------------------------------------------------------------------

func (c *Collate[T]) renderPanelFooter() *Node {
	orderID := c.id + "-pending-order"

	// Hidden input to track pending sort order
	hiddenOrder := IHidden().ID(orderID).Attr("value", c.order)

	return Div("flex flex-col gap-3 mt-4 pt-3 border-t border-gray-200 dark:border-gray-700").Render(
		hiddenOrder,
		Div("flex items-center justify-between").Render(
			// Reset button
			Button(
				"inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md cursor-pointer "+
					"border border-gray-300 dark:border-gray-600 "+
					"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 "+
					"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
			).Attr("type", "button").
				OnClick(JS(c.resetJS())).
				Render(
					Span("text-base leading-none").
						Style("font-family", "Material Icons Round").
						Text("undo"),
					Span().Text(c.loc().Reset),
				),

			// Apply button
			Button(
				"inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md cursor-pointer "+
					"bg-gray-900 dark:bg-gray-600 text-white "+
					"hover:bg-gray-800 dark:hover:bg-gray-500 transition-colors",
			).Attr("type", "button").
				OnClick(JS(c.applyJS())).
				Render(
					Span("text-base leading-none").
						Style("font-family", "Material Icons Round").
						Text("check"),
					Span().Text(c.loc().Apply),
				),
		),
	)
}

// ---------------------------------------------------------------------------
// Internal: empty state
// ---------------------------------------------------------------------------

func (c *Collate[T]) renderEmpty() *Node {
	bodyID := c.id + "-body"

	return Div("flex flex-col items-center justify-center py-16 "+
		"border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-xl "+
		"bg-white dark:bg-gray-800/50").ID(bodyID).Render(
		Span("text-6xl text-gray-300 dark:text-gray-600 mb-4").
			Style("font-family", "Material Icons Round").
			Text(c.emptyIcon),
		Span("text-gray-500 dark:text-gray-400 text-lg font-medium").
			Text(c.getEmptyText()),
	)
}

// ---------------------------------------------------------------------------
// Internal: footer (export + count + load more)
// ---------------------------------------------------------------------------

func (c *Collate[T]) renderFooter() *Node {
	footerID := c.id + "-footer"
	items := make([]*Node, 0, 4)

	// Spacer to push items right
	items = append(items, Div("flex-1"))

	// Count
	if c.totalItems > 0 {
		showing := c.page * c.limit
		if showing > c.totalItems {
			showing = c.totalItems
		}
		countText := Span("text-sm text-gray-500 dark:text-gray-400").
			Text(c.loc().ItemCount(showing, c.totalItems))
		items = append(items, countText)
	}

	// Reset paging button
	if c.page > 1 {
		resetBtn := Button(
			"inline-flex items-center justify-center w-8 h-8 text-sm font-medium rounded-md cursor-pointer " +
				"border border-gray-300 dark:border-gray-600 " +
				"bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400 " +
				"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
		).Text("×").OnClick(JS(c.resetPagingJS()))
		items = append(items, resetBtn)
	}

	// Load more button
	if c.hasMore {
		loadMoreBtn := Button(
			"inline-flex items-center gap-1 px-3 py-1.5 text-sm font-medium rounded-md cursor-pointer " +
				"border border-gray-300 dark:border-gray-600 " +
				"bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 " +
				"hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors",
		).Text(c.loc().LoadMore).OnClick(JS(c.loadMoreJS()))
		items = append(items, loadMoreBtn)
	}

	return Div("flex items-center gap-3 mt-3").ID(footerID).Render(items...)
}

// ---------------------------------------------------------------------------
// Internal: JS generators
// ---------------------------------------------------------------------------

func (c *Collate[T]) searchEnterJS(searchID string) string {
	if c.action == "" {
		return ""
	}
	return fmt.Sprintf(
		"if(event.key==='Enter'){event.preventDefault();"+
			"__ws.call('%s',{operation:'search',search:document.getElementById('%s').value,"+
			"page:1,limit:%d,order:'%s'})}",
		escJS(c.action), escJS(searchID), c.limit, escJS(c.order),
	)
}

func (c *Collate[T]) searchImmediateJS(searchID string) string {
	if c.action == "" {
		return ""
	}
	return fmt.Sprintf(
		"__ws.call('%s',{operation:'search',search:document.getElementById('%s').value,"+
			"page:1,limit:%d,order:'%s'})",
		escJS(c.action), escJS(searchID), c.limit, escJS(c.order),
	)
}

// applyJS collects all filter/sort state from the panel and sends it to the server.
func (c *Collate[T]) applyJS() string {
	if c.action == "" {
		return ""
	}
	panelID := c.id + "-panel"
	orderID := c.id + "-pending-order"
	searchID := c.id + "-search"

	return fmt.Sprintf(
		`(function(){`+
			`var panel=document.getElementById('%s');`+
			`var order=(document.getElementById('%s')||{}).value||'';`+
			`var search=(document.getElementById('%s')||{}).value||'';`+
			`var filters=[];`+
			// Collect bool filters (checkboxes)
			`panel.querySelectorAll('[data-filter-type="bool"]').forEach(function(el){`+
			`filters.push({field:el.getAttribute('data-filter-field'),type:'bool',bool:el.checked})`+
			`});`+
			// Collect date filters
			`panel.querySelectorAll('[data-filter-type="date-from"]').forEach(function(el){`+
			`var field=el.getAttribute('data-filter-field');`+
			`var toEl=panel.querySelector('[data-filter-type="date-to"]');`+
			`var fromVal=el.value||'';`+
			`var toVal='';`+
			// Find the matching to-input (next sibling container)
			`var toInput=el.closest('.flex.flex-col').querySelector('[data-filter-type="date-to"]');`+
			`if(toInput)toVal=toInput.value||'';`+
			`if(fromVal||toVal)filters.push({field:field,type:'date',from:fromVal,to:toVal})`+
			`});`+
			// Collect select filters
			`panel.querySelectorAll('[data-filter-type="select"]').forEach(function(el){`+
			`if(el.value)filters.push({field:el.getAttribute('data-filter-field'),type:'select',value:el.value})`+
			`});`+
			`panel.classList.add('hidden');`+
			`__ws.call('%s',{operation:'filter',search:search,page:1,limit:%d,order:order,filters:filters})`+
			`})()`,
		escJS(panelID), escJS(orderID), escJS(searchID),
		escJS(c.action), c.limit,
	)
}

func (c *Collate[T]) resetJS() string {
	if c.action == "" {
		return ""
	}
	panelID := c.id + "-panel"
	return fmt.Sprintf(
		`document.getElementById('%s').classList.add('hidden');`+
			`__ws.call('%s',{operation:'reset',page:1,limit:%d})`,
		escJS(panelID), escJS(c.action), c.limit,
	)
}

func (c *Collate[T]) exportJS() string {
	if c.action == "" {
		return ""
	}
	return fmt.Sprintf(
		"__ws.call('%s',{operation:'export',search:'%s',order:'%s'})",
		escJS(c.action), escJS(c.search), escJS(c.order),
	)
}

func (c *Collate[T]) loadMoreJS() string {
	if c.action == "" {
		return ""
	}
	return fmt.Sprintf(
		"var cp=parseInt(document.getElementById('%s').getAttribute('data-page'))||%d;"+
			"document.getElementById('%s').setAttribute('data-page',cp+1);"+
			"__ws.call('%s',{operation:'loadmore',search:'%s',page:cp+1,limit:%d,order:'%s'})",
		escJS(c.id), c.page,
		escJS(c.id),
		escJS(c.action), escJS(c.search), c.limit, escJS(c.order),
	)
}

// RenderFooter builds just the footer node (for replacing after load-more).
func (c *Collate[T]) RenderFooter() *Node {
	return c.renderFooter()
}

func (c *Collate[T]) resetPagingJS() string {
	if c.action == "" {
		return ""
	}
	return fmt.Sprintf(
		"document.getElementById('%s').setAttribute('data-page','1');"+
			"__ws.call('%s',{operation:'search',search:'%s',page:1,limit:%d,order:'%s'})",
		escJS(c.id),
		escJS(c.action), escJS(c.search), c.limit, escJS(c.order),
	)
}
