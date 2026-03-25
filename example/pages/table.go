package pages

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/go-pdf/fpdf"
	r "github.com/michalCapo/g-sui/ui"
)

// Product is the sample type for the DataTable demo.
type Product struct {
	ID           int
	Name         string
	Price        float64
	Stock        int
	CreatedAt    string // Date column for filter demo
	Category     string // Select filter demo
	Status       string // Status badge demo
	ReleaseMonth string // MonthYear filter demo (format: "2026-03")
}

func TablePage(ctx *r.Context) *r.Node {
	// Section 1: Basic — SimpleTable with action buttons
	viewBtn := r.NewButton("View").BtnColor(r.BtnBlue).BtnSize(r.BtnXS).Build()
	editBtn := r.NewButton("Edit").BtnColor(r.BtnGreen).BtnSize(r.BtnXS).Build()

	basicTable := r.NewSimpleTable(4).
		SimpleHeader("ID", "Name", "Email", "Actions").
		CellText("1").CellText("John Doe").CellText("john@example.com").Cell(viewBtn).
		CellText("2").CellText("Jane Roe").CellText("jane@example.com").Cell(editBtn).
		Build()

	// Section 2: Colspan — raw table (SimpleTable doesn't support colspan)
	colspanTable := r.Table("w-full table-auto text-sm").Render(
		r.Tbody().Render(
			r.Tr().Render(
				r.Td("text-blue-700 font-semibold p-2 border-b border-gray-200").Attr("colspan", "4").Text("Full-width notice"),
			),
			r.Tr().Render(
				r.Td("p-2 border-b border-gray-200").Attr("colspan", "2").Text("Left span 2"),
				r.Td("p-2 border-b border-gray-200").Attr("colspan", "2").Text("Right span 2"),
			),
			r.Tr().Render(
				r.Td("p-2 border-b border-gray-200").Attr("colspan", "3").Text("Span 3"),
				r.Td("p-2 border-b border-gray-200").Text("End"),
			),
		),
	)

	// Section 3: Numeric alignment — SimpleTable with 3 cols
	numTable := r.NewSimpleTable(3).
		SimpleHeader("Item", "Qty", "Amount").
		CellText("Apples").CellText("3").CellText("$6.00").
		CellText("Oranges").CellText("2").CellText("$5.00").
		CellText("Total").CellText("").CellText("$11.00").
		Build()

	// Section 4: DataTable (Generic) — typed table with search, sort, load-more, export
	pageSize := 10
	initialData := allProducts[:pageSize]
	hasMore := len(allProducts) > pageSize

	dataTable := NewData().
		Page(1).TotalItems(len(allProducts)).HasMore(hasMore).
		Render(initialData)

	wrapCard := func(title string, table *r.Node) *r.Node {
		return r.Div("bg-white rounded shadow p-4 border border-gray-200 overflow-hidden").Render(
			r.Div("text-lg font-bold mb-2").Text(title),
			table,
		)
	}

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Table"),
		r.Div("text-gray-600").Text("Simple table utility with colspan, empty cells, and alignment."),
		wrapCard("Basic", basicTable),
		wrapCard("Colspan", colspanTable),
		wrapCard("Numeric alignment", numTable),
		wrapCard("DataTable (Generic)", dataTable),
	)
}

