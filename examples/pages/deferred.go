package pages

import (
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// DeferredContent demonstrates ctx.Defer with a skeleton that is replaced by
// server-rendered content via WebSocket patch when ready.
func DeferredContent(ctx *ui.Context) string {
	target := ui.Target()

	heavy := func(c *ui.Context) string {
		time.Sleep(900 * time.Millisecond)
		return ui.Div("bg-white dark:bg-gray-900 p-4 rounded shadow border", target)(
			ui.Div("font-semibold")("Deferred content loaded"),
			ui.Div("text-gray-500 text-sm")("Replaced via WS patch"),
		)
	}

	// Show a component skeleton immediately, replace when the content is ready
	placeholder := ctx.Defer(heavy).SkeletonComponent().Replace(target)

	return ui.Div("max-w-5xl mx-auto p-6 flex flex-col gap-4")(
		ui.Div("text-xl font-bold")("Deferred Fragment"),
		ui.Div("text-gray-600")("Shows a skeleton that is replaced when the server finishes rendering."),
		placeholder,
	)
}
