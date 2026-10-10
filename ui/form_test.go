package ui

import "testing"

func TestFormValidate(t *testing.T) {
	f := NewForm("f")
	f.Text("Name", "name").Required()
	f.Text("Code", "code").PatternValidation("[0-9]+").Err("Digits only")
	errs := f.Validate(map[string]any{"code": "a1"})
	if errs["name"] != "Name is required" || errs["code"] != "Digits only" {
		t.Fatalf("errors %v", errs)
	}
	if errs := f.Validate(map[string]any{"name": "Ann", "code": "12"}); len(errs) != 0 {
		t.Fatalf("errors %v", errs)
	}
}