// allProducts is the full dataset for the DataTable demo
var allProducts = []*Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Stock: 25, CreatedAt: "2026-01-15", Category: "Electronics", Status: "Draft", ReleaseMonth: "2026-01"},
	{ID: 2, Name: "Mouse", Price: 29.50, Stock: 150, CreatedAt: "2026-01-20", Category: "Accessories", Status: "Sent", ReleaseMonth: "2026-01"},
	{ID: 3, Name: "Keyboard", Price: 79.00, Stock: 80, CreatedAt: "2026-02-05", Category: "Accessories", Status: "Paid", ReleaseMonth: "2026-02"},
	{ID: 4, Name: "Monitor", Price: 449.99, Stock: 30, CreatedAt: "2026-02-10", Category: "Electronics", Status: "Paid", ReleaseMonth: "2026-02"},
	{ID: 5, Name: "Headphones", Price: 59.95, Stock: 200, CreatedAt: "2026-02-15", Category: "Accessories", Status: "Overdue", ReleaseMonth: "2026-02"},
	{ID: 6, Name: "Webcam", Price: 89.99, Stock: 45, CreatedAt: "2026-03-01", Category: "Electronics", Status: "Sent", ReleaseMonth: "2026-03"},
	{ID: 7, Name: "USB Cable", Price: 12.99, Stock: 500, CreatedAt: "2026-03-05", Category: "Accessories", Status: "Paid", ReleaseMonth: "2026-03"},
	{ID: 8, Name: "Desk Lamp", Price: 34.50, Stock: 60, CreatedAt: "2026-03-10", Category: "Office", Status: "Draft", ReleaseMonth: "2026-03"},
	{ID: 9, Name: "Notebook", Price: 8.99, Stock: 300, CreatedAt: "2026-03-15", Category: "Office", Status: "Paid", ReleaseMonth: "2026-03"},
	{ID: 10, Name: "Pen Set", Price: 15.00, Stock: 120, CreatedAt: "2026-03-20", Category: "Office", Status: "Sent", ReleaseMonth: "2026-03"},
	{ID: 11, Name: "Monitor Stand", Price: 45.00, Stock: 40, CreatedAt: "2026-04-01", Category: "Accessories", Status: "Overdue", ReleaseMonth: "2026-04"},
	{ID: 12, Name: "Laptop Bag", Price: 55.00, Stock: 75, CreatedAt: "2026-04-05", Category: "Accessories", Status: "Paid", ReleaseMonth: "2026-04"},
	{ID: 13, Name: "Tablet", Price: 599.99, Stock: 35, CreatedAt: "2026-05-01", Category: "Electronics", Status: "Draft", ReleaseMonth: "2026-05"},
	{ID: 14, Name: "Mouse Pad", Price: 14.99, Stock: 220, CreatedAt: "2026-05-10", Category: "Accessories", Status: "Sent", ReleaseMonth: "2026-05"},
	{ID: 15, Name: "Printer", Price: 249.00, Stock: 15, CreatedAt: "2026-05-15", Category: "Electronics", Status: "Paid", ReleaseMonth: "2026-05"},
	{ID: 16, Name: "Stapler", Price: 9.50, Stock: 180, CreatedAt: "2026-06-01", Category: "Office", Status: "Overdue", ReleaseMonth: "2026-06"},
	{ID: 17, Name: "Router", Price: 129.99, Stock: 55, CreatedAt: "2026-06-10", Category: "Electronics", Status: "Sent", ReleaseMonth: "2026-06"},
	{ID: 18, Name: "USB Hub", Price: 24.99, Stock: 140, CreatedAt: "2026-06-20", Category: "Accessories", Status: "Paid", ReleaseMonth: "2026-06"},
	{ID: 19, Name: "Desk Chair", Price: 349.00, Stock: 20, CreatedAt: "2026-07-05", Category: "Office", Status: "Draft", ReleaseMonth: "2026-07"},
	{ID: 20, Name: "Microphone", Price: 89.00, Stock: 65, CreatedAt: "2026-07-15", Category: "Electronics", Status: "Paid", ReleaseMonth: "2026-07"},
	{ID: 21, Name: "Whiteboard", Price: 42.00, Stock: 30, CreatedAt: "2026-08-01", Category: "Office", Status: "Sent", ReleaseMonth: "2026-08"},
	{ID: 22, Name: "HDMI Cable", Price: 11.99, Stock: 400, CreatedAt: "2026-08-10", Category: "Accessories", Status: "Paid", ReleaseMonth: "2026-08"},
	{ID: 23, Name: "Speaker", Price: 69.95, Stock: 90, CreatedAt: "2026-09-01", Category: "Electronics", Status: "Overdue", ReleaseMonth: "2026-09"},
	{ID: 24, Name: "Paper Tray", Price: 18.50, Stock: 110, CreatedAt: "2026-09-15", Category: "Office", Status: "Paid", ReleaseMonth: "2026-09"},
}

// TableDataRequest represents the incoming data from table operations
type TableDataRequest struct {
	Operation string   `json:"operation"`
	Search    string   `json:"search"`
	Page      int      `json:"page"`
	PageSize  int      `json:"pageSize"`
	Sort      int      `json:"sort"`
	Dir       string   `json:"dir"`
	Col       int      `json:"col"`
	Type      string   `json:"type"`
	Op        string   `json:"op"`
	Val       string   `json:"val"`
	Vals      []string `json:"vals"`
	From      string   `json:"from"`
	To        string   `json:"to"`
}

// activeFilter stores a currently active column filter
type activeFilter struct {
	Col  int
	Type string
	Op   string
	Val  string
	Vals []string
	From string
	To   string
}

