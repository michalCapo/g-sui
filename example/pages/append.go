package pages

import (
	"fmt"
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

func Append(ctx *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Append / Prepend Demo"),
		r.Div("text-gray-600").Text("Click buttons to insert items at the beginning or end of the list."),
		r.Div("flex gap-2").Render(
			r.Button("px-4 py-2 rounded cursor-pointer bg-blue-600 text-white hover:bg-blue-700 text-sm").
				Text("Add at end").
				OnClick(&r.Action{Name: "append.end"}),
			r.Button("px-4 py-2 rounded cursor-pointer bg-green-600 text-white hover:bg-green-700 text-sm").
				Text("Add at start").
				OnClick(&r.Action{Name: "append.start"}),
		),
		r.Div("space-y-2").ID("append-list").Render(
			r.Div("p-2 rounded border bg-white").Render(
				r.Span("text-sm text-gray-600").Text("Initial item"),
			),
		),
	)
}

func HandleAppendEnd(ctx *r.Context) string {
	now := time.Now().Format("15:04:05")
	item := r.Div("p-2 rounded border bg-white").Render(
		r.Span("text-sm text-gray-600").Text(fmt.Sprintf("Appended at %s", now)),
	)
	return item.ToJSAppend("append-list")
}

func HandleAppendStart(ctx *r.Context) string {
	now := time.Now().Format("15:04:05")
	item := r.Div("p-2 rounded border bg-white").Render(
		r.Span("text-sm text-gray-600").Text(fmt.Sprintf("Prepended at %s", now)),
	)
	return item.ToJSPrepend("append-list")
}
