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



package core

import (
	"io"
	"os"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/manager"
	"github.com/mmonterroca/docxgo/internal/serializer"
	"github.com/mmonterroca/docxgo/internal/writer"
	xmlstructs "github.com/mmonterroca/docxgo/internal/xml"
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
}

// NewDocument creates a new Document.
func NewDocument() domain.Document {
	idGen := manager.NewIDGenerator()
	return &document{
		paragraphs:   make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		tables:       make([]domain.Table, 0, constants.DefaultTableCapacity),
		sections:     make([]domain.Section, 0, 1),
		metadata:     &domain.Metadata{},
		idGen:        idGen,
		relManager:   manager.NewRelationshipManager(idGen),
		mediaManager: manager.NewMediaManager(idGen),
	}
}

// AddParagraph adds a new paragraph to the document.
func (d *document) AddParagraph() (domain.Paragraph, error) {
	id := d.idGen.NextParagraphID()
	para := NewParagraph(id, d.idGen, d.relManager)
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
	table := NewTable(id, rows, cols, d.idGen, d.relManager)
	d.tables = append(d.tables, table)
	return table, nil
}

// AddSection adds a new section to the document.
func (d *document) AddSection() (domain.Section, error) {
	// TODO: Implement section creation
	return nil, errors.Unsupported("Document.AddSection", "sections not yet implemented")
}

// DefaultSection returns the default (first) section of the document.
func (d *document) DefaultSection() (domain.Section, error) {
	if len(d.sections) == 0 {
		// Create default section if it doesn't exist
		section := NewSection(*d.relManager, *d.idGen)
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

// WriteTo writes the document to the provided writer in .docx format.
func (d *document) WriteTo(w io.Writer) (int64, error) {
	// Serialize domain objects to XML structures
	ser := serializer.NewDocumentSerializer()
	xmlDoc := ser.SerializeDocument(d)

	// Create ZIP writer
	zipWriter := writer.NewZipWriter(w)
	defer zipWriter.Close()

	// Build relationships from relationship manager
	rels := d.relManager.ToXML()

	// Serialize metadata
	coreProps := ser.SerializeCoreProperties(d.metadata)
	appProps := ser.SerializeAppProperties(d)

	// Write document structure
	if err := zipWriter.WriteDocument(xmlDoc, rels, coreProps, appProps); err != nil {
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
