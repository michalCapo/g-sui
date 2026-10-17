package ui

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"compress/gzip"
)

// An asset URL holds the content hash and is cached for a year.
func TestAssetIsServedWithLongCache(t *testing.T) {
	app := NewApp()
	url := app.Asset("app.js", []byte(strings.Repeat("console.log(1);", 50)))
	if !strings.HasPrefix(url, "/__gsui/asset/") || !strings.HasSuffix(url, "/app.js") {
		t.Fatalf("asset url = %q", url)
	}
	if other := app.Asset("app.js", []byte("console.log(2)")); other == url {
		t.Fatal("changed content must get a new url")
	}
	req := httptest.NewRequest("GET", url, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, req)
	if rr.Code != 200 || rr.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("status %d, cache %q", rr.Code, rr.Header().Get("Cache-Control"))
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "javascript") || rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("headers %v", rr.Header())
	}
	gz, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := io.ReadAll(gz); !strings.HasPrefix(string(b), "console.log(1);") {
		t.Fatalf("body %q", b)
	}
	rr = httptest.NewRecorder()
	app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "/__gsui/asset/000000000000/app.js", nil))
	if rr.Code != 404 {
		t.Fatalf("unknown asset status %d", rr.Code)
	}
}

// A compiled stylesheet and a local icon font replace the CDN engine and
// Google Fonts.
func TestCompiledStylesReplaceCDN(t *testing.T) {
	app := NewApp()
	body := shellBody(t, app)
	for _, want := range []string{"@tailwindcss/browser@4", "fonts.googleapis.com/icon", `type="text/tailwindcss"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("default shell missing %q", want)
		}
	}
	app = NewApp()
	app.TailwindCSS([]byte(".flex{display:flex}"))
	app.IconFont([]byte("wOF2"))
	body = shellBody(t, app)
	for _, unwanted := range []string{"cdn.jsdelivr.net", "fonts.googleapis.com", "fonts.gstatic.com", "text/tailwindcss"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("compiled shell still has %q", unwanted)
		}
	}
	for _, want := range []string{`<link rel="stylesheet" href="/__gsui/asset/`, `as="font" type="font/woff2" crossorigin`, "font-family:'Material Icons Round'", "font-display:block"} {
		if !strings.Contains(body, want) {
			t.Errorf("compiled shell missing %q", want)
		}
	}
}

func TestAcceptsGzipHonoursQuality(t *testing.T) {
	for header, want := range map[string]bool{
		"gzip": true, "br, gzip": true, "gzip;q=0.5": true, "*": true, "": false,
		"gzip;q=0": false, "gzip;q=0.0, br": false, "br": false, "*;q=0": false, "*, gzip;q=0": false,
	} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", header)
		if got := acceptsGzip(req); got != want {
			t.Errorf("acceptsGzip(%q) = %v, want %v", header, got, want)
		}
	}
}
