// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 Misael Monterroca
//
// See LICENSE for the full copyright notice, including predecessor authors,
// and CREDITS.md for the project genealogy.

package core

import (
	"sync"

	"github.com/mmonterroca/docxgo/v2/domain"
	"github.com/mmonterroca/docxgo/v2/internal/manager"
	"github.com/mmonterroca/docxgo/v2/pkg/constants"
	"github.com/mmonterroca/docxgo/v2/pkg/errors"
)

// docxSection implements the Section interface.
type docxSection struct {
	mu           sync.RWMutex
	pageSize     domain.PageSize
	margins      domain.Margins
	orientation  domain.Orientation
	columns      int
	headers      map[domain.HeaderType]*docxHeader
	footers      map[domain.FooterType]*docxFooter
	relationMgr  *manager.RelationshipManager
	idGen        *manager.IDGenerator
	mediaManager *manager.MediaManager
}

// NewSection creates a new section with default settings.
func NewSection(relationMgr *manager.RelationshipManager, idGen *manager.IDGenerator, mediaManager *manager.MediaManager) domain.Section {
	return &docxSection{
		pageSize:     domain.PageSizeA4,
		margins:      domain.DefaultMargins,
		orientation:  domain.OrientationPortrait,
		columns:      1,
		headers:      make(map[domain.HeaderType]*docxHeader),
		footers:      make(map[domain.FooterType]*docxFooter),
		relationMgr:  relationMgr,
		idGen:        idGen,
		mediaManager: mediaManager,
	}
}

// PageSize returns the page size for this section.
func (s *docxSection) PageSize() domain.PageSize {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pageSize
}

