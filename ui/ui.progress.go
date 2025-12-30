package ui

import (
	"fmt"
	"strconv"
	"strings"
)

type progress struct {
	value         int
	color         string
	gradient      []string
	striped       bool
	animated      bool
	indeterminate bool
	size          string
	label         string
	labelPosition string // "inside", "outside"
	class         string
	visible       bool
}

func ProgressBar() *progress {
	return &progress{
		value:         0,
		color:         "bg-blue-600",
		size:          "md",
		labelPosition: "inside",
		visible:       true,
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

func (p *progress) Gradient(colors ...string) *progress {
	p.gradient = colors
	return p
}

func (p *progress) Size(value string) *progress {
	p.size = value
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

func (p *progress) Indeterminate(value bool) *progress {
	p.indeterminate = value
	return p
}

func (p *progress) Label(value string) *progress {
	p.label = value
	return p
}

func (p *progress) LabelPosition(value string) *progress {
	p.labelPosition = value
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

	// Determine height based on size
	heightClass := "h-2"
	switch p.size {
	case "xs":
		heightClass = "h-1"
	case "sm":
		heightClass = "h-1.5"
	case "md":
		heightClass = "h-2.5"
	case "lg":
		heightClass = "h-4"
	case "xl":
		heightClass = "h-6"
	}

	// Build base container classes
	containerClasses := []string{
		"w-full",
		"overflow-hidden",
		"bg-gray-200/50",
		"dark:bg-gray-800/50",
		"rounded-full",
		heightClass,
	}

	if p.class != "" {
		containerClasses = append(containerClasses, p.class)
	}

	// Build bar classes
	barClasses := []string{
		"h-full",
		"rounded-full",
	}

	// Use color if no gradient is provided
	if len(p.gradient) == 0 {
		barClasses = append(barClasses, p.color)
	}

	if !p.indeterminate {
		barClasses = append(barClasses, "transition-all", "duration-500", "ease-out")
	} else {
		barClasses = append(barClasses, "w-1/3", "animate-progress-indeterminate", p.color)
	}

	// Build inline style for width and gradient
	var barStyle string
	if !p.indeterminate {
		barStyle = fmt.Sprintf("width: %d%%", p.value)
	}

	if len(p.gradient) > 0 {
		barStyle += fmt.Sprintf("; background: linear-gradient(90deg, %s)", strings.Join(p.gradient, ", "))
	}

	// Add striped pattern using inline style
	if p.striped {
		barStyle += "; background-image: linear-gradient(45deg, rgba(255,255,255,.15) 25%, transparent 25%, transparent 50%, rgba(255,255,255,.15) 50%, rgba(255,255,255,.15) 75%, transparent 75%, transparent); background-size: 1rem 1rem"
	}

	// Add animation using inline style
	if p.animated && p.striped && !p.indeterminate {
		barStyle += "; animation: progress-stripes 1s linear infinite"
	}

	// Build the progress bar HTML
	barHTML := fmt.Sprintf(
		`<div class="%s" style="%s"></div>`,
		Classes(barClasses...),
		barStyle,
	)

	// Add label if provided
	labelHTML := ""
	if p.label != "" && !p.indeterminate {
		if p.labelPosition == "inside" {
			labelHTML = fmt.Sprintf(
				`<div class="absolute inset-0 flex items-center justify-center text-[10px] font-bold text-white mix-blend-difference pointer-events-none">%s</div>`,
				escapeAttr(p.label),
			)
		} else {
			// Outside label requires a wrapper or different positioning
			return Div("flex flex-col gap-1.5")(
				Div("flex justify-between items-center text-xs font-semibold")(
					Span("")(p.label),
					Span("text-gray-500")(fmt.Sprintf("%d%%", p.value)),
				),
				Div(Classes(containerClasses...), Attr{Style: "position: relative;"})(barHTML),
			)
		}
	}

	container := Div(
		Classes(containerClasses...),
		Attr{Style: "position: relative;"},
	)(barHTML + labelHTML)

	// Add animation stylesheet
	animationStyle := p.getAnimationStyle()

	return container + animationStyle
}

// getAnimationStyle returns CSS style tag for animating progress bar
func (p *progress) getAnimationStyle() string {
	styles := []string{}

	if (p.animated && p.striped) || p.indeterminate {
		styles = append(styles, "@keyframes progress-stripes{0%{background-position:1rem 0}100%{background-position:0 0}}")
		styles = append(styles, "@keyframes progress-indeterminate{0%{left:-33%;}100%{left:100%;}}")
		styles = append(styles, ".animate-progress-indeterminate{position:absolute; animation: progress-indeterminate 1.8s infinite cubic-bezier(0.65, 0.815, 0.735, 0.395);}")
	}

	if len(styles) == 0 {
		return ""
	}

	return fmt.Sprintf(`<style id="__progress-anim__">%s</style>`, strings.Join(styles, ""))
}

// ProgressWithLabel creates a progress bar with percentage label
func ProgressWithLabel(percent int) *progress {
	return ProgressBar().Value(percent).Label(strconv.Itoa(percent) + "%")
}
