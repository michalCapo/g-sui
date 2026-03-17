package pages

import r "github.com/michalCapo/g-sui/ui"

func Radio(ctx *r.Context) *r.Node {
	card := func(title string, body *r.Node) *r.Node {
		return r.Div("bg-white p-4 rounded-lg shadow flex flex-col gap-3").Render(
			r.Div("text-sm font-bold text-gray-700").Text(title),
			body,
		)
	}

	radioLabel := func(name, value, text string, checked bool) *r.Node {
		cb := r.IRadio("w-4 h-4 cursor-pointer").Attr("name", name).Attr("value", value)
		if checked {
			cb.Attr("checked", "true")
		}
		return r.Label("flex items-center gap-2 text-sm cursor-pointer").Render(cb, r.Span().Text(text))
	}

	radioLabelSize := func(name, value, text, size string) *r.Node {
		inputCls := "cursor-pointer"
		switch size {
		case "sm":
			inputCls += " w-3 h-3"
		case "xs":
			inputCls += " w-2.5 h-2.5"
		default:
			inputCls += " w-4 h-4"
		}
		return r.Label("flex items-center gap-2 text-sm cursor-pointer").Render(
			r.IRadio(inputCls).Attr("name", name).Attr("value", value),
			r.Span().Text(text),
		)
	}

	// Standalone radios with selection
	single := r.Div("flex flex-col gap-2").Render(
		radioLabel("Gender", "male", "Male", true),
		radioLabel("Gender", "female", "Female", false),
		radioLabel("Gender", "other", "Other", false),
	)

	// Button-group style radios
	group := r.Div("flex gap-2").Render(
		r.Label("flex items-center gap-1 px-3 py-2 border rounded cursor-pointer hover:bg-gray-50").Render(
			r.IRadio("w-4 h-4").Attr("name", "Group").Attr("value", "male"),
			r.Span("text-sm").Text("Male"),
		),
		r.Label("flex items-center gap-1 px-3 py-2 border rounded cursor-pointer hover:bg-gray-50").Render(
			r.IRadio("w-4 h-4").Attr("name", "Group").Attr("value", "female"),
			r.Span("text-sm").Text("Female"),
		),
		r.Label("flex items-center gap-1 px-3 py-2 border rounded cursor-pointer hover:bg-gray-50").Render(
			r.IRadio("w-4 h-4").Attr("name", "Group").Attr("value", "other"),
			r.Span("text-sm").Text("Other"),
		),
	)

	// Custom card-style radios (RadioDiv equivalent)
	customRadio := func(name, value string, content *r.Node) *r.Node {
		return r.Label("cursor-pointer block").Render(
			r.IRadio("peer hidden").Attr("name", name).Attr("value", value),
			r.Div("border-2 border-gray-200 rounded-xl transition-all peer-checked:border-blue-500 peer-checked:shadow-md hover:border-gray-300").Render(
				content,
			),
		)
	}

	custom := r.Div("grid grid-cols-3 gap-4").Render(
		customRadio("Custom", "car1",
			r.Div("w-full h-48 rounded-xl relative overflow-hidden").Render(
				r.Div("absolute inset-0 bg-gradient-to-br from-gray-100 to-gray-200 opacity-60"),
				r.Div("absolute top-4 left-4 z-10").Render(
					r.Div("text-2xl font-bold text-gray-800").Text("11-AA"),
					r.Div("text-lg text-gray-700").Text("Skoda"),
				),
			),
		),
		customRadio("Custom", "car2",
			r.Div("w-full h-48 rounded-xl relative bg-white").Render(
				r.Div("absolute top-4 left-4 z-10").Render(
					r.Div("text-2xl font-bold text-gray-800").Text("22aa"),
				),
			),
		),
		customRadio("Custom", "car3",
			r.Div("w-full h-48 rounded-xl relative overflow-hidden").Render(
				r.Div("absolute inset-0 bg-gradient-to-br from-amber-100 to-amber-200 opacity-60"),
				r.Div("absolute top-4 left-4 z-10").Render(
					r.Div("text-2xl font-bold text-gray-800").Text("ABC-123"),
					r.Div("text-lg text-gray-700").Text("Volvo"),
				),
			),
		),
	)

	// Size variants
	sizes := r.Div("flex flex-col gap-2").Render(
		radioLabelSize("SizeA", "a", "Default", "md"),
		radioLabelSize("SizeB", "b", "Small (SM)", "sm"),
		radioLabelSize("SizeC", "c", "Extra Small (XS)", "xs"),
	)

	// Validation - required
	validation := r.Div("flex flex-col gap-2").Render(
		r.Div("text-sm text-gray-700").Text("Required group (no selection)"),
		r.Div("flex gap-2").Render(
			r.Label("flex items-center gap-1 px-3 py-2 border rounded cursor-pointer hover:bg-gray-50").Render(
				r.IRadio("w-4 h-4").Attr("name", "ReqGroup").Attr("value", "a").Attr("required", "true"),
				r.Span("text-sm").Text("Option A"),
			),
			r.Label("flex items-center gap-1 px-3 py-2 border rounded cursor-pointer hover:bg-gray-50").Render(
				r.IRadio("w-4 h-4").Attr("name", "ReqGroup").Attr("value", "b").Attr("required", "true"),
				r.Span("text-sm").Text("Option B"),
			),
		),
		r.Div("text-sm text-gray-700 mt-2").Text("Required standalone radios"),
		r.Div("flex flex-col gap-1").Render(
			r.Label("flex items-center gap-2 text-sm cursor-pointer").Render(
				r.IRadio("w-4 h-4").Attr("name", "ReqSingle").Attr("value", "a").Attr("required", "true"),
				r.Span().Text("Option A"),
			),
			r.Label("flex items-center gap-2 text-sm cursor-pointer").Render(
				r.IRadio("w-4 h-4").Attr("name", "ReqSingle").Attr("value", "b").Attr("required", "true"),
				r.Span().Text("Option B"),
			),
			r.Label("flex items-center gap-2 text-sm cursor-pointer").Render(
				r.IRadio("w-4 h-4").Attr("name", "ReqSingle").Attr("value", "c").Attr("required", "true"),
				r.Span().Text("Option C"),
			),
		),
	)

	// Disabled
	disabled := r.Div("flex flex-col gap-2").Render(
		r.Label("flex items-center gap-2 text-sm opacity-50").Render(
			r.IRadio("w-4 h-4").Attr("name", "Dis").Attr("value", "a").Attr("disabled", "true"),
			r.Span().Text("Disabled A"),
		),
		r.Label("flex items-center gap-2 text-sm opacity-50").Render(
			r.IRadio("w-4 h-4").Attr("name", "Dis").Attr("value", "b").Attr("disabled", "true"),
			r.Span().Text("Disabled B"),
		),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Radio"),
		r.Div("text-gray-600").Text("Single radio inputs, grouped radio buttons, and custom card-style radios."),
		card("Standalone radios (with selection)", single),
		card("Radio buttons group", group),
		card("Custom card radios (RadioDiv)", custom),
		card("Sizes", sizes),
		card("Validation (required)", validation),
		card("Disabled", disabled),
	)
}
