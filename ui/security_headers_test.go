package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	get := func(app *App, path string) http.Header {
		t.Helper()
		rr := httptest.NewRecorder()
		app.Handler().ServeHTTP(rr, httptest.NewRequest("GET", "http://example.test"+path, nil))
		return rr.Header()
	}
	page := func(_ *Context) *Node { return Div().Text("page") }

	app := NewApp()
	app.Page("/", page)
	for _, path := range []string{"/", "/__ws.js", "/missing"} {
		h := get(app, path)
		if got := h.Get("X-Frame-Options"); got != "SAMEORIGIN" {
			t.Fatalf("%s: X-Frame-Options = %q, want SAMEORIGIN", path, got)
		}
		if got := h.Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("%s: X-Content-Type-Options = %q, want nosniff", path, got)
		}
		if got := h.Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
			t.Fatalf("%s: Referrer-Policy = %q", path, got)
		}
	}

	deny := NewApp()
	deny.FrameOptions = "DENY"
	deny.Page("/", page)
	if got := get(deny, "/").Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("FrameOptions DENY: got %q", got)
	}

	none := NewApp()
	none.FrameOptions = "-"
	none.Page("/", page)
	if got := get(none, "/").Values("X-Frame-Options"); len(got) != 0 {
		t.Fatalf("FrameOptions -: got %q, want no header", got)
	}

	custom := NewApp()
	custom.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Del("Referrer-Policy")
			next.ServeHTTP(w, r)
		})
	})
	custom.Page("/", page)
	h := get(custom, "/")
	if got := h.Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("middleware override: X-Frame-Options = %q, want DENY", got)
	}
	if got := h.Values("Referrer-Policy"); len(got) != 0 {
		t.Fatalf("middleware delete: Referrer-Policy = %q, want none", got)
	}

	// Stricter values set in front of Handler are kept, and an empty
	// value suppresses the default.
	outer := NewApp()
	outer.Page("/", page)
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header()["X-Content-Type-Options"] = nil
		outer.Handler().ServeHTTP(w, r)
	})
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, httptest.NewRequest("GET", "http://example.test/", nil))
	if got := rr.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("outer middleware: X-Frame-Options = %q, want DENY", got)
	}
	if got := rr.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("outer middleware: Referrer-Policy = %q, want no-referrer", got)
	}
	if got := rr.Header().Values("X-Content-Type-Options"); len(got) != 0 {
		t.Fatalf("outer middleware: X-Content-Type-Options = %q, want none", got)
	}
}
