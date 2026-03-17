package pages

import (
	"fmt"
	"strconv"
	"strings"

	r "github.com/michalCapo/g-sui/ui"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type InvoiceItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Total       float64
}

type Invoice struct {
	ID          uint
	Number      string
	Company     string
	Description string
	Amount      float64
	Tax         float64
	Total       float64
	Status      string
	DueDate     string
	CreatedAt   string
	Items       []InvoiceItem
}

// ---------------------------------------------------------------------------
// Invoice List Content
// ---------------------------------------------------------------------------

func InvoiceListContent(invoices []Invoice) *r.Node {
	return r.Div("flex flex-col gap-6").Render(
		r.Div("flex items-center justify-between").Render(
			r.Div().Render(
				r.Div("text-2xl font-bold tracking-tight").Text("Invoices"),
				r.Div("text-sm text-gray-500 mt-1").Text("Manage and review all invoices."),
			),
			r.Button("px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 cursor-pointer").
				Text("+ New Invoice").
				OnClick(&r.Action{Name: "nav.create"}),
		),
		r.Div("bg-white rounded-lg shadow overflow-hidden").Render(
			InvoiceTable(invoices),
		),
	)
}

// ---------------------------------------------------------------------------
// Invoice Table
// ---------------------------------------------------------------------------

func InvoiceTable(invoices []Invoice) *r.Node {
	return r.Table("w-full text-sm").Render(
		r.Thead("bg-gray-50 border-b").Render(
			r.Tr().Render(
				Th("Invoice #", "w-28"),
				Th("Company", ""),
				Th("Total", "text-right"),
				Th("Status", ""),
				Th("Due Date", ""),
				Th("", "w-20"),
			),
		),
		r.Tbody("divide-y divide-gray-100").Render(
			r.Map(invoices, func(inv Invoice, _ int) *r.Node {
				return InvoiceRow(inv)
			})...,
		),
	)
}

func InvoiceRow(inv Invoice) *r.Node {
	return r.Tr("hover:bg-gray-50 cursor-pointer").
		OnClick(&r.Action{Name: "invoice.view", Data: map[string]any{"ID": inv.ID}}).
		Render(
			Td(inv.Number, "font-mono font-medium"),
			Td(inv.Company, ""),
			Td(fmt.Sprintf("%.2f", inv.Total), "text-right tabular-nums"),
			r.Td("px-4 py-3").Render(InvoiceStatusBadge(inv.Status)),
			Td(inv.DueDate, ""),
			r.Td("px-4 py-3 text-gray-500 text-xs").Text(">"),
		)
}

func InvoiceStatusBadge(status string) *r.Node {
	colors := map[string]string{
		"draft":   "bg-gray-100 text-gray-700",
		"sent":    "bg-blue-100 text-blue-800",
		"paid":    "bg-green-100 text-green-800",
		"overdue": "bg-red-100 text-red-800",
	}
	c := colors[status]
	if c == "" {
		c = "bg-gray-100 text-gray-700"
	}
	return r.Span("px-2 py-0.5 rounded-full text-xs font-medium").
		Class(c).
		Text(strings.ToUpper(status[:1]) + status[1:])
}

// ---------------------------------------------------------------------------
// Invoice Detail
// ---------------------------------------------------------------------------

