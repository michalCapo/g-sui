package pages

import "github.com/michalCapo/g-sui/ui"

func Select(ctx *ui.Context) string {
	row := func(title string, content string) string {
		return ui.Div("bg-white p-4 rounded-lg shadow border border-gray-200 flex flex-col gap-3")(
			ui.Div("text-sm font-bold text-gray-700")(title),
			content,
		)
	}

	ex := func(label string, control string, extra ...string) string {
		var trailing string
		if len(extra) > 0 {
			trailing = extra[0]
		}
		return ui.Div("flex items-center justify-between gap-4 w-full")(
			ui.Div("text-sm text-gray-600")(label),
			ui.Div("flex items-center gap-3")(ui.Div("w-64")(control), trailing),
		)
	}

	opts := []ui.AOption{{ID: "", Value: "Select..."}, {ID: "one", Value: "One"}, {ID: "two", Value: "Two"}, {ID: "three", Value: "Three"}}
	data := struct{ Country string }{Country: ""}
	optsNoPlaceholder := []ui.AOption{{ID: "one", Value: "One"}, {ID: "two", Value: "Two"}, {ID: "three", Value: "Three"}}

	target := ui.Target()
	onCountryChange := func(inner *ui.Context) string {
		_ = inner.Body(&data)
		return ui.Div("text-sm text-gray-700")("Selected: " + func() string {
			if data.Country == "" {
				return "(none)"
			}
			return data.Country
		}())
	}

	basics := ui.Div("flex flex-col gap-2")(
		ex("Default", ui.ISelect("Country", &data).Options(opts).Render("Country")),
		ex(
			"Placeholder + change",
			ui.ISelect("Country", &data).Options(opts).Placeholder("Pick one").Change(ctx.Call(onCountryChange, &data).Replace(target)).Render("Choose"),
			ui.Div("", target)(""),
		),
	)

	validation := ui.Div("flex flex-col gap-2")(
		ex("Error state", ui.ISelect("Err").Options(opts).Placeholder("Please select").Required(true).Render("Invalid")),
		ex("Required + empty", ui.ISelect("Z").Options(opts).Empty().Required().Render("Required")),
		ex("Disabled", ui.ISelect("Y").Options(opts).Disabled().Render("Disabled")),
	)

	variants := ui.Div("flex flex-col gap-2")(
		ex("No placeholder + <empty>", ui.ISelect("Country", &data).Options(optsNoPlaceholder).Empty().Render("Choose")),
	)

	return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Select"),
		ui.Div("text-gray-600")("Select input variations, validation, and sizing."),
		row("Basics", basics),
		row("Validation", validation),
		row("Variants", variants),
	)
}
