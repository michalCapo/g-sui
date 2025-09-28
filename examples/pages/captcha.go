package pages

import (
	"net/http"

	"github.com/michalCapo/g-sui/ui"
)

func renderCaptchaMessage(message string) string {
	if message == "" {
		return ""
	}

	className := "p-3 rounded border"
	switch message {
	case "Captcha validated successfully! Form submission would proceed here.", "Simple captcha validated successfully!":
		className += " bg-green-50 text-green-800 border-green-200"
	case "Incorrect captcha. Please try again.":
		className += " bg-red-50 text-red-800 border-red-200"
	default:
		className += " bg-blue-50 text-blue-800 border-blue-200"
	}

	return ui.Div(className)(message)
}

func handleServerSubmission(ctx *ui.Context, captcha *ui.Captcha2Component) string {
	valid, err := captcha.ValidateValues(
		ctx.Request.FormValue(captcha.SessionFieldName()),
		ctx.Request.FormValue(captcha.AnswerFieldName()),
	)
	if err != nil {
		message := "Captcha validation error: " + err.Error()
		ctx.Error("Captcha validation failed: " + err.Error())
		return message
	}

	if valid {
		ctx.Success("Captcha solved correctly!")
		return "Captcha validated successfully! Form submission would proceed here."
	}

	ctx.Error("Incorrect captcha answer")
	return "Incorrect captcha. Please try again."
}

func handleSimpleSubmission(ctx *ui.Context, captcha *ui.Captcha2Component) string {
	valid, err := captcha.ValidateValues(
		ctx.Request.FormValue(captcha.SessionFieldName()),
		ctx.Request.FormValue(captcha.AnswerFieldName()),
	)
	if err != nil {
		message := "Captcha validation error: " + err.Error()
		ctx.Error("Simple captcha validation failed: " + err.Error())
		return message
	}

	if valid {
		ctx.Success("Simple captcha solved correctly!")
		return "Simple captcha validated successfully!"
	}

	ctx.Error("Incorrect captcha answer")
	return "Incorrect captcha. Please try again."
}

func Captcha(ctx *ui.Context) string {
	serverCaptcha := ui.Captcha2().
		AnswerField("CaptchaAnswer").
		SessionField("CaptchaSession").
		ClientVerifiedField("CaptchaClientVerified")

	simpleCaptcha := ui.Captcha2().
		AnswerField("SimpleCaptchaAnswer").
		SessionField("SimpleCaptchaSession").
		ClientVerifiedField("SimpleCaptchaVerified")

	var serverMessage string
	var simpleMessage string

	if ctx.Request.Method == http.MethodPost {
		if err := ctx.Request.ParseForm(); err != nil {
			message := "Error processing form: " + err.Error()
			if ctx.Request.FormValue("form_id") == "simple" {
				simpleMessage = message
			} else {
				serverMessage = message
			}
			ctx.Error(message)
		} else {
			switch ctx.Request.FormValue("form_id") {
			case "simple":
				simpleMessage = handleSimpleSubmission(ctx, simpleCaptcha)
			default:
				serverMessage = handleServerSubmission(ctx, serverCaptcha)
			}
		}
	}

	serverTarget := ui.Target()
	serverForm := ui.Form("bg-white p-6 rounded-lg shadow flex flex-col gap-4 w-full border", serverTarget,
		ctx.Submit(func(ctx *ui.Context) string {
			return Captcha(ctx)
		}).Replace(serverTarget),
	)(
		ui.Div("text-xl font-bold")("Server-Validated CAPTCHA Form"),
		ui.Div("text-gray-600")("Enter the text shown in the image below to validate the captcha."),
		ui.If(serverMessage != "", func() string { return renderCaptchaMessage(serverMessage) }),
		serverCaptcha.Render(),
		ui.Input("", ui.Attr{Type: "hidden", Name: "form_id", Value: "server"}),
		ui.Div("flex items-center gap-4")(
			ui.Button().Submit().Color(ui.Blue).Class("rounded px-6 py-2").Render("Validate Captcha"),
			ui.Button().Color(ui.Gray).Class("rounded px-4 py-2").Click("window.location.reload()").Render("Reset"),
		),
		ui.Div("text-sm text-gray-500 mt-2")(
			"Enter the text from the image above. The validation is performed server-side for security.",
		),
	)

	simpleTarget := ui.Target()
	simpleForm := ui.Form("bg-white p-6 rounded-lg shadow flex flex-col gap-4 w-full border", simpleTarget,
		ctx.Submit(func(ctx *ui.Context) string {
			return Captcha(ctx)
		}).Replace(simpleTarget),
	)(
		ui.Div("text-xl font-bold")("Reusable Component Example"),
		ui.Div("text-gray-600")("This form reuses the same component with different field names."),
		ui.If(simpleMessage != "", func() string { return renderCaptchaMessage(simpleMessage) }),
		simpleCaptcha.Render(),
		ui.Input("", ui.Attr{Type: "hidden", Name: "form_id", Value: "simple"}),
		ui.Div("flex items-center gap-4")(
			ui.Button().Submit().Color(ui.Green).Class("rounded px-6 py-2").Render("Validate Simple Captcha"),
			ui.Button().Color(ui.Gray).Class("rounded px-4 py-2").Click("window.location.reload()").Render("Reset"),
		),
		ui.Div("text-sm text-gray-500 mt-2")(
			"Reusing the component keeps validation logic in one place while customising form field names.",
		),
	)

	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("CAPTCHA Component Examples"),
		ui.Div("text-gray-600")("Demonstrates the reusable CAPTCHA component with server-side validation helpers."),
		serverForm,
		simpleForm,
	)
}
