package pages

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/michalCapo/g-sui/ui"
)

// Demo form showcasing inputs and validation
type DemoForm struct {
	Name      string `validate:"required"`
	Email     string `validate:"required,email"`
	Phone     string
	Password  string  `validate:"required,min=6"`
	Age       int     `validate:"gte=0,lte=120"`
	Price     float64 `validate:"gte=0"`
	Bio       string
	Gender    string `validate:"oneof=male female other"`
	Country   string
	Agree     bool `validate:"eq=true"`
	BirthDate time.Time
	AlarmTime time.Time
	Meeting   time.Time
}

var (
	demoTarget = ui.Target()
)

func (f *DemoForm) Submit(ctx *ui.Context) string {
	if err := ctx.Body(f); err != nil {
		return f.Render(ctx, &err)
	}
	v := validator.New()
	if err := v.Struct(f); err != nil {
		return f.Render(ctx, &err)
	}
	ctx.Success("Form submitted successfully")
	return f.Render(ctx, nil)
}

func (f *DemoForm) Render(ctx *ui.Context, err *error) string {
	countries := ui.MakeOptions([]string{"", "USA", "Slovakia", "Germany", "Japan"})
	genders := []ui.AOption{{ID: "male", Value: "Male"}, {ID: "female", Value: "Female"}, {ID: "other", Value: "Other"}}

	// Full-width section (no grid), single-column form card
	return ui.Div("w-full", demoTarget)(
		ui.Form("flex flex-col gap-4 bg-white p-6 rounded-lg shadow w-full", demoTarget, ctx.Submit(f.Submit).Replace(demoTarget))(
			ui.ErrorForm(err, nil),

			ui.IText("Name", f).Required().Render("Name"),
			ui.IEmail("Email", f).Required().Render("Email"),
			ui.IPhone("Phone", f).Render("Phone"),
			ui.IPassword("Password").Required().Render("Password"),

			ui.INumber("Age", f).Numbers(0, 120, 1).Render("Age"),
			ui.INumber("Price", f).Format("%.2f").Render("Price (USD)"),
			ui.IArea("Bio", f).Rows(4).Render("Short Bio"),

			ui.Div("block sm:hidden")(
				ui.Div("text-sm font-bold")("Gender"),
				ui.IRadio("Gender", f).Value("male").Render("Male"),
				ui.IRadio("Gender", f).Value("female").Render("Female"),
				ui.IRadio("Gender", f).Value("other").Render("Other"),
			),
			ui.Div("hidden sm:block overflow-x-auto")(
				ui.IRadioButtons("Gender", f).Options(genders).Render("Gender"),
			),
			ui.ISelect("Country", f).Options(countries).Render("Country"),
			ui.ICheckbox("Agree", f).Required().Render("I agree to the terms"),

			ui.IDate("BirthDate", f).Render("Birth Date"),
			ui.ITime("AlarmTime", f).Render("Alarm Time"),
			ui.IDateTime("Meeting", f).Render("Meeting (Local)"),

			ui.Div("flex gap-2 mt-2")(
				ui.Button().Submit().Color(ui.Blue).Class("rounded").Render("Submit"),
				ui.Button().Reset().Color(ui.Gray).Class("rounded").Render("Reset"),
			),
		),
	)
}

func Showcase(ctx *ui.Context) string {
	form := &DemoForm{}
	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-6 w-full")(
		ui.Div("text-3xl font-bold")("Component Showcase"),
		ui.Div("text-gray-600")("Inputs, validation, and common field widgets."),
		form.Render(ctx, nil),
	)
}
