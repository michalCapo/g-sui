package pages

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/michalCapo/g-sui/ui"
)

type formData struct {
	Title string `validate:"required"`
	Some  int    `validate:"required"`
}

func Submit(ctx *ui.Context) string {
	form := formData{}
	err := ctx.Body(&form)
	if err != nil {
		return render(ctx, &form, &err)
	}

	v := validator.New()
	if err := v.Struct(form); err != nil {
		return render(ctx, &form, &err)
	}

	ctx.Success("Form submitted successfully")

	return render(ctx, &form, nil)
}

func render(ctx *ui.Context, data *formData, err *error) string {
	target := ui.Target()
	form := ui.FormNew(ctx.Submit(Submit).Replace(target))

	result := "Form submit result will be displayed here."

	if err == nil {
		result = fmt.Sprintf("Form data: %+v", data)
	}

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4", target)(
		ui.Div("text-2xl font-bold")("Form missmatch"),
		ui.Div("text-gray-600")("Form input fields and submit button is defined outside html form element. This is useful when you want to reuse the form in multiple places."),

		ui.Div("border rounded-lg p-4 bg-white dark:bg-gray-900 shadow-lg border rounded-lg flex flex-col gap-4")(
			ui.Div("flex flex-col")(
				ui.Div("text-lg font-semibold")("Form creation example"),
				ui.Div("text-gray-600 text-sm mb-4")("Form example with input fields and submit button."),
			),

			ui.ErrorForm(err, nil),

			ui.Div("flex flex-col")(
				result,
			),

			form.Render(),
			form.Text("Title", data).Required().Render("Title"),
			form.Hidden("Some", "int", 123),
			form.Button().Color(ui.Blue).Submit().Render("Submit"),
		),
	)
}

func FormContent(ctx *ui.Context) string {
	return render(ctx, &formData{}, nil)
}
