package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// Table (Generic) Tests
// ============================================================================

type Person struct {
	Name string
	Age  int
	City string
}

func TestTable_Generic_BasicRendering(t *testing.T) {
	table := Table[Person]("table-class").
		Head("Name", "text-left").
		Head("Age", "text-center").
		Head("City", "text-left").
		Field(func(p *Person) string { return p.Name }, "text-left").
		Field(func(p *Person) string { return string(rune('0' + p.Age)) }, "text-center").
		Field(func(p *Person) string { return p.City }, "text-left")

	data := []*Person{
		{Name: "Alice", Age: 3, City: "NYC"},
	}

	html := table.Render(data)

	// Should have table element
	assertContains(t, html, `<table`)
	assertContains(t, html, `table-class`)

	// Should have thead and tbody
	assertContains(t, html, `<thead`)
	assertContains(t, html, `<tbody`)

	// Should have data
	assertContains(t, html, "Alice")
	assertContains(t, html, "NYC")
}

func TestTable_Generic_HeadMethod(t *testing.T) {
	table := Table[Person]("").
		Head("Name", "class-name").
		Head("Age", "class-age")

	html := table.Render([]*Person{})

	assertContains(t, html, `class-name`)
	assertContains(t, html, `Name`)
	assertContains(t, html, `class-age`)
	assertContains(t, html, `Age`)
}

func TestTable_Generic_FieldMethod(t *testing.T) {
	table := Table[Person]("").
		Head("Name", "").
		Field(func(p *Person) string { return p.Name }, "p-2")

	data := []*Person{
		{Name: "Bob"},
	}

	html := table.Render(data)

	assertContains(t, html, `Bob`)
	assertContains(t, html, `class="p-2"`)
}

func TestTable_Generic_RenderWithData(t *testing.T) {
	table := Table[Person]("w-full").
		Head("Name", "text-left").
		Head("Age", "text-center").
		Head("City", "text-left").
		Field(func(p *Person) string { return p.Name }, "text-left").
		Field(func(p *Person) string { return string(rune('0' + p.Age)) }, "text-center").
		Field(func(p *Person) string { return p.City }, "text-left")

	data := []*Person{
		{Name: "Alice", Age: 3, City: "NYC"},
		{Name: "Bob", Age: 5, City: "LA"},
	}

	html := table.Render(data)

	assertContains(t, html, "Alice")
	assertContains(t, html, "Bob")
	assertContains(t, html, "NYC")
	assertContains(t, html, "LA")

	// Should have tr elements
	assertContains(t, html, `<tr>`)
}

func TestTable_Generic_EmptyData(t *testing.T) {
	table := Table[Person]("").
		Head("Name", "").
		Field(func(p *Person) string { return p.Name }, "")

	html := table.Render([]*Person{})

	// Should still render table structure
	assertContains(t, html, `<table`)
	assertContains(t, html, `<thead`)
	assertContains(t, html, `<tbody`)
}

func TestTable_Generic_HeadHTML_RawHTML(t *testing.T) {
	table := Table[Person]("").
		HeadHTML("<strong>Name</strong>", "text-left")

	html := table.Render([]*Person{})

	assertContains(t, html, `<strong>Name</strong>`)
}

func TestTable_Generic_FieldText_Escaping(t *testing.T) {
	table := Table[Person]("").
		Head("Name", "").
		FieldText(func(p *Person) string { return p.Name }, "")

	data := []*Person{
		{Name: "<script>alert('xss')</script>"},
	}

	html := table.Render(data)

	// Should escape HTML
	assertNotContains(t, html, `<script>`)
	assertContains(t, html, `&lt;script&gt;`)
}

// ============================================================================
// SimpleTable Tests
// ============================================================================

func TestTableSimple_BasicRendering(t *testing.T) {
	table := SimpleTable(2, "table-class").
		Field("Name").
		Field("Age")

	html := table.Render()

	// Should have table element
	assertContains(t, html, `<table`)
	assertContains(t, html, `table-class`)

	// Should have tbody
	assertContains(t, html, `<tbody`)

	// Should have fields
	assertContains(t, html, "Name")
	assertContains(t, html, "Age")
}

