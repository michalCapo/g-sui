package pages

import (
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// ClientTable demonstrates a client-side rendered table with sorting, search and pagination.
// Data is fetched from /api/client-demo/invoices and all interactions happen client-side.
func ClientTable(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-3xl font-bold")("Client Table"),
		ui.Div("text-gray-600")("Data fetched from API, sorted/filtered/paginated entirely on the client."),
		ui.Client(ctx).
			Source("/api/client-demo/invoices").
			Loading(ui.SkeletonTable).
			Empty("receipt_long", "No invoices found").
			Table(
				ui.ClientCol("Number").Label("#").Sortable(true),
				ui.ClientCol("Company").Label("Company").Sortable(true),
				ui.ClientCol("Amount").Label("Amount").Type("number").Format("amount").Sortable(true),
				ui.ClientCol("Date").Label("Date").Type("date").Sortable(true),
				ui.ClientCol("Status").Label("Status"),
			).
			Search(true).
			Pagination(8).
			Render(),
	)
}

// ClientCharts demonstrates all four chart types: bar, donut, area, and horizontal bar.
// Each chart fetches from a different API endpoint and renders as pure SVG.
func ClientCharts(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Client Charts"),
		ui.Div("text-gray-600")("SVG charts rendered client-side from API data."),
		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-6")(
			// Bar chart
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-2")("Monthly Revenue"),
				ui.Client(ctx).
					Source("/api/client-demo/revenue-monthly").
					Chart(ui.BarChart).
					ChartOptions(ui.ClientOpts{"height": 250, "valueFormat": "amount"}).
					Render(),
			),
			// Donut chart
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-2")("Revenue by Category"),
				ui.Client(ctx).
					Source("/api/client-demo/revenue-category").
					Chart(ui.DonutChart).
					ChartOptions(ui.ClientOpts{"height": 250}).
					Render(),
			),
			// Area chart
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-2")("Trend"),
				ui.Client(ctx).
					Source("/api/client-demo/trend").
					Chart(ui.AreaChart).
					ChartOptions(ui.ClientOpts{"height": 250}).
					Render(),
			),
			// Horizontal bar chart
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-2")("Top Customers"),
				ui.Client(ctx).
					Source("/api/client-demo/top-customers").
					Chart(ui.HBarChart).
					ChartOptions(ui.ClientOpts{"height": 250, "valueFormat": "amount"}).
					Render(),
			),
		),
	)
}

// ClientDashboard shows multiple independent client zones on one page:
// a KPI stats bar, a revenue chart, and a recent invoices table.
func ClientDashboard(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Client Dashboard"),
		ui.Div("text-gray-600")("Multiple independent client zones on one page: stats, chart, and table."),

		// KPI stats row - using generic component
		ui.Client(ctx).
			Source("/api/client-demo/stats").
			Loading(ui.SkeletonCards).
			Component("kpi-bar", ui.ClientOpts{
				"items": []ui.ClientOpts{
					{"key": "total", "label": "Total Revenue", "format": "amount", "icon": "payments"},
					{"key": "count", "label": "Invoices", "icon": "receipt_long"},
					{"key": "avg", "label": "Average", "format": "amount", "icon": "trending_up"},
					{"key": "overdue", "label": "Overdue", "icon": "warning", "color": "red"},
				},
			}).
			Render(),

		// Revenue chart
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
			ui.Div("text-lg font-semibold mb-2")("Revenue Over Time"),
			ui.Client(ctx).
				Source("/api/client-demo/revenue-monthly").
				Chart(ui.BarChart).
				ChartOptions(ui.ClientOpts{"height": 280, "valueFormat": "amount"}).
				Render(),
		),

		// Recent invoices table
		ui.Client(ctx).
			Source("/api/client-demo/invoices").
			Loading(ui.SkeletonTable).
			Table(
				ui.ClientCol("Number").Label("#").Sortable(true),
				ui.ClientCol("Company").Label("Company").Sortable(true),
				ui.ClientCol("Amount").Label("Amount").Type("number").Format("amount").Sortable(true),
				ui.ClientCol("Date").Label("Date").Type("date").Sortable(true),
				ui.ClientCol("Status").Label("Status"),
			).
			Pagination(5).
			Render(),
	)
}

// ClientPolling demonstrates auto-refreshing data with Poll().
// The table re-fetches from the API every 3 seconds, showing live process data.
func ClientPolling(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-3xl font-bold")("Client Polling"),
		ui.Div("text-gray-600")("Table auto-refreshes every 3 seconds. Watch the timestamps update."),
		ui.Client(ctx).
			Source("/api/client-demo/live").
			Loading(ui.SkeletonTable).
			Poll(3*time.Second).
			Table(
				ui.ClientCol("Name").Label("Process").Sortable(true),
				ui.ClientCol("Status").Label("Status"),
				ui.ClientCol("CPU").Label("CPU %").Type("number"),
				ui.ClientCol("Memory").Label("Memory").Format("amount"),
				ui.ClientCol("Updated").Label("Last Update").Type("date"),
			).
			Render(),
	)
}

// ClientEmpty demonstrates empty and error states.
// One zone returns no data (empty state), the other hits a failing endpoint (error state).
func ClientEmpty(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Client States"),
		ui.Div("text-gray-600")("Demonstrates loading, empty, and error states."),
		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-6")(
			ui.Div("")(
				ui.Div("text-lg font-semibold mb-2")("Empty State"),
				ui.Client(ctx).
					Source("/api/client-demo/empty").
					Loading(ui.SkeletonTable).
					Empty("search_off", "No results match your criteria").
					Table(ui.ClientCol("Name").Label("Name")).
					Render(),
			),
			ui.Div("")(
				ui.Div("text-lg font-semibold mb-2")("Error State"),
				ui.Client(ctx).
					Source("/api/client-demo/error").
					Loading(ui.SkeletonTable).
					Table(ui.ClientCol("Name").Label("Name")).
					Render(),
			),
		),
	)
}
