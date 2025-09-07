package pages

import (
	"time"

	"github.com/michalCapo/g-sui/ui"
)

// form used to choose skeleton type via ctx.Body
type deferForm struct {
	As ui.Skeleton
}

var target = ui.Target()

// Deffered matches the TS example: returns a skeleton immediately, then
// pushes two patches when data is ready (replace + append).
func DefferedComponent(ctx *ui.Context) string {
	// read optional skeleton preference from body
	form := deferForm{}
	err := ctx.Body(&form)

	if err != nil {
		panic(err)
	}

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

	// LazyMoreData: append controls after ~2s
	go func() {
		defer func() { recover() }()
		time.Sleep(2100 * time.Millisecond)

		controls := ui.Div("grid grid-cols-5 gap-4")(
			ui.
				Button().
				Color(ui.Blue).
				Class("rounded text-sm").
				Click(ctx.Call(DefferedComponent, &deferForm{}).Replace(target)).
				Render("Default skeleton"),

			ui.
				Button().
				Color(ui.Blue).
				Class("rounded text-sm").
				Click(ctx.Call(DefferedComponent, &deferForm{As: ui.SkeletonComponent}).Replace(target)).
				Render("Component skeleton"),

			ui.
				Button().
				Color(ui.Blue).
				Class("rounded text-sm").
				Click(ctx.Call(DefferedComponent, &deferForm{As: ui.SkeletonList}).Replace(target)).
				Render("List skeleton"),

			ui.
				Button().
				Color(ui.Blue).
				Class("rounded text-sm").
				Click(ctx.Call(DefferedComponent, &deferForm{As: ui.SkeletonPage}).Replace(target)).
				Render("Page skeleton"),

			ui.
				Button().
				Color(ui.Blue).
				Class("rounded text-sm").
				Click(ctx.Call(DefferedComponent, &deferForm{As: ui.SkeletonForm}).Replace(target)).
				Render("Form skeleton"),
		)

		ctx.Patch(target.Append(), controls)
	}()

	// return the chosen skeleton immediately
	return target.Skeleton(form.As)
}

func Deffered(ctx *ui.Context) string {

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Deferred"),
		ui.Div("text-gray-600")("Returns a skeleton immediately, then replaces the target and appends controls after a short delay."),
		DefferedComponent(ctx),
	)
}
