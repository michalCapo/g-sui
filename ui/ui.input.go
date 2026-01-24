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
	c.onchange = action
	return c
}

func (c *TInput) Click(action string) *TInput {
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
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				value = fmt.Sprintf("%v", tmp.Interface())
			}
		}
		if value == "" && c.value != "" {
			value = c.value
		}

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		inputJS := Input(
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
		)

		return Div(c.class)(labelJS, inputJS)
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
		}

		rows := uint8(5)
		if c.target.Rows > 0 {
			rows = uint8(c.target.Rows)
		}

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		textareaJS := Textarea(
			Classes(AREA, c.size,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
				If(c.readonly, func() string { return "cursor-default" }),
			),
			Attr{
				Rows:        rows,
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
		)(Text(value))

		return Div(c.class)(labelJS, textareaJS)
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
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				value = fmt.Sprintf("%v", tmp.Interface())
			}
		}

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		inputJS := Input(
			Classes(INPUT, c.size, c.class,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
			),
			Attr{
				Value:       value,
				Type:        c.as,
				ID:          c.target.ID,
				Name:        c.name,
				OnClick:     c.onclick,
				Required:    c.required,
				Disabled:    c.disabled,
				Placeholder: c.placeholder,
				Form:        c.form,
			},
		)

		return Div("")(labelJS, inputJS)
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
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				if timeValue, ok := tmp.Interface().(time.Time); ok {
					if !timeValue.IsZero() {
						value = timeValue.Format(time.DateOnly)
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

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		inputJS := Input(
			Classes(INPUT, c.size,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
				"min-w-0 max-w-full", c.classInput,
			),
			Attr{
				Min:         min,
				Max:         max,
				Value:       value,
				Type:        c.as,
				ID:          c.target.ID,
				Name:        c.name,
				OnClick:     c.onclick,
				OnChange:    c.onchange,
				Required:    c.required,
				Disabled:    c.disabled,
				Placeholder: c.placeholder,
				Form:        c.form,
			},
		)

		return Div(Classes(c.class, "min-w-0"))(labelJS, inputJS)
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
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				if timeValue, ok := tmp.Interface().(time.Time); ok {
					if !timeValue.IsZero() {
						value = timeValue.Format("15:04")
					} else {
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

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		inputJS := Input(
			Classes(INPUT, c.size, c.class,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
			),
			Attr{
				Min:         min,
				Max:         max,
				Value:       value,
				Type:        c.as,
				ID:          c.target.ID,
				Name:        c.name,
				OnClick:     c.onclick,
				Required:    c.required,
				Disabled:    c.disabled,
				Placeholder: c.placeholder,
				Form:        c.form,
			},
		)

		return Div("")(labelJS, inputJS)
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
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				if timeValue, ok := tmp.Interface().(time.Time); ok {
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

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		inputJS := Input(
			Classes(INPUT, c.size, c.class,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
			),
			Attr{
				Min:         min,
				Max:         max,
				Value:       value,
				Type:        c.as,
				ID:          c.target.ID,
				Name:        c.name,
				OnClick:     c.onclick,
				Required:    c.required,
				Disabled:    c.disabled,
				Placeholder: c.placeholder,
				Form:        c.form,
			},
		)

		return Div("")(labelJS, inputJS)
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
			tmp, err := PathValue(c.data, c.name)
			if err == nil {
				val := tmp.Interface()
				if c.emptyOnDefault && isZeroValue(val) {
					// Leave value empty
				} else {
					if c.valueFormat != "%v" && c.valueFormat != "" {
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

		hasNumbers := c.numbers.Max != 0 || c.numbers.Step != 0 || c.numbers.Min != 0
		if hasNumbers {
			min = fmt.Sprintf("%v", c.numbers.Min)
		}
		if c.numbers.Max != 0 {
			max = fmt.Sprintf("%v", c.numbers.Max)
		}
		if c.numbers.Step != 0 {
			step = fmt.Sprintf("%v", c.numbers.Step)
		}

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		inputJS := Input(
			Classes(INPUT, c.size,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
			),
			Attr{
				Min:         min,
				Max:         max,
				Step:        step,
				Value:       value,
				Type:        c.as,
				ID:          c.target.ID,
				Name:        c.name,
				OnClick:     c.onclick,
				Required:    c.required,
				Disabled:    c.disabled,
				Placeholder: c.placeholder,
				Form:        c.form,
			},
		)

		return Div(c.class)(labelJS, inputJS)
	}
	return c
}

var Hidden = func(name string, value any, attr ...Attr) string {
	return Input("hidden", append(attr, Attr{Name: name, Type: "hidden", Value: fmt.Sprintf("%v", value)})...)
}

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

		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		divJS := Div(
			Classes(VALUE, c.size,
				If(c.disabled, func() string { return DISABLED }),
				If(c.error != nil, func() string { return "border-l-8 border-red-600" }),
				c.classInput,
			),
			attr...,
		)(Text(c.value))

		return Div(c.class)(labelJS, divJS)
	}

	return c
}

// TFile represents a file input component
type TFile struct {
	target     Attr
	name       string
	form       string
	class      string
	classLabel string
	classInput string
	accept     string // MIME types: "image/*", ".pdf,.doc", etc.
	multiple   bool
	required   bool
	disabled   bool
	visible    bool
	onchange   string
	Render     func(text string) string
	// Zone mode fields
	zoneEnabled bool
	zoneIcon    string // Icon CSS classes
	zoneTitle   string // Primary text
	zoneHint    string // Secondary text
	zoneContent string // Custom HTML content (overrides icon/title/hint)
	classZone   string // Zone container classes
}

func IFile(name string) *TFile {
	c := &TFile{
		target:  Target(),
		name:    name,
		visible: true,
	}

	c.Render = func(text string) string {
		if !c.visible {
			return ""
		}

		// Zone mode: render dropzone-style UI
		if c.zoneEnabled || c.zoneContent != "" {
			// Render label above zone (if provided)
			labelJS := ""
			if text != "" {
				labelJS = Label(&c.target).
					Class(c.classLabel).
					ClassLabel("text-gray-600").
					Required(c.required).
					Render(text)
			}

			// Build input attributes
			attr := Attr{
				ID:       c.target.ID,
				Name:     c.name,
				Type:     "file",
				Form:     c.form,
				Required: c.required,
				Disabled: c.disabled,
				OnChange: c.onchange,
			}

			// Hide the actual file input
			inputClass := Classes(
				"sr-only",
				c.classInput,
				If(c.disabled, func() string { return DISABLED }),
			)
			inputHTML := Input(inputClass, attr)

			// Build zone container content
			var zoneContent string
			if c.zoneContent != "" {
				// Use custom content
				zoneContent = c.zoneContent
			} else {
				// Build default content with icon, title, hint
				var iconHTML string
				if c.zoneIcon != "" {
					iconHTML = Div(Classes("flex items-center justify-center", c.zoneIcon))()
				} else {
					// Default upload icon (simple circle with arrow)
					iconHTML = Div("w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center")(
						Div("w-6 h-6 text-gray-500 dark:text-gray-400")(
							`<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path></svg>`,
						),
					)
				}

				var titleHTML string
				if c.zoneTitle != "" {
					titleHTML = Div("font-semibold text-gray-900 dark:text-gray-100")(Text(c.zoneTitle))
				}

				var hintHTML string
				if c.zoneHint != "" {
					hintHTML = Div("text-sm text-gray-500 dark:text-gray-400")(Text(c.zoneHint))
				}

				zoneContent = Div("flex flex-col items-center justify-center gap-2")(
					iconHTML,
					titleHTML,
					hintHTML,
				)
			}

			// Build zone container with default styling
			defaultZoneClass := Classes(
				"border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl p-8",
				"hover:border-gray-400 hover:bg-gray-50 dark:hover:bg-gray-900",
				"dark:hover:border-gray-500",
				"transition-colors cursor-pointer",
				"flex flex-col items-center justify-center gap-2",
				If(c.disabled, func() string { return "opacity-50 cursor-not-allowed" }),
			)
			zoneClass := Classes(defaultZoneClass, c.classZone)

			// Create clickable label wrapper for the zone
			zoneLabel := El("label",
				zoneClass,
				Attr{For: c.target.ID},
			)(zoneContent)

			// Add accept and multiple attributes via script
			var script string
			if c.accept != "" || c.multiple {
				var attrs []string
				if c.accept != "" {
					acceptEscaped := strings.ReplaceAll(c.accept, `\`, `\\`)
					acceptEscaped = strings.ReplaceAll(acceptEscaped, `'`, `\'`)
					acceptEscaped = strings.ReplaceAll(acceptEscaped, `"`, `\"`)
					attrs = append(attrs, fmt.Sprintf(`el.setAttribute('accept','%s');`, acceptEscaped))
				}
				if c.multiple {
					attrs = append(attrs, `el.setAttribute('multiple','multiple');`)
				}
				idEscaped := strings.ReplaceAll(c.target.ID, `\`, `\\`)
				idEscaped = strings.ReplaceAll(idEscaped, `'`, `\'`)
				idEscaped = strings.ReplaceAll(idEscaped, `"`, `\"`)
				script = Script(fmt.Sprintf(`(function(){var el=document.getElementById('%s');if(el){%s}})();`,
					idEscaped, strings.Join(attrs, "")))
			}

			return Div(c.class)(labelJS, inputHTML, zoneLabel, script)
		}

		// Standard mode: render traditional file input
		labelJS := Label(&c.target).
			Class(c.classLabel).
			ClassLabel("text-gray-600").
			Required(c.required).
			Render(text)

		// Build input attributes
		inputClass := Classes(
			"block w-full text-sm text-gray-500",
			"file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0",
			"file:text-sm file:font-semibold",
			"file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100",
			"dark:file:bg-blue-900 dark:file:text-blue-300 dark:hover:file:bg-blue-800",
			"cursor-pointer border border-gray-300 rounded",
			c.classInput,
			If(c.disabled, func() string { return DISABLED }),
		)

		attr := Attr{
			ID:       c.target.ID,
			Name:     c.name,
			Type:     "file",
			Form:     c.form,
			Required: c.required,
			Disabled: c.disabled,
			OnChange: c.onchange,
		}

		// Generate input HTML
		inputHTML := Input(inputClass, attr)

		// Add accept and multiple attributes via script (since Attr doesn't have these)
		var script string
		if c.accept != "" || c.multiple {
			var attrs []string
			if c.accept != "" {
				// Escape the accept string for JavaScript
				acceptEscaped := strings.ReplaceAll(c.accept, `\`, `\\`)
				acceptEscaped = strings.ReplaceAll(acceptEscaped, `'`, `\'`)
				acceptEscaped = strings.ReplaceAll(acceptEscaped, `"`, `\"`)
				attrs = append(attrs, fmt.Sprintf(`el.setAttribute('accept','%s');`, acceptEscaped))
			}
			if c.multiple {
				attrs = append(attrs, `el.setAttribute('multiple','multiple');`)
			}
			// Escape the ID for JavaScript
			idEscaped := strings.ReplaceAll(c.target.ID, `\`, `\\`)
			idEscaped = strings.ReplaceAll(idEscaped, `'`, `\'`)
			idEscaped = strings.ReplaceAll(idEscaped, `"`, `\"`)
			script = Script(fmt.Sprintf(`(function(){var el=document.getElementById('%s');if(el){%s}})();`,
				idEscaped, strings.Join(attrs, "")))
		}

		return Div(c.class)(labelJS, inputHTML, script)
	}

	return c
}

// Accept sets allowed file types (MIME types or extensions)
// Examples: "image/*", ".pdf,.doc", "image/png,image/jpeg"
func (c *TFile) Accept(types string) *TFile {
	c.accept = types
	return c
}

// Multiple allows selecting multiple files
func (c *TFile) Multiple() *TFile {
	c.multiple = true
	return c
}

// Required marks the field as required
func (c *TFile) Required(value ...bool) *TFile {
	if len(value) == 0 {
		c.required = true
	} else {
		c.required = value[0]
	}
	return c
}

// Disabled disables the input
func (c *TFile) Disabled(value ...bool) *TFile {
	if len(value) == 0 {
		c.disabled = true
	} else {
		c.disabled = value[0]
	}
	return c
}

// Form associates the input with a form by ID
func (c *TFile) Form(value string) *TFile {
	c.form = value
	return c
}

// Class sets the wrapper div class
func (c *TFile) Class(value ...string) *TFile {
	c.class = strings.Join(value, " ")
	return c
}

// ClassInput sets the input element class
func (c *TFile) ClassInput(value ...string) *TFile {
	c.classInput = strings.Join(value, " ")
	return c
}

// ClassLabel sets the label class
func (c *TFile) ClassLabel(value ...string) *TFile {
	c.classLabel = strings.Join(value, " ")
	return c
}

// If conditionally renders the component
func (c *TFile) If(value bool) *TFile {
	c.visible = value
	return c
}

// Change sets the onchange handler
func (c *TFile) Change(action string) *TFile {
	c.onchange = action
	return c
}

// ID sets a custom ID for the file input (useful for linking with ImagePreview)
// If not set, an auto-generated ID will be used
func (c *TFile) ID(id string) *TFile {
	c.target.ID = id
	return c
}

// GetID returns the file input's ID (useful for linking with ImagePreview)
func (c *TFile) GetID() string {
	return c.target.ID
}

// Zone enables dropzone mode with title and hint text
func (c *TFile) Zone(title, hint string) *TFile {
	c.zoneEnabled = true
	c.zoneTitle = title
	c.zoneHint = hint
	return c
}

// ZoneIcon sets the icon CSS classes for zone mode
func (c *TFile) ZoneIcon(classes string) *TFile {
	c.zoneIcon = classes
	return c
}

// ZoneContent sets completely custom HTML content for zone mode (overrides icon/title/hint)
func (c *TFile) ZoneContent(html string) *TFile {
	c.zoneContent = html
	c.zoneEnabled = true
	return c
}

// ClassZone sets the zone container CSS classes
func (c *TFile) ClassZone(classes ...string) *TFile {
	c.classZone = strings.Join(classes, " ")
	return c
}

// TImagePreview represents an image preview component for file inputs
type TImagePreview struct {
	inputID  string // ID of the file input to listen to
	id       string // Container ID
	maxSize  string // Max dimensions (e.g., "320px")
	multiple bool   // Show grid for multiple images
	class    string // Wrapper classes
	visible  bool
}

// ImagePreview creates a new image preview component linked to a file input
func ImagePreview(inputID string) *TImagePreview {
	return &TImagePreview{
		inputID: inputID,
		id:      Target().ID,
		maxSize: "320px",
		visible: true,
	}
}

// Multiple enables multi-image grid layout
func (c *TImagePreview) Multiple() *TImagePreview {
	c.multiple = true
	return c
}

// MaxSize sets the maximum image dimensions
func (c *TImagePreview) MaxSize(size string) *TImagePreview {
	c.maxSize = size
	return c
}

// Class sets the wrapper div classes
func (c *TImagePreview) Class(value ...string) *TImagePreview {
	c.class = strings.Join(value, " ")
	return c
}

// If conditionally renders the component
func (c *TImagePreview) If(value bool) *TImagePreview {
	c.visible = value
	return c
}

// Render generates the HTML and JavaScript for the image preview
func (c *TImagePreview) Render() string {
	if !c.visible {
		return ""
	}

	// Escape IDs for JavaScript
	inputIDEscaped := escapeJS(c.inputID)
	containerIDEscaped := escapeJS(c.id)
	maxSizeEscaped := escapeJS(c.maxSize)

	// Determine container classes based on multiple mode
	containerClass := "mt-2"
	if c.multiple {
		containerClass = Classes(containerClass, "grid grid-cols-3 gap-2")
	} else {
		containerClass = Classes(containerClass, "flex justify-center")
	}

	// Add custom classes
	if c.class != "" {
		containerClass = Classes(containerClass, c.class)
	}

	// Build the JavaScript that handles file preview
	jsCode := fmt.Sprintf(`(function(){
		var input = document.getElementById('%s');
		var container = document.getElementById('%s');
		if (!input || !container) return;
		
		input.addEventListener('change', function(e) {
			var files = e.target.files;
			if (!files || files.length === 0) {
				container.innerHTML = '';
				return;
			}
			
			container.innerHTML = '';
			
			for (var i = 0; i < files.length; i++) {
				var file = files[i];
				if (!file.type.startsWith('image/')) continue;
				
				var reader = new FileReader();
				reader.onload = function(readerEvent) {
					var img = document.createElement('img');
					img.src = readerEvent.target.result;
					img.alt = 'Preview';
					img.className = 'rounded border border-gray-300 dark:border-gray-700 object-cover';
					img.style.maxWidth = '%s';
					img.style.maxHeight = '%s';
					if (%t) {
						img.className += ' w-full h-32';
					} else {
						img.className += ' max-w-full';
					}
					container.appendChild(img);
				};
				reader.readAsDataURL(file);
			}
		});
	})();`, inputIDEscaped, containerIDEscaped, maxSizeEscaped, maxSizeEscaped, c.multiple)

	// Generate the container div and script
	containerHTML := Div(containerClass, Attr{ID: c.id})()
	scriptHTML := Script(jsCode)

	return containerHTML + scriptHTML
}

// TImageUpload represents a combined image upload component with built-in preview
// It combines TFile functionality with TImagePreview for a single, image-specific component
type TImageUpload struct {
	target       Attr
	name         string
	form         string
	class        string
	classLabel   string
	classInput   string
	accept       string // defaults to "image/*"
	multiple     bool
	required     bool
	disabled     bool
	visible      bool
	onchange     string
	// Zone mode fields
	zoneEnabled  bool
	zoneIcon     string
	zoneTitle    string
	zoneHint     string
	zoneContent  string
	classZone    string
	// Preview fields (from TImagePreview)
	maxSize      string // e.g., "320px"
	classPreview string
	previewID    string // Container ID for preview (deprecated, kept for compatibility)
	zoneUploadID string // ID for upload zone container
	zonePreviewID string // ID for preview zone container
	Render       func(text string) string
}

// IImageUpload creates a new image upload component with default accept="image/*"
func IImageUpload(name string) *TImageUpload {
	c := &TImageUpload{
		target:       Target(),
		name:         name,
		accept:       "image/*",
		maxSize:      "320px",
		visible:      true,
		previewID:    Target().ID,
		zoneUploadID: Target().ID,
		zonePreviewID: Target().ID,
	}

	c.Render = func(text string) string {
		if !c.visible {
			return ""
		}

		// Always use zone mode for inline preview
		// If zone is not explicitly enabled, enable it with defaults
		useZone := c.zoneEnabled || c.zoneContent != ""
		if !useZone {
			// Auto-enable zone mode with default icon/title/hint
			useZone = true
			if c.zoneTitle == "" {
				c.zoneTitle = "Add Image"
			}
			if c.zoneHint == "" {
				c.zoneHint = "Click to upload"
			}
		}

		var labelJS string
		if text != "" {
			labelJS = Label(&c.target).
				Class(c.classLabel).
				ClassLabel("text-gray-600").
				Required(c.required).
				Render(text)
		}

		// Build input attributes
		attr := Attr{
			ID:       c.target.ID,
			Name:     c.name,
			Type:     "file",
			Form:     c.form,
			Required: c.required,
			Disabled: c.disabled,
			OnChange: c.onchange,
		}

		// Hide the actual file input
		inputClass := Classes(
			"sr-only",
			c.classInput,
			If(c.disabled, func() string { return DISABLED }),
		)
		inputHTML := Input(inputClass, attr)

		// Build upload zone content (shown initially)
		var zoneContent string
		if c.zoneContent != "" {
			// Use custom content
			zoneContent = c.zoneContent
		} else {
			// Build default content with icon, title, hint
			var iconHTML string
			if c.zoneIcon != "" {
				iconHTML = Div(Classes("flex items-center justify-center", c.zoneIcon))()
			} else {
				// Default upload icon (simple circle with arrow)
				iconHTML = Div("w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center")(
					Div("w-6 h-6 text-gray-500 dark:text-gray-400")(
						`<svg fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path></svg>`,
					),
				)
			}

			var titleHTML string
			if c.zoneTitle != "" {
				titleHTML = Div("font-semibold text-gray-900 dark:text-gray-100")(Text(c.zoneTitle))
			}

			var hintHTML string
			if c.zoneHint != "" {
				hintHTML = Div("text-sm text-gray-500 dark:text-gray-400")(Text(c.zoneHint))
			}

			zoneContent = Div("flex flex-col items-center justify-center gap-2")(
				iconHTML,
				titleHTML,
				hintHTML,
			)
		}

		// Build zone container with default styling
		defaultZoneClass := Classes(
			"border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-xl p-8",
			"hover:border-gray-400 hover:bg-gray-50 dark:hover:bg-gray-900",
			"dark:hover:border-gray-500",
			"transition-colors cursor-pointer",
			"flex flex-col items-center justify-center gap-2",
			"relative",
			If(c.disabled, func() string { return "opacity-50 cursor-not-allowed" }),
		)
		zoneClass := Classes(defaultZoneClass, c.classZone)

		// Create upload zone (initially visible)
		uploadZoneLabel := El("label",
			Classes(zoneClass, ""),
			Attr{For: c.target.ID, ID: c.zoneUploadID},
		)(zoneContent)

		// Create preview zone (initially hidden) with image and change button
		previewZoneClass := Classes(
			zoneClass,
			"hidden",
			"border-solid",
		)
		changeButtonID := Target().ID
		previewImgID := Target().ID

		// Create change button using El
		changeBtnHTML := El("button",
			"mt-3 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm font-medium transition-colors",
			Attr{
				ID:    changeButtonID,
				Type:  "button",
			},
		)(Text("Change Image"))

		previewImgHTML := Img("rounded border border-gray-300 dark:border-gray-700 object-cover max-w-full", Attr{
			ID:    previewImgID,
			Alt:   "Preview",
			Style: fmt.Sprintf("max-width: %s; max-height: %s;", c.maxSize, c.maxSize),
		})

		previewZoneContent := Div("flex flex-col items-center justify-center gap-2 w-full")(
			previewImgHTML,
			changeBtnHTML,
		)

		previewZoneDiv := Div(previewZoneClass, Attr{ID: c.zonePreviewID})(previewZoneContent)

		// Add accept and multiple attributes via script
		var script string
		if c.accept != "" || c.multiple {
			var attrs []string
			if c.accept != "" {
				acceptEscaped := strings.ReplaceAll(c.accept, `\`, `\\`)
				acceptEscaped = strings.ReplaceAll(acceptEscaped, `'`, `\'`)
				acceptEscaped = strings.ReplaceAll(acceptEscaped, `"`, `\"`)
				attrs = append(attrs, fmt.Sprintf(`el.setAttribute('accept','%s');`, acceptEscaped))
			}
			if c.multiple {
				attrs = append(attrs, `el.setAttribute('multiple','multiple');`)
			}
			idEscaped := strings.ReplaceAll(c.target.ID, `\`, `\\`)
			idEscaped = strings.ReplaceAll(idEscaped, `'`, `\'`)
			idEscaped = strings.ReplaceAll(idEscaped, `"`, `\"`)
			script = Script(fmt.Sprintf(`(function(){var el=document.getElementById('%s');if(el){%s}})();`,
				idEscaped, strings.Join(attrs, "")))
		}

		// Escape IDs for JavaScript
		inputIDEscaped := escapeJS(c.target.ID)
		uploadZoneIDEscaped := escapeJS(c.zoneUploadID)
		previewZoneIDEscaped := escapeJS(c.zonePreviewID)
		previewImgIDEscaped := escapeJS(previewImgID)
		changeBtnIDEscaped := escapeJS(changeButtonID)
		maxSizeEscaped := escapeJS(c.maxSize)

		// Build JavaScript to handle file selection and toggle between states
		jsCode := fmt.Sprintf(`(function(){
			var input = document.getElementById('%s');
			var uploadZone = document.getElementById('%s');
			var previewZone = document.getElementById('%s');
			var previewImg = document.getElementById('%s');
			var changeBtn = document.getElementById('%s');
			
			if (!input || !uploadZone || !previewZone || !previewImg || !changeBtn) return;
			
			// Handle file selection
			input.addEventListener('change', function(e) {
				var files = e.target.files;
				if (!files || files.length === 0) {
					// No file selected - show upload zone, hide preview
					uploadZone.classList.remove('hidden');
					previewZone.classList.add('hidden');
					return;
				}
				
				var file = files[0];
				if (!file.type.startsWith('image/')) {
					// Not an image - show upload zone
					uploadZone.classList.remove('hidden');
					previewZone.classList.add('hidden');
					return;
				}
				
				var reader = new FileReader();
				reader.onload = function(readerEvent) {
					previewImg.src = readerEvent.target.result;
					previewImg.style.maxWidth = '%s';
					previewImg.style.maxHeight = '%s';
					// Hide upload zone, show preview
					uploadZone.classList.add('hidden');
					previewZone.classList.remove('hidden');
				};
				reader.readAsDataURL(file);
			});
			
			// Handle change button click
			changeBtn.addEventListener('click', function(e) {
				e.preventDefault();
				e.stopPropagation();
				input.click();
			});
		})();`, inputIDEscaped, uploadZoneIDEscaped, previewZoneIDEscaped, previewImgIDEscaped, changeBtnIDEscaped, maxSizeEscaped, maxSizeEscaped)

		previewScript := Script(jsCode)

		return Div(c.class)(labelJS, inputHTML, uploadZoneLabel, previewZoneDiv, script, previewScript)
	}

	return c
}

// Accept sets allowed file types (MIME types or extensions)
// Defaults to "image/*" but can be overridden (e.g., "image/png,image/jpeg")
func (c *TImageUpload) Accept(types string) *TImageUpload {
	c.accept = types
	return c
}

// Multiple allows selecting multiple files
func (c *TImageUpload) Multiple() *TImageUpload {
	c.multiple = true
	return c
}

// Required marks the field as required
func (c *TImageUpload) Required(value ...bool) *TImageUpload {
	if len(value) == 0 {
		c.required = true
	} else {
		c.required = value[0]
	}
	return c
}

// Disabled disables the input
func (c *TImageUpload) Disabled(value ...bool) *TImageUpload {
	if len(value) == 0 {
		c.disabled = true
	} else {
		c.disabled = value[0]
	}
	return c
}

// Form associates the input with a form by ID
func (c *TImageUpload) Form(value string) *TImageUpload {
	c.form = value
	return c
}

// Class sets the wrapper div class
func (c *TImageUpload) Class(value ...string) *TImageUpload {
	c.class = strings.Join(value, " ")
	return c
}

// ClassInput sets the input element class
func (c *TImageUpload) ClassInput(value ...string) *TImageUpload {
	c.classInput = strings.Join(value, " ")
	return c
}

// ClassLabel sets the label class
func (c *TImageUpload) ClassLabel(value ...string) *TImageUpload {
	c.classLabel = strings.Join(value, " ")
	return c
}

// If conditionally renders the component
func (c *TImageUpload) If(value bool) *TImageUpload {
	c.visible = value
	return c
}

// Change sets the onchange handler
func (c *TImageUpload) Change(action string) *TImageUpload {
	c.onchange = action
	return c
}

// ID sets a custom ID for the file input
func (c *TImageUpload) ID(id string) *TImageUpload {
	c.target.ID = id
	return c
}

// GetID returns the file input's ID
func (c *TImageUpload) GetID() string {
	return c.target.ID
}

// Zone enables dropzone mode with title and hint text
func (c *TImageUpload) Zone(title, hint string) *TImageUpload {
	c.zoneEnabled = true
	c.zoneTitle = title
	c.zoneHint = hint
	return c
}

// ZoneIcon sets the icon CSS classes for zone mode
func (c *TImageUpload) ZoneIcon(classes string) *TImageUpload {
	c.zoneIcon = classes
	return c
}

// ZoneContent sets completely custom HTML content for zone mode (overrides icon/title/hint)
func (c *TImageUpload) ZoneContent(html string) *TImageUpload {
	c.zoneContent = html
	c.zoneEnabled = true
	return c
}

// ClassZone sets the zone container CSS classes
func (c *TImageUpload) ClassZone(classes ...string) *TImageUpload {
	c.classZone = strings.Join(classes, " ")
	return c
}

// MaxSize sets the maximum image dimensions for preview (e.g., "320px")
func (c *TImageUpload) MaxSize(size string) *TImageUpload {
	c.maxSize = size
	return c
}

// ClassPreview sets the preview container CSS classes
func (c *TImageUpload) ClassPreview(classes ...string) *TImageUpload {
	c.classPreview = strings.Join(classes, " ")
	return c
}
