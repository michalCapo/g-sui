package ui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// Form: declarative builder for validated forms
// ---------------------------------------------------------------------------

// FieldType enumerates supported field kinds.
type FieldType int

const (
	FieldText      FieldType = iota // <input type="text">
	FieldPassword                   // <input type="password">
	FieldEmail                      // <input type="email">
	FieldNumber                     // <input type="number">
	FieldPhone                      // <input type="tel">
	FieldDate                       // <input type="date">
	FieldTime                       // <input type="time">
	FieldDatetime                   // <input type="datetime-local">
	FieldUrl                        // <input type="url">
	FieldSearch                     // <input type="search">
	FieldTextarea                   // <textarea>
	FieldSelect                     // <select>
	FieldRadio                      // inline radios
	FieldRadioBtn                   // button-style radios (visible border cards)
	FieldRadioCard                  // hidden-peer card radios (peer-checked styling)
	FieldCheckbox                   // <input type="checkbox">
	FieldHidden                     // <input type="hidden">
)

// RadioStyle is an alias kept for readability in the API.
type RadioStyle = FieldType

const (
	RadioInline = FieldRadio
	RadioButton = FieldRadioBtn
	RadioCard   = FieldRadioCard
)

// FieldOption represents a value/label pair for select and radio fields.
type FieldOption struct {
	Value string
	Label string
}

// Field holds the declarative definition for one form field.
type Field struct {
	Type        FieldType
	Name        string // maps to Attr("name",...) and struct json tag
	Label       string
	Placeholder string
	Value       string
	Checked     bool // only for FieldCheckbox
	Required    bool
	Pattern     string // regex pattern for client+server validation
	ErrMsg      string // custom error message (defaults to "<Label> is required")
	Options     []FieldOption
	Class       string // override input class
	WrapClass   string // override wrapper class
}

// FormBuilder accumulates fields and submit buttons, then renders to Nodes.
type FormBuilder struct {
	id         string
	fields     []Field
	buttons    []formButton
	actionName string
	class      string // wrapper div class
	fieldClass string // default input class
	errClass   string // error text class
}

type formButton struct {
	action string // value sent as "Action" field
	label  string
	class  string
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// NewForm creates a new form builder with the given container ID.
func NewForm(id string) *FormBuilder {
	return &FormBuilder{
		id:         id,
		fieldClass: "w-full border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100",
		errClass:   "text-xs text-red-600 mt-1 hidden",
		class:      "flex flex-col gap-4",
	}
}

// ---------------------------------------------------------------------------
// Form-level configuration
// ---------------------------------------------------------------------------

// FormClass overrides the form wrapper div class.
func (f *FormBuilder) FormClass(cls string) *FormBuilder {
	f.class = cls
	return f
}

// InputClass sets the default CSS class for all text-like inputs.
func (f *FormBuilder) InputClass(cls string) *FormBuilder {
	f.fieldClass = cls
	return f
}

// ErrClass sets the CSS class for error messages.
func (f *FormBuilder) ErrClass(cls string) *FormBuilder {
	f.errClass = cls
	return f
}

// Action sets the WS action name that the form submits to.
func (f *FormBuilder) Action(name string) *FormBuilder {
	f.actionName = name
	return f
}

// ---------------------------------------------------------------------------
// Field adders — each returns a *FieldBuilder for chained configuration
// ---------------------------------------------------------------------------

// FieldBuilder provides chained configuration for a single field.
type FieldBuilder struct {
	form  *FormBuilder
	index int
}

func (f *FormBuilder) addField(ft FieldType, label, name string) *FieldBuilder {
	f.fields = append(f.fields, Field{
		Type:  ft,
		Name:  name,
		Label: label,
	})
	return &FieldBuilder{form: f, index: len(f.fields) - 1}
}

func (fb *FieldBuilder) field() *Field { return &fb.form.fields[fb.index] }

// Required marks the field as required for validation.
func (fb *FieldBuilder) Required() *FieldBuilder {
	fb.field().Required = true
	return fb
}

// Placeholder sets the input placeholder text.
func (fb *FieldBuilder) Placeholder(p string) *FieldBuilder {
	fb.field().Placeholder = p
	return fb
}

// Value sets the current field value (for re-rendering after submit).
func (fb *FieldBuilder) Value(v string) *FieldBuilder {
	fb.field().Value = v
	return fb
}

// IsChecked sets the checkbox checked state.
func (fb *FieldBuilder) IsChecked(c bool) *FieldBuilder {
	fb.field().Checked = c
	return fb
}

// Err sets a custom error message for the field.
func (fb *FieldBuilder) Err(msg string) *FieldBuilder {
	fb.field().ErrMsg = msg
	return fb
}

// PatternValidation sets a regex pattern for client+server validation.
func (fb *FieldBuilder) PatternValidation(p string) *FieldBuilder {
	fb.field().Pattern = p
	return fb
}

// Class overrides the input element's CSS class.
func (fb *FieldBuilder) Class(cls string) *FieldBuilder {
	fb.field().Class = cls
	return fb
}

// WrapClass overrides the wrapper div's CSS class.
func (fb *FieldBuilder) WrapClass(cls string) *FieldBuilder {
	fb.field().WrapClass = cls
	return fb
}

// Opts adds value:label option pairs. Format: "value:Label" or just "Label"
// (in which case value = lowercase label). An empty value like ":Select..."
// creates a placeholder option.
func (fb *FieldBuilder) Opts(pairs ...string) *FieldBuilder {
	for _, p := range pairs {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) == 2 {
			fb.field().Options = append(fb.field().Options, FieldOption{Value: parts[0], Label: parts[1]})
		} else {
			fb.field().Options = append(fb.field().Options, FieldOption{Value: strings.ToLower(p), Label: p})
		}
	}
	return fb
}

