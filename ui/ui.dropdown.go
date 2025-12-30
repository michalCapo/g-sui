package ui

import (
	"fmt"
	"strings"
)

type dropdownItem struct {
	label     string
	onclick   string
	isDivider bool
}

type dropdown struct {
	trigger  string
	items    []dropdownItem
	position string
	visible  bool
	class    string
	target   Attr
}

// Dropdown creates a new dropdown component with default settings
func Dropdown() *dropdown {
	return &dropdown{
		position: "bottom-left",
		visible:  true,
		target:   Target(),
		items:    make([]dropdownItem, 0),
	}
}

// Trigger sets the HTML element that opens the dropdown
func (d *dropdown) Trigger(html string) *dropdown {
	d.trigger = html
	return d
}

// Item adds a menu item with the given label and onclick handler
func (d *dropdown) Item(label string, onclick string) *dropdown {
	d.items = append(d.items, dropdownItem{
		label:     label,
		onclick:   onclick,
		isDivider: false,
	})
	return d
}

// Divider adds a visual separator between menu item groups
func (d *dropdown) Divider() *dropdown {
	d.items = append(d.items, dropdownItem{
		isDivider: true,
	})
	return d
}

// Position sets the dropdown position: "bottom-left", "bottom-right", "top-left", or "top-right"
func (d *dropdown) Position(value string) *dropdown {
	d.position = value
	return d
}

// If conditionally shows the dropdown based on the boolean value
func (d *dropdown) If(value bool) *dropdown {
	d.visible = value
	return d
}

// Class adds additional CSS classes to the dropdown menu
func (d *dropdown) Class(value ...string) *dropdown {
	d.class = strings.Join(value, " ")
	return d
}

// Render generates the HTML for the dropdown component
func (d *dropdown) Render() string {
	if !d.visible || d.trigger == "" {
		return ""
	}

	dropdownID := "dropdown_" + d.target.ID
	triggerID := "dropdown_trigger_" + d.target.ID

	// Build position-specific classes
	positionClasses := d.getPositionClasses()

	// Build dropdown menu classes
	menuClasses := Classes(
		"absolute z-50",
		"min-w-[12rem]",
		"bg-white",
		"dark:bg-gray-900",
		"border",
		"border-gray-300",
		"dark:border-gray-600",
		"rounded-lg",
		"shadow-lg",
		"py-1",
		"hidden",
		positionClasses,
		d.class,
	)

	// Build menu items HTML
	itemsHTML := d.renderItems()

	// Build the dropdown menu HTML
	menuHTML := fmt.Sprintf(
		`<div id="%s" class="%s">%s</div>`,
		escapeAttr(dropdownID),
		escapeAttr(menuClasses),
		itemsHTML,
	)

	// Make trigger interactive by wrapping it with onclick
	triggerWrapper := fmt.Sprintf(
		`<div id="%s" class="relative inline-block" data-dropdown-open="false">%s%s</div>`,
		escapeAttr(triggerID),
		d.trigger,
		menuHTML,
	)

	// Add JavaScript for toggle and click-outside-to-close functionality
	script := d.renderScript(dropdownID, triggerID)

	return triggerWrapper + script
}

// getPositionClasses returns the position-specific classes for the dropdown menu
func (d *dropdown) getPositionClasses() string {
	switch d.position {
	case "bottom-right":
		return "left-full mt-1"
	case "top-left":
		return "right-0 bottom-full mb-1"
	case "top-right":
		return "left-0 bottom-full mb-1"
	default: // "bottom-left"
		return "left-0 mt-1"
	}
}

// renderItems generates the HTML for all menu items
func (d *dropdown) renderItems() string {
	if len(d.items) == 0 {
		return ""
	}

	var builder strings.Builder

	for _, item := range d.items {
		if item.isDivider {
			builder.WriteString(d.renderDivider())
		} else {
			builder.WriteString(d.renderItem(item.label, item.onclick))
		}
	}

	return builder.String()
}

// renderItem generates the HTML for a single menu item
func (d *dropdown) renderItem(label, onclick string) string {
	itemClass := Classes(
		"block",
		"w-full",
		"text-left",
		"px-4",
		"py-2",
		"text-sm",
		"cursor-pointer",
		"hover:bg-gray-100",
		"dark:hover:bg-gray-800",
		"text-gray-800",
		"dark:text-gray-200",
		"transition-colors",
		"duration-150",
	)

	return fmt.Sprintf(
		`<button class="%s" onclick="%s">%s</button>`,
		escapeAttr(itemClass),
		escapeAttr(onclick),
		escapeAttr(label),
	)
}

// renderDivider generates the HTML for a divider line
func (d *dropdown) renderDivider() string {
	return `<div class="border-t border-gray-200 dark:border-gray-700 my-1 mx-2"></div>`
}

// renderScript generates JavaScript for dropdown toggle and click-outside-to-close
func (d *dropdown) renderScript(dropdownID, triggerID string) string {
	scriptJS := fmt.Sprintf(
		`(function(){ var t=document.getElementById('%s'); if(!t) return; var d=document.getElementById('%s'); if(!d) return; var o=false; t.addEventListener('click',function(e){ e.stopPropagation(); o=!o; if(o){ d.classList.remove('hidden'); }else{ d.classList.add('hidden'); } }); d.addEventListener('click',function(e){ e.stopPropagation(); }); document.addEventListener('click',function(){ if(o){ o=false; d.classList.add('hidden'); } }); var p='%s'; window.addEventListener('resize',function(){ if(o){ var r=d.getBoundingClientRect(); var w=window.innerWidth; var h=window.innerHeight; if(r.left<0){ d.style.left='0px'; d.style.right='auto'; } if(r.right>w){ d.style.left='auto'; d.style.right='0px'; } if(r.bottom>h){ d.classList.remove('mt-1'); d.classList.add('bottom-full'); d.classList.add('mb-1'); } } }); })();`,
		escapeJS(triggerID),
		escapeJS(dropdownID),
		escapeJS(d.position),
	)

	return Script(scriptJS)
}
