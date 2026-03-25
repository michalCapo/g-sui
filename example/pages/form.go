package pages

import (
	"fmt"

	r "github.com/michalCapo/g-sui/ui"
)

type FormData struct {
	Action     string `json:"Action"`
	Title      string `json:"Title"`
	GenderNext string `json:"GenderNext"`
	Gender     string `json:"Gender"`
	Country    string `json:"Country"`
	Some       string `json:"Some"`
	Number     string `json:"Number"`
	Agree      bool   `json:"Agree"`
}

// formBuilder returns a reusable form definition. The form ID scopes all
// element IDs, radio group names, and the HTML form attribute — so multiple
// instances with different IDs can coexist on the same page without collision.
func formBuilder(formID string, data FormData) *r.FormBuilder {
	return r.NewForm(formID).
		Action("form.submit").

		// Title (text, required)
		Text("Title", "Title").Required().Placeholder("Enter title").Value(data.Title).Render().

		// Gender inline radios (required)
		Radio("Gender (radio)", "GenderNext").Required().
		Opts("male:Male", "female:Female").Value(data.GenderNext).Render().

		// Agree checkbox (not required)
		Checkbox("I agree", "Agree").IsChecked(data.Agree).Render().

		// Gender button-style radios (required)
		RadioBtn("Gender (buttons)", "Gender").Required().
		Opts("male:Male", "female:Female", "other:Other").Value(data.Gender).Render().

		// Country select (required)
		SelectField("Country", "Country").Required().
		Opts(":Select country...", "1:USA", "2:Slovakia", "3:Germany", "4:Japan").
		Value(data.Country).Render().

		// Hidden field
		Hidden("Some").Value("123").Render().

		// Number card-style radios (required)
		RadioCard("Number", "Number").Required().
		Opts("1:1", "2:2", "3:3").Value(data.Number).Render().

		// Submit buttons
		Submit("save", "Save", "px-4 py-2 rounded bg-blue-600 text-white cursor-pointer hover:bg-blue-700 text-sm").
		Submit("preview", "Preview", "px-4 py-2 rounded bg-purple-600 text-white cursor-pointer hover:bg-purple-700 text-sm").
		Submit("", "Submit", "px-4 py-2 rounded border border-gray-300 text-gray-700 cursor-pointer hover:bg-gray-50 text-sm")
}

func FormPage(ctx *r.Context) *r.Node {
	return formPageContent("", FormData{}, "", FormData{})
}

func formPageContent(result1 string, data1 FormData, result2 string, data2 FormData) *r.Node {
	if result1 == "" {
		result1 = "Form 1 result will be displayed here."
	}
	if result2 == "" {
		result2 = "Form 2 result will be displayed here."
	}

	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").ID("form-page").Render(
		r.Div("text-2xl font-bold").Text("Form association"),
		r.Div("text-gray-600").Text("Two independent forms on the same page. Each form uses HTML form association — inputs are scoped by form ID so radio groups, IDs, and validation are fully isolated."),

		// Form 1
		r.Div("rounded-lg p-4 bg-white shadow-lg flex flex-col gap-4").Render(
			r.Div("flex flex-col").Render(
				r.Div("text-lg font-semibold").Text("Form 1"),
				r.Div("text-gray-600 text-sm mb-4").Text("First form instance with its own state."),
			),
			r.Div("flex flex-col").ID("form1-result").Render(
				r.Div("text-sm text-gray-600").Text(result1),
			),
			formBuilder("form1", data1).Build(),
		),

		// Form 2 — same field definitions, different form ID
		r.Div("rounded-lg p-4 bg-white shadow-lg flex flex-col gap-4").Render(
			r.Div("flex flex-col").Render(
				r.Div("text-lg font-semibold").Text("Form 2"),
				r.Div("text-gray-600 text-sm mb-4").Text("Second form instance — fully isolated from Form 1."),
			),
			r.Div("flex flex-col").ID("form2-result").Render(
				r.Div("text-sm text-gray-600").Text(result2),
			),
			formBuilder("form2", data2).Build(),
		),
	)
}

func HandleFormSubmit(ctx *r.Context) string {
	var data FormData
	ctx.Body(&data)

	result := fmt.Sprintf(
		"Action=%s  Title=%s  GenderNext=%s  Gender=%s  Country=%s  Some=%s  Number=%s  Agree=%v",
		data.Action, data.Title, data.GenderNext, data.Gender, data.Country, data.Some, data.Number, data.Agree,
	)

	var toast string
	switch data.Action {
	case "save":
		toast = "Form saved successfully"
	case "preview":
		toast = "Form preview displayed"
	default:
		toast = "Form submitted successfully"
	}

	return r.NewResponse().
		Replace("form-page", formPageContent(result, data, "", FormData{})).
		Toast("success", toast).
		Build()
}

func RegisterForm(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/form", func(ctx *r.Context) *r.Node { return layout(ctx, FormPage(ctx)) })
	app.Action("nav.form", NavTo("/form", func() *r.Node { return FormPage(nil) }))
	app.Action("form.submit", HandleFormSubmit)
}
