package pages

import "github.com/michalCapo/g-sui/ui"

func IconsDemo() string {
	return ui.Div("bg-white p-6 rounded-lg shadow w-full")(
		ui.Div("text-lg font-bold")("Icons"),
		ui.Div("flex flex-col gap-3 mt-2")(
			ui.Div("flex items-center gap-3")(ui.Icon("w-6 h-6 bg-gray-400 rounded"), ui.Div("text-sm text-gray-700")("Icon (basic)")),
			ui.Div("flex items-center gap-3")(ui.Icon2("w-6 h-6 bg-blue-600 rounded", "Centered with icon left")),
			ui.Div("flex items-center gap-3")(ui.Icon3("w-6 h-6 bg-green-600 rounded", "Centered with icon right")),
			ui.Div("flex items-center gap-3")(ui.Icon4("w-6 h-6 bg-purple-600 rounded", "End-aligned icon")),
		),
	)
}

func Others(ctx *ui.Context) string {
	hello := ui.Div("bg-white p-6 rounded-lg shadow w-full")(ui.Div("text-lg font-bold")("Hello"), HelloContent(ctx))
	counter := ui.Div("bg-white p-6 rounded-lg shadow w-full")(ui.Div("text-lg font-bold")("Counter"), CounterContent(ctx))
	login := ui.Div("bg-white p-6 rounded-lg shadow w-full")(ui.Div("text-lg font-bold")("Login"), LoginContent(ctx))
	icons := IconsDemo()

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
		icons,
		markdown,
		captcha,
	)
}
