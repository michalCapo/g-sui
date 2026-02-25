package ui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMultiInstanceSeparatePorts verifies that two App instances can each
// register routes and serve their own content without interfering.
func TestMultiInstanceSeparatePorts(t *testing.T) {
	app1 := MakeApp("en")
	app1.Page("/", "App1 Home", func(ctx *Context) string {
		return Div("", Attr{})("app1-content")
	})

	app2 := MakeApp("en")
	app2.Page("/", "App2 Home", func(ctx *Context) string {
		return Div("", Attr{})("app2-content")
	})

	// Both should produce handlers without panic
	h1 := app1.Handler()
	h2 := app2.Handler()

	srv1 := httptest.NewServer(h1)
	defer srv1.Close()
	srv2 := httptest.NewServer(h2)
	defer srv2.Close()

	// Verify app1 serves its own content
	resp1, err := http.Get(srv1.URL + "/")
	if err != nil {
		t.Fatalf("app1 GET /: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	if resp1.StatusCode != 200 {
		t.Errorf("app1: expected 200, got %d", resp1.StatusCode)
	}
	if !strings.Contains(string(body1), "app1-content") {
		t.Error("app1: expected body to contain 'app1-content'")
	}
	if strings.Contains(string(body1), "app2-content") {
		t.Error("app1: body should NOT contain 'app2-content'")
	}

	// Verify app2 serves its own content
	resp2, err := http.Get(srv2.URL + "/")
	if err != nil {
		t.Fatalf("app2 GET /: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	if resp2.StatusCode != 200 {
		t.Errorf("app2: expected 200, got %d", resp2.StatusCode)
	}
	if !strings.Contains(string(body2), "app2-content") {
		t.Error("app2: expected body to contain 'app2-content'")
	}
	if strings.Contains(string(body2), "app1-content") {
		t.Error("app2: body should NOT contain 'app1-content'")
	}
}

// TestMultiInstanceContentIDIsolation verifies that each App gets its own
// ContentID, so body element IDs don't collide.
func TestMultiInstanceContentIDIsolation(t *testing.T) {
	app1 := MakeApp("en")
	app2 := MakeApp("en")

	if app1.ContentID.ID == app2.ContentID.ID {
		t.Errorf("expected different ContentIDs, both got %q", app1.ContentID.ID)
	}
}

// TestMultiInstanceSharedMux verifies that two apps can be mounted on a shared
// mux under different path prefixes and serve their own content.
func TestMultiInstanceSharedMux(t *testing.T) {
	mux := http.NewServeMux()

	app1 := MakeApp("en")
	app1.Page("/", "Admin", func(ctx *Context) string {
		return Div("", Attr{})("admin-content")
	})
	app1.Mount("/admin", mux)

	app2 := MakeApp("en")
	app2.Page("/", "Portal", func(ctx *Context) string {
		return Div("", Attr{})("portal-content")
	})
	app2.Mount("/portal", mux)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Verify /admin/ returns admin content
	resp1, err := http.Get(srv.URL + "/admin/")
	if err != nil {
		t.Fatalf("GET /admin/: %v", err)
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	if resp1.StatusCode != 200 {
		t.Errorf("/admin/: expected 200, got %d", resp1.StatusCode)
	}
	if !strings.Contains(string(body1), "admin-content") {
		t.Errorf("/admin/: expected 'admin-content' in body, got: %s", string(body1))
	}

	// Verify /portal/ returns portal content
	resp2, err := http.Get(srv.URL + "/portal/")
	if err != nil {
		t.Fatalf("GET /portal/: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	if resp2.StatusCode != 200 {
		t.Errorf("/portal/: expected 200, got %d", resp2.StatusCode)
	}
	if !strings.Contains(string(body2), "portal-content") {
		t.Errorf("/portal/: expected 'portal-content' in body, got: %s", string(body2))
	}
}

// TestMultiInstanceStoredIsolation verifies that actions registered on one app
// do not leak into another.
func TestMultiInstanceStoredIsolation(t *testing.T) {
	app1 := MakeApp("en")
	app2 := MakeApp("en")

	handler1 := func(ctx *Context) string { return "action1" }
	handler2 := func(ctx *Context) string { return "action2" }

	app1.Action("/test-action", handler1)
	app2.Action("/other-action", handler2)

	// app1 should have /test-action but NOT /other-action
	app1.storedMu.Lock()
	foundInApp1 := false
	leakedToApp1 := false
	for _, path := range app1.stored {
		if path == "/test-action" {
			foundInApp1 = true
		}
		if path == "/other-action" {
			leakedToApp1 = true
		}
	}
	app1.storedMu.Unlock()

	if !foundInApp1 {
		t.Error("app1: expected /test-action to be registered")
	}
	if leakedToApp1 {
		t.Error("app1: /other-action should NOT be registered (leaked from app2)")
	}

	// app2 should have /other-action but NOT /test-action
	app2.storedMu.Lock()
	foundInApp2 := false
	leakedToApp2 := false
	for _, path := range app2.stored {
		if path == "/other-action" {
			foundInApp2 = true
		}
		if path == "/test-action" {
			leakedToApp2 = true
		}
	}
	app2.storedMu.Unlock()

	if !foundInApp2 {
		t.Error("app2: expected /other-action to be registered")
	}
	if leakedToApp2 {
		t.Error("app2: /test-action should NOT be registered (leaked from app1)")
	}
}
