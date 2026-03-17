package pages

import (
	r "github.com/michalCapo/g-sui/ui"
)

func sharedForm(formID, title, description string) *r.Node {
	inputCls := "w-full border border-gray-300 rounded px-3 py-2 text-sm"

	return r.Div("flex flex-col gap-4").ID(formID).Render(
		r.Div("flex flex-col gap-1").Render(
			r.Div("text-gray-600 text-sm").Text("Title"),
			r.IText(inputCls).ID(formID+"-title").Attr("name", "Title").Attr("value", title).Attr("placeholder", "Title"),
		),
		r.Div("flex flex-col gap-1").Render(
			r.Div("text-gray-600 text-sm").Text("Description"),
			r.Textarea(inputCls).ID(formID+"-desc").Attr("name", "Description").Attr("placeholder", "Description").Text(description),
		),
		r.Div("flex flex-row gap-4 justify-end").Render(
			r.Button("rounded-lg hover:text-red-700 hover:underline text-gray-400 px-3 py-1 cursor-pointer text-sm").
				Text("Reset").
				OnClick(&r.Action{Name: "shared.reset", Data: map[string]any{"formID": formID}}),
			r.Button("rounded-lg px-4 py-2 bg-blue-600 text-white cursor-pointer hover:bg-blue-700 text-sm").
				Text("Submit").
				OnClick(&r.Action{
					Name:    "shared.submit",
					Data:    map[string]any{"formID": formID},
					Collect: []string{formID + "-title", formID + "-desc"},
				}),
		),
	)
}

func Shared(ctx *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Shared"),
		r.Div("text-gray-600").Text("Reused form template in multiple places."),

		r.Div("border rounded-lg p-4 bg-white shadow-lg").Render(
			r.Div("text-lg font-semibold").Text("Form 1"),
			r.Div("text-gray-600 text-sm mb-4").Text("This form is reused."),
			sharedForm("form1", "Hello", "What a nice day"),
		),
		r.Div("border rounded-lg p-4 bg-white shadow-lg").Render(
			r.Div("text-lg font-semibold").Text("Form 2"),
			r.Div("text-gray-600 text-sm mb-4").Text("This form is reused."),
			sharedForm("form2", "Next Title", "Next Description"),
		),
	)
}

func HandleSharedSubmit(ctx *r.Context) string {
	var data map[string]any
	ctx.Body(&data)

	formID, _ := data["formID"].(string)
	title, _ := data[formID+"-title"].(string)
	desc, _ := data[formID+"-desc"].(string)

	if formID == "form1" {
		return r.NewResponse().
			Replace(formID, sharedForm(formID, title, desc)).
			Toast("error", "Data not stored").
			Build()
	}

	return r.NewResponse().
		Replace(formID, sharedForm(formID, title, desc)).
		Toast("success", "Data stored").
		Build()
}

func HandleSharedReset(ctx *r.Context) string {
	var data map[string]any
	ctx.Body(&data)

	formID, _ := data["formID"].(string)
	return sharedForm(formID, "", "").ToJSReplace(formID)
}
