package pages

import r "github.com/michalCapo/g-sui/ui"

func Date(ctx *r.Context) *r.Node {
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
		row("Date", labeled("Birth date", r.IDate(inputCls))),
		row("Time", labeled("Alarm", r.ITime(inputCls))),
		row("DateTime", labeled("Meeting time", r.IDatetime(inputCls))),
		row("Required date", labeled("Required date", r.IDate(inputCls).Attr("required", "true"))),
		row("Readonly time", labeled("Readonly time", r.ITime(inputCls).Attr("readonly", "true").Attr("value", "10:00"))),
		row("Disabled datetime", labeled("Disabled datetime", r.IDatetime(inputCls+" opacity-50").Attr("disabled", "true"))),
	)

	styling := r.Div("flex flex-col gap-2").Render(
		row("Styled wrapper", labeled("Styled wrapper", r.IDate("w-full border border-gray-300 rounded px-3 py-2 text-sm bg-yellow-50"))),
		row("Custom label", r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-purple-700 font-bold").Text("Custom label"),
			r.ITime(inputCls),
		)),
		row("Custom input background", labeled("Custom input background", r.IDatetime("w-full border border-gray-300 rounded px-3 py-2 text-sm bg-blue-50"))),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Date, Time, DateTime"),
		r.Div("text-gray-600").Text("Common attributes across temporal inputs."),
		card("Basics & states", basics),
		card("Styling", styling),
	)
}

func RegisterDate(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/date", func(ctx *r.Context) *r.Node { return layout(ctx, Date(ctx)) })
	app.Action("nav.date", NavTo("/date", func() *r.Node { return Date(nil) }))
}
