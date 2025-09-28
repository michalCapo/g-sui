package ui

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// CaptchaSession stores server-side CAPTCHA information
type CaptchaSession struct {
	Text      string
	CreatedAt time.Time
	Attempts  int
	Solved    bool
}

// Global CAPTCHA session store (in production, use Redis or database)
var captchaSessions = make(map[string]*CaptchaSession)

// generateSecureCaptchaText creates a cryptographically secure random CAPTCHA text
func generateSecureCaptchaText(length int) (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[int(b[i])%len(chars)]
	}

	return string(result), nil
}

// generateSecureID creates a cryptographically secure ID with the given prefix
func generateSecureID(prefix string) (string, error) {
	b := make([]byte, 8) // 64 bits of entropy
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure ID: %w", err)
	}
	
	return fmt.Sprintf("%s%x", prefix, b), nil
}

// StoreCaptchaSession manually stores a captcha session (for testing purposes)
func StoreCaptchaSession(sessionID string, session *CaptchaSession) {
	cleanupExpiredCaptchaSessions()
	captchaSessions[sessionID] = session
}

// CreateCaptchaSession creates a new server-side CAPTCHA session
func CreateCaptchaSession(sessionID string) (*CaptchaSession, error) {
	captchaText, err := generateSecureCaptchaText(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CAPTCHA text: %w", err)
	}

	session := &CaptchaSession{
		Text:      captchaText,
		CreatedAt: time.Now(),
		Attempts:  0,
		Solved:    false,
	}

	// Clean up old sessions
	cleanupExpiredCaptchaSessions()

	captchaSessions[sessionID] = session
	return session, nil
}

// ValidateCaptcha validates a CAPTCHA answer against the server-side session
func ValidateCaptcha(sessionID, answer string) (bool, error) {
	session, exists := captchaSessions[sessionID]
	if !exists {
		return false, fmt.Errorf("CAPTCHA session not found")
	}

	// Check if session is expired (5 minutes)
	if time.Since(session.CreatedAt) > 5*time.Minute {
		delete(captchaSessions, sessionID)
		return false, fmt.Errorf("CAPTCHA session expired")
	}

	// Check attempt limit (prevent brute force)
	session.Attempts++
	if session.Attempts > 3 {
		delete(captchaSessions, sessionID)
		return false, fmt.Errorf("too many CAPTCHA attempts")
	}

	// Already solved
	if session.Solved {
		return true, nil
	}

	// Check answer (case insensitive)
	if strings.ToLower(strings.TrimSpace(answer)) == strings.ToLower(session.Text) {
		session.Solved = true
		return true, nil
	}

	return false, nil
}

// cleanupExpiredCaptchaSessions removes expired sessions
func cleanupExpiredCaptchaSessions() {
	now := time.Now()
	for id, session := range captchaSessions {
		if now.Sub(session.CreatedAt) > 10*time.Minute {
			delete(captchaSessions, id)
		}
	}
}

