package ui

import "strings"

type card struct {
	header  string
	body    string
	footer  string
	variant string
	class   string
	visible bool
}

const (
	// Card variants
	CardBordered  = "bordered"
	CardShadowed  = "shadowed"
	CardFlat      = "flat"
)

func Card() *card {
	return &card{
		variant: CardShadowed,
		visible: true,
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
		"rounded-lg",
		"overflow-hidden",
	}

	// Variant-specific classes
	switch c.variant {
	case CardBordered:
		baseClasses = append(baseClasses,
			"border",
			"border-gray-300",
			"dark:border-gray-700",
		)
	case CardShadowed:
		baseClasses = append(baseClasses,
			"shadow-md",
			"border",
			"border-gray-200",
			"dark:border-gray-800",
		)
	case CardFlat:
		// Flat has no border or shadow
	default:
		// Default to shadowed for unknown variants
		baseClasses = append(baseClasses,
			"shadow-md",
			"border",
			"border-gray-200",
			"dark:border-gray-800",
		)
	}

	// Add custom classes
	if c.class != "" {
		baseClasses = append(baseClasses, c.class)
	}

	cardClass := Classes(baseClasses...)

	var sections []string

	// Header section
	if c.header != "" {
		headerHtml := Div(
			Classes(
				"px-6",
				"py-4",
				"border-b",
				"border-gray-200",
				"dark:border-gray-800",
				"bg-gray-50",
				"dark:bg-gray-800/50",
			),
		)(c.header)
		sections = append(sections, headerHtml)
	}

	// Body section
	if c.body != "" {
		bodyHtml := Div(
			Classes(
				"px-6",
				"py-4",
			),
		)(c.body)
		sections = append(sections, bodyHtml)
	}

	// Footer section
	if c.footer != "" {
		footerHtml := Div(
			Classes(
				"px-6",
				"py-4",
				"border-t",
				"border-gray-200",
				"dark:border-gray-800",
				"bg-gray-50",
				"dark:bg-gray-800/50",
			),
		)(c.footer)
		sections = append(sections, footerHtml)
	}

	return Div(cardClass)(strings.Join(sections, ""))
}
