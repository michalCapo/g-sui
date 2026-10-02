package pages

import (
	"fmt"
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

const appendListID = "append-list"

var appendEnd, appendStart r.ActionRef[struct{}]

func Append(_ *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Append / Prepend Demo"),
		r.Div("text-gray-600").Text("Click buttons to insert items at the beginning or end of the list."),
		r.Div("flex gap-2").Render(
			r.Button("px-4 py-2 rounded cursor-pointer bg-blue-600 text-white hover:bg-blue-700 text-sm").
				Text("Add at end").
				OnClick(appendEnd.Call(struct{}{})),
			r.Button("px-4 py-2 rounded cursor-pointer bg-green-600 text-white hover:bg-green-700 text-sm").
				Text("Add at start").
				OnClick(appendStart.Call(struct{}{})),
		),
		r.Div("space-y-2").ID(appendListID).Render(
			r.Div("p-2 rounded border bg-white").Render(
				r.Span("text-sm text-gray-600").Text("Initial item"),
			),
		),
	)
}

func HandleAppendEnd(_ *r.Context, _ struct{}) (r.Result, error) {
	now := time.Now().Format("15:04:05")
	item := r.Div("p-2 rounded border bg-white").Render(
		r.Span("text-sm text-gray-600").Text(fmt.Sprintf("Appended at %s", now)),
	)
	return r.Result{}.Append(appendListID, item), nil
}

func HandleAppendStart(_ *r.Context, _ struct{}) (r.Result, error) {
	now := time.Now().Format("15:04:05")
	item := r.Div("p-2 rounded border bg-white").Render(
		r.Span("text-sm text-gray-600").Text(fmt.Sprintf("Prepended at %s", now)),
	)
	return r.Result{}.Prepend(appendListID, item), nil
}

func RegisterAppend(app *r.App) {
	appendEnd = r.RegisterAction(app, "append.end", HandleAppendEnd)
	appendStart = r.RegisterAction(app, "append.start", HandleAppendStart)
	app.Page("/append", Append)
}
