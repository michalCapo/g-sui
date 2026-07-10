package ui

import (
	"strings"
	"testing"
)

type scopeRow struct{ Name string }

func TestTwoTablesHaveScopedFilterControlIDs(t *testing.T) {
	build := func(id string) string {
		dt := NewDataTable[scopeRow](id).Action("t.data").
			Col("Name", ColOpt[scopeRow]{Filter: FilterTypeText, Text: func(r *scopeRow) *Node { return Span().Text(r.Name) }})
		return dt.Render([]*scopeRow{{Name: "x"}}).ToJS()
	}
	a, b := build("tblA"), build("tblB")
	if !strings.Contains(a, "tblA-filter-0-") {
		t.Fatalf("table A filter controls not scoped with table ID:\n%s", a[:min(2000, len(a))])
	}
	if strings.Contains(a, "'filter-0-") {
		t.Fatalf("table A still has unscoped filter control IDs")
	}
	if !strings.Contains(b, "tblB-filter-0-") {
		t.Fatalf("table B filter controls not scoped with table ID")
	}
}
