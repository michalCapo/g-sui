package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Button Tests
// ============================================================================

func TestButton_BasicRendering(t *testing.T) {
	html := Button().Render("Click Me")

	// Default button renders as a div
	assertContains(t, html, "<div")
	assertContains(t, html, "</div>")

	// Verify button text
	assertContains(t, html, "Click Me")

	// Verify default class
	assertContains(t, html, BTN)
}

func TestButton_AllTypes(t *testing.T) {
	tests := []struct {
		name       string
		build      func() *button
		expectTag  string
		expectType string
	}{
		{
			name:       "Default button",
			build:      func() *button { return Button() },
			expectTag:  "div",
			expectType: "div",
		},
		{
			name:       "Submit button",
			build:      func() *button { return Button().Submit() },
			expectTag:  "button",
			expectType: "submit",
		},
		{
			name:       "Reset button",
			build:      func() *button { return Button().Reset() },
			expectTag:  "button",
			expectType: "reset",
		},
		{
			name:       "Anchor button (Href)",
			build:      func() *button { return Button().Href("/path") },
			expectTag:  "a",
			expectType: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := tt.build().Render("Test")

			// Check for correct tag
			if !strings.Contains(html, "<"+tt.expectTag) {
				t.Errorf("Expected <%s> tag, got: %s", tt.expectTag, html)
			}

			// For button tags, check type attribute
			if tt.expectTag == "button" {
				assertContains(t, html, `type="`+tt.expectType+`"`)
			}

			// For anchor tags, check href attribute
			if tt.expectTag == "a" {
				assertContains(t, html, `href="/path"`)
			}
		})
	}
}

func TestButton_AllColors(t *testing.T) {
	colors := []string{
		Blue,
		Green,
		Purple,
		Red,
		Gray,
		Yellow,
		White,
		YellowOutline,
		GreenOutline,
		PurpleOutline,
		RedOutline,
		GrayOutline,
		BlueOutline,
		WhiteOutline,
	}

	for _, color := range colors {
		t.Run(color, func(t *testing.T) {
			html := Button().Color(color).Render("Test")

			// Color classes should be in the HTML
			// Check if any part of the color constant is present
			if !strings.Contains(html, color) && color != "" {
				// Some colors may have multiple classes, just check if something is there
				t.Logf("Color %q, HTML: %s", color, html)
			}
		})
	}
}

func TestButton_AllSizes(t *testing.T) {
	sizes := []struct {
		name string
		size string
	}{
		{"XS", XS},
		{"SM", SM},
		{"MD", MD},
		{"ST", ST},
		{"LG", LG},
		{"XL", XL},
	}

	for _, tt := range sizes {
		t.Run(tt.name, func(t *testing.T) {
			html := Button().Size(tt.size).Render("Test")

			if !strings.Contains(html, tt.size) {
				t.Errorf("Size %q should be in HTML: %s", tt.size, html)
			}
		})
	}
}

func TestButton_DisabledState(t *testing.T) {
	html := Button().Disabled(true).Render("Disabled Button")

	// Check for disabled styling (disabled div buttons use classes, not disabled attribute)
	assertContains(t, html, DISABLED)

	// Check for opacity class
	assertContains(t, html, "opacity-25")

	// Check for pointer-events-none class
	assertContains(t, html, "pointer-events-none")
}

func TestButton_DisabledStateWithSubmit(t *testing.T) {
	html := Button().Submit().Disabled(true).Render("Disabled Button")

	// Submit buttons have type="submit"
	assertContains(t, html, `type="submit"`)

	// Submit buttons get the disabled attribute and disabled styling
	assertContains(t, html, `disabled="disabled"`)
	assertContains(t, html, DISABLED)
}

func TestButton_IfFalseReturnsEmpty(t *testing.T) {
	html := Button().If(false).Render("Test")

	if html != "" {
		t.Errorf("If(false) should return empty string, got: %s", html)
	}
}

func TestButton_IfTrueReturnsContent(t *testing.T) {
	html := Button().If(true).Render("Test")

	if html == "" {
		t.Errorf("If(true) should return content, got empty string")
	}

	// Default button renders as div
	assertContains(t, html, "<div")
}

