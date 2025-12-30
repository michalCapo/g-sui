package ui

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// Captcha3Component provides a drag-and-drop captcha experience where users
// rearrange characters to match the target sequence rendered inline.
type Captcha3Component struct {
	sessionFieldName        string
	arrangementFieldName    string
	clientVerifiedFieldName string
	characterCount          int
	sessionLifetime         time.Duration
	attemptLimit            int
	onValidated             Callable
}

// Captcha3 constructs a configurable CAPTCHA component featuring drag & drop
// reordering with a styled tile board that reflects the current arrangement.
func Captcha3(onValidated Callable) *Captcha3Component {
	component := &Captcha3Component{
		sessionFieldName:        "captcha_session",
		arrangementFieldName:    "captcha_arrangement",
		clientVerifiedFieldName: "captcha_client_verified",
		characterCount:          4,
		sessionLifetime:         defaultCaptchaLifetime,
		attemptLimit:            defaultCaptchaAttempts,
		onValidated:             onValidated,
	}

	if component.onValidated == nil {
		component.onValidated = func(*Context) string {
			return Div("text-green-600")("Captcha validated successfully!")
		}
	}

	return component
}

// SessionField overrides the hidden form field used to transport the session ID.
func (c *Captcha3Component) SessionField(name string) *Captcha3Component {
	if name != "" {
		c.sessionFieldName = name
	}
	return c
}

// ArrangementField overrides the hidden field storing the current character order.
func (c *Captcha3Component) ArrangementField(name string) *Captcha3Component {
	if name != "" {
		c.arrangementFieldName = name
	}
	return c
}

// ClientVerifiedField overrides the optional hidden flag toggled when solved client-side.
func (c *Captcha3Component) ClientVerifiedField(name string) *Captcha3Component {
	if name != "" {
		c.clientVerifiedFieldName = name
	}
	return c
}

// Count configures how many characters are generated for the captcha challenge.
func (c *Captcha3Component) Count(n int) *Captcha3Component {
	if n > 0 {
		c.characterCount = n
	}
	return c
}

// Lifetime configures how long the generated captcha session remains valid.
func (c *Captcha3Component) Lifetime(d time.Duration) *Captcha3Component {
	if d > 0 {
		c.sessionLifetime = d
	}
	return c
}

// Attempts configures how many validation attempts are permitted for the session.
func (c *Captcha3Component) Attempts(limit int) *Captcha3Component {
	if limit > 0 {
		c.attemptLimit = limit
	}
	return c
}

// SessionFieldName returns the configured session hidden input name.
func (c *Captcha3Component) SessionFieldName() string {
	if c == nil || c.sessionFieldName == "" {
		return "captcha_session"
	}
	return c.sessionFieldName
}

// ArrangementFieldName returns the configured arrangement hidden input name.
func (c *Captcha3Component) ArrangementFieldName() string {
	if c == nil || c.arrangementFieldName == "" {
		return "captcha_arrangement"
	}
	return c.arrangementFieldName
}

// ClientVerifiedFieldName returns the configured client verification hidden input name.
func (c *Captcha3Component) ClientVerifiedFieldName() string {
	if c == nil || c.clientVerifiedFieldName == "" {
		return "captcha_client_verified"
	}
	return c.clientVerifiedFieldName
}

func (c *Captcha3Component) characterCountValue() int {
	if c == nil || c.characterCount <= 0 {
		return 5
	}
	return c.characterCount
}

func (c *Captcha3Component) lifetimeValue() time.Duration {
	if c == nil || c.sessionLifetime <= 0 {
		return defaultCaptchaLifetime
	}
	return c.sessionLifetime
}

func (c *Captcha3Component) attemptLimitValue() int {
	if c == nil || c.attemptLimit <= 0 {
		return defaultCaptchaAttempts
	}
	return c.attemptLimit
}

