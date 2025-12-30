package ui

import (
	"fmt"
	"strings"
)

type card struct {
	header   string
	body     string
	footer   string
	image    string
	imageAlt string
	variant  string
	class    string
	padding  string
	hover    bool
	compact  bool
	visible  bool
}

const (
	// Card variants
	CardBordered = "bordered"
	CardShadowed = "shadowed"
	CardFlat     = "flat"
	CardGlass    = "glass"
)

func Card() *card {
	return &card{
		variant: CardShadowed,
		visible: true,
		padding: "p-6",
	}
}

func (c *card) Header(html string) *card {
	c.header = html
	return c
}

func (c *card) Body(html string) *card {
	c.body = html
	return c
}

func (c *card) Footer(html string) *card {
	c.footer = html
	return c
}

func (c *card) Image(src string, alt string) *card {
	c.image = src
	c.imageAlt = alt
	return c
}

func (c *card) Padding(value string) *card {
	c.padding = value
	return c
}

func (c *card) Hover(value bool) *card {
	c.hover = value
	return c
}

func (c *card) Compact(value bool) *card {
	c.compact = value
	if value {
		c.padding = "p-4"
	}
	return c
}

func (c *card) Variant(value string) *card {
	c.variant = value
	return c
}

func (c *card) If(value bool) *card {
	c.visible = value
	return c
}

func (c *card) Class(value ...string) *card {
	c.class = strings.Join(value, " ")
	return c
}

func (c *card) Render() string {
	if !c.visible {
		return ""
	}

	// Base classes
	baseClasses := []string{
		"bg-white",
		"dark:bg-gray-900",
		"rounded-xl",
		"overflow-hidden",
	}

	// Variant-specific classes
	switch c.variant {
	case CardBordered:
		baseClasses = append(baseClasses,
			"border",
			"border-gray-200",
			"dark:border-gray-800",
		)
	case CardShadowed:
		baseClasses = append(baseClasses,
			"shadow-sm",
			"border",
			"border-gray-100",
			"dark:border-gray-800/50",
		)
	case CardFlat:
		// Flat has no border or shadow
	case CardGlass:
		baseClasses = []string{
			"bg-white/70",
			"dark:bg-gray-900/70",
			"backdrop-blur-md",
			"rounded-xl",
			"overflow-hidden",
			"border",
			"border-white/20",
			"dark:border-gray-800/50",
		}
	default:
		// Default to shadowed for unknown variants
		baseClasses = append(baseClasses,
			"shadow-sm",
			"border",
			"border-gray-100",
			"dark:border-gray-800/50",
		)
	}

	if c.hover {
		baseClasses = append(baseClasses, "transition-all duration-300 hover:shadow-lg hover:-translate-y-1")
	}

	// Add custom classes
	if c.class != "" {
		baseClasses = append(baseClasses, c.class)
	}

	cardClass := Classes(baseClasses...)

	var sections []string

	// Image section
	if c.image != "" {
		height := "h-48"
		if c.compact {
			height = "h-32"
		}
		sections = append(sections, fmt.Sprintf(`<img src="%s" alt="%s" class="w-full %s object-cover">`, escapeAttr(c.image), escapeAttr(c.imageAlt), height))
	}

	// Header section
	if c.header != "" {
		padding := "px-6 py-4"
		if c.compact {
			padding = "px-4 py-3"
		}
		headerHtml := Div(
			Classes(
				padding,
				"border-b",
				"border-gray-100/80",
				"dark:border-gray-800/80",
				"bg-gray-50/30",
				"dark:bg-gray-800/30",
			),
		)(c.header)
		sections = append(sections, headerHtml)
	}

	// Body section
	if c.body != "" {
		bodyHtml := Div(
			Classes(c.padding),
		)(c.body)
		sections = append(sections, bodyHtml)
	}

	// Footer section
	if c.footer != "" {
		padding := "px-6 py-4"
		if c.compact {
			padding = "px-4 py-3"
		}
		footerHtml := Div(
			Classes(
				padding,
				"border-t",
				"border-gray-100/80",
				"dark:border-gray-800/80",
				"bg-gray-50/30",
				"dark:bg-gray-800/30",
			),
		)(c.footer)
		sections = append(sections, footerHtml)
	}

	return Div(cardClass)(strings.Join(sections, ""))
}
