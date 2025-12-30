package ui

import (
	"fmt"
)

func ICheckbox(name string, data ...any) *TInput {
	c := &TInput{
		as:      "checkbox",
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

		if value == "true" {
			checked = "checked"
		}

		c.class = Classes(c.class, "flex items-center gap-2")

		if c.required && value == "false" {
			c.class = Classes(c.class, "invalid")
		}

		return Div(Classes(c.class, c.size))(
			Input(
				Classes("cursor-pointer select-none", If(c.disabled, func() string { return DISABLED })),
				Attr{
					Value:   value,
					Checked: checked,

					Type:     c.as,
					ID:       c.target.ID,
					Name:     c.name,
					Required: c.required,
					Disabled: c.disabled,
					OnClick:  c.onclick,
					Form:     c.form,
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
				ClassLabel("text-gray-600 cursor-pointer select-none").
				Render(text),
		)
	}

	return c
}
