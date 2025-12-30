package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Label Tests
// ============================================================================

func TestLabel_BasicRendering(t *testing.T) {
	target := Target()
	html := Label(&target).Render("Username")

	// Verify label element
	assertContains(t, html, `<label`)
	assertContains(t, html, `</label>`)

	// Verify for attribute matches target ID
	assertContains(t, html, `for="`+target.ID+`"`)

	// Verify label text
	assertContains(t, html, "Username")

	// Verify default class
	assertContains(t, html, `text-sm`)
}

func TestLabel_Required(t *testing.T) {
	target := Target()
	html := Label(&target).Required(true).Render("Username")

	// Should contain span for asterisk
	assertContains(t, html, `<span`)

	// Should contain asterisk
	assertContains(t, html, `*`)

	// Should have red color class
	assertContains(t, html, `text-red-700`)
}

func TestLabel_RequiredFalse(t *testing.T) {
	target := Target()
	html := Label(&target).Required(false).Render("Username")

	// Should NOT contain asterisk
	assertNotContains(t, html, `*`)

	// Should NOT have red color class
	assertNotContains(t, html, `text-red-700`)
}

func TestLabel_RequiredNotShownWhenDisabled(t *testing.T) {
	target := Target()
	html := Label(&target).Required(true).Disabled(true).Render("Username")

	// When disabled, the asterisk should not appear
	assertNotContains(t, html, `text-red-700`)
	assertNotContains(t, html, `*`)
}

func TestLabel_RequiredShownWhenNotDisabled(t *testing.T) {
	target := Target()
	html := Label(&target).Required(true).Disabled(false).Render("Username")

	// When not disabled, the asterisk should appear
	assertContains(t, html, `text-red-700`)
	assertContains(t, html, `*`)
}

func TestLabel_Disabled(t *testing.T) {
	target := Target()
	html := Label(&target).Disabled(true).Render("Username")

	// Should render normally (disabled affects required indicator)
	assertContains(t, html, `<label`)
	assertContains(t, html, "Username")
}

func TestLabel_ClassMethods(t *testing.T) {
	target := Target()

	// Test Class method
	html := Label(&target).Class("custom-class").Render("Username")
	assertContains(t, html, "custom-class")

	// Test ClassLabel method
	html2 := Label(&target).ClassLabel("label-custom").Render("Username")
	assertContains(t, html2, "label-custom")
}

func TestLabel_MultipleClassMethods(t *testing.T) {
	target := Target()
	html := Label(&target).
		Class("custom-class1", "custom-class2").
		ClassLabel("label-custom1", "label-custom2").
		Render("Username")

	assertContains(t, html, "custom-class1")
	assertContains(t, html, "custom-class2")
	assertContains(t, html, "label-custom1")
	assertContains(t, html, "label-custom2")
}

func TestLabel_EmptyTextReturnsEmpty(t *testing.T) {
	target := Target()
	html := Label(&target).Render("")

	if html != "" {
		t.Errorf("Empty label text should return empty string, got: %s", html)
	}
}

func TestLabel_WithNilTarget(t *testing.T) {
	html := Label(nil).Render("Username")

	// Should still render label even with nil target
	assertContains(t, html, `<label`)
	assertContains(t, html, "Username")

	// Should not have for attribute
	assertNotContains(t, html, `for=`)
}

func TestLabel_SpecialCharactersInText(t *testing.T) {
	target := Target()
	html := Label(&target).Render("User's Name & More")

	// Should contain the text (labels don't escape HTML by design for label text)
	assertContains(t, html, `User's`)
	assertContains(t, html, `Name`)
	assertContains(t, html, `More`)
}

func TestLabel_DefaultClasses(t *testing.T) {
	target := Target()
	html := Label(&target).Render("Test")

	// Should have default text-sm class
	assertContains(t, html, `text-sm`)

	// Should have relative class for positioning
	assertContains(t, html, `relative`)
}

func TestLabel_LongText(t *testing.T) {
	target := Target()
	longText := strings.Repeat("A very long label text ", 10)
	html := Label(&target).Render(longText)

	assertContains(t, html, longText)
}
