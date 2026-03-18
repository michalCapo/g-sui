package pages

import (
	r "github.com/michalCapo/g-sui/ui"
)

func SelectPage(ctx *r.Context) *r.Node {
	row := func(title string, content *r.Node) *r.Node {
		return r.Div("bg-white p-4 rounded-lg shadow border border-gray-200 flex flex-col gap-3").Render(
			r.Div("text-sm font-bold text-gray-700").Text(title),
			content,
		)
	}

	ex := func(label string, controls ...*r.Node) *r.Node {
		right := r.Div("flex items-center gap-3")
		for _, c := range controls {
			right.Render(c)
		}
		return r.Div("flex items-center justify-between gap-4 w-full").Render(
			r.Div("text-sm text-gray-600").Text(label),
			right,
		)
	}

	selectCls := "w-full border border-gray-300 rounded px-3 py-2 text-sm"

	labeled := func(labelText string, sel *r.Node) *r.Node {
		return r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm text-gray-700 font-medium").Text(labelText),
			sel,
		)
	}

	makeSelect := func(name string, opts []string, placeholder string) *r.Node {
		s := r.Select(selectCls).Attr("name", name).ID(name)
		children := make([]*r.Node, 0, len(opts)+1)
		if placeholder != "" {
			children = append(children, r.Option().Attr("value", "").Text(placeholder))
		}
		for _, opt := range opts {
			children = append(children, r.Option().Attr("value", opt).Text(opt))
		}
		return s.Render(children...)
	}

	opts := []string{"One", "Two", "Three"}

	basics := r.Div("flex flex-col gap-2").Render(
		ex("Default", r.Div("w-64").Render(labeled("Country", makeSelect("Country", opts, "Select...")))),
		ex(
			"Placeholder + change handler",
			r.Div("w-64").Render(
				labeled("Choose",
					r.Select(selectCls).Attr("name", "ChooseField").ID("select-choose").
						On("change", &r.Action{
							Name:    "select.change",
							Collect: []string{"select-choose"},
						}).
						Render(
							r.Option().Attr("value", "").Text("Pick one"),
							r.Option().Attr("value", "One").Text("One"),
							r.Option().Attr("value", "Two").Text("Two"),
							r.Option().Attr("value", "Three").Text("Three"),
						),
				),
			),
			r.Div("text-sm text-gray-700").ID("select-display").Text("Selected: (none)"),
		),
	)

	validation := r.Div("flex flex-col gap-2").Render(
		ex("Required", r.Div("w-64").Render(labeled("Required",
			r.Select(selectCls).Attr("name", "Req").Attr("required", "true").Render(
				r.Option().Attr("value", "").Text("Please select"),
				r.Option().Attr("value", "one").Text("One"),
				r.Option().Attr("value", "two").Text("Two"),
			),
		))),
		ex("Disabled", r.Div("w-64").Render(labeled("Disabled",
			r.Select(selectCls+" opacity-50").Attr("disabled", "true").Render(
				r.Option().Text("One"),
				r.Option().Text("Two"),
			),
		))),
		ex("Error state", r.Div("w-64").Render(labeled("Invalid",
			r.Select(selectCls+" border-red-400").Attr("name", "Err").Attr("required", "true").Render(
				r.Option().Attr("value", "").Text("Please select"),
				r.Option().Attr("value", "one").Text("One"),
			),
		))),
	)

	variants := r.Div("flex flex-col gap-2").Render(
		ex("No placeholder", r.Div("w-64").Render(labeled("Choose",
			r.Select(selectCls).Attr("name", "NoPH").Render(
				r.Option().Attr("value", "One").Text("One"),
				r.Option().Attr("value", "Two").Text("Two"),
				r.Option().Attr("value", "Three").Text("Three"),
			),
		))),
	)

	return r.Div("max-w-5xl mx-auto flex flex-col gap-6").Render(
		r.Div("text-3xl font-bold").Text("Select"),
		r.Div("text-gray-600").Text("Select input variations, validation, and sizing."),
		row("Basics", basics),
		row("Validation", validation),
		row("Variants", variants),
	)
}

func HandleSelectChange(ctx *r.Context) string {
	var data map[string]any
	ctx.Body(&data)

	val, _ := data["ChooseField"].(string)
	if val == "" {
		val = "(none)"
	}

	return r.SetText("select-display", "Selected: "+val)
}

func RegisterSelect(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/select", func(ctx *r.Context) *r.Node { return layout(SelectPage(ctx)) })
	app.Action("nav.select", NavTo("/select", func() *r.Node { return SelectPage(nil) }))
	app.Action("select.change", HandleSelectChange)
}
