package ui

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

type testUser struct {
	ID       uint
	Name     string
	Email    string
	Active   bool
	Posts    []testPost
	Created  time.Time
	NilField *string
}

type testPost struct {
	ID    uint
	Title string
	Text  string
}

func TestEscapeAttr_HTMLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "script tag injection",
			input:    `<script>alert(1)</script>`,
			contains: []string{"&lt;", "&gt;"},
			excludes: []string{"<script>", "</script>"},
		},
		{
			name:     "onload attribute injection",
			input:    `onload="alert(1)"`,
			contains: []string{"&#34;", "onload=", "alert(1)"},
			excludes: []string{`onload="alert(1)"`},
		},
		{
			name:     "multiple angle brackets",
			input:    `<<<>>>`,
			contains: []string{"&lt;&lt;&lt;", "&gt;&gt;&gt;"},
			excludes: []string{"<<<", ">>>"},
		},
		{
			name:     "ampersand",
			input:    `&`,
			contains: []string{"&amp;"},
			excludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAttr(tt.input)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("escapeAttr(%q) = %q, expected to contain %q", tt.input, result, substr)
				}
			}

			for _, substr := range tt.excludes {
				if strings.Contains(result, substr) {
					t.Errorf("escapeAttr(%q) = %q, should NOT contain %q", tt.input, result, substr)
				}
			}
		})
	}
}

func TestEscapeAttr_Unicode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"chinese", "中文"},
		{"japanese", "日本語"},
		{"korean", "한글"},
		{"emoji", "😀🎉"},
		{"arabic", "مرحبا"},
		{"cyrillic", "Привет"},
		{"mixed", "Hello 世界 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAttr(tt.input)
			if result != tt.input {
				t.Errorf("escapeAttr(%q) = %q, expected %q (unicode should be preserved)", tt.input, result, tt.input)
			}
		})
	}
}

func TestEscapeAttr_SpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"double quote", `"`, "&#34;"},
		{"single quote", "'", "&#39;"},
		{"less than", "<", "&lt;"},
		{"greater than", ">", "&gt;"},
		{"ampersand", "&", "&amp;"},
		{"all special chars", `&<>"'`, "&amp;&lt;&gt;&#34;&#39;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAttr(tt.input)
			if result != tt.expected {
				t.Errorf("escapeAttr(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeJS_DoubleQuotes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple string", "test"},
		{"with double quotes", `"hello"`},
		{"with single quotes", `'hello'`},
		{"with both quotes", `"hello'world"`},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)

			marshaled, _ := json.Marshal(tt.input)
			expected := string(marshaled[1 : len(marshaled)-1])

			if result != expected {
				t.Errorf("escapeJS(%q) = %q, expected %q", tt.input, result, expected)
			}

			resultWithQuotes := `"` + result + `"`
			var unmarshaled string
			if err := json.Unmarshal([]byte(resultWithQuotes), &unmarshaled); err != nil {
				t.Errorf("escapeJS(%q) produced invalid JSON string: %v", tt.input, err)
			} else if unmarshaled != tt.input {
				t.Errorf("escapeJS(%q) roundtrip failed: got %q", tt.input, unmarshaled)
			}
		})
	}
}

