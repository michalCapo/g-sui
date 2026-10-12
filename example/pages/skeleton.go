package pages

import (
	"time"

	r "github.com/michalCapo/g-sui/ui"
)

// themed is the app's skeleton look. Copy it to change a field for one call.
var themed = r.SkeletonStyle{
	BarClass:     "bg-zinc-200 dark:bg-zinc-800",
	RoundClass:   "rounded-md",
	CardClass:    "bg-zinc-50 dark:bg-zinc-900 rounded-xl border border-zinc-200 dark:border-zinc-800",
	DividerClass: "border-zinc-200 dark:border-zinc-800",
	NoAvatar:     true,
}

func Skeleton(_ *r.Context) *r.Node {
	short := themed
	short.Rows = 3

	return r.Div("max-w-6xl mx-auto flex flex-col gap-8 w-full").Render(
		r.Div("text-3xl font-bold").Text("Skeleton Loaders"),
		r.Div("text-gray-600").Text("Loading placeholder components for various content types."),

		r.Div("flex flex-col gap-2 rounded-lg border border-blue-200 dark:border-blue-900 p-4").Render(
			r.Div("font-semibold").Text("Navigation skeleton"),
			r.Div("text-sm text-gray-600 dark:text-gray-400").Text("This page takes 1.5 seconds to render. The app registers a themed skeleton for it with app.Skeleton, so the skeleton replaces the content while it loads. Back and forward show it too."),
			r.NavLink("/skeleton/slow", "text-blue-600 dark:text-blue-400 underline w-fit").Text("Open the slow page"),
		),

		section("Table", r.SkeletonTable()),
		section("List", r.SkeletonList()),
		section("Cards", r.SkeletonCards()),
		section("Form", r.SkeletonForm()),
		section("Component", r.SkeletonComponent()),
		section("Page Layout", r.SkeletonPage()),

		section("Themed List (SkeletonStyle, 3 rows, no avatar)", short.List()),
		section("Themed Cards", short.Cards()),
		section("Custom (SkeletonBlock + SkeletonBar)", r.SkeletonBlock("flex items-center gap-3",
			themed.Bar("h-12 w-12 rounded-full"),
			r.Div("flex-1 flex flex-col gap-2").Render(themed.Bar("h-4 w-1/3"), themed.Bar("h-3 w-2/3")),
		)),
	)
}

func slowPage(_ *r.Context) *r.Node {
	time.Sleep(1500 * time.Millisecond)
	return r.Div("max-w-3xl mx-auto flex flex-col gap-4").Render(
		r.H1("text-3xl font-bold").Text("Slow page"),
		r.P("text-gray-600 dark:text-gray-400").Text("The server took 1.5 seconds. Go back, then forward, to see the skeleton again."),
		r.NavLink("/skeleton", "text-blue-600 dark:text-blue-400 underline w-fit").Text("Back to skeletons"),
	)
}

func section(title string, content *r.Node) *r.Node {
	return r.Div("flex flex-col gap-3").Render(
		r.Div("text-sm font-bold text-gray-500 uppercase").Text(title),
		content,
	)
}

func RegisterSkeleton(app *r.App) {
	app.Skeleton("/skeleton/slow", themed.Component())
	app.Page("/skeleton", Skeleton)
	app.Page("/skeleton/slow", slowPage)
}
