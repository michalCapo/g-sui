package ui

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

type TTable[T any] struct {
	heads []string
	slots []struct {
		slot func(item *T) string
		cls  string
	}
	// data []*T
	cls string
	// rows []string
}

func Table[T any](cls string) *TTable[T] {
	return &TTable[T]{
		heads: []string{},
		slots: []struct {
			slot func(item *T) string
			cls  string
		}{},
		// data: data,
		cls: cls,
	}
}

func NewTable[T any](cls string) *TTable[T] {
	return &TTable[T]{}
}

func (t *TTable[T]) Head(value string, cls string) *TTable[T] {
	t.heads = append(t.heads, fmt.Sprintf(`<th class="%s">%s</th>`, escapeAttr(cls), html.EscapeString(value)))
	return t
}

// HeadHTML adds a table header with raw HTML content (use with caution)
func (t *TTable[T]) HeadHTML(value string, cls string) *TTable[T] {
	t.heads = append(t.heads, fmt.Sprintf(`<th class="%s">%s</th>`, escapeAttr(cls), value))
	return t
}

func (t *TTable[T]) Field(slot func(item *T) string, cls string) *TTable[T] {
	t.slots = append(t.slots, struct {
		slot func(item *T) string
		cls  string
	}{slot, cls})
	return t
}

// FieldText adds a table field with automatic HTML escaping for safe text content
func (t *TTable[T]) FieldText(slot func(item *T) string, cls string) *TTable[T] {
	safeSlot := func(item *T) string {
		return html.EscapeString(slot(item))
	}
	t.slots = append(t.slots, struct {
		slot func(item *T) string
		cls  string
	}{safeSlot, cls})
	return t
}

func (t *TTable[T]) Row(slot func(item *T) []string, cls string) *TTable[T] {
	return t
}

func (t *TTable[T]) Render(data []*T) string {
	var headsBuilder strings.Builder
	for _, head := range t.heads {
		headsBuilder.WriteString(head)
	}

	var rowsBuilder strings.Builder
	for _, row := range data {
		rowsBuilder.WriteString("<tr>")
		for _, slot := range t.slots {
			rowsBuilder.WriteString(fmt.Sprintf(`<td class="%s">%s</td>`, escapeAttr(slot.cls), slot.slot(row)))
		}
		rowsBuilder.WriteString("</tr>")
	}

	return fmt.Sprintf(
		`<div><table class="table-auto %s"><thead><tr>%s</tr></thead><tbody>%s</tbody></table></div>`,
		t.cls, headsBuilder.String(), rowsBuilder.String(),
	)
}

type TTableSimple struct {
	numCols    int
	cls        string
	rows       [][]string
	colClasses []string   // Store column classes
	cellAttrs  [][]string // Store cell attributes
}

func SimpleTable(numColumns int, cls ...string) *TTableSimple {
	return &TTableSimple{
		numCols:    numColumns,
		cls:        Classes(cls...),
		rows:       [][]string{},
		colClasses: make([]string, numColumns), // Initialize column classes slice
		cellAttrs:  [][]string{},               // Initialize cell attributes slice
	}
}

func (t *TTableSimple) Empty() *TTableSimple {
	t.Field("")
	return t
}

func (t *TTableSimple) Class(column int, classes ...string) *TTableSimple {
	if column >= 0 && column < t.numCols {
		t.colClasses[column] = Classes(classes...)
	}
	return t
}

func (t *TTableSimple) Field(value string, cls ...string) *TTableSimple {
	if len(t.rows) == 0 || len(t.rows[len(t.rows)-1]) == t.numCols {
		t.rows = append(t.rows, []string{})
		t.cellAttrs = append(t.cellAttrs, []string{})
	}

	cellClass := Classes(cls...)
	if cellClass != "" {
		value = fmt.Sprintf(`<div class="%s">%s</div>`, escapeAttr(cellClass), value)
	}

	t.rows[len(t.rows)-1] = append(t.rows[len(t.rows)-1], value)
	t.cellAttrs[len(t.cellAttrs)-1] = append(t.cellAttrs[len(t.cellAttrs)-1], "") // Initialize empty attributes for this cell

	return t
}

