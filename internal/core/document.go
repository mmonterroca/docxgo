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

// Package core provides concrete implementations of domain interfaces for go-docx v2.
//
// This package contains the core document model implementations including:
// - Document: The main document structure
// - Paragraph: Paragraph implementation with formatting
// - Run: Text run implementation with character formatting
// - Table: Table implementation with cells and rows
// - Section: Section implementation with page settings
// - Image: Image embedding and positioning
// - Field: Field implementation (TOC, page numbers, etc.)
//
// These implementations handle the business logic and coordinate with
// internal managers (ID generation, relationships, media, styles) and
// serialization to XML structures.
package core

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/manager"
	"github.com/mmonterroca/docxgo/internal/serializer"
	"github.com/mmonterroca/docxgo/internal/writer"
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// document implements the domain.Document interface.
type document struct {
	paragraphs   []domain.Paragraph
	tables       []domain.Table
	sections     []domain.Section
	metadata     *domain.Metadata
	idGen        *manager.IDGenerator
	relManager   *manager.RelationshipManager
	mediaManager *manager.MediaManager
	styleManager domain.StyleManager
	headerCount  int
	footerCount  int
}

// NewDocument creates a new Document.
func NewDocument() domain.Document {
	idGen := manager.NewIDGenerator()
	relManager := manager.NewRelationshipManager(idGen)
	doc := &document{
		paragraphs:   make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		tables:       make([]domain.Table, 0, constants.DefaultTableCapacity),
		sections:     make([]domain.Section, 0, 1),
		metadata:     &domain.Metadata{},
		idGen:        idGen,
		relManager:   relManager,
		mediaManager: manager.NewMediaManager(idGen),
		styleManager: manager.NewStyleManager(),
	}

	// Ensure core document relationships exist (styles, fonts, theme)
	doc.ensureDefaultRelationships()

	return doc
}

// AddParagraph adds a new paragraph to the document.
func (d *document) AddParagraph() (domain.Paragraph, error) {
	id := d.idGen.NextParagraphID()
	para := NewParagraph(id, d.idGen, d.relManager, d.mediaManager)
	d.paragraphs = append(d.paragraphs, para)
	return para, nil
}

// AddTable adds a new table with the specified dimensions.
func (d *document) AddTable(rows, cols int) (domain.Table, error) {
	if rows < constants.MinTableRows || rows > constants.MaxTableRows {
		return nil, errors.InvalidArgument("Document.AddTable", "rows", rows,
			"rows must be between 1 and 1000")
	}
	if cols < constants.MinTableCols || cols > constants.MaxTableCols {
		return nil, errors.InvalidArgument("Document.AddTable", "cols", cols,
			"columns must be between 1 and 63")
	}

	id := d.idGen.NextTableID()
	table := NewTable(id, rows, cols, d.idGen, d.relManager, d.mediaManager)
	d.tables = append(d.tables, table)
	return table, nil
}

// AddSection adds a new section to the document.
// Note: Currently only DefaultSection() is fully supported.
// Multi-section documents will be implemented in a future release.
func (d *document) AddSection() (domain.Section, error) {
	return nil, errors.Unsupported("Document.AddSection", "multi-section documents not yet implemented - use DefaultSection() instead")
}

// AddPageBreak adds a page break to the document.
func (d *document) AddPageBreak() error {
	// Create a new paragraph
	para, err := d.AddParagraph()
	if err != nil {
		return err
	}

	// Add a run with a page break
	run, err := para.AddRun()
	if err != nil {
		return err
	}

	return run.AddBreak(domain.BreakTypePage)
}

// DefaultSection returns the default (first) section of the document.
func (d *document) DefaultSection() (domain.Section, error) {
	if len(d.sections) == 0 {
		// Create default section if it doesn't exist
		section := NewSection(d.relManager, d.idGen, d.mediaManager)
		d.sections = append(d.sections, section)
		return section, nil
	}
	return d.sections[0], nil
}

// Paragraphs returns all paragraphs in the document.
func (d *document) Paragraphs() []domain.Paragraph {
	// Return a copy to prevent external modification
	paras := make([]domain.Paragraph, len(d.paragraphs))
	copy(paras, d.paragraphs)
	return paras
}

// Tables returns all tables in the document.
func (d *document) Tables() []domain.Table {
	tables := make([]domain.Table, len(d.tables))
	copy(tables, d.tables)
	return tables
}

// Sections returns all sections in the document.
func (d *document) Sections() []domain.Section {
	sections := make([]domain.Section, len(d.sections))
	copy(sections, d.sections)
	return sections
}

// generateHeadingBookmarks generates bookmarks for all headings in the document.
// This is required for Table of Contents (TOC) fields to work properly.
// Bookmarks are named _Toc{sequential_number} and only applied to paragraphs with Heading styles.
func (d *document) generateHeadingBookmarks() {
	bookmarkCounter := 0

	for _, para := range d.paragraphs {
		// Type assert to access internal paragraph methods
		if p, ok := para.(*paragraph); ok {
			styleName := p.StyleName()

			// Check if this paragraph has a Heading style
			if strings.HasPrefix(styleName, "Heading") {
				bookmarkID := fmt.Sprintf("%d", bookmarkCounter)
				bookmarkName := fmt.Sprintf("_Toc%d", bookmarkCounter)
				p.SetBookmark(bookmarkID, bookmarkName)
				bookmarkCounter++
			}
		}
	}
}

