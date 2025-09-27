package pages

import (
	"github.com/go-playground/validator/v10"
	"github.com/michalCapo/g-sui/ui"
)

func LoginContent(ctx *ui.Context) string {
	return LoginForm("user").render(ctx, nil)
}

func LoginForm(name string) *TLoginForm { return &TLoginForm{Name: name} }

// defining login form with validations for given fields
type TLoginForm struct {
	Name     string `validate:"required,oneof=user"`
	Password string `validate:"required,oneof=password"`
}

// we want to display success message
func (form *TLoginForm) Success(ctx *ui.Context) string {
	return ui.Div("text-green-600 max-w-md p-8 text-center font-bold rounded-lg bg-white shadow-xl")("Success")
}

// Login action
func (form *TLoginForm) Login(ctx *ui.Context) string {
	if err := ctx.Body(form); err != nil {
		return form.render(ctx, &err)
	}

	v := validator.New()
	if err := v.Struct(form); err != nil {
		return form.render(ctx, &err)
	}

	return form.Success(ctx)
}

// translations for login form
var translations = map[string]string{
	"Name":              "User name",
	"has invalid value": "is invalid",
}

// temporary id
var loginTarget = ui.Target()

func (form *TLoginForm) render(ctx *ui.Context, err *error) string {
	return ui.Form("max-w-md bg-white p-8 rounded-lg shadow-xl", loginTarget, ctx.Submit(form.Login).Replace(loginTarget))(
		ui.ErrorForm(err, &translations),
		ui.IText("Name", form).Required().Error(err).Render("Name"),
		ui.IPassword("Password").Required().Error(err).Render("Password"),
		ui.Button().Submit().Color(ui.Blue).Class("rounded").Render("Login"),
	)
}
