package ui

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

//go:embed pwa.js
var pwaJS string

// PWA makes the app installable: a web app manifest, a service worker with an
// offline page, the head tags phones read and install support. Empty fields
// get the defaults in their comments. See App.PWA.
type PWA struct {
	Name        string // app name; default App.Title, else "App"
	ShortName   string // name under the home screen icon; default Name
	Description string
	ID          string // stable app identity; default StartURL
	StartURL    string // page the installed app opens; default Scope
	Scope       string // paths that belong to the app; default "/"

	// ThemeColor colors the browser UI and the title bar of the installed
	// app. Default "#ffffff".
	ThemeColor string
	// DarkThemeColor is ThemeColor in dark mode and the offline page's dark
	// background. Default ThemeColor.
	DarkThemeColor string
	// BackgroundColor is the splash screen and the offline page background.
	// Default ThemeColor.
	BackgroundColor string

	// Icons for the manifest. Browsers want at least a 192x192 and a
	// 512x512 PNG, and a separate maskable one for Android.
	Icons []PWAIcon
	// AppleTouchIcon is the URL of a 180x180 PNG for iPhone and iPad.
	AppleTouchIcon string
	Shortcuts      []PWAShortcut
	Categories     []string

	// SafeAreas lets pages draw under the notch and the home indicator
	// (viewport-fit=cover). Pad the layout with env(safe-area-inset-*).
	SafeAreas bool

	// OfflineText is shown on the page that opens when a page load finds
	// no network. Default English.
	OfflineText string

	// NotificationIcon is the image in notifications. Default the first
	// 192x192 icon that is not maskable.
	NotificationIcon string
	// NotificationBadge is the small one-color image Android shows in the
	// status bar. Default none: Android shows a generic one.
	NotificationBadge string
	// Push turns on Web Push. Without it notifications show only while a
	// page of the app is open (ShowNotification).
	Push *PWAPush

	// Worker is JavaScript appended to the service worker. g-sui already
	// shows pushes and handles notification clicks.
	//
	// This is a trusted raw API: never pass untrusted/user-controlled input to it.
	Worker string
}

// PWAIcon is a manifest icon.
type PWAIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes,omitempty"`   // "192x192", "any" for SVG
	Type    string `json:"type,omitempty"`    // "image/png"
	Purpose string `json:"purpose,omitempty"` // "any", "maskable", "monochrome"
}

// PWAShortcut is an entry of the installed app's icon menu.
type PWAShortcut struct {
	Name      string `json:"name"`
	ShortName string `json:"short_name,omitempty"`
	URL       string `json:"url"`
}

// PWA makes the app installable. It serves /manifest.webmanifest and
// /sw.js, adds the manifest, theme color and Apple tags to every page, and
// registers the service worker. The service worker only answers page loads
// that fail for lack of network, with an offline page; everything else goes
// to the network and nothing is cached. Use InstallApp and
// ShowIfInstallable for an install button. Configure before serving.
//
//	app.PWA(ui.PWA{
//		Name:  "Mail",
//		Icons: []ui.PWAIcon{{Src: "/assets/icon-512.png", Sizes: "512x512", Type: "image/png"}},
//	})
func (app *App) PWA(p PWA) {
	app.mu.Lock()
	app.pwa = &p
	app.mu.Unlock()
}

// InstallApp installs the app: it shows the browser's install prompt when
// there is one, else opens the dialog helpDialogID (the app's own steps,
// e.g. "Share, then Add to Home Screen"), if not empty. Needs App.PWA.
func InstallApp(helpDialogID string) *Action {
	return wrap(op("install", helpDialogID), "if(__gsui.install){", "}")
}

// ShowIfInstallable hides the node unless the app can be installed: the
// browser offered its install prompt, or the device installs from a menu
// (iPhone, iPad, macOS Safari). It stays hidden in the installed app. Needs
// App.PWA. Pair it with InstallApp.
func (n *Node) ShowIfInstallable() *Node { return n.Attr("data-gsui-installable", "") }

// pwaFiles are App.PWA's parts of the page shell, built once at setup.
type pwaFiles struct {
	scope    string // pages under it get the files
	head     string // tags after the description in <head>
	script   string // runs after local.js
	viewport string // added to the viewport meta content
}

