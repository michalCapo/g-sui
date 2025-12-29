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
	Limit  int64
	Offset int64
	Order  string
	Search string
	Filter []TField
}

type TCollateResult[T any] struct {
	Total    int64
	Filtered int64
	Data     []T
	Query    *TQuery
}

type collate[T any] struct {
	Init         *TQuery
	Target       Attr
	TargetFilter Attr
	Database     *gorm.DB
	SearchFields []TField
	SortFields   []TField
	FilterFields []TField
	ExcelFields  []TField
	OnRow        func(*T, int) string
	OnExcel      func(*[]T) (string, io.Reader, error)
}

// Collate constructs a new collate with sensible defaults using the provided init query.
func Collate[T any](init *TQuery) *collate[T] {
	return &collate[T]{
		Init:         init,
		Target:       Target(),
		TargetFilter: Target(),
	}
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

	err := ctx.Body(query)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	query.Limit *= 2

	return collate.ui(ctx, query)
}

func (collate *collate[T]) onSort(ctx *Context) string {
	query := makeQuery(collate.Init)

	body := &TQuery{}
	err := ctx.Body(body)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	query.Order = body.Order

	return collate.ui(ctx, query)
}

func (collate *collate[T]) onSearch(ctx *Context) string {
	query := makeQuery(collate.Init)

	err := ctx.Body(query)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	return collate.ui(ctx, query)
}

func (collate *collate[T]) onReset(ctx *Context) string {
	query := makeQuery(collate.Init)

	return collate.ui(ctx, query)
}

// Search sets searchable fields.
func (c *collate[T]) Search(fields ...TField) { c.SearchFields = fields }

// Sort sets sortable fields.
func (c *collate[T]) Sort(fields ...TField) { c.SortFields = fields }

// Filter sets filterable fields.
func (c *collate[T]) Filter(fields ...TField) { c.FilterFields = fields }

// Excel sets fields to be exported to Excel.
func (c *collate[T]) Excel(fields ...TField) { c.ExcelFields = fields }

// Row sets the row rendering function.
func (c *collate[T]) Row(fn func(*T, int) string) { c.OnRow = fn }

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
			Div("flex gap-x-2")(
				Sorting(ctx, collate, query),
				Flex1,
				Searching(ctx, collate, query),
			),
			Div("flex justify-end")(
				Filtering(ctx, collate, query),
			),
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
		Limit:  def.Limit,
		Offset: def.Offset,
		Order:  def.Order,
		Search: def.Search,
		Filter: def.Filter,
	}

	return query
}

func Empty[T any](result *TCollateResult[T]) string {
	if result.Total == 0 {
		return Div("mt-2 py-24 rounded text-xl flex justify-center items-center bg-white rounded-lg")(
			Div("")(
				Div("text-black text-2xl p-4 mb-2 font-bold flex justify-center items-center")(("No records found")),
			),
		)
	}

	if result.Filtered == 0 {
		return Div("mt-2 py-24 rounded text-xl flex justify-center items-center bg-white rounded-lg")(
			Div("flex gap-x-px items-center justify-center text-2xl")(
				Icon("fa fa-fw fa-exclamation-triangle text-yellow-500"),
				Div("text-black p-4 mb-2 font-bold flex justify-center items-center")("No records found for the selected filter"),
			),
		)
	}

	return ""
}

func Filtering[T any](ctx *Context, collate *collate[T], query *TQuery) string {
	if len(collate.FilterFields) == 0 {
		return ""
	}

	// c.Query = Query(def)
	// ctx.Session(database, name, c.Query)

	return Div("col-span-2 relative h-0 hidden z-20", collate.TargetFilter)(
		Div("absolute top-2 right-0 w-96 rounded-xl bg-white border border-gray-200 shadow-2xl p-4")(
			// Header with title and close button
			Div("flex items-center justify-between mb-2")(
				Div("text-sm font-semibold text-gray-700")("Filters"),
				Button().
					Class("rounded-full w-9 h-9 border border-gray-200 bg-white hover:bg-gray-50 flex items-center justify-center").
					Click(fmt.Sprintf("window.document.getElementById('%s')?.classList.toggle('hidden');", collate.TargetFilter.ID)).
					Render(Icon("fa fa-fw fa-times")),
			),

			Form("flex flex-col", ctx.Submit(collate.onSearch).Replace(collate.Target))(
				Hidden("Search", "string", query.Search),

				// Filters content
				Div("flex flex-col gap-2 mt-2")(
					Map2(collate.FilterFields, func(item TField, index int) []string {
						if item.DB == "" {
							item.DB = item.Field
						}

						position := fmt.Sprintf("Filter[%d]", index)

						return []string{
							Iff(item.As == ZERO_DATE)(
								Div("flex items-center")(
									Hidden(position+".Field", "string", item.DB),
									Hidden(position+".As", "uint", item.As),
									ICheckbox(position+".Bool", query).Render(item.Text),
								),
							),

							Iff(item.As == NOT_ZERO_DATE)(
								Div("flex items-center")(
									Hidden(position+".Field", "string", item.DB),
									Hidden(position+".As", "uint", item.As),
									ICheckbox(position+".Bool", query).Render(item.Text),
								),
							),

							Iff(item.As == DATES)(
								Div("")(
									Label(nil).Class("text-xs mt-1 font-bold").Render(item.Text),
									Div("grid grid-cols-2 gap-2")(
										Hidden(position+".Field", "string", item.DB),
										Hidden(position+".As", "uint", item.As),
										IDate(position+".Dates.From", query).Class("").Render("From"),
										IDate(position+".Dates.To", query).Class("").Render("To"),
									),
								),
							),

							Iff(item.As == BOOL)(
								Div("flex items-center")(
									Hidden(position+".Field", "string", item.DB),
									Hidden(position+".As", "uint", item.As),
									Hidden(position+".Condition", "string", item.Condition),
									ICheckbox(position+".Bool", query).Render(item.Text),
								),
							),

							Iff(item.As == SELECT && len(item.Options) > 0)(
								Div("")(
									Hidden(position+".Field", "string", item.DB),
									Hidden(position+".As", "uint", item.As),
									ISelect(position+".Value", query).
										Class("flex-1").
										Options(item.Options).
										Render(item.Text),
								),
							),
						}
					}),
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
						Class("flex items-center gap-2 rounded-lg px-4 h-10").
						Color(Purple).
						Render(IconLeft("fa fa-fw fa-check", "Apply")),
				),
			),
		),
	)
}

