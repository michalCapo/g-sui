package ui

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ASelect struct {
	as          string
	data        any
	error       validator.FieldError
	name        string
	class       string
	size        string
	onchange    string
	placeholder string
	target      Attr
	options     []AOption
	empty       bool
	disabled    bool
	required    bool
	visible     bool
	form        string
}

func (c *ASelect) Error(errs *error) *ASelect {
	if errs == nil {
		return c
	}

	temp := (*errs).(validator.ValidationErrors)

	for _, err := range temp {
		if err.Field() == c.name {
			c.error = err
		}
	}

	return c
}

func (c *ASelect) Type(value string) *ASelect {
	c.as = value
	return c
}

func (c *ASelect) If(value bool) *ASelect {
	c.visible = value
	return c
}

func (c *ASelect) Placeholder(value string) *ASelect {
	c.placeholder = value
	return c
}

func (c *ASelect) Class(value ...string) *ASelect {
	c.class = strings.Join(value, " ")
	return c
}

func (c *ASelect) Required(value ...bool) *ASelect {
	if value == nil {
		c.required = true
		return c
	}

	c.required = value[0]
	return c
}

func (c *ASelect) Disabled(value ...bool) *ASelect {
	if value == nil {
		c.disabled = true
		return c
	}

	c.disabled = value[0]
	return c
}

func (c *ASelect) Change(action string) *ASelect {
	// if action.Target.Id == "" {
	// 	action.Target = c.target
	// }

	c.onchange = action

	return c
}

func (c *ASelect) Empty() *ASelect {
	c.empty = true
	return c
}

func (c *ASelect) Options(options []AOption) *ASelect {
	c.options = options
	return c
}

func (c *ASelect) Form(form string) *ASelect {
	c.form = form
	return c
}

func (c *ASelect) Render(text string) string {
	value := ""

	if c.data != nil {
		// v := reflect.ValueOf(c.data)

		// if v.Kind() == reflect.Ptr {
		// 	v = v.Elem()
		// }

		// tmp := v.FieldByName(c.name)

		tmp, err := PathValue(c.data, c.name)

		if err == nil {
			value = fmt.Sprintf("%v", tmp.Interface())
		}
	}

	return Div(c.class)(
		Label(&c.target).
			Required(c.required).
			Render(text),

		Select(
			Classes(INPUT, c.size, If(c.disabled, func() string { return "cursor-text bg-gray-100" }), If(c.error != nil, func() string { return "border-l-8 border-red-600" })),
			Attr{
				Form:        c.form,
				Type:        c.as,
				ID:          c.target.ID,
				Name:        c.name,
				Required:    c.required,
				Placeholder: c.placeholder,
				Disabled:    c.disabled,
				OnChange:    c.onchange,
			},
		)(
			If(c.empty, func() string { return Option("", Attr{Value: ""})() }),
			Map(c.options, func(option *AOption, index int) string {
				return Option("", Attr{Value: option.ID, Selected: If(option.ID == value, func() string { return "selected" })})(option.Value)
			}),
		),
	)
}

func ISelect(name string, data ...any) *ASelect {
	var temp any

	if len(data) > 0 {
		temp = data[0]
	}

	return &ASelect{
		as:      "text",
		name:    name,
		size:    MD,
		target:  Target(),
		visible: true,
		data:    temp,
	}
}
