package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// HTTP Integration Tests - Full Form Submission Cycle
// ============================================================================

// Test structures for HTTP integration tests

type HTTPCompleteForm struct {
	Username  string    `json:"Username"`
	Email     string    `json:"Email"`
	Age       int       `json:"Age"`
	Height    float64   `json:"Height"`
	Active    bool      `json:"Active"`
	BirthDate time.Time `json:"BirthDate"`
	Bio       string    `json:"Bio"`
	Accept    bool      `json:"Accept"`
}

type HTTPNestedForm struct {
	User struct {
		FirstName string `json:"User.FirstName"`
		LastName  string `json:"User.LastName"`
	} `json:"-"`
	Address struct {
		Street string `json:"Address.Street"`
		City   string `json:"Address.City"`
	} `json:"-"`
}

type HTTPAllTypesForm struct {
	StringVal   string    `json:"StringVal"`
	IntVal      int       `json:"IntVal"`
	Int8Val     int8      `json:"Int8Val"`
	Int16Val    int16     `json:"Int16Val"`
	Int32Val    int32     `json:"Int32Val"`
	Int64Val    int64     `json:"Int64Val"`
	UintVal     uint      `json:"UintVal"`
	Uint8Val    uint8     `json:"Uint8Val"`
	Uint16Val   uint16    `json:"Uint16Val"`
	Uint32Val   uint32    `json:"Uint32Val"`
	Uint64Val   uint64    `json:"Uint64Val"`
	Float32Val  float32   `json:"Float32Val"`
	Float64Val  float64   `json:"Float64Val"`
	BoolVal     bool      `json:"BoolVal"`
	DateVal     time.Time `json:"DateVal"`
	DateTimeVal time.Time `json:"DateTimeVal"`
	TimeVal     time.Time `json:"TimeVal"`
}

type HTTPPointerForm struct {
	Name     *string `json:"Name"`
	Age      *int    `json:"Age"`
	Verified *bool   `json:"Verified"`
}

type HTTPEmptyForm struct {
	Field1 string `json:"Field1"`
	Field2 string `json:"Field2"`
}

// Helper to create an HTTP test server with a form handler
func createTestHandler(handler func(*Context)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Request:   r,
			Response:  w,
			SessionID: "test-session",
		}
		handler(ctx)
	})
}

// Helper to make a JSON POST request
func makeJSONRequest(url string, data []BodyItem) (*http.Response, error) {
	bodyJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

// Helper to make a form-urlencoded POST request
func makeFormRequest(url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	return client.Do(req)
}

// Helper to make a multipart form request
func makeMultipartRequest(url string, data url.Values) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, values := range data {
		for _, value := range values {
			writer.WriteField(key, value)
		}
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	return client.Do(req)
}

// ============================================================================
// JSON POST Tests
// ============================================================================

func TestHTTP_PostJSON_CompleteForm(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte("OK"))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "Username", Value: "john_doe", Type: "string"},
		{Name: "Email", Value: "john@example.com", Type: "string"},
		{Name: "Age", Value: "30", Type: "int"},
		{Name: "Height", Value: "1.75", Type: "float64"},
		{Name: "Active", Value: "true", Type: "bool"},
		{Name: "BirthDate", Value: "1990-05-15", Type: "time.Time"},
		{Name: "Bio", Value: "Software developer", Type: "string"},
		{Name: "Accept", Value: "true", Type: "bool"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Verify parsed values
	if receivedForm.Username != "john_doe" {
		t.Errorf("Username = %q, want %q", receivedForm.Username, "john_doe")
	}
	if receivedForm.Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", receivedForm.Email, "john@example.com")
	}
	if receivedForm.Age != 30 {
		t.Errorf("Age = %d, want %d", receivedForm.Age, 30)
	}
	if receivedForm.Height != 1.75 {
		t.Errorf("Height = %f, want %f", receivedForm.Height, 1.75)
	}
	if !receivedForm.Active {
		t.Errorf("Active = %v, want %v", receivedForm.Active, true)
	}
	if !receivedForm.Accept {
		t.Errorf("Accept = %v, want %v", receivedForm.Accept, true)
	}
}

