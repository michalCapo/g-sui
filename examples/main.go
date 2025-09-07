package main

import (
	"embed"
	"time"

	"github.com/michalCapo/g-sui/examples/pages"
	"github.com/michalCapo/g-sui/ui"
)

//go:embed assets/*
var assets embed.FS

// simple registry of routes for menu rendering
type route struct {
	Path  string
	Title string
}

// align the navigation with the TS examples
var routes = []route{
	{Path: "/", Title: "Showcase"},
	{Path: "/button", Title: "Button"},
	{Path: "/text", Title: "Text"},
	{Path: "/password", Title: "Password"},
	{Path: "/number", Title: "Number"},
	{Path: "/date", Title: "Date & Time"},
	{Path: "/area", Title: "Textarea"},
	{Path: "/select", Title: "Select"},
	{Path: "/checkbox", Title: "Checkbox"},
	{Path: "/radio", Title: "Radio"},
	{Path: "/table", Title: "Table"},
	{Path: "/others", Title: "Others"},
	{Path: "/append", Title: "Append/Prepend"},
	{Path: "/clock", Title: "Clock"},
	{Path: "/deferred", Title: "Deferred"},
}

func main() {
	app := ui.MakeApp("en")
	// Example favicon (served at /favicon.ico) and link tag
	app.Favicon(assets, "assets/favicon.svg", 24*time.Hour)
	app.AutoRestart(true) // enable if you want the examples to rebuild on changes

	// layout builder with top menu styled like TS examples
	layout := func(title string, body func(*ui.Context) string) ui.Callable {
		return func(ctx *ui.Context) string {
			nav := ui.Div("bg-white shadow mb-6")(
				ui.Div("max-w-5xl mx-auto px-4 py-2 flex items-center gap-2")(
					// top bar
					ui.Div("flex flex-wrap gap-1 mt-2 md:mt-0")(
						ui.Map(routes, func(r *route, _ int) string {
							base := "px-2 py-1 rounded text-sm whitespace-nowrap"
							cls := base + " hover:bg-gray-200"
							// Highlight the currently selected route
							if ctx != nil && ctx.Request != nil && r.Path == ctx.Request.URL.Path {
								cls = base + " bg-blue-700 text-white hover:bg-blue-600"
							}

							return ui.A(cls, ui.Href(r.Path), ctx.Load(r.Path))(r.Title)
						}),
					),
					ui.Flex1,
					ui.ThemeSwitcher(""),
				),
			)

			content := body(ctx)
			return app.HTML(title, "bg-gray-100 min-h-screen", nav+ui.Div("max-w-5xl mx-auto px-2")(content))
		}
	}

	// Individual example pages
	app.Page("/", layout("Showcase", pages.Showcase))
	app.Page("/button", layout("Button", pages.Button))
	app.Page("/text", layout("Text", pages.Text))
	app.Page("/password", layout("Password", pages.Password))
	app.Page("/number", layout("Number", pages.Number))
	app.Page("/date", layout("Date & Time", pages.Date))
	app.Page("/area", layout("Textarea", pages.Area))
	app.Page("/select", layout("Select", pages.Select))
	app.Page("/checkbox", layout("Checkbox", pages.Checkbox))
	app.Page("/radio", layout("Radio", pages.Radio))
	app.Page("/table", layout("Table", pages.Table))
	app.Page("/others", layout("Others", pages.Others))
	app.Page("/append", layout("Append / Prepend", pages.Append))
	app.Page("/clock", layout("Clock", pages.Clock))
	app.Page("/deferred", layout("Deferred", pages.Deffered))

	app.Listen(":1422")
}
