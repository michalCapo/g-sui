package ui

import (
	"fmt"
	"strings"
)

type dropdownItem struct {
	label     string
	onclick   string
	icon      string
	variant   string // "default", "danger"
	isDivider bool
	isHeader  bool
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
func (d *dropdown) Item(label string, onclick string, icon ...string) *dropdown {
	item := dropdownItem{
		label:     label,
		onclick:   onclick,
		variant:   "default",
		isDivider: false,
	}
	if len(icon) > 0 {
		item.icon = icon[0]
	}
	d.items = append(d.items, item)
	return d
}

// Danger adds a danger-variant menu item
func (d *dropdown) Danger(label string, onclick string, icon ...string) *dropdown {
	item := dropdownItem{
		label:     label,
		onclick:   onclick,
		variant:   "danger",
		isDivider: false,
	}
	if len(icon) > 0 {
		item.icon = icon[0]
	}
	d.items = append(d.items, item)
	return d
}

// Header adds a non-interactive header label to the menu
func (d *dropdown) Header(label string) *dropdown {
	d.items = append(d.items, dropdownItem{
		label:    label,
		isHeader: true,
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
		"border-gray-200",
		"dark:border-gray-800",
		"rounded-xl",
		"shadow-xl",
		"py-1.5",
		"hidden",
		"opacity-0 scale-95 origin-top",
		"transition-all duration-200 ease-out",
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
		`<div id="%s" class="relative inline-block">%s%s</div>`,
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
		return "right-0 top-full mt-2 origin-top-right"
	case "top-left":
		return "left-0 bottom-full mb-2 origin-bottom-left"
	case "top-right":
		return "right-0 bottom-full mb-2 origin-bottom-right"
	default: // "bottom-left"
		return "left-0 top-full mt-2 origin-top-left"
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
		} else if item.isHeader {
			builder.WriteString(d.renderHeader(item.label))
		} else {
			builder.WriteString(d.renderItem(item.label, item.onclick, item.icon, item.variant))
		}
	}

	return builder.String()
}

// renderHeader generates the HTML for a menu header
func (d *dropdown) renderHeader(label string) string {
	return fmt.Sprintf(
		`<div class="px-4 py-1.5 text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">%s</div>`,
		escapeAttr(label),
	)
}

// renderItem generates the HTML for a single menu item
func (d *dropdown) renderItem(label, onclick, icon, variant string) string {
	itemClass := Classes(
		"flex",
		"items-center",
		"gap-2",
		"w-full",
		"text-left",
		"px-3",
		"py-2",
		"mx-1",
		"w-[calc(100%-0.5rem)]",
		"text-sm",
		"font-bold",
		"cursor-pointer",
		"rounded-md",
		"transition-all",
		"duration-150",
		"whitespace-nowrap",
	)

	if variant == "danger" {
		itemClass = Classes(itemClass, "text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950/30")
	} else {
		itemClass = Classes(itemClass, "text-gray-700 dark:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800")
	}

	iconHTML := `<span class="w-5 h-5 flex-shrink-0 flex items-center justify-center opacity-70">`
	if icon != "" {
		iconHTML += icon
	}
	iconHTML += `</span>`

	return fmt.Sprintf(
		`<button class="%s" onclick="%s">%s<span class="flex-1">%s</span></button>`,
		escapeAttr(itemClass),
		escapeAttr(onclick),
		iconHTML,
		escapeAttr(label),
	)
}

// renderDivider generates the HTML for a divider line
func (d *dropdown) renderDivider() string {
	return `<div class="border-t border-gray-100 dark:border-gray-800 my-1.5 mx-2"></div>`
}

// renderScript generates JavaScript for dropdown toggle and click-outside-to-close
func (d *dropdown) renderScript(dropdownID, triggerID string) string {
	scriptJS := fmt.Sprintf(
		`(function(){ 
			var t=document.getElementById('%s'); if(!t) return; 
			var d=document.getElementById('%s'); if(!d) return; 
			var o=false; 
			
			function show(){
				o=true;
				d.classList.remove('hidden');
				setTimeout(function(){
					d.classList.remove('opacity-0', 'scale-95');
					d.classList.add('opacity-100', 'scale-100');
				}, 10);
			}
			
			function hide(){
				o=false;
				d.classList.remove('opacity-100', 'scale-100');
				d.classList.add('opacity-0', 'scale-95');
				setTimeout(function(){
					if(!o) d.classList.add('hidden');
				}, 200);
			}
			
			t.addEventListener('click',function(e){ 
				e.stopPropagation(); 
				if(o) hide(); else show();
			});
			
			document.addEventListener('click',function(){ if(o) hide(); });
			document.addEventListener('keydown', function(e){ if(e.key==='Escape' && o) hide(); });
		})();`,
		escapeJS(triggerID),
		escapeJS(dropdownID),
	)

	return Script(scriptJS)
}
