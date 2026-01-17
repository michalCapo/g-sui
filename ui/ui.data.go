package ui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// Note: This package uses SQLite-compatible SQL functions.
// For PostgreSQL-specific functions like unaccent(), consider using
// database-specific query builders or implementing diacritic handling in Go.

// Global flag to track if we've registered the function
var normalizeRegistered = false

// NormalizeForSearch normalizes a search term to handle diacritics and special characters
// This makes searches more user-friendly by matching accented characters
func NormalizeForSearch(search string) string {
	// Convert to lowercase first
	search = strings.ToLower(search)

	// Replace accented characters with their basic equivalents
	replacements := map[string]string{
		"á": "a", "ä": "a", "à": "a", "â": "a", "ã": "a", "å": "a", "æ": "ae",
		"č": "c", "ć": "c", "ç": "c",
		"ď": "d", "đ": "d",
		"é": "e", "ë": "e", "è": "e", "ê": "e", "ě": "e",
		"í": "i", "ï": "i", "ì": "i", "î": "i",
		"ľ": "l", "ĺ": "l", "ł": "l",
		"ň": "n", "ń": "n", "ñ": "n",
		"ó": "o", "ö": "o", "ò": "o", "ô": "o", "õ": "o", "ø": "o", "œ": "oe",
		"ř": "r", "ŕ": "r",
		"š": "s", "ś": "s", "ş": "s", "ș": "s",
		"ť": "t", "ț": "t",
		"ú": "u", "ü": "u", "ù": "u", "û": "u", "ů": "u",
		"ý": "y", "ÿ": "y",
		"ž": "z", "ź": "z", "ż": "z",
	}

	for accented, basic := range replacements {
		search = strings.ReplaceAll(search, accented, basic)
	}

	return search
}

// RegisterSQLiteNormalize registers a custom SQLite function 'normalize' for diacritic removal
// This function should be called after establishing the database connection
func RegisterSQLiteNormalize(db *gorm.DB) error {
	if normalizeRegistered {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// Get the SQLite connection with proper context
	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Register the custom normalize function
	err = conn.Raw(func(driverConn any) error {
		sqliteConn, ok := driverConn.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("connection is not sqlite3")
		}

		// Register the normalize function

		return sqliteConn.RegisterFunc("normalize", NormalizeForSearch, true)
	})

	if err == nil {
		normalizeRegistered = true

	} else {
		log.Printf("Failed to register normalize function: %v", err)
	}

	return err
}

// type TSort struct {
// 	Field string
// 	Text  string
// }

const (
	BOOL = iota
	// BOOL_NEGATIVE
	// BOOL_ZERO
	NOT_ZERO_DATE
	ZERO_DATE
	DATES
	SELECT
)

type TField struct {
	DB    string
	Field string
	Text  string

	Value     string
	As        uint
	Condition string
	Options   []AOption

	Bool bool
	// Value string
	Dates struct {
		From time.Time
		To   time.Time
	}
	// Options []AOption
}

var BOOL_ZERO_OPTIONS = []AOption{
	{
		ID:    "",
		Value: "All",
	},
	{
		ID:    "yes",
		Value: "On",
	},
	{
		ID:    "no",
		Value: "Off",
	},
}

type TQuery struct {
	Limit        int64
	Offset       int64
	Order        string
	PendingOrder string // Pending sort order that will be applied on "Apply" click
	Search       string
	Filter       []TField
}

// QueryHiddenFields generates hidden form fields for preserving query state
// including all filter values. This ensures filter state persists across form submissions.
func QueryHiddenFields(query *TQuery) string {
	var fields []string

	fields = append(fields, Hidden("Limit", query.Limit))
	fields = append(fields, Hidden("Offset", query.Offset))
	fields = append(fields, Hidden("Order", query.Order))
	fields = append(fields, Hidden("PendingOrder", query.PendingOrder))
	fields = append(fields, Hidden("Search", query.Search))
	fields = append(fields, FilterHiddenFields(query))

	return strings.Join(fields, "")
}

