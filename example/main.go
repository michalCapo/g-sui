package main

import (
	"embed"

	"github.com/michalCapo/g-sui/example/pages"
	r "github.com/michalCapo/g-sui/ui"
)

//go:embed assets/*
var assets embed.FS

func main() {
	store := pages.NewInvoiceStore()
	pages.InitInvoiceStore(store)

	app := r.NewApp()

	// Pages (GET routes – full HTML shell with layout)
	app.Page("/", func(ctx *r.Context) *r.Node { return layout(pages.Showcase(ctx)) })
	app.Page("/invoices", pages.InvoiceListPage(layout))
	app.Page("/invoices/new", pages.InvoiceCreatePage(layout))

	// Invoice actions (WS – replace only content area, no layout)
	app.Action("nav.list", navTo("/invoices", pages.InvoiceListNav))
	app.Action("nav.create", navTo("/invoices/new", pages.InvoiceCreateNav))
	app.Action("invoice.view", pages.HandleInvoiceView)
	app.Action("invoice.delete", pages.HandleInvoiceDelete)
	app.Action("invoice.create", pages.HandleInvoiceCreate)
	app.Action("invoice.addRow", pages.HandleInvoiceAddRow)

	// --- Example pages (ported from examples/) ---

	// Component demos
	app.Page("/button", func(ctx *r.Context) *r.Node { return layout(pages.Button(ctx)) })
	app.Page("/text", func(ctx *r.Context) *r.Node { return layout(pages.Text(ctx)) })
	app.Page("/password", func(ctx *r.Context) *r.Node { return layout(pages.Password(ctx)) })
	app.Page("/number", func(ctx *r.Context) *r.Node { return layout(pages.Number(ctx)) })
	app.Page("/date", func(ctx *r.Context) *r.Node { return layout(pages.Date(ctx)) })
	app.Page("/area", func(ctx *r.Context) *r.Node { return layout(pages.Area(ctx)) })
	app.Page("/select", func(ctx *r.Context) *r.Node { return layout(pages.SelectPage(ctx)) })
	app.Page("/checkbox", func(ctx *r.Context) *r.Node { return layout(pages.Checkbox(ctx)) })
	app.Page("/radio", func(ctx *r.Context) *r.Node { return layout(pages.Radio(ctx)) })
	app.Page("/icons", func(ctx *r.Context) *r.Node { return layout(pages.Icons(ctx)) })
	app.Page("/table", func(ctx *r.Context) *r.Node { return layout(pages.TablePage(ctx)) })
	app.Page("/form", func(ctx *r.Context) *r.Node { return layout(pages.FormPage(ctx)) })
	app.Page("/counter", func(ctx *r.Context) *r.Node { return layout(pages.Counter(ctx)) })
	app.Page("/hello", func(ctx *r.Context) *r.Node { return layout(pages.Hello(ctx)) })
	app.Page("/others", func(ctx *r.Context) *r.Node { return layout(pages.Others(ctx)) })
	app.Page("/append", func(ctx *r.Context) *r.Node { return layout(pages.Append(ctx)) })
	app.Page("/clock", func(ctx *r.Context) *r.Node { return layout(pages.Clock(ctx)) })
	app.Page("/shared", func(ctx *r.Context) *r.Node { return layout(pages.Shared(ctx)) })
	app.Page("/reload-redirect", func(ctx *r.Context) *r.Node { return layout(pages.ReloadRedirect(ctx)) })
	app.Page("/routes", func(ctx *r.Context) *r.Node { return layout(pages.RoutesExample(ctx)) })
	app.Page("/login", func(ctx *r.Context) *r.Node { return layout(pages.LoginPage(ctx)) })
	app.Page("/skeleton", func(ctx *r.Context) *r.Node { return layout(pages.Skeleton(ctx)) })

	// Example page navigation actions (WS – replace content area)
	app.Action("nav.showcase", navTo("/", func() *r.Node { return pages.Showcase(nil) }))
	app.Action("nav.button", navTo("/button", func() *r.Node { return pages.Button(nil) }))
	app.Action("nav.text", navTo("/text", func() *r.Node { return pages.Text(nil) }))
	app.Action("nav.password", navTo("/password", func() *r.Node { return pages.Password(nil) }))
	app.Action("nav.number", navTo("/number", func() *r.Node { return pages.Number(nil) }))
	app.Action("nav.date", navTo("/date", func() *r.Node { return pages.Date(nil) }))
	app.Action("nav.area", navTo("/area", func() *r.Node { return pages.Area(nil) }))
	app.Action("nav.select", navTo("/select", func() *r.Node { return pages.SelectPage(nil) }))
	app.Action("nav.checkbox", navTo("/checkbox", func() *r.Node { return pages.Checkbox(nil) }))
	app.Action("nav.radio", navTo("/radio", func() *r.Node { return pages.Radio(nil) }))
	app.Action("nav.icons", navTo("/icons", func() *r.Node { return pages.Icons(nil) }))
	app.Action("nav.table", navTo("/table", func() *r.Node { return pages.TablePage(nil) }))
	app.Action("nav.form", navTo("/form", func() *r.Node { return pages.FormPage(nil) }))
	app.Action("nav.counter", navTo("/counter", func() *r.Node { return pages.Counter(nil) }))
	app.Action("nav.hello", navTo("/hello", func() *r.Node { return pages.Hello(nil) }))
	app.Action("nav.others", navTo("/others", func() *r.Node { return pages.Others(nil) }))
	app.Action("nav.append", navTo("/append", func() *r.Node { return pages.Append(nil) }))
	app.Action("nav.clock", navTo("/clock", func() *r.Node { return pages.Clock(nil) }))
	app.Action("nav.shared", navTo("/shared", func() *r.Node { return pages.Shared(nil) }))
	app.Action("nav.reload", navTo("/reload-redirect", func() *r.Node { return pages.ReloadRedirect(nil) }))
	app.Action("nav.routes", navTo("/routes", func() *r.Node { return pages.RoutesExample(nil) }))
	app.Action("nav.login", navTo("/login", func() *r.Node { return pages.LoginPage(nil) }))
	app.Action("nav.skeleton", navTo("/skeleton", func() *r.Node { return pages.Skeleton(nil) }))

	// Example page interactive actions (WS)
	app.Action("counter.inc", pages.HandleCounterInc)
	app.Action("counter.dec", pages.HandleCounterDec)
	app.Action("hello.ok", pages.HandleHelloOk)
	app.Action("hello.error", pages.HandleHelloError)
	app.Action("hello.delay", pages.HandleHelloDelay)
	app.Action("hello.crash", pages.HandleHelloCrash)
	app.Action("append.end", pages.HandleAppendEnd)
	app.Action("append.start", pages.HandleAppendStart)
	app.Action("form.submit", pages.HandleFormSubmit)
	app.Action("shared.submit", pages.HandleSharedSubmit)
	app.Action("shared.reset", pages.HandleSharedReset)
	app.Action("redirect.dashboard", pages.HandleRedirectDashboard)
	app.Action("redirect.invoices", pages.HandleRedirectInvoices)
	app.Action("redirect.button", pages.HandleRedirectButton)
	app.Action("routes.user", pages.HandleRoutesUser)
	app.Action("routes.userpost", pages.HandleRoutesUserPost)
	app.Action("routes.product", pages.HandleRoutesProduct)
	app.Action("routes.search", pages.HandleRoutesSearch)
	app.Action("login.submit", pages.HandleLoginSubmit)
	app.Action("select.change", pages.HandleSelectChange)
	app.Action("clock.start", pages.HandleClockStart)

	// Serve embedded static assets (favicon, images, etc.)
	app.Assets(assets, "assets", "/assets/")
	app.Favicon = "/assets/favicon.svg"
	app.Title = "g-sui Component Showcase"
	app.Description = "A server-rendered Go UI framework with live WebSocket updates, Tailwind CSS, and interactive components."

	app.Listen(":1423")
}

// ---------------------------------------------------------------------------
// Navigation helper
// ---------------------------------------------------------------------------

func navTo(url string, content func() *r.Node) r.ActionHandler {
	return func(ctx *r.Context) string {
		return r.NewResponse().
			Inner(pages.ContentID, content()).
			Navigate(url).
			Build()
	}
}

// ---------------------------------------------------------------------------
// Layout
// ---------------------------------------------------------------------------

func layout(content *r.Node) *r.Node {
	return r.Div("min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors").Render(
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
					navLink("Invoices", "nav.list"),
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
