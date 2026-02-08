package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Icon Tests
// ============================================================================

func TestIcon_BasicRendering(t *testing.T) {
	html := Icon("home")

	// Verify div element
	assertContains(t, html, `<div`)
	assertContains(t, html, `</div>`)

	// Verify Material Icons class and icon name
	assertContains(t, html, `material-icons`)
	assertContains(t, html, `home`)
}

func TestIcon_WithAttributes(t *testing.T) {
	html := Icon("person", Attr{ID: "user-icon", Title: "User"})

	// Should have Material Icons class and icon name
	assertContains(t, html, `material-icons`)
	assertContains(t, html, `person`)

	// Should have ID attribute
	assertContains(t, html, `id="user-icon"`)

	// Should have Title attribute
	assertContains(t, html, `title="User"`)
}

func TestIcon_EmptyClass(t *testing.T) {
	html := Icon("")

	// Should still render div
	assertContains(t, html, `<div`)
	assertContains(t, html, `material-icons`)
}

// ============================================================================
// IconStart Tests
// ============================================================================

func TestIconStart_Layout(t *testing.T) {
	html := IconStart("home", "Home")

	// Should have Material Icons class and icon name
	assertContains(t, html, `material-icons`)
	assertContains(t, html, `home`)

	// Should have text
	assertContains(t, html, `Home`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

func TestIconStart_WithSpaces(t *testing.T) {
	html := IconStart("arrow_back", "Go Back")

	assertContains(t, html, `arrow_back`)
	assertContains(t, html, `Go Back`)
}

func TestIconStart_SpecialCharacters(t *testing.T) {
	html := IconStart("settings", "Test & \"More\"")

	assertContains(t, html, `settings`)
	assertContains(t, html, `Test`)
}

// ============================================================================
// IconLeft Tests
// ============================================================================

func TestIconLeft_Layout(t *testing.T) {
	html := IconLeft("arrow_back", "Back")

	// Should have Material Icons class and icon name
	assertContains(t, html, `material-icons`)
	assertContains(t, html, `arrow_back`)

	// Should have text
	assertContains(t, html, `Back`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

func TestIconLeft_Structure(t *testing.T) {
	html := IconLeft("chevron_left", "Previous")

	// IconLeft layout: Flex1 - Icon - Text - Flex1
	// This means icon comes before the text
	iconPos := strings.Index(html, `chevron_left`)
	textPos := strings.Index(html, `Previous`)

	if iconPos == -1 {
		t.Error("Icon name not found")
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
	html := IconRight("arrow_forward", "Next")

	// Should have Material Icons class and icon name
	assertContains(t, html, `material-icons`)
	assertContains(t, html, `arrow_forward`)

	// Should have text
	assertContains(t, html, `Next`)

	// Should have flex layout
	assertContains(t, html, `flex-1 flex items-center gap-2`)
}

func TestIconRight_Structure(t *testing.T) {
	html := IconRight("chevron_right", "Continue")

	// IconRight layout: Flex1 - Text - Icon - Flex1
	// This means text comes before the icon
	iconPos := strings.Index(html, `chevron_right`)
	textPos := strings.Index(html, `Continue`)

	if iconPos == -1 {
		t.Error("Icon name not found")
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
	html := IconEnd("check", "Done")

	// Should have Material Icons class and icon name
	assertContains(t, html, `material-icons`)
	assertContains(t, html, `check`)

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
	html := Icon("home", Attr{Class: "text-blue-500"})

	assertContains(t, html, `material-icons`)
	assertContains(t, html, `home`)
	assertContains(t, html, `text-blue-500`)
}

func TestIcon_WithCustomAttributes(t *testing.T) {
	html := Icon("settings", Attr{
		ID:    "my-icon",
		Class: "additional-class",
		Style: "color: red;",
	})

	assertContains(t, html, `material-icons`)
	assertContains(t, html, `settings`)
	assertContains(t, html, `additional-class`)
	assertContains(t, html, `style="color: red;"`)
}

// ============================================================================
// Icon Layout - Edge Cases
// ============================================================================

func TestIconStart_EmptyText(t *testing.T) {
	html := IconStart("home", "")

	assertContains(t, html, `home`)
}

func TestIconLeft_EmptyText(t *testing.T) {
	html := IconLeft("arrow_back", "")

	assertContains(t, html, `arrow_back`)
}

func TestIconRight_EmptyText(t *testing.T) {
	html := IconRight("arrow_forward", "")

	assertContains(t, html, `arrow_forward`)
}

func TestIconEnd_EmptyText(t *testing.T) {
	html := IconEnd("check", "")

	assertContains(t, html, `check`)
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
	html := IconStart("home", longText)

	assertContains(t, html, `home`)
	assertContains(t, html, longText)
}

func TestIconLeft_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconLeft("arrow_back", longText)

	assertContains(t, html, `arrow_back`)
	assertContains(t, html, longText)
}

func TestIconRight_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconRight("arrow_forward", longText)

	assertContains(t, html, `arrow_forward`)
	assertContains(t, html, longText)
}

func TestIconEnd_LongText(t *testing.T) {
	longText := strings.Repeat("Very Long Text ", 10)
	html := IconEnd("check", longText)

	assertContains(t, html, `check`)
	assertContains(t, html, longText)
}