// Render returns the FormBuilder to continue chaining at the form level.
func (fb *FieldBuilder) Render() *FormBuilder {
	return fb.form
}

// ---------------------------------------------------------------------------
// Typed field shortcuts (return *FieldBuilder for chaining)
// ---------------------------------------------------------------------------

func (f *FormBuilder) Text(label, name string) *FieldBuilder {
	return f.addField(FieldText, label, name)
}
func (f *FormBuilder) Password(label, name string) *FieldBuilder {
	return f.addField(FieldPassword, label, name)
}
func (f *FormBuilder) Email(label, name string) *FieldBuilder {
	return f.addField(FieldEmail, label, name)
}
func (f *FormBuilder) Number(label, name string) *FieldBuilder {
	return f.addField(FieldNumber, label, name)
}
func (f *FormBuilder) Phone(label, name string) *FieldBuilder {
	return f.addField(FieldPhone, label, name)
}
func (f *FormBuilder) DateField(label, name string) *FieldBuilder {
	return f.addField(FieldDate, label, name)
}
func (f *FormBuilder) TimeField(label, name string) *FieldBuilder {
	return f.addField(FieldTime, label, name)
}
func (f *FormBuilder) DatetimeField(label, name string) *FieldBuilder {
	return f.addField(FieldDatetime, label, name)
}
func (f *FormBuilder) UrlField(label, name string) *FieldBuilder {
	return f.addField(FieldUrl, label, name)
}
func (f *FormBuilder) SearchField(label, name string) *FieldBuilder {
	return f.addField(FieldSearch, label, name)
}
func (f *FormBuilder) Area(label, name string) *FieldBuilder {
	return f.addField(FieldTextarea, label, name)
}
func (f *FormBuilder) SelectField(label, name string) *FieldBuilder {
	return f.addField(FieldSelect, label, name)
}
func (f *FormBuilder) Radio(label, name string) *FieldBuilder {
	return f.addField(FieldRadio, label, name)
}
func (f *FormBuilder) RadioBtn(label, name string) *FieldBuilder {
	return f.addField(FieldRadioBtn, label, name)
}
func (f *FormBuilder) RadioCard(label, name string) *FieldBuilder {
	return f.addField(FieldRadioCard, label, name)
}
func (f *FormBuilder) Checkbox(label, name string) *FieldBuilder {
	return f.addField(FieldCheckbox, label, name)
}
func (f *FormBuilder) Hidden(name string) *FieldBuilder {
	return f.addField(FieldHidden, "", name)
}

// Submit adds a submit button with an action value, label, and CSS class.
func (f *FormBuilder) Submit(action, label, class string) *FormBuilder {
	f.buttons = append(f.buttons, formButton{action: action, label: label, class: class})
	return f
}

// ---------------------------------------------------------------------------
// Render — builds the full Node tree
// ---------------------------------------------------------------------------

