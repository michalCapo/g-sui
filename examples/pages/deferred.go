package pages

import (
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// form used to choose skeleton type via ctx.Body
type deferForm struct {
	As ui.Skeleton
}

// Deffered matches the TS example: returns a skeleton immediately, then
// pushes two patches when data is ready (replace + append).
func Deffered(ctx *ui.Context) string {
	target := ui.Target()

	// read optional skeleton preference from body
	form := deferForm{}
	_ = ctx.Body(&form)

	// LazyLoadData: replace after ~2s
	go func() {
		defer func() { recover() }()
		time.Sleep(2 * time.Second)
		html := ui.Div("space-y-4", target)(
			ui.Div("bg-gray-50 dark:bg-gray-900 p-4 rounded shadow border rounded p-4")(
				ui.Div("text-lg font-semibold")("Deferred content loaded"),
				ui.Div("text-gray-600 text-sm")("This block replaced the skeleton via WebSocket patch."),
			),
		)
		ctx.Patch(target.Replace(), html)
	}()

	// LazyMoreData: append controls after ~3s
	go func() {
		defer func() { recover() }()
		time.Sleep(2100 * time.Millisecond)
		// helpers for calls with different skeletons
		callDefault := ctx.Call(Deffered, &deferForm{}).Replace(target)
		callComponent := ctx.Call(Deffered, &deferForm{As: ui.SkeletonComponent}).Replace(target)
		callList := ctx.Call(Deffered, &deferForm{As: ui.SkeletonList}).Replace(target)
		callPage := ctx.Call(Deffered, &deferForm{As: ui.SkeletonPage}).Replace(target)
		callForm := ctx.Call(Deffered, &deferForm{As: ui.SkeletonForm}).Replace(target)

		controls := ui.Div("grid grid-cols-5 gap-4")(
			ui.Button().Color(ui.Blue).Class("rounded text-sm").Click(callDefault).Render("Default skeleton"),
			ui.Button().Color(ui.Blue).Class("rounded text-sm").Click(callComponent).Render("Component skeleton"),
			ui.Button().Color(ui.Blue).Class("rounded text-sm").Click(callList).Render("List skeleton"),
			ui.Button().Color(ui.Blue).Class("rounded text-sm").Click(callPage).Render("Page skeleton"),
			ui.Button().Color(ui.Blue).Class("rounded text-sm").Click(callForm).Render("Form skeleton"),
		)
		ctx.Patch(target.Append(), controls)
	}()

	// return the chosen skeleton immediately
	return target.Skeleton(form.As)
}

// DeferredContent keeps the route stable by delegating to Deffered.
func DeferredContent(ctx *ui.Context) string { return Deffered(ctx) }
