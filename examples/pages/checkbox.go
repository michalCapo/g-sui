package pages

import "github.com/michalCapo/g-sui/ui"

func CheckboxContent(_ *ui.Context) string {
    row := func(title string, content string) string {
        return ui.Div("bg-white p-4 rounded-lg shadow border border-gray-200 flex flex-col gap-3")(
            ui.Div("text-sm font-bold text-gray-700")(title),
            content,
        )
    }

    ex := func(label string, control string) string {
        return ui.Div("flex items-center justify-between gap-4 w-full")(
            ui.Div("text-sm text-gray-600")(label),
            control,
        )
    }

    data := struct{ Agree bool }{Agree: true}

    basics := ui.Div("flex flex-col gap-2")(
        ex("Default", ui.ICheckbox("Agree", &data).Render("I agree")),
        ex("Required", ui.ICheckbox("Terms").Required().Render("Accept terms")),
        ex("Unchecked", ui.ICheckbox("X").Render("Unchecked")),
        ex("Disabled", ui.ICheckbox("D").Disabled().Render("Disabled")),
    )

    sizes := ui.Div("flex flex-col gap-2")(
        ex("Small (SM)", ui.ICheckbox("S").Size(ui.SM).Render("Small")),
        ex("Extra small (XS)", ui.ICheckbox("XS").Size(ui.XS).Render("Extra small")),
    )

    return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
        ui.Div("text-3xl font-bold")("Checkbox"),
        ui.Div("text-gray-600")("Checkbox states, sizes, and required validation."),
        row("Basics", basics),
        row("Sizes", sizes),
    )
}
