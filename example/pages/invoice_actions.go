package pages

import (
	"fmt"
	"strconv"

	r "github.com/michalCapo/g-sui/ui"
)

// ContentID is the shared target ID for the main content area, generated once
// via r.Target() so every page and action handler references the same element.
var ContentID = r.Target()

// invoiceStore is the package-level store reference, set via InitInvoiceStore.
var invoiceStore *InvoiceStore

// InitInvoiceStore wires the store into the pages package so action handlers
// can access it without import cycles.
func InitInvoiceStore(s *InvoiceStore) {
	invoiceStore = s
}

// ---------------------------------------------------------------------------
// Page constructors (used by main for app.Page registrations)
// ---------------------------------------------------------------------------

func InvoiceListPage(layout func(*r.Node) *r.Node) func(ctx *r.Context) *r.Node {
	return func(ctx *r.Context) *r.Node {
		return layout(InvoiceListContent(invoiceStore.All()))
	}
}

func InvoiceCreatePage(layout func(*r.Node) *r.Node) func(ctx *r.Context) *r.Node {
	return func(ctx *r.Context) *r.Node {
		return layout(InvoiceCreateForm(1, ""))
	}
}

// ---------------------------------------------------------------------------
// Navigation content helpers (used by main for navTo wrappers)
// ---------------------------------------------------------------------------

func InvoiceListNav() *r.Node {
	return InvoiceListContent(invoiceStore.All())
}

func InvoiceCreateNav() *r.Node {
	return InvoiceCreateForm(1, "")
}

// ---------------------------------------------------------------------------
// Action handlers
// ---------------------------------------------------------------------------

func HandleInvoiceView(ctx *r.Context) string {
	var req struct{ ID float64 }
	ctx.Body(&req)

	inv := invoiceStore.ByID(uint(req.ID))
	if inv == nil {
		return r.Notify("error", "Invoice not found")
	}

	return InvoiceDetail(inv).ToJSInner(ContentID) +
		r.SetLocation(fmt.Sprintf("/invoices/%d", inv.ID))
}

func HandleInvoiceDelete(ctx *r.Context) string {
	var req struct{ ID float64 }
	ctx.Body(&req)

	if !invoiceStore.Delete(uint(req.ID)) {
		return r.Notify("error", "Invoice not found")
	}

	return r.NewResponse().
		Inner(ContentID, InvoiceListContent(invoiceStore.All())).
		Navigate("/invoices").
		Toast("success", "Invoice deleted").
		Build()
}

func HandleInvoiceAddRow(ctx *r.Context) string {
	var req struct{ ItemCount float64 }
	ctx.Body(&req)
	newCount := int(req.ItemCount) + 1

	return InvoiceCreateForm(newCount, "").ToJSReplace("create-form")
}

func HandleInvoiceCreate(ctx *r.Context) string {
	var data map[string]any
	ctx.Body(&data)

	company, _ := data["company"].(string)
	status, _ := data["status"].(string)
	dueDate, _ := data["dueDate"].(string)
	description, _ := data["description"].(string)

	count := ItemCountFrom(data)

	if company == "" {
		return InvoiceCreateForm(count, "Company is required").
			ToJSReplace("create-form")
	}
	if dueDate == "" {
		return InvoiceCreateForm(count, "Due date is required").
			ToJSReplace("create-form")
	}

	var items []InvoiceItem
	for i := 0; i < count; i++ {
		desc, _ := data[fmt.Sprintf("ItemDesc_%d", i)].(string)
		qtyStr, _ := data[fmt.Sprintf("ItemQty_%d", i)].(string)
		priceStr, _ := data[fmt.Sprintf("ItemPrice_%d", i)].(string)

		qty, _ := strconv.Atoi(qtyStr)
		price, _ := strconv.ParseFloat(priceStr, 64)

		if desc == "" && qty == 0 && price == 0 {
			continue
		}
		items = append(items, InvoiceItem{Description: desc, Quantity: qty, UnitPrice: price})
	}

	if len(items) == 0 {
		return InvoiceCreateForm(count, "At least one line item is required").
			ToJSReplace("create-form")
	}

	inv := invoiceStore.Create(Invoice{
		Company:     company,
		Description: description,
		Status:      status,
		DueDate:     dueDate,
		Items:       items,
	})

	return r.NewResponse().
		Inner(ContentID, InvoiceListContent(invoiceStore.All())).
		Navigate("/invoices").
		Toast("success", fmt.Sprintf("Invoice %s created", inv.Number)).
		Build()
}
