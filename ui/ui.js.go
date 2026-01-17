package ui

import (
	"fmt"
	"strings"
)

// Element constructors that output HTML strings

// El creates an HTML element constructor for any tag
func El(tag string, class string, attr ...Attr) func(elements ...string) string {
	return func(elements ...string) string {
		attr = append(attr, Attr{Class: class})
		return element(tag, attr, elements...)
	}
}

// ElClosed creates an HTML element constructor for self-closing tags
func ElClosed(tag string, class string, attr ...Attr) string {
	attr = append(attr, Attr{Class: class})
	return elementClosed(tag, attr)
}

// element generates HTML for an element with children
func element(tag string, attrs []Attr, children ...string) string {
	attrStr := buildAttrs(attrs)
	if attrStr != "" {
		attrStr = " " + attrStr
	}
	return fmt.Sprintf("<%s%s>%s</%s>", tag, attrStr, strings.Join(children, ""), tag)
}

// elementClosed generates HTML for a self-closing element
func elementClosed(tag string, attrs []Attr) string {
	attrStr := buildAttrs(attrs)
	if attrStr != "" {
		attrStr = " " + attrStr
	}
	return fmt.Sprintf("<%s%s>", tag, attrStr)
}

// buildAttrs builds the HTML attribute string from Attr slice
func buildAttrs(attrs []Attr) string {
	var parts []string

	for _, attr := range attrs {
		if attr.ID != "" {
			parts = append(parts, fmt.Sprintf(`id="%s"`, escapeAttr(attr.ID)))
		}
		if attr.Class != "" {
			parts = append(parts, fmt.Sprintf(`class="%s"`, escapeAttr(attr.Class)))
		}
		if attr.Style != "" {
			parts = append(parts, fmt.Sprintf(`style="%s"`, escapeAttr(attr.Style)))
		}
		if attr.Href != "" {
			parts = append(parts, fmt.Sprintf(`href="%s"`, escapeAttr(attr.Href)))
		}
		if attr.Src != "" {
			parts = append(parts, fmt.Sprintf(`src="%s"`, escapeAttr(attr.Src)))
		}
		if attr.Alt != "" {
			parts = append(parts, fmt.Sprintf(`alt="%s"`, escapeAttr(attr.Alt)))
		}
		if attr.Title != "" {
			parts = append(parts, fmt.Sprintf(`title="%s"`, escapeAttr(attr.Title)))
		}
		if attr.Type != "" {
			parts = append(parts, fmt.Sprintf(`type="%s"`, escapeAttr(attr.Type)))
		}
		if attr.Name != "" {
			parts = append(parts, fmt.Sprintf(`name="%s"`, escapeAttr(attr.Name)))
		}
		if attr.Value != "" {
			parts = append(parts, fmt.Sprintf(`value="%s"`, escapeAttr(attr.Value)))
		}
		if attr.Placeholder != "" {
			parts = append(parts, fmt.Sprintf(`placeholder="%s"`, escapeAttr(attr.Placeholder)))
		}
		if attr.Pattern != "" {
			parts = append(parts, fmt.Sprintf(`pattern="%s"`, escapeAttr(attr.Pattern)))
		}
		if attr.Autocomplete != "" {
			parts = append(parts, fmt.Sprintf(`autocomplete="%s"`, escapeAttr(attr.Autocomplete)))
		}
		if attr.For != "" {
			parts = append(parts, fmt.Sprintf(`for="%s"`, escapeAttr(attr.For)))
		}
		if attr.Form != "" {
			parts = append(parts, fmt.Sprintf(`form="%s"`, escapeAttr(attr.Form)))
		}
		if attr.Target != "" {
			parts = append(parts, fmt.Sprintf(`target="%s"`, escapeAttr(attr.Target)))
		}
		if attr.OnClick != "" {
			parts = append(parts, fmt.Sprintf(`onclick="%s"`, escapeAttr(attr.OnClick)))
		}
		if attr.OnChange != "" {
			parts = append(parts, fmt.Sprintf(`onchange="%s"`, escapeAttr(attr.OnChange)))
		}
		if attr.OnSubmit != "" {
			parts = append(parts, fmt.Sprintf(`onsubmit="%s"`, escapeAttr(attr.OnSubmit)))
		}
		if attr.Min != "" {
			parts = append(parts, fmt.Sprintf(`min="%s"`, escapeAttr(attr.Min)))
		}
		if attr.Max != "" {
			parts = append(parts, fmt.Sprintf(`max="%s"`, escapeAttr(attr.Max)))
		}
		if attr.Step != "" {
			parts = append(parts, fmt.Sprintf(`step="%s"`, escapeAttr(attr.Step)))
		}
		if attr.Checked != "" {
			parts = append(parts, "checked")
		}
		if attr.Selected != "" {
			parts = append(parts, "selected")
		}
		if attr.Disabled {
			parts = append(parts, "disabled")
		}
		if attr.Required {
			parts = append(parts, "required")
		}
		if attr.Readonly {
			parts = append(parts, "readonly")
		}
		if attr.Rows != 0 {
			parts = append(parts, fmt.Sprintf(`rows="%d"`, attr.Rows))
		}
		if attr.Cols != 0 {
			parts = append(parts, fmt.Sprintf(`cols="%d"`, attr.Cols))
		}
		if attr.Width != 0 {
			parts = append(parts, fmt.Sprintf(`width="%d"`, attr.Width))
		}
		if attr.Height != 0 {
			parts = append(parts, fmt.Sprintf(`height="%d"`, attr.Height))
		}
		// Data attributes
		if attr.DataAccordion != "" {
			parts = append(parts, fmt.Sprintf(`data-accordion="%s"`, escapeAttr(attr.DataAccordion)))
		}
		if attr.DataAccordionItem != "" {
			parts = append(parts, fmt.Sprintf(`data-accordion-item="%s"`, escapeAttr(attr.DataAccordionItem)))
		}
		if attr.DataAccordionContent != "" {
			parts = append(parts, fmt.Sprintf(`data-accordion-content="%s"`, escapeAttr(attr.DataAccordionContent)))
		}
		if attr.DataTabs != "" {
			parts = append(parts, fmt.Sprintf(`data-tabs="%s"`, escapeAttr(attr.DataTabs)))
		}
		if attr.DataTabsIndex != "" {
			parts = append(parts, fmt.Sprintf(`data-tabs-index="%s"`, escapeAttr(attr.DataTabsIndex)))
		}
		if attr.DataTabsPanel != "" {
			parts = append(parts, fmt.Sprintf(`data-tabs-panel="%s"`, escapeAttr(attr.DataTabsPanel)))
		}
	}

	return strings.Join(parts, " ")
}

