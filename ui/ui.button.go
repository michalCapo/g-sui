package ui

import "strings"

type button struct {
	size     string
	color    string
	onclick  string
	class    string
	as       string
	target   Attr
	visible  bool
	disabled bool
	attr     []Attr
}

func Button(attr ...Attr) *button {
	return &button{
		as:      "div",
		size:    MD,
		visible: true,
		target:  Target(),
		attr:    attr,
	}
}

func (b *button) Submit() *button {
	b.as = "submit"
	return b
}

func (b *button) Reset() *button {
	b.as = "reset"
	return b
}

func (b *button) If(value bool) *button {
	b.visible = value
	return b
}

func (b *button) Disabled(value bool) *button {
	b.disabled = value
	return b
}

func (b *button) Class(value ...string) *button {
	b.class = strings.Join(value, " ")
	return b
}

func (b *button) Color(value string) *button {
	b.color = value
	return b
}

func (b *button) Size(value string) *button {
	b.size = value
	return b
}

func (b *button) Click(onclick string) *button {
	// if action.Target.Id == "" {
	// 	action.Target = b.target
	// }

	b.onclick = onclick

	return b
}

func (b *button) Href(value string) *button {
	b.as = "a"
	b.attr = append(b.attr, Attr{Href: value})
	return b
}

func (b *button) Name(value string) *button {
	b.attr = append(b.attr, Attr{Name: value})
	return b
}

func (b *button) Val(value string) *button {
	b.attr = append(b.attr, Attr{Value: value})
	return b
}

func (b *button) Render(text string) string {
	if !b.visible {
		return ""
	}

	class := Classes(BTN, b.size, b.color, b.class, If(b.disabled, func() string { return DISABLED + " opacity-25" }))

	if b.as == "a" {
		return A(
			class,
			append(b.attr, Attr{
				ID: b.target.ID,
			})...,
		)(text)
	}

	if b.as == "div" {
		return Div(
			class,
			append(b.attr, Attr{
				ID:      b.target.ID,
				OnClick: b.onclick,
			})...,
		)(text)
	}

	return open("button")(
		class,
		append(b.attr,
			Attr{
				ID:      b.target.ID,
				Type:    b.as,
				OnClick: b.onclick,
			})...,
	)(text)
}