// SetPageSize sets the page size.
func (s *docxSection) SetPageSize(size domain.PageSize) error {
	if size.Width <= 0 || size.Height <= 0 {
		return errors.NewValidationError(
			"SetPageSize",
			"page size",
			size,
			"width and height must be positive",
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pageSize = size
	return nil
}

// Margins returns the page margins.
func (s *docxSection) Margins() domain.Margins {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.margins
}

// SetMargins sets the page margins.
func (s *docxSection) SetMargins(margins domain.Margins) error {
	if margins.Top < 0 || margins.Right < 0 || margins.Bottom < 0 || margins.Left < 0 {
		return errors.NewValidationError(
			"SetMargins",
			"margins",
			margins,
			"margins cannot be negative",
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.margins = margins
	return nil
}

// Orientation returns the page orientation.
func (s *docxSection) Orientation() domain.Orientation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orientation
}

// SetOrientation sets the page orientation.
func (s *docxSection) SetOrientation(orient domain.Orientation) error {
	if orient != domain.OrientationPortrait && orient != domain.OrientationLandscape {
		return errors.NewValidationError(
			"SetOrientation",
			"orientation",
			orient,
			"must be Portrait or Landscape",
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.orientation = orient
	return nil
}

// Columns returns the number of columns.
func (s *docxSection) Columns() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.columns
}

// SetColumns sets the number of columns.
func (s *docxSection) SetColumns(count int) error {
	if count < 1 || count > constants.MaxColumns {
		return errors.NewValidationError(
			"SetColumns",
			"columns",
			count,
			"must be between 1 and "+string(rune(constants.MaxColumns)),
		)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.columns = count
	return nil
}

// Header returns the header for this section.
func (s *docxSection) Header(headerType domain.HeaderType) (domain.Header, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if header already exists
	if header, exists := s.headers[headerType]; exists {
		return header, nil
	}

	// Create new header
	header := &docxHeader{
		headerType:   headerType,
		paragraphs:   make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		tables:       make([]domain.Table, 0),
		blocks:       make([]domain.Block, 0, constants.DefaultParagraphCapacity),
		relationMgr:  s.relationMgr,
		idGen:        s.idGen,
		mediaManager: s.mediaManager,
	}

	s.headers[headerType] = header
	return header, nil
}

// Footer returns the footer for this section.
func (s *docxSection) Footer(footerType domain.FooterType) (domain.Footer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if footer already exists
	if footer, exists := s.footers[footerType]; exists {
		return footer, nil
	}

	// Create new footer
	footer := &docxFooter{
		footerType:   footerType,
		paragraphs:   make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		tables:       make([]domain.Table, 0),
		blocks:       make([]domain.Block, 0, constants.DefaultParagraphCapacity),
		relationMgr:  s.relationMgr,
		idGen:        s.idGen,
		mediaManager: s.mediaManager,
	}

	s.footers[footerType] = footer
	return footer, nil
}

// HeadersAll returns a copy of all headers defined for the section.
func (s *docxSection) HeadersAll() map[domain.HeaderType]domain.Header {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[domain.HeaderType]domain.Header, len(s.headers))
	for k, v := range s.headers {
		result[k] = v
	}
	return result
}

// FootersAll returns a copy of all footers defined for the section.
func (s *docxSection) FootersAll() map[domain.FooterType]domain.Footer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[domain.FooterType]domain.Footer, len(s.footers))
	for k, v := range s.footers {
		result[k] = v
	}
	return result
}

// docxHeader implements the Header interface.
type docxHeader struct {
	mu           sync.RWMutex
	headerType   domain.HeaderType
	paragraphs   []domain.Paragraph
	tables       []domain.Table
	blocks       []domain.Block
	relationMgr  *manager.RelationshipManager
	idGen        *manager.IDGenerator
	relID        string
	targetPath   string
	mediaManager *manager.MediaManager
}

// AddParagraph adds a paragraph to the header.
func (h *docxHeader) AddParagraph() (domain.Paragraph, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.idGen.NextParagraphID()
	para := NewParagraph(id, h.idGen, h.relationMgr, h.mediaManager)
	markHeaderFooterParagraph(para)
	h.paragraphs = append(h.paragraphs, para)
	h.blocks = append(h.blocks, domain.Block{Paragraph: para})
	return para, nil
}

// Paragraphs returns all top-level paragraphs in the header.
func (h *docxHeader) Paragraphs() []domain.Paragraph {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return defensive copy
	result := make([]domain.Paragraph, len(h.paragraphs))
	copy(result, h.paragraphs)
	return result
}

// AddTable adds a table with the given number of rows and columns to the header.
func (h *docxHeader) AddTable(rows, cols int) (domain.Table, error) {
	if rows < constants.MinTableRows || rows > constants.MaxTableRows {
		return nil, errors.InvalidArgument("Header.AddTable", "rows", rows,
			"rows must be between 1 and 1000")
	}
	if cols < constants.MinTableCols || cols > constants.MaxTableCols {
		return nil, errors.InvalidArgument("Header.AddTable", "cols", cols,
			"columns must be between 1 and 63")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.idGen.NextTableID()
	tbl := NewTable(id, rows, cols, h.idGen, h.relationMgr, h.mediaManager)
	markHeaderFooterTable(tbl)
	h.tables = append(h.tables, tbl)
	h.blocks = append(h.blocks, domain.Block{Table: tbl})
	return tbl, nil
}

// RemoveLastTable drops the most recently added table from both tables and
// blocks, but only if it is still the very last block -- i.e. nothing else
// was added to the header since. It exists for the reader's best-effort
// header/footer hydration (see hydrateTable in internal/reader): AddTable
// attaches a table to the header immediately, before its rows/cells are
// populated, so a hydration failure partway through a table (e.g. an
// unparseable gridSpan) must undo that attach -- otherwise the header is
// left holding a table whose later rows or cells were silently never
// populated, even though the caller tolerated the error and got no failure
// back from OpenDocument. Safe under hydration's synchronous, single-caller
// use: hydrateTable returns immediately on error, before anything else can
// be added to the container.
func (h *docxHeader) RemoveLastTable() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.tables) == 0 || len(h.blocks) == 0 {
		return
	}
	last := h.tables[len(h.tables)-1]
	if h.blocks[len(h.blocks)-1].Table != last {
		return
	}
	h.tables = h.tables[:len(h.tables)-1]
	h.blocks = h.blocks[:len(h.blocks)-1]
}

// Tables returns all top-level tables in the header.
func (h *docxHeader) Tables() []domain.Table {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]domain.Table, len(h.tables))
	copy(result, h.tables)
	return result
}

// Blocks returns every top-level element in the header, in insertion order.
func (h *docxHeader) Blocks() []domain.Block {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]domain.Block, len(h.blocks))
	copy(result, h.blocks)
	return result
}

