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
	}

	if len(data) > 0 {
		c.data = data[0]
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
					Checked:  checked,
					Form:     c.form,
					Value:    c.value,
					Type:     c.as,
					ID:       c.target.ID,
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
					`, c.target.ID)),
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
	error          validator.FieldError
	data           any
	name           string
	class          string
	size           string
	onchange       string
	placeholder    string
	buttonInactive string
	buttonActive   string
	button         string
	target         Attr
	options        []AOption
	disabled       bool
	required       bool
	classLabel     string
	visible        bool
	empty          bool
	radioPosition  string
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

func (c *ARadio) ClassLabel(value ...string) *ARadio {
	c.classLabel = strings.Join(value, " ")
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

func (c *ARadio) Form(form string) *ARadio {
	c.target.Form = form
	return c
}

func (c *ARadio) RadioPosition(class string) *ARadio {
	c.radioPosition = class
	return c
}

func (c *ARadio) Render(text string) string {
	value := ""

	if c.data != nil && c.data != "" {
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
				Class(c.classLabel).
				Required(c.required).
				Render(text)
		}),

		Hidden(c.name, "radio", value, Attr{
			ID:       c.target.ID,
			OnChange: c.onchange,
			Form:     c.target.Form,
		}),

		Or(c.size == "", func() string {
			// Card-style layout for IRadioDiv
			radioPos := c.radioPosition
			if radioPos == "" {
				radioPos = "absolute bottom-2 right-2"
			}

			return Div(Classes("w-full flex flex-col gap-4", If(c.error != nil, func() string { return "border-l-8 border-red-600" })))(
				Map(c.options, func(option *AOption, index int) string {
					radioID := fmt.Sprintf("%s_%d", c.target.ID, index)
					checked := ""
					if value == option.ID {
						checked = "checked"
					}

					return Div(
						Classes("relative", c.button, If(c.disabled, func() string { return "opacity-50 pointer-events-none" }), Or(value == option.ID, func() string { return c.buttonActive }, func() string { return c.buttonInactive })),
						Attr{
							Target: c.target.ID,
							OnClick: Trim(fmt.Sprintf(`
								const cards = document.querySelectorAll('[target=%s]');
								cards.forEach(card => {
									card.classList.value = '%s';
									const radio = card.querySelector('input[type="radio"]');
									if (radio) radio.checked = false;
								});

								event.currentTarget.classList.value = '%s';
								const radio = event.currentTarget.querySelector('input[type="radio"]');
								if (radio) {
									radio.checked = true;
									const el = document.getElementById('%s');
									if (el != null) {
										el.value = '%s';
										el.dispatchEvent(new Event('change'));
									}
								}
							`, c.target.ID, Classes("relative", c.button, c.buttonInactive), Classes("relative", c.button, c.buttonActive), c.target.ID, option.ID)),
						},
					)(
						option.Value,
						Input(
							Classes("hover:cursor-pointer", radioPos, If(c.disabled, func() string { return DISABLED })),
							Attr{
								Checked:  checked,
								Form:     c.target.Form,
								Value:    option.ID,
								Type:     "radio",
								ID:       radioID,
								Name:     c.name,
								Required: c.required,
								Disabled: c.disabled,
								OnChange: Trim(fmt.Sprintf(`
									const el = document.getElementById('%s');
									if (el == null) return;
									el.value = '%s';
									el.dispatchEvent(new Event('change'));
								`, c.target.ID, option.ID)),
							},
						),
					)
				}),
			)
		}, func() string {
			// Grid layout for IRadioButtons
			return Div(Classes("w-full grid grid-flow-col justify-stretch gap-px", If(c.error != nil, func() string { return "border-l-8 border-red-600" })))(
				Map(c.options, func(option *AOption, index int) string {
					return Div(
						Classes(c.size, c.button, If(c.disabled, func() string { return "opacity-50 pointer-events-none" }), Or(value == option.ID, func() string { return c.buttonActive }, func() string { return c.buttonInactive })),
						Attr{
							Target: c.target.ID,
							OnClick: Trim(fmt.Sprintf(`
								const buttons = document.querySelectorAll('[target=%s]');
								buttons.forEach(button => button.classList.value = '%s');

								event.target.classList.value = '%s';

								const el = document.getElementById('%s');
								if (el == null)
									return;
							
								el.value = '%s';
								el.dispatchEvent(new Event('change'));

							`, c.target.ID, Classes(c.size, c.button, c.buttonInactive), Classes(c.size, c.button, c.buttonActive), c.target.ID, option.ID)),
						},
					)(option.Value)
				}),
			)
		}),
	)
}

func IRadioButtons(name string, data ...any) *ARadio {
	var temp any

	if len(data) > 0 {
		temp = data[0]
	}

	return &ARadio{
		name:           name,
		size:           MD,
		target:         Target(),
		visible:        true,
		data:           temp,
		classLabel:     "text-sm text-xs mt-3 font-bold relative",
		button:         "border cursor-pointer flex items-center justify-center text-center",
		buttonActive:   "bg-gray-600 text-white border-black",
		buttonInactive: "bg-white text-black hover:bg-gray-600 hover:text-white",
	}
}

func IRadioDiv(name string, data ...any) *ARadio {
	var temp any

	if len(data) > 0 {
		temp = data[0]
	}

	return &ARadio{
		name:           name,
		size:           "",
		target:         Target(),
		visible:        true,
		data:           temp,
		classLabel:     "text-sm text-xs mt-3 font-bold relative",
		button:         "cursor-pointer flex items-center justify-center text-center rounded-xl",
		buttonActive:   "border border-blue-400 rounded-xl",
		buttonInactive: "",
		radioPosition:  "absolute bottom-2 right-2",
	}
}
