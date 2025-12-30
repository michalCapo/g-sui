package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

// Test data structures for form parsing tests

type CompleteForm struct {
	Username  string    `json:"Username"`
	Email     string    `json:"Email"`
	Age       int       `json:"Age"`
	Height    float64   `json:"Height"`
	Active    bool      `json:"Active"`
	BirthDate time.Time `json:"BirthDate"`
	LastLogin time.Time `json:"LastLogin"`
	Bio       string    `json:"Bio"`
	Accept    bool      `json:"Accept"`
	Score     int8      `json:"Score"`
	Count     uint16    `json:"Count"`
	BigNumber uint64    `json:"BigNumber"`
	Price     float32   `json:"Price"`
}

type NestedForm struct {
	User struct {
		FirstName string `json:"User.FirstName"`
		LastName  string `json:"User.LastName"`
	} `json:"-"`
	Address struct {
		Street string `json:"Address.Street"`
		City   string `json:"Address.City"`
	} `json:"-"`
}

type SliceForm struct {
	Tags   []string `json:"Tags[0]"`
	Scores []int    `json:"Scores[0]"`
}

type PointerForm struct {
	Name     *string `json:"Name"`
	Age      *int    `json:"Age"`
	Verified *bool   `json:"Verified"`
}

type DateTimeForm struct {
	StartDate time.Time `json:"StartDate"`
	EndDate   time.Time `json:"EndDate"`
	EventTime time.Time `json:"EventTime"`
}

type OptionalFieldsForm struct {
	Required string `json:"Required"`
	Optional string `json:"Optional"`
}

type DeletedAtForm struct {
	DeletedAt gorm.DeletedAt `json:"DeletedAt"`
}

type SkeletonTypeForm struct {
	Type Skeleton `json:"Type"`
}

// Helper function to create a test Context with JSON body
func makeTestContext(bodyData []BodyItem) *Context {
	bodyJSON, _ := json.Marshal(bodyData)
	req, _ := http.NewRequest("POST", "/test", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")

	return &Context{
		Request:   req,
		Response:  nil,
		SessionID: "test-session",
	}
}

// ============================================================================
// Context.Body Tests - Complete Form
// ============================================================================

func TestContextBody_CompleteForm(t *testing.T) {
	tests := []struct {
		name    string
		data    []BodyItem
		verify  func(*testing.T, *CompleteForm)
		wantErr bool
	}{
		{
			name: "All fields populated",
			data: []BodyItem{
				{Name: "Username", Value: "john_doe", Type: "string"},
				{Name: "Email", Value: "john@example.com", Type: "string"},
				{Name: "Age", Value: "30", Type: "int"},
				{Name: "Height", Value: "1.75", Type: "float64"},
				{Name: "Active", Value: "true", Type: "bool"},
				{Name: "BirthDate", Value: "1990-05-15", Type: "time.Time"},
				{Name: "Bio", Value: "Software developer", Type: "string"},
				{Name: "Accept", Value: "true", Type: "checkbox"},
				{Name: "Score", Value: "42", Type: "int8"},
				{Name: "Count", Value: "1000", Type: "uint16"},
				{Name: "BigNumber", Value: "18446744073709551615", Type: "uint64"},
				{Name: "Price", Value: "19.99", Type: "float32"},
			},
			verify: func(t *testing.T, form *CompleteForm) {
				if form.Username != "john_doe" {
					t.Errorf("Username = %q, want %q", form.Username, "john_doe")
				}
				if form.Email != "john@example.com" {
					t.Errorf("Email = %q, want %q", form.Email, "john@example.com")
				}
				if form.Age != 30 {
					t.Errorf("Age = %d, want 30", form.Age)
				}
				if form.Height != 1.75 {
					t.Errorf("Height = %v, want 1.75", form.Height)
				}
				if !form.Active {
					t.Errorf("Active = %v, want true", form.Active)
				}
				if form.BirthDate.Year() != 1990 || form.BirthDate.Month() != 5 || form.BirthDate.Day() != 15 {
					t.Errorf("BirthDate = %v, want 1990-05-15", form.BirthDate)
				}
				if form.Bio != "Software developer" {
					t.Errorf("Bio = %q, want 'Software developer'", form.Bio)
				}
				if !form.Accept {
					t.Errorf("Accept = %v, want true", form.Accept)
				}
				if form.Score != 42 {
					t.Errorf("Score = %d, want 42", form.Score)
				}
				if form.Count != 1000 {
					t.Errorf("Count = %d, want 1000", form.Count)
				}
				if form.BigNumber != 18446744073709551615 {
					t.Errorf("BigNumber = %d, want 1844674073709551615", form.BigNumber)
				}
				if form.Price != 19.99 {
					t.Errorf("Price = %v, want 19.99", form.Price)
				}
			},
			wantErr: false,
		},
		{
			name: "Boolean false",
			data: []BodyItem{
				{Name: "Active", Value: "false", Type: "bool"},
			},
			verify: func(t *testing.T, form *CompleteForm) {
				if form.Active {
					t.Errorf("Active = %v, want false", form.Active)
				}
			},
			wantErr: false,
		},
		{
			name: "Empty string",
			data: []BodyItem{
				{Name: "Username", Value: "", Type: "string"},
			},
			verify: func(t *testing.T, form *CompleteForm) {
				if form.Username != "" {
					t.Errorf("Username = %q, want empty string", form.Username)
				}
			},
			wantErr: false,
		},
		{
			name: "Invalid boolean",
			data: []BodyItem{
				{Name: "Active", Value: "yes", Type: "bool"},
			},
			verify:  func(t *testing.T, form *CompleteForm) {},
			wantErr: true, // HTML forms submit "on" or value, library expects "true" or "false"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &CompleteForm{}
			ctx := makeTestContext(tt.data)

			err := ctx.Body(form)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ctx.Body() error = %v, want nil", err)
			}

			tt.verify(t, form)
		})
	}
}

