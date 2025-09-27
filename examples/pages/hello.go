package pages

import (
	"time"

	"github.com/michalCapo/g-sui/ui"
)

func HelloContent(ctx *ui.Context) string {
	sayHello := func(ctx *ui.Context) string { ctx.Success("Hello"); return "" }
	sayDelay := func(ctx *ui.Context) string { time.Sleep(2 * time.Second); ctx.Info("Information"); return "" }
	sayError := func(ctx *ui.Context) string { ctx.Error("Hello error"); return "" }
	sayCrash := func(_ *ui.Context) string { panic("Hello again") }

	return ui.Div("flex justify-start gap-4 items-center")(
		ui.Button().
			Color(ui.GreenOutline).
			Class("rounded").
			Click(ctx.Call(sayHello).None()).
			Render("with ok"),

		ui.Button().
			Color(ui.RedOutline).
			Class("rounded").
			Click(ctx.Call(sayError).None()).
			Render("with error"),

		ui.Button().
			Color(ui.BlueOutline).
			Class("rounded").
			Click(ctx.Call(sayDelay).None()).
			Render("with delay"),

		ui.Button().
			Color(ui.YellowOutline).
			Class("rounded").
			Click(ctx.Call(sayCrash).None()).
			Render("with crash"),
	)
}
