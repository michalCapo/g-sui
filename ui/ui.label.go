package ui

import "strings"

type ALabel struct {
	id         string
	class      string
	classLabel string
	required   bool
	disabled   bool
}

func Label(target *Attr) *ALabel {
	tmp := &ALabel{
		class: "text-sm",
	}

	if target != nil {
		tmp.id = target.ID
	}

	return tmp
}

func (c *ALabel) Required(value bool) *ALabel {
	c.required = value
	return c
}

func (c *ALabel) Disabled(value bool) *ALabel {
	c.disabled = value
	return c
}

func (c *ALabel) Class(value ...string) *ALabel {
	c.class = Classes(append(strings.Split(c.class, " "), value...)...)
	return c
}

func (c *ALabel) ClassLabel(value ...string) *ALabel {
	c.classLabel = Classes(append(strings.Split(c.classLabel, " "), value...)...)
	return c
}

// Render generates JavaScript code that creates the label element
func (c *ALabel) Render(text string) string {
	if text == "" {
		return ""
	}

	labelJS := El("label", c.classLabel, Attr{For: c.id})(Text(text))
	asteriskJS := ""
	if c.required && !c.disabled {
		asteriskJS = Span("ml-1 text-red-700")(Text("*"))
	}

	return Div(Classes(c.class, "relative"))(labelJS, asteriskJS)
}
