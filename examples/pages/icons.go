package pages

import "github.com/michalCapo/g-sui/ui"

func Icons() string {
	icon := ui.Div("flex items-center gap-3 border rounded p-4")

	return ui.Div("bg-white rounded-lg shadow w-full")(
		ui.Div("flex flex-col gap-3")(
			icon(
				ui.IconBasic("w-6 h-6 bg-gray-400 rounded", "Basic icon"),
			),
			icon(
				ui.IconStart("w-6 h-6 bg-gray-400 rounded", "Start aligned icon"),
			),
			icon(
				ui.IconLeft("w-6 h-6 bg-blue-600 rounded", "Centered with icon left"),
			),
			icon(
				ui.IconRight("w-6 h-6 bg-green-600 rounded", "Centered with icon right"),
			),
			icon(
				ui.IconEnd("w-6 h-6 bg-purple-600 rounded", "End-aligned icon"),
			),
		),
	)
}

func IconsContent(ctx *ui.Context) string {
	return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Icons"),
		ui.Div("text-gray-600")(
			"Icon positioning helpers and layouts.",
		),
		Icons(),
	)
}