// RelationshipID returns the relationship ID associated with this header.
func (h *docxHeader) RelationshipID() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.relID
}

// TargetPath returns the target part path for this header within the DOCX package.
func (h *docxHeader) TargetPath() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.targetPath
}

// setRelationship stores the relationship metadata for the header.
func (h *docxHeader) setRelationship(relID, target string) {
	h.mu.Lock()
	h.relID = relID
	h.targetPath = target
	h.mu.Unlock()
}

// SetExistingRelationship seeds the header with relationship metadata read from an existing document.
func (h *docxHeader) SetExistingRelationship(relID, target string) {
	if h == nil {
		return
	}
	h.setRelationship(relID, target)
}

// docxFooter implements the Footer interface.
type docxFooter struct {
	mu           sync.RWMutex
	footerType   domain.FooterType
	paragraphs   []domain.Paragraph
	tables       []domain.Table
	blocks       []domain.Block
	relationMgr  *manager.RelationshipManager
	idGen        *manager.IDGenerator
	relID        string
	targetPath   string
	mediaManager *manager.MediaManager
}

// AddParagraph adds a paragraph to the footer.
func (f *docxFooter) AddParagraph() (domain.Paragraph, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.idGen.NextParagraphID()
	para := NewParagraph(id, f.idGen, f.relationMgr, f.mediaManager)
	markHeaderFooterParagraph(para)
	f.paragraphs = append(f.paragraphs, para)
	f.blocks = append(f.blocks, domain.Block{Paragraph: para})
	return para, nil
}

// Paragraphs returns all top-level paragraphs in the footer.
func (f *docxFooter) Paragraphs() []domain.Paragraph {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return defensive copy
	result := make([]domain.Paragraph, len(f.paragraphs))
	copy(result, f.paragraphs)
	return result
}

// AddTable adds a table with the given number of rows and columns to the footer.
func (f *docxFooter) AddTable(rows, cols int) (domain.Table, error) {
	if rows < constants.MinTableRows || rows > constants.MaxTableRows {
		return nil, errors.InvalidArgument("Footer.AddTable", "rows", rows,
			"rows must be between 1 and 1000")
	}
	if cols < constants.MinTableCols || cols > constants.MaxTableCols {
		return nil, errors.InvalidArgument("Footer.AddTable", "cols", cols,
			"columns must be between 1 and 63")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	id := f.idGen.NextTableID()
	tbl := NewTable(id, rows, cols, f.idGen, f.relationMgr, f.mediaManager)
	markHeaderFooterTable(tbl)
	f.tables = append(f.tables, tbl)
	f.blocks = append(f.blocks, domain.Block{Table: tbl})
	return tbl, nil
}

// RemoveLastTable is docxHeader.RemoveLastTable's footer twin -- see there.
func (f *docxFooter) RemoveLastTable() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.tables) == 0 || len(f.blocks) == 0 {
		return
	}
	last := f.tables[len(f.tables)-1]
	if f.blocks[len(f.blocks)-1].Table != last {
		return
	}
	f.tables = f.tables[:len(f.tables)-1]
	f.blocks = f.blocks[:len(f.blocks)-1]
}

// Tables returns all top-level tables in the footer.
func (f *docxFooter) Tables() []domain.Table {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]domain.Table, len(f.tables))
	copy(result, f.tables)
	return result
}

// Blocks returns every top-level element in the footer, in insertion order.
func (f *docxFooter) Blocks() []domain.Block {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make([]domain.Block, len(f.blocks))
	copy(result, f.blocks)
	return result
}

// RelationshipID returns the relationship ID associated with this footer.
func (f *docxFooter) RelationshipID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.relID
}

// TargetPath returns the target part path for this footer within the DOCX package.
func (f *docxFooter) TargetPath() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.targetPath
}

// setRelationship stores the relationship metadata for the footer.
func (f *docxFooter) setRelationship(relID, target string) {
	f.mu.Lock()
	f.relID = relID
	f.targetPath = target
	f.mu.Unlock()
}

// SetExistingRelationship seeds the footer with relationship metadata read from an existing document.
func (f *docxFooter) SetExistingRelationship(relID, target string) {
	if f == nil {
		return
	}
	f.setRelationship(relID, target)
}
