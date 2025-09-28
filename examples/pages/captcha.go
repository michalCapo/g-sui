package pages

import (
	"github.com/michalCapo/g-sui/ui"
)

func validated(ctx *ui.Context) string {
	return ui.Div("text-green-600")("Captcha validated successfully!")
}

func Captcha(ctx *ui.Context) string {
	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("CAPTCHA Component Examples"),
		ui.Div("text-gray-600")("Demonstrates the reusable CAPTCHA component with server-side validation helpers."),

		ui.Div("bg-white p-4 rounded-lg shadow-md")(
			ui.Captcha2(validated).Render(ctx),
		),
	)
}
