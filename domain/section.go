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

// DefaultMargins provides standard 1-inch margins (1440 twips).
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

// Page orientation constants.
const (
	OrientationPortrait  Orientation = iota // Portrait orientation
	OrientationLandscape                    // Landscape orientation
)

// HeaderType represents different header types.
type HeaderType int

// Header type constants for different page scenarios.
const (
	HeaderDefault HeaderType = iota // Default header for all pages
	HeaderFirst                      // Header for first page
	HeaderEven                       // Header for even pages
)

// FooterType represents different footer types.
type FooterType int

// Footer type constants for different page scenarios.
const (
	FooterDefault FooterType = iota // Default footer for all pages
	FooterFirst                      // Footer for first page
	FooterEven                       // Footer for even pages
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
	// ID returns the unique style identifier.
	ID() string

	// Name returns the style display name.
	Name() string

	// Type returns the style type (paragraph or character).
	Type() StyleType

	// BasedOn returns the style ID this style is based on.
	BasedOn() string

	// SetBasedOn sets the parent style.
	SetBasedOn(styleID string) error

	// Next returns the style ID to use for the next paragraph.
	Next() string

	// SetNext sets the next paragraph style.
	SetNext(styleID string) error

	// Font returns the font settings for this style.
	Font() Font

	// SetFont sets the font settings.
	SetFont(font Font) error

	// IsDefault returns whether this is a default style.
	IsDefault() bool

	// SetDefault marks this style as default for its type.
	SetDefault(isDefault bool) error

	// IsCustom returns whether this is a custom (user-defined) style.
	IsCustom() bool
}

// StyleType represents the type of style.
type StyleType int

// Style type constants for different document elements.
const (
	StyleTypeParagraph StyleType = iota // Paragraph style
	StyleTypeCharacter                  // Character/run style
	StyleTypeTable                      // Table style
	StyleTypeNumbering                  // Numbering style
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
