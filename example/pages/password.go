package pages

import r "github.com/michalCapo/g-sui/ui"

func Password(ctx *r.Context) *r.Node {
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

	labeled := func(labelText string, input *r.Node) *r.Node {
		return r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-gray-700 font-medium").Text(labelText),
			input,
		)
	}

	basics := r.Div("flex flex-col gap-2").Render(
		row("Default", labeled("Password", r.IPassword(inputCls))),
		row("With placeholder", labeled("Password", r.IPassword(inputCls).Attr("placeholder", "••••••••"))),
		row("Required", labeled("Password (required)", r.IPassword(inputCls).Attr("required", "true"))),
		row("Readonly", labeled("Readonly password", r.IPassword(inputCls).Attr("readonly", "true").Attr("value", "secret"))),
		row("Disabled", labeled("Password (disabled)", r.IPassword(inputCls+" opacity-50").Attr("disabled", "true"))),
		row("Preset value", labeled("Preset value", r.IPassword(inputCls).Attr("value", "topsecret"))),
		row("Visible (type=text)", labeled("As text", r.IText(inputCls).Attr("value", "visible value"))),
	)

	styling := r.Div("flex flex-col gap-2").Render(
		row("Styled wrapper", labeled("Styled wrapper", r.IPassword("w-64 border border-gray-300 rounded px-3 py-2 text-sm bg-yellow-50"))),
		row("Custom label", r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-purple-700 font-bold").Text("Custom label"),
			r.IPassword(inputCls),
		)),
		row("Custom input background", labeled("Custom input background", r.IPassword("w-64 border border-gray-300 rounded px-3 py-2 text-sm bg-blue-50"))),
		row("Size: XS", labeled("XS", r.IPassword("w-64 border border-gray-300 rounded px-2 py-0.5 text-xs"))),
		row("Size: XL", labeled("XL", r.IPassword("w-64 border border-gray-300 rounded px-4 py-3 text-lg"))),
	)

	behavior := r.Div("flex flex-col gap-2").Render(
		row("Autocomplete (new-password)", labeled("New password", r.IPassword(inputCls).Attr("autocomplete", "new-password"))),
		row("Pattern (min 8 chars)", labeled("Min length pattern", r.IPassword(inputCls).Attr("pattern", ".{8,}").Attr("placeholder", "at least 8 characters"))),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Password"),
		r.Div("text-gray-600").Text("Common features and states for password inputs."),
		card("Basics & states", basics),
		card("Styling", styling),
		card("Behavior & attributes", behavior),
	)
}