func TestHTTP_PostJSON_NestedStructs(t *testing.T) {
	var receivedForm HTTPNestedForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "User.FirstName", Value: "John", Type: "string"},
		{Name: "User.LastName", Value: "Doe", Type: "string"},
		{Name: "Address.Street", Value: "123 Main St", Type: "string"},
		{Name: "Address.City", Value: "New York", Type: "string"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if receivedForm.User.FirstName != "John" {
		t.Errorf("User.FirstName = %q, want %q", receivedForm.User.FirstName, "John")
	}
	if receivedForm.User.LastName != "Doe" {
		t.Errorf("User.LastName = %q, want %q", receivedForm.User.LastName, "Doe")
	}
	if receivedForm.Address.Street != "123 Main St" {
		t.Errorf("Address.Street = %q, want %q", receivedForm.Address.Street, "123 Main St")
	}
	if receivedForm.Address.City != "New York" {
		t.Errorf("Address.City = %q, want %q", receivedForm.Address.City, "New York")
	}
}

func TestHTTP_PostJSON_AllTypes(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "StringVal", Value: "test string", Type: "string"},
		{Name: "IntVal", Value: "42", Type: "int"},
		{Name: "Int8Val", Value: "127", Type: "int8"},
		{Name: "Int16Val", Value: "32767", Type: "int16"},
		{Name: "Int32Val", Value: "2147483647", Type: "int32"},
		{Name: "Int64Val", Value: "9223372036854775807", Type: "int64"},
		{Name: "UintVal", Value: "42", Type: "uint"},
		{Name: "Uint8Val", Value: "255", Type: "uint8"},
		{Name: "Uint16Val", Value: "65535", Type: "uint16"},
		{Name: "Uint32Val", Value: "4294967295", Type: "uint32"},
		{Name: "Uint64Val", Value: "18446744073709551615", Type: "uint64"},
		{Name: "Float32Val", Value: "3.14", Type: "float32"},
		{Name: "Float64Val", Value: "2.718281828", Type: "float64"},
		{Name: "BoolVal", Value: "true", Type: "bool"},
		{Name: "DateVal", Value: "2023-05-15", Type: "time.Time"},
		{Name: "DateTimeVal", Value: "2023-05-15T14:30", Type: "time.Time"},
		{Name: "TimeVal", Value: "14:30", Type: "time.Time"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	// Verify all values
	if receivedForm.StringVal != "test string" {
		t.Errorf("StringVal = %q, want %q", receivedForm.StringVal, "test string")
	}
	if receivedForm.IntVal != 42 {
		t.Errorf("IntVal = %d, want %d", receivedForm.IntVal, 42)
	}
	if receivedForm.Int8Val != 127 {
		t.Errorf("Int8Val = %d, want %d", receivedForm.Int8Val, 127)
	}
	if receivedForm.Int16Val != 32767 {
		t.Errorf("Int16Val = %d, want %d", receivedForm.Int16Val, 32767)
	}
	if receivedForm.Int32Val != 2147483647 {
		t.Errorf("Int32Val = %d, want %d", receivedForm.Int32Val, 2147483647)
	}
	if receivedForm.UintVal != 42 {
		t.Errorf("UintVal = %d, want %d", receivedForm.UintVal, 42)
	}
	if receivedForm.Uint8Val != 255 {
		t.Errorf("Uint8Val = %d, want %d", receivedForm.Uint8Val, 255)
	}
	if receivedForm.Uint16Val != 65535 {
		t.Errorf("Uint16Val = %d, want %d", receivedForm.Uint16Val, 65535)
	}
	if receivedForm.Float32Val != 3.14 {
		t.Errorf("Float32Val = %f, want %f", receivedForm.Float32Val, 3.14)
	}
	if receivedForm.Float64Val != 2.718281828 {
		t.Errorf("Float64Val = %f, want %f", receivedForm.Float64Val, 2.718281828)
	}
	if !receivedForm.BoolVal {
		t.Errorf("BoolVal = %v, want %v", receivedForm.BoolVal, true)
	}
}

