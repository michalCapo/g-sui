package ui

import "fmt"

type FormInstance struct {
	FormId   string
	OnSubmit Attr
}

func FormNew(onSubmit Attr) *FormInstance {
	return &FormInstance{
		FormId:   "i" + RandomString(15),
		OnSubmit: onSubmit,
	}
}

func (f *FormInstance) Text(name string, data ...any) *TInput {
	return IText(name, data...).Form(f.FormId)
}

func (f *FormInstance) Area(name string, data ...any) *TInput {
	return IArea(name, data...).Form(f.FormId)
}

func (f *FormInstance) Password(name string, data ...any) *TInput {
	return IPassword(name, data...).Form(f.FormId)
}

func (f *FormInstance) Number(name string, data ...any) *TInput {
	return INumber(name, data...).Form(f.FormId)
}

func (f *FormInstance) Select(name string, data ...any) *ASelect {
	return ISelect(name, data...).Form(f.FormId)
}

func (f *FormInstance) Checkbox(name string, data ...any) *TInput {
	return ICheckbox(name, data...).Form(f.FormId)
}

func (f *FormInstance) Radio(name string, data ...any) *TInput {
	return IRadio(name, data...).Form(f.FormId)
}

func (f *FormInstance) RadioButtons(name string, data ...any) *ARadio {
	return IRadioButtons(name, data...).Form(f.FormId)
}

func (f *FormInstance) Date(name string, data ...any) *TInput {
	return IDate(name, data...).Form(f.FormId)
}

func (f *FormInstance) Time(name string, data ...any) *TInput {
	return ITime(name, data...).Form(f.FormId)
}

func (f *FormInstance) DateTime(name string, data ...any) *TInput {
	return IDateTime(name, data...).Form(f.FormId)
}

func (f *FormInstance) Phone(name string, data ...any) *TInput {
	return IPhone(name, data...).Form(f.FormId)
}

func (f *FormInstance) Email(name string, data ...any) *TInput {
	return IEmail(name, data...).Form(f.FormId)
}

func (f *FormInstance) Hidden(name string, typ string, value any, attr ...Attr) string {
	attr = append(attr, Attr{Name: name, Type: typ, Value: fmt.Sprintf("%v", value), Form: f.FormId})
	return Input("hidden", attr...)
}

func (f *FormInstance) Button() *button {
	return Button().Form(f.FormId)
}

func (f *FormInstance) Render() string {
	return Form("hidden", Attr{ID: f.FormId}, f.OnSubmit)()
}
