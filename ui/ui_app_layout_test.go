package ui

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAppWithoutLayout verifies that the app works without a defined layout.
// When no layout is set, the page handler content is wrapped in a default content div.
func TestAppWithoutLayout(t *testing.T) {
	app := MakeApp("en")
	app.Page("/", "Home", func(ctx *Context) string {
		return Div("test-content", Attr{})("Test Content")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := app.Handler()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("Expected HTML response, got empty body")
	}

	// Verify essential elements are present
	if !strings.Contains(body, "Test Content") {
		t.Error("Expected page content to be present")
	}

	// Verify content is wrapped in ContentID div
	if !strings.Contains(body, app.ContentID.ID) {
		t.Error("Expected content div with ContentID to be present")
	}

	// Verify HTML structure
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Expected DOCTYPE declaration")
	}

	if !strings.Contains(body, "<title>Home</title>") {
		t.Error("Expected page title")
	}
}

// TestAppWithLayout verifies that the app still works with a defined layout.
func TestAppWithLayout(t *testing.T) {
	app := MakeApp("en")
	app.Layout(func(ctx *Context) string {
		return Div("layout-wrapper", Attr{})(
			Div("header", Attr{})("Header"),
			Div("", Attr{ID: "__content__"})(),
			Div("footer", Attr{})("Footer"),
		)
	})
	app.Page("/", "Home", func(ctx *Context) string {
		return Div("page-content", Attr{})("Page Content")
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler := app.Handler()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("Expected HTML response, got empty body")
	}

	// Verify layout elements are present
	if !strings.Contains(body, "Header") {
		t.Error("Expected layout header to be present")
	}

	if !strings.Contains(body, "Footer") {
		t.Error("Expected layout footer to be present")
	}

	// Verify HTML structure
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Expected DOCTYPE declaration")
	}

	if !strings.Contains(body, "<title>Home</title>") {
		t.Error("Expected page title")
	}
}

// TestAppWithoutLayout_MultiplePages verifies multiple pages work without layout
func TestAppWithoutLayout_MultiplePages(t *testing.T) {
	app := MakeApp("en")
	app.Page("/", "Home", func(ctx *Context) string {
		return Div("", Attr{})("Home Page")
	})
	app.Page("/about", "About", func(ctx *Context) string {
		return Div("", Attr{})("About Page")
	})

	// Test home page
	req1 := httptest.NewRequest("GET", "/", nil)
	w1 := httptest.NewRecorder()
	handler := app.Handler()
	handler.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("Expected status 200 for home, got %d", w1.Code)
	}
	if !strings.Contains(w1.Body.String(), "Home Page") {
		t.Error("Expected home page content")
	}
	if !strings.Contains(w1.Body.String(), "<title>Home</title>") {
		t.Error("Expected home page title")
	}

	// Test about page
	req2 := httptest.NewRequest("GET", "/about", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("Expected status 200 for about, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "About Page") {
		t.Error("Expected about page content")
	}
	if !strings.Contains(w2.Body.String(), "<title>About</title>") {
		t.Error("Expected about page title")
	}
}
