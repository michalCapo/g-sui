package pages

import (
	"fmt"
	"sort"
	"strings"

	r "github.com/michalCapo/g-sui/ui"
)

// Product is the sample type for the DataTable demo.
type Product struct {
	ID    int
	Name  string
	Price float64
	Stock int
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

	// Section 4: DataTable (Generic) — typed table with search, sort, pagination, export
	// Show first page of data (5 items per page)
	pageSize := 5
	initialData := allProducts[:pageSize]
	totalPages := (len(allProducts) + pageSize - 1) / pageSize

	// todo use the smae table definition inside this file
	dataTable := NewData().
		Page(1).TotalPages(totalPages).TotalItems(len(allProducts)).
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
	{ID: 1, Name: "Laptop", Price: 999.99, Stock: 25},
	{ID: 2, Name: "Mouse", Price: 29.50, Stock: 150},
	{ID: 3, Name: "Keyboard", Price: 79.00, Stock: 80},
	{ID: 4, Name: "Monitor", Price: 449.99, Stock: 30},
	{ID: 5, Name: "Headphones", Price: 59.95, Stock: 200},
	{ID: 6, Name: "Webcam", Price: 89.99, Stock: 45},
	{ID: 7, Name: "USB Cable", Price: 12.99, Stock: 500},
	{ID: 8, Name: "Desk Lamp", Price: 34.50, Stock: 60},
	{ID: 9, Name: "Notebook", Price: 8.99, Stock: 300},
	{ID: 10, Name: "Pen Set", Price: 15.00, Stock: 120},
	{ID: 11, Name: "Monitor Stand", Price: 45.00, Stock: 40},
	{ID: 12, Name: "Laptop Bag", Price: 55.00, Stock: 75},
}

// TableDataRequest represents the incoming data from table operations
type TableDataRequest struct {
	Operation string `json:"operation"`
	Search    string `json:"search"`
	Page      int    `json:"page"`
	Sort      int    `json:"sort"`
	Dir       string `json:"dir"`
}

func handleTableData(ctx *r.Context) string {
	var req TableDataRequest
	ctx.Body(&req)

	// Filter by search
	filtered := make([]*Product, 0)
	if req.Search != "" {
		searchLower := strings.ToLower(req.Search)
		for _, p := range allProducts {
			if strings.Contains(strings.ToLower(p.Name), searchLower) ||
				strings.Contains(fmt.Sprintf("%d", p.ID), searchLower) {
				filtered = append(filtered, p)
			}
		}
	} else {
		filtered = append(filtered, allProducts...)
	}

	// Sort
	if req.Sort >= 0 && req.Sort < 4 {
		sort.Slice(filtered, func(i, j int) bool {
			var cmp int
			switch req.Sort {
			case 0: // ID
				cmp = filtered[i].ID - filtered[j].ID
			case 1: // Name
				cmp = strings.Compare(filtered[i].Name, filtered[j].Name)
			case 2: // Price
				if filtered[i].Price < filtered[j].Price {
					cmp = -1
				} else if filtered[i].Price > filtered[j].Price {
					cmp = 1
				}
			case 3: // Stock
				cmp = filtered[i].Stock - filtered[j].Stock
			}
			if req.Dir == "desc" {
				return cmp > 0
			}
			return cmp < 0
		})
	}

	// Handle export operation
	if req.Operation == "export" {
		// In a real app, this would generate a CSV/Excel file
		// For demo, we'll just return a console log
		return fmt.Sprintf("console.log('Exporting %d items');", len(filtered))
	}

	// Paginate
	pageSize := 5
	totalItems := len(filtered)
	totalPages := max(1, (totalItems+pageSize-1)/pageSize)

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Page > totalPages {
		req.Page = totalPages
	}

	start := (req.Page - 1) * pageSize
	end := min(start+pageSize, totalItems)
	pageData := filtered[start:end]

	// Build the table
	dataTable := NewData().
		Page(req.Page).TotalPages(totalPages).TotalItems(totalItems).
		Sort(req.Sort, req.Dir).
		Search(req.Search).
		Render(pageData)

	return dataTable.ToJSReplace("products-table")
}

func NewData() *r.DataTable[Product] {
	dataTable := r.NewDataTable[Product]("products-table").
		Head("ID").Head("Name").Head("Price").Head("Stock").
		FieldText(func(p *Product) string { return fmt.Sprintf("%d", p.ID) }).
		FieldText(func(p *Product) string { return p.Name }).
		FieldText(func(p *Product) string { return fmt.Sprintf("$%.2f", p.Price) }).
		FieldText(func(p *Product) string { return fmt.Sprintf("%d", p.Stock) }).
		Sortable(0, 1, 2, 3).
		Action("table.data")

	return dataTable
}

func RegisterTable(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/table", func(ctx *r.Context) *r.Node { return layout(TablePage(ctx)) })
	app.Action("nav.table", NavTo("/table", func() *r.Node { return TablePage(nil) }))
	app.Action("table.data", handleTableData)
}