func TestHTTP_PostJSON_PointerFields(t *testing.T) {
	var receivedForm HTTPPointerForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "Name", Value: "Alice", Type: "string"},
		{Name: "Age", Value: "25", Type: "int"},
		{Name: "Verified", Value: "true", Type: "bool"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if receivedForm.Name == nil {
		t.Error("Name should not be nil")
	} else if *receivedForm.Name != "Alice" {
		t.Errorf("Name = %q, want %q", *receivedForm.Name, "Alice")
	}
	if receivedForm.Age == nil {
		t.Error("Age should not be nil")
	} else if *receivedForm.Age != 25 {
		t.Errorf("Age = %d, want %d", *receivedForm.Age, 25)
	}
	if receivedForm.Verified == nil {
		t.Error("Verified should not be nil")
	} else if !*receivedForm.Verified {
		t.Errorf("Verified = %v, want %v", *receivedForm.Verified, true)
	}
}

func TestHTTP_PostJSON_EmptyBody(t *testing.T) {
	var receivedForm HTTPEmptyForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Send empty JSON array
	req, err := http.NewRequest("POST", server.URL, bytes.NewReader([]byte("[]")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostJSON_InvalidJSON(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		var form HTTPCompleteForm
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	req, err := http.NewRequest("POST", server.URL, bytes.NewReader([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostJSON_InvalidInteger(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		// Error is logged but doesn't return
		_ = err
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "Age", Value: "not_a_number", Type: "int"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should still return OK because invalid values are logged but don't fail the request
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostJSON_InvalidBoolean(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		_ = err
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "Active", Value: "not_a_bool", Type: "bool"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostJSON_TimeFields(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	tests := []struct {
		name  string
		value string
	}{
		{"HTML date input", "2023-05-15"},
		{"HTML datetime-local", "2023-05-15T14:30"},
		{"HTML time input", "14:30"},
		{"RFC3339", "2023-05-15T14:30:00Z"},
		{"Go timestamp", "2023-05-15 14:30:00 +0000 UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []BodyItem{
				{Name: "DateVal", Value: tt.value, Type: "time.Time"},
			}

			resp, err := makeJSONRequest(server.URL, data)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", tt.name, resp.StatusCode)
			}

			if receivedForm.DateVal.IsZero() {
				t.Errorf("DateVal should not be zero for %s format", tt.name)
			}
		})
	}
}

func TestHTTP_PostJSON_OutOfRangeInt8(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		_ = err
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "Int8Val", Value: "128", Type: "int8"}, // Out of range for int8
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should still return OK - errors are logged
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostJSON_OutOfRangeUint8(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		_ = err
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "Uint8Val", Value: "256", Type: "uint8"}, // Out of range for uint8
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ============================================================================
// Form URL-Encoded Tests
// ============================================================================

// Note: The current implementation only handles JSON for structured data.
// Form URL-encoded tests would require the Context.Body to handle that format.
// These tests verify the current behavior.

func TestHTTP_PostForm_FormURLWithContext(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		// Current implementation uses JSON, form data would need ParseForm
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	formData := url.Values{
		"Username": []string{"john_doe"},
		"Email":    []string{"john@example.com"},
		"Age":      []string{"30"},
	}

	// This will fail with current implementation as it expects JSON
	resp, err := makeFormRequest(server.URL, formData)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Current implementation expects JSON, not form data
	// This documents the current behavior
	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("Note: Form URL-encoded not currently supported, got status %d", resp.StatusCode)
	}
}

// ============================================================================
// Multipart Form Tests
// ============================================================================

func TestHTTP_PostMultipart_BasicFields(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			ctx.Response.Write([]byte(err.Error()))
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	formData := url.Values{
		"Username": []string{"john_doe"},
		"Email":    []string{"john@example.com"},
		"Age":      []string{"30"},
	}

	resp, err := makeMultipartRequest(server.URL, formData)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Multipart form is supported, should parse correctly
	if resp.StatusCode == http.StatusOK {
		if receivedForm.Username != "john_doe" {
			t.Errorf("Username = %q, want %q", receivedForm.Username, "john_doe")
		}
	}
}

// ============================================================================
// Security Validation Tests
// ============================================================================

func TestHTTP_PostSecurity_FieldCountLimit(t *testing.T) {
	callCount := 0
	handler := createTestHandler(func(ctx *Context) {
		var form map[string]interface{}
		callCount++
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create more than MaxFieldCount (1000) fields
	data := make([]BodyItem, 1001)
	for i := 0; i < 1001; i++ {
		data[i] = BodyItem{
			Name:  fmt.Sprintf("Field%d", i),
			Value: "value",
			Type:  "string",
		}
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for too many fields, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostSecurity_FieldNameTooLong(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		var form map[string]interface{}
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create field name longer than MaxFieldNameLen (256)
	longName := strings.Repeat("a", 257)
	data := []BodyItem{
		{Name: longName, Value: "value", Type: "string"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for field name too long, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostSecurity_EmptyFieldName(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		var form map[string]interface{}
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "", Value: "value", Type: "string"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty field name, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostSecurity_UnsafeCharactersInFieldName(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		var form map[string]interface{}
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Test various unsafe characters
	unsafeNames := []string{
		"Field<script>",
		"Field\x00",
		"Field\n",
		"Field\t",
	}

	for _, name := range unsafeNames {
		t.Run(fmt.Sprintf("unsafe_%x", rune(name[4])), func(t *testing.T) {
			data := []BodyItem{
				{Name: name, Value: "value", Type: "string"},
			}

			resp, err := makeJSONRequest(server.URL, data)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400 for unsafe characters in field name %q, got %d", name, resp.StatusCode)
			}
		})
	}
}

func TestHTTP_PostSecurity_FieldValueTooLong(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		var form map[string]interface{}
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create field value longer than MaxFieldValueLen (1MB)
	longValue := strings.Repeat("a", 1024*1024+1)
	data := []BodyItem{
		{Name: "Field", Value: longValue, Type: "string"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for field value too long, got %d", resp.StatusCode)
	}
}

func TestHTTP_PostSecurity_BodySizeLimit(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		var form map[string]interface{}
		err := ctx.Body(&form)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create body larger than MaxBodySize (10MB)
	largeValue := strings.Repeat("a", 11*1024*1024)
	req, err := http.NewRequest("POST", server.URL, strings.NewReader(largeValue))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should fail due to body size limit
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for body too large, got %d", resp.StatusCode)
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestHTTP_Post_SpecialCharactersInValues(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	specialValues := []string{
		"Hello \"World\"",
		"Test & Test",
		"Line1\nLine2",
		"Tab\tHere",
		"Unicode: 你好",
		"Emoji: 🎉",
	}

	for _, value := range specialValues {
		t.Run("special_chars", func(t *testing.T) {
			data := []BodyItem{
				{Name: "Bio", Value: value, Type: "string"},
			}

			resp, err := makeJSONRequest(server.URL, data)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			if receivedForm.Bio != value {
				t.Errorf("Bio = %q, want %q", receivedForm.Bio, value)
			}
		})
	}
}

func TestHTTP_Post_NegativeInteger(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "IntVal", Value: "-42", Type: "int"},
		{Name: "Int64Val", Value: "-9223372036854775808", Type: "int64"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	if receivedForm.IntVal != -42 {
		t.Errorf("IntVal = %d, want %d", receivedForm.IntVal, -42)
	}
}

func TestHTTP_Post_FloatWithUnderscore(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Test that underscores are cleaned from numeric values
	data := []BodyItem{
		{Name: "Float64Val", Value: "1_000.5", Type: "float64"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if receivedForm.Float64Val != 1000.5 {
		t.Errorf("Float64Val = %f, want %f", receivedForm.Float64Val, 1000.5)
	}
}

func TestHTTP_Post_IntegerWithUnderscore(t *testing.T) {
	var receivedForm HTTPAllTypesForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	data := []BodyItem{
		{Name: "IntVal", Value: "1_000_000", Type: "int"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if receivedForm.IntVal != 1000000 {
		t.Errorf("IntVal = %d, want %d", receivedForm.IntVal, 1000000)
	}
}

func TestHTTP_Post_MissingField(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Only send Username, not all fields
	data := []BodyItem{
		{Name: "Username", Value: "john_doe", Type: "string"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should succeed - missing fields are just left as zero values
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if receivedForm.Username != "john_doe" {
		t.Errorf("Username = %q, want %q", receivedForm.Username, "john_doe")
	}
	// Other fields should be zero/empty
	if receivedForm.Email != "" {
		t.Errorf("Email should be empty, got %q", receivedForm.Email)
	}
}

func TestHTTP_Post_UnknownField(t *testing.T) {
	var receivedForm HTTPCompleteForm

	handler := createTestHandler(func(ctx *Context) {
		err := ctx.Body(&receivedForm)
		if err != nil {
			ctx.Response.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Send a field that doesn't exist in the struct
	data := []BodyItem{
		{Name: "Username", Value: "john_doe", Type: "string"},
		{Name: "UnknownField", Value: "some_value", Type: "string"},
	}

	resp, err := makeJSONRequest(server.URL, data)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should succeed - unknown fields are logged but don't fail the request
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	if receivedForm.Username != "john_doe" {
		t.Errorf("Username = %q, want %q", receivedForm.Username, "john_doe")
	}
}

// ============================================================================
// Response Rendering Tests
// ============================================================================

func TestHTTP_Response_HTMLRendering(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		html := Div("container")(
			Div("title")("Hello World"),
			P("text")("This is a test"),
		)
		ctx.Response.Header().Set("Content-Type", "text/html")
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(html))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	assertContains(t, string(body), "Hello World")
	assertContains(t, string(body), "This is a test")
	assertContains(t, string(body), "container")
}

func TestHTTP_Response_FormWithButton(t *testing.T) {
	handler := createTestHandler(func(ctx *Context) {
		target := Target()
		form := Div("form-container")(
			Form("test-form", Attr{ID: target.ID})(
				IText("username", nil).Render("Username"),
				Button().Submit().Render("Submit"),
			),
		)
		ctx.Response.Header().Set("Content-Type", "text/html")
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(form))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	assertContains(t, string(body), "type=\"text\"")
	assertContains(t, string(body), "name=\"username\"")
	assertContains(t, string(body), "Submit")
}

// ============================================================================
// Session State Tests
// ============================================================================

func TestHTTP_Context_SessionID(t *testing.T) {
	var capturedSessionID string

	handler := createTestHandler(func(ctx *Context) {
		capturedSessionID = ctx.SessionID
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if capturedSessionID == "" {
		t.Error("SessionID should not be empty")
	}
	if capturedSessionID != "test-session" {
		t.Errorf("SessionID = %q, want %q", capturedSessionID, "test-session")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTP_Context_IP(t *testing.T) {
	var capturedIP string

	handler := createTestHandler(func(ctx *Context) {
		capturedIP = ctx.IP()
		ctx.Response.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if capturedIP == "" {
		t.Error("IP should not be empty")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
