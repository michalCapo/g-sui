package pages

import (
	"time"

	"github.com/michalCapo/g-sui/ui"
)

type CaptchaForm struct {
	CaptchaAnswer  string
	CaptchaSession string
	Message        string
}

func Captcha(ctx *ui.Context) string {
	form := CaptchaForm{}

	// Handle form submission
	if ctx.Request.Method == "POST" {
		err := ctx.Body(&form)
		if err != nil {
			form.Message = "Error processing form: " + err.Error()
		} else {
			// Validate captcha on server side
			valid, err := ui.ValidateCaptcha(form.CaptchaSession, form.CaptchaAnswer)
			if err != nil {
				form.Message = "Captcha validation error: " + err.Error()
				ctx.Error("Captcha validation failed: " + err.Error())
			} else if valid {
				form.Message = "Captcha validated successfully! Form submission would proceed here."
				ctx.Success("Captcha solved correctly!")
			} else {
				form.Message = "Incorrect captcha. Please try again."
				ctx.Error("Incorrect captcha answer")
			}
		}
	}

	// Generate a unique session ID for the captcha (use the user's session ID plus a suffix for uniqueness)
	sessionID := ctx.SessionID + "_captcha_" + ui.RandomString(8)

	// Create the captcha session manually with known text for testing
	captchaText := sessionID[len(sessionID)-6:] // Use last 6 chars as captcha text
	session := &ui.CaptchaSession{
		Text:      captchaText,
		CreatedAt: time.Now(),
		Attempts:  0,
		Solved:    false,
	}
	ui.StoreCaptchaSession(sessionID, session) // We'll need to create this function

	// Captcha demo section
	captchaDemo := ui.Div("bg-white p-6 rounded-lg shadow flex flex-col gap-4 w-full border")(
		ui.Div("text-xl font-bold")("Client-Side CAPTCHA Demo"),
		ui.Div("text-gray-600")("This is a demo of the legacy client-side captcha (for display purposes only)."),
		ui.Div("w-full overflow-x-auto")(ui.Captcha2Legacy()),
		ui.Div("text-sm text-amber-600 bg-amber-50 p-3 rounded border border-amber-200")(
			"Note: This is a client-side only demo. For security, use the server-validated captcha below.",
		),
	)

	// Server-validated captcha form
	formTarget := ui.Target()
	captchaForm := ui.Form("bg-white p-6 rounded-lg shadow flex flex-col gap-4 w-full border", formTarget,
		ctx.Submit(func(ctx *ui.Context) string {
			return Captcha(ctx)
		}, form).Replace(formTarget),
	)(
		ui.Div("text-xl font-bold")("Server-Validated CAPTCHA Form"),
		ui.Div("text-gray-600")("Enter the text shown in the image below to validate the captcha."),

		// Display message if any
		ui.If(form.Message != "", func() string {
			messageClass := "p-3 rounded border"
			if len(form.Message) > 0 && form.Message == "Captcha validated successfully! Form submission would proceed here." {
				messageClass += " bg-green-50 text-green-800 border-green-200"
			} else if len(form.Message) > 0 && form.Message == "Incorrect captcha. Please try again." {
				messageClass += " bg-red-50 text-red-800 border-red-200"
			} else {
				messageClass += " bg-blue-50 text-blue-800 border-blue-200"
			}
			return ui.Div(messageClass)(form.Message)
		}),

		// Captcha component - using custom inputs with Go-friendly field names
		ui.Div("", ui.Attr{Style: "display: flex; flex-wrap: wrap; align-items: center; gap: 10px; margin-bottom: 10px; width: 100%;"})(
			ui.Canvas("", ui.Attr{ID: "captcha-canvas", Style: "border: 1px solid #ccc; width: 100%; max-width: 320px; height: 96px;"})(),
			ui.Input("w-full sm:w-auto flex-1 min-w-0", ui.Attr{Type: "text", Name: "CaptchaAnswer", Placeholder: "Enter text from image", Required: true, Autocomplete: "off"}),
			ui.Input("", ui.Attr{Type: "hidden", Name: "CaptchaSession", Value: sessionID}),
			ui.Script(`
				setTimeout(function() {
					const canvas = document.getElementById('captcha-canvas');
					if (!canvas) return;
					const ctx = canvas.getContext('2d');
					const captchaText = '`+sessionID[len(sessionID)-6:]+`'; 
					
					function drawCaptcha() {
						const w = 320, h = 96;
						canvas.width = w; canvas.height = h;
						ctx.clearRect(0, 0, w, h);
						ctx.fillStyle = '#f0f0f0';
						ctx.fillRect(0, 0, w, h);
						
						ctx.font = 'bold 24px Arial';
						ctx.textBaseline = 'middle';
						ctx.textAlign = 'center';
						
						for (let i = 0; i < captchaText.length; i++) {
							const char = captchaText[i];
							const x = (w / captchaText.length) * i + (w / captchaText.length) / 2;
							const y = h / 2 + (Math.random() * 10 - 5);
							
							ctx.save();
							ctx.translate(x, y);
							ctx.rotate((Math.random() * 0.5 - 0.25));
							ctx.fillStyle = 'rgb(' + Math.floor(Math.random() * 200) + ',' + Math.floor(Math.random() * 200) + ',' + Math.floor(Math.random() * 200) + ')';
							ctx.fillText(char, 0, 0);
							ctx.restore();
						}
						
						// Add some noise
						for (let i = 0; i < 20; i++) {
							ctx.beginPath();
							ctx.arc(Math.random() * w, Math.random() * h, Math.random() * 2, 0, Math.PI * 2);
							ctx.fillStyle = 'rgba(0,0,0,0.3)';
							ctx.fill();
						}
					}
					
					drawCaptcha();
				}, 300);
			`),
		),

		// Submit button
		ui.Div("flex items-center gap-4")(
			ui.Button().Submit().Color(ui.Blue).Class("rounded px-6 py-2").Render("Validate Captcha"),
			ui.Button().Color(ui.Gray).Class("rounded px-4 py-2").Click("window.location.reload()").Render("Reset"),
		),

		// Instructions
		ui.Div("text-sm text-gray-500 mt-2")(
			"Enter the text from the image above. The validation is performed server-side for security.",
		),
	)

	// Simple captcha alternative
	simpleCaptchaTarget := ui.Target()
	simpleCaptchaForm := ui.Form("bg-white p-6 rounded-lg shadow flex flex-col gap-4 w-full border", simpleCaptchaTarget,
		ctx.Submit(func(ctx *ui.Context) string {
			form := CaptchaForm{}
			err := ctx.Body(&form)
			if err != nil {
				ctx.Error("Error processing form: " + err.Error())
				return SimpleCaptchaSection(ctx, "Error processing form: "+err.Error())
			}

			valid, err := ui.ValidateCaptcha(form.CaptchaSession, form.CaptchaAnswer)
			if err != nil {
				ctx.Error("Captcha validation failed: " + err.Error())
				return SimpleCaptchaSection(ctx, "Captcha validation error: "+err.Error())
			} else if valid {
				ctx.Success("Simple captcha solved correctly!")
				return SimpleCaptchaSection(ctx, "Simple captcha validated successfully!")
			} else {
				ctx.Error("Incorrect captcha answer")
				return SimpleCaptchaSection(ctx, "Incorrect captcha. Please try again.")
			}
		}, form).Replace(simpleCaptchaTarget),
	)(
		SimpleCaptchaSection(ctx, ""),
	)

	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("CAPTCHA Examples"),
		ui.Div("text-gray-600")("Different types of CAPTCHA implementations with server-side validation."),
		captchaDemo,
		captchaForm,
		simpleCaptchaForm,
	)
}

