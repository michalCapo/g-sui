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

		// Step Progress Section
		renderStepProgress(),

		// Buttons with Tooltips
		renderTooltips(),

		// Tabs Section
		renderTabs(),

		// Accordion Section
		renderAccordion(),

		// Dropdown Section
		renderDropdowns(),

		// Form inputs
		form.Render(ctx, nil),
	)
}

func renderAlerts() string {
	return ui.Div("flex flex-col gap-4")(
		ui.Div("text-2xl font-bold")("Alerts"),
		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-4")(
			ui.Div("flex flex-col gap-2")(
				ui.Div("text-sm font-bold text-gray-500 uppercase mb-1")("With Titles"),
				ui.Alert().Variant("info").Title("Heads up!").Message("This is an info alert with important information.").Dismissible(true).Render(),
				ui.Alert().Variant("success").Title("Great success!").Message("Your changes have been saved successfully.").Dismissible(true).Render(),
			),
			ui.Div("flex flex-col gap-2")(
				ui.Div("text-sm font-bold text-gray-500 uppercase mb-1")("Outline Variants"),
				ui.Alert().Variant("warning-outline").Title("Warning").Message("Please review your input before proceeding.").Dismissible(true).Render(),
				ui.Alert().Variant("error-outline").Title("Error occurred").Message("Something went wrong while saving your data.").Dismissible(true).Render(),
			),
		),
	)
}

func renderBadges() string {
	icon := ui.Icon("check_circle", ui.Attr{Class: "text-xs"})

	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Badges"),
		ui.Div("flex flex-col gap-6")(
			ui.Div("flex flex-wrap items-center gap-4")(
				ui.Div("text-sm font-bold text-gray-500 uppercase w-full mb-1")("Variants & Icons"),
				ui.Badge().Color("green-soft").Text("Verified").Icon(icon).Render(),
				ui.Badge().Color("blue").Text("New").Size("lg").Render(),
				ui.Badge().Color("red").Text("Urgent").Square().Render(),
				ui.Badge().Color("yellow-soft").Text("Warning").Size("sm").Render(),
			),
			ui.Div("flex flex-wrap items-center gap-4")(
				ui.Div("text-sm font-bold text-gray-500 uppercase w-full mb-1")("Dots & Sizes"),
				ui.Badge().Color("green").Dot().Size("sm").Render(),
				ui.Badge().Color("blue").Dot().Size("md").Render(),
				ui.Badge().Color("red").Dot().Size("lg").Render(),
				ui.Badge().Color("purple-soft").Text("Large Badge").Size("lg").Render(),
			),
		),
	)
}

func renderCards() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Cards"),
		ui.Div("grid grid-cols-1 md:grid-cols-3 gap-6")(
			ui.Card().Header("<h3 class='font-bold'>Standard Card</h3>").
				Body("<p class='text-gray-600 dark:text-gray-400'>This is a standard shadowed card with default padding.</p>").
				Footer("<div class='text-xs text-gray-500'>Card Footer</div>").
				Render(),
			ui.Card().Image("https://images.unsplash.com/photo-1506744038136-46273834b3fb?w=800&auto=format&fit=crop", "Landscape").
				Header("<h3 class='font-bold'>Card with Image</h3>").
				Body("<p class='text-gray-600 dark:text-gray-400'>Cards can now display images at the top.</p>").
				Hover(true).
				Render(),
			ui.Card().Variant(ui.CardGlass).
				Header("<h3 class='font-bold'>Glass Variant</h3>").
				Body("<p class='text-gray-600 dark:text-gray-400'>This card uses a glassmorphism effect with backdrop blur.</p>").
				Hover(true).
				Render(),
		),
	)
}

