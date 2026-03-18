package pages

import r "github.com/michalCapo/g-sui/ui"

func icon(name, colorCls string) *r.Node {
	return r.I("material-icons-round " + colorCls).Text(name)
}

func Icons(ctx *r.Context) *r.Node {
	iconRow := func(children ...*r.Node) *r.Node {
		return r.Div("relative flex items-center justify-center border rounded-lg p-4 bg-white border-gray-300 min-h-[3rem]").Render(children...)
	}

	return r.Div("max-w-6xl mx-auto flex flex-col gap-8 w-full").Render(
		r.Div("text-3xl font-bold").Text("Icons"),
		r.Div("text-gray-600").Text("Material icons, positioning helpers, and icon+text pairs."),

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
