package pages

import (
	"time"

	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// NewFilters demonstrates advanced table filtering: column filter dropdowns,
// URL sync, localStorage persistence, and async row detail.
func NewFilters(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Advanced Filters"),
		ui.Div("text-gray-600")("Column filter dropdowns with per-type UIs, URL synchronization, and localStorage persistence. Try filtering — the URL updates in real-time. Refresh the page to see state restored."),

		js.Client(ctx).
			Source("/api/new/filterable-invoices").
			Loading(ui.SkeletonTable).
			Empty("filter_list_off", "No invoices match your filters").
			Table(
				js.Col("id").Label("ID").Sortable(true).Filterable(true),
				js.Col("company").Label("Company").Sortable(true).Filterable(true),
				js.Col("amount").Label("Amount").Type("number").Format("amount").Sortable(true).Filterable(true),
				js.Col("date").Label("Date").Type("date").Sortable(true).Filterable(true),
				js.Col("status").Label("Status").Type("enum").Filterable(true).EnumOptions(
					js.Option{Value: "paid", Label: "Paid"},
					js.Option{Value: "pending", Label: "Pending"},
					js.Option{Value: "overdue", Label: "Overdue"},
					js.Option{Value: "cancelled", Label: "Cancelled"},
				),
				js.Col("verified").Label("Verified").Type("bool").Filterable(true),
				js.Col("priority").Label("Priority").Type("number").Sortable(true),
			).
			Search(true).
			Pagination(10).
			Render(),

		// Usage guide
		ui.Div("bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 text-sm")(
			ui.Div("font-semibold text-blue-800 dark:text-blue-300 mb-2")("How it works"),
			ui.Div("text-blue-700 dark:text-blue-400")(
				"Click the filter icon (funnel) in any column header to open a filter dropdown. Each column type has its own UI: text fields for text, number comparisons for numbers, date ranges with presets for dates, checkboxes for enums, and radio buttons for booleans. Active filters turn the icon blue. The URL updates via history.replaceState and state is saved to localStorage.",
			),
		),
	)
}

// NewAsyncDetail demonstrates async expandable row detail and tabbed detail views.
func NewAsyncDetail(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Async Row Detail"),
		ui.Div("text-gray-600")("Click any row to expand it. Detail data is fetched asynchronously from the server and cached."),

		js.Client(ctx).
			Source("/api/new/filterable-invoices").
			Loading(ui.SkeletonTable).
			Table(
				js.Col("id").Label("ID").Sortable(true),
				js.Col("company").Label("Company").Sortable(true),
				js.Col("amount").Label("Amount").Type("number").Format("amount").Sortable(true),
				js.Col("date").Label("Date").Type("date").Sortable(true),
				js.Col("status").Label("Status"),
			).
			Search(true).
			Pagination(8).
			Component("table", js.Opts{
				"columns": []map[string]any{
					{"key": "id", "label": "ID", "sortable": true},
					{"key": "company", "label": "Company", "sortable": true},
					{"key": "amount", "label": "Amount", "type": "number", "format": "amount", "sortable": true},
					{"key": "date", "label": "Date", "type": "date", "sortable": true},
					{"key": "status", "label": "Status"},
				},
				"expandable":   true,
				"detailSource": "/api/new/invoice-detail?id={id}",
				"pageSize":     8,
				"search":       true,
			}).
			Render(),
	)
}

// NewConditionalPolling demonstrates polling that stops when a condition is met.
func NewConditionalPolling(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Conditional Polling"),
		ui.Div("text-gray-600")("The component polls every 2 seconds but stops when status reaches 'completed'. Reload the page to restart."),

		js.Client(ctx).
			Source("/api/new/job-status").
			Loading(ui.SkeletonComponent).
			Poll(2*time.Second).
			PollWhile("data && data.status !== 'completed'").
			Component("kpi-bar", js.Opts{
				"items": []js.Opts{
					{"key": "status", "label": "Status", "icon": "sync"},
					{"key": "progress", "label": "Progress", "icon": "speed"},
					{"key": "message", "label": "Message", "icon": "info"},
					{"key": "updated", "label": "Last Update", "icon": "schedule"},
				},
			}).
			Render(),
	)
}
