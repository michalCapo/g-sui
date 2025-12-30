package ui

import (
	"math"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestSetFieldValue_String(t *testing.T) {
	type TestStruct struct {
		Str string
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty string", "", "", false},
		{"simple string", "hello", "hello", false},
		{"unicode string", "Hello 世界 🌍", "Hello 世界 🌍", false},
		{"special chars", "test\"quote\nnewline", "test\"quote\nnewline", false},
		{"very long string", string(make([]byte, 10000)), string(make([]byte, 10000)), false},
		{"whitespace only", "   \t\n  ", "   \t\n  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName("Str")
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && obj.Str != tt.want {
				t.Errorf("setFieldValue() = %q, want %q", obj.Str, tt.want)
			}
		})
	}
}

func TestSetFieldValue_SignedIntegers(t *testing.T) {
	type TestStruct struct {
		Int   int
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64
	}

	tests := []struct {
		name     string
		field    string
		input    string
		wantErr  bool
		validate func(t *testing.T, obj *TestStruct)
	}{
		{"int zero", "Int", "0", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int != 0 {
				t.Errorf("Int = %d, want 0", obj.Int)
			}
		}},
		{"int positive", "Int", "42", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int != 42 {
				t.Errorf("Int = %d, want 42", obj.Int)
			}
		}},
		{"int negative", "Int", "-42", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int != -42 {
				t.Errorf("Int = %d, want -42", obj.Int)
			}
		}},
		{"int with underscores", "Int", "1_000_000", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int != 1000000 {
				t.Errorf("Int = %d, want 1000000", obj.Int)
			}
		}},
		{"int8 max", "Int8", "127", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int8 != 127 {
				t.Errorf("Int8 = %d, want 127", obj.Int8)
			}
		}},
		{"int8 min", "Int8", "-128", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int8 != -128 {
				t.Errorf("Int8 = %d, want -128", obj.Int8)
			}
		}},
		{"int8 overflow", "Int8", "128", true, nil},
		{"int8 underflow", "Int8", "-129", true, nil},
		{"int16 max", "Int16", "32767", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int16 != 32767 {
				t.Errorf("Int16 = %d, want 32767", obj.Int16)
			}
		}},
		{"int16 min", "Int16", "-32768", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int16 != -32768 {
				t.Errorf("Int16 = %d, want -32768", obj.Int16)
			}
		}},
		{"int32 max", "Int32", "2147483647", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int32 != 2147483647 {
				t.Errorf("Int32 = %d, want 2147483647", obj.Int32)
			}
		}},
		{"int32 min", "Int32", "-2147483648", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int32 != -2147483648 {
				t.Errorf("Int32 = %d, want -2147483648", obj.Int32)
			}
		}},
		{"int64 max", "Int64", "9223372036854775807", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int64 != 9223372036854775807 {
				t.Errorf("Int64 = %d, want 9223372036854775807", obj.Int64)
			}
		}},
		{"int64 min", "Int64", "-9223372036854775808", false, func(t *testing.T, obj *TestStruct) {
			if obj.Int64 != -9223372036854775808 {
				t.Errorf("Int64 = %d, want -9223372036854775808", obj.Int64)
			}
		}},
		{"invalid int", "Int", "not_a_number", true, nil},
		{"empty int", "Int", "", true, nil},
		{"int too long", "Int", "123456789012345678901234567890", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName(tt.field)
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, obj)
			}
		})
	}
}

