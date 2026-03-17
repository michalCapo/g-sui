package pages

import r "github.com/michalCapo/g-sui/ui"

func Skeleton(ctx *r.Context) *r.Node {
	return r.Div("max-w-6xl mx-auto flex flex-col gap-8 w-full").Render(
		r.Div("text-3xl font-bold").Text("Skeleton Loaders"),
		r.Div("text-gray-600").Text("Loading placeholder components for various content types."),

		section("Table", r.SkeletonTable()),
		section("List", r.SkeletonList()),
		section("Cards", r.SkeletonCards()),
		section("Form", r.SkeletonForm()),
		section("Component", r.SkeletonComponent()),
		section("Page Layout", r.SkeletonPage()),
	)
}

func section(title string, content *r.Node) *r.Node {
	return r.Div("flex flex-col gap-3").Render(
		r.Div("text-sm font-bold text-gray-500 uppercase").Text(title),
		content,
	)
}
