package main

import (
	"github.com/michalCapo/g-sui/ui"
)

func main() {
	app := ui.MakeApp("en")

	// Initialize shared store
	invoiceStore = NewStore()

	// Layout: minimal nav + content area
	app.Layout(func(ctx *ui.Context) string {
		nav := ui.Div("bg-white dark:bg-gray-900 shadow")(
			ui.Div("max-w-5xl mx-auto px-4 py-3 flex items-center justify-between")(
				ui.Div("flex items-center gap-6")(
					ui.Div("text-lg font-bold tracking-tight text-gray-900 dark:text-white")("Invoices"),
					ui.Div("flex gap-1")(
						ui.A("px-3 py-1.5 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300",
							ctx.Load("/invoices"),
						)(ui.IconLeft("receipt_long", "All Invoices")),
						ui.A("px-3 py-1.5 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300",
							ctx.Load("/invoices/new"),
						)(ui.IconLeft("add", "Create")),
					),
				),
				ui.ThemeSwitcher(""),
			),
		)

		content := ui.Div("max-w-5xl mx-auto px-4 py-8", ui.Attr{ID: "__content__"})()

		return nav + content
	})

	// Pages
	app.Page("/", "Invoices", InvoiceList)
	app.Page("/invoices", "Invoices", InvoiceList)
	app.Page("/invoices/new", "New Invoice", InvoiceCreate)

	// REST API
	RegisterAPI(app, invoiceStore)

	app.Listen(":1423")
}
