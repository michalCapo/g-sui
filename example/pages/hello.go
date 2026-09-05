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

func HandleHelloOk(ctx *r.Context, _ struct{}) (r.Result, error) {
	return r.Result{}.Notify("success", "Hello"), nil
}

func HandleHelloError(ctx *r.Context, _ struct{}) (r.Result, error) {
	return r.Result{}.Notify("error", "Hello error"), nil
}

func HandleHelloDelay(ctx *r.Context, _ struct{}) (r.Result, error) {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Context().Done():
		return r.Result{}, ctx.Context().Err()
	case <-timer.C:
	}
	return r.Result{}.Notify("info", "Information (after 2s delay)"), nil
}

func HandleHelloCrash(ctx *r.Context, _ struct{}) (r.Result, error) {
	panic("Hello again")
}

func RegisterHello(app *r.App) {
	app.Page("/hello", Hello)
	r.RegisterAction(app, "hello.ok", HandleHelloOk)
	r.RegisterAction(app, "hello.error", HandleHelloError)
	r.RegisterAction(app, "hello.delay", HandleHelloDelay)
	r.RegisterAction(app, "hello.crash", HandleHelloCrash)
}
