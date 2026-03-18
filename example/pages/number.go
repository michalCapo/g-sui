package pages

import r "github.com/michalCapo/g-sui/ui"

func Number(ctx *r.Context) *r.Node {
	card := func(title string, body *r.Node) *r.Node {
		return r.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3").Render(
			r.Div("text-sm font-bold text-gray-700").Text(title),
			body,
		)
	}

	row := func(label string, control *r.Node) *r.Node {
		return r.Div("flex items-center justify-between gap-4").Render(
			r.Div("text-sm text-gray-600").Text(label),
			r.Div("w-64").Render(control),
		)
	}

	inputCls := "w-full border border-gray-300 rounded px-3 py-2 text-sm"

	labeled := func(labelText string, input *r.Node) *r.Node {
		return r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-gray-700 font-medium").Text(labelText),
			input,
		)
	}

	basics := r.Div("flex flex-col gap-2").Render(
		row("Integer with range/step", labeled("Age", r.INumber(inputCls).Attr("name", "Age").Attr("value", "30").Attr("min", "0").Attr("max", "120").Attr("step", "1"))),
		row("Float formatted", labeled("Price", r.INumber(inputCls).Attr("name", "Price").Attr("value", "19.90").Attr("step", "0.01"))),
		row("Required", labeled("Required", r.INumber(inputCls).Attr("required", "true"))),
		row("Readonly", labeled("Readonly", r.INumber(inputCls).Attr("readonly", "true").Attr("value", "42"))),
		row("Disabled", labeled("Disabled", r.INumber(inputCls+" opacity-50").Attr("disabled", "true"))),
		row("Placeholder", labeled("Number", r.INumber(inputCls).Attr("placeholder", "0..100"))),
	)

	styling := r.Div("flex flex-col gap-2").Render(
		row("Styled wrapper", labeled("Styled wrapper", r.INumber("w-full border border-gray-300 rounded px-3 py-2 text-sm bg-yellow-50"))),
		row("Custom label", r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-purple-700 font-bold").Text("Custom label"),
			r.INumber(inputCls),
		)),
		row("Custom input background", labeled("Custom input background", r.INumber("w-full border border-gray-300 rounded px-3 py-2 text-sm bg-blue-50"))),
		row("Size: LG", labeled("Large size", r.INumber("w-full border border-gray-300 rounded px-4 py-3 text-lg"))),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Number input"),
		r.Div("text-gray-600").Text("Ranges, formatting, and common attributes."),
		card("Basics & states", basics),
		card("Styling", styling),
	)
}

func RegisterNumber(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/number", func(ctx *r.Context) *r.Node { return layout(Number(ctx)) })
	app.Action("nav.number", NavTo("/number", func() *r.Node { return Number(nil) }))
}
