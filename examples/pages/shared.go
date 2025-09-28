package pages

import "github.com/michalCapo/g-sui/ui"

type someForm struct {
	Title       string
	Description string
	onCancel    string
	onSubmit    ui.Attr
}

func (f *someForm) Render(ctx *ui.Context, target ui.Attr) string {

	return ui.Form("flex flex-col gap-4", target, f.onSubmit)(
		ui.Div("")(
			ui.Div("text-gray-600 text-sm")("Title"),
			ui.IText("Title", f).
				Class("w-full").
				Placeholder("Title").
				Render(""),
		),
		ui.Div("")(
			ui.Div("text-gray-600 text-sm")("Description"),
			ui.IArea("Description", f).
				Class("w-full").
				Placeholder("Description").
				Render(""),
		),

		// Buttons
		ui.Div("flex flex-row gap-4 justify-end")(
			ui.Button().
				// Class("rounded text-sm").
				Class("rounded-lg hover:text-red-700 hover:underline text-gray-400").
				// Color(ui.Gray).
				Click(f.onCancel).
				Render("Reset"),

			ui.Button().
				Submit().
				// Class("rounded text-sm").
				Class("rounded-lg").
				Color(ui.Blue).
				Render("Submit"),
		),
	)
}

func Shared(ctx *ui.Context) string {
	target := ui.Target()
	target2 := ui.Target()

	form1 := &someForm{
		Title:       "Hello",
		Description: "What a nice day",
	}
	form2 := &someForm{
		Title:       "Next Title",
		Description: "Next Description",
	}

	onCancel := func(ctx *ui.Context) string {
		form1.Title = ""
		form1.Description = ""

		return form1.Render(ctx, target)
	}

	onSubmit := func(ctx *ui.Context) string {
		ctx.Success("Data stored")

		return form1.Render(ctx, target)
	}

	onCancel2 := func(ctx *ui.Context) string {
		form2.Title = ""
		form2.Description = ""

		return form2.Render(ctx, target2)
	}

	onSubmit2 := func(ctx *ui.Context) string {
		ctx.Success("Data stored")

		return form2.Render(ctx, target2)
	}

	form1.onCancel = ctx.Call(onCancel).Replace(target)
	form1.onSubmit = ctx.Submit(onSubmit).Replace(target)

	form2.onCancel = ctx.Call(onCancel2).Replace(target2)
	form2.onSubmit = ctx.Submit(onSubmit2).Replace(target2)

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Complex"),
		ui.Div("text-gray-600")("Tries to mimmic real application: reused form in multiplate places"),

		ui.Div("border rounded-lg p-4 bg-white dark:bg-gray-900 shadow-lg border rounded-lg")(
			ui.Div("text-lg font-semibold")("Form 1"),
			ui.Div("text-gray-600 text-sm mb-4")("This form is reused."),
			form1.Render(ctx, target),
		),
		ui.Div("border rounded-lg p-4 bg-white dark:bg-gray-900 shadow-lg border rounded-lg")(
			ui.Div("text-lg font-semibold")("Form 2"),
			ui.Div("text-gray-600 text-sm mb-4")("This form is reused."),
			form2.Render(ctx, target2),
		),
	)
}
