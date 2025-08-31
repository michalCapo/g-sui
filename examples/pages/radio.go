package pages

import "github.com/michalCapo/go-srui/ui"

func RadioContent(_ *ui.Context) string {
    card := func(title string, body string) string {
        return ui.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3")(
            ui.Div("text-sm font-bold text-gray-700")(title),
            body,
        )
    }

    genders := []ui.AOption{{ID: "male", Value: "Male"}, {ID: "female", Value: "Female"}, {ID: "other", Value: "Other"}}
    selected := struct{ Gender string }{Gender: "male"}

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

    return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
        ui.Div("text-3xl font-bold")("Radio"),
        ui.Div("text-gray-600")("Single radio inputs and grouped radio buttons with a selected state."),
        card("Standalone radios (with selection)", single),
        card("Radio buttons group (no selection)", radiosDefault),
        card("Radio buttons group (with selection)", radiosSelected),
        card("Sizes", sizes),
        card("Validation", validation),
        card("Disabled", disabled),
    )
}

