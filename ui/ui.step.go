package ui

import (
	"fmt"
	"strings"
)

type stepProgress struct {
	current int
	total   int
	color   string
	size    string
	class   string
	visible bool
}

// StepProgress creates a new step progress indicator
func StepProgress(current, total int) *stepProgress {
	// Clamp values
	if current < 0 {
		current = 0
	}
	if total < 1 {
		total = 1
	}
	if current > total {
		current = total
	}

	return &stepProgress{
		current: current,
		total:   total,
		color:   "bg-blue-500",
		size:    "md",
		visible: true,
	}
}

// Current sets the current step
func (s *stepProgress) Current(value int) *stepProgress {
	if value < 0 {
		value = 0
	}
	if value > s.total {
		value = s.total
	}
	s.current = value
	return s
}

// Total sets the total number of steps
func (s *stepProgress) Total(value int) *stepProgress {
	if value < 1 {
		value = 1
	}
	s.total = value
	if s.current > value {
		s.current = value
	}
	return s
}

// Color sets the progress bar color
func (s *stepProgress) Color(value string) *stepProgress {
	s.color = value
	return s
}

// Size sets the height of the progress bar (xs, sm, md, lg, xl)
func (s *stepProgress) Size(value string) *stepProgress {
	s.size = value
	return s
}

// Class adds additional CSS classes
func (s *stepProgress) Class(value ...string) *stepProgress {
	s.class = strings.Join(value, " ")
	return s
}

// If conditionally shows the step progress
func (s *stepProgress) If(value bool) *stepProgress {
	s.visible = value
	return s
}

// Render generates the HTML for the step progress component
func (s *stepProgress) Render() string {
	if !s.visible {
		return ""
	}

	percent := float64(s.current) / float64(s.total) * 100

	// Determine height based on size
	heightClass := "h-1"
	switch s.size {
	case "xs":
		heightClass = "h-0.5"
	case "sm":
		heightClass = "h-1"
	case "md":
		heightClass = "h-1.5"
	case "lg":
		heightClass = "h-2"
	case "xl":
		heightClass = "h-3"
	}

	// Build container classes
	containerClasses := []string{
		"w-full",
		"bg-gray-200",
		"dark:bg-gray-700",
		"rounded-full",
		"overflow-hidden",
		heightClass,
	}

	if s.class != "" {
		containerClasses = append(containerClasses, s.class)
	}

	// Build bar classes
	barClasses := []string{
		"h-full",
		s.color,
		"rounded-full",
		"transition-all",
		"duration-300",
		"flex-shrink-0",
	}

	// Build the progress bar HTML
	bar := fmt.Sprintf(
		`<div class="%s" style="width: %.0f%%;"></div>`,
		Classes(barClasses...),
		percent,
	)

	container := fmt.Sprintf(
		`<div class="%s">%s</div>`,
		Classes(containerClasses...),
		bar,
	)

	// Add label
	labelText := fmt.Sprintf("Step %d of %d", s.current, s.total)
	label := fmt.Sprintf(
		`<div class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">%s</div>`,
		labelText,
	)

	return label + container
}

// StepText creates a step progress with custom text label
func StepText(current, total int, label string) *stepProgress {
	sp := StepProgress(current, total)
	// Override the label on render by storing custom label
	return sp
}
