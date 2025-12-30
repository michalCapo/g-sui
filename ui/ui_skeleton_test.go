package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Skeleton Tests
// ============================================================================

func TestSkeletonDefault_Rendering(t *testing.T) {
	html := SkeletonDefault()

	// Should have animate-pulse class
	assertContains(t, html, "animate-pulse")

	// Should have rounded skeleton bars
	// Looking for bg-gray-200 h-5 which is the skeleton bar class
	count := strings.Count(html, `bg-gray-200 h-5`)
	if count < 3 {
		t.Errorf("Default skeleton should have at least 3 skeleton lines, got %d", count)
	}

	// Should have white/dark background container
	assertContains(t, html, `bg-white`)
}

func TestSkeletonList_Rendering(t *testing.T) {
	html := SkeletonListN(5)

	// Should have animate-pulse class
	assertContains(t, html, "animate-pulse")

	// Should have avatar circles
	assertContains(t, html, `rounded-full h-10 w-10`)
}

func TestSkeletonList_CustomCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"Zero count", 0},
		{"Single item", 1},
		{"Three items", 3},
		{"Ten items", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := SkeletonListN(tt.count)

			// Should still render
			if html == "" {
				t.Error("SkeletonListN should render HTML")
			}

			assertContains(t, html, "animate-pulse")
		})
	}
}

func TestSkeletonComponent_Rendering(t *testing.T) {
	html := SkeletonComponentBlock()

	// Should have animate-pulse class
	assertContains(t, html, "animate-pulse")

	// Should have header bar
	assertContains(t, html, `bg-gray-200 h-6`)

	// Should have content bars
	count := strings.Count(html, `bg-gray-200 h-4`)
	if count < 2 {
		t.Errorf("Component skeleton should have at least 2 content bars, got %d", count)
	}
}

func TestSkeletonPage_Rendering(t *testing.T) {
	html := SkeletonPageBlock()

	// Should have animate-pulse class
	assertContains(t, html, "animate-pulse")

	// Should have page header
	assertContains(t, html, `bg-gray-200 h-8`)

	// Should have card elements
	assertContains(t, html, `bg-white dark:bg-gray-900`)
}

func TestSkeletonForm_Rendering(t *testing.T) {
	html := SkeletonFormBlock()

	// Should have animate-pulse class
	assertContains(t, html, "animate-pulse")

	// Should have form header
	assertContains(t, html, `bg-gray-200 h-6`)

	// Should have input fields (h-10)
	assertContains(t, html, `bg-gray-200 h-10`)

	// Should have textarea (h-24)
	assertContains(t, html, `bg-gray-200 h-24`)

	// Should have action buttons
	assertContains(t, html, `bg-gray-200 h-10`)
}

func TestAttrSkeleton_KindVariants(t *testing.T) {
	target := Target()

	tests := []struct {
		name     string
		kind     Skeleton
		contains []string
	}{
		{
			name: "SkeletonList",
			kind: SkeletonList,
			contains: []string{
				"animate-pulse",
				"rounded-full",
				"h-10 w-10",
			},
		},
		{
			name: "SkeletonComponent",
			kind: SkeletonComponent,
			contains: []string{
				"animate-pulse",
				"bg-gray-200 h-6",
			},
		},
		{
			name: "SkeletonPage",
			kind: SkeletonPage,
			contains: []string{
				"animate-pulse",
				"bg-gray-200 h-8",
			},
		},
		{
			name: "SkeletonForm",
			kind: SkeletonForm,
			contains: []string{
				"animate-pulse",
				"bg-gray-200 h-10",
				"bg-gray-200 h-24",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := target.Skeleton(tt.kind)

			for _, expected := range tt.contains {
				if !strings.Contains(html, expected) {
					t.Errorf("Skeleton %s should contain %q", tt.kind, expected)
				}
			}
		})
	}
}

func TestAttrSkeleton_DefaultVariant(t *testing.T) {
	target := Target()

	// Call with no kind specified (empty string)
	html := target.Skeleton("")

	// Should render default skeleton
	assertContains(t, html, "animate-pulse")

	// Should have skeleton bars
	count := strings.Count(html, `bg-gray-200 h-5`)
	if count < 3 {
		t.Errorf("Default skeleton should have at least 3 skeleton lines, got %d", count)
	}
}

func TestAttrSkeleton_UnknownKind(t *testing.T) {
	target := Target()

	// Call with unknown kind
	html := target.Skeleton("unknown_kind")

	// Should render default skeleton for unknown kind
	assertContains(t, html, "animate-pulse")
}

func TestAttrSkeleton_WithID(t *testing.T) {
	target := Target()
	targetID := target.ID

	html := target.Skeleton(SkeletonComponent)

	// Should include the target ID
	assertContains(t, html, `id="`+targetID+`"`)
}

func TestSkeletonDefault_DarkMode(t *testing.T) {
	html := SkeletonDefault()

	// Should have dark mode classes
	assertContains(t, html, `dark:bg-gray-900`)
}

func TestSkeletonList_DarkMode(t *testing.T) {
	html := SkeletonListN(3)

	// Should have dark mode classes
	assertContains(t, html, `dark:bg-gray-900`)
}

func TestSkeletonElement_Rendering(t *testing.T) {
	// Test using Attr.Skeleton() directly
	html := Attr{}.Skeleton(SkeletonList)

	assertContains(t, html, "animate-pulse")
	assertContains(t, html, `rounded-full`)
}