// Build renders the form into a Node tree. It emits a <form> element
// (with onsubmit prevented) and sets the HTML "form" attribute on every
// input, select, and textarea so they are formally associated. This lets
// multiple independent forms coexist on the same page without nesting
// <form> elements — inputs can live anywhere in the DOM.
func (f *FormBuilder) Build() *Node {
	children := make([]*Node, 0, len(f.fields)+1)

	for i := range f.fields {
		node := f.renderField(&f.fields[i])
		if node != nil {
			children = append(children, node)
		}
	}

	// Submit buttons — associated with the <form> via the form attribute
	if len(f.buttons) > 0 {
		btns := make([]*Node, len(f.buttons))
		for i, btn := range f.buttons {
			cls := btn.class
			if cls == "" {
				cls = "px-4 py-2 rounded bg-blue-600 text-white cursor-pointer hover:bg-blue-700 text-sm"
			}
			btns[i] = Button(cls).
				Attr("type", "button").
				Attr("form", f.id).
				Text(btn.label).
				OnClick(JS(f.buildValidateJS(btn.action)))
		}
		children = append(children, Div("flex gap-2").Render(btns...))
	}

	return Form(f.class).ID(f.id).
		OnSubmit(JS("event.preventDefault()")).
		Render(children...)
}

// ---------------------------------------------------------------------------
// Field rendering
// ---------------------------------------------------------------------------

// fieldID returns a unique element ID scoped to this form.
func (f *FormBuilder) fieldID(fld *Field) string {
	return f.id + "-" + fld.Name
}

// radioName returns a form-scoped radio group name so multiple forms on the
// same page don't share radio groups. The original field Name is sent to the
// server in the collected data object (see buildValidateJS).
func (f *FormBuilder) radioName(fld *Field) string {
	return f.id + "-" + fld.Name
}

func (f *FormBuilder) errID(fld *Field) string {
	return "err-" + f.id + "-" + fld.Name
}

func (f *FormBuilder) inputClass(fld *Field) string {
	if fld.Class != "" {
		return fld.Class
	}
	return f.fieldClass
}

// formAttr sets the HTML "form" attribute on a node to associate it with
// this form's <form> element. This allows inputs placed anywhere in the DOM
// to belong to a specific form — critical when multiple forms coexist.
func (f *FormBuilder) formAttr(n *Node) *Node {
	return n.Attr("form", f.id)
}

func (f *FormBuilder) renderField(fld *Field) *Node {
	switch fld.Type {
	case FieldHidden:
		return f.formAttr(IHidden().ID(f.fieldID(fld)).Attr("name", fld.Name).Attr("value", fld.Value))

	case FieldCheckbox:
		return f.renderCheckbox(fld)

	case FieldSelect:
		return f.renderSelect(fld)

	case FieldRadio:
		return f.renderRadioInline(fld)

	case FieldRadioBtn:
		return f.renderRadioButton(fld)

	case FieldRadioCard:
		return f.renderRadioCard(fld)

	case FieldTextarea:
		return f.renderTextarea(fld)

	default:
		return f.renderInput(fld)
	}
}

func (f *FormBuilder) renderInput(fld *Field) *Node {
	cls := f.inputClass(fld)
	var input *Node
	switch fld.Type {
	case FieldPassword:
		input = IPassword(cls)
	case FieldEmail:
		input = IEmail(cls)
	case FieldNumber:
		input = INumber(cls)
	case FieldPhone:
		input = IPhone(cls)
	case FieldDate:
		input = IDate(cls)
	case FieldTime:
		input = ITime(cls)
	case FieldDatetime:
		input = IDatetime(cls)
	case FieldUrl:
		input = IUrl(cls)
	case FieldSearch:
		input = ISearch(cls)
	default:
		input = IText(cls)
	}
	f.formAttr(input.ID(f.fieldID(fld)).Attr("name", fld.Name))
	if fld.Placeholder != "" {
		input.Attr("placeholder", fld.Placeholder)
	}
	if fld.Value != "" {
		input.Attr("value", fld.Value)
	}
	if fld.Pattern != "" {
		input.Attr("pattern", fld.Pattern)
	}

	wrapCls := fld.WrapClass
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-1"
	}

	errMsg := fld.ErrMsg
	if errMsg == "" && fld.Required {
		errMsg = fld.Label + " is required"
	}

	children := []*Node{
		Label("text-sm font-medium text-gray-700 dark:text-gray-300").Attr("for", f.fieldID(fld)).Text(fld.Label + f.reqSuffix(fld)),
		input,
	}
	if errMsg != "" {
		children = append(children, Div(f.errClass).ID(f.errID(fld)).Text(errMsg))
	}
	return Div(wrapCls).Render(children...)
}

