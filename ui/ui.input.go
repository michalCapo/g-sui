package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type TInput struct {
	dates struct {
		Min time.Time
		Max time.Time
	}
	data         any
	Render       func(text string) string
	placeholder  string
	class        string
	classLabel   string
	classInput   string
	autocomplete string
	size         string
	onclick      string
	onchange     string
	as           string
	name         string
	pattern      string
	value        string
	valueFormat  string
	form         string
	error        validator.FieldError
	target       Attr
	numbers      struct {
		Min  float64
		Max  float64
		Step float64
	}
	visible        bool
	required       bool
	disabled       bool
	readonly       bool
	emptyOnDefault bool
}

func (c *TInput) Format(value string) *TInput {
	c.valueFormat = value
	return c
}

func (c *TInput) Rows(value uint8) *TInput {
	c.target.Rows = value
	return c
}

func (c *TInput) If(value bool) *TInput {
	c.visible = value
	return c
}

func (c *TInput) Value(value string) *TInput {
	c.value = value
	return c
}

func (c *TInput) Form(value string) *TInput {
	c.form = value
	return c
}

func (c *TInput) Type(value string) *TInput {
	c.as = value
	return c
}

func (c *TInput) Class(value ...string) *TInput {
	c.class = strings.Join(value, " ")
	return c
}

func (c *TInput) ClassInput(value ...string) *TInput {
	c.classInput = strings.Join(value, " ")
	return c
}

func (c *TInput) ClassLabel(value ...string) *TInput {
	c.classLabel = strings.Join(value, " ")
	return c
}

func (c *TInput) Size(value string) *TInput {
	c.size = value
	return c
}

func (c *TInput) Placeholder(value string) *TInput {
	c.placeholder = value
	return c
}

func (c *TInput) Pattern(value string) *TInput {
	c.pattern = value
	return c
}

func (c *TInput) Autocomplete(value string) *TInput {
	c.autocomplete = value
	return c
}

func (c *TInput) Required(value ...bool) *TInput {
	if value == nil {
		c.required = true
		return c
	}

	c.required = value[0]
	return c
}

func (c *TInput) Error(errs *error) *TInput {
	if errs == nil || *errs == nil {
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

func (c *TInput) Readonly(value ...bool) *TInput {
	if value == nil {
		c.readonly = true
		return c
	}

	c.readonly = value[0]
	return c
}

func (c *TInput) Disabled(value ...bool) *TInput {
	if value == nil {
		c.disabled = true
		return c
	}

	c.disabled = value[0]
	return c
}

func (c *TInput) EmptyOnDefault() *TInput {
	c.emptyOnDefault = true
	return c
}

func (c *TInput) Change(action string) *TInput {
	// if action.Target.Id == "" {
	// 	action.Target = c.target
	// }

	c.onchange = action

	return c
}

func (c *TInput) Click(action string) *TInput {
	// if action.Target.Id == "" {
	// 	action.Target = c.target
	// }

	c.onclick = action

	return c
}

func (c *TInput) Numbers(min float64, max float64, step float64) *TInput {
	c.numbers.Min = min
	c.numbers.Max = max
	c.numbers.Step = step

	return c
}

func (c *TInput) Dates(min time.Time, max time.Time) *TInput {
	c.dates.Min = min
	c.dates.Max = max

	return c
}

func IText(name string, data ...any) *TInput {
	c := &TInput{
		as:      "text",
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

		if c.data != nil {
			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			// if tmp.IsValid() {
			// 	value = tmp.String()
			// }

			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				value = fmt.Sprintf("%v", tmp.Interface())
			}
		}

		// Use c.value as fallback if value is still empty
		if value == "" && c.value != "" {
			value = c.value
		}

		// if c.error != nil { c.class = Classes(c.class, "border-4 border-red-600") }

		return Div(c.class)(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Input(
				Classes(INPUT, c.size, c.classInput,
					If(c.disabled, func() string { return DISABLED }),
					If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
					If(c.readonly, func() string { return "cursor-text pointer-events-none" }),
				),
				Attr{
					ID:           c.target.ID,
					Name:         c.name,
					Type:         c.as,
					OnChange:     c.onchange,
					OnClick:      c.onclick,
					Required:     c.required,
					Disabled:     c.disabled,
					Readonly:     c.readonly,
					Value:        value,
					Pattern:      c.pattern,
					Placeholder:  c.placeholder,
					Autocomplete: c.autocomplete,
					Form:         c.form,
				},
			),

			// ErrorField(c.error),
		)
	}

	return c
}

var IPhone = func(name string, data ...any) *TInput {
	return IText(name, data...).
		Type("tel").
		Autocomplete("tel").
		Placeholder("+421").
		Pattern("\\+[0-9]{10,14}")
}

var IEmail = func(name string, data ...any) *TInput {
	return IText(name, data...).
		Type("email").
		Autocomplete("email").
		Placeholder("name@gmail.com")
	// Pattern("^[a-z0-9._%+-]+@[a-z0-9-]+\\.[a-z0-9]{2,}$")
}

func IArea(name string, data ...any) *TInput {
	c := &TInput{
		as:      "text",
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

		if c.data != nil {
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				value = fmt.Sprintf("%v", tmp.Interface())
			}

			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			// if tmp.IsValid() {
			// 	value = tmp.String()
			// }
		}

		rows := uint8(5)

		if c.target.Rows > 0 {
			rows = uint8(c.target.Rows)
		}

		return Div(c.class)(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Textarea(
				Classes(AREA, c.size,
					If(c.disabled, func() string { return DISABLED }),
					If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
					If(c.readonly, func() string { return "cursor-default" }),
				),
				Attr{
					Rows: rows,

					Type:        c.as,
					ID:          c.target.ID,
					Name:        c.name,
					OnClick:     c.onclick,
					Required:    c.required,
					Disabled:    c.disabled,
					Readonly:    c.readonly,
					Placeholder: c.placeholder,
					Form:        c.form,
				},
			)(value),
		)
	}

	return c
}

