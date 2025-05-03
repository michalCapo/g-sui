package ui

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func IRadio(name string, data ...any) *TInput {
	c := &TInput{
		as:      "radio",
		target:  Target(),
		name:    name,
		size:    MD,
		visible: true,
		data:    data[0],
	}

	c.Render = func(text string) string {
		if !c.visible {
			return ""
		}

		value := ""
		checked := ""

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

		if len(value) > 0 && value == c.value {
			checked = "checked"
		}

		c.class = Classes(c.class, "flex items-center gap-2")

		if c.required && value == "false" {
			c.class = Classes(c.class, "invalid")
		}

		return Div(Classes(c.class, c.size))(
			Input(
				Classes("hover:cursor-pointer", If(c.disabled, func() string { return DISABLED })),
				Attr{
					Checked: checked,

					Value:    c.value,
					Type:     c.as,
					Id:       c.target.Id,
					Name:     c.name,
					Required: c.required,
					Disabled: c.disabled,
					OnClick:  c.onclick,
					OnChange: Trim(fmt.Sprintf(`
						const el = document.getElementById('%s');

						if (el == null)
							return;

						if (el.required !== true || el.disabled === true) 
							return;

						const div = el.parentElement;

						if (el.checked) {
							div.classList.remove('invalid');
						} else {
							div.classList.add('invalid');
						}
					`, c.target.Id)),
				},
			),

			Label(&c.target).
				Required(c.required).
				ClassLabel("hover:cursor-pointer").
				Render(text),
		)
	}

	return c
}

type ARadio struct {
	error           validator.FieldError
	data            any
	name            string
	class           string
	size            string
	onchange        string
	placeholder     string
	button_inactive string
	button_active   string
	button          string
	target          Attr
	options         []AOption
	disabled        bool
	required        bool
	visible         bool
	empty           bool
}

func (c *ARadio) Error(errs *error) *ARadio {
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

func (c *ARadio) If(value bool) *ARadio {
	c.visible = value
	return c
}

func (c *ARadio) Placeholder(value string) *ARadio {
	c.placeholder = value
	return c
}

func (c *ARadio) Class(value ...string) *ARadio {
	c.class = strings.Join(value, " ")
	return c
}

func (c *ARadio) Required(value ...bool) *ARadio {
	if value == nil {
		c.required = true
		return c
	}

	c.required = value[0]
	return c
}

func (c *ARadio) Disabled(value ...bool) *ARadio {
	if value == nil {
		c.disabled = true
		return c
	}

	c.disabled = value[0]
	return c
}

func (c *ARadio) Change(action string) *ARadio {
	// if action.Target.Id == "" {
	// 	action.Target = c.target
	// }

	c.onchange = action

	return c
}

func (c *ARadio) Empty() *ARadio {
	c.empty = true
	return c
}

func (c *ARadio) Options(options []AOption) *ARadio {
	c.options = options
	return c
}

func (c *ARadio) Render(text string) string {
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
		If(text != "", func() string {
			return Label(&c.target).
				Required(c.required).
				Render(text)
		}),

		Hidden(c.name, "radio", value, Attr{
			Id:       c.target.Id,
			OnChange: c.onchange,
		}),

		Div(Classes("grid grid-flow-col justify-stretch gap-2", If(c.error != nil, func() string { return "border-l-8 border-red-600" })))(

			Map(c.options, func(option *AOption, index int) string {
				return Div(
					Classes(c.size, c.button, If(c.disabled, func() string { return "opacity-50 pointer-events-none" }), Or(value == option.Id, func() string { return c.button_active }, func() string { return c.button_inactive })),
					Attr{
						Target: c.target.Id,
						OnClick: Trim(fmt.Sprintf(`
							const buttons = document.querySelectorAll('[target=%s]');
							buttons.forEach(button => button.classList.value = '%s');

							event.target.classList.value = '%s';

							const el = document.getElementById('%s');
							if (el == null)
								return;
							
							el.value = '%s';
							el.dispatchEvent(new Event('change'));

						`, c.target.Id, Classes(c.size, c.button, c.button_inactive), Classes(c.size, c.button, c.button_active), c.target.Id, option.Id)),
					},
				)(option.Value)
			}),

			// return Button().
			// 	Class(c.button, If(c.disabled, "opacity-50 pointer-events-none"), If(value == option.Id, c.button_active, c.button_inactive)).
			// 	Render(option.Value)

			// return (&AButton{
			// 	as:      "div",
			// 	size:    MD,
			// 	visible: true,
			// 	target:  Target(),
			// 	class:   Classes(c.button, If(c.disabled, "opacity-50 pointer-events-none"), If(value == option.Id, c.button_active, c.button_inactive)),
			// 	onclick: Trim(fmt.Sprintf(`
			// 	`)),
			// }).Render(option.Value)

			// Map(c.options, func(option AOption, index int) string {
			// 	button_id := fmt.Sprintf("%s_%d_btn", c.target.Id, index)
			// 	option_id := fmt.Sprintf("%s_%d_rad", c.target.Id, index)

			// 	return Div(
			// 		Classes(c.size, c.button, If(c.disabled, "opacity-50 pointer-events-none"), If(value == option.Id, c.button_active, c.button_inactive)),
			// 		Attr{
			// 			Id: button_id,
			// 			OnClick: Trim(fmt.Sprintf(`
			// 				const el = document.getElementById('%s');

			// 				if (el == null)
			// 					return;

			// 				el.click();
			// 			`, option_id)),
			// 		},
			// 	)(
			// 		Div("")(option.Value),
			// 		Input("",
			// 			Attr{
			// 				Checked:  If(option.Id == value, "checked", ""),
			// 				Value:    option.Value,
			// 				Type:     "radio",
			// 				Id:       option_id,
			// 				Name:     c.name,
			// 				Required: c.required,
			// 				Disabled: c.disabled,
			// 				// OnClick:  c.onclick,
			// 				OnChange: Trim(fmt.Sprintf(`
			// 					const btn = document.getElementById('%s');

			// 					if (btn == null)
			// 						return;

			// 					btn.classList.value = '%s';
			// 				`, button_id, Classes(c.size, c.button, c.button_active))),
			// 			},
			// 		),
			// 	)
			// }),
		),
	)
}

func IRadioButtons(name string, data ...any) *ARadio {
	return &ARadio{
		name:            name,
		size:            MD,
		target:          Target(),
		visible:         true,
		data:            data[0],
		button:          "border text-center rounded cursor-pointer",
		button_active:   "bg-gray-600 text-white border-black",
		button_inactive: "bg-white text-black hover:bg-gray-600 hover:text-white",
	}
}