// Primary element constructors (generate HTML)
var (
	I        = func(class string, attr ...Attr) func(elements ...string) string { return El("i", class, attr...) }
	A        = func(class string, attr ...Attr) func(elements ...string) string { return El("a", class, attr...) }
	P        = func(class string, attr ...Attr) func(elements ...string) string { return El("p", class, attr...) }
	Div      = func(class string, attr ...Attr) func(elements ...string) string { return El("div", class, attr...) }
	Span     = func(class string, attr ...Attr) func(elements ...string) string { return El("span", class, attr...) }
	Form     = func(class string, attr ...Attr) func(elements ...string) string { return El("form", class, attr...) }
	Textarea = func(class string, attr ...Attr) func(elements ...string) string { return El("textarea", class, attr...) }
	Select   = func(class string, attr ...Attr) func(elements ...string) string { return El("select", class, attr...) }
	Option   = func(class string, attr ...Attr) func(elements ...string) string { return El("option", class, attr...) }
	List     = func(class string, attr ...Attr) func(elements ...string) string { return El("ul", class, attr...) }
	ListItem = func(class string, attr ...Attr) func(elements ...string) string { return El("li", class, attr...) }
	Canvas   = func(class string, attr ...Attr) func(elements ...string) string { return El("canvas", class, attr...) }

	Img   = func(class string, attr ...Attr) string { return ElClosed("img", class, attr...) }
	Input = func(class string, attr ...Attr) string { return ElClosed("input", class, attr...) }
)

// Text is a pass-through for plain text (no transformation needed in HTML mode)
func Text(text string) string {
	return text
}

// Flex1 creates a flex-1 div element
var Flex1 = Div("flex-1")()

// Space represents a non-breaking space
var Space = "&nbsp;"
