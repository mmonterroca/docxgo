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
	"io"
	"os"

	"github.com/SlideLang/go-docx/domain"
	"github.com/SlideLang/go-docx/internal/manager"
	"github.com/SlideLang/go-docx/internal/serializer"
	"github.com/SlideLang/go-docx/internal/writer"
	xmlstructs "github.com/SlideLang/go-docx/internal/xml"
	"github.com/SlideLang/go-docx/pkg/constants"
	"github.com/SlideLang/go-docx/pkg/errors"
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

	// Build relationships
	rels := &xmlstructs.Relationships{
		Xmlns:         constants.NamespacePackageRels,
		Relationships: []*xmlstructs.Relationship{},
	}

	// TODO: Add relationships from relManager

	// Write document structure
	var coreProps *xmlstructs.CoreProperties // TODO: Add metadata support
	var appProps *xmlstructs.AppProperties   // TODO: Add metadata support

	if err := zipWriter.WriteDocument(xmlDoc, rels, coreProps, appProps); err != nil {
		return 0, errors.WrapWithCode(err, errors.ErrCodeIO, "Document.WriteTo")
	}

	// TODO: Return actual byte count
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
