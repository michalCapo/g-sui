package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPWA(t *testing.T) {
	app := MakeApp("en")
	config := PWAConfig{
		Name:                  "Test App",
		ShortName:             "Test",
		ThemeColor:            "#ff0000",
		BackgroundColor:       "#ffffff",
		GenerateServiceWorker: true,
	}

	app.PWA(config)

	// Test Manifest Generation
	if app.pwaManifest == nil {
		t.Fatal("pwaManifest should not be nil")
	}

	var manifest PWAConfig
	err := json.Unmarshal(app.pwaManifest, &manifest)
	if err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if manifest.Name != config.Name {
		t.Errorf("Expected name %s, got %s", config.Name, manifest.Name)
	}
	if manifest.Display != "standalone" {
		t.Errorf("Expected default display 'standalone', got %s", manifest.Display)
	}

	// Test HTML Head Tags
	head := strings.Join(app.HTMLHead, "")
	if !strings.Contains(head, `link rel="manifest" href="/manifest.webmanifest"`) {
		t.Error("HTML head missing manifest link")
	}
	if !strings.Contains(head, `meta name="theme-color" content="#ff0000"`) {
		t.Error("HTML head missing theme-color meta")
	}
	if !strings.Contains(head, "navigator.serviceWorker.register('/sw.js')") {
		t.Error("HTML head missing service worker registration")
	}

	// Test Manifest Endpoint using app's TestHandler()
	handler := app.TestHandler()
	req := httptest.NewRequest("GET", "/manifest.webmanifest", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK for manifest, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/manifest+json" {
		t.Errorf("Expected content-type application/manifest+json, got %s", w.Header().Get("Content-Type"))
	}

	// Test Service Worker Endpoint
	reqSW := httptest.NewRequest("GET", "/sw.js", nil)
	wSW := httptest.NewRecorder()
	handler.ServeHTTP(wSW, reqSW)

	if wSW.Code != http.StatusOK {
		t.Errorf("Expected status OK for sw.js, got %d", wSW.Code)
	}
	if wSW.Header().Get("Content-Type") != "application/javascript" {
		t.Errorf("Expected content-type application/javascript, got %s", wSW.Header().Get("Content-Type"))
	}
	if !strings.Contains(wSW.Body.String(), "CACHE_NAME") {
		t.Error("Service worker body missing expected content")
	}
}
