package main

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// invoiceStore is the shared in-memory data store, set in main.go.
var invoiceStore *InvoiceStore

// ---------------------------------------------------------------------------
// Invoice List Page
// ---------------------------------------------------------------------------

// InvoiceList renders the invoice listing page with a client-rendered table
// featuring column filters (date range, text, number, enum).
func InvoiceList(ctx *ui.Context) string {
	return ui.Div("flex flex-col gap-6")(
		// Header
		ui.Div("flex items-center justify-between")(
			ui.Div("")(
				ui.Div("text-2xl font-bold tracking-tight")("Invoices"),
				ui.Div("text-sm text-gray-500 mt-1")("Manage and review all invoices. Click a row to expand details."),
			),
			ui.A(ui.Classes(ui.BTN, ui.SM, ui.Blue), ctx.Load("/invoices/new"))(
				ui.IconLeft("add", "New Invoice"),
			),
		),

		// Client-rendered table with column filters
		js.Client(ctx).
			Source("/api/invoices").
			Loading(ui.SkeletonTable).
			Empty("receipt_long", "No invoices found").
			Table(
				js.Col("number").Label("Invoice #").Sortable(true).Filterable(true).Class("w-32").CellClass("font-mono font-medium"),
				js.Col("company").Label("Company").Sortable(true).Filterable(true),
				js.Col("total").Label("Total").Type("number").Format("amount").Sortable(true).Filterable(true).CellClass("text-right tabular-nums"),
				js.Col("status").Label("Status").Type("enum").Sortable(true).Filterable(true).EnumOptions(
					js.Option{Value: "draft", Label: "Draft"},
					js.Option{Value: "sent", Label: "Sent"},
					js.Option{Value: "paid", Label: "Paid"},
					js.Option{Value: "overdue", Label: "Overdue"},
				),
				js.Col("dueDate").Label("Due Date").Type("date").Sortable(true).Filterable(true),
				js.Col("createdAt").Label("Created").Type("date").Sortable(true).Filterable(true),
			).
			Search(true).
			Pagination(10).
			Render(),
	)
}

// ---------------------------------------------------------------------------
// Create Invoice Page
// ---------------------------------------------------------------------------

// InvoiceForm holds the data for the create invoice form.
type InvoiceForm struct {
	Company     string `validate:"required"`
	Description string
	DueDate     time.Time `validate:"required"`
	Status      string    `validate:"required,oneof=draft sent"`

	// Line items (parallel slices -- simpler for form binding)
	ItemDesc  []string
	ItemQty   []int
	ItemPrice []float64

	// Number of item rows to display
	ItemCount int
}

// createFormTarget is a stable target for the create form page.
var createFormTarget = ui.Target()

// InvoiceCreate renders the create invoice page.
func InvoiceCreate(ctx *ui.Context) string {
	form := &InvoiceForm{
		Status:    "draft",
		ItemCount: 1,
	}
	return form.Render(ctx, nil)
}

// addItem is the server action to add another line item row.
func addItem(ctx *ui.Context) string {
	form := &InvoiceForm{}
	ctx.Body(form)
	form.ItemCount++
	return form.Render(ctx, nil)
}

// submitInvoice is the server action to validate and create the invoice.
func submitInvoice(ctx *ui.Context) string {
	form := &InvoiceForm{}
	if err := ctx.Body(form); err != nil {
		return form.Render(ctx, &err)
	}

	v := validator.New()
	if err := v.Struct(form); err != nil {
		return form.Render(ctx, &err)
	}

	// Build line items from parallel slices
	var items []InvoiceItem
	for i := 0; i < len(form.ItemDesc); i++ {
		desc := form.ItemDesc[i]
		qty := 0
		price := 0.0
		if i < len(form.ItemQty) {
			qty = form.ItemQty[i]
		}
		if i < len(form.ItemPrice) {
			price = form.ItemPrice[i]
		}
		if desc == "" && qty == 0 && price == 0 {
			continue
		}
		items = append(items, InvoiceItem{
			Description: desc,
			Quantity:    qty,
			UnitPrice:   price,
		})
	}

	if len(items) == 0 {
		errMsg := fmt.Errorf("at least one line item is required")
		var err error = errMsg
		return form.Render(ctx, &err)
	}

	// Create via the store directly (server-side)
	inv := Invoice{
		Company:     form.Company,
		Description: form.Description,
		Status:      form.Status,
		DueDate:     form.DueDate.Format("2006-01-02"),
		Items:       items,
	}
	created := invoiceStore.Create(inv)

	ctx.Success(fmt.Sprintf("Invoice %s created", created.Number))
	ctx.Redirect("/invoices")
	return ""
}

