package pages

import (
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

func Hello(ctx *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Hello Actions"),
		r.Div("text-gray-600").Text("Click buttons to trigger different server action responses."),
		r.Div("flex justify-start gap-4 items-center").Render(
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-green-600 text-green-600 hover:bg-green-50").
				Text("with ok").
				OnClick(&r.Action{Name: "hello.ok"}),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-red-600 text-red-600 hover:bg-red-50").
				Text("with error").
				OnClick(&r.Action{Name: "hello.error"}),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-blue-600 text-blue-600 hover:bg-blue-50").
				Text("with delay").
				OnClick(&r.Action{Name: "hello.delay"}),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-yellow-600 text-yellow-600 hover:bg-yellow-50").
				Text("with crash").
				OnClick(&r.Action{Name: "hello.crash"}),
		),
	)
}

func HandleHelloOk(ctx *r.Context) string {
	return r.Notify("success", "Hello")
}

func HandleHelloError(ctx *r.Context) string {
	return r.Notify("error", "Hello error")
}

func HandleHelloDelay(ctx *r.Context) string {
	time.Sleep(2 * time.Second)
	return r.Notify("info", "Information (after 2s delay)")
}

func HandleHelloCrash(ctx *r.Context) string {
	panic("Hello again")
}

func RegisterHello(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/hello", func(ctx *r.Context) *r.Node { return layout(ctx, Hello(ctx)) })
	app.Action("nav.hello", NavTo("/hello", func() *r.Node { return Hello(nil) }))
	app.Action("hello.ok", HandleHelloOk)
	app.Action("hello.error", HandleHelloError)
	app.Action("hello.delay", HandleHelloDelay)
	app.Action("hello.crash", HandleHelloCrash)
}
