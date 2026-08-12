// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// table implements the domain.Table interface.
type table struct {
	id           string
	rows         []domain.TableRow
	cols         int
	width        domain.TableWidth
	alignment    domain.Alignment
	style        domain.TableStyle
	borders      domain.TableLevelBorders
	idGen        *manager.IDGenerator
	relManager   *manager.RelationshipManager
	mediaManager *manager.MediaManager
}

// NewTable creates a new Table.
func NewTable(id string, rows, cols int, idGen *manager.IDGenerator, relManager *manager.RelationshipManager, mediaManager *manager.MediaManager) domain.Table {
	t := &table{
		id:           id,
		rows:         make([]domain.TableRow, 0, rows),
		cols:         cols,
		width:        domain.TableWidth{Type: domain.WidthAuto, Value: 0},
		alignment:    domain.AlignmentLeft,
		style:        domain.TableStyle{},
		idGen:        idGen,
		relManager:   relManager,
		mediaManager: mediaManager,
	}

	// Create initial rows
	for i := 0; i < rows; i++ {
		rowID := idGen.NextRowID()
		row := NewTableRow(t, rowID, cols, idGen, relManager, mediaManager)
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
	row := NewTableRow(t, rowID, t.cols, t.idGen, t.relManager, t.mediaManager)
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
	row := NewTableRow(t, rowID, t.cols, t.idGen, t.relManager, t.mediaManager)

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

// Borders returns the table-level borders.
func (t *table) Borders() domain.TableLevelBorders {
	return t.borders
}

// SetBorders sets the table-level borders.
func (t *table) SetBorders(borders domain.TableLevelBorders) error {
	t.borders = borders
	return nil
}

// tableRow implements the domain.TableRow interface.
type tableRow struct {
	id           string
	cells        []domain.TableCell
	height       int
	table        *table
	idGen        *manager.IDGenerator
	relManager   *manager.RelationshipManager
	mediaManager *manager.MediaManager
}

// NewTableRow creates a new TableRow.
func NewTableRow(tbl *table, id string, cols int, idGen *manager.IDGenerator, relManager *manager.RelationshipManager, mediaManager *manager.MediaManager) domain.TableRow {
	row := &tableRow{
		id:           id,
		cells:        make([]domain.TableCell, 0, cols),
		height:       0, // Auto height
		table:        tbl,
		idGen:        idGen,
		relManager:   relManager,
		mediaManager: mediaManager,
	}

	// Create cells
	for i := 0; i < cols; i++ {
		cellID := idGen.NextCellID()
		cell := NewTableCell(row, cellID, idGen, relManager, mediaManager)
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
	tables            []domain.Table
	width             int
	verticalAlignment domain.VerticalAlignment
	borders           domain.TableBorders
	shading           domain.Color
	// shadingSet mirrors capsSetter (internal/serializer): whether SetShading
	// was ever called, so the serializer can tell "never touched" from
	// "explicitly shaded white" -- the only way to represent a cell shaded
	// pure white on purpose, distinct from the auto-white every new cell
	// starts with.
	shadingSet bool
	// themeFill/themeFillTint/themeFillShade cache a theme reference
	// hydrated for this cell's shading -- domain.Color is a bare, comparable
	// {R,G,B} struct compared with != as a sentinel throughout this package
	// and the serializer, so it cannot itself carry a theme link without
	// becoming a breaking change to public API. Always the "fill" slot
	// regardless of whether the source used w:val="solid"+w:themeColor or
	// w:val="clear"+w:themeFill -- see HydrateThemeFill. SetShading clears
	// these: an explicit caller colour means the cell no longer follows the
	// theme.
	themeFill      string
	themeFillTint  string
	themeFillShade string
	gridSpan       int
	vMerge         domain.VerticalMergeType
	row            *tableRow
	hMergeParent   *tableCell
	idGen          *manager.IDGenerator
	relManager     *manager.RelationshipManager
	mediaManager   *manager.MediaManager
}

// NewTableCell creates a new TableCell.
func NewTableCell(row *tableRow, id string, idGen *manager.IDGenerator, relManager *manager.RelationshipManager, mediaManager *manager.MediaManager) domain.TableCell {
	return &tableCell{
		id:                id,
		paragraphs:        make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		tables:            make([]domain.Table, 0, 1),
		width:             0, // Auto width
		verticalAlignment: domain.VerticalAlignTop,
		borders:           domain.TableBorders{},
		shading:           domain.ColorWhite,
		gridSpan:          1, // Default: no horizontal merge
		vMerge:            domain.VMergeNone,
		row:               row,
		idGen:             idGen,
		relManager:        relManager,
		mediaManager:      mediaManager,
	}
}

// AddParagraph adds a paragraph to this cell.
func (c *tableCell) AddParagraph() (domain.Paragraph, error) {
	id := c.idGen.NextParagraphID()
	para := NewParagraph(id, c.idGen, c.relManager, c.mediaManager)
	c.paragraphs = append(c.paragraphs, para)
	return para, nil
}

// RemoveParagraph removes the paragraph at the given index.
func (c *tableCell) RemoveParagraph(index int) error {
	if index < 0 || index >= len(c.paragraphs) {
		return errors.InvalidArgument("TableCell.RemoveParagraph", "index", index,
			"paragraph index out of bounds")
	}

	c.paragraphs = append(c.paragraphs[:index], c.paragraphs[index+1:]...)
	return nil
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

// SetShading sets the cell background color, clearing any theme link a
// hydrated shading may have carried -- an explicit colour from the caller
// means the cell no longer follows the document's theme. See
// HydrateThemeFill.
func (c *tableCell) SetShading(color domain.Color) error {
	c.shading = color
	c.shadingSet = true
	c.themeFill = ""
	c.themeFillTint = ""
	c.themeFillShade = ""
	return nil
}

// ShadingSet reports whether SetShading was ever called on this cell. This
// is an internal method used by the serializer to distinguish a cell
// explicitly shaded white from one that was simply never touched (both
// otherwise report Shading() == domain.ColorWhite).
func (c *tableCell) ShadingSet() bool {
	return c.shadingSet
}

// ThemeFill returns the theme reference cached for this cell's shading --
// hydrated from a source w:themeFill (or, for a w:val="solid" pattern,
// w:themeColor, normalized onto this same slot) -- and its tint/shade, or
// three empty strings if none was hydrated. This is an internal method used
// by the serializer.
func (c *tableCell) ThemeFill() (fill, tint, shade string) {
	return c.themeFill, c.themeFillTint, c.themeFillShade
}

// HydrateThemeFill records the theme reference a source document's cell
// shading resolved from, alongside the concrete colour SetShading also
// received for it (see applyCellShading). Only the reader should call this
// -- SetShading clears it, since a caller-supplied colour has nothing to do
// with the source's theme. This is an internal method used by the reader.
func (c *tableCell) HydrateThemeFill(fill, tint, shade string) {
	c.themeFill = fill
	c.themeFillTint = tint
	c.themeFillShade = shade
}

// Merge merges this cell with adjacent cells.
func (c *tableCell) Merge(cols, rows int) error {
	const op = "TableCell.Merge"

	if cols < 1 {
		return errors.InvalidArgument(op, "cols", cols,
			"cols must be at least 1")
	}
	if rows < 1 {
		return errors.InvalidArgument(op, "rows", rows,
			"rows must be at least 1")
	}

	if c.row == nil || c.row.table == nil {
		return errors.InvalidState(op, "cell is not attached to a table")
	}

	if c.hMergeParent != nil {
		return errors.InvalidState(op, "cannot merge from a horizontally merged cell")
	}
	if c.vMerge == domain.VMergeContinue {
		return errors.InvalidState(op, "cannot merge from a vertically continued cell")
	}

	row := c.row
	colIndex := -1
	for idx, candidate := range row.cells {
		if tc, ok := candidate.(*tableCell); ok && tc == c {
			colIndex = idx
			break
		}
	}
	if colIndex == -1 {
		return errors.InvalidState(op, "cell not found in parent row")
	}

	if colIndex+cols > len(row.cells) {
		return errors.InvalidArgument(op, "cols", cols,
			"merge exceeds row column count")
	}

	tbl := row.table
	rowIndex := -1
	for idx, candidate := range tbl.rows {
		if tr, ok := candidate.(*tableRow); ok && tr == row {
			rowIndex = idx
			break
		}
	}
	if rowIndex == -1 {
		return errors.InvalidState(op, "parent row not found in table")
	}

	if rowIndex+rows > len(tbl.rows) {
		return errors.InvalidArgument(op, "rows", rows,
			"merge exceeds available rows")
	}

	// Validate that the merge region is free of existing merges
	for rOffset := 0; rOffset < rows; rOffset++ {
		targetRow, ok := tbl.rows[rowIndex+rOffset].(*tableRow)
		if !ok {
			return errors.InvalidState(op, "unexpected row implementation type")
		}

		for cOffset := 0; cOffset < cols; cOffset++ {
			targetCell, ok := targetRow.cells[colIndex+cOffset].(*tableCell)
			if !ok {
				return errors.InvalidState(op, "unexpected cell implementation type")
			}

			if targetCell == c {
				continue
			}

			if targetCell.hMergeParent != nil {
				return errors.InvalidState(op, "merge region overlaps an existing horizontal merge")
			}
			if targetCell.vMerge != domain.VMergeNone {
				return errors.InvalidState(op, "merge region overlaps an existing vertical merge")
			}
			if targetCell.gridSpan > 1 {
				return errors.InvalidState(op, "merge region overlaps an existing grid span")
			}
		}
	}

	// Configure the primary cell
	c.hMergeParent = nil
	c.gridSpan = cols
	if rows > 1 {
		c.vMerge = domain.VMergeRestart
	} else {
		c.vMerge = domain.VMergeNone
	}

	// Apply horizontal merge within the current row
	for offset := 1; offset < cols; offset++ {
		sibling := row.cells[colIndex+offset].(*tableCell)
		sibling.hMergeParent = c
		sibling.gridSpan = 1
		sibling.vMerge = domain.VMergeNone
	}

	// Apply vertical merges across subsequent rows
	if rows > 1 {
		for rOffset := 1; rOffset < rows; rOffset++ {
			targetRow := tbl.rows[rowIndex+rOffset].(*tableRow)
			leading := targetRow.cells[colIndex].(*tableCell)
			leading.hMergeParent = nil
			leading.gridSpan = cols
			leading.vMerge = domain.VMergeContinue

			for cOffset := 1; cOffset < cols; cOffset++ {
				neighbor := targetRow.cells[colIndex+cOffset].(*tableCell)
				neighbor.hMergeParent = leading
				neighbor.gridSpan = 1
				neighbor.vMerge = domain.VMergeContinue
			}
		}
	}

	return nil
}

// GridSpan returns the number of grid columns spanned by this cell.
func (c *tableCell) GridSpan() int {
	return c.gridSpan
}

// SetGridSpan sets the horizontal merge span.
func (c *tableCell) SetGridSpan(span int) error {
	if span < 1 {
		return errors.InvalidArgument("TableCell.SetGridSpan", "span", span,
			"span must be at least 1")
	}
	c.gridSpan = span
	return nil
}

// VMerge returns the vertical merge type.
func (c *tableCell) VMerge() domain.VerticalMergeType {
	return c.vMerge
}

// SetVMerge sets the vertical merge type.
func (c *tableCell) SetVMerge(mergeType domain.VerticalMergeType) error {
	if mergeType < domain.VMergeNone || mergeType > domain.VMergeContinue {
		return errors.InvalidArgument("TableCell.SetVMerge", "mergeType", mergeType,
			"invalid vertical merge type")
	}
	c.vMerge = mergeType
	return nil
}

// AddTable adds a nested table to this cell.
func (c *tableCell) AddTable(rows, cols int) (domain.Table, error) {
	if rows < 1 {
		return nil, errors.InvalidArgument("TableCell.AddTable", "rows", rows,
			"rows must be at least 1")
	}
	if cols < 1 {
		return nil, errors.InvalidArgument("TableCell.AddTable", "cols", cols,
			"cols must be at least 1")
	}

	table := NewTable(c.idGen.GenerateID("table"), rows, cols, c.idGen, c.relManager, c.mediaManager)
	c.tables = append(c.tables, table)
	return table, nil
}

// Tables returns all nested tables in this cell.
func (c *tableCell) Tables() []domain.Table {
	// Return a defensive copy
	result := make([]domain.Table, len(c.tables))
	copy(result, c.tables)
	return result
}

// IsHorizontallyMergedContinuation reports whether this cell is hidden by a
// horizontal merge originating from another cell in the same row.
func (c *tableCell) IsHorizontallyMergedContinuation() bool {
	return c.hMergeParent != nil
}
