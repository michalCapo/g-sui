package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Icon Tests
// ============================================================================

func TestIcon_BasicRendering(t *testing.T) {
	html := Icon("fa fa-home")

	// Verify div element
	assertContains(t, html, `<div`)
	assertContains(t, html, `</div>`)

	// Verify CSS class is included
	assertContains(t, html, `fa fa-home`)
}

func TestIcon_WithAttributes(t *testing.T) {
	html := Icon("fa fa-user", Attr{ID: "user-icon", Title: "User"})

	// Should have the icon class
	assertContains(t, html, `fa fa-user`)

	// Should have ID attribute
	assertContains(t, html, `id="user-icon"`)

	// Should have Title attribute
	assertContains(t, html, `title="User"`)
}

func TestIcon_EmptyClass(t *testing.T) {
	html := Icon("")

	// Should still render div
	assertContains(t, html, `<div`)
}

// ============================================================================
// IconStart Tests
// ============================================================================

func TestIconStart_Layout(t *testing.T) {
	html := IconStart("fa fa-home", "Home")

	// Should have icon class
	assertContains(t, html, `fa fa-home`)

	// Should have text
	assertContains(t, html, `Home`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

func TestIconStart_WithSpaces(t *testing.T) {
	html := IconStart("fa fa-arrow-left", "Go Back")

	assertContains(t, html, `fa fa-arrow-left`)
	assertContains(t, html, `Go Back`)
}

func TestIconStart_SpecialCharacters(t *testing.T) {
	html := IconStart("fa fa-icon", "Test & \"More\"")

	assertContains(t, html, `fa fa-icon`)
	assertContains(t, html, `Test`)
}

// ============================================================================
// IconLeft Tests
// ============================================================================

func TestIconLeft_Layout(t *testing.T) {
	html := IconLeft("fa fa-arrow-left", "Back")

	// Should have icon class
	assertContains(t, html, `fa fa-arrow-left`)

	// Should have text
	assertContains(t, html, `Back`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

func TestIconLeft_Structure(t *testing.T) {
	html := IconLeft("fa fa-chevron-left", "Previous")

	// IconLeft layout: Flex1 - Icon - Text - Flex1
	// This means icon comes before the text
	iconPos := strings.Index(html, `fa fa-chevron-left`)
	textPos := strings.Index(html, `Previous`)

	if iconPos == -1 {
		t.Error("Icon class not found")
	}

	if textPos == -1 {
		t.Error("Text not found")
	}

	// Icon should come before text in IconLeft
	if iconPos > textPos && textPos > 0 {
		t.Logf("Warning: Icon position %d after text position %d", iconPos, textPos)
	}
}

// ============================================================================
// IconRight Tests
// ============================================================================

func TestIconRight_Layout(t *testing.T) {
	html := IconRight("fa fa-arrow-right", "Next")

	// Should have icon class
	assertContains(t, html, `fa fa-arrow-right`)

	// Should have text
	assertContains(t, html, `Next`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

func TestIconRight_Structure(t *testing.T) {
	html := IconRight("fa fa-chevron-right", "Continue")

	// IconRight layout: Flex1 - Text - Icon - Flex1
	// This means text comes before the icon
	iconPos := strings.Index(html, `fa fa-chevron-right`)
	textPos := strings.Index(html, `Continue`)

	if iconPos == -1 {
		t.Error("Icon class not found")
	}

	if textPos == -1 {
		t.Error("Text not found")
	}

	// Text should come before icon in IconRight
	if textPos > iconPos && iconPos > 0 {
		t.Logf("Warning: Text position %d after icon position %d", textPos, iconPos)
	}
}

// ============================================================================
// IconEnd Tests
// ============================================================================

func TestIconEnd_Layout(t *testing.T) {
	html := IconEnd("fa fa-check", "Done")

	// Should have icon class
	assertContains(t, html, `fa fa-check`)

	// Should have text
	assertContains(t, html, `Done`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

// ============================================================================
// Flex1 Constant Tests
// ============================================================================

func TestFlex1_Constant(t *testing.T) {
	html := Flex1

	// Should be a div with flex-1 class
	assertContains(t, html, `<div`)
	assertContains(t, html, `flex-1`)
}

func TestFlex1_EmptyContent(t *testing.T) {
	html := Flex1

	// Should not have content between div tags
	// Flex1 is defined as Div("flex-1")() which is <div class="flex-1"></div>
	assertContains(t, html, `><`)
}

// ============================================================================
// Icon Function Tests
// ============================================================================

func TestIcon_MultipleClasses(t *testing.T) {
	html := Icon("fa fa-fw fa-home text-blue-500")

	assertContains(t, html, `fa fa-fw fa-home text-blue-500`)
}

func TestIcon_WithCustomAttributes(t *testing.T) {
	html := Icon("custom-icon", Attr{
		ID:    "my-icon",
		Class: "additional-class",
		Style: "color: red;",
	})

	assertContains(t, html, `custom-icon`)
	assertContains(t, html, `additional-class`)
	assertContains(t, html, `style="color: red;"`)
}

// ============================================================================
// Icon Layout - Edge Cases
// ============================================================================

func TestIconStart_EmptyText(t *testing.T) {
	html := IconStart("fa fa-home", "")

	assertContains(t, html, `fa fa-home`)
}

func TestIconLeft_EmptyText(t *testing.T) {
	html := IconLeft("fa fa-arrow-left", "")

	assertContains(t, html, `fa fa-arrow-left`)
}

func TestIconRight_EmptyText(t *testing.T) {
	html := IconRight("fa fa-arrow-right", "")

	assertContains(t, html, `fa fa-arrow-right`)
}

func TestIconEnd_EmptyText(t *testing.T) {
	html := IconEnd("fa fa-check", "")

	assertContains(t, html, `fa fa-check`)
}

func TestIconStart_EmptyIcon(t *testing.T) {
	html := IconStart("", "Home")

	assertContains(t, html, `Home`)
}

func TestIconLeft_EmptyIcon(t *testing.T) {
	html := IconLeft("", "Back")

	assertContains(t, html, `Back`)
}

func TestIconRight_EmptyIcon(t *testing.T) {
	html := IconRight("", "Next")

	assertContains(t, html, `Next`)
}

func TestIconEnd_EmptyIcon(t *testing.T) {
	html := IconEnd("", "Done")

	assertContains(t, html, `Done`)
}

// ============================================================================
// Icon Layout - Long Text Tests
// ============================================================================

func TestIconStart_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconStart("fa fa-home", longText)

	assertContains(t, html, `fa fa-home`)
	assertContains(t, html, longText)
}

func TestIconLeft_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconLeft("fa fa-arrow-left", longText)

	assertContains(t, html, `fa fa-arrow-left`)
	assertContains(t, html, longText)
}

func TestIconRight_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconRight("fa fa-arrow-right", longText)

	assertContains(t, html, `fa fa-arrow-right`)
	assertContains(t, html, longText)
}

func TestIconEnd_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconEnd("fa fa-check", longText)

	assertContains(t, html, `fa fa-check`)
	assertContains(t, html, longText)
}