// SimpleCaptchaSection renders the simple captcha section
func SimpleCaptchaSection(ctx *ui.Context, message string) string {
	sessionID := ctx.SessionID + "_simple_captcha_" + ui.RandomString(8)

	// Create the captcha session manually with known text for testing
	captchaText := sessionID[len(sessionID)-6:] // Use last 6 chars as captcha text
	session := &ui.CaptchaSession{
		Text:      captchaText,
		CreatedAt: time.Now(),
		Attempts:  0,
		Solved:    false,
	}
	ui.StoreCaptchaSession(sessionID, session)

	return ui.Div("flex flex-col gap-4")(
		ui.Div("text-xl font-bold")("Simple Text CAPTCHA"),
		ui.Div("text-gray-600")("A simpler captcha without canvas - just plain text to copy."),

		// Display message if any
		ui.If(message != "", func() string {
			messageClass := "p-3 rounded border"
			if message == "Simple captcha validated successfully!" {
				messageClass += " bg-green-50 text-green-800 border-green-200"
			} else if message == "Incorrect captcha. Please try again." {
				messageClass += " bg-red-50 text-red-800 border-red-200"
			} else {
				messageClass += " bg-blue-50 text-blue-800 border-blue-200"
			}
			return ui.Div(messageClass)(message)
		}),

		// Simple captcha with Go-friendly field names
		ui.Div("flex items-center gap-4 p-4 bg-gray-50 rounded border")(
			ui.Div("text-lg font-mono bg-gray-200 px-4 py-2 rounded border-2 border-dashed select-none")(
				sessionID[len(sessionID)-6:], // Use last 6 chars of session ID as captcha text
			),
			ui.Input("flex-1", ui.Attr{Type: "text", Name: "CaptchaAnswer", Placeholder: "Enter the code above", Required: true, Autocomplete: "off"}),
			ui.Input("", ui.Attr{Type: "hidden", Name: "CaptchaSession", Value: sessionID}),
		),

		ui.Div("flex items-center gap-4")(
			ui.Button().Submit().Color(ui.Green).Class("rounded px-6 py-2").Render("Validate Simple Captcha"),
		),

		ui.Div("text-sm text-gray-500")(
			"Copy the text shown above and enter it in the input field.",
		),
	)
}
