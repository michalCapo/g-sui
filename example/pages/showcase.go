package pages

import (
	r "github.com/michalCapo/g-sui/ui"
)

func Showcase(ctx *r.Context) *r.Node {
	return r.Div("max-w-6xl mx-auto flex flex-col gap-8 w-full").Render(
		r.Div("text-3xl font-bold").Text("Component Showcase"),
		r.Div("text-gray-600").Text("A collection of reusable UI components built with the rework framework."),

		renderAlertSection(),
		renderBadgeSection(),
		renderCardSection(),
		renderProgressSection(),
		renderStepProgressSection(),
		renderTooltipSection(),
		renderTabsSection(),
		renderAccordionSection(),
		renderDropdownSection(),
		renderConfirmDialogSection(),
	)
}

// ---------------------------------------------------------------------------
// Alerts
// ---------------------------------------------------------------------------

func renderAlertSection() *r.Node {
	return r.Div("flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Alerts"),
		r.Div("grid grid-cols-1 md:grid-cols-2 gap-4").Render(
			r.Div("flex flex-col gap-2").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase mb-1").Text("With Titles (Dismissible)"),
				r.NewAlert().
					Title("Heads up!").
					Message("This is an info alert with important information.").
					Variant("info").
					Dismissible(true).
					Build(),
				r.NewAlert().
					Title("Great success!").
					Message("Your changes have been saved successfully.").
					Variant("success").
					Dismissible(true).
					Build(),
			),
			r.Div("flex flex-col gap-2").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase mb-1").Text("Outline Variants"),
				r.NewAlert().
					Title("Warning").
					Message("Please review your input before proceeding.").
					Variant("warning-outline").
					Build(),
				r.NewAlert().
					Title("Error occurred").
					Message("Something went wrong while saving your data.").
					Variant("error-outline").
					Build(),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Badges
// ---------------------------------------------------------------------------

func renderBadgeSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Badges"),
		r.Div("flex flex-col gap-6").Render(
			r.Div("flex flex-wrap items-center gap-4").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase w-full mb-1").Text("Variants & Icons"),
				r.NewBadge("Verified").Color("green").BadgeIcon("check_circle").Build(),
				r.NewBadge("New").Color("blue").Build(),
				r.NewBadge("Urgent").Color("red").Square().Build(),
				r.NewBadge("Warning").Color("yellow").Build(),
			),
			r.Div("flex flex-wrap items-center gap-4").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase w-full mb-1").Text("Soft Variants"),
				r.NewBadge("Soft Green").Color("green-soft").Build(),
				r.NewBadge("Soft Blue").Color("blue-soft").Build(),
				r.NewBadge("Soft Purple").Color("purple-soft").Build(),
				r.NewBadge("Soft Yellow").Color("yellow-soft").Build(),
			),
			r.Div("flex flex-wrap items-center gap-4").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase w-full mb-1").Text("Sizes"),
				r.NewBadge("Small").Color("gray").BadgeSize("sm").Build(),
				r.NewBadge("Default").Color("gray").BadgeSize("md").Build(),
				r.NewBadge("Large").Color("purple").BadgeSize("lg").Build(),
			),
			r.Div("flex flex-wrap items-center gap-4").Render(
				r.Div("text-sm font-bold text-gray-500 uppercase w-full mb-1").Text("Dots"),
				r.NewBadge("").Color("green").BadgeSize("sm").Dot().Build(),
				r.NewBadge("").Color("blue").BadgeSize("md").Dot().Build(),
				r.NewBadge("").Color("red").BadgeSize("lg").Dot().Build(),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Cards
// ---------------------------------------------------------------------------

func renderCardSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Cards"),
		r.Div("grid grid-cols-1 md:grid-cols-3 gap-6").Render(
			r.NewCard().
				CardVariant("shadowed").
				CardBody(r.Div().Render(
					r.Div("font-bold mb-2").Text("Standard Card"),
					r.P("text-gray-600 text-sm").Text("A standard shadowed card with default padding and footer."),
				)).
				CardFooter(r.Div("text-xs text-gray-500").Text("Card Footer")).
				Build(),
			r.NewCard().
				CardVariant("shadowed").
				CardImage(
					"https://images.unsplash.com/photo-1506744038136-46273834b3fb?w=640&auto=format&fit=crop&q=75",
					"Landscape",
				).
				CardImageSize("800", "493").
				CardImagePriority(true).
				CardBody(r.Div().Render(
					r.Div("font-bold mb-2").Text("Card with Image"),
					r.P("text-gray-600 text-sm").Text("Cards can display images at the top with hover effects."),
				)).
				CardHover(true).
				Build(),
			r.NewCard().
				CardVariant("glass").
				CardBody(r.Div().Render(
					r.Div("font-bold mb-2").Text("Glass Variant"),
					r.P("text-gray-600 text-sm").Text("This card uses a glassmorphism effect with backdrop blur."),
				)).
				CardHover(true).
				Build(),
		),
	)
}