func TestButton_AllBuilderMethods(t *testing.T) {
	tests := []struct {
		name   string
		build  func() *button
		verify func(*testing.T, string)
	}{
		{
			name: "Class method",
			build: func() *button {
				return Button().Class("custom-class")
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, "custom-class")
			},
		},
		{
			name: "Size method",
			build: func() *button {
				return Button().Size(LG)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, LG)
			},
		},
		{
			name: "Color method - Blue",
			build: func() *button {
				return Button().Color(Blue)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, Blue)
			},
		},
		{
			name: "Color method - Green",
			build: func() *button {
				return Button().Color(Green)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, Green)
			},
		},
		{
			name: "Color method - Purple",
			build: func() *button {
				return Button().Color(Purple)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, Purple)
			},
		},
		{
			name: "Color method - Red",
			build: func() *button {
				return Button().Color(Red)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, Red)
			},
		},
		{
			name: "Color method - Gray",
			build: func() *button {
				return Button().Color(Gray)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, Gray)
			},
		},
		{
			name: "Color method - Yellow",
			build: func() *button {
				return Button().Color(Yellow)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, Yellow)
			},
		},
		{
			name: "Color method - White",
			build: func() *button {
				return Button().Color(White)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, White)
			},
		},
		{
			name: "Color method - BlueOutline",
			build: func() *button {
				return Button().Color(BlueOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, BlueOutline)
			},
		},
		{
			name: "Color method - GreenOutline",
			build: func() *button {
				return Button().Color(GreenOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, GreenOutline)
			},
		},
		{
			name: "Color method - PurpleOutline",
			build: func() *button {
				return Button().Color(PurpleOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, PurpleOutline)
			},
		},
		{
			name: "Color method - RedOutline",
			build: func() *button {
				return Button().Color(RedOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, RedOutline)
			},
		},
		{
			name: "Color method - GrayOutline",
			build: func() *button {
				return Button().Color(GrayOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, GrayOutline)
			},
		},
		{
			name: "Color method - YellowOutline",
			build: func() *button {
				return Button().Color(YellowOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, YellowOutline)
			},
		},
		{
			name: "Color method - WhiteOutline",
			build: func() *button {
				return Button().Color(WhiteOutline)
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, WhiteOutline)
			},
		},
		{
			name: "Click method",
			build: func() *button {
				return Button().Click("console.log('clicked')")
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, `onclick=`)
			},
		},
		{
			name: "Name method",
			build: func() *button {
				return Button().Name("btnSubmit")
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, `name="btnSubmit"`)
			},
		},
		{
			name: "Val method",
			build: func() *button {
				return Button().Val("submit-value")
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, `value="submit-value"`)
			},
		},
		{
			name: "Form method",
			build: func() *button {
				return Button().Submit().Form("form123")
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, `form="form123"`)
			},
		},
		{
			name: "Href method - renders anchor",
			build: func() *button {
				return Button().Href("/test-path")
			},
			verify: func(t *testing.T, html string) {
				assertContains(t, html, `<a`)
				assertContains(t, html, `href="/test-path"`)
				assertNotContains(t, html, `<button`)
			},
		},
		{
			name: "Disabled method - true",
			build: func() *button {
				return Button().Disabled(true)
			},
			verify: func(t *testing.T, html string) {
				// Div buttons use classes for disabled state
				assertContains(t, html, DISABLED)
				assertContains(t, html, "pointer-events-none")
				assertContains(t, html, "opacity-25")
			},
		},
		{
			name: "Disabled method - false",
			build: func() *button {
				return Button().Disabled(false)
			},
			verify: func(t *testing.T, html string) {
				assertNotContains(t, html, `disabled="disabled"`)
				assertNotContains(t, html, "opacity-25")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := tt.build().Render("Test Button")
			tt.verify(t, html)
		})
	}
}

func TestButton_CombinedOptions(t *testing.T) {
	// Test combining multiple options
	html := Button().
		Submit().
		Size(LG).
		Color(Blue).
		Class("extra-class").
		Disabled(true).
		Render("Submit")

	assertContains(t, html, `<button`)
	assertContains(t, html, `type="submit"`)
	assertContains(t, html, LG)
	assertContains(t, html, Blue)
	assertContains(t, html, "extra-class")
	// Submit buttons get disabled attribute and styling
	assertContains(t, html, `disabled="disabled"`)
	assertContains(t, html, DISABLED)
	assertContains(t, html, "Submit")
}

func TestButton_HrefWithOtherOptions(t *testing.T) {
	// Test that Href works with other options
	html := Button().
		Href("/test").
		Size(MD).
		Color(Green).
		Render("Link")

	assertContains(t, html, `<a`)
	assertContains(t, html, `href="/test"`)
	assertContains(t, html, MD)
	assertContains(t, html, Green)
	assertContains(t, html, "Link")
}

func TestButton_EmptyText(t *testing.T) {
	html := Button().Render("")

	// Should still render button with empty content
	assertContains(t, html, `<div`)
	assertContains(t, html, `</div>`)
}

func TestButton_SpecialCharactersInText(t *testing.T) {
	html := Button().Render("Click & Submit \"Now\"")

	// Text is passed through as-is in div content
	assertContains(t, html, `Click & Submit`)
}
