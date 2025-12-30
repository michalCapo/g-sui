package ui

import (
	"fmt"
	"strings"
)

type tooltip struct {
	content  string
	position string
	variant  string
	delay    int // in milliseconds
	visible  bool
	class    string
	target   Attr
}

// Tooltip creates a new tooltip component with default settings
func Tooltip() *tooltip {
	return &tooltip{
		position: "top",
		variant:  "dark",
		delay:    200,
		visible:  true,
		target:   Target(),
	}
}

// Content sets the tooltip text content
func (t *tooltip) Content(value string) *tooltip {
	t.content = value
	return t
}

// Position sets the tooltip position: "top", "bottom", "left", or "right"
func (t *tooltip) Position(value string) *tooltip {
	t.position = value
	return t
}

// Variant sets the tooltip appearance: "dark", "light", "blue", "green", "red", "yellow"
func (t *tooltip) Variant(value string) *tooltip {
	t.variant = value
	return t
}

// Delay sets the show/hide delay in milliseconds
func (t *tooltip) Delay(ms int) *tooltip {
	t.delay = ms
	return t
}

// If conditionally shows the tooltip based on the boolean value
func (t *tooltip) If(value bool) *tooltip {
	t.visible = value
	return t
}

// Class adds additional CSS classes to the tooltip
func (t *tooltip) Class(value ...string) *tooltip {
	t.class = strings.Join(value, " ")
	return t
}

// Render generates the HTML for the tooltip component wrapping the provided content
func (t *tooltip) Render(wrappedHTML string) string {
	if !t.visible || t.content == "" {
		return wrappedHTML
	}

	tooltipID := "tt_" + RandomString(8)

	// Build position-specific classes
	positionClasses, arrowClasses := t.getPositionClasses()

	// Build variant-specific classes
	variantClasses := t.getVariantClasses()

	// Build tooltip classes
	tooltipClasses := Classes(
		"absolute z-[100]",
		"px-2.5 py-1.5",
		"text-[11px] font-bold leading-none whitespace-nowrap",
		"rounded-md shadow-lg",
		"opacity-0 scale-95",
		"invisible",
		"transition-all duration-200 ease-out",
		"pointer-events-none",
		positionClasses,
		variantClasses,
		t.class,
	)

	// Build arrow
	arrow := t.renderArrow(arrowClasses)

	// Wrap content in a relative positioned container with inline-block
	wrapperClasses := Classes(
		"relative",
		"inline-block",
	)

	// Build tooltip HTML - using data attributes for JavaScript
	tooltipHTML := fmt.Sprintf(
		`<div id="%s" class="%s" data-tooltip-delay="%d">%s%s</div>`,
		escapeAttr(tooltipID),
		escapeAttr(tooltipClasses),
		t.delay,
		t.content,
		arrow,
	)

	// Combine wrapper, tooltip, and content
	wrapper := fmt.Sprintf(
		`<div class="%s" onmouseenter="gSuiShowTooltip('%s')" onmouseleave="gSuiHideTooltip('%s')">%s%s</div>`,
		escapeAttr(wrapperClasses),
		escapeJS(tooltipID),
		escapeJS(tooltipID),
		wrappedHTML,
		tooltipHTML,
	)

	// Add JavaScript for tooltip functionality
	script := t.renderTooltipScript(tooltipID)

	return wrapper + script
}

// getPositionClasses returns the position-specific classes for tooltip and arrow
func (t *tooltip) getPositionClasses() (tooltipClasses, arrowClasses string) {
	switch t.position {
	case "bottom":
		return "left-1/2 -translate-x-1/2 top-full mt-2.5",
			"absolute left-1/2 -translate-x-1/2 -top-1 w-2 h-2 rotate-45"
	case "left":
		return "right-full top-1/2 -translate-y-1/2 mr-2.5",
			"absolute right-0 top-1/2 -translate-y-1/2 translate-x-1 w-2 h-2 rotate-45"
	case "right":
		return "left-full top-1/2 -translate-y-1/2 ml-2.5",
			"absolute left-0 top-1/2 -translate-y-1/2 -translate-x-1 w-2 h-2 rotate-45"
	default: // "top"
		return "left-1/2 -translate-x-1/2 bottom-full mb-2.5",
			"absolute left-1/2 -translate-x-1/2 -bottom-1 w-2 h-2 rotate-45"
	}
}

// getVariantClasses returns the variant-specific styling classes
func (t *tooltip) getVariantClasses() string {
	switch t.variant {
	case "light":
		return "bg-white text-gray-800 border border-gray-100 dark:bg-gray-800 dark:text-gray-100 dark:border-gray-700 shadow-lg"
	case "blue":
		return "bg-blue-600 text-white shadow-lg shadow-blue-500/20"
	case "green":
		return "bg-green-600 text-white shadow-lg shadow-green-500/20"
	case "red":
		return "bg-red-600 text-white shadow-lg shadow-red-500/20"
	case "yellow":
		return "bg-yellow-500 text-gray-900 shadow-lg shadow-yellow-500/20"
	default: // "dark"
		return "bg-gray-900 text-white dark:bg-white dark:text-gray-900 shadow-lg"
	}
}

// renderArrow generates the arrow element
func (t *tooltip) renderArrow(arrowClasses string) string {
	arrowColor := t.getArrowColor()
	return fmt.Sprintf(`<div class="%s %s"></div>`, arrowClasses, arrowColor)
}

// getArrowColor returns the background color for the arrow to match the tooltip
func (t *tooltip) getArrowColor() string {
	switch t.variant {
	case "light":
		return "bg-white dark:bg-gray-800 border-l border-t border-gray-100 dark:border-gray-700"
	case "blue":
		return "bg-blue-600"
	case "green":
		return "bg-green-600"
	case "red":
		return "bg-red-600"
	case "yellow":
		return "bg-yellow-500"
	default: // "dark"
		return "bg-gray-900 dark:bg-white"
	}
}

// renderTooltipScript generates JavaScript for tooltip functionality
func (t *tooltip) renderTooltipScript(tooltipID string) string {
	return Script(`
		(function() {
			window.gSuiTooltipTimers = window.gSuiTooltipTimers || {};
			
			window.gSuiShowTooltip = window.gSuiShowTooltip || function(id) {
				var tt = document.getElementById(id);
				if (!tt) return;
				
				if (window.gSuiTooltipTimers[id]) {
					clearTimeout(window.gSuiTooltipTimers[id]);
				}
				
				var delay = parseInt(tt.getAttribute('data-tooltip-delay')) || 0;
				window.gSuiTooltipTimers[id] = setTimeout(function() {
					tt.classList.remove('opacity-0', 'invisible', 'scale-95');
					tt.classList.add('opacity-100', 'visible', 'scale-100');
				}, delay);
			};

			window.gSuiHideTooltip = window.gSuiHideTooltip || function(id) {
				var tt = document.getElementById(id);
				if (!tt) return;
				
				if (window.gSuiTooltipTimers[id]) {
					clearTimeout(window.gSuiTooltipTimers[id]);
				}
				
				tt.classList.remove('opacity-100', 'visible', 'scale-100');
				tt.classList.add('opacity-0', 'invisible', 'scale-95');
			};
		})();
	`)
}
