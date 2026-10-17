package ui

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"
)

// assetFile is a file served by Asset.
type assetFile struct {
	data, gzipped []byte
	contentType   string
}

const assetPrefix = "/__gsui/asset/"

// Asset serves data at a URL that holds its content hash and returns that
// URL. Browsers cache it for a year: a changed file gets a new URL, so no
// version needs bumping. Use it for scripts, stylesheets and fonts a page
// links to. name gives the file name and content type. Register assets
// before serving requests.
func (app *App) Asset(name string, data []byte) string {
	name = path.Base("/" + name)
	sum := sha256.Sum256(data)
	key := hex.EncodeToString(sum[:6]) + "/" + name
	file := &assetFile{data: data, contentType: mime.TypeByExtension(path.Ext(name))}
	if file.contentType == "" {
		file.contentType = http.DetectContentType(data)
	}
	if compressible(file.contentType) {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, _ = gz.Write(data)
		_ = gz.Close()
		if b.Len() < len(data) {
			file.gzipped = b.Bytes()
		}
	}
	app.mu.Lock()
	defer app.mu.Unlock()
	if app.assets == nil {
		app.assets = make(map[string]*assetFile)
	}
	app.assets[key] = file
	return assetPrefix + key
}

func compressible(contentType string) bool {
	return strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "javascript") ||
		strings.Contains(contentType, "json") || strings.Contains(contentType, "svg")
}

func (app *App) serveAsset(w http.ResponseWriter, r *http.Request) {
	app.mu.RLock()
	file := app.assets[strings.TrimPrefix(r.URL.Path, assetPrefix)]
	app.mu.RUnlock()
	if file == nil {
		http.NotFound(w, r)
		return
	}
	h := w.Header()
	h.Set("Content-Type", file.contentType)
	h.Set("Cache-Control", "public, max-age=31536000, immutable")
	h.Del("Pragma")
	h.Del("Expires")
	h.Add("Vary", "Accept-Encoding")
	body := file.data
	if file.gzipped != nil && acceptsGzip(r) {
		h.Set("Content-Encoding", "gzip")
		body = file.gzipped
	}
	h.Set("Content-Length", fmt.Sprint(len(body)))
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

// TailwindCSS makes pages link a compiled Tailwind stylesheet, served with
// Asset, instead of compiling classes in the browser with the Tailwind
// engine from a CDN. Pages then paint as soon as the stylesheet is in, with
// no hidden start. css must hold every class the pages use, those of g-sui
// components included (scan the g-sui ui directory as a Tailwind source),
// and the class based dark variant:
//
//	@custom-variant dark (&:where(.dark, .dark *));
func (app *App) TailwindCSS(css []byte) {
	url := app.Asset("app.css", css)
	app.mu.Lock()
	app.tailwindURL = url
	app.mu.Unlock()
}

// IconFont serves the Material Icons Round font (woff2) used by Icon with
// the app instead of loading it from Google Fonts. Icons stay blank, not
// spelled out, until the font is in; it is preloaded.
func (app *App) IconFont(woff2 []byte) {
	url := app.Asset("material-icons-round.woff2", woff2)
	app.mu.Lock()
	app.iconFontURL = url
	app.mu.Unlock()
}

// styleHeadHTML returns the head tags that style pages: the Tailwind engine
// and Google's icon font, or the app's own compiled stylesheet and font.
func (app *App) styleHeadHTML() string {
	app.mu.RLock()
	tailwindURL, iconURL := app.tailwindURL, app.iconFontURL
	app.mu.RUnlock()
	var b strings.Builder
	if tailwindURL == "" || iconURL == "" {
		b.WriteString(`<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
`)
	}
	if tailwindURL == "" {
		b.WriteString(`<link rel="preconnect" href="https://cdn.jsdelivr.net" crossorigin>
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4" data-gsui-style-engine async onload="this.dataset.gsuiLoaded='true'" onerror="this.dataset.gsuiLoaded='error'"></script>
`)
	} else {
		fmt.Fprintf(&b, "<link rel=\"stylesheet\" href=\"%s\">\n", tailwindURL)
	}
	if iconURL == "" {
		b.WriteString(`<link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons+Round" media="print" onload="this.media='all';this.dataset.gsuiLoaded='true'" onerror="this.dataset.gsuiLoaded='error'">
<noscript><link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons+Round"></noscript>
`)
	} else {
		fmt.Fprintf(&b, `<link rel="preload" href="%s" as="font" type="font/woff2" crossorigin>
<style>@font-face{font-family:'Material Icons Round';font-style:normal;font-weight:400;font-display:block;src:url(%s) format('woff2')}.material-icons-round{font-family:'Material Icons Round';font-weight:normal;font-style:normal;font-size:24px;line-height:1;letter-spacing:normal;text-transform:none;display:inline-block;white-space:nowrap;word-wrap:normal;direction:ltr;-webkit-font-feature-settings:'liga';font-feature-settings:'liga';-webkit-font-smoothing:antialiased}</style>
`, iconURL, iconURL)
	}
	if tailwindURL == "" {
		b.WriteString(`<style type="text/tailwindcss">
@custom-variant dark (&:where(.dark, .dark *));
@theme{--font-sans:ui-sans-serif,system-ui,-apple-system,'Segoe UI',sans-serif;}
</style>
`)
	}
	return b.String()
}

// acceptsGzip reports whether the request accepts a gzip body. "gzip;q=0"
// refuses it.
func acceptsGzip(r *http.Request) bool {
	gzipQ, starQ := "", ""
	for part := range strings.SplitSeq(r.Header.Get("Accept-Encoding"), ",") {
		name, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		q := "1"
		for param := range strings.SplitSeq(params, ";") {
			if key, value, ok := strings.Cut(strings.TrimSpace(param), "="); ok && strings.EqualFold(strings.TrimSpace(key), "q") {
				q = strings.TrimSpace(value)
			}
		}
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "gzip", "x-gzip":
			gzipQ = q
		case "*":
			starQ = q
		}
	}
	if gzipQ == "" {
		gzipQ = starQ
	}
	return gzipQ != "" && strings.Trim(gzipQ, "0.") != ""
}
