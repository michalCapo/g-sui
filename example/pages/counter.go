package pages

import (
	"fmt"

	r "github.com/michalCapo/g-sui/ui"
)

func counterWidget(id string, count int) *r.Node {
	return r.Div("flex gap-2 items-center bg-purple-500 rounded text-white p-px").ID(id).Render(
		r.Button("rounded-l px-5 cursor-pointer hover:bg-purple-600").
			Text("-").
			OnClick(&r.Action{Name: "counter.dec", Data: map[string]any{"id": id, "count": count}}),
		r.Div("text-2xl px-3").Text(fmt.Sprintf("%d", count)),
		r.Button("rounded-r px-5 cursor-pointer hover:bg-purple-600").
			Text("+").
			OnClick(&r.Action{Name: "counter.inc", Data: map[string]any{"id": id, "count": count}}),
	)
}

func Counter(ctx *r.Context) *r.Node {
	id1 := r.Target()
	id2 := r.Target()

	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Counter"),
		r.Div("text-gray-600").Text("Stateful counter using WebSocket actions."),
		r.Div("flex gap-4").Render(
			counterWidget(id1, 3),
			counterWidget(id2, 5),
		),
	)
}

func HandleCounterInc(ctx *r.Context) string {
	var data struct {
		ID    string  `json:"id"`
		Count float64 `json:"count"`
	}
	ctx.Body(&data)
	newCount := int(data.Count) + 1
	return counterWidget(data.ID, newCount).ToJSReplace(data.ID)
}

func HandleCounterDec(ctx *r.Context) string {
	var data struct {
		ID    string  `json:"id"`
		Count float64 `json:"count"`
	}
	ctx.Body(&data)
	newCount := int(data.Count) - 1
	if newCount < 0 {
		newCount = 0
	}
	return counterWidget(data.ID, newCount).ToJSReplace(data.ID)
}
