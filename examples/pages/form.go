package pages

import "github.com/michalCapo/g-sui/ui"

func Submit(ctx *ui.Context) string {
	return "ok"
}

func FormContent(ctx *ui.Context) string {
	target := ui.Target()
	some := ui.FormNew(ctx.Submit(Submit).Replace(target))

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Form missmatch"),
		ui.Div("text-gray-600")("Form input fields and submit button is defined outside html form element. This is useful when you want to reuse the form in multiple places."),

		ui.Div("border rounded-lg p-4 bg-white dark:bg-gray-900 shadow-lg border rounded-lg flex flex-col gap-4")(
			ui.Div("flex flex-col")(
				ui.Div("text-lg font-semibold")("Form creation example"),
				ui.Div("text-gray-600 text-sm mb-4")("Form example with input fields and submit button."),
			),

			ui.Div("flex flex-col", target)(
				"Form submit result will be displayed here.",
			),

			some.Render(),
			some.Text("Title").Required().Render("Title"),
			some.Number("Number").Render("Number"),
			some.Hidden("Some", "number", "value", ui.Attr{Step: "123"}),
			some.Button().Color(ui.Blue).Submit().Render("Submit"),
		),
	)
}
