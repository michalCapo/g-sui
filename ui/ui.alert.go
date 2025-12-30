package ui

import (
	"fmt"
	"strings"
)

type alert struct {
	message     string
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
		"relative flex items-center gap-3 p-4 rounded-lg border shadow-sm",
		If(a.class != "", func() string { return a.class }),
	)

	// Build the alert content
	content := Div(alertClasses, Attr{ID: alertID})(
		a.renderIcon(iconHTML, iconClasses),
		a.renderMessage(),
		a.renderDismissButton(alertID),
	)

	// Add JavaScript for dismiss functionality
	script := a.renderDismissScript(alertID)

	return content + script
}

// getVariantStyles returns the base classes, icon SVG, and icon classes for each variant
func (a *alert) getVariantStyles() (baseClasses, iconHTML, iconClasses string) {
	switch a.variant {
	case "success":
		return "bg-green-50 border-green-200 text-green-800 dark:bg-green-900 dark:border-green-600 dark:text-green-100",
			`<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/></svg>`,
			"text-green-600 dark:text-green-400"
	case "warning":
		return "bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-900 dark:border-yellow-600 dark:text-yellow-100",
			`<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/></svg>`,
			"text-yellow-600 dark:text-yellow-400"
	case "error":
		return "bg-red-50 border-red-200 text-red-800 dark:bg-red-900 dark:border-red-600 dark:text-red-100",
			`<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/></svg>`,
			"text-red-600 dark:text-red-400"
	default: // "info"
		return "bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900 dark:border-blue-600 dark:text-blue-100",
			`<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/></svg>`,
			"text-blue-600 dark:text-blue-400"
	}
}

// renderIcon generates the icon element
func (a *alert) renderIcon(iconHTML, iconClasses string) string {
	return Div("flex-shrink-0 " + iconClasses)(
		iconHTML,
	)
}

// renderMessage generates the message content
func (a *alert) renderMessage() string {
	return Div("flex-1 text-sm font-medium")(
		a.message,
	)
}

// renderDismissButton generates the dismiss button if dismissible
func (a *alert) renderDismissButton(alertID string) string {
	if !a.dismissible {
		return ""
	}

	closeIcon := `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"/></svg>`

	// Escape single quotes in alertID for JavaScript
	escapedID := strings.ReplaceAll(alertID, "'", "\\'")

	return fmt.Sprintf(
		`<button type="button" onclick="gSuiDismissAlert('%s', %s)" class="flex-shrink-0 p-1 rounded-md hover:bg-black/10 dark:hover:bg-white/10 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-transparent transition-colors" aria-label="Close alert">%s</button>`,
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
