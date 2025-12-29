package pages

import "github.com/michalCapo/g-sui/ui"

func Radio(_ *ui.Context) string {
	card := func(title string, body string) string {
		return ui.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3")(
			ui.Div("text-sm font-bold text-gray-700")(title),
			body,
		)
	}

	selected := struct{ Gender string }{Gender: "male"}
	genders := []ui.AOption{{ID: "male", Value: "Male"}, {ID: "female", Value: "Female"}, {ID: "other", Value: "Other"}}
	customs := []ui.AOption{
		{
			ID: "car1",
			Value: ui.Div("w-full h-48 rounded-xl border border-gray-300 relative overflow-hidden")(
				ui.Div("absolute inset-0 bg-gradient-to-br from-gray-100 to-gray-200 opacity-60")(""),
				ui.Div("absolute top-4 left-4 z-10")(
					ui.Div("text-2xl font-bold text-gray-800")("11-AA"),
					ui.Div("text-lg text-gray-700")("Skoda"),
				),
			),
		},
		{
			ID: "car2",
			Value: ui.Div("w-full h-48 rounded-xl border border-gray-300 relative bg-white")(
				ui.Div("absolute top-4 left-4 z-10")(
					ui.Div("text-2xl font-bold text-gray-800")("22aa"),
				),
			),
		},
		{
			ID: "car3",
			Value: ui.Div("w-full h-48 rounded-xl border border-gray-300 relative overflow-hidden")(
				ui.Div("absolute inset-0 bg-gradient-to-br from-amber-100 to-amber-200 opacity-60")(""),
				ui.Div("absolute top-4 left-4 z-10")(
					ui.Div("text-2xl font-bold text-gray-800")("ABC-123"),
					ui.Div("text-lg text-gray-700")("Volvo"),
				),
			),
		},
	}

	single := ui.Div("flex flex-col gap-2")(
		ui.IRadio("Gender", &selected).Value("male").Render("Male"),
		ui.IRadio("Gender", &selected).Value("female").Render("Female"),
		ui.IRadio("Gender", &selected).Value("other").Render("Other"),
	)

	radiosDefault := ui.IRadioButtons("Group").Options(genders).Render("Gender")
	radiosSelected := ui.IRadioButtons("Group2", &struct{ Group2 string }{Group2: "female"}).Options(genders).Render("Gender")

	validation := ui.Div("flex flex-col gap-2")(
		ui.Div("text-sm text-gray-700")("Required group (no selection)"),
		ui.IRadioButtons("ReqGroup").Options(genders).Required().Render("Gender (required)"),
		ui.Div("text-sm text-gray-700")("Required standalone radios (no selection)"),
		ui.Div("flex flex-col gap-1")(
			ui.IRadio("ReqSingle").Required().Value("a").Render("Option A"),
			ui.IRadio("ReqSingle").Required().Value("b").Render("Option B"),
			ui.IRadio("ReqSingle").Required().Value("c").Render("Option C"),
		),
	)

	sizes := ui.Div("flex flex-col gap-2")(
		ui.IRadio("SizesA").Value("a").Render("Default"),
		ui.IRadio("SizesB").Size(ui.SM).ClassLabel("text-sm").Value("b").Render("Small (SM)"),
		ui.IRadio("SizesC").Size(ui.XS).ClassLabel("text-sm").Value("c").Render("Extra small (XS)"),
	)

	disabled := ui.Div("flex flex-col gap-2")(
		ui.IRadio("DisA").Disabled().Value("a").Render("Disabled A"),
		ui.IRadio("DisB").Disabled().Value("b").Render("Disabled B"),
	)

	custom := ui.Div("flex flex-col gap-2")(
		ui.IRadioDiv("Custom").Options(customs).Render("Custom"),
	)

	return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
		ui.Div("text-3xl font-bold")("Radio"),
		ui.Div("text-gray-600")("Single radio inputs and grouped radio buttons with a selected state."),
		card("Standalone radios (with selection)", single),
		card("Radio buttons group (no selection)", radiosDefault),
		card("Radio buttons group (with selection)", radiosSelected),
		card("Custom radios", custom),
		card("Sizes", sizes),
		card("Validation", validation),
		card("Disabled", disabled),
	)
}
