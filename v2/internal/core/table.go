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

package core

import (
	"github.com/SlideLang/go-docx/v2/domain"
	"github.com/SlideLang/go-docx/v2/internal/manager"
	"github.com/SlideLang/go-docx/v2/pkg/constants"
	"github.com/SlideLang/go-docx/v2/pkg/errors"
)

// table implements the domain.Table interface.
type table struct {
	id         string
	rows       []domain.TableRow
	cols       int
	width      domain.TableWidth
	alignment  domain.Alignment
	style      domain.TableStyle
	idGen      *manager.IDGenerator
	relManager *manager.RelationshipManager
}

// NewTable creates a new Table.
func NewTable(id string, rows, cols int, idGen *manager.IDGenerator, relManager *manager.RelationshipManager) domain.Table {
	t := &table{
		id:         id,
		rows:       make([]domain.TableRow, 0, rows),
		cols:       cols,
		width:      domain.TableWidth{Type: domain.WidthAuto, Value: 0},
		alignment:  domain.AlignmentLeft,
		style:      domain.TableStyle{},
		idGen:      idGen,
		relManager: relManager,
	}

	// Create initial rows
	for i := 0; i < rows; i++ {
		rowID := idGen.NextRowID()
		row := NewTableRow(rowID, cols, idGen, relManager)
		t.rows = append(t.rows, row)
	}

	return t
}

// Row returns the row at the specified index.
func (t *table) Row(index int) (domain.TableRow, error) {
	if index < 0 || index >= len(t.rows) {
		return nil, errors.InvalidArgument("Table.Row", "index", index,
			"row index out of bounds")
	}
	return t.rows[index], nil
}

// Rows returns all rows in the table.
func (t *table) Rows() []domain.TableRow {
	rows := make([]domain.TableRow, len(t.rows))
	copy(rows, t.rows)
	return rows
}

// AddRow adds a new row to the end of the table.
func (t *table) AddRow() (domain.TableRow, error) {
	rowID := t.idGen.NextRowID()
	row := NewTableRow(rowID, t.cols, t.idGen, t.relManager)
	t.rows = append(t.rows, row)
	return row, nil
}

// InsertRow inserts a new row at the specified index.
func (t *table) InsertRow(index int) (domain.TableRow, error) {
	if index < 0 || index > len(t.rows) {
		return nil, errors.InvalidArgument("Table.InsertRow", "index", index,
			"row index out of bounds")
	}

	rowID := t.idGen.NextRowID()
	row := NewTableRow(rowID, t.cols, t.idGen, t.relManager)

	// Insert at index
	t.rows = append(t.rows[:index], append([]domain.TableRow{row}, t.rows[index:]...)...)

	return row, nil
}

// DeleteRow deletes the row at the specified index.
func (t *table) DeleteRow(index int) error {
	if index < 0 || index >= len(t.rows) {
		return errors.InvalidArgument("Table.DeleteRow", "index", index,
			"row index out of bounds")
	}

	t.rows = append(t.rows[:index], t.rows[index+1:]...)
	return nil
}

// RowCount returns the number of rows in the table.
func (t *table) RowCount() int {
	return len(t.rows)
}

// ColumnCount returns the number of columns in the table.
func (t *table) ColumnCount() int {
	return t.cols
}

// Width returns the table width.
func (t *table) Width() domain.TableWidth {
	return t.width
}

// SetWidth sets the table width.
func (t *table) SetWidth(width domain.TableWidth) error {
	if width.Type < domain.WidthAuto || width.Type > domain.WidthPct {
		return errors.InvalidArgument("Table.SetWidth", "width.Type", width.Type,
			"invalid width type")
	}
	t.width = width
	return nil
}

// Alignment returns the table alignment.
func (t *table) Alignment() domain.Alignment {
	return t.alignment
}

// SetAlignment sets the table alignment.
func (t *table) SetAlignment(align domain.Alignment) error {
	if align < domain.AlignmentLeft || align > domain.AlignmentDistribute {
		return errors.InvalidArgument("Table.SetAlignment", "align", align,
			"invalid alignment value")
	}
	t.alignment = align
	return nil
}

// Style returns the table style.
func (t *table) Style() domain.TableStyle {
	return t.style
}

// SetStyle sets the table style.
func (t *table) SetStyle(style domain.TableStyle) error {
	t.style = style
	return nil
}

// tableRow implements the domain.TableRow interface.
type tableRow struct {
	id         string
	cells      []domain.TableCell
	height     int
	idGen      *manager.IDGenerator
	relManager *manager.RelationshipManager
}

