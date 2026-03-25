package pages

import r "github.com/michalCapo/g-sui/ui"

func Area(ctx *r.Context) *r.Node {
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

	textareaCls := "w-64 border border-gray-300 rounded px-3 py-2 text-sm"

	labeled := func(labelText string, ta *r.Node) *r.Node {
		return r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-gray-700 font-medium").Text(labelText),
			ta,
		)
	}

	basics := r.Div("flex flex-col gap-2").Render(
		row("Default", labeled("Bio", r.Textarea(textareaCls).Attr("rows", "3").Attr("name", "Bio").Text("Short text"))),
		row("Placeholder", labeled("Your bio", r.Textarea(textareaCls).Attr("rows", "3").Attr("placeholder", "Tell us something"))),
		row("Required", labeled("Required", r.Textarea(textareaCls).Attr("rows", "3").Attr("required", "true"))),
		row("Readonly", labeled("Readonly", r.Textarea(textareaCls).Attr("rows", "3").Attr("readonly", "true").Text("Read-only text"))),
		row("Disabled", labeled("Disabled", r.Textarea(textareaCls+" opacity-50").Attr("rows", "3").Attr("disabled", "true"))),
		row("With preset value", labeled("With value", r.Textarea(textareaCls).Attr("rows", "3").Text("Initial text value"))),
	)

	styling := r.Div("flex flex-col gap-2").Render(
		row("Styled wrapper", labeled("Styled wrapper", r.Textarea("w-64 border border-gray-300 rounded px-3 py-2 text-sm bg-yellow-50").Attr("rows", "3"))),
		row("Custom label", r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-purple-700 font-bold").Text("Custom label"),
			r.Textarea(textareaCls).Attr("rows", "3"),
		)),
		row("Custom input background", labeled("Custom input background", r.Textarea("w-64 border border-gray-300 rounded px-3 py-2 text-sm bg-blue-50").Attr("rows", "3"))),
		row("Size: XL", labeled("XL size", r.Textarea("w-64 border border-gray-300 rounded px-4 py-3 text-lg").Attr("rows", "3"))),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Textarea"),
		r.Div("text-gray-600").Text("Common features supported by textarea."),
		card("Basics & states", basics),
		card("Styling", styling),
	)
}

func RegisterArea(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/area", func(ctx *r.Context) *r.Node { return layout(ctx, Area(ctx)) })
	app.Action("nav.area", NavTo("/area", func() *r.Node { return Area(nil) }))
}