func TestTableSimple_AutoRowWrap(t *testing.T) {
	table := SimpleTable(2, "w-full").
		Field("Cell1").
		Field("Cell2").
		Field("Cell3").
		Field("Cell4").
		Field("Cell5").
		Field("Cell6")

	html := table.Render()

	// Count rows - should have multiple rows
	rowCount := strings.Count(html, `<tr>`)
	if rowCount < 2 {
		t.Errorf("Expected at least 2 rows, got %d", rowCount)
	}
}

func TestTableSimple_ClassMethod(t *testing.T) {
	table := SimpleTable(3, "").
		Class(0, "col-class-1").
		Class(1, "col-class-2").
		Class(2, "col-class-3").
		Field("A").
		Field("B").
		Field("C")

	html := table.Render()

	assertContains(t, html, `col-class-1`)
	assertContains(t, html, `col-class-2`)
	assertContains(t, html, `col-class-3`)
}

func TestTableSimple_FieldTextEscaping(t *testing.T) {
	table := SimpleTable(1, "").
		FieldText("<script>alert('xss')</script>", "")

	html := table.Render()

	// Should escape HTML
	assertNotContains(t, html, `<script>`)
	assertContains(t, html, `&lt;script&gt;`)
}

func TestTableSimple_EmptyMethod(t *testing.T) {
	table := SimpleTable(2, "").
		Field("Cell1").
		Empty().
		Field("Cell2")

	html := table.Render()

	assertContains(t, html, "Cell1")
	assertContains(t, html, "Cell2")
}

func TestTableSingle_AttrMethod(t *testing.T) {
	table := SimpleTable(2, "").
		Field("Cell1", "cell-class").
		Attr(`colspan="2"`)

	html := table.Render()

	assertContains(t, html, `colspan="2"`)
}

func TestTableSimple_MultipleRows(t *testing.T) {
	table := SimpleTable(2, "w-full").
		Field("A1").Field("A2").
		Field("B1").Field("B2").
		Field("C1").Field("C2")

	html := table.Render()

	assertContains(t, html, "A1")
	assertContains(t, html, "A2")
	assertContains(t, html, "B1")
	assertContains(t, html, "B2")
	assertContains(t, html, "C1")
	assertContains(t, html, "C2")
}

func TestTableSimple_NewTable(t *testing.T) {
	table := NewTable[Person]("")

	if table == nil {
		t.Error("NewTable should return a non-nil table")
	}
}

func TestTableSimple_FieldWithClass(t *testing.T) {
	table := SimpleTable(2, "").
		Field("Styled", "bg-blue-500 text-white")

	html := table.Render()

	assertContains(t, html, `bg-blue-500`)
	assertContains(t, html, `text-white`)
	assertContains(t, html, "Styled")
}

func TestTableSimple_IncompleteRow(t *testing.T) {
	table := SimpleTable(3, "").
		Field("Cell1").
		Field("Cell2")
		// Only 2 cells in a 3-column table

	html := table.Render()

	// Should still render and fill with empty cells
	assertContains(t, html, `Cell1`)
	assertContains(t, html, `Cell2`)
}

func TestTableSimple_ColspanHandling(t *testing.T) {
	table := SimpleTable(3, "").
		Field("Full width").
		Attr(`colspan="3"`)

	html := table.Render()

	// The Attr method adds the attribute to the cell
	assertContains(t, html, `colspan`)
	assertContains(t, html, "Full width")
}

func TestTableSimple_WithCustomClasses(t *testing.T) {
	table := SimpleTable(2, "custom-table-class").
		Field("A", "cell-a").
		Field("B", "cell-b")

	html := table.Render()

	assertContains(t, html, `custom-table-class`)
	assertContains(t, html, `cell-a`)
	assertContains(t, html, `cell-b`)
}

func TestTableSimple_FieldTextWithClass(t *testing.T) {
	table := SimpleTable(2, "").
		FieldText("Safe text", "p-2 bg-gray-100")

	html := table.Render()

	assertContains(t, html, "Safe text")
	assertContains(t, html, `p-2`)
	assertContains(t, html, `bg-gray-100`)
}

func TestTableSimple_EmptyTable(t *testing.T) {
	table := SimpleTable(2, "")
	html := table.Render()

	// Should render empty table
	assertContains(t, html, `<table`)
	assertContains(t, html, `<tbody`)
}