func TestEscapeJS_UnicodeEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"quote escape", `"`},
		{"less than escape", `<`},
		{"greater than escape", `>`},
		{"backslash escape", `\`},
		{"newline escape", "\n"},
		{"tab escape", "\t"},
		{"carriage return", "\r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)

			resultWithQuotes := `"` + result + `"`
			var unmarshaled string
			if err := json.Unmarshal([]byte(resultWithQuotes), &unmarshaled); err != nil {
				t.Errorf("escapeJS(%q) produced invalid JSON: %v", tt.input, err)
			} else if unmarshaled != tt.input {
				t.Errorf("escapeJS(%q) = %q, roundtrip to %q", tt.input, result, unmarshaled)
			}
		})
	}
}

func TestEscapeJS_NewlinesAndTabs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"newline", "line1\nline2"},
		{"tab", "col1\tcol2"},
		{"carriage return", "text\r"},
		{"mixed whitespace", "line1\n\tline2\r\nline3"},
		{"multiple newlines", "\n\n\n"},
		{"multiple tabs", "\t\t\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)

			if strings.Contains(result, "\n") {
				t.Errorf("escapeJS(%q) contains literal newline", tt.input)
			}
			if strings.Contains(result, "\t") {
				t.Errorf("escapeJS(%q) contains literal tab", tt.input)
			}
			if strings.Contains(result, "\r") {
				t.Errorf("escapeJS(%q) contains literal carriage return", tt.input)
			}

			resultWithQuotes := `"` + result + `"`
			var unmarshaled string
			if err := json.Unmarshal([]byte(resultWithQuotes), &unmarshaled); err != nil {
				t.Errorf("escapeJS(%q) produced invalid JSON: %v", tt.input, err)
			} else if unmarshaled != tt.input {
				t.Errorf("escapeJS(%q) roundtrip failed: got %q, want %q", tt.input, unmarshaled, tt.input)
			}
		})
	}
}

func TestEscapeJS_Chinese(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple chinese", "中文"},
		{"chinese with quotes", `"中文"`},
		{"chinese with mixed", "Hello 世界"},
		{"japanese", "日本語"},
		{"korean", "한글"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)

			resultWithQuotes := `"` + result + `"`
			var unmarshaled string
			if err := json.Unmarshal([]byte(resultWithQuotes), &unmarshaled); err != nil {
				t.Errorf("escapeJS(%q) produced invalid JSON: %v", tt.input, err)
			} else if unmarshaled != tt.input {
				t.Errorf("escapeJS(%q) roundtrip failed: got %q, want %q", tt.input, unmarshaled, tt.input)
			}
		})
	}
}

func TestEscapeJS_ScriptTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"opening script", "<script>"},
		{"closing script", "</script>"},
		{"full script", "<script>alert(1)</script>"},
		{"script tag variant", "<Script>alert(1)</Script>"},
		{"javascript protocol", "javascript:alert(1)"},
		{"onclick handler", "onclick='alert(1)'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)

			resultWithQuotes := `"` + result + `"`
			var unmarshaled string
			if err := json.Unmarshal([]byte(resultWithQuotes), &unmarshaled); err != nil {
				t.Errorf("escapeJS(%q) produced invalid JSON: %v", tt.input, err)
			} else if unmarshaled != tt.input {
				t.Errorf("escapeJS(%q) roundtrip failed", tt.input)
			}
		})
	}
}

func TestEscapeJS_XSSVectors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"basic script", "<script>alert('XSS')</script>"},
		{"img onerror", "<img onerror='alert(1)'>"},
		{"svg on load", "<svg onload='alert(1)'>"},
		{"expression", "<img src=x onerror=alert(1)>"},
		{"data URI", "data:text/html,<script>alert(1)</script>"},
		{"event handler", "javascript:void(document.location='http://evil.com')"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)

			resultWithQuotes := `"` + result + `"`
			var unmarshaled string
			if err := json.Unmarshal([]byte(resultWithQuotes), &unmarshaled); err != nil {
				t.Errorf("escapeJS(%q) produced invalid JSON: %v", tt.input, err)
			} else if unmarshaled != tt.input {
				t.Errorf("escapeJS(%q) roundtrip failed", tt.input)
			}
		})
	}
}

func TestValidateFieldAccess_RejectDangerousPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"os package access", "os.Exec", true},
		{"exec package", "exec.Command", true},
		{"syscall package", "syscall.Write", true},
		{"runtime package", "runtime.GOOS", true},
		{"unsafe package", "unsafe.Pointer", true},
		{"reflect package", "reflect.ValueOf", true},
		{"double underscore", "User.__private", true},
		{"os mixed case", "User.os.Exec", true},
		{"exec mixed case", "User.exec.Command", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldAccess(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldAccess(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFieldAccess_LongPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"exactly 256 chars", strings.Repeat("a.", 127) + "a", false},
		{"257 chars", strings.Repeat("a.", 128) + "a", true},
		{"1000 chars", strings.Repeat("a.", 500), true},
		{"short path", "User.Name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldAccess(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldAccess(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFieldAcceptSafePaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"simple field", "User.Name", false},
		{"nested field", "User.Post.Comments", false},
		{"slice index", "Posts[0]", false},
		{"nested slice", "Posts[0].Comments[1]", false},
		{"field starting with underscore", "User._id", false},
		{"mixed case", "User.PostTitle", false},
		{"deep nesting", "A.B.C.D.E", false},
		{"with numbers", "Post1.User2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldAccess(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldAccess(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFieldAccess_InvalidFieldNames(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"starts with number", "User.1Name", true},
		{"starts with special char", "User.$name", true},
		{"contains dash", "User.First-Name", true},
		{"contains space", "User.First Name", true},
		{"empty field", "User..Name", true},
		{"just dot", ".", true},
		{"ends with dot", "User.", true},
		{"starts with dot", ".User", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldAccess(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldAccess(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestPathValue_SafeFieldAccess(t *testing.T) {
	user := testUser{
		ID:     1,
		Name:   "John Doe",
		Email:  "john@example.com",
		Active: true,
	}

	tests := []struct {
		name    string
		obj     interface{}
		path    string
		wantVal interface{}
		wantErr bool
	}{
		{"simple field", user, "Name", "John Doe", false},
		{"numeric field", user, "ID", uint(1), false},
		{"boolean field", user, "Active", true, false},
		{"nested field", user, "Name", "John Doe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := PathValue(tt.obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathValue(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if !tt.wantErr && val.Interface() != tt.wantVal {
				t.Errorf("PathValue(%q) = %v, want %v", tt.path, val.Interface(), tt.wantVal)
			}
		})
	}
}

func TestPathValue_SliceAccess(t *testing.T) {
	posts := []testPost{
		{ID: 1, Title: "Post 1", Text: "Content 1"},
		{ID: 2, Title: "Post 2", Text: "Content 2"},
	}
	user := testUser{
		Name:  "John",
		Posts: posts,
	}

	tests := []struct {
		name    string
		obj     interface{}
		path    string
		wantVal interface{}
		wantErr bool
	}{
		{"first item", user, "Posts[0].Title", "Post 1", false},
		{"second item", user, "Posts[1].Title", "Post 2", false},
		{"deep slice access", user, "Posts[0].Text", "Content 1", false},
		{"non-existent field", user, "Posts[0].Missing", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := PathValue(tt.obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathValue(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if !tt.wantErr && val.Interface() != tt.wantVal {
				t.Errorf("PathValue(%q) = %v, want %v", tt.path, val.Interface(), tt.wantVal)
			}
		})
	}
}

func TestPathValue_ErrorHandling(t *testing.T) {
	user := testUser{
		Name: "John",
	}

	tests := []struct {
		name    string
		obj     interface{}
		path    string
		wantErr bool
	}{
		{"invalid field", user, "InvalidField", true},
		{"nil pointer", user, "NilField.Value", true},
		{"unsafe path", user, "os.Exec", true},
		{"out of bounds", user, "Posts[999]", true},
		{"empty path", user, "", true},
		{"double underscore", user, "User.__private", true},
		{"too long path", user, strings.Repeat("a.", 100) + "b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := PathValue(tt.obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathValue(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if tt.wantErr && val != nil {
				t.Errorf("PathValue(%q) on error should return nil, got %v", tt.path, val)
			}
		})
	}
}

func TestPathValue_NilPointerHandling(t *testing.T) {
	user := testUser{
		Name:     "John",
		NilField: nil,
	}

	tests := []struct {
		name    string
		obj     interface{}
		path    string
		wantErr bool
	}{
		{"nil pointer field", user, "NilField", false},
		{"access nil pointer", user, "NilField.Value", true},
		{"nil nested pointer", user, "NilField.Nested", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := PathValue(tt.obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathValue(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if !tt.wantErr && !val.IsValid() {
				t.Errorf("PathValue(%q) returned invalid value", tt.path)
			}
		})
	}
}

func TestPathValue_DynamicSliceExpansion(t *testing.T) {
	posts := []testPost{{ID: 1, Title: "First"}}

	tests := []struct {
		name     string
		obj      interface{}
		path     string
		wantType reflect.Kind
		wantErr  bool
	}{
		{"existing index", posts, "[0]", reflect.Struct, false},
		{"next index expands", posts, "[1]", reflect.Struct, true},
		{"expand multiple", posts, "[2]", reflect.Struct, true},
		{"excessive expansion", posts, "[2000]", reflect.Struct, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := PathValue(tt.obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathValue(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if !tt.wantErr && val.Kind() != tt.wantType {
				t.Errorf("PathValue(%q) kind = %v, want %v", tt.path, val.Kind(), tt.wantType)
			}
		})
	}
}

func TestPathValue_UnexportedFieldAccess(t *testing.T) {
	type testStruct struct {
		Exported   string
		unexported string
	}

	obj := testStruct{
		Exported:   "public",
		unexported: "private",
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"access exported field", "Exported", false},
		{"access unexported field", "unexported", true},
		{"mixed case unexported", "Unexported", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := PathValue(obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathValue(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !val.IsValid() {
				t.Errorf("PathValue(%q) returned invalid value", tt.path)
			}
		})
	}
}

func TestIsValidFieldName(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  bool
	}{
		{"simple", "Name", true},
		{"with underscore", "first_name", true},
		{"with numbers", "Post1", true},
		{"starts with letter", "User", true},
		{"starts with underscore", "_id", true},
		{"starts with number", "1name", false},
		{"starts with dash", "-name", false},
		{"with space", "first name", false},
		{"with special char", "name$", false},
		{"empty", "", false},
		{"single letter", "x", true},
		{"single underscore", "_", true},
		{"camel case", "FirstName", true},
		{"snake case", "first_name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFieldName(tt.field)
			if result != tt.want {
				t.Errorf("isValidFieldName(%q) = %v, want %v", tt.field, result, tt.want)
			}
		})
	}
}