func TestSetFieldValue_UnsignedIntegers(t *testing.T) {
	type TestStruct struct {
		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
	}

	tests := []struct {
		name     string
		field    string
		input    string
		wantErr  bool
		validate func(t *testing.T, obj *TestStruct)
	}{
		{"uint zero", "Uint", "0", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint != 0 {
				t.Errorf("Uint = %d, want 0", obj.Uint)
			}
		}},
		{"uint positive", "Uint", "42", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint != 42 {
				t.Errorf("Uint = %d, want 42", obj.Uint)
			}
		}},
		{"uint negative", "Uint", "-1", true, nil},
		{"uint8 max", "Uint8", "255", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint8 != 255 {
				t.Errorf("Uint8 = %d, want 255", obj.Uint8)
			}
		}},
		{"uint8 overflow", "Uint8", "256", true, nil},
		{"uint16 max", "Uint16", "65535", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint16 != 65535 {
				t.Errorf("Uint16 = %d, want 65535", obj.Uint16)
			}
		}},
		{"uint16 overflow", "Uint16", "65536", true, nil},
		{"uint32 max", "Uint32", "4294967295", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint32 != 4294967295 {
				t.Errorf("Uint32 = %d, want 4294967295", obj.Uint32)
			}
		}},
		{"uint32 overflow", "Uint32", "4294967296", true, nil},
		{"uint64 max", "Uint64", "18446744073709551615", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint64 != 18446744073709551615 {
				t.Errorf("Uint64 = %d, want 18446744073709551615", obj.Uint64)
			}
		}},
		{"uint with underscores", "Uint", "1_000_000", false, func(t *testing.T, obj *TestStruct) {
			if obj.Uint != 1000000 {
				t.Errorf("Uint = %d, want 1000000", obj.Uint)
			}
		}},
		{"invalid uint", "Uint", "not_a_number", true, nil},
		{"empty uint", "Uint", "", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName(tt.field)
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, obj)
			}
		})
	}
}

func TestSetFieldValue_Floats(t *testing.T) {
	type TestStruct struct {
		Float32 float32
		Float64 float64
	}

	tests := []struct {
		name     string
		field    string
		input    string
		wantErr  bool
		validate func(t *testing.T, obj *TestStruct)
	}{
		{"float32 zero", "Float32", "0", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float32 != 0 {
				t.Errorf("Float32 = %g, want 0", obj.Float32)
			}
		}},
		{"float32 positive", "Float32", "3.14", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float32 != 3.14 {
				t.Errorf("Float32 = %g, want 3.14", obj.Float32)
			}
		}},
		{"float32 negative", "Float32", "-3.14", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float32 != -3.14 {
				t.Errorf("Float32 = %g, want -3.14", obj.Float32)
			}
		}},
		{"float32 scientific", "Float32", "1.5e10", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float32 != 1.5e10 {
				t.Errorf("Float32 = %g, want 1.5e10", obj.Float32)
			}
		}},
		{"float64 zero", "Float64", "0", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float64 != 0 {
				t.Errorf("Float64 = %g, want 0", obj.Float64)
			}
		}},
		{"float64 positive", "Float64", "3.141592653589793", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float64 != 3.141592653589793 {
				t.Errorf("Float64 = %g, want 3.141592653589793", obj.Float64)
			}
		}},
		{"float64 negative", "Float64", "-3.141592653589793", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float64 != -3.141592653589793 {
				t.Errorf("Float64 = %g, want -3.141592653589793", obj.Float64)
			}
		}},
		{"float64 scientific", "Float64", "1.7976931348623157e+308", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float64 != 1.7976931348623157e+308 {
				t.Errorf("Float64 = %g, want 1.7976931348623157e+308", obj.Float64)
			}
		}},
		{"float64 with underscores", "Float64", "1_000.5", false, func(t *testing.T, obj *TestStruct) {
			if obj.Float64 != 1000.5 {
				t.Errorf("Float64 = %g, want 1000.5", obj.Float64)
			}
		}},
		{"float64 infinity", "Float64", "+Inf", false, func(t *testing.T, obj *TestStruct) {
			if !math.IsInf(float64(obj.Float64), 1) {
				t.Errorf("Float64 = %g, want +Inf", obj.Float64)
			}
		}},
		{"float64 negative infinity", "Float64", "-Inf", false, func(t *testing.T, obj *TestStruct) {
			if !math.IsInf(float64(obj.Float64), -1) {
				t.Errorf("Float64 = %g, want -Inf", obj.Float64)
			}
		}},
		{"float64 NaN", "Float64", "NaN", false, func(t *testing.T, obj *TestStruct) {
			if !math.IsNaN(float64(obj.Float64)) {
				t.Errorf("Float64 = %g, want NaN", obj.Float64)
			}
		}},
		{"invalid float", "Float64", "not_a_number", true, nil},
		{"empty float", "Float64", "", true, nil},
		{"float too long", "Float64", string(make([]byte, 100)), true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName(tt.field)
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, obj)
			}
		})
	}
}

