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

	app.Listen(":1424")
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

func layout(ctx *r.Context, content *r.Node) *r.Node {
	// Keep the nav highlight in sync with the URL: SPA navigation only swaps
	// the content area, so the server-rendered active state goes stale. This
	// re-applies it on every pushState (menu clicks) and popstate (back/forward).
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
var ps=history.pushState;history.pushState=function(){ps.apply(this,arguments);upd()};
window.addEventListener('popstate',upd);
window.addEventListener('gsui:updated',upd);
if(document.readyState==='loading')document.addEventListener('DOMContentLoaded',upd);else upd();
})();`)
	return r.Div("min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors").Render(
		r.Nav("bg-white dark:bg-gray-900 shadow dark:shadow-gray-800/50").Attr("aria-label", "Main navigation").Render(
			r.Div("mx-auto px-4 py-3 flex items-start gap-2").Render(
				r.Div("flex flex-wrap gap-1 flex-1").Render(
					navLink(ctx, "Showcase", "nav.showcase", "/"), navLink(ctx, "Icons", "nav.icons", "/icons"), navLink(ctx, "Button", "nav.button", "/button"),
					navLink(ctx, "Text", "nav.text", "/text"), navLink(ctx, "Password", "nav.password", "/password"), navLink(ctx, "Number", "nav.number", "/number"),
					navLink(ctx, "Date", "nav.date", "/date"), navLink(ctx, "Textarea", "nav.area", "/area"), navLink(ctx, "Select", "nav.select", "/select"),
					navLink(ctx, "Checkbox", "nav.checkbox", "/checkbox"), navLink(ctx, "Radio", "nav.radio", "/radio"), navLink(ctx, "Table", "nav.table", "/table"),
					navLink(ctx, "Form", "nav.form", "/form"), navLink(ctx, "Login", "nav.login", "/login"), navLink(ctx, "Others", "nav.others", "/others"),
					navLink(ctx, "Append", "nav.append", "/append"), navLink(ctx, "Clock", "nav.clock", "/clock"), navLink(ctx, "Shared", "nav.shared", "/shared"),
					navLink(ctx, "Reload", "nav.reload", "/reload-redirect"), navLink(ctx, "Routes", "nav.routes", "/routes"), navLink(ctx, "Skeleton", "nav.skeleton", "/skeleton"),
					navLink(ctx, "Collate", "nav.collate", "/collate"),
				),
				r.ThemeSwitcher(),
			),
		),
		r.Main("max-w-5xl mx-auto px-4 py-8").ID(pages.ContentID).Render(
			content,
		),
	)
}

func navLink(ctx *r.Context, label, action, path string) *r.Node {
	cls := "px-3 py-1.5 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-700 dark:text-gray-300 cursor-pointer"
	active := ctx != nil && ctx.Request != nil && ctx.Request.URL.Path == path
	if active {
		cls += " bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 font-medium"
	}
	n := r.Button(cls).
		Attr("data-nav-path", path).
		Text(label).
		OnClick(&r.Action{Name: action})
	if active {
		n.Attr("aria-current", "page")
	}
	return n
}
