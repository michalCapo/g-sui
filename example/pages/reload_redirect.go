package pages

import r "github.com/michalCapo/g-sui/ui"

var redirectDashboard, redirectButton r.ActionRef[struct{}]

func ReloadRedirect(_ *r.Context) *r.Node {
	return r.Div("max-w-6xl mx-auto flex flex-col gap-6 w-full").Render(
		r.Div("text-3xl font-bold").Text("Reload & Redirect"),
		r.Div("text-gray-600").Text("Demonstrates page reload and redirect functionality."),

		r.Div("bg-white p-6 rounded-lg shadow w-full").Render(
			r.Div("text-lg font-bold mb-4").Text("Reload Example"),
			r.Div("text-gray-600 mb-4").Text("Click the button below to reload the current page."),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-blue-600 text-blue-600 hover:bg-blue-50 text-sm").
				Text("Reload Page").
				OnClick(r.Reload()),
		),

		r.Div("bg-white p-6 rounded-lg shadow w-full").Render(
			r.Div("text-lg font-bold mb-4").Text("Redirect Examples"),
			r.Div("text-gray-600 mb-4").Text("Click any button to redirect to a different page."),
			r.Div("flex flex-row gap-4 flex-wrap").Render(
				r.Button("px-4 py-2 rounded cursor-pointer border-2 border-green-600 text-green-600 hover:bg-green-50 text-sm").
					Text("Redirect to Dashboard").
					OnClick(redirectDashboard.Call(struct{}{})),
				r.Button("px-4 py-2 rounded cursor-pointer border-2 border-yellow-600 text-yellow-600 hover:bg-yellow-50 text-sm").
					Text("Redirect to Button").
					OnClick(redirectButton.Call(struct{}{})),
			),
		),
	)
}

func HandleRedirectDashboard(_ *r.Context, _ struct{}) (r.Result, error) {
	return r.Result{}.Notify("info", "Redirecting to dashboard...").Navigate("/"), nil
}

func HandleRedirectButton(_ *r.Context, _ struct{}) (r.Result, error) {
	return r.Result{}.Notify("info", "Redirecting to button page...").Navigate("/button"), nil
}

func RegisterReloadRedirect(app *r.App) {
	redirectDashboard = r.RegisterAction(app, "redirect.dashboard", HandleRedirectDashboard)
	redirectButton = r.RegisterAction(app, "redirect.button", HandleRedirectButton)
	app.Page("/reload-redirect", ReloadRedirect)
}