// Global active filters state (per-session in real app, global here for demo)
var activeFilters = map[int]*activeFilter{}

func filterProducts(search string, filters map[int]*activeFilter) []*Product {
	filtered := make([]*Product, 0, len(allProducts))
	for _, p := range allProducts {
		// Text search
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(p.Name), searchLower) &&
				!strings.Contains(fmt.Sprintf("%d", p.ID), searchLower) {
				continue
			}
		}

		// Column filters
		if !applyColumnFilters(p, filters) {
			continue
		}

		filtered = append(filtered, p)
	}
	return filtered
}

func applyColumnFilters(p *Product, filters map[int]*activeFilter) bool {
	for _, f := range filters {
		switch f.Type {
		case "date":
			val := p.CreatedAt
			if f.From != "" && val < f.From {
				return false
			}
			if f.To != "" && val > f.To {
				return false
			}
		case "monthyear":
			val := p.ReleaseMonth
			if f.From != "" && val < f.From {
				return false
			}
			if f.To != "" && val > f.To {
				return false
			}
		case "number":
			var numVal float64
			switch f.Col {
			case 0:
				numVal = float64(p.ID)
			case 2:
				numVal = p.Price
			case 3:
				numVal = float64(p.Stock)
			default:
				continue
			}
			from := parseFloat(f.From)
			to := parseFloat(f.To)
			switch f.Op {
			case "range":
				if f.From != "" && numVal < from {
					return false
				}
				if f.To != "" && numVal > to {
					return false
				}
			case "gte":
				if f.From != "" && numVal < from {
					return false
				}
			case "lte":
				if f.From != "" && numVal > from {
					return false
				}
			case "gt":
				if f.From != "" && numVal <= from {
					return false
				}
			case "lt":
				if f.From != "" && numVal >= from {
					return false
				}
			case "equals":
				if f.From != "" && numVal != from {
					return false
				}
			}
		case "text":
			var val string
			switch f.Col {
			case 1:
				val = p.Name
			default:
				continue
			}
			valLower := strings.ToLower(val)
			fValLower := strings.ToLower(f.Val)
			switch f.Op {
			case "contains":
				if !strings.Contains(valLower, fValLower) {
					return false
				}
			case "startswith":
				if !strings.HasPrefix(valLower, fValLower) {
					return false
				}
			case "equals":
				if valLower != fValLower {
					return false
				}
			}
		case "select":
			if len(f.Vals) == 0 {
				continue
			}
			var val string
			switch f.Col {
			case 5:
				val = p.Category
			case 6:
				val = p.Status
			default:
				continue
			}
			found := false
			for _, v := range f.Vals {
				if v == val {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func sortProducts(data []*Product, col int, dir string) {
	if col < 0 || col > 7 {
		return
	}
	sort.Slice(data, func(i, j int) bool {
		var cmp int
		switch col {
		case 0:
			cmp = data[i].ID - data[j].ID
		case 1:
			cmp = strings.Compare(data[i].Name, data[j].Name)
		case 2:
			if data[i].Price < data[j].Price {
				cmp = -1
			} else if data[i].Price > data[j].Price {
				cmp = 1
			}
		case 3:
			cmp = data[i].Stock - data[j].Stock
		case 4:
			cmp = strings.Compare(data[i].CreatedAt, data[j].CreatedAt)
		case 5:
			cmp = strings.Compare(data[i].Category, data[j].Category)
		case 6:
			cmp = strings.Compare(data[i].Status, data[j].Status)
		case 7:
			cmp = strings.Compare(data[i].ReleaseMonth, data[j].ReleaseMonth)
		}
		if dir == "desc" {
			return cmp > 0
		}
		return cmp < 0
	})
}

func handleTableData(ctx *r.Context) string {
	var req TableDataRequest
	ctx.Body(&req)

	pageSize := 10
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	// Handle filter operation: update active filters
	if req.Operation == "filter" {
		if req.Type != "" {
			// Apply a specific column filter
			if req.Type == "select" && len(req.Vals) == 0 {
				delete(activeFilters, req.Col)
			} else if (req.Type == "date" || req.Type == "monthyear") && req.From == "" && req.To == "" {
				delete(activeFilters, req.Col)
			} else if req.Type == "number" && req.From == "" && req.To == "" {
				delete(activeFilters, req.Col)
			} else if req.Type == "text" && req.Val == "" {
				delete(activeFilters, req.Col)
			} else {
				activeFilters[req.Col] = &activeFilter{
					Col: req.Col, Type: req.Type, Op: req.Op,
					Val: req.Val, Vals: req.Vals, From: req.From, To: req.To,
				}
			}
		} else {
			// Reset all filters
			activeFilters = map[int]*activeFilter{}
		}
	}

	// Handle removing a single filter (badge × button)
	if req.Operation == "removeFilter" {
		delete(activeFilters, req.Col)
	}

	filtered := filterProducts(req.Search, activeFilters)
	sortProducts(filtered, req.Sort, req.Dir)
	totalItems := len(filtered)

	// Handle export (CSV)
	if req.Operation == "export" {
		return exportProductsCSV(filtered)
	}

	// Handle PDF export
	if req.Operation == "export-pdf" {
		return exportProductsPDF(filtered)
	}

	// Handle "load more": append rows to existing tbody + replace footer
	if req.Operation == "loadmore" {
		if req.Page < 1 {
			req.Page = 1
		}
		start := (req.Page - 1) * pageSize
		end := min(start+pageSize, totalItems)
		if start >= totalItems {
			return ""
		}
		pageData := filtered[start:end]
		hasMore := end < totalItems

		dt := newDataWithFilters().
			Page(req.Page).TotalItems(totalItems).HasMore(hasMore).
			Sort(req.Sort, req.Dir).Search(req.Search).
			RowOffset(start)

		resp := r.NewResponse()
		rows := dt.RenderRows(pageData)
		for _, row := range rows {
			resp.Append(dt.TbodyID(), row)
		}
		resp.Replace(dt.FooterID(), dt.RenderFooter())
		return resp.Build()
	}

	// Default: full table re-render (search, sort, filter)
	if req.Page < 1 {
		req.Page = 1
	}
	end := min(req.Page*pageSize, totalItems)
	pageData := filtered[:end]
	hasMore := end < totalItems

	dataTable := newDataWithFilters().
		Page(req.Page).TotalItems(totalItems).HasMore(hasMore).
		Sort(req.Sort, req.Dir).Search(req.Search).
		Render(pageData)

	return dataTable.ToJSReplace("products-table")
}

// newDataWithFilters creates a DataTable with current active filter state applied
func newDataWithFilters() *r.DataTable[Product] {
	dt := NewData()

	// Build filter badges and set filter values
	var badges []r.FilterBadge
	colLabels := map[int]string{0: "ID", 1: "Name", 2: "Price", 3: "Stock", 4: "Created", 5: "Category", 6: "Status", 7: "Release"}

	for col, f := range activeFilters {
		label := colLabels[col]
		var valueStr string

		switch f.Type {
		case "date":
			valueStr = f.From + " – " + f.To
		case "monthyear":
			valueStr = f.From + " – " + f.To
		case "number":
			if f.Op == "range" {
				valueStr = f.From + " – " + f.To
			} else {
				valueStr = f.Op + " " + f.From
			}
		case "text":
			valueStr = f.Val
		case "select":
			valueStr = strings.Join(f.Vals, ", ")
		}

		dt.SetFilterValue(col, &r.FilterValue{
			Operator: f.Op, Value: f.Val, Values: f.Vals, From: f.From, To: f.To,
		})

		badges = append(badges, r.FilterBadge{
			Label:  label,
			Value:  valueStr,
			Column: col,
			OnRemove: fmt.Sprintf(
				"__ws.call('table.data',{operation:'removeFilter',col:%d,search:'',page:1,sort:-1,dir:'asc'})",
				col,
			),
		})
	}

	if len(badges) > 0 {
		dt.SetFilterLabels(badges)
	}

	return dt
}

func NewData() *r.DataTable[Product] {
	dataTable := r.NewDataTable[Product]("products-table").
		Col("ID", r.ColOpt[Product]{
			Sortable: true,
			Filter:   r.NumFilter,
			Text:     func(p *Product) *r.Node { return r.Span().Text(fmt.Sprintf("%d", p.ID)) },
		}).
		Col("Name", r.ColOpt[Product]{
			Sortable: true,
			Filter:   r.TxtFilter,
			Text:     func(p *Product) *r.Node { return r.Span().Text(p.Name) },
		}).
		Col("Price", r.ColOpt[Product]{
			Sortable: true,
			Filter:   r.NumFilter,
			Text:     func(p *Product) *r.Node { return r.Span().Text(fmt.Sprintf("$%.2f", p.Price)) },
		}).
		Col("Stock", r.ColOpt[Product]{
			Sortable: true,
			Filter:   r.NumFilter,
			Text:     func(p *Product) *r.Node { return r.Span().Text(fmt.Sprintf("%d", p.Stock)) },
		}).
		Col("Created", r.ColOpt[Product]{
			Sortable: true,
			Filter:   r.DateFilter,
			Text:     func(p *Product) *r.Node { return r.Span().Text(p.CreatedAt) },
		}).
		Col("Category", r.ColOpt[Product]{
			Sortable:      true,
			Filter:        r.SelectFilter,
			FilterOptions: []string{"Electronics", "Accessories", "Office"},
			Text:          func(p *Product) *r.Node { return r.Span().Text(p.Category) },
		}).
		Col("Status", r.ColOpt[Product]{
			Sortable:      true,
			Filter:        r.SelectFilter,
			FilterOptions: []string{"Draft", "Sent", "Paid", "Overdue"},
			Text: func(p *Product) *r.Node {
				color := "gray-soft"
				switch p.Status {
				case "Draft":
					color = "gray-soft"
				case "Sent":
					color = "blue-soft"
				case "Paid":
					color = "green-soft"
				case "Overdue":
					color = "red-soft"
				}
				return r.NewBadge(p.Status).Color(color).BadgeSize("sm").Build()
			},
		}).
		Col("Release", r.ColOpt[Product]{
			Sortable: true,
			Filter:   r.MonthYearFilter,
			Text:     func(p *Product) *r.Node { return r.Span().Text(p.ReleaseMonth) },
		}).
		Detail(productDetail).
		Action("table.data")

	return dataTable
}

// productDetail renders the expandable detail content for a product row.
func productDetail(p *Product) *r.Node {
	field := func(label, value string) *r.Node {
		return r.Div("flex gap-2").Render(
			r.Span("text-gray-500 dark:text-gray-400 font-medium min-w-[80px]").Text(label+":"),
			r.Span("text-gray-800 dark:text-gray-200").Text(value),
		)
	}

	return r.Div("grid grid-cols-2 gap-3 text-sm").Render(
		field("ID", fmt.Sprintf("%d", p.ID)),
		field("Name", p.Name),
		field("Price", fmt.Sprintf("$%.2f", p.Price)),
		field("Stock", fmt.Sprintf("%d units", p.Stock)),
		field("Created", p.CreatedAt),
		field("Category", p.Category),
		field("Status", p.Status),
		field("Release", p.ReleaseMonth),
		field("Value", fmt.Sprintf("$%.2f", p.Price*float64(p.Stock))),
	)
}

func exportProductsPDF(products []*Product) string {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// Title
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, "Products", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Table header
	headers := []string{"ID", "Name", "Price", "Stock", "Created", "Category", "Status", "Release"}
	widths := []float64{15, 55, 25, 20, 30, 35, 25, 25}

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(240, 240, 240)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Helvetica", "", 9)
	for _, p := range products {
		row := []string{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			fmt.Sprintf("$%.2f", p.Price),
			fmt.Sprintf("%d", p.Stock),
			p.CreatedAt,
			p.Category,
			p.Status,
			p.ReleaseMonth,
		}
		for i, cell := range row {
			align := "L"
			if i == 0 || i == 2 || i == 3 {
				align = "R"
			}
			pdf.CellFormat(widths[i], 7, cell, "1", 0, align, false, 0, "")
		}
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return fmt.Sprintf("console.error('PDF error: %s');", err)
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return r.Download("products.pdf", "application/pdf", b64)
}

func exportProductsCSV(products []*Product) string {
	var buf bytes.Buffer
	buf.WriteString("ID,Name,Price,Stock,Created,Category,Status,ReleaseMonth\n")
	for _, p := range products {
		name := strings.ReplaceAll(p.Name, "\"", "\"\"")
		buf.WriteString(fmt.Sprintf("%d,\"%s\",%.2f,%d,%s,%s,%s,%s\n",
			p.ID, name, p.Price, p.Stock, p.CreatedAt, p.Category, p.Status, p.ReleaseMonth))
	}
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return r.Download("products.csv", "text/csv", b64)
}

func RegisterTable(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/table", func(ctx *r.Context) *r.Node { return layout(ctx, TablePage(ctx)) })
	app.Action("nav.table", NavTo("/table", func() *r.Node { return TablePage(nil) }))
	app.Action("table.data", handleTableData)
}
