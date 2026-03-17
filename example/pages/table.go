package pages

import (
	"fmt"

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
	products := []*Product{
		{ID: 1, Name: "Laptop", Price: 999.99, Stock: 25},
		{ID: 2, Name: "Mouse", Price: 29.50, Stock: 150},
		{ID: 3, Name: "Keyboard", Price: 79.00, Stock: 80},
		{ID: 4, Name: "Monitor", Price: 449.99, Stock: 30},
		{ID: 5, Name: "Headphones", Price: 59.95, Stock: 200},
	}

	dataTable := r.NewDataTable[Product]("products-table").
		Head("ID").Head("Name").Head("Price").Head("Stock").
		FieldText(func(p *Product) string { return fmt.Sprintf("%d", p.ID) }).
		FieldText(func(p *Product) string { return p.Name }).
		FieldText(func(p *Product) string { return fmt.Sprintf("$%.2f", p.Price) }).
		FieldText(func(p *Product) string { return fmt.Sprintf("%d", p.Stock) }).
		Searchable("table.search").
		Sortable(0, 1, 2, 3).SortAction("table.sort").
		Paginated("table.page", 1, 3).TotalItems(15).
		Export("table.export").
		Render(products)

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
