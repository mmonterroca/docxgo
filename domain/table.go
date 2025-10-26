/*
MIT License

Copyright (c) 2025 Misael Montero
Copyright (c) 2020-2023 fumiama (original go-docx)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/



package domain

// Table represents a table in a document.
type Table interface {
	// Row returns the row at the specified index.
	Row(index int) (TableRow, error)

	// Rows returns all rows in the table.
	Rows() []TableRow

	// AddRow adds a new row to the end of the table.
	AddRow() (TableRow, error)

	// InsertRow inserts a new row at the specified index.
	InsertRow(index int) (TableRow, error)

	// DeleteRow deletes the row at the specified index.
	DeleteRow(index int) error

	// RowCount returns the number of rows in the table.
	RowCount() int

	// ColumnCount returns the number of columns in the table.
	ColumnCount() int

	// Width returns the table width.
	Width() TableWidth

	// SetWidth sets the table width.
	SetWidth(width TableWidth) error

	// Alignment returns the table alignment.
	Alignment() Alignment

	// SetAlignment sets the table alignment.
	SetAlignment(align Alignment) error

	// Style returns the table style.
	Style() TableStyle

	// SetStyle sets the table style.
	SetStyle(style TableStyle) error
}

// TableRow represents a row in a table.
type TableRow interface {
	// Cell returns the cell at the specified column index.
	Cell(col int) (TableCell, error)

	// Cells returns all cells in this row.
	Cells() []TableCell

	// Height returns the row height.
	Height() int

	// SetHeight sets the row height in twips.
	SetHeight(twips int) error
}

// TableCell represents a cell in a table.
type TableCell interface {
	// AddParagraph adds a paragraph to this cell.
	AddParagraph() (Paragraph, error)

	// Paragraphs returns all paragraphs in this cell.
	Paragraphs() []Paragraph

	// Width returns the cell width.
	Width() int

	// SetWidth sets the cell width in twips.
	SetWidth(twips int) error

	// VerticalAlignment returns the vertical alignment of content.
	VerticalAlignment() VerticalAlignment

	// SetVerticalAlignment sets the vertical alignment.
	SetVerticalAlignment(align VerticalAlignment) error

	// Borders returns the cell borders.
	Borders() TableBorders

	// SetBorders sets the cell borders.
	SetBorders(borders TableBorders) error

	// Shading returns the cell background color.
	Shading() Color

	// SetShading sets the cell background color.
	SetShading(color Color) error

	// Merge merges this cell with adjacent cells.
	// cols and rows specify how many cells to merge in each direction.
	Merge(cols, rows int) error

	// GridSpan returns the number of columns this cell spans.
	GridSpan() int

	// SetGridSpan sets the number of columns to span (horizontal merge).
	SetGridSpan(span int) error

	// VMerge returns the vertical merge status.
	VMerge() VerticalMergeType

	// SetVMerge sets the vertical merge type.
	SetVMerge(mergeType VerticalMergeType) error

	// AddTable adds a nested table to this cell.
	AddTable(rows, cols int) (Table, error)

	// Tables returns all nested tables in this cell.
	Tables() []Table
}

// TableWidth represents table width settings.
type TableWidth struct {
	Type  WidthType
	Value int
}

// WidthType defines how width is calculated.
type WidthType int

const (
	WidthAuto WidthType = iota // Auto width
	WidthDXA                   // Fixed width in twips
	WidthPct                   // Percentage (value = percentage * 50)
)

// VerticalAlignment represents vertical alignment in a cell.
type VerticalAlignment int

const (
	VerticalAlignTop VerticalAlignment = iota
	VerticalAlignCenter
	VerticalAlignBottom
)

// TableBorders represents borders for a table or cell.
type TableBorders struct {
	Top    BorderStyle
	Left   BorderStyle
	Bottom BorderStyle
	Right  BorderStyle
}

// BorderStyle represents a border's appearance.
type BorderStyle struct {
	Style BorderLineStyle
	Width int // Width in eighths of a point
	Color Color
}

// BorderLineStyle represents border line styles.
type BorderLineStyle int

const (
	BorderNone BorderLineStyle = iota
	BorderSingle
	BorderDotted
	BorderDashed
	BorderDouble
	BorderTriple
	BorderThick
)

// TableStyle represents table styling options.
type TableStyle struct {
	Name string
	// Could be expanded with more style properties
}

// VerticalMergeType represents vertical merge status for table cells.
type VerticalMergeType int

const (
	VMergeNone    VerticalMergeType = iota // No vertical merge
	VMergeRestart                           // Start of vertical merge
	VMergeContinue                          // Continuation of vertical merge
)

// CellMergeInfo represents cell merge information.
type CellMergeInfo struct {
	GridSpan int               // Horizontal span (number of columns)
	VMerge   VerticalMergeType // Vertical merge type
	RowSpan  int               // Vertical span (number of rows) - calculated
	ColSpan  int               // Horizontal span (number of columns) - same as GridSpan
}
