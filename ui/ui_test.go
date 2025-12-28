package ui

import (
	"strings"
	"testing"
)

func TestEscapeJS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double quotes",
			input:    `test"quote`,
			expected: `test\"quote`,
		},
		{
			name:     "backslash",
			input:    `test\slash`,
			expected: `test\\slash`,
		},
		{
			name:     "single quotes",
			input:    `test'quote`,
			expected: `test'quote`,
		},
		{
			name:     "newline",
			input:    "test\nnewline",
			expected: `test\nnewline`,
		},
		{
			name:     "carriage return",
			input:    "test\rcarriage",
			expected: `test\rcarriage`,
		},
		{
			name:     "tab",
			input:    "test\ttab",
			expected: `test\ttab`,
		},
		{
			name:     "chinese characters",
			input:    "中文测试",
			expected: "中文测试",
		},
		{
			name:     "mixed unicode",
			input:    "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
		},
		{
			name:     "script tag",
			input:    "</script>",
			expected: `\u003c/script\u003e`,
		},
		{
			name:     "javascript unicode escape for quote",
			input:    `\u0022`,
			expected: `\\u0022`,
		},
		{
			name:     "multiple backslashes",
			input:    `\\\\`,
			expected: `\\\\\\\\`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "complex mixed",
			input:    `test"quote\nwith\newline</script>`,
			expected: `test\"quote\\nwith\\newline\u003c/script\u003e`,
		},
		{
			name:     "emoji",
			input:    "😀🎉",
			expected: "😀🎉",
		},
		{
			name:     "unicode surrogate pair",
			input:    "\U0001F600",
			expected: "\U0001F600",
		},
		{
			name:     "null character",
			input:    "\x00",
			expected: `\u0000`,
		},
		{
			name:     "backspace",
			input:    "\b",
			expected: `\b`,
		},
		{
			name:     "form feed",
			input:    "\f",
			expected: `\f`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)
			if result != tt.expected {
				t.Errorf("escapeJS(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeJS_XSS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "unicode double quote escape attempt",
			input:    `\u0022`,
			contains: `\\u0022`,
		},
		{
			name:     "unicode single quote escape attempt",
			input:    `\u0027`,
			contains: `\\u0027`,
		},
		{
			name:     "unicode backslash escape attempt",
			input:    `\u005c`,
			contains: `\\u005c`,
		},
		{
			name:     "javascript injection attempt",
			input:    `";alert('XSS');//`,
			contains: `\";alert('XSS');//`,
		},
		{
			name:     "backtick injection",
			input:    "`alert('XSS')`",
			contains: "`alert('XSS')`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeJS(tt.input)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("escapeJS(%q) = %q, expected to contain %q", tt.input, result, tt.contains)
			}
		})
	}
}