// ---------------------------------------------------------------------------
// Progress Bars
// ---------------------------------------------------------------------------

func renderProgressSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Progress Bars"),
		r.Div("grid grid-cols-1 md:grid-cols-2 gap-8").Render(
			r.Div("flex flex-col gap-4").Render(
				r.NewProgress().
					ProgressValue(75).
					ProgressGradient("#3b82f6", "#8b5cf6").
					ProgressLabel("Gradient Style (75%)").
					LabelPosition("outside").
					Build(),
				r.NewProgress().
					ProgressValue(45).
					ProgressColor("bg-indigo-600").
					ProgressLabel("Outside Label (45%)").
					LabelPosition("outside").
					Build(),
			),
			r.Div("flex flex-col gap-4").Render(
				r.NewProgress().
					ProgressValue(65).
					ProgressColor("bg-green-500").
					ProgressLabel("Animated Stripes (65%)").
					LabelPosition("outside").
					Striped(true).
					Animated(true).
					ProgressSize("sm").
					Build(),
				r.NewProgress().
					ProgressColor("bg-blue-600").
					ProgressLabel("Indeterminate").
					LabelPosition("outside").
					Indeterminate(true).
					Build(),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Step Progress
// ---------------------------------------------------------------------------

func renderStepProgressSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Step Progress"),
		r.Div("grid grid-cols-1 md:grid-cols-2 gap-8").Render(
			r.Div("flex flex-col gap-4").Render(
				r.NewStepProgress(1, 4).Build(),
				r.NewStepProgress(2, 4).Build(),
				r.NewStepProgress(3, 4).Build(),
				r.NewStepProgress(4, 4).StepColor("bg-green-500").Build(),
			),
			r.Div("flex flex-col gap-4").Render(
				r.NewStepProgress(1, 5).StepColor("bg-purple-500").StepSize("sm").Build(),
				r.NewStepProgress(2, 5).StepColor("bg-yellow-500").StepSize("lg").Build(),
				r.NewStepProgress(3, 5).StepColor("bg-red-500").StepSize("xl").Build(),
				r.NewStepProgress(7, 10).StepColor("bg-indigo-500").Build(),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Tooltips
// ---------------------------------------------------------------------------

func renderTooltipSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Tooltips"),
		r.Div("flex flex-wrap gap-4").Render(
			r.NewTooltip("Default tooltip (top)").
				TooltipPosition("top").
				Wrap(r.NewButton("Hover me").BtnColor(r.BtnBlue).Build()),
			r.NewTooltip("Bottom position").
				TooltipPosition("bottom").
				Wrap(r.NewButton("Bottom").BtnColor(r.BtnGreen).Build()),
			r.NewTooltip("Success variant").
				TooltipVariant("green").
				Wrap(r.NewButton("Success").BtnColor(r.BtnGreenOutline).Build()),
			r.NewTooltip("Danger variant").
				TooltipVariant("red").
				Wrap(r.NewButton("Danger").BtnColor(r.BtnRedOutline).Build()),
			r.NewTooltip("Light variant").
				TooltipVariant("light").
				Wrap(r.NewButton("Light").BtnColor(r.BtnWhite).Build()),
		),
	)
}

// ---------------------------------------------------------------------------
// Tabs
// ---------------------------------------------------------------------------

func renderTabsSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Tabs"),
		r.Div("flex flex-col gap-8").Render(
			r.Div().Render(
				r.Div("text-sm font-bold text-gray-500 uppercase mb-3").Text("Boxed Style with Icons"),
				r.NewTabs().
					Tab("Home", r.Div().Render(
						r.Div("text-lg font-bold mb-2").Text("Home"),
						r.Div("text-gray-600").Text("Welcome to your central dashboard. This panel demonstrates how tabs can wrap complex content."),
					), "home").
					Tab("Profile", r.Div().Render(
						r.Div("text-lg font-bold mb-2").Text("Profile"),
						r.Div("text-gray-600").Text("Manage your personal information, display name, and avatar settings here."),
					), "person").
					Tab("Settings", r.Div().Render(
						r.Div("text-lg font-bold mb-2").Text("Settings"),
						r.Div("text-gray-600").Text("Fine-tune application behavior, notification preferences, and privacy controls."),
					), "settings").
					Active(0).
					TabStyle("boxed").
					Build(),
			),
			r.Div("grid grid-cols-1 md:grid-cols-2 gap-8").Render(
				r.Div().Render(
					r.Div("text-sm font-bold text-gray-500 uppercase mb-3").Text("Underline Style"),
					r.NewTabs().
						Tab("General", r.Div().Render(
							r.Div("text-lg font-bold mb-2").Text("General"),
							r.Div("text-gray-600").Text("Basic configuration for your workspace."),
						)).
						Tab("Security", r.Div().Render(
							r.Div("text-lg font-bold mb-2").Text("Security"),
							r.Div("text-gray-600").Text("Manage passwords, two-factor authentication and active sessions."),
						)).
						Active(0).
						TabStyle("underline").
						Build(),
				),
				r.Div().Render(
					r.Div("text-sm font-bold text-gray-500 uppercase mb-3").Text("Pills Style"),
					r.NewTabs().
						Tab("Daily", r.Div().Render(
							r.Div("text-lg font-bold mb-2").Text("Daily"),
							r.Div("text-gray-600").Text("Detailed activities for the last 24 hours."),
						)).
						Tab("Weekly", r.Div().Render(
							r.Div("text-lg font-bold mb-2").Text("Weekly"),
							r.Div("text-gray-600").Text("Summary of performance over the past 7 days."),
						)).
						Tab("Monthly", r.Div().Render(
							r.Div("text-lg font-bold mb-2").Text("Monthly"),
							r.Div("text-gray-600").Text("Strategic overview of goals achieved this month."),
						)).
						Active(1).
						TabStyle("pills").
						Build(),
				),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Accordion
// ---------------------------------------------------------------------------

func renderAccordionSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Accordion"),
		r.Div("grid grid-cols-1 md:grid-cols-2 gap-8").Render(
			r.Div().Render(
				r.Div("text-sm font-bold text-gray-500 uppercase mb-3").Text("Bordered with Default Open"),
				r.NewAccordion().
					Item("What is g-sui?",
						r.Div().Text("g-sui is a Go-based server-rendered UI framework that provides a component-based approach to building web applications."),
						true,
					).
					Item("How do I get started?",
						r.Div().Text("Simply import the ui package and start composing components."),
					).
					Item("Is it responsive?",
						r.Div().Text("All components are built with responsive design in mind, using Tailwind's responsive modifiers."),
					).
					Variant("bordered").
					Build(),
			),
			r.Div().Render(
				r.Div("text-sm font-bold text-gray-500 uppercase mb-3").Text("Separated Variant"),
				r.NewAccordion().
					Item("Separated Section 1",
						r.Div().Text("In the separated variant, each item is its own card."),
					).
					Item("Separated Section 2",
						r.Div().Text("Multiple sections can be open at once with the details element."),
					).
					Multiple(true).
					Variant("separated").
					Build(),
			),
		),
	)
}

// ---------------------------------------------------------------------------
// Dropdown Menus
// ---------------------------------------------------------------------------

func renderDropdownSection() *r.Node {
	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Dropdown Menus"),
		r.Div("flex flex-wrap gap-4").Render(
			r.NewDropdown(
				r.NewButton("Actions").BtnColor(r.BtnBlue).Build(),
			).
				DropdownHeader("General").
				DropdownItem("Edit Profile", r.JS("alert('Edit Profile')"), "edit").
				DropdownItem("Account Settings", r.JS("alert('Account Settings')")).
				DropdownDivider().
				DropdownHeader("Danger Zone").
				DropdownDanger("Delete Account", r.JS("alert('Delete Account')"), "delete").
				Build(),
			r.NewDropdown(
				r.NewButton("Options").BtnColor(r.BtnWhite).Build(),
			).
				DropdownItem("Share", r.JS("alert('Share')")).
				DropdownItem("Download", r.JS("alert('Download')")).
				Build(),
		),
	)
}

// ---------------------------------------------------------------------------
// Confirm Dialog
// ---------------------------------------------------------------------------

func renderConfirmDialogSection() *r.Node {
	dialogTargetID := r.Target()

	return r.Div().Render(
		r.Div("text-2xl font-bold mb-4").Text("Confirm Dialog"),
		r.Div("text-gray-600 text-sm mb-4").Text("Click the button below to show a confirmation dialog overlay."),
		r.Div().ID(dialogTargetID),
		r.NewButton("Delete Item").
			BtnColor(r.BtnRed).
			BtnIcon("delete").
			OnBtnClick(r.JS(
				r.ConfirmDialog(
					"Delete this item?",
					"This action cannot be undone. The item will be permanently removed.",
					r.JS("alert('Confirmed!');"+r.RemoveEl(dialogTargetID)),
				).ToJSInner(dialogTargetID),
			)).
			Build(),
	)
}

func RegisterShowcase(app *r.App, layout func(*r.Context, *r.Node) *r.Node) {
	app.Page("/", func(ctx *r.Context) *r.Node { return layout(ctx, Showcase(ctx)) })
	app.Action("nav.showcase", NavTo("/", func() *r.Node { return Showcase(nil) }))
}
