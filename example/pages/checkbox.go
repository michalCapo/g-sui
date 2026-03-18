package pages

import r "github.com/michalCapo/g-sui/ui"

func Checkbox(ctx *r.Context) *r.Node {
	row := func(title string, content *r.Node) *r.Node {
		return r.Div("bg-white p-4 rounded-lg shadow border border-gray-200 flex flex-col gap-3").Render(
			r.Div("text-sm font-bold text-gray-700").Text(title),
			content,
		)
	}

	ex := func(label string, control *r.Node) *r.Node {
		return r.Div("flex items-center justify-between gap-4 w-full").Render(
			r.Div("text-sm text-gray-600").Text(label),
			control,
		)
	}

	checkboxLabel := func(text string, attrs ...func(*r.Node)) *r.Node {
		cb := r.ICheckbox("w-4 h-4 cursor-pointer")
		for _, fn := range attrs {
			fn(cb)
		}
		return r.Label("flex items-center gap-2 text-sm cursor-pointer").Render(cb, r.Span().Text(text))
	}

	basics := r.Div("flex flex-col gap-2").Render(
		ex("Default (checked)", checkboxLabel("I agree", func(n *r.Node) { n.Attr("checked", "true") })),
		ex("Required", checkboxLabel("Accept terms", func(n *r.Node) { n.Attr("required", "true") })),
		ex("Unchecked", checkboxLabel("Unchecked")),
		ex("Disabled", checkboxLabel("Disabled", func(n *r.Node) { n.Attr("disabled", "true") })),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Checkbox"),
		r.Div("text-gray-600").Text("Checkbox states, sizes, and required validation."),
		row("Basics", basics),
	)
}

func RegisterCheckbox(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/checkbox", func(ctx *r.Context) *r.Node { return layout(Checkbox(ctx)) })
	app.Action("nav.checkbox", NavTo("/checkbox", func() *r.Node { return Checkbox(nil) }))
}