// ============================================================================
// Context.Body Tests - Nested Structs
// ============================================================================

func TestContextBody_NestedStructs(t *testing.T) {
	tests := []struct {
		name   string
		data   []BodyItem
		verify func(*testing.T, *NestedForm)
	}{
		{
			name: "Nested user fields",
			data: []BodyItem{
				{Name: "User.FirstName", Value: "John", Type: "string"},
				{Name: "User.LastName", Value: "Doe", Type: "string"},
				{Name: "Address.Street", Value: "123 Main St", Type: "string"},
				{Name: "Address.City", Value: "Boston", Type: "string"},
			},
			verify: func(t *testing.T, form *NestedForm) {
				if form.User.FirstName != "John" {
					t.Errorf("User.FirstName = %q, want John", form.User.FirstName)
				}
				if form.User.LastName != "Doe" {
					t.Errorf("User.LastName = %q, want Doe", form.User.LastName)
				}
				if form.Address.Street != "123 Main St" {
					t.Errorf("Address.Street = %q, want '123 Main St'", form.Address.Street)
				}
				if form.Address.City != "Boston" {
					t.Errorf("Address.City = %q, want Boston", form.Address.City)
				}
			},
		},
		{
			name: "Partial nested data",
			data: []BodyItem{
				{Name: "User.FirstName", Value: "Jane", Type: "string"},
				// User.LastName not provided - should remain zero
				{Name: "Address.City", Value: "NYC", Type: "string"},
			},
			verify: func(t *testing.T, form *NestedForm) {
				if form.User.FirstName != "Jane" {
					t.Errorf("User.FirstName = %q, want Jane", form.User.FirstName)
				}
				if form.User.LastName != "" {
					t.Errorf("User.LastName = %q, want empty string", form.User.LastName)
				}
				if form.Address.City != "NYC" {
					t.Errorf("Address.City = %q, want NYC", form.Address.City)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &NestedForm{}
			ctx := makeTestContext(tt.data)

			err := ctx.Body(form)
			if err != nil {
				t.Fatalf("ctx.Body() error = %v", err)
			}

			tt.verify(t, form)
		})
	}
}

// ============================================================================
// Context.Body Tests - Time Fields
// ============================================================================

func TestContextBody_TimeFields(t *testing.T) {
	tests := []struct {
		name    string
		data    []BodyItem
		verify  func(*testing.T, *DateTimeForm)
		wantErr bool
	}{
		{
			name: "Date format",
			data: []BodyItem{
				{Name: "StartDate", Value: "2023-12-25", Type: "time.Time"},
			},
			verify: func(t *testing.T, form *DateTimeForm) {
				expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
				if !form.StartDate.Equal(expected) {
					t.Errorf("StartDate = %v, want %v", form.StartDate, expected)
				}
			},
			wantErr: false,
		},
		{
			name: "DateTime-local format",
			data: []BodyItem{
				{Name: "EventTime", Value: "2023-06-15T14:30", Type: "time.Time"},
			},
			verify: func(t *testing.T, form *DateTimeForm) {
				expected := time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC)
				if !form.EventTime.Equal(expected) {
					t.Errorf("EventTime = %v, want %v", form.EventTime, expected)
				}
			},
			wantErr: false,
		},
		{
			name: "Time format",
			data: []BodyItem{
				{Name: "EventTime", Value: "14:30", Type: "time.Time"},
			},
			verify: func(t *testing.T, form *DateTimeForm) {
				// Time-only format - just check it parsed
				if form.EventTime.IsZero() {
					t.Errorf("EventTime should not be zero, got %v", form.EventTime)
				}
			},
			wantErr: false,
		},
		{
			name: "RFC3339 format",
			data: []BodyItem{
				{Name: "StartDate", Value: "2023-03-20T10:00:00Z", Type: "time.Time"},
			},
			verify: func(t *testing.T, form *DateTimeForm) {
				expected, _ := time.Parse(time.RFC3339, "2023-03-20T10:00:00Z")
				if !form.StartDate.Equal(expected) {
					t.Errorf("StartDate = %v, want %v", form.StartDate, expected)
				}
			},
			wantErr: false,
		},
		{
			name: "Invalid time",
			data: []BodyItem{
				{Name: "StartDate", Value: "not a time", Type: "time.Time"},
			},
			verify:  func(t *testing.T, form *DateTimeForm) {},
			wantErr: true, // Invalid time format should return error
		},
		{
			name: "Empty time",
			data: []BodyItem{
				{Name: "StartDate", Value: "", Type: "time.Time"},
			},
			verify:  func(t *testing.T, form *DateTimeForm) {},
			wantErr: true, // Empty time should return error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &DateTimeForm{}
			ctx := makeTestContext(tt.data)

			err := ctx.Body(form)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ctx.Body() error = %v", err)
			}

			tt.verify(t, form)
		})
	}
}

