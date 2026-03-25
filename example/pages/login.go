package pages

import (
	"fmt"

	r "github.com/michalCapo/g-sui/ui"
)

func LoginPage(ctx *r.Context) *r.Node {
	return loginForm("", nil)
}

func loginForm(errMsg string, data map[string]any) *r.Node {
	inputCls := "w-full border border-gray-300 rounded px-3 py-2 text-sm"
	errCls := "border-red-400"

	nameVal := ""
	if data != nil {
		if v, ok := data["Name"].(string); ok {
			nameVal = v
		}
	}

	nameInput := r.IText(inputCls).ID("login-name").Attr("name", "Name").Attr("placeholder", "Username")
	if nameVal != "" {
		nameInput.Attr("value", nameVal)
	}
	if errMsg != "" {
		nameInput.Class(errCls)
	}

	passInput := r.IPassword(inputCls).ID("login-pass").Attr("name", "Password").Attr("placeholder", "Password")
	if errMsg != "" {
		passInput.Class(errCls)
	}

	return r.Div("max-w-md bg-white p-8 rounded-lg shadow-xl flex flex-col gap-4").ID("login-form").Render(
		r.Div("text-2xl font-bold").Text("Login"),
		r.Div("text-sm text-gray-500").Text("Enter user / password to test validation."),

		r.If(errMsg != "",
			r.Div("bg-red-50 border border-red-200 rounded px-4 py-3 text-sm text-red-700").Text(errMsg),
		),

		r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm font-medium text-gray-700").Text("User name"),
			nameInput,
		),
		r.Div("flex flex-col gap-1").Render(
			r.Label("text-sm font-medium text-gray-700").Text("Password"),
			passInput,
		),

		r.Button("px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 cursor-pointer").
			Text("Login").
			OnClick(&r.Action{
				Name:    "login.submit",
				Collect: []string{"login-name", "login-pass"},
			}),

		r.Div("text-xs text-gray-500 text-center").Text(
			fmt.Sprintf("Hint: use %q / %q", "user", "password"),
		),
	)
}

func loginSuccess() *r.Node {
	return r.Div("text-green-600 max-w-md p-8 text-center font-bold rounded-lg bg-white shadow-xl").
		ID("login-form").
		Text("Success")
}

func HandleLoginSubmit(ctx *r.Context) string {
	var data map[string]any
	ctx.Body(&data)

	name, _ := data["Name"].(string)
	pass, _ := data["Password"].(string)

	if name == "" || pass == "" {
		return r.NewResponse().
			Replace("login-form", loginForm("User name and password are required", data)).
			Build()
	}

	if name != "user" || pass != "password" {
		return r.NewResponse().
			Replace("login-form", loginForm("Invalid credentials. Name must be 'user' and password 'password'.", data)).
			Build()
	}

	return r.NewResponse().
		Replace("login-form", loginSuccess()).
		Toast("success", "Login successful").
		Build()
}

func RegisterLogin(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/login", func(ctx *r.Context) *r.Node { return layout(ctx, LoginPage(ctx)) })
	app.Action("nav.login", NavTo("/login", func() *r.Node { return LoginPage(nil) }))
	app.Action("login.submit", HandleLoginSubmit)
}
