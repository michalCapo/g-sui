package pages

import (
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

// ActionsPage shows local UI actions. None of them needs JavaScript or a
// server round trip.
func ActionsPage(_ *r.Context) *r.Node {
	btn := "px-3 py-1.5 rounded border border-gray-300 dark:border-gray-600 text-sm cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
	input := "border border-gray-300 dark:border-gray-600 rounded px-3 py-1.5 text-sm bg-white dark:bg-gray-900"
	card := func(title string, nodes ...*r.Node) *r.Node {
		return r.Div("flex flex-col gap-3 bg-white dark:bg-gray-900 rounded-lg shadow p-4").Render(
			r.Div("font-semibold").Text(title),
			r.Div("flex flex-wrap items-center gap-2").Render(nodes...),
		)
	}

	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Title("Actions").Render(
		r.Div("text-3xl font-bold").Text("Local actions"),
		r.Div("text-gray-600 dark:text-gray-400").Text("Browser-side UI built from Go actions. Press Ctrl+K (Cmd+K) to focus the input."),

		card("Menu",
			r.Div("relative").Render(
				r.Button(btn).Text("Menu").OnClick(r.Toggle("actions-menu")),
				r.Div("hidden absolute z-10 mt-1 w-40 bg-white dark:bg-gray-800 rounded shadow-lg border border-gray-200 dark:border-gray-700").
					ID("actions-menu").
					OnOutsideClick(r.Hide("actions-menu")).
					OnKey("Escape", r.Hide("actions-menu")).
					Render(
						r.Button("block w-full text-left px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer").
							Text("Say hello").OnClick(r.Seq(r.Toast("Hello"), r.Hide("actions-menu"))),
						r.Button("block w-full text-left px-3 py-2 text-sm text-red-600 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer").
							Text("Delete").OnClick(r.Confirm("Delete it?", r.Notify("error", "Deleted"))),
					),
			),
		),

		card("Inputs",
			r.IText(input).ID("actions-input").Attr("placeholder", "Type, then Enter").
				Shortcut("mod+k", r.Focus("actions-input")).
				OnKey("Enter", r.CopyFrom("actions-input")).
				OnKey("Escape", r.SetValue("actions-input", "")),
			r.Button(btn).Text("Copy").OnClick(r.CopyFrom("actions-input")),
			r.Button(btn).Text("Clear").OnClick(r.Seq(r.SetValue("actions-input", ""), r.Focus("actions-input"))),
			r.IPassword(input).ID("actions-password").Attr("value", "secret"),
			r.Button(btn).Text("Show password").OnClick(r.TogglePassword("actions-password")),
		),

		card("Classes and text",
			r.Div("px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-800 text-sm").ID("actions-box").Text("Box"),
			r.Button(btn).Text("Highlight").OnClick(r.ToggleClass("actions-box", "ring-2 ring-blue-500")),
			r.Button(btn).Text("Set text").OnClick(r.SetText("actions-box", "Changed")),
			r.Button(btn).Text("Later").OnClick(r.Delay(time.Second, r.SetText("actions-box", "One second later"))),
		),

		card("Dialog",
			r.Button(btn).Text("Open dialog").OnClick(r.OpenDialog("actions-dialog")),
			r.El("dialog", "rounded-lg p-6 shadow-xl backdrop:bg-black/50 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100").
				ID("actions-dialog").Render(
				r.Div("flex flex-col gap-4").Render(
					r.Div().Text("A native dialog. Escape closes it too."),
					r.Button(btn).Text("Close").OnClick(r.CloseDialog("actions-dialog")),
				),
			),
		),

		card("Page",
			r.Button(btn).Text("Toggle theme").OnClick(r.ToggleTheme()),
			r.Button(btn).Text("Set title").OnClick(r.SetTitle("Actions changed")),
			r.Button(btn).Text("Go to counter").OnClick(r.Navigate("/counter")),
			r.Button(btn).Text("Scroll to top").OnClick(r.ScrollTop()),
		),

		card("Widget",
			r.Widget("ticker", map[string]any{"label": "Mounted for"}, "text-sm"),
		),
	)
}

func RegisterActions(app *r.App) {
	// A widget wraps third-party JavaScript with mount and cleanup.
	app.Widget("ticker", `function(el, props) {
		var n = 0;
		el.textContent = props.label + " 0s";
		var t = setInterval(function() { el.textContent = props.label + " " + (++n) + "s"; }, 1000);
		return function() { clearInterval(t); };
	}`)
	app.Page("/actions", ActionsPage)
}
