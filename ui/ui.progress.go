package ui

import (
	"fmt"
	"strconv"
	"strings"
)

type progress struct {
	value    int
	color    string
	striped  bool
	animated bool
	label    string
	class    string
	visible  bool
}

func ProgressBar() *progress {
	return &progress{
		value:   0,
		color:   "bg-blue-600",
		visible: true,
	}
}

func (p *progress) Value(percent int) *progress {
	// Clamp value between 0 and 100
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	p.value = percent
	return p
}

func (p *progress) Color(value string) *progress {
	p.color = value
	return p
}

func (p *progress) Striped(value bool) *progress {
	p.striped = value
	return p
}

func (p *progress) Animated(value bool) *progress {
	p.animated = value
	return p
}

func (p *progress) Label(value string) *progress {
	p.label = value
	return p
}

func (p *progress) If(value bool) *progress {
	p.visible = value
	return p
}

func (p *progress) Class(value ...string) *progress {
	p.class = strings.Join(value, " ")
	return p
}

func (p *progress) Render() string {
	if !p.visible {
		return ""
	}

	// Build base container classes
	containerClasses := []string{
		"w-full",
		"overflow-hidden",
		"bg-gray-200",
		"dark:bg-gray-700",
		"rounded-full",
		"h-4",
	}

	if p.class != "" {
		containerClasses = append(containerClasses, p.class)
	}

	// Build bar classes
	barClasses := []string{
		"h-full",
		"rounded-full",
		"transition-all",
		"duration-300",
		"ease-out",
		p.color,
	}

	// Build inline style for width
	barStyle := fmt.Sprintf("width: %d%%", p.value)

	// Add striped pattern using inline style
	if p.striped {
		barStyle += "; background-image: linear-gradient(45deg, rgba(255,255,255,.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,.15) 50%, rgba(255,255,255,.15) 75%, transparent 75%, transparent); background-size: 1rem 1rem"
	}

	// Add animation using inline style
	if p.animated && p.striped {
		barStyle += "; animation: progress-stripes 1s linear infinite"
	}

	// Build the progress bar HTML
	container := Div(
		Classes(containerClasses...),
		Attr{Style: "position: relative;"},
	)(
		func() string {
			// Create bar element with data attribute for JS animation
			bar := fmt.Sprintf(
				`<div class="%s" style="%s" data-progress-value="%d"></div>`,
				Classes(barClasses...),
				barStyle,
				p.value,
			)

			// Add label if provided
			if p.label != "" {
				label := fmt.Sprintf(
					`<div class="absolute inset-0 flex items-center justify-center text-xs font-semibold text-gray-700 dark:text-gray-200" style="pointer-events: none;">%s</div>`,
					escapeAttr(p.label),
				)
				return bar + label
			}

			return bar
		}(),
	)

	// Add animation stylesheet for animated progress bars
	animationStyle := p.getAnimationStyle()

	return container + animationStyle
}

// getAnimationStyle returns CSS style tag for animating progress bar
func (p *progress) getAnimationStyle() string {
	if !p.animated || !p.striped {
		return ""
	}

	// Return style tag with animation keyframes
	return Trim(`<style id="__progress-anim__">@keyframes progress-stripes{0%{background-position:1rem 0}100%{background-position:0 0}}</style>`)
}

// ProgressWithLabel creates a progress bar with percentage label
func ProgressWithLabel(percent int) *progress {
	return ProgressBar().Value(percent).Label(strconv.Itoa(percent) + "%")
}
