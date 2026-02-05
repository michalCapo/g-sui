package pages

import "github.com/michalCapo/g-sui/ui"

func Icons() string {
	icon := ui.Div("flex items-center gap-3 border rounded p-4 bg-white rounded-lg border-gray-300")

	return ui.Div("")(
		ui.Div("flex flex-col gap-3")(
			icon(
				ui.IconStart("home text-gray-600", "Start aligned icon"),
			),
			icon(
				ui.IconLeft("person text-blue-600", "Centered with icon left"),
			),
			icon(
				ui.IconRight("check_circle text-green-600", "Centered with icon right"),
			),
			icon(
				ui.IconEnd("settings text-purple-600", "End-aligned icon"),
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
