// Package pages provides example pages demonstrating form usage with FormInstance.
// This file demonstrates creating forms where input fields and submit buttons
// are defined outside the HTML form element for better reusability.
package pages

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/michalCapo/g-sui/ui"
)

type formData struct {
	Title   string `validate:"required"`
	Some    int    `validate:"required"`
	Gender  string `validate:"required"`
	Number  int    `validate:"required"`
	Country uint   `validate:"required"`
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

	genders := []ui.AOption{{ID: "male", Value: "Male"}, {ID: "female", Value: "Female"}, {ID: "other", Value: "Other"}}
	numbers := []ui.AOption{{ID: "1", Value: "1"}, {ID: "2", Value: "2"}, {ID: "3", Value: "3"}}
	countries := []ui.AOption{{ID: "1", Value: "USA"}, {ID: "2", Value: "Slovakia"}, {ID: "3", Value: "Germany"}, {ID: "4", Value: "Japan"}}

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4", target)(
		ui.Div("text-2xl font-bold")("Form association"),
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
			form.RadioButtons("Gender", data).Options(genders).Render("Male"),
			form.RadioButtons("Number", data).Options(numbers, "int").Render("Number"),
			form.Select("Country", data).Options(countries, "uint").Render("Country"),
			form.Hidden("Some", "int", 123),
			form.Button().Color(ui.Blue).Submit().Render("Submit"),
		),
	)
}

func FormContent(ctx *ui.Context) string {
	return render(ctx, &formData{Gender: "male", Number: 2}, nil)
}
