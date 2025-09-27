package pages

import (
	"fmt"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// Clock demonstrates server-initiated WS patches updating a target every second.
func Clock(ctx *ui.Context) string {
	target := ui.Target()

	pad2 := func(n int) string {
		if n < 10 {
			return fmt.Sprintf("0%d", n)
		}
		return fmt.Sprintf("%d", n)
	}

	clockUI := func() string {
		t := time.Now()
		hh, mm, ss := pad2(t.Hour()), pad2(t.Minute()), pad2(t.Second())
		return ui.Div("font-mono text-3xl bg-white p-4 border rounded", target)(hh + ":" + mm + ":" + ss)
	}

	// Start pushes a patch every second; stops automatically when the target disappears (invalid target).
	start := func(ctx *ui.Context) {
		stop := make(chan struct{})
		// Register clear() so the server can stop the ticker when the browser reports target id invalid.

		ctx.Patch(target.Replace(), clockUI(), func() { close(stop) })

		go func() {
			tick := time.NewTicker(time.Second)
			defer tick.Stop()
			for {
				select {
				case <-stop:
					return
				case <-tick.C:
					ctx.Patch(target.Replace(), clockUI())
				}
			}
		}()
	}

	// Autostart live updates on initial render
	start(ctx)

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Live Clock (WS patches)"),
		ui.Div("text-gray-600")("Updates replace the target via WebSocket patches every second. Background updates stop automatically when the element disappears (invalid target)."),
		// initial render
		clockUI(),
	)
}
