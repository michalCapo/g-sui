package ui

import "fmt"

// Additional skeleton types for client zones.
const (
	SkeletonTable Skeleton = "table"
	SkeletonCards Skeleton = "cards"
)

// ---------------------------------------------------------------------------
// Skeleton renderers for SkeletonTable and SkeletonCards
// ---------------------------------------------------------------------------

// SkeletonTableBlock renders a table-shaped skeleton with default attributes.
func SkeletonTableBlock() string { return Attr{}.SkeletonTable() }

// SkeletonTable renders a table-shaped skeleton placeholder.
func (a Attr) SkeletonTable() string {
	headerCells := ""
	for i := 0; i < 4; i++ {
		headerCells += `<th class="p-3"><div class="bg-gray-200 dark:bg-gray-700 h-4 rounded w-20"></div></th>`
	}
	rows := ""
	for r := 0; r < 5; r++ {
		cells := ""
		for c := 0; c < 4; c++ {
			w := "w-24"
			if c == 0 {
				w = "w-32"
			}
			if c == 2 {
				w = "w-16"
			}
			cells += fmt.Sprintf(`<td class="p-3"><div class="bg-gray-200 dark:bg-gray-700 h-4 rounded %s"></div></td>`, w)
		}
		rows += `<tr class="border-t border-gray-100 dark:border-gray-800">` + cells + `</tr>`
	}
	return Div("animate-pulse", a)(
		`<div class="bg-white dark:bg-gray-900 rounded-lg shadow overflow-hidden">` +
			`<table class="w-full"><thead><tr class="border-b border-gray-200 dark:border-gray-700">` +
			headerCells + `</tr></thead><tbody>` + rows + `</tbody></table></div>`,
	)
}

// SkeletonCardsBlock renders a card grid skeleton with default attributes.
func SkeletonCardsBlock() string { return Attr{}.SkeletonCards() }

// SkeletonCards renders a card grid skeleton placeholder.
func (a Attr) SkeletonCards() string {
	cards := ""
	for i := 0; i < 6; i++ {
		cards += `<div class="bg-white dark:bg-gray-900 rounded-lg p-4 shadow">` +
			`<div class="bg-gray-200 dark:bg-gray-700 h-5 rounded w-3/4 mb-3"></div>` +
			`<div class="bg-gray-200 dark:bg-gray-700 h-4 rounded w-1/2 mb-2"></div>` +
			`<div class="bg-gray-200 dark:bg-gray-700 h-4 rounded w-2/3"></div></div>`
	}
	return Div("animate-pulse", a)(
		`<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">` + cards + `</div>`,
	)
}
