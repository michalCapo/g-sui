package pages

import (
	"fmt"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// ClockContent demonstrates server-initiated WS patches updating a target every second.
func ClockContent(ctx *ui.Context) string {
	target := ui.Target()

	pad2 := func(n int) string {
		if n < 10 {
			return fmt.Sprintf("0%d", n)
		}
		return fmt.Sprintf("%d", n)
	}

	render := func(t time.Time) string {
		hh, mm, ss := pad2(t.Hour()), pad2(t.Minute()), pad2(t.Second())
		return ui.Div("font-mono text-3xl", target)(hh + ":" + mm + ":" + ss)
	}

    // Start pushes a patch every second; stops automatically when the target disappears (invalid target).
    start := func(c *ui.Context) string {
        stop := make(chan struct{})
        // Register clear() so the server can stop the ticker when the browser reports target id invalid.
        c.Patch(target, ui.OUTLINE, render(time.Now()), func() { close(stop) })
        go func() {
            tick := time.NewTicker(time.Second)
            defer tick.Stop()
            for {
                select {
                case <-stop:
                    return
                case <-tick.C:
                    c.Patch(target, ui.OUTLINE, render(time.Now()))
                }
            }
        }()
        return ""
    }

	return ui.Div("max-w-5xl mx-auto p-6 flex flex-col gap-4")(
		ui.Div("text-xl font-bold")("Live Clock (WS patches)"),
        ui.Div("text-gray-600")("Updates replace the target via WebSocket patches every second. Background updates stop automatically when the element disappears (invalid target)."),
		// initial render
		render(time.Now()),
		// control
		ui.Div("flex gap-2")(
			ui.Button().Color(ui.Blue).Class("rounded").Click(ctx.Call(start).None()).Render("Start Live"),
		),
	)
}