// NewTableRow creates a new TableRow.
func NewTableRow(id string, cols int, idGen *manager.IDGenerator, relManager *manager.RelationshipManager) domain.TableRow {
	row := &tableRow{
		id:         id,
		cells:      make([]domain.TableCell, 0, cols),
		height:     0, // Auto height
		idGen:      idGen,
		relManager: relManager,
	}

	// Create cells
	for i := 0; i < cols; i++ {
		cellID := idGen.NextCellID()
		cell := NewTableCell(cellID, idGen, relManager)
		row.cells = append(row.cells, cell)
	}

	return row
}

// Cell returns the cell at the specified column index.
func (r *tableRow) Cell(col int) (domain.TableCell, error) {
	if col < 0 || col >= len(r.cells) {
		return nil, errors.InvalidArgument("TableRow.Cell", "col", col,
			"column index out of bounds")
	}
	return r.cells[col], nil
}

// Cells returns all cells in this row.
func (r *tableRow) Cells() []domain.TableCell {
	cells := make([]domain.TableCell, len(r.cells))
	copy(cells, r.cells)
	return cells
}

// Height returns the row height.
func (r *tableRow) Height() int {
	return r.height
}

// SetHeight sets the row height in twips.
func (r *tableRow) SetHeight(twips int) error {
	if twips < 0 {
		return errors.InvalidArgument("TableRow.SetHeight", "twips", twips,
			"height cannot be negative")
	}
	r.height = twips
	return nil
}

// tableCell implements the domain.TableCell interface.
type tableCell struct {
	id                string
	paragraphs        []domain.Paragraph
	width             int
	verticalAlignment domain.VerticalAlignment
	borders           domain.TableBorders
	shading           domain.Color
	idGen             IDGenerator
	relManager        *manager.RelationshipManager
}

// NewTableCell creates a new TableCell.
func NewTableCell(id string, idGen IDGenerator, relManager *manager.RelationshipManager) domain.TableCell {
	return &tableCell{
		id:                id,
		paragraphs:        make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		width:             0, // Auto width
		verticalAlignment: domain.VerticalAlignTop,
		borders:           domain.TableBorders{},
		shading:           domain.ColorWhite,
		idGen:             idGen,
		relManager:        relManager,
	}
}

// AddParagraph adds a paragraph to this cell.
func (c *tableCell) AddParagraph() (domain.Paragraph, error) {
	id := c.idGen.NextParagraphID()
	para := NewParagraph(id, c.idGen, c.relManager)
	c.paragraphs = append(c.paragraphs, para)
	return para, nil
}

// Paragraphs returns all paragraphs in this cell.
func (c *tableCell) Paragraphs() []domain.Paragraph {
	paras := make([]domain.Paragraph, len(c.paragraphs))
	copy(paras, c.paragraphs)
	return paras
}

// Width returns the cell width.
func (c *tableCell) Width() int {
	return c.width
}

// SetWidth sets the cell width in twips.
func (c *tableCell) SetWidth(twips int) error {
	if twips < 0 {
		return errors.InvalidArgument("TableCell.SetWidth", "twips", twips,
			"width cannot be negative")
	}
	c.width = twips
	return nil
}

// VerticalAlignment returns the vertical alignment of content.
func (c *tableCell) VerticalAlignment() domain.VerticalAlignment {
	return c.verticalAlignment
}

// SetVerticalAlignment sets the vertical alignment.
func (c *tableCell) SetVerticalAlignment(align domain.VerticalAlignment) error {
	if align < domain.VerticalAlignTop || align > domain.VerticalAlignBottom {
		return errors.InvalidArgument("TableCell.SetVerticalAlignment", "align", align,
			"invalid vertical alignment value")
	}
	c.verticalAlignment = align
	return nil
}

// Borders returns the cell borders.
func (c *tableCell) Borders() domain.TableBorders {
	return c.borders
}

// SetBorders sets the cell borders.
func (c *tableCell) SetBorders(borders domain.TableBorders) error {
	c.borders = borders
	return nil
}

// Shading returns the cell background color.
func (c *tableCell) Shading() domain.Color {
	return c.shading
}

// SetShading sets the cell background color.
func (c *tableCell) SetShading(color domain.Color) error {
	c.shading = color
	return nil
}

// Merge merges this cell with adjacent cells.
func (c *tableCell) Merge(cols, rows int) error {
	// TODO: Implement cell merging
	return errors.Unsupported("TableCell.Merge", "cell merging not yet implemented")
}
