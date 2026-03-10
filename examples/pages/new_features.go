package pages

import (
	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// NewUtilities demonstrates the core JS utilities added to g-sui:
// __debounce, __clipboard, __cfmt.relativeTime, and __cfmt.datePreset.
func NewUtilities(ctx *ui.Context) string {
	card := "bg-white dark:bg-gray-900 rounded-lg shadow p-6"
	presetBtn := "px-3 py-1.5 rounded bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-sm cursor-pointer"

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-8")(
		ui.Div("text-3xl font-bold")("JS Utilities"),
		ui.Div("text-gray-600")("Core JavaScript utilities: debounce, clipboard, relative time formatting, and date presets."),

		// --- Debounce ---
		ui.Div(card)(
			ui.Div("text-xl font-semibold mb-1")("Debounce"),
			ui.Div("text-gray-500 text-sm mb-4")("Type in the input below. The output updates only after 500ms of inactivity (debounced)."),
			ui.Input("bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-80",
				ui.Attr{ID: "debounce-input", Type: "text", Placeholder: "Type something..."},
			),
			ui.Div("mt-3 text-sm text-gray-700 dark:text-gray-300")(
				"Output: ", ui.Span("font-mono text-blue-600 dark:text-blue-400", ui.Attr{ID: "debounce-output"})("—"),
			),
			ui.Div("mt-2 text-xs text-gray-400")(
				"Trigger count: ", ui.Span("font-mono", ui.Attr{ID: "debounce-count"})("0"),
			),
			js.Script(`
				var count = 0;
				var inp = document.getElementById("debounce-input");
				var out = document.getElementById("debounce-output");
				var cnt = document.getElementById("debounce-count");
				var update = __debounce(function(val) {
					count++;
					out.textContent = val || "\u2014";
					cnt.textContent = String(count);
				}, 500);
				inp.addEventListener("input", function(){ update(inp.value); });
			`),
		),

		// --- Clipboard ---
		ui.Div(card)(
			ui.Div("text-xl font-semibold mb-1")("Clipboard"),
			ui.Div("text-gray-500 text-sm mb-4")("Click any button to copy its text to your clipboard. A toast notification will confirm the copy."),
			ui.Div("flex flex-wrap gap-3")(
				ui.Button().Color(ui.Blue).Size(ui.SM).Click("__clipboard('Hello from g-sui!')").Render(`Copy "Hello from g-sui!"`),
				ui.Button().Color(ui.Green).Size(ui.SM).Click("__clipboard('npm install g-sui')").Render(`Copy "npm install g-sui"`),
				ui.Button().Color(ui.Purple).Size(ui.SM).Click("__clipboard(document.getElementById('copy-target').textContent)").Render("Copy text below"),
			),
			ui.Pre("mt-3 p-3 bg-gray-100 dark:bg-gray-800 rounded text-sm font-mono text-gray-700 dark:text-gray-300", ui.Attr{ID: "copy-target"})(
				"func main() {\n    app := ui.MakeApp(\"en\")\n    app.Listen(\":8080\")\n}",
			),
		),

		// --- Relative Time ---
		ui.Div(card)(
			ui.Div("text-xl font-semibold mb-1")("Relative Time Formatting"),
			ui.Div("text-gray-500 text-sm mb-4")("__cfmt.relativeTime() converts dates to human-readable relative strings."),
			ui.Div("grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm", ui.Attr{ID: "relative-time-demo"})(),
			js.Script(`
				var el = document.getElementById("relative-time-demo");
				var now = new Date();
				var tests = [
					{ label: "Just now", date: new Date(now - 5000) },
					{ label: "30 seconds ago", date: new Date(now - 30000) },
					{ label: "5 minutes ago", date: new Date(now - 5*60*1000) },
					{ label: "2 hours ago", date: new Date(now - 2*60*60*1000) },
					{ label: "Yesterday", date: new Date(now - 24*60*60*1000) },
					{ label: "3 days ago", date: new Date(now - 3*24*60*60*1000) },
					{ label: "Last week", date: new Date(now - 7*24*60*60*1000) },
					{ label: "2 weeks ago", date: new Date(now - 14*24*60*60*1000) },
				];
				for (var i = 0; i < tests.length; i++) {
					var row = document.createElement("div");
					row.className = "flex justify-between py-1.5 px-3 bg-gray-50 dark:bg-gray-800 rounded";
					row.innerHTML = '<span class="text-gray-500 dark:text-gray-400">' + tests[i].label + '</span>' +
						'<span class="font-mono text-blue-600 dark:text-blue-400">' + __cfmt.relativeTime(tests[i].date) + '</span>';
					el.appendChild(row);
				}
			`),
		),

		// --- Date Presets ---
		ui.Div(card)(
			ui.Div("text-xl font-semibold mb-1")("Date Presets"),
			ui.Div("text-gray-500 text-sm mb-4")("__cfmt.datePreset(name) returns {from, to} date ranges for common time periods."),
			ui.Div("flex flex-wrap gap-2 mb-4")(
				ui.Button().Class("rounded", presetBtn).Click("showPreset('today')").Render("Today"),
				ui.Button().Class("rounded", presetBtn).Click("showPreset('thisWeek')").Render("This Week"),
				ui.Button().Class("rounded", presetBtn).Click("showPreset('thisMonth')").Render("This Month"),
				ui.Button().Class("rounded", presetBtn).Click("showPreset('thisQuarter')").Render("This Quarter"),
				ui.Button().Class("rounded", presetBtn).Click("showPreset('thisYear')").Render("This Year"),
				ui.Button().Class("rounded", presetBtn).Click("showPreset('lastMonth')").Render("Last Month"),
				ui.Button().Class("rounded", presetBtn).Click("showPreset('lastYear')").Render("Last Year"),
			),
			ui.Div("p-3 bg-gray-50 dark:bg-gray-800 rounded text-sm font-mono text-gray-700 dark:text-gray-300", ui.Attr{ID: "date-preset-output"})(
				"Click a preset to see the date range",
			),
			js.Script(`
				window.showPreset = function(name) {
					var result = __cfmt.datePreset(name);
					document.getElementById("date-preset-output").innerHTML =
						'<span class="text-blue-600 dark:text-blue-400">' + name + '</span>: ' +
						'from <strong>' + result.from + '</strong> to <strong>' + result.to + '</strong>';
				};
			`),
		),
	)
}
