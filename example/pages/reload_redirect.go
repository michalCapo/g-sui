package pages

import r "github.com/michalCapo/g-sui/ui"

func ReloadRedirect(ctx *r.Context) *r.Node {
	return r.Div("max-w-6xl mx-auto flex flex-col gap-6 w-full").Render(
		r.Div("text-3xl font-bold").Text("Reload & Redirect"),
		r.Div("text-gray-600").Text("Demonstrates page reload and redirect functionality."),

		r.Div("bg-white p-6 rounded-lg shadow w-full").Render(
			r.Div("text-lg font-bold mb-4").Text("Reload Example"),
			r.Div("text-gray-600 mb-4").Text("Click the button below to reload the current page."),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-blue-600 text-blue-600 hover:bg-blue-50 text-sm").
				Text("Reload Page").
				OnClick(r.JS("location.reload()")),
		),

		r.Div("bg-white p-6 rounded-lg shadow w-full").Render(
			r.Div("text-lg font-bold mb-4").Text("Redirect Examples"),
			r.Div("text-gray-600 mb-4").Text("Click any button to redirect to a different page."),
			r.Div("flex flex-row gap-4 flex-wrap").Render(
				r.Button("px-4 py-2 rounded cursor-pointer border-2 border-green-600 text-green-600 hover:bg-green-50 text-sm").
					Text("Redirect to Dashboard").
					OnClick(&r.Action{Name: "redirect.dashboard"}),
				r.Button("px-4 py-2 rounded cursor-pointer border-2 border-yellow-600 text-yellow-600 hover:bg-yellow-50 text-sm").
					Text("Redirect to Button").
					OnClick(&r.Action{Name: "redirect.button"}),
			),
		),
	)
}

func HandleRedirectDashboard(ctx *r.Context) string {
	return r.Notify("info", "Redirecting to dashboard...") + r.Redirect("/")
}

func HandleRedirectButton(ctx *r.Context) string {
	return r.Notify("info", "Redirecting to button page...") + r.Redirect("/button")
}

func RegisterReloadRedirect(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/reload-redirect", func(ctx *r.Context) *r.Node { return layout(ReloadRedirect(ctx)) })
	app.Action("nav.reload", NavTo("/reload-redirect", func() *r.Node { return ReloadRedirect(nil) }))
	app.Action("redirect.dashboard", HandleRedirectDashboard)
	app.Action("redirect.button", HandleRedirectButton)
}
