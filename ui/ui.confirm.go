package ui

import "fmt"

// ConfirmDialog renders a server-rendered modal overlay with confirm/cancel buttons.
// confirmAction is the POST URL for the confirm button.
// cancelURL is the navigation URL for the cancel button (or empty to just close).
func ConfirmDialog(title, message, confirmAction, cancelURL, class string) string {
	if class == "" {
		class = "bg-white dark:bg-gray-900 rounded-xl shadow-2xl p-6 max-w-md w-full"
	}

	cancelBtn := ""
	if cancelURL != "" {
		cancelBtn = fmt.Sprintf(`<a href="%s" class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-100 dark:hover:bg-gray-800">Cancel</a>`, cancelURL)
	} else {
		cancelBtn = `<button type="button" class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-100 dark:hover:bg-gray-800" onclick="this.closest('[data-confirm-overlay]').remove()">Cancel</button>`
	}

	return fmt.Sprintf(`<div data-confirm-overlay class="fixed inset-0 z-[10000] flex items-center justify-center bg-black/50">
	<div class="%s">
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">%s</h3>
		<p class="text-sm text-gray-600 dark:text-gray-400 mb-6">%s</p>
		<div class="flex justify-end gap-3">
			%s
			<form method="POST" action="%s" class="inline">
				<button type="submit" class="px-4 py-2 rounded-lg bg-blue-600 text-white text-sm font-medium hover:bg-blue-700">Confirm</button>
			</form>
		</div>
	</div>
</div>`, class, title, message, cancelBtn, confirmAction)
}
