package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Skeleton renderers for client zones
// ============================================================================

func TestSkeletonTableBlock_Rendering(t *testing.T) {
	html := SkeletonTableBlock()

	assertContains(t, html, "animate-pulse")
	assertContains(t, html, "<table")
	assertContains(t, html, "<thead")
	assertContains(t, html, "<tbody")
	assertContains(t, html, "<th")
	assertContains(t, html, "<td")

	// Should have 4 header cells
	if c := strings.Count(html, "<th "); c != 4 {
		t.Errorf("expected 4 <th> cells, got %d", c)
	}

	// Should have 5 rows × 4 cells = 20 <td> cells
	if c := strings.Count(html, "<td "); c != 20 {
		t.Errorf("expected 20 <td> cells, got %d", c)
	}
}

func TestSkeletonCardsBlock_Rendering(t *testing.T) {
	html := SkeletonCardsBlock()

	assertContains(t, html, "animate-pulse")
	assertContains(t, html, "grid")
	assertContains(t, html, "grid-cols-1")
	assertContains(t, html, "md:grid-cols-2")
	assertContains(t, html, "lg:grid-cols-3")

	// Should have 6 card items
	if c := strings.Count(html, "rounded-lg p-4 shadow"); c != 6 {
		t.Errorf("expected 6 card skeletons, got %d", c)
	}
}

func TestSkeletonTable_DarkMode(t *testing.T) {
	html := SkeletonTableBlock()
	assertContains(t, html, "dark:bg-gray-700")
	assertContains(t, html, "dark:bg-gray-900")
}

func TestSkeletonCards_DarkMode(t *testing.T) {
	html := SkeletonCardsBlock()
	assertContains(t, html, "dark:bg-gray-700")
	assertContains(t, html, "dark:bg-gray-900")
}

func TestAttrSkeletonTable_WithID(t *testing.T) {
	html := Attr{ID: "my-zone"}.SkeletonTable()
	assertContains(t, html, `id="my-zone"`)
	assertContains(t, html, "animate-pulse")
	assertContains(t, html, "<table")
}

func TestAttrSkeletonCards_WithID(t *testing.T) {
	html := Attr{ID: "cards-zone"}.SkeletonCards()
	assertContains(t, html, `id="cards-zone"`)
	assertContains(t, html, "animate-pulse")
	assertContains(t, html, "grid")
}
