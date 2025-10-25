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

// Section represents a section in a document.
// Sections allow different page layouts within the same document.
type Section interface {
	// PageSize returns the page size for this section.
	PageSize() PageSize

	// SetPageSize sets the page size.
	SetPageSize(size PageSize) error

	// Margins returns the page margins.
	Margins() Margins

	// SetMargins sets the page margins.
	SetMargins(margins Margins) error

	// Orientation returns the page orientation.
	Orientation() Orientation

	// SetOrientation sets the page orientation.
	SetOrientation(orient Orientation) error

	// Columns returns the number of columns.
	Columns() int

	// SetColumns sets the number of columns.
	SetColumns(count int) error

	// Header returns the header for this section.
	Header(headerType HeaderType) (Header, error)

	// Footer returns the footer for this section.
	Footer(footerType FooterType) (Footer, error)
}

// PageSize represents page dimensions in twips.
type PageSize struct {
	Width  int
	Height int
}

// Common page sizes (in twips: 1440 = 1 inch)
var (
	PageSizeA4      = PageSize{Width: 11906, Height: 16838} // 210mm x 297mm
	PageSizeLetter  = PageSize{Width: 12240, Height: 15840} // 8.5" x 11"
	PageSizeLegal   = PageSize{Width: 12240, Height: 20160} // 8.5" x 14"
	PageSizeA3      = PageSize{Width: 16838, Height: 23811} // 297mm x 420mm
	PageSizeTableid = PageSize{Width: 15840, Height: 24480} // 11" x 17"
)

// Margins represents page margins in twips.
type Margins struct {
	Top    int
	Right  int
	Bottom int
	Left   int
	Header int // Distance from top edge to header
	Footer int // Distance from bottom edge to footer
}

// Default margins (1 inch = 1440 twips)
var DefaultMargins = Margins{
	Top:    1440,
	Right:  1440,
	Bottom: 1440,
	Left:   1440,
	Header: 720,
	Footer: 720,
}

// Orientation represents page orientation.
type Orientation int

const (
	OrientationPortrait Orientation = iota
	OrientationLandscape
)

// HeaderType represents different header types.
type HeaderType int

const (
	HeaderDefault HeaderType = iota // Default header for all pages
	HeaderFirst                     // Header for first page
	HeaderEven                      // Header for even pages
)

// FooterType represents different footer types.
type FooterType int

const (
	FooterDefault FooterType = iota // Default footer for all pages
	FooterFirst                     // Footer for first page
	FooterEven                      // Footer for even pages
)

// Header represents a page header.
type Header interface {
	// AddParagraph adds a paragraph to the header.
	AddParagraph() (Paragraph, error)

	// Paragraphs returns all paragraphs in the header.
	Paragraphs() []Paragraph
}

// Footer represents a page footer.
type Footer interface {
	// AddParagraph adds a paragraph to the footer.
	AddParagraph() (Paragraph, error)

	// Paragraphs returns all paragraphs in the footer.
	Paragraphs() []Paragraph
}

// Style represents a paragraph or character style.
type Style interface {
	// Name returns the style name.
	Name() string

	// Type returns the style type (paragraph or character).
	Type() StyleType

	// BasedOn returns the style this style is based on.
	BasedOn() string

	// Next returns the style to use for the next paragraph.
	Next() string
}

// StyleType represents the type of style.
type StyleType int

const (
	StyleTypeParagraph StyleType = iota
	StyleTypeCharacter
	StyleTypeTable
	StyleTypeNumbering
)

// Field represents a field in a document (TOC, page number, etc.)
type Field interface {
	// Type returns the field type.
	Type() FieldType

	// Code returns the field code.
	Code() string

	// SetCode sets the field code.
	SetCode(code string) error

	// Result returns the field result (calculated value).
	Result() string

	// Update recalculates the field result.
	Update() error
}
