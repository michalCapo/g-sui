// Run with: go run ./example/live
package main

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	ui "github.com/michalCapo/g-sui/ui"
)

type renameInput struct {
	Name string `json:"name"`
}

func (in *renameInput) Validate() error {
	if len(in.Name) < 2 {
		return ui.ValidationError{Fields: ui.FormErrors{"name": "Enter at least two characters"}}
	}
	return nil
}

type counter struct{ Count int }

type product struct{ Name string }

func (v *counter) Render(ctx *ui.ViewContext) *ui.Node {
	return ui.Div("space-y-4").Title("Counter").Render(
		ui.H1("text-2xl font-semibold").Text("Counter"),
		ui.P().Text("This count belongs to this tab. Returning to this page or reconnecting starts a fresh view."),
		ui.Span("text-3xl").ID("count").Text(strconv.Itoa(v.Count)),
		ui.Button("border rounded px-4 py-2").ID("increment").Text("+1").OnClick(ctx.Event("increment")),
		ui.Label().Attr("for", "draft").Text("Unfinished note"),
		ui.IText("border rounded p-2").ID("draft").Attr("placeholder", "Kept when the count updates"),
		ui.NavLink("/counter?tab=details").Attr("data-gsui-patch", "").Text("Change URL without resetting the count"),
	)
}
func (v *counter) Handle(ctx *ui.ViewContext, event ui.Event) error {
	if event.Name == "increment" {
		v.Count++
	}
	return nil
}

func newApp() *ui.App {
	app := ui.NewApp()
	app.Title = "Server-driven Go app"
	app.Layout(func(ctx *ui.Context) *ui.Node {
		return ui.Div("max-w-3xl mx-auto p-6 space-y-6").Render(
			ui.Nav("flex gap-6 border-b pb-4").Render(
				ui.NavLink("/").Text("Profile"),
				ui.NavLink("/counter").Text("Counter"),
				ui.NavLink("/products").Text("Products"),
			),
			ui.Main().ID("__content__"),
		)
	})
	rename := ui.RegisterAction(app, "profile.rename", func(ctx *ui.Context, in renameInput) (ui.Result, error) {
		// Demo-only, connection-scoped state. Real applications write to their
		// database using ctx.User() and ctx.Context().
		ctx.Session["name"] = in.Name
		return ui.Refresh("profile").Toast("Saved"), nil
	})
	app.Subscription("clock", func(ctx *ui.Context) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Context().Done():
				return ctx.Context().Err()
			case now := <-ticker.C:
				if err := ctx.Push(ui.SetText("clock", now.Format("15:04:05"))); err != nil {
					return err
				}
			}
		}
	})
	app.Page("/", func(ctx *ui.Context) *ui.Node {
		name, _ := ctx.Session["name"].(string)
		if name == "" {
			name = "Guest"
		}
		return ui.Div("space-y-6").Title("Profile").Render(
			ui.H1("text-2xl font-semibold").Text("Profile"),
			ui.Region("profile", ui.P().ID("greeting").Text("Hello, "+name)),
			ui.FormFor[renameInput]("rename", "space-y-3").Render(
				ui.Label().Attr("for", "name").Text("Name"),
				ui.IText("border rounded p-2").ID("name").Attr("name", "name").Attr("required", ""),
				ui.Button("border rounded px-4 py-2").Attr("type", "submit").Text("Save"),
			).Submit(rename),
			ui.P().ID("clock").Text(time.Now().Format("15:04:05")).Subscribe("clock"),
		)
	})
	app.Live("/counter", func() ui.View { return &counter{} })
	table := ui.RegisterTable(app, "products", func(ctx *ui.Context, q ui.TableQuery) (ui.TablePage[product], error) {
		var rows []*product
		for _, name := range []string{"Apple", "Apricot", "Banana", "Cherry", "Grape", "Lemon", "Orange", "Peach"} {
			if strings.Contains(strings.ToLower(name), strings.ToLower(q.Search)) {
				rows = append(rows, &product{Name: name})
			}
		}
		if q.SortColumn == 0 {
			sort.Slice(rows, func(i, j int) bool {
				if q.Direction == "desc" {
					return rows[i].Name > rows[j].Name
				}
				return rows[i].Name < rows[j].Name
			})
		}
		total := len(rows)
		return ui.TablePage[product]{Rows: rows[:min(q.Limit(), total)], Total: total}, nil
	}, func(table *ui.DataTable[product]) {
		table.PageSize(3).RowKey(func(p *product) string { return p.Name }).Col("Name", ui.ColOpt[product]{Sortable: true, Text: func(p *product) *ui.Node { return ui.Span().Text(p.Name) }})
	})
	app.Page("/products", func(ctx *ui.Context) *ui.Node {
		node, err := table.Render(ctx)
		if err != nil {
			return ui.P().Text("Products unavailable")
		}
		return ui.Div().Title("Products").Render(ui.H1("text-2xl font-semibold mb-4").Text("Products"), node)
	})
	return app
}

func main() {
	app := newApp()
	if err := app.Listen("127.0.0.1:1425"); err != nil {
		log.Fatal(err)
	}
}
