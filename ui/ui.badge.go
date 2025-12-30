package ui

import (
	"fmt"
	"strings"
)

type badge struct {
	text    string
	color   string
	dot     bool
	icon    string
	size    string
	rounded bool
	visible bool
	class   string
	attr    []Attr
}

func Badge(attr ...Attr) *badge {
	return &badge{
		color:   "gray",
		size:    "md",
		rounded: true,
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

func (b *badge) Icon(html string) *badge {
	b.icon = html
	return b
}

func (b *badge) Size(value string) *badge {
	b.size = value
	return b
}

func (b *badge) Square() *badge {
	b.rounded = false
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

	isOutline := strings.HasSuffix(b.color, "-outline")
	isSoft := strings.HasSuffix(b.color, "-soft")
	colorName := strings.TrimSuffix(strings.TrimSuffix(b.color, "-outline"), "-soft")

	if b.dot {
		// Dot variant: small circular notification indicator
		baseClass := "inline-flex items-center justify-center rounded-full"
		sizeClass := "h-2 w-2"
		if b.size == "lg" {
			sizeClass = "h-3 w-3"
		} else if b.size == "sm" {
			sizeClass = "h-1.5 w-1.5"
		}
		colorClass := b.getColorClasses(colorName, isOutline, isSoft)
		extraClass := b.class

		return Span(Classes(baseClass, sizeClass, colorClass, extraClass), b.attr...)()
	}

	// Text badge variant: pill-shaped with text/number
	roundedClass := "rounded-full"
	if !b.rounded {
		roundedClass = "rounded-md"
	}

	sizeClass := "px-2 py-0.5 text-[10px] h-5"
	if b.size == "lg" {
		sizeClass = "px-3 py-1 text-xs h-6"
	} else if b.size == "sm" {
		sizeClass = "px-1.5 py-0 text-[9px] h-4"
	}

	baseClass := "inline-flex items-center justify-center font-bold tracking-wide uppercase"
	colorClass := b.getColorClasses(colorName, isOutline, isSoft)
	extraClass := b.class

	content := b.text
	if b.icon != "" {
		iconSize := "w-3 h-3"
		if b.size == "sm" {
			iconSize = "w-2.5 h-2.5"
		}
		content = fmt.Sprintf(`<span class="%s mr-1 flex items-center justify-center">%s</span>%s`, iconSize, b.icon, b.text)
	}

	return Span(Classes(baseClass, sizeClass, roundedClass, colorClass, extraClass), b.attr...)(content)
}

func (b *badge) getColorClasses(colorName string, isOutline, isSoft bool) string {
	switch colorName {
	case "red", Red:
		if isOutline {
			return "bg-transparent text-red-600 border border-red-600 dark:text-red-400 dark:border-red-400"
		}
		if isSoft {
			return "bg-red-50 text-red-700 dark:bg-red-950/40 dark:text-red-300 border border-red-200/50 dark:border-red-800/50"
		}
		return "bg-red-600 text-white dark:bg-red-700 dark:text-red-100"
	case "green", Green:
		if isOutline {
			return "bg-transparent text-green-600 border border-green-600 dark:text-green-400 dark:border-green-400"
		}
		if isSoft {
			return "bg-green-50 text-green-700 dark:bg-green-950/40 dark:text-green-300 border border-green-200/50 dark:border-green-800/50"
		}
		return "bg-green-600 text-white dark:bg-green-700 dark:text-green-100"
	case "blue", Blue:
		if isOutline {
			return "bg-transparent text-blue-600 border border-blue-600 dark:text-blue-400 dark:border-blue-400"
		}
		if isSoft {
			return "bg-blue-50 text-blue-700 dark:bg-blue-950/40 dark:text-blue-300 border border-blue-200/50 dark:border-blue-800/50"
		}
		return "bg-blue-600 text-white dark:bg-blue-700 dark:text-blue-100"
	case "yellow", Yellow:
		if isOutline {
			return "bg-transparent text-yellow-600 border border-yellow-600 dark:text-yellow-400 dark:border-yellow-400"
		}
		if isSoft {
			return "bg-yellow-50 text-yellow-700 dark:bg-yellow-950/40 dark:text-yellow-300 border border-yellow-200/50 dark:border-yellow-800/50"
		}
		return "bg-yellow-500 text-gray-900 dark:bg-yellow-600 dark:text-gray-100"
	case "purple":
		if isOutline {
			return "bg-transparent text-purple-600 border border-purple-600 dark:text-purple-400 dark:border-purple-400"
		}
		if isSoft {
			return "bg-purple-50 text-purple-700 dark:bg-purple-950/40 dark:text-purple-300 border border-purple-200/50 dark:border-purple-800/50"
		}
		return "bg-purple-600 text-white dark:bg-purple-700 dark:text-white"
	case "gray", Gray:
		if isOutline {
			return "bg-transparent text-gray-600 border border-gray-600 dark:text-gray-400 dark:border-gray-400"
		}
		if isSoft {
			return "bg-gray-100 text-gray-700 dark:bg-gray-800/60 dark:text-gray-300 border border-gray-200 dark:border-gray-700"
		}
		return "bg-gray-600 text-white dark:bg-gray-700 dark:text-gray-100"
	default:
		// Custom color or default fallback
		if strings.HasPrefix(b.color, "bg-") {
			return b.color
		}
		return "bg-gray-600 text-white dark:bg-gray-700 dark:text-gray-100"
	}
}