func (f *FormBuilder) renderTextarea(fld *Field) *Node {
	cls := f.inputClass(fld)
	ta := Textarea(cls).ID(f.fieldID(fld)).Attr("name", fld.Name).Attr("rows", "3")
	f.formAttr(ta)
	if fld.Placeholder != "" {
		ta.Attr("placeholder", fld.Placeholder)
	}
	if fld.Value != "" {
		ta.Text(fld.Value)
	}

	wrapCls := fld.WrapClass
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-1"
	}

	errMsg := fld.ErrMsg
	if errMsg == "" && fld.Required {
		errMsg = fld.Label + " is required"
	}

	children := []*Node{
		Label("text-sm font-medium text-gray-700 dark:text-gray-300").Attr("for", f.fieldID(fld)).Text(fld.Label + f.reqSuffix(fld)),
		ta,
	}
	if errMsg != "" {
		children = append(children, Div(f.errClass).ID(f.errID(fld)).Text(errMsg))
	}
	return Div(wrapCls).Render(children...)
}

func (f *FormBuilder) renderCheckbox(fld *Field) *Node {
	cb := ICheckbox("w-4 h-4 cursor-pointer").ID(f.fieldID(fld)).Attr("name", fld.Name)
	f.formAttr(cb)
	if fld.Checked {
		cb.Attr("checked", "true")
	}
	return Label("flex items-center gap-2 text-sm cursor-pointer").Attr("for", f.fieldID(fld)).Render(
		cb,
		Span().Text(fld.Label),
	)
}

func (f *FormBuilder) renderSelect(fld *Field) *Node {
	cls := f.inputClass(fld)
	opts := make([]*Node, len(fld.Options))
	for i, o := range fld.Options {
		opt := Option().Attr("value", o.Value).Text(o.Label)
		if o.Value == fld.Value {
			opt.Attr("selected", "true")
		}
		opts[i] = opt
	}
	sel := Select(cls).ID(f.fieldID(fld)).Attr("name", fld.Name).Render(opts...)
	f.formAttr(sel)

	wrapCls := fld.WrapClass
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-1"
	}

	errMsg := fld.ErrMsg
	if errMsg == "" && fld.Required {
		errMsg = fld.Label + " is required"
	}

	children := []*Node{
		Label("text-sm font-medium text-gray-700 dark:text-gray-300").Attr("for", f.fieldID(fld)).Text(fld.Label + f.reqSuffix(fld)),
		sel,
	}
	if errMsg != "" {
		children = append(children, Div(f.errClass).ID(f.errID(fld)).Text(errMsg))
	}
	return Div(wrapCls).Render(children...)
}

// renderRadioInline creates simple inline radio buttons: <label><input radio> text</label>
// Radio names are scoped with the form ID to prevent cross-form collisions.
func (f *FormBuilder) renderRadioInline(fld *Field) *Node {
	rName := f.radioName(fld)
	radios := make([]*Node, len(fld.Options))
	for i, o := range fld.Options {
		radio := f.formAttr(IRadio("w-4 h-4").Attr("name", rName).Attr("value", o.Value))
		if i == 0 {
			radio.ID(f.fieldID(fld))
		}
		if fld.Value == o.Value {
			radio.Attr("checked", "true")
		}
		radios[i] = Label("flex items-center gap-2 text-sm cursor-pointer").Render(
			radio,
			Span().Text(o.Label),
		)
	}

	wrapCls := fld.WrapClass
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-1"
	}

	errMsg := fld.ErrMsg
	if errMsg == "" && fld.Required {
		errMsg = fld.Label + " is required"
	}

	children := []*Node{
		Label("text-sm font-medium text-gray-700 dark:text-gray-300").Text(fld.Label + f.reqSuffix(fld)),
		Div("flex gap-4").Render(radios...),
	}
	if errMsg != "" {
		children = append(children, Div(f.errClass).ID(f.errID(fld)).Text(errMsg))
	}
	return Div(wrapCls).Render(children...)
}

