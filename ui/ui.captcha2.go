package ui

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CaptchaSession stores server-side CAPTCHA information
type CaptchaSession struct {
	Text        string
	CreatedAt   time.Time
	Attempts    int
	Solved      bool
	ExpiresAt   time.Time
	MaxAttempts int
}

// Global CAPTCHA session store (in production, use Redis or database)
var (
	captchaSessions   = make(map[string]*CaptchaSession)
	captchaSessionsMu sync.RWMutex
)

const (
	defaultCaptchaLength   = 6
	defaultCaptchaLifetime = 5 * time.Minute
	cleanupGracePeriod     = 10 * time.Minute
	defaultCaptchaAttempts = 3
)

// generateSecureCaptchaText creates a cryptographically secure random CAPTCHA text
func generateSecureCaptchaText(length int) (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	result := make([]byte, length)
	for i := range length {
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
func storeCaptchaSession(sessionID string, session *CaptchaSession) {
	cleanupExpiredCaptchaSessions()
	captchaSessionsMu.Lock()
	captchaSessions[sessionID] = session
	captchaSessionsMu.Unlock()
}

// CreateCaptchaSession creates a new server-side CAPTCHA session
func createCaptchaSession(sessionID string, length int, lifetime time.Duration, attemptLimit int) (*CaptchaSession, error) {
	if length <= 0 {
		length = defaultCaptchaLength
	}

	captchaText, err := generateSecureCaptchaText(length)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CAPTCHA text: %w", err)
	}

	if lifetime <= 0 {
		lifetime = defaultCaptchaLifetime
	}

	if attemptLimit <= 0 {
		attemptLimit = defaultCaptchaAttempts
	}

	session := &CaptchaSession{
		Text:        captchaText,
		CreatedAt:   time.Now(),
		Attempts:    0,
		Solved:      false,
		ExpiresAt:   time.Now().Add(lifetime),
		MaxAttempts: attemptLimit,
	}

	storeCaptchaSession(sessionID, session)
	return session, nil
}

// ValidateCaptcha validates a CAPTCHA answer against the server-side session
func validateCaptcha(sessionID, answer string) (bool, error) {
	captchaSessionsMu.Lock()
	defer captchaSessionsMu.Unlock()

	session, exists := captchaSessions[sessionID]
	if !exists {
		return false, fmt.Errorf("CAPTCHA session not found")
	}

	now := time.Now()
	if sessionExpired(session, now) {
		delete(captchaSessions, sessionID)
		return false, fmt.Errorf("CAPTCHA session expired")
	}

	limit := session.MaxAttempts
	if limit <= 0 {
		limit = defaultCaptchaAttempts
	}

	session.Attempts++
	if session.Attempts > limit {
		delete(captchaSessions, sessionID)
		return false, fmt.Errorf("too many CAPTCHA attempts")
	}

	if session.Solved {
		return true, nil
	}

	if strings.EqualFold(strings.TrimSpace(answer), session.Text) {
		session.Solved = true
		return true, nil
	}

	return false, nil
}

func sessionExpired(session *CaptchaSession, now time.Time) bool {
	if session == nil {
		return true
	}

	if !session.ExpiresAt.IsZero() {
		return now.After(session.ExpiresAt)
	}

	return now.Sub(session.CreatedAt) > defaultCaptchaLifetime
}

// cleanupExpiredCaptchaSessions removes expired sessions
func cleanupExpiredCaptchaSessions() {
	now := time.Now()
	captchaSessionsMu.Lock()
	defer captchaSessionsMu.Unlock()

	for id, session := range captchaSessions {
		if sessionExpired(session, now) || now.Sub(session.CreatedAt) > cleanupGracePeriod {
			delete(captchaSessions, id)
		}
	}
}

type Captcha2Component struct {
	answerFieldName         string
	sessionFieldName        string
	clientVerifiedFieldName string
	codeLength              int
	sessionLifetime         time.Duration
	attemptLimit            int
}

// Captcha2 constructs a configurable CAPTCHA component with built-in session
// storage and validation helpers. Each call to Render() creates a new
// challenge and stores it in the in-memory session map.
func Captcha2() *Captcha2Component {
	return &Captcha2Component{
		answerFieldName:         "captcha_answer",
		sessionFieldName:        "captcha_session",
		clientVerifiedFieldName: "captcha_client_verified",
		codeLength:              defaultCaptchaLength,
		sessionLifetime:         defaultCaptchaLifetime,
		attemptLimit:            defaultCaptchaAttempts,
	}
}

// AnswerField sets the form field used for the user-supplied CAPTCHA answer.
func (c *Captcha2Component) AnswerField(name string) *Captcha2Component {
	if name != "" {
		c.answerFieldName = name
	}
	return c
}

// SessionField sets the hidden form field used to transport the CAPTCHA session ID.
func (c *Captcha2Component) SessionField(name string) *Captcha2Component {
	if name != "" {
		c.sessionFieldName = name
	}
	return c
}

// ClientVerifiedField customises the optional hidden field used for client-side hints.
func (c *Captcha2Component) ClientVerifiedField(name string) *Captcha2Component {
	if name != "" {
		c.clientVerifiedFieldName = name
	}
	return c
}

// Length configures the number of characters in the generated CAPTCHA challenge.
func (c *Captcha2Component) Length(n int) *Captcha2Component {
	if n > 0 {
		c.codeLength = n
	}
	return c
}

// Lifetime configures how long the generated CAPTCHA session remains valid.
func (c *Captcha2Component) Lifetime(d time.Duration) *Captcha2Component {
	if d > 0 {
		c.sessionLifetime = d
	}
	return c
}

// Attempts configures how many attempts are permitted before the session is discarded.
func (c *Captcha2Component) Attempts(limit int) *Captcha2Component {
	if limit > 0 {
		c.attemptLimit = limit
	}
	return c
}

func (c *Captcha2Component) AnswerFieldName() string {
	if c == nil || c.answerFieldName == "" {
		return "captcha_answer"
	}
	return c.answerFieldName
}

func (c *Captcha2Component) SessionFieldName() string {
	if c == nil || c.sessionFieldName == "" {
		return "captcha_session"
	}
	return c.sessionFieldName
}

func (c *Captcha2Component) ClientVerifiedFieldName() string {
	if c == nil || c.clientVerifiedFieldName == "" {
		return "captcha_client_verified"
	}
	return c.clientVerifiedFieldName
}

func (c *Captcha2Component) codeLengthValue() int {
	if c == nil || c.codeLength <= 0 {
		return defaultCaptchaLength
	}
	return c.codeLength
}

func (c *Captcha2Component) lifetimeValue() time.Duration {
	if c == nil || c.sessionLifetime <= 0 {
		return defaultCaptchaLifetime
	}
	return c.sessionLifetime
}

func (c *Captcha2Component) attemptLimitValue() int {
	if c == nil || c.attemptLimit <= 0 {
		return defaultCaptchaAttempts
	}
	return c.attemptLimit
}

func (c *Captcha2Component) Render() string {
	if c == nil {
		return renderCaptchaError("Captcha component not initialised")
	}

	sessionID, err := generateSecureID("captcha_session_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}

	session, err := createCaptchaSession(sessionID, c.codeLengthValue(), c.lifetimeValue(), c.attemptLimitValue())
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA. Please refresh the page and try again.")
	}

	canvasID, err := generateSecureID("captchaCanvas_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}
	inputID, err := generateSecureID("captchaInput_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}
	hiddenFieldID, err := generateSecureID("captchaVerified_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}

	text := escapeJS(session.Text)

	return Div("", Attr{Style: "display: flex; flex-wrap: wrap; align-items: center; gap: 10px; margin-bottom: 10px; width: 100%;"})(
		Canvas("", Attr{ID: canvasID, Style: "border: 1px solid #ccc; width: 100%; max-width: 320px; height: 96px;"})(),
		Input("w-full sm:w-auto flex-1 min-w-0", Attr{ID: inputID, Type: "text", Name: c.AnswerFieldName(), Placeholder: "Enter text from image", Required: true, Autocomplete: "off"}),
		Input("", Attr{Type: "hidden", Name: c.SessionFieldName(), Value: sessionID}),
		Input("", Attr{ID: hiddenFieldID, Type: "hidden", Name: c.ClientVerifiedFieldName(), Value: "false"}),
		Script(fmt.Sprintf(`
            setTimeout(function() {
                var canvas = document.getElementById('%s');
                if (!canvas) { return; }
                var ctx = canvas.getContext('2d');
                if (!ctx) { return; }
                var input = document.getElementById('%s');
                var hiddenField = document.getElementById('%s');
                var captchaText = '%s';

                function sizeCanvas() {
                    var ratio = window.devicePixelRatio || 1;
                    var displayWidth = Math.min(320, canvas.clientWidth || 320);
                    var displayHeight = 96;
                    canvas.width = Math.floor(displayWidth * ratio);
                    canvas.height = Math.floor(displayHeight * ratio);
                    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
                    canvas.style.width = displayWidth + 'px';
                    canvas.style.height = displayHeight + 'px';
                }

                function drawCaptcha() {
                    sizeCanvas();
                    var w = canvas.clientWidth || 320;
                    var h = canvas.clientHeight || 96;
                    ctx.clearRect(0, 0, w, h);
                    ctx.fillStyle = '#f0f0f0';
                    ctx.fillRect(0, 0, w, h);

                    ctx.font = 'bold 24px Arial';
                    ctx.textBaseline = 'middle';
                    ctx.textAlign = 'center';

                    for (var i = 0; i < captchaText.length; i++) {
                        var char = captchaText[i];
                        var x = (w / captchaText.length) * i + (w / captchaText.length) / 2;
                        var y = h / 2 + (Math.random() * 10 - 5);

                        ctx.save();
                        ctx.translate(x, y);
                        ctx.rotate((Math.random() * 0.5 - 0.25));
                        ctx.fillStyle = 'rgb(' + Math.floor(Math.random() * 200) + ',' + Math.floor(Math.random() * 200) + ',' + Math.floor(Math.random() * 200) + ')';
                        ctx.fillText(char, 0, 0);
                        ctx.restore();
                    }

                    for (var i = 0; i < 20; i++) {
                        ctx.beginPath();
                        ctx.arc(Math.random() * w, Math.random() * h, Math.random() * 2, 0, Math.PI * 2);
                        ctx.fillStyle = 'rgba(0,0,0,0.3)';
                        ctx.fill();
                    }
                }

                function validateCaptcha() {
                    if (!input) { return; }
                    if (input.value.toLowerCase() === captchaText.toLowerCase()) {
                        if (hiddenField) { hiddenField.value = 'true'; }
                        input.style.borderColor = 'green';
                    } else {
                        if (hiddenField) { hiddenField.value = 'false'; }
                        input.style.borderColor = 'red';
                    }
                }

                if (input) {
                    input.addEventListener('input', validateCaptcha);
                }

                drawCaptcha();
                window.addEventListener('resize', drawCaptcha);
            }, 300);
        `, canvasID, inputID, hiddenFieldID, text)),
	)
}

func renderCaptchaError(message string) string {
	return Div("text-red-600 bg-red-50 p-2 border border-red-200 rounded")(message)
}

// ValidateValues verifies a supplied answer against a stored CAPTCHA session.
func (c *Captcha2Component) ValidateValues(sessionID, answer string) (bool, error) {
	if sessionID == "" {
		return false, fmt.Errorf("CAPTCHA session missing")
	}
	return validateCaptcha(sessionID, answer)
}

// Validate provides a convenience alias for ValidateValues.
func (c *Captcha2Component) Validate(sessionID, answer string) (bool, error) {
	return c.ValidateValues(sessionID, answer)
}

// ValidateRequest extracts the CAPTCHA fields from an HTTP request and validates them.
func (c *Captcha2Component) ValidateRequest(r *http.Request) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("missing request")
	}

	if err := r.ParseForm(); err != nil {
		return false, fmt.Errorf("failed to parse form: %w", err)
	}

	return c.ValidateValues(
		r.FormValue(c.SessionFieldName()),
		r.FormValue(c.AnswerFieldName()),
	)
}
