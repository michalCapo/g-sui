package pages

import (
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

var helloOk, helloError, helloDelay, helloCrash r.ActionRef[struct{}]

func Hello(_ *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Hello Actions"),
		r.Div("text-gray-600").Text("Click buttons to trigger different server action responses."),
		r.Div("flex justify-start gap-4 items-center").Render(
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-green-600 text-green-600 hover:bg-green-50").
				Text("with ok").
				OnClick(helloOk.Call(struct{}{})),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-red-600 text-red-600 hover:bg-red-50").
				Text("with error").
				OnClick(helloError.Call(struct{}{})),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-blue-600 text-blue-600 hover:bg-blue-50").
				Text("with delay").
				OnClick(helloDelay.Call(struct{}{})),
			r.Button("px-4 py-2 rounded cursor-pointer border-2 border-yellow-600 text-yellow-600 hover:bg-yellow-50").
				Text("with crash").
				OnClick(helloCrash.Call(struct{}{})),
		),
	)
}

func HandleHelloOk(_ *r.Context, _ struct{}) (r.Result, error) {
	return r.Result{}.Notify("success", "Hello"), nil
}

func HandleHelloError(_ *r.Context, _ struct{}) (r.Result, error) {
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

func HandleHelloCrash(_ *r.Context, _ struct{}) (r.Result, error) {
	panic("Hello again")
}

func RegisterHello(app *r.App) {
	helloOk = r.RegisterAction(app, "hello.ok", HandleHelloOk)
	helloError = r.RegisterAction(app, "hello.error", HandleHelloError)
	helloDelay = r.RegisterAction(app, "hello.delay", HandleHelloDelay)
	helloCrash = r.RegisterAction(app, "hello.crash", HandleHelloCrash)
	app.Page("/hello", Hello)
}
