package ui

import (
	"fmt"
	"strings"
)

type accordionItem struct {
	title   string
	content string
	open    bool
}

type accordion struct {
	items    []accordionItem
	multiple bool
	variant  string
	visible  bool
	class    string
	id       string
}

const (
	AccordionBordered = "bordered"
	AccordionGhost    = "ghost"
	AccordionSeparated = "separated"
)

// Accordion creates a new accordion component
func Accordion() *accordion {
	return &accordion{
		multiple: false,
		visible:  true,
		variant:  AccordionBordered,
		items:    make([]accordionItem, 0),
		id:       "acc_" + RandomString(8),
	}
}

// Item adds a new section to the accordion
func (a *accordion) Item(title, content string, open ...bool) *accordion {
	isOpen := false
	if len(open) > 0 {
		isOpen = open[0]
	}
	a.items = append(a.items, accordionItem{
		title:   title,
		content: content,
		open:    isOpen,
	})
	return a
}

// Multiple sets whether multiple sections can be open at once
func (a *accordion) Multiple(value bool) *accordion {
	a.multiple = value
	return a
}

// Variant sets the accordion style: "bordered", "ghost", or "separated"
func (a *accordion) Variant(value string) *accordion {
	a.variant = value
	return a
}

// If conditionally renders the accordion
func (a *accordion) If(value bool) *accordion {
	a.visible = value
	return a
}

// Class adds custom CSS classes to the accordion
func (a *accordion) Class(value ...string) *accordion {
	a.class = strings.Join(value, " ")
	return a
}

// Render generates the HTML for the accordion component
func (a *accordion) Render() string {
	if !a.visible || len(a.items) == 0 {
		return ""
	}

	// Generate unique IDs for each item
	itemIds := make([]string, len(a.items))
	contentIds := make([]string, len(a.items))
	for i := range a.items {
		itemIds[i] = fmt.Sprintf("%s_item_%d", a.id, i)
		contentIds[i] = fmt.Sprintf("%s_content_%d", a.id, i)
	}

	// Build the items HTML
	itemsHTML := make([]string, len(a.items))
	for i, item := range a.items {
		itemsHTML[i] = a.renderItem(itemIds[i], contentIds[i], item.title, item.content, i)
	}

	// Container classes
	containerClass := Classes(
		"accordion",
		"w-full",
		If(a.variant == AccordionBordered, func() string {
			return "border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden"
		}),
		If(a.variant == AccordionSeparated, func() string {
			return "flex flex-col gap-2"
		}),
		a.class,
	)

	return Div(containerClass, Attr{ID: a.id, DataAccordion: a.multipleValue()})(
		strings.Join(itemsHTML, ""),
	) + a.script(itemIds, contentIds)
}

// multipleValue returns the data attribute value for multiple mode
func (a *accordion) multipleValue() string {
	if a.multiple {
		return "multiple"
	}
	return "single"
}

// renderItem renders a single accordion item
func (a *accordion) renderItem(itemId, contentId, title, content string, index int) string {
	isSeparated := a.variant == AccordionSeparated
	isOpen := a.items[index].open

	headerClass := Classes(
		"accordion-header",
		"flex",
		"items-center",
		"justify-between",
		"w-full",
		"px-5",
		"py-4",
		"cursor-pointer",
		"select-none",
		"transition-all",
		"duration-200",
		"group",
		If(isSeparated, func() string {
			return "bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg shadow-sm"
		}),
		If(!isSeparated && a.variant != AccordionGhost, func() string {
			return "bg-white dark:bg-gray-900 hover:bg-gray-50/50 dark:hover:bg-gray-800/30"
		}),
		If(a.variant == AccordionGhost, func() string {
			return "hover:bg-gray-100/50 dark:hover:bg-gray-800/30 rounded-lg"
		}),
		If(!isSeparated && index > 0 && a.variant == AccordionBordered, func() string {
			return "border-t border-gray-100 dark:border-gray-800"
		}),
		If(isOpen, func() string {
			return "active-item"
		}),
	)

	iconClass := Classes(
		"accordion-icon",
		"transform",
		"transition-transform",
		"duration-300",
		Or(isOpen, func() string { return "rotate-180" }, func() string { return "rotate-0" }),
		"text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300",
	)

	contentClass := Classes(
		"accordion-content",
		If(isOpen, func() string { return "open" }),
		"overflow-hidden",
		"transition-all",
		"duration-300",
		"ease-in-out",
		"px-5",
		If(isSeparated, func() string {
			return "bg-white dark:bg-gray-900 border-x border-b border-gray-100 dark:border-gray-800 rounded-b-lg -mt-2 pt-2 shadow-sm"
		}),
		If(!isSeparated, func() string {
			return "bg-white dark:bg-gray-900"
		}),
		If(a.variant == AccordionGhost, func() string {
			return "bg-transparent"
		}),
	)

	maxHeight := "max-height: 0px;"
	if isOpen {
		maxHeight = "max-height: 1000px;" // Reasonable default for pre-rendered open state
	}

	return Div(If(isSeparated, func() string { return "mb-2" }))(
		Div(headerClass)(
			Div("font-bold text-gray-700 dark:text-gray-200 tracking-tight")(title),
			Div(iconClass)(
				`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>`,
			),
		),
		Div(contentClass, Attr{
			ID:    contentId,
			Style: maxHeight,
		})(
			Div("py-4 text-sm text-gray-600 dark:text-gray-400 leading-relaxed")(content),
		),
	)
}

// script generates the JavaScript for accordion functionality
func (a *accordion) script(itemIds, contentIds []string) string {
	js := fmt.Sprintf(`
		(function() {
			var accordionId = '%s';
			var multiple = %t;

			var accordion = document.getElementById(accordionId);
			if (!accordion) return;

			var headers = accordion.querySelectorAll('.accordion-header');
			var contents = accordion.querySelectorAll('.accordion-content');

			headers.forEach(function(header, index) {
				var content = contents[index];
				if (content.classList.contains('open')) {
					content.style.maxHeight = content.scrollHeight + 'px';
				}

				header.addEventListener('click', function(e) {
					e.preventDefault();
					var icon = header.querySelector('.accordion-icon');
					var isOpen = content.classList.contains('open');

					if (!multiple) {
						headers.forEach(function(h, i) {
							if (i !== index) {
								var c = contents[i];
								c.style.maxHeight = '0px';
								c.classList.remove('open');
								h.classList.remove('active-item');
								var hi = h.querySelector('.accordion-icon');
								if (hi) hi.classList.remove('rotate-180');
							}
						});
					}

					if (isOpen) {
						content.style.maxHeight = '0px';
						content.classList.remove('open');
						header.classList.remove('active-item');
						if (icon) icon.classList.remove('rotate-180');
					} else {
						content.classList.add('open');
						header.classList.add('active-item');
						content.style.maxHeight = content.scrollHeight + 'px';
						if (icon) icon.classList.add('rotate-180');
					}
				});
			});

			window.addEventListener('resize', function() {
				contents.forEach(function(content) {
					if (content.classList.contains('open')) {
						content.style.maxHeight = content.scrollHeight + 'px';
					}
				});
			});
		})();
	`, a.id, a.multiple)

	return Script(js)
}
