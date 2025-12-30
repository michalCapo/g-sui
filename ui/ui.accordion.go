package ui

import (
	"fmt"
	"strings"
)

type accordionItem struct {
	title   string
	content string
}

type accordion struct {
	items    []accordionItem
	multiple bool
	visible  bool
	class    string
	id       string
}

// Accordion creates a new accordion component
func Accordion() *accordion {
	return &accordion{
		multiple: false,
		visible:  true,
		items:    make([]accordionItem, 0),
		id:       "acc_" + RandomString(8),
	}
}

// Item adds a new section to the accordion
func (a *accordion) Item(title, content string) *accordion {
	a.items = append(a.items, accordionItem{
		title:   title,
		content: content,
	})
	return a
}

// Multiple sets whether multiple sections can be open at once
func (a *accordion) Multiple(value bool) *accordion {
	a.multiple = value
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
		"border",
		"border-gray-300",
		"dark:border-gray-600",
		"rounded-lg",
		"overflow-hidden",
		"bg-white",
		"dark:bg-gray-900",
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
	headerClass := Classes(
		"accordion-header",
		"flex",
		"items-center",
		"justify-between",
		"w-full",
		"px-4",
		"py-3",
		"cursor-pointer",
		"select-none",
		"transition-colors",
		"duration-200",
		"bg-gray-50",
		"hover:bg-gray-100",
		"dark:bg-gray-800",
		"dark:hover:bg-gray-700",
		If(index > 0, func() string {
			return "border-t border-gray-300 dark:border-gray-600"
		}),
	)

	iconClass := Classes(
		"accordion-icon",
		"transform",
		"transition-transform",
		"duration-300",
		"rotate-0",
	)

	contentClass := Classes(
		"accordion-content",
		"max-h-0",
		"overflow-hidden",
		"transition-all",
		"duration-300",
		"ease-in-out",
		"px-4",
		"bg-white",
		"dark:bg-gray-900",
	)

	return Div("")(
		Div(headerClass)(
			Div("font-medium text-gray-800 dark:text-gray-200")(title),
			Div(iconClass)(
				`<svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
				</svg>`,
			),
		),
		Div(contentClass, Attr{
			ID: contentId,
		})(
			Div("py-3")(content),
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
				header.addEventListener('click', function(e) {
					e.preventDefault();
					var content = contents[index];
					var icon = header.querySelector('.accordion-icon');

					var isOpen = content.classList.contains('open');

					if (!multiple) {
						// Close all other items
						contents.forEach(function(c, i) {
							if (i !== index) {
								c.style.maxHeight = '';
								c.classList.remove('open');
								var inner = c.querySelector('.py-3');
								if (inner) inner.classList.remove('py-3');
								var h = headers[i].querySelector('.accordion-icon');
								if (h) h.classList.remove('rotate-180');
							}
						});
					}

					// Toggle current item
					if (isOpen) {
						content.style.maxHeight = '';
						content.classList.remove('open');
						var inner = content.querySelector('.py-3');
						if (inner) inner.classList.remove('py-3');
						if (icon) icon.classList.remove('rotate-180');
					} else {
						content.classList.add('open');
						var inner = content.querySelector('.py-3');
						if (inner) inner.classList.add('py-3');
						content.style.maxHeight = content.scrollHeight + 'px';
						if (icon) icon.classList.add('rotate-180');
					}
				});
			});

			// Handle window resize to adjust heights
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