// ============================================================================
// Context.Body Tests - Pointer Fields
// ============================================================================

func TestContextBody_PointerFields(t *testing.T) {
	tests := []struct {
		name   string
		data   []BodyItem
		verify func(*testing.T, *PointerForm)
	}{
		{
			name: "String pointer with value",
			data: []BodyItem{
				{Name: "Name", Value: "Alice", Type: "string"},
			},
			verify: func(t *testing.T, form *PointerForm) {
				if form.Name == nil {
					t.Error("Name pointer is nil")
				}
				if *form.Name != "Alice" {
					t.Errorf("Name = %q, want Alice", *form.Name)
				}
			},
		},
		{
			name: "Int pointer with value",
			data: []BodyItem{
				{Name: "Age", Value: "25", Type: "int"},
			},
			verify: func(t *testing.T, form *PointerForm) {
				if form.Age == nil {
					t.Error("Age pointer is nil")
				}
				if *form.Age != 25 {
					t.Errorf("Age = %d, want 25", *form.Age)
				}
			},
		},
		{
			name: "Bool pointer true",
			data: []BodyItem{
				{Name: "Verified", Value: "true", Type: "bool"},
			},
			verify: func(t *testing.T, form *PointerForm) {
				if form.Verified == nil {
					t.Error("Verified pointer is nil")
				}
				if !*form.Verified {
					t.Errorf("Verified = %v, want true", *form.Verified)
				}
			},
		},
		{
			name: "Bool pointer false",
			data: []BodyItem{
				{Name: "Verified", Value: "false", Type: "bool"},
			},
			verify: func(t *testing.T, form *PointerForm) {
				if form.Verified == nil {
					t.Error("Verified pointer is nil")
				}
				if *form.Verified {
					t.Errorf("Verified = %v, want false", *form.Verified)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &PointerForm{}
			ctx := makeTestContext(tt.data)

			err := ctx.Body(form)
			if err != nil {
				t.Fatalf("ctx.Body() error = %v", err)
			}

			tt.verify(t, form)
		})
	}
}

// ============================================================================
// Context.Body Tests - Empty and Invalid Values
// ============================================================================

func TestContextBody_EmptyValues(t *testing.T) {
	tests := []struct {
		name    string
		data    []BodyItem
		verify  func(*testing.T, *CompleteForm)
		wantErr bool
	}{
		{
			name: "Empty string for optional field",
			data: []BodyItem{
				{Name: "Username", Value: "", Type: "string"},
			},
			verify: func(t *testing.T, form *CompleteForm) {
				if form.Username != "" {
					t.Errorf("Username = %q, want empty string", form.Username)
				}
			},
			wantErr: false,
		},
		{
			name: "Zero values",
			data: []BodyItem{
				{Name: "Age", Value: "0", Type: "int"},
				{Name: "Height", Value: "0", Type: "float64"},
				{Name: "Active", Value: "false", Type: "bool"},
			},
			verify: func(t *testing.T, form *CompleteForm) {
				if form.Age != 0 {
					t.Errorf("Age = %d, want 0", form.Age)
				}
				if form.Height != 0 {
					t.Errorf("Height = %v, want 0", form.Height)
				}
				if form.Active {
					t.Errorf("Active = %v, want false", form.Active)
				}
			},
			wantErr: false,
		},
		{
			name: "Negative number",
			data: []BodyItem{
				{Name: "Age", Value: "-5", Type: "int"},
				{Name: "Height", Value: "-1.5", Type: "float64"},
			},
			verify: func(t *testing.T, form *CompleteForm) {
				if form.Age != -5 {
					t.Errorf("Age = %d, want -5", form.Age)
				}
				if form.Height != -1.5 {
					t.Errorf("Height = %v, want -1.5", form.Height)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := &CompleteForm{}
			ctx := makeTestContext(tt.data)

			err := ctx.Body(form)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ctx.Body() error = %v", err)
			}

			tt.verify(t, form)
		})
	}
}

// ============================================================================
// Context.Body Tests - Special Types
// ============================================================================

func TestContextBody_DeletedAt(t *testing.T) {
	data := []BodyItem{
		{Name: "DeletedAt", Value: "2023-01-15", Type: "gorm.DeletedAt"},
	}

	form := &DeletedAtForm{}
	ctx := makeTestContext(data)

	err := ctx.Body(form)
	if err != nil {
		t.Fatalf("ctx.Body() error = %v", err)
	}

	if !form.DeletedAt.Valid {
		t.Errorf("DeletedAt.Valid should be true")
	}

	expected := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
	if !form.DeletedAt.Time.Equal(expected) {
		t.Errorf("DeletedAt.Time = %v, want %v", form.DeletedAt.Time, expected)
	}
}

func TestContextBody_SkeletonType(t *testing.T) {
	data := []BodyItem{
		{Name: "Type", Value: "list", Type: "string"},
	}

	form := &SkeletonTypeForm{}
	ctx := makeTestContext(data)

	err := ctx.Body(form)
	if err != nil {
		t.Fatalf("ctx.Body() error = %v", err)
	}

	if form.Type != "list" {
		t.Errorf("Type = %q, want 'list'", form.Type)
	}
}

// ============================================================================
// Context.Body Tests - Error Cases
// ============================================================================

func TestContextBody_InvalidIntegerOverflow(t *testing.T) {
	data := []BodyItem{
		{Name: "Age", Value: "999999999999999999999999999", Type: "int"},
	}

	form := &CompleteForm{}
	ctx := makeTestContext(data)

	err := ctx.Body(form)
	// Library returns error for integer overflow (value too long)
	if err == nil {
		t.Error("Expected error for integer overflow, got nil")
	}
}

func TestContextBody_InvalidFloat(t *testing.T) {
	data := []BodyItem{
		{Name: "Height", Value: "not_a_number", Type: "float64"},
	}

	form := &CompleteForm{}
	ctx := makeTestContext(data)

	err := ctx.Body(form)
	// Library returns error for invalid float values
	if err == nil {
		t.Error("Expected error for invalid float, got nil")
	}
}

func TestContextBody_InvalidUintWithNegative(t *testing.T) {
	data := []BodyItem{
		{Name: "Count", Value: "-1", Type: "uint"},
	}

	form := &CompleteForm{}
	ctx := makeTestContext(data)

	err := ctx.Body(form)
	// Library returns error for negative uint values
	if err == nil {
		t.Error("Expected error for negative uint, got nil")
	}
}

func TestContextBody_EmptyBody(t *testing.T) {
	form := &CompleteForm{}
	ctx := makeTestContext([]BodyItem{})

	err := ctx.Body(form)
	if err != nil {
		t.Errorf("ctx.Body() with empty body error = %v", err)
	}

	// All fields should remain zero values
	if form.Username != "" {
		t.Errorf("Username should be empty, got %q", form.Username)
	}
	if form.Age != 0 {
		t.Errorf("Age should be 0, got %d", form.Age)
	}
}

// ============================================================================
// Context.Body Tests - Field Name Safety
// ============================================================================

func TestValidateInputSafety_ValidFieldNames(t *testing.T) {
	tests := []struct {
		name     string
		data     []BodyItem
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Valid field names",
			data: []BodyItem{
				{Name: "Username", Value: "test", Type: "string"},
				{Name: "User.FirstName", Value: "test", Type: "string"},
				{Name: "Tags[0]", Value: "tag1", Type: "string"},
				{Name: "user_profile", Value: "value", Type: "string"},
			},
			wantErr: false,
		},
		{
			name:     "Too many fields",
			data:     make([]BodyItem, 1001),
			wantErr:  true,
			errorMsg: "too many fields",
		},
		{
			name: "Field name too long",
			data: []BodyItem{
				{Name: strings.Repeat("a", 257), Value: "test", Type: "string"},
			},
			wantErr:  true,
			errorMsg: "field name too long",
		},
		{
			name: "Empty field name",
			data: []BodyItem{
				{Name: "", Value: "test", Type: "string"},
			},
			wantErr:  true,
			errorMsg: "empty field name",
		},
		{
			name: "Invalid characters in field name",
			data: []BodyItem{
				{Name: "user<script>", Value: "test", Type: "string"},
			},
			wantErr:  true,
			errorMsg: "unsafe character",
		},
		{
			name: "Value too long",
			data: []BodyItem{
				{Name: "Username", Value: strings.Repeat("a", 1024*1024+1), Type: "string"}, // MaxFieldValueLen is 1MB
			},
			wantErr:  true,
			errorMsg: "field value too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputSafety(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing %q, but got nil", tt.errorMsg)
				}
				if err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error message = %q, expected to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateNumericInput_ValidInputs(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		inputType string
		expected  any
		wantErr   bool
	}{
		{
			name:      "Valid int",
			input:     "42",
			inputType: "int",
			expected:  42,
			wantErr:   false,
		},
		{
			name:      "Valid int with underscore",
			input:     "1_000_000",
			inputType: "int",
			expected:  1000000,
			wantErr:   false,
		},
		{
			name:      "Valid negative int",
			input:     "-42",
			inputType: "int",
			expected:  -42,
			wantErr:   false,
		},
		{
			name:      "Int overflow",
			input:     "999999999999999999999999",
			inputType: "int",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Invalid int",
			input:     "not_a_number",
			inputType: "int",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Valid float64",
			input:     "3.14159",
			inputType: "float64",
			expected:  3.14159,
			wantErr:   false,
		},
		{
			name:      "Float64 scientific notation",
			input:     "1.5e10",
			inputType: "float64",
			expected:  1.5e10,
			wantErr:   false,
		},
		{
			name:      "Invalid float64",
			input:     "not_a_number",
			inputType: "float64",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Valid uint",
			input:     "42",
			inputType: "uint",
			expected:  uint(42),
			wantErr:   false,
		},
		{
			name:      "Invalid uint negative",
			input:     "-1",
			inputType: "uint",
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Valid number",
			input:     "123",
			inputType: "number",
			expected:  123,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateNumericInput(tt.input, tt.inputType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %q, but got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("validateNumericInput(%q, %s) error = %v", tt.input, tt.inputType, err)
			}

			// Type assert the result
			switch tt.inputType {
			case "int":
				actual, ok := result.(int)
				if !ok || actual != tt.expected.(int) {
					t.Errorf("validateNumericInput(%q) = %v (%T), want %d (int)", tt.input, result, actual, tt.expected.(int))
				}
			case "uint":
				actual, ok := result.(uint)
				if !ok || actual != tt.expected.(uint) {
					t.Errorf("validateNumericInput(%q) = %v (%T), want %d (uint)", tt.input, result, actual, tt.expected.(uint))
				}
			case "float64":
				actual, ok := result.(float64)
				if !ok {
					t.Errorf("validateNumericInput(%q) = %v (%T), want float64", tt.input, result, tt.expected.(float64))
				} else {
					if actual != tt.expected.(float64) {
						t.Errorf("validateNumericInput(%q) = %v, want %v", tt.input, actual, tt.expected)
					}
				}
			case "number":
				actual, ok := result.(int)
				if !ok || actual != tt.expected.(int) {
					t.Errorf("validateNumericInput(%q) = %v (%T), want %d (int)", tt.input, result, actual, tt.expected.(int))
				}
			}
		})
	}
}