func TestSetFieldValue_Bool(t *testing.T) {
	type TestStruct struct {
		Bool bool
	}

	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{"true", "true", true, false},
		{"false", "false", false, false},
		{"True (case)", "True", false, true},
		{"False (case)", "False", false, true},
		{"1", "1", false, true},
		{"0", "0", false, true},
		{"yes", "yes", false, true},
		{"no", "no", false, true},
		{"empty", "", false, true},
		{"invalid", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName("Bool")
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && obj.Bool != tt.want {
				t.Errorf("setFieldValue() = %v, want %v", obj.Bool, tt.want)
			}
		})
	}
}

func TestSetFieldValue_Time(t *testing.T) {
	type TestStruct struct {
		Time time.Time
	}

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, obj *TestStruct)
	}{
		{"date format", "2006-01-02", false, func(t *testing.T, obj *TestStruct) {
			expected := time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)
			if !obj.Time.Equal(expected) {
				t.Errorf("Time = %v, want %v", obj.Time, expected)
			}
		}},
		{"datetime-local format", "2006-01-02T15:04", false, func(t *testing.T, obj *TestStruct) {
			expected := time.Date(2006, 1, 2, 15, 4, 0, 0, time.UTC)
			if !obj.Time.Equal(expected) {
				t.Errorf("Time = %v, want %v", obj.Time, expected)
			}
		}},
		{"time format", "15:04", false, func(t *testing.T, obj *TestStruct) {
			// Time-only format sets to today's date with the given time
			if obj.Time.Hour() != 15 || obj.Time.Minute() != 4 {
				t.Errorf("Time = %v, want hour=15 min=4", obj.Time)
			}
		}},
		{"full timestamp", "2006-01-02T15:04:05-07:00", false, func(t *testing.T, obj *TestStruct) {
			// RFC3339 format with timezone offset
			if obj.Time.Year() != 2006 || obj.Time.Month() != 1 || obj.Time.Day() != 2 {
				t.Errorf("Time = %v, want year=2006 month=1 day=2", obj.Time)
			}
			if obj.Time.Hour() != 15 || obj.Time.Minute() != 4 {
				t.Errorf("Time hour=%d min=%d, want hour=15 min=4", obj.Time.Hour(), obj.Time.Minute())
			}
		}},
		{"RFC3339", "2006-01-02T15:04:05Z", false, func(t *testing.T, obj *TestStruct) {
			// RFC3339 format
			if obj.Time.Year() != 2006 || obj.Time.Month() != 1 || obj.Time.Day() != 2 {
				t.Errorf("Time = %v, want year=2006 month=1 day=2", obj.Time)
			}
		}},
		{"invalid time", "not a time", true, nil},
		{"empty time", "", false, func(t *testing.T, obj *TestStruct) {
			// Empty string should set zero time value
			if !obj.Time.IsZero() {
				t.Errorf("Time = %v, want zero time", obj.Time)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName("Time")
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, obj)
			}
		})
	}
}

func TestSetFieldValue_DeletedAt(t *testing.T) {
	type TestStruct struct {
		DeletedAt gorm.DeletedAt
	}

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, obj *TestStruct)
	}{
		{"valid date", "2006-01-02", false, func(t *testing.T, obj *TestStruct) {
			expected := time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)
			if !obj.DeletedAt.Time.Equal(expected) {
				t.Errorf("DeletedAt.Time = %v, want %v", obj.DeletedAt.Time, expected)
			}
			if !obj.DeletedAt.Valid {
				t.Errorf("DeletedAt.Valid = %v, want true", obj.DeletedAt.Valid)
			}
		}},
		{"datetime format", "2006-01-02T15:04", false, func(t *testing.T, obj *TestStruct) {
			expected := time.Date(2006, 1, 2, 15, 4, 0, 0, time.UTC)
			if !obj.DeletedAt.Time.Equal(expected) {
				t.Errorf("DeletedAt.Time = %v, want %v", obj.DeletedAt.Time, expected)
			}
			if !obj.DeletedAt.Valid {
				t.Errorf("DeletedAt.Valid = %v, want true", obj.DeletedAt.Valid)
			}
		}},
		{"invalid date", "not a date", true, nil},
		{"empty date", "", false, func(t *testing.T, obj *TestStruct) {
			// Empty string should set zero DeletedAt value
			if obj.DeletedAt.Valid {
				t.Errorf("DeletedAt.Valid = %v, want false", obj.DeletedAt.Valid)
			}
			if !obj.DeletedAt.Time.IsZero() {
				t.Errorf("DeletedAt.Time = %v, want zero time", obj.DeletedAt.Time)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName("DeletedAt")
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, obj)
			}
		})
	}
}