// Render renders the create invoice form.
func (f *InvoiceForm) Render(ctx *ui.Context, err *error) string {
	statuses := ui.MakeOptions([]string{"draft", "sent"})

	// Ensure at least 1 item row
	if f.ItemCount < 1 {
		f.ItemCount = 1
	}

	// Build item rows
	itemRows := ""
	for i := 0; i < f.ItemCount; i++ {
		desc := ""
		qty := 0
		price := 0.0
		if i < len(f.ItemDesc) {
			desc = f.ItemDesc[i]
		}
		if i < len(f.ItemQty) {
			qty = f.ItemQty[i]
		}
		if i < len(f.ItemPrice) {
			price = f.ItemPrice[i]
		}

		itemRows += ui.Div("grid grid-cols-12 gap-3 items-end")(
			ui.Div("col-span-6")(
				ui.If(i == 0, func() string {
					return ui.Div("text-xs font-medium text-gray-500 mb-1")("Description")
				}),
				ui.Input("w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm",
					ui.Attr{
						Type:        "text",
						Name:        fmt.Sprintf("ItemDesc[%d]", i),
						Value:       desc,
						Placeholder: "Service or product...",
					}),
			),
			ui.Div("col-span-2")(
				ui.If(i == 0, func() string {
					return ui.Div("text-xs font-medium text-gray-500 mb-1")("Qty")
				}),
				ui.Input("w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm",
					ui.Attr{
						Type:  "number",
						Name:  fmt.Sprintf("ItemQty[%d]", i),
						Value: fmt.Sprintf("%d", qty),
						Min:   "0",
					}),
			),
			ui.Div("col-span-3")(
				ui.If(i == 0, func() string {
					return ui.Div("text-xs font-medium text-gray-500 mb-1")("Unit Price")
				}),
				ui.Input("w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm",
					ui.Attr{
						Type:  "number",
						Name:  fmt.Sprintf("ItemPrice[%d]", i),
						Value: fmt.Sprintf("%.2f", price),
						Min:   "0",
						Step:  "0.01",
					}),
			),
			ui.Div("col-span-1 flex items-center justify-center text-gray-400 text-sm pb-1")(
				fmt.Sprintf("#%d", i+1),
			),
		)
	}

	// Hidden field to preserve item count across re-renders
	itemCountHidden := ui.Hidden("ItemCount", f.ItemCount)

	return ui.Div("flex flex-col gap-6", createFormTarget)(
		// Header
		ui.Div("flex items-center justify-between")(
			ui.Div("")(
				ui.Div("text-2xl font-bold tracking-tight")("New Invoice"),
				ui.Div("text-sm text-gray-500 mt-1")("Create a new invoice with line items."),
			),
			ui.A("text-sm text-blue-600 hover:text-blue-800 cursor-pointer", ctx.Load("/invoices"))(
				ui.IconLeft("arrow_back", "Back to list"),
			),
		),

		// Form card
		ui.Form("bg-white dark:bg-gray-900 rounded-lg shadow p-6 flex flex-col gap-5",
			createFormTarget,
			ctx.Submit(submitInvoice).Replace(createFormTarget),
		)(
			ui.ErrorForm(err, nil),
			itemCountHidden,

			// Invoice details section
			ui.Div("flex flex-col gap-4")(
				ui.Div("text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide")("Invoice Details"),
				ui.Div("grid grid-cols-1 sm:grid-cols-2 gap-4")(
					ui.IText("Company", f).Required().Render("Company"),
					ui.ISelect("Status", f).Options(statuses).Render("Status"),
				),
				ui.IDate("DueDate", f).Required().Render("Due Date"),
				ui.IArea("Description", f).Rows(2).Placeholder("Notes or description...").Render("Description"),
			),

			// Divider
			ui.ElClosed("hr", "border-gray-200 dark:border-gray-700"),

			// Line items section
			ui.Div("flex flex-col gap-3")(
				ui.Div("flex items-center justify-between")(
					ui.Div("text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide")("Line Items"),
					ui.Button().Color(ui.GrayOutline).Size(ui.XS).
						Click(ctx.Call(addItem, f).Replace(createFormTarget)).
						Render(ui.IconLeft("add", "Add Row")),
				),
				itemRows,
			),

			// Divider
			ui.ElClosed("hr", "border-gray-200 dark:border-gray-700"),

			// Actions
			ui.Div("flex items-center gap-3 justify-end")(
				ui.A("px-4 py-2 text-sm text-gray-600 hover:text-gray-800 cursor-pointer", ctx.Load("/invoices"))("Cancel"),
				ui.Button().Submit().Color(ui.Blue).Size(ui.SM).Render("Create Invoice"),
			),
		),
	)
}