// FilterHiddenFields generates hidden form fields for filter state only.
// Use this when other query fields (Search, Order, etc.) are handled separately.
func FilterHiddenFields(query *TQuery) string {
	var fields []string

	for i, filter := range query.Filter {
		position := fmt.Sprintf("Filter[%d]", i)
		fields = append(fields, Hidden(position+".DB", filter.DB))
		fields = append(fields, Hidden(position+".Field", filter.Field))
		fields = append(fields, Hidden(position+".As", filter.As))
		fields = append(fields, Hidden(position+".Condition", filter.Condition))
		fields = append(fields, Hidden(position+".Value", filter.Value))
		fields = append(fields, Hidden(position+".Bool", filter.Bool))
		if !filter.Dates.From.IsZero() {
			fields = append(fields, Hidden(position+".Dates.From", filter.Dates.From.Format("2006-01-02")))
		}
		if !filter.Dates.To.IsZero() {
			fields = append(fields, Hidden(position+".Dates.To", filter.Dates.To.Format("2006-01-02")))
		}
	}

	return strings.Join(fields, "")
}

type TCollateResult[T any] struct {
	Total    int64
	Filtered int64
	Data     []T
	Query    *TQuery
}

// CollateColors holds all color-related CSS classes for theming collate components.
// Use predefined color sets (CollateBlue, CollateGreen, etc.) or create custom ones.
type CollateColors struct {
	Button        string // Button color constant (e.g., Blue, Green)
	ButtonOutline string // Outline button color constant (e.g., BlueOutline, GreenOutline)
	ActiveBg      string // Active state background (e.g., "bg-blue-600")
	ActiveBorder  string // Active state border (e.g., "border-blue-600")
	ActiveHover   string // Active state hover (e.g., "hover:bg-blue-700")
}

// Predefined color schemes for collate components
var (
	CollateBlue = CollateColors{
		Button:        Blue,
		ButtonOutline: BlueOutline,
		ActiveBg:      "bg-blue-800",
		ActiveBorder:  "border-blue-600",
		ActiveHover:   "hover:bg-blue-700",
	}
	CollateGreen = CollateColors{
		Button:        Green,
		ButtonOutline: GreenOutline,
		ActiveBg:      "bg-green-600",
		ActiveBorder:  "border-green-600",
		ActiveHover:   "hover:bg-green-700",
	}
	CollatePurple = CollateColors{
		Button:        Purple,
		ButtonOutline: PurpleOutline,
		ActiveBg:      "bg-purple-500",
		ActiveBorder:  "border-purple-500",
		ActiveHover:   "hover:bg-purple-700",
	}
	CollateRed = CollateColors{
		Button:        Red,
		ButtonOutline: RedOutline,
		ActiveBg:      "bg-red-600",
		ActiveBorder:  "border-red-600",
		ActiveHover:   "hover:bg-red-700",
	}
	CollateYellow = CollateColors{
		Button:        Yellow,
		ButtonOutline: YellowOutline,
		ActiveBg:      "bg-yellow-400",
		ActiveBorder:  "border-yellow-400",
		ActiveHover:   "hover:bg-yellow-500",
	}
	CollateGray = CollateColors{
		Button:        Gray,
		ButtonOutline: GrayOutline,
		ActiveBg:      "bg-gray-600",
		ActiveBorder:  "border-gray-600",
		ActiveHover:   "hover:bg-gray-700",
	}
)

type collate[T any] struct {
	Init          *TQuery
	Target        Attr
	TargetFilter  Attr
	Database      *gorm.DB
	SearchFields  []TField
	SortFields    []TField
	FilterFields  []TField
	ExcelFields   []TField
	OnRow         func(*T, int) string
	OnExcel       func(*[]T) (string, io.Reader, error)
	Colors        CollateColors
	OnEmpty       func(*Context) string
	IconEmpty     string
	TextEmpty     string
	ActionEmpty   string
	OnActionEmpty func(*Context, Attr) string
}

// Collate constructs a new collate with sensible defaults using the provided init query.
// Default color scheme is blue.
func Collate[T any](init *TQuery) *collate[T] {
	return &collate[T]{
		Init:         init,
		Target:       Target(),
		TargetFilter: Target(),
		Colors:       CollateBlue,
	}
}

// SetColor sets the color scheme for the collate component.
// Use predefined schemes: CollateBlue, CollateGreen, CollatePurple, CollateRed, CollateYellow, CollateGray.
func (c *collate[T]) SetColor(colors CollateColors) *collate[T] {
	c.Colors = colors
	return c
}