// renderRadioButton creates visible card-style radios with border + hover.
// Radio names are scoped with the form ID.
func (f *FormBuilder) renderRadioButton(fld *Field) *Node {
	rName := f.radioName(fld)
	cards := make([]*Node, len(fld.Options))
	for i, o := range fld.Options {
		radio := f.formAttr(IRadio("w-4 h-4").Attr("name", rName).Attr("value", o.Value))
		if i == 0 {
			radio.ID(f.fieldID(fld))
		}
		if fld.Value == o.Value {
			radio.Attr("checked", "true")
		}
		cards[i] = Label("flex items-center gap-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800").Render(
			radio,
			Span("text-sm").Text(o.Label),
		)
	}

	wrapCls := fld.WrapClass
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-1"
	}

	errMsg := fld.ErrMsg
	if errMsg == "" && fld.Required {
		errMsg = fld.Label + " is required"
	}

	children := []*Node{
		Label("text-sm font-medium text-gray-700 dark:text-gray-300").Text(fld.Label + f.reqSuffix(fld)),
		Div("flex gap-2").Render(cards...),
	}
	if errMsg != "" {
		children = append(children, Div(f.errClass).ID(f.errID(fld)).Text(errMsg))
	}
	return Div(wrapCls).Render(children...)
}

// renderRadioCard creates hidden-input card radios with peer-checked CSS.
// Radio names are scoped with the form ID.
func (f *FormBuilder) renderRadioCard(fld *Field) *Node {
	rName := f.radioName(fld)
	cards := make([]*Node, len(fld.Options))
	for i, o := range fld.Options {
		radio := f.formAttr(IRadio("peer hidden").Attr("name", rName).Attr("value", o.Value))
		if i == 0 {
			radio.ID(f.fieldID(fld))
		}
		if fld.Value == o.Value {
			radio.Attr("checked", "true")
		}
		cards[i] = Label("cursor-pointer block").Render(
			radio,
			Div("h-10 py-2 px-4 rounded-md border border-gray-300 dark:border-gray-600 text-center text-sm transition-all peer-checked:border-blue-500 peer-checked:bg-blue-50 dark:peer-checked:bg-blue-900/30 peer-checked:font-bold hover:border-gray-400 dark:hover:border-gray-500 dark:text-gray-200").Text(o.Label),
		)
	}

	wrapCls := fld.WrapClass
	if wrapCls == "" {
		wrapCls = "flex flex-col gap-1"
	}

	errMsg := fld.ErrMsg
	if errMsg == "" && fld.Required {
		errMsg = fld.Label + " is required"
	}

	children := []*Node{
		Label("text-sm font-medium text-gray-700 dark:text-gray-300").Text(fld.Label + f.reqSuffix(fld)),
		Div("flex gap-3").Render(cards...),
	}
	if errMsg != "" {
		children = append(children, Div(f.errClass).ID(f.errID(fld)).Text(errMsg))
	}
	return Div(wrapCls).Render(children...)
}

func (f *FormBuilder) reqSuffix(fld *Field) string {
	if fld.Required {
		return " *"
	}
	return ""
}

// ---------------------------------------------------------------------------
// Client-side validation JS generation
// ---------------------------------------------------------------------------

