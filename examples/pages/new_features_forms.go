package pages

import (
	"github.com/michalCapo/g-sui/js"
	"github.com/michalCapo/g-sui/ui"
)

// NewForms demonstrates form and navigation features: AJAX form, async button,
// autofill, keyboard shortcuts, SPA links, and external links.
func NewForms(ctx *ui.Context) string {
	// AutoFill mappings
	emailMappings := map[string]map[string]string{
		"gmail": {
			"imap-host": "imap.gmail.com",
			"imap-port": "993",
			"smtp-host": "smtp.gmail.com",
			"smtp-port": "587",
		},
		"outlook": {
			"imap-host": "outlook.office365.com",
			"imap-port": "993",
			"smtp-host": "smtp-mail.outlook.com",
			"smtp-port": "587",
		},
		"yahoo": {
			"imap-host": "imap.mail.yahoo.com",
			"imap-port": "993",
			"smtp-host": "smtp.mail.yahoo.com",
			"smtp-port": "465",
		},
	}

	cls := "w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm"
	lbl := "block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
	lblSm := "block text-xs text-gray-500 dark:text-gray-400 mb-1"
	kbd := "px-3 py-1.5 bg-gray-100 dark:bg-gray-700 rounded text-sm font-mono"

	return ui.Div("max-w-5xl mx-auto flex flex-col gap-8")(
		ui.Div("text-3xl font-bold")("Forms & Navigation"),
		ui.Div("text-gray-600")("AJAX form submission, async buttons, auto-fill, keyboard shortcuts, and navigation helpers."),

		// --- AJAX Form ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("AJAX Form"),
			ui.Div("text-gray-500 text-sm mb-4")("Form submits via fetch() instead of full page reload. Shows loading state on submit button."),
			ui.Form("flex flex-col gap-3 max-w-md", ui.Attr{ID: "ajax-demo-form", Method: "POST", Action: "/api/new/submit-form"})(
				ui.Div("")(
					ui.El("label", lbl)("Name"),
					ui.Input(cls, ui.Attr{Type: "text", Name: "name", Placeholder: "Enter your name", Required: true}),
				),
				ui.Div("")(
					ui.El("label", lbl)("Email"),
					ui.Input(cls, ui.Attr{Type: "email", Name: "email", Placeholder: "you@example.com", Required: true}),
				),
				ui.Div("")(
					ui.El("label", lbl)("Message"),
					ui.Textarea(cls, ui.Attr{Name: "message", Rows: 3, Placeholder: "Your message..."})(""),
				),
				ui.Button().Submit().Color(ui.Blue).Size(ui.SM).Render("Submit via AJAX"),
			),
			js.AjaxForm("ajax-demo-form", nil),
		),

		// --- Async Button ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Async Button"),
			ui.Div("text-gray-500 text-sm mb-4")("Button POSTs to a URL on click, shows loading state, and displays result inline."),
			js.AsyncButton("Run Server Action", "/api/new/async-action", "async-result", "bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700 cursor-pointer"),
		),

		// --- AutoFill ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("AutoFill"),
			ui.Div("text-gray-500 text-sm mb-4")("Select an email provider to auto-populate the IMAP/SMTP settings below."),
			ui.Div("flex flex-col gap-3 max-w-md")(
				ui.Div("")(
					ui.El("label", lbl)("Email Provider"),
					ui.Select(cls, ui.Attr{ID: "email-provider"})(
						ui.Option("", ui.Attr{Value: ""})("Select provider..."),
						ui.Option("", ui.Attr{Value: "gmail"})("Gmail"),
						ui.Option("", ui.Attr{Value: "outlook"})("Outlook"),
						ui.Option("", ui.Attr{Value: "yahoo"})("Yahoo"),
					),
				),
				ui.Div("grid grid-cols-2 gap-3")(
					ui.Div("")(
						ui.El("label", lblSm)("IMAP Host"),
						ui.Input(cls, ui.Attr{Type: "text", ID: "imap-host", Readonly: true}),
					),
					ui.Div("")(
						ui.El("label", lblSm)("IMAP Port"),
						ui.Input(cls, ui.Attr{Type: "text", ID: "imap-port", Readonly: true}),
					),
					ui.Div("")(
						ui.El("label", lblSm)("SMTP Host"),
						ui.Input(cls, ui.Attr{Type: "text", ID: "smtp-host", Readonly: true}),
					),
					ui.Div("")(
						ui.El("label", lblSm)("SMTP Port"),
						ui.Input(cls, ui.Attr{Type: "text", ID: "smtp-port", Readonly: true}),
					),
				),
			),
			js.AutoFill("email-provider", emailMappings),
		),

		// --- Keyboard Shortcuts ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Keyboard Shortcuts"),
			ui.Div("text-gray-500 text-sm mb-4")("Press ? to see all registered shortcuts. Try pressing the keys below."),
			js.Shortcuts(),
			js.RegisterShortcut("n", "__toast.info('Shortcut N pressed: New item')", "Create new item"),
			js.RegisterShortcut("d", "__toast.info('Shortcut D pressed: Delete')", "Delete selected"),
			js.RegisterShortcut("Ctrl+s", "__toast.success('Ctrl+S pressed: Saved!')", "Save document"),
			js.RegisterShortcut("Ctrl+k", "document.getElementById('live-search-input') && document.getElementById('live-search-input').focus()", "Focus search"),
			ui.Div("flex flex-wrap gap-2")(
				ui.El("kbd", kbd)("?"), ui.Span("text-sm text-gray-500 self-center")("Help"),
				ui.El("kbd", kbd)("n"), ui.Span("text-sm text-gray-500 self-center")("New"),
				ui.El("kbd", kbd)("d"), ui.Span("text-sm text-gray-500 self-center")("Delete"),
				ui.El("kbd", kbd)("Ctrl+S"), ui.Span("text-sm text-gray-500 self-center")("Save"),
			),
		),

		// --- Navigation Helpers ---
		ui.Div("bg-white dark:bg-gray-900 rounded-lg shadow p-6")(
			ui.Div("text-xl font-semibold mb-1")("Navigation Helpers"),
			ui.Div("text-gray-500 text-sm mb-4")("SPA links and external links."),
			ui.Div("flex flex-wrap gap-3")(
				ui.A("px-4 py-2 bg-blue-600 text-white rounded text-sm hover:bg-blue-700 cursor-pointer inline-block", ctx.Load("/"))("SPA: Go to Showcase"),
				ui.A("px-4 py-2 bg-purple-600 text-white rounded text-sm hover:bg-purple-700 cursor-pointer inline-block", ctx.Load("/client-charts"))("SPA: Go to Charts"),
				js.ExternalLink("https://github.com", "px-4 py-2 bg-gray-600 text-white rounded text-sm hover:bg-gray-700 cursor-pointer inline-block", "External: GitHub"),
			),
		),
	)
}
