package main

import (
	"embed"

	"github.com/michalCapo/g-sui/example/pages"
	r "github.com/michalCapo/g-sui/ui"
)

//go:embed assets/*
var assets embed.FS

const contentID = "__content__"

func main() {
	app := r.NewApp()
	app.Layout(layout)

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

	// Serve embedded static assets (favicon, images, etc.)
	app.Assets(assets, "assets", "/assets/")
	app.Favicon = "/assets/favicon.svg"
	app.Title = "g-sui Component Showcase"
	app.Description = "A server-rendered Go UI framework with live WebSocket updates, Tailwind CSS, and interactive components."

	app.Listen(":1424")
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

func layout(ctx *r.Context) *r.Node {
	// The layout persists while the runtime swaps pages.
	ctx.HeadJS(`(function(){
if(window.__navHl)return;window.__navHl=true;
var ACT=['bg-blue-100','dark:bg-blue-900/40','text-blue-700','dark:text-blue-300','font-medium'];
var INACT=['text-gray-700','dark:text-gray-300'];
function upd(){var p=location.pathname;document.querySelectorAll('[data-nav-path]').forEach(function(b){
var on=b.getAttribute('data-nav-path')===p;
ACT.forEach(function(c){b.classList.toggle(c,on)});
INACT.forEach(function(c){b.classList.toggle(c,!on)});
if(on)b.setAttribute('aria-current','page');else b.removeAttribute('aria-current');
})}
window.addEventListener('gsui:navigated',upd);
if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',upd);else upd();
})();`)
	return r.Div("min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors").Render(
		r.Nav("bg-white dark:bg-gray-900 shadow dark:shadow-gray-800/50").Attr("aria-label", "Main navigation").Render(
			r.Div("mx-auto px-4 py-3 flex items-start gap-2").Render(
				r.Div("flex flex-wrap gap-1 flex-1").Render(
					navLink(ctx, "Showcase", "/"), navLink(ctx, "Icons", "/icons"), navLink(ctx, "Button", "/button"),
					navLink(ctx, "Text", "/text"), navLink(ctx, "Password", "/password"), navLink(ctx, "Number", "/number"),
					navLink(ctx, "Date", "/date"), navLink(ctx, "Textarea", "/area"), navLink(ctx, "Select", "/select"),
					navLink(ctx, "Checkbox", "/checkbox"), navLink(ctx, "Radio", "/radio"), navLink(ctx, "Table", "/table"),
					navLink(ctx, "Form", "/form"), navLink(ctx, "Login", "/login"), navLink(ctx, "Others", "/others"),
					navLink(ctx, "Append", "/append"), navLink(ctx, "Clock", "/clock"), navLink(ctx, "Shared", "/shared"),
					navLink(ctx, "Reload", "/reload-redirect"), navLink(ctx, "Routes", "/routes"), navLink(ctx, "Skeleton", "/skeleton"),
					navLink(ctx, "Counter", "/counter"),
					navLink(ctx, "Collate", "/collate"),
				),
				r.ThemeSwitcher(),
			),
		),
		r.Main("max-w-5xl mx-auto px-4 py-8").ID(contentID),
	)
}

func navLink(ctx *r.Context, label, path string) *r.Node {
	cls := "px-3 py-1.5 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300 cursor-pointer"
	active := ctx != nil && ctx.Request != nil && ctx.Request.URL.Path == path
	if active {
		cls += " bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 font-medium"
	}
	n := r.NavLink(path, cls).
		Attr("data-nav-path", path).
		Text(label)
	if active {
		n.Attr("aria-current", "page")
	}
	return n
}
