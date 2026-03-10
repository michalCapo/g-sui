package pages

import (
	"time"

	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// NewComponents demonstrates new UI components: file upload, confirm dialog,
// modal/preview, badge counter, and toast notifications.
func NewComponents(ctx *ui.Context) string {
	return ui.Div("max-w-5xl mx-auto flex flex-col gap-8")(
		ui.Div("text-3xl font-bold")("New Components"),
		ui.Div("text-gray-600")("File upload, confirm dialog, modal/preview, badge counter, and toast/notifications."),

		// --- File Upload ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("File Upload"),
			ui.Div("text-gray-500 text-sm mb-4")("Drag & drop or click to browse. Validates file size (max 10MB). Uploads to /api/new/upload."),
			js.Client(ctx).
				Source("/api/new/upload").
				Upload(10*1024*1024, ".pdf,.jpg,.png,.txt,.csv").
				MaxFiles(5).
				Render(),
		),

		// --- Confirm Dialog ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Confirm Dialog"),
			ui.Div("text-gray-500 text-sm mb-4")("Styled replacement for browser confirm(). Returns a Promise."),
			ui.Div("flex flex-wrap gap-3")(
				ui.Button().Color(ui.Blue).Size(ui.SM).
					Click("__confirm('Are you sure you want to proceed?', function(){ __notify('Confirmed!','success'); }, {title:'Confirm Action'})").
					Render("Default Confirm"),
				ui.Button().Color(ui.Red).Size(ui.SM).
					Click("__confirm('This action cannot be undone. All data will be permanently deleted.', function(){ __notify('Deleted!','success'); }, {title:'Delete Item', variant:'danger', confirmText:'Delete', cancelText:'Keep'})").
					Render("Danger Confirm"),
				ui.Button().Color(ui.Gray).Size(ui.SM).
					Click("__confirm('Do you want to save changes?').then(function(ok){ __notify(ok ? 'Saved!' : 'Cancelled', ok ? 'success' : 'info'); })").
					Render("Promise-based"),
			),
		),

		// --- Server-rendered Confirm Dialog ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Server-Rendered Confirm Dialog"),
			ui.Div("text-gray-500 text-sm mb-4")("ui.ConfirmDialog() renders a server-side modal with a POST form. The dialog below is rendered server-side."),
			ui.Div("", ui.Attr{ID: "confirm-container", Style: "display:none"})(
				ui.ConfirmDialog("Confirm Submission", "Are you sure you want to submit this form? This will send the data to the server.", "/api/new/submit-form", "", ""),
			),
			ui.Button().Color(ui.Blue).Size(ui.SM).
				Click("document.getElementById('confirm-container').style.display='block'").
				Render("Show Server Confirm"),
		),

		// --- Modal / Preview ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Modal / Preview"),
			ui.Div("text-gray-500 text-sm mb-4")("Lightweight modal overlay for images, text, and HTML. Close with Escape or click outside."),
			ui.Div("flex flex-wrap gap-3")(
				ui.Button().Color(ui.Blue).Size(ui.SM).
					Click("__cmodal.open('https://picsum.photos/800/600')").
					Render("Image Preview"),
				ui.Button().Color(ui.Green).Size(ui.SM).
					Click(`__cmodal.open('This is a plain text preview.\n\nIt preserves whitespace and line breaks.\nThe modal auto-sizes to content.')`).
					Render("Text Preview"),
				ui.Button().Color(ui.Purple).Size(ui.SM).
					Click(`__cmodal.open('<h2 style=\'margin:0 0 12px\'>Custom HTML</h2><p>This modal displays <strong>HTML content</strong> with full styling support.</p><ul style=\'margin:8px 0;padding-left:20px\'><li>Feature 1</li><li>Feature 2</li><li>Feature 3</li></ul>')`).
					Render("HTML Content"),
			),
		),

		// --- Badge / Counter ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Badge / Counter"),
			ui.Div("text-gray-500 text-sm mb-4")("Live badge that polls a count endpoint every 5 seconds. The badge hides when count is 0."),
			ui.Div("flex items-center gap-4")(
				ui.Div("flex items-center gap-2")(
					ui.Span("text-sm text-gray-600 dark:text-gray-400")("Notifications:"),
					js.Client(ctx).
						Source("/api/new/notification-count").
						Component("badge", js.Opts{}).
						Poll(5*time.Second).
						Render(),
				),
				ui.Div("flex items-center gap-2")(
					ui.Span("text-sm text-gray-600 dark:text-gray-400")("Messages:"),
					js.Client(ctx).
						Source("/api/new/notification-count").
						Component("badge", js.Opts{"bg": "bg-blue-500"}).
						Poll(8*time.Second).
						Render(),
				),
			),
		),

		// --- Toast / Notifications ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Toast / Notifications"),
			ui.Div("text-gray-500 text-sm mb-4")("The __toast API wraps __notify with convenience methods."),
			ui.Div("flex flex-wrap gap-3")(
				ui.Button().Color(ui.Green).Size(ui.SM).Click("__toast.success('Operation completed successfully!')").Render("Success"),
				ui.Button().Color(ui.Red).Size(ui.SM).Click("__toast.error('Something went wrong!')").Render("Error"),
				ui.Button().Color(ui.Blue).Size(ui.SM).Click("__toast.info('Here is some useful information.')").Render("Info"),
				ui.Button().Color(ui.Gray).Size(ui.SM).Click("__toast.show('Custom toast message', 'info', 3000)").Render("Custom (3s)"),
			),
		),
	)
}