// prepareHeaderFooterRelationships ensures that every header/footer defined in the
// document has an associated relationship and target part name within the DOCX
// package. This must run before serialization so both section references and the
// document relationships list are consistent.
func (d *document) prepareHeaderFooterRelationships() {
	for _, sec := range d.sections {
		coreSection, ok := sec.(*docxSection)
		if !ok {
			continue
		}

		coreSection.mu.Lock()

		for _, header := range coreSection.headers {
			if header == nil {
				continue
			}

			if header.TargetPath() == "" {
				d.headerCount++
				target := fmt.Sprintf("header%d.xml", d.headerCount)
				header.setRelationship(header.RelationshipID(), target)
			}

			if header.RelationshipID() == "" {
				if relID, err := d.relManager.AddHeader(header.TargetPath()); err == nil {
					header.setRelationship(relID, header.TargetPath())
				}
			}
		}

		for _, footer := range coreSection.footers {
			if footer == nil {
				continue
			}

			if footer.TargetPath() == "" {
				d.footerCount++
				target := fmt.Sprintf("footer%d.xml", d.footerCount)
				footer.setRelationship(footer.RelationshipID(), target)
			}

			if footer.RelationshipID() == "" {
				if relID, err := d.relManager.AddFooter(footer.TargetPath()); err == nil {
					footer.setRelationship(relID, footer.TargetPath())
				}
			}
		}

		coreSection.mu.Unlock()
	}
}

// ensureDefaultRelationships guarantees that the DOCX package contains the
// required relationships for styles, fonts, and theme assets. Without these
// entries Word falls back to implicit defaults and style assignments appear as
// "Normal", which breaks features such as the Table of Contents.
func (d *document) ensureDefaultRelationships() {
	if d == nil || d.relManager == nil {
		return
	}

	// Track existing relationship targets to avoid duplicates when called
	// multiple times (e.g. SaveAs after WriteTo).
	existing := make(map[string]bool)
	for _, rel := range d.relManager.All() {
		existing[rel.Target] = true
	}

	baseRels := []struct {
		relType string
		target  string
	}{
		{constants.RelTypeStyles, "styles.xml"},
		{constants.RelTypeFontTable, "fontTable.xml"},
		{constants.RelTypeTheme, "theme/theme1.xml"},
	}

	for _, rel := range baseRels {
		if existing[rel.target] {
			continue
		}

		// Ignore the error because the inputs are fixed and validated. In the
		// unlikely event of a failure we still prefer to continue writing the
		// document instead of aborting.
		_, _ = d.relManager.Add(rel.relType, rel.target, "Internal")
	}
}

// WriteTo writes the document to the provided writer in .docx format.
func (d *document) WriteTo(w io.Writer) (int64, error) {
	if len(d.sections) == 0 {
		if _, err := d.DefaultSection(); err != nil {
			return 0, errors.Wrap(err, "Document.WriteTo")
		}
	}

	// Generate bookmarks for headings (needed for TOC)
	d.generateHeadingBookmarks()

	// Ensure headers and footers have relationships/targets before serialization
	d.prepareHeaderFooterRelationships()

	// Ensure required base relationships are present before serialization
	d.ensureDefaultRelationships()

	// Serialize domain objects to XML structures
	ser := serializer.NewDocumentSerializer()
	xmlDoc := ser.SerializeDocument(d)
	headers, footers := ser.SerializeSectionParts(d)

	// Create ZIP writer
	zipWriter := writer.NewZipWriter(w)
	defer zipWriter.Close()

	// Build relationships from relationship manager
	rels := d.relManager.ToXML()

	// Serialize metadata
	coreProps := ser.SerializeCoreProperties(d.metadata)
	appProps := ser.SerializeAppProperties(d)

	// Serialize styles
	styles := ser.SerializeStyles(d.styleManager)

	mediaFiles := d.mediaManager.All()

	// Write document structure
	if err := zipWriter.WriteDocument(xmlDoc, rels, coreProps, appProps, styles, mediaFiles, headers, footers); err != nil {
		return 0, errors.WrapWithCode(err, errors.ErrCodeIO, "Document.WriteTo")
	}

	// Get byte count from writer if available
	// For now, return 0 as ZipWriter doesn't track total bytes
	// This could be enhanced by wrapping the writer with a counting writer
	return 0, nil
}

// SaveAs saves the document to the specified file path.
func (d *document) SaveAs(path string) error {
	if path == "" {
		return errors.InvalidArgument("Document.SaveAs", "path", path, "path cannot be empty")
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeIO, "Document.SaveAs")
	}
	defer file.Close()

	// Write document
	_, err = d.WriteTo(file)
	if err != nil {
		return errors.Wrap(err, "Document.SaveAs")
	}

	return nil
}

// Validate checks if the document structure is valid.
func (d *document) Validate() error {
	// Basic validation
	if len(d.paragraphs) == 0 && len(d.tables) == 0 {
		return errors.InvalidState("Document.Validate", "document is empty")
	}

	// Validate each paragraph
	for i, para := range d.paragraphs {
		if para == nil {
			return errors.InvalidState("Document.Validate",
				"paragraph at index "+string(rune(i))+" is nil")
		}
	}

	// Validate each table
	for i, table := range d.tables {
		if table == nil {
			return errors.InvalidState("Document.Validate",
				"table at index "+string(rune(i))+" is nil")
		}
	}

	return nil
}

// Metadata returns the document's metadata.
func (d *document) Metadata() *domain.Metadata {
	return d.metadata
}

// SetMetadata updates the document's metadata.
func (d *document) SetMetadata(meta *domain.Metadata) error {
	if meta == nil {
		return errors.InvalidArgument("Document.SetMetadata", "meta", meta, "metadata cannot be nil")
	}
	d.metadata = meta
	return nil
}

// StyleManager returns the style manager for this document.
func (d *document) StyleManager() domain.StyleManager {
	return d.styleManager
}
