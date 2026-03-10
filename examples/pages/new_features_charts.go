package pages

import (
	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// NewCharts demonstrates chart enhancements: two-series bar charts and SVG tooltips.
func NewCharts(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Chart Enhancements"),
		ui.Div("text-gray-600")("Two-series bar chart comparison and native SVG tooltips on all chart types. Hover over any bar, dot, or donut segment to see the tooltip."),

		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-6")(
			// Two-series bar chart
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-1")("Revenue: Current vs Previous Year"),
				ui.Div("text-sm text-gray-500 mb-3")("Two-series bar chart with value2 field and legend."),
				js.Client(ctx).
					Source("/api/new/revenue-comparison").
					Chart(js.BarChart).
					ChartOptions(js.Opts{
						"height":      280,
						"valueFormat": "amount",
						"seriesName":  "2025",
						"series2Name": "2024",
					}).
					Render(),
			),

			// Single-series bar with tooltips
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-1")("Monthly Revenue"),
				ui.Div("text-sm text-gray-500 mb-3")("Hover over bars to see native SVG tooltips."),
				js.Client(ctx).
					Source("/api/client-demo/revenue-monthly").
					Chart(js.BarChart).
					ChartOptions(js.Opts{"height": 280, "valueFormat": "amount"}).
					Render(),
			),

			// Area chart with tooltips
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-1")("Trend Line"),
				ui.Div("text-sm text-gray-500 mb-3")("Hover over dots to see values."),
				js.Client(ctx).
					Source("/api/client-demo/trend").
					Chart(js.AreaChart).
					ChartOptions(js.Opts{"height": 280}).
					Render(),
			),

			// Donut chart with tooltips
			ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-4")(
				ui.Div("text-lg font-semibold mb-1")("Revenue by Category"),
				ui.Div("text-sm text-gray-500 mb-3")("Hover segments to see category, value, and percentage."),
				js.Client(ctx).
					Source("/api/client-demo/revenue-category").
					Chart(js.DonutChart).
					ChartOptions(js.Opts{"height": 280}).
					Render(),
			),
		),
	)
}