// setupPWA builds the PWA files from app.pwa and registers their routes.
func (app *App) setupPWA() {
	app.mu.RLock()
	cfg := app.pwa
	app.mu.RUnlock()
	if cfg == nil {
		return
	}
	p := *cfg
	p.Name = cmp.Or(p.Name, app.Title, "App")
	p.ShortName = cmp.Or(p.ShortName, p.Name)
	p.Scope = cmp.Or(p.Scope, "/")
	p.StartURL = cmp.Or(p.StartURL, p.Scope)
	p.ID = cmp.Or(p.ID, p.StartURL)
	p.ThemeColor = cmp.Or(p.ThemeColor, "#ffffff")
	p.DarkThemeColor = cmp.Or(p.DarkThemeColor, p.ThemeColor)
	p.BackgroundColor = cmp.Or(p.BackgroundColor, p.ThemeColor)
	p.OfflineText = cmp.Or(p.OfflineText, "You are offline. Check your connection and try again.")
	p.NotificationIcon = cmp.Or(p.NotificationIcon, pwaNotificationIcon(p.Icons))

	lang, texts := app.locale(nil)
	offline := app.Asset("offline.html", pwaOfflinePage(p, lang, texts.Reload))
	manifest, _ := json.Marshal(struct {
		ID              string        `json:"id"`
		Name            string        `json:"name"`
		ShortName       string        `json:"short_name"`
		Description     string        `json:"description,omitempty"`
		Lang            string        `json:"lang"`
		StartURL        string        `json:"start_url"`
		Scope           string        `json:"scope"`
		Display         string        `json:"display"`
		ThemeColor      string        `json:"theme_color"`
		BackgroundColor string        `json:"background_color"`
		Icons           []PWAIcon     `json:"icons,omitempty"`
		Shortcuts       []PWAShortcut `json:"shortcuts,omitempty"`
		Categories      []string      `json:"categories,omitempty"`
	}{p.ID, p.Name, p.ShortName, p.Description, lang, p.StartURL, p.Scope, "standalone",
		p.ThemeColor, p.BackgroundColor, p.Icons, p.Shortcuts, p.Categories})
	// The cache name holds the offline page's content hash, so a changed
	// page replaces the old cache.
	worker := pwaWorker(offline, "gsui-offline-"+path.Base(path.Dir(offline))) + notifyWorker(p)
	if p.Worker != "" {
		worker += "\n" + p.Worker + "\n"
	}
	app.mux.HandleFunc("GET /manifest.webmanifest", pwaServe("application/manifest+json", manifest))
	serveWorker := pwaServe("application/javascript; charset=utf-8", []byte(worker))
	app.mux.HandleFunc("GET /sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Service-Worker-Allowed", "/")
		serveWorker(w, r)
	})

	var head strings.Builder
	esc := html.EscapeString
	head.WriteString("\n" + `<link rel="manifest" href="/manifest.webmanifest">` + "\n")
	// The first theme-color whose media matches wins.
	if p.DarkThemeColor != p.ThemeColor {
		fmt.Fprintf(&head, `<meta name="theme-color" content="%s" media="(prefers-color-scheme: dark)">`+"\n", esc(p.DarkThemeColor))
	}
	fmt.Fprintf(&head, `<meta name="theme-color" content="%s">
<meta name="application-name" content="%s">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-title" content="%s">
<meta name="apple-mobile-web-app-status-bar-style" content="default">`, esc(p.ThemeColor), esc(p.Name), esc(p.ShortName))
	if p.AppleTouchIcon != "" {
		fmt.Fprintf(&head, "\n"+`<link rel="apple-touch-icon" sizes="180x180" href="%s">`, esc(p.AppleTouchIcon))
	}
	head.WriteString("\n<style>html:not([data-gsui-install]) [data-gsui-installable]{display:none!important}" + notifyStyle + "</style>")
	scope, _ := json.Marshal(p.Scope)
	files := &pwaFiles{scope: p.Scope, head: head.String(), script: "\n(" + pwaJS + ")(" + string(scope) + ");" + notifyScript(p)}
	if p.Push != nil && p.Push.Save != nil {
		app.mux.HandleFunc("POST /__gsui/push", app.pushHandler(p.Push.Save))
	}
	if p.SafeAreas {
		files.viewport = ",viewport-fit=cover"
	}
	app.pwaFiles = files
}

// pwaPage reports whether the page at path gets App.PWA's files: it is in
// PWA.Scope.
func (app *App) pwaPage(path string) bool {
	return app.pwaFiles != nil && strings.HasPrefix(path, app.pwaFiles.scope)
}

// pwaNotificationIcon is the first 192x192 icon that is not maskable, else
// the first one that is not maskable.
func pwaNotificationIcon(icons []PWAIcon) string {
	first := ""
	for _, icon := range icons {
		if strings.Contains(icon.Purpose, "maskable") {
			continue
		}
		if slices.Contains(strings.Fields(icon.Sizes), "192x192") {
			return icon.Src
		}
		first = cmp.Or(first, icon.Src)
	}
	return first
}

// pwaServe serves a generated file that browsers revalidate on every use.
func pwaServe(contentType string, body []byte) http.HandlerFunc {
	sum := sha256.Sum256(body)
	tag := `"` + hex.EncodeToString(sum[:12]) + `"`
	return func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", contentType)
		h.Set("Cache-Control", "no-cache")
		h.Set("ETag", tag)
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(body))
	}
}

