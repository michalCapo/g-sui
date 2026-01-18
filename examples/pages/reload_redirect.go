package pages

import "github.com/michalCapo/g-sui/ui"

func ReloadRedirect(ctx *ui.Context) string {
	// Reload example
	reloadPage := func(ctx *ui.Context) string {
		ctx.Success("Reloading page...")
		return ctx.Reload()
	}

	// Redirect examples
	redirectToHome := func(ctx *ui.Context) string {
		ctx.Info("Redirecting to home page...")
		return ctx.Redirect("/")
	}

	redirectToOthers := func(ctx *ui.Context) string {
		ctx.Info("Redirecting to others page...")
		return ctx.Redirect("/others")
	}

	redirectToButton := func(ctx *ui.Context) string {
		ctx.Info("Redirecting to button page...")
		return ctx.Redirect("/button")
	}

	reloadSection := ui.Div("bg-white p-6 rounded-lg shadow w-full")(
		ui.Div("text-lg font-bold mb-4")("Reload Example"),
		ui.Div("text-gray-600 mb-4")("Click the button below to reload the current page. The page will refresh and any state will be reset."),
		ui.Div("flex flex-row gap-4")(
			ui.Button().
				Color(ui.BlueOutline).
				Class("rounded").
				Click(ctx.Call(reloadPage).None()).
				Render("Reload Page"),
		),
	)

	redirectSection := ui.Div("bg-white p-6 rounded-lg shadow w-full")(
		ui.Div("text-lg font-bold mb-4")("Redirect Examples"),
		ui.Div("text-gray-600 mb-4")("Click any button below to redirect to a different page. The browser will navigate to the specified URL."),
		ui.Div("flex flex-row gap-4 flex-wrap")(
			ui.Button().
				Color(ui.GreenOutline).
				Class("rounded").
				Click(ctx.Call(redirectToHome).None()).
				Render("Redirect to Home"),
			ui.Button().
				Color(ui.PurpleOutline).
				Class("rounded").
				Click(ctx.Call(redirectToOthers).None()).
				Render("Redirect to Others"),
			ui.Button().
				Color(ui.YellowOutline).
				Class("rounded").
				Click(ctx.Call(redirectToButton).None()).
				Render("Redirect to Button"),
		),
	)

	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("Reload & Redirect"),
		ui.Div("text-gray-600")("Demonstrates page reload and redirect functionality using ctx.Reload() and ctx.Redirect()."),
		reloadSection,
		redirectSection,
	)
}
