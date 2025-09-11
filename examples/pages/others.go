package pages

import "github.com/michalCapo/g-sui/ui"

func Others(ctx *ui.Context) string {
	hello := ui.Div("bg-white p-6 rounded-lg shadow w-full")(ui.Div("text-lg font-bold")("Hello"), HelloContent(ctx))
	counter := ui.Div("bg-white p-6 rounded-lg shadow w-full")(ui.Div("text-lg font-bold")("Counter"), CounterContent(ctx))
	login := ui.Div("bg-white p-6 rounded-lg shadow w-full")(ui.Div("text-lg font-bold")("Login"), LoginContent(ctx))

	markdown := ui.Div("bg-white p-6 rounded-lg shadow flex flex-col gap-3 w-full")(
		ui.Div("text-xl font-bold")("Markdown"),
		ui.Markdown("prose prose-sm sm:prose max-w-none")(`# Heading

- Item 1
- Item 2

**Bold** and _italic_.`),
	)

	captcha := ui.Div("bg-white p-6 rounded-lg shadow flex flex-col gap-3 w-full")(
		ui.Div("text-xl font-bold")("Client CAPTCHA (demo)"),
		ui.Div("w-full overflow-x-auto")(ui.Captcha2()),
	)

	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("Others"),
		ui.Div("text-gray-600")("Miscellaneous demos: Hello, Counter, Login, and icon helpers."),
		hello,
		counter,
		login,
		markdown,
		captcha,
	)
}
