package ui

import (
	"strings"
	"testing"
)

func TestInputComponents_FormAttribute(t *testing.T) {
	formID := "test-form-123"

	tests := []struct {
		name           string
		component      func() string
		expectedFormID string
	}{
		{
			name: "IText with Form",
			component: func() string {
				return IText("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IArea with Form",
			component: func() string {
				return IArea("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IPassword with Form",
			component: func() string {
				return IPassword("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "INumber with Form",
			component: func() string {
				return INumber("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IDate with Form",
			component: func() string {
				return IDate("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "ITime with Form",
			component: func() string {
				return ITime("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IDateTime with Form",
			component: func() string {
				return IDateTime("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "ICheckbox with Form",
			component: func() string {
				return ICheckbox("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IRadio with Form",
			component: func() string {
				return IRadio("test", nil).Form(formID).Value("value1").Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IPhone with Form",
			component: func() string {
				return IPhone("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IEmail with Form",
			component: func() string {
				return IEmail("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "ISelect with Form",
			component: func() string {
				return ISelect("test", nil).Form(formID).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IRadioButtons with Form",
			component: func() string {
				options := []AOption{{ID: "1", Value: "Option 1"}}
				return IRadioButtons("test", nil).Form(formID).Options(options).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "IRadioDiv with Form",
			component: func() string {
				options := []AOption{{ID: "1", Value: "Option 1"}}
				return IRadioDiv("test", nil).Form(formID).Options(options).Render("Test Label")
			},
			expectedFormID: formID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := tt.component()
			
			// Check that the form attribute is present in the HTML
			formAttr := `form="` + tt.expectedFormID + `"`
			if !strings.Contains(html, formAttr) {
				t.Errorf("%s: expected to find %q in HTML output, but it was not found.\nHTML output:\n%s", tt.name, formAttr, html)
			}
		})
	}
}

func TestFormInstance_AllComponents(t *testing.T) {
	// Create a FormInstance to test that all methods properly set the Form ID
	form := FormNew(Attr{})
	formID := form.FormId

	tests := []struct {
		name           string
		component      func() string
		expectedFormID string
	}{
		{
			name: "FormInstance.Text",
			component: func() string {
				return form.Text("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Area",
			component: func() string {
				return form.Area("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Password",
			component: func() string {
				return form.Password("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Number",
			component: func() string {
				return form.Number("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Date",
			component: func() string {
				return form.Date("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Time",
			component: func() string {
				return form.Time("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.DateTime",
			component: func() string {
				return form.DateTime("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Checkbox",
			component: func() string {
				return form.Checkbox("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Radio",
			component: func() string {
				return form.Radio("test", nil).Value("value1").Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Phone",
			component: func() string {
				return form.Phone("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Email",
			component: func() string {
				return form.Email("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.Select",
			component: func() string {
				return form.Select("test", nil).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.RadioButtons",
			component: func() string {
				options := []AOption{{ID: "1", Value: "Option 1"}}
				return form.RadioButtons("test", nil).Options(options).Render("Test Label")
			},
			expectedFormID: formID,
		},
		{
			name: "FormInstance.RadioDiv",
			component: func() string {
				options := []AOption{{ID: "1", Value: "Option 1"}}
				return form.RadioDiv("test", nil).Options(options).Render("Test Label")
			},
			expectedFormID: formID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := tt.component()
			
			// Check that the form attribute is present in the HTML
			formAttr := `form="` + tt.expectedFormID + `"`
			if !strings.Contains(html, formAttr) {
				t.Errorf("%s: expected to find %q in HTML output, but it was not found.\nHTML output:\n%s", tt.name, formAttr, html)
			}
		})
	}
}

