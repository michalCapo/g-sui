package pages

import "github.com/michalCapo/g-sui/ui"

func Others(ctx *ui.Context) string {
	hello := ui.Div("bg-white p-6 rounded-lg shadow w-full")(
		ui.Div("text-lg font-bold")("Hello"),
		ui.Div("flex flex-row gap-4")(
			HelloContent(ctx),
		),
	)
	counter := ui.Div("bg-white p-6 rounded-lg shadow w-full")(
		ui.Div("text-lg font-bold")("Counter"),
		ui.Div("flex flex-row gap-4")(
			Counter(3).render(ctx),
			Counter(5).render(ctx),
		),
	)
	login := ui.Div("bg-white p-6 rounded-lg shadow w-full")(
		ui.Div("text-lg font-bold")("Login"),
		ui.Div("flex flex-row gap-4")(
			LoginContent(ctx),
		),
	)

	markdown := ui.Div("bg-white p-6 rounded-lg shadow flex flex-col gap-3 w-full")(
		ui.Div("text-xl font-bold")("Markdown"),
		ui.Markdown("prose prose-sm sm:prose max-w-none")(`# Heading

- Item 1
- Item 2

**Bold** and _italic_.`),
	)

	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("Others"),
		ui.Div("text-gray-600")("Miscellaneous demos: Hello, Counter, Login, and icon helpers."),
		hello,
		counter,
		login,
		markdown,
	)
}