// buildValidateJS generates a self-executing JS function that:
// 1. Clears all error messages
// 2. Validates required fields
// 3. Validates pattern constraints
// 4. Collects all field values into an object
// 5. Calls __ws.call(actionName, data) only if valid
func (f *FormBuilder) buildValidateJS(actionValue string) string {
	var b strings.Builder
	b.WriteString("(function(){")
	b.WriteString("var ok=true;")

	// Helper functions
	b.WriteString("function err(id,show){var e=document.getElementById(id);if(e){e.classList.toggle('hidden',!show)}}")
	b.WriteString("function val(id){var e=document.getElementById(id);if(!e)return '';if(e.tagName.toLowerCase()==='textarea')return e.value.trim();return e.value.trim()}")
	// radioVal uses the scoped name (formID-fieldName) to query only radios
	// belonging to this form, preventing cross-form interference.
	b.WriteString("function radioVal(name){var c=document.querySelector('input[type=radio][name=\"'+name+'\"]:checked');return c?c.value:''}")
	b.WriteString("function selVal(id){var e=document.getElementById(id);return e?e.value:''}")
	b.WriteString("function checkVal(id){var e=document.getElementById(id);return e?e.checked:false}")

	// Phase 1: clear all errors
	for i := range f.fields {
		fld := &f.fields[i]
		if fld.Required || fld.Pattern != "" {
			fmt.Fprintf(&b, "err('%s',false);", escJS(f.errID(fld)))
		}
	}

	// Phase 2: validate required fields
	for i := range f.fields {
		fld := &f.fields[i]
		if !fld.Required && fld.Pattern == "" {
			continue
		}

		errID := f.errID(fld)
		fieldID := f.fieldID(fld)
		rName := f.radioName(fld) // scoped radio name

		switch fld.Type {
		case FieldRadio, FieldRadioBtn, FieldRadioCard:
			if fld.Required {
				fmt.Fprintf(&b, "if(!radioVal('%s')){err('%s',true);ok=false}", escJS(rName), escJS(errID))
			}
		case FieldSelect:
			if fld.Required {
				fmt.Fprintf(&b, "if(!selVal('%s')){err('%s',true);ok=false}", escJS(fieldID), escJS(errID))
			}
		case FieldCheckbox:
			if fld.Required {
				fmt.Fprintf(&b, "if(!checkVal('%s')){err('%s',true);ok=false}", escJS(fieldID), escJS(errID))
			}
		default:
			if fld.Required {
				fmt.Fprintf(&b, "if(!val('%s')){err('%s',true);ok=false}", escJS(fieldID), escJS(errID))
			}
			if fld.Pattern != "" {
				patternJSON, _ := json.Marshal(fld.Pattern)
				fmt.Fprintf(&b, "if(val('%s')&&!new RegExp(%s).test(val('%s'))){err('%s',true);ok=false}",
					escJS(fieldID), string(patternJSON), escJS(fieldID), escJS(errID))
			}
		}
	}

	b.WriteString("if(!ok)return;")

	// Phase 3: collect all values into data object.
	// Keys use the original field Name (not the scoped radio name) so the
	// server receives clean names like "Gender", not "myform-Gender".
	fmt.Fprintf(&b, "var d={Action:'%s'};", escJS(actionValue))

	for i := range f.fields {
		fld := &f.fields[i]
		fieldID := f.fieldID(fld)
		name := fld.Name
		rName := f.radioName(fld)

		switch fld.Type {
		case FieldRadio, FieldRadioBtn, FieldRadioCard:
			// Query the scoped radio name, store under the clean field name
			fmt.Fprintf(&b, "d['%s']=radioVal('%s');", escJS(name), escJS(rName))
		case FieldCheckbox:
			fmt.Fprintf(&b, "d['%s']=checkVal('%s');", escJS(name), escJS(fieldID))
		case FieldSelect:
			fmt.Fprintf(&b, "d['%s']=selVal('%s');", escJS(name), escJS(fieldID))
		default:
			fmt.Fprintf(&b, "d['%s']=val('%s');", escJS(name), escJS(fieldID))
		}
	}

	// Phase 4: call WS action
	fmt.Fprintf(&b, "__ws.call('%s',d);", escJS(f.actionName))
	b.WriteString("})()")

	return b.String()
}

// ---------------------------------------------------------------------------
// Server-side validation
// ---------------------------------------------------------------------------

// FormErrors holds field-level validation errors.
type FormErrors map[string]string

// HasErrors returns true if any field has an error.
func (fe FormErrors) HasErrors() bool { return len(fe) > 0 }

// Get returns the error for a specific field name, or empty string.
func (fe FormErrors) Get(name string) string { return fe[name] }

// Validate checks the data map against the form's field definitions.
// It returns a FormErrors map (empty if all valid).
// The data parameter should be the map[string]any from ctx.Body().
func (f *FormBuilder) Validate(data map[string]any) FormErrors {
	errs := make(FormErrors)

	for i := range f.fields {
		fld := &f.fields[i]

		if fld.Required {
			switch fld.Type {
			case FieldCheckbox:
				v, _ := data[fld.Name].(bool)
				if !v {
					errs[fld.Name] = f.errMsg(fld)
				}
			default:
				v := fmt.Sprintf("%v", data[fld.Name])
				if v == "" || v == "<nil>" {
					errs[fld.Name] = f.errMsg(fld)
				}
			}
		}

		// Pattern validation (server-side) for string fields
		if fld.Pattern != "" {
			v := fmt.Sprintf("%v", data[fld.Name])
			if v != "" && v != "<nil>" {
				// Use Go regexp for server-side pattern check
				// We import regexp lazily to avoid import cost when not used
				matched := matchPattern(fld.Pattern, v)
				if !matched {
					msg := fld.ErrMsg
					if msg == "" {
						msg = fld.Label + " format is invalid"
					}
					errs[fld.Name] = msg
				}
			}
		}
	}

	return errs
}

func (f *FormBuilder) errMsg(fld *Field) string {
	if fld.ErrMsg != "" {
		return fld.ErrMsg
	}
	return fld.Label + " is required"
}

// matchPattern matches a regex pattern against a string value.
func matchPattern(pattern, value string) bool {
	matched, err := regexp.MatchString("^(?:"+pattern+")$", value)
	if err != nil {
		return false
	}
	return matched
}