func TestSetFieldValue_Skeleton(t *testing.T) {
	type TestStruct struct {
		Skeleton Skeleton
	}

	tests := []struct {
		name    string
		input   string
		want    Skeleton
		wantErr bool
	}{
		{"simple skeleton", "list", Skeleton("list"), false},
		{"component skeleton", "component", Skeleton("component"), false},
		{"page skeleton", "page", Skeleton("page"), false},
		{"custom skeleton", "custom_value", Skeleton("custom_value"), false},
		{"empty skeleton", "", Skeleton(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName("Skeleton")
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && obj.Skeleton != tt.want {
				t.Errorf("setFieldValue() = %q, want %q", obj.Skeleton, tt.want)
			}
		})
	}
}

func TestSetFieldValue_TypeAlias(t *testing.T) {
	type MyString string
	type TestStruct struct {
		MyStr MyString
	}

	obj := &TestStruct{}
	fieldValue := reflect.ValueOf(obj).Elem().FieldByName("MyStr")
	err := setFieldValue(&fieldValue, "hello")
	if err != nil {
		t.Errorf("setFieldValue() error = %v, want nil", err)
	}
	if obj.MyStr != "hello" {
		t.Errorf("MyStr = %q, want %q", obj.MyStr, "hello")
	}
}

func TestSetFieldValue_Pointer(t *testing.T) {
	type TestStruct struct {
		StrPtr  *string
		IntPtr  *int
		BoolPtr *bool
	}

	tests := []struct {
		name     string
		field    string
		input    string
		wantErr  bool
		validate func(t *testing.T, obj *TestStruct)
	}{
		{"string pointer", "StrPtr", "hello", false, func(t *testing.T, obj *TestStruct) {
			if obj.StrPtr == nil {
				t.Error("StrPtr is nil")
				return
			}
			if *obj.StrPtr != "hello" {
				t.Errorf("StrPtr = %q, want %q", *obj.StrPtr, "hello")
			}
		}},
		{"int pointer", "IntPtr", "42", false, func(t *testing.T, obj *TestStruct) {
			if obj.IntPtr == nil {
				t.Error("IntPtr is nil")
				return
			}
			if *obj.IntPtr != 42 {
				t.Errorf("IntPtr = %d, want 42", *obj.IntPtr)
			}
		}},
		{"bool pointer true", "BoolPtr", "true", false, func(t *testing.T, obj *TestStruct) {
			if obj.BoolPtr == nil {
				t.Error("BoolPtr is nil")
				return
			}
			if *obj.BoolPtr != true {
				t.Errorf("BoolPtr = %v, want true", *obj.BoolPtr)
			}
		}},
		{"bool pointer false", "BoolPtr", "false", false, func(t *testing.T, obj *TestStruct) {
			if obj.BoolPtr == nil {
				t.Error("BoolPtr is nil")
				return
			}
			if *obj.BoolPtr != false {
				t.Errorf("BoolPtr = %v, want false", *obj.BoolPtr)
			}
		}},
		{"invalid int pointer", "IntPtr", "not_a_number", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName(tt.field)
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, obj)
			}
		})
	}
}

