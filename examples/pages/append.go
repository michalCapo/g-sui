package pages

import (
	"fmt"
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// Demonstrates Append and Prepend swaps on a target container.
func Append(ctx *ui.Context) string {
	target := ui.Target()

	addEnd := func(c *ui.Context) string {
		now := time.Now().Format("15:04:05")
		return ui.Div("p-2 rounded border bg-white dark:bg-gray-900")(
			ui.Span("text-sm text-gray-600")(
				fmt.Sprintf("Appended at %s", now),
			),
		)
	}

	addStart := func(c *ui.Context) string {
		now := time.Now().Format("15:04:05")
		return ui.Div("p-2 rounded border bg-white dark:bg-gray-900")(
			ui.Span("text-sm text-gray-600")(
				fmt.Sprintf("Prepended at %s", now),
			),
		)
	}

	// Controls and container
	controls := ui.Div("flex gap-2")(
		ui.Button().Color(ui.Blue).Class("rounded").Click(ctx.Call(addEnd).Append(target)).Render("Add at end"),
		ui.Button().Color(ui.Green).Class("rounded").Click(ctx.Call(addStart).Prepend(target)).Render("Add at start"),
	)

	container := ui.Div("space-y-2", target)(
		ui.Div("p-2 rounded border bg-white dark:bg-gray-900")(
			ui.Span("text-sm text-gray-600")("Initial item"),
		),
	)

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Append / Prepend Demo"),
		ui.Div("text-gray-600")("Click buttons to insert items at the beginning or end of the list."),
		controls,
		container,
	)
}
