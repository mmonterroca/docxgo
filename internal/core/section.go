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
	"sync"

	"github.com/mmonterroca/docxgo/domain"
	"github.com/mmonterroca/docxgo/internal/manager"
	"github.com/mmonterroca/docxgo/pkg/constants"
	"github.com/mmonterroca/docxgo/pkg/errors"
)

// docxSection implements the Section interface.
type docxSection struct {
	mu            sync.RWMutex
	pageSize      domain.PageSize
	margins       domain.Margins
	orientation   domain.Orientation
	columns       int
	headers       map[domain.HeaderType]*docxHeader
	footers       map[domain.FooterType]*docxFooter
	relationMgr   manager.RelationshipManager
	idGen         manager.IDGenerator
}

// NewSection creates a new section with default settings.
func NewSection(relationMgr manager.RelationshipManager, idGen manager.IDGenerator) domain.Section {
	return &docxSection{
		pageSize:    domain.PageSizeA4,
		margins:     domain.DefaultMargins,
		orientation: domain.OrientationPortrait,
		columns:     1,
		headers:     make(map[domain.HeaderType]*docxHeader),
		footers:     make(map[domain.FooterType]*docxFooter),
		relationMgr: relationMgr,
		idGen:       idGen,
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
		headerType:  headerType,
		paragraphs:  make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		relationMgr: s.relationMgr,
		idGen:       s.idGen,
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
		footerType:  footerType,
		paragraphs:  make([]domain.Paragraph, 0, constants.DefaultParagraphCapacity),
		relationMgr: s.relationMgr,
		idGen:       s.idGen,
	}

	s.footers[footerType] = footer
	return footer, nil
}

// docxHeader implements the Header interface.
type docxHeader struct {
	mu          sync.RWMutex
	headerType  domain.HeaderType
	paragraphs  []domain.Paragraph
	relationMgr manager.RelationshipManager
	idGen       manager.IDGenerator
}

// AddParagraph adds a paragraph to the header.
func (h *docxHeader) AddParagraph() (domain.Paragraph, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	para := NewParagraph(h.relationMgr, h.idGen)
	h.paragraphs = append(h.paragraphs, para)
	return para, nil
}

// Paragraphs returns all paragraphs in the header.
func (h *docxHeader) Paragraphs() []domain.Paragraph {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return defensive copy
	result := make([]domain.Paragraph, len(h.paragraphs))
	copy(result, h.paragraphs)
	return result
}

// docxFooter implements the Footer interface.
type docxFooter struct {
	mu          sync.RWMutex
	footerType  domain.FooterType
	paragraphs  []domain.Paragraph
	relationMgr manager.RelationshipManager
	idGen       manager.IDGenerator
}

// AddParagraph adds a paragraph to the footer.
func (f *docxFooter) AddParagraph() (domain.Paragraph, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	para := NewParagraph(f.relationMgr, f.idGen)
	f.paragraphs = append(f.paragraphs, para)
	return para, nil
}

// Paragraphs returns all paragraphs in the footer.
func (f *docxFooter) Paragraphs() []domain.Paragraph {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return defensive copy
	result := make([]domain.Paragraph, len(f.paragraphs))
	copy(result, f.paragraphs)
	return result
}