func InvoiceDetail(inv *Invoice) *r.Node {
	return r.Div("flex flex-col gap-6").Render(
		r.Div("flex items-center justify-between").Render(
			r.Div().Render(
				r.Div("text-2xl font-bold tracking-tight").Text(inv.Number),
				r.Div("text-sm text-gray-500 mt-1").Text(inv.Company),
			),
			r.Div("flex gap-2").Render(
				r.Button("px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800 cursor-pointer").
					Text("Back to list").
					OnClick(r.JS("history.back()")),
				r.Button("px-3 py-1.5 text-sm bg-red-600 text-white rounded hover:bg-red-700 cursor-pointer").
					Text("Delete").
					OnClick(&r.Action{Name: "invoice.delete", Data: map[string]any{"ID": inv.ID}}),
			),
		),
		r.Div("bg-white rounded-lg shadow p-6 flex flex-col gap-5").Render(
			r.Div("grid grid-cols-2 sm:grid-cols-4 gap-4").Render(
				InfoField("Status", strings.ToUpper(inv.Status[:1])+inv.Status[1:]),
				InfoField("Due Date", inv.DueDate),
				InfoField("Created", inv.CreatedAt),
				InfoField("Description", inv.Description),
			),
			r.Hr(),
			r.Div("grid grid-cols-3 gap-4 text-right").Render(
				TotalField("Subtotal", fmt.Sprintf("%.2f", inv.Amount)),
				TotalField("Tax (20%)", fmt.Sprintf("%.2f", inv.Tax)),
				TotalField("Total", fmt.Sprintf("%.2f", inv.Total)),
			),
			r.Hr(),
			r.Div("text-sm font-semibold text-gray-700 uppercase tracking-wide").Text("Line Items"),
			r.Table("w-full text-sm").Render(
				r.Thead("bg-gray-50 border-b").Render(
					r.Tr().Render(
						Th("Description", ""),
						Th("Qty", "text-right"),
						Th("Unit Price", "text-right"),
						Th("Total", "text-right"),
					),
				),
				r.Tbody("divide-y divide-gray-100").Render(
					r.Map(inv.Items, func(item InvoiceItem, _ int) *r.Node {
						return r.Tr().Render(
							Td(item.Description, ""),
							Td(strconv.Itoa(item.Quantity), "text-right"),
							Td(fmt.Sprintf("%.2f", item.UnitPrice), "text-right"),
							Td(fmt.Sprintf("%.2f", item.Total), "text-right"),
						)
					})...,
				),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Create Invoice Form
// ---------------------------------------------------------------------------

func InvoiceCreateForm(itemCount int, errMsg string) *r.Node {
	return r.Div("flex flex-col gap-6").ID("create-form").Render(
		r.Div("flex items-center justify-between").Render(
			r.Div().Render(
				r.Div("text-2xl font-bold tracking-tight").Text("New Invoice"),
				r.Div("text-sm text-gray-500 mt-1").Text("Create a new invoice with line items."),
			),
			r.Button("text-sm text-blue-600 hover:text-blue-800 cursor-pointer").
				Text("Back to list").
				OnClick(r.JS("history.back()")),
		),

		r.If(errMsg != "",
			r.Div("bg-red-50 border border-red-200 rounded px-4 py-3 text-sm text-red-700").Text(errMsg),
		),

		r.Div("bg-white rounded-lg shadow p-6 flex flex-col gap-5").Render(
			r.Div("text-sm font-semibold text-gray-700 uppercase tracking-wide").Text("Invoice Details"),
			r.Div("grid grid-cols-1 sm:grid-cols-2 gap-4").Render(
				FormField("Company", "company", "text", "Company name..."),
				FormSelect("Status", "status", []string{"draft", "sent"}),
			),
			FormField("Due Date", "dueDate", "date", ""),
			FormTextarea("Description", "description", "Notes or description..."),

			r.Hr(),

			r.Div("flex items-center justify-between").Render(
				r.Div("text-sm font-semibold text-gray-700 uppercase tracking-wide").Text("Line Items"),
				r.Button("px-3 py-1 text-xs border rounded text-gray-600 hover:bg-gray-50 cursor-pointer").
					Text("+ Add Row").
					OnClick(&r.Action{Name: "invoice.addRow", Data: map[string]any{"ItemCount": itemCount}}),
			),
			r.Div("flex flex-col gap-2").ID("item-rows").Render(
				itemRowNodes(itemCount)...,
			),

			r.Hr(),

			r.Div("flex items-center gap-3 justify-end").Render(
				r.Button("px-4 py-2 text-sm text-gray-600 hover:text-gray-800 cursor-pointer").
					Text("Cancel").
					OnClick(&r.Action{Name: "nav.list"}),
				r.Button("px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 cursor-pointer").
					Text("Create Invoice").
					OnClick(&r.Action{
						Name:    "invoice.create",
						Data:    map[string]any{"ItemCount": itemCount},
						Collect: CollectIDs(itemCount),
					}),
			),
		),
	)
}

func itemRowNodes(count int) []*r.Node {
	rows := make([]*r.Node, count)
	for i := 0; i < count; i++ {
		rows[i] = itemRow(i)
	}
	return rows
}

func itemRow(i int) *r.Node {
	return r.Div("grid grid-cols-12 gap-3 items-end").Render(
		r.Div("col-span-6").Render(
			r.If(i == 0, r.Div("text-xs font-medium text-gray-500 mb-1").Text("Description")),
			r.IText("w-full border border-gray-300 rounded px-3 py-2 text-sm").
				ID(fmt.Sprintf("item-desc-%d", i)).
				Attr("name", fmt.Sprintf("ItemDesc_%d", i)).
				Attr("placeholder", "Service or product..."),
		),
		r.Div("col-span-2").Render(
			r.If(i == 0, r.Div("text-xs font-medium text-gray-500 mb-1").Text("Qty")),
			r.INumber("w-full border border-gray-300 rounded px-3 py-2 text-sm").
				ID(fmt.Sprintf("item-qty-%d", i)).
				Attr("name", fmt.Sprintf("ItemQty_%d", i)).
				Attr("min", "0").
				Attr("value", "1"),
		),
		r.Div("col-span-3").Render(
			r.If(i == 0, r.Div("text-xs font-medium text-gray-500 mb-1").Text("Unit Price")),
			r.INumber("w-full border border-gray-300 rounded px-3 py-2 text-sm").
				ID(fmt.Sprintf("item-price-%d", i)).
				Attr("name", fmt.Sprintf("ItemPrice_%d", i)).
				Attr("min", "0").
				Attr("step", "0.01").
				Attr("value", "0.00"),
		),
		r.Div("col-span-1 flex items-center justify-center text-gray-400 text-xs pb-1").
			Text(fmt.Sprintf("#%d", i+1)),
	)
}

func CollectIDs(itemCount int) []string {
	ids := []string{"company", "status", "dueDate", "description"}
	for i := 0; i < itemCount; i++ {
		ids = append(ids,
			fmt.Sprintf("item-desc-%d", i),
			fmt.Sprintf("item-qty-%d", i),
			fmt.Sprintf("item-price-%d", i),
		)
	}
	return ids
}

func ItemCountFrom(data map[string]any) int {
	if v, ok := data["ItemCount"]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return 1
}

// ---------------------------------------------------------------------------
// Shared form helpers
// ---------------------------------------------------------------------------

func Th(label, extra string) *r.Node {
	cls := "px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wide"
	if extra != "" {
		cls += " " + extra
	}
	return r.Th().Class(cls).Text(label)
}

func Td(text, extra string) *r.Node {
	cls := "px-4 py-3"
	if extra != "" {
		cls += " " + extra
	}
	return r.Td().Class(cls).Text(text)
}

func InfoField(label, value string) *r.Node {
	return r.Div().Render(
		r.Div("text-xs font-medium text-gray-500 uppercase tracking-wide").Text(label),
		r.Div("text-sm mt-1").Text(value),
	)
}

func TotalField(label, value string) *r.Node {
	return r.Div().Render(
		r.Div("text-xs text-gray-500").Text(label),
		r.Div("text-lg font-semibold mt-0.5 tabular-nums").Text(value),
	)
}

func FormField(label, id, inputType, placeholder string) *r.Node {
	cls := "w-full border border-gray-300 rounded px-3 py-2 text-sm mt-1"
	var input *r.Node
	switch inputType {
	case "text":
		input = r.IText(cls)
	case "password":
		input = r.IPassword(cls)
	case "email":
		input = r.IEmail(cls)
	case "number":
		input = r.INumber(cls)
	case "date":
		input = r.IDate(cls)
	case "time":
		input = r.ITime(cls)
	case "tel":
		input = r.IPhone(cls)
	case "url":
		input = r.IUrl(cls)
	case "search":
		input = r.ISearch(cls)
	case "hidden":
		input = r.IHidden(cls)
	default:
		input = r.Input(cls).Attr("type", inputType)
	}
	return r.Div().Render(
		r.Label("text-sm font-medium text-gray-700").Text(label),
		input.ID(id).Attr("name", id).Attr("placeholder", placeholder),
	)
}

func FormSelect(label, id string, options []string) *r.Node {
	opts := r.Map(options, func(opt string, _ int) *r.Node {
		return r.Option().Attr("value", opt).
			Text(strings.ToUpper(opt[:1]) + opt[1:])
	})
	return r.Div().Render(
		r.Label("text-sm font-medium text-gray-700").Text(label),
		r.Select("w-full border border-gray-300 rounded px-3 py-2 text-sm mt-1").
			ID(id).Attr("name", id).Render(opts...),
	)
}

func FormTextarea(label, id, placeholder string) *r.Node {
	return r.Div().Render(
		r.Label("text-sm font-medium text-gray-700").Text(label),
		r.Textarea("w-full border border-gray-300 rounded px-3 py-2 text-sm mt-1").
			ID(id).
			Attr("name", id).
			Attr("rows", "2").
			Attr("placeholder", placeholder),
	)
}
