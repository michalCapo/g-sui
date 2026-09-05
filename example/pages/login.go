package pages

import r "github.com/michalCapo/g-sui/ui"

const (
	loginFormID = "login-form"
	loginNameID = "login-name"
	loginPassID = "login-pass"
)

type loginInput struct {
	Name     string `json:"Name"`
	Password string `json:"Password"`
}

func (input *loginInput) Validate() error {
	errs := r.FormErrors{}
	if input.Name == "" {
		errs["Name"] = "User name is required"
	}
	if input.Password == "" {
		errs["Password"] = "Password is required"
	}
	if errs.HasErrors() {
		return r.ValidationError{Fields: errs}
	}
	return nil
}

func loginForm(submit r.ActionRef[loginInput]) *r.Node {
	inputCls := "w-full border border-gray-300 rounded px-3 py-2 text-sm"
	return r.FormFor[loginInput](loginFormID, "max-w-md bg-white p-8 rounded-lg shadow-xl flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Login"),
		r.Div("text-sm text-gray-500").Text("Enter user / password to test validation."),
		r.Label("text-sm font-medium text-gray-700").Attr("for", loginNameID).Text("User name"),
		r.IText(inputCls).ID(loginNameID).Attr("name", "Name").Attr("required", "").Attr("autocomplete", "username"),
		r.Label("text-sm font-medium text-gray-700").Attr("for", loginPassID).Text("Password"),
		r.IPassword(inputCls).ID(loginPassID).Attr("name", "Password").Attr("required", "").Attr("autocomplete", "current-password"),
		r.Button("px-4 py-2 bg-blue-600 text-white rounded text-sm font-medium hover:bg-blue-700 cursor-pointer").Attr("type", "submit").Text("Login"),
		r.Div("text-xs text-gray-500 text-center").Text("Hint: use user / password"),
	).Submit(submit)
}

func RegisterLogin(app *r.App) {
	submit := r.RegisterAction(app, "login.submit", func(ctx *r.Context, input loginInput) (r.Result, error) {
		if input.Name != "user" || input.Password != "password" {
			return r.Result{}, r.ValidationError{Fields: r.FormErrors{"Name": "Use user / password for this demo"}}
		}
		return r.Result{}.Replace(loginFormID, r.Div("text-green-600 max-w-md p-8 text-center font-bold rounded-lg bg-white shadow-xl").ID(loginFormID).Text("Success")).Toast("Login successful"), nil
	})
	app.Page("/login", func(ctx *r.Context) *r.Node { return loginForm(submit) })
}
