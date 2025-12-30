package ui

import "strings"

type badge struct {
	text    string
	color   string
	dot     bool
	visible bool
	class   string
	attr    []Attr
}

func Badge(attr ...Attr) *badge {
	return &badge{
		color:   "bg-gray-600",
		visible: true,
		attr:    attr,
	}
}

func (b *badge) Text(value string) *badge {
	b.text = value
	return b
}

func (b *badge) Color(value string) *badge {
	b.color = value
	return b
}

func (b *badge) Dot() *badge {
	b.dot = true
	return b
}

func (b *badge) If(value bool) *badge {
	b.visible = value
	return b
}

func (b *badge) Class(value ...string) *badge {
	b.class = strings.Join(value, " ")
	return b
}

func (b *badge) Render() string {
	if !b.visible {
		return ""
	}

	if b.dot {
		// Dot variant: small circular notification indicator
		baseClass := "inline-flex items-center justify-center rounded-full"
		sizeClass := "h-2 w-2"
		colorClass := b.getColorClasses()
		extraClass := b.class

		return Span(Classes(baseClass, sizeClass, colorClass, extraClass), b.attr...)()
	}

	// Text badge variant: pill-shaped with text/number
	baseClass := "inline-flex items-center justify-center rounded-full px-2 py-0.5 text-xs font-medium"
	sizeClass := "min-w-[1.25rem] h-5"
	colorClass := b.getColorClasses()
	extraClass := b.class

	return Span(Classes(baseClass, sizeClass, colorClass, extraClass), b.attr...)(b.text)
}

func (b *badge) getColorClasses() string {
	switch b.color {
	case "red", Red:
		return "bg-red-600 text-white dark:bg-red-700 dark:text-red-100"
	case "green", Green:
		return "bg-green-600 text-white dark:bg-green-700 dark:text-green-100"
	case "blue", Blue:
		return "bg-blue-600 text-white dark:bg-blue-700 dark:text-blue-100"
	case "yellow", Yellow:
		return "bg-yellow-500 text-gray-900 dark:bg-yellow-600 dark:text-gray-100"
	case "gray", Gray:
		return "bg-gray-600 text-white dark:bg-gray-700 dark:text-gray-100"
	default:
		// Custom color or default fallback
		if strings.HasPrefix(b.color, "bg-") {
			return b.color
		}
		return "bg-gray-600 text-white dark:bg-gray-700 dark:text-gray-100"
	}
}
