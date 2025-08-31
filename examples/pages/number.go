package pages

import "github.com/michalCapo/g-sui/ui"

func NumberContent(_ *ui.Context) string {
    card := func(title string, body string) string {
        return ui.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3")(
            ui.Div("text-sm font-bold text-gray-700")(title),
            body,
        )
    }

    row := func(label string, control string) string {
        return ui.Div("flex items-center justify-between gap-4")(
            ui.Div("text-sm text-gray-600")(label),
            ui.Div("w-64")(control),
        )
    }

    data := struct {
        Age   int
        Price float64
    }{Age: 30, Price: 19.9}

    basics := ui.Div("flex flex-col gap-2")(
        row("Integer with range/step", ui.INumber("Age", &data).Numbers(0, 120, 1).Render("Age")),
        row("Float formatted (%.2f)", ui.INumber("Price", &data).Format("%.2f").Render("Price")),
        row("Required", ui.INumber("Req").Required().Render("Required")),
        row("Readonly", ui.INumber("RO").Readonly().Value("42").Render("Readonly")),
        row("Disabled", ui.INumber("D").Disabled().Render("Disabled")),
        row("Placeholder", ui.INumber("PH").Placeholder("0..100").Render("Number")),
    )

    styling := ui.Div("flex flex-col gap-2")(
        row("Wrapper .Class()", ui.INumber("C").Class("p-2 rounded bg-yellow-50").Render("Styled wrapper")),
        row("Label .ClassLabel()", ui.INumber("CL").ClassLabel("text-purple-700 font-bold").Render("Custom label")),
        row("Input .ClassInput()", ui.INumber("CI").ClassInput("bg-blue-50").Render("Custom input background")),
        row("Size: LG", ui.INumber("S").Size(ui.LG).Render("Large size")),
    )

    behavior := ui.Div("flex flex-col gap-2")(
        row("Change handler (console.log)", ui.INumber("Change").Change("console.log('changed', this && this.value)").Render("On change, log")),
        row("Click handler (console.log)", ui.INumber("Click").Click("console.log('clicked number')").Render("On click, log")),
    )

    return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
        ui.Div("text-3xl font-bold")("Number input"),
        ui.Div("text-gray-600")("Ranges, formatting, and common attributes."),
        card("Basics & states", basics),
        card("Styling", styling),
        card("Behavior & attributes", behavior),
    )
}