// FieldText adds a field with automatic HTML escaping for safe text content
func (t *TTableSimple) FieldText(value string, cls ...string) *TTableSimple {
	if len(t.rows) == 0 || len(t.rows[len(t.rows)-1]) == t.numCols {
		t.rows = append(t.rows, []string{})
		t.cellAttrs = append(t.cellAttrs, []string{})
	}

	// Escape the text content for safety
	escapedValue := html.EscapeString(value)
	
	cellClass := Classes(cls...)
	if cellClass != "" {
		escapedValue = fmt.Sprintf(`<div class="%s">%s</div>`, escapeAttr(cellClass), escapedValue)
	}

	t.rows[len(t.rows)-1] = append(t.rows[len(t.rows)-1], escapedValue)
	t.cellAttrs[len(t.cellAttrs)-1] = append(t.cellAttrs[len(t.cellAttrs)-1], "") // Initialize empty attributes for this cell

	return t
}

// func (t *TTableSimple) Field2(attrs string, value string, cls ...string) *TTableSimple {
// 	if len(t.rows) == 0 || len(t.rows[len(t.rows)-1]) == t.numCols {
// 		t.rows = append(t.rows, []string{})
// 	}

// 	t.rows[len(t.rows)-1] = append(t.rows[len(t.rows)-1], fmt.Sprintf(`<div class="%s"%s>%s</div>`, Classes(cls...), attrs, value))

// 	return t
// }

func (t *TTableSimple) Attr(attrs string) *TTableSimple {
	// Add attributes to the last defined cell
	if len(t.cellAttrs) > 0 && len(t.cellAttrs[len(t.cellAttrs)-1]) > 0 {
		lastRowIndex := len(t.cellAttrs) - 1
		lastCellIndex := len(t.cellAttrs[lastRowIndex]) - 1
		if t.cellAttrs[lastRowIndex][lastCellIndex] == "" {
			t.cellAttrs[lastRowIndex][lastCellIndex] = attrs
		} else {
			t.cellAttrs[lastRowIndex][lastCellIndex] += " " + attrs
		}
	}
	return t
}

func (t *TTableSimple) Render() string {
    var rowsBuilder strings.Builder
    colspanRegex := regexp.MustCompile(`colspan=['"]?(\d+)['"]?`)

    for rowIndex, row := range t.rows {
        rowsBuilder.WriteString("<tr>")
        usedCols := 0

        for i, cell := range row {
            class := ""
            if t.colClasses[i] != "" {
                class = fmt.Sprintf(` class="%s"`, t.colClasses[i])
            }

            // Add cell-specific attributes if they exist
            attrs := ""
            if rowIndex < len(t.cellAttrs) && i < len(t.cellAttrs[rowIndex]) && t.cellAttrs[rowIndex][i] != "" {
                attrs = " " + t.cellAttrs[rowIndex][i]
            }

            // Determine colspan for this cell before placing it
            nextColspan := 1
            if attrs != "" {
                matches := colspanRegex.FindStringSubmatch(attrs)
                if len(matches) > 1 {
                    if parsedColspan, err := strconv.Atoi(matches[1]); err == nil {
                        nextColspan = parsedColspan
                    }
                }
            }

            // If this cell would overflow the current row, wrap to a new row first
            if usedCols > 0 && usedCols+nextColspan > t.numCols {
                // pad the remaining cells if needed
                for i := usedCols; i < t.numCols; i++ {
                    padClass := ""
                    if i < len(t.colClasses) && t.colClasses[i] != "" {
                        padClass = fmt.Sprintf(` class="%s"`, t.colClasses[i])
                    }
                    rowsBuilder.WriteString(fmt.Sprintf(`<td%s></td>`, padClass))
                }
                rowsBuilder.WriteString("</tr><tr>")
                usedCols = 0
            }

            // Place the cell
            rowsBuilder.WriteString(fmt.Sprintf(`<td%s%s>%s</td>`, class, attrs, cell))
            usedCols += nextColspan
        }

        // Only add empty cells if we haven't reached the total number of columns
        for i := usedCols; i < t.numCols; i++ {
            class := ""
            if i < len(t.colClasses) && t.colClasses[i] != "" {
                class = fmt.Sprintf(` class="%s"`, t.colClasses[i])
            }
            rowsBuilder.WriteString(fmt.Sprintf(`<td%s></td>`, class))
        }
        rowsBuilder.WriteString("</tr>")
    }

	return fmt.Sprintf(
		`<table class="table-auto %s"><tbody>%s</tbody></table>`,
		t.cls, rowsBuilder.String(),
	)
}