func IPassword(name string, data ...any) *TInput {
	c := &TInput{
		as:      "password",
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

		if c.data != nil {
			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			// if tmp.IsValid() {
			// 	value = tmp.String()
			// }

			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				value = fmt.Sprintf("%v", tmp.Interface())
			}
		}

		return Div("")(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Input(
				Classes(INPUT, c.size, c.class, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" })),
				Attr{
					Value: value,

					Type:        c.as,
					ID:          c.target.ID,
					Name:        c.name,
					OnClick:     c.onclick,
					Required:    c.required,
					Disabled:    c.disabled,
					Placeholder: c.placeholder,
					Form:        c.form,
				},
			),
		)
	}

	return c
}

func IDate(name string, data ...any) *TInput {
	c := &TInput{
		as:         "date",
		target:     Target(),
		name:       name,
		size:       MD,
		visible:    true,
		classInput: " text-left min-w-0 appearance-none max-w-full",
	}

	if len(data) > 0 {
		c.data = data[0]
	}

	c.Render = func(text string) string {
		if !c.visible {
			return ""
		}

		min := ""
		max := ""
		value := ""

		if c.data != nil {
			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			tmp, err := PathValue(c.data, c.name)

			if err == nil {
				if timeValue, ok := tmp.Interface().(time.Time); ok {
					if !timeValue.IsZero() {
						value = timeValue.Format(time.DateOnly)
						// value = fmt.Sprintf("%04d-%02d-%02d", timeValue.Year(), timeValue.Month(), timeValue.Day())
					}
				}
			}
		}

		if !c.dates.Min.IsZero() {
			min = c.dates.Min.Format(time.DateOnly)
		}

		if !c.dates.Max.IsZero() {
			max = c.dates.Max.Format(time.DateOnly)
		}

		// Generate JavaScript validation/clamping for Safari where min/max may be ignored in picker UI
		onChangeWithValidation := c.onchange
		if !c.dates.Min.IsZero() || !c.dates.Max.IsZero() {
			minDate := ""
			maxDate := ""
			if !c.dates.Min.IsZero() {
				minDate = c.dates.Min.Format(time.DateOnly)
			}
			if !c.dates.Max.IsZero() {
				maxDate = c.dates.Max.Format(time.DateOnly)
			}

			validationJS := fmt.Sprintf(`
				(function(){
					var v = this.value;
					var min = '%s';
					var max = '%s';
					if (!v) { return; }
					this.setCustomValidity('');
					// Compare ISO dates lexicographically (YYYY-MM-DD)
					if (min && v < min) { this.value = min; v = min; }
					if (max && v > max) { this.value = max; v = max; }
					if (this.reportValidity) { this.reportValidity(); }
				}).call(this)
			`, minDate, maxDate)

			if c.onchange != "" {
				onChangeWithValidation = validationJS + "; " + c.onchange
			} else {
				onChangeWithValidation = validationJS
			}
		}

		return Div(Classes(c.class, "min-w-0"))(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Input(
				Classes(INPUT, c.size, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" }), "min-w-0 max-w-full", c.classInput),
				Attr{
					Min:   min,
					Max:   max,
					Value: value,

					Type:        c.as,
					ID:          c.target.ID,
					Name:        c.name,
					OnClick:     c.onclick,
					OnChange:    onChangeWithValidation,
					Required:    c.required,
					Disabled:    c.disabled,
					Placeholder: c.placeholder,
					Form:        c.form,
				},
			),
		)
	}
	return c
}

func ITime(name string, data ...any) *TInput {
	c := &TInput{
		as:      "time",
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

		min := ""
		max := ""
		value := ""

		if c.data != nil {
			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			tmp, err := PathValue(c.data, c.name)

			if err == nil {
				if timeValue, ok := tmp.Interface().(time.Time); ok {
					// Only set value if the time is not zero
					if !timeValue.IsZero() {
						value = timeValue.Format("15:04")
					} else {
						// Set default time to 00:00 for time inputs
						value = "00:00"
					}
				}
			}
		}

		if !c.dates.Min.IsZero() {
			min = c.dates.Min.Format("15:04")
		}

		if !c.dates.Max.IsZero() {
			max = c.dates.Max.Format("15:04")
		}

		return Div("")(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Input(
				Classes(INPUT, c.size, c.class, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" })),
				Attr{
					Min:   min,
					Max:   max,
					Value: value,

					Type:        c.as,
					ID:          c.target.ID,
					Name:        c.name,
					OnClick:     c.onclick,
					Required:    c.required,
					Disabled:    c.disabled,
					Placeholder: c.placeholder,
					Form:        c.form,
				},
			),
		)
	}
	return c
}

func IDateTime(name string, data ...any) *TInput {
	c := &TInput{
		as:      "datetime-local",
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

		min := ""
		max := ""
		value := ""

		if c.data != nil {
			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			tmp, err := PathValue(c.data, c.name)

			if err == nil {
				if timeValue, ok := tmp.Interface().(time.Time); ok {
					// Only set value if the time is not zero
					if !timeValue.IsZero() {
						value = timeValue.Format("2006-01-02T15:04")
					}
				}
			}
		}

		if !c.dates.Min.IsZero() {
			min = c.dates.Min.Format("2006-01-02T15:04")
		}

		if !c.dates.Max.IsZero() {
			max = c.dates.Max.Format("2006-01-02T15:04")
		}

		return Div("")(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Input(
				Classes(INPUT, c.size, c.class, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" })),
				Attr{
					Min:   min,
					Max:   max,
					Value: value,

					Type:        c.as,
					ID:          c.target.ID,
					Name:        c.name,
					OnClick:     c.onclick,
					Required:    c.required,
					Disabled:    c.disabled,
					Placeholder: c.placeholder,
					Form:        c.form,
				},
			),
		)
	}
	return c
}

func isZeroValue(val any) bool {
	switch v := val.(type) {
	case int:
		return v == 0
	case int8:
		return v == 0
	case int16:
		return v == 0
	case int32:
		return v == 0
	case int64:
		return v == 0
	case uint:
		return v == 0
	case uint8:
		return v == 0
	case uint16:
		return v == 0
	case uint32:
		return v == 0
	case uint64:
		return v == 0
	case float32:
		return v == 0
	case float64:
		return v == 0
	}
	return false
}

func INumber(name string, data ...any) *TInput {
	c := &TInput{
		as:          "number",
		target:      Target(),
		name:        name,
		size:        MD,
		visible:     true,
		valueFormat: "%v",
	}

	if len(data) > 0 {
		c.data = data[0]
	}

	c.Render = func(text string) string {
		if !c.visible {
			return ""
		}

		min := ""
		max := ""
		step := ""
		value := ""

		if c.data != nil {
			// v := reflect.ValueOf(c.data)

			// if v.Kind() == reflect.Ptr {
			// 	v = v.Elem()
			// }

			// tmp := v.FieldByName(c.name)

			// if tmp.IsValid() {
			// 	value = fmt.Sprintf("%v", tmp.Interface()) }

			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				// Convert to float64 if format requires it (e.g., %.2f)
				val := tmp.Interface()
				
				// Check if we should skip zero values
				if c.emptyOnDefault && isZeroValue(val) {
					// Leave value empty to show placeholder
				} else {
					if c.valueFormat != "%v" && c.valueFormat != "" {
						// Try to convert to float64 for numeric formatting
						switch v := val.(type) {
						case int:
							val = float64(v)
						case int8:
							val = float64(v)
						case int16:
							val = float64(v)
						case int32:
							val = float64(v)
						case int64:
							val = float64(v)
						case uint:
							val = float64(v)
						case uint8:
							val = float64(v)
						case uint16:
							val = float64(v)
						case uint32:
							val = float64(v)
						case uint64:
							val = float64(v)
						case float32:
							val = float64(v)
						}
					}
					value = fmt.Sprintf(c.valueFormat, val)
				}
			}
		}

		// Check if Numbers() was called by checking if max or step are set
		// (min can be 0, so we can't rely on it alone)
		hasNumbers := c.numbers.Max != 0 || c.numbers.Step != 0 || c.numbers.Min != 0

		if hasNumbers {
			// If Numbers() was called, render min even if it's 0
			min = fmt.Sprintf("%v", c.numbers.Min)
		}

		if c.numbers.Max != 0 {
			max = fmt.Sprintf("%v", c.numbers.Max)
		}

		if c.numbers.Step != 0 {
			step = fmt.Sprintf("%v", c.numbers.Step)
		}

		return Div(c.class)(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Input(
				Classes(INPUT, c.size, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" })),
				Attr{
					Min:   min,
					Max:   max,
					Step:  step,
					Value: value,

					Type:        c.as,
					ID:          c.target.ID,
					Name:        c.name,
					OnClick:     c.onclick,
					Required:    c.required,
					Disabled:    c.disabled,
					Placeholder: c.placeholder,
					Form:        c.form,
				},
			),

			// Script(fmt.Sprintf(`
			// 	(function() {
			// 		const input = document.getElementById('%v');
			// 		if (input.value) {
			// 			input.value = parseFloat(input.value).toFixed(2);
			// 		}
			// 	})();
			// `, c.target.ID)),
		)
	}
	return c
}

var Hidden = func(name string, value any, attr ...Attr) string {
	return Input("hidden", append(attr, Attr{Name: name, Type: "hidden", Value: fmt.Sprintf("%v", value)})...)
}

// func IValue(name string, data ...any) *TInput {
// 	c := &TInput{
// 		as:      "text",
// 		target:  Target(),
// 		name:    name,
// 		size:    MD,
// 		visible: true,
// 		data:    data[0],
// 	}

// 	c.Render = func(text string) string {
// 		if !c.visible {
// 			return ""
// 		}

// 		value := ""

// 		if c.data != nil {
// 			tmp, err := PathValue(c.data, c.name)
// 			if err == nil {
// 				value = fmt.Sprintf("%v", tmp.Interface())
// 			}
// 		}

// 		return Div(c.class)(
// 			Label(c.target).
// 				Required(c.required).
// 				Render(text),

// 			Div(
// 				Classes(VALUE, c.size, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" })),
// 				Attr{
// 					Id:           c.target.Id,
// 					Name:         c.name,
// 					Type:         c.as,
// 					OnChange:     c.onchange,
// 					OnClick:      c.onclick,
// 					Required:     c.required,
// 					Disabled:     c.disabled,
// 					Value:        value,
// 					Pattern:      c.pattern,
// 					Placeholder:  c.placeholder,
// 					Autocomplete: c.autocomplete,
// 				},
// 			)(),
// 		)
// 	}

// 	return c
// }

func IValue(attr ...Attr) *TInput {
	c := &TInput{
		target:  Target(),
		size:    MD,
		visible: true,
	}

	c.Render = func(text string) string {
		if !c.visible {
			return ""
		}

		attr = append(attr, Attr{
			ID:          c.target.ID,
			Name:        c.name,
			Required:    c.required,
			Disabled:    c.disabled,
			Pattern:     c.pattern,
			Placeholder: c.placeholder,
		})

		return Div(c.class)(
			Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text),

			Div(
				Classes(VALUE, c.size, If(c.disabled, func() string { return DISABLED }), If(c.error != nil, func() string { return "border-l-8 border-red-600" }), c.classInput),
				attr...,
			)(c.value),
		)
	}

	return c
}
