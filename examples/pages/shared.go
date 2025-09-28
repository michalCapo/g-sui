package pages

import "github.com/michalCapo/g-sui/ui"

type TemplateForm struct {
	target      ui.Attr
	Title       string
	Description string
	onSubmit    func(*ui.Context) string
}

func NewTemplate(title string, description string) *TemplateForm {
	target := ui.Target()

	return &TemplateForm{
		target:      target,
		Title:       title,
		Description: description,
	}
}

func (f *TemplateForm) OnCancel(ctx *ui.Context) string {
	f.Title = ""
	f.Description = ""

	return f.Render(ctx)
}

func (f *TemplateForm) Render(ctx *ui.Context) string {

	return ui.Form("flex flex-col gap-4", f.target, ctx.Submit(f.onSubmit).Replace(f.target))(
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
				Click(ctx.Call(f.OnCancel).Replace(f.target)).
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
	form1 := NewTemplate(
		"Hello",
		"What a nice day",
	)
	form2 := NewTemplate(
		"Next Title",
		"Next Description",
	)

	form1.onSubmit = func(ctx *ui.Context) string {
		ctx.Error("Data not stored")

		return form1.Render(ctx)
	}

	form2.onSubmit = func(ctx *ui.Context) string {
		ctx.Success("Data stored but do not shared")

		return form2.Render(ctx)
	}

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Shared"),
		ui.Div("text-gray-600")("Tries to mimmic real application: reused form in multiplate places"),

		ui.Div("border rounded-lg p-4 bg-white dark:bg-gray-900 shadow-lg border rounded-lg")(
			ui.Div("text-lg font-semibold")("Form 1"),
			ui.Div("text-gray-600 text-sm mb-4")("This form is reused."),
			form1.Render(ctx),
		),
		ui.Div("border rounded-lg p-4 bg-white dark:bg-gray-900 shadow-lg border rounded-lg")(
			ui.Div("text-lg font-semibold")("Form 2"),
			ui.Div("text-gray-600 text-sm mb-4")("This form is reused."),
			form2.Render(ctx),
		),
	)
}
