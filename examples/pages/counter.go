package pages

import (
	"fmt"

	"github.com/michalCapo/g-sui/ui"
)

// we want to create counter with initial count value
func Counter(count int) *TCounter {
	return &TCounter{Count: count}
}

// struct definition
type TCounter struct {
	Count int
}

func (counter *TCounter) Increment(ctx *ui.Context) string {
	err := ctx.Body(counter)
	if err != nil {
		panic(err)
	}

	counter.Count++
	return counter.render(ctx)
}

func (counter *TCounter) Decrement(ctx *ui.Context) string {
	err := ctx.Body(counter)
	if err != nil {
		panic(err)
	}

	counter.Count--
	if counter.Count < 0 {
		counter.Count = 0
	}
	return counter.render(ctx)
}

func (counter *TCounter) render(ctx *ui.Context) string {
	target := ui.Target()
	return ui.Div("flex gap-2 items-center bg-purple-500 rounded text-white p-px", target)(
		ui.Button().
			Click(ctx.Call(counter.Decrement, counter).Replace(target)).
			Class("rounded-l px-5").
			Render("-"),
		ui.Div("text-2xl")(fmt.Sprintf("%d", counter.Count)),
		ui.Button().
			Click(ctx.Call(counter.Increment, counter).Replace(target)).
			Class("rounded-r px-5").
			Render("+"),
	)
}
