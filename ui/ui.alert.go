package ui

import (
	"fmt"
	"strings"
)

type alert struct {
	message     string
	title       string
	variant     string
	dismissible bool
	persistKey  string
	visible     bool
	class       string
	target      Attr
}

// Alert creates a new alert component with default settings
func Alert() *alert {
	return &alert{
		variant:     "info",
		dismissible: false,
		visible:     true,
		target:      Target(),
	}
}

// Message sets the alert message content
func (a *alert) Message(value string) *alert {
	a.message = value
	return a
}

// Title sets the alert title
func (a *alert) Title(value string) *alert {
	a.title = value
	return a
}

// Variant sets the alert type: "info", "success", "warning", or "error"
func (a *alert) Variant(value string) *alert {
	a.variant = value
	return a
}

// Dismissible sets whether the alert can be dismissed by the user
func (a *alert) Dismissible(value bool) *alert {
	a.dismissible = value
	return a
}

// Persist sets a localStorage key to remember dismissal ("don't show again")
func (a *alert) Persist(key string) *alert {
	a.persistKey = key
	return a
}

// If conditionally shows the alert based on the boolean value
func (a *alert) If(value bool) *alert {
	a.visible = value
	return a
}

// Class adds additional CSS classes to the alert
func (a *alert) Class(value ...string) *alert {
	a.class = strings.Join(value, " ")
	return a
}

// Render generates the HTML for the alert component
func (a *alert) Render() string {
	if !a.visible || a.message == "" {
		return ""
	}

	// Generate unique IDs for this alert instance
	alertID := a.target.ID
	if alertID == "" {
		alertID = "alert_" + RandomString(8)
	}

	// Get variant-specific styling
	baseClasses, iconHTML, iconClasses := a.getVariantStyles()

	// Build the main alert classes
	alertClasses := Classes(
		baseClasses,
		"relative flex items-start gap-3 p-4 rounded-lg border shadow-sm",
		If(a.class != "", func() string { return a.class }),
	)

	// Build the alert content
	content := Div(alertClasses, Attr{ID: alertID})(
		a.renderIcon(iconHTML, iconClasses),
		Div("flex-1 min-w-0")(
			a.renderTitle(),
			a.renderMessage(),
		),
		a.renderDismissButton(alertID),
	)

	// Add JavaScript for dismiss functionality
	script := a.renderDismissScript(alertID)

	return content + script
}

// getVariantStyles returns the base classes, icon SVG, and icon classes for each variant
func (a *alert) getVariantStyles() (baseClasses, iconHTML, iconClasses string) {
	isOutline := strings.HasSuffix(a.variant, "-outline")
	variantName := strings.TrimSuffix(a.variant, "-outline")

	switch variantName {
	case "success":
		if isOutline {
			return "bg-white border-green-500 text-green-700 dark:bg-gray-950 dark:border-green-500 dark:text-green-400",
				Icon("check_circle", Attr{Class: "text-lg"}),
				"text-green-500"
		}
		return "bg-green-50 border-green-200 text-green-800 dark:bg-green-950/40 dark:border-green-900/50 dark:text-green-100",
			Icon("check_circle", Attr{Class: "text-lg"}),
			"text-green-600 dark:text-green-400"
	case "warning":
		if isOutline {
			return "bg-white border-yellow-500 text-yellow-700 dark:bg-gray-950 dark:border-yellow-500 dark:text-yellow-400",
				Icon("warning", Attr{Class: "text-lg"}),
				"text-yellow-500"
		}
		return "bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-950/40 dark:border-yellow-900/50 dark:text-yellow-100",
			Icon("warning", Attr{Class: "text-lg"}),
			"text-yellow-600 dark:text-yellow-400"
	case "error":
		if isOutline {
			return "bg-white border-red-500 text-red-700 dark:bg-gray-950 dark:border-red-500 dark:text-red-400",
				Icon("error", Attr{Class: "text-lg"}),
				"text-red-500"
		}
		return "bg-red-50 border-red-200 text-red-800 dark:bg-red-950/40 dark:border-red-900/50 dark:text-red-100",
			Icon("error", Attr{Class: "text-lg"}),
			"text-red-600 dark:text-red-400"
	default: // "info"
		if isOutline {
			return "bg-white border-blue-500 text-blue-700 dark:bg-gray-950 dark:border-blue-500 dark:text-blue-400",
				Icon("info", Attr{Class: "text-lg"}),
				"text-blue-500"
		}
		return "bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-950/40 dark:border-blue-900/50 dark:text-blue-100",
			Icon("info", Attr{Class: "text-lg"}),
			"text-blue-600 dark:text-blue-400"
	}
}

// renderIcon generates the icon element
func (a *alert) renderIcon(iconHTML, iconClasses string) string {
	return Div("flex-shrink-0 mt-0.5 " + iconClasses)(
		iconHTML,
	)
}

// renderTitle generates the title content
func (a *alert) renderTitle() string {
	if a.title == "" {
		return ""
	}
	return Div("text-sm font-bold mb-1")(
		a.title,
	)
}

// renderMessage generates the message content
func (a *alert) renderMessage() string {
	textClass := "text-sm"
	if a.title != "" {
		textClass = "text-xs opacity-90"
	}
	return Div(textClass)(
		a.message,
	)
}

// renderDismissButton generates the dismiss button if dismissible
func (a *alert) renderDismissButton(alertID string) string {
	if !a.dismissible {
		return ""
	}

	closeIcon := Icon("close", Attr{Class: "text-base"})

	// Escape single quotes in alertID for JavaScript
	escapedID := strings.ReplaceAll(alertID, "'", "\\'")

	return fmt.Sprintf(
		`<button type="button" onclick="gSuiDismissAlert('%s', %s)" class="flex-shrink-0 ml-auto -mr-1 p-1 rounded-md opacity-50 hover:opacity-100 hover:bg-black/5 dark:hover:bg-white/5 focus:outline-none transition-all" aria-label="Close alert">%s</button>`,
		escapedID,
		Or(a.persistKey != "", func() string { return "'" + escapeJS(a.persistKey) + "'" }, func() string { return "null" }),
		closeIcon,
	)
}

// renderDismissScript generates the JavaScript for dismiss functionality
func (a *alert) renderDismissScript(alertID string) string {
	if !a.dismissible {
		return ""
	}

	var persistCheck string
	if a.persistKey != "" {
		// Add localStorage persistence check
		persistCheck = fmt.Sprintf(
			`try { if (localStorage.getItem('%s') === 'dismissed') { document.getElementById('%s').remove(); return; } } catch (_) {}`,
			escapeJS(a.persistKey),
			escapeJS(alertID),
		)
	}

	persistAction := ""
	if a.persistKey != "" {
		persistAction = fmt.Sprintf(
			`try { localStorage.setItem('%s', 'dismissed'); } catch (_) {}`,
			escapeJS(a.persistKey),
		)
	}

	scriptJS := fmt.Sprintf(
		`(function(){ var el=document.getElementById('%s'); if(!el) return; %s window.gSuiDismissAlert=function(id,persist){ var e=document.getElementById(id); if(e){ e.remove(); %s } }; })();`,
		escapeJS(alertID),
		persistCheck,
		persistAction,
	)

	return Script(scriptJS)
}