// pwaWorker is the service worker. It precaches the offline page and serves
// it only for page loads that fail for lack of network. Other requests
// (assets, actions, the WebSocket, APIs) never reach respondWith, so they go
// to the network as without a worker, and no private page is ever cached.
func pwaWorker(offline, cache string) string {
	o, _ := json.Marshal(offline)
	c, _ := json.Marshal(cache)
	return `// g-sui service worker: an offline page for failed page loads, no other caching.
var OFFLINE = ` + string(o) + `, CACHE = ` + string(c) + `;
self.addEventListener('install', function (event) {
  event.waitUntil(caches.open(CACHE).then(function (cache) { return cache.add(new Request(OFFLINE, {cache: 'reload'})); })
    .then(function () { return self.skipWaiting(); }));
});
self.addEventListener('activate', function (event) {
  event.waitUntil(caches.keys().then(function (keys) {
    return Promise.all(keys.map(function (key) { return key.indexOf('gsui-offline-') === 0 && key !== CACHE ? caches.delete(key) : null; }));
  }).then(function () {
    if (self.registration.navigationPreload) return self.registration.navigationPreload.enable().catch(function () {});
  }).then(function () { return self.clients.claim(); }));
});
self.addEventListener('fetch', function (event) {
  var request = event.request;
  if (request.mode !== 'navigate' || request.method !== 'GET') return;
  event.respondWith(Promise.resolve(event.preloadResponse).then(function (preloaded) { return preloaded || fetch(request); })
    .catch(function () {
      return caches.open(CACHE).then(function (cache) { return cache.match(OFFLINE); })
        .then(function (page) { return page || Response.error(); });
    }));
});
`
}

// pwaOfflinePage is a static page without the g-sui runtime or external
// resources. It reloads when the network is back.
func pwaOfflinePage(p PWA, lang, reload string) []byte {
	esc := html.EscapeString
	light := pwaColor(p.BackgroundColor, "#ffffff")
	dark := pwaColor(p.DarkThemeColor, light)
	return fmt.Appendf(nil, `<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">
<meta name="theme-color" content="%s">
<title>%s</title>
<style>
:root{color-scheme:light dark}
body{box-sizing:border-box;margin:0;min-height:100vh;min-height:100dvh;display:flex;align-items:center;justify-content:center;padding:24px max(24px,env(safe-area-inset-right)) max(24px,env(safe-area-inset-bottom)) max(24px,env(safe-area-inset-left));background:%s;color:%s;font:16px/1.5 ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif;text-align:center;-webkit-text-size-adjust:100%%}
main{max-width:22rem}
h1{margin:0 0 8px;font-size:22px;line-height:1.3}
p{margin:0 0 24px;opacity:.75}
button{min-height:44px;padding:0 24px;border:0;border-radius:9999px;background:currentColor;font:inherit;font-weight:600;cursor:pointer}
button span{color:%s}
button:focus-visible{outline:2px solid currentColor;outline-offset:3px}
@media (prefers-color-scheme:dark){body{background:%s;color:%s}button span{color:%s}}
</style>
</head>
<body>
<main>
<h1>%s</h1>
<p>%s</p>
<button type="button" onclick="location.reload()"><span>%s</span></button>
</main>
<script>addEventListener('online',function(){location.reload()})</script>
</body>
</html>
`, esc(lang), esc(p.ThemeColor), esc(p.Name),
		light, pwaInk(light), light,
		dark, pwaInk(dark), dark,
		esc(p.Name), esc(p.OfflineText), esc(reload))
}

// pwaColorRE accepts CSS colors such as #fff, rgb(0 0 0 / .5) or white, and
// nothing that could end the declaration.
var pwaColorRE = regexp.MustCompile(`^[#a-zA-Z0-9(),.%/ -]{1,64}$`)

func pwaColor(c, fallback string) string {
	if pwaColorRE.MatchString(c) {
		return c
	}
	return fallback
}

// pwaInk is dark or light text for a background. Colors other than hex
// count as light.
func pwaInk(bg string) string {
	hexColor := strings.TrimPrefix(bg, "#")
	if len(hexColor) == 3 {
		hexColor = string([]byte{hexColor[0], hexColor[0], hexColor[1], hexColor[1], hexColor[2], hexColor[2]})
	}
	v, err := strconv.ParseUint(hexColor, 16, 32)
	if !strings.HasPrefix(bg, "#") || len(hexColor) != 6 || err != nil {
		return "#111827"
	}
	r, g, b := float64(v>>16&255), float64(v>>8&255), float64(v&255)
	if 0.299*r+0.587*g+0.114*b < 140 {
		return "#f3f4f6"
	}
	return "#111827"
}