// Render builds the captcha markup and accompanying behaviour script.
func (c *Captcha3Component) Render(ctx *Context) string {
	if c == nil {
		return renderCaptchaError("Captcha component not initialised")
	}

	sessionID, err := generateSecureID("captcha_session_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}

	session, err := createCaptchaSession(sessionID, c.characterCountValue(), c.lifetimeValue(), c.attemptLimitValue())
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA. Please refresh the page and try again.")
	}

	rootID, err := generateSecureID("captcha3Root_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}
	tilesID, err := generateSecureID("captcha3Tiles_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}
	targetID, err := generateSecureID("captcha3Target_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}
	arrangementFieldID, err := generateSecureID("captcha3Arrangement_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}
	clientFlagID, err := generateSecureID("captcha3Verified_")
	if err != nil {
		return renderCaptchaError("Error generating CAPTCHA IDs")
	}

	successPath := ""
	if ctx != nil && ctx.App != nil && c.onValidated != nil {
		if callable := ctx.Callable(c.onValidated); callable != nil {
			if path, ok := stored[*callable]; ok {
				successPath = path
			}
		}
	}

	scrambled := shuffleStringSecure(session.Text)
	defaultSuccess := Div("text-green-600")("Captcha validated successfully!")

	return Div("flex flex-col items-start gap-3 w-full", Attr{ID: rootID})(
		Div("")(
			Span("text-sm text-gray-600 mb-2")("Drag and drop the characters on the canvas until they match the sequence below."),
		),

		Div("flex flex-col w-full border border-gray-300 rounded-lg")(
			Div("flex flex-wrap gap-2 justify-center items-center m-4", Attr{ID: targetID})(),
			Div("flex flex-wrap gap-3 justify-center items-center rounded-b-lg border bg-gray-200 shadow-sm p-4 min-h-[7.5rem] transition-colors duration-300", Attr{ID: tilesID})(),
		),
		Hidden(c.SessionFieldName(), sessionID),
		Hidden(c.ArrangementFieldName(), scrambled, Attr{ID: arrangementFieldID}),
		Hidden(c.ClientVerifiedFieldName(), "false", Attr{ID: clientFlagID}),
		Script(fmt.Sprintf(`
            setTimeout(function () {
                var root = document.getElementById('%s');
                var tilesContainer = document.getElementById('%s');
                var targetContainer = document.getElementById('%s');
                var arrangementInput = document.getElementById('%s');
                var verifiedInput = document.getElementById('%s');
                if (!root || !tilesContainer) { return; }

                var captchaText = '%s';
                var scrambled = '%s';
                var successPath = '%s';
                var defaultSuccess = '%s';

                var solved = false;
                var tiles = scrambled ? scrambled.split('') : [];
                if (!tiles.length) { tiles = captchaText.split(''); }

                var uniqueChars = Object.create(null);
                captchaText.split('').forEach(function (c) { uniqueChars[c] = true; });
                if (tiles.join('') === captchaText && Object.keys(uniqueChars).length > 1) {
                    tiles = captchaText.split('').reverse();
                }

                function renderTarget() {
                    if (!targetContainer) { return; }
                    targetContainer.innerHTML = '';
                    captchaText.split('').forEach(function (char) {
                        var item = document.createElement('div');
                        item.className = 'inline-flex items-center justify-center px-3 py-2 rounded border text-sm font-semibold tracking-wide uppercase';
                        item.textContent = char;
                        targetContainer.appendChild(item);
                    });
                    targetContainer.setAttribute('aria-hidden', 'false');
                }

                function syncHidden() {
                    if (arrangementInput) { arrangementInput.value = tiles.join(''); }
                    if (!solved && verifiedInput) { verifiedInput.value = 'false'; }
                }

                function updateContainerAppearance() {
                    if (!tilesContainer) { return; }
                    tilesContainer.classList.toggle('border-slate-300', !solved);
                    tilesContainer.classList.toggle('bg-white', !solved);
                    tilesContainer.classList.toggle('border-green-500', solved);
                    tilesContainer.classList.toggle('bg-emerald-50', solved);
                }

                var baseTileClass = 'cursor-move select-none inline-flex items-center justify-center w-12 px-3 py-2 rounded border border-dashed border-gray-400 bg-white text-lg font-semibold shadow-sm transition-all duration-150';
                var solvedTileClass = ' bg-green-600 text-white border-green-600 shadow-none cursor-default';

                function renderTiles() {
                    if (!tilesContainer) { return; }
                    tilesContainer.innerHTML = '';
                    updateContainerAppearance();
                    for (var i = 0; i < tiles.length; i++) {
                        var tile = document.createElement('div');
                        tile.className = baseTileClass;
                        tile.textContent = tiles[i];
                        tile.setAttribute('data-index', String(i));
                        tile.setAttribute('draggable', solved ? 'false' : 'true');
                        tile.setAttribute('aria-grabbed', 'false');
                        tilesContainer.appendChild(tile);
                    }
                    tilesContainer.setAttribute('tabindex', '0');
                    tilesContainer.setAttribute('aria-live', 'polite');
                    tilesContainer.setAttribute('aria-label', 'Captcha character tiles');
                    syncHidden();
                }

                function injectSuccess(html) {
                    if (!root) { return; }
                    var output = (html && html.trim()) ? html : defaultSuccess;
                    root.innerHTML = output;
                }

                function markSolved() {
                    if (solved) { return; }
                    solved = true;
                    if (verifiedInput) { verifiedInput.value = 'true'; }
                    if (arrangementInput) { arrangementInput.value = captchaText; }

                    if (tilesContainer) {
                        var nodes = tilesContainer.children;
                        for (var i = 0; i < nodes.length; i++) {
                            var node = nodes[i];
                            node.className = baseTileClass + solvedTileClass;
                            node.setAttribute('draggable', 'false');
                        }
                    }

                    updateContainerAppearance();

                    if (successPath) {
                        fetch(successPath, {
                            method: 'POST',
                            credentials: 'same-origin',
                            headers: { 'Content-Type': 'application/json' },
                            body: '[]'
                        })
                            .then(function (resp) { if (!resp.ok) { throw new Error('HTTP ' + resp.status); } return resp.text(); })
                            .then(injectSuccess)
                            .catch(function () { injectSuccess(defaultSuccess); });
                    } else {
                        injectSuccess(defaultSuccess);
                    }
                }

                function checkSolved() {
                    if (tiles.join('') === captchaText) {
                        markSolved();
                    }
                }

                tilesContainer.addEventListener('dragstart', function (event) {
                    if (solved) { event.preventDefault(); return; }
                    var tile = event.target && event.target.closest('[data-index]');
                    if (!tile) { return; }
                    tile.setAttribute('aria-grabbed', 'true');
                    tile.classList.add('ring-2', 'ring-blue-300');
                    event.dataTransfer.effectAllowed = 'move';
                    event.dataTransfer.setData('text/plain', tile.getAttribute('data-index') || '0');
                });

                tilesContainer.addEventListener('dragover', function (event) {
                    if (solved) { return; }
                    event.preventDefault();
                    event.dataTransfer.dropEffect = 'move';
                });

                tilesContainer.addEventListener('drop', function (event) {
                    if (solved) { return; }
                    event.preventDefault();
                    var payload = event.dataTransfer.getData('text/plain');
                    var from = parseInt(payload, 10);
                    if (isNaN(from) || from < 0 || from >= tiles.length) { return; }

                    var target = event.target && event.target.closest('[data-index]');
                    var to = target ? parseInt(target.getAttribute('data-index') || '0', 10) : tiles.length;
                    if (isNaN(to)) { to = tiles.length; }
                    if (to > tiles.length) { to = tiles.length; }

                    var char = tiles.splice(from, 1)[0];
                    if (from < to) { to -= 1; }
                    tiles.splice(to, 0, char);

                    renderTiles();
                    checkSolved();
                });

                tilesContainer.addEventListener('dragend', function (event) {
                    var tile = event.target && event.target.closest('[data-index]');
                    if (tile) {
                        tile.setAttribute('aria-grabbed', 'false');
                        tile.classList.remove('ring-2', 'ring-blue-300');
                    }
                });

                tilesContainer.addEventListener('dragleave', function (event) {
                    var tile = event.target && event.target.closest('[data-index]');
                    if (tile) {
                        tile.classList.remove('ring-2', 'ring-blue-300');
                    }
                });

                renderTarget();
                renderTiles();
                checkSolved();
            }, 250);
        `, rootID, tilesID, targetID, arrangementFieldID, clientFlagID, escapeJS(session.Text), escapeJS(scrambled), escapeJS(successPath), escapeJS(defaultSuccess))),
	)
}

// ValidateValues provides server-side validation by comparing stored captcha text with the supplied arrangement.
func (c *Captcha3Component) ValidateValues(sessionID, arrangement string) (bool, error) {
	if sessionID == "" {
		return false, fmt.Errorf("CAPTCHA session missing")
	}
	return validateCaptcha(sessionID, arrangement)
}

// Validate offers a convenience alias for ValidateValues.
func (c *Captcha3Component) Validate(sessionID, arrangement string) (bool, error) {
	return c.ValidateValues(sessionID, arrangement)
}

// ValidateRequest extracts captcha fields from the HTTP request and validates them.
func (c *Captcha3Component) ValidateRequest(r *http.Request) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("missing request")
	}

	if err := r.ParseForm(); err != nil {
		return false, fmt.Errorf("failed to parse form: %w", err)
	}

	return c.ValidateValues(
		r.FormValue(c.SessionFieldName()),
		r.FormValue(c.ArrangementFieldName()),
	)
}

// shuffleStringSecure returns a securely shuffled version of the provided string.
func shuffleStringSecure(input string) string {
	runes := []rune(input)
	length := len(runes)
	if length <= 1 {
		return input
	}

	for i := length - 1; i > 0; i-- {
		j := secureRandomIndex(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}

	if string(runes) == input {
		if hasMultipleUniqueRunes(runes) {
			runes[0], runes[length-1] = runes[length-1], runes[0]
			if string(runes) == input && length > 3 {
				runes[1], runes[length-2] = runes[length-2], runes[1]
			}
		}
	}

	return string(runes)
}

func secureRandomIndex(n int) int {
	if n <= 0 {
		return 0
	}
	max := big.NewInt(int64(n))
	value, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		fallback := time.Now().UnixNano()
		if fallback < 0 {
			fallback = -fallback
		}
		return int(fallback % int64(n))
	}
	return int(value.Int64())
}

func hasMultipleUniqueRunes(runes []rune) bool {
	if len(runes) <= 1 {
		return false
	}
	seen := make(map[rune]struct{})
	for _, r := range runes {
		seen[r] = struct{}{}
		if len(seen) > 1 {
			return true
		}
	}
	return false
}
