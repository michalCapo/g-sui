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
	{Path: "/icons", Title: "Icons"},
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
	{Path: "/form", Title: "Form"},
	{Path: "/image-upload", Title: "Image Upload"},
	{Path: "/captcha", Title: "Captcha"},
	{Path: "/others", Title: "Others"},
	{Path: "/append", Title: "Append/Prepend"},
	{Path: "/clock", Title: "Clock"},
	{Path: "/deferred", Title: "Deferred"},
	{Path: "/shared", Title: "Shared"},
	{Path: "/collate", Title: "Collate"},
	{Path: "/collate-empty", Title: "Collate Empty"},
	{Path: "/spa", Title: "SPA"},
	{Path: "/reload-redirect", Title: "Reload & Redirect"},
	{Path: "/routes", Title: "Route Params"},
}

func main() {
	app := ui.MakeApp("en")
	app.Assets(assets, "assets", 24*time.Hour)
	app.Favicon(assets, "assets/favicon.svg", 24*time.Hour)

	// Enable PWA
	// app.PWA(ui.PWAConfig{
	// 	Name:                  "g-sui Showcase",
	// 	ShortName:             "g-sui",
	// 	Description:           "Go Server-Rendered UI Showcase",
	// 	ThemeColor:            "#1d4ed8",
	// 	BackgroundColor:       "#ffffff",
	// 	Display:               "standalone",
	// 	GenerateServiceWorker: true,
	// 	Icons: []ui.PWAIcon{
	// 		{Src: "/favicon.ico", Sizes: "any", Type: "image/x-icon"},
	// 	},
	// })

	// app.AutoRestart(true) // enable if you want the examples to rebuild on changes

	// Define persistent layout with content slot
	app.Layout(func(ctx *ui.Context) string {
		nav := ui.Div("bg-white shadow")(
			ui.Div("mx-auto px-4 py-2 flex items-start gap-2")(
				// top bar
				ui.Div("flex flex-wrap gap-1 mt-2 md:mt-0 w-full")(
					ui.Map(routes, func(r *route, _ int) string {
						base := "px-2 py-1 rounded text-sm whitespace-nowrap"
						cls := base + " hover:bg-gray-200"
						// Highlight will be handled by JS router based on current path
						return ui.A(cls, ctx.Load(r.Path))(r.Title)
					}),
				),
				ui.Flex1,
				ui.ThemeSwitcher(""),
			),
		)

		// Content slot where pages will be rendered
		contentSlot := ui.Div("max-w-5xl mx-auto px-2 py-8", ui.Attr{ID: "__content__"})()

		return nav + contentSlot
	})

	// Individual example pages (content only, no layout wrapper)
	app.Page("/", "Showcase", pages.Showcase)
	app.Page("/icons", "Icons", pages.IconsContent)
	app.Page("/button", "Button", pages.Button)
	app.Page("/text", "Text", pages.Text)
	app.Page("/password", "Password", pages.Password)
	app.Page("/number", "Number", pages.Number)
	app.Page("/date", "Date & Time", pages.Date)
	app.Page("/area", "Textarea", pages.Area)
	app.Page("/select", "Select", pages.Select)
	app.Page("/checkbox", "Checkbox", pages.Checkbox)
	app.Page("/radio", "Radio", pages.Radio)
	app.Page("/form", "Form", pages.FormContent)
	app.Page("/image-upload", "Image Upload", pages.ImageUploadContent)
	app.Page("/table", "Table", pages.Table)
	app.Page("/captcha", "Captcha", pages.Captcha)
	app.Page("/others", "Others", pages.Others)
	app.Page("/append", "Append / Prepend", pages.Append)
	app.Page("/clock", "Clock", pages.Clock)
	app.Page("/deferred", "Deferred", pages.Deffered)
	app.Page("/shared", "Shared", pages.Shared)
	app.Page("/collate", "Collate", pages.Collate)
	app.Page("/collate-empty", "Collate Empty", pages.CollateEmpty)
	app.Page("/spa", "SPA", pages.SpaExample)
	app.Page("/reload-redirect", "Reload & Redirect", pages.ReloadRedirect)
	app.Page("/routes", "Route Parameters", pages.RoutesExample)
	app.Page("/routes/search", "Search", pages.SearchExample)
	app.Page("/routes/user/{id}", "User Detail", pages.UserDetail)
	app.Page("/routes/user/{userId}/post/{postId}", "User Post Detail", pages.UserPostDetail)
	app.Page("/routes/category/{category}/product/{product}", "Category Product Detail", pages.CategoryProductDetail)

	app.Listen(":1422")
}
