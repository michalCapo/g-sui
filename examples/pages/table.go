package pages

import "github.com/michalCapo/g-sui/ui"

func Table(_ *ui.Context) string {
	table := ui.SimpleTable(4, "w-full table-auto")
	table.Class(0, "text-left font-bold p-2 border-b border-gray-200").
		Class(1, "text-left p-2 border-b border-gray-200").
		Class(2, "text-left p-2 border-b border-gray-200").
		Class(3, "text-right p-2 border-b border-gray-200")

	table.Field("ID").Field("Name").Field("Email").Field("Actions")

	table.Field("1").Field("John Doe").Field("john@example.com").Field(
		ui.Button().Class("px-3 py-1 rounded").Color(ui.Blue).Render("View"),
	)

	table.Field("2").Field("Jane Roe").Field("jane@example.com").Field(
		ui.Button().Class("px-3 py-1 rounded").Color(ui.Green).Render("Edit"),
	)

	// Demonstrate colspan row; renderer now wraps to a new row automatically
	table.Field("Notice", "text-blue-700 font-semibold text-center").Attr("colspan=\"4\"")

	// Mixed row where middle spans 2 cols, button stays in last col
	table.Field("3").Field("No Email User").Attr("colspan=\"2\"").Field(
		ui.Button().Class("px-3 py-1 rounded").Color(ui.Gray).Render("Disabled"),
	)

	// Additional rows with different colspans
	table.Field("Span 2", "text-center").Attr("colspan=\"2\"").Field("Right side").Attr("colspan=\"2\"")
	table.Empty().Field("Span across 3 columns").Attr("colspan=\"3\"")

	card := ui.Div("bg-white rounded shadow p-4 border border-gray-200 overflow-hidden")(
		ui.Div("text-lg font-bold")("Basic"),
		table.Render(),
	)

	// Colspan demos
	t2 := ui.SimpleTable(4, "w-full table-auto")
	t2.Class(0, "p-2 border-b border-gray-200").
		Class(1, "p-2 border-b border-gray-200").
		Class(2, "p-2 border-b border-gray-200").
		Class(3, "p-2 border-b border-gray-200")
	t2.Field("Full-width notice", "text-blue-700 font-semibold").Attr("colspan=\"4\"")
	t2.Field("Left span 2").Attr("colspan=\"2\"").Field("Right span 2").Attr("colspan=\"2\"")
	t2.Field("Span 3").Attr("colspan=\"3\"").Field("End")
	card2 := ui.Div("bg-white rounded shadow p-4 border border-gray-200 overflow-hidden")(
		ui.Div("text-lg font-bold")("Colspan"),
		t2.Render(),
	)

	// Numeric alignment + totals
	t3 := ui.SimpleTable(3, "w-full table-auto")
	t3.Class(0, "text-left p-2 border-b border-gray-200").
		Class(1, "text-right p-2 border-b border-gray-200").
		Class(2, "text-right p-2 border-b border-gray-200")
	t3.Field("Item").Field("Qty").Field("Amount")
	t3.Field("Apples").Field("3").Field("$6.00")
	t3.Field("Oranges").Field("2").Field("$5.00")
	t3.Field("Total", "font-semibold")
	t3.Empty()
	t3.Field("$11.00")
	card3 := ui.Div("bg-white rounded shadow p-4 border border-gray-200 overflow-hidden")(
		ui.Div("text-lg font-bold")("Numeric alignment"),
		t3.Render(),
	)

	return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Table"),
		ui.Div("text-gray-600")("Simple table utility with colspan, empty cells, and alignment."),
		card,
		card2,
		card3,
	)
}