func (collate *collate[T]) onXLS(ctx *Context) string {
	// Set query for all records
	query := makeQuery(collate.Init)
	err := ctx.Body(query)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	query.Limit = 1000000
	result := collate.Load(query)

	var filename string
	var reader io.Reader

	if collate.OnExcel != nil {
		var err error

		filename, reader, err = collate.OnExcel(&result.Data)
		if err != nil {
			log.Printf("Error: %v", err)
			return "Error generating Excel file"
		}
	} else {
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("Error closing Excel file: %v", err)
			}
		}()

		for i, header := range collate.ExcelFields {
			if header.Text == "" {
				header.Text = header.Field
			}

			cell := string(rune('A'+i)) + "1"
			f.SetCellValue("Sheet1", cell, header.Text)
		}

		styleDate, err := f.NewStyle(&excelize.Style{NumFmt: 14})
		if err != nil {
			log.Printf("Error: %v", err)
		}

		// Write data rows
		for rowIndex, item := range result.Data {
			v := reflect.ValueOf(item)

			for colIndex, header := range collate.ExcelFields {
				col := string(rune('A' + colIndex))
				cell := col + strconv.Itoa(rowIndex+2)
				value := v.FieldByName(header.Field).Interface()
				typ := v.FieldByName(header.Field).Type().String()

				switch typ {
				case "time.Time":
					if !value.(time.Time).IsZero() {
						// value = value.(time.Time).Format("2006-01-02")

						f.SetCellValue("Sheet1", cell, value)
						f.SetCellStyle("Sheet1", cell, cell, styleDate)

						f.SetColWidth("Sheet1", col, col, 15)
					}

				default:
					f.SetCellValue("Sheet1", cell, value)
				}
			}
		}

		// Set filename with timestamp
		filename = fmt.Sprintf("export_%s.xlsx", time.Now().Format("20060102_150405"))

		fileBytes, err := f.WriteToBuffer()
		if err != nil {
			return "Error generating Excel file"
		}

		reader = io.Reader(bytes.NewReader(fileBytes.Bytes()))
	}

	ctx.DownloadAs(&reader, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", filename)

	return ""
}

func (collate *collate[T]) onResize(ctx *Context) string {
	query := makeQuery(collate.Init)

	body := &TQuery{}
	err := ctx.Body(body)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Preserve all state from the request body
	query.Offset = body.Offset
	query.Order = body.Order
	query.PendingOrder = body.PendingOrder
	query.Filter = body.Filter
	query.Search = body.Search

	// Double the limit, with fallback to init value
	if body.Limit > 0 {
		query.Limit = body.Limit * 2
	} else {
		query.Limit = collate.Init.Limit * 2
	}

	return collate.ui(ctx, query)
}

func (collate *collate[T]) onSort(ctx *Context) string {
	query := makeQuery(collate.Init)

	body := &TQuery{}
	err := ctx.Body(body)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Preserve all state from the request body
	query.Limit = body.Limit
	query.Offset = body.Offset
	query.Order = body.Order
	query.PendingOrder = body.PendingOrder
	query.Filter = body.Filter
	query.Search = body.Search

	// Ensure valid limit
	if query.Limit <= 0 {
		query.Limit = collate.Init.Limit
	}

	return collate.ui(ctx, query)
}

func (collate *collate[T]) onSearch(ctx *Context) string {
	query := makeQuery(collate.Init)

	err := ctx.Body(query)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	// Apply pending order when Apply button is clicked
	if query.PendingOrder != "" {
		query.Order = query.PendingOrder
	}

	return collate.ui(ctx, query)
}

func (collate *collate[T]) onReset(ctx *Context) string {
	query := makeQuery(collate.Init)

	return collate.ui(ctx, query)
}

// Search sets searchable fields.
func (c *collate[T]) Search(fields ...TField) *collate[T] {
	c.SearchFields = fields
	return c
}

// Sort sets sortable fields.
func (c *collate[T]) Sort(fields ...TField) *collate[T] {
	c.SortFields = fields
	return c
}

// Filter sets filterable fields.
func (c *collate[T]) Filter(fields ...TField) *collate[T] {
	c.FilterFields = fields
	return c
}

// Excel sets fields to be exported to Excel.
func (c *collate[T]) Excel(fields ...TField) *collate[T] {
	c.ExcelFields = fields
	return c
}

