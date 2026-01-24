package pages

import (
	"fmt"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

func SpaExample(ctx *ui.Context) string {
	target := ui.Target()

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Single Page Application (SPA)"),
		ui.Div("text-gray-600")("This page demonstrates g-sui's SPA capabilities using explicit navigation via ctx.Load()."),

		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-6")(
			// Feature 1: Smooth Transitions
			ui.Div("bg-white dark:bg-gray-800 p-6 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700")(
				ui.Div("text-xl font-semibold mb-2")("Seamless Transitions"),
				ui.P("text-gray-500 mb-4")("Navigate between pages without a full browser reload. The scroll position and application state can be preserved better than with traditional multi-page apps."),
				ui.A("text-blue-600 hover:underline", ui.Href("/"))("Back to Showcase (Smoothly)"),
			),

			// Feature 2: Background Loading
			ui.Div("bg-white dark:bg-gray-800 p-6 rounded-xl shadow-sm border border-gray-100 dark:border-gray-700")(
				ui.Div("text-xl font-semibold mb-2")("Background Loading"),
				ui.P("text-gray-500 mb-4")("Resources are fetched in the background. A smart loader appears only if the transition takes longer than 50ms."),
				ui.Button().
					Color(ui.Blue).
					Click(ctx.Call(func(ctx *ui.Context) string {
						time.Sleep(1 * time.Second) // Simulate lag
						return ui.Div("text-green-600 font-medium")("Content loaded after simulation!")
					}).Render(target)).
					Render("Trigger Delayed Content"),
				ui.Div("mt-4", target)(),
			),
		),

		ui.Div("bg-blue-50 dark:bg-blue-900/20 p-6 rounded-xl border border-blue-100 dark:border-blue-800")(
			ui.Div("font-semibold text-blue-800 dark:text-blue-300 mb-2")("How it works"),
			ui.Div("text-blue-700 dark:text-blue-400 text-sm space-y-2")(
				ui.P("")("1. Use `ctx.Load(\"/path\")` to enable smooth navigation on specific links."),
				ui.P("")("2. Clicking a link with `ctx.Load()` triggers a background `fetch`."),
				ui.P("")("3. The server returns the partial or full HTML."),
				ui.P("")("4. The client updates the DOM and browser history."),
			),
		),

		ui.Div("mt-10 pt-6 border-t border-gray-100 dark:border-gray-800 text-center")(
			ui.Div("text-gray-400 text-xs")(fmt.Sprintf("Page rendered at %s", time.Now().Format("15:04:05"))),
		),
	)
}
