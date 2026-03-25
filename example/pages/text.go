package pages

import r "github.com/michalCapo/g-sui/ui"

func Text(ctx *r.Context) *r.Node {
	card := func(title string, body *r.Node) *r.Node {
		return r.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3").Render(
			r.Div("text-sm font-bold text-gray-700").Text(title),
			body,
		)
	}

	row := func(label string, control *r.Node) *r.Node {
		return r.Div("flex items-center justify-between gap-4").Render(
			r.Div("text-sm text-gray-600").Text(label),
			control,
		)
	}

	inputCls := "w-64 border border-gray-300 rounded px-3 py-2 text-sm"

	labeledInput := func(labelText string, input *r.Node) *r.Node {
		return r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-gray-700 font-medium").Text(labelText),
			input,
		)
	}

	basics := r.Div("flex flex-col gap-2").Render(
		row("Default", labeledInput("Name", r.IText(inputCls).Attr("name", "Name").Attr("value", "John Doe"))),
		row("With placeholder", labeledInput("Your name", r.IText(inputCls).Attr("placeholder", "Type your name"))),
		row("Required field", labeledInput("Required field", r.IText(inputCls).Attr("required", "true"))),
		row("Readonly", labeledInput("Readonly field", r.IText(inputCls).Attr("readonly", "true").Attr("value", "Read-only value"))),
		row("Disabled", labeledInput("Disabled", r.IText(inputCls+" opacity-50").Attr("disabled", "true").Attr("placeholder", "Cannot type"))),
		row("With preset value", labeledInput("Preset", r.IText(inputCls).Attr("value", "Preset text"))),
	)

	styling := r.Div("flex flex-col gap-2").Render(
		row("Styled wrapper", labeledInput("Styled wrapper", r.IText("w-64 border border-gray-300 rounded px-3 py-2 text-sm bg-yellow-50"))),
		row("Custom label", r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-purple-700 font-bold").Text("Custom label"),
			r.IText(inputCls),
		)),
		row("Custom input background", labeledInput("Custom input background", r.IText("w-64 border border-gray-300 rounded px-3 py-2 text-sm bg-blue-50"))),
		row("Size: XS", labeledInput("XS", r.IText("w-64 border border-gray-300 rounded px-2 py-0.5 text-xs"))),
		row("Size: MD (default)", labeledInput("MD", r.IText(inputCls))),
		row("Size: XL", labeledInput("XL", r.IText("w-64 border border-gray-300 rounded px-4 py-3 text-lg"))),
	)

	behavior := r.Div("flex flex-col gap-2").Render(
		row("Autocomplete", labeledInput("Name (autocomplete)", r.IText(inputCls).Attr("autocomplete", "name"))),
		row("Pattern (email-like)", labeledInput("Email", r.IEmail(inputCls).Attr("pattern", "[^@]+@[^@]+\\.[^@]+").Attr("placeholder", "user@example.com"))),
		row("Type switch (password)", labeledInput("Password-like", r.IPassword(inputCls))),
		row("Change handler", labeledInput("On change, log value", r.IText(inputCls).OnClick(r.JS("console.log('clicked input')")))),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Text input"),
		r.Div("text-gray-600").Text("Common features supported by text-like inputs."),
		card("Basics & states", basics),
		card("Styling", styling),
		card("Behavior & attributes", behavior),
	)
}

func RegisterText(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/text", func(ctx *r.Context) *r.Node { return layout(ctx, Text(ctx)) })
	app.Action("nav.text", NavTo("/text", func() *r.Node { return Text(nil) }))
}
