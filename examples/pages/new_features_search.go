package pages

import (
	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// NewSearch demonstrates the search components: LiveSearch, ContentSearch, and Autocomplete.
func NewSearch(ctx *ui.Context) string {
	// Sample items for LiveSearch
	items := []struct{ Name, Role, Dept string }{
		{"Alice Johnson", "Engineer", "Engineering"},
		{"Bob Smith", "Designer", "Design"},
		{"Carol Williams", "Manager", "Engineering"},
		{"David Brown", "Analyst", "Finance"},
		{"Eva Martinez", "Developer", "Engineering"},
		{"Frank Davis", "Designer", "Design"},
		{"Grace Wilson", "Director", "Operations"},
		{"Henry Taylor", "Intern", "Engineering"},
		{"Iris Anderson", "Accountant", "Finance"},
		{"Jack Thomas", "Developer", "Engineering"},
	}

	var itemsHTML string
	for _, item := range items {
		itemsHTML += ui.Div("list-item flex items-center justify-between py-2 px-3 border-b border-gray-100 dark:border-gray-800")(
			ui.Div("")(
				ui.Div("font-medium text-gray-800 dark:text-gray-200")(item.Name),
				ui.Div("text-sm text-gray-500 dark:text-gray-400")(item.Role+" — "+item.Dept),
			),
		)
	}

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-8")(
		ui.Div("text-3xl font-bold")("Search Components"),
		ui.Div("text-gray-600")("LiveSearch for filtering lists, ContentSearch for in-page text search, and Autocomplete with datalist."),

		// --- LiveSearch ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("LiveSearch"),
			ui.Div("text-gray-500 text-sm mb-4")("Filters server-rendered HTML elements by text content. Type to filter the list below."),
			js.LiveSearch(".list-item", "live-search-input", "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-80 mb-4"),
			ui.Div("bg-gray-50 dark:bg-gray-800 rounded-lg")(itemsHTML),
		),

		// --- ContentSearch ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Content Search"),
			ui.Div("text-gray-500 text-sm mb-4")("Press / to open the in-page search bar. Matches are highlighted. Navigate with Enter/Shift+Enter. Close with Escape."),
			ui.Div("prose dark:prose-invert max-w-none", ui.Attr{ID: "search-content"})(
				ui.Div("text-gray-700 dark:text-gray-300 leading-relaxed space-y-4")(
					`<p>The g-sui framework provides a comprehensive set of server-rendered UI components written in Go. Unlike traditional JavaScript frameworks, g-sui renders all HTML on the server and uses WebSocket connections for real-time updates.</p>`,
					`<p>Key features include automatic dark mode support, responsive layouts with Tailwind CSS, and a powerful client-side rendering engine for data-heavy components like tables and charts. The framework handles form validation, file uploads, and complex data grids out of the box.</p>`,
					`<p>The architecture follows a simple pattern: Go functions return HTML strings. Components like buttons, inputs, tables, and cards are all built using this composable approach. The result is type-safe, fast, and requires zero JavaScript knowledge from the developer.</p>`,
					`<p>For interactive features, g-sui provides WebSocket-based actions (ctx.Call, ctx.Submit) and a client-side component system (js.Client) that can render tables, charts, and custom components from JSON API data.</p>`,
				),
			),
			js.ContentSearch("#search-content", "/"),
		),

		// --- Autocomplete ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Autocomplete"),
			ui.Div("text-gray-500 text-sm mb-4")("Browser-native autocomplete using datalist. Start typing a city name."),
			ui.Div("flex flex-col gap-2")(
				ui.Div("text-sm text-gray-600 dark:text-gray-400")("City:"),
				js.Autocomplete("city-input", "/api/new/autocomplete-cities", "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-80"),
			),
		),
	)
}
