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

func TestPWA_ID(t *testing.T) {
	app := MakeApp("en")

	t.Run("ID defaults to StartURL when empty", func(t *testing.T) {
		config := PWAConfig{
			Name:     "Test App",
			StartURL: "/app",
		}
		app.PWA(config)

		var manifest PWAConfig
		err := json.Unmarshal(app.pwaManifest, &manifest)
		if err != nil {
			t.Fatalf("Failed to unmarshal manifest: %v", err)
		}

		if manifest.ID != "/app" {
			t.Errorf("Expected ID to default to '/app', got %s", manifest.ID)
		}
	})

	t.Run("ID uses provided value", func(t *testing.T) {
		app2 := MakeApp("en")
		config := PWAConfig{
			Name:     "Test App",
			ID:       "/custom-id",
			StartURL: "/app",
		}
		app2.PWA(config)

		var manifest PWAConfig
		err := json.Unmarshal(app2.pwaManifest, &manifest)
		if err != nil {
			t.Fatalf("Failed to unmarshal manifest: %v", err)
		}

		if manifest.ID != "/custom-id" {
			t.Errorf("Expected ID to be '/custom-id', got %s", manifest.ID)
		}
	})

	t.Run("ID defaults to '/' when both ID and StartURL are empty", func(t *testing.T) {
		app3 := MakeApp("en")
		config := PWAConfig{
			Name: "Test App",
		}
		app3.PWA(config)

		var manifest PWAConfig
		err := json.Unmarshal(app3.pwaManifest, &manifest)
		if err != nil {
			t.Fatalf("Failed to unmarshal manifest: %v", err)
		}

		if manifest.ID != "/" {
			t.Errorf("Expected ID to default to '/', got %s", manifest.ID)
		}
	})
}

func TestPWA_IconPurpose(t *testing.T) {
	app := MakeApp("en")
	config := PWAConfig{
		Name:  "Test App",
		Icons: []PWAIcon{
			{Src: "/icon-192.png", Sizes: "192x192", Type: "image/png", Purpose: "any"},
			{Src: "/icon-512.png", Sizes: "512x512", Type: "image/png", Purpose: "any maskable"},
		},
	}

	app.PWA(config)

	var manifest PWAConfig
	err := json.Unmarshal(app.pwaManifest, &manifest)
	if err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if len(manifest.Icons) != 2 {
		t.Fatalf("Expected 2 icons, got %d", len(manifest.Icons))
	}

	if manifest.Icons[0].Purpose != "any" {
		t.Errorf("Expected first icon purpose 'any', got %s", manifest.Icons[0].Purpose)
	}

	if manifest.Icons[1].Purpose != "any maskable" {
		t.Errorf("Expected second icon purpose 'any maskable', got %s", manifest.Icons[1].Purpose)
	}
}

func TestPWA_IconPurpose_OmittedWhenEmpty(t *testing.T) {
	app := MakeApp("en")
	config := PWAConfig{
		Name:  "Test App",
		Icons: []PWAIcon{
			{Src: "/icon.png", Sizes: "192x192", Type: "image/png"},
		},
	}

	app.PWA(config)

	// Parse raw JSON to verify purpose field is not present when empty
	var rawManifest map[string]interface{}
	err := json.Unmarshal(app.pwaManifest, &rawManifest)
	if err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	icons, ok := rawManifest["icons"].([]interface{})
	if !ok || len(icons) == 0 {
		t.Fatal("No icons found in manifest")
	}

	firstIcon, ok := icons[0].(map[string]interface{})
	if !ok {
		t.Fatal("First icon is not an object")
	}

	if _, hasPurpose := firstIcon["purpose"]; hasPurpose {
		t.Error("Purpose field should be omitted when empty")
	}
}
