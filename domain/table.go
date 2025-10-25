/*
   Copyright (c) 2025 SlideLang

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
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
