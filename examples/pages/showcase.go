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
	return ui.Div("max-w-full sm:max-w-6xl mx-auto flex flex-col gap-8 w-full")(
		ui.Div("text-3xl font-bold")("Component Showcase"),
		ui.Div("text-gray-600")("A collection of reusable UI components."),

		// Alerts Section
		renderAlerts(),

		// Badges Section
		renderBadges(),

		// Cards Section
		renderCards(),

		// Progress Bars Section
		renderProgress(),

		// Buttons with Tooltips
		renderTooltips(),

		// Tabs Section
		renderTabs(),

		// Accordion Section
		renderAccordion(),

		// Form inputs
		form.Render(ctx, nil),

		// Dropdown Section
		renderDropdowns(),
	)
}

func renderAlerts() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Alerts"),
		ui.Alert().Variant("info").Message("This is an info alert with important information.").Dismissible(true).Render(),
		ui.Alert().Variant("success").Message("Success! Your changes have been saved.").Dismissible(true).Render(),
		ui.Alert().Variant("warning").Message("Warning: Please review before proceeding.").Dismissible(true).Render(),
		ui.Alert().Variant("error").Message("Error: Something went wrong. Please try again.").Dismissible(true).Render(),
	)
}

func renderBadges() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Badges"),
		ui.Div("flex flex-wrap items-center gap-4")(
			ui.Div("flex items-center gap-2")(
				ui.Span("text-gray-700")("Notification"),
				ui.Badge().Color("red").Dot().Render(),
			),
			ui.Div("flex items-center gap-2")(
				ui.Span("text-gray-700")("Messages"),
				ui.Badge().Color("blue").Text("3").Render(),
			),
			ui.Div("flex items-center gap-2")(
				ui.Span("text-gray-700")("Status"),
				ui.Badge().Color("green").Text("Online").Render(),
			),
			ui.Div("flex items-center gap-2")(
				ui.Span("text-gray-700")("Beta"),
				ui.Badge().Color("yellow").Text("Beta").Render(),
			),
			ui.Div("flex items-center gap-2")(
				ui.Span("text-gray-700")("New"),
				ui.Badge().Color("purple").Text("New").Render(),
			),
		),
	)
}

func renderCards() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Cards"),
		ui.Div("grid grid-cols-1 md:grid-cols-3 gap-4")(
			ui.Card().Header("<h3 class='font-bold'>Bordered Card</h3>").
				Body("<p class='text-gray-600'>This card has a border but no shadow.</p>").
				Variant(ui.CardBordered).Render(),
			ui.Card().Header("<h3 class='font-bold'>Shadowed Card</h3>").
				Body("<p class='text-gray-600'>This card has a shadow with subtle border.</p>").
				Variant(ui.CardShadowed).Render(),
			ui.Card().Header("<h3 class='font-bold'>Flat Card</h3>").
				Body("<p class='text-gray-600'>This card has no border or shadow.</p>").
				Variant(ui.CardFlat).Render(),
		),
	)
}

func renderProgress() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Progress Bars"),
		ui.Div("flex flex-col gap-4")(
			ui.Div("")(
				ui.Div("mb-1 font-medium")("Progress: 25%"),
				ui.ProgressBar().Value(25).Render(),
			),
			ui.Div("")(
				ui.Div("mb-1 font-medium")("Progress: 50% (Striped)"),
				ui.ProgressBar().Value(50).Striped(true).Render(),
			),
			ui.Div("")(
				ui.Div("mb-1 font-medium")("Progress: 75% (Animated)"),
				ui.ProgressBar().Value(75).Striped(true).Animated(true).Render(),
			),
			ui.Div("")(
				ui.Div("mb-1 font-medium")("Complete: 100%"),
				ui.ProgressBar().Value(100).Color("bg-green-600").Striped(true).Animated(true).Render(),
			),
		),
	)
}

func renderTooltips() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Tooltips"),
		ui.Div("flex flex-wrap gap-4")(
			ui.Tooltip().Content("This tooltip appears on top").Position("top").Render(
				ui.Button().Color(ui.Blue).Class("rounded").Render("Hover me (Top)"),
			),
			ui.Tooltip().Content("This tooltip appears below").Position("bottom").Render(
				ui.Button().Color(ui.Green).Class("rounded").Render("Hover me (Bottom)"),
			),
			ui.Tooltip().Content("This tooltip appears on the left").Position("left").Render(
				ui.Button().Color(ui.Yellow).Class("rounded").Render("Hover me (Left)"),
			),
			ui.Tooltip().Content("This tooltip appears on the right").Position("right").Render(
				ui.Button().Color(ui.Red).Class("rounded").Render("Hover me (Right)"),
			),
			ui.Tooltip().Content("Light variant tooltip").Variant("light").Render(
				ui.Button().Color(ui.Purple).Class("rounded").Render("Light Tooltip"),
			),
		),
	)
}

func renderTabs() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Tabs"),
		ui.Tabs().
			Tab("Overview", "<div class='p-4'><p>This is the overview tab content. You can put any HTML content here.</p></div>").
			Tab("Features", "<div class='p-4'><ul class='list-disc list-inside'><li>Feature 1</li><li>Feature 2</li><li>Feature 3</li></ul></div>").
			Tab("Settings", "<div class='p-4'><p>Configure your settings here.</p></div>").
			Active(0).
			Style("underline").
			Render(),

		ui.Div("text-lg font-bold mt-6 mb-2")("Pills Style"),
		ui.Tabs().
			Tab("Home", "<div class='p-4'><p>Welcome home!</p></div>").
			Tab("Profile", "<div class='p-4'><p>User profile information</p></div>").
			Tab("Messages", "<div class='p-4'><p>Your messages</p></div>").
			Active(0).
			Style("pills").
			Render(),
	)
}

func renderAccordion() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Accordion"),
		ui.Accordion().
			Item("What is g-sui?", "g-sui is a Go-based server-rendered UI framework that provides a component-based approach to building web applications with minimal JavaScript.").
			Item("How do I get started?", "Simply import the ui package and start composing components. The framework handles HTML generation and provides type-safe components.").
			Item("Can I use custom CSS?", "Yes! g-sui uses Tailwind CSS by default, but you can add custom classes to any component using the Class() method.").
			Item("Is it responsive?", "All components are built with responsive design in mind, using Tailwind's responsive modifiers.").
			Render(),

		ui.Div("text-lg font-bold mt-6 mb-2")("Multiple Open Sections"),
		ui.Accordion().
			Item("Section 1", "Content for section 1. You can have multiple sections open at once with Multiple(true).").
			Item("Section 2", "Content for section 2.").
			Item("Section 3", "Content for section 3.").
			Multiple(true).
			Render(),
	)
}

func renderDropdowns() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Dropdown Menus"),
		ui.Div("flex flex-wrap gap-4")(
			ui.Dropdown().
				Trigger(ui.Button().Color(ui.Blue).Class("rounded").Render("Options ▼")).
				Item("Edit", "alert('Edit')").
				Item("Duplicate", "alert('Duplicate')").
				Divider().
				Item("Delete", "alert('Delete')").
				Position("bottom-left").
				Render(),

			ui.Dropdown().
				Trigger(ui.Button().Color(ui.Green).Class("rounded").Render("Actions ▼")).
				Item("View", "alert('View')").
				Item("Download", "alert('Download')").
				Item("Share", "alert('Share')").
				Position("bottom-right").
				Render(),
		),
	)
}
