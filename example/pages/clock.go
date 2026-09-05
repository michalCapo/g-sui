package pages

import (
	r "github.com/michalCapo/g-sui/ui"
	"time"
)

const liveClockID = "live-clock"

func Clock(ctx *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Live Clock (WS patches)"),
		r.Div("text-gray-600").Text("Updates via WebSocket patches every second."),
		r.Div("font-mono text-3xl bg-white p-4 border rounded").ID(liveClockID).Text(time.Now().Format("15:04:05")).Subscribe("clock"),
	)
}

func RegisterClock(app *r.App) {
	app.Page("/clock", Clock)
	app.Subscription("clock", func(ctx *r.Context) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Context().Done():
				return ctx.Context().Err()
			case now := <-ticker.C:
				if err := ctx.Push(r.Result{}.SetText(liveClockID, now.Format("15:04:05"))); err != nil {
					return err
				}
			}
		}
	})
}
