package pages

import (
	r "github.com/michalCapo/g-sui/ui"
	"strconv"
)

const counterDraftID = "draft"

type counterView struct{ Counts [2]int }

func (v *counterView) Render(ctx *r.ViewContext) *r.Node {
	widgets := make([]*r.Node, 0, len(v.Counts))
	for i, count := range v.Counts {
		id := "counter-" + strconv.Itoa(i)
		widgets = append(widgets, r.Div("flex gap-2 items-center bg-purple-500 rounded text-white p-px").ID(id).Render(
			r.Button("rounded-l px-5 cursor-pointer hover:bg-purple-600").ID(id+"-dec").Text("-").OnClick(ctx.Event(id+"-dec")),
			r.Div("text-2xl px-3").ID(id+"-count").Text(strconv.Itoa(count)),
			r.Button("rounded-r px-5 cursor-pointer hover:bg-purple-600").ID(id+"-inc").Text("+").OnClick(ctx.Event(id+"-inc")),
		))
	}
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Title("Counter").Render(
		r.Div("text-2xl font-bold").Text("Counter"),
		r.Div("text-gray-600").Text("Each tab has its own counters. Returning to this page or reconnecting resets them."),
		r.Div("flex gap-4").Render(widgets...),
		r.Label().Attr("for", counterDraftID).Text("Unfinished note"),
		r.IText("border rounded p-2").ID(counterDraftID).Attr("placeholder", "Kept when the count updates"),
		r.NavLink("/counter?tab=details").Attr("data-gsui-patch", "").Text("Change URL without resetting the count"),
	)
}

func (v *counterView) Handle(ctx *r.ViewContext, event r.Event) error {
	for i := range v.Counts {
		switch event.Name {
		case "counter-" + strconv.Itoa(i) + "-inc":
			v.Counts[i]++
		case "counter-" + strconv.Itoa(i) + "-dec":
			v.Counts[i] = max(v.Counts[i]-1, 0)
		}
	}
	return nil
}

func RegisterCounter(app *r.App) {
	app.Live("/counter", func() r.View { return &counterView{Counts: [2]int{3, 5}} })
}