// Captcha2 creates a client-side JavaScript CAPTCHA with server-side validation support.
// It returns HTML containing a canvas for the CAPTCHA image, an input field,
// and inline JavaScript to handle the CAPTCHA logic.
//
// IMPORTANT: This now includes server-side validation support.
// Use ValidateCaptcha() server-side to verify the CAPTCHA solution.
func Captcha2(sessionID string) string {
	// Create server-side CAPTCHA session
	session, err := CreateCaptchaSession(sessionID)
	if err != nil {
		return Div("text-red-600 bg-red-50 p-2 border border-red-200 rounded")(
			"Error generating CAPTCHA. Please refresh the page and try again.",
		)
	}

	// Generate cryptographically secure IDs for HTML elements
	canvasID, err := generateSecureID("captchaCanvas_")
	if err != nil {
		return Div("text-red-600")("Error generating CAPTCHA IDs")
	}
	inputID, err := generateSecureID("captchaInput_")
	if err != nil {
		return Div("text-red-600")("Error generating CAPTCHA IDs")
	}
	hiddenFieldID, err := generateSecureID("captchaVerified_")
	if err != nil {
		return Div("text-red-600")("Error generating CAPTCHA IDs")
	}

    return Div("", Attr{Style: "display: flex; flex-wrap: wrap; align-items: center; gap: 10px; margin-bottom: 10px; width: 100%;"})(
        // Responsive canvas: width 100% up to a max; height via CSS, real pixel size set in JS
        Canvas("", Attr{ID: canvasID, Style: "border: 1px solid #ccc; width: 100%; max-width: 320px; height: 96px;"})(),
        // Text input becomes full-width on narrow screens
        Input("w-full sm:w-auto flex-1 min-w-0", Attr{ID: inputID, Type: "text", Name: "captcha_answer", Placeholder: "Enter text from image", Required: true, Autocomplete: "off"}),
        // Hidden session ID for server-side validation
        Input("", Attr{Type: "hidden", Name: "captcha_session", Value: sessionID}),
        // Hidden verification field (client-side indicator only)
        Input("", Attr{ID: hiddenFieldID, Type: "hidden", Name: "captcha_client_verified", Value: "false"}),
        Script(fmt.Sprintf(`
        setTimeout(function() {
            const canvas = document.getElementById('%s');
            const ctx = canvas.getContext('2d');
            const input = document.getElementById('%s');
            const hiddenField = document.getElementById('%s');
            const captchaText = '%s';

            function sizeCanvas() {
                const ratio = window.devicePixelRatio || 1;
                const displayWidth = Math.min(320, canvas.clientWidth || 320);
                const displayHeight = 96;
                canvas.width = Math.floor(displayWidth * ratio);
                canvas.height = Math.floor(displayHeight * ratio);
                ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
                canvas.style.width = displayWidth + 'px';
                canvas.style.height = displayHeight + 'px';
            }

            function drawCaptcha() {
                sizeCanvas();
                const w = canvas.clientWidth || 320;
                const h = canvas.clientHeight || 96;
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

                for (let i = 0; i < 20; i++) {
                    ctx.beginPath();
                    ctx.arc(Math.random() * w, Math.random() * h, Math.random() * 2, 0, Math.PI * 2);
                    ctx.fillStyle = 'rgba(0,0,0,0.3)';
                    ctx.fill();
                }
            }

            function validateCaptcha() {
				// Client-side validation for immediate feedback only
				// Server-side validation is required for security
				if (input.value.toLowerCase() === captchaText.toLowerCase()) {
					hiddenField.value = 'true';
					input.style.borderColor = 'green';
				} else {
					hiddenField.value = 'false';
					input.style.borderColor = 'red';
				}
			}

            input.addEventListener('input', validateCaptcha);
            drawCaptcha();
            window.addEventListener('resize', drawCaptcha);
        }, 300);
        `, canvasID, inputID, hiddenFieldID, session.Text)),
    )
}

// SimpleCaptcha creates a simple server-side validated CAPTCHA without canvas
func SimpleCaptcha(sessionID string) string {
	session, err := CreateCaptchaSession(sessionID)
	if err != nil {
		return Div("text-red-600 bg-red-50 p-2 border border-red-200 rounded")(
			"Error generating CAPTCHA. Please refresh the page and try again.",
		)
	}

	return Div("flex items-center gap-4 p-4 bg-gray-50 rounded border")(
		Div("text-lg font-mono bg-gray-200 px-4 py-2 rounded border-2 border-dashed select-none")(
			session.Text,
		),
		Input("flex-1", Attr{
			Type:        "text", 
			Name:        "captcha_answer", 
			Placeholder: "Enter the code above", 
			Required:    true, 
			Autocomplete: "off",
		}),
		Input("", Attr{Type: "hidden", Name: "captcha_session", Value: sessionID}),
	)
}

// Legacy function for backward compatibility - now uses server-side validation
func Captcha2Legacy() string {
	// Generate a unique session ID for this CAPTCHA
	sessionID := RandomString(16)
	return Captcha2(sessionID)
}