// ============================================================================
// ParseTimeValue Tests
// ============================================================================

func TestParseTimeValue_AllFormats(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantYear   int
		wantMonth  time.Month
		wantDay    int
		wantHour   int
		wantMinute int
		wantErr    bool
	}{
		{"Date format", "2006-01-02", 2006, time.January, 2, 0, 0, false},
		{"DateTime-local", "2006-01-02T15:04", 2006, time.January, 2, 15, 4, false},
		{"Time format", "15:04", 0, time.January, 1, 15, 4, false}, // Year 0 is expected for time-only
		{"RFC3339", "2006-01-02T10:00:00Z", 2006, time.January, 2, 10, 0, false},
		{"RFC3339Nano", "2006-01-02T10:00:00.123Z", 2006, time.January, 2, 10, 0, false},
		{"Invalid time", "not a time", 0, 0, 0, 0, 0, true},
		{"Empty time", "", 0, 0, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeValue(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %q, but got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseTimeValue(%q) error = %v", tt.input, err)
			}

			if result.Year() != tt.wantYear {
				t.Errorf("parseTimeValue(%q) year = %d, want %d", tt.input, result.Year(), tt.wantYear)
			}
			if result.Month() != tt.wantMonth {
				t.Errorf("parseTimeValue(%q) month = %d, want %d", tt.input, result.Month(), tt.wantMonth)
			}
			if result.Day() != tt.wantDay {
				t.Errorf("parseTimeValue(%q) day = %d, want %d", tt.input, result.Day(), tt.wantDay)
			}
			if result.Hour() != tt.wantHour {
				t.Errorf("parseTimeValue(%q) hour = %d, want %d", tt.input, result.Hour(), tt.wantHour)
			}
			if result.Minute() != tt.wantMinute {
				t.Errorf("parseTimeValue(%q) minute = %d, want %d", tt.input, result.Minute(), tt.wantMinute)
			}
		})
	}
}

// ============================================================================
// RandomString Tests
// ============================================================================

func TestRandomString_GeneratesCorrectLength(t *testing.T) {
	lengths := []int{1, 5, 10, 15, 20, 30}

	for _, length := range lengths {
		result := RandomString(length)

		if len(result) != length {
			t.Errorf("RandomString(%d) length = %d, want %d", length, len(result), length)
		}

		// Should only contain URL-safe base64 characters (alphanumeric + X, Y, Z replacements)
		for _, r := range result {
			isAlphanumeric := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
			if !isAlphanumeric {
				t.Errorf("RandomString(%d) contains invalid character: %c", length, r)
			}
		}
	}
}

func TestRandomString_DefaultLength(t *testing.T) {
	result := RandomString()

	if len(result) != 20 {
		t.Errorf("RandomString() default length = %d, want 20", len(result))
	}
}

func TestRandomString_Uniqueness(t *testing.T) {
	results := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result := RandomString(10)
		if results[result] {
			t.Errorf("RandomString() generated duplicate: %s", result)
		}
		results[result] = true
	}
}
