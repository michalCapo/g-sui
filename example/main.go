// Command example runs the g-sui component showcase.
package main

import (
	"embed"
	"log"

	"github.com/michalCapo/g-sui/example/pages"
	r "github.com/michalCapo/g-sui/ui"
)

//go:embed assets/*
var assets embed.FS

const contentID = "__content__"

func main() {
	app := r.NewApp()
	app.Layout(layout)
	// Pages with a skeleton (see pages/skeleton.go) show it while they load;
	// no page-wide overlay.
	app.Loader = r.LoaderNone

	// Register all pages (each file owns its routes + actions)
	pages.RegisterShowcase(app)
	pages.RegisterIcons(app)
	pages.RegisterButton(app)
	pages.RegisterText(app)
	pages.RegisterPassword(app)
	pages.RegisterNumber(app)
	pages.RegisterDate(app)
	pages.RegisterArea(app)
	pages.RegisterSelect(app)
	pages.RegisterCheckbox(app)
	pages.RegisterRadio(app)
	pages.RegisterTable(app)
	pages.RegisterForm(app)
	pages.RegisterLogin(app)
	pages.RegisterOthers(app)
	pages.RegisterAppend(app)
	pages.RegisterClock(app)
	pages.RegisterShared(app)
	pages.RegisterReloadRedirect(app)
	pages.RegisterRoutes(app)
	pages.RegisterSkeleton(app)
	pages.RegisterCounter(app)
	pages.RegisterHello(app)
	pages.RegisterCollate(app)
	pages.RegisterActions(app)

	// Serve embedded static assets (favicon, images, etc.)
	app.Assets(assets, "assets", "/assets/")
	app.Favicon = "/assets/favicon.svg"
	app.Title = "g-sui Component Showcase"
	app.Description = "A server-rendered Go UI framework with live WebSocket updates, Tailwind CSS, and interactive components."
	// Installable app with an offline page; the Install button is in the layout.
	app.PWA(r.PWA{
		Name:           "g-sui Component Showcase",
		ShortName:      "g-sui",
		ThemeColor:     "#f9fafb",
		DarkThemeColor: "#030712",
		Icons:          []r.PWAIcon{{Src: "/assets/favicon.svg", Sizes: "any", Type: "image/svg+xml"}},
	})

	log.Fatal(app.Listen(":1424"))
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

func layout(_ *r.Context) *r.Node {
	// The layout persists while the runtime swaps pages.
	return r.Div("min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors").Render(
		r.Nav("bg-white dark:bg-gray-900 shadow dark:shadow-gray-800/50").Attr("aria-label", "Main navigation").Render(
			r.Div("mx-auto px-4 py-3 flex items-start gap-2").Render(
				r.Div("flex flex-wrap gap-1 flex-1").Render(
					navLink("Showcase", "/"), navLink("Icons", "/icons"), navLink("Button", "/button"),
					navLink("Text", "/text"), navLink("Password", "/password"), navLink("Number", "/number"),
					navLink("Date", "/date"), navLink("Textarea", "/area"), navLink("Select", "/select"),
					navLink("Checkbox", "/checkbox"), navLink("Radio", "/radio"), navLink("Table", "/table"),
					navLink("Form", "/form"), navLink("Login", "/login"), navLink("Others", "/others"),
					navLink("Append", "/append"), navLink("Clock", "/clock"), navLink("Shared", "/shared"),
					navLink("Reload", "/reload-redirect"), navLink("Routes", "/routes"), navLink("Skeleton", "/skeleton"),
					navLink("Counter", "/counter"),
					navLink("Collate", "/collate"),
					navLink("Actions", "/actions"),
				),
				r.Button("px-3 py-1.5 rounded text-sm bg-blue-600 text-white").Text("Install").
					ShowIfInstallable().OnClick(r.InstallApp("install-help")),
				r.ThemeSwitcher(),
			),
		),
		// Steps for browsers without an install prompt (iPhone, iPad, Safari).
		r.El("dialog", "m-auto max-w-sm rounded-lg p-6 bg-white dark:bg-gray-900 text-gray-800 dark:text-gray-100 backdrop:bg-black/40").ID("install-help").Render(
			r.P().Text("Tap Share, then Add to Home Screen. In Safari on a Mac, choose File, then Add to Dock."),
			r.Button("mt-4 px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-800").Text("Close").OnClick(r.CloseDialog("install-help")),
		),
		r.Main("max-w-5xl mx-auto px-4 py-8").ID(contentID),
	)
}

func navLink(label, path string) *r.Node {
	return r.NavLink(path, "px-3 py-1.5 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer").
		ActiveClass("bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 font-medium", "text-gray-700 dark:text-gray-300").
		Text(label)
}