// Row sets the row rendering function.
func (c *collate[T]) Row(fn func(*T, int) string) *collate[T] {
	c.OnRow = fn
	return c
}

// Empty sets a custom empty state renderer.
func (c *collate[T]) Empty(fn func(*Context) string) *collate[T] {
	c.OnEmpty = fn
	return c
}

// EmptyIcon sets the icon for the default empty state.
func (c *collate[T]) EmptyIcon(icon string) *collate[T] {
	c.IconEmpty = icon
	return c
}

// EmptyText sets the text for the default empty state.
func (c *collate[T]) EmptyText(text string) *collate[T] {
	c.TextEmpty = text
	return c
}

// EmptyAction sets the action button for the default empty state.
// The function receives both context and target, allowing it to create proper click handlers.
func (c *collate[T]) EmptyAction(text string, fn func(*Context, Attr) string) *collate[T] {
	c.ActionEmpty = text
	c.OnActionEmpty = fn
	return c
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

// validateFieldName checks if a field name is in the allowlist of configured fields
func (c *collate[T]) validateFieldName(fieldName string) bool {
	// Check against SearchFields
	for _, field := range c.SearchFields {
		if field.DB == fieldName || field.Field == fieldName {
			return true
		}
	}
	// Check against FilterFields
	for _, field := range c.FilterFields {
		if field.DB == fieldName || field.Field == fieldName {
			return true
		}
	}
	// Check against SortFields
	for _, field := range c.SortFields {
		if field.DB == fieldName || field.Field == fieldName {
			return true
		}
	}
	return false
}

func (c *collate[T]) Load(query *TQuery) *TCollateResult[T] {
	result := &TCollateResult[T]{
		Total:    0,
		Filtered: 0,
		Data:     []T{},
		Query:    query,
	}

	c.Database.Model(result.Data).Count(&result.Total)

	temp := c.Database.Model(&result.Data).
		Session(&gorm.Session{}).
		Order(query.Order).
		Limit(int(query.Limit)).
		Offset(int(query.Offset))

	// Apply filters using parameterized queries
	for _, filter := range query.Filter {
		if filter.DB == "" {
			filter.DB = filter.Field
		}

		// Validate field name against allowlist
		if !c.validateFieldName(filter.DB) {
			fmt.Printf("WARNING: Rejecting invalid field name: %s\n", filter.DB)
			continue
		}

		if filter.As == BOOL && filter.Bool && filter.Condition != "" {
			// Validate condition contains only safe operators
			if strings.Contains(filter.Condition, " = 1") || strings.Contains(filter.Condition, " = 0") ||
				strings.Contains(filter.Condition, " IS NULL") || strings.Contains(filter.Condition, " IS NOT NULL") {
				temp = temp.Where(filter.DB + filter.Condition)
			} else {
				fmt.Printf("WARNING: Rejecting unsafe condition: %s\n", filter.Condition)
			}
			continue
		}

		if filter.As == BOOL && filter.Bool {
			temp = temp.Where(filter.DB+" = ?", 1)
		}

		if filter.As == ZERO_DATE && filter.Bool {
			temp = temp.Where(filter.DB+" <= ?", "0001-01-01 00:00:00+00:00")
		}

		if filter.As == NOT_ZERO_DATE && filter.Bool {
			temp = temp.Where(filter.DB+" > ?", "0001-01-01 00:00:00+00:00")
		}

		if filter.As == DATES && !filter.Dates.From.IsZero() {
			temp = temp.Where(filter.DB+" >= ?", startOfDay(filter.Dates.From))
		}

		if filter.As == DATES && !filter.Dates.To.IsZero() {
			temp = temp.Where(filter.DB+" <= ?", endOfDay(filter.Dates.To))
		}

		if filter.As == SELECT && filter.Value != "" {
			temp = temp.Where(filter.DB+" = ?", filter.Value)
		}
	}

	// Apply search using parameterized queries
	if len(query.Search) > 0 {
		// Normalize search term to handle accented characters
		normalizedSearch := NormalizeForSearch(query.Search)

		var searchConditions []string
		var searchArgs []any

		for _, field := range c.SearchFields {
			if field.DB == "" {
				field.DB = field.Field
			}

			// Validate field name against allowlist
			if !c.validateFieldName(field.DB) {
				fmt.Printf("WARNING: Rejecting invalid search field: %s\n", field.DB)
				continue
			}

			// Primary approach: Use custom normalize function with parameterized query
			// Note: We still need to build the CAST part as a string since it's a SQL function,
			// but the search value is parameterized
			searchConditions = append(searchConditions, "normalize(CAST("+field.DB+" AS TEXT)) LIKE ?")
			searchArgs = append(searchArgs, "%"+normalizedSearch+"%")

			// Fallback approach: Simple case-insensitive search with parameterized query
			searchConditions = append(searchConditions, "LOWER(CAST("+field.DB+" AS TEXT)) LIKE ?")
			searchArgs = append(searchArgs, "%"+strings.ToLower(query.Search)+"%")
		}

		if len(searchConditions) > 0 {
			whereClause := "(" + strings.Join(searchConditions, " OR ") + ")"
			temp = temp.Where(whereClause, searchArgs...)
		}
	}

	temp.Count(&result.Filtered)
	temp.Find(&result.Data)

	return result
}

// Render is the public entry point used by pages. It binds the database,
// prepares the query from Init and renders the collated UI.
func (collate *collate[T]) Render(ctx *Context, database *gorm.DB) string {
	collate.Database = database

	// Establish base query
	query := makeQuery(collate.Init)

	return collate.ui(ctx, query)
}

// ui executes the query and renders the UI. Internal use only.
func (collate *collate[T]) ui(ctx *Context, query *TQuery) string {
	result := collate.Load(query)

	return Div("flex flex-col gap-2 mt-2", collate.Target)(
		Div("flex flex-col")(
			Header(ctx, collate, query),
			Filtering(ctx, collate, query),
		),
		Map(result.Data, collate.OnRow),
		Paging(ctx, collate, result),
	)
}

func makeQuery(def *TQuery) *TQuery {
	if def.Offset < 0 {
		def.Offset = 0
	}

	if def.Limit <= 0 {
		def.Limit = 10
	}

	query := &TQuery{
		Limit:        def.Limit,
		Offset:       def.Offset,
		Order:        def.Order,
		PendingOrder: def.PendingOrder,
		Search:       def.Search,
		Filter:       def.Filter,
	}

	// Initialize PendingOrder to Order if not set
	if query.PendingOrder == "" {
		query.PendingOrder = query.Order
	}

	return query
}

// renderEmpty renders the empty state when no records are found.
func (collate *collate[T]) renderEmpty(ctx *Context, result *TCollateResult[T]) string {
	if collate.OnEmpty != nil {
		return collate.OnEmpty(ctx)
	}

	icon := collate.IconEmpty
	if icon == "" {
		icon = "fa fa-inbox" // Generic empty state icon
	}

	title := collate.TextEmpty
	if title == "" {
		if result.Total == 0 {
			title = "No records found"
		} else {
			title = "No records found for the selected filter"
		}
	}

	emptyStateContent := []string{
		Div("text-gray-300 text-7xl mb-6")(Icon(icon)),
		Div("text-gray-600 text-xl font-medium mb-6 text-center")(title),
	}

	// Only add action button if both text and handler are provided
	if collate.ActionEmpty != "" && collate.OnActionEmpty != nil {
		emptyStateContent = append(emptyStateContent,
			Button().
				Class("rounded-lg px-6 h-12 font-bold").
				Color(Black).
				Click(collate.OnActionEmpty(ctx, collate.Target)).
				Render(IconLeft("fa fa-plus", collate.ActionEmpty)),
		)
	}

	return Div("mt-2 py-20 border-2 border-dashed border-gray-200 rounded-xl flex flex-col items-center justify-center bg-white")(
		emptyStateContent...,
	)
}

func Filtering[T any](ctx *Context, collate *collate[T], query *TQuery) string {
	if len(collate.FilterFields) == 0 && len(collate.SortFields) == 0 {
		return ""
	}

	// Calculate dynamic width based on content
	// Base width for minimal content, then add for each filter/sort field
	totalFields := len(collate.FilterFields) + len(collate.SortFields)
	widthClass := "w-96" // default width for minimal content

	if totalFields > 8 {
		widthClass = "w-[38rem]" // extra large for many fields
	} else if totalFields > 5 {
		widthClass = "w-[28rem]" // large for several fields
	} else if totalFields > 2 {
		widthClass = "w-96" // medium for a few fields
	} else {
		widthClass = "w-[22rem]" // small for minimal fields
	}

	return Div("col-span-2 relative h-0 hidden z-20", collate.TargetFilter)(
		Div(fmt.Sprintf("absolute top-2 right-0 rounded-xl bg-white border shadow-2xl p-4 %s", widthClass))(
			// Header with title and close button
			Div("flex items-center justify-between mb-2")(
				Div("text-sm font-semibold text-gray-700")("Filters & Options"),
				Button().
					Class("rounded-full w-9 h-9 border bg-white hover:bg-gray-50 flex items-center justify-center").
					Click(fmt.Sprintf("window.document.getElementById('%s')?.classList.toggle('hidden');", collate.TargetFilter.ID)).
					Render(Icon("fa fa-fw fa-times")),
			),

			Form("flex flex-col", ctx.Submit(collate.onSearch).Replace(collate.Target))(
				Hidden("Search", query.Search),
				Hidden("PendingOrder", query.PendingOrder),

				// Sort section
				Iff(len(collate.SortFields) > 0)(
					Div("flex flex-col gap-2 mb-3")(
						Div("text-xs font-bold text-gray-600 mb-1")("Sort By"),
						Div("flex flex-wrap gap-1", Attr{ID: "sort-buttons-container"})(
							Map(collate.SortFields, func(sort *TField, index int) string {
								if sort.DB == "" {
									sort.DB = sort.Field
								}

								// Use PendingOrder for visual state instead of Order
								direction := ""
								pendingOrderStr := query.PendingOrder
								if pendingOrderStr == "" {
									pendingOrderStr = query.Order
								}

								// Parse order string: "field asc" or "field desc"
								orderParts := strings.Fields(strings.TrimSpace(pendingOrderStr))
								if len(orderParts) >= 2 {
									orderField := strings.ToLower(orderParts[0])
									orderDir := strings.ToLower(orderParts[1])
									if orderField == strings.ToLower(sort.DB) {
										if orderDir == "asc" {
											direction = "asc"
										} else if orderDir == "desc" {
											direction = "desc"
										}
									}
								}

								// Each button will have a unique ID
								btnID := fmt.Sprintf("sort-btn-%s", sort.DB)
								iconID := fmt.Sprintf("sort-icon-%s", sort.DB)

								// JavaScript to cycle through: none -> asc -> desc -> none
								// Use configured colors for active state
								activeClass := fmt.Sprintf("rounded text-sm %s %s text-white %s cursor-pointer font-bold text-center select-none p-3 flex items-center justify-center",
									collate.Colors.ActiveBg, collate.Colors.ActiveBorder, collate.Colors.ActiveHover)
								inactiveClass := "rounded text-sm border border-gray-300 bg-white text-gray-700 hover:bg-gray-50 cursor-pointer font-bold text-center select-none p-3 flex items-center justify-center"

								jsUpdateOrder := fmt.Sprintf(
									`(function(){
var form=document.getElementById('sort-btn-%s').closest('form');
if(!form){console.log('no form');return;}
var hidden=form.querySelector('input[name="PendingOrder"]');
if(!hidden){console.log('no hidden');return;}

var field='%s';
var current=(hidden.value||'').trim();
var parts=current.split(/\s+/);
var currentField=parts[0]||'';
var currentDir=(parts[1]||'').toLowerCase();
var newOrder='';
var newDir='';

if(currentField.toLowerCase()===field.toLowerCase()){
	if(currentDir==='asc'){
		newOrder=field+' desc';
		newDir='desc';
	}else if(currentDir==='desc'){
		newOrder='';
		newDir='';
	}else{
		newOrder=field+' asc';
		newDir='asc';
	}
}else{
	newOrder=field+' asc';
	newDir='asc';
}
hidden.value=newOrder;

var allBtns=form.querySelectorAll('[id^="sort-btn-"]');
for(var i=0;i<allBtns.length;i++){
	var w=allBtns[i];
	var f=w.id.replace('sort-btn-','');
	var ic=document.getElementById('sort-icon-'+f);
	var bt=w.querySelector('[onclick]');
	if(!bt)continue;
	
	var isActive=(f.toLowerCase()===field.toLowerCase() && newDir!=='');
	var dir=(f.toLowerCase()===field.toLowerCase())?newDir:'';
	
	if(ic){
		if(dir==='asc'){
			ic.className='fa fa-fw fa-sort-amount-asc';
		}else if(dir==='desc'){
			ic.className='fa fa-fw fa-sort-amount-desc';
		}else{
			ic.className='fa fa-fw fa-sort';
		}
	}
	
	if(isActive){
		bt.className='%s';
	}else{
		bt.className='%s';
	}
}
})();`,
									sort.DB, sort.DB, activeClass, inactiveClass,
								)

								buttonClass := inactiveClass
								if direction == "asc" || direction == "desc" {
									buttonClass = activeClass
								}

								iconClass := "fa fa-fw fa-sort"
								if direction == "asc" {
									iconClass = "fa fa-fw fa-sort-amount-asc"
								} else if direction == "desc" {
									iconClass = "fa fa-fw fa-sort-amount-desc"
								}

								return Div("", Attr{ID: btnID})(
									Button().
										Class(buttonClass).
										Click(jsUpdateOrder).
										Render(
											Div("flex gap-2 items-center")(
												I(iconClass, Attr{ID: iconID})(),
												sort.Text,
											),
										),
								)
							}),
						),
					),
				),

				// Filters section
				Iff(len(collate.FilterFields) > 0)(
					Div("flex flex-col gap-2 mt-2 pt-3 border-t border-gray-200")(
						Div("text-xs font-bold text-gray-600 mb-1")("Filters"),
						Map2(collate.FilterFields, func(item TField, index int) []string {
							if item.DB == "" {
								item.DB = item.Field
							}

							position := fmt.Sprintf("Filter[%d]", index)

							return []string{
								Iff(item.As == ZERO_DATE)(
									Div("flex items-center")(
										Hidden(position+".Field", item.DB),
										Hidden(position+".As", item.As),
										ICheckbox(position+".Bool", query).Render(item.Text),
									),
								),

								Iff(item.As == NOT_ZERO_DATE)(
									Div("flex items-center")(
										Hidden(position+".Field", item.DB),
										Hidden(position+".As", item.As),
										ICheckbox(position+".Bool", query).Render(item.Text),
									),
								),

								Iff(item.As == DATES)(
									Div("")(
										Label(nil).Class("text-xs mt-1 font-bold").Render(item.Text),
										Div("grid grid-cols-2 gap-2")(
											Hidden(position+".Field", item.DB),
											Hidden(position+".As", item.As),
											IDate(position+".Dates.From", query).Class("").Render("From"),
											IDate(position+".Dates.To", query).Class("").Render("To"),
										),
									),
								),

								Iff(item.As == BOOL)(
									Div("flex items-center")(
										Hidden(position+".Field", item.DB),
										Hidden(position+".As", item.As),
										Hidden(position+".Condition", item.Condition),
										ICheckbox(position+".Bool", query).Render(item.Text),
									),
								),

								Iff(item.As == SELECT && len(item.Options) > 0)(
									Div("")(
										Hidden(position+".Field", item.DB),
										Hidden(position+".As", item.As),
										ISelect(position+".Value", query).
											Class("flex-1").
											Options(item.Options).
											Render(item.Text),
									),
								),
							}
						}),
					),
				),

				// Footer actions
				Div("flex items-center justify-between mt-4 pt-3 border-t border-gray-200")(
					Button().
						Color(White).
						Class("flex items-center gap-2 rounded-full px-4 h-10 border border-gray-300 bg-white hover:bg-gray-50").
						Click(ctx.Call(collate.onReset).Replace(collate.Target)).
						Render(IconLeft("fa fa-fw fa-undo", "Reset")),

					Button().
						Submit().
						Class("flex items-center gap-2 rounded-full px-4 h-10").
						Color(collate.Colors.Button).
						Render(IconLeft("fa fa-fw fa-check", "Apply")),
				),
			),
		),
	)
}

func Header[T any](ctx *Context, collate *collate[T], query *TQuery) string {
	// Build form class conditionally using configured color
	formClass := fmt.Sprintf("flex %s rounded", collate.Colors.ActiveBg)

	return Div("flex w-full")(
		// Excel button at the start of the row
		If(len(collate.ExcelFields) > 0 || collate.OnExcel != nil, func() string {
			return Button().
				Class("rounded shadow").
				Color(collate.Colors.Button).
				Click(ctx.Call(collate.onXLS, query).None()).
				Render(IconLeft("fa fa-download", "Export"))
		}),

		Flex1,

		If(collate.SearchFields != nil, func() string {
			return Form(formClass, ctx.Submit(collate.onSearch).Replace(collate.Target))(
				// Preserve current filter state using hidden fields
				Hidden("Limit", query.Limit),
				Hidden("Offset", query.Offset),
				Hidden("Order", query.Order),
				Hidden("PendingOrder", query.PendingOrder),
				FilterHiddenFields(query),

				IText("Search", query).
					Class("p-px").
					ClassInput("cursor-pointer bg-white border-gray-300 hover:border-blue-500 block w-full p-3").
					Placeholder(ctx.Translate("Search")).
					Render(""),

				Button().
					Submit().
					Class("shadow").
					Color(collate.Colors.Button).
					Render(Icon("fa fa-fw fa-search")),
			)
		}),

		If(len(collate.FilterFields) > 0 || len(collate.SortFields) > 0, func() string {
			return Button().
				Submit().
				Class("rounded-r shadow").
				Color(collate.Colors.Button).
				Click(fmt.Sprintf("window.document.getElementById('%s')?.classList.toggle('hidden');", collate.TargetFilter.ID)).
				Render(IconLeft("fa fa-fw fa-sliders", "Filter"))
		}),
	)
}

func Sorting[T any](ctx *Context, collate *collate[T], query *TQuery) string {
	if len(collate.SortFields) == 0 {
		return ""
	}

	return Div("flex gap-px")(
		Map(collate.SortFields, func(sort *TField, index int) string {
			if sort.DB == "" {
				sort.DB = sort.Field
			}

			direction := ""
			color := GrayOutline
			field := strings.ToLower(sort.DB)
			orderStr := strings.ToLower(query.Order)

			if strings.Contains(orderStr, field) {
				if strings.Contains(orderStr, "asc") {
					direction = "asc"
				} else {
					direction = "desc"
				}

				color = collate.Colors.Button
			}

			reverse := "desc"

			if direction == "desc" {
				reverse = "asc"
			}

			// Create a copy of the current query with updated order to preserve filters and search
			sortQuery := *query
			sortQuery.Order = sort.DB + " " + reverse

			// Use form to properly preserve filter state across sort changes
			return Form("inline-flex", ctx.Submit(collate.onSort).Replace(collate.Target))(
				QueryHiddenFields(&sortQuery),
				Button().
					Submit().
					Class("rounded bg-white").
					Color(color).
					Render(
						Div("flex gap-2 items-center")(
							Iff(direction == "asc")(Icon("fa fa-fw fa-sort-amount-asc")),
							Iff(direction == "desc")(Icon("fa fa-fw fa-sort-amount-desc")),
							Iff(direction == "")(Icon("fa fa-fw fa-sort")),
							sort.Text,
						),
					),
			)
		}),
	)
}

func Paging[T any](ctx *Context, collate *collate[T], result *TCollateResult[T]) string {
	if result.Filtered == 0 {
		return collate.renderEmpty(ctx, result)
	}

	size := len(result.Data)
	more := ctx.Translate("Load more items")
	count := ctx.Translate("Showing %d / %d of %d in total", size, result.Filtered, result.Total)

	if result.Filtered == result.Total {
		count = ctx.Translate("Showing %d / %d", size, result.Total)
	}

	return Div("flex items-center justify-center")(
		// showing information
		Div("mx-4 font-bold text-lg")(count),

		Div("flex gap-px flex-1 justify-end")(
			// reset
			Button().
				Class("bg-white rounded-l").
				Color(collate.Colors.ButtonOutline).
				Disabled(size == 0 || size <= int(collate.Init.Limit)).
				Click(ctx.Call(collate.onReset).Replace(collate.Target)).
				Render(
					Icon("fa fa-fw fa-undo"),
				),

			// load more - use form to properly preserve filter state
			Form("inline-flex", ctx.Submit(collate.onResize).Replace(collate.Target))(
				QueryHiddenFields(result.Query),
				Button().
					Submit().
					Class("rounded-r bg-white").
					Color(collate.Colors.ButtonOutline).
					Disabled(size >= int(result.Filtered)).
					Render(
						Div("flex gap-2 items-center")(
							Icon("fa fa-arrow-down"), more,
						),
					),
			),
		),
	)
}