func TestSetFieldValue_AllTypes(t *testing.T) {
	type AllTypes struct {
		String    string
		Int       int
		Int8      int8
		Int16     int16
		Int32     int32
		Int64     int64
		Uint      uint
		Uint8     uint8
		Uint16    uint16
		Uint32    uint32
		Uint64    uint64
		Float32   float32
		Float64   float64
		Bool      bool
		Time      time.Time
		DeletedAt gorm.DeletedAt
		Skeleton  Skeleton
		StrPtr    *string
		IntPtr    *int
	}

	obj := &AllTypes{}
	objValue := reflect.ValueOf(obj).Elem()

	// Test all fields can be set
	tests := []struct {
		field string
		value string
	}{
		{"String", "test string"},
		{"Int", "42"},
		{"Int8", "127"},
		{"Int16", "32767"},
		{"Int32", "2147483647"},
		{"Int64", "9223372036854775807"},
		{"Uint", "42"},
		{"Uint8", "255"},
		{"Uint16", "65535"},
		{"Uint32", "4294967295"},
		{"Uint64", "18446744073709551615"},
		{"Float32", "3.14"},
		{"Float64", "3.141592653589793"},
		{"Bool", "true"},
		{"Time", "2006-01-02"},
		{"DeletedAt", "2006-01-02"},
		{"Skeleton", "list"},
		{"StrPtr", "pointer string"},
		{"IntPtr", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			fieldValue := objValue.FieldByName(tt.field)
			if !fieldValue.IsValid() {
				t.Fatalf("Field %s not found", tt.field)
			}
			err := setFieldValue(&fieldValue, tt.value)
			if err != nil {
				t.Errorf("setFieldValue() error = %v for field %s", err, tt.field)
			}
		})
	}
}

func TestSetFieldValue_EdgeCases(t *testing.T) {
	type TestStruct struct {
		Str   string
		Int   int
		Float float64
		Bool  bool
	}

	tests := []struct {
		name    string
		field   string
		input   string
		wantErr bool
	}{
		{"empty string", "Str", "", false},
		{"whitespace string", "Str", "   ", false},
		{"zero int", "Int", "0", false},
		{"zero float", "Float", "0", false},
		{"zero float decimal", "Float", "0.0", false},
		{"negative zero float", "Float", "-0", false},
		{"very large int", "Int", "999999999999999999999999999999999999999", true}, // Too large for int64
		{"very small int", "Int", "-999999999999999999999999999999999999999", true},
		{"float with many decimals", "Float", "3.141592653589793238462643383279", false},
		{"scientific notation small", "Float", "1e-10", false},
		{"scientific notation large", "Float", "1e10", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName(tt.field)
			err := setFieldValue(&fieldValue, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("setFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetFieldValue_UnsupportedType(t *testing.T) {
	type TestStruct struct {
		Slice []string
		Map   map[string]int
	}

	tests := []struct {
		name  string
		field string
		input string
	}{
		{"slice", "Slice", "test"},
		{"map", "Map", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &TestStruct{}
			fieldValue := reflect.ValueOf(obj).Elem().FieldByName(tt.field)
			err := setFieldValue(&fieldValue, tt.input)
			if err == nil {
				t.Errorf("setFieldValue() expected error for unsupported type %s", tt.field)
			}
		})
	}
}

func TestParseTimeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, result time.Time)
	}{
		{"date format", "2006-01-02", false, func(t *testing.T, result time.Time) {
			expected := time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)
			if !result.Equal(expected) {
				t.Errorf("parseTimeValue() = %v, want %v", result, expected)
			}
		}},
		{"datetime-local", "2006-01-02T15:04", false, func(t *testing.T, result time.Time) {
			expected := time.Date(2006, 1, 2, 15, 4, 0, 0, time.UTC)
			if !result.Equal(expected) {
				t.Errorf("parseTimeValue() = %v, want %v", result, expected)
			}
		}},
		{"time only", "15:04", false, func(t *testing.T, result time.Time) {
			if result.Hour() != 15 || result.Minute() != 4 {
				t.Errorf("parseTimeValue() hour=%d min=%d, want hour=15 min=4", result.Hour(), result.Minute())
			}
		}},
		{"full timestamp", "2006-01-02T15:04:05-07:00", false, func(t *testing.T, result time.Time) {
			// RFC3339 format with timezone offset
			if result.Year() != 2006 || result.Month() != 1 || result.Day() != 2 {
				t.Errorf("parseTimeValue() = %v, want year=2006 month=1 day=2", result)
			}
			if result.Hour() != 15 || result.Minute() != 4 {
				t.Errorf("parseTimeValue() hour=%d min=%d, want hour=15 min=4", result.Hour(), result.Minute())
			}
		}},
		{"RFC3339", "2006-01-02T15:04:05Z", false, func(t *testing.T, result time.Time) {
			expected, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
			if err != nil {
				t.Fatalf("Failed to parse expected time: %v", err)
			}
			if !result.Equal(expected) {
				t.Errorf("parseTimeValue() = %v, want %v", result, expected)
			}
		}},
		{"invalid", "not a time", true, nil},
		{"empty", "", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimeValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
