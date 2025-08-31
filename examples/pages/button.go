package pages

import "github.com/michalCapo/g-sui/ui"

func ButtonContent(_ *ui.Context) string {
    row := func(title string, content string) string {
        return ui.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3")(
            ui.Div("text-sm font-bold text-gray-700")(title),
            content,
        )
    }

    ex := func(label string, btn string) string {
        return ui.Div("flex items-center justify-between gap-4 w-full")(
            ui.Div("text-sm text-gray-600")(label),
            btn,
        )
    }

    // Colors grid: solid + outline
    solid := []struct{ c, t string }{
        {ui.Blue, "Blue"},
        {ui.Green, "Green"},
        {ui.Red, "Red"},
        {ui.Purple, "Purple"},
        {ui.Yellow, "Yellow"},
        {ui.Gray, "Gray"},
        {ui.White, "White"},
    }
    outline := []struct{ c, t string }{
        {ui.BlueOutline, "Blue (outline)"},
        {ui.GreenOutline, "Green (outline)"},
        {ui.RedOutline, "Red (outline)"},
        {ui.PurpleOutline, "Purple (outline)"},
        {ui.YellowOutline, "Yellow (outline)"},
        {ui.GrayOutline, "Gray (outline)"},
        {ui.WhiteOutline, "White (outline)"},
    }

    var colorsGrid string
    for _, it := range solid {
        colorsGrid += ui.Button().Color(it.c).Class("rounded w-full").Render(it.t)
    }
    for _, it := range outline {
        colorsGrid += ui.Button().Color(it.c).Class("rounded w-full").Render(it.t)
    }
    colorsGrid = ui.Div("grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2")(colorsGrid)

    // Sizes
    sizes := []struct{ k string; t string }{
        {ui.XS, "Extra small"},
        {ui.SM, "Small"},
        {ui.MD, "Medium (default)"},
        {ui.ST, "Standard"},
        {ui.LG, "Large"},
        {ui.XL, "Extra large"},
    }
    var sizesGrid string
    for _, it := range sizes {
        sizesGrid += ex(it.t, ui.Button().Size(it.k).Class("rounded").Color(ui.Blue).Render("Click me"))
    }
    sizesGrid = ui.Div("flex flex-col gap-2")(sizesGrid)

    // Basics
    basics := ui.Div("flex flex-col gap-2")(
        ex("Button", ui.Button().Class("rounded").Color(ui.Blue).Render("Click me")),
        ex("Button — disabled", ui.Button().Disabled(true).Class("rounded").Color(ui.Blue).Render("Unavailable")),
        ex("Button as link", ui.Button().Href("https://example.com").Class("rounded").Color(ui.Blue).Render("Visit example.com")),
        ex("Submit button (visual)", ui.Button().Submit().Class("rounded").Color(ui.Green).Render("Submit")),
        ex("Reset button (visual)", ui.Button().Reset().Class("rounded").Color(ui.Gray).Render("Reset")),
    )

    return ui.Div("max-w-full sm:max-w-5xl mx-auto flex flex-col gap-6")(
        ui.Div("text-3xl font-bold")("Button"),
        ui.Div("text-gray-600")("Common button states and variations. Clicks here are for visual demo only."),
        row("Basics", basics),
        row("Colors (solid and outline)", colorsGrid),
        row("Sizes", sizesGrid),
    )
}