func renderProgress() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Progress Bars"),
		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-8")(
			ui.Div("flex flex-col gap-4")(
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Gradient Style (75%)"),
					ui.ProgressBar().Value(75).Gradient("#3b82f6", "#8b5cf6").Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Outside Label"),
					ui.ProgressBar().Value(45).Label("System Update").LabelPosition("outside").Color("bg-indigo-600").Render(),
				),
			),
			ui.Div("flex flex-col gap-4")(
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Animated Stripes"),
					ui.ProgressBar().Value(65).Color("bg-green-500").Striped(true).Animated(true).Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Indeterminate"),
					ui.ProgressBar().Indeterminate(true).Color("bg-blue-600").Render(),
				),
			),
		),
	)
}

func renderStepProgress() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Step Progress"),
		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-8")(
			ui.Div("flex flex-col gap-4")(
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Step 1 of 4"),
					ui.StepProgress(1, 4).Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Step 2 of 4"),
					ui.StepProgress(2, 4).Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Step 3 of 4"),
					ui.StepProgress(3, 4).Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Step 4 of 4 (Complete)"),
					ui.StepProgress(4, 4).Color("bg-green-500").Render(),
				),
			),
			ui.Div("flex flex-col gap-4")(
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Small Size - Step 1 of 5"),
					ui.StepProgress(1, 5).Size("sm").Color("bg-purple-500").Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Large Size - Step 2 of 5"),
					ui.StepProgress(2, 5).Size("lg").Color("bg-yellow-500").Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Extra Large - Step 3 of 5"),
					ui.StepProgress(3, 5).Size("xl").Color("bg-red-500").Render(),
				),
				ui.Div("")(
					ui.Div("mb-1 text-sm font-medium")("Custom Step Progress"),
					ui.StepProgress(7, 10).Color("bg-indigo-500").Size("md").Render(),
				),
			),
		),
	)
}

func renderTooltips() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Tooltips"),
		ui.Div("flex flex-wrap gap-4")(
			ui.Tooltip().Content("Delayed tooltip").Delay(500).Render(
				ui.Button().Color(ui.Blue).Class("rounded-lg").Render("500ms Delay"),
			),
			ui.Tooltip().Content("Bottom position").Position("bottom").Render(
				ui.Button().Color(ui.Green).Class("rounded-lg").Render("Bottom"),
			),
			ui.Tooltip().Content("Success variant").Variant("green").Render(
				ui.Button().Color(ui.GreenOutline).Class("rounded-lg").Render("Success"),
			),
			ui.Tooltip().Content("Danger variant").Variant("red").Render(
				ui.Button().Color(ui.RedOutline).Class("rounded-lg").Render("Danger"),
			),
			ui.Tooltip().Content("Light variant").Variant("light").Render(
				ui.Button().Color(ui.GrayOutline).Class("rounded-lg").Render("Light"),
			),
		),
	)
}