func Searching[T any](ctx *Context, collate *collate[T], query *TQuery) string {
	if collate.SearchFields == nil {
		return ""
	}

	// reset := TQuery{
	// 	Search: "",
	// 	Filter: query.Filter,
	// 	Order:  query.Order,
	// 	Offset: query.Offset,
	// 	Limit:  query.Limit,
	// }

	return Div("flex-1 xl:flex gap-px hidden")(

		// Button().
		// 	Class("rounded shadow bg-white").
		// 	Color(Blue).
		// 	Click(ctx.Call(collate.onSearch, reset).Replace(collate.Target)).
		// 	Render(Icon("fa fa-times")),

		Form("flex-1 flex bg-blue-800 rounded-l-lg", ctx.Submit(collate.onSearch).Replace(collate.Target))(
			Map2(collate.FilterFields, func(item TField, index int) []string {
				if item.DB == "" {
					item.DB = item.Field
				}

				position := fmt.Sprintf("Filter[%d]", index)

				return []string{
					Iff(item.As == ZERO_DATE)(
						Hidden(position+".Field", "string", item.DB),
						Hidden(position+".As", "uint", item.As),
						Hidden(position+".Value", "string", item.Value),
					),

					Iff(item.As == NOT_ZERO_DATE)(
						Hidden(position+".Field", "string", item.DB),
						Hidden(position+".As", "uint", item.As),
						Hidden(position+".Value", "string", item.Value),
					),

					Iff(item.As == DATES)(
						Hidden(position+".Field", "string", item.DB),
						Hidden(position+".As", "uint", item.As),
						Hidden(position+".Value", "string", item.Value),
					),

					Iff(item.As == BOOL)(
						Hidden(position+".Field", "string", item.DB),
						Hidden(position+".As", "uint", item.As),
						Hidden(position+".Condition", "string", item.Condition),
						Hidden(position+".Value", "string", item.Value),
					),

					Iff(item.As == SELECT && len(item.Options) > 0)(
						Hidden(position+".Field", "string", item.DB),
						Hidden(position+".As", "uint", item.As),
						Hidden(position+".Value", "string", item.Value),
					),
				}
			}),

			IText("Search", query).
				Class("flex-1 p-1 w-72").
				ClassInput("cursor-pointer bg-white border-gray-300 hover:border-blue-500 block w-full p-3").
				Placeholder(ctx.Translate("Search")).
				Render(""),

			Button().
				Submit().
				Class("rounded shadow bg-white").
				Color(Blue).
				Render(Icon("fa fa-fw fa-search")),
		),

		If(len(collate.ExcelFields) > 0 || collate.OnExcel != nil, func() string {
			return Button().
				Color(Blue).
				Click(ctx.Call(collate.onXLS, query).None()).
				Render(IconStart("fa fa-download", "XLS"))
		}),

		If(len(collate.FilterFields) > 0, func() string {
			return Button().
				Submit().
				Class("rounded-r-lg shadow bg-white").
				Color(Blue).
				Click(fmt.Sprintf("window.document.getElementById('%s')?.classList.toggle('hidden');", collate.TargetFilter.ID)).
				Render(IconLeft("fa fa-fw fa-chevron-down", "Filter"))
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
			query := strings.ToLower(query.Order)

			if strings.Contains(query, field) {
				if strings.Contains(query, "asc") {
					direction = "asc"
				} else {
					direction = "desc"
				}

				color = Purple
			}

			reverse := "desc"

			if direction == "desc" {
				reverse = "asc"
			}

			return Button().
				Class("rounded bg-white").
				Color(color).
				Click(ctx.Call(collate.onSort, TQuery{Order: sort.DB + " " + reverse}).Replace(collate.Target)).
				Render(
					Div("flex gap-2 items-center")(
						Iff(direction == "asc")(Icon("fa fa-fw fa-sort-amount-asc")),
						Iff(direction == "desc")(Icon("fa fa-fw fa-sort-amount-desc")),
						Iff(direction == "")(Icon("fa fa-fw fa-sort")),
						sort.Text,
					),
				)
		}),
	)
}

func Paging[T any](ctx *Context, collate *collate[T], result *TCollateResult[T]) string {
	if result.Filtered == 0 {
		return Empty(result)
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
				Color(PurpleOutline).
				Disabled(size == 0 || size <= int(collate.Init.Limit)).
				Click(ctx.Call(collate.onReset).Replace(collate.Target)).
				Render(
					Icon("fa fa-fw fa-undo"),
					// Div("flex gap-2 items-center")(
					// 	Icon("fa fa-repeat"), reset,
					// ),
				),

			// load more
			Button().
				Class("rounded-r").
				Color(Purple).
				Disabled(size >= int(result.Filtered)).
				Click(ctx.Call(collate.onResize, result.Query).Replace(collate.Target)).
				Render(
					Div("flex gap-2 items-center")(
						Icon("fa fa-arrow-down"), more,
					),
				),
		),
	)
}
