package main

import (
	"embed"

	"github.com/michalCapo/g-sui/example/pages"
	r "github.com/michalCapo/g-sui/ui"
)

//go:embed assets/*
var assets embed.FS

func main() {
	app := r.NewApp()

	// Register all pages (each file owns its routes + actions)
	pages.RegisterShowcase(app, layout)
	pages.RegisterIcons(app, layout)
	pages.RegisterButton(app, layout)
	pages.RegisterText(app, layout)
	pages.RegisterPassword(app, layout)
	pages.RegisterNumber(app, layout)
	pages.RegisterDate(app, layout)
	pages.RegisterArea(app, layout)
	pages.RegisterSelect(app, layout)
	pages.RegisterCheckbox(app, layout)
	pages.RegisterRadio(app, layout)
	pages.RegisterTable(app, layout)
	pages.RegisterForm(app, layout)
	pages.RegisterLogin(app, layout)
	pages.RegisterOthers(app, layout)
	pages.RegisterAppend(app, layout)
	pages.RegisterClock(app, layout)
	pages.RegisterShared(app, layout)
	pages.RegisterReloadRedirect(app, layout)
	pages.RegisterRoutes(app, layout)
	pages.RegisterSkeleton(app, layout)
	pages.RegisterCounter(app, layout)
	pages.RegisterHello(app, layout)
	pages.RegisterCollate(app, layout)

	// Serve embedded static assets (favicon, images, etc.)
	app.Assets(assets, "assets", "/assets/")
	app.Favicon = "/assets/favicon.svg"
	app.Title = "g-sui Component Showcase"
	app.Description = "A server-rendered Go UI framework with live WebSocket updates, Tailwind CSS, and interactive components."

	app.Listen(":1423")
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

func layout(content *r.Node) *r.Node {
	return r.Div("min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors overflow-y-scroll").Render(
		r.Nav("bg-white dark:bg-gray-900 shadow dark:shadow-gray-800/50").Attr("aria-label", "Main navigation").Render(
			r.Div("mx-auto px-4 py-3 flex items-start gap-2").Render(
				r.Div("flex flex-wrap gap-1 flex-1").Render(
					navLink("Showcase", "nav.showcase"),
					navLink("Icons", "nav.icons"),
					navLink("Button", "nav.button"),
					navLink("Text", "nav.text"),
					navLink("Password", "nav.password"),
					navLink("Number", "nav.number"),
					navLink("Date", "nav.date"),
					navLink("Textarea", "nav.area"),
					navLink("Select", "nav.select"),
					navLink("Checkbox", "nav.checkbox"),
					navLink("Radio", "nav.radio"),
					navLink("Table", "nav.table"),
					navLink("Form", "nav.form"),
					navLink("Login", "nav.login"),
					navLink("Others", "nav.others"),
					navLink("Append", "nav.append"),
					navLink("Clock", "nav.clock"),
					navLink("Shared", "nav.shared"),
					navLink("Reload", "nav.reload"),
					navLink("Routes", "nav.routes"),
					navLink("Skeleton", "nav.skeleton"),
					navLink("Collate", "nav.collate"),
				),
				r.ThemeSwitcher(),
			),
		),
		r.Main("max-w-5xl mx-auto px-4 py-8").ID(pages.ContentID).Render(
			content,
		),
	)
}

func navLink(label, action string) *r.Node {
	return r.Button("px-3 py-1.5 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300 cursor-pointer").
		Text(label).
		OnClick(&r.Action{Name: action})
}