func renderTabs() string {
	iconHome := ui.Icon("home", ui.Attr{Class: "text-sm"})
	iconUser := ui.Icon("person", ui.Attr{Class: "text-sm"})
	iconSettings := ui.Icon("settings", ui.Attr{Class: "text-sm"})

	contentClass := "p-6 bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-100 dark:border-gray-800"

	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Tabs"),
		ui.Div("grid grid-cols-1 gap-8")(
			ui.Div("")(
				ui.Div("text-sm font-bold text-gray-500 uppercase mb-3")("Boxed Style with Icons"),
				ui.Tabs().
					Tab("Home", ui.Div(contentClass)(
						ui.Div("text-lg font-bold mb-2")("🏠 Dashboard Home"),
						ui.Div("text-gray-600 dark:text-gray-400")("Welcome to your central dashboard. This panel demonstrates how tabs can wrap complex HTML content with a clean white background."),
					), iconHome).
					Tab("Profile", ui.Div(contentClass)(
						ui.Div("text-lg font-bold mb-2")("👤 User Profile"),
						ui.Div("text-gray-600 dark:text-gray-400")("Manage your personal information, display name, and avatar settings here."),
					), iconUser).
					Tab("Settings", ui.Div(contentClass)(
						ui.Div("text-lg font-bold mb-2")("⚙️ System Settings"),
						ui.Div("text-gray-600 dark:text-gray-400")("Fine-tune application behavior, notification preferences, and privacy controls."),
					), iconSettings).
					Active(0).
					Style("boxed").
					Render(),
			),
			ui.Div("grid grid-cols-1 md:grid-cols-2 gap-8")(
				ui.Div("")(
					ui.Div("text-sm font-bold text-gray-500 uppercase mb-3")("Underline Style"),
					ui.Tabs().
						Tab("General", ui.Div(contentClass)(
							ui.Div("font-bold")("General Info"),
							ui.P("mt-2")("Basic configuration for your workspace."),
						)).
						Tab("Security", ui.Div(contentClass)(
							ui.Div("font-bold")("Privacy & Security"),
							ui.P("mt-2")("Manage passwords, two-factor authentication and active sessions."),
						)).
						Active(0).
						Style("underline").
						Render(),
				),
				ui.Div("")(
					ui.Div("text-sm font-bold text-gray-500 uppercase mb-3")("Pills Style"),
					ui.Tabs().
						Tab("Daily", ui.Div(contentClass)(
							ui.Div("font-bold text-blue-600")("Today's Progress"),
							ui.P("mt-1")("Detailed activities for the last 24 hours."),
						)).
						Tab("Weekly", ui.Div(contentClass)(
							ui.Div("font-bold text-green-600")("Weekly Trends"),
							ui.P("mt-1")("Summary of performance over the past 7 days."),
						)).
						Tab("Monthly", ui.Div(contentClass)(
							ui.Div("font-bold text-purple-600")("Monthly Report"),
							ui.P("mt-1")("Strategic overview of goals achieved this month."),
						)).
						Active(1).
						Style("pills").
						Render(),
				),
			),
		),
	)
}

func renderAccordion() string {
	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Accordion"),
		ui.Div("grid grid-cols-1 md:grid-cols-2 gap-8")(
			ui.Div("")(
				ui.Div("text-sm font-bold text-gray-500 uppercase mb-3")("Bordered with Default Open"),
				ui.Accordion().Variant(ui.AccordionBordered).
					Item("What is g-sui?", "g-sui is a Go-based server-rendered UI framework that provides a component-based approach to building web applications.", true).
					Item("How do I get started?", "Simply import the ui package and start composing components.").
					Item("Is it responsive?", "All components are built with responsive design in mind, using Tailwind's responsive modifiers.").
					Render(),
			),
			ui.Div("")(
				ui.Div("text-sm font-bold text-gray-500 uppercase mb-3")("Separated Variant (Multiple)"),
				ui.Accordion().Variant(ui.AccordionSeparated).Multiple(true).
					Item("Separated Section 1", "In the separated variant, each item is its own card.").
					Item("Separated Section 2", "Multiple sections can be open at once when Multiple(true) is used.").
					Render(),
			),
		),
	)
}

func renderDropdowns() string {
	iconEdit := ui.Icon("edit", ui.Attr{Class: "text-sm"})
	iconDelete := ui.Icon("delete", ui.Attr{Class: "text-sm"})

	return ui.Div("")(
		ui.Div("text-2xl font-bold mb-4")("Dropdown Menus"),
		ui.Div("flex flex-wrap gap-4")(
			ui.Dropdown().
				Trigger(ui.Button().Color(ui.Blue).Class("rounded-lg").Render("Actions ▼")).
				Header("General").
				Item("Edit Profile", "alert('Edit')", iconEdit).
				Item("Account Settings", "alert('Settings')").
				Divider().
				Header("Danger Zone").
				Danger("Delete Account", "alert('Delete')", iconDelete).
				Position("bottom-left").
				Render(),

			ui.Dropdown().
				Trigger(ui.Button().Color(ui.GrayOutline).Class("rounded-lg").Render("Options ▼")).
				Item("Share", "alert('Share')").
				Item("Download", "alert('Download')").
				Position("bottom-right").
				Render(),
		),
	)
}
