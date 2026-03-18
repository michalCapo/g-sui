package pages

import (
	"fmt"
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

func Clock(ctx *r.Context) *r.Node {
	t := time.Now()
	timeStr := pad2(t.Hour()) + ":" + pad2(t.Minute()) + ":" + pad2(t.Second())

	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Live Clock (WS patches)"),
		r.Div("text-gray-600").Text("Updates via WebSocket patches every second."),
		r.Div("font-mono text-3xl bg-white p-4 border rounded").ID("live-clock").Text(timeStr).
			JS("__ws.callSilent('clock.start')"),
	)
}

// HandleClockStart is a WS action that spawns a goroutine to push
// live clock updates every second over the caller's WebSocket connection.
func HandleClockStart(ctx *r.Context) string {
	go func() {
		defer func() { recover() }()
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for range tick.C {
			now := time.Now()
			ts := pad2(now.Hour()) + ":" + pad2(now.Minute()) + ":" + pad2(now.Second())
			node := r.Div("font-mono text-3xl bg-white p-4 border rounded").ID("live-clock").Text(ts)
			if err := ctx.Push(node.ToJSReplace("live-clock")); err != nil {
				return
			}
		}
	}()
	return ""
}

func pad2(n int) string {
	if n < 10 {
		return fmt.Sprintf("0%d", n)
	}
	return fmt.Sprintf("%d", n)
}

func RegisterClock(app *r.App, layout func(*r.Node) *r.Node) {
	app.Page("/clock", func(ctx *r.Context) *r.Node { return layout(Clock(ctx)) })
	app.Action("nav.clock", NavTo("/clock", func() *r.Node { return Clock(nil) }))
	app.Action("clock.start", HandleClockStart)
}
