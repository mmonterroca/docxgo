/*
MIT License

Copyright (c) 2025 Misael Monterroca
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

package docx

import (
	"fmt"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// DocumentBuilder provides a fluent API for building Word documents.
// It accumulates errors during construction and surfaces them in Build().
//
// Example:
//
//	builder := docx.NewDocumentBuilder()
//	builder.AddParagraph().
//	    Text("Hello, World!").
//	    Bold().
//	    FontSize(14).
//	    End()
//
//	doc, err := builder.Build()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	doc.SaveAs("output.docx")
type DocumentBuilder struct {
	doc    domain.Document
	errors []error
}

// SectionBuilder provides a fluent API for configuring document sections.
type SectionBuilder struct {
	section domain.Section
	parent  *DocumentBuilder
	err     error
}

func (sb *SectionBuilder) recordError(err error) {
	if err == nil {
		return
	}
	if sb.err == nil {
		sb.err = err
	}
	if sb.parent != nil {
		sb.parent.errors = append(sb.parent.errors, err)
	}
}

func (sb *SectionBuilder) ensureSection(op string) bool {
	if sb == nil {
		return false
	}
	if sb.err != nil {
		return false
	}
	if sb.section == nil {
		sb.recordError(errors.InvalidState(op, "section is nil"))
		return false
	}
	return true
}

// PageSize sets the page size for the section (e.g., PageSizeA4, PageSizeLetter).
func (sb *SectionBuilder) PageSize(size domain.PageSize) *SectionBuilder {
	if !sb.ensureSection("SectionBuilder.PageSize") {
		return sb
	}

	if err := sb.section.SetPageSize(size); err != nil {
		sb.recordError(err)
	}
	return sb
}

// Orientation sets the page orientation for the section.
func (sb *SectionBuilder) Orientation(orient domain.Orientation) *SectionBuilder {
	if !sb.ensureSection("SectionBuilder.Orientation") {
		return sb
	}
	if err := sb.section.SetOrientation(orient); err != nil {
		sb.recordError(err)
	}
	return sb
}

// Margins sets the page margins for the section.
func (sb *SectionBuilder) Margins(margins domain.Margins) *SectionBuilder {
	if !sb.ensureSection("SectionBuilder.Margins") {
		return sb
	}
	if err := sb.section.SetMargins(margins); err != nil {
		sb.recordError(err)
	}
	return sb
}

// Columns sets the column layout for the section.
func (sb *SectionBuilder) Columns(count int) *SectionBuilder {
	if !sb.ensureSection("SectionBuilder.Columns") {
		return sb
	}
	if err := sb.section.SetColumns(count); err != nil {
		sb.recordError(err)
	}
	return sb
}

// Header returns the requested header for direct manipulation.
func (sb *SectionBuilder) Header(headerType domain.HeaderType) (domain.Header, error) {
	if sb == nil {
		return nil, errors.InvalidState("SectionBuilder.Header", "section is nil")
	}
	if !sb.ensureSection("SectionBuilder.Header") {
		return nil, sb.err
	}
	head, err := sb.section.Header(headerType)
	if err != nil {
		sb.recordError(err)
		return nil, err
	}
	return head, nil
}

// Footer returns the requested footer for direct manipulation.
func (sb *SectionBuilder) Footer(footerType domain.FooterType) (domain.Footer, error) {
	if sb == nil {
		return nil, errors.InvalidState("SectionBuilder.Footer", "section is nil")
	}
	if !sb.ensureSection("SectionBuilder.Footer") {
		return nil, sb.err
	}
	foot, err := sb.section.Footer(footerType)
	if err != nil {
		sb.recordError(err)
		return nil, err
	}
	return foot, nil
}

// Section exposes the underlying domain.Section for advanced scenarios.
func (sb *SectionBuilder) Section() domain.Section {
	if sb == nil {
		return nil
	}
	return sb.section
}

// End returns control to the DocumentBuilder.
func (sb *SectionBuilder) End() *DocumentBuilder {
	if sb == nil {
		return nil
	}
	return sb.parent
}

// NewDocumentBuilder creates a new document builder with optional configuration.
//
// Example:
//
//	builder := docx.NewDocumentBuilder(
//	    docx.WithDefaultFont("Arial"),
//	    docx.WithPageSize(docx.A4),
//	)
func NewDocumentBuilder(opts ...Option) *DocumentBuilder {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Create document with configuration
	doc := NewDocument()

	builder := &DocumentBuilder{
		doc:    doc,
		errors: make([]error, 0),
	}

	// Apply configuration to document
	if config.Metadata != nil {
		if err := doc.SetMetadata(config.Metadata); err != nil {
			// Note: This error is intentionally ignored during builder initialization
			// as metadata errors are non-critical for document creation
			_ = err
		}
	}

	// Apply theme if provided
	if config.Theme != nil {
		// Use type assertion to get Theme interface
		// This avoids import cycle between docx and themes packages
		if theme, ok := config.Theme.(interface {
			ApplyTo(domain.Document) error
		}); ok {
			if err := theme.ApplyTo(doc); err != nil {
				builder.errors = append(builder.errors, err)
			}
		}
	}

	return builder
}

// AddParagraph adds a new paragraph to the document and returns a ParagraphBuilder.
// Errors are accumulated and returned in Build().
//
// Example:
//
//	builder.AddParagraph().
//	    Text("This is a paragraph.").
//	    Alignment(docx.AlignmentCenter).
//	    End()
func (b *DocumentBuilder) AddParagraph() *ParagraphBuilder {
	para, err := b.doc.AddParagraph()
	if err != nil {
		b.errors = append(b.errors, err)
		return &ParagraphBuilder{
			parent: b,
			err:    err,
		}
	}

	return &ParagraphBuilder{
		para:   para,
		parent: b,
	}
}

// AddTable adds a new table with the specified dimensions and returns a TableBuilder.
//
// Example:
//
//	builder.AddTable(3, 3).
//	    Row(0).Cell(0).Text("Header 1").Bold().End().
//	    Row(0).Cell(1).Text("Header 2").Bold().End().
//	    End()
func (b *DocumentBuilder) AddTable(rows, cols int) *TableBuilder {
	table, err := b.doc.AddTable(rows, cols)
	if err != nil {
		b.errors = append(b.errors, err)
		return &TableBuilder{
			parent: b,
			err:    err,
		}
	}

	return &TableBuilder{
		table:  table,
		parent: b,
	}
}

// DefaultSection returns a SectionBuilder for configuring the document's default section.
// Errors are accumulated and surfaced during Build().
func (b *DocumentBuilder) DefaultSection() *SectionBuilder {
	section, err := b.doc.DefaultSection()
	sb := &SectionBuilder{
		section: section,
		parent:  b,
	}
	if err != nil {
		sb.recordError(err)
	}
	return sb
}

// AddSection inserts a new section with the specified break type (default NextPage).
// Returns a SectionBuilder for configuring the new section.
func (b *DocumentBuilder) AddSection(breakType ...domain.SectionBreakType) *SectionBuilder {
	bt := domain.SectionBreakTypeNextPage
	if len(breakType) > 0 {
		bt = breakType[0]
	}

	section, err := b.doc.AddSectionWithBreak(bt)
	sb := &SectionBuilder{
		section: section,
		parent:  b,
	}
	if err != nil {
		sb.recordError(err)
	}
	return sb
}

// SetMetadata sets the document metadata.
func (b *DocumentBuilder) SetMetadata(meta *domain.Metadata) *DocumentBuilder {
	if err := b.doc.SetMetadata(meta); err != nil {
		b.errors = append(b.errors, err)
	}
	return b
}

// Build validates the document and returns it.
// All accumulated errors are returned here.
func (b *DocumentBuilder) Build() (domain.Document, error) {
	// Return first error if any accumulated
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("document build failed with %d error(s): %w", len(b.errors), b.errors[0])
	}

	// Validate document structure
	if err := b.doc.Validate(); err != nil {
		return nil, errors.Wrap(err, "DocumentBuilder.Build")
	}

	return b.doc, nil
}

// ParagraphBuilder provides a fluent API for building paragraphs.
type ParagraphBuilder struct {
	para   domain.Paragraph
	parent *DocumentBuilder
	err    error
}

// Text adds text to the paragraph.
// If multiple Text() calls are made, each creates a new run.
//
// Example:
//
//	para.Text("Hello ").Text("World")  // Creates 2 runs
func (pb *ParagraphBuilder) Text(text string) *ParagraphBuilder {
	if pb.err != nil {
		return pb // Propagate error
	}

	run, err := pb.para.AddRun()
	if err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
		return pb
	}

	if err := run.SetText(text); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// Bold makes the last run bold.
func (pb *ParagraphBuilder) Bold() *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	runs := pb.para.Runs()
	if len(runs) == 0 {
		pb.err = errors.InvalidState("ParagraphBuilder.Bold", "no runs to make bold")
		pb.parent.errors = append(pb.parent.errors, pb.err)
		return pb
	}

	if err := runs[len(runs)-1].SetBold(true); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// Italic makes the last run italic.
func (pb *ParagraphBuilder) Italic() *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	runs := pb.para.Runs()
	if len(runs) == 0 {
		pb.err = errors.InvalidState("ParagraphBuilder.Italic", "no runs to make italic")
		pb.parent.errors = append(pb.parent.errors, pb.err)
		return pb
	}

	if err := runs[len(runs)-1].SetItalic(true); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// Color sets the color of the last run.
func (pb *ParagraphBuilder) Color(color domain.Color) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	runs := pb.para.Runs()
	if len(runs) == 0 {
		pb.err = errors.InvalidState("ParagraphBuilder.Color", "no runs to colorize")
		pb.parent.errors = append(pb.parent.errors, pb.err)
		return pb
	}

	if err := runs[len(runs)-1].SetColor(color); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// FontSize sets the font size of the last run in points.
func (pb *ParagraphBuilder) FontSize(points int) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	runs := pb.para.Runs()
	if len(runs) == 0 {
		pb.err = errors.InvalidState("ParagraphBuilder.FontSize", "no runs to set font size")
		pb.parent.errors = append(pb.parent.errors, pb.err)
		return pb
	}

	// Convert points to half-points
	halfPoints := points * 2
	if err := runs[len(runs)-1].SetSize(halfPoints); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// Alignment sets the paragraph alignment.
func (pb *ParagraphBuilder) Alignment(align domain.Alignment) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	if err := pb.para.SetAlignment(align); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// Underline sets the underline style of the last run.
func (pb *ParagraphBuilder) Underline(style domain.UnderlineStyle) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	runs := pb.para.Runs()
	if len(runs) == 0 {
		pb.err = errors.InvalidState("ParagraphBuilder.Underline", "no runs to underline")
		pb.parent.errors = append(pb.parent.errors, pb.err)
		return pb
	}

	if err := runs[len(runs)-1].SetUnderline(style); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// AddImage adds an image from a file path to the paragraph.
func (pb *ParagraphBuilder) AddImage(path string) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	if _, err := pb.para.AddImage(path); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// AddImageWithSize adds an image with custom dimensions.
func (pb *ParagraphBuilder) AddImageWithSize(path string, size domain.ImageSize) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	if _, err := pb.para.AddImageWithSize(path, size); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// AddImageWithPosition adds a floating image with custom positioning.
func (pb *ParagraphBuilder) AddImageWithPosition(path string, size domain.ImageSize, pos domain.ImagePosition) *ParagraphBuilder {
	if pb.err != nil {
		return pb
	}

	if _, err := pb.para.AddImageWithPosition(path, size, pos); err != nil {
		pb.err = err
		pb.parent.errors = append(pb.parent.errors, err)
	}

	return pb
}

// End returns to the DocumentBuilder for further operations.
func (pb *ParagraphBuilder) End() *DocumentBuilder {
	return pb.parent
}

// TableBuilder provides a fluent API for building tables.
type TableBuilder struct {
	table  domain.Table
	parent *DocumentBuilder
	err    error
}

// Row returns a RowBuilder for the specified row index.
func (tb *TableBuilder) Row(index int) *RowBuilder {
	if tb.err != nil {
		return &RowBuilder{parent: tb, err: tb.err}
	}

	row, err := tb.table.Row(index)
	if err != nil {
		tb.err = err
		tb.parent.errors = append(tb.parent.errors, err)
		return &RowBuilder{parent: tb, err: err}
	}

	return &RowBuilder{
		row:    row,
		parent: tb,
	}
}

// Width sets the table width.
func (tb *TableBuilder) Width(widthType domain.WidthType, value int) *TableBuilder {
	if tb.err != nil {
		return tb
	}

	if err := tb.table.SetWidth(domain.TableWidth{Type: widthType, Value: value}); err != nil {
		tb.err = err
		tb.parent.errors = append(tb.parent.errors, err)
	}

	return tb
}

// Alignment sets the table alignment.
func (tb *TableBuilder) Alignment(align domain.Alignment) *TableBuilder {
	if tb.err != nil {
		return tb
	}

	if err := tb.table.SetAlignment(align); err != nil {
		tb.err = err
		tb.parent.errors = append(tb.parent.errors, err)
	}

	return tb
}

// Style sets the table style.
func (tb *TableBuilder) Style(style domain.TableStyle) *TableBuilder {
	if tb.err != nil {
		return tb
	}

	if err := tb.table.SetStyle(style); err != nil {
		tb.err = err
		tb.parent.errors = append(tb.parent.errors, err)
	}

	return tb
}

// End returns to the DocumentBuilder.
func (tb *TableBuilder) End() *DocumentBuilder {
	return tb.parent
}

// RowBuilder provides a fluent API for building table rows.
type RowBuilder struct {
	row    domain.TableRow
	parent *TableBuilder
	err    error
}

// Cell returns a CellBuilder for the specified cell index.
func (rb *RowBuilder) Cell(index int) *CellBuilder {
	if rb.err != nil {
		return &CellBuilder{parent: rb, err: rb.err}
	}

	cell, err := rb.row.Cell(index)
	if err != nil {
		rb.err = err
		rb.parent.parent.errors = append(rb.parent.parent.errors, err)
		return &CellBuilder{parent: rb, err: err}
	}

	return &CellBuilder{
		cell:   cell,
		parent: rb,
	}
}

// Height sets the row height.
func (rb *RowBuilder) Height(twips int) *RowBuilder {
	if rb.err != nil {
		return rb
	}

	if err := rb.row.SetHeight(twips); err != nil {
		rb.err = err
		rb.parent.parent.errors = append(rb.parent.parent.errors, err)
	}

	return rb
}

// End returns to the TableBuilder.
func (rb *RowBuilder) End() *TableBuilder {
	return rb.parent
}

// CellBuilder provides a fluent API for building table cells.
type CellBuilder struct {
	cell   domain.TableCell
	parent *RowBuilder
	err    error
}

// Text adds text to the cell (creates a paragraph with a run).
func (cb *CellBuilder) Text(text string) *CellBuilder {
	if cb.err != nil {
		return cb
	}

	para, err := cb.cell.AddParagraph()
	if err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
		return cb
	}

	run, err := para.AddRun()
	if err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
		return cb
	}

	if err := run.SetText(text); err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
	}

	return cb
}

// Bold makes the last run in the last paragraph bold.
func (cb *CellBuilder) Bold() *CellBuilder {
	if cb.err != nil {
		return cb
	}

	paragraphs := cb.cell.Paragraphs()
	if len(paragraphs) == 0 {
		cb.err = errors.InvalidState("CellBuilder.Bold", "no paragraphs in cell")
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, cb.err)
		return cb
	}

	runs := paragraphs[len(paragraphs)-1].Runs()
	if len(runs) == 0 {
		cb.err = errors.InvalidState("CellBuilder.Bold", "no runs in paragraph")
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, cb.err)
		return cb
	}

	if err := runs[len(runs)-1].SetBold(true); err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
	}

	return cb
}

// Width sets the cell width.
func (cb *CellBuilder) Width(twips int) *CellBuilder {
	if cb.err != nil {
		return cb
	}

	if err := cb.cell.SetWidth(twips); err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
	}

	return cb
}

// VerticalAlignment sets the cell vertical alignment.
func (cb *CellBuilder) VerticalAlignment(align domain.VerticalAlignment) *CellBuilder {
	if cb.err != nil {
		return cb
	}

	if err := cb.cell.SetVerticalAlignment(align); err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
	}

	return cb
}

// Shading sets the cell background color.
func (cb *CellBuilder) Shading(color domain.Color) *CellBuilder {
	if cb.err != nil {
		return cb
	}

	if err := cb.cell.SetShading(color); err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
	}

	return cb
}

// Merge merges this cell with adjacent cells.
// colspan: number of columns to span (1 = no horizontal merge)
// rowspan: number of rows to span (1 = no vertical merge)
func (cb *CellBuilder) Merge(colspan, rowspan int) *CellBuilder {
	if cb.err != nil {
		return cb
	}

	if err := cb.cell.Merge(colspan, rowspan); err != nil {
		cb.err = err
		cb.parent.parent.parent.errors = append(cb.parent.parent.parent.errors, err)
	}

	return cb
}

// End returns to the RowBuilder.
func (cb *CellBuilder) End() *RowBuilder {
	return cb.parent
}
