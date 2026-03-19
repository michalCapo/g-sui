package pages

import r "github.com/michalCapo/g-sui/ui"

func icon(name, colorCls string) *r.Node {
	return r.I("material-icons-round " + colorCls).Text(name)
}

// svgIcon wraps an inline SVG in a flex-centered container with a label.
func svgIcon(label string, svg *r.Node) *r.Node {
	return r.Div("flex flex-col items-center gap-2").Render(
		svg,
		r.Span("text-xs text-gray-500").Text(label),
	)
}

func Icons(ctx *r.Context) *r.Node {
	iconRow := func(children ...*r.Node) *r.Node {
		return r.Div("relative flex items-center justify-center border rounded-lg p-4 bg-white border-gray-300 min-h-[3rem]").Render(children...)
	}

	return r.Div("max-w-6xl mx-auto flex flex-col gap-8 w-full").Render(
		r.Div("text-3xl font-bold").Text("Icons"),
		r.Div("text-gray-600").Text("Material icons, inline SVG, positioning helpers, and icon+text pairs."),

		// Standalone icons
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase").Text("Standalone Icons"),
			r.Div("flex flex-wrap items-center gap-6").Render(
				r.Icon("home", "text-2xl text-gray-700"),
				r.Icon("settings", "text-2xl text-blue-600"),
				r.Icon("favorite", "text-2xl text-red-500"),
				r.Icon("star", "text-2xl text-yellow-500"),
				r.Icon("delete", "text-2xl text-gray-400"),
				r.Icon("check_circle", "text-2xl text-green-600"),
			),
		),

		// Inline SVG icons — uses createElementNS under the hood
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase").Text("Inline SVG Icons"),
			r.Div("text-xs text-gray-400").Text("Built with r.SVG() — proper createElementNS, no innerHTML workaround."),
			r.Div("flex flex-wrap items-center gap-8").Render(

				// Checkmark circle
				svgIcon("Check",
					r.SVG("w-8 h-8 text-green-600").
						Attr("viewBox", "0 0 24 24").Attr("fill", "none").
						Attr("stroke", "currentColor").Attr("stroke-width", "2").
						Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").
						Render(
							r.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10"),
							r.El("path").Attr("d", "M9 12l2 2 4-4"),
						),
				),

				// X circle
				svgIcon("Close",
					r.SVG("w-8 h-8 text-red-500").
						Attr("viewBox", "0 0 24 24").Attr("fill", "none").
						Attr("stroke", "currentColor").Attr("stroke-width", "2").
						Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").
						Render(
							r.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10"),
							r.El("path").Attr("d", "M15 9l-6 6"),
							r.El("path").Attr("d", "M9 9l6 6"),
						),
				),

				// Warning triangle
				svgIcon("Warning",
					r.SVG("w-8 h-8 text-yellow-500").
						Attr("viewBox", "0 0 24 24").Attr("fill", "none").
						Attr("stroke", "currentColor").Attr("stroke-width", "2").
						Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").
						Render(
							r.El("path").Attr("d", "M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"),
							r.El("line").Attr("x1", "12").Attr("y1", "9").Attr("x2", "12").Attr("y2", "13"),
							r.El("line").Attr("x1", "12").Attr("y1", "17").Attr("x2", "12.01").Attr("y2", "17"),
						),
				),

				// Info circle
				svgIcon("Info",
					r.SVG("w-8 h-8 text-blue-500").
						Attr("viewBox", "0 0 24 24").Attr("fill", "none").
						Attr("stroke", "currentColor").Attr("stroke-width", "2").
						Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").
						Render(
							r.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10"),
							r.El("line").Attr("x1", "12").Attr("y1", "16").Attr("x2", "12").Attr("y2", "12"),
							r.El("line").Attr("x1", "12").Attr("y1", "8").Attr("x2", "12.01").Attr("y2", "8"),
						),
				),

				// Heart (filled)
				svgIcon("Heart",
					r.SVG("w-8 h-8 text-pink-500").
						Attr("viewBox", "0 0 24 24").Attr("fill", "currentColor").
						Render(
							r.El("path").Attr("d", "M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78L12 21.23l8.84-8.84a5.5 5.5 0 000-7.78z"),
						),
				),

				// Star (filled)
				svgIcon("Star",
					r.SVG("w-8 h-8 text-yellow-400").
						Attr("viewBox", "0 0 24 24").Attr("fill", "currentColor").
						Render(
							r.El("polygon").Attr("points", "12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"),
						),
				),

				// Arrow right
				svgIcon("Arrow",
					r.SVG("w-8 h-8 text-gray-600").
						Attr("viewBox", "0 0 24 24").Attr("fill", "none").
						Attr("stroke", "currentColor").Attr("stroke-width", "2").
						Attr("stroke-linecap", "round").Attr("stroke-linejoin", "round").
						Render(
							r.El("line").Attr("x1", "5").Attr("y1", "12").Attr("x2", "19").Attr("y2", "12"),
							r.El("polyline").Attr("points", "12 5 19 12 12 19"),
						),
				),

				// Spinner (animated)
				svgIcon("Spinner",
					r.SVG("w-8 h-8 text-indigo-500 animate-spin").
						Attr("viewBox", "0 0 24 24").Attr("fill", "none").
						Render(
							r.El("circle").Attr("cx", "12").Attr("cy", "12").Attr("r", "10").
								Attr("stroke", "currentColor").Attr("stroke-width", "3").
								Style("opacity", "0.25"),
							r.El("path").Attr("d", "M4 12a8 8 0 018-8").
								Attr("stroke", "currentColor").Attr("stroke-width", "3").
								Attr("stroke-linecap", "round"),
						),
				),
			),
		),

		// Icon + text pairs
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase").Text("Icon + Text Pairs"),
			r.Div("flex flex-wrap items-center gap-6").Render(
				r.IconText("home", "Home", "text-gray-700"),
				r.IconText("settings", "Settings", "text-blue-600"),
				r.IconText("favorite", "Favorites", "text-red-500"),
				r.IconText("notifications", "Alerts", "text-yellow-600"),
			),
		),

		// Positioning helpers
		r.Div("flex flex-col gap-4").Render(
			r.Div("text-sm font-bold text-gray-500 uppercase").Text("Positioning Helpers"),
			r.Div("flex flex-col gap-3").Render(
				iconRow(
					r.Div("absolute left-4 top-0 bottom-0 flex items-center").Render(icon("home", "text-gray-600")),
					r.Span("text-center").Text("Start aligned icon"),
				),
				iconRow(
					r.Span("inline-flex items-center gap-2").Render(
						icon("person", "text-blue-600"),
						r.Span().Text("Centered with icon left"),
					),
				),
				iconRow(
					r.Span("inline-flex items-center gap-2").Render(
						r.Span().Text("Centered with icon right"),
						icon("check_circle", "text-green-600"),
					),
				),
				iconRow(
					r.Span("text-center").Text("End-aligned icon"),
					r.Div("absolute right-4 top-0 bottom-0 flex items-center").Render(icon("settings", "text-purple-600")),
				),
			),
		),
	)
}

func RegisterIcons(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/icons", func(ctx *r.Context) *r.Node { return layout(Icons(ctx)) })
	app.Action("nav.icons", NavTo("/icons", func() *r.Node { return Icons(nil) }))
}
