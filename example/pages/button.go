package pages

import r "github.com/michalCapo/g-sui/ui"

func Button() *r.Node {
	row := func(title string, content *r.Node) *r.Node {
		return r.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3").Render(
			r.Div("text-sm font-bold text-gray-700").Text(title),
			content,
		)
	}

	ex := func(label string, btn *r.Node) *r.Node {
		return r.Div("flex items-center justify-between gap-4 w-full").Render(
			r.Div("text-sm text-gray-600").Text(label),
			btn,
		)
	}

	// Basics
	basics := r.Div("flex flex-col gap-2").Render(
		ex("Button", r.NewButton("Click me").Build()),
		ex("Button — disabled", r.NewButton("Unavailable").Disabled(true).Build()),
		ex("Button as link", r.NewButton("Visit example.com").Href("https://example.com").Build()),
		ex("Submit button (visual)", r.NewButton("Submit").BtnColor(r.BtnGreen).Submit().Build()),
		ex("Reset button (visual)", r.NewButton("Reset").BtnColor(r.BtnGray).Reset().Build()),
	)

	// Colors — solid + outline
	type colorEntry struct {
		color string
		title string
	}
	solid := []colorEntry{
		{r.BtnBlue, "Blue"},
		{r.BtnGreen, "Green"},
		{r.BtnRed, "Red"},
		{r.BtnPurple, "Purple"},
		{r.BtnYellow, "Yellow"},
		{r.BtnGray, "Gray"},
		{r.BtnWhite, "White"},
	}
	outline := []colorEntry{
		{r.BtnBlueOutline, "Blue (outline)"},
		{r.BtnGreenOutline, "Green (outline)"},
		{r.BtnRedOutline, "Red (outline)"},
		{"bg-transparent hover:bg-purple-50 text-purple-600 border border-purple-600", "Purple (outline)"},
		{"bg-transparent hover:bg-yellow-50 text-yellow-600 border border-yellow-500", "Yellow (outline)"},
		{"bg-transparent hover:bg-gray-50 text-gray-600 border border-gray-600", "Gray (outline)"},
		{"bg-transparent hover:bg-gray-50 text-gray-500 border border-gray-300", "White (outline)"},
	}

	colorBtns := make([]*r.Node, 0, len(solid)+len(outline))
	for _, it := range solid {
		colorBtns = append(colorBtns, r.NewButton(it.title).BtnColor(it.color).BtnClass("w-full").Build())
	}
	for _, it := range outline {
		colorBtns = append(colorBtns, r.NewButton(it.title).BtnColor(it.color).BtnClass("w-full").Build())
	}
	colorsGrid := r.Div("grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2").Render(colorBtns...)

	// Sizes
	type sizeEntry struct {
		size  string
		title string
	}
	sizes := []sizeEntry{
		{r.BtnXS, "Extra small"},
		{r.BtnSM, "Small"},
		{r.BtnMD, "Medium (default)"},
		{r.BtnLG, "Standard"},
		{r.BtnXL, "Large"},
		{"px-8 py-4 text-xl", "Extra large"},
	}
	sizeNodes := make([]*r.Node, 0, len(sizes))
	for _, it := range sizes {
		sizeNodes = append(sizeNodes, ex(it.title,
			r.NewButton("Click me").BtnSize(it.size).Build(),
		))
	}
	sizesGrid := r.Div("flex flex-col gap-2").Render(sizeNodes...)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Button"),
		r.Div("text-gray-600").Text("Common button states and variations. Clicks here are for visual demo only."),
		row("Basics", basics),
		row("Colors (solid and outline)", colorsGrid),
		row("Sizes", sizesGrid),
	)
}

func RegisterButton(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/button", func(ctx *r.Context) *r.Node { return layout(Button()) })
	app.Action("nav.button", NavTo("/button", func() *r.Node { return Button() }))
}